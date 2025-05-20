package tome

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

var templarBin string

func TestMain(m *testing.M) {
	// Build templar binary
	tmpDir, err := os.MkdirTemp("", "templar_integration_tests_*")
	defer os.RemoveAll(tmpDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	templarBin = filepath.Join(tmpDir, "templar")
	cmd := exec.Command("go", "build", "-o", templarBin, "../../cmd/main.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build helper: %v\n", err)
		os.Exit(1)
	}

	// 2) run the tests
	code := m.Run()

	// 4) exit with the right code
	os.Exit(code)
}

// func Scenario1(t *testing.T) {
// 	// Create a temporary directory for the test
// 	tmpDir := t.TempDir()
// 	cmd := exec.Command(templarBin, "--verbose", "--values=inputs/scenario1/values.yaml", "--set test=foo", "--out=outputs/scenario1_a", "inputs/scenario1/")
// 	cmd.Stdout = os.Stdout
// 	cmd.Stderr = os.Stderr
// 	if err := cmd.Run(); err != nil {
// 		fmt.Fprintf(os.Stderr, "failed to build helper: %v\n", err)
// 		os.Exit(1)
// 	}

// }
