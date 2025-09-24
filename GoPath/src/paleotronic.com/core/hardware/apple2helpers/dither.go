package apple2helpers

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"math"
	"strings"

	"github.com/nfnt/resize"
	"paleotronic.com/core/hires"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/fmt"
)

//import "os"

type ColorChannel struct {
	Data []float32
	W, H int
}

func NewColorChannel(w, h int) *ColorChannel {
	return &ColorChannel{Data: make([]float32, w*h), W: w, H: h}
}

func (cc *ColorChannel) Set(x, y int, c float32) {
	if (x < 0) || (x >= cc.W) || (y < 0) || (y >= cc.H) {
		//log.Printf("(%d,%d) not in bounds (%d, %d)\n", x, y, cc.W, cc.H)
		return
	}
	//log.Printf( "(%d,%d) -> %f\n", x, y, c )
	cc.Data[x+y*cc.W] = c
}

func (cc *ColorChannel) Get(x, y int) float32 {
	if (x < 0) || (x >= cc.W) || (y < 0) || (y >= cc.H) {
		//log.Printf("(%d,%d) not in bounds (%d, %d)\n", x, y, cc.W, cc.H)
		return 0
	}
	return cc.Data[x+y*cc.W]
}

func (cc *ColorChannel) GetClamp(x, y int, l, u float32) float32 {
	if (x < 0) || (x >= cc.W) || (y < 0) || (y >= cc.H) {
		//log.Printf("(%d,%d) not in bounds (%d, %d)\n", x, y, cc.W, cc.H)
		return l
	}
	v := cc.Data[x+y*cc.W]
	if v < l {
		v = l
	}
	if v > u {
		v = u
	}
	return v
}

type RelPos struct {
	X int
	Y int
}

type DiffusionMatrix struct {
	Quotients map[RelPos]float32
	Divisor   float32
}

// Apply the configured diffusion matrix to distribute the error amount on a channel
func (m *DiffusionMatrix) Apply(data *ColorChannel, x, y int, err float32) {

	var rx, ry int
	var value float32
	for loc, mult := range m.Quotients {
		rx, ry = x+loc.X, y+loc.Y // relative positions to current x, y
		value = data.Get(rx, ry)
		value += (err * mult) / m.Divisor
		data.Set(rx, ry, value)
	}

}

func NewDiffusionMatrix(div float32, matrix map[RelPos]float32) *DiffusionMatrix {
	return &DiffusionMatrix{Divisor: div, Quotients: matrix}
}

var Atkinson *DiffusionMatrix = NewDiffusionMatrix(
	8,
	map[RelPos]float32{
		{+1, +0}: 1,
		{+2, +0}: 1,
		{-1, +1}: 1,
		{+0, +1}: 1,
		{+1, +1}: 1,
		{+0, +2}: 1,
	},
)

var JaJuNi *DiffusionMatrix = NewDiffusionMatrix(
	48,
	map[RelPos]float32{
		{+1, +0}: 7,
		{+2, +0}: 5,
		{-2, +1}: 1,
		{+1, +1}: 5,
		{+0, +1}: 7,
		{+1, +1}: 5,
		{+2, +1}: 1,
		{-2, +2}: 1,
		{+1, +2}: 1,
		{+0, +2}: 5,
		{+1, +2}: 1,
		{+2, +2}: 1,
	},
)

var SuperFrog *DiffusionMatrix = NewDiffusionMatrix(
	32,
	map[RelPos]float32{
		{+1, +0}: 5,
		{+2, +0}: 3,
		{+3, +0}: 1,
		{-1, +1}: 3,
		{+0, +1}: 5,
		{+1, +1}: 3,
		{+2, +1}: 1,
		{+0, +2}: 3,
		{+1, +2}: 1,
	},
)

