package types

import (
	"errors"
	"image/color"
	"math"

	"paleotronic.com/fmt"
	//"github.com/lucasb-eyer/go-colorful"
)

type VideoColor struct {
	Red, Green, Blue, Alpha uint8
	Offset                  int8
	Depth                   uint8
}

func NewVideoColor(i uint8, j uint8, k uint8, l uint8) *VideoColor {
	this := &VideoColor{}

	/* vars */

	this.Red = i
	this.Green = j
	this.Blue = k
	this.Alpha = l

	return this

}

func (this VideoColor) ToColorRGBA() color.RGBA {
	return color.RGBA{this.Red, this.Green, this.Blue, this.Alpha}
}

func (this VideoColor) ToColorNRGBA(alpha uint8) color.RGBA {
	return color.RGBA{
		uint8(float32(this.Red) * (float32(alpha) / 255)),
		uint8(float32(this.Green) * (float32(alpha) / 255)),
		uint8(float32(this.Blue) * (float32(alpha) / 255)),
		255,
	}
}

func (this VideoColor) ToRGBA() (int32, int32, int32, int32) {
	return int32(this.Red)*256 + 255, int32(this.Green)*256 + 255, int32(this.Blue)*256 + 255, int32(this.Alpha)*256 + 255
}

func (this VideoColor) ToFRGBA() (float32, float32, float32, float32) {
	return float32(this.Red) / 255,
		float32(this.Green) / 255,
		float32(this.Blue) / 255,
		float32(this.Alpha) / 255
}

func (this VideoColor) ToFRGBAI(c uint64) (float32, float32, float32, float32) {
	amod := float32(255-(c>>24)) / 255
	return float32(this.Red) / 255,
		float32(this.Green) / 255,
		float32(this.Blue) / 255,
		(float32(this.Alpha) / 255) * amod
}

func (this VideoColor) ToFRGBA64() (float64, float64, float64, float64) {
	return float64(this.Red) / 255,
		float64(this.Green) / 255,
		float64(this.Blue) / 255,
		float64(this.Alpha) / 255
}

const Pr = .299
const Pg = .587
const Pb = .114

func changeSaturation(R float64, G float64, B float64, change float64) (float64, float64, float64) {

	var P float64 = math.Sqrt(R*R*Pr + G*G*Pg + B*B*Pb)

	R = P + (R-P)*change
	G = P + (G-P)*change
	B = P + (B-P)*change

	return R, G, B
}

func (this *VideoColor) ChangeSaturation(change float64) *VideoColor {

	r, g, b, _ := this.ToFRGBA64()

	nr, ng, nb := changeSaturation(r, g, b, change)

	return &VideoColor{
		Red:   uint8(255 * nr),
		Green: uint8(255 * ng),
		Blue:  uint8(255 * nb),
		Alpha: this.Alpha,
	}

}

func (this VideoColor) MarshalBinary() ([]byte, error) {
	return []byte{byte(this.Red), byte(this.Green), byte(this.Blue), byte(this.Alpha), byte(this.Offset), byte(this.Depth)}, nil
}

func (this *VideoColor) UnmarshalBinary(data []byte) error {
	if len(data) < 6 {
		return errors.New("Not enough data")
	}
	this.Red = data[0]
	this.Green = data[1]
	this.Blue = data[2]
	this.Alpha = data[3]
	this.Offset = int8(data[4])
	this.Depth = uint8(data[5])
	return nil
}

func (this *VideoColor) String() string {
	return fmt.Sprintf("RGBAOD(%d,%d,%d,%d,%d,%d)", this.Red, this.Green, this.Blue, this.Alpha, this.Offset, this.Depth)
}

func (this *VideoColor) ToUintRGBA() uint64 {
	return (uint64(this.Red) << 24) | (uint64(this.Green) << 16) | (uint64(this.Blue) << 8) | uint64(this.Alpha)
}

//~ func (a *VideoColor) PerceptualDistance(b *VideoColor) float64 {

//~ ar, ag, ab, _ :=  a.ToFRGBA64()
//~ br, bg, bb, _ :=  b.ToFRGBA64()

//~ c2a := colorful.Color{ ar, ag, ab }
//~ c2b := colorful.Color{ br, bg, bb }

//~ return c2a.DistanceCIE76( c2b )

//~ }

func (vc *VideoColor) ToCIELAB() (float64, float64, float64) {
	return rgb2lab(vc.Red, vc.Green, vc.Blue)
}

