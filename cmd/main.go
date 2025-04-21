package main

import (
	"fmt"
	"os"

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

	baseTome, err := tome.New(
		args[0],
		args[1],
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
