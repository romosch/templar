// internal/walker/walker.go
package walker

import (
	"os"
	"path/filepath"
)

func CollectFilesAndDirs(paths []string) ([]string, error) {
	var allFiles []string

	for _, root := range paths {
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				allFiles = append(allFiles, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return allFiles, nil
}
