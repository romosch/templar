package tome

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"templar/internal/options"

	"gopkg.in/yaml.v3"
)

type FileRolePermission struct {
	Read    *bool `yaml:"read"`
	Write   *bool `yaml:"write"`
	Execute *bool `yaml:"execute"`
}

type FilePermissions struct {
	User  FileRolePermission `yaml:"user"`
	Group FileRolePermission `yaml:"group"`
	Other FileRolePermission `yaml:"other"`
}

type Tome struct {
	Permissions FilePermissions `yaml:"permissions"`
	Source      string
	Target      string         `yaml:"target"`
	Strip       string         `yaml:"strip"`
	Include     []string       `yaml:"include"`
	Exclude     []string       `yaml:"exclude"`
	Copy        []string       `yaml:"copy"`
	Temp        []string       `yaml:"temp"`
	Values      map[string]any `yaml:"values"`
}

func (t *Tome) ShouldInclude(name string) bool {
	if len(t.Include) > 0 {
		for _, pattern := range t.Include {
			if matches, _ := filepath.Match(pattern, name); matches {
				return true
			}
		}
		return false
	}

	if len(t.Exclude) > 0 {
		for _, pattern := range t.Exclude {
			if matches, _ := filepath.Match(pattern, name); matches {
				return false
			}
		}
	}

	return true
}

func (t *Tome) shouldCopy(name string) bool {
	if len(t.Copy) > 0 {
		for _, pattern := range t.Copy {
			if matches, _ := filepath.Match(pattern, name); matches {
				return true
			}
		}
		return false
	}

	if len(t.Temp) > 0 {
		for _, pattern := range t.Temp {
			if matches, _ := filepath.Match(pattern, name); matches {
				return false
			}
		}
		return true
	}

	return false
}

func Load(file string, base *Tome) ([]Tome, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read tome file: %w", err)
	}
	var templatedData bytes.Buffer
	if options.Verbose() {
		fmt.Printf("Templated tomes file %s\n", templatedData.String())
	}
	err = base.Template(&templatedData, data)

	var tome Tome
	var tomes []Tome
	err = yaml.Unmarshal(templatedData.Bytes(), &tomes)
	if err != nil {
		err = yaml.Unmarshal(templatedData.Bytes(), &tomes)
		if err != nil {
			return nil, fmt.Errorf("invalid YAML in tome file: %w", err)
		}
		tomes = []Tome{tome}
	}

	dir := filepath.Dir(file)
	rel, err := filepath.Rel(base.Source, dir)
	if err != nil {
		return nil, fmt.Errorf("failed to get relative path: %w", err)
	}

	for i, tome := range tomes {
		tome.Source = dir
		if tome.Target == "" {
			tome.Target = filepath.Join(base.Target, rel)
		} else if tome.Target[0] != '/' {
			tome.Target = filepath.Join(filepath.Dir(filepath.Join(base.Target, rel)), tome.Target)
		}

		if len(tome.Values) == 0 {
			tome.Values = base.Values
		}

		if tome.Strip == "" {
			tome.Strip = base.Strip
		}

		if len(tome.Include) == 0 && len(tome.Exclude) == 0 {
			tome.Include = base.Include
			tome.Exclude = base.Exclude
		}

		if len(tome.Include) > 0 && len(tome.Exclude) > 0 {
			return nil, fmt.Errorf("cannot use both include and exclude patterns in tome %d", i+1)
		}

		if len(tome.Copy) == 0 && len(tome.Temp) == 0 {
			tome.Copy = base.Copy
			tome.Temp = base.Temp
		}

		if len(tome.Copy) > 0 && len(tome.Temp) > 0 {
			return nil, fmt.Errorf("cannot use both copy-only and template-only patterns in tome %d", i+1)
		}

		tomes[i] = tome
	}

	return tomes, nil
}
