package apple2

import (
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/hardware/cpu/z80"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/types"
	"paleotronic.com/log"
)

type IOCardSoftCard struct {
	IOCard
	// data
	//
	Int interfaces.Interpretable
	cpu *z80.CoreZ80
}

func (d *IOCardSoftCard) Init(slot int) {
	d.IOCard.Init(slot)
	d.Log("Initialising z80 softcard...")
	apple2helpers.TrashZ80CPU(d.Int)
	d.cpu = apple2helpers.GetZ80CPU(d.Int)
}

func (d *IOCardSoftCard) Done(slot int) {
	apple2helpers.TrashZ80CPU(d.Int)
}

func (d *IOCardSoftCard) HandleIO(register int, value *uint64, eventType IOType) {

	//log.Printf("%s of register %.2x (%.2x:%s)\n", eventType.String(), register, *value, string(rune(*value-128)))

	// switch eventType {
	// case IOT_READ:
	// 	switch register {
	// 	default:
	// 		log.Printf("Softcard: %s of register %.2x (%.2x:%s)\n", eventType.String(), register, *value, string(rune(*value-128)))
	// 	}
	// case IOT_WRITE:
	// 	switch register {
	// 	default:
	// 		log.Printf("Softcard: %s of register %.2x (%.2x:%s)\n", eventType.String(), register, *value, string(rune(*value-128)))
	// 	}
	// }

}

func (d *IOCardSoftCard) FirmwareRead(offset int) uint64 {
	log.Printf("Softcard firmware read @ %.2x", offset)
	return 0xff
}

func (d *IOCardSoftCard) FirmwareWrite(offset int, value uint64) {
	//
	log.Printf("Softcard firmware write @ 0x%.2x < 0x%.2x", offset, value)
	os := d.Int.GetState()
	switch os {
	case types.EXEC6502:
		d.Int.SetState(types.EXECZ80)
	case types.DIRECTEXEC6502:
		d.Int.SetState(types.DIRECTEXECZ80)
	case types.EXECZ80:
		d.Int.SetState(types.EXEC6502)
	case types.DIRECTEXECZ80:
		d.Int.SetState(types.DIRECTEXEC6502)
	default:
		return
	}
	log.Printf("State change: %s -> %s", os.String(), d.Int.GetState().String())
}

func (d *IOCardSoftCard) FirmwareExec(
	offset int,
	PC, A, X, Y, SP, P *int,
) int64 {
	log.Printf("Softcard firmware exec @ %.2x", offset)
	return 1
}

func NewIOCardSoftCard(mm *memory.MemoryMap, index int, ent interfaces.Interpretable) *IOCardSoftCard {
	this := &IOCardSoftCard{}
	this.Int = ent
	this.Name = "IOCardSoftCard"
	for i, _ := range this.ROM {
		this.ROM[i] = 0xff
	}
	this.IsFWHandler = true

	return this
}
