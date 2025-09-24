package spectrum

import (
	"sync"

	"paleotronic.com/fmt"
	"paleotronic.com/log"

	"paleotronic.com/core/hardware/common"
	"paleotronic.com/core/hardware/servicebus"
	"paleotronic.com/core/settings"
	"paleotronic.com/octalyzer/bus"

	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/hardware/cpu/mos6502"
	"paleotronic.com/core/hardware/cpu/z80"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/memory"

	z "github.com/remogatto/z80"
)

type IOChipAction struct {
	Name    string
	Actions []*IOActionCommand
}

type IOActionCommand struct {
	Command string
	Params  []string
}

type ZXSpectrumModel int

const (
	model48k ZXSpectrumModel = iota
	model128k
	modelPlus2
)

type ZXSpectrum struct {
	*memory.MappedRegion
	readhandlers        map[int]*IOChipAction
	writehandlers       map[int]*IOChipAction
	e                   interfaces.Interpretable
	mm                  *memory.MemoryMap
	index               int
	model               ZXSpectrumModel
	cpu                 *z80.CoreZ80
	m                   sync.Mutex
	injectedBusRequests []*servicebus.ServiceBusRequest
	keyboard            *Keyboard
	memory              *ZXMemory48K
	ports               z.PortAccessor
	beeper              *ZXBeeper
	config              *SpectrumConfig
	joystickState       servicebus.JoyLine
	framecounter        uint64
	events              []*servicebus.ServiceBusRequest
	delay_table         []byte
	pageState           byte
	ay38910             [2]*common.AY38910
	ayReg               int
	border              int
	loadedFile          string
}

func NewZXSpectrum(mm *memory.MemoryMap, globalbase int, base int, ent interfaces.Interpretable, misc map[string]map[interface{}]interface{}) *ZXSpectrum {

	for ent.GetChild() != nil {
		ent = ent.GetChild()
	}

	this := &ZXSpectrum{}

	settings.UnifiedRender[ent.GetMemIndex()] = false

	var rsh [256]memory.ReadSubscriptionHandler
	var esh [256]memory.ExecSubscriptionHandler
	var wsh [256]memory.WriteSubscriptionHandler

	this.MappedRegion = memory.NewMappedRegion(
		mm,
		ent.GetMemIndex(),
		globalbase,
		base,
		256,
		"ZXSpectrum",
		rsh,
		esh,
		wsh,
	)

	this.model = model48k
	if misc != nil {
		m, ok := misc["config"]
		if ok {
			cfg, ok := m["model"]
			if ok {
				model, ok := cfg.(int)
				if ok {
					this.model = ZXSpectrumModel(model)
					log.Printf("setting spectrum model id to %d", this.model)
				}
			}
		}
	}

	switch this.model {
	case model48k:
		this.config = NewSpectrum48KConfig()
	case model128k:
		this.config = NewSpectrum128KConfig()
	}
	this.delay_table = BuildDelayTable(this.config)

	this.e = ent
	this.mm = mm
	this.index = ent.GetMemIndex()
	this.joystickState = 0

	this.readhandlers = make(map[int]*IOChipAction)
	this.writehandlers = make(map[int]*IOChipAction)
	this.events = make([]*servicebus.ServiceBusRequest, 0, 16)

	// Setup speaker..
	mbm := mm.BlockMapper[this.e.GetMemIndex()]
	mbm.ResetFunc = this.ResetMemory

	ent.SetCycleCounter(this)

	this.SetupSpeaker(this.e)
	this.keyboard = NewKeyboard(this.e, this)
	this.memory = NewZXMemory48K(this.mm, this.index, this.e, this)
	switch this.model {
	case model48k:
		this.ports = NewZXPorts48K(this.e, this)
	case model128k:
		this.ports = NewZXPorts128K(this.e, this)
	}

	log.Printf("ay cpu clock is %d", int64(settings.CPUClock[this.e.GetMemIndex()]))

	for i, _ := range this.ay38910 {
		this.ay38910[i] = common.NewAY38910(
			fmt.Sprintf("SPECCAY%d", i),
			i*0x80,
			int64(settings.CPUClock[this.e.GetMemIndex()])/2,
			int(settings.SampleRate),
			0xff,
			i,
			this.e,
		)
		this.ay38910[i].Reset()
	}

	this.cpu = apple2helpers.GetZ80CPU(this.e)
	this.cpu.Z80().SetMemoryAccessor(this.memory)
	this.cpu.Z80().SetPortAccessor(this.ports)

	servicebus.Subscribe(this.e.GetMemIndex(), servicebus.SpectrumLoadSnapshot, this)
	servicebus.Subscribe(this.e.GetMemIndex(), servicebus.SpectrumSaveSnapshot, this)
	servicebus.Subscribe(this.e.GetMemIndex(), servicebus.KeyEvent, this)
	servicebus.Subscribe(this.e.GetMemIndex(), servicebus.JoyEvent, this)
	servicebus.Subscribe(this.e.GetMemIndex(), servicebus.Z80SpeedChange, this)

	this.ResetMemory(true)

	this.DumpZero()

	bus.StopClock()

	// and return
	return this
}

