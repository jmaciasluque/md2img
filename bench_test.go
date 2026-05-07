package md2img

import (
	"path/filepath"
	"testing"
)

// Benchmark inputs of increasing complexity.
var (
	benchSimple = "# Hello World"

	benchTable = `## Scores

| Name  | Score | Grade |
|-------|-------|-------|
| Alice | 95    | A     |
| Bob   | 87    | B     |
| Carol | 92    | A-    |
| Dave  | 78    | C+    |
| Eve   | 88    | B+    |
`

	benchComplex = `# Monthly Report

## Summary

This month we shipped three features and closed 12 bugs.

| Feature         | Status  | Author | Days |
|-----------------|---------|--------|------|
| Auth v2         | Shipped | Alice  | 5    |
| Dashboard rev   | Shipped | Bob    | 8    |
| API rate limits | In progress | Carol | 3 |
| Mobile fixes    | Shipped | Dave   | 2    |
| Docs update     | In progress | Eve | 1 |

## Code Changes

` + "```" + `go
func main() {
    fmt.Println("hello world")
}
` + "```" + `

> Important: Remember to update the changelog.

- Item one
- Item two
- Item three

---

Final notes go here.
`
)

func benchRender(b *testing.B, md string, cfg *Config) {
	b.Helper()
	out := filepath.Join(b.TempDir(), "bench.png")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var err error
		if cfg != nil {
			err = RenderWithConfig(md, out, *cfg)
		} else {
			err = Render(md, out)
		}
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRenderSimple(b *testing.B) {
	benchRender(b, benchSimple, nil)
}

func BenchmarkRenderTable(b *testing.B) {
	benchRender(b, benchTable, nil)
}

func BenchmarkRenderComplex(b *testing.B) {
	benchRender(b, benchComplex, nil)
}

func BenchmarkRenderPDF(b *testing.B) {
	cfg := DefaultConfig()
	cfg.AsPDF = true
	benchRender(b, benchComplex, &cfg)
}

func BenchmarkRenderDPI100(b *testing.B) {
	cfg := DefaultConfig()
	cfg.DPI = 100
	benchRender(b, benchComplex, &cfg)
}

func BenchmarkRenderDPI300(b *testing.B) {
	cfg := DefaultConfig()
	cfg.DPI = 300
	benchRender(b, benchComplex, &cfg)
}

func BenchmarkRenderTrimmed(b *testing.B) {
	cfg := DefaultConfig()
	cfg.Trim = true
	benchRender(b, benchComplex, &cfg)
}
