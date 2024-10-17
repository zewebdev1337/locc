# locc

## WARNING

This is basically a stripped down version of [promptify](https://github.com/zewebdev1337/promptify) and as such, it contains the same buggy behavior in merging configs and excluding directories (I'm looking at you `node_modules`), might fix later, hopefully.

## Description
locc (Lines of Code Counter) is a command-line tool designed to traverse subdirectories in the current working directory, identifying code files to calculate and present non-empty line counts.

The tool incorporates a configuration system that merges global settings located at `~/.locc.yaml` with local project-level configurations, which default to `.locc.yaml` within the current directory. This local configuration file path can be customized using the provided command-line flags during execution.

## Features

- Determination of programming languages based on file extensions.
- Implementation of file and directory inclusion/exclusion rules.
- Definition of a maximum file size limit for processing.
- Generation of verbose output detailing individual file analysis.

## Usage

```
locc [flags]
```

## Flags

```
  -c, --config string      Path to the local configuration file (Not required if your config file is `./.locc.yaml`).
      --data               Enable processing of data store files (e.g., JSON, YAML).
      --docs               Enable processing of document files (e.g., plain text, Markdown).
  -h, --help               Display help information.
      --init               Generate a local configuration file template.
  -o, --output string      Output the results to the specified file name.
  -v, --verbose            Enable verbose output for detailed file information.
```

## Configuration

The configuration file adheres to the YAML format and supports the following parameters:

- **languages**: This section defines a mapping of language-specific settings, encompassing file extensions associated with each language and their corresponding single-line and multi-line comment syntax.
- **stores**: Similar to the 'languages' section, this defines configurations for data storage files.
- **documents**: Analogous to the 'languages' and 'stores' sections, this configures settings related to document files.
- **excludes**: This section enables the specification of files or directories to be excluded from line counting. These exclusions can be defined globally or on a per-language basis.
- **max_file_size**: This parameter sets a limit, expressed in bytes, on the size of files considered for processing. Files exceeding this threshold are disregarded.


## Examples

```
# Analyze the current directory with default settings
locc

# Analyze the current directory using a specific configuration file
locc --config myconfig.yaml

# Analyze the current directory and include data store files
locc --data

# Analyze the current directory, enable verbose output, and write the results to a file
locc --verbose --output results.txt
```
