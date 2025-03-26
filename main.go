package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"templater/internal/config"
	"templater/internal/loader"
	"templater/internal/renderer"
	"templater/internal/walker"
)

var (
	dryRun               bool
	verbose              bool
	force                bool
	readonly             bool
	values               multiFlag
	setVals              multiFlag
	includePatterns      multiFlag
	excludePatterns      multiFlag
	copyOnlyPatterns     multiFlag
	templateOnlyPatterns multiFlag
	stripSuffix          string
	outputDir            string
	configPath           string
)

type multiFlag []string

func (m *multiFlag) String() string         { return fmt.Sprint(*m) }
func (m *multiFlag) Set(value string) error { *m = append(*m, value); return nil }

func init() {
	flag.BoolVar(&dryRun, "dry-run", false, "Simulate actions without writing files")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	flag.BoolVar(&readonly, "readonly", false, "Make all generated files read-only")
	flag.StringVar(&configPath, "config", ".templar.yaml", "Path to Templar config file")
	flag.Var(&values, "values", "Path to values YAML file (can be repeated)")
	flag.Var(&setVals, "set", "Set a value (key=value) (can be repeated)")
	flag.Var(&includePatterns, "include", "Glob pattern of files to include (can be repeated)")
	flag.Var(&excludePatterns, "exclude", "Glob pattern of files to exclude (can be repeated)")
	flag.Var(&copyOnlyPatterns, "copy-only", "Glob pattern for files to copy without templating (can be repeated)")
	flag.Var(&templateOnlyPatterns, "template-only", "Glob pattern for files to template; others copied as-is (mutually exclusive with --copy-only)")
	flag.StringVar(&stripSuffix, "strip", "", "Suffix to strip from output filenames if templated")
	flag.StringVar(&outputDir, "o", "out", "Output directory")
	flag.BoolVar(&force, "force", false, "Overwrite files in output directory without confirmation")

}

func confirmOverwrite(path string) bool {
	fmt.Printf("[templar] ⚠️  Non-empty output directory '%s' already exists. Overwrite? [y/N]: ", path)
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes"
}

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: templater [--dry-run] [--verbose] [--values] [--include] [--exclude] [--force] [--readonly] <input-dir>")
		fmt.Println("args: ", args)
		os.Exit(1)
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil && !os.IsNotExist(err) {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Use cfg to populate your variables if the CLI didn’t set them
	if cfg != nil {
		// Example: if CLI values are empty, use config
		if len(values) == 0 {
			values = cfg.Values
		}
		if len(setVals) == 0 {
			setVals = cfg.Set
		}
		if len(includePatterns) == 0 {
			includePatterns = cfg.Include
		}
		if len(excludePatterns) == 0 {
			excludePatterns = cfg.Exclude
		}
		if len(copyOnlyPatterns) == 0 {
			copyOnlyPatterns = cfg.CopyOnly
		}
		if len(templateOnlyPatterns) == 0 {
			templateOnlyPatterns = cfg.TemplateOnly
		}
		if stripSuffix == "" {
			stripSuffix = cfg.Strip
		}
		if outputDir == "" && cfg.OutputDir != "" {
			outputDir = cfg.OutputDir
		}
		if !force {
			force = cfg.Force
		}
		if !readonly {
			readonly = cfg.Readonly
		}
		if !dryRun {
			dryRun = cfg.DryRun
		}

		if len(flag.Args()) == 0 && len(cfg.InputPaths) > 0 {
			flag.CommandLine.Parse(cfg.InputPaths)
		}
	}

	if len(copyOnlyPatterns) > 0 && len(templateOnlyPatterns) > 0 {
		fmt.Println("[templar] ❌ Cannot use both --copy-only and --template-only")
		os.Exit(1)
	}

	if stat, err := os.Stat(outputDir); err == nil && stat.IsDir() && !force {
		entries, _ := os.ReadDir(outputDir)
		if len(entries) > 0 {
			if !confirmOverwrite(outputDir) {
				fmt.Println("[templar] Aborted.")
				os.Exit(0)
			}
		}
	}

	inputPath := args[0]

	values, err := loader.LoadAndMergeValues(values, setVals)
	if err != nil {
		log.Fatalf("failed to load values: %v", err)
	}
	if verbose {
		fmt.Printf("Loaded values: %v\n", values)
		fmt.Printf("Includes: %v\n", includePatterns)
	}

	absInputPath, _ := filepath.Abs(inputPath)
	if verbose {
		fmt.Printf("Processing input path: %s\n", absInputPath)
	}
	files, err := walker.CollectFilesAndDirs([]string{absInputPath})
	if err != nil {
		log.Fatalf("error walking files: %v", err)
	}

	for _, file := range files {
		relPath, _ := filepath.Rel(absInputPath, file)

		if !walker.ShouldInclude(relPath, includePatterns, excludePatterns) {
			if verbose {
				fmt.Printf("Skipping (excluded): %s\n", relPath)
			}
			continue
		}

		destPath := filepath.Join(outputDir, relPath)

		shouldTemplate := true // default: template all

		if len(copyOnlyPatterns) > 0 {
			shouldTemplate = !walker.MatchesAny(copyOnlyPatterns, relPath)
		} else if len(templateOnlyPatterns) > 0 {
			shouldTemplate = walker.MatchesAny(templateOnlyPatterns, relPath)
		}

		if shouldTemplate {
			if verbose {
				fmt.Printf("Templating: %s\n", relPath)
			}
			err := renderer.RenderFile(file, absInputPath, outputDir, values, dryRun, verbose, stripSuffix, readonly)
			if err != nil {
				log.Printf("error rendering file %s: %v", file, err)
			}
		} else {
			if verbose {
				fmt.Printf("Copying without templating: %s\n", relPath)
			}
			err := renderer.CopyFile(file, destPath, readonly)
			if err != nil {
				log.Printf("error copying file %s: %v", file, err)
			}
		}
	}

	fmt.Println("Template rendering complete.")
}
