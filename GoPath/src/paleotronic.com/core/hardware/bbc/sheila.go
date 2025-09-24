package bbc

import (
	"log"
	"sync"

	"paleotronic.com/core/hardware/servicebus"

	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/hardware/cpu/mos6502"
)

type IOChipAction struct {
	Name    string
	Actions []*IOActionCommand
}

type IOActionCommand struct {
	Command string
	Params  []string
}

type SHIELA struct {
	*memory.MappedRegion
	readhandlers  map[int]*IOChipAction
	writehandlers map[int]*IOChipAction
	e             interfaces.Interpretable
	mm            *memory.MemoryMap
	index         int
	SystemVIA     *SystemVIA
	cpu           *mos6502.Core6502
	// registers
	IC32State             int
	SlowDataBusWriteValue uint64
	KeysDown              int
	KeyRow, KeyCol        int
	SysViaKbdState        [10][8]bool // col, row addressing
	m                     sync.Mutex
	injectedBusRequests   []*servicebus.ServiceBusRequest
}

func NewSHIELA(mm *memory.MemoryMap, globalbase int, base int, ent interfaces.Interpretable, misc map[string]map[interface{}]interface{}) *SHIELA {

	for ent.GetChild() != nil {
		ent = ent.GetChild()
	}

	this := &SHIELA{}

	var rsh [256]memory.ReadSubscriptionHandler
	var esh [256]memory.ExecSubscriptionHandler
	var wsh [256]memory.WriteSubscriptionHandler

	this.MappedRegion = memory.NewMappedRegion(
		mm,
		globalbase,
		base,
		256,
		"SHIELA",
		rsh,
		esh,
		wsh,
	)

	this.e = ent
	this.mm = mm
	this.index = ent.GetMemIndex()

	this.readhandlers = make(map[int]*IOChipAction)
	this.writehandlers = make(map[int]*IOChipAction)

	// Setup speaker..
	mbm := mm.BlockMapper[this.e.GetMemIndex()]
	mbm.ResetFunc = this.ResetMemory

	ent.SetCycleCounter(this)

	this.cpu = apple2helpers.GetCPU(ent)
	this.cpu.DoneFunc = this.AfterTask
	this.cpu.InitFunc = this.BeforeTask

	this.SystemVIA = NewSystemVIA(this.cpu)

	servicebus.Subscribe(this.e.GetMemIndex(), servicebus.BBCKeyPressed, this)

	// and return
	return this
}

func (this *SHIELA) BeforeTask(cpu *mos6502.Core6502) {
	//
}

func (this *SHIELA) AfterTask(cpu *mos6502.Core6502) {
	//
}

func (this *SHIELA) ResetMemory(all bool) {
	// Reset mapped memory to defaults..
	mbm := this.mm.BlockMapper[this.e.GetMemIndex()]
	main := mbm.Get("main")
	os12 := mbm.Get("rom.os12")
	basic := mbm.Get("rom.basic")
	shiela := mbm.Get("shiela")

	for bank := 0x00; bank < 0x100; bank++ {
		if bank < 0x80 {
			mbm.PageREAD[bank] = main
			mbm.PageWRITE[bank] = main
		} else if bank < 0xC0 {
			mbm.PageREAD[bank] = basic
		} else if bank == 0xFE {
			mbm.PageREAD[bank] = shiela
			mbm.PageWRITE[bank] = shiela
		} else {
			mbm.PageREAD[bank] = os12
		}
	}
}

func (this *SHIELA) Done() {

	// Any tear down needed

}

func (mr *SHIELA) ProcessEvent(name string, addr int, value *uint64, action memory.MemoryAction) (bool, bool) {
	// switch name {
	// case "0xCXXX read":
	// 	mr.AddressRead_Cxxx(addr, value)
	// 	return false, false
	// }

	return true, true
}

func (mr *SHIELA) ImA() string {
	return "SHIELA"
}

func (mr *SHIELA) Increment(n int) {
	// send clocks to system and user via chips
	mr.SystemVIA.DoCycles(n)
	mr.HandleServiceBusInjection(mr.HandleServiceBusRequest)
}

func (mr *SHIELA) Decrement(n int) {
	//
}

func (mr *SHIELA) AdjustClock(n int) {
	//
}

func (mr *SHIELA) Read(address int) uint64 {
	return mr.RelativeRead(address - mr.Base)
}

func (mr *SHIELA) Exec(address int) {
	mr.RelativeExec(address - mr.Base)
}

func (mr *SHIELA) Write(address int, value uint64) {
	mr.RelativeWrite(address-mr.Base, value)
}

func (mr *SHIELA) RelativeWrite(offset int, value uint64) {
	if offset >= mr.Size {
		return // ignore write outside our bounds
	}

	switch {
	case offset >= 0x40 && offset <= 0x4F:
		log.Printf("SHIELA write - System VIA register %d <- %.2X", offset&0xf, value&0xff)
		mr.SystemVIA.WriteRegister(offset&0xF, byte(value))
	case offset >= 0x60 && offset <= 0x6F:
		log.Printf("SHIELA write - User/Printer VIA register %d <- %.2X", offset&0xf, value&0xff)
		//mr.SystemVIA.Write(offset&0xF, int(value))
	default:
		log.Printf("SHIELA write offset (FE)%.2X <- %.2X", offset, value)
	}

}

/* RelativeRead handles a read within this regions address space */
func (mr *SHIELA) RelativeRead(offset int) uint64 {
	if offset >= mr.Size {
		return 0 // ignore read outside our bounds
	}

	mr.SystemVIA.KeyPress(5, 4)

	switch {
	case offset >= 0x40 && offset <= 0x4F:
		log.Printf("SHIELA read - System VIA register %d", offset&0xf)
		return uint64(mr.SystemVIA.ReadRegister(offset & 0xF))
	// case offset >= 0x60 && offset <= 0x6F:
	// 	log.Printf("SHIELA read - User/Printer VIA register %d", offset&0xf)
	// 	return uint64(mr.SystemVIA.ReadRegister(offset & 0xF))
	default:
		log.Printf("SHIELA read offset (FE)%.2X", offset)
	}

	return mr.Data.Read(offset)
}

func (mr *SHIELA) HandleServiceBusRequest(r *servicebus.ServiceBusRequest) (*servicebus.ServiceBusResponse, bool) {

	log.Printf("Got ServiceBusRequest(%s)", r)

	//index := mr.e.GetMemIndex()

	switch r.Type {
	case servicebus.BBCKeyPressed:
		ch := r.Payload.(rune)
		log.Printf("Keypress code = %d", ch)
		mr.SystemVIA.KeyPressRune(ch)
		// TODO: Somehow put this ready for receipt..
	}

	return &servicebus.ServiceBusResponse{
		Payload: 0,
	}, true
}

func (c *SHIELA) InjectServiceBusRequest(r *servicebus.ServiceBusRequest) {
	log.Printf("Injecting sb request: %+v", r)
	c.m.Lock()
	defer c.m.Unlock()
	if c.injectedBusRequests == nil {
		c.injectedBusRequests = make([]*servicebus.ServiceBusRequest, 0, 16)
	}
	c.injectedBusRequests = append(c.injectedBusRequests, r)
}

func (c *SHIELA) HandleServiceBusInjection(handler func(r *servicebus.ServiceBusRequest) (*servicebus.ServiceBusResponse, bool)) {
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
