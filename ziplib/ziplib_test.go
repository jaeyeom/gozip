package ziplib

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupTestDir creates a temporary directory with test files and returns its path.
func setupTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Create files.
	writeFile(t, filepath.Join(dir, "hello.txt"), "hello world\n")
	writeFile(t, filepath.Join(dir, "foo.go"), "package foo\n")

	// Create a subdirectory with a file.
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(sub, "nested.txt"), "nested content\n")

	return dir
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestZipUnzipRoundTrip(t *testing.T) {
	src := setupTestDir(t)
	zipPath := filepath.Join(t.TempDir(), "test.zip")
	extractDir := t.TempDir()

	// Zip individual files.
	err := Zip(zipPath, []string{
		filepath.Join(src, "hello.txt"),
		filepath.Join(src, "foo.go"),
	}, ZipOptions{})
	if err != nil {
		t.Fatalf("Zip: %v", err)
	}

	// Unzip.
	err = Unzip(zipPath, UnzipOptions{OutputDir: extractDir, Overwrite: true})
	if err != nil {
		t.Fatalf("Unzip: %v", err)
	}

	// Verify content. The extracted paths include the original absolute path
	// structure, so we check that the files exist somewhere under extractDir.
	// Since we zipped with absolute paths, the archive contains them as-is.
	got := readFile(t, filepath.Join(extractDir, filepath.Join(src, "hello.txt")))
	if got != "hello world\n" {
		t.Errorf("hello.txt content = %q, want %q", got, "hello world\n")
	}
}

func TestZipRecursive(t *testing.T) {
	src := setupTestDir(t)
	zipPath := filepath.Join(t.TempDir(), "recursive.zip")

	// Change to src dir so paths are relative.
	orig, _ := os.Getwd()
	if err := os.Chdir(src); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig)

	var buf bytes.Buffer
	err := Zip(zipPath, []string{"."}, ZipOptions{
		Recursive: true,
		Output:    &buf,
	})
	if err != nil {
		t.Fatalf("Zip recursive: %v", err)
	}

	entries, err := List(zipPath)
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	names := make(map[string]bool)
	for _, e := range entries {
		names[e.Name] = true
	}

	for _, want := range []string{"hello.txt", "foo.go", "sub/nested.txt"} {
		if !names[want] {
			t.Errorf("archive missing %s; got %v", want, names)
		}
	}
}

func TestZipNonRecursiveSkipsDir(t *testing.T) {
	src := setupTestDir(t)
	zipPath := filepath.Join(t.TempDir(), "nonrec.zip")

	var buf bytes.Buffer
	err := Zip(zipPath, []string{
		filepath.Join(src, "sub"),
		filepath.Join(src, "hello.txt"),
	}, ZipOptions{
		Recursive: false,
		Output:    &buf,
	})
	if err != nil {
		t.Fatalf("Zip: %v", err)
	}

	if !strings.Contains(buf.String(), "skipped") {
		t.Errorf("expected skip message, got: %s", buf.String())
	}

	entries, err := List(zipPath)
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	for _, e := range entries {
		if strings.Contains(e.Name, "nested") {
			t.Errorf("archive should not contain nested.txt without -r")
		}
	}
}

func TestZipExcludePatterns(t *testing.T) {
	src := setupTestDir(t)
	zipPath := filepath.Join(t.TempDir(), "exclude.zip")

	orig, _ := os.Getwd()
	if err := os.Chdir(src); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig)

	err := Zip(zipPath, []string{"."}, ZipOptions{
		Recursive:       true,
		ExcludePatterns: []string{"*.go"},
	})
	if err != nil {
		t.Fatalf("Zip: %v", err)
	}

	entries, err := List(zipPath)
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	for _, e := range entries {
		if strings.HasSuffix(e.Name, ".go") {
			t.Errorf("archive should not contain .go files, found: %s", e.Name)
		}
	}
}

func TestZipCompressionLevels(t *testing.T) {
	src := setupTestDir(t)

	tests := []struct {
		name  string
		level int
	}{
		{"store", 0},
		{"default", -1},
		{"best speed", 1},
		{"best compression", 9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zipPath := filepath.Join(t.TempDir(), "level.zip")
			err := Zip(zipPath, []string{filepath.Join(src, "hello.txt")}, ZipOptions{
				CompressionLevel: tt.level,
			})
			if err != nil {
				t.Fatalf("Zip level %d: %v", tt.level, err)
			}

			entries, err := List(zipPath)
			if err != nil {
				t.Fatalf("List: %v", err)
			}
			if len(entries) != 1 {
				t.Fatalf("expected 1 entry, got %d", len(entries))
			}
		})
	}
}

