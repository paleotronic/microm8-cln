package common

import (
	"image/color"
)

func clip(v int, max int) int {
	if v > max {
		return max
	}
	return v
}

func mixCMYK(a, b color.CMYK) color.CMYK {

	c := clip(int(a.C)+int(b.C), 255)
	m := clip(int(a.M)+int(b.M), 255)
	y := clip(int(a.Y)+int(b.Y), 255)
	k := clip(int(a.K)+int(b.K), 255)

	return color.CMYK{uint8(c), uint8(m), uint8(y), uint8(k)}

}
