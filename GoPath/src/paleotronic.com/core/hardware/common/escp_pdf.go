package common

import (
	"bytes"
	"image"
	"image/color"
	"time"

	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/interfaces"

	"paleotronic.com/files"
	"paleotronic.com/spooler"

	"paleotronic.com/fmt"

	"github.com/signintech/gopdf"
)

type R struct {
	x, y, w, h float64
}

type ElementType int

const (
	etRect  ElementType = 0
	etText  ElementType = 1
	etPoint ElementType = 2
)

type Element struct {
	color   color.CMYK
	kind    ElementType
	payload *ElementText
}

func (e *Element) render(rr R, pdf *gopdf.GoPdf) {
	r, g, b := color.CMYKToRGB(e.color.C, e.color.M, e.color.Y, e.color.K)
	if e.kind == etPoint {
		pdf.SetFillColor(r, g, b)
		pdf.Oval(
			rr.x,
			rr.y,
			rr.x+rr.w/6,
			rr.y+rr.h/6,
		)
	} else if e.kind == etRect {
		pdf.SetFillColor(r, g, b)
		pdf.RectFromUpperLeftWithStyle(
			rr.x,
			rr.y,
			rr.w,
			rr.h,
			"F",
		)
	} else if e.kind == etText {
		pdf.SetTextColor(r, g, b)
		cell := &gopdf.Rect{W: rr.w, H: rr.h}
		pdf.SetFont(e.payload.fontName, "", e.payload.size)
		pdf.SetX(rr.x)
		pdf.SetY(rr.y)
		pdf.CellWithOption(cell, e.payload.text, e.payload.options)
	}
}

type ElementText struct {
	text     string
	size     int
	fontName string
	options  gopdf.CellOption
}

type PDFOutput struct {
	w, h           float64
	hdpi, vdpi     float64
	pw, ph         float64
	page           int
	pdfcount       int
	plotted        bool
	valid          bool
	image          *image.RGBA
	pdf            *gopdf.GoPdf
	b              *bytes.Buffer
	vscale, hscale float64
	fonts          map[string]bool
	c              map[R]*Element
	list           []R
	inpage         bool
}

func (p *PDFOutput) LetterAt(x, y float64, ch rune, style fxFontStyle, c fxColor) {

	sx, sy := (x * p.hscale), (y * p.vscale)

	cmyk := p.fxColorCMYK(c)

	fontname := style.String()
	if !p.fonts[fontname] {
		p.pdf.AddTTFFontByReader(
			fontname,
			getFXFontData(fontname),
		)
		p.fonts[fontname] = true
	}

	if style.doubleStrike == true || style.style == fxStyleBold || style.style == fxStyleBoldItalic {
		cmyk = mixCMYK(cmyk, cmyk)
	}

	size := 12
	if style.imgWriter {
		size = 14
	}

	cell := &gopdf.Rect{
		H: 12,
		W: float64(72) / style.Cpi(),
	}

	options := gopdf.CellOption{}
	if style.underline == true {
		options.Border = gopdf.Bottom
	}

	if style.submode == fxSubModeSubscript {
		size = 9
		options.Align = gopdf.Bottom
	}
	if style.submode == fxSubModeSuperscript {
		size = 9
		options.Align = gopdf.Top
	}

	// p.pdf.SetFont(fontname, "", size)
	// p.pdf.SetX(sx * 72)
	// p.pdf.SetY(sy * 72)
	// p.pdf.CellWithOption(cell, string(ch), options)
	// if style.doubleStrike == true {
	// 	p.pdf.SetX((sx * 72) + 1)
	// 	p.pdf.CellWithOption(cell, string(ch), options)
	// }

	p.logText(
		R{sx * 72, sy * 72, cell.W, cell.H},
		string(ch),
		fontname,
		size,
		cmyk,
		options,
	)
	if style.doubleStrike == true {
		p.logText(
			R{sx * 72, (sy * 72) + 1, cell.W, cell.H},
			string(ch),
			fontname,
			size,
			cmyk,
			options,
		)
	}

	p.plotted = true
	p.valid = true

}

func (p *PDFOutput) logRoundPoint(rect R, c color.CMYK) {
	data, ok := p.c[rect]
	if ok && data.kind == etRect {
		data.color = mixCMYK(data.color, c)
	} else {
		p.c[rect] = &Element{kind: etRect, color: c}
		p.list = append(p.list, rect)
	}
}

