package tome

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"templar/internal/options"
)

func (t *Tome) Walk(root string) error {
	if filepath.Base(root) == ".tome.yaml" {
		return nil
	}
	if !t.ShouldInclude(filepath.Base(root)) {
		if options.Verbose() {
			log.Print("Skipping:", filepath.Base(root))
		}
		return nil
	}
	info, err := os.Stat(root)
	if err != nil {
		return fmt.Errorf("failed to stat %s: %w", root, err)
	}

	if !info.IsDir() {
		// If root is a file, apply the logic for individual files
		err = t.Render(root, options.Verbose(), options.DryRun(), options.Force())
		if err != nil {
			return fmt.Errorf("failed to render file: %w", err)
		}
		return nil
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	tomesFile := filepath.Join(root, ".tome.yaml")
	if _, err := os.Stat(tomesFile); errors.Is(err, os.ErrNotExist) {
		// No tome
		for _, entry := range entries {
			err = t.Walk(filepath.Join(root, entry.Name()))
			if err != nil {
				return err
			}
		}
	} else {
		// Tome found
		subTomes, err := LoadTomeFile(tomesFile, t)
		if err != nil {
			return fmt.Errorf("failed to load tomes from %s: %w", tomesFile, err)
		}
		for _, subTome := range subTomes {
			for _, entry := range entries {
				err = subTome.Walk(filepath.Join(root, entry.Name()))
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
