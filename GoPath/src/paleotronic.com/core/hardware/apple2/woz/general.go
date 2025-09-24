package woz

import (
	"bytes"

	"paleotronic.com/core/memory"

	"paleotronic.com/log"
)

// Apple ][ address and data prologues - standard
var addrPrologue = []byte{0xD5, 0xAA, 0x96}
var addrPrologue13 = []byte{0xD5, 0xAA, 0xb5}
var addrEpilogue = []byte{0xDE, 0xAA, 0xEB}
var dataPrologue = []byte{0xD5, 0xAA, 0xAD}
var dataEpilogue = []byte{0xDE, 0xAA, 0xEB}
var addrEpilogueFallback = []byte{0xDE, 0xAA}

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

// NibblesToBitStream -> bytes, bytecount, bitcount
func NibblesToBitstream(nibbles []byte, syncs []byte) ([]byte, uint16, uint16) {
	out := make([]byte, 0, 6656)
	bitptr := 0

	var writeBit = func(bit byte) {
		for len(out)*8 < bitptr+8 {
			out = append(out, 0x00)
		}
		byteindex := bitptr / 8
		bitindex := uint(bitptr % 8)
		bitsetmask := byte(1 << (7 - bitindex))
		bitclrmask := ^bitsetmask
		out[byteindex] = (out[byteindex] & bitclrmask) | (bitsetmask * bit)
		bitptr++
	}

	synccount := 0
	for i, n := range nibbles {
		var needZeros = syncs[i] == 0xff
		for i := 7; i >= 0; i-- {
			if n&(1<<uint(i)) != 0 {
				writeBit(1)
			} else {
				writeBit(0)
			}
		}
		if needZeros {
			writeBit(0)
			writeBit(0)
			synccount++
		}
		// enforce limit -- base on woz spacing - probably overformat at this point
		if len(out) >= 6643 {
			break
		}
	}
	log.Printf("Sync bytes: %d", synccount)

	return out, uint16(len(out)), uint16(bitptr)
}

func CalcSyncsForNibbles(nibbles []byte) []byte {
	out := make([]byte, len(nibbles))
	syncStart := 0
	syncCount := 0
	for i, v := range nibbles {
		switch {
		case v == 0xff:
			if syncCount == 0 {
				syncStart = i
			}
			syncCount++
		case v != 0xff:
			if syncCount > 5 {
				//log.Printf("Offset %d, %d sync values", syncStart, syncCount)
				for j := 0; j < syncCount; j++ {
					out[syncStart+j] = 0xff
				}
			}
			syncCount = 0
		}
	}
	if syncCount > 5 {
		for j := 0; j < syncCount; j++ {
			out[syncStart+j] = 0xff
		}
	}
	return out
}

func StandardizeNibbleSpacing(data []byte) []byte {

	const gap1 = 48
	const aplen = 14
	const dtlen = 0x15d
	const gap23 = 33

	out := bytes.NewBuffer([]byte(nil))

	for i := 0; i < gap1; i++ {
		out.Write([]byte{0xFF})
	}

	a := bytes.Split(data, addrPrologue)
	for i, chunk := range a {
		if i == 0 {
			// skip leader
			continue
		}
		sdata := []byte{}
		sdata = append(sdata, addrPrologue...)
		for bytes.HasSuffix(chunk, []byte{0xFF}) {
			chunk = bytes.TrimSuffix(chunk, []byte{0xFF})
		}
		sdata = append(sdata, chunk...)
		gap2 := len(sdata) - aplen - dtlen
		gap3 := gap23 - gap2

		// write sector
		out.Write(sdata)
		// gap3
		for i := 0; i < gap3; i++ {
			out.Write([]byte{0xFF})
		}
	}

	b := out.Bytes()
	for len(b) < 0x1a00 {
		b = append(b, 0xff)
	}
	return b
}

// CollectUntilPattern bytes starting at pos, until we see pattern... return bytes
func CollectUntilPattern(src []byte, start int, pattern []byte, include bool) (int, int) {
	collected := make([]byte, 0, 1024)
	ptr := start - 1
	found := false
	for ptr < len(src)-1 && !found {
		ptr++
		collected = append(collected, src[ptr])
		if bytes.HasSuffix(collected, pattern) {
			found = true
			break
		}
	}
	if !found {
		return -1, -1
	}
	if found && include {
		return start, start + len(collected)
	}
	return start, start + len(collected) - len(pattern)
}
