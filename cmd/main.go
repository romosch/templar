package main

import (
	"fmt"
	"os"
	"path/filepath"

	"templar/internal/options"
	"templar/internal/tome"
	"templar/internal/values"
)

const Version = "v0.1.0"

func main() {

	options.Init()

	if options.Version() {
		fmt.Printf("templar %s\n", Version)
		os.Exit(0)
	}

	args := options.Args()
	if options.ShowHelp() || len(args) < 2 {
		fmt.Println("Usage: templar [flags] <input-dir> <output-dir>")
		options.PrintDefaults()
		if len(args) < 2 {
			os.Exit(1)
		}
	}

	values, err := values.LoadAndMerge(options.Values(), options.SetVals())
	if err != nil {
		fmt.Printf("[templar] ❌  failed to load values: %v\n", err)
		os.Exit(1)
	}

	info, err := os.Stat(args[0])
	if err != nil {
		fmt.Printf("[templar] ❌  failed to access input path: %v\n", err)
		os.Exit(1)
	}
	source := args[0]
	if !info.IsDir() {
		source = filepath.Dir(args[0])
	}
	target := args[1]
	info, err = os.Stat(args[1])
	if err == nil && !info.IsDir() {
		target = filepath.Dir(args[1])
	}

	baseTome, err := tome.New(
		source,
		target,
		options.Mode(),
		options.StripSuffix(),
		options.IncludePatterns(),
		options.ExcludePatterns(),
		options.CopyPatterns(),
		options.TempPatterns(),
		values,
	)

	if err != nil {
		fmt.Printf("[templar] ❌  failed to create base tome: %v\n", err)
		os.Exit(1)
	}

	if options.Verbose() {
		fmt.Printf("Base Tome: %+v\n", baseTome)
	}

	err = baseTome.Walk(args[0])
	if err != nil {
		fmt.Printf("[templar] ❌  error walking files: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("[templar] ✅  Template rendering complete.")
	os.Exit(0)
}