func BuildDelayTable(config *SpectrumConfig) []byte {
	var delay_table = make([]byte, config.TStatesPerFrame+100)
	tstate := config.FirstScreenByte - 1
	for y := 0; y < config.ScreenHeight; y++ {
		for x := 0; x < config.ScreenWidth; x += 16 {
			tstate_x := x / config.PixelsPerTState
			delay_table[tstate+tstate_x+0] = 6
			delay_table[tstate+tstate_x+1] = 5
			delay_table[tstate+tstate_x+2] = 4
			delay_table[tstate+tstate_x+3] = 3
			delay_table[tstate+tstate_x+4] = 2
			delay_table[tstate+tstate_x+5] = 1
		}
		tstate += config.TStatesPerLine
	}
	return delay_table
}

func (this *ZXSpectrum) SetupSpeaker(ent interfaces.Interpretable) {
	if settings.UseHQAudio {
		this.beeper = NewZXBeeperHQ(
			ent,
			"OUT",
			0,
			44100,
			3494400,
			64,
			32,
		)
	} else {

		this.beeper = NewZXBeeper(
			ent,
			"OUT",
			0,
			22050,
			3494400,
			settings.BPSBufferSize,
		)
	}
	this.beeper.Bind(
		this.e.PassWaveBuffer,
		this.e.PassWaveBufferNB,
		this.e.PassWaveBufferCompressed,
	)
}

func (this *ZXSpectrum) DumpZero() {
	for i := 0; i < 256; i++ {
		if i%16 == 0 {
			if i > 0 {
				fmt.Println()
			}
			fmt.Printf("%.4x:", i)
		}

		var value uint64
		this.mm.BlockMapper[this.index].Do(i, memory.MA_READ, &value)
		fmt.Printf(" %.2x", value)
	}
	fmt.Println()
}

func (this *ZXSpectrum) BeforeTask(cpu *mos6502.Core6502) {
	//
}

func (this *ZXSpectrum) AfterTask(cpu *mos6502.Core6502) {
	//
}

func (this *ZXSpectrum) ResetMemory(all bool) {
	switch this.model {
	case model48k:
		this.ResetMemory48K(all)
	case model128k:
		this.ResetMemory128K(all)
	}
}

