// cmd/config.go
package cmd

import (
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// defaultConfig is an embedded file system that contains the default configuration file.
// This variable is used to read the default configuration file when it is needed.
//
//go:embed default_config.yaml
var defaultConfig embed.FS

// defaultConfig is an embedded file system that contains the default configuration file.
// This variable is used to read the default configuration file when it is needed.
//
//go:embed default_local_config.yaml
var defaultLocalConfig embed.FS

// Config struct represents the configuration for the program.
// Languages field is a map that contains the configuration for each programming language.
// The key of the map is the language name, and the value is a LanguageConfig struct.
// Stores field is a map that contains the configuration for each store type.
// The key of the map is the store name, and the value is a LanguageConfig struct.
// Documents field is a map that contains the configuration for each document type.
// The key of the map is the document name, and the value is a LanguageConfig struct.
// Excludes field is a map that contains the exclusion configuration for specific files.
// The key of the map is the file name, and the value is an interface{} that can be either a slice of interfaces or a map of interfaces.
// Includes field is a map that contains the inclusion configuration for specific files.
// The key of the map is the file name, and the value is an interface{} that can be either a slice of interfaces or a map of interfaces.
// MaxFileSize field is an int64 that represents the maximum size of a file that can be processed.
type Config struct {
	Languages   map[string]LanguageConfig `yaml:"languages"`
	Stores      map[string]LanguageConfig `yaml:"stores"`
	Documents   map[string]LanguageConfig `yaml:"documents"`
	Excludes    map[string]interface{}    `yaml:"excludes"`
	Includes    map[string]interface{}    `yaml:"includes"`
	MaxFileSize int64                     `yaml:"max_file_size"`
}

// LanguageConfig struct represents the configuration for a specific programming language.
// Extensions field is a slice of strings that contains the file extensions associated with the language.
// Comment field is a slice of strings that contains the comment symbols used in the language.
type LanguageConfig struct {
	// Extensions is a slice of strings that contains the file extensions associated with the language.
	// For example, for Go language, this field might contain ["go"].
	Extensions []string `yaml:"extensions"`

	// Comment is a slice of strings that contains the comment symbols used in the language.
	// For example, for HTML language, this field would contain ["<!--", "-->"], for Go it would contain ["//"].
	Comment []string `yaml:"comment"`
}

// FileExclusion struct represents the exclusion configuration for a specific file.
// Wordlists field is of type ExclusionWordlists, which contains the lists of words to exclude and include.
type FileExclusion struct {
	// Wordlists is of type ExclusionWordlists, which contains the lists of words to exclude and include.
	// This field is used to specify the words that should be excluded or included from a file during processing.
	Wordlists ExclusionWordlists `yaml:"wordlists"`
}

// ExclusionWordlists struct represents the wordlists configuration for exclusions.
// Exclude field is a slice of strings that contains the list of words to exclude.
// Include field is a slice of strings that contains the list of words to include.
// These fields are used to specify the words that should be excluded or included from a file during processing.
type ExclusionWordlists struct {
	// Exclude is a slice of strings that contains the list of words to exclude.
	// If a word in this list is found in a file, it will be excluded during processing.
	Exclude []string `yaml:"exclude,omitempty"`

	// Include is a slice of strings that contains the list of words to include.
	// If a word in this list is found in a file, it will be included during processing,
	// even if it is also present in the Exclude list.
	Include []string `yaml:"include,omitempty"`
}

// getUserHomeDir is a function that retrieves the user's home directory.
// It uses the os.UserHomeDir function to get the home directory.
// If there is an error in retrieving the home directory, it logs the error and exits the program.
func getUserHomeDir() string {
	// Call os.UserHomeDir to get the user's home directory
	homeDir, err := os.UserHomeDir()
	// If there is an error, log the error and exit the program
	if err != nil {
		log.Fatal(fmt.Errorf("failed to get user home directory: %w", err))
	}
	// If there is no error, return the home directory
	return homeDir
}

// getGlobalConfigPath is a function that retrieves the path to the global configuration file.
// It uses the filepath.Join function to join the user's home directory with the ".locc.yaml" filename.
// The user's home directory is retrieved using the getUserHomeDir function.
// The function returns the path to the global configuration file.
func getGlobalConfigPath() string {
	// Call getUserHomeDir to get the user's home directory
	// Join the home directory with the ".locc.yaml" filename using filepath.Join
	// Return the path to the global configuration file
	return filepath.Join(getUserHomeDir(), ".locc.yaml")
}

// loadGlobalConfig is a function that loads the global configuration file.
// It first retrieves the path to the global configuration file using the getGlobalConfigPath function.
// It then checks if the global configuration file exists at that path.
// If it exists, it reads the file, unmarshals the YAML data into a Config struct, and returns the Config.
// If the global configuration file does not exist, it loads the default configuration file using the retrieveDefaultConfig function,
// marshals the default configuration into YAML data, writes the YAML data to the global configuration file path,
// and returns the default configuration.
// If there is an error in any of these steps, it returns the error.
func loadGlobalConfig() (*Config, error) {
	// Get the path to the global configuration file
	globalConfigPath := getGlobalConfigPath()

	// Check if the global configuration file exists at that path
	if _, err := os.Stat(globalConfigPath); err == nil {
		// If it exists, read the file
		globalConfigData, err := os.ReadFile(globalConfigPath)
		if err != nil {
			// If there is an error reading the file, return the error
			return nil, fmt.Errorf("failed to read global config: %w", err)
		}

		// Unmarshal the YAML data into a Config struct
		var config Config
		err = yaml.Unmarshal(globalConfigData, &config)
		if err != nil {
			// If there is an error unmarshalling the data, return the error
			return nil, fmt.Errorf("failed to parse global config: %w", err)
		}

		// Initialize maps if they are nil
		if config.Languages == nil {
			config.Languages = make(map[string]LanguageConfig)
		}
		if config.Stores == nil {
			config.Stores = make(map[string]LanguageConfig)
		}
		if config.Documents == nil {
			config.Documents = make(map[string]LanguageConfig)
		}
		if config.Excludes == nil {
			config.Excludes = make(map[string]interface{})
		}
		if config.Includes == nil {
			config.Includes = make(map[string]interface{})
		}

		return &config, nil
	} else if os.IsNotExist(err) {
		// If global config doesn't exist, load default config to be written it to global config path
		defaultConfig, err := retrieveDefaultConfig()
		if err != nil {
			// If there is an error retrieving the default config, return the error
			return nil, err
		}

		// Marshal the default configuration into YAML data
		defaultConfigData, err := yaml.Marshal(defaultConfig)
		if err != nil {
			// If there is an error marshalling the data, return the error
			return nil, fmt.Errorf("failed to marshal default config: %w", err)
		}

		// Write the YAML data to the global configuration file path
		err = os.WriteFile(globalConfigPath, defaultConfigData, 0644)
		if err != nil {
			// If there is an error writing the file, return the error
			return nil, fmt.Errorf("failed to write default config to global config path: %w", err)
		}
		// Return the default configuration
		return defaultConfig, nil
	} else {
		// If there is an error other than the file not existing, return the error
		return nil, err
	}
}

// retrieveDefaultConfig is a function that retrieves the default configuration.
// It uses the defaultConfig embed.FS variable to read the "default_config.yaml" file.
// If there is an error in reading the file, it returns the error.
// It then unmarshals the YAML data into a Config struct and returns the Config.
// If there is an error in unmarshalling the data, it returns the error.
func retrieveDefaultConfig() (*Config, error) {
	// Use the defaultConfig embed.FS variable to read the "default_config.yaml" file
	defaultConfigContent, err := defaultConfig.ReadFile("default_config.yaml")
	// If there is an error in reading the file, return the error
	if err != nil {
		return nil, fmt.Errorf("failed to read default config: %w", err)
	}

	// Declare a Config variable to hold the default configuration
	var config Config
	// Unmarshal the YAML data into the Config variable
	err = yaml.Unmarshal(defaultConfigContent, &config)
	// If there is an error in unmarshalling the data, return the error
	if err != nil {
		return nil, fmt.Errorf("failed to parse default config: %w", err)
	}

	// If there is no error, return the Config variable
	return &config, nil
}

// loadLocalConfig is a function that loads the local configuration file.
// It takes a filename as an argument, which is the path to the local configuration file.
// If the filename is not provided, it defaults to "./.locc.yaml".
// The function checks if the local configuration file exists at the provided path.
// If it exists, it reads the file, unmarshals the YAML data into a Config struct, and returns the Config.
// If the local configuration file does not exist, it returns nil and nil for the Config and error.
// If there is an error in any of these steps, it returns the error.
func loadLocalConfig(filename string) (*Config, error) {
	// If a filename is provided, use that as the local configuration file path.
	// Otherwise, use "./.locc.yaml" as the local configuration file path.
	var localConfigPath string
	if filename != "" {
		localConfigPath = filename
	} else {
		localConfigPath = "./.locc.yaml"
	}

	// Check if the local configuration file exists at the provided path
	if _, err := os.Stat(localConfigPath); err == nil {
		// If it exists, read the file
		localConfigData, err := os.ReadFile(localConfigPath)
		if err != nil {
			// If there is an error reading the file, return the error
			return nil, fmt.Errorf("failed to read local config: %w", err)
		}

		// Declare a Config variable to hold the local configuration
		var localConfig Config
		// Unmarshal the YAML data into the Config variable
		err = yaml.Unmarshal(localConfigData, &localConfig)
		if err != nil {
			// If there is an error unmarshalling the data, return the error
			return nil, fmt.Errorf("failed to parse local config: %w", err)
		}
		// If there is no error, return the Config variable
		// Initialize maps if they are nil
		if localConfig.Languages == nil {
			localConfig.Languages = make(map[string]LanguageConfig)
		}
		if localConfig.Stores == nil {
			localConfig.Stores = make(map[string]LanguageConfig)
		}
		if localConfig.Documents == nil {
			localConfig.Documents = make(map[string]LanguageConfig)
		}
		if localConfig.Excludes == nil {
			localConfig.Excludes = make(map[string]interface{})
		}
		if localConfig.Includes == nil {
			localConfig.Includes = make(map[string]interface{})
		}

		return &localConfig, nil
	}
	// If the local configuration file does not exist, return nil and nil for the Config and error
	return nil, nil
}

// mergeConfigs function merges the global and local configurations.
// It takes two pointers to Config structs as arguments: globalConfig and localConfig.
// If globalConfig is nil, it returns localConfig.
// If localConfig is nil, it returns globalConfig.
// If both configurations are not nil, it merges the local configuration into the global configuration.
// It checks if the local configuration has any language configurations, exclusions, or max file size.
// If it does, it updates the corresponding fields in the global configuration.
// Finally, it returns the merged global configuration.
func mergeConfigs(globalConfig, localConfig *Config) *Config {
	// If globalConfig is nil, return localConfig
	if globalConfig == nil {
		return localConfig
	}
	// If localConfig is nil, return globalConfig
	if localConfig == nil {
		return globalConfig
	}

	// If localConfig has any language configurations, merge them into globalConfig
	if localConfig.Languages != nil {
		for lang, langConfig := range localConfig.Languages {
			globalConfig.Languages[lang] = langConfig
		}
	}
	// If localConfig has any exclusions, merge them into globalConfig
	if localConfig.Excludes != nil {
		for lang, exclusions := range localConfig.Excludes {
			globalConfig.Excludes[lang] = exclusions
		}
	}

	// If localConfig has any inclusions, merge them into globalConfig
	if localConfig.Includes != nil {
		for lang, inclusions := range localConfig.Includes {
			globalConfig.Includes[lang] = inclusions
		}
	}
	// If localConfig has a max file size, update the max file size in globalConfig
	if localConfig.MaxFileSize > 0 {
		globalConfig.MaxFileSize = localConfig.MaxFileSize
	}

	// Return the merged global configuration
	return globalConfig
}

func processFilters(config *Config) {
	config.Excludes = processFilter(config.Excludes)
	config.Includes = processFilter(config.Includes)
}

func processFilter(filter map[string]interface{}) map[string]interface{} {
	for lang, rules := range filter {
		switch v := rules.(type) {
		case []interface{}:
			filter[lang] = processSimpleFilter(v)
		case map[interface{}]interface{}:
			filter[lang] = processDetailedFilter(v)
		}
	}
	return filter
}

func processSimpleFilter(rules []interface{}) map[string][]string {
	filterMap := make(map[string][]string)
	for _, item := range rules {
		if filename, ok := item.(string); ok {
			filterMap[filename] = []string{}
		}
	}
	return filterMap
}

func processDetailedFilter(rules map[interface{}]interface{}) map[string][]string {
	filterMap := make(map[string][]string)
	for filename, details := range rules {
		if detailsSlice, ok := details.([]interface{}); ok {
			var wordlist []string
			for _, word := range detailsSlice {
				if str, ok := word.(string); ok {
					wordlist = append(wordlist, str)
				}
			}
			filterMap[filename.(string)] = wordlist
		}
	}
	return filterMap
}
