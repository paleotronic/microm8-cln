package apple2

import (
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/memory"
)

type Apple2ROMChip struct {
	memory.MappedRegion
}

/* Would sit up at 57344 */
func NewApple2ROMChip(mm *memory.MemoryMap, ent interfaces.Interpretable, globalbase int, base int) *Apple2ROMChip {

	this := &Apple2ROMChip{}

	var rsh [256]memory.ReadSubscriptionHandler
	var esh [256]memory.ExecSubscriptionHandler
	var wsh [256]memory.WriteSubscriptionHandler

	this.MappedRegion = *memory.NewMappedRegion(
		mm,
		ent.GetMemIndex(),
		globalbase,
		base,
		256,
		"Apple2IOChip",
		rsh,
		esh,
		wsh,
	)

	// Execs - adjusted based on 57344
	this.SubscribeExecHandler(7256, Call_64600)

	// and return
	return this
}

/* call -936 code */
func Call_64600(mm *memory.MappedRegion, offset int) {

}
