package decalfont

import (
	"paleotronic.com/core/types"
)

func XYToOffset80(x, y int) int {
	var offset int = 0
	if (x % 2) == 0 {
		offset = 1024
	}

	return ((y % 8) * 128) + ((y / 8) * 40) + (x / 2) + offset
}

func XYToOffset40(x, y int) int {
	return ((y % 8) * 128) + ((y / 8) * 40) + x
}

//func XYToOffset(cols int, x, y int) int {
//	if cols == 80 {
//		return XYToOffset80(x, y)
//	} else {
//		return XYToOffset40(x, y)
//	}
//}

func Between(v, lo, hi int) bool {
	return ((v >= lo) && (v <= hi))
}

func PokeToAscii(v int) int {
	v = v & 0xffff

	if Between(v, 0, 31) {
		return 64 + (v % 32)
	}

	if Between(v, 32, 63) {
		return 32 + (v % 32)
	}

	if Between(v, 64, 95) {
		return 64 + (v % 32)
	}

	if Between(v, 96, 127) {
		return 32 + (v % 32)
	}

	if Between(v, 128, 159) {
		return 64 + (v % 32)
	}

	if Between(v, 160, 191) {
		return 32 + (v % 32)
	}

	if Between(v, 192, 223) {
		return 64 + (v % 32)
	}

	if Between(v, 224, 255) {
		return 96 + (v % 32)
	}

	if Between(v, 256, 287) {
		return v
	}

	return v
}

func PokeToAttribute(v int) types.VideoAttribute {
	va := types.VA_INVERSE
	if (v & 64) > 0 {
		va = types.VA_BLINK
	}
	if (v & 128) > 0 {
		va = types.VA_NORMAL
	}
	if (v & 256) > 0 {
		va = types.VA_NORMAL
	}
	return va
}

func OffsetToX(cols int, address int) int {
	if cols == 80 {
		return OffsetToX80(address)
	} else {
		return OffsetToX40(address)
	}
}

func OffsetToX40(address int) int {
	return int((address % 128) % 40)
}

func OffsetToX80(address int) int {
	var off int = 1
	if address >= 1024 {
		off = 0
	}

	return (2 * ((address % 128) % 40)) + off
}

func OffsetToY(address int) int {
	return ((address % 128) / 40 * 8) + ((address / 128) % 8)
}
