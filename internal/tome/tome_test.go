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
  strip: "custom-strip"
  include: ["custom-include"]
  values:
    key: "custom-value"
`,
			base: Tome{
				Target: "/tmp",
			},
			expected: []Tome{
				{
					Target:  "/tmp/custom-target",
					Strip:   "custom-strip",
					Include: []string{"custom-include"},
					Values:  map[string]any{"key": "custom-value"},
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
				Target:  "/tmp/target",
				Strip:   "default-strip",
				Exclude: []string{"default-exclude"},
				Temp:    []string{"default-temp"},
				Values:  map[string]any{"key": "value"},
			},
			expected: []Tome{
				{
					Target:  "/tmp/target/{{ .tempdir }}",
					Strip:   "default-strip",
					Exclude: []string{"default-exclude"},
					Temp:    []string{"default-temp"},
					Values:  map[string]any{"key": "value"},
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
			tt.base.Source = filepath.Dir(tempDir)

			tomes, err := Load(tempFile.Name(), &tt.base)
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
				assert.Equal(t, strings.Replace(expectedTome.Target, "{{ .tempdir }}", filepath.Base(tempDir), -1), tomes[i].Target, "Target mismatch")
				assert.Equal(t, expectedTome.Strip, tomes[i].Strip, "Strip mismatch")
				assert.Equal(t, expectedTome.Include, tomes[i].Include, "Include mismatch")
				assert.Equal(t, expectedTome.Exclude, tomes[i].Exclude, "Exclude mismatch")
				assert.Equal(t, expectedTome.Copy, tomes[i].Copy, "Copy mismatch")
				assert.Equal(t, expectedTome.Temp, tomes[i].Temp, "Temp mismatch")
				assert.Equal(t, expectedTome.Values, tomes[i].Values, "Values mismatch")
			}
		})
	}
}
