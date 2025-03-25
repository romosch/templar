package loader

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

func LoadAndMergeValues(valueFiles []string, setVals []string) (map[string]interface{}, error) {
	final := map[string]interface{}{}

	// Load values from files
	for _, file := range valueFiles {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read values file %s: %w", file, err)
		}

		// 🌱 Substitute environment variables before parsing
		yamlText := substituteEnvVars(string(data))

		var parsed map[string]interface{}
		if err := yaml.Unmarshal([]byte(yamlText), &parsed); err != nil {
			return nil, fmt.Errorf("invalid YAML in file %s: %w", file, err)
		}

		mergeMaps(final, parsed)
	}

	// Merge --set values
	for _, setVal := range setVals {
		parts := strings.SplitN(setVal, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("--set must be in key=value format")
		}
		setNestedValue(final, parts[0], parts[1])
	}

	return final, nil
}

// Support nested keys like "app.name=foo"
func setNestedValue(m map[string]interface{}, key string, value string) {
	keys := strings.Split(key, ".")
	last := len(keys) - 1

	curr := m
	for i, k := range keys {
		if i == last {
			curr[k] = value
			return
		}
		if next, ok := curr[k].(map[string]interface{}); ok {
			curr = next
		} else {
			next := make(map[string]interface{})
			curr[k] = next
			curr = next
		}
	}
}

// Merge src into dst
func mergeMaps(dst, src map[string]interface{}) {
	for k, v := range src {
		if vMap, ok := v.(map[string]interface{}); ok {
			if dstMap, ok := dst[k].(map[string]interface{}); ok {
				mergeMaps(dstMap, vMap)
			} else {
				dst[k] = vMap
			}
		} else {
			dst[k] = v
		}
	}
}
