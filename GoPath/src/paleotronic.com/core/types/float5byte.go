package types

/*
     exp       m4       m3       m2       m1
      98       35       44       7A       00      - some constant in hex
     152       53       68      122        0      - same in dec
10011000 00110101 01000100 01111010 00000000      - same in bin
         ^sign bit

In this case:
Exponent = 152 - 128 = 24                                      ; dec
Mantissa = 0.10110101010001000111101000000000                  ; bin
Mantissa = +1 * (53 >> 8 + 68 >> 16 + 122 >> 24 + 0 >> 32)     ; dec
         = 1 * 2^-1 + 0 * 2^-2 + 1 * 2^-3 + 1 * 2^-4 + 0 * 2^-5 + ...
         = 0.70807611942291259765

So the number is...
0.70807611942291259765 * 2^24 = 11879546.0
*/

import (
	"math"
	"strconv"

	"paleotronic.com/core/memory"
	"paleotronic.com/utils"
	//       "encoding/binary"
)

type Float5b struct {
	exp            byte
	m4, m3, m2, m1 byte
}

func NewFloat5b(f float64) *Float5b {
	this := &Float5b{}
	this.SetValue(f)
	return this
}

func (f *Float5b) GetExponent() int {
	return int(f.exp) - 128
}

func (f *Float5b) SetExponent(e int) {
	f.exp = byte(128 + e)
}

func (f *Float5b) GetValueCrufty() float64 {

	s := float64(1)
	if f.m4&128 != 0 {
		s = -1
	}

	//if f.m1 == 0 && f.m2 == 0 && f.m3 == 0 && f.m4 == 0 {
	// return 0
	//}

	m := 0.5 + float64(f.m4&0x7f)/math.Pow(2, 8) + float64(f.m3)/math.Pow(2, 16) + float64(f.m2)/math.Pow(2, 24) + float64(f.m1)/math.Pow(2, 32)
	v := s * m * math.Pow(2, float64(f.GetExponent()))
	return v
}

func (f *Float5b) GetValue() float64 {

	var exp int = f.GetExponent()
	var mantissa uint64 = (uint64(f.m4&127) << 45) | (uint64(f.m3) << 37) | (uint64(f.m2) << 29) | (uint64(f.m1) << 21)

	var expbig uint64 = uint64(exp+1022) << 52
	var sign uint64 = 0
	if f.m4&128 == 128 {
		sign = (1 << 63)
	}

	var value uint64 = mantissa | expbig | sign

	float := math.Float64frombits(value)

	return float
}

func (f *Float5b) SetBytes(exp, m4, m3, m2, m1 byte) {
	f.exp, f.m4, f.m3, f.m2, f.m1 = exp, m4, m3, m2, m1
}

func (f *Float5b) SetValue(arg float64) {

	if arg == 0 {
		f.exp = 0
		f.m1 = 0
		f.m2 = 0
		f.m3 = 0
		f.m4 = 0
		return
	}

	_, expnt := math.Frexp(math.Abs(arg))

	sign := byte(0)
	if arg < 0 {
		sign = 128
	}

	mantissa := (math.Float64bits(math.Abs(arg)) >> 21) & 0x00000000ffffffff

	f.exp = byte(128 + expnt)
	f.m1 = byte(mantissa & 0xff)
	f.m2 = byte((mantissa >> 8) & 0xff)
	f.m3 = byte((mantissa >> 16) & 0xff)
	f.m4 = byte((mantissa>>24)&0x7f) | sign
}

func (f *Float5b) Add(v *Float5b) *Float5b {

	return NewFloat5b(f.GetValue() + v.GetValue())

}

func (f *Float5b) AddF(v float64) *Float5b {

	return NewFloat5b(f.GetValue() + v)

}

func (f *Float5b) Sub(v *Float5b) *Float5b {

	return NewFloat5b(f.GetValue() - v.GetValue())

}

func (f *Float5b) SubF(v float64) *Float5b {

	return NewFloat5b(f.GetValue())

}

func (f *Float5b) Mul(v *Float5b) *Float5b {

	return NewFloat5b(f.GetValue() * v.GetValue())

}

func (f *Float5b) MulF(v float64) *Float5b {

	return NewFloat5b(f.GetValue() * v)

}

func (f *Float5b) Div(v *Float5b) *Float5b {

	return NewFloat5b(f.GetValue() / v.GetValue())

}

