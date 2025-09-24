package memory

import (
	"testing"
	"paleotronic.com/fmt"
)

func TestMMCreate(t *testing.T) {

	mm := NewMemoryMap()

	if len(mm.Data) != OCTALYZER_MEMORY_SIZE {
		t.Errorf( "Global memory size wrong, expected %d, got %d", OCTALYZER_MEMORY_SIZE, len(mm.Data) )
	}

}

func TestMMAddMapping(t *testing.T) {

	mm := NewMemoryMap()

	mr := NewMappedRegion(
		mm,
		49152,
		256,
		"Apple2IOChip",
		make(map[int]ReadSubscriptionHandler),
		make(map[int]WriteSubscriptionHandler))

	mm.MapInterpreterRegion(0, MemoryRange{Base: mr.Base, Size: mr.Size}, mr)

	mp, ok := mm.InterpreterMappableAtAddress(0, 49200)
	if !ok {
		t.Error("No mappable found within expected int address range")
	}

	//fmt.Printf("Found mapped object: %s at %d-%d\n", mp.GetLabel(), mp.GetBase(), mp.GetBase()+mp.GetSize()-1)

}

func speaker(mm *MappedRegion, offset int) uint {
	//fmt.Println("--> Triggering speaker click shim")
	return 255 // so we can see if we got here
}

func TestRelativeReadHandler(t *testing.T) {

	mm := NewMemoryMap()

	mr := NewMappedRegion(mm, 49152, 256, "Apple2IOChip", make(map[int]ReadSubscriptionHandler), make(map[int]WriteSubscriptionHandler))

	mm.MapInterpreterRegion(0, MemoryRange{Base: mr.Base, Size: mr.Size}, mr)

	mr.SubscribeReadHandler(0x30, speaker)  // realtime function handler offset 0x30 into region

	v := mr.RelativeRead( 0x30 )
	v2 := mr.Read(49200)

	if v != 255 {
		t.Error( "Failed to trigger realtime READ response handler via relative address" )
	}

	if v2 != 255 {
		t.Error( "Failed to trigger realtime READ response handler via absolute address" )
	}

}

func TestRelativeWriteHandler(t *testing.T) {

	mm := NewMemoryMap()

	mr := NewMappedRegion(mm, 49152, 256, "Apple2IOChip", make(map[int]ReadSubscriptionHandler), make(map[int]WriteSubscriptionHandler))

	mm.MapInterpreterRegion(0, MemoryRange{Base: mr.Base, Size: mr.Size}, mr)

	mr.Write(49152, 77) // <-- this should flow to core memory as the slice is just a reflection

	// Check value in global
	if mm.Data[49152] != 77 {
		t.Error("Slice based update did not succeed")
	}

}

func TestGlobalWrite(t *testing.T) {

	mm := NewMemoryMap()

	mr := NewMappedRegion(mm, 49152, 256, "Apple2IOChip", make(map[int]ReadSubscriptionHandler), make(map[int]WriteSubscriptionHandler))

	mm.MapInterpreterRegion(0, MemoryRange{Base: mr.Base, Size: mr.Size}, mr)

	mm.Data[49152] = 66 // Direct to global address space

	// Check value in memory region
	if mr.Read(49152) != 66 {
		t.Error("Global based update propagation did not succeed")
	}

}

func TestFailForce(t *testing.T) {
	t.Error("forced fail to show output")
}
