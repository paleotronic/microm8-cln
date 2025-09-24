package apple2

import (
	log2 "log"

	"paleotronic.com/core/hardware/common"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/memory"
	"paleotronic.com/fmt"
)

type IOCardUthernet struct {
	IOCard
	// data
	//
	Int       interfaces.Interpretable
	terminate chan bool
	w5100     *common.W5100EthernetController
}

func (d *IOCardUthernet) Init(slot int) {
	d.IOCard.Init(slot)
	d.Log("Initialising uthernet...")
	d.terminate = make(chan bool)
	d.w5100 = common.NewW5100EthernetController()
}

func (d *IOCardUthernet) Done(slot int) {
	//d.terminate <- true
	if d.w5100 != nil {
		d.w5100.Stop()
	}
}

func (d *IOCardUthernet) Handl0eIO(register int, value *uint64, eventType IOType) {

	switch eventType {
	case IOT_READ:
		switch register {
		case 0x04:
			*value = uint64(d.w5100.GetMode())
		case 0x05:
			*value = uint64(d.w5100.GetAddressHigh())
		case 0x06:
			*value = uint64(d.w5100.GetAddressLow())
		case 0x07:
			*value = uint64(d.w5100.GetDataPort())
		default:
			fmt.RPrintf("%s of register %.2x (%.2x:%s)\n", eventType.String(), register, *value, string(rune(*value-128)))
		}
	case IOT_WRITE:
		switch register {
		case 0x04:
			d.w5100.SetMode(common.W5100Mode(*value))
		case 0x05:
			d.w5100.SetAddressHigh(byte(*value))
		case 0x06:
			d.w5100.SetAddressLow(byte(*value))
		case 0x07:
			d.w5100.SetDataPort(byte(*value))
		default:
			fmt.RPrintf("%s of register %.2x (%.2x:%s)\n", eventType.String(), register, *value, string(rune(*value-128)))
		}
	}

}

func (d *IOCardUthernet) FirmwareRead(register int) uint64 {
	//d.Log("Firmware read of offset %.2x", offset)
	log2.Printf("Uthernet: FR of address %d", register)

	// var value uint64
	// switch register {
	// case 0x04:
	// 	value = uint64(d.w5100.GetMode())
	// case 0x05:
	// 	value = uint64(d.w5100.GetAddressHigh())
	// case 0x06:
	// 	value = uint64(d.w5100.GetAddressLow())
	// case 0x07:
	// 	value = uint64(d.w5100.GetDataPort())
	// default:
	// 	//fmt.RPrintf("%s of register %.2x (%.2x:%s)\n", eventType.String(), register, value, string(rune(*value-128)))
	// }

	return 0
}

func (d *IOCardUthernet) FirmwareWrite(register int, value uint64) {
	//d.Log("Firmware write to offset %.2x (value %.2x)", register, value)
	// log2.Printf("Uthernet: FW of address %d", register)

	// switch register {
	// case 0x04:
	// 	d.w5100.SetMode(common.W5100Mode(value))
	// case 0x05:
	// 	d.w5100.SetAddressHigh(byte(value))
	// case 0x06:
	// 	d.w5100.SetAddressLow(byte(value))
	// case 0x07:
	// 	d.w5100.SetDataPort(byte(value))
	// default:
	// 	//fmt.RPrintf("%s of register %.2x (%.2x:%s)\n", eventType.String(), register, value, string(rune(*value-128)))
	// }
}

func NewIOCardUthernet(mm *memory.MemoryMap, index int, ent interfaces.Interpretable) *IOCardUthernet {
	this := &IOCardUthernet{}
	this.SetMemory(mm, index)
	this.Int = ent
	//this.SetROM([]uint64{0x24, 0xEA, 0x4C})
	this.Name = "IOCardUthernet"
	// this.IsFWHandler = true

	return this
}
