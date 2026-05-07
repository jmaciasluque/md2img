package md2img

import (
	"image"
	"image/draw"
	"image/png"
	"os"
)

// trimPNG reads a PNG, finds the bounding box of non-white content,
// and rewrites it cropped with padding.
func trimPNG(path string, dpi int, paddingMM float64) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		return err
	}

	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	// Padding in pixels.
	padding := int(float64(dpi) * paddingMM / 25.4)
	if padding < 1 {
		padding = 1
	}

	// Convert to NRGBA for direct Pix buffer access (no per-pixel interface dispatch).
	nrgba := image.NewNRGBA(bounds)
	draw.Draw(nrgba, bounds, img, bounds.Min, draw.Src)

	pix := nrgba.Pix
	stride := nrgba.Stride

	top := h
	bottom := 0
	left := w
	right := 0

	for y := 0; y < h; y++ {
		rowOff := y * stride
		for x := 0; x < w; x++ {
			off := rowOff + x*4
			r := pix[off]
			g := pix[off+1]
			b := pix[off+2]
			a := pix[off+3]
			// Skip fully transparent or near-white pixels.
			if a < 128 {
				continue
			}
			if r >= 0xF0 && g >= 0xF0 && b >= 0xF0 {
				continue
			}
			if y < top {
				top = y
			}
			if y > bottom {
				bottom = y
			}
			if x < left {
				left = x
			}
			if x > right {
				right = x
			}
		}
	}

	if top > bottom {
		return nil
	}

	// Apply padding, clamping to image bounds.
	x0 := bounds.Min.X + left - padding
	if x0 < bounds.Min.X {
		x0 = bounds.Min.X
	}
	y0 := bounds.Min.Y + top - padding
	if y0 < bounds.Min.Y {
		y0 = bounds.Min.Y
	}
	x1 := bounds.Min.X + right + padding + 1
	if x1 > bounds.Max.X {
		x1 = bounds.Max.X
	}
	y1 := bounds.Min.Y + bottom + padding + 1
	if y1 > bounds.Max.Y {
		y1 = bounds.Max.Y
	}

	cropped := nrgba.SubImage(image.Rect(x0, y0, x1, y1))

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	return png.Encode(out, cropped)
}
