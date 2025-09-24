package accelimage

import (
	"image"
	"image/color"
	"image/draw"
)

/*
accelimage
==========
A simple rendering library for fast image writes.
*/
var TintCache map[*image.RGBA]map[color.RGBA]*image.RGBA

func init() {
	TintCache = make(map[*image.RGBA]map[color.RGBA]*image.RGBA)
}

func GetTintedVariant(i *image.RGBA, c color.RGBA) *image.RGBA {

	if i == nil {
		return nil
	}

	cache, ok := TintCache[i]
	if !ok {
		cache = make(map[color.RGBA]*image.RGBA)
		TintCache[i] = cache
	}
	ti, ok := cache[c]
	if !ok {
		ti = image.NewRGBA(i.Bounds())
		cache[c] = ti
		tr := float32(c.R) / 255
		tg := float32(c.G) / 255
		tb := float32(c.B) / 255
		ta := float32(c.A) / 255
		for j := 0; j < len(ti.Pix); j++ {
			switch j % 4 {
			case 0:
				ti.Pix[j] = uint8(float32(i.Pix[j]) * tr)
			case 1:
				ti.Pix[j] = uint8(float32(i.Pix[j]) * tg)
			case 2:
				ti.Pix[j] = uint8(float32(i.Pix[j]) * tb)
			case 3:
				ti.Pix[j] = uint8(float32(i.Pix[j]) * ta)
			}
		}
	}

	return ti

}

type Filter func(ax, ay int, c color.RGBA) color.RGBA

func FillRGBAWithFilter(dest *image.RGBA, destrect image.Rectangle, c color.RGBA, f Filter) {

	destrect = dest.Bounds().Intersect(destrect)

	if destrect.Max.X > dest.Bounds().Max.X {
		return
	}

	for y := destrect.Min.Y; y < destrect.Max.Y; y++ {
		ptr := dest.PixOffset(destrect.Min.X, y)
		for x := 0; x < destrect.Dx(); x++ {
			fc := f(destrect.Min.X+x, y, c)
			dest.Pix[ptr+0] = fc.R
			dest.Pix[ptr+1] = fc.G
			dest.Pix[ptr+2] = fc.B
			dest.Pix[ptr+3] = fc.A
			ptr += 4
		}
	}

}

func FillRGBA(dest *image.RGBA, destrect image.Rectangle, c color.RGBA) {

	destrect = dest.Bounds().Intersect(destrect)

	if destrect.Max.X > dest.Bounds().Max.X {
		return
	}

	for y := destrect.Min.Y; y < destrect.Max.Y; y++ {
		ptr := dest.PixOffset(destrect.Min.X, y)
		for x := 0; x < destrect.Dx(); x++ {
			dest.Pix[ptr+0] = c.R
			dest.Pix[ptr+1] = c.G
			dest.Pix[ptr+2] = c.B
			dest.Pix[ptr+3] = c.A
			ptr += 4
		}
	}

}

// DrawImage puts src at point
func DrawImageRGBA(dest *image.RGBA, point image.Point, src *image.RGBA) {

	destrect := src.Bounds().Add(point)
	destrect = dest.Bounds().Intersect(destrect)

	for y := destrect.Min.Y; y < destrect.Max.Y; y++ {
		ptr := dest.PixOffset(destrect.Min.X, y)
		sptr := src.PixOffset(0, y-destrect.Min.Y)
		for x := 0; x < destrect.Dx(); x++ {
			dest.Pix[ptr+0] = src.Pix[sptr+0]
			dest.Pix[ptr+1] = src.Pix[sptr+1]
			dest.Pix[ptr+2] = src.Pix[sptr+2]
			dest.Pix[ptr+3] = src.Pix[sptr+3]
			ptr += 4
			sptr += 4
			if ptr >= len(dest.Pix) {
				break
			}
		}
	}

}

// DrawImage puts src at point
func DrawImageRGBAAlpha(dest *image.RGBA, point image.Point, src *image.RGBA, threshold uint8) {

	destrect := src.Bounds().Add(point)
	destrect = dest.Bounds().Intersect(destrect)

	for y := destrect.Min.Y; y < destrect.Max.Y; y++ {
		ptr := dest.PixOffset(destrect.Min.X, y)
		sptr := src.PixOffset(0, y-destrect.Min.Y)
		for x := 0; x < destrect.Dx(); x++ {
			alpha := src.Pix[sptr+3]
			if alpha > threshold {
				dest.Pix[ptr+0] = src.Pix[sptr+0]
				dest.Pix[ptr+1] = src.Pix[sptr+1]
				dest.Pix[ptr+2] = src.Pix[sptr+2]
				dest.Pix[ptr+3] = src.Pix[sptr+3]
			}
			ptr += 4
			sptr += 4
		}
	}

}

