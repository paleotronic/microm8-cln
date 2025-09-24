package common

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"

	"paleotronic.com/core/interfaces"

	"paleotronic.com/fmt"
	"paleotronic.com/octalyzer/video/font"
)

type PNGOutput struct {
	w, h           float64
	hdpi, vdpi     float64
	pw, ph         float64
	page           int
	plotted        bool
	image          *image.RGBA
	vscale, hscale float64
}

func (p *PNGOutput) LetterAt(x, y float64, ch rune, style fxFontStyle, color fxColor) {

	hcpi := style.Cpi()

	hdpi := hcpi * float64(PixelTextFont.TextWidth)
	vdpi := float64(60)

	sx, sy := (x*p.hscale/p.w)*p.pw, (y*p.vscale/p.h)*p.ph
	dw, dh := p.hscale*p.pw/(p.w*hdpi), p.vscale*p.ph/(p.h*vdpi)

	p.drawGraphicsText(ch, false, sx, sy, dw, dh)

}

func (p *PNGOutput) PlotDots(x, y float64, chunk []byte, hdpi, vdpi float64, c fxColor) {

	sx, sy := (x*p.hscale/p.w)*p.pw, (y*p.vscale/p.h)*p.ph

	dw, dh := p.hscale*p.pw/(p.w*hdpi), p.vscale*p.ph/(p.h*vdpi)

	fmt.RPrintf("sx, sy = %f, %f\n", sx, sy)
	fmt.RPrintf("dw, dh = %f, %f\n", dw, dh)

	for col, v := range chunk {
		for bit := 0; bit < 8; bit++ {
			xx := sx + float64(col)*dw
			yy := sy + float64(7-bit)*dh
			//

			if v&(1<<uint(bit)) != 0 {
				dr := image.Rect(
					int(xx),
					int(yy),
					int(xx+dw),
					int(yy+dh),
				)
				fmt.RPrintf("Point at %d, %d\n", int(xx), int(yy))
				draw.Draw(p.image, dr, image.NewUniform(color.RGBA{0, 0, 0, 255}), image.ZP, draw.Src)
			}
		}
	}

	p.plotted = true
}

func (p *PNGOutput) SetPageSize(w, h float64, hdpi, vdpi float64) {

	p.pw = w * hdpi
	p.ph = h * vdpi
	p.w = w
	p.h = h
	p.hdpi = hdpi
	p.vdpi = vdpi
	p.vscale = 0.82
	p.hscale = 1

}

func (p *PNGOutput) FinalizePage() {

	if p.plotted {
		filename := fmt.Sprintf("page%.3d.png", p.page)
		f, err := os.Create(filename)
		if err != nil {
			return
		}
		defer f.Close()
		err = png.Encode(f, p.image)
		fmt.RPrintf("Writing page to %s...\n", filename)
		p.NewPage()
	}

}

func (p *PNGOutput) NewPage() {

	p.page++
	p.plotted = false
	dr := image.Rect(
		0, 0,
		int(p.pw), int(p.ph),
	)
	p.image = image.NewRGBA(
		dr,
	)
	draw.Draw(p.image, dr, image.NewUniform(color.RGBA{255, 255, 255, 255}), image.ZP, draw.Src)
}

func (p *PNGOutput) Flush(ent interfaces.Interpretable) {

}

func LoadNormalFontGlyphs() *font.DecalFont {
	this := font.NewDecalFont(
		&font.Font{
			Name:    "Apple ][",
			Width:   7,
			Height:  8,
			Inverse: true,
			Sources: []font.FontRange{
				font.FontRange{Source: "fonts/Pr21Normal_0.png", Start: 32, End: 288},
				font.FontRange{Source: "fonts/Pr21Alt_0.png", Start: 1024 + 32, End: 1024 + 512},
			},
		},
	)

	return this
}

var PixelTextFont *font.DecalFont = LoadNormalFontGlyphs()

func (p *PNGOutput) drawGraphicsText(ch rune, inverse bool, sx, sy, dw, dh float64) {

	var data []font.DecalPoint
	var ok bool

	if inverse {
		data, ok = PixelTextFont.GlyphsI[ch]
	} else {
		data, ok = PixelTextFont.GlyphsN[ch]
	}

	if !ok {
		return
	}

	// draw glyph here
	for _, dp := range data {

		xx := sx + float64(dp.X)*dw
		yy := sy + float64(dp.Y)*dh
		dr := image.Rect(
			int(xx),
			int(yy),
			int(xx+dw),
			int(yy+dh),
		)

		fmt.RPrintf("Point at %d, %d\n", int(xx), int(yy))
		draw.Draw(p.image, dr, image.NewUniform(color.RGBA{0, 0, 0, 255}), image.ZP, draw.Src)
		p.plotted = true

	}

}
