package types

import (
	"errors"
	"image/color"
	"math"

	"paleotronic.com/log"
)

type RGB323 byte
type RGB332 byte
type RGB233 byte

type VideoPalette struct {
	Items         []*VideoColor
	DitherPalette map[byte][]int
	ColorSpace    byte

	PCache map[color.Color]int
	ECache map[color.Color]int
}

func (this *VideoPalette) AddRGB12(c int) {
	var r, g, b, a uint8
	r = uint8((c&0xf00)>>4) | 0xf
	g = uint8(c&0x0f0) | 0xf
	b = uint8((c&0x00f)<<4) | 0xf
	a = 0xff
	if c == 0x000 {
		a = 0x00
	}
	this.Add(&VideoColor{Red: r, Green: g, Blue: b, Alpha: a, Offset: 0, Depth: 20})
}

func (this *VideoPalette) GetRGB12(index int) int {
	if (index >= 0) && (index < len(this.Items)) {
		c := this.Items[index]
		return ((int(c.Red) & 0xf0) << 4) | (int(c.Green) & 0xf0) | (int(c.Blue) >> 4)
	}
	return 0
}

func (this *VideoPalette) ToRGB12() []int {
	r := make([]int, this.Size())
	for i, _ := range r {
		r[i] = this.GetRGB12(i)
	}
	return r
}

func (this *VideoPalette) FromRGB12(r []int) {
	this.Items = []*VideoColor{}
	for _, c := range r {
		this.AddRGB12(c)
	}
}

func (this *VideoPalette) Get(index int) *VideoColor {

	/* vars */
	var result *VideoColor

	index = index % this.Size()

	result = nil
	if (index >= 0) && (index < len(this.Items)) {
		result = this.Items[index]
	}

	/* enforce non void return */
	return result

}

func (this *VideoPalette) SetColor(index int, vc *VideoColor) {

	/* vars */

	if (index >= 0) && (index < len(this.Items)) {
		this.Items[index] = vc
	}

}

func NewVideoPalette() *VideoPalette {
	this := &VideoPalette{}

	/* vars */

	this.Items = make([]*VideoColor, 0)

	this.PCache = make(map[color.Color]int)
	this.ECache = make(map[color.Color]int)

	return this
}

func (this *VideoPalette) Size() int {

	return len(this.Items)
}

func (this *VideoPalette) Add(c *VideoColor) {

	/* vars */

	this.Items = append(this.Items, c)

}

func (this *VideoPalette) Desaturate() *VideoPalette {
	p := NewVideoPalette()
	for _, vc := range this.Items {
		p.Add(vc.Desaturate())
	}
	return p
}

func (this *VideoPalette) Tint(r, g, b uint8) *VideoPalette {
	p := NewVideoPalette()
	for _, vc := range this.Items {
		p.Add(vc.Desaturate().Tint(r, g, b))
	}
	return p
}

func (this VideoPalette) MarshalBinary() ([]byte, error) {
	data := []byte{byte(this.Size())}

	for _, cc := range this.Items {
		b, _ := cc.MarshalBinary()
		data = append(data, b...)
	}

	return data, nil
}

func (this *VideoPalette) UnmarshalBinary(data []byte) error {

	if len(data) == 0 {
		return errors.New("Not enough data")
	}

	count := int(data[0])

	if len(data) < (1 + count*5) {
		return errors.New("Not enough data")
	}

	this.Items = make([]*VideoColor, 0)

	for i := 0; i < count; i++ {
		s := 1 + (5 * i)
		e := s + 5
		b := data[s:e]
		cc := &VideoColor{}
		err := cc.UnmarshalBinary(b)
		if err != nil {
			log.Fatal(err.Error())
		}
		this.Items = append(this.Items, cc)
	}

	return nil

}

func cbyte(v uint32) byte {

	if v < 256 {
		return byte(v & 0xff)
	}

	return 255

}

func (this *VideoPalette) GetMatch(c color.Color, perceptual bool) int {

	if perceptual {

		i, ok := this.PCache[c]
		if ok {
			return i
		}

	} else {

		i, ok := this.ECache[c]
		if ok {
			return i
		}

	}

	r, g, b, _ := c.RGBA()

	a := &VideoColor{Red: uint8(r / 256), Green: uint8(g / 256), Blue: uint8(b / 256), Alpha: 255}

	var lowdiff float64 = 999999999
	var lowindex int = -1

	var v float64
	for i, b := range this.Items {
		if perceptual {
			v = a.PerceptualDistance(b)
		} else {
			v = a.EuclideanDistance(b)
		}

		if v < lowdiff {
			lowindex = i
			lowdiff = v
		}
	}

	if perceptual {
		this.PCache[c] = lowindex
	} else {
		this.ECache[c] = lowindex
	}

	return lowindex

}

