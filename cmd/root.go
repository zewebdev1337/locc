// cmd/root.go
package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	outputFile      string
	configFile      string
	enableStores    bool
	enableDocuments bool
	initLocalConfig bool
	verbose         bool
)

// TODO: Fix/Define include behavior
// TODO:➚ Complimentary to ↑, review config file structure and modify if necessary

var rootCmd = &cobra.Command{
	Use:   "locc",
	Short: "Count lines of code in a project",
	Long: `locc (Lines of Code Counter) is a tool that scans the current directory and its subdirectories for code files,
counts the number of non-empty lines in each file, and outputs the result.

It uses a configuration system that combines global and local settings:
- Global configuration: Stored in ~/.locc.yaml
- Local configuration: Defaults to ./.locc.yaml but can be specified with the --config flag

Features:
- Language detection based on file extensions
- Simple, filename-based rules available for folder and file inclusion and exclusion system
- Complex, wordlist-based rules available for file inclusion and exclusion system
- Maximum file size limit
- Verbose output option

Configuration:
The configuration file should be in YAML format and can include:
- languages: Map of language configurations (extensions and comment syntax)
- stores: Map of data store configurations (extensions and comment syntax)
- documents: Map of document/plain text configurations (extensions and comment syntax)
- exclusions: Map of file/directory exclusions (global and language-specific)
- max_file_size: Maximum file size to process (in bytes)

For more detailed information, please refer to the documentation.`,
	Run: func(cmd *cobra.Command, args []string) {
		if initLocalConfig {
			err := runInit("./.locc.yaml")
			if err != nil {
				log.Fatal(err)
			}
			return // Exit after initialization
		}
		globalConfig, err := loadGlobalConfig()
		if err != nil {
			log.Fatal(err)
		}
		localConfig, err := loadLocalConfig(configFile)
		if err != nil {
			log.Fatal(err)
		}

		config := mergeConfigs(globalConfig, localConfig)
		processFilters(config)

		err = countLinesOfCode(config, enableStores, enableDocuments)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	},
}

// Execute is the entry point for the locc tool.
// It calls the Execute method of the rootCmd object, which starts the command line interface.
// If an error occurs, it prints the error message and exits with a non-zero status.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

// buildFileList is a function that constructs a list of files to process based on the configuration and the current working directory.
// It takes a configuration object and the root directory as input.
// It returns a slice of strings containing the paths of the files to process and an error if one occurs.
func buildFileList(config *Config, rootDir string) ([]string, error) {
	// Initialize a slice to store the paths of the files to process
	var filesToProcess []string

	// Use the filepath.Walk function to traverse the directory tree rooted at rootDir
	// For each file or directory encountered, the function calls the anonymous function provided as the second argument
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		// If an error occurs, return it
		if err != nil {
			return err
		}

		// Get the relative path of the file or directory
		relPath, _ := filepath.Rel(rootDir, path)

		// If the file is a directory
		if info.IsDir() {
			// Check if the directory should be excluded based on the configuration
			if shouldExcludeDir(config, relPath) {
				// If the directory should be excluded, skip it and its subdirectories
				return filepath.SkipDir
			}
			// If the directory should not be excluded, continue traversing it
			return nil
		}

		// If the file is not a directory
		// Check if the file should be included based on the configuration
		if shouldIncludeFile(config, relPath, info, enableStores, enableDocuments) {
			// If the file should be included, add its path to the slice of files to process
			filesToProcess = append(filesToProcess, path)
		}

		// Continue traversing the directory tree
		return nil
	})

	// If an error occurs during the traversal, return it
	if err != nil {
		return nil, fmt.Errorf("failed to build file list: %w", err)
	}

	// Return the slice of files to process and nil to indicate that the function completed successfully
	return filesToProcess, nil
}

// shouldExcludeDir is a function that checks whether a directory should be excluded from the process based on the configuration.
// It takes a configuration object and the relative path of the directory as input.
// It returns a boolean value indicating whether the directory should be excluded.
func shouldExcludeDir(config *Config, relPath string) bool {
	// Check global exclusions
	// If the configuration contains a map of exclusions for the "locc" key,
	// check if the relative path of the directory matches any of the exclusions.
	// If it does, return true to indicate that the directory should be excluded.
	if exclusions, ok := config.Excludes["locc"]; ok {
		if containsPath(exclusions, relPath) {
			return true
		}
	}

	// Check language-specific exclusions
	// Iterate over the map of exclusions in the configuration.
	// For each language, check if the relative path of the directory matches any of the exclusions for that language.
	// If it does, return true to indicate that the directory should be excluded.
	for _, exclusions := range config.Excludes {
		if containsPath(exclusions, relPath) {
			return true
		}
	}

	// If the directory does not match any of the exclusions, return false to indicate that it should not be excluded.
	return false
}