var FloydSteinberg *DiffusionMatrix = NewDiffusionMatrix(
	16,
	map[RelPos]float32{
		{+1, +0}: 7,
		{-1, +1}: 3,
		{+0, +1}: 5,
		{+1, +1}: 1,
	},
)

var Burkes *DiffusionMatrix = NewDiffusionMatrix(
	32,
	map[RelPos]float32{
		{+1, +0}: 8,
		{+2, +0}: 4,
		{-2, +1}: 2,
		{-1, +1}: 4,
		{+0, +1}: 8,
		{+1, +1}: 4,
		{+2, +1}: 2,
	},
)

var Sierra3Row *DiffusionMatrix = NewDiffusionMatrix(
	32,
	map[RelPos]float32{
		{+1, +0}: 5,
		{+2, +0}: 3,
		{-2, +1}: 2,
		{-1, +1}: 4,
		{+0, +1}: 5,
		{+1, +1}: 4,
		{+2, +1}: 2,
		{-1, +2}: 2,
		{+0, +2}: 3,
		{+1, +2}: 2,
	},
)

var Stucki *DiffusionMatrix = NewDiffusionMatrix(
	42,
	map[RelPos]float32{
		{+1, +0}: 8,
		{+2, +0}: 4,
		{-2, +1}: 2,
		{-1, +1}: 4,
		{+0, +1}: 8,
		{+1, +1}: 4,
		{+2, +2}: 2,
		{-2, +2}: 1,
		{-1, +2}: 2,
		{+0, +2}: 4,
		{+1, +2}: 2,
		{+2, +2}: 1,
	},
)

var Bayer4x4 [][]float32 = [][]float32{
	[]float32{1, 9, 3, 11},
	[]float32{13, 5, 15, 7},
	[]float32{4, 12, 2, 10},
	[]float32{16, 8, 14, 6},
}

var None *DiffusionMatrix = NewDiffusionMatrix(
	1,
	map[RelPos]float32{
		{+1, +0}: 0,
		{+2, +0}: 0,
		{-2, +1}: 0,
		{-1, +1}: 0,
		{+0, +1}: 0,
		{+1, +1}: 0,
		{+2, +1}: 0,
	},
)

func RGB12ToVideoColor(c int) *types.VideoColor {
	var r, g, b, a uint8
	r = uint8(c>>4) | 0x0f
	g = uint8(c&0xf0) | 0x0f
	b = (uint8(c&0x0f) << 4) | 0x0f
	a = uint8(0xff)
	if r == 0xf && g == 0xf && b == 0xf {
		a = 0x00
	}
	return &types.VideoColor{
		r, g, b, a, 0, 20,
	}
}

