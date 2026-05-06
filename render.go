package md2img

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/jung-kurt/gofpdf/v2"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"
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
	r, err := strconv.ParseUint(s[0:2], 16, 8)
	if err != nil {
		return Color{}, fmt.Errorf("invalid hex color: %s", s)
	}
	g, err := strconv.ParseUint(s[2:4], 16, 8)
	if err != nil {
		return Color{}, fmt.Errorf("invalid hex color: %s", s)
	}
	b, err := strconv.ParseUint(s[4:6], 16, 8)
	if err != nil {
		return Color{}, fmt.Errorf("invalid hex color: %s", s)
	}
	return Color{R: int(r), G: int(g), B: int(b)}, nil
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
	MarginBottom float64 // Bottom margin in mm (used for page break check)

	// Text colors
	TextColor Color // Default body text color

	// Heading colors and sizes
	HeadingColor   Color    // Heading text color
	HeadingSizes   [6]float64 // Font sizes for H1–H6
	HeadingFont    string   // Heading font family override ("", same as FontFamily)

	// Table
	TableHeaderBg    Color   // Table header background
	TableHeaderFg    Color   // Table header text color
	TableHeaderFont  string  // Table header font family ("", same as FontFamily)
	TableHeaderSize  float64 // Table header font size
	TableCellHeight  float64 // Row height in mm
	TableRowEven     Color   // Even row background
	TableRowOdd      Color   // Odd row background

	// Code block
	CodeBg          Color   // Code block background
	CodeFont        string  // Code font family (default: "Courier")
	CodeFontSize    float64 // Code font size
	CodeLineHeight  float64 // Code line height in mm

	// Blockquote
	BlockquoteLineColor Color  // Left border color
	BlockquoteTextColor Color  // Quote text color
	BlockquoteFont      string // Quote font (default: same as FontFamily, italic)

	// Horizontal rule
	HRColor     Color   // HR line color
	HRLineWidth float64 // HR line thickness in mm

	// Output
	DPI     int    // Ghostscript DPI (default: 200)
	AsPDF   bool   // Output PDF instead of PNG
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

		TableHeaderBg:   Color{50, 50, 80},
		TableHeaderFg:   Color{200, 200, 255},
		TableHeaderFont: "",
		TableHeaderSize: 10,
		TableCellHeight: 8,
		TableRowEven:    Color{245, 245, 250},
		TableRowOdd:     Color{255, 255, 255},

		CodeBg:         Color{240, 240, 240},
		CodeFont:       "Courier",
		CodeFontSize:   9,
		CodeLineHeight: 4.5,

		BlockquoteLineColor: Color{100, 100, 200},
		BlockquoteTextColor: Color{100, 100, 100},
		BlockquoteFont:      "",

		HRColor:     Color{180, 180, 180},
		HRLineWidth: 0.3,

		DPI:   200,
		AsPDF: false,
	}
}

type renderer struct {
	pdf *gofpdf.Fpdf
	w   float64 // content width (page width minus margins)
	src []byte
	cfg Config
}

func newRenderer(cfg Config) *renderer {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(false, 0)
	// Add the first page. Use AddPageFormat for custom sizes.
	isA4 := cfg.PageWidth == 210 && cfg.PageHeight == 297
	if isA4 {
		pdf.AddPage()
	} else {
		pdf.AddPageFormat("P", gofpdf.SizeType{Wd: cfg.PageWidth, Ht: cfg.PageHeight})
	}
	pdf.SetMargins(cfg.MarginLeft, cfg.MarginTop, cfg.MarginRight)
	w, _ := pdf.GetPageSize()
	return &renderer{pdf: pdf, w: w - cfg.MarginLeft - cfg.MarginRight, cfg: cfg}
}

func (r *renderer) ensureSpace(h float64) {
	if r.pdf.GetY()+h > r.cfg.PageHeight-r.cfg.MarginBottom {
		r.pdf.AddPage()
	}
}

func (r *renderer) renderNodes(n ast.Node) {
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		r.renderNode(child)
	}
}

func (r *renderer) renderNode(n ast.Node) {
	kind := n.Kind().String()
	switch {
	case kind == "Heading":
		r.renderHeading(n)
	case kind == "Paragraph":
		r.renderParagraph(n)
	case kind == "CodeBlock" || kind == "FencedCodeBlock":
		r.renderCodeBlock(n)
	case kind == "Table":
		r.renderTable(n)
	case kind == "List":
		r.renderList(n)
	case kind == "Blockquote":
		r.renderBlockquote(n)
	case kind == "ThematicBreak":
		r.renderHR()
	}
}

