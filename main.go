package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"
)

var parser = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
).Parser()

func run(input, output string) error {
	r := newRenderer()
	r.src = []byte(input)
	reader := text.NewReader(r.src)
	doc := parser.Parse(reader)
	r.renderNodes(doc)

	pdfPath := strings.TrimSuffix(output, ".png") + ".pdf"
	if err := r.pdf.OutputFileAndClose(pdfPath); err != nil {
		return fmt.Errorf("PDF error: %w", err)
	}

	cmd := exec.Command("gs",
		"-dNOPAUSE", "-dBATCH", "-dQUIET",
		"-sDEVICE=png16m", "-r200",
		"-sOutputFile="+output, pdfPath,
	)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("PNG conversion error (is ghostscript installed?): %w", err)
	}

	os.Remove(pdfPath)
	return nil
}

func main() {
	var md string
	var output string

	// Parse args: md2img [-o output.png] [input.md]
	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		if args[i] == "-o" && i+1 < len(args) {
			output = args[i+1]
			i++
		} else {
			data, err := os.ReadFile(args[i])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", args[i], err)
				os.Exit(1)
			}
			md = string(data)
		}
	}

	if output == "" {
		output = "/tmp/md2img_output.png"
	}

	// Read from stdin if no file argument
	if md == "" {
		var buf bytes.Buffer
		buf.ReadFrom(os.Stdin)
		md = buf.String()
	}

	if strings.TrimSpace(md) == "" {
		fmt.Fprintf(os.Stderr, "Error: empty markdown input\n")
		os.Exit(1)
	}

	if err := run(md, output); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Done: %s\n", output)
}
