package options

import (
	"flag"
	"fmt"
)

var (
	dryRun          bool
	verbose         bool
	force           bool
	showVersion     bool
	showHelp        bool
	strict          bool
	mode            string
	stripSuffix     string
	args            []string
	values          multiFlag
	setValues       multiFlag
	includePatterns multiFlag
	excludePatterns multiFlag
	copyPatterns    multiFlag
	tempPatterns    multiFlag
)

type multiFlag []string

func (m *multiFlag) String() string         { return fmt.Sprint(*m) }
func (m *multiFlag) Set(value string) error { *m = append(*m, value); return nil }

func Init() {
	flag.BoolVar(&showVersion, "version", false, "Show version and exit")
	flag.BoolVar(&showHelp, "help", false, "Show help and exit")
	flag.BoolVar(&dryRun, "dry-run", false, "Simulate actions without writing files")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	flag.BoolVar(&strict, "strict", false, "Fail on missing values")
	flag.Var(&values, "values", "Path to values YAML file (can be repeated)")
	flag.Var(&setValues, "set", "Set a value (key=value) (can be repeated)")
	flag.Var(&includePatterns, "include", "Glob pattern of files to include (can be repeated)")
	flag.Var(&excludePatterns, "exclude", "Glob pattern of files to exclude (can be repeated)")
	flag.Var(&copyPatterns, "copy", "Glob pattern for files to copy without templating (can be repeated)")
	flag.Var(&tempPatterns, "temp", "Glob pattern for files to template; others are copied as-is (mutually exclusive with --copy)")
	flag.StringVar(&stripSuffix, "strip", "", "Suffix to strip from output filenames if templated")
	flag.StringVar(&mode, "mode", "", "Set file mode (permissions) for created files (octal or symbolic)")
	flag.BoolVar(&force, "force", false, "Overwrite files in output directory without confirmation")

	flag.Parse()
	args = flag.Args()

}

func DryRun() bool {
	return dryRun
}

func Verbose() bool {
	return verbose
}

func Force() bool {
	return force
}

func Version() bool {
	return showVersion
}

func Strict() bool {
	return strict
}

func ShowHelp() bool {
	return showHelp
}

func StripSuffix() string {
	return stripSuffix
}

func Values() []string {
	return values
}

func SetVals() []string {
	return setValues
}

func IncludePatterns() []string {
	return includePatterns
}

func ExcludePatterns() []string {
	return excludePatterns
}

func CopyPatterns() []string {
	return copyPatterns
}

func TempPatterns() []string {
	return tempPatterns
}

func Args() []string {
	return args
}

func PrintDefaults() {
	flag.PrintDefaults()
}

func Mode() string {
	return mode
}