func (r *renderer) extractText(n ast.Node) string {
	var buf bytes.Buffer
	ast.Walk(n, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			switch t := n.(type) {
			case *ast.Text:
				buf.Write(t.Segment.Value(r.src))
			case *ast.String:
				buf.Write(t.Value)
			}
		}
		return ast.WalkContinue, nil
	})
	return sanitize(buf.String())
}

func (r *renderer) renderHeading(n ast.Node) {
	h := n.(*ast.Heading)
	text := r.extractText(n)
	idx := h.Level - 1
	if idx < 0 {
		idx = 0
	}
	if idx > 5 {
		idx = 5
	}
	size := r.cfg.HeadingSizes[idx]

	font := r.cfg.HeadingFont
	if font == "" {
		font = r.cfg.FontFamily
	}
	r.pdf.SetFont(font, "B", size)
	r.pdf.SetTextColor(r.cfg.HeadingColor.R, r.cfg.HeadingColor.G, r.cfg.HeadingColor.B)
	r.ensureSpace(size + 5)
	r.pdf.MultiCell(r.w, size*0.6, text, "", "L", false)
	r.pdf.Ln(3)
}

func (r *renderer) renderParagraph(n ast.Node) {
	text := r.extractText(n)
	r.pdf.SetFont(r.cfg.FontFamily, "", r.cfg.FontSize)
	r.pdf.SetTextColor(r.cfg.TextColor.R, r.cfg.TextColor.G, r.cfg.TextColor.B)
	r.ensureSpace(8)
	r.pdf.MultiCell(r.w, 6, text, "", "L", false)
	r.pdf.Ln(3)
}

func (r *renderer) renderCodeBlock(n ast.Node) {
	var lines []string
	switch block := n.(type) {
	case *ast.FencedCodeBlock:
		for i := 0; i < block.Lines().Len(); i++ {
			seg := block.Lines().At(i)
			lines = append(lines, strings.TrimRight(string(seg.Value(r.src)), "\n"))
		}
	case *ast.CodeBlock:
		for i := 0; i < block.Lines().Len(); i++ {
			seg := block.Lines().At(i)
			lines = append(lines, strings.TrimRight(string(seg.Value(r.src)), "\n"))
		}
	default:
		lines = strings.Split(r.extractText(n), "\n")
	}
	lh := r.cfg.CodeLineHeight
	h := float64(len(lines))*lh + 6
	r.ensureSpace(h)
	r.pdf.SetFillColor(r.cfg.CodeBg.R, r.cfg.CodeBg.G, r.cfg.CodeBg.B)
	r.pdf.Rect(r.cfg.MarginLeft, r.pdf.GetY(), r.w, h, "F")
	r.pdf.SetFont(r.cfg.CodeFont, "", r.cfg.CodeFontSize)
	r.pdf.SetTextColor(r.cfg.TextColor.R, r.cfg.TextColor.G, r.cfg.TextColor.B)
	r.pdf.SetXY(r.cfg.MarginLeft+3, r.pdf.GetY()+3)
	for _, line := range lines {
		r.pdf.Cell(r.w-6, lh, line)
		r.pdf.Ln(lh)
	}
	r.pdf.Ln(4)
}

