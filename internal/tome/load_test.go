package tome

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {

	tests := []struct {
		name        string
		fileContent string
		base        Tome
		expected    []Tome
		expectError bool
	}{
		{
			name: "Valid single tome",
			fileContent: `
- target: "custom-target"
  strip: 
  - "custom-strip"
  include: ["custom-include"]
  values:
    key: "custom-value"
`,
			base: Tome{
				target: "/tmp",
			},
			expected: []Tome{
				{
					target:  "/tmp/custom-target",
					strip:   []string{"custom-strip"},
					include: []string{"custom-include"},
					values:  map[string]any{"key": "custom-value"},
				},
			},
			expectError: false,
		},
		{
			name: "Empty tome inherits base",
			fileContent: `
- {}
`,
			base: Tome{
				target:  "/tmp/target",
				strip:   []string{"default-strip"},
				exclude: []string{"default-exclude"},
				temp:    []string{"default-temp"},
				values:  map[string]any{"key": "value"},
			},
			expected: []Tome{
				{
					target:  "/tmp/target/{{ .tempdir }}",
					strip:   []string{"default-strip"},
					exclude: []string{"default-exclude"},
					temp:    []string{"default-temp"},
					values:  map[string]any{"key": "value"},
				},
			},
			expectError: false,
		},
		{
			name: "Invalid YAML",
			fileContent: `
- target: "custom-target
`,
			expected:    nil,
			expectError: true,
		},
		{
			name: "Include and exclude conflict",
			fileContent: `
- include: ["include-pattern"]
  exclude: ["exclude-pattern"]
`,
			expected:    nil,
			expectError: true,
		},
		{
			name: "Copy and temp conflict",
			fileContent: `
- copy: ["copy-pattern"]
  temp: ["temp-pattern"]
`,
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			tempFile, err := os.CreateTemp(tempDir, ".tome.yaml")
			if err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}

			_, err = tempFile.WriteString(tt.fileContent)
			if err != nil {
				t.Fatalf("failed to write to temp file: %v", err)
			}
			tempFile.Close()
			tt.base.source = filepath.Dir(tempDir)

			tomes, err := LoadTomeFile(tempFile.Name(), &tt.base)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected an error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(tomes) != len(tt.expected) {
				t.Errorf("expected %d tomes, got %d", len(tt.expected), len(tomes))
				return
			}

			for i, expectedTome := range tt.expected {
				assert.Equal(t, strings.Replace(expectedTome.target, "{{ .tempdir }}", filepath.Base(tempDir), -1), tomes[i].target, "Target mismatch")
				assert.Equal(t, expectedTome.strip, tomes[i].strip, "Strip mismatch")
				assert.Equal(t, expectedTome.include, tomes[i].include, "Include mismatch")
				assert.Equal(t, expectedTome.exclude, tomes[i].exclude, "Exclude mismatch")
				assert.Equal(t, expectedTome.copy, tomes[i].copy, "Copy mismatch")
				assert.Equal(t, expectedTome.temp, tomes[i].temp, "Temp mismatch")
				assert.Equal(t, expectedTome.values, tomes[i].values, "Values mismatch")
			}
		})
	}
}
