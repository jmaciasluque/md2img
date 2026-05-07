# md2img

Convert Markdown to styled PNG images. No browser, no Python — just a Go binary.

```
markdown → goldmark (parser) → Go image (render) → PNG
```

## Install

### Homebrew (macOS/Linux)

```bash
brew install jmaciasluque/tap/md2img
```

### Pre-built binaries

Download the latest release for your platform from the [Releases page](https://github.com/jmaciasluque/md2img/releases).

### From source

```bash
git clone https://github.com/jmaciasluque/md2img.git
cd md2img
make build    # → ./md2img
make install  # → ~/go/bin/md2img
```

### With Go

```bash
go install github.com/jmaciasluque/md2img/cmd/md2img@latest
```

**No external dependencies.** Just a single Go binary. TTF fonts are loaded from your system (Arial, Courier New, Times New Roman on macOS).

## Usage

```bash
# From stdin
echo "| Name | Score |\n|------|-------|\n| Alice | 95 |" | md2img -o scores.png

# Explicit stdin marker
echo "# Hello" | md2img -o hello.png -

# From file
md2img -o output.png input.md

# Default output: /tmp/md2img_output.png
echo "# Hello" | md2img

# Trim whitespace (tight crop)
echo "| A | B |" | md2img -o tight.png -trim
```

## CLI Flags

### Output

| Flag | Description | Default |
|------|-------------|---------|
| `-o`, `--output` | Output file path | `/tmp/md2img_output.png` |
| `-trim`, `--trim` | Auto-crop whitespace from PNG output | `false` |
| `-trim-padding`, `--trim-padding` | Padding around content after trim (mm) | `5` |
| `-dpi`, `--dpi` | Image resolution (DPI) | `200` |
| `-version`, `--version` | Print version | — |

### Font

| Flag | Description | Default |
|------|-------------|---------|
| `-font`, `--font` | Body font family (`Helvetica`, `Times`, `Courier`) | `Helvetica` |
| `-font-file` | Body font TTF/OTF path | — |
| `-font-size`, `--font-size` | Body font size (points) | `14` |
| `-heading-font`, `--heading-font` | Heading font (empty = same as body) | (same as body) |
| `-heading-font-file` | Heading font TTF/OTF path | — |

### Page Layout

| Flag | Description | Default |
|------|-------------|---------|
| `-page-w`, `--page-w` | Page width in mm | `210` (A4) |
| `-page-h`, `--page-h` | Page height in mm | `297` (A4) |
| `-margin`, `--margin` | All margins in mm | `15` |

### Table

| Flag | Description | Default |
|------|-------------|---------|
| `-table-full-width`, `--table-full-width` | Stretch table to fill available width | `false` (auto-width) |
| `-table-header-font` | Table header font | (same as body) |
| `-table-header-font-file` | Table header font TTF/OTF path | — |
| `-table-header-size` | Table header font size | `12` |

Tables auto-size to fit their content by default. Use `-table-full-width` to stretch across the full width between margins.

### Theme

| Flag | Description | Default |
|------|-------------|---------|
| `-theme` | Theme preset: `light`, `dark`, `github`, `slack` | `light` |

### Colors

All color flags accept hex values: `#333366`, `333366`, or shorthand `#fff`.

| Flag | Description | Default |
|------|-------------|---------|
| `-text-color` | Body text color | `#282828` |
| `-link-color` | Link text color | `#2850B4` |
| `-heading-color` | Heading text color | `#282850` |
| `-table-header-bg` | Table header background | `#323250` |
| `-table-header-fg` | Table header text color | `#C8C8FF` |
| `-table-row-even` | Even table row background | `#F5F5FA` |
| `-table-row-odd` | Odd table row background | `#FFFFFF` |
| `-code-bg` | Code block background | `#F0F0F0` |
| `-code-font` | Code block font | `Courier` |
| `-code-font-file` | Code block font TTF/OTF path | — |
| `-code-font-size` | Code block font size | `11` |
| `-blockquote-line-color` | Blockquote left border | `#6464C8` |
| `-blockquote-text-color` | Blockquote text color | `#646464` |
| `-hr-color` | Horizontal rule color | `#B4B4B4` |

### Examples

```bash
# Dark theme table
echo "| Name | Score |\n|------|-------|\n| Alice | 95 |" | md2img -theme dark -o dark_table.png

# US Letter, high resolution
md2img -o report.png -page-w 215.9 -page-h 279.4 -dpi 300 report.md

# Trim with custom padding (in mm)
echo "| A | B |" | md2img -o padded.png -trim -trim-padding 10

# Times font, large text
md2img -o big.png -font Times -font-size 16 -heading-font Helvetica input.md

# Use a specific Unicode-capable font file
md2img -o unicode.png -font-file /path/to/NotoSans-Regular.ttf input.md

# Full-width table (stretches to fill page)
echo "| A | B |\n|---|---|\n| 1 | 2 |" | md2img -o wide.png -table-full-width
```

## As a library

```go
import md2img "github.com/jmaciasluque/md2img"

// Simple usage
err := md2img.Render("# Hello\n\nWorld.", "output.png")

// With custom config
cfg := md2img.DefaultConfig()
cfg.DPI = 300
cfg.FontFamily = "Times"
cfg.TableHeaderBg = md2img.Color{R: 45, G: 55, B: 72}
cfg.TableHeaderFg = md2img.Color{R: 226, G: 232, B: 240}
cfg.Trim = true  // auto-crop whitespace

err := md2img.RenderWithConfig("# Report\n\n| A | B |\n|---|---|\n| 1 | 2 |", "report.png", cfg)

// Render without touching the filesystem
img, err := md2img.RenderImage("# Preview", cfg)
err = md2img.RenderPNG(w, "# Stream me", cfg)
```

### Color helpers

```go
// Parse hex colors
c, err := md2img.HexToColor("#333366")

// Or construct directly
c := md2img.Color{R: 51, G: 51, B: 102}
```

## Chat and AI agents

Matrix, Slack, and most chat platforms don't render HTML tables — they weren't built for structured data. If you're an AI agent that needs to present a comparison or summary in a conversation, you're stuck with code blocks (ugly, no alignment) or images.

md2img fills that gap: markdown in, styled PNG out, send it as a message attachment.

```bash
# Quick table for a chat message
echo "| Model      | Speed    | Quality |
|------------|----------|---------|
| Qwen3-14B  | 11 tok/s | Good    |
| Gemma3-12B | 13 tok/s | Good    |" | md2img -trim -o /tmp/table.png

# Then send MEDIA:/tmp/table.png in your message
```

Works for longer reports too:

```bash
cat << 'EOF' | md2img -trim -o /tmp/report.png
## Weekly Summary

| Task            | Status  | Hours |
|-----------------|---------|-------|
| Blog post       | Done    | 4     |
| API refactor    | In progress | 6 |
| Deploy staging  | Blocked | 0     |
EOF
```

## Supported Markdown

| Element | Rendering |
|---------|-----------|
| Headers (H1–H6) | Bold, sized proportionally |
| Tables | Auto-sized columns (or full-width), configurable header/row colors |
| Table alignment | Left, center, and right alignment markers |
| Bullet lists | `*` prefix, including nested lists |
| Numbered lists | `1.` `2.` `3.` prefix |
| Task lists | `[x]` and `[ ]` markers |
| Code blocks | Monospace font, configurable background |
| Blockquotes | Configurable left border, italic |
| Horizontal rules | Configurable color and thickness |
| Bold / italic | Supported inline within paragraphs |
| Links | Colored and underlined link text |
| Strikethrough | Horizontal strike through text |
| Inline code | Monospace font at body text size |

## Limitations

- **Font dependent Unicode** — glyphs render when the selected system/custom font supports them. Unsupported glyphs fall back to ASCII placeholders.
- **No inline images** — text-based rendering only.

## Benchmarks

No external dependencies keeps rendering simple. These results are from an Apple M4 at Go 1.26.2:

```
BenchmarkRenderSimple           6.0ms    16.6MB/op
BenchmarkRenderTable           14.5ms    17.1MB/op
BenchmarkRenderComplex         39.5ms    22.2MB/op
BenchmarkRenderDPI100          18.2ms     7.2MB/op
BenchmarkRenderDPI300          68.5ms    47.0MB/op
BenchmarkRenderTrimmed         46.6ms    22.2MB/op
BenchmarkRenderInline           6.6ms    16.7MB/op
BenchmarkRenderFullWidthTable  20.9ms    19.2MB/op
```

## Project Structure

```
md2img/
├── cmd/md2img/     # CLI entry point
│   ├── main.go
│   └── main_test.go
├── render.go       # Direct-to-image rendering engine + Config (library API)
├── fonts.go        # TTF font loading with system fallback
├── trim.go         # Auto-crop whitespace
├── theme.go        # Built-in theme presets
├── sanitize.go     # Unicode → ASCII mapping
├── sanitize_test.go
├── render_test.go
├── bench_test.go
├── Makefile
├── .goreleaser.yaml
├── .github/workflows/
│   ├── ci.yml      # Build + test on macOS & Ubuntu
│   ├── bench.yml   # Benchmark suite with benchstat
│   └── release.yml # Multi-platform release builds
└── README.md
```

## How It Works

1. **Parse** — [goldmark](https://github.com/yuin/goldmark) parses Markdown into an AST (with GFM table support)
2. **Render** — Direct rendering to `image.RGBA` using [golang.org/x/image](https://pkg.go.dev/golang.org/x/image) for TTF font rendering
3. **Output** — PNG encoding with optional auto-crop

The binary is a few MB depending on OS and architecture. No external dependencies — fonts are loaded from your system.

## Development

```bash
# Run tests
make test

# Run go vet
make vet

# Build
make build

# Install locally
make install

# Run benchmarks
make bench

# Validate release configuration locally
make release-snapshot

# Compare against a previous run
go test -bench=. -benchmem -count=5 ./... | tee new.txt
benchstat old.txt new.txt
```

## License

MIT
