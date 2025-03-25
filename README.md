# Templater

Templater is a command-line tool for rendering templates with dynamic values. It processes input files, applies templates, and generates output files based on provided values.

## Features

- Supports YAML-based value files and inline key-value pairs.
- Includes and excludes files using glob patterns.
- Supports dry-run mode for testing without writing files.
- Verbose logging for detailed output.
- Strips suffixes from output filenames.

## Installation

1. Clone the repository:
    ```bash
    git clone https://github.com/your-username/templater.git
    cd templater
    ```

2. Build the binary:
    ```bash
    go build -o templater
    ```

## Usage

```bash
templater [options] <input-dir>
```

### Options

- `--dry-run`: Simulate actions without writing files.
- `--verbose`: Enable verbose logging.
- `--values <file>`: Path to a values YAML file (can be repeated).
- `--set <key=value>`: Set a value inline (can be repeated).
- `--include <pattern>`: Glob pattern of files to include (can be repeated).
- `--exclude <pattern>`: Glob pattern of files to exclude (can be repeated).
- `--strip <suffix>`: Suffix to strip from output filenames.
- `-o <dir>`: Output directory (default: `out`).

### Example

```bash
templater --values values.yaml --set app.name=myapp --include "**/*.tmpl" --exclude "test/*" -o output templates/
```

## Project Structure

- `main.go`: Entry point for the CLI.
- `internal/loader`: Handles loading and merging of values.
- `internal/renderer`: Renders templates to output files.
- `internal/walker`: Collects files and applies include/exclude filters.
- `values.yaml`: Example values file.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.  