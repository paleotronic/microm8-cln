package restalgia

import (
	"math"
)

const RESOLUTION = 16384
const lv = 90000000000000

var sval [RESOLUTION]float64
var cval [RESOLUTION]float64
var asval [RESOLUTION]float64
var acval [RESOLUTION]float64
var atval [RESOLUTION]float64

func init() {
	for i, _ := range sval {
		f := (2 * math.Pi) * (float64(i) / RESOLUTION)
		sval[i] = math.Sin(f)
		cval[i] = math.Cos(f)
	}
	for i, _ := range asval {
		f := ((float64(i) / RESOLUTION) * 2) - 1
		asval[i] = math.Asin(f)
		acval[i] = math.Acos(f)
	}
	for i, _ := range atval {
		f := ((float64(i) / RESOLUTION) * 2 * lv) - lv
		atval[i] = math.Atan(f)
	}
}

func Sin(r float64) float64 {
	index := int(r/(2*math.Pi)*RESOLUTION) % RESOLUTION
	return sval[index]
}

func Cos(r float64) float64 {
	index := int(r/(2*math.Pi)*RESOLUTION) % RESOLUTION
	return cval[index]
}

func Tan(r float64) float64 {
	index := int(r/(2*math.Pi)*RESOLUTION) % RESOLUTION
	return sval[index] / cval[index]
}

func Asin(r float64) float64 {
	index := int(((r+1)/2)*RESOLUTION) % RESOLUTION
	return asval[index]
}

func Acos(r float64) float64 {
	index := int(((r+1)/2)*RESOLUTION) % RESOLUTION
	return acval[index]
}

func Atan(r float64) float64 {
	// if r < -lv {
	// 	r = -lv
	// }
	// if r > lv {
	// 	r = lv
	// }
	// index := int(((r + lv) / (2 * lv)) * RESOLUTION)
	// if index >= RESOLUTION {
	// 	index = RESOLUTION - 1
	// }
	// if index < 0 {
	// 	index = 0
	// }
	// return atval[index]
	return math.Atan(r)
}
