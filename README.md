# templar

templar is a command-line tool for rendering templates with dynamic values. It processes input files, applies templates, and generates output files based on provided values.

## Features

- Supports YAML-based value files and inline key-value pairs.
- Includes and excludes files using glob patterns.
- Supports dry-run mode for testing without writing files.
- Verbose logging for detailed output.
- Strips suffixes from output filenames.
- Allows copying files without templating based on patterns.
- Supports environment variable substitution in YAML files.

## Installation

1. Clone the repository:
    ```bash
    git clone https://github.com/your-username/templar.git
    cd templar
    ```

2. Build the binary:
    ```bash
    go build -o templar
    ```

## Usage

```bash
templar [options] <input-dir>
```

### Options

- `--dry-run`: Simulate actions without writing files.
- `--verbose`: Enable verbose logging.
- `--values <file>`: Path to a values YAML file (can be repeated).
- `--set <key=value>`: Set a value inline (can be repeated).
- `--include <pattern>`: Glob pattern of files to include (can be repeated).
- `--exclude <pattern>`: Glob pattern of files to exclude (can be repeated).
- `--copy-only <pattern>`: Glob pattern for files to copy without templating (can be repeated).
- `--template-only <pattern>`: Glob pattern for files to template; others copied as-is.
- `--strip <suffix>`: Suffix to strip from output filenames.
- `--readonly`: Make all generated files read-only.
- `-o <dir>`: Output directory (default: `out`).

### Example

```bash
templar --values values.yaml --set app.name=myapp --include "**/*.tmpl" --exclude "test/*" --copy-only "**/*.txt" -o output templates/
```

## Project Structure

- `main.go`: Entry point for the CLI.
- `internal/loader`: Handles loading and merging of values.
- `internal/renderer`: Renders templates to output files.
- `internal/walker`: Collects files and applies include/exclude filters.
- `internal/config`: Loads configuration from a YAML file.
- `values.yaml`: Example values file.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.