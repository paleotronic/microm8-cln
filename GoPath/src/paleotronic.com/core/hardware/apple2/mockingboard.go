package apple2

import (
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/hardware/common"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/settings"
	"paleotronic.com/fmt"
)

const MBSampleRate = 48000
const MBSamplesBuffer = 4096

type IOCardMockingBoard struct {
	IOCard
	// data
	controllers     [2]*common.R6522
	chips           [2]*common.AY38910
	Int             interfaces.Interpretable
	terminate       chan bool
	cyclesPerSample float64
	cycleCount      float64
	sampleCount     int
	nzCount         int
	buffer          [MBSamplesBuffer]float32
	buffA, buffB    [3][]float32
}

func (d *IOCardMockingBoard) Init(slot int) {
	d.IOCard.Init(slot)
	d.Log("Initialising mockingboard...")
	d.terminate = make(chan bool)
	cpu := apple2helpers.GetCPU(d.Int)
	for i := 0; i < len(d.chips); i++ {
		d.chips[i] = common.NewAY38910(
			fmt.Sprintf("ay.%d", i),
			(i%1)*0x80,
			cpu.BaseSpeed,
			int(settings.SampleRate),
			0xff,
			i,
			d.Int,
		)
	}
	for i := 0; i < len(d.controllers); i++ {
		j := i
		label := fmt.Sprintf("6522.%d", j)
		d.controllers[j] = common.NewR6522(label, cpu)
		d.controllers[j].IOBaseAddress = (d.Slot * 0x100) + 0xc000 + (0x80 * j)
		//d.Int.SetCycleCounter(d.controllers[j])
		d.controllers[j].SetBindings(
			d.chips[j].SetBus,
			d.chips[j].SetControl,
			d.chips[j].GetBus,
			d.chips[j].GetBus,
		)
	}
	for i := 0; i < 3; i++ {
		d.buffA[i] = make([]float32, MBSamplesBuffer)
		d.buffB[i] = make([]float32, MBSamplesBuffer)
	}

	d.cyclesPerSample = float64(cpu.BaseSpeed) / MBSampleRate

	d.Int.SetCycleCounter(d)

	//go d.Run(slot)
}

func (d *IOCardMockingBoard) GetChip(idx int) *common.R6522 {
	return d.controllers[idx%2]
}

func (d *IOCardMockingBoard) GetYAML() []byte {
	return append(d.controllers[0].Bytes(), d.controllers[1].Bytes()...)
}

func (d *IOCardMockingBoard) SetYAML(b []byte) {
	if len(b) < 30 {
		return
	}
	d.controllers[0].FromBytes(b[0:15])
	d.controllers[1].FromBytes(b[15:30])
}

func (d IOCardMockingBoard) ImA() string {
	return "Mockingboard"
}

func (d *IOCardMockingBoard) AdjustClock(speed int) {
	//
}

func (d *IOCardMockingBoard) Decrement(cycles int) {
	//
}

func (d *IOCardMockingBoard) Increment(cycles int) {
	//
	c := cycles
	for c > 0 {
		d.controllers[0].DoCycle()
		d.controllers[1].DoCycle()
		c--
	}

}

func (d *IOCardMockingBoard) Done(slot int) {
	//d.terminate <- true
	for i := 0; i < len(d.chips); i++ {
		d.chips[i].Reset()
	}
}

// func (d *IOCardMockingBoard) Run(slot int) {
// 	<-d.terminate
// }

func (d *IOCardMockingBoard) HandleIO(register int, value *uint64, eventType IOType) {

	//fmt.RPrintf("%s of register %.2x (%.2x:%s)\n", eventType.String(), register, *value, string(rune(*value-128)))

	switch eventType {
	case IOT_READ:
		switch register {
		default:
			fmt.RPrintf("%s of register %.2x (%.2x:%s)\n", eventType.String(), register, *value, string(rune(*value-128)))
		}
	case IOT_WRITE:
		switch register {
		default:
			fmt.RPrintf("%s of register %.2x (%.2x:%s)\n", eventType.String(), register, *value, string(rune(*value-128)))
		}
	}

}

func (d *IOCardMockingBoard) FirmwareRead(offset int) uint64 {
	d.Log("Firmware read of offset %.2x", offset)

	var chip int
	var psg *common.AY38910
	for chip, psg = range d.chips {
		if psg.GetBaseReg() == (offset & 0x0f0) {
			break
		}
	}
	if chip >= 2 {
		d.Log("Could not determine which PSG to communicate to")
		return 0xff
	}
	return uint64(d.controllers[chip&1].ReadRegister(offset & 0x0f))
}

func (d *IOCardMockingBoard) FirmwareWrite(offset int, value uint64) {
	d.Log("Firmware write to offset %.2x (value %.2x)", offset, value)

	//fmt2.Printf("Firmware write to offset %.2x (value %.2x)\n", offset, value)

	var chip int
	var psg *common.AY38910
	for chip, psg = range d.chips {
		if psg.GetBaseReg() == (offset & 0x0f0) {
			break
		}
	}
	if chip >= 2 {
		d.Log("Could not determine which PSG to communicate to")
		return
	}

	d.controllers[chip&1].WriteRegister(offset&0x0f, int(value))
}

func NewIOCardMockingBoard(mm *memory.MemoryMap, index int, ent interfaces.Interpretable) *IOCardMockingBoard {
	this := &IOCardMockingBoard{}
	this.SetMemory(mm, index)
	this.Int = ent
	this.Name = "IOCardMockingBoard"
	this.IsFWHandler = true

	return this
}
