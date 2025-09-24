package types

import (
	"testing"

	"paleotronic.com/fmt"
)

func TestVectorMarshalUnmarshal(t *testing.T) {
	v := NewVector(
		VT_LINE,
		14,
		0, 0, 0,
		100, 100, 100,
	)
	data, _ := v.MarshalBinary()
	v2 := &Vector{}
	err := v2.UnmarshalBinary(data)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if v.Type != v2.Type {
		t.Fatalf("expected vector type to be %s, but got %s", v.Type, v2.Type)
	}

	if v.VectorCount != v2.VectorCount {
		t.Fatalf("expected vector count to be %d, but got %d", v.VectorCount, v2.VectorCount)
	}

	if v.String() != v2.String() {
		t.Fatalf("expected string to be %s, but got %s", v.String(), v2.String())
	}
}

func TestVectorListMarshalUnmarshal(t *testing.T) {
	vl := &VectorList{
		NewVector(
			VT_LINE,
			14,
			0, 0, 0,
			100, 100, 100,
		),
		NewVector(
			VT_LINE,
			14,
			100, 100, 100,
			200, 200, 200,
		),
	}
	data, err := vl.MarshalBinary(4096)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	vl2 := &VectorList{}
	err = vl2.UnmarshalBinary(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if len(*vl) != len(*vl2) {
		t.Fatalf("expected list len %d, but got %d", len(*vl), len(*vl2))
	}
	s1 := fmt.Sprintf("%+v", *vl)
	s2 := fmt.Sprintf("%+v", *vl2)
	t.Logf("s1 = %+v", *vl)
	t.Logf("s2 = %+v", *vl2)
	if s1 != s2 {
		t.Fatalf("expected list %s, but got %s", *vl, *vl2)
	}
}
