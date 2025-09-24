package apple2

import (
	"sync"

	"paleotronic.com/core/hardware/servicebus"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/memory"
	"paleotronic.com/fmt"
	"paleotronic.com/log"
)

type IOType int

const (
	IOT_READ IOType = iota
	IOT_WRITE
)

func (iot IOType) String() string {
	if iot == IOT_READ {
		return "READ"
	} else {
		return "WRITE"
	}
	return "UNKNOWN"
}

type IOCard struct {
	ROM                 [256]uint64
	C8ROM               [0x800]uint64
	Name                string
	Slot                int
	mm                  *memory.MemoryMap
	index               int
	HasExpROM           bool
	IsFWHandler         bool
	m                   sync.Mutex
	injectedBusRequests []*servicebus.ServiceBusRequest
}

type SlotCard interface {
	SetROM(data []uint64)
	SetC8ROM(data []uint64)
	Init(slot int) // init card, load rom
	HandleIO(register int, value *uint64, eventType IOType)
	Done(slot int)
	Log(format string, items ...interface{})
	LoadROM(ent interfaces.Interpretable, addr int)
	LoadFW(ent interfaces.Interpretable, addr int, ref memory.Firmware)
	LoadExpROM(ent interfaces.Interpretable, addr int, slot int)
	CardName() string
	IsFW() bool
	FirmwareRead(offset int) uint64
	FirmwareWrite(offset int, value uint64)
	FirmwareExec(offset int, PC, A, X, Y, SP, P *int) int64
	RedirectedRead(offset int) uint64
	RedirectedWrite(offset int, value uint64)
	GetYAML() []byte
	SetYAML(b []byte)
	GetID() int
	Reset()
	Increment(n int)
	Decrement(n int)
	AdjustClock(n int)
	ImA() string
}

func (c *IOCard) InjectServiceBusRequest(r *servicebus.ServiceBusRequest) {
	c.m.Lock()
	defer c.m.Unlock()
	if c.injectedBusRequests == nil {
		c.injectedBusRequests = make([]*servicebus.ServiceBusRequest, 0, 16)
	}
	c.injectedBusRequests = append(c.injectedBusRequests, r)
}

func (c *IOCard) GetID() int {
	return c.Slot
}

func (c *IOCard) GetYAML() []byte {
	return []byte(nil)
}

func (c *IOCard) SetYAML(b []byte) {
	// nothing here
}

func (c *IOCard) Increment(n int) {

}

func (c *IOCard) Decrement(n int) {

}

func (c *IOCard) AdjustClock(n int) {

}

func (c *IOCard) ImA() string {
	return "generic-io"
}

func (c *IOCard) HandleServiceBusInjection(handler func(r *servicebus.ServiceBusRequest) (*servicebus.ServiceBusResponse, bool)) {
	if c.injectedBusRequests == nil || len(c.injectedBusRequests) == 0 {
		return
	}
	c.m.Lock()
	defer c.m.Unlock()
	for _, r := range c.injectedBusRequests {
		if handler != nil {
			handler(r)
		}
	}
	c.injectedBusRequests = make([]*servicebus.ServiceBusRequest, 0, 16)
}

func (c *IOCard) IsFW() bool {
	return c.IsFWHandler
}

func (c *IOCard) SetROM(data []uint64) {
	for i, v := range data {
		if i < 256 {
			c.ROM[i] = v
		}
	}
}

func (c *IOCard) SetC8ROM(data []uint64) {
	for i, v := range data {
		if i < 0x800 {
			c.C8ROM[i] = v
		}
	}
	c.HasExpROM = true
}

func (c *IOCard) SetMemory(mm *memory.MemoryMap, index int) {
	c.mm = mm
	c.index = index
}

func (c *IOCard) Log(format string, items ...interface{}) {
	log.Printf(c.Name+": "+format, items...)
}

func (c *IOCard) Init(slot int) {
	c.Log("Startup for %s in slot %d", c.Name, slot)
	c.Slot = slot
}

