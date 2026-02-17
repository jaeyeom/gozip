# gozip

A Go implementation of zip and unzip, compatible with standard tools.

## Why?

Some Docker and cloud environments come with the Go toolchain but lack
`zip`/`unzip`, and installing system packages may require root access or
pulling in extra dependencies. With gozip, a single `go install` gives you
both tools — no package manager, no `apt-get`, no source compilation needed.
It is designed as a lightweight, quick-to-install replacement for environments
where the standard utilities are not readily available.

## Installation

```sh
go install github.com/jaeyeom/gozip/cmd/gozip@latest
go install github.com/jaeyeom/gozip/cmd/gounzip@latest
```

## Usage

### gozip — create zip archives

```sh
# Zip individual files
gozip archive.zip file1.txt file2.txt

# Zip a directory recursively
gozip -r archive.zip mydir/

# Set compression level (0=store, 1=fastest, 9=best)
gozip -r -9 archive.zip mydir/

# Exclude files by pattern
gozip -r -x '*.log' archive.zip mydir/
```

### gounzip — extract zip archives

```sh
# Extract an archive
gounzip archive.zip

# Extract to a specific directory
gounzip -d output/ archive.zip

# Overwrite existing files
gounzip -o archive.zip

# List archive contents
gounzip -l archive.zip

# Extract only matching files
gounzip archive.zip '*.txt'

# Strip directory paths on extraction
gounzip -j archive.zip
```

## Library

The `ziplib` package provides the core functionality for use in other Go programs:

```go
import "github.com/jaeyeom/gozip/ziplib"

// Create a zip archive.
err := ziplib.Zip("archive.zip", []string{"file1.txt", "dir/"}, ziplib.ZipOptions{
    Recursive:        true,
    CompressionLevel: 6,
    ExcludePatterns:  []string{"*.log"},
    Output:           os.Stdout,
})

// Extract a zip archive.
err := ziplib.Unzip("archive.zip", ziplib.UnzipOptions{
    OutputDir: "output/",
    Overwrite: true,
    Output:    os.Stdout,
})

// List archive entries.
entries, err := ziplib.List("archive.zip")
```

## Development

```sh
# Format, lint, test, and build
make all

# CI-friendly checks (no mutation)
make check

# Run tests only
make test

# Build binaries to bin/
make build
```

## License

[MIT](LICENSE)
