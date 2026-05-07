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
	HeadingColor Color
	HeadingSizes [6]float64
	HeadingFont  string // Heading font family override
	HeadingBold  bool   // Render headings in bold (default: true)

	// Table
	TableHeaderBg   Color
	TableHeaderFg   Color
	TableHeaderFont string
	TableHeaderSize float64
	TableCellHeight float64
	TableAutoWidth  bool // size columns to fit content
	TableRowEven    Color
	TableRowOdd     Color

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
		FontSize:   14,

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
		TableHeaderSize: 12,
		TableAutoWidth:  true,
		TableCellHeight: 5,
		TableRowEven:    Color{245, 245, 250},
		TableRowOdd:     Color{255, 255, 255},

		CodeBg:         Color{240, 240, 240},
		CodeFont:       "Courier",
		CodeFontSize:   11,
		CodeLineHeight: 6.5,

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
	img         *image.RGBA
	width       int // canvas width in pixels
	x, y        int // current position
	margin      int // left margin in pixels
	marginRight int // right margin in pixels
	marginTop   int // top margin in pixels
	fonts       fontSet
	cfg         Config
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
	marginRight := mmToPx(cfg.MarginRight, dpi)
	mt := mmToPx(cfg.MarginTop, dpi)

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	fonts := loadFonts(cfg.FontFamily, cfg.CodeFont, cfg.FontSize, cfg.CodeFontSize)

	return &canvas{
		img:         img,
		width:       w,
		x:           margin,
		y:           mt,
		margin:      margin,
		marginRight: marginRight,
		marginTop:   mt,
		fonts:       fonts,
		cfg:         cfg,
	}
}

func (c *canvas) contentRight() int {
	return c.width - c.marginRight
}

func (c *canvas) contentWidth() int {
	return c.contentRight() - c.margin
}

// ensureHeight grows the canvas if needed.
func (c *canvas) ensureHeight(needed int) {
	if c.y+needed < c.img.Bounds().Dy() {
		return
	}
	newH := c.img.Bounds().Dy() * 2
	for c.y+needed >= newH {
		newH *= 2
	}
	newImg := image.NewRGBA(image.Rect(0, 0, c.width, newH))
	draw.Draw(newImg, newImg.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
	draw.Draw(newImg, c.img.Bounds(), c.img, image.Point{}, draw.Src)
	c.img = newImg
}

// drawString draws text at the current position using the given face and color.
// It advances c.x past the drawn text. The y position uses baseFaceH for consistent
// baseline when mixing different font sizes (e.g. inline code).
func (c *canvas) drawString(s string, face font.Face, col color.RGBA) {
	c.drawStringAt(s, face, col, 0)
}

// drawStringAt draws text with a y-offset from the current c.y, using baseFaceH for
// the baseline calculation so all inline elements sit on the same line.
func (c *canvas) drawStringAt(s string, face font.Face, col color.RGBA, yOff int) {
	d := &font.Drawer{
		Dst:  c.img,
		Face: face,
		Src:  image.NewUniform(col),
		Dot:  fixed.P(c.x, c.y+faceHeight(face)-2+yOff),
	}
	d.DrawString(s)
	c.x = int(d.Dot.X >> 6)
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
	c.ensureHeight(lh*2 + 16)
	c.y += lh / 2
	c.drawString(text, face, c.cfg.HeadingColor.toRGBA())
	c.x = c.margin
	c.y += lh/2 + 16
}

func (c *canvas) renderParagraph(n ast.Node, src []byte) {
	face := c.fonts.regular
	lh := faceHeight(face)
	segments := c.inlineSegments(n, src, face, c.cfg.TextColor.toRGBA())
	c.renderSegmentsWrapped(segments, lh+6)
	c.x = c.margin
	c.y += lh + 12
}

type inlineSegment struct {
	text string
	face font.Face
	col  color.RGBA
}

func (c *canvas) inlineSegments(n ast.Node, src []byte, baseFace font.Face, baseCol color.RGBA) []inlineSegment {
	var segments []inlineSegment
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		switch child.Kind().String() {
		case "Text":
			t := child.(*ast.Text)
			segments = append(segments, inlineSegment{
				text: sanitize(string(t.Segment.Value(src))),
				face: baseFace,
				col:  baseCol,
			})
		case "String":
			s := child.(*ast.String)
			segments = append(segments, inlineSegment{
				text: sanitize(string(s.Value)),
				face: baseFace,
				col:  baseCol,
			})
		case "Emphasis":
			em := child.(*ast.Emphasis)
			face := baseFace
			if em.Level == 2 {
				face = c.fonts.bold
			} else if em.Level == 1 && c.fonts.italic != nil {
				face = c.fonts.italic
			}
			segments = append(segments, c.inlineSegments(child, src, face, baseCol)...)
		case "CodeSpan":
			codeFace := loadFace(systemFonts["Courier"].regular, c.cfg.FontSize)
			if codeFace == nil {
				codeFace = c.fonts.code
			}
			segments = append(segments, inlineSegment{
				text: extractText(child, src),
				face: codeFace,
				col:  c.cfg.TextColor.toRGBA(),
			})
		default:
			segments = append(segments, c.inlineSegments(child, src, baseFace, baseCol)...)
		}
	}
	return segments
}