func HGRDither(ent interfaces.Interpretable, pngfile io.Reader, colors []int, gamma float32, matrix *DiffusionMatrix, perceptual bool, cliprect *image.Rectangle) error {

	cp := ent.GetCurrentPage()

	fmt.Println("========================================>", cp)

	page := GETGFX(ent, cp)
	if page == nil || (page.HControl == nil && page.Control == nil) {
		panic("missing control interface "+cp)
	}

	//page.HControl.Fill(0) // clear
	var maxColors = page.GetPaletteSize()
	if strings.HasPrefix(cp, "HGR") {
		maxColors = 8
	}

	colors = make([]int, 0)
	for i := 0; i < maxColors; i++ {
		colors = append(colors, i)
	}

	palette := types.NewVideoPalette()
	var cmap = make(map[int]int)

	if strings.HasPrefix(cp, "SHR") {
		pp := page.HControl.(*hires.SuperHiResBuffer).GetPalette()
		for i, c := range pp {
			cmap[palette.Size()] = i
			palette.Add(RGB12ToVideoColor(int(c)))
		}
	} else {
		for _, c := range colors {
			cmap[palette.Size()] = c
			palette.Add(page.GetPaletteColor(c))
		}
	}

	img, e := png.Decode(pngfile)
	if e != nil {
		return e
	}

	w := int(page.GetWidth())
	h := int(page.GetHeight())

	dw, dh := w, h
	xm := 1
	ym := 1

	transparent := map[int]int{
		0: 0,
	}

	sx, sy := 0, 0


	var useLo = false
	var useDblLo = false
	if strings.HasPrefix(cp, "SHR") {
		dw = page.HControl.(*hires.SuperHiResBuffer).GetWidth()
		xm = 1
	}
	if strings.HasPrefix(cp, "DHR") {
		dw = 140
		xm = 1
	}
	if cp == "LOGR" || cp == "LGR2" {
		dw = 40
		xm = 1
		_,_,_,ymax := page.GetBounds()
		dh = int(ymax+1)
		useLo = true
	}
	if cp == "DLGR" || cp == "DLG2" {
		dw = 80
		xm = 1
		_,_,_,ymax := page.GetBounds()
		dh = int(ymax+1)
		useDblLo = true
	}
	if strings.HasPrefix(cp, "HGR") {
		dw = 280
		xm = 1
		transparent[4] = 0
	}

	if cliprect != nil {
		// override
		dw = cliprect.Size().X
		dh = cliprect.Size().Y
		sx = cliprect.Min.X
		sy = cliprect.Min.Y
	}

	nimg := resize.Resize(uint(dw), uint(dh), img, resize.Bilinear)
	pimg := image.NewRGBA(image.Rect(0, 0, dw, dh))
	draw.Draw(pimg, pimg.Bounds(), nimg, image.Pt(0, 0), draw.Over)

	//	var size int = w * h1
	var r *ColorChannel = NewColorChannel(dw, dh)
	var g *ColorChannel = NewColorChannel(dw, dh)
	var b *ColorChannel = NewColorChannel(dw, dh)
	//	var ci int

	var tmap = make([]bool, dw*dh)

	// parse into floats x over y first because we can dither better in woz mode
	for x := 0; x < dw; x++ {
		for y := 0; y < dh; y++ {
			pxcol := pimg.At(x, y) //pixel from image (color.Color)
			rc, gc, bc, ac := pxcol.RGBA()
			r.Set(x, y, float32(math.Pow(float64(rc), float64(gamma))))
			g.Set(x, y, float32(math.Pow(float64(gc), float64(gamma))))
			b.Set(x, y, float32(math.Pow(float64(bc), float64(gamma))))
			tmap[x+dw*y] = (ac == 0)
		}
	}

	// process
	var trnsp bool
	for x := 0; x < dw; x++ {
		for y := 0; y < dh; y++ {

			rcc, gcc, bcc := r.GetClamp(x, y, 0, 65535), g.GetClamp(x, y, 0, 65535), b.GetClamp(x, y, 0, 65535)

			trnsp = tmap[x+y*dw]

			rc := palette.GetMatch(
				color.RGBA{R: uint8(rcc / 256), G: uint8(gcc / 256), B: uint8(bcc / 256), A: 255},
				perceptual,
			)

			ccc := cmap[rc]
			if newrc, ok := transparent[ccc]; ok {
				ccc = newrc
			}

			// plot pixel
			if !trnsp {
				if useLo {					
					xx := x*xm+sx*xm
					yy := y*ym+sy*ym
					c := uint64(ccc)
					
					px := int((xx * 2) % 80)
					py := int(((yy / 2) * 2) % 48)

					v := page.Control.GetValueXY(px, py)
					c0 := v & 0xf
					c1 := (v & 0xf0) >> 4
					switch y % 2 {
					case 0:
						c0 = c
					case 1:
						c1 = c
					}
					v = (v & 0xffff0000) | (c1 << 4) | c0
					page.Control.PutValueXY(px, py, v)
				} else if useDblLo {
					xx := x*xm+sx*xm
					yy := y*ym+sy*ym
					c := uint64(ccc)
					
					px := int(xx % 80)
					py := int(((yy / 2) * 2) % 48)

					v := page.Control.GetValueXY(px, py)
					cv := c
					if (px % 2) == 0 {
						cv = uint64(ror4bit(int(cv)))
					}
					c0 := v & 0xf
					c1 := (v & 0xf0) >> 4
					switch y % 2 {
					case 0:
						c0 = cv
					case 1:
						c1 = cv
					}
					v = (v & 0xffff0000) | (c1 << 4) | c0
					page.Control.PutValueXY(px, py, v)				
				} else {
					page.HControl.Plot(x*xm+sx*xm, y*ym+sy*ym, ccc)
				}
			}

			rrc, grc, brc, _ := palette.Get(rc).ToRGBA()

			// calc errors
			rerr, gerr, berr := rcc-float32(rrc), gcc-float32(grc), bcc-float32(brc)

			r.Set(x, y, float32(rrc))
			g.Set(x, y, float32(grc))
			b.Set(x, y, float32(brc))

			matrix.Apply(r, x, y, rerr)
			matrix.Apply(g, x, y, gerr)
			matrix.Apply(b, x, y, berr)

		}

	}

	settings.SHRFrameForce[ent.GetMemIndex()] = true

	return nil
}

