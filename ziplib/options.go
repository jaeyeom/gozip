// Package ziplib provides core zip/unzip functionality.
package ziplib

import (
	"io"
	"time"
)

// ZipOptions configures the behavior of the Zip function.
type ZipOptions struct {
	// Recursive enables recursive directory traversal.
	Recursive bool
	// CompressionLevel sets the flate compression level (0-9).
	// -1 means default compression.
	CompressionLevel int
	// ExcludePatterns is a list of glob patterns to exclude from the archive.
	ExcludePatterns []string
	// Output is where status messages are written. If nil, messages are discarded.
	Output io.Writer
}

// UnzipOptions configures the behavior of the Unzip function.
type UnzipOptions struct {
	// OutputDir is the directory to extract files into. Defaults to ".".
	OutputDir string
	// Overwrite allows overwriting existing files.
	Overwrite bool
	// JunkPaths strips directory components from file names on extraction.
	JunkPaths bool
	// FilePatterns filters which files to extract. Empty means extract all.
	FilePatterns []string
	// Output is where status messages are written. If nil, messages are discarded.
	Output io.Writer
}

// ListEntry holds metadata about a single entry in a zip archive.
type ListEntry struct {
	Name             string
	UncompressedSize uint64
	CompressedSize   uint64
	Modified         time.Time
	IsDir            bool
}