func (c *canvas) renderSegmentsWrapped(segments []inlineSegment, lineHeight int) {
	c.ensureHeight(lineHeight + 20)
	for _, segment := range segments {
		parts := strings.Split(segment.text, "\n")
		for i, part := range parts {
			c.renderWordsWrapped(part, segment.face, segment.col, lineHeight)
			if i < len(parts)-1 {
				c.newWrappedLine(lineHeight)
			}
		}
	}
}

func (c *canvas) renderWordsWrapped(text string, face font.Face, col color.RGBA, lineHeight int) {
	for _, word := range strings.Fields(text) {
		prefix := ""
		if c.x > c.margin {
			prefix = " "
		}
		c.drawWrappedToken(prefix+word, face, col, lineHeight)
	}
}

func (c *canvas) drawWrappedToken(token string, face font.Face, col color.RGBA, lineHeight int) {
	maxX := c.contentRight()
	if c.x > c.margin && c.x+measure(face, token) > maxX {
		c.newWrappedLine(lineHeight)
		token = strings.TrimLeft(token, " ")
	}
	if measure(face, token) <= c.contentWidth() {
		c.drawString(token, face, col)
		return
	}

	var chunk string
	for _, r := range token {
		next := chunk + string(r)
		if chunk != "" && c.x+measure(face, next) > maxX {
			c.drawString(chunk, face, col)
			c.newWrappedLine(lineHeight)
			chunk = strings.TrimLeft(string(r), " ")
			continue
		}
		chunk = next
	}
	if chunk != "" {
		c.drawString(chunk, face, col)
	}
}

func (c *canvas) newWrappedLine(lineHeight int) {
	c.x = c.margin
	c.y += lineHeight
	c.ensureHeight(lineHeight + 20)
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
	if faceHeight(c.fonts.code) > lh {
		lh = faceHeight(c.fonts.code)
	}
	padding := mmToPx(1.5, c.cfg.DPI)
	maxLineW := c.contentWidth() - 2*padding
	if maxLineW < 1 {
		maxLineW = 1
	}
	var wrapped []string
	for _, line := range lines {
		wrapped = append(wrapped, wrapHard(line, c.fonts.code, maxLineW)...)
	}
	blockH := len(wrapped)*lh + padding*2
	c.ensureHeight(blockH + 24)

	// Background
	bgRect := image.Rect(c.margin-4, c.y, c.contentRight()+4, c.y+blockH)
	c.drawRect(bgRect, c.cfg.CodeBg.toRGBA())

	// Text
	c.x = c.margin + padding
	c.y += padding
	for _, line := range wrapped {
		c.x = c.margin + padding
		c.drawString(line, c.fonts.code, c.cfg.TextColor.toRGBA())
		c.y += lh
	}
	c.x = c.margin
	c.y += padding + 12
}

