package tome

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"log"

	"github.com/Masterminds/sprig/v3"
	"gopkg.in/yaml.v3"
)

func toYaml(v interface{}) (string, error) {
	buf := new(bytes.Buffer)
	enc := yaml.NewEncoder(buf)
	enc.SetIndent(2)
	err := enc.Encode(v)
	return buf.String(), err
}

func indent(spaces int, text string) string {
	padding := strings.Repeat(" ", spaces)
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) != "" {
			lines[i] = padding + line
		}
	}
	return strings.Join(lines, "\n")
}

var funcMap template.FuncMap

func init() {
	funcMap = sprig.TxtFuncMap() // All Sprig helpers
	funcMap["toYaml"] = toYaml
	funcMap["indent"] = indent
}

func (t *Tome) Render(inputPath string, verbose, dryRun, force bool) error {
	// Get current file mode
	info, err := os.Stat(inputPath)
	if err != nil {
		return fmt.Errorf("Error stating input file: %w", err)
	}
	mode := info.Mode().Perm() // os.FileMode with only permission bits

	content, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("Error reading input file: %w", err)
	}

	relPath, err := filepath.Rel(t.source, inputPath)
	if err != nil {
		return fmt.Errorf("Error getting relative path: %w", err)
	}

	// Template the file name
	var templatedPath bytes.Buffer
	err = t.Template(&templatedPath, []byte(relPath))
	if err != nil {
		return fmt.Errorf("Error templating file name: %w", err)
	}

	outputPath := filepath.Join(t.target, templatedPath.String())

	if t.strip != "" {
		// Split the output path into directories
		parts := strings.Split(outputPath, string(filepath.Separator))
		for i, part := range parts {
			// Strip the suffix from each part of the path
			parts[i] = strings.TrimSuffix(part, t.strip)
		}
		// Rejoin the parts to form the new output path
		outputPath = "/" + filepath.Join(parts...)
	}

	copy := t.shouldCopy(inputPath)
	if verbose {
		if copy {
			log.Printf("Copying %s -> %s\n", inputPath, outputPath)
		} else {
			log.Printf("Templating %s -> %s\n", inputPath, outputPath)
		}
	}

	if dryRun {
		return nil
	}

	err = os.MkdirAll(filepath.Dir(outputPath), 0755)
	if err != nil {
		return fmt.Errorf("Error creating output directory: %w", err)
	}

	if _, err := os.Stat(outputPath); !errors.Is(err, os.ErrNotExist) &&
		!force && !confirmOverwrite(outputPath) {
		return nil
	}

	if copy {
		err = os.WriteFile(outputPath, content, mode)
		if err != nil {
			return err
		}
	} else {
		outFile, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("Error creating output file: %w", err)
		}
		err = t.Template(outFile, content)
		outFile.Close()
		if err != nil {
			return fmt.Errorf("Error templating file: %w", err)
		}
	}

	if t.mode != 0 {
		mode = t.mode
	}

	// Apply new permissions
	if err := os.Chmod(outputPath, mode); err != nil {
		return fmt.Errorf("Error setting file permissions: %w", err)
	}

	return nil
}

func (t *Tome) Template(writer io.Writer, data []byte) error {
	tmpl, err := template.New("").Funcs(funcMap).Parse(string(data))
	if err != nil {
		return err
	}
	return tmpl.Execute(writer, t.values)
}

func ParseFileMode(modeStr string) (os.FileMode, error) {
	// Try parsing as octal
	if n, err := strconv.ParseUint(modeStr, 8, 32); err == nil {
		return os.FileMode(n), nil
	}

	// If not octal, parse symbolic string
	if len(modeStr) != 9 {
		return 0, fmt.Errorf("invalid symbolic file mode: %s", modeStr)
	}

	var mode os.FileMode

	symbols := []struct {
		char byte
		bit  os.FileMode
	}{
		{'r', 0400}, {'w', 0200}, {'x', 0100}, // user
		{'r', 0040}, {'w', 0020}, {'x', 0010}, // group
		{'r', 0004}, {'w', 0002}, {'x', 0001}, // others
	}

	for i, sym := range symbols {
		if modeStr[i] == sym.char {
			mode |= sym.bit
		} else if modeStr[i] != '-' {
			return 0, fmt.Errorf("unexpected character '%c' in symbolic mode", modeStr[i])
		}
	}

	return mode, nil
}

func confirmOverwrite(path string) bool {
	log.Printf("[templar] ⚠️ '%s' already exists. Overwrite? [y/N]: ", path)
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes"
}
