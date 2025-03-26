// internal/renderer/renderer.go
package renderer

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

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

var funcMap = sprig.TxtFuncMap() // All Sprig helpers

func RenderFile(inputPath, baseInputDir, outputDir string,
	values map[string]interface{}, dryRun, verbose bool,
	stripSuffix string, readonly bool) error {
	content, err := os.ReadFile(inputPath)
	if err != nil {
		return err
	}

	funcMap["toYaml"] = toYaml
	funcMap["indent"] = indent

	tmpl, err := template.New(filepath.Base(inputPath)).Funcs(funcMap).Option("missingkey=error").Parse(string(content))
	if err != nil {
		return err
	}

	relPath, err := filepath.Rel(baseInputDir, inputPath)
	if err != nil {
		return err
	}

	if stripSuffix != "" && filepath.Ext(relPath) == stripSuffix {
		relPath = relPath[:len(relPath)-len(stripSuffix)]
	}

	outputPath := filepath.Join(outputDir, relPath)

	if verbose {
		fmt.Printf("Templating %s -> %s\n", inputPath, outputPath)
	}

	if dryRun {
		return nil
	}

	err = os.MkdirAll(filepath.Dir(outputPath), 0755)
	if err != nil {
		return err
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	err = tmpl.Execute(outFile, values)
	if err != nil {
		return err
	}

	// After tmpl.Execute(outFile, values)...
	if readonly {
		err = os.Chmod(outputPath, 0444) // Owner/group/others: read-only
		if err != nil {
			return fmt.Errorf("failed to make file readonly: %w", err)
		}
	}
	return nil
}

func CopyFile(src, dst string, readonly bool) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(dst), 0755)
	if err != nil {
		return err
	}

	err = os.WriteFile(dst, input, 0644)
	if err != nil {
		return err
	}

	if readonly {
		err = os.Chmod(dst, 0444)
		if err != nil {
			return fmt.Errorf("failed to make file readonly: %w", err)
		}
	}

	return nil
}
