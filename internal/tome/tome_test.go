package tome

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldInclude(t *testing.T) {
	tests := []struct {
		name     string
		tome     Tome
		input    string
		expected bool
	}{
		{
			name: "Match in include",
			tome: Tome{
				Include: []string{"*.txt"},
			},
			input:    "file.txt",
			expected: true,
		},
		{
			name: "No Match in include",
			tome: Tome{
				Include: []string{"*.txt"},
			},
			input:    "file.jpg",
			expected: false,
		},
		{
			name: "Match in exclude",
			tome: Tome{
				Exclude: []string{"*.txt"},
			},
			input:    "file.txt",
			expected: false,
		},
		{
			name: "No Match in exclude",
			tome: Tome{
				Exclude: []string{"*.txt"},
			},
			input:    "file.jpg",
			expected: true,
		},
		{
			name:     "No include or exclude",
			tome:     Tome{},
			input:    "file.txt",
			expected: true,
		},
		{
			name: "Match in subdir include",
			tome: Tome{
				Include: []string{"**/*.txt"},
			},
			input:    "test.txt",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.tome.ShouldInclude(tt.input)
			assert.Equal(t, tt.expected, result, "Expected %v, got %v", tt.expected, result)
		})
	}
}

func TestShouldCopy(t *testing.T) {
	tests := []struct {
		name     string
		tome     Tome
		input    string
		expected bool
	}{
		{
			name: "Match in copy",
			tome: Tome{
				Copy: []string{"**/*.txt"},
			},
			input:    "test/file.txt",
			expected: true,
		},
		{
			name: "No Match in copy",
			tome: Tome{
				Copy: []string{"*.txt"},
			},
			input:    "file.jpg",
			expected: false,
		},
		{
			name: "Match in temp",
			tome: Tome{
				Temp: []string{"*.txt"},
			},
			input:    "file.txt",
			expected: false,
		},
		{
			name: "No Match in temp",
			tome: Tome{
				Temp: []string{"*.txt"},
			},
			input:    "file.jpg",
			expected: true,
		},
		{
			name:     "No copy or temp",
			tome:     Tome{},
			input:    "file.txt",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.tome.shouldCopy(tt.input)
			assert.Equal(t, tt.expected, result, "Expected %v, got %v", tt.expected, result)
		})
	}
}
