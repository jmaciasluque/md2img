package md2img

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// Version is set at build time via ldflags.
var Version = "dev"

var parser = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
).Parser()

// Color represents an RGB color with values 0–255.
type Color struct {
	R, G, B int
}

// HexToColor parses a hex color string like "#333366" or "333366" into a Color.
func HexToColor(s string) (Color, error) {
	s = strings.TrimPrefix(s, "#")
	if len(s) == 3 {
		s = string([]byte{s[0], s[0], s[1], s[1], s[2], s[2]})
	}
	if len(s) != 6 {
		return Color{}, fmt.Errorf("invalid hex color: %s", s)
	}
	r, err := hexByte(s[0:2])
	if err != nil {
		return Color{}, fmt.Errorf("invalid hex color: %s", s)
	}
	g, err := hexByte(s[2:4])
	if err != nil {
		return Color{}, fmt.Errorf("invalid hex color: %s", s)
	}
	b, err := hexByte(s[4:6])
	if err != nil {
		return Color{}, fmt.Errorf("invalid hex color: %s", s)
	}
	return Color{R: int(r), G: int(g), B: int(b)}, nil
}

func hexByte(s string) (byte, error) {
	var v byte
	for _, c := range s {
		v *= 16
		switch {
		case c >= '0' && c <= '9':
			v += byte(c - '0')
		case c >= 'a' && c <= 'f':
			v += byte(c-'a') + 10
		case c >= 'A' && c <= 'F':
			v += byte(c-'A') + 10
		default:
			return 0, fmt.Errorf("invalid hex digit: %c", c)
		}
	}
	return v, nil
}

func (c Color) toRGBA() color.RGBA {
	return color.RGBA{R: byte(c.R), G: byte(c.G), B: byte(c.B), A: 255}
}

// Config holds all customizable rendering options.
type Config struct {
	// Font
	FontFamily string  // "Helvetica", "Times", or "Courier"
	FontSize   float64 // Body text size in points

	// Page
	PageWidth    float64 // Page width in mm (default: 210 = A4)
	PageHeight   float64 // Page height in mm (default: 297 = A4)
	MarginTop    float64 // Top margin in mm
	MarginLeft   float64 // Left margin in mm
	MarginRight  float64 // Right margin in mm
	MarginBottom float64 // Bottom margin in mm

	// Text colors
	TextColor Color // Default body text color

	// Heading colors and sizes
	HeadingColor   Color
	HeadingSizes   [6]float64
	HeadingFont    string // Heading font family override
	HeadingBold    bool   // Render headings in bold (default: true)

	// Table
	TableHeaderBg    Color
	TableHeaderFg    Color
	TableHeaderFont  string
	TableHeaderSize  float64
	TableCellHeight  float64
	TableRowEven     Color
	TableRowOdd      Color

	// Code block
	CodeBg         Color
	CodeFont       string
	CodeFontSize   float64
	CodeLineHeight float64

	// Blockquote
	BlockquoteLineColor Color
	BlockquoteTextColor Color
	BlockquoteFont      string

	// Horizontal rule
	HRColor     Color
	HRLineWidth float64

	// Output
	DPI         int
	Trim        bool
	TrimPadding float64
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		FontFamily: "Helvetica",
		FontSize:   11,

		PageWidth:    210,
		PageHeight:   297,
		MarginTop:    15,
		MarginLeft:   15,
		MarginRight:  15,
		MarginBottom: 15,

		TextColor: Color{40, 40, 40},

		HeadingColor: Color{40, 40, 80},
		HeadingSizes: [6]float64{22, 18, 14, 12, 11, 11},
		HeadingFont:  "",
		HeadingBold:  true,

		TableHeaderBg:   Color{50, 50, 80},
		TableHeaderFg:   Color{200, 200, 255},
		TableHeaderFont: "",
		TableHeaderSize: 10,
		TableCellHeight: 10,
		TableRowEven:    Color{245, 245, 250},
		TableRowOdd:     Color{255, 255, 255},

		CodeBg:         Color{240, 240, 240},
		CodeFont:       "Courier",
		CodeFontSize:   9,
		CodeLineHeight: 5.5,

		BlockquoteLineColor: Color{100, 100, 200},
		BlockquoteTextColor: Color{100, 100, 100},
		BlockquoteFont:      "",

		HRColor:     Color{180, 180, 180},
		HRLineWidth: 0.3,

		DPI:         200,
		Trim:        false,
		TrimPadding: 5,
	}
}