func (c *IOCard) Done(slot int) {
	c.Log("Shutdown for %s in slot %d", c.Name, slot)
}

func (c *IOCard) CardName() string {
	return c.Name
}

func (c *IOCard) LoadROM(ent interfaces.Interpretable, addr int) {

	mm := ent.GetMemoryMap()
	index := ent.GetMemIndex()

	//mm.BlockWrite( mm.MEMBASE(ent.GetMemIndex())+addr, c.ROM[:] )

	slot := (addr - 0xc000) / 256

	mb := memory.NewMemoryBlockROM(mm, index, mm.MEMBASE(index), addr, len(c.ROM), false, fmt.Sprintf("slot%d.rom", slot), c.ROM[:])
	if slot == 3 {
		mb.SetState("off")
	}

	mm.BlockMapper[index].Register(mb)

	c.Log("Added rom to mapper for %s at 0x%x", c.Name, addr)

}

func (c *IOCard) LoadFW(ent interfaces.Interpretable, addr int, ref memory.Firmware) {

	mm := ent.GetMemoryMap()
	index := ent.GetMemIndex()

	//mm.BlockWrite( mm.MEMBASE(ent.GetMemIndex())+addr, c.ROM[:] )

	slot := (addr - 0xc000) / 256

	mb := memory.NewMemoryBlockFirmware(mm, index, mm.MEMBASE(index), addr, len(c.ROM), false, fmt.Sprintf("slot%d.firmware", slot), ref)

	mm.BlockMapper[index].Register(mb)

	c.Log("Added firmware to mapper for %s at 0x%x", c.Name, addr)

}

func (c *IOCard) LoadExpROM(ent interfaces.Interpretable, addr int, slot int) {

	mm := ent.GetMemoryMap()
	index := ent.GetMemIndex()

	if c.HasExpROM {
		mbexp := memory.NewMemoryBlockROM(mm, index, mm.MEMBASE(index), addr, len(c.C8ROM), false, fmt.Sprintf("slotexp%d.rom", slot), c.C8ROM[:])
		mbexp.SetState("off")

		mm.BlockMapper[index].Register(mbexp)
		c.Log("Added EXPANSION rom to mapper for %s at 0x%x", c.Name, 0xc800)
	}

}

func (c *IOCard) HandleIO(register int, value *uint64, eventType IOType) {
	//c.Log("%s event for register 0x%x / %d", eventType.String(), register, value)
}

func CardFactory(name string, mm *memory.MemoryMap, index int, acb func(b bool), ent interfaces.Interpretable) SlotCard {
	switch name {
	case "softcard":
		return NewIOCardSoftCard(mm, index, ent)
	case "mousecard":
		return NewIOCardMouse(mm, index, ent)
	case "mockingboard":
		return NewIOCardMockingBoard(mm, index, ent)
	case "parallel":
		return NewIOCardParallel(mm, index, ent)
	case "smartport":
		return NewIOCardSmartPort(mm, index, ent)
	case "superserialcard":
		return NewIOCardSSC(mm, index, ent)
	case "diskiicard":
		return NewIOCardDiskII(mm, ent, index, acb)
	case "slot3dummy":
		return NewIOCardSlot3()
	case "uthernet":
		return NewIOCardUthernet(mm, index, ent)
	}
	return nil
}

func (d *IOCard) FirmwareRead(offset int) uint64 {
	return 0
}

func (d *IOCard) FirmwareWrite(offset int, value uint64) {
	//
}

func (d *IOCard) FirmwareExec(
	offset int,
	PC, A, X, Y, SP, P *int,
) int64 {
	return 0
}

func (d *IOCard) RedirectedRead(offset int) uint64 {
	return 0
}

func (d *IOCard) RedirectedWrite(offset int, value uint64) {
	//
}

func (d *IOCard) Reset() {
	// bespoke reset handler for card
}
