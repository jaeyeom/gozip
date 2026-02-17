package ziplib

import (
	"archive/zip"
	"compress/flate"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Zip creates a zip archive at zipPath containing the given files.
// Directories are included recursively only if opts.Recursive is true;
// otherwise a warning is printed and the directory is skipped.
func Zip(zipPath string, files []string, opts ZipOptions) error {
	out := opts.Output
	if out == nil {
		out = io.Discard
	}

	f, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("creating archive: %w", err)
	}
	defer f.Close()

	w := zip.NewWriter(f)
	defer w.Close()

	// Register custom compressor for the requested level.
	level := opts.CompressionLevel
	if level < -1 || level > 9 {
		level = -1
	}
	w.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
		return flate.NewWriter(out, level)
	})

	for _, name := range files {
		if err := addToZip(w, name, opts, out); err != nil {
			return err
		}
	}
	return nil
}

func addToZip(w *zip.Writer, path string, opts ZipOptions, out io.Writer) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat %s: %w", path, err)
	}

	if info.IsDir() {
		if !opts.Recursive {
			fmt.Fprintf(out, "  adding: %s/ (skipped, not recursive)\n", path)
			return nil
		}
		return filepath.Walk(path, func(p string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if matchesAny(p, opts.ExcludePatterns) {
				if fi.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if fi.IsDir() {
				return nil
			}
			return writeFileToZip(w, p, fi, opts, out)
		})
	}

	if matchesAny(path, opts.ExcludePatterns) {
		return nil
	}
	return writeFileToZip(w, path, info, opts, out)
}

func writeFileToZip(w *zip.Writer, path string, info os.FileInfo, opts ZipOptions, out io.Writer) error {
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return fmt.Errorf("file header %s: %w", path, err)
	}
	header.Name = filepath.ToSlash(path)

	if opts.CompressionLevel == 0 {
		header.Method = zip.Store
	} else {
		header.Method = zip.Deflate
	}

	fw, err := w.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("create header %s: %w", path, err)
	}

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	if _, err := io.Copy(fw, f); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}

	fmt.Fprintf(out, "  adding: %s\n", path)
	return nil
}

// Unzip extracts the contents of a zip archive.
func Unzip(zipPath string, opts UnzipOptions) error {
	out := opts.Output
	if out == nil {
		out = io.Discard
	}

	outputDir := opts.OutputDir
	if outputDir == "" {
		outputDir = "."
	}

	absOutputDir, err := filepath.Abs(outputDir)
	if err != nil {
		return fmt.Errorf("resolve output dir: %w", err)
	}

	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("open archive: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		if len(opts.FilePatterns) > 0 && !matchesAny(f.Name, opts.FilePatterns) {
			continue
		}

		name := f.Name
		if opts.JunkPaths {
			name = filepath.Base(name)
		}

		destPath := filepath.Join(outputDir, name)

		// Zip-slip prevention.
		absDest, err := filepath.Abs(destPath)
		if err != nil {
			return fmt.Errorf("resolve path: %w", err)
		}
		if !strings.HasPrefix(absDest, absOutputDir+string(os.PathSeparator)) && absDest != absOutputDir {
			return fmt.Errorf("illegal file path: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(destPath, f.Mode()); err != nil {
				return fmt.Errorf("mkdir %s: %w", destPath, err)
			}
			continue
		}

		if err := extractFile(f, destPath, opts.Overwrite, out); err != nil {
			return err
		}

		// Restore modification time.
		if err := os.Chtimes(destPath, f.Modified, f.Modified); err != nil {
			return fmt.Errorf("chtimes %s: %w", destPath, err)
		}
	}
	return nil
}

func extractFile(f *zip.File, destPath string, overwrite bool, out io.Writer) error {
	if !overwrite {
		if _, err := os.Stat(destPath); err == nil {
			return fmt.Errorf("file exists: %s (use overwrite option)", destPath)
		}
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return fmt.Errorf("mkdir for %s: %w", destPath, err)
	}

	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("open entry %s: %w", f.Name, err)
	}
	defer rc.Close()

	w, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
	if err != nil {
		return fmt.Errorf("create %s: %w", destPath, err)
	}
	defer w.Close()

	if _, err := io.Copy(w, rc); err != nil {
		return fmt.Errorf("extract %s: %w", f.Name, err)
	}

	fmt.Fprintf(out, "  inflating: %s\n", destPath)
	return nil
}

// List returns metadata for all entries in a zip archive.
func List(zipPath string) ([]ListEntry, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("open archive: %w", err)
	}
	defer r.Close()

	entries := make([]ListEntry, 0, len(r.File))
	for _, f := range r.File {
		entries = append(entries, ListEntry{
			Name:             f.Name,
			UncompressedSize: f.UncompressedSize64,
			CompressedSize:   f.CompressedSize64,
			Modified:         f.Modified,
			IsDir:            f.FileInfo().IsDir(),
		})
	}
	return entries, nil
}
