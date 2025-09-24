package types

import (
	//"paleotronic.com/fmt"
	"math"
	"testing"
)

func DumpUint64Bits(bits uint64) string {
	out := ""

	for i := 0; i < 64; i++ {
		if bits&(1<<63) != 0 {
			out = out + "1"
		} else {
			out = out + "0"
		}
		bits = bits << 1

		if i == 0 || i == 11 {
			out += " "
		}
	}

	return out
}

func DumpFloatBits(f float64) string {

	bits := math.Float64bits(f)

	return DumpUint64Bits(bits)

}

func TestFloat5b(t *testing.T) {

//	var ff float64 = 811.9999999

	//fmt.Printf("Binary of %f is %s\n", ff, DumpFloatBits(ff))

//	fb := NewFloat5b(ff)

//	exp := uint64(fb.GetExponent()+1022) << 52
	//fmt.Printf("Exp bits = %s\n", DumpUint64Bits(uint64(exp)))

//	gg := fb.GetValue()
	//fmt.Println(gg)
	//fmt.Printf("Binary of %f is %s\n", gg, DumpFloatBits(gg))
    //fmt.Println( math.Floor(gg) )

	//t.Fail()

}
