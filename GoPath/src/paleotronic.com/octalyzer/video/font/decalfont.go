package font

import (
	"io/ioutil"

	//	"os"
	"io"
	"math"

	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"

	// "paleotronic.com/glumby"
	"paleotronic.com/log"
	"paleotronic.com/octalyzer/assets"

	"github.com/go-gl/mathgl/mgl32"

	yaml "gopkg.in/yaml.v2"
)

const (
	SW = 560 * 2
	SH = 384
)

var NormalFont *DecalFont

type Font struct {
	Name    string
	Height  int
	Width   int
	Inverse bool
	Sources []FontRange
}

type FontRange struct {
	Start  int
	End    int
	Source string
	XPad   int
	YPad   int
}

type DecalFont struct {
	Border       int
	GlyphWidth   int
	DotHeight    float32
	ScaleX       float32
	SpacingH     int
	GlyphHeight  int
	DotWidth     float32
	ScaleY       float32
	Pix          image.Image
	TextHeight   int
	ScaleZ       float32
	cpl          int
	TextWidth    int
	SpacingV     int
	ScreenWidth  int
	ScreenHeight int
	baseLine     int
	//used        []*glumby.Mesh
	batchCount  int
	InverseLoad bool
	GlyphsN     map[rune][]DecalPoint
	GlyphsI     map[rune][]DecalPoint
}

func LoadPNG(filename string) (image.Image, error) {

	data, err := assets.Asset(filename)
	if err != nil {
		return nil, err
	}
	r := bytes.NewBuffer(data)

	return png.Decode(r)
}

func fillRGBA(img *image.RGBA, c color.RGBA) {
	draw.Draw(img, img.Bounds(), &image.Uniform{c}, image.ZP, draw.Src)
}

func (this *DecalFont) GetScaleZ() float32 {
	return this.ScaleZ
}

type DecalPoint struct {
	X, Y int
}

func (this *DecalFont) ReadPoints(pix image.Image, ox, oy int, w, h int) []DecalPoint {

	glyphdata := make([]DecalPoint, 0)

	for r := 0; r < h; r++ {
		for c := 0; c < w; c++ {

			cc := pix.At(ox+c, oy+r)
			_, _, _, aa := cc.RGBA()

			dotset := (aa >= 32768)

			// autoinvert
			if this.InverseLoad {
				dotset = !dotset
			}

			if dotset {
				glyphdata = append(glyphdata, DecalPoint{X: c, Y: r})
			}
		}
	}

	//tx := glumby.NewTextureFromRGBA(sb)

	return glyphdata
}

func (this *DecalFont) LoadGlyphs(start int, end int) {

	// Normal
	this.InverseLoad = false
	for idx := start; idx <= end; idx++ {

		log.Printf("%d,", idx)

		v := idx - start
		ox := (v % this.cpl) * (this.TextWidth + this.SpacingH)
		oy := (v / this.cpl) * (this.TextHeight + this.SpacingV)

		g := this.ReadPoints(this.Pix, ox, oy, this.TextWidth, this.TextHeight)

		this.GlyphsN[rune(idx)] = g

	}

	// Inverted
	this.InverseLoad = true
	for idx := start; idx <= end; idx++ {
		v := idx - start
		ox := (v % this.cpl) * (this.TextWidth + this.SpacingH)
		oy := (v / this.cpl) * (this.TextHeight + this.SpacingV)

		g := this.ReadPoints(this.Pix, ox, oy, this.TextWidth, this.TextHeight)

		this.GlyphsI[rune(idx)] = g

	}

}

func round(f float32) float32 {
	return float32(math.Floor(float64(f) + 0.5))
}

func (this *DecalFont) GetScaleX() float32 {
	return this.ScaleX
}

func (this *DecalFont) SetScaleX(scaleX float32) {
	this.ScaleX = scaleX
}

func (this *DecalFont) SetScale(x float32, y float32) {
	this.SetScaleX(x)
	this.SetScaleY(y)
}

func NewDecaleFontFromReader(r io.Reader) (*DecalFont, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	f := &Font{}
	err = yaml.Unmarshal(b, f)
	if err != nil {
		return nil, err
	}
	font := NewDecalFont(f)
	return font, nil
}

func NewDecalFont(font *Font) *DecalFont {

	this := &DecalFont{}

	this.Border = 0
	this.GlyphHeight = 56
	this.GlyphWidth = 49
	this.ScreenWidth = SW
	this.ScreenHeight = SH

	this.TextWidth = font.Width
	this.TextHeight = font.Height
	this.InverseLoad = false

	this.GlyphsN = make(map[rune][]DecalPoint)
	this.GlyphsI = make(map[rune][]DecalPoint)

	for _, r := range font.Sources {
		start := r.Start
		end := r.End
		file := r.Source
		this.SpacingH = r.XPad
		this.SpacingV = r.YPad

		pix, err := LoadPNG(file)
		if err != nil {
			log.Fatalf("Unable to load image buffer from %s: %s\n", file, err.Error())
		}

		this.Pix = pix

		width := pix.Bounds().Max.X
		this.cpl = width / (this.TextWidth + this.SpacingH)

		this.LoadGlyphs(start, end)
	}

	return this
}

func (this *DecalFont) GetScaleY() float32 {
	return this.ScaleY
}

func (this *DecalFont) SetScaleY(scaleY float32) {
	this.ScaleY = scaleY
}

