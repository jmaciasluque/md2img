package md2img

import (
	"os"
	"path/filepath"
	"testing"
)

func renderToFile(t *testing.T, md, filename string) string {
	t.Helper()
	out := filepath.Join(t.TempDir(), filename)
	if err := Render(md, out); err != nil {
		t.Fatalf("Render() error: %v", err)
	}
	return out
}

func renderToFileWithConfig(t *testing.T, md, filename string, cfg Config) string {
	t.Helper()
	out := filepath.Join(t.TempDir(), filename)
	if err := RenderWithConfig(md, out, cfg); err != nil {
		t.Fatalf("RenderWithConfig() error: %v", err)
	}
	return out
}

func requireFile(t *testing.T, path string) os.FileInfo {
	t.Helper()
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		t.Fatalf("output file not created: %s", path)
	}
	if err != nil {
		t.Fatalf("stat error: %v", err)
	}
	if info.Size() == 0 {
		t.Fatalf("output file is empty: %s", path)
	}
	return info
}

func TestRenderHeading(t *testing.T) {
	out := renderToFile(t, "# Hello World", "heading.png")
	info := requireFile(t, out)
	if info.Size() < 500 {
		t.Errorf("heading PNG too small: %d bytes", info.Size())
	}
}

func TestRenderParagraph(t *testing.T) {
	out := renderToFile(t, "Just a simple paragraph.", "para.png")
	requireFile(t, out)
}

func TestRenderTable(t *testing.T) {
	md := `## My Table

| Name  | Age |
|-------|-----|
| Alice | 30  |
| Bob   | 25  |
`
	out := renderToFile(t, md, "table.png")
	info := requireFile(t, out)
	if info.Size() < 1000 {
		t.Errorf("table PNG too small: %d bytes", info.Size())
	}
}

func TestRenderTableSingleColumn(t *testing.T) {
	md := `| Item |
|------|
| One  |
| Two  |
`
	out := renderToFile(t, md, "single_col.png")
	requireFile(t, out)
}

func TestRenderCodeBlock(t *testing.T) {
	md := "```go\nfmt.Println(\"hello\")\n```"
	out := renderToFile(t, md, "code.png")
	requireFile(t, out)
}

func TestRenderIndentedCodeBlock(t *testing.T) {
	md := "    indented code\n    more code"
	out := renderToFile(t, md, "indented_code.png")
	requireFile(t, out)
}

func TestRenderBulletList(t *testing.T) {
	md := "- Item one\n- Item two\n- Item three"
	out := renderToFile(t, md, "list.png")
	requireFile(t, out)
}

func TestRenderOrderedList(t *testing.T) {
	md := "1. First\n2. Second\n3. Third"
	out := renderToFile(t, md, "ordered.png")
	requireFile(t, out)
}

func TestRenderBlockquote(t *testing.T) {
	md := "> This is a wise quote."
	out := renderToFile(t, md, "quote.png")
	requireFile(t, out)
}

func TestRenderHR(t *testing.T) {
	md := "Before\n\n---\n\nAfter"
	out := renderToFile(t, md, "hr.png")
	requireFile(t, out)
}

func TestRenderUnicode(t *testing.T) {
	md := "## café — 'hello'\n\n| ✓ | ✗ |\n|----|----|\n| ok | no |"
	out := renderToFile(t, md, "unicode.png")
	requireFile(t, out)
}

func TestRenderComplexDocument(t *testing.T) {
	md := `# Main Title

## Sub heading

Some paragraph text with **bold** and *italic*.

| Col A | Col B | Col C |
|-------|-------|-------|
| 1     | 2     | 3     |
| 4     | 5     | 6     |

- List item one
- List item two

> A blockquote

` + "```" + `
code here
` + "```" + `

---

Final paragraph.
`
	out := renderToFile(t, md, "complex.png")
	info := requireFile(t, out)
	if info.Size() < 2000 {
		t.Errorf("complex doc PNG too small: %d bytes", info.Size())
	}
}

func TestRenderEmptyInput(t *testing.T) {
	out := filepath.Join(t.TempDir(), "empty.png")
	err := Render("   ", out)
	if err == nil {
		t.Log("Render() with whitespace did not error (acceptable)")
	}
}

func TestRenderLongDocument(t *testing.T) {
	// Test page break handling
	md := "# Title\n\n"
	for i := 0; i < 50; i++ {
		md += "This is paragraph number " + string(rune('0'+i%10)) + ".\n\n"
	}
	out := renderToFile(t, md, "long.png")
	info := requireFile(t, out)
	// Should produce a multi-page PDF converted to PNG
	if info.Size() < 5000 {
		t.Errorf("long doc PNG too small: %d bytes", info.Size())
	}
}

// --- Tests for custom Config ---

func TestRenderCustomFont(t *testing.T) {
	cfg := DefaultConfig()
	cfg.FontFamily = "Times"
	out := renderToFileWithConfig(t, "# Times Font\n\nUsing Times New Roman style.", "custom_font.png", cfg)
	requireFile(t, out)
}

func TestRenderCustomFontSize(t *testing.T) {
	cfg := DefaultConfig()
	cfg.FontSize = 16
	out := renderToFileWithConfig(t, "Large body text for testing.", "custom_fontsize.png", cfg)
	requireFile(t, out)
}

