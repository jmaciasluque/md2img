package main

import (
	"bytes"
	"fmt"
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
	var md string
	var output string

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-o":
			if i+1 < len(args) {
				output = args[i+1]
				i++
			}
		case "-version":
			fmt.Printf("md2img %s\n", md2img.Version)
			return nil
		default:
			data, err := os.ReadFile(args[i])
			if err != nil {
				return fmt.Errorf("reading %s: %w", args[i], err)
			}
			md = string(data)
		}
	}

	if output == "" {
		output = "/tmp/md2img_output.png"
	}

	if md == "" {
		var buf bytes.Buffer
		buf.ReadFrom(os.Stdin)
		md = buf.String()
	}

	if err := md2img.Render(md, output); err != nil {
		return err
	}

	fmt.Printf("Done: %s\n", output)
	return nil
}
