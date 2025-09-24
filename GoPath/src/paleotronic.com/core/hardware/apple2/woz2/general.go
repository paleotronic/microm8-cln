package woz2

import "paleotronic.com/core/memory"

func getLE32(data memory.MemBytes, index int) uint32 {
	if data.Len()-index < 4 {
		return 0
	}
	return uint32(data.Get(index+0)) |
		(uint32(data.Get(index+1)) << 8) |
		(uint32(data.Get(index+2)) << 16) |
		(uint32(data.Get(index+3)) << 24)
}

func getLE16(data memory.MemBytes, index int) uint16 {
	return uint16(data.Get(index)) | (uint16(data.Get(index+1)) << 8)
}

func setLE16(data memory.MemBytes, index int, value uint16) {
	data.Set(index+0, byte(value&0xff))
	data.Set(index+1, byte((value>>8)&0xff))
}

func setLE32(data memory.MemBytes, index int, value uint32) {
	data.Set(index+0, byte(value&0xff))
	data.Set(index+1, byte((value>>8)&0xff))
	data.Set(index+2, byte((value>>16)&0xff))
	data.Set(index+3, byte((value>>24)&0xff))
}
