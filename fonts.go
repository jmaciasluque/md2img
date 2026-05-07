package md2img

import (
	"os"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/opentype"
)

// fontFamily maps Config font names to system font paths.
type fontFamily struct {
	regular    string
	bold       string
	italic     string
	boldItalic string
}

// systemFonts lists font search paths by family name, checked in order.
var systemFonts = map[string]fontFamily{
	"Helvetica": {
		regular:    findFirst("/System/Library/Fonts/Supplemental/Arial.ttf", "/usr/share/fonts/truetype/msttcorefonts/Arial.ttf", "/usr/share/fonts/truetype/liberation2/LiberationSans-Regular.ttf", "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf"),
		bold:       findFirst("/System/Library/Fonts/Supplemental/Arial Bold.ttf", "/usr/share/fonts/truetype/msttcorefonts/Arial_Bold.ttf", "/usr/share/fonts/truetype/liberation2/LiberationSans-Bold.ttf", "/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf"),
		italic:     findFirst("/System/Library/Fonts/Supplemental/Arial Italic.ttf", "/usr/share/fonts/truetype/msttcorefonts/Arial_Italic.ttf", "/usr/share/fonts/truetype/liberation2/LiberationSans-Italic.ttf", "/usr/share/fonts/truetype/dejavu/DejaVuSans-Oblique.ttf"),
		boldItalic: findFirst("/System/Library/Fonts/Supplemental/Arial Bold Italic.ttf", "/usr/share/fonts/truetype/msttcorefonts/Arial_Bold_Italic.ttf", "/usr/share/fonts/truetype/liberation2/LiberationSans-BoldItalic.ttf", "/usr/share/fonts/truetype/dejavu/DejaVuSans-BoldOblique.ttf"),
	},
	"Arial": {
		regular:    findFirst("/System/Library/Fonts/Supplemental/Arial.ttf", "/usr/share/fonts/truetype/msttcorefonts/Arial.ttf", "/usr/share/fonts/truetype/liberation2/LiberationSans-Regular.ttf", "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf"),
		bold:       findFirst("/System/Library/Fonts/Supplemental/Arial Bold.ttf", "/usr/share/fonts/truetype/msttcorefonts/Arial_Bold.ttf", "/usr/share/fonts/truetype/liberation2/LiberationSans-Bold.ttf", "/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf"),
		italic:     findFirst("/System/Library/Fonts/Supplemental/Arial Italic.ttf", "/usr/share/fonts/truetype/msttcorefonts/Arial_Italic.ttf", "/usr/share/fonts/truetype/liberation2/LiberationSans-Italic.ttf", "/usr/share/fonts/truetype/dejavu/DejaVuSans-Oblique.ttf"),
		boldItalic: findFirst("/System/Library/Fonts/Supplemental/Arial Bold Italic.ttf", "/usr/share/fonts/truetype/msttcorefonts/Arial_Bold_Italic.ttf", "/usr/share/fonts/truetype/liberation2/LiberationSans-BoldItalic.ttf", "/usr/share/fonts/truetype/dejavu/DejaVuSans-BoldOblique.ttf"),
	},
	"Times": {
		regular:    findFirst("/System/Library/Fonts/Supplemental/Times New Roman.ttf", "/usr/share/fonts/truetype/msttcorefonts/times.ttf", "/usr/share/fonts/truetype/liberation2/LiberationSerif-Regular.ttf", "/usr/share/fonts/truetype/dejavu/DejaVuSerif.ttf"),
		bold:       findFirst("/System/Library/Fonts/Supplemental/Times New Roman Bold.ttf", "/usr/share/fonts/truetype/msttcorefonts/timesbd.ttf", "/usr/share/fonts/truetype/liberation2/LiberationSerif-Bold.ttf", "/usr/share/fonts/truetype/dejavu/DejaVuSerif-Bold.ttf"),
		italic:     findFirst("/System/Library/Fonts/Supplemental/Times New Roman Italic.ttf", "/usr/share/fonts/truetype/msttcorefonts/timesi.ttf", "/usr/share/fonts/truetype/liberation2/LiberationSerif-Italic.ttf", "/usr/share/fonts/truetype/dejavu/DejaVuSerif-Italic.ttf"),
		boldItalic: findFirst("/System/Library/Fonts/Supplemental/Times New Roman Bold Italic.ttf", "/usr/share/fonts/truetype/msttcorefonts/timesbi.ttf", "/usr/share/fonts/truetype/liberation2/LiberationSerif-BoldItalic.ttf", "/usr/share/fonts/truetype/dejavu/DejaVuSerif-BoldItalic.ttf"),
	},
	"Courier": {
		regular:    findFirst("/System/Library/Fonts/Supplemental/Courier New.ttf", "/usr/share/fonts/truetype/msttcorefonts/cour.ttf", "/usr/share/fonts/truetype/liberation2/LiberationMono-Regular.ttf", "/usr/share/fonts/truetype/dejavu/DejaVuSansMono.ttf"),
		bold:       findFirst("/System/Library/Fonts/Supplemental/Courier New Bold.ttf", "/usr/share/fonts/truetype/msttcorefonts/courbd.ttf", "/usr/share/fonts/truetype/liberation2/LiberationMono-Bold.ttf", "/usr/share/fonts/truetype/dejavu/DejaVuSansMono-Bold.ttf"),
		italic:     findFirst("/System/Library/Fonts/Supplemental/Courier New Italic.ttf", "/usr/share/fonts/truetype/msttcorefonts/couri.ttf", "/usr/share/fonts/truetype/liberation2/LiberationMono-Italic.ttf", "/usr/share/fonts/truetype/dejavu/DejaVuSansMono-Oblique.ttf"),
		boldItalic: findFirst("/System/Library/Fonts/Supplemental/Courier New Bold Italic.ttf", "/usr/share/fonts/truetype/msttcorefonts/courbi.ttf", "/usr/share/fonts/truetype/liberation2/LiberationMono-BoldItalic.ttf", "/usr/share/fonts/truetype/dejavu/DejaVuSansMono-BoldOblique.ttf"),
	},
	"Courier New": {
		regular:    findFirst("/System/Library/Fonts/Supplemental/Courier New.ttf", "/usr/share/fonts/truetype/msttcorefonts/cour.ttf", "/usr/share/fonts/truetype/liberation2/LiberationMono-Regular.ttf", "/usr/share/fonts/truetype/dejavu/DejaVuSansMono.ttf"),
		bold:       findFirst("/System/Library/Fonts/Supplemental/Courier New Bold.ttf", "/usr/share/fonts/truetype/msttcorefonts/courbd.ttf", "/usr/share/fonts/truetype/liberation2/LiberationMono-Bold.ttf", "/usr/share/fonts/truetype/dejavu/DejaVuSansMono-Bold.ttf"),
		italic:     findFirst("/System/Library/Fonts/Supplemental/Courier New Italic.ttf", "/usr/share/fonts/truetype/msttcorefonts/couri.ttf", "/usr/share/fonts/truetype/liberation2/LiberationMono-Italic.ttf", "/usr/share/fonts/truetype/dejavu/DejaVuSansMono-Oblique.ttf"),
		boldItalic: findFirst("/System/Library/Fonts/Supplemental/Courier New Bold Italic.ttf", "/usr/share/fonts/truetype/msttcorefonts/courbi.ttf", "/usr/share/fonts/truetype/liberation2/LiberationMono-BoldItalic.ttf", "/usr/share/fonts/truetype/dejavu/DejaVuSansMono-BoldOblique.ttf"),
	},
}

