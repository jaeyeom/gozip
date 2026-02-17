package ziplib

import "path/filepath"

// matchesAny reports whether name matches any of the given glob patterns.
func matchesAny(name string, patterns []string) bool {
	for _, p := range patterns {
		if matched, err := filepath.Match(p, filepath.Base(name)); err == nil && matched {
			return true
		}
	}
	return false
}