func (p *PDFOutput) logText(rect R, text string, fontName string, size int, c color.CMYK, options gopdf.CellOption) {
	data, ok := p.c[rect]
	if ok && data.kind == etText && fontName == data.payload.fontName && data.payload.size == size {
		data.color = mixCMYK(data.color, c)
	} else {
		p.c[rect] = &Element{
			color: c,
			kind:  etText,
			payload: &ElementText{
				options:  options,
				fontName: fontName,
				text:     text,
				size:     size,
			},
		}
		p.list = append(p.list, rect)
	}
}

func (p *PDFOutput) fxColorCMYK(c fxColor) color.CMYK {
	switch c {
	case fxColorBlack:
		return color.CMYK{0, 0, 0, 192}
	case fxColorCyan:
		return color.CMYK{192, 61, 0, 0}
	case fxColorMagenta:
		return color.CMYK{0, 192, 96, 0}
	case fxColorYellow:
		return color.CMYK{0, 0, 192, 0}
	case fxColorRed:
		return color.CMYK{0, 192, 192, 0}
	case fxColorViolet:
		return color.CMYK{192, 192, 0, 0}
	case fxColorGreen:
		return color.CMYK{192, 0, 192, 0}
	case fxColorOrange:
		return color.CMYK{0, 125, 192, 0}
	case fxColorBlue:
		return color.CMYK{192, 192, 0, 0}
	}
	return color.CMYK{0, 0, 0, 192}
}

func (p *PDFOutput) PlotDots(x, y float64, chunk []byte, hdpi, vdpi float64, c fxColor) {

	sx, sy := (x * p.hscale), (y * p.vscale)

	dw, dh := float64(p.hscale)/hdpi, float64(p.vscale)/vdpi

	//fmt.RPrintf("sx, sy = %f, %f\n", sx, sy)
	//fmt.RPrintf("dw, dh = %f, %f\n", dw, dh)

	for bit := 0; bit < 8; bit++ {
		for col, v := range chunk {

			xx := sx + float64(col)*dw
			yy := sy + float64(7-bit)*dh

			if v&(1<<uint(bit)) != 0 {
				//p.pdf.Rect(xx, yy, dw, dh, "F")
				cmyk := p.fxColorCMYK(c)

				p.logRoundPoint(R{xx * 72, yy * 72, dw * 72, dh * 72}, cmyk)

				p.valid = true
				p.plotted = true

			}
		}
	}

}

func (p *PDFOutput) SetPageSize(w, h float64, hdpi, vdpi float64) {

	p.pw = w * hdpi
	p.ph = h * vdpi
	p.w = w
	p.h = h
	p.hdpi = hdpi
	p.vdpi = vdpi

	p.valid = false

	p.vscale = 1
	p.hscale = 1

	p.fonts = make(map[string]bool)
	p.pdf = &gopdf.GoPdf{}
	p.pdf.Start(
		gopdf.Config{
			PageSize: gopdf.Rect{
				W: w * 72,
				H: h * 72, // convert inches to points
			},
		},
	)

	p.c = make(map[R]*Element)
	p.list = make([]R, 0)

}

func (p *PDFOutput) FinalizePage() {

	if p.plotted {

		p.pdf.AddPage()

		for _, r := range p.list {
			p.c[r].render(r, p.pdf)
		}

		p.c = make(map[R]*Element)
		p.list = make([]R, 0)

		p.page++

		p.plotted = false
		p.inpage = false
	}

}

func (p *PDFOutput) NewPage() {
}

func (p *PDFOutput) Flush(ent interfaces.Interpretable) {

	p.FinalizePage()

	if p.valid {

		p.pdfcount++

		t := time.Now()

		filename := fmt.Sprintf("print-%s.pdf", t.Format("20060102_150405"))

		path := files.GetUserDirectory(files.BASEDIR + "/MyPrints/" + filename)

		p.pdf.WritePdf(path)

		fmt.RPrintf("Wrote %s...\n", filename)

		apple2helpers.OSDShow(ent, filename+" -> MyPrints")

		if settings.PDFSpool {
			sp := spooler.NewSpooler()
			sp.SpoolPDF(path)
		}

		p.SetPageSize(p.w, p.h, p.hdpi, p.vdpi)

		p.plotted = false
		p.valid = false
	}
}
