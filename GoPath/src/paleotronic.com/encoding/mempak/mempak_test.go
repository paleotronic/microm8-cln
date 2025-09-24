package mempak

import "testing"
import "paleotronic.com/fmt"

func TestEncodeDecode8x8(t *testing.T) {

	var addr, daddr, count int
	var value, dvalue uint64
	var slot int
	var e error
	var read bool
	var data []byte

	slot = 1
	addr = 0xff
	value = 0xfe
	data = Encode(slot, addr, value, read)

	_, daddr, dvalue, read, count, e = Decode(data)
	if e != nil {
		t.Error(e.Error())
	}

	if count != len(data) {
		t.Error(
			fmt.Sprintf("Expected len %d, got %d", len(data), count),
		)
	}

	if daddr != addr {
		t.Error(
			fmt.Sprintf("Expected addr %d, got %d", addr, daddr),
		)
	}

	if dvalue != value {
		t.Error(
			fmt.Sprintf("Expected value %d, got %d", value, dvalue),
		)
	}
}

func TestEncodeDecode16x8(t *testing.T) {

	var addr, daddr, count int
	var value, dvalue uint64
	var slot int
	var e error
	var read bool
	var data []byte

	slot = 1
	addr = 0xaaff
	value = 0xfe
	data = Encode(slot, addr, value, read)

	_, daddr, dvalue, read, count, e = Decode(data)
	if e != nil {
		t.Error(e.Error())
	}

	if count != len(data) {
		t.Error(
			fmt.Sprintf("Expected len %d, got %d", len(data), count),
		)
	}

	if daddr != addr {
		t.Error(
			fmt.Sprintf("Expected addr %d, got %d", addr, daddr),
		)
	}

	if dvalue != value {
		t.Error(
			fmt.Sprintf("Expected value %d, got %d", value, dvalue),
		)
	}
}

func TestEncodeDecode20x8(t *testing.T) {

	var addr, daddr, count int
	var value, dvalue uint64
	var slot int
	var e error
	var data []byte
	var read bool

	slot = 3
	addr = 0xabbcc
	value = 0xfe
	data = Encode(slot, addr, value, read)

	_, daddr, dvalue, read, count, e = Decode(data)
	if e != nil {
		t.Error(e.Error())
	}

	if count != len(data) {
		t.Error(
			fmt.Sprintf("Expected len %d, got %d", len(data), count),
		)
	}

	if daddr != addr {
		t.Error(
			fmt.Sprintf("Expected addr %d, got %d", addr, daddr),
		)
	}

	if dvalue != value {
		t.Error(
			fmt.Sprintf("Expected value %d, got %d", value, dvalue),
		)
	}
}

func TestEncodeDecode8x64(t *testing.T) {

	var addr, daddr, count int
	var value, dvalue uint64
	var slot int
	var e error
	var data []byte
	var read bool

	slot = 3
	addr = 0xCCDDEE
	value = 0x1122334455667788
	data = Encode(slot, addr, value, read)

	_, daddr, dvalue, read, count, e = Decode(data)
	if e != nil {
		t.Error(e.Error())
	}

	fmt.Println(len(data))

	if count != len(data) {
		t.Error(
			fmt.Sprintf("Expected len %d, got %d", len(data), count),
		)
	}

	if daddr != addr {
		t.Error(
			fmt.Sprintf("Expected addr %d, got %d", addr, daddr),
		)
	}

	fmt.Printf("addr = %x\n", daddr)
	fmt.Printf("value = %x\n", dvalue)

	if dvalue != value {
		t.Error(
			fmt.Sprintf("Expected value %d, got %d", value, dvalue),
		)
	}
}

func TestEncodeDecode20x32(t *testing.T) {

	var addr, daddr, count int
	var value, dvalue uint64
	var slot int
	var e error
	var data []byte
	var read bool

	slot = 3
	addr = 0xabbcc
	value = 0x4411223344
	data = Encode(slot, addr, value, read)

	_, daddr, dvalue, read, count, e = Decode(data)
	if e != nil {
		t.Error(e.Error())
	}

	if count != len(data) {
		t.Error(
			fmt.Sprintf("Expected len %d, got %d", len(data), count),
		)
	}

	if daddr != addr {
		t.Error(
			fmt.Sprintf("Expected addr %d, got %d", addr, daddr),
		)
	}

	if dvalue != value {
		t.Error(
			fmt.Sprintf("Expected value %d, got %d", value, dvalue),
		)
	}
}

func TestEncodeDecode20x32Block(t *testing.T) {

	var addr, daddr int
	var value, dvalue []uint64
	var slot int
	var e error
	var data []byte

	value = make([]uint64, 34819)
	for i := 0; i < len(value); i++ {
		value[i] = 0x665500000000
	}

	slot = 3
	addr = 0xabbcc
	data = EncodeBlock(slot, addr, value)

	fmt.Printf("%d uints encoded to %d bytes\n", len(value), len(data))

	daddr, dvalue, e = DecodeBlock(data)
	if e != nil {
		t.Error(e.Error())
	}

	if daddr != addr {
		t.Error(
			fmt.Sprintf("Expected addr %d, got %d", addr, daddr),
		)
	}

	if daddr != addr {
		t.Error(
			fmt.Sprintf("Expected addr %d, got %d", addr, daddr),
		)
	}

	same := true
	for i := 0; i < len(value); i++ {
		if value[i] != dvalue[i] {
			same = false
			break
		}
	}

	if !same {
		t.Error(
			fmt.Sprintf("Expected value %d, got %d", value, dvalue),
		)
	}

}
