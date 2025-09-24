package vduproto

import (
	"bytes"
	"paleotronic.com/fmt"
	"testing"

	//"paleotronic.com/core/types"
)

func TestStreamPackUnpack(t *testing.T) {

	s := []byte{0,0,0,0,0,0,0,0}

	var b StreamPack
	b.AddSlice(s)

	fmt.Println(b.Data)

	a, _ := b.Unwind()

	fmt.Println(s)
	fmt.Println(a)

	if !bytes.Equal(s, a) {
		t.Error("Difference between RLE pack and unpack ")
	}

	//t.Error("frog")

}

func TestStreamPackUnpackUint(t *testing.T) {

	s := []uint{
		0x00000000,
		0x00011000,
		0x00000000,
		0x00000022,
	}

	data :=PackSliceUints(s)

	fmt.Println(data)
	fmt.Println(len(data))

	u, e := UnpackSliceUints(data)
	if e != nil {
		t.Error(e.Error())
	}

	fmt.Println(u)

	//t.Error("frog")

}
