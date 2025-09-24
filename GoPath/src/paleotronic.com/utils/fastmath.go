package utils

import "math"

var sine [90]float64
var cosine [90]float64

func init() {
	preComputeSinCos()
}

func preComputeSinCos() {
	for i:=0; i<90; i++ {
		r := float64(i) * 0.0174533
		sine[i] = math.Sin(r)
		cosine[i] = math.Cos(r)
	}
}

func Sin( r float64 ) float64 {
	d := int( math.Floor(r * 57.2958 + 0.5) ) % 90
	return sine[d]
}

func Cos( r float64 ) float64 {
	d := int( math.Floor(r * 57.2958 + 0.5) ) % 90
	return cosine[d]
}



