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
	cmd := exec.Command("go", "build", "-o", templarBin, "../cmd/main.go")
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

func TestScenario1a(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir := t.TempDir()
	cmd := exec.Command(templarBin, "--verbose", "--set=test=a", "--set=num=01914634", "--out="+tmpDir, "inputs/scenario1/templates")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatal("failed to run templar command:", err)
	}

	compareDirectories(t, "outputs/scenario1_a", tmpDir)
}

func TestScenario1b(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir := t.TempDir()
	cmd := exec.Command(templarBin, "--verbose", "--set=test=b", "--out="+tmpDir, "inputs/scenario1/templates")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatal("failed to run templar command:", err)
	}

	compareDirectories(t, "outputs/scenario1_b", tmpDir)
}

func compareDirectories(t *testing.T, wantPath, gotPath string) {
	got, err := os.ReadDir(gotPath)
	if err != nil {
		t.Fatalf("failed to read tmpDir: %v", err)
	}
	want, err := os.ReadDir(wantPath)
	if err != nil {
		t.Fatalf("failed to read outputs/scenario1_a: %v", err)
	}

	exptedFilesString := ""
	for _, entry := range want {
		exptedFilesString += fmt.Sprintf("  %s\n", entry.Name())
	}
	producedFilesString := ""
	for _, entry := range got {
		producedFilesString += fmt.Sprintf("  %s\n", entry.Name())
	}

	if len(got) != len(want) {
		t.Fatalf(`
directory file count mismatch:
expected %s (%d):
%s
produced %s (%d):
%s
`,
			wantPath, len(want), exptedFilesString, gotPath, len(got), producedFilesString)
	}
	for i, entry := range got {
		if entry.Name() != want[i].Name() {
			t.Errorf("file name mismatch at index %d: got %s, want %s", i, entry.Name(), want[i].Name())
		}
		gotEntryPath := filepath.Join(gotPath, entry.Name())
		wantEntryPath := filepath.Join(wantPath, want[i].Name())

		gotEntryInfo, err := os.Lstat(gotEntryPath)
		if err != nil {
			t.Fatalf("failed to lstat %s: %v", gotEntryPath, err)
		}
		wantEntryInfo, err := os.Lstat(wantEntryPath)
		if err != nil {
			if os.IsNotExist(err) {
				t.Fatalf("expected file %s does not exist in %s", wantEntryPath, gotPath)
			} else {
				t.Fatalf("failed to lstat %s: %v", wantEntryPath, err)
			}

		}
		if gotEntryInfo.IsDir() {
			if !wantEntryInfo.IsDir() {
				t.Errorf("expected directory but got file for %s", entry.Name())
			}
			compareDirectories(t, wantEntryPath, gotEntryPath)
			continue
		}

		if gotEntryInfo.Mode()&os.ModeType != wantEntryInfo.Mode()&os.ModeType {
			t.Errorf("file type mismatch for %s: got %v, want %v", entry.Name(), gotEntryInfo.Mode()&os.ModeType, wantEntryInfo.Mode()&os.ModeType)
			continue
		}

		if gotEntryInfo.Mode()&os.ModeSymlink != 0 {
			gotLink, err := os.Readlink(gotEntryPath)
			if err != nil {
				t.Fatalf("failed to readlink %s: %v", gotEntryPath, err)
			}
			wantLink, err := os.Readlink(wantEntryPath)
			if err != nil {
				t.Fatalf("failed to readlink %s: %v", wantEntryPath, err)
			}
			if gotLink != wantLink {
				t.Errorf("symlink target mismatch for %s: got %s, want %s", entry.Name(), gotLink, wantLink)
			}
			continue
		}

		gotData, err := os.ReadFile(gotEntryPath)
		if err != nil {
			t.Fatalf("failed to read %s: %v", gotEntryPath, err)
		}
		wantData, err := os.ReadFile(wantEntryPath)
		if err != nil {
			t.Fatalf("failed to read %s: %v", wantEntryPath, err)
		}
		if string(gotData) != string(wantData) {
			t.Errorf("file content mismatch for %s: got: %s, want: %s", entry.Name(), string(gotData), string(wantData))
		}

		if gotEntryInfo.Mode().Perm() != wantEntryInfo.Mode().Perm() {
			t.Errorf("file mode mismatch for %s: got %v, want %v", entry.Name(), gotEntryInfo.Mode().Perm(), wantEntryInfo.Mode().Perm())
		}
	}
}
