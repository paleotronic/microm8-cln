package mempak

import (
	"errors"
)

/*
leader id:
 	index (0-7) = 3 bits
	asize (0-2) = 2 bits
	vsize (0-3) = 3 bits

	SSSAAVVV
	00011000 = 8 + 16 = 24 decimal

	vsize

bytes (asize+1 bytes)

value (vsize+1 bytes)
*/

/* New mempak 64 bit specifications

RAAVVV
268421
31
Address (2 bits)
	0 = 1 byte  (8 bit address)
	1 = 2 bytes (16 bit address)
	2 = 3 bytes (24 bit address)
	3 = 4 bytes (32 bit address)
Value (6 bits)
	0 = 1 byte
	1 = 2 byte
	2 = 3 byte
	3 = 4 byte
	4 = 5 byte
	5 = 6 byte
	6 = 7 byte
	7 = 8 byte

*/

func getAddrSize(addr int) int {
	switch {
	case addr >= 1<<24:
		return 3
	case addr >= 1<<16:
		return 2
	case addr >= 1<<8:
		return 1
	default:
		return 0
	}
}

func getValueSize(value uint64) int {
	switch {
	case value >= 1<<56:
		return 7
	case value >= 1<<48:
		return 6
	case value >= 1<<40:
		return 5
	case value >= 1<<32:
		return 4
	case value >= 1<<24:
		return 3
	case value >= 1<<16:
		return 2
	case value >= 1<<8:
		return 1
	default:
		return 0
	}
}

func Encode(slot int, addr int, value uint64, read bool) []byte {

	data := make([]byte, 13)

	var asize, vsize int

	asize = getAddrSize(addr)

	vsize = getValueSize(value)

	data[0] = (byte(asize) << 3) | byte(vsize)
	if read {
		data[0] = data[0] | 32
	}

	for i := 0; i < asize+1; i++ {
		data[1+i] = byte(addr & 0xff)
		addr >>= 8
	}

	var count int
	if !read {
		for i := 0; i < vsize+1; i++ {
			data[2+i+asize] = byte(value & 0xff)
			value >>= 8
		}
		count = 3 + asize + vsize
	} else {
		count = 3 + asize
	}

	return data[0:count]

}

func Decode(data []byte) (int, int, uint64, bool, int, error) {

	leader := int(data[0])
	vsize := leader & 7
	asize := (leader >> 3) & 3
	slot := 0
	var read bool = (leader & 32) != 0

	count := 3 + asize + vsize
	if read {
		count = 3 + asize
	}

	if len(data) < count {
		return slot, 0, 0, false, 0, errors.New("Too little data for mempak value")
	}

	var addr int
	for i := 0; i < asize+1; i++ {
		n := uint64(i * 8)
		addr = addr | int(data[1+i])<<n
	}

	var value uint64
	if !read {
		for i := 0; i < vsize+1; i++ {
			n := uint64(i * 8)
			value = value | uint64(data[2+i+asize])<<n
		}
	}

	return slot, addr, value, read, count, nil

}

/*
	Sequential encoding scheme: -
	=============================

	0xff = switch to 4 byte values
	0xfe = switch to 3 byte values
	0xfd = switch to 2 byte values
	0xfc = switch to 1 byte values


*/

func EncodeBlock(slot int, addr int, values []uint64) []byte {

	data := make([]byte, 3)
	data[0] = byte(addr & 0xff)
	data[1] = byte((addr >> 8) & 0xff)
	data[2] = byte((addr >> 16) & 0xff)

	b := PackSliceUints(values)

	data = append(data, b...)

	return data

}

func DecodeBlock(data []byte) (int, []uint64, error) {

	var addr int
	var out []uint64

	if len(data) < 3 {
		return addr, out, errors.New("Not enough data")
	}

	addr = int(data[0]) | (int(data[1]) << 8) | (int(data[2]) << 16)

	out, e := UnpackSliceUints(data[3:])

	return addr, out, e

}