func shouldIncludeFile(config *Config, relPath string, info os.FileInfo, enableStores, enableDocuments bool) bool {
	if config.MaxFileSize > 0 && info.Size() > config.MaxFileSize {
		return false
	}

	fileName := filepath.Base(relPath)
	lang, _ := detectLanguage(relPath, config, enableStores, enableDocuments)

	// Check global includes
	if globalIncludes, ok := config.Includes["locc"]; ok {
		if matchesFilter(globalIncludes, fileName, relPath) {
			return true
		}
	}

	// Check language-specific includes
	if includes, ok := config.Includes[lang]; ok {
		if matchesFilter(includes, fileName, relPath) {
			return true
		}
	}

	// Check global excludes
	if globalExcludes, ok := config.Excludes["locc"]; ok {
		if matchesFilter(globalExcludes, fileName, relPath) {
			return false
		}
	}

	// Check language-specific excludes
	if excludes, ok := config.Excludes[lang]; ok {
		if matchesFilter(excludes, fileName, relPath) {
			return false
		}
	}

	return lang != ""
}

func matchesFilter(filter interface{}, fileName, relPath string) bool {
	switch v := filter.(type) {
	case map[string][]string:
		if wordlist, ok := v[fileName]; ok {
			if len(wordlist) == 0 {
				return true
			}
			content, err := os.ReadFile(relPath)
			if err != nil {
				return false
			}
			for _, word := range wordlist {
				if strings.Contains(string(content), word) {
					return true
				}
			}
		}
	case []string:
		for _, pattern := range v {
			if matchesExclusion(pattern, fileName) {
				return true
			}
		}
	}
	return false
}

// containsPath is a function that checks whether a path matches any exclusion pattern in a configuration.
// It takes a configuration object and a path as input.
// The configuration object can be a map of exclusions for a specific language, a map of exclusions for the "locc" key, or a slice of exclusions.
// The function returns a boolean value indicating whether the path matches any of the exclusions.
func containsPath(exclusions interface{}, path string) bool {
	// Switch on the type of the exclusions object.
	switch v := exclusions.(type) {
	// If the exclusions object is a map of exclusions for a specific language, iterate over the keys of the map.
	case map[string]FileExclusion:
		for key := range v {
			// If the path matches the exclusion pattern, return true.
			if matchesExclusion(key, path) {
				return true
			}
		}
	// If the exclusions object is a map of exclusions for the "locc" key, iterate over the keys of the map.
	case map[string]interface{}:
		for key := range v {
			// If the path matches the exclusion pattern, return true.
			if matchesExclusion(key, path) {
				return true
			}
		}
	// If the exclusions object is a slice of exclusions, iterate over the slice.
	case []interface{}:
		for _, item := range v {
			// If the item is a string, check if it matches the exclusion pattern.
			if str, ok := item.(string); ok {
				if matchesExclusion(str, path) {
					// If the path matches the exclusion pattern, return true.
					return true
				}
			}
		}
	}
	// If the path does not match any of the exclusions, return false.
	return false
}

// matchesExclusion is a function that checks whether a path matches an exclusion pattern.
// It takes a pattern and a path as input.
// The pattern can be a directory path with a trailing slash, indicating that all files and subdirectories within that directory should be excluded,
// or it can be a file path, indicating that that specific file should be excluded.
// The function returns a boolean value indicating whether the path matches the exclusion pattern.
func matchesExclusion(pattern, path string) bool {
	// If the pattern ends with a slash, it indicates a directory path.
	// In this case, the function checks whether the path starts with the pattern (indicating that it is within the excluded directory)
	// or whether the path is equal to the pattern with the trailing slash removed (indicating that it is the excluded directory itself).
	if strings.HasSuffix(pattern, "/") {
		return strings.HasPrefix(path, pattern) || path == strings.TrimSuffix(pattern, "/")
	}
	// If the pattern does not end with a slash, it indicates a file path.
	// In this case, the function checks whether the path is equal to the pattern.
	return pattern == path
}

// detectLanguage is a function that detects the language of a file based on its extension and the configuration.
// It takes a file name and a configuration object as input.
// It returns the language of the file and the comment syntax for that language.
func detectLanguage(filename string, config *Config, enableStores, enableDocuments bool) (string, []string) {
	// Get the extension of the file from its name and convert it to lower case.
	ext := strings.ToLower(filepath.Ext(filename))

	// Iterate over the map of languages in the configuration.
	for lang, langConfig := range config.Languages {
		// For each language, iterate over the slice of extensions for that language.
		for _, e := range langConfig.Extensions {
			// If the extension of the file matches an extension for the language, return the language and the comment syntax for that language.
			if e == ext {
				return lang, langConfig.Comment
			}
		}
	}

	// Check stores if enabled
	if enableStores {
		for store, storeConfig := range config.Stores {
			for _, e := range storeConfig.Extensions {
				if e == ext {
					return store, storeConfig.Comment
				}
			}
		}
	}

	// Check documents if enabled
	if enableDocuments {
		for doc, docConfig := range config.Documents {
			for _, e := range docConfig.Extensions {
				if e == ext {
					return doc, docConfig.Comment
				}
			}
		}
	}

	// If the extension of the file does not match any of the extensions in the configuration, return an empty string and nil to indicate that the language is not supported.
	return "", nil
}

