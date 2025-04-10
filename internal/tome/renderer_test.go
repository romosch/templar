package tome

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRender(t *testing.T) {
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "input.txt")
	outputDir := filepath.Join(tempDir, "output")
	outputFile := filepath.Join(outputDir, "input.txt")

	// Create a mock input file
	err := os.WriteFile(inputFile, []byte("Hello, {{ .Name }}!"), 0644)
	if err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	tome := &Tome{
		Source: tempDir,
		Target: outputDir,
		Values: map[string]interface{}{
			"Name": "World",
		},
	}

	err = tome.Render(inputFile, true, false, true)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Verify the output file exists
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Fatalf("Output file was not created")
	}

	// Verify the content of the output file
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	expectedContent := "Hello, World!"
	if string(content) != expectedContent {
		t.Errorf("Output content mismatch. Expected: %q, Got: %q", expectedContent, string(content))
	}
}

func TestRenderDryRun(t *testing.T) {
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "input.txt")
	outputDir := filepath.Join(tempDir, "output")

	// Create a mock input file
	err := os.WriteFile(inputFile, []byte("Hello, {{ .Name }}!"), 0644)
	if err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	tome := &Tome{
		Source: tempDir,
		Target: outputDir,
		Values: map[string]interface{}{
			"Name": "World",
		},
	}

	err = tome.Render(inputFile, true, true, false)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Verify the output file does not exist
	outputFile := filepath.Join(outputDir, "input.txt")
	if _, err := os.Stat(outputFile); !os.IsNotExist(err) {
		t.Fatalf("Output file should not have been created in dry-run mode")
	}
}
