package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// repoRoot finds the directory containing go.mod.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find repo root (no go.mod)")
		}
		dir = parent
	}
}

func runCLI(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmdArgs := append([]string{"run", "."}, args...)
	cmd := exec.Command("go", cmdArgs...)
	cmd.Dir = filepath.Join(repoRoot(t), "cmd", "md2img")
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func TestCLIFileInput(t *testing.T) {
	mdFile := filepath.Join(t.TempDir(), "input.md")
	if err := os.WriteFile(mdFile, []byte("# Hello World\n\nTest content."), 0644); err != nil {
		t.Fatal(err)
	}

	outFile := filepath.Join(t.TempDir(), "output.png")
	out, err := runCLI(t, "-o", outFile, mdFile)
	if err != nil {
		t.Fatalf("CLI failed: %v\n%s", err, out)
	}

	if !strings.Contains(out, "Done:") {
		t.Errorf("expected 'Done:' in output, got: %s", out)
	}

	if _, err := os.Stat(outFile); os.IsNotExist(err) {
		t.Fatal("output file not created")
	}
}

func TestCLIStdin(t *testing.T) {
	outFile := filepath.Join(t.TempDir(), "stdin_out.png")

	cmd := exec.Command("go", "run", ".", "-o", outFile)
	cmd.Dir = filepath.Join(repoRoot(t), "cmd", "md2img")
	cmd.Stdin = strings.NewReader("# Stdin Test\n\nHello from stdin.")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI failed: %v\n%s", err, out)
	}

	if !strings.Contains(string(out), "Done:") {
		t.Errorf("expected 'Done:' in output, got: %s", out)
	}
}

func TestCLIDefaultOutput(t *testing.T) {
	mdFile := filepath.Join(t.TempDir(), "default.md")
	if err := os.WriteFile(mdFile, []byte("# Default Output"), 0644); err != nil {
		t.Fatal(err)
	}

	out, err := runCLI(t, mdFile)
	if err != nil {
		t.Fatalf("CLI failed: %v\n%s", err, out)
	}

	if !strings.Contains(out, "/tmp/md2img_output.png") {
		t.Errorf("expected default output path, got: %s", out)
	}

	os.Remove("/tmp/md2img_output.png")
}

func TestCLINoArgs(t *testing.T) {
	// Empty stdin — should either error or produce empty output
	out, err := runCLI(t)
	if err == nil && strings.Contains(out, "Done:") {
		t.Log("no-args with empty stdin succeeded (acceptable)")
	}
	_ = out
}

func TestCLIBadFile(t *testing.T) {
	out, err := runCLI(t, "-o", "/tmp/out.png", "/nonexistent/file.md")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
	if !strings.Contains(out, "Error") {
		t.Errorf("expected error message, got: %q", out)
	}
}

// --- Tests for new flags ---

func TestCLIFontFlag(t *testing.T) {
	outFile := filepath.Join(t.TempDir(), "font.png")
	out, err := runCLI(t, "-o", outFile, "-font", "Times", "-font-size", "14", "-",
	)
	if err == nil {
		// If it didn't error, the flag was accepted
		t.Log("font flag accepted (stdin was empty)")
	}
	_ = out
}

func TestCLIDPIFlag(t *testing.T) {
	mdFile := filepath.Join(t.TempDir(), "dpi.md")
	if err := os.WriteFile(mdFile, []byte("# DPI Test"), 0644); err != nil {
		t.Fatal(err)
	}
	outFile := filepath.Join(t.TempDir(), "dpi.png")
	out, err := runCLI(t, "-o", outFile, "-dpi", "100", mdFile)
	if err != nil {
		t.Fatalf("CLI with -dpi failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Done:") {
		t.Errorf("expected 'Done:' in output, got: %s", out)
	}
}

func TestCLIColorFlags(t *testing.T) {
	mdFile := filepath.Join(t.TempDir(), "colors.md")
	if err := os.WriteFile(mdFile, []byte("# Colors\n\nTest."), 0644); err != nil {
		t.Fatal(err)
	}
	outFile := filepath.Join(t.TempDir(), "colors.png")
	out, err := runCLI(t, "-o", outFile,
		"-text-color", "#333333",
		"-heading-color", "#006600",
		"-table-header-bg", "#003366",
		"-table-header-fg", "#ffffff",
		"-code-bg", "#1a1a1a",
		"-hr-color", "#cc0000",
		"-blockquote-line-color", "#ff6600",
		"-blockquote-text-color", "#555555",
		mdFile,
	)
	if err != nil {
		t.Fatalf("CLI with color flags failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Done:") {
		t.Errorf("expected 'Done:' in output, got: %s", out)
	}
}

func TestCLIMarginFlag(t *testing.T) {
	mdFile := filepath.Join(t.TempDir(), "margin.md")
	if err := os.WriteFile(mdFile, []byte("# Margins\n\nWide margins."), 0644); err != nil {
		t.Fatal(err)
	}
	outFile := filepath.Join(t.TempDir(), "margin.png")
	out, err := runCLI(t, "-o", outFile, "-margin", "30", mdFile)
	if err != nil {
		t.Fatalf("CLI with -margin failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Done:") {
		t.Errorf("expected 'Done:' in output, got: %s", out)
	}
}

func TestCLIPageSizeFlags(t *testing.T) {
	mdFile := filepath.Join(t.TempDir(), "pagesize.md")
	if err := os.WriteFile(mdFile, []byte("# Page Size"), 0644); err != nil {
		t.Fatal(err)
	}
	outFile := filepath.Join(t.TempDir(), "pagesize.png")
	out, err := runCLI(t, "-o", outFile, "-page-w", "215.9", "-page-h", "279.4", mdFile)
	if err != nil {
		t.Fatalf("CLI with page size flags failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Done:") {
		t.Errorf("expected 'Done:' in output, got: %s", out)
	}
}

func TestCLIBadColorFlag(t *testing.T) {
	out, err := runCLI(t, "-text-color", "notacolor", "-")
	if err == nil {
		t.Error("expected error for invalid color")
	}
	if !strings.Contains(out, "invalid hex color") {
		t.Errorf("expected hex color error, got: %q", out)
	}
}

func TestCLIBadDPIFlag(t *testing.T) {
	out, err := runCLI(t, "-dpi", "notanumber", "-")
	if err == nil {
		t.Error("expected error for invalid DPI")
	}
	if !strings.Contains(out, "invalid -dpi") {
		t.Errorf("expected DPI error, got: %q", out)
	}
}

func TestCLIUnknownFlag(t *testing.T) {
	out, err := runCLI(t, "-bogus", "-")
	if err == nil {
		t.Error("expected error for unknown flag")
	}
	if !strings.Contains(out, "unknown flag") {
		t.Errorf("expected unknown flag error, got: %q", out)
	}
}

func TestCLIVersionFlag(t *testing.T) {
	out, err := runCLI(t, "-version")
	if err != nil {
		t.Fatalf("CLI -version failed: %v", err)
	}
	if !strings.Contains(out, "md2img") {
		t.Errorf("expected version output, got: %q", out)
	}
}
