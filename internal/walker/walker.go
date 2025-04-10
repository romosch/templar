package walker

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"templar/internal/options"
	"templar/internal/tome"
)

func Walk(root string, t *tome.Tome) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			if !t.ShouldInclude(root) {
				log.Print("Skipping file:", name)
				continue
			}
			tomesFile := filepath.Join(root, name, ".tome.yaml")
			if _, err := os.Stat(tomesFile); errors.Is(err, os.ErrNotExist) {
				// No tome
				err = Walk(filepath.Join(root, name), t)
				if err != nil {
					return fmt.Errorf("failed to walk directory: %w", err)
				}
			} else {
				// Tome found
				subTomes, err := tome.Load(tomesFile, t)
				if err != nil {
					return fmt.Errorf("failed to load tomes from %s: %w", tomesFile, err)
				}
				for _, subTome := range subTomes {
					err = Walk(filepath.Join(root, name), &subTome)
					if err != nil {
						return fmt.Errorf("failed to walk directory: %w", err)
					}
				}
			}
		} else {
			if !t.ShouldInclude(name) {
				if options.Verbose() {
					log.Print("Skipping file:", name)
				}
				continue
			}
			err = t.Render(filepath.Join(root, name), options.Verbose(), options.DryRun(), options.Force())
			if err != nil {
				return fmt.Errorf("failed to render file: %w", err)
			}
		}
	}
	return nil
}
