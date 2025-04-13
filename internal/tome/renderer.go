package tome

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
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

	relPath, err := filepath.Rel(t.Source, inputPath)
	if err != nil {
		return fmt.Errorf("Error getting relative path: %w", err)
	}

	if t.Strip != "" && filepath.Ext(relPath) == t.Strip {
		relPath = relPath[:len(relPath)-len(t.Strip)]
	}

	// Template the file name
	var templatedPath bytes.Buffer
	tmpl, err := template.New("filename").Funcs(funcMap).Parse(relPath)
	if err != nil {
		return fmt.Errorf("Error parsing filename template: %w", err)
	}
	err = tmpl.Execute(&templatedPath, t.Values)
	if err != nil {
		return fmt.Errorf("Error templating filename: %w", err)
	}

	outputPath := filepath.Join(t.Target, templatedPath.String())
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

	// Apply updates
	mode = updateRolePerms(mode, t.Permissions.User, 6)
	mode = updateRolePerms(mode, t.Permissions.Group, 3)
	mode = updateRolePerms(mode, t.Permissions.Other, 0)

	// Apply new permissions
	if err := os.Chmod(outputPath, mode); err != nil {
		return fmt.Errorf("Error setting file permissions: %w", err)
	}

	return nil
}

func (t *Tome) Template(writer io.Writer, data []byte) error {
	fmt.Printf("[templar] Templating %s\n", string(data))
	tmpl, err := template.New("").Funcs(funcMap).Parse(string(data))
	if err != nil {
		return err
	}
	return tmpl.Execute(writer, t.Values)
}

func updateRolePerms(mode os.FileMode, frp FileRolePermission, shift uint) os.FileMode {
	// Current bits
	readBit := os.FileMode(04 << shift)
	writeBit := os.FileMode(02 << shift)
	execBit := os.FileMode(01 << shift)

	// Clear and set only if value is provided
	if frp.Read != nil {
		mode &^= readBit
		if *frp.Read {
			mode |= readBit
		}
	}
	if frp.Write != nil {
		mode &^= writeBit
		if *frp.Write {
			mode |= writeBit
		}
	}
	if frp.Execute != nil {
		mode &^= execBit
		if *frp.Execute {
			mode |= execBit
		}
	}
	return mode
}

func confirmOverwrite(path string) bool {
	log.Printf("[templar] ⚠️ '%s' already exists. Overwrite? [y/N]: ", path)
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes"
}