func (f *Float5b) DivF(v float64) *Float5b {

	return NewFloat5b(f.GetValue() / v)

}

func (f *Float5b) Int() int {

	return int(f.GetValue())

}

func (f *Float5b) Sin() *Float5b {

	return NewFloat5b(math.Sin(f.GetValue()))

}

func (f *Float5b) Cos() *Float5b {

	return NewFloat5b(math.Cos(f.GetValue()))

}

func (f *Float5b) Tan() *Float5b {

	return NewFloat5b(math.Tan(f.GetValue()))

}

func (f *Float5b) Atan() *Float5b {

	return NewFloat5b(math.Atan(f.GetValue()))

}

func (f *Float5b) Floor() *Float5b {

	return NewFloat5b(math.Floor(f.GetValue()))

}

func (f *Float5b) Ceil() *Float5b {

	return NewFloat5b(math.Ceil(f.GetValue()))

}

func (f *Float5b) Round() *Float5b {

	v := f.GetValue()

	if v > 0 {
		v = math.Floor(v + 0.5)
	} else {
		v = math.Floor(v - 0.5)
	}

	return NewFloat5b(v)

}

func (f *Float5b) Sqrt() *Float5b {

	return NewFloat5b(math.Sqrt(f.GetValue()))

}

func (f *Float5b) Pow(v *Float5b) *Float5b {

	return NewFloat5b(math.Pow(f.GetValue(), v.GetValue()))

}

func (f Float5b) String() string {

	s := strconv.FormatFloat(f.GetValue(), 'f', 17, 64)

	return utils.StrToFloatStrApple(s)

}

func (f *Float5b) WriteMemory(mm *memory.MemoryMap, index int, address int) {
	mm.WriteInterpreterMemory(index, address+0, uint64(f.exp))
	mm.WriteInterpreterMemory(index, address+1, uint64(f.m4))
	mm.WriteInterpreterMemory(index, address+2, uint64(f.m3))
	mm.WriteInterpreterMemory(index, address+3, uint64(f.m2))
	mm.WriteInterpreterMemory(index, address+4, uint64(f.m1))
}

func (f *Float5b) ReadMemory(mm *memory.MemoryMap, index int, address int) {
	f.exp = byte(mm.ReadInterpreterMemory(index, address+0))
	f.m4 = byte(mm.ReadInterpreterMemory(index, address+1))
	f.m3 = byte(mm.ReadInterpreterMemory(index, address+2))
	f.m2 = byte(mm.ReadInterpreterMemory(index, address+3))
	f.m1 = byte(mm.ReadInterpreterMemory(index, address+4))
}

func (f *Float5b) WriteMemoryFACFormat(mm *memory.MemoryMap, index int, address int) {

	// 6 bytes
	// EXP  M1		M2		M3 		M4 		SGN
	// s    b7=1    s    	s		s		(neg=0xff, pos=0x00)

	mm.WriteInterpreterMemory(index, address+0, uint64(f.exp))
	mm.WriteInterpreterMemory(index, address+1, uint64(f.m4)|0x80) // set bit 7
	mm.WriteInterpreterMemory(index, address+2, uint64(f.m3))
	mm.WriteInterpreterMemory(index, address+3, uint64(f.m2))
	mm.WriteInterpreterMemory(index, address+4, uint64(f.m1))

	if f.GetValue() < 0 {
		mm.WriteInterpreterMemory(index, address+5, 0xff)
	} else {
		mm.WriteInterpreterMemory(index, address+5, 0x00)
	}
}

func (f *Float5b) ReadMemoryFACFormat(mm *memory.MemoryMap, index int, address int) {

	// 6 bytes
	// EXP  M1		M2		M3 		M4 		SGN
	// s    b7=1    s    	s		s		(neg=0xff, pos=0x00)

	f.exp = byte(mm.ReadInterpreterMemory(index, address+0))

	f.m3 = byte(mm.ReadInterpreterMemory(index, address+2))
	f.m2 = byte(mm.ReadInterpreterMemory(index, address+3))
	f.m1 = byte(mm.ReadInterpreterMemory(index, address+4))

	v := byte(mm.ReadInterpreterMemory(index, address+5)) // sgn
	if v == 0xff {
		f.m4 = byte(mm.ReadInterpreterMemory(index, address+1))
	} else {
		f.m4 = byte(mm.ReadInterpreterMemory(index, address+1)) & 0x7f // strip bit 7
	}

}
