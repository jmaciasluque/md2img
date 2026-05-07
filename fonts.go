package md2img

import (
	"fmt"
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/opentype"
)

// fontFamily maps Config font names to system font paths.
type fontFamily struct {
	regular string
	bold    string
	italic  string
	boldItalic string
}

// systemFonts lists font search paths by family name.
var systemFonts = map[string]fontFamily{
	"Helvetica": {
		regular:     "/System/Library/Fonts/Supplemental/Arial.ttf",
		bold:        "/System/Library/Fonts/Supplemental/Arial Bold.ttf",
		italic:      "/System/Library/Fonts/Supplemental/Arial Italic.ttf",
		boldItalic:  "/System/Library/Fonts/Supplemental/Arial Bold Italic.ttf",
	},
	"Arial": {
		regular:     "/System/Library/Fonts/Supplemental/Arial.ttf",
		bold:        "/System/Library/Fonts/Supplemental/Arial Bold.ttf",
		italic:      "/System/Library/Fonts/Supplemental/Arial Italic.ttf",
		boldItalic:  "/System/Library/Fonts/Supplemental/Arial Bold Italic.ttf",
	},
	"Times": {
		regular:     "/System/Library/Fonts/Supplemental/Times New Roman.ttf",
		bold:        "/System/Library/Fonts/Supplemental/Times New Roman Bold.ttf",
		italic:      "/System/Library/Fonts/Supplemental/Times New Roman Italic.ttf",
		boldItalic:  "/System/Library/Fonts/Supplemental/Times New Roman Bold Italic.ttf",
	},
	"Courier": {
		regular:     "/System/Library/Fonts/Supplemental/Courier New.ttf",
		bold:        "/System/Library/Fonts/Supplemental/Courier New Bold.ttf",
		italic:      "/System/Library/Fonts/Supplemental/Courier New Italic.ttf",
		boldItalic:  "/System/Library/Fonts/Supplemental/Courier New Bold Italic.ttf",
	},
	"Courier New": {
		regular:     "/System/Library/Fonts/Supplemental/Courier New.ttf",
		bold:        "/System/Library/Fonts/Supplemental/Courier New Bold.ttf",
		italic:      "/System/Library/Fonts/Supplemental/Courier New Italic.ttf",
		boldItalic:  "/System/Library/Fonts/Supplemental/Courier New Bold Italic.ttf",
	},
}

// loadFace loads a TTF font at the given size. Falls back to basicfont.
func loadFace(path string, size float64) font.Face {
	if path == "" {
		return basicfont.Face7x13
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return basicfont.Face7x13
	}
	f, err := opentype.Parse(data)
	if err != nil {
		return basicfont.Face7x13
	}
	face, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return basicfont.Face7x13
	}
	return face
}

// fontSet holds loaded faces for a font family.
type fontSet struct {
	regular font.Face
	bold    font.Face
	italic  font.Face
	code    font.Face
}

// loadFonts loads font faces for the configured families.
func loadFonts(bodyFamily, codeFamily string, bodySize, codeSize float64) fontSet {
	body := systemFonts[bodyFamily]
	if body.regular == "" {
		body = systemFonts["Helvetica"]
	}
	code := systemFonts[codeFamily]
	if code.regular == "" {
		code = systemFonts["Courier"]
	}

	return fontSet{
		regular: loadFace(body.regular, bodySize),
		bold:    loadFace(body.bold, bodySize),
		italic:  loadFace(body.italic, bodySize),
		code:    loadFace(code.regular, codeSize),
	}
}

// faceHeight returns the approximate line height for a font.Face.
func faceHeight(f font.Face) int {
	bounds, _, _ := f.GlyphBounds('M')
	h := bounds.Max.Y - bounds.Min.Y
	if h <= 0 {
		return 16 // fallback
	}
	// Convert from 26.6 fixed-point to pixels.
	return int(h>>6) + 4
}

// measure returns the width of a string in pixels.
func measure(f font.Face, s string) int {
	w := 0
	for _, r := range s {
		adv, _ := f.GlyphAdvance(r)
		w += int(adv >> 6)
	}
	return w
}

// Ensure fmt is used.
var _ = fmt.Sprintf
