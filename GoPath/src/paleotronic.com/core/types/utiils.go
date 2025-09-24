package types

import (
	"bytes"
	"encoding/binary"
	"math"
	"strings"
)

// Helper for converting floats to []byte
func Float2uint_old(f float32) uint64 {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, f)
	if err != nil {
		return 0
	}
	b := buf.Bytes()
	return uint64(b[0])<<24 | uint64(b[1])<<16 | uint64(b[2])<<8 | uint64(b[3])
}

func Float2uint(f float32) uint64 {
	return uint64(math.Float32bits(f))
}

func Float642uint(f float64) uint64 {
	return uint64(math.Float64bits(f))
}

func bytes2Uint(b []byte) uint64 {
	return uint64(b[0])<<24 | uint64(b[1])<<16 | uint64(b[2])<<8 | uint64(b[3])
}

func uint2Bytes(u uint64) []byte {
	data := make([]byte, 4)

	data[0] = byte((u & 0xff000000) >> 24)
	data[1] = byte((u & 0x00ff0000) >> 16)
	data[2] = byte((u & 0x0000ff00) >> 8)
	data[3] = byte(u & 0x000000ff)

	return data
}

func Uint2Float_old(u uint64) float32 {
	data := make([]byte, 4)

	data[0] = byte((u & 0xff000000) >> 24)
	data[1] = byte((u & 0x00ff0000) >> 16)
	data[2] = byte((u & 0x0000ff00) >> 8)
	data[3] = byte(u & 0x000000ff)

	var f float32
	b := bytes.NewBuffer(data)
	_ = binary.Read(b, binary.LittleEndian, &f)
	return f
}

func Uint2Float(u uint64) float32 {
	return math.Float32frombits(uint32(u))
}

func Uint2Float64(u uint64) float64 {
	return math.Float64frombits(u)
}

func UintSlice2Float(u []uint64) []float32 {
	f := make([]float32, len(u))
	for i, v := range u {
		f[i] = math.Float32frombits(uint32(v))
	}
	return f
}

func UintSlice2FloatBP(u []uint64) []float32 {
	f := make([]float32, len(u))
	for i, v := range u {
		//f[i] = math.Float32frombits( uint32(v) )
		switch v {
		case 2:
			f[i] = 1
		case 1:
			f[i] = 0
		case 0:
			f[i] = -1
		}
	}
	return f
}

func FloatSlice2UintBP(u []float32) []uint64 {
	f := make([]uint64, len(u))
	for i, v := range u {
		switch {
		case v < 0:
			f[i] = 0
		case v == 0:
			f[i] = 1
		case v > 0:
			f[i] = 2
		}
	}
	return f
}

func FloatSlice2Uint(u []float32) []uint64 {
	f := make([]uint64, len(u))
	for i, v := range u {
		f[i] = uint64(math.Float32bits(v))
	}
	return f
}

func PackName(name string, l int) []uint64 {
	if len(name) > l {
		name = name[0:l]
	}
	for len(name) < l {
		name += string(rune(0))
	}
	b := []byte(name)

	data := make([]uint64, 0)

	for len(b) > 0 {
		if len(b) >= 4 {
			chunk := b[0:4]
			b = b[4:]
			data = append(data, bytes2Uint(chunk))
		}
	}

	return data
}

func UnpackName(data []uint64) string {

	out := make([]byte, 0)

	for _, u := range data {
		b := uint2Bytes(u)

		for _, bb := range b {
			if bb != 0 {
				out = append(out, bb)
			}
		}
	}

	s := string(out)

	return strings.Trim(s, " ")
}
