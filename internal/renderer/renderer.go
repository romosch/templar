// internal/renderer/renderer.go
package renderer

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"text/template/parse"

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

// Collect all .Values.x references
func collectUsedVariables(tmpl *template.Template) []string {
	var vars []string
	for _, tree := range tmpl.Templates() {
		ast := tree.Tree.Root
		visitNodes(ast, func(n parse.Node) {
			if node, ok := n.(*parse.FieldNode); ok {
				// Look for things like .Values.db_user
				if len(node.Ident) >= 2 && node.Ident[0] == "Values" {
					vars = append(vars, "."+joinDotPath(node.Ident))
				}
			}
		})
	}
	return vars
}

func visitNodes(n parse.Node, visit func(parse.Node)) {
	visit(n)
	switch node := n.(type) {
	case *parse.ListNode:
		for _, item := range node.Nodes {
			visitNodes(item, visit)
		}
	case *parse.IfNode:
		visitNodes(node.List, visit)
		if node.ElseList != nil {
			visitNodes(node.ElseList, visit)
		}
	case *parse.RangeNode:
		visitNodes(node.List, visit)
		if node.ElseList != nil {
			visitNodes(node.ElseList, visit)
		}
	case *parse.WithNode:
		visitNodes(node.List, visit)
		if node.ElseList != nil {
			visitNodes(node.ElseList, visit)
		}
	case *parse.TemplateNode:
		// You can handle nested templates here if needed
	}
}

func joinDotPath(parts []string) string {
	return fmt.Sprintf("%s", parts[0]) + "." + strings.Join(parts[1:], ".")
}

func valueExists(m map[string]interface{}, path []string) bool {
	curr := m
	for i, key := range path {
		val, ok := curr[key]
		if !ok {
			return false
		}
		if i == len(path)-1 {
			return true
		}
		next, ok := val.(map[string]interface{})
		if !ok {
			return false
		}
		curr = next
	}
	return false
}

func RenderFile(inputPath, baseInputDir, outputDir string, values map[string]interface{}, dryRun, verbose bool, stripSuffix string) error {
	content, err := os.ReadFile(inputPath)
	if err != nil {
		return err
	}

	funcMap["toYaml"] = toYaml
	funcMap["indent"] = indent

	tmpl, err := template.New(filepath.Base(inputPath)).Funcs(funcMap).Parse(string(content))
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

	unset := 0
	for _, v := range collectUsedVariables(tmpl) {
		pathParts := strings.Split(v, ".")[1:] // remove leading dot
		if !valueExists(values, pathParts) {
			unset++
			if verbose {
				fmt.Fprintf(os.Stderr, "[templar] ⚠️  Warning: %s is unset\n", v)
			}
		}
	}
	if unset > 0 {
		fmt.Fprintf(os.Stderr, "[templar] ⚠️  Warning: found %d unset variables. Use --verbose to list them.\n", unset)
	}

	return tmpl.Execute(outFile, values)
}