func (this *ZXSpectrum) ConfigureMemory128K(b byte) {
	if this.model != model128k {
		return
	}

	// log.Printf("setting page state to %.2x", b)
	//fmt.Printf(" PAGE BANK -> %d", b&7)
	// log.Printf(" - rom -> %d", (b&16)/16)

	// if this.pageState == b {
	// 	return
	// }

	cpu := apple2helpers.GetZ80CPU(this.e)
	if z80.TRACE {
		cpu.TraceEvent(
			"MEMORY",
			fmt.Sprintf("changing page state from %.2x to %.2x", this.pageState, b),
		)
	}

	mbm := this.mm.BlockMapper[this.e.GetMemIndex()]
	main0 := mbm.Get("main.block0")
	main1 := mbm.Get("main.block1")
	rom0 := mbm.Get("rom.basic")
	rom1 := mbm.Get("rom.dos")
	bank0 := mbm.Get("bank.0")
	bank1 := mbm.Get("bank.1")
	bank2 := mbm.Get("bank.2")
	bank3 := mbm.Get("bank.3")
	bank4 := mbm.Get("bank.4")
	bank5 := mbm.Get("bank.5")
	bank6 := mbm.Get("bank.6")
	bank7 := mbm.Get("bank.7")
	//speccy := mbm.Get("zxspectrum")

	var rom, upper *memory.MemoryBlock

	switch b & 7 {
	case 0:
		upper = bank0
	case 1:
		upper = bank1
	case 2:
		upper = bank2
	case 3:
		upper = bank3
	case 4:
		upper = bank4
	case 5:
		upper = bank5
	case 6:
		upper = bank6
	case 7:
		upper = bank7
	}

	switch b & 8 {
	case 0:
		// normal screen
		//log.Printf("normal screen selected")
		scrn := apple2helpers.GETGFX(this.e, "SCRN")
		if scrn != nil {
			scrn.SetActive(true)
		}
		scrn2 := apple2helpers.GETGFX(this.e, "SCR2")
		if scrn2 != nil {
			scrn2.SetActive(false)
		}
	case 8:
		// shadow screen
		//.Printf("shadow screen selected")
		scrn := apple2helpers.GETGFX(this.e, "SCRN")
		if scrn != nil {
			scrn.SetActive(false)
		}
		scrn2 := apple2helpers.GETGFX(this.e, "SCR2")
		if scrn2 != nil {
			scrn2.SetActive(true)
		}
	}

	switch b & 16 {
	case 0:
		//log.Printf("128k rom selected")
		rom = rom0
	case 16:
		//log.Printf("48k rom selected")
		rom = rom1
	}

	for bank := 0x00; bank < 0x100; bank++ {
		if bank < 0x40 {
			mbm.PageREAD[bank] = rom
			mbm.PageWRITE[bank] = nil
		} else if bank < 0x80 {
			mbm.PageREAD[bank] = main0
			mbm.PageWRITE[bank] = main0
			// } else if bank == 0x80 {
			// 	mbm.PageREAD[bank] = speccy
			// 	mbm.PageWRITE[bank] = speccy
		} else if bank < 0xc0 {
			mbm.PageREAD[bank] = main1
			mbm.PageWRITE[bank] = main1
		} else {
			// C0 RAM PAGING
			mbm.PageREAD[bank] = upper
			mbm.PageWRITE[bank] = upper
		}
	}

	this.pageState = b

	//log.Printf("mapped memory banks for Spectrum 128k")
}

func (this *ZXSpectrum) ResetMemory128K(all bool) {
	//Reset mapped memory to defaults..
	// mbm := this.mm.BlockMapper[this.e.GetMemIndex()]
	// main0 := mbm.Get("main.block0")
	// main1 := mbm.Get("main.block1")
	// rom := mbm.Get("rom.basic")
	// bank0 := mbm.Get("bank.0")
	// speccy := mbm.Get("zxspectrum")

	// for bank := 0x00; bank < 0x100; bank++ {
	// 	if bank < 0x40 {
	// 		mbm.PageREAD[bank] = rom
	// 		mbm.PageWRITE[bank] = nil
	// 	} else if bank < 0x80 {
	// 		mbm.PageREAD[bank] = main0
	// 		mbm.PageWRITE[bank] = main0
	// 	} else if bank == 0x80 {
	// 		mbm.PageREAD[bank] = speccy
	// 		mbm.PageWRITE[bank] = speccy
	// 	} else if bank < 0xc0 {
	// 		mbm.PageREAD[bank] = main1
	// 		mbm.PageWRITE[bank] = main1
	// 	} else {
	// 		mbm.PageREAD[bank] = bank0
	// 		mbm.PageWRITE[bank] = bank0
	// 	}
	// }

	// scrn := apple2helpers.GETGFX(this.e, "SCRN")
	// if scrn != nil {
	// 	scrn.SetActive(true)
	// }

	// log.Printf("mapped memory banks for Spectrum 128k")
	this.ConfigureMemory128K(0)
}

func (this *ZXSpectrum) ResetMemory48K(all bool) {
	// Reset mapped memory to defaults..
	mbm := this.mm.BlockMapper[this.e.GetMemIndex()]
	main := mbm.Get("main.all")
	rom := mbm.Get("rom.basic")
	speccy := mbm.Get("zxspectrum")

	for bank := 0x00; bank < 0x100; bank++ {
		if bank < 0x40 {
			mbm.PageREAD[bank] = rom
			mbm.PageWRITE[bank] = nil
		} else if bank == 0xFF {
			mbm.PageREAD[bank] = speccy
			mbm.PageWRITE[bank] = speccy
		} else {
			mbm.PageREAD[bank] = main
			mbm.PageWRITE[bank] = main
		}
	}

	scrn := apple2helpers.GETGFX(this.e, "SCRN")
	if scrn != nil {
		scrn.SetActive(true)
	}

	log.Printf("mapped memory banks for Spectrum")
}

func (this *ZXSpectrum) Done() {

	// Any tear down needed
	servicebus.UnsubscribeAll(this.e.GetMemIndex())

}

func (mr *ZXSpectrum) ProcessEvent(name string, addr int, value *uint64, action memory.MemoryAction) (bool, bool) {
	// switch name {
	// case "0xCXXX read":
	// 	mr.AddressRead_Cxxx(addr, value)
	// 	return false, false
	// }

	return true, true
}

func (mr *ZXSpectrum) ImA() string {
	return "ZXSpectrum"
}

func (mr *ZXSpectrum) Increment(n int) {
	// send clocks to system and user via chips
	mr.HandleServiceBusInjection(mr.HandleServiceBusRequest)
}

func (mr *ZXSpectrum) Decrement(n int) {
	//
}

func (mr *ZXSpectrum) AdjustClock(n int) {
	//
}

func (mr *ZXSpectrum) Read(address int) uint64 {
	return mr.RelativeRead(address - mr.Base)
}

func (mr *ZXSpectrum) Exec(address int) {
	mr.RelativeExec(address - mr.Base)
}

func (mr *ZXSpectrum) Write(address int, value uint64) {
	mr.RelativeWrite(address-mr.Base, value)
}

func (mr *ZXSpectrum) RelativeWrite(offset int, value uint64) {
	if offset >= mr.Size {
		return // ignore write outside our bounds
	}

	mr.Data.Write(offset, value)
}

/* RelativeRead handles a read within this regions address space */
func (mr *ZXSpectrum) RelativeRead(offset int) uint64 {
	if offset >= mr.Size {
		return 0 // ignore read outside our bounds
	}

	return mr.Data.Read(offset)
}

func (mr *ZXSpectrum) HandleServiceBusRequest(r *servicebus.ServiceBusRequest) (*servicebus.ServiceBusResponse, bool) {

	//log.Printf("Got ServiceBusRequest(%s)", r)

	//index := mr.e.GetMemIndex()

	switch r.Type {
	case servicebus.Z80SpeedChange:
		speed := r.Payload.(int)
		mr.beeper.AdjustClock(speed)
	case servicebus.JoyEvent:
		ev := r.Payload.(*servicebus.JoystickEventData)
		//log.Printf("Got SERVICEBUS joyevent: %+v", *ev)

		if ev.Stick == 0 {
			mr.joystickState = ev.Line
		}

	case servicebus.KeyEvent:
		ev := r.Payload.(*servicebus.KeyEventData)
		//log.Printf("Got SERVICEBUS keyevent: %+v", *ev)

		mr.keyboard.ProcessKeyEvent(ev)

	case servicebus.SpectrumLoadSnapshot:
		file := r.Payload.(string)
		mr.LoadFile(file)

	case servicebus.SpectrumSaveSnapshot:
		file := r.Payload.(string)
		mr.CreateSnapshot(file)

	default:
		// something
	}

	return &servicebus.ServiceBusResponse{
		Payload: 0,
	}, true
}

func (c *ZXSpectrum) InjectServiceBusRequest(r *servicebus.ServiceBusRequest) {
	log.Printf("Injecting sb request: %+v", r)
	c.m.Lock()
	defer c.m.Unlock()
	if c.injectedBusRequests == nil {
		c.injectedBusRequests = make([]*servicebus.ServiceBusRequest, 0, 16)
	}
	c.injectedBusRequests = append(c.injectedBusRequests, r)
}

func (c *ZXSpectrum) HandleServiceBusInjection(handler func(r *servicebus.ServiceBusRequest) (*servicebus.ServiceBusResponse, bool)) {
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
