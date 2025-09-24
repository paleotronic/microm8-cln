package interpreter

import (
	"testing"

	"paleotronic.com/freeze"
)

func TestSimpleDelta(t *testing.T) {

	s1 := &freeze.CPURegs{
		A:  0x01,
		X:  0x02,
		Y:  0x03,
		PC: 0x4000,
		P:  128,
	}

	s2 := &freeze.CPURegs{
		A:  0xff,
		X:  0x02,
		Y:  0x03,
		PC: 0x4002,
		P:  159,
	}

	d := getDelta(s1, s2)

	b := d.ToBytes()

	t.Logf("Delta encoded bytes 1 = %+v", b)

	d2 := &CPUDelta{}
	d2.FromBytes(b)

	b2 := d.ToBytes()

	t.Logf("Delta encoded bytes 2 = %+v", b2)

	t.Fail()

}
