package tome

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"templar/internal/options"
)

// Render traverses the file system starting from the specified root path.
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
func (t *Tome) Render(inputPath string) error {
	if filepath.Base(inputPath) == ".tome.yaml" {
		return nil
	}
	if !t.ShouldInclude(inputPath) {
		if options.Verbose {
			fmt.Println("[templar] Skipping:", filepath.Base(inputPath))
		}
		return nil
	}
	info, err := os.Lstat(inputPath)
	if err != nil {
		return fmt.Errorf("failed to stat %s: %w", inputPath, err)
	}

	outputPath, err := t.formatPath(inputPath)
	if err != nil {
		return fmt.Errorf("error formatting path: %w", err)
	}

	// Determine the file mode
	mode := info.Mode().Perm()
	if t.mode != 0 {
		mode = t.mode
	}

	if info.IsDir() {
		if options.Verbose {
			fmt.Printf("[templar] Creating directory %s -> %v %s\n", inputPath, mode, outputPath)
		}
		if !options.DryRun {
			err = os.MkdirAll(outputPath, mode)
			if err != nil {
				return fmt.Errorf("error creating output directory: %w", err)
			}
		}
		// Input is a directory, iterate over its contents
		entries, err := os.ReadDir(inputPath)
		if err != nil {
			return fmt.Errorf("failed to read directory: %w", err)
		}

		tomesFile := filepath.Join(inputPath, ".tome.yaml")
		if _, err := os.Stat(tomesFile); errors.Is(err, os.ErrNotExist) {
			// No tomefile, render dir entries using the current tome
			for _, entry := range entries {

				err = t.Render(filepath.Join(inputPath, entry.Name()))
				if err != nil {
					return err
				}
			}
		} else {
			// tomefile exists, load it and render dir entries with each sub-tome
			subTomes, err := LoadTomeFile(tomesFile, t)
			if err != nil {
				return fmt.Errorf("failed to load tomes from %s: %w", tomesFile, err)
			}
			for _, subTome := range subTomes {
				if options.Verbose {
					fmt.Printf("[templar] Tome %s\n", subTome.source)
				}
				for _, entry := range entries {
					err = subTome.Render(filepath.Join(inputPath, entry.Name()))
					if err != nil {
						return err
					}
				}
			}
		}
		return nil
	}

	// Root is a file, render it

	// Determine whether to copy or template the file/symlink
	symlink := (info.Mode() & os.ModeSymlink) != 0
	copy := t.shouldCopy(inputPath)

	if options.Verbose {
		if symlink {
			fmt.Printf("[templar] Recreating symlink %s -> %s\n", inputPath, outputPath)
		} else if copy {
			fmt.Printf("[templar] Copying %s -> %v %s\n", inputPath, mode, outputPath)
		} else {
			fmt.Printf("[templar] Templating %s -> %v %s\n", inputPath, mode, outputPath)
		}
	}

	if options.DryRun {
		return nil
	}

	err = os.MkdirAll(filepath.Dir(outputPath), mode)
	if err != nil {
		return fmt.Errorf("error creating output directory: %w", err)
	}

	// Check if the output file already exists and handle it based on the options
	if _, err := os.Stat(outputPath); !errors.Is(err, os.ErrNotExist) &&
		!options.Force && !confirmOverwrite(outputPath) {
		return nil
	}

	// If the input is a symlink, read the target and create a new symlink
	if symlink {
		target, err := os.Readlink(inputPath)
		if err != nil {
			return fmt.Errorf("readlink %q: %w", inputPath, err)
		}
		if err := os.Symlink(target, outputPath); err != nil {
			return fmt.Errorf("symlink %q -> %q at %q: %w", inputPath, target, outputPath, err)
		}
		return nil
	}

	// If the input is a regular file, read its contents
	// and either copy or template it to the output path
	content, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("error reading input file: %w", err)
	}

	if copy {
		err = os.WriteFile(outputPath, content, mode)
		if err != nil {
			return err
		}
	} else {
		outFile, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("error creating output file: %w", err)
		}
		err = t.Template(outFile, string(content), inputPath)
		outFile.Close()
		if err != nil {
			return fmt.Errorf("error templating contents: %w", err)
		}
	}

	// Set the file permissions
	if err := os.Chmod(outputPath, mode); err != nil {
		return fmt.Errorf("error setting file permissions: %w", err)
	}

	return nil
}

func confirmOverwrite(path string) bool {
	fmt.Printf("[templar] ⚠️  '%s' already exists. Overwrite? [y/N]: ", path)
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes"
}
