package main

import (
	"os"

	"log"
	"templar/internal/options"
	"templar/internal/tome"
	"templar/internal/values"
)

const Version = "v0.1.0"

func main() {

	options.Init()

	args := options.Args()
	if options.ShowHelp() || len(args) < 2 {
		log.Print("Usage: templar [flags] <input-dir> <output-dir>")
		options.PrintDefaults()
		if len(args) < 2 {
			os.Exit(1)
		}
	}

	values, err := values.LoadAndMerge(options.Values(), options.SetVals())
	if err != nil {
		log.Fatalf("failed to load values: %v", err)
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
		log.Fatalf("failed to create base tome: %v", err)
	}

	if options.Verbose() {
		log.Printf("Base Tome: %+v", baseTome)
	}

	err = baseTome.Walk(args[0])
	if err != nil {
		log.Fatalf("error walking files: %v", err)
	}

	log.Print("Template rendering complete.")
	os.Exit(0)
}