func Clamp(in float32, l, u float32) float32 {
	if in < l {
		in = l
	}
	if in > u {
		in = u
	}
	return in
}

func HGRDitherBayer4x4(ent interfaces.Interpretable, pngfile io.Reader, colors []int, gamma float32, perceptual bool) error {
	cp := ent.GetCurrentPage()
	page := GETGFX(ent, cp)
	if page == nil || page.HControl == nil {
		panic("missing control interface")
	}

	page.HControl.Fill(0) // clear

	var maxColors = page.GetPaletteSize()
	if strings.HasPrefix(cp, "HGR") {
		maxColors = 8
	}

	colors = make([]int, 0)
	for i := 0; i < maxColors; i++ {
		colors = append(colors, i)
	}

	palette := types.NewVideoPalette()
	var cmap = make(map[int]int)
	for _, c := range colors {
		cmap[palette.Size()] = c
		palette.Add(page.GetPaletteColor(c))
	}

	img, e := png.Decode(pngfile)
	if e != nil {
		return e
	}

	w := int(page.GetWidth())
	h := int(page.GetHeight())
	
	var useLo = false
	var useDblLo = false
	if cp == "LOGR" || cp == "LGR2" {
		_,_,_,ymax := page.GetBounds()
		h = int(ymax+1)
		useLo = true
	}
	if cp == "DLGR" || cp == "DLG2" {
		_,_,_,ymax := page.GetBounds()
		h = int(ymax+1)
		useDblLo = true
	}

	nimg := resize.Resize(uint(w), uint(h), img, resize.Bilinear)
	pimg := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(pimg, pimg.Bounds(), nimg, image.Pt(0, 0), draw.Over)

	//	var size int = w * h
	var r *ColorChannel = NewColorChannel(w, h)
	var g *ColorChannel = NewColorChannel(w, h)
	var b *ColorChannel = NewColorChannel(w, h)
	//	var ci int

	// parse into floats
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {

			pxcol := pimg.At(x, y) //pixel from image (color.Color)
			rc, gc, bc, _ := pxcol.RGBA()
			r.Set(x, y, float32(math.Pow(float64(rc), float64(gamma))))
			g.Set(x, y, float32(math.Pow(float64(gc), float64(gamma))))
			b.Set(x, y, float32(math.Pow(float64(bc), float64(gamma))))
		}
	}

	// process
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {

			tv := Bayer4x4[x%4][y%4] / 17

			rcc, gcc, bcc := r.GetClamp(x, y, 0, 65535), g.GetClamp(x, y, 0, 65535), b.GetClamp(x, y, 0, 65535)

			rcc = Clamp(rcc+(rcc*tv), 0, 65535)
			gcc = Clamp(gcc+(gcc*tv), 0, 65535)
			bcc = Clamp(bcc+(bcc*tv), 0, 65535)

			rc := palette.GetMatch(
				color.RGBA{R: uint8(rcc / 256), G: uint8(gcc / 256), B: uint8(bcc / 256), A: 255},
				perceptual,
			)

			// plot pixel
			if useLo {					
				c := uint64(cmap[rc])
				
				px := int((x * 2) % 80)
				py := int(((y / 2) * 2) % 48)

				v := page.Control.GetValueXY(px, py)
				c0 := v & 0xf
				c1 := (v & 0xf0) >> 4
				switch y % 2 {
				case 0:
					c0 = c
				case 1:
					c1 = c
				}
				v = (v & 0xffff0000) | (c1 << 4) | c0
				page.Control.PutValueXY(px, py, v)
			} else if useDblLo {
				c := uint64(cmap[rc])
				
				px := int(x % 80)
				py := int(((y / 2) * 2) % 48)

				v := page.Control.GetValueXY(px, py)
				cv := c
				if (px % 2) == 0 {
					cv = uint64(ror4bit(int(cv)))
				}
				c0 := v & 0xf
				c1 := (v & 0xf0) >> 4
				switch y % 2 {
				case 0:
					c0 = cv
				case 1:
					c1 = cv
				}
				v = (v & 0xffff0000) | (c1 << 4) | c0
				page.Control.PutValueXY(px, py, v)				
			} else {
				page.HControl.Plot(x, y, cmap[rc])
			}

		}

	}

	settings.SHRFrameForce[ent.GetMemIndex()] = true

	return nil
}

