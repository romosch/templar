package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"templar/internal/config"
	"templar/internal/loader"
	"templar/internal/renderer"
	"templar/internal/walker"
)

const Version = "v0.1.0"

var (
	dryRun               bool
	verbose              bool
	force                bool
	readonly             bool
	showVersion          bool
	showHelp             bool
	values               multiFlag
	setVals              multiFlag
	includePatterns      multiFlag
	excludePatterns      multiFlag
	copyOnlyPatterns     multiFlag
	templateOnlyPatterns multiFlag
	stripSuffix          string
	configPath           string
)

type multiFlag []string

func (m *multiFlag) String() string         { return fmt.Sprint(*m) }
func (m *multiFlag) Set(value string) error { *m = append(*m, value); return nil }

func init() {
	flag.BoolVar(&showVersion, "version", false, "Show version and exit")
	flag.BoolVar(&showHelp, "help", false, "Show help and exit")
	flag.BoolVar(&dryRun, "dry-run", false, "Simulate actions without writing files")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	flag.BoolVar(&readonly, "readonly", false, "Make all generated files read-only")
	flag.StringVar(&configPath, "config", ".templar.yml", "Path to Templar config file")
	flag.Var(&values, "values", "Path to values YAML file (can be repeated)")
	flag.Var(&setVals, "set", "Set a value (key=value) (can be repeated)")
	flag.Var(&includePatterns, "include", "Glob pattern of files to include (can be repeated)")
	flag.Var(&excludePatterns, "exclude", "Glob pattern of files to exclude (can be repeated)")
	flag.Var(&copyOnlyPatterns, "copy-only", "Glob pattern for files to copy without templating (can be repeated)")
	flag.Var(&templateOnlyPatterns, "template-only", "Glob pattern for files to template; others copied as-is (mutually exclusive with --copy-only)")
	flag.StringVar(&stripSuffix, "strip", "", "Suffix to strip from output filenames if templated")
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

	if showVersion {
		fmt.Printf("Templar version %s\n", Version)
		os.Exit(0)
	}

	args := flag.Args()
	if showHelp || len(args) < 2 {
		fmt.Println("Usage: templar [flags] <input-dir> <output-dir>")
		flag.PrintDefaults()
		os.Exit(1)
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil && !os.IsNotExist(err) && configPath != flag.Lookup("config").DefValue {
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

		if !force {
			force = cfg.Force
		}
		if !readonly {
			readonly = cfg.Readonly
		}
		if !dryRun {
			dryRun = cfg.DryRun
		}
	}

	if len(copyOnlyPatterns) > 0 && len(templateOnlyPatterns) > 0 {
		fmt.Println("[templar] ❌ Cannot use both --copy-only and --template-only")
		os.Exit(1)
	}

	inputDir := args[0]
	outputDir := args[1]

	if stat, err := os.Stat(outputDir); err == nil && stat.IsDir() && !force && !dryRun {
		entries, _ := os.ReadDir(outputDir)
		if len(entries) > 0 {
			if !confirmOverwrite(outputDir) {
				fmt.Println("[templar] Aborted.")
				os.Exit(0)
			}
		}
	}

	values, err := loader.LoadAndMergeValues(values, setVals)
	if err != nil {
		log.Fatalf("failed to load values: %v", err)
	}
	if verbose {
		printConfig()
	}

	absInputPath, _ := filepath.Abs(inputDir)
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
			err := renderer.RenderFile(file, absInputPath, outputDir, values, dryRun, verbose, stripSuffix, readonly)
			if err != nil {
				log.Printf("error rendering file %s: %v", file, err)
			}
		} else {
			if verbose {
				fmt.Printf("Copying: %s\n", relPath)
			}
			if dryRun {
				continue
			}
			err := renderer.CopyFile(file, destPath, readonly)
			if err != nil {
				log.Printf("error copying file %s: %v", file, err)
			}
		}
	}

	fmt.Println("Template rendering complete.")
}

func printConfig() {
	fmt.Println("dryRun: ", dryRun)
	fmt.Println("verbose: ", verbose)
	fmt.Println("force: ", force)
	fmt.Println("readonly: ", readonly)
	fmt.Println("showVersion: ", showVersion)
	fmt.Println("showHelp: ", showHelp)
	fmt.Println("values: ", values)
	fmt.Println("setVals: ", setVals)
	fmt.Println("includePatterns: ", includePatterns)
	fmt.Println("excludePatterns: ", excludePatterns)
	fmt.Println("copyOnlyPatterns: ", copyOnlyPatterns)
	fmt.Println("templateOnlyPatterns: ", templateOnlyPatterns)
	fmt.Println("stripSuffix: ", stripSuffix)
	fmt.Println("configPath: ", configPath)
}
