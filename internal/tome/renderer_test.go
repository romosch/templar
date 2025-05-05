package tome

import (
	"flag"
	"io"
	"os"
	"path/filepath"
	"templar/internal/options"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderMatrix(t *testing.T) {
	tests := []struct {
		dryRun              bool
		force               bool
		inputMode           os.FileMode
		expectedMode        os.FileMode
		name                string
		inputDirName        string
		inputFileName       string
		inputFileContent    string
		expectedDirName     string
		expectedFileContent string
		expectedFileName    string
		tome                Tome
	}{
		{
			name:                "Replace placeholders and change mode",
			inputMode:           0644,
			expectedMode:        0777,
			inputDirName:        "input-{{.name}}.tmpl",
			inputFileName:       "input-{{.name}}.txt.tmpl",
			inputFileContent:    "Hello, {{ .msg }}!",
			expectedDirName:     "input-test",
			expectedFileName:    "input-test.txt",
			expectedFileContent: "Hello, World!",
			tome: Tome{
				mode:  0777,
				strip: []string{".tmpl"},
				values: map[string]interface{}{
					"msg":  "World",
					"name": "test",
				},
			},
		},
		{
			name:                "Keep Original Mode",
			inputMode:           0444,
			expectedMode:        0444,
			inputFileName:       "input.txt",
			inputFileContent:    "Hello, World!",
			expectedFileName:    "input.txt",
			expectedFileContent: "Hello, World!",
		},
		{
			name:             "Dry run without creating file",
			inputMode:        0644,
			expectedMode:     0644,
			inputFileName:    "input.txt",
			expectedFileName: "input.txt",
			dryRun:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			inputDir := filepath.Join(tempDir, tt.inputDirName)
			inputFile := filepath.Join(inputDir, tt.inputFileName)
			outputDir := filepath.Join(tempDir, "output")
			expectedDir := outputDir
			if tt.expectedDirName != "" {
				expectedDir = filepath.Join(outputDir, tt.expectedDirName)
			}

			expectedFile := filepath.Join(expectedDir, tt.expectedFileName)

			// Create a mock input file with specific permissions
			err := os.MkdirAll(inputDir, 0755)
			if err != nil {
				t.Fatalf("Failed to create input directory: %v", err)
			}
			err = os.WriteFile(inputFile, []byte(tt.inputFileContent), tt.inputMode)
			if err != nil {
				t.Fatalf("Failed to create input file: %v", err)
			}

			tt.tome.source = tempDir
			tt.tome.target = outputDir

			if tt.dryRun {
				flag.CommandLine.Set("dry-run", "true")
			}
			if tt.force {
				flag.CommandLine.Set("force", "true")
			}
			options.Init()
			err = tt.tome.Render(inputFile)
			assert.NoError(t, err, "Render should not return an error")

			if tt.dryRun {
				// Verify the output file does not exist
				if _, err := os.Stat(expectedFile); !os.IsNotExist(err) {
					t.Fatalf("Output file should not have been created")
				}
				return
			}

			// Verify the output file exists
			info, err := os.Stat(expectedFile)
			assert.NoError(t, err, "Output file should exist")

			// Verify the file name
			assert.Equal(t, tt.expectedFileName, info.Name(), "Output file name should match the expected name")

			// Verify the file mode
			assert.Equal(t, tt.expectedMode, info.Mode().Perm(), "File mode should match the specified mode")

			// Verify the content of the output file
			content, err := os.ReadFile(expectedFile)
			assert.NoError(t, err, "Failed to read output file")
			assert.Equal(t, tt.expectedFileContent, string(content), "Output file content should match the expected content")
		})
	}
}

func TestRenderImport(t *testing.T) {
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "input.txt")
	inputFileContent := `{{ include "import.txt" }}`
	importFile := filepath.Join(tempDir, "import.txt")
	importFileContent := "{{ .msg }}"

	expectedFileContent := "Hello, World!"

	err := os.WriteFile(inputFile, []byte(inputFileContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	err = os.WriteFile(importFile, []byte(importFileContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create import file: %v", err)
	}

	tome := Tome{
		source: tempDir,
		target: tempDir,
		values: map[string]interface{}{
			"msg": "Hello, World!",
		},
	}

	err = tome.Render(inputFile)
	assert.NoError(t, err, "Render should not return an error")

	// Verify the output file exists
	_, err = os.Stat(inputFile)
	assert.NoError(t, err, "Output file should exist")

	// Verify the content of the output file
	content, err := os.ReadFile(inputFile)
	assert.NoError(t, err, "Failed to read output file")
	assert.Equal(t, expectedFileContent, string(content), "Output file content should match the expected content")
}

func TestRenderRequired(t *testing.T) {
	tome := Tome{
		values: map[string]interface{}{
			"msg": "Hello, World!",
		},
	}
	err := tome.Template(io.Discard, "{{ required .msg }}", "input.txt")
	assert.NoError(t, err, "Render should not return an error")

	err = tome.Template(io.Discard, "{{ required .missing }}", "input.txt")
	assert.Error(t, err, "Render should return an error for missing required value")
}

func TestRenderSymLink(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := t.TempDir()
	inputFile := filepath.Join(inputDir, "input.txt")
	inputFileContent := "Hello, World!"
	symlinkFile := filepath.Join(inputDir, "symlink.txt")
	outputFile := filepath.Join(outputDir, "symlink.txt")

	err := os.WriteFile(inputFile, []byte(inputFileContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	err = os.Symlink(inputFile, symlinkFile)
	if err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	tome := Tome{
		source: inputDir,
		target: outputDir,
	}

	err = tome.Render(symlinkFile)
	assert.NoError(t, err, "Render should not return an error")

	// Verify the output file exists
	_, err = os.Stat(outputFile)
	assert.NoError(t, err, "Output file should exist")

	// Verify the content of the output file
	content, err := os.ReadFile(outputFile)
	assert.NoError(t, err, "Failed to read output file")
	assert.Equal(t, inputFileContent, string(content), "Output file content should match the expected content")

	// Verify the symlink points to the correct file
	resolvedPath, err := filepath.EvalSymlinks(outputFile)
	if err != nil {
		t.Fatalf("Failed to resolve symlink: %v", err)
	}
	assert.Equal(t, inputFile, resolvedPath, "Symlink should point to the original file")

}