// canvas holds rendering state.
type canvas struct {
	img    *image.RGBA
	width  int // canvas width in pixels
	x, y   int // current position
	margin int // margin in pixels
	fonts  fontSet
	cfg    Config
}

// mmToPx converts millimeters to pixels at the given DPI.
func mmToPx(mm float64, dpi int) int {
	return int(mm * float64(dpi) / 25.4)
}

func newCanvas(cfg Config) *canvas {
	// Use DPI from config, or infer from PageWidth for reasonable pixel density.
	dpi := cfg.DPI
	if dpi <= 0 {
		dpi = 200
	}
	w := mmToPx(cfg.PageWidth, dpi)
	h := mmToPx(cfg.PageHeight, dpi)
	margin := mmToPx(cfg.MarginLeft, dpi)

	img := image.NewRGBA(image.Rect(0, 0, w, h*4)) // 4x height for growth
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	fonts := loadFonts(cfg.FontFamily, cfg.CodeFont, cfg.FontSize, cfg.CodeFontSize)

	return &canvas{
		img:    img,
		width:  w,
		margin: margin,
		fonts:  fonts,
		cfg:    cfg,
	}
}

// ensureHeight grows the canvas if needed.
func (c *canvas) ensureHeight(needed int) {
	if c.y+needed < c.img.Bounds().Dy() {
		return
	}
	newH := c.img.Bounds().Dy() * 2
	newImg := image.NewRGBA(image.Rect(0, 0, c.width, newH))
	draw.Draw(newImg, newImg.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
	draw.Draw(newImg, c.img.Bounds(), c.img, image.Point{}, draw.Src)
	c.img = newImg
}

// drawString draws text at the current position using the given face and color.
func (c *canvas) drawString(s string, face font.Face, col color.RGBA) {
	d := &font.Drawer{
		Dst:  c.img,
		Face: face,
		Src:  image.NewUniform(col),
		Dot:  fixed.P(c.x, c.y+faceHeight(face)-2),
	}
	d.DrawString(s)
}

// drawLine draws a horizontal line.
func (c *canvas) drawHorizontalLine(x0, x1, y int, col color.RGBA, width float64) {
	lw := int(width)
	if lw < 1 {
		lw = 1
	}
	for i := 0; i < lw; i++ {
		for x := x0; x < x1; x++ {
			if x >= 0 && x < c.img.Bounds().Dx() && y+i >= 0 && y+i < c.img.Bounds().Dy() {
				c.img.Set(x, y+i, col)
			}
		}
	}
}

// drawVerticalLine draws a vertical line.
func (c *canvas) drawVerticalLine(x, y0, y1 int, col color.RGBA) {
	for y := y0; y < y1; y++ {
		if x >= 0 && x < c.img.Bounds().Dx() && y >= 0 && y < c.img.Bounds().Dy() {
			c.img.Set(x, y, col)
		}
	}
}

// drawRect fills a rectangle.
func (c *canvas) drawRect(r image.Rectangle, col color.RGBA) {
	draw.Draw(c.img, r, &image.Uniform{col}, image.Point{}, draw.Src)
}

// fillBackground fills the entire canvas with white.
func (c *canvas) fillBackground() {
	draw.Draw(c.img, c.img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
}

// --- Element renderers ---

func (c *canvas) renderHeading(n ast.Node, src []byte) {
	h := n.(*ast.Heading)
	text := extractText(n, src)
	idx := h.Level - 1
	if idx < 0 {
		idx = 0
	}
	if idx > 5 {
		idx = 5
	}
	size := c.cfg.HeadingSizes[idx]

	fontFamily := c.cfg.HeadingFont
	if fontFamily == "" {
		fontFamily = c.cfg.FontFamily
	}
	family := systemFonts[fontFamily]
	if family.regular == "" {
		family = systemFonts["Helvetica"]
	}

	face := loadFace(family.regular, size)
	if c.cfg.HeadingBold {
		face = loadFace(family.bold, size)
	}

	lh := faceHeight(face)
	c.y += lh / 2
	c.drawString(text, face, c.cfg.HeadingColor.toRGBA())
	c.y += lh/2 + 10
}

func (c *canvas) renderParagraph(n ast.Node, src []byte) {
	text := extractText(n, src)
	face := c.fonts.regular
	lh := faceHeight(face)
	c.drawString(text, face, c.cfg.TextColor.toRGBA())
	c.y += lh + 6
}

func (c *canvas) renderCodeBlock(n ast.Node, src []byte) {
	var lines []string
	switch block := n.(type) {
	case *ast.FencedCodeBlock:
		for i := 0; i < block.Lines().Len(); i++ {
			seg := block.Lines().At(i)
			lines = append(lines, strings.TrimRight(string(seg.Value(src)), "\n"))
		}
	case *ast.CodeBlock:
		for i := 0; i < block.Lines().Len(); i++ {
			seg := block.Lines().At(i)
			lines = append(lines, strings.TrimRight(string(seg.Value(src)), "\n"))
		}
	default:
		lines = strings.Split(extractText(n, src), "\n")
	}

	lh := mmToPx(c.cfg.CodeLineHeight, c.cfg.DPI)
	padding := mmToPx(4, c.cfg.DPI)
	blockH := len(lines)*lh + padding*2
	lineH := faceHeight(c.fonts.code)

	// Background
	bgRect := image.Rect(c.margin-4, c.y, c.width-c.margin+4, c.y+blockH)
	c.drawRect(bgRect, c.cfg.CodeBg.toRGBA())

	// Text
	c.x = c.margin + padding
	c.y += padding
	for _, line := range lines {
		c.drawString(line, c.fonts.code, c.cfg.TextColor.toRGBA())
		c.y += lineH
	}
	c.x = c.margin
	c.y += padding + 6
}

func (c *canvas) renderTable(n ast.Node, src []byte) {
	var rows [][]string
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		kind := child.Kind().String()
		if kind == "TableHeader" || kind == "TableRow" {
			var cells []string
			for cell := child.FirstChild(); cell != nil; cell = cell.NextSibling() {
				if cell.Kind().String() == "TableCell" {
					cells = append(cells, strings.TrimSpace(extractText(cell, src)))
				}
			}
			if len(cells) > 0 {
				rows = append(rows, cells)
			}
		}
	}
	if len(rows) == 0 {
		return
	}

	numCols := len(rows[0])
	tableW := c.width - 2*c.margin
	colW := tableW / numCols
	rowH := mmToPx(c.cfg.TableCellHeight, c.cfg.DPI)
	if rowH < 20 {
		rowH = 20
	}

	// Load header font
	headerFamily := c.cfg.TableHeaderFont
	if headerFamily == "" {
		headerFamily = c.cfg.FontFamily
	}
	family := systemFonts[headerFamily]
	if family.regular == "" {
		family = systemFonts["Helvetica"]
	}
	headerFace := loadFace(family.bold, c.cfg.TableHeaderSize)

	for ri, row := range rows {
		c.ensureHeight(rowH + 10)
		for ci := 0; ci < numCols && ci < len(row); ci++ {
			x := c.margin + ci*colW
			y := c.y

			// Background
			var bg color.RGBA
			if ri == 0 {
				bg = c.cfg.TableHeaderBg.toRGBA()
			} else if ri%2 == 0 {
				bg = c.cfg.TableRowEven.toRGBA()
			} else {
				bg = c.cfg.TableRowOdd.toRGBA()
			}
			c.drawRect(image.Rect(x, y, x+colW, y+rowH), bg)

			// Borders
			borderCol := color.RGBA{180, 180, 180, 255}
			c.drawHorizontalLine(x, x+colW, y, borderCol, 0.5)
			c.drawHorizontalLine(x, x+colW, y+rowH, borderCol, 0.5)
			c.drawVerticalLine(x, y, y+rowH, borderCol)
			c.drawVerticalLine(x+colW, y, y+rowH, borderCol)

			// Text
			fg := c.cfg.TextColor.toRGBA()
			face := c.fonts.regular
			if ri == 0 {
				fg = c.cfg.TableHeaderFg.toRGBA()
				face = headerFace
			}
			d := &font.Drawer{
				Dst:  c.img,
				Face: face,
				Src:  image.NewUniform(fg),
				Dot:  fixed.P(x+8, y+rowH-int(float64(rowH)*0.3)),
			}
			d.DrawString(row[ci])
		}
		c.y += rowH
	}
	c.y += 10
}

func (c *canvas) renderList(n ast.Node, src []byte) {
	l := n.(*ast.List)
	face := c.fonts.regular
	lh := faceHeight(face)
	c.y += 2
	i := 1
	for item := l.FirstChild(); item != nil; item = item.NextSibling() {
		text := extractText(item, src)
		bullet := "•  "
		if l.IsOrdered() {
			bullet = fmt.Sprintf("%d.  ", i)
			i++
		}
		c.drawString(bullet+text, face, c.cfg.TextColor.toRGBA())
		c.y += lh + 3
	}
	c.y += 6
}

func (c *canvas) renderBlockquote(n ast.Node, src []byte) {
	lc := c.cfg.BlockquoteLineColor.toRGBA()
	c.drawVerticalLine(c.margin+2, c.y, c.y+24, lc)

	saved := c.x
	c.x = c.margin + 14
	face := c.fonts.italic
	if face == nil {
		face = c.fonts.regular
	}
	text := extractText(n, src)
	c.drawString(text, face, c.cfg.BlockquoteTextColor.toRGBA())
	c.x = saved
	c.y += faceHeight(face) + 4
}

func (c *canvas) renderHR() {
	c.y += 8
	y := c.y
	hc := c.cfg.HRColor.toRGBA()
	lw := int(c.cfg.HRLineWidth * float64(c.cfg.DPI) / 25.4)
	if lw < 1 {
		lw = 1
	}
	c.drawHorizontalLine(c.margin, c.width-c.margin, y, hc, float64(lw))
	c.y += 12
}

// --- Top-level render ---

func (c *canvas) renderNodes(n ast.Node, src []byte) {
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		c.renderNode(child, src)
	}
}

