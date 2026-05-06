package md2img

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/jung-kurt/gofpdf/v2"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"
)

// Version is set at build time via ldflags.
var Version = "dev"

const (
	pageMargin = 15.0
	cellH      = 8.0
	a4Height   = 297.0
)

var parser = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
).Parser()

type renderer struct {
	pdf *gofpdf.Fpdf
	w   float64
	src []byte
}

func newRenderer() *renderer {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddPage()
	pdf.SetMargins(pageMargin, pageMargin, pageMargin)
	w, _ := pdf.GetPageSize()
	return &renderer{pdf: pdf, w: w - 2*pageMargin}
}

func (r *renderer) ensureSpace(h float64) {
	if r.pdf.GetY()+h > a4Height-pageMargin {
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
	var size float64
	switch h.Level {
	case 1:
		size = 22
	case 2:
		size = 18
	case 3:
		size = 14
	default:
		size = 12
	}
	r.pdf.SetFont("Helvetica", "B", size)
	r.pdf.SetTextColor(40, 40, 80)
	r.ensureSpace(size + 5)
	r.pdf.MultiCell(r.w, size*0.6, text, "", "L", false)
	r.pdf.Ln(3)
}

func (r *renderer) renderParagraph(n ast.Node) {
	text := r.extractText(n)
	r.pdf.SetFont("Helvetica", "", 11)
	r.pdf.SetTextColor(40, 40, 40)
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
	h := float64(len(lines))*4.5 + 6
	r.ensureSpace(h)
	r.pdf.SetFillColor(240, 240, 240)
	r.pdf.Rect(pageMargin, r.pdf.GetY(), r.w, h, "F")
	r.pdf.SetFont("Courier", "", 9)
	r.pdf.SetTextColor(40, 40, 40)
	r.pdf.SetXY(pageMargin+3, r.pdf.GetY()+3)
	for _, line := range lines {
		r.pdf.Cell(r.w-6, 4.5, line)
		r.pdf.Ln(4.5)
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
	totalH := float64(len(rows)) * cellH
	r.ensureSpace(totalH + 4)

	y := r.pdf.GetY()

	for ri, row := range rows {
		for ci := 0; ci < numCols && ci < len(row); ci++ {
			x := pageMargin + float64(ci)*colW

			if ri == 0 {
				r.pdf.SetFillColor(50, 50, 80)
				r.pdf.SetTextColor(200, 200, 255)
				r.pdf.SetFont("Helvetica", "B", 10)
			} else {
				if ri%2 == 0 {
					r.pdf.SetFillColor(245, 245, 250)
				} else {
					r.pdf.SetFillColor(255, 255, 255)
				}
				r.pdf.SetTextColor(40, 40, 40)
				r.pdf.SetFont("Helvetica", "", 10)
			}

			r.pdf.SetXY(x, y)
			r.pdf.CellFormat(colW, cellH, "  "+row[ci], "1", 0, "L", true, 0, "")
		}
		y += cellH
	}
	r.pdf.SetY(y + 4)
}

func (r *renderer) renderList(n ast.Node) {
	l := n.(*ast.List)
	r.pdf.SetFont("Helvetica", "", 11)
	r.pdf.SetTextColor(40, 40, 40)
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
	r.pdf.SetDrawColor(100, 100, 200)
	r.pdf.SetLineWidth(1)
	y := r.pdf.GetY()
	r.pdf.Line(pageMargin+2, y, pageMargin+2, y+6)
	r.pdf.SetX(pageMargin + 8)
	r.pdf.SetFont("Helvetica", "I", 11)
	r.pdf.SetTextColor(100, 100, 100)
	text := r.extractText(n)
	r.ensureSpace(8)
	r.pdf.MultiCell(r.w-10, 6, text, "", "L", false)
	r.pdf.Ln(3)
}

func (r *renderer) renderHR() {
	y := r.pdf.GetY()
	r.pdf.SetDrawColor(180, 180, 180)
	r.pdf.SetLineWidth(0.3)
	r.pdf.Line(pageMargin, y, pageMargin+r.w, y)
	r.pdf.Ln(5)
}

// Render converts markdown input to a PNG file at the given output path.
// It requires Ghostscript (gs) to be installed on the system.
func Render(input, output string) error {
	r := newRenderer()
	r.src = []byte(input)
	reader := text.NewReader(r.src)
	doc := parser.Parse(reader)
	r.renderNodes(doc)

	pdfPath := strings.TrimSuffix(output, ".png") + ".pdf"
	if err := r.pdf.OutputFileAndClose(pdfPath); err != nil {
		return fmt.Errorf("PDF error: %w", err)
	}
	defer os.Remove(pdfPath)

	cmd := exec.Command("gs",
		"-dNOPAUSE", "-dBATCH", "-dQUIET",
		"-sDEVICE=png16m", "-r200",
		"-sOutputFile="+output, pdfPath,
	)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("PNG conversion error (is ghostscript installed?): %w", err)
	}

	return nil
}