func TestUnzipOverwrite(t *testing.T) {
	src := setupTestDir(t)
	zipPath := filepath.Join(t.TempDir(), "overwrite.zip")
	extractDir := t.TempDir()

	err := Zip(zipPath, []string{filepath.Join(src, "hello.txt")}, ZipOptions{})
	if err != nil {
		t.Fatalf("Zip: %v", err)
	}

	// First extraction.
	err = Unzip(zipPath, UnzipOptions{OutputDir: extractDir, Overwrite: true})
	if err != nil {
		t.Fatalf("Unzip first: %v", err)
	}

	// Second extraction without overwrite should fail.
	err = Unzip(zipPath, UnzipOptions{OutputDir: extractDir, Overwrite: false})
	if err == nil {
		t.Fatal("expected error on overwrite=false with existing file")
	}
	if !strings.Contains(err.Error(), "file exists") {
		t.Errorf("expected 'file exists' error, got: %v", err)
	}

	// Second extraction with overwrite should succeed.
	err = Unzip(zipPath, UnzipOptions{OutputDir: extractDir, Overwrite: true})
	if err != nil {
		t.Fatalf("Unzip with overwrite: %v", err)
	}
}

func TestUnzipJunkPaths(t *testing.T) {
	src := setupTestDir(t)
	zipPath := filepath.Join(t.TempDir(), "junk.zip")
	extractDir := t.TempDir()

	orig, _ := os.Getwd()
	if err := os.Chdir(src); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig)

	err := Zip(zipPath, []string{"."}, ZipOptions{Recursive: true})
	if err != nil {
		t.Fatalf("Zip: %v", err)
	}

	err = Unzip(zipPath, UnzipOptions{
		OutputDir: extractDir,
		JunkPaths: true,
		Overwrite: true,
	})
	if err != nil {
		t.Fatalf("Unzip junk paths: %v", err)
	}

	// nested.txt should be at top level, not in sub/.
	if _, err := os.Stat(filepath.Join(extractDir, "nested.txt")); err != nil {
		t.Errorf("expected nested.txt at top level: %v", err)
	}
}

func TestUnzipFilePatterns(t *testing.T) {
	src := setupTestDir(t)
	zipPath := filepath.Join(t.TempDir(), "filter.zip")
	extractDir := t.TempDir()

	orig, _ := os.Getwd()
	if err := os.Chdir(src); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig)

	err := Zip(zipPath, []string{"."}, ZipOptions{Recursive: true})
	if err != nil {
		t.Fatalf("Zip: %v", err)
	}

	err = Unzip(zipPath, UnzipOptions{
		OutputDir:    extractDir,
		FilePatterns: []string{"*.txt"},
		Overwrite:    true,
	})
	if err != nil {
		t.Fatalf("Unzip with patterns: %v", err)
	}

	// .go file should not be extracted.
	goPath := filepath.Join(extractDir, "foo.go")
	if _, err := os.Stat(goPath); !os.IsNotExist(err) {
		t.Errorf("foo.go should not be extracted")
	}
}

func TestUnzipZipSlipPrevention(t *testing.T) {
	// Create a malicious zip with a path traversal entry.
	zipPath := filepath.Join(t.TempDir(), "evil.zip")
	extractDir := t.TempDir()

	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(f)
	fw, err := w.Create("../../etc/passwd")
	if err != nil {
		t.Fatal(err)
	}
	fw.Write([]byte("evil"))
	w.Close()
	f.Close()

	err = Unzip(zipPath, UnzipOptions{OutputDir: extractDir, Overwrite: true})
	if err == nil {
		t.Fatal("expected zip-slip error")
	}
	if !strings.Contains(err.Error(), "illegal file path") {
		t.Errorf("expected 'illegal file path' error, got: %v", err)
	}
}

func TestZipNonexistentFile(t *testing.T) {
	zipPath := filepath.Join(t.TempDir(), "bad.zip")
	err := Zip(zipPath, []string{"/nonexistent/file.txt"}, ZipOptions{})
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestUnzipInvalidArchive(t *testing.T) {
	// Create a file that is not a zip.
	badPath := filepath.Join(t.TempDir(), "notazip.zip")
	os.WriteFile(badPath, []byte("this is not a zip"), 0o644)

	err := Unzip(badPath, UnzipOptions{})
	if err == nil {
		t.Fatal("expected error for invalid archive")
	}
}

func TestListEntries(t *testing.T) {
	src := setupTestDir(t)
	zipPath := filepath.Join(t.TempDir(), "list.zip")

	orig, _ := os.Getwd()
	if err := os.Chdir(src); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig)

	err := Zip(zipPath, []string{"."}, ZipOptions{Recursive: true})
	if err != nil {
		t.Fatalf("Zip: %v", err)
	}

	entries, err := List(zipPath)
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if len(entries) < 3 {
		t.Fatalf("expected at least 3 entries, got %d", len(entries))
	}

	found := false
	for _, e := range entries {
		if e.Name == "hello.txt" {
			found = true
			if e.UncompressedSize == 0 {
				t.Error("expected non-zero uncompressed size for hello.txt")
			}
		}
	}
	if !found {
		t.Error("hello.txt not found in list")
	}
}

func TestListNonexistent(t *testing.T) {
	_, err := List("/nonexistent/file.zip")
	if err == nil {
		t.Fatal("expected error for nonexistent archive")
	}
}
