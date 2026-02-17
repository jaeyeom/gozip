package ziplib

import "testing"

func TestMatchesAny(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		patterns []string
		want     bool
	}{
		{"no patterns", "foo.txt", nil, false},
		{"empty patterns", "foo.txt", []string{}, false},
		{"exact match", "foo.txt", []string{"foo.txt"}, true},
		{"glob match", "foo.txt", []string{"*.txt"}, true},
		{"no match", "foo.txt", []string{"*.go"}, false},
		{"multiple patterns first", "foo.txt", []string{"*.txt", "*.go"}, true},
		{"multiple patterns second", "foo.go", []string{"*.txt", "*.go"}, true},
		{"multiple patterns none", "foo.rs", []string{"*.txt", "*.go"}, false},
		{"path uses base name", "dir/foo.txt", []string{"*.txt"}, true},
		{"invalid pattern ignored", "foo.txt", []string{"[invalid"}, false},
		{"question mark", "foo.txt", []string{"fo?.txt"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesAny(tt.input, tt.patterns)
			if got != tt.want {
				t.Errorf("matchesAny(%q, %v) = %v, want %v", tt.input, tt.patterns, got, tt.want)
			}
		})
	}
}
