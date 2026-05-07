---
name: md2img
description: Convert markdown to styled PNG images via a compiled Go binary. Renders tables, headers, code blocks, lists, blockquotes to styled images for Matrix, Slack, Discord, and any platform that doesn't render HTML.
category: creative
tags: [markdown, png, image, rendering, matrix, slack, discord, telegram]
---

# md2img — Markdown to PNG

Renders markdown to styled PNG images using a pure Go binary (`~/bin/md2img`). No external runtime dependencies.

## Pipeline

```
markdown → goldmark (parser) → canvas (golang.org/x/image + image/draw) → PNG
```

**No Ghostscript, no PDF intermediate.** The renderer draws directly to an `image.NRGBA` canvas using `golang.org/x/image/font` for TTF text and `image/draw` for shapes.

**Fonts**: Loads system TTF fonts via `findFirst()` — checks macOS paths (`/Library/Fonts/`, `~/Library/Fonts/`) then Ubuntu paths (`/usr/share/fonts/truetype/`). Falls back to `golang.org/x/image/font/basicfont` (bitmap, fixed-size — smaller output on systems without TTF fonts).

## Install

```bash
# Homebrew
brew install jmaciasluque/tap/md2img

# From source
cd ~/src/md2img && make build && make install
```

## Usage

```bash
# From file (RELIABLE — always works)
md2img -o output.png input.md

# From stdin
echo "## Hello" | md2img -o output.png

# With customization flags
md2img -o dark.png -heading-color "#006600" -table-header-bg "#2d3748" -dpi 300 input.md

# Auto-crop whitespace (tight around content)
echo "| A | B |" | md2img -o tight.png -trim

# Trim with custom padding (mm)
md2img -o padded.png -trim -trim-padding 10 input.md
```

## CLI Flags

Key groups:

- **Output**: `-o`, `-trim`, `-trim-padding`, `-dpi`, `-version`
- **Font**: `-font`, `-font-size`, `-heading-font`
- **Page**: `-page-w`, `-page-h`, `-margin`
- **Colors** (all accept hex like `#333366`): `-text-color`, `-heading-color`, `-table-header-bg`, `-table-header-fg`, `-table-row-even`, `-table-row-odd`, `-code-bg`, `-blockquote-line-color`, `-blockquote-text-color`, `-hr-color`
- **Table**: `-table-header-font`, `-table-header-size`, `-table-full-width` (opt-in to stretch tables across full width; default is auto-width fitting content)
- **Code**: `-code-font`, `-code-font-size`

**Note**: Input is a positional arg (not `-input`). Output is `-o` (not `-output`). Stdin is used when no positional arg is given.

## Library API

```go
import md2img "github.com/jmaciasluque/md2img"

// Simple
err := md2img.Render("# Hello\n\nWorld.", "output.png")

// With config
cfg := md2img.DefaultConfig()
cfg.DPI = 300
cfg.TableHeaderBg = md2img.Color{R: 45, G: 55, B: 72}
err := md2img.RenderWithConfig(input, output, cfg)
```

- `Config` struct holds all options, `DefaultConfig()` returns sensible defaults
- `HexToColor("#333366")` parses hex strings to `Color`
- `Color{R, G, B int}` — uses `int` (not `uint8`)

## Supported Markdown

- **Headers** (H1–H6) — bold, sized proportionally
- **Tables** — auto-width by default (columns fit content), configurable header/row colors, cell borders, zebra striping. Use `-table-full-width` to stretch across page width.
- **Bullet & numbered lists**
- **Code blocks** — monospace font, configurable background
- **Blockquotes** — configurable left border, italic
- **Horizontal rules** — configurable color and thickness
- **Bold**, **italic**, and `inline code` — full inline formatting support
- **Inline code** — monospace font at body size for consistent baseline alignment

## Limitations

- **No inline images** — only text-based rendering.
- **No nested lists** — flat lists only.
- **Font availability varies by platform** — macOS has good TTF coverage; Ubuntu needs `fonts-liberation` package for proper rendering.

## Source & Repo

**GitHub:** https://github.com/jmaciasluque/md2img
**Tap:** https://github.com/jmaciasluque/homebrew-tap
**Binary:** `~/bin/md2img`
