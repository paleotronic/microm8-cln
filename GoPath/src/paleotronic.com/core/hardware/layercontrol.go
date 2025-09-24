package hardware

/*
LayerControl is a memory mapped layer configuration
*/

import "paleotronic.com/core/memory"

type LayerControl struct {
	memory.MappedRegion
}

func NewLayerControl(mm *memory.MemoryMap, index int, globalbase int, base int) *LayerControl {

	this := &LayerControl{}

	var rsh [256]memory.ReadSubscriptionHandler
	var esh [256]memory.ExecSubscriptionHandler
	var wsh [256]memory.WriteSubscriptionHandler

	this.MappedRegion = *memory.NewMappedRegion(
		mm,
		index,
		globalbase,
		base,
		256,
		"LayerControl",
		rsh,
		esh,
		wsh,
	)

	// and return
	return this
}

func (l *LayerControl) SetIndex(index uint64) {
	l.Data.Write(0, index)
}

func (l *LayerControl) GetIndex() uint64 {
	return l.Data.Read(0)
}
