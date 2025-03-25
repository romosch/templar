// internal/loader/loader.go
package loader

import (
	"os"

	"gopkg.in/yaml.v3"
)

func LoadYAML(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	err = yaml.Unmarshal(data, &result)
	return result, err
}
