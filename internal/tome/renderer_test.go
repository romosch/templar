package tome

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderMatrix(t *testing.T) {
	tests := []struct {
		dryRun           bool
		force            bool
		inputMode        os.FileMode
		expectedMode     os.FileMode
		name             string
		inputFileName    string
		inputFileContent string
		expectedOutput   string
		expectedFileName string
		tome             Tome
	}{
		{
			name:             "Replace placeholders and change mode",
			inputMode:        0644,
			expectedMode:     0777,
			inputFileName:    "input-{{.name}}.txt",
			inputFileContent: "Hello, {{ .msg }}!",
			expectedFileName: "input-test.txt",
			expectedOutput:   "Hello, World!",
			tome: Tome{
				mode: 0777,
				values: map[string]interface{}{
					"msg":  "World",
					"name": "test",
				},
			},
		},
		{
			name:             "Keep Original Mode",
			inputMode:        0444,
			expectedMode:     0444,
			inputFileName:    "input.txt",
			inputFileContent: "Hello, World!",
			expectedFileName: "input.txt",
			expectedOutput:   "Hello, World!",
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
			inputFile := filepath.Join(tempDir, tt.inputFileName)
			outputDir := filepath.Join(tempDir, "output")
			outputFile := filepath.Join(outputDir, tt.expectedFileName)

			// Create a mock input file with specific permissions
			err := os.WriteFile(inputFile, []byte(tt.inputFileContent), tt.inputMode)
			if err != nil {
				t.Fatalf("Failed to create input file: %v", err)
			}

			tt.tome.source = tempDir
			tt.tome.target = outputDir

			err = tt.tome.Render(inputFile, true, tt.dryRun, tt.force)
			assert.NoError(t, err, "Render should not return an error")

			if tt.dryRun {
				// Verify the output file does not exist
				if _, err := os.Stat(outputFile); !os.IsNotExist(err) {
					t.Fatalf("Output file should not have been created")
				}
				return
			}

			// Verify the output file exists
			info, err := os.Stat(outputFile)
			assert.NoError(t, err, "Output file should exist")

			// Verify the file name
			assert.Equal(t, tt.expectedFileName, info.Name(), "Output file name should match the expected name")

			// Verify the file mode
			assert.Equal(t, tt.expectedMode, info.Mode().Perm(), "File mode should match the specified mode")

			// Verify the content of the output file
			content, err := os.ReadFile(outputFile)
			assert.NoError(t, err, "Failed to read output file")
			assert.Equal(t, tt.expectedOutput, string(content), "Output file content should match the expected content")
		})
	}
}
