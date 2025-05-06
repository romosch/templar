package values

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

var envVarRegexp = regexp.MustCompile(`\$\{([^}]+)\}`)

func LoadAndMerge(valueFiles []string, setVals []string) (map[string]any, error) {
	final := map[string]any{}

	// Load values from files
	for _, file := range valueFiles {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read values file %s: %w", file, err)
		}

		// Substitute environment variables before parsing
		yamlText := SubstituteEnvVars(string(data))

		var parsed map[string]any
		if err := yaml.Unmarshal([]byte(yamlText), &parsed); err != nil {
			return nil, fmt.Errorf("invalid YAML in file %s: %w", file, err)
		}

		MergeMaps(final, parsed)
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
func setNestedValue(m map[string]any, key string, value string) {
	keys := strings.Split(key, ".")
	last := len(keys) - 1

	curr := m
	for i, k := range keys {
		if i == last {
			curr[k] = parseYAMLValue(value)
			return
		}
		if next, ok := curr[k].(map[string]any); ok {
			curr = next
		} else {
			next := make(map[string]any)
			curr[k] = next
			curr = next
		}
	}
}

// Merge src into dst
func MergeMaps(dst, src map[string]any) {
	for k, v := range src {
		if vMap, ok := v.(map[string]any); ok {
			if dstMap, ok := dst[k].(map[string]any); ok {
				MergeMaps(dstMap, vMap)
			} else {
				dst[k] = vMap
			}
		} else {
			dst[k] = v
		}
	}
}

// SubstituteEnvVars replaces ${VAR} with the corresponding environment variable.
// Escaped form ${{VAR}} is preserved as literal ${VAR}.
func SubstituteEnvVars(yamlContent string) string {
	// Step 1: Escape ${{VAR}} to a temporary placeholder
	yamlContent = strings.ReplaceAll(yamlContent, "${{", "__ESCAPED_VAR__START__")
	yamlContent = strings.ReplaceAll(yamlContent, "}}", "__ESCAPED_VAR__END__")

	// Step 2: Substitute all ${VAR}
	yamlContent = envVarRegexp.ReplaceAllStringFunc(yamlContent, func(m string) string {
		key := envVarRegexp.FindStringSubmatch(m)[1]
		return os.Getenv(key)
	})

	// Step 3: Restore escaped ${VAR}
	yamlContent = strings.ReplaceAll(yamlContent, "__ESCAPED_VAR__START__", "${")
	yamlContent = strings.ReplaceAll(yamlContent, "__ESCAPED_VAR__END__", "}")

	return yamlContent
}

func parseYAMLValue(value string) any {
	// Check for booleans
	if value == "true" || value == "True" {
		return true
	}
	if value == "false" || value == "False" {
		return false
	}

	// Check for null
	if value == "null" || value == "~" {
		return nil
	}

	// Check for integers
	if intVal, err := strconv.Atoi(value); err == nil {
		return intVal
	}

	// Check for floats
	if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
		return floatVal
	}

	// Check for lists
	if strings.HasPrefix(value, "{") && strings.HasSuffix(value, "}") {
		trimmed := strings.Trim(value, "{}")
		parts := strings.Split(trimmed, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		return parts
	}

	// Default to string
	return value
}
