package md2img

import (
	"image"
	"image/draw"
	"image/png"
	"os"
)

// writeTrimmedPNG crops an image to its content bounds and writes it.
func writeTrimmedPNG(img image.Image, path string, dpi int, paddingMM float64) error {
	bounds := img.Bounds()

	// Convert to NRGBA for pixel access.
	nrgba := image.NewNRGBA(bounds)
	draw.Draw(nrgba, bounds, img, bounds.Min, draw.Src)

	cropped := trimImage(nrgba, dpi, paddingMM)

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()
	return png.Encode(out, cropped)
}

// trimImage scans for non-white content and crops with padding.
func trimImage(nrgba *image.NRGBA, dpi int, paddingMM float64) image.Image {
	bounds := nrgba.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	padding := int(float64(dpi) * paddingMM / 25.4)
	if padding < 1 {
		padding = 1
	}

	pix := nrgba.Pix
	stride := nrgba.Stride

	top, bottom := h, 0
	left, right := w, 0

	for y := 0; y < h; y++ {
		rowOff := y * stride
		for x := 0; x < w; x++ {
			off := rowOff + x*4
			r := pix[off]
			g := pix[off+1]
			b := pix[off+2]
			a := pix[off+3]
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
		return nrgba
	}

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

	return nrgba.SubImage(image.Rect(x0, y0, x1, y1))
}
