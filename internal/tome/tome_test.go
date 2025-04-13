package tome

import (
	"testing"
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
				include: []string{"*.txt"},
			},
			input:    "file.txt",
			expected: true,
		},
		{
			name: "No Match in include",
			tome: Tome{
				include: []string{"*.txt"},
			},
			input:    "file.jpg",
			expected: false,
		},
		{
			name: "Match in exclude",
			tome: Tome{
				exclude: []string{"*.txt"},
			},
			input:    "file.txt",
			expected: false,
		},
		{
			name: "No Match in exclude",
			tome: Tome{
				exclude: []string{"*.txt"},
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.tome.ShouldInclude(tt.input)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
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
				copy: []string{"*.txt"},
			},
			input:    "file.txt",
			expected: true,
		},
		{
			name: "No Match in copy",
			tome: Tome{
				copy: []string{"*.txt"},
			},
			input:    "file.jpg",
			expected: false,
		},
		{
			name: "Match in temp",
			tome: Tome{
				temp: []string{"*.txt"},
			},
			input:    "file.txt",
			expected: false,
		},
		{
			name: "No Match in temp",
			tome: Tome{
				temp: []string{"*.txt"},
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
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
