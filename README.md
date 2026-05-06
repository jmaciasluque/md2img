# md2img

Convert Markdown to styled PNG images. No browser, no Python — just a Go binary and Ghostscript.

```
markdown → goldmark (parser) → gofpdf (PDF) → Ghostscript (PNG)
```

## Install

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

## As a library

```go
import md2img "github.com/jmaciasluque/md2img"

err := md2img.Render("# Hello\n\nWorld.", "output.png")
```

## Supported Markdown

| Element | Rendering |
|---------|-----------|
| Headers (H1–H6) | Bold, blue, sized proportionally |
| Tables | Navy header, zebra-striped rows, cell borders |
| Bullet lists | `*` prefix |
| Numbered lists | `1.` `2.` `3.` prefix |
| Code blocks | Monospace font, gray background |
| Blockquotes | Blue left border, italic |
| Horizontal rules | Thin gray line |
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
├── render.go       # PDF rendering engine (library API)
├── sanitize.go     # Unicode → ASCII mapping
├── sanitize_test.go
├── render_test.go
├── Makefile
├── .github/workflows/
│   ├── ci.yml      # Build + test on push/PR
│   └── release.yml # Multi-platform release builds
└── README.md
```

## How It Works

1. **Parse** — [goldmark](https://github.com/yuin/goldmark) parses Markdown into an AST (with GFM table support)
2. **Render** — [gofpdf](https://github.com/jung-kurt/gofpdf) draws the AST onto a PDF page with styled fonts and colors
3. **Convert** — Ghostscript rasterizes the PDF to a 200 DPI PNG

The binary is ~5MB. Ghostscript is the only external dependency.

## Development

```bash
# Run tests
make test

# Build
make build

# Install locally
make install
```

## License

MIT
