package tome

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"templar/internal/values"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Mode    string         `yaml:"mode"`
	Target  string         `yaml:"target"`
	Strip   []string       `yaml:"strip"`
	Include []string       `yaml:"include"`
	Exclude []string       `yaml:"exclude"`
	Copy    []string       `yaml:"copy"`
	Temp    []string       `yaml:"temp"`
	Values  map[string]any `yaml:"values"`
}

func LoadTomeFile(file string, base *Tome) ([]*Tome, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read tome file: %w", err)
	}
	var templatedData bytes.Buffer
	err = base.Template(&templatedData, string(data), file)
	if err != nil {
		return nil, fmt.Errorf("failed to template tome file: %w", err)
	}
	var tomeConfig Config
	var tomeConfigs []Config
	err = yaml.Unmarshal(templatedData.Bytes(), &tomeConfigs)
	if err != nil {
		err = yaml.Unmarshal(templatedData.Bytes(), &tomeConfig)
		if err != nil {
			return nil, fmt.Errorf("invalid YAML in tome file: %w", err)
		}
		tomeConfigs = []Config{tomeConfig}
	}
	if len(tomeConfigs) == 0 {
		return nil, fmt.Errorf("no tomes found in tome file")
	}

	dir := filepath.Dir(file)
	formattedDir, err := base.FormatPath(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to format target path: %w", err)
	}
	tomes := make([]*Tome, len(tomeConfigs))

	for i, tomeConfig := range tomeConfigs {
		if tomeConfig.Target == "" {
			tomeConfig.Target = formattedDir
		} else if tomeConfig.Target[0] != '/' {
			tomeConfig.Target = filepath.Join(filepath.Dir(formattedDir), tomeConfig.Target)
		}

		mergedValues := make(map[string]any, len(base.values))
		for key, value := range base.values {
			mergedValues[key] = value
		}

		values.MergeMaps(mergedValues, tomeConfig.Values)

		if len(tomeConfig.Strip) == 0 {
			tomeConfig.Strip = base.strip
		}

		if len(tomeConfig.Include) == 0 && len(tomeConfig.Exclude) == 0 {
			tomeConfig.Include = base.include
			tomeConfig.Exclude = base.exclude
		}

		if len(tomeConfig.Copy) == 0 && len(tomeConfig.Temp) == 0 {
			tomeConfig.Copy = base.copy
			tomeConfig.Temp = base.temp
		}

		tomes[i], err = New(dir, tomeConfig.Target, tomeConfig.Mode, tomeConfig.Strip,
			tomeConfig.Include, tomeConfig.Exclude, tomeConfig.Copy, tomeConfig.Temp, mergedValues)
		if err != nil {
			return nil, fmt.Errorf("failed to create tome %d: %w", i+1, err)
		}
	}

	return tomes, nil
}