func wrapHard(text string, face font.Face, maxWidth int) []string {
	if text == "" {
		return []string{""}
	}
	var lines []string
	var chunk string
	for _, r := range text {
		next := chunk + string(r)
		if chunk != "" && measure(face, next) > maxWidth {
			lines = append(lines, chunk)
			chunk = string(r)
			continue
		}
		chunk = next
	}
	if chunk != "" {
		lines = append(lines, chunk)
	}
	return lines
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
	tableW := c.contentWidth()
	minRowH := mmToPx(c.cfg.TableCellHeight, c.cfg.DPI)
	if minRowH < 20 {
		minRowH = 20
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

	// Cell internal padding (pixels per side)
	cellPad := 12

	// Compute per-column widths
	colWidths := make([]int, numCols)
	if c.cfg.TableAutoWidth {
		for ci := 0; ci < numCols; ci++ {
			maxW := 0
			for ri, row := range rows {
				if ci < len(row) {
					face := c.fonts.regular
					if ri == 0 {
						face = headerFace
					}
					w := measure(face, row[ci])
					if w > maxW {
						maxW = w
					}
				}
			}
			colWidths[ci] = maxW + 2*cellPad
		}
		// Ensure total doesn't exceed available width; shrink proportionally if needed.
		totalW := 0
		for _, w := range colWidths {
			totalW += w
		}
		if totalW > tableW {
			scale := float64(tableW) / float64(totalW)
			for i := range colWidths {
				colWidths[i] = int(float64(colWidths[i]) * scale)
			}
		}
	} else {
		colW := tableW / numCols
		for i := range colWidths {
			colWidths[i] = colW
		}
	}

	xOffset := c.margin
	for ri, row := range rows {
		rowLines := make([][]string, numCols)
		rowH := minRowH
		for ci := 0; ci < numCols && ci < len(row); ci++ {
			face := c.fonts.regular
			if ri == 0 {
				face = headerFace
			}
			maxTextW := colWidths[ci] - 2*cellPad
			if maxTextW < 1 {
				maxTextW = 1
			}
			rowLines[ci] = wrapWords(row[ci], face, maxTextW)
			lineH := faceHeight(face) + 2
			cellH := len(rowLines[ci])*lineH + 2*cellPad
			if cellH > rowH {
				rowH = cellH
			}
		}
		c.ensureHeight(rowH + 10)
		x := xOffset
		for ci := 0; ci < numCols && ci < len(row); ci++ {
			cw := colWidths[ci]
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
			c.drawRect(image.Rect(x, y, x+cw, y+rowH), bg)

			// Borders
			borderCol := color.RGBA{180, 180, 180, 255}
			c.drawHorizontalLine(x, x+cw, y, borderCol, 0.5)
			c.drawHorizontalLine(x, x+cw, y+rowH, borderCol, 0.5)
			c.drawVerticalLine(x, y, y+rowH, borderCol)
			c.drawVerticalLine(x+cw, y, y+rowH, borderCol)

			// Text
			fg := c.cfg.TextColor.toRGBA()
			face := c.fonts.regular
			if ri == 0 {
				fg = c.cfg.TableHeaderFg.toRGBA()
				face = headerFace
			}
			lineH := faceHeight(face) + 2
			for li, line := range rowLines[ci] {
				d := &font.Drawer{
					Dst:  c.img,
					Face: face,
					Src:  image.NewUniform(fg),
					Dot:  fixed.P(x+cellPad, y+cellPad+faceHeight(face)+li*lineH),
				}
				d.DrawString(line)
			}
			x += cw
		}
		c.y += rowH
	}
	c.y += 14
}

func wrapWords(text string, face font.Face, maxWidth int) []string {
	if maxWidth < 1 {
		maxWidth = 1
	}
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}
	var lines []string
	line := words[0]
	for _, word := range words[1:] {
		next := line + " " + word
		if measure(face, next) <= maxWidth {
			line = next
			continue
		}
		lines = append(lines, wrapHard(line, face, maxWidth)...)
		line = word
	}
	lines = append(lines, wrapHard(line, face, maxWidth)...)
	return lines
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
		c.renderSegmentsWrapped([]inlineSegment{{
			text: bullet + text,
			face: face,
			col:  c.cfg.TextColor.toRGBA(),
		}}, lh+6)
		c.x = c.margin
		c.y += lh + 6
	}
	c.y += 10
}

func (c *canvas) renderBlockquote(n ast.Node, src []byte) {
	lc := c.cfg.BlockquoteLineColor.toRGBA()
	face := c.fonts.italic
	if face == nil {
		face = c.fonts.regular
	}
	lh := faceHeight(face) + 6
	text := extractText(n, src)
	lines := wrapWords(text, face, c.contentWidth()-14)
	quoteH := len(lines) * lh
	c.ensureHeight(quoteH + 12)
	c.drawVerticalLine(c.margin+2, c.y, c.y+quoteH, lc)

	x := c.margin + 14
	for _, line := range lines {
		d := &font.Drawer{
			Dst:  c.img,
			Face: face,
			Src:  image.NewUniform(c.cfg.BlockquoteTextColor.toRGBA()),
			Dot:  fixed.P(x, c.y+faceHeight(face)),
		}
		d.DrawString(line)
		c.y += lh
	}
	c.x = c.margin
	c.y += 12
}

func (c *canvas) renderHR() {
	c.ensureHeight(28)
	c.y += 12
	y := c.y
	hc := c.cfg.HRColor.toRGBA()
	lw := int(c.cfg.HRLineWidth * float64(c.cfg.DPI) / 25.4)
	if lw < 1 {
		lw = 1
	}
	c.drawHorizontalLine(c.margin, c.contentRight(), y, hc, float64(lw))
	c.y += 16
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

	// Ensure symmetric left/right padding in the crop.
	rightEdge := right + 1 + left // right padding = left padding
	if rightEdge > bounds.Dx() {
		rightEdge = bounds.Dx()
	}
	cropped := c.img.SubImage(image.Rect(0, 0, rightEdge, contentH))

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
