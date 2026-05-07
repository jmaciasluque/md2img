---
name: md2img
description: Convert markdown to styled PNG images via a compiled Go binary. Renders tables, headers, code blocks, lists, blockquotes to styled images for Matrix, Slack, Discord, and any platform that doesn't render HTML.
category: creative
tags: [markdown, png, image, rendering, matrix, slack, discord, telegram]
---

# md2img — Markdown to PNG

Renders markdown to styled PNG images using a pure Go binary (`md2img`). No external runtime dependencies.

## Pipeline

```
markdown → goldmark (parser) → canvas (golang.org/x/image + image/draw) → PNG
```

**No Ghostscript, no PDF intermediate.** The renderer draws directly to a Go image canvas using `golang.org/x/image/font` for TTF text and `image/draw` for shapes.

**Fonts**: Loads system TTF fonts via `findFirst()` and supports explicit TTF/OTF paths with `-font-file`, `-heading-font-file`, `-table-header-font-file`, and `-code-font-file`. Falls back to `golang.org/x/image/font/basicfont` if no configured font can be loaded.

## Install

```bash
# Homebrew
brew install jmaciasluque/tap/md2img

# From source
cd ~/src/md2img && make build && make install
```

If `md2img` is not on PATH, build it from the repo and use `./md2img`.

## Usage

```bash
# From file (RELIABLE — always works)
md2img -o output.png input.md

# From stdin
echo "## Hello" | md2img -o output.png

# Explicit stdin marker
echo "## Hello" | md2img -o output.png -

# With customization flags
md2img -o dark.png -theme dark -dpi 300 input.md

# With a specific Unicode-capable font
md2img -o unicode.png -font-file /path/to/NotoSans-Regular.ttf input.md

# Auto-crop whitespace (tight around content)
echo "| A | B |" | md2img -o tight.png -trim

# Trim with custom padding (mm)
md2img -o padded.png -trim -trim-padding 10 input.md
```

## CLI Flags

Key groups:

- **Output**: `-o`, `-trim`, `-trim-padding`, `-dpi`, `-version`
- **Theme**: `-theme` (`light`, `dark`, `github`, `slack`)
- **Font**: `-font`, `-font-file`, `-font-size`, `-heading-font`, `-heading-font-file`
- **Page**: `-page-w`, `-page-h`, `-margin`
- **Colors** (all accept hex like `#333366`): `-text-color`, `-link-color`, `-heading-color`, `-table-header-bg`, `-table-header-fg`, `-table-row-even`, `-table-row-odd`, `-code-bg`, `-blockquote-line-color`, `-blockquote-text-color`, `-hr-color`
- **Table**: `-table-header-font`, `-table-header-font-file`, `-table-header-size`, `-table-full-width` (opt-in to stretch tables across full width; default is auto-width fitting content)
- **Code**: `-code-font`, `-code-font-file`, `-code-font-size`

**Note**: Input is a positional arg (not `-input`). Output is `-o` or `--output`. Stdin is used when no positional arg is given, or when the positional arg is `-`.

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
- **Table alignment** — left, center, and right alignment markers.
- **Bullet & numbered lists** — including nested lists.
- **Task lists** — `[x]` and `[ ]` markers.
- **Code blocks** — monospace font, configurable background
- **Blockquotes** — configurable left border, italic
- **Horizontal rules** — configurable color and thickness
- **Bold**, **italic**, and `inline code` — full inline formatting support
- **Links** — colored and underlined link text
- **Strikethrough** — horizontal strike through text
- **Inline code** — monospace font at body size for consistent baseline alignment

## Limitations

- **No inline images** — only text-based rendering.
- **Font availability varies by platform** — pass `-font-file` for predictable Unicode rendering across machines.

## Agent Workflow

1. Check `md2img -version` or build with `make build`.
2. Render markdown with `-trim` for chat attachments unless full-page output is needed.
3. Verify the output path exists and is non-empty before sending or attaching it.
4. For wide tables or long prose, inspect the PNG to confirm the wrapped layout is readable.

## Source & Repo

**GitHub:** https://github.com/jmaciasluque/md2img
**Tap:** https://github.com/jmaciasluque/homebrew-tap
**Binary:** `~/bin/md2img`
