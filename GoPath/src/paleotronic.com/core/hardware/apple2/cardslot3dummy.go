package apple2

import (
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/memory"
	"paleotronic.com/fmt"
)

type Slot3DummyCard struct {
	*memory.MappedRegion
}

var SLOT3ROM = []uint64{}

func NewSlot3DummyCard(mm *memory.MemoryMap, globalbase int, base int, ent interfaces.Interpretable) *Slot3DummyCard {

	this := &Slot3DummyCard{}

	var rsh [256]memory.ReadSubscriptionHandler
	var esh [256]memory.ExecSubscriptionHandler
	var wsh [256]memory.WriteSubscriptionHandler

	this.MappedRegion = memory.NewMappedRegion(
		mm,
		ent.GetMemIndex(),
		globalbase,
		base,
		256,
		"Slot3DummyCard",
		rsh,
		esh,
		wsh,
	)

	// Write the rom to the slot..
	this.Log("Loading ROM into address %x...", base)
	index := ent.GetMemIndex()
	mm.BlockWrite(index, mm.MEMBASE(index)+base, SLOT3ROM)

	this.Init()

	return this

}

func (d *Slot3DummyCard) Log(format string, items ...interface{}) {
	fmt.Printf("Slot3DummyCard: "+format+"\n", items...)
}

func (d *Slot3DummyCard) Init() {
}

func (d *Slot3DummyCard) Read(address int) uint64 {
	return d.RelativeRead(address - d.Base)
}

func (d *Slot3DummyCard) Exec(address int) {
	d.RelativeExec(address - d.Base)
}

func (d *Slot3DummyCard) Write(address int, value uint64) {
	d.RelativeWrite(address-d.Base, value)
}

func (mr *Slot3DummyCard) RelativeWrite(offset int, value uint64) {
}

/* RelativeRead handles a read within this regions address space */
func (mr *Slot3DummyCard) RelativeRead(offset int) uint64 {
	return 0
}

func (mr *Slot3DummyCard) Reset() {

}
