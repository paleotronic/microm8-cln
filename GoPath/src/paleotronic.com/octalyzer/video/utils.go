package video

import (
	"paleotronic.com/core/types"
)

func (t *BaseLayer) XYToOffset80(x, y int) int {
	var offset int = 0
	if (x % 2) == 0 {
		offset = 1024
	}

	return ((y % 8) * 128) + ((y / 8) * 40) + (x / 2) + offset
}

func (t *BaseLayer) XYToOffset40(x, y int) int {
	return ((y % 8) * 128) + ((y / 8) * 40) + x
}

func (t *BaseLayer) Between(v, lo, hi uint64) bool {
	return ((v >= lo) && (v <= hi))
}

func (t *BaseLayer) PokeToAsciiBBC(v uint64, usealt bool) int {
	return int(v & 0x7f)
}

func (t *BaseLayer) PokeToAsciiApple(v uint64, usealt bool) int {
	highbit := v & 1024

	v = v & 1023

	if t.Between(v, 0, 31) {
		return int((64 + (v % 32)) | highbit)
	}

	if t.Between(v, 32, 63) {
		return int((32 + (v % 32)) | highbit)
	}

	if t.Between(v, 64, 95) {
		if usealt {
			return int((128 + (v % 32)) | highbit)
		} else {
			return int((64 + (v % 32)) | highbit)
		}
	}

	if t.Between(v, 96, 127) {
		if usealt {
			return int((96 + (v % 32)) | highbit)
		} else {
			return int((32 + (v % 32)) | highbit)
		}
	}

	if t.Between(v, 128, 159) {
		return int((64 + (v % 32)) | highbit)
	}

	if t.Between(v, 160, 191) {
		return int((32 + (v % 32)) | highbit)
	}

	if t.Between(v, 192, 223) {
		return int((64 + (v % 32)) | highbit)
	}

	if t.Between(v, 224, 255) {
		return int((96 + (v % 32)) | highbit)
	}

	return int(v | highbit)
}

func (t *BaseLayer) PokeToAttribute(v uint64, usealt bool) types.VideoAttribute {

	v = v & 1023

	va := types.VA_INVERSE
	if (v & 64) > 0 {
		if usealt {
			if v < 96 {
				va = types.VA_NORMAL
			} else {
				va = types.VA_INVERSE
			}
		} else {
			va = types.VA_BLINK
		}
	}
	if (v & 128) > 0 {
		va = types.VA_NORMAL
	}
	if (v & 256) > 0 {
		va = types.VA_NORMAL
	}
	return va
}

func (t *BaseLayer) OffsetToX(cols int, address int) int {
	if cols == 80 {
		return t.OffsetToX80(address)
	} else {
		return t.OffsetToX40(address)
	}
}

func (t *BaseLayer) OffsetToX40(address int) int {
	return int((address % 128) % 40)
}

func (t *BaseLayer) OffsetToX80(address int) int {
	var off int = 1
	if address >= 1024 {
		off = 0
	}

	return (2 * ((address % 128) % 40)) + off
}

func (t *BaseLayer) OffsetToY(address int) int {
	return ((address % 128) / 40 * 8) + ((address / 128) % 8)
}
