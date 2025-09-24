package apple2helpers

import (
	"paleotronic.com/core/hires"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/settings"
	"paleotronic.com/octalyzer/video/font"
)

func LoadGraphicsFont(index int, fontid int) {
	if fontid >= 0 && fontid < len(settings.AuxFonts[index]) {
		fontName := settings.AuxFonts[index][fontid]
		f, err := font.LoadFromFile(fontName)
		if err == nil {
			PixelTextFont = f
		}
	}
}

func LoadNormalFontGlyphs() *font.DecalFont {

	this, err := font.LoadFromFile("fonts/osdfont.yaml")
	if err != nil {
		panic(err)
	}

	return this
}

var PixelTextWidth, PixelTextHeight int = 1, 1
var PixelTextColor int = 3
var PixelTextX, PixelTextY int = 1, 1
var PixelTextInverse bool = false
var PixelTextFont *font.DecalFont = LoadNormalFontGlyphs()

func DrawGraphicsText(ent interfaces.Interpretable, text string) {

	cp := GetVideoMode(ent)

	//log.Printf("DrawGraphicsText called in mode %s", cp)

	if cp == "TEXT" {
		return
	}

	page := GETGFX(ent, cp)
	if page == nil {
		panic("missing control interface")
	}

	var mx, my int
	var mc int

	switch cp {
	case "LOGR":
		mx, my, mc = 40, 48, 16
	case "DLGR":
		mx, my, mc = 80, 48, 16
	case "HGR1":
		mx, my, mc = 280, 192, 8
	case "HGR2":
		mx, my, mc = 280, 192, 8
	case "DHR1":
		mx, my, mc = 560, 192, 16
	case "DHR2":
		mx, my, mc = 560, 192, 16
	case "SHR1":
		mx, my, mc = page.HControl.(*hires.SuperHiResBuffer).GetWidth(), 200, 16
	default:
		return
	}

	tw := PixelTextFont.TextWidth
	th := PixelTextFont.TextHeight

	dx, dy := PixelTextX, PixelTextY

	var data []font.DecalPoint
	var ok bool

	for _, ch := range text {

		if PixelTextInverse {
			data, ok = PixelTextFont.GlyphsI[ch]
		} else {
			data, ok = PixelTextFont.GlyphsN[ch]
		}

		if !ok {
			if ch == 13 {
				dx = PixelTextX
				dy += (PixelTextHeight * th)
			}
			continue // skip any draw operation
		}

		// draw glyph here
		for _, dp := range data {
			for y := 0; y < PixelTextHeight; y++ {
				for x := 0; x < PixelTextWidth; x++ {

					zx := dx + (PixelTextWidth * dp.X) + x
					zy := dy + (PixelTextHeight * dp.Y) + y

					if zx >= mx || zx < 0 || zy < 0 || zy >= my {
						continue
					}

					if cp == "HGR1" || cp == "HGR2" || cp == "DHR1" || cp == "DHR2" || cp == "SHR1" {
						page.HControl.Plot(zx, zy, PixelTextColor%mc)
					} else if cp == "LOGR" {
						LOGRPlot40(ent, uint64(zx), uint64(zy), uint64(PixelTextColor%mc))
					} else if cp == "DLGR" {
						LOGRPlot80(ent, uint64(zx), uint64(zy), uint64(PixelTextColor%mc))
					}

				}
			}
		}

		dx += (PixelTextWidth * tw)
	}

	dy += (PixelTextHeight * th)
	PixelTextY = dy // update so we linefeed

}
