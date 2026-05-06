package main

import (
	"os"
	"strings"
	"testing"
)

func TestSanitize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"em\u2014dash", "em-dash"},
		{"en\u2013dash", "en-dash"},
		{"ellipsis\u2026", "ellipsis..."},
		{"right\u2192arrow", "right->arrow"},
		{"left\u2190arrow", "left<-arrow"},
		{"not\u2260equal", "not!=equal"},
		{"bullet\u2022item", "bullet*item"},
		{"check\u2713mark", "check[OK]mark"},
		{"cross\u2717mark", "cross[X]mark"},
		{"\u201cquoted\u201d", "\"quoted\""},
		{"\u2018single\u2019", "'single'"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := sanitize(tt.input)
			if got != tt.expected {
				t.Errorf("sanitize(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSanitizePassthrough(t *testing.T) {
	// ASCII text should pass through unchanged
	input := "Hello World 123 !@#$%^&*()"
	got := sanitize(input)
	if got != input {
		t.Errorf("sanitize modified ASCII: got %q, want %q", got, input)
	}
}

func TestRenderCreatesFile(t *testing.T) {
	md := "# Hello World\n\nThis is a test."
	out := t.TempDir() + "/test.png"

	if err := run(md, out); err != nil {
		t.Fatalf("run() error: %v", err)
	}

	if _, err := os.Stat(out); os.IsNotExist(err) {
		t.Fatalf("output file not created: %s", out)
	}

	info, _ := os.Stat(out)
	if info.Size() == 0 {
		t.Fatal("output file is empty")
	}
}

func TestRenderTable(t *testing.T) {
	md := `## My Table

| Name  | Age |
|-------|-----|
| Alice | 30  |
| Bob   | 25  |
`
	out := t.TempDir() + "/table.png"

	if err := run(md, out); err != nil {
		t.Fatalf("run() error: %v", err)
	}

	info, _ := os.Stat(out)
	if info.Size() < 1000 {
		t.Errorf("table PNG seems too small: %d bytes", info.Size())
	}
}

func TestRenderCodeBlock(t *testing.T) {
	md := "```go\nfmt.Println(\"hello\")\n```"
	out := t.TempDir() + "/code.png"

	if err := run(md, out); err != nil {
		t.Fatalf("run() error: %v", err)
	}

	if _, err := os.Stat(out); os.IsNotExist(err) {
		t.Fatal("output file not created")
	}
}

func TestRenderList(t *testing.T) {
	md := "- Item one\n- Item two\n- Item three"
	out := t.TempDir() + "/list.png"

	if err := run(md, out); err != nil {
		t.Fatalf("run() error: %v", err)
	}

	if _, err := os.Stat(out); os.IsNotExist(err) {
		t.Fatal("output file not created")
	}
}

func TestRenderBlockquote(t *testing.T) {
	md := "> This is a wise quote."
	out := t.TempDir() + "/quote.png"

	if err := run(md, out); err != nil {
		t.Fatalf("run() error: %v", err)
	}

	if _, err := os.Stat(out); os.IsNotExist(err) {
		t.Fatal("output file not created")
	}
}

func TestRenderEmptyInput(t *testing.T) {
	// Empty input should be caught before run()
	// But run() itself should handle gracefully
	out := t.TempDir() + "/empty.png"
	err := run("   ", out)
	if err == nil {
		t.Log("run() with whitespace input did not error (acceptable)")
	}
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
	out := t.TempDir() + "/complex.png"

	if err := run(md, out); err != nil {
		t.Fatalf("run() error: %v", err)
	}

	info, _ := os.Stat(out)
	if info.Size() < 2000 {
		t.Errorf("complex doc PNG seems too small: %d bytes", info.Size())
	}
}

func TestUnicodeSanitized(t *testing.T) {
	// Ensure Unicode-heavy input doesn't crash
	md := "## caf\u00e9 \u2014 \u2018hello\u2019\n\n| \u2713 | \u2717 |\n|----|----|\n| ok | no |"
	out := t.TempDir() + "/unicode.png"

	if err := run(md, out); err != nil {
		t.Fatalf("run() with unicode error: %v", err)
	}

	if _, err := os.Stat(out); os.IsNotExist(err) {
		t.Fatal("output file not created")
	}
}

func TestOutputFlag(t *testing.T) {
	// Test that -o flag works via main() parsing
	// We can't easily test main() directly, but we test run()
	out := t.TempDir() + "/flagged.png"
	if err := run("# Test", out); err != nil {
		t.Fatalf("run() error: %v", err)
	}

	if _, err := os.Stat(out); os.IsNotExist(err) {
		t.Fatal("output file not created at specified path")
	}
}

// TestRenderOrderedList is a separate test for ordered lists
func TestRenderOrderedList(t *testing.T) {
	md := "1. First item\n2. Second item\n3. Third item"
	out := t.TempDir() + "/ordered.png"

	if err := run(md, out); err != nil {
		t.Fatalf("run() error: %v", err)
	}

	if _, err := os.Stat(out); os.IsNotExist(err) {
		t.Fatal("output file not created")
	}
}

func TestRenderHR(t *testing.T) {
	md := "Before\n\n---\n\nAfter"
	out := t.TempDir() + "/hr.png"

	if err := run(md, out); err != nil {
		t.Fatalf("run() error: %v", err)
	}

	if _, err := os.Stat(out); os.IsNotExist(err) {
		t.Fatal("output file not created")
	}
}

func TestSanitizeComplete(t *testing.T) {
	// Test a realistic mixed-content string
	input := "The \u201cbest\u201d talk \u2014 \u2713 recommended"
	got := sanitize(input)
	if !strings.Contains(got, "\"best\"") {
		t.Errorf("quotes not sanitized: %q", got)
	}
	if !strings.Contains(got, "talk -") {
		t.Errorf("em dash not sanitized: %q", got)
	}
	if !strings.Contains(got, "[OK]") {
		t.Errorf("check mark not sanitized: %q", got)
	}
}
