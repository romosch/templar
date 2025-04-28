# Templar

Templar is a fast, lightweight templating engine for directories and files, built with Go’s text/template and Sprig.
It lets you define how directories are generated using simple, declarative control files, enabling dynamic, repeatable, and fully configurable project structures.

Templar is ideal for generating config files, scaffolding projects, or automating deployments — all with clean, composable templates.

## ✨ Features

- Directory-aware templating: Template entire directory trees, not just individual files.
- Control files (Tomes) define generation rules locally, per directory, similar to a Makefile for templates.
- Tomes can be templated themselves, enabling dynamic output directroy structures.
- Multiple outputs: Generate many outputs from a single template source using dynamic values.
- Go templates + Sprig functions: Powerful templating features out of the box.

## 📦 Installation

Download a prebuilt binary from the Releases page.

1. Clone the repository:
    ```bash
    git clone https://github.com/your-username/templar.git
    cd templar
    ```

2. Build the binary:
    ```bash
    go build -o templar
    ```

## 🛠️ Usage

```bash
templar [options] <input> <output>
```

### 🧰 Options

- `--dry-run`: Simulate actions without writing files.
- `--verbose`: Enable verbose logging.
- `--values <file>`: Path to a values YAML file (can be repeated).
- `--set <key=value>`: Set a value inline (can be repeated).
- `--include <pattern>`: Glob pattern of files to include (can be repeated).
- `--exclude <pattern>`: Glob pattern of files to exclude (can be repeated).
- `--copy <pattern>`: Glob pattern for files to copy without templating (can be repeated).
- `--temp <pattern>`: Glob pattern for files to template; others copied.
- `--strip <suffix>`: Suffix to strip from output filenames.
- `--mode`: Set file mode (permissions) for created files (octal or symbolic)

### 🧾 Tomes
A Tome is a special YAML file (`.tome.yaml`) placed inside any template directory.
It acts as a blueprint for rendering, telling Templar how the contents of that directory should be processed and where the generated outputs should be written.

Tomes enable local control and dynamic generation. A single template directory can produce one, many, or differently customized outputs, all from the same source.

Tome files themselves are templates. Before being evaluated, a .tome.yaml is rendered just like any other file — allowing using input variables, conditional logic, and Sprig functions to control how the directory behaves based on the provided values.

#### Properties
A .tome.yaml file can be composed of a single tome, or a list of them, each with the following properties:
| Property  | Type          | Description                                                                 | Default          |
|-----------|---------------|-----------------------------------------------------------------------------|------------------|
| `mode`    | `string`      | Octal/symbolic file-mode specifying rendered files type and permissions     | Same as template |
| `target`  | `string`      | Target directory (relative or absolute)                                     | Same as template |
| `strip`   | `string`      | Suffix to strip from output filenames                                       | None             |
| `include` | `[]string`    | Glob patterns of files to include (can be repeated)                         | All              |
| `exclude` | `[]string`    | Glob patterns of files to exclude (can be repeated)                         | None             |
| `copy`    | `[]string`    | Glob patterns for files to copy without templating (can be repeated)        | None             |
| `temp`    | `[]string`    | Glob patterns for files to template; others copied                          | All              |
| `values`  | `map[string]` | Key-value map containing the (default) values for rendering. Overwritten by higher-level values | None |

### 🧪 Example

```bash
templar --values values.yaml --set app.name=myapp --include "**/*.tmpl" --copy-only "**/*.txt" templates out
```

## 🤝 Contributions

Contributions are welcome! Please open an issue or submit a pull request.

## 📜 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.