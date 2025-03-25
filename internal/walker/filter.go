package walker

import (
	"github.com/bmatcuk/doublestar/v4"
)

func ShouldInclude(path string, includes, excludes []string) bool {
	// Match excludes first
	for _, pattern := range excludes {
		match, _ := doublestar.PathMatch(pattern, path)
		if match {
			return false
		}
	}

	// If no includes, include all
	if len(includes) == 0 {
		return true
	}

	for _, pattern := range includes {
		match, _ := doublestar.PathMatch(pattern, path)
		if match {
			return true
		}
	}
	return false
}
