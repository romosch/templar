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
	"templar/internal/options"
	"text/template"
	"text/template/parse"
)

func (t *Tome) Render(inputPath string, verbose, dryRun, force bool) error {
	relPath, err := filepath.Rel(t.source, inputPath)
	if err != nil {
		return fmt.Errorf("error getting relative path: %w", err)
	}

	// Template the file name
	var templatedPath bytes.Buffer
	err = t.Template(&templatedPath, relPath, inputPath)
	if err != nil {
		return fmt.Errorf("error templating name: %w", err)
	}

	outputPath := filepath.Join(t.target, templatedPath.String())

	if len(t.strip) > 0 {
		// Split the output path into directories
		parts := strings.Split(outputPath, string(filepath.Separator))
		for i, part := range parts {
			// Strip the suffix from each part of the path
			for _, s := range t.strip {
				if strings.HasSuffix(part, s) {
					parts[i] = strings.TrimSuffix(part, s)
					break
				}
			}
		}
		// Rejoin the parts to form the new output path
		if outputPath[0] == '/' {
			outputPath = "/" + filepath.Join(parts...)
		} else {
			outputPath = filepath.Join(parts...)
		}
	}

	// Get current file mode
	info, err := os.Lstat(inputPath)
	if err != nil {
		return fmt.Errorf("error stating input file: %w", err)
	}

	symlink := (info.Mode() & os.ModeSymlink) != 0
	copy := t.shouldCopy(inputPath)
	if verbose {
		if copy || symlink {
			fmt.Printf("Copying %s -> %s\n", inputPath, outputPath)
		} else {
			fmt.Printf("Templating %s -> %s\n", inputPath, outputPath)
		}
	}

	if dryRun {
		return nil
	}

	err = os.MkdirAll(filepath.Dir(outputPath), 0755)
	if err != nil {
		return fmt.Errorf("error creating output directory: %w", err)
	}

	if _, err := os.Stat(outputPath); !errors.Is(err, os.ErrNotExist) &&
		!force && !confirmOverwrite(outputPath) {
		return nil
	}

	if symlink {
		target, err := os.Readlink(inputPath)
		if err != nil {
			return fmt.Errorf("readlink %q: %w", inputPath, err)
		}
		if err := os.Symlink(target, outputPath); err != nil {
			return fmt.Errorf("symlink %q -> %q at %q: %w", inputPath, target, outputPath, err)
		}
		return nil
	}

	content, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("error reading input file: %w", err)
	}
	mode := info.Mode().Perm()
	if copy {
		err = os.WriteFile(outputPath, content, mode)
		if err != nil {
			return err
		}
	} else {
		outFile, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("error creating output file: %w", err)
		}
		err = t.Template(outFile, string(content), inputPath)
		outFile.Close()
		if err != nil {
			return fmt.Errorf("error templating contents: %w", err)
		}
	}

	if t.mode != 0 {
		mode = t.mode
	}

	// Apply new permissions
	if err := os.Chmod(outputPath, mode); err != nil {
		return fmt.Errorf("error setting file permissions: %w", err)
	}

	return nil
}

func (t *Tome) Template(writer io.Writer, text string, name string) error {
	tmpl, err := template.New(name).Funcs(t.funcMap(filepath.Dir(name))).Parse(text)
	if err != nil {
		return err
	}

	missingTemplateKeys, err := findMissingTemplateKeys(tmpl, text, t.values)
	if err != nil {
		return fmt.Errorf("error finding missing template keys for %s: %w", name, err)
	}
	if len(missingTemplateKeys) > 0 {
		for _, missingKey := range missingTemplateKeys {
			fmt.Printf("[templar] ⚠️  %s:%d:%d missing key '%s'\n", name, missingKey.Line, missingKey.Column, missingKey.Name)
		}
		if options.Strict() {
			return errors.New("missing template keys not allowed in strict mode")
		}
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
	fmt.Printf("[templar] ⚠️  '%s' already exists. Overwrite? [y/N]: ", path)
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes"
}

// MissingKey holds the name and position of a missing template key
type MissingKey struct {
	Name   string
	Line   int
	Column int
}

// FindMissingTemplateKeysWithPos parses the template and returns all keys
// that are missing in the given values map, along with their line and column.
func findMissingTemplateKeys(tmpl *template.Template, tmplStr string, values map[string]interface{}) ([]MissingKey, error) {
	var missing []MissingKey

	// Split the template into lines for positional tracking
	lines := strings.Split(tmplStr, "\n")

	// Traverse the template AST
	var walk func(node parse.Node)
	walk = func(node parse.Node) {
		switch n := node.(type) {
		case *parse.ListNode:
			for _, sub := range n.Nodes {
				walk(sub)
			}
		case *parse.ActionNode:
			walk(n.Pipe)
		case *parse.PipeNode:
			for _, cmd := range n.Cmds {
				walk(cmd)
			}
		case *parse.CommandNode:
			for _, arg := range n.Args {
				walk(arg)
			}
		case *parse.FieldNode:
			// .Field format, check the first part only (top-level)
			if len(n.Ident) > 0 {
				key := n.Ident[0]
				if _, ok := values[key]; !ok {
					line, col := positionFromOffset(n.Pos, lines)
					missing = append(missing, MissingKey{Name: key, Line: line, Column: col})
				}
			}
		case *parse.TemplateNode:
			// Not processing nested templates for now
		}
	}

	walk(tmpl.Tree.Root)

	return missing, nil
}

// positionFromOffset returns the line and column number based on Pos
func positionFromOffset(pos parse.Pos, lines []string) (int, int) {
	offset := int(pos) - 1 // Pos is 1-indexed

	count := 0
	for i, line := range lines {
		lineLen := len(line) + 1 // +1 for the newline
		if count+lineLen > offset {
			col := offset - count + 1
			return i + 1, col
		}
		count += lineLen
	}
	return 0, 0 // fallback
}
