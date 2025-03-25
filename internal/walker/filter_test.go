package walker

import (
	"testing"
)

func TestShouldInclude(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		includes []string
		excludes []string
		expected bool
	}{
		{
			name:     "Exclude match",
			path:     "test/file.txt",
			includes: []string{"**/*.txt"},
			excludes: []string{"test/*"},
			expected: false,
		},
		{
			name:     "Include match",
			path:     "test/file.txt",
			includes: []string{"**/*.txt"},
			excludes: []string{},
			expected: true,
		},
		{
			name:     "No includes, no excludes",
			path:     "test/file.txt",
			includes: []string{},
			excludes: []string{},
			expected: true,
		},
		{
			name:     "No includes, exclude match",
			path:     "test/file.txt",
			includes: []string{},
			excludes: []string{"test/*"},
			expected: false,
		},
		{
			name:     "No match in includes",
			path:     "test/file.txt",
			includes: []string{"**/*.go"},
			excludes: []string{},
			expected: false,
		},
		{
			name:     "Exclude takes precedence over include",
			path:     "test/file.txt",
			includes: []string{"**/*.txt"},
			excludes: []string{"test/file.txt"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldInclude(tt.path, tt.includes, tt.excludes)
			if result != tt.expected {
				t.Errorf("ShouldInclude(%q, %v, %v) = %v; want %v", tt.path, tt.includes, tt.excludes, result, tt.expected)
			}
		})
	}
}
