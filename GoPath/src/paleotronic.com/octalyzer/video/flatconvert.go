package video

var HGRToDHGR [512][256]uint64
var HGRToDHGRMono [256][256]uint64

func byteDoubler(b int) int {
	num := ((b & 0x040) << 6) |
		((b & 0x020) << 5) |
		((b & 0x010) << 4) |
		((b & 0x08) << 3) |
		((b & 0x04) << 2) |
		((b & 0x02) << 1) |
		(b & 0x01)
	return num | (num << 1)
}

func init() {

	var b1, b2 int
	for bb1 := 0; bb1 < 512; bb1++ {
		for bb2 := 0; bb2 < 256; bb2++ {
			var value int
			if (bb1 & 0x0181) >= 0x0101 {
				value = 1
			}
			b1 = byteDoubler((bb1 & 0x07f))
			if (bb1 & 0x080) != 0 {
				b1 <<= 1
			}
			b2 = byteDoubler((bb2 & 0x07f))
			if (bb2 & 0x080) != 0 {
				b2 <<= 1
			}
			if (bb1&0x040) == 0x040 && (bb2&1) != 0 {
				b2 |= 1
			}
			value |= b1 | (b2 << 14)
			if (bb2 & 0x040) != 0 {
				value |= 0x10000000
			}
			HGRToDHGR[bb1][bb2] = uint64(value)
			mvalue := byteDoubler(bb1) | (byteDoubler(bb2) << 14)
			HGRToDHGRMono[bb1&0x0ff][bb2] = uint64(mvalue)
		}
	}

}

func ConvertHGRToDHGRPattern(b []uint64, mono bool) []uint64 {
	// input is 40 bytes of data
	out := make([]uint64, 80)
	var b1, b2 int
	var extraHalfBit = false
	var idx int
	var dhgrWord uint64
	for i := 0; i < 40; i += 2 {
		b1 = int(b[i+0])
		b2 = int(b[i+1])
		if extraHalfBit && i > 0 {
			idx = b1 | 0x0100
		} else {
			idx = b1
		}
		if mono {
			dhgrWord = HGRToDHGRMono[b1][b2]
		} else {
			dhgrWord = HGRToDHGR[idx][b2]
		}
		extraHalfBit = (dhgrWord & 0x10000000) != 0
		out[i*2+0] = dhgrWord & 0x7f
		out[i*2+1] = (dhgrWord >> 7) & 0x7f
		out[i*2+2] = (dhgrWord >> 14) & 0x7f
		out[i*2+3] = (dhgrWord >> 21) & 0x7f
	}

	return out
}

func HalfBitShift(b []uint64) []uint64 {
	var fbp uint64
	var out = make([]uint64, len(b))
	for i, v := range b {
		//v := b[i]
		out[i] = ((v << 1) & 0x7f) | fbp
		fbp = (v & 64) >> 6
	}
	return out
}

var DHGRPaletteToLores = map[int]int{
	0x0: 0x0,
	0x1: 0x1,
	0x8: 0x2,
	0x9: 0x3,
	0x4: 0x4,
	0x5: 0x5,
	0xc: 0x6,
	0xd: 0x7,
	0x2: 0x8,
	0x3: 0x9,
	0xa: 0xa,
	0xb: 0xb,
	0x6: 0xc,
	0x7: 0xd,
	0xe: 0xe,
	0xf: 0xf,
}

func ColorFlip(b []int) []int {
	out := make([]int, len(b))
	for i, v := range b {
		out[i] = rol4bit(v)
		out[i] = ((out[i] << 2) & 0xf) | ((out[i] >> 2) & 0xf)
	}
	return out
}
