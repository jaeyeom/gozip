// Package e2e tests interoperability between gozip/gounzip and system zip/unzip.
package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// testFiles defines the set of files created for each test.
var testFiles = map[string]string{
	"hello.txt":         "hello world\n",
	"empty.txt":         "",
	"sub/nested.txt":    "nested content\n",
	"sub/deep/deep.txt": "deep content\n",
}

// setupTestData creates a directory with test files and returns its path.
func setupTestData(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range testFiles {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// buildBinaries builds gozip and gounzip into a temp directory and returns the paths.
func buildBinaries(t *testing.T) (gozip, gounzip string) {
	t.Helper()
	binDir := t.TempDir()
	gozip = filepath.Join(binDir, "gozip")
	gounzip = filepath.Join(binDir, "gounzip")

	projectRoot := ".."
	for _, bin := range []struct {
		output, pkg string
	}{
		{gozip, "./cmd/gozip"},
		{gounzip, "./cmd/gounzip"},
	} {
		cmd := exec.Command("go", "build", "-o", bin.output, bin.pkg) //nolint:gosec // Test-only; args are not user-controlled.
		cmd.Dir = projectRoot
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("build %s: %v\n%s", bin.pkg, err, out)
		}
	}
	return gozip, gounzip
}

// requireCmd skips the test if a system command is not available.
func requireCmd(t *testing.T, name string) {
	t.Helper()
	if _, err := exec.LookPath(name); err != nil {
		t.Skipf("%s not found in PATH, skipping", name)
	}
}

// verifyExtracted checks that all testFiles exist with correct content under dir.
func verifyExtracted(t *testing.T, dir string) {
	t.Helper()
	for name, wantContent := range testFiles {
		p := filepath.Join(dir, name)
		got, err := os.ReadFile(p)
		if err != nil {
			t.Errorf("read %s: %v", name, err)
			continue
		}
		if string(got) != wantContent {
			t.Errorf("%s: got %q, want %q", name, string(got), wantContent)
		}
	}
}

// TestSystemZipToGounzip creates an archive with system zip and extracts with gounzip.
func TestSystemZipToGounzip(t *testing.T) {
	requireCmd(t, "zip")
	_, gounzipBin := buildBinaries(t)

	srcDir := setupTestData(t)
	zipPath := filepath.Join(t.TempDir(), "system.zip")
	extractDir := t.TempDir()

	// Create archive with system zip.
	//   zip -r archive.zip .   (from within srcDir)
	cmd := exec.Command("zip", "-r", zipPath, ".")
	cmd.Dir = srcDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("system zip: %v\n%s", err, out)
	}

	// Extract with gounzip.
	cmd = exec.Command(gounzipBin, "-o", "-d", extractDir, zipPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("gounzip: %v\n%s", err, out)
	}

	verifyExtracted(t, extractDir)
}

// TestGozipToSystemUnzip creates an archive with gozip and extracts with system unzip.
func TestGozipToSystemUnzip(t *testing.T) {
	requireCmd(t, "unzip")
	gozipBin, _ := buildBinaries(t)

	srcDir := setupTestData(t)
	zipPath := filepath.Join(t.TempDir(), "gozip.zip")
	extractDir := t.TempDir()

	// Create archive with gozip.
	//   gozip -r archive.zip .   (from within srcDir)
	cmd := exec.Command(gozipBin, "-r", zipPath, ".")
	cmd.Dir = srcDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("gozip: %v\n%s", err, out)
	}

	// Extract with system unzip.
	//   unzip -o archive.zip -d extractDir
	cmd = exec.Command("unzip", "-o", zipPath, "-d", extractDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("system unzip: %v\n%s", err, out)
	}

	verifyExtracted(t, extractDir)
}

// TestGozipToGounzipRoundTrip creates an archive with gozip and extracts with gounzip.
func TestGozipToGounzipRoundTrip(t *testing.T) {
	gozipBin, gounzipBin := buildBinaries(t)

	srcDir := setupTestData(t)
	zipPath := filepath.Join(t.TempDir(), "roundtrip.zip")
	extractDir := t.TempDir()

	// Create archive with gozip.
	cmd := exec.Command(gozipBin, "-r", zipPath, ".")
	cmd.Dir = srcDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("gozip: %v\n%s", err, out)
	}

	// Extract with gounzip.
	cmd = exec.Command(gounzipBin, "-o", "-d", extractDir, zipPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("gounzip: %v\n%s", err, out)
	}

	verifyExtracted(t, extractDir)
}

// TestSystemZipToSystemUnzipBaseline confirms the system tools work as expected.
func TestSystemZipToSystemUnzipBaseline(t *testing.T) {
	requireCmd(t, "zip")
	requireCmd(t, "unzip")

	srcDir := setupTestData(t)
	zipPath := filepath.Join(t.TempDir(), "baseline.zip")
	extractDir := t.TempDir()

	cmd := exec.Command("zip", "-r", zipPath, ".")
	cmd.Dir = srcDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("system zip: %v\n%s", err, out)
	}

	cmd = exec.Command("unzip", "-o", zipPath, "-d", extractDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("system unzip: %v\n%s", err, out)
	}

	verifyExtracted(t, extractDir)
}

// TestGozipCompressionLevelsWithSystemUnzip tests that various compression levels
// produce archives that system unzip can handle.
func TestGozipCompressionLevelsWithSystemUnzip(t *testing.T) {
	requireCmd(t, "unzip")
	gozipBin, _ := buildBinaries(t)

	srcDir := setupTestData(t)

	levels := []string{"-0", "-1", "-5", "-9"}
	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			zipPath := filepath.Join(t.TempDir(), "level.zip")
			extractDir := t.TempDir()

			cmd := exec.Command(gozipBin, "-r", level, zipPath, ".")
			cmd.Dir = srcDir
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("gozip %s: %v\n%s", level, err, out)
			}

			cmd = exec.Command("unzip", "-o", zipPath, "-d", extractDir)
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("system unzip: %v\n%s", err, out)
			}

			verifyExtracted(t, extractDir)
		})
	}
}

// TestGounzipListMatchesSystemUnzip verifies that gounzip -l produces output
// for an archive created by system zip.
func TestGounzipListMatchesSystemUnzip(t *testing.T) {
	requireCmd(t, "zip")
	_, gounzipBin := buildBinaries(t)

	srcDir := setupTestData(t)
	zipPath := filepath.Join(t.TempDir(), "list.zip")

	cmd := exec.Command("zip", "-r", zipPath, ".")
	cmd.Dir = srcDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("system zip: %v\n%s", err, out)
	}

	// gounzip -l should succeed and show entries.
	cmd = exec.Command(gounzipBin, "-l", zipPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("gounzip -l: %v\n%s", err, out)
	}

	output := string(out)
	for name := range testFiles {
		base := filepath.Base(name)
		if !containsString(output, base) {
			t.Errorf("gounzip -l output missing %q", base)
		}
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
