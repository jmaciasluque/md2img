package main

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"

	md2img "github.com/jmaciasluque/md2img"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var md string
	var output string
	cfg := md2img.DefaultConfig()

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-o", "--output":
			if i+1 < len(args) {
				output = args[i+1]
				i++
			}
		case "-version", "--version":
			fmt.Printf("md2img %s\n", md2img.Version)
			return nil

		// Font
		case "-font", "--font":
			if i+1 < len(args) {
				cfg.FontFamily = args[i+1]
				i++
			}
		case "-font-size", "--font-size":
			if i+1 < len(args) {
				v, err := strconv.ParseFloat(args[i+1], 64)
				if err != nil {
					return fmt.Errorf("invalid -font-size: %v", err)
				}
				cfg.FontSize = v
				i++
			}

		// Page
		case "-page-w", "--page-w":
			if i+1 < len(args) {
				v, err := strconv.ParseFloat(args[i+1], 64)
				if err != nil {
					return fmt.Errorf("invalid -page-w: %v", err)
				}
				cfg.PageWidth = v
				i++
			}
		case "-page-h", "--page-h":
			if i+1 < len(args) {
				v, err := strconv.ParseFloat(args[i+1], 64)
				if err != nil {
					return fmt.Errorf("invalid -page-h: %v", err)
				}
				cfg.PageHeight = v
				i++
			}
		case "-margin", "--margin":
			if i+1 < len(args) {
				v, err := strconv.ParseFloat(args[i+1], 64)
				if err != nil {
					return fmt.Errorf("invalid -margin: %v", err)
				}
				cfg.MarginTop = v
				cfg.MarginLeft = v
				cfg.MarginRight = v
				cfg.MarginBottom = v
				i++
			}

		// Text
		case "-text-color", "--text-color":
			if i+1 < len(args) {
				c, err := md2img.HexToColor(args[i+1])
				if err != nil {
					return err
				}
				cfg.TextColor = c
				i++
			}

		// Headings
		case "-heading-color", "--heading-color":
			if i+1 < len(args) {
				c, err := md2img.HexToColor(args[i+1])
				if err != nil {
					return err
				}
				cfg.HeadingColor = c
				i++
			}
		case "-heading-font", "--heading-font":
			if i+1 < len(args) {
				cfg.HeadingFont = args[i+1]
				i++
			}

		// Table
		case "-table-header-bg", "--table-header-bg":
			if i+1 < len(args) {
				c, err := md2img.HexToColor(args[i+1])
				if err != nil {
					return err
				}
				cfg.TableHeaderBg = c
				i++
			}
		case "-table-header-fg", "--table-header-fg":
			if i+1 < len(args) {
				c, err := md2img.HexToColor(args[i+1])
				if err != nil {
					return err
				}
				cfg.TableHeaderFg = c
				i++
			}
		case "-table-header-font", "--table-header-font":
			if i+1 < len(args) {
				cfg.TableHeaderFont = args[i+1]
				i++
			}
		case "-table-header-size", "--table-header-size":
			if i+1 < len(args) {
				v, err := strconv.ParseFloat(args[i+1], 64)
				if err != nil {
					return fmt.Errorf("invalid -table-header-size: %v", err)
				}
				cfg.TableHeaderSize = v
				i++
			}
		case "-table-row-even", "--table-row-even":
			if i+1 < len(args) {
				c, err := md2img.HexToColor(args[i+1])
				if err != nil {
					return err
				}
				cfg.TableRowEven = c
				i++
			}
		case "-table-row-odd", "--table-row-odd":
			if i+1 < len(args) {
				c, err := md2img.HexToColor(args[i+1])
				if err != nil {
					return err
				}
				cfg.TableRowOdd = c
				i++
			}

		// Code
		case "-code-bg", "--code-bg":
			if i+1 < len(args) {
				c, err := md2img.HexToColor(args[i+1])
				if err != nil {
					return err
				}
				cfg.CodeBg = c
				i++
			}
		case "-code-font", "--code-font":
			if i+1 < len(args) {
				cfg.CodeFont = args[i+1]
				i++
			}
		case "-code-font-size", "--code-font-size":
			if i+1 < len(args) {
				v, err := strconv.ParseFloat(args[i+1], 64)
				if err != nil {
					return fmt.Errorf("invalid -code-font-size: %v", err)
				}
				cfg.CodeFontSize = v
				i++
			}

		// Blockquote
		case "-blockquote-line-color", "--blockquote-line-color":
			if i+1 < len(args) {
				c, err := md2img.HexToColor(args[i+1])
				if err != nil {
					return err
				}
				cfg.BlockquoteLineColor = c
				i++
			}
		case "-blockquote-text-color", "--blockquote-text-color":
			if i+1 < len(args) {
				c, err := md2img.HexToColor(args[i+1])
				if err != nil {
					return err
				}
				cfg.BlockquoteTextColor = c
				i++
			}

		// Horizontal rule
		case "-hr-color", "--hr-color":
			if i+1 < len(args) {
				c, err := md2img.HexToColor(args[i+1])
				if err != nil {
					return err
				}
				cfg.HRColor = c
				i++
			}

		// Output
		case "-dpi", "--dpi":
			if i+1 < len(args) {
				v, err := strconv.Atoi(args[i+1])
				if err != nil {
					return fmt.Errorf("invalid -dpi: %v", err)
				}
				cfg.DPI = v
				i++
			}
		case "-pdf", "--pdf":
			cfg.AsPDF = true
		case "-trim", "--trim":
			cfg.Trim = true

		default:
			if strings.HasPrefix(args[i], "-") {
				return fmt.Errorf("unknown flag: %s", args[i])
			}
			data, err := os.ReadFile(args[i])
			if err != nil {
				return fmt.Errorf("reading %s: %w", args[i], err)
			}
			md = string(data)
		}
	}

	if output == "" {
		if cfg.AsPDF {
			output = "/tmp/md2img_output.pdf"
		} else {
			output = "/tmp/md2img_output.png"
		}
	}

	if md == "" {
		var buf bytes.Buffer
		buf.ReadFrom(os.Stdin)
		md = buf.String()
	}

	if err := md2img.RenderWithConfig(md, output, cfg); err != nil {
		return err
	}

	fmt.Printf("Done: %s\n", output)
	return nil
}