func DitherImage(palette *types.VideoPalette, pimg image.Image, gamma float32, matrix *DiffusionMatrix, perceptual bool) image.Image {

	dw := int(pimg.Bounds().Max.X)
	dh := int(pimg.Bounds().Max.Y)

	newimage := image.NewRGBA(pimg.Bounds())

	//	var size int = w * h
	var r *ColorChannel = NewColorChannel(dw, dh)
	var g *ColorChannel = NewColorChannel(dw, dh)
	var b *ColorChannel = NewColorChannel(dw, dh)
	//	var ci int

	// parse into floats x over y first because we can dither better in woz mode
	for x := 0; x < dw; x++ {
		for y := 0; y < dh; y++ {
			pxcol := pimg.At(x, y) //pixel from image (color.Color)
			rc, gc, bc, _ := pxcol.RGBA()
			r.Set(x, y, float32(math.Pow(float64(rc), float64(gamma))))
			g.Set(x, y, float32(math.Pow(float64(gc), float64(gamma))))
			b.Set(x, y, float32(math.Pow(float64(bc), float64(gamma))))
		}
	}

	// process
	for x := 0; x < dw; x++ {

		for y := 0; y < dh; y++ {

			rcc, gcc, bcc := r.GetClamp(x, y, 0, 65535), g.GetClamp(x, y, 0, 65535), b.GetClamp(x, y, 0, 65535)

			rc := palette.GetMatch(
				color.RGBA{R: uint8(rcc / 256), G: uint8(gcc / 256), B: uint8(bcc / 256), A: 255},
				perceptual,
			)

			// plot pixel
			newimage.Set(x, y, palette.Get(rc).ToColorRGBA())

			rrc, grc, brc, _ := palette.Get(rc).ToRGBA()

			// calc errors
			rerr, gerr, berr := rcc-float32(rrc), gcc-float32(grc), bcc-float32(brc)

			r.Set(x, y, float32(rrc))
			g.Set(x, y, float32(grc))
			b.Set(x, y, float32(brc))

			matrix.Apply(r, x, y, rerr)
			matrix.Apply(g, x, y, gerr)
			matrix.Apply(b, x, y, berr)

		}

	}

	return newimage
}