func TestRenderCustomColors(t *testing.T) {
	cfg := DefaultConfig()
	cfg.TextColor = Color{R: 100, G: 0, B: 0}
	cfg.HeadingColor = Color{R: 0, G: 100, B: 0}
	out := renderToFileWithConfig(t, "# Green Headline\n\nRed body text.", "custom_colors.png", cfg)
	requireFile(t, out)
}

func TestRenderCustomTableColors(t *testing.T) {
	cfg := DefaultConfig()
	cfg.TableHeaderBg = Color{R: 0, G: 100, B: 0}
	cfg.TableHeaderFg = Color{R: 255, G: 255, B: 255}
	cfg.TableRowEven = Color{R: 220, G: 255, B: 220}
	cfg.TableRowOdd = Color{R: 255, G: 255, B: 255}

	md := `| Col A | Col B |
|-------|-------|
| 1     | 2     |
| 3     | 4     |
| 5     | 6     |
`
	out := renderToFileWithConfig(t, md, "custom_table.png", cfg)
	info := requireFile(t, out)
	if info.Size() < 500 {
		t.Errorf("custom table PNG too small: %d bytes", info.Size())
	}
}

func TestRenderCustomCodeStyle(t *testing.T) {
	cfg := DefaultConfig()
	cfg.CodeBg = Color{R: 30, G: 30, B: 30}
	cfg.CodeFontSize = 10

	md := "```go\nfmt.Println(\"dark code block\")\n```"
	out := renderToFileWithConfig(t, md, "custom_code.png", cfg)
	requireFile(t, out)
}

func TestRenderCustomMargins(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MarginLeft = 30
	cfg.MarginRight = 30
	cfg.MarginTop = 20

	out := renderToFileWithConfig(t, "# Wide Margins\n\nThis has extra margins on each side.", "custom_margins.png", cfg)
	requireFile(t, out)
}

func TestRenderCustomPageSize(t *testing.T) {
	cfg := DefaultConfig()
	// Letter size: 215.9 x 279.4 mm
	cfg.PageWidth = 215.9
	cfg.PageHeight = 279.4

	out := renderToFileWithConfig(t, "# Letter Size\n\nThis renders on US Letter.", "custom_page.png", cfg)
	requireFile(t, out)
}

func TestRenderCustomDPI(t *testing.T) {
	cfg := DefaultConfig()
	cfg.DPI = 100 // Lower DPI for faster test

	out := renderToFileWithConfig(t, "# Low DPI", "custom_dpi.png", cfg)
	requireFile(t, out)
}

func TestRenderBlockquoteColors(t *testing.T) {
	cfg := DefaultConfig()
	cfg.BlockquoteLineColor = Color{R: 200, G: 0, B: 0}
	cfg.BlockquoteTextColor = Color{R: 50, G: 50, B: 50}

	out := renderToFileWithConfig(t, "> Custom blockquote styling.", "custom_quote.png", cfg)
	requireFile(t, out)
}

func TestRenderHRColor(t *testing.T) {
	cfg := DefaultConfig()
	cfg.HRColor = Color{R: 200, G: 0, B: 0}
	cfg.HRLineWidth = 1.0

	out := renderToFileWithConfig(t, "Before\n\n---\n\nAfter", "custom_hr.png", cfg)
	requireFile(t, out)
}

func TestRenderTrim(t *testing.T) {
	md := "# Title\n\n| A | B |\n|---|---|\n| 1 | 2 |"
	cfg := DefaultConfig()
	cfg.Trim = true
	out := renderToFileWithConfig(t, md, "trimmed.png", cfg)
	info := requireFile(t, out)
	// Trimmed output should be significantly smaller than full page.
	if info.Size() > 5000 {
		t.Errorf("trimmed PNG too large (%d bytes) — trim may not have worked", info.Size())
	}
}

func TestHexToColor(t *testing.T) {
	tests := []struct {
		input string
		want  Color
		err   bool
	}{
		{"#333366", Color{51, 51, 102}, false},
		{"ff0000", Color{255, 0, 0}, false},
		{"#fff", Color{255, 255, 255}, false},
		{"#GG0000", Color{}, true},
		{"abc", Color{170, 187, 204}, false}, // 3-char shorthand → aabbcc
		{"12", Color{}, true},   // too short
	}
	for _, tt := range tests {
		c, err := HexToColor(tt.input)
		if tt.err && err == nil {
			t.Errorf("HexToColor(%q) expected error, got %v", tt.input, c)
		}
		if !tt.err && err != nil {
			t.Errorf("HexToColor(%q) unexpected error: %v", tt.input, err)
		}
		if !tt.err && err == nil && c != tt.want {
			t.Errorf("HexToColor(%q) = %v, want %v", tt.input, c, tt.want)
		}
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.FontFamily != "Helvetica" {
		t.Errorf("default FontFamily = %q, want Helvetica", cfg.FontFamily)
	}
	if cfg.DPI != 200 {
		t.Errorf("default DPI = %d, want 200", cfg.DPI)
	}
	if cfg.PageWidth != 210 || cfg.PageHeight != 297 {
		t.Errorf("default page = %.0fx%.0f, want 210x297", cfg.PageWidth, cfg.PageHeight)
	}
}