func (r *renderer) renderTable(n ast.Node) {
	var rows [][]string
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		kind := child.Kind().String()
		if kind == "TableHeader" || kind == "TableRow" {
			var cells []string
			for cell := child.FirstChild(); cell != nil; cell = cell.NextSibling() {
				if cell.Kind().String() == "TableCell" {
					cells = append(cells, strings.TrimSpace(r.extractText(cell)))
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
	colW := r.w / float64(numCols)
	ch := r.cfg.TableCellHeight
	totalH := float64(len(rows)) * ch
	r.ensureSpace(totalH + 4)

	y := r.pdf.GetY()

	for ri, row := range rows {
		for ci := 0; ci < numCols && ci < len(row); ci++ {
			x := r.cfg.MarginLeft + float64(ci)*colW

			if ri == 0 {
				r.pdf.SetFillColor(r.cfg.TableHeaderBg.R, r.cfg.TableHeaderBg.G, r.cfg.TableHeaderBg.B)
				r.pdf.SetTextColor(r.cfg.TableHeaderFg.R, r.cfg.TableHeaderFg.G, r.cfg.TableHeaderFg.B)
				font := r.cfg.TableHeaderFont
				if font == "" {
					font = r.cfg.FontFamily
				}
				r.pdf.SetFont(font, "B", r.cfg.TableHeaderSize)
			} else {
				if ri%2 == 0 {
					r.pdf.SetFillColor(r.cfg.TableRowEven.R, r.cfg.TableRowEven.G, r.cfg.TableRowEven.B)
				} else {
					r.pdf.SetFillColor(r.cfg.TableRowOdd.R, r.cfg.TableRowOdd.G, r.cfg.TableRowOdd.B)
				}
				r.pdf.SetTextColor(r.cfg.TextColor.R, r.cfg.TextColor.G, r.cfg.TextColor.B)
				r.pdf.SetFont(r.cfg.FontFamily, "", r.cfg.TableHeaderSize)
			}

			r.pdf.SetXY(x, y)
			r.pdf.CellFormat(colW, ch, "  "+row[ci], "1", 0, "L", true, 0, "")
		}
		y += ch
	}
	r.pdf.SetY(y + 4)
}

func (r *renderer) renderList(n ast.Node) {
	l := n.(*ast.List)
	r.pdf.SetFont(r.cfg.FontFamily, "", r.cfg.FontSize)
	r.pdf.SetTextColor(r.cfg.TextColor.R, r.cfg.TextColor.G, r.cfg.TextColor.B)
	i := 1
	for item := l.FirstChild(); item != nil; item = item.NextSibling() {
		text := r.extractText(item)
		bullet := "* "
		if l.IsOrdered() {
			bullet = fmt.Sprintf("%d. ", i)
			i++
		}
		r.ensureSpace(8)
		r.pdf.MultiCell(r.w, 6, bullet+text, "", "L", false)
		r.pdf.Ln(1)
	}
	r.pdf.Ln(3)
}

func (r *renderer) renderBlockquote(n ast.Node) {
	lc := r.cfg.BlockquoteLineColor
	r.pdf.SetDrawColor(lc.R, lc.G, lc.B)
	r.pdf.SetLineWidth(1)
	y := r.pdf.GetY()
	r.pdf.Line(r.cfg.MarginLeft+2, y, r.cfg.MarginLeft+2, y+6)
	r.pdf.SetX(r.cfg.MarginLeft + 8)
	font := r.cfg.BlockquoteFont
	if font == "" {
		font = r.cfg.FontFamily
	}
	r.pdf.SetFont(font, "I", r.cfg.FontSize)
	tc := r.cfg.BlockquoteTextColor
	r.pdf.SetTextColor(tc.R, tc.G, tc.B)
	text := r.extractText(n)
	r.ensureSpace(8)
	r.pdf.MultiCell(r.w-10, 6, text, "", "L", false)
	r.pdf.Ln(3)
}

func (r *renderer) renderHR() {
	y := r.pdf.GetY()
	hc := r.cfg.HRColor
	r.pdf.SetDrawColor(hc.R, hc.G, hc.B)
	r.pdf.SetLineWidth(r.cfg.HRLineWidth)
	r.pdf.Line(r.cfg.MarginLeft, y, r.cfg.MarginLeft+r.w, y)
	r.pdf.Ln(5)
}

// Render converts markdown input to a PNG file at the given output path.
// It requires Ghostscript (gs) to be installed on the system.
// Uses DefaultConfig(). For custom options, use RenderWithConfig.
func Render(input, output string) error {
	return RenderWithConfig(input, output, DefaultConfig())
}

// RenderWithConfig converts markdown input to a PNG or PDF file using the
// given configuration. If cfg.AsPDF is true, output is a PDF file directly
// (Ghostscript is not needed). Otherwise, Ghostscript converts the PDF to PNG.
func RenderWithConfig(input, output string, cfg Config) error {
	r := newRenderer(cfg)
	r.src = []byte(input)
	reader := text.NewReader(r.src)
	doc := parser.Parse(reader)
	r.renderNodes(doc)

	pdfPath := strings.TrimSuffix(output, ".png") + ".pdf"
	if err := r.pdf.OutputFileAndClose(pdfPath); err != nil {
		return fmt.Errorf("PDF error: %w", err)
	}

	if cfg.AsPDF {
		// Rename PDF to the requested output path if it differs
		if pdfPath != output {
			if err := os.Rename(pdfPath, output); err != nil {
				return fmt.Errorf("PDF rename error: %w", err)
			}
		}
		return nil
	}

	defer os.Remove(pdfPath)

	cmd := exec.Command("gs",
		"-dNOPAUSE", "-dBATCH", "-dQUIET",
		"-sDEVICE=png16m",
		fmt.Sprintf("-r%d", cfg.DPI),
		"-sOutputFile="+output, pdfPath,
	)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("PNG conversion error (is ghostscript installed?): %w", err)
	}

	return nil
}