func (c *canvas) renderNode(n ast.Node, src []byte) {
	switch n.Kind().String() {
	case "Heading":
		c.renderHeading(n, src)
	case "Paragraph":
		c.renderParagraph(n, src)
	case "CodeBlock", "FencedCodeBlock":
		c.renderCodeBlock(n, src)
	case "Table":
		c.renderTable(n, src)
	case "List":
		c.renderList(n, src)
	case "Blockquote":
		c.renderBlockquote(n, src)
	case "ThematicBreak":
		c.renderHR()
	}
}

// extractText collects all text content from a node.
func extractText(n ast.Node, src []byte) string {
	var buf bytes.Buffer
	ast.Walk(n, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			switch t := n.(type) {
			case *ast.Text:
				buf.Write(t.Segment.Value(src))
			case *ast.String:
				buf.Write(t.Value)
			}
		}
		return ast.WalkContinue, nil
	})
	return sanitize(buf.String())
}

// Render converts markdown input to a PNG file at the given output path.
func Render(input, output string) error {
	return RenderWithConfig(input, output, DefaultConfig())
}

// RenderWithConfig converts markdown input to a PNG file using the given configuration.
func RenderWithConfig(input, output string, cfg Config) error {
	c := newCanvas(cfg)
	c.fillBackground()

	src := []byte(input)
	reader := text.NewReader(src)
	doc := parser.Parse(reader)
	c.renderNodes(doc, src)

	// Crop to content height.
	bounds := c.img.Bounds()
	contentH := c.y + 20 // some bottom padding
	if contentH > bounds.Dy() {
		contentH = bounds.Dy()
	}

	// Auto-crop to content bounds (left/right).
	left, right := bounds.Dx(), 0
	for y := 0; y < contentH; y++ {
		for x := 0; x < bounds.Dx(); x++ {
			r, g, b, a := c.img.At(x, y).RGBA()
			if a < 128 {
				continue
			}
			if r < 0xF000 || g < 0xF000 || b < 0xF000 {
				if x < left {
					left = x
				}
				if x > right {
					right = x
				}
			}
		}
	}

	if left > right {
		// Empty — write a 1x1 white pixel.
		return writePNG(image.NewRGBA(image.Rect(0, 0, 1, 1)), output)
	}

	cropped := c.img.SubImage(image.Rect(0, 0, right+1, contentH))

	if cfg.Trim {
		return writeTrimmedPNG(cropped, output, cfg.DPI, cfg.TrimPadding)
	}

	// Convert back to RGBA for PNG encoding.
	b := cropped.Bounds()
	rgba := image.NewRGBA(b)
	draw.Draw(rgba, b, cropped, b.Min, draw.Src)
	return writePNG(rgba, output)
}

func writePNG(img *image.RGBA, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