func (this *DecalFont) GetBytes(ch rune, inverted bool) []byte {
	var b = []byte{0, 0, 0, 0, 0, 0, 0, 0}
	var p []DecalPoint
	if inverted {
		p = this.GlyphsI[ch]
	} else {
		p = this.GlyphsN[ch]
	}
	for _, pt := range p {
		x := pt.X % 7
		y := pt.Y % 8
		b[y] = b[y] | (1 << uint(x))
	}
	return b
}

func (this *DecalFont) GetFontBitplanes() (map[rune][]byte, map[rune][]byte) {
	m := map[rune][]byte{}
	i := map[rune][]byte{}
	for ch := 32; ch <= 288; ch++ {
		m[rune(ch)] = this.GetBytes(rune(ch), false)
		i[rune(ch)] = this.GetBytes(rune(ch), true)
	}
	return m, i
}

var HighChar2Unicode map[rune]rune = map[rune]rune{
	1024: 9787, // filled smiley
	1025: 9824, // filled spade
	1026: 9475, // vertical bar
	1027: 9473, // horizontal bar
	1028: 9620, // horizontal bar ^
	1029: 9620, // horizontal bar ^^
	1030: 9601, // horizontal bar vv
	1031: 9615, // vertical bar <
	1032: 9621, // vertical bar >
	1033: 9582,
	1034: 9584,
	1035: 9583,
	1045: 9581,
	1043: 9829,
	1048: 9827,
	1050: 9830,
	1041: 9899,
	1047: 9898,
	1052: 11104,
	1053: 11106,
	1054: 11105,
	1055: 11107,
	1057: 9618,
	1058: 9618,
	0:    9725,
}

func LoadFromFile(filename string) (*DecalFont, error) {
	b, err := assets.Asset(filename)
	if err != nil {
		return nil, err
	}
	f, err := NewDecaleFontFromReader(bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	return f, err
}

func GetFontName(filename string) string {
	b, err := assets.Asset(filename)
	if err != nil {
		return "unknown"
	}
	f := &Font{}
	err = yaml.Unmarshal(b, f)
	if err != nil {
		return "unknown"
	}
	return f.Name
}

type GlyphGrid [16][16]bool

func (f *DecalFont) GlyphToVectors(ch rune, inverted bool) []*mgl32.Vec3 {
	var p []DecalPoint
	var ok bool
	if inverted {
		p, ok = f.GlyphsI[ch]
	} else {
		p, ok = f.GlyphsN[ch]
	}
	if !ok {
		return nil
	}
	var grid GlyphGrid
	var lines = make([]*mgl32.Vec3, 0)
	for _, point := range p {
		grid[point.X][point.Y] = true
	}
	lines = append(lines, grid.verticalVectors(f.TextWidth, f.TextHeight)...)
	lines = append(lines, grid.horizontalVectors(f.TextWidth, f.TextHeight)...)
	return lines
}

func (g GlyphGrid) horizontalVectors(w, h int) []*mgl32.Vec3 {
	var out = []*mgl32.Vec3{}
	for y := 0; y < h; y++ {
		var sx, ex = -1, -1
		for x := 0; x < w; x++ {
			if g[x][y] {
				if sx == -1 {
					sx = x
					ex = x
				} else {
					ex = x
				}
			} else {
				if sx != -1 {
					out = append(
						out,
						&mgl32.Vec3{float32(sx), float32(y), 0},
						&mgl32.Vec3{float32(ex), float32(y), 0},
					)
					sx = -1
					ex = -1
				}
			}
		}
		if sx != -1 {
			out = append(
				out,
				&mgl32.Vec3{float32(sx), float32(y), 0},
				&mgl32.Vec3{float32(ex), float32(y), 0},
			)
			sx = -1
			ex = -1
		}
	}
	return out
}

func (g GlyphGrid) verticalVectors(w, h int) []*mgl32.Vec3 {
	var out = []*mgl32.Vec3{}
	for x := 0; x < w; x++ {
		var sy, ey = -1, -1
		for y := 0; y < h; y++ {
			if g[x][y] {
				if sy == -1 {
					sy = y
					ey = y
				} else {
					ey = y
				}
			} else {
				if sy != -1 {
					out = append(
						out,
						&mgl32.Vec3{float32(x), float32(sy), 0},
						&mgl32.Vec3{float32(x), float32(ey), 0},
					)
					sy = -1
					ey = -1
				}
			}
		}
		if sy != -1 {
			out = append(
				out,
				&mgl32.Vec3{float32(x), float32(sy), 0},
				&mgl32.Vec3{float32(x), float32(ey), 0},
			)
			sy = -1
			ey = -1
		}
	}
	return out
}

func (g GlyphGrid) diagVectors(w, h int, hd int) []*mgl32.Vec3 {

	var out = []*mgl32.Vec3{}
	// for y := 0; y <  {
	// 	var sx, ex = -1, -1
	// 	for x := 0; x < w; x++ {
	// 		if g[x][y] {
	// 			if sx == -1 {
	// 				sx = x
	// 				ex = x
	// 			} else {
	// 				ex = x
	// 			}
	// 		} else {
	// 			if sx != -1 {
	// 				if ex-sx > 0 {
	// 					out = append(
	// 						out,
	// 						&mgl32.Vec3{sx, y, 0},
	// 						&mgl32.Vec3{ex, y, 0},
	// 					)
	// 				}
	// 				sx = -1
	// 				ex = -1
	// 			}
	// 		}
	// 	}
	// }
	return out

}
