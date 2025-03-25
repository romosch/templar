package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"templater/internal/loader"
	"templater/internal/renderer"
	"templater/internal/walker"
)

var (
	dryRun          bool
	verbose         bool
	force           bool
	values          multiFlag
	setVals         multiFlag
	includePatterns multiFlag
	excludePatterns multiFlag
	stripSuffix     string
	outputDir       string
)

type multiFlag []string

func (m *multiFlag) String() string         { return fmt.Sprint(*m) }
func (m *multiFlag) Set(value string) error { *m = append(*m, value); return nil }

func init() {
	flag.BoolVar(&dryRun, "dry-run", false, "Simulate actions without writing files")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	flag.Var(&values, "values", "Path to values YAML file (can be repeated)")
	flag.Var(&setVals, "set", "Set a value (key=value) (can be repeated)")
	flag.Var(&includePatterns, "include", "Glob pattern of files to include (can be repeated)")
	flag.Var(&excludePatterns, "exclude", "Glob pattern of files to exclude (can be repeated)")
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
		fmt.Println("Usage: templater [--dry-run] [--verbose] [--verbose] [--include] [--exclude] [--force] [--readonly] <input-dir>")
		fmt.Println("args: ", args)
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
		if verbose {
			fmt.Printf("Processing input path: %s\n", relPath)
		}
		if !walker.ShouldInclude(relPath, includePatterns, excludePatterns) {
			if verbose {
				fmt.Printf("Skipping: %s\n", relPath)
			}
			continue
		}

		if verbose {
			fmt.Printf("Processing: %s\n", relPath)
		}

		err := renderer.RenderFile(file, absInputPath, outputDir, values, dryRun, verbose, stripSuffix)
		if err != nil {
			log.Printf("error rendering file %s: %v", file, err)
		}
	}

	fmt.Println("Template rendering complete.")
}
