package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"

	md2img "github.com/jmaciasluque/md2img"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	return runWithArgs(os.Args[1:], os.Stdin, os.Stdout)
}

func runWithArgs(args []string, stdin io.Reader, stdout io.Writer) error {
	cfg := md2img.DefaultConfig()
	output := "/tmp/md2img_output.png"
	version := false
	tableFullWidth := false

	var textColor, headingColor string
	var tableHeaderBg, tableHeaderFg, tableRowEven, tableRowOdd string
	var codeBg, blockquoteLineColor, blockquoteTextColor, hrColor string
	var margin float64

	fs := flag.NewFlagSet("md2img", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	fs.StringVar(&output, "o", output, "output file path")
	fs.StringVar(&output, "output", output, "output file path")
	fs.BoolVar(&version, "version", false, "print version")

	fs.StringVar(&cfg.FontFamily, "font", cfg.FontFamily, "body font family")
	fs.Float64Var(&cfg.FontSize, "font-size", cfg.FontSize, "body font size in points")
	fs.StringVar(&cfg.HeadingFont, "heading-font", cfg.HeadingFont, "heading font family")

	fs.Float64Var(&cfg.PageWidth, "page-w", cfg.PageWidth, "page width in mm")
	fs.Float64Var(&cfg.PageHeight, "page-h", cfg.PageHeight, "page height in mm")
	fs.Float64Var(&margin, "margin", cfg.MarginTop, "all margins in mm")

	fs.StringVar(&textColor, "text-color", "", "body text color")
	fs.StringVar(&headingColor, "heading-color", "", "heading text color")
	fs.StringVar(&tableHeaderBg, "table-header-bg", "", "table header background color")
	fs.StringVar(&tableHeaderFg, "table-header-fg", "", "table header foreground color")
	fs.StringVar(&cfg.TableHeaderFont, "table-header-font", cfg.TableHeaderFont, "table header font family")
	fs.Float64Var(&cfg.TableHeaderSize, "table-header-size", cfg.TableHeaderSize, "table header font size")
	fs.StringVar(&tableRowEven, "table-row-even", "", "even table row background color")
	fs.StringVar(&tableRowOdd, "table-row-odd", "", "odd table row background color")
	fs.BoolVar(&tableFullWidth, "table-full-width", false, "stretch tables to full width")

	fs.StringVar(&codeBg, "code-bg", "", "code block background color")
	fs.StringVar(&cfg.CodeFont, "code-font", cfg.CodeFont, "code block font family")
	fs.Float64Var(&cfg.CodeFontSize, "code-font-size", cfg.CodeFontSize, "code block font size")

	fs.StringVar(&blockquoteLineColor, "blockquote-line-color", "", "blockquote line color")
	fs.StringVar(&blockquoteTextColor, "blockquote-text-color", "", "blockquote text color")
	fs.StringVar(&hrColor, "hr-color", "", "horizontal rule color")

	fs.IntVar(&cfg.DPI, "dpi", cfg.DPI, "image resolution in DPI")
	fs.BoolVar(&cfg.Trim, "trim", cfg.Trim, "auto-crop whitespace from PNG output")
	fs.Float64Var(&cfg.TrimPadding, "trim-padding", cfg.TrimPadding, "padding around content after trim in mm")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if version {
		fmt.Fprintf(stdout, "md2img %s\n", md2img.Version)
		return nil
	}

	cfg.MarginTop = margin
	cfg.MarginLeft = margin
	cfg.MarginRight = margin
	cfg.MarginBottom = margin
	if tableFullWidth {
		cfg.TableAutoWidth = false
	}

	if err := applyColors(&cfg, map[string]string{
		"text-color":            textColor,
		"heading-color":         headingColor,
		"table-header-bg":       tableHeaderBg,
		"table-header-fg":       tableHeaderFg,
		"table-row-even":        tableRowEven,
		"table-row-odd":         tableRowOdd,
		"code-bg":               codeBg,
		"blockquote-line-color": blockquoteLineColor,
		"blockquote-text-color": blockquoteTextColor,
		"hr-color":              hrColor,
	}); err != nil {
		return err
	}
	if err := validateConfig(cfg, output); err != nil {
		return err
	}

	md, err := readInput(fs.Args(), stdin)
	if err != nil {
		return err
	}
	if err := md2img.RenderWithConfig(md, output, cfg); err != nil {
		return err
	}

	fmt.Fprintf(stdout, "Done: %s\n", output)
	return nil
}

func applyColors(cfg *md2img.Config, values map[string]string) error {
	for name, value := range values {
		if value == "" {
			continue
		}
		c, err := md2img.HexToColor(value)
		if err != nil {
			return fmt.Errorf("invalid -%s: %w", name, err)
		}
		switch name {
		case "text-color":
			cfg.TextColor = c
		case "heading-color":
			cfg.HeadingColor = c
		case "table-header-bg":
			cfg.TableHeaderBg = c
		case "table-header-fg":
			cfg.TableHeaderFg = c
		case "table-row-even":
			cfg.TableRowEven = c
		case "table-row-odd":
			cfg.TableRowOdd = c
		case "code-bg":
			cfg.CodeBg = c
		case "blockquote-line-color":
			cfg.BlockquoteLineColor = c
		case "blockquote-text-color":
			cfg.BlockquoteTextColor = c
		case "hr-color":
			cfg.HRColor = c
		}
	}
	return nil
}

func validateConfig(cfg md2img.Config, output string) error {
	if output == "" {
		return fmt.Errorf("output path cannot be empty")
	}
	if cfg.DPI <= 0 {
		return fmt.Errorf("invalid -dpi: must be greater than 0")
	}
	if cfg.FontSize <= 0 {
		return fmt.Errorf("invalid -font-size: must be greater than 0")
	}
	if cfg.TableHeaderSize <= 0 {
		return fmt.Errorf("invalid -table-header-size: must be greater than 0")
	}
	if cfg.CodeFontSize <= 0 {
		return fmt.Errorf("invalid -code-font-size: must be greater than 0")
	}
	if cfg.PageWidth <= 0 {
		return fmt.Errorf("invalid -page-w: must be greater than 0")
	}
	if cfg.PageHeight <= 0 {
		return fmt.Errorf("invalid -page-h: must be greater than 0")
	}
	if cfg.MarginTop < 0 || cfg.MarginLeft < 0 || cfg.MarginRight < 0 || cfg.MarginBottom < 0 {
		return fmt.Errorf("invalid -margin: must be 0 or greater")
	}
	if cfg.TrimPadding < 0 {
		return fmt.Errorf("invalid -trim-padding: must be 0 or greater")
	}
	return nil
}

func readInput(args []string, stdin io.Reader) (string, error) {
	if len(args) > 1 {
		return "", fmt.Errorf("expected at most one input file, got %d", len(args))
	}
	if len(args) == 1 && args[0] != "-" {
		data, err := os.ReadFile(args[0])
		if err != nil {
			return "", fmt.Errorf("reading %s: %w", args[0], err)
		}
		return string(data), nil
	}

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(stdin); err != nil {
		return "", err
	}
	return buf.String(), nil
}
