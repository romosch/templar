package options

import (
	"fmt"

	flag "github.com/spf13/pflag"
)

var (
	dryRun          bool
	verbose         bool
	force           bool
	showVersion     bool
	showHelp        bool
	strict          bool
	mode            string
	out             string
	args            []string
	stripSuffix     []string
	values          []string
	setValues       []string
	includePatterns []string
	excludePatterns []string
	copyPatterns    []string
	tempPatterns    []string
)

type multiFlag []string

func (m *multiFlag) String() string         { return fmt.Sprint(*m) }
func (m *multiFlag) Set(value string) error { *m = append(*m, value); return nil }

func Init() {
	flag.BoolVarP(&showVersion, "version", "V", false, "Show version and exit")
	flag.BoolVarP(&showHelp, "help", "h", false, "Show help and exit")
	flag.BoolVarP(&dryRun, "dry-run", "d", false, "Simulate actions without writing files")
	flag.BoolVarP(&verbose, "verbose", "D", false, "Enable verbose logging")
	flag.BoolVarP(&strict, "strict", "S", false, "Fail on missing values")
	flag.BoolVarP(&force, "force", "F", false, "Overwrite files in output directory without confirmation")
	flag.StringVarP(&mode, "mode", "m", "", "Set file mode (permissions) for created files (octal or symbolic)")
	flag.StringVarP(&out, "out", "o", "", "Output directory for generated files (default: standard output)")
	flag.StringSliceVarP(&values, "values", "v", []string{}, "Path to values YAML file (can be repeated)")
	flag.StringSliceVarP(&setValues, "s", "s", []string{}, "Set a value (key=value) (can be repeated)")
	flag.StringSliceVarP(&includePatterns, "include", "i", []string{}, "Glob pattern of files to include (can be repeated)")
	flag.StringSliceVarP(&excludePatterns, "exclude", "e", []string{}, "Glob pattern of files to exclude (can be repeated)")
	flag.StringSliceVarP(&copyPatterns, "copy", "c", []string{}, "Glob pattern for files to copy without templating (can be repeated)")
	flag.StringSliceVarP(&tempPatterns, "temp", "t", []string{}, "Glob pattern for files to template; others are copied as-is (mutually exclusive with --copy)")
	flag.StringSliceVarP(&stripSuffix, "strip", "r", []string{}, "Suffix to strip from output filenames if templated (can be repeated)")

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

func StripSuffix() []string {
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

func Output() string {
	return out
}
