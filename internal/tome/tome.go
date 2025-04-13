package tome

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

type Tome struct {
	source  string
	target  string
	mode    os.FileMode
	strip   string
	include []string
	exclude []string
	copy    []string
	temp    []string
	values  map[string]any
}

func New(source, target, mode, strip string, include, exclude, copy, temp []string, values map[string]any) (*Tome, error) {

	if len(include) > 0 && len(exclude) > 0 {
		return nil, fmt.Errorf("cannot use both include and exclude patterns")
	}

	if len(copy) > 0 && len(temp) > 0 {
		return nil, fmt.Errorf("cannot use both copy-only and template-only patterns")
	}

	var fileMode os.FileMode
	var err error
	if mode != "" {
		fileMode, err = parseFileMode(mode)
		if err != nil {
			return nil, fmt.Errorf("invalid file mode: %w", err)
		}
	}

	return &Tome{
		source:  source,
		target:  target,
		mode:    fileMode,
		strip:   strip,
		include: include,
		exclude: exclude,
		copy:    copy,
		temp:    temp,
		values:  values,
	}, nil
}

func parseFileMode(modeStr string) (os.FileMode, error) {
	// Try parsing as octal
	if n, err := strconv.ParseUint(modeStr, 8, 32); err == nil {
		return os.FileMode(n), nil
	}

	// If not octal, parse symbolic string
	if len(modeStr) != 9 {
		return 0, fmt.Errorf("invalid symbolic file mode: %s", modeStr)
	}

	var mode os.FileMode

	symbols := []struct {
		char byte
		bit  os.FileMode
	}{
		{'r', 0400}, {'w', 0200}, {'x', 0100}, // user
		{'r', 0040}, {'w', 0020}, {'x', 0010}, // group
		{'r', 0004}, {'w', 0002}, {'x', 0001}, // others
	}

	for i, sym := range symbols {
		if modeStr[i] == sym.char {
			mode |= sym.bit
		} else if modeStr[i] != '-' {
			return 0, fmt.Errorf("unexpected character '%c' in symbolic mode", modeStr[i])
		}
	}

	return mode, nil
}

func (t *Tome) ShouldInclude(name string) bool {
	if len(t.include) > 0 {
		for _, pattern := range t.include {
			if matches, _ := filepath.Match(pattern, name); matches {
				return true
			}
		}
		return false
	}

	if len(t.exclude) > 0 {
		for _, pattern := range t.exclude {
			if matches, _ := filepath.Match(pattern, name); matches {
				return false
			}
		}
	}

	return true
}

func (t *Tome) shouldCopy(name string) bool {
	if len(t.copy) > 0 {
		for _, pattern := range t.copy {
			if matches, _ := filepath.Match(pattern, name); matches {
				return true
			}
		}
		return false
	}

	if len(t.temp) > 0 {
		for _, pattern := range t.temp {
			if matches, _ := filepath.Match(pattern, name); matches {
				return false
			}
		}
		return true
	}

	return false
}
