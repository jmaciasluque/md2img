# md2img

Convert Markdown to styled PNG images. No browser, no Python — just a Go binary and Ghostscript.

```
markdown → goldmark (parser) → gofpdf (PDF) → Ghostscript (PNG)
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

**Dependencies:** [Ghostscript](https://www.ghostscript.com/) (`gs`) must be installed.

```bash
# macOS
brew install ghostscript

# Ubuntu/Debian
sudo apt install ghostscript

# Arch
sudo pacman -S ghostscript
```

## Usage

```bash
# From stdin
echo "| Name | Score |\n|------|-------|\n| Alice | 95 |" | md2img -o scores.png

# From file
md2img -o output.png input.md

# Default output: /tmp/md2img_output.png
echo "# Hello" | md2img
```

## CLI Flags

### Output

| Flag | Description | Default |
|------|-------------|---------|
| `-o`, `--output` | Output file path | `/tmp/md2img_output.png` |
| `-pdf`, `--pdf` | Output PDF directly (no Ghostscript needed) | `false` |
| `-trim`, `--trim` | Auto-crop whitespace from PNG output | `false` |
| `-dpi`, `--dpi` | PNG resolution (DPI) | `200` |
| `-version`, `--version` | Print version | — |

### Font

| Flag | Description | Default |
|------|-------------|---------|
| `-font`, `--font` | Body font family (`Helvetica`, `Times`, `Courier`) | `Helvetica` |
| `-font-size`, `--font-size` | Body font size (points) | `11` |
| `-heading-font`, `--heading-font` | Heading font (empty = same as body) | (same as body) |

### Page Layout

| Flag | Description | Default |
|------|-------------|---------|
| `-page-w`, `--page-w` | Page width in mm | `210` (A4) |
| `-page-h`, `--page-h` | Page height in mm | `297` (A4) |
| `-margin`, `--margin` | All margins in mm | `15` |

### Colors

All color flags accept hex values: `#333366`, `333366`, or shorthand `#fff`.

| Flag | Description | Default |
|------|-------------|---------|
| `-text-color` | Body text color | `#282828` |
| `-heading-color` | Heading text color | `#282850` |
| `-table-header-bg` | Table header background | `#323250` |
| `-table-header-fg` | Table header text color | `#C8C8FF` |
| `-table-header-font` | Table header font | (same as body) |
| `-table-header-size` | Table header font size | `10` |
| `-table-row-even` | Even table row background | `#F5F5FA` |
| `-table-row-odd` | Odd table row background | `#FFFFFF` |
| `-code-bg` | Code block background | `#F0F0F0` |
| `-code-font` | Code block font | `Courier` |
| `-code-font-size` | Code block font size | `9` |
| `-blockquote-line-color` | Blockquote left border | `#6464C8` |
| `-blockquote-text-color` | Blockquote text color | `#646464` |
| `-hr-color` | Horizontal rule color | `#B4B4B4` |

### Examples

```bash
# Dark theme table
echo "| Name | Score |\n|------|-------|\n| Alice | 95 |" | md2img \
  -o dark_table.png \
  -text-color "#E2E8F0" \
  -table-header-bg "#2D3748" \
  -table-header-fg "#E2E8F0" \
  -table-row-even "#1A202C" \
  -table-row-odd "#2D3748"

# US Letter, high resolution
md2img -o report.png -page-w 215.9 -page-h 279.4 -dpi 300 report.md

# Direct PDF output (no Ghostscript needed)
echo "# Title" | md2img -o output.pdf -pdf

# Trim whitespace (tight crop around content)
echo "| A | B |" | md2img -o tight.png -trim

# Times font, large text
md2img -o big.png -font Times -font-size 16 -heading-font Helvetica input.md
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
cfg.AsPDF = true  // output PDF directly
cfg.Trim = true   // auto-crop whitespace

err := md2img.RenderWithConfig("# Report\n\n| A | B |\n|---|---|\n| 1 | 2 |", "report.pdf", cfg)
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
| Gemma3-12B | 13 tok/s | Good    |" | md2img -o /tmp/table.png

# Then send MEDIA:/tmp/table.png in your message
```

Works for longer reports too:

```bash
cat << 'EOF' | md2img -o /tmp/report.png
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
| Tables | Configurable header/row colors, cell borders |
| Bullet lists | `*` prefix |
| Numbered lists | `1.` `2.` `3.` prefix |
| Code blocks | Monospace font, configurable background |
| Blockquotes | Configurable left border, italic |
| Horizontal rules | Configurable color and thickness |
| Bold / italic | Supported via markdown syntax |

## Limitations

- **ASCII only** — Unicode characters (emojis, em dashes, special symbols) are mapped to ASCII equivalents. Full Unicode support requires embedding a TTF font.
- **No inline images** — text-based rendering only.
- **No nested lists** — flat lists only.

## Examples

### Table

```bash
cat << 'EOF' | md2img -o comparison.png
## STACKIT vs Scaleway

| Feature    | STACKIT      | Scaleway    |
|------------|--------------|-------------|
| Free Tier  | No           | Yes         |
| Kubernetes | SKE          | Kapsule     |
| Best For   | Government   | Everyone    |
EOF
```

### Code Block

```bash
echo '```go
fmt.Println("hello world")
```' | md2img -o code.png
```

### Full Document

```bash
md2img -o report.png report.md
```

## Project Structure

```
md2img/
├── cmd/md2img/     # CLI entry point
│   ├── main.go
│   └── main_test.go
├── render.go       # PDF rendering engine + Config (library API)
├── sanitize.go     # Unicode → ASCII mapping
├── sanitize_test.go
├── render_test.go
├── Makefile
├── .github/workflows/
│   ├── ci.yml      # Build + test + flag tests on macOS & Ubuntu
│   └── release.yml # Multi-platform release builds
└── README.md
```

## How It Works

1. **Parse** — [goldmark](https://github.com/yuin/goldmark) parses Markdown into an AST (with GFM table support)
2. **Render** — [gofpdf](https://github.com/jung-kurt/gofpdf) draws the AST onto a PDF page with styled fonts and colors
3. **Convert** — Ghostscript rasterizes the PDF to a configurable DPI PNG (or output PDF directly with `-pdf`)

The binary is ~5MB. Ghostscript is the only external dependency (not needed for `-pdf` mode).

## Development

```bash
# Run tests
make test

# Build
make build

# Install locally
make install

# Run benchmarks
go test -bench=. -benchmem -count=3 ./...

# Compare against a previous run
go test -bench=. -benchmem -count=5 ./... | tee new.txt
benchstat old.txt new.txt
```

## License

MIT