func rgb2lab(R uint8, G uint8, B uint8) (float64, float64, float64) {
	//http://www.brucelindbloom.com

	var r, g, b, X, Y, Z, fx, fy, fz, xr, yr, zr float64
	var Ls, as, bs float64
	var eps float64 = 216 / 24389
	var k float64 = 24389 / 27

	var Xr float64 = 0.964221 // reference white D50
	var Yr float64 = 1.0
	var Zr float64 = 0.825211

	// RGB to XYZ
	r = float64(R) / 255 //R 0..1
	g = float64(G) / 255 //G 0..1
	b = float64(B) / 255 //B 0..1

	// assuming sRGB (D65)
	if r <= 0.04045 {
		r = r / 12
	} else {
		r = math.Pow((r+0.055)/1.055, 2.4)
	}

	if g <= 0.04045 {
		g = g / 12
	} else {
		g = math.Pow((g+0.055)/1.055, 2.4)
	}

	if b <= 0.04045 {
		b = b / 12
	} else {
		b = math.Pow((b+0.055)/1.055, 2.4)
	}

	X = 0.436052025*r + 0.385081593*g + 0.143087414*b
	Y = 0.222491598*r + 0.71688606*g + 0.060621486*b
	Z = 0.013929122*r + 0.097097002*g + 0.71418547*b

	// XYZ to Lab
	xr = X / Xr
	yr = Y / Yr
	zr = Z / Zr

	if xr > eps {
		fx = math.Pow(xr, 1/3.)
	} else {
		fx = ((k*xr + 16.) / 116.)
	}

	if yr > eps {
		fy = math.Pow(yr, 1/3.)
	} else {
		fy = ((k*yr + 16.) / 116.)
	}

	if zr > eps {
		fz = math.Pow(zr, 1/3.)
	} else {
		fz = ((k*zr + 16.) / 116)
	}

	Ls = (116 * fy) - 16
	as = 500 * (fx - fy)
	bs = 200 * (fy - fz)

	return (2.55*Ls + .5),
		(as + .5),
		(bs + .5)
}

func (e1 *VideoColor) PerceptualDistance(e2 *VideoColor) float64 {

	l1, a1, b1 := e1.ToCIELAB()
	l2, a2, b2 := e2.ToCIELAB()

	return math.Sqrt(
		math.Pow(l2-l1, 2) +
			math.Pow(a2-a1, 2) +
			math.Pow(b2-b1, 2),
	)

}

func (a *VideoColor) EuclideanDistance(b *VideoColor) float64 {

	ar, ag, ab, _ := a.ToFRGBA64()
	br, bg, bb, _ := b.ToFRGBA64()

	return math.Sqrt(
		math.Pow(br-ar, 2) + math.Pow(bg-ag, 2) + math.Pow(bb-ab, 2),
	)

}

func (this *VideoColor) FromUintRGBA(u uint64) {
	this.Alpha = uint8(u & 0xff)
	this.Blue = uint8((u >> 8) & 0xff)
	this.Green = uint8((u >> 16) & 0xff)
	this.Red = uint8((u >> 24) & 0xff)
}

func (this *VideoColor) RGB233() RGB233 {

	return RGB233((this.Red & 192) | ((this.Green >> 2) & 56) | ((this.Blue >> 5) & 7))

}

func (this *VideoColor) RGB323() RGB323 {

	return RGB323((this.Red & 224) | ((this.Green >> 3) & 24) | ((this.Blue >> 5) & 7))

}

func (this *VideoColor) RGB332() RGB332 {

	return RGB332((this.Red & 224) | ((this.Green >> 3) & 28) | ((this.Blue >> 6) & 3))

}

func (this *VideoColor) Desaturate() *VideoColor {

	level := uint8((int(this.Red) + int(this.Green) + int(this.Blue)) / 3)
	return NewVideoColor(level, level, level, this.Alpha)

}

func (this *VideoColor) Tint(r, g, b uint8) *VideoColor {
	ra := float32(r) / 255
	ba := float32(b) / 255
	ga := float32(g) / 255
	return NewVideoColor(
		uint8(float32(this.Red)*ra),
		uint8(float32(this.Green)*ga),
		uint8(float32(this.Blue)*ba),
		this.Alpha,
	)
}

func Color2RGB233(c color.Color) RGB233 {

	r, g, b, _ := c.RGBA()

	// RR------ -------- GGG----- -------- BBB----- --------
	rr := (r >> 14)
	gg := (g >> 13)
	bb := (b >> 13)

	return RGB233((rr << 6) | (gg << 3) | bb)

}

func Color2RGB323(c color.Color) RGB323 {

	r, g, b, _ := c.RGBA()

	// RRR----- -------- GG------ -------- BBB----- --------
	rr := (r >> 13)
	gg := (g >> 14)
	bb := (b >> 13)

	return RGB323((rr << 5) | (gg << 3) | bb)

}

func Color2RGB332(c color.Color) RGB332 {

	r, g, b, _ := c.RGBA()

	// RRR----- -------- GGG----- -------- BB------ --------
	rr := (r >> 13)
	gg := (g >> 13)
	bb := (b >> 14)

	return RGB332((rr << 5) | (gg << 2) | bb)

}

func RGB233ToVideoColor(c RGB233) VideoColor {

	r := uint8(c&192) | 63
	g := uint8((c&56)<<2) | 31
	b := uint8((c&7)<<5) | 31

	return VideoColor{Red: r, Green: g, Blue: b, Alpha: 255}

}

func RGB323ToVideoColor(c RGB323) VideoColor {

	r := uint8(c&224) | 31
	g := uint8((c&24)<<3) | 63
	b := uint8((c&7)<<5) | 31

	return VideoColor{Red: r, Green: g, Blue: b, Alpha: 255}

}

func RGB332ToVideoColor(c RGB332) VideoColor {

	r := uint8(c&224) | 31
	g := uint8((c&28)<<3) | 31
	b := uint8((c&7)<<6) | 63

	return VideoColor{Red: r, Green: g, Blue: b, Alpha: 255}

}