func DrawImageRGBAOffset(dest *image.RGBA, destrect image.Rectangle, src *image.RGBA, point image.Point) {

	destrect = dest.Bounds().Intersect(destrect)

	if destrect.Max.X > dest.Bounds().Max.X {
		return
	}

	for y := destrect.Min.Y; y < destrect.Max.Y; y++ {
		ptr := dest.PixOffset(destrect.Min.X, y)
		sptr := src.PixOffset(point.X, y-destrect.Min.Y+point.Y)
		for x := 0; x < destrect.Dx(); x++ {

			dest.Pix[ptr+0] = src.Pix[sptr+0]
			dest.Pix[ptr+1] = src.Pix[sptr+1]
			dest.Pix[ptr+2] = src.Pix[sptr+2]
			dest.Pix[ptr+3] = src.Pix[sptr+3]
			ptr += 4
			sptr += 4
		}
	}

}

// DrawImage puts src at point
func DrawImageRGBATint(dest *image.RGBA, point image.Point, src *image.RGBA, tint color.RGBA) {

	tr := float32(tint.R) / 255
	tg := float32(tint.G) / 255
	tb := float32(tint.B) / 255
	ta := float32(tint.A) / 255

	destrect := src.Bounds().Add(point)
	destrect = dest.Bounds().Intersect(destrect)

	if destrect.Max.X > dest.Bounds().Max.X {
		return
	}

	for y := destrect.Min.Y; y < destrect.Max.Y; y++ {
		ptr := dest.PixOffset(destrect.Min.X, y)
		sptr := src.PixOffset(0, y-destrect.Min.Y)
		for x := 0; x < destrect.Dx(); x++ {
			dest.Pix[ptr+0] = uint8(float32(src.Pix[sptr+0]) * tr)
			dest.Pix[ptr+1] = uint8(float32(src.Pix[sptr+1]) * tg)
			dest.Pix[ptr+2] = uint8(float32(src.Pix[sptr+2]) * tb)
			dest.Pix[ptr+3] = uint8(float32(src.Pix[sptr+3]) * ta)
			ptr += 4
			sptr += 4
		}
	}

}

func DrawImageRGBATintAlpha(dest *image.RGBA, point image.Point, src *image.RGBA, tint color.RGBA, threshold uint8) {

	tr := float32(tint.R) / 255
	tg := float32(tint.G) / 255
	tb := float32(tint.B) / 255
	ta := float32(tint.A) / 255

	destrect := src.Bounds().Add(point)
	destrect = dest.Bounds().Intersect(destrect)

	if destrect.Max.X > dest.Bounds().Max.X {
		return
	}

	var alpha uint8

	for y := destrect.Min.Y; y < destrect.Max.Y; y++ {
		ptr := dest.PixOffset(destrect.Min.X, y)
		sptr := src.PixOffset(0, y-destrect.Min.Y)
		for x := 0; x < destrect.Dx(); x++ {
			alpha = src.Pix[sptr+3]
			if alpha >= threshold {
				dest.Pix[ptr+0] = uint8(float32(src.Pix[sptr+0]) * tr)
				dest.Pix[ptr+1] = uint8(float32(src.Pix[sptr+1]) * tg)
				dest.Pix[ptr+2] = uint8(float32(src.Pix[sptr+2]) * tb)
				dest.Pix[ptr+3] = uint8(float32(alpha) * ta)
			}
			ptr += 4
			sptr += 4
		}
	}

}

// Make sure image is *image.RGBA -- might have to conver it or just return
// it depending on underlying types
func ImageRGBA(src image.Image) *image.RGBA {
	switch src.(type) {
	case *image.RGBA:
		return src.(*image.RGBA)
	default:
		b := src.Bounds()
		dst := image.NewRGBA(b)
		draw.Draw(dst, b, src, b.Min, draw.Src)
		return dst
	}
	return nil
}