// findFirst returns the first path that exists on disk.
func findFirst(paths ...string) string {
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

var fontCache sync.Map

// loadFace loads a TTF/OTF font at the given size. Falls back to basicfont.
func loadFace(path string, size float64) font.Face {
	if path == "" {
		return basicfont.Face7x13
	}
	var f *opentype.Font
	if cached, ok := fontCache.Load(path); ok {
		f = cached.(*opentype.Font)
	} else {
		data, err := os.ReadFile(path)
		if err != nil {
			return basicfont.Face7x13
		}
		parsed, err := opentype.Parse(data)
		if err != nil {
			return basicfont.Face7x13
		}
		fontCache.Store(path, parsed)
		f = parsed
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
func loadFonts(cfg Config) fontSet {
	body := resolveFamily(cfg.FontFamily, cfg.FontFile)
	code := resolveFamily(cfg.CodeFont, cfg.CodeFontFile)

	return fontSet{
		regular: loadFace(body.regular, cfg.FontSize),
		bold:    loadFace(body.bold, cfg.FontSize),
		italic:  loadFace(body.italic, cfg.FontSize),
		code:    loadFace(code.regular, cfg.CodeFontSize),
	}
}

func resolveFamily(name, regularOverride string) fontFamily {
	if regularOverride != "" {
		return fontFamily{
			regular:    regularOverride,
			bold:       regularOverride,
			italic:     regularOverride,
			boldItalic: regularOverride,
		}
	}
	body := systemFonts[name]
	if body.regular == "" {
		body = systemFonts["Helvetica"]
	}
	return body
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
		adv, ok := f.GlyphAdvance(r)
		if !ok {
			w += measure(f, fallbackGlyph(r))
			continue
		}
		w += int(adv >> 6)
	}
	return w
}