func (this *VideoPalette) GetTwoMatch(c color.Color) (int, int) {

	r, g, b, _ := c.RGBA()
	cr, cg, cb := float64(r), float64(g), float64(b)

	var lowdiff float64 = 999999999
	//var plowdiff float64 = 999999999
	var lowindex int = -1
	var plowindex int = -1

	for i, c := range this.Items {
		pr, pg, pb := float64(c.Red)*256+255, float64(c.Green)*256+255, float64(c.Blue)*256+255

		v := math.Sqrt((cr-pr)*(cr-pr) + (cg-pg)*(cg-pg) + (cb-pb)*(cb-pb))

		if v < lowdiff {
			plowindex = lowindex
			lowindex = i
			lowdiff = v
		}
	}

	return lowindex, plowindex

}

// Palette matching function for color.Color
func (this *VideoPalette) GetClosestIndex(colorspace byte, x, y int, col color.Color, boost uint32) int {

	var cx byte

	switch colorspace {
	case 0:
		cx = byte(Color2RGB332(col))
	case 1:
		cx = byte(Color2RGB323(col))
	case 2:
		cx = byte(Color2RGB233(col))
	}
	rgbpal := this.GetDitherPalette(colorspace)

	matches := rgbpal[cx]

	if len(matches) > 1 {

		return matches[(x+y)%2]

	}

	return matches[0]
}

func (this *VideoPalette) GetDitherPalette(cs byte) map[byte][]int {
	if this.DitherPalette != nil && cs == this.ColorSpace {
		return this.DitherPalette
	}

	// Need to create the dither palette
	this.DitherPalette = make(map[byte][]int)
	this.ColorSpace = cs

	for idx := 0; idx < 256; idx++ {

		var vc VideoColor
		var cx byte
		switch this.ColorSpace {
		case 0:
			cx = byte(RGB332(idx))
			vc = RGB332ToVideoColor(RGB332(cx))
		case 1:
			cx = byte(RGB323(idx))
			vc = RGB323ToVideoColor(RGB323(cx))
		case 2:
			cx = byte(RGB233(idx))
			vc = RGB233ToVideoColor(RGB233(cx))
		}

		// find closest match to this index in palette
		match := 0
		//		pmatch := 0
		var low_dc uint32 = (1 << 32) - 1
		//		var plow_dc uint32 = (1 << 32) - 1

		r, g, b := vc.Red, vc.Green, vc.Blue

		for i, ccz := range this.Items {

			var cc VideoColor

			switch this.ColorSpace {
			case 0:
				cc = RGB332ToVideoColor(ccz.RGB332()) // flatten
			case 1:
				cc = RGB323ToVideoColor(ccz.RGB323())
			case 2:
				cc = RGB233ToVideoColor(ccz.RGB233())
			}
			crr, cgg, cbb := cc.Red, cc.Green, cc.Blue

			v := (float64(uint32(crr)-uint32(r)) * float64(uint32(crr)-uint32(r))) +
				(float64(uint32(cgg)-uint32(g)) * float64(uint32(cgg)-uint32(g))) +
				(float64(uint32(cbb)-uint32(b)) * float64(uint32(cbb)-uint32(b)))

			dc := uint32(math.Sqrt(v))

			if dc < low_dc {

				//				pmatch = match
				//				plow_dc = low_dc

				match = i
				low_dc = dc

			}

		}

		//		cdiff := math.Abs(float64(plow_dc - low_dc))

		//		if low_dc < 10 {
		this.DitherPalette[cx] = []int{match}
		//		} else {
		//	if cdiff < 600 {
		//		this.DitherPalette[cx] = []int{match, pmatch}
		//	} else {
		//		this.DitherPalette[cx] = []int{match}
		//	}
		//}

		//log.Printf("palette map ---> idx = %d, r = %d, g = %d, b = %d, low = %d, plow = %d, cdiff = %f, match=%d, pmatch=%d, pi = %v\n", idx, r, g, b, low_dc, plow_dc, cdiff, match, pmatch, this.DitherPalette[cx])

	}

	return this.DitherPalette
}

func (this *VideoPalette) String() string {

	out := ""
	for i, c := range this.Items {
		if i > 0 {
			out += ";"
		}
		out += c.String()
	}
	return out
}