// Counts the number of lines of code in a project based on the configuration.
// It takes a configuration object and two boolean values indicating whether to enable stores and documents as input.
// It returns an error if one occurs.
func countLinesOfCode(config *Config, enableStores, enableDocuments bool) error {
	// Get the current working directory
	cwd, err := os.Getwd()
	// If an error occurs, return it
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Build a list of files to process based on the configuration and the current working directory
	filesToProcess, err := buildFileList(config, cwd)
	// If an error occurs, return it
	if err != nil {
		return fmt.Errorf("failed to build file list: %w", err)
	}

	// Initialize a variable to store the total number of lines of code
	totalLines := 0
	// Initialize a strings.Builder object to store the output
	var output strings.Builder

	// Iterate over the slice of files to process
	for _, path := range filesToProcess {
		// Get the relative path of the file
		relPath, _ := filepath.Rel(cwd, path)
		// Detect the language of the file based on its extension and the configuration
		lang, _ := detectLanguage(path, config, enableStores, enableDocuments)
		// If the language is not supported, skip the file
		if lang == "" {
			continue
		}

		// Read the content of the file
		content, err := os.ReadFile(path)
		// If an error occurs, return it
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}

		// Count the number of non-empty lines in the content
		lineCount := countNonEmptyLines(content)
		// Add the line count to the total number of lines of code
		totalLines += lineCount

		// If verbose output is enabled, print the file name, language, and line count
		if verbose {
			fmt.Printf("File: %s, Language: %s, Lines: %d\n", relPath, lang, lineCount)
		}

		// Write the file name, language, and line count to the output
		output.WriteString(fmt.Sprintf("%s,%s,%d\n", relPath, lang, lineCount))
	}

	// Print the total number of lines of code
	fmt.Printf("Total lines of code: %d\n", totalLines)

	// If an output file is specified, write the output to the file
	if outputFile != "" {
		err = os.WriteFile(outputFile, []byte(output.String()), 0644)
		// If an error occurs, return it
		if err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		// Print the name of the output file
		fmt.Printf("Output written to %s\n", outputFile)
	}

	// If no error occurs, return nil
	return nil
}

// Counts the number of non-empty lines in a byte slice.
// It takes a byte slice containing the content of a file as input.
// It returns an integer representing the number of non-empty lines in the content.
func countNonEmptyLines(content []byte) int {
	// Create a new scanner that reads from a new reader that reads from the content byte slice.
	scanner := bufio.NewScanner(bytes.NewReader(content))
	// Initialize a variable to store the number of non-empty lines.
	lineCount := 0
	// Use a for loop to iterate over the lines in the content.
	for scanner.Scan() {
		// Trim any whitespace from the current line and check if its length is greater than 0.
		if len(strings.TrimSpace(scanner.Text())) > 0 {
			// If the length is greater than 0, increment the line count.
			lineCount++
		}
	}
	// Return the line count.
	return lineCount
}

// Registers command-line flags for the rootCmd object.
func init() {
	// If the flag is not provided, the output will be printed to the console.
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file name (optional)")
	// Allows the user to specify a local configuration file.
	// If the flag is not provided, the tool will use the default local configuration file (.locc.yaml).
	rootCmd.Flags().StringVarP(&configFile, "config", "c", "", "Local configuration file (optional)")
	// Enables the processing of data stores (JSON, YAML, etc.).
	// If the flag is not provided, the tool will not process data stores.
	rootCmd.Flags().BoolVar(&enableStores, "data", false, "Enable processing of data stores (JSON, YAML, etc.)")
	// Enables the processing of documents (plain text, Markdown, etc.).
	// If the flag is not provided, the tool will not process documents.
	rootCmd.Flags().BoolVar(&enableDocuments, "docs", false, "Enable processing of documents (plain text, Markdown, etc.)")
	// Initializes a local configuration file.
	// If the flag is provided, the tool will create a default local configuration file (.locc.yaml) in the current directory and exit.
	rootCmd.Flags().BoolVar(&initLocalConfig, "init", false, "Initialize a local configuration file")
	// Enables verbose output.
	// If the flag is provided, the tool will print the number of lines of code for each file it processes in addition to the total lines of code.
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
}

func runInit(filename string) error {
	// Check if local config already exists
	if _, err := os.Stat(filename); err == nil {
		return fmt.Errorf("local config file already exists at %s", filename)
	}

	// Load default local config
	defaultLocalConfigContent, err := defaultLocalConfig.ReadFile("default_local_config.yaml")
	if err != nil {
		return fmt.Errorf("failed to read default local config: %w", err)
	}

	// Write default local config to file
	err = os.WriteFile(filename, defaultLocalConfigContent, 0644)
	if err != nil {
		return fmt.Errorf("failed to write default local config to file: %w", err)
	}

	fmt.Printf("Successfully initialized local configuration file at %s\n", filename)
	return nil
}
