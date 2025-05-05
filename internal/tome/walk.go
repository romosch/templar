package tome

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"templar/internal/options"
)

// Walk traverses the file system starting from the specified root directory.
// It processes files and directories based on the rules defined in the Tome instance.
//
// If the root is a file, it applies rendering logic to the file. If the root is a directory,
// it recursively processes its contents. The function skips files or directories that do not
// meet the inclusion criteria defined by the Tome's include or exclude patterns.
//
// If a ".tome.yaml" file is found in a directory, it is treated as a configuration file
// for sub-Tomes. The function loads the sub-Tomes and delegates the traversal to them.
//
// Parameters:
//   - root: The starting path for the traversal.
//
// Returns:
//   - An error if any issues occur during traversal, file rendering, or sub-Tome loading.
func (t *Tome) Walk(root string) error {
	if filepath.Base(root) == ".tome.yaml" {
		return nil
	}
	if !t.ShouldInclude(root) {
		if options.Verbose() {
			fmt.Println("[templar] Skipping:", filepath.Base(root))
		}
		return nil
	}
	info, err := os.Lstat(root)
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
