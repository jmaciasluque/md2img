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
	md := "## caf\u00e9 \u2014 \u2018hello\u2019\n\n| \u2713 | \u2717 |\n|----|----|\n| ok | no |"
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
