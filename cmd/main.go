package main

import (
	"os"

	"log"
	"templar/internal/options"
	"templar/internal/tome"
	"templar/internal/values"
	"templar/internal/walker"
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

	baseTome := tome.Tome{
		Source:      args[0],
		Target:      args[1],
		Strip:       options.StripSuffix(),
		Include:     options.IncludePatterns(),
		Exclude:     options.ExcludePatterns(),
		Copy:        options.CopyPatterns(),
		Temp:        options.TempPatterns(),
		Permissions: tome.FilePermissions{},
		Values:      values,
	}

	if options.Verbose() {
		log.Printf("Base Tome: %+v", baseTome)
	}

	err = walker.Walk(args[0], &baseTome)
	if err != nil {
		log.Fatalf("error walking files: %v", err)
	}

	log.Print("Template rendering complete.")
	os.Exit(0)
}
