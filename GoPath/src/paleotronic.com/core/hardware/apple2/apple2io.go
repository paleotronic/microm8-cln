package apple2

import (
	"strings"
	"sync"

	log2 "log"

	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/hardware/common"
	"paleotronic.com/core/hardware/cpu/mos6502"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/vduconst"
	"paleotronic.com/utils"

	"paleotronic.com/core/hardware/servicebus"
)

//var mm2e map[*memory.MappedRegion]interfaces.Interpretable
//var mm2ss map[*memory.MappedRegion]apple2helpers.SoftSwitchConfig

var PDLTIME float64 = 2886

const PDLTICKS = 16
const MaxScan = 312 // 262 for ntsc, but take higher

var PDL_CNTR_INTERVAL float64 = 2816.0 / 255.0

type Apple2IOChip struct {
	*memory.MappedRegion
	cards         [8]SlotCard
	readhandlers  map[int]*IOChipAction
	writehandlers map[int]*IOChipAction
	Flags         map[string]bool
	e             interfaces.Interpretable
	ss            apple2helpers.SoftSwitchConfig
	//KeyVal            uint64
	speaker, cassette *AppleSpeaker
	memmode           MemoryFlag
	vidmode           VideoFlag
	//IOSelect          int
	//IOInternalROM     int
	//PeripheralROMSlot uint64
	//C8Rom C8ROMType
	VideoSW         string
	GlobalCycles    int64
	LastC083CO8BHit int64
	LastBlankState  bool
	VerticalRetrace int64
	VBlankLength    int64
	ScanCycles      int64
	Clocks          int64
	FPS             int64
	FBAddr, FBSize  int

	// Options
	DisableSW80  bool
	DisableSWAlt bool
	DisableSWDbl bool
	DisableSWAux bool
	HQMode       bool
	UseHighROMs  bool

	//LastScanLine    int
	//LastScanSegment int
	//LastScanHPOS    int
	//ScanData        [MaxScan][ScanSegments]*ScanlineState
	//PrevScanData    [MaxScan][ScanSegments]*ScanlineState
	//ScanChanged     [MaxScan]bool
	//ScanLayerCache  [2][SMMaxMode]*types.LayerSpecMapped
	UnifiedFrame *ScanRenderer

	// Tape
	Tape *common.Cassette

	NextVideoLatch int64
	LastVidMode    VideoFlag

	sm sync.Mutex
}

const (
	REG_IOSELECT          = 0
	REG_IOINTERNALROM     = 1
	REG_PERIPHERALROMSLOT = 2
	REG_C8ROM             = 3
	REG_MEMMODE           = 4
	REG_VIDMODE           = 5
	REG_KEYVAL            = 6
)

var (
	PDL_TARGET [4]int64
	vb         int
	mutex      sync.Mutex
	ssmutex    sync.Mutex
)

func init() {
	//~ mm2e = make(map[*memory.MappedRegion]interfaces.Interpretable)
	//~ mm2ss = make(map[*memory.MappedRegion]apple2helpers.SoftSwitchConfig)
}

type IOChipAction struct {
	Name    string
	Actions []*IOActionCommand
}

type IOActionCommand struct {
	Command string
	Params  []string
}

func (mr *Apple2IOChip) IsTapeAttached() bool {
	return mr.Tape != nil
}

func (mr *Apple2IOChip) TapeDetach() {
	mr.Tape = nil
}

func (mr *Apple2IOChip) TapeStart() {
	if mr.IsTapeAttached() {
		mr.Tape.Start()
	}
}

func (mr *Apple2IOChip) TapeStop() {
	if mr.IsTapeAttached() {
		mr.Tape.Stop()
	}
}

func (mr *Apple2IOChip) TapeNext() {
	if mr.IsTapeAttached() {
		mr.Tape.NextSilence()
	}
}

func (mr *Apple2IOChip) TapePrev() {
	if mr.IsTapeAttached() {
		mr.Tape.NextSilence()
	}
}

func (mr *Apple2IOChip) TapeAttach(data []byte) error {
	cpu := apple2helpers.GetCPU(mr.e)
	if mr.IsTapeAttached() {
		mr.e.ClearCycleCounter(mr.Tape)
	}
	mr.TapeDetach()
	c, err := common.NewCassetteFromData(data, cpu.BaseSpeed)
	if err != nil {
		return err
	}
	mr.Tape = c
	mr.e.SetCycleCounter(c)
	return nil
}

func (mr *Apple2IOChip) TapeBegin() {
	if mr.IsTapeAttached() {
		mr.Tape.Begin()
	}
}

func (mr *Apple2IOChip) TapeEnd() {
	if mr.IsTapeAttached() {
		mr.Tape.End()
	}
}

func (mr *Apple2IOChip) SetRegKeyVal(v uint64) {
	mr.e.GetMemoryMap().WriteInterpreterMemory(mr.e.GetMemIndex(), memory.MICROM8_KEYCODE, uint64(v))
}

func (mr *Apple2IOChip) GetRegKeyVal() uint64 {
	return uint64(mr.e.GetMemoryMap().ReadInterpreterMemory(mr.e.GetMemIndex(), memory.MICROM8_KEYCODE))
}

func (mr *Apple2IOChip) SetRegIOSelect(v int) {
	mr.e.GetMemoryMap().WriteInterpreterMemory(mr.e.GetMemIndex(), memory.OCTALYZER_IO_BASE+REG_IOSELECT, uint64(v))
}

func (mr *Apple2IOChip) GetRegIOSelect() int {
	return int(mr.e.GetMemoryMap().ReadInterpreterMemory(mr.e.GetMemIndex(), memory.OCTALYZER_IO_BASE+REG_IOSELECT))
}

func (mr *Apple2IOChip) SetRegIOInternalROM(v int) {
	mr.e.GetMemoryMap().WriteInterpreterMemory(mr.e.GetMemIndex(), memory.OCTALYZER_IO_BASE+REG_IOINTERNALROM, uint64(v))
}

func (mr *Apple2IOChip) GetRegIOInternalROM() int {
	return int(mr.e.GetMemoryMap().ReadInterpreterMemory(mr.e.GetMemIndex(), memory.OCTALYZER_IO_BASE+REG_IOINTERNALROM))
}

func (mr *Apple2IOChip) SetRegPeripheralROMSlot(v uint64) {
	mr.e.GetMemoryMap().WriteInterpreterMemory(mr.e.GetMemIndex(), memory.OCTALYZER_IO_BASE+REG_PERIPHERALROMSLOT, uint64(v))
}

func (mr *Apple2IOChip) GetRegPeripheralROMSlot() uint64 {
	return uint64(mr.e.GetMemoryMap().ReadInterpreterMemory(mr.e.GetMemIndex(), memory.OCTALYZER_IO_BASE+REG_PERIPHERALROMSLOT))
}

// func (mr *Apple2IOChip) SetRegVidMode(v VideoFlag) {
// 	mr.e.GetMemoryMap().WriteInterpreterMemory(mr.e.GetMemIndex(), memory.OCTALYZER_IO_BASE+REG_VIDMODE, uint64(v))
// }

// func (mr *Apple2IOChip) GetRegVidMode() VideoFlag {
// 	return VideoFlag(mr.e.GetMemoryMap().ReadInterpreterMemory(mr.e.GetMemIndex(), memory.OCTALYZER_IO_BASE+REG_VIDMODE))
// }

// func (mr *Apple2IOChip) SetRegMemMode(v MemoryFlag) {
// 	mr.e.GetMemoryMap().WriteInterpreterMemory(mr.e.GetMemIndex(), memory.OCTALYZER_IO_BASE+REG_MEMMODE, uint64(v))
// }

// func (mr *Apple2IOChip) GetRegMemMode() MemoryFlag {
// 	return MemoryFlag(mr.e.GetMemoryMap().ReadInterpreterMemory(mr.e.GetMemIndex(), memory.OCTALYZER_IO_BASE+REG_MEMMODE))
// }

func (mr *Apple2IOChip) SetRegC8ROMType(v C8ROMType) {
	mr.e.GetMemoryMap().WriteInterpreterMemory(mr.e.GetMemIndex(), memory.OCTALYZER_IO_BASE+REG_C8ROM, uint64(v))
	mr.SetExpRomState(v)
	//fmt.Printf("*** Setting C8 ROM to %s\n", v.String())
}

func (mr *Apple2IOChip) GetRegC8ROMType() C8ROMType {
	return C8ROMType(mr.e.GetMemoryMap().ReadInterpreterMemory(mr.e.GetMemIndex(), memory.OCTALYZER_IO_BASE+REG_C8ROM))
}

type ROMDef struct {
	Source string
	Base   int
	Size   int
}

type SlotDef struct {
	Driver  string
	Options map[string]interface{}
	StdROM  *ROMDef
	ExpROM  *ROMDef
}

func NewApple2IOChip(mm *memory.MemoryMap, globalbase int, base int, ent interfaces.Interpretable, misc map[string]map[interface{}]interface{}, options string, clocks, retrace, vblank, scan, fps int64) *Apple2IOChip {

	for ent.GetChild() != nil {
		ent = ent.GetChild()
	}

	this := &Apple2IOChip{}

	var rsh [256]memory.ReadSubscriptionHandler
	var esh [256]memory.ExecSubscriptionHandler
	var wsh [256]memory.WriteSubscriptionHandler

	this.MappedRegion = memory.NewMappedRegion(
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

	this.e = ent

	servicebus.UnsubscribeType(
		this.e.GetMemIndex(),
		servicebus.UnifiedPlaybackSync,
	)
	servicebus.UnsubscribeType(
		this.e.GetMemIndex(),
		servicebus.Cycles6502Update,
	)
	servicebus.UnsubscribeType(
		this.e.GetMemIndex(),
		servicebus.Clocks6502Update,
	)

	this.memmode = MF_DEFAULT
	this.vidmode = VF_DEFAULT

	this.Flags = make(map[string]bool)
	this.readhandlers = make(map[int]*IOChipAction)
	this.writehandlers = make(map[int]*IOChipAction)

	f := apple2helpers.SoftSwitchConfig{}

	// Soft switch defaults - Power on behaviour
	f.SoftSwitch_MIXED = true
	f.SoftSwitch_GRAPHICS = false
	f.SoftSwitch_HIRES = false
	f.SoftSwitch_PAGE2 = false
	f.SoftSwitch_HRAMWRT = true
	f.SoftSwitch_BSRBANK2 = true
	f.SoftSwitch_SLOTC3ROM = false

	this.VerticalRetrace = retrace
	this.VBlankLength = vblank
	this.Clocks = clocks
	this.GlobalCycles = 0
	this.FPS = fps
	this.ScanCycles = scan

	this.ss = f

	var parseOptions = func(in map[interface{}]interface{}) map[string]interface{} {
		m := map[string]interface{}{}
		for key, value := range in {
			m[key.(string)] = value
		}
		return m
	}

	var parseRom = func(in map[interface{}]interface{}) *ROMDef {
		romDef := &ROMDef{}
		for key, value := range in {
			switch key.(string) {
			case "source":
				romDef.Source = common.ResolveString(ent.GetMemIndex(), value.(string))
			case "base":
				romDef.Base = value.(int)
			case "size":
				romDef.Size = value.(int)
			}
		}
		return romDef
	}

	var parseSlotDef = func(slotDefinition map[interface{}]interface{}) *SlotDef {
		slotDef := &SlotDef{}

		for keyname, value := range slotDefinition {

			switch keyname.(string) {
			case "driver":
				slotDef.Driver = value.(string)
			case "slotrom":
				slotDef.StdROM = parseRom(value.(map[interface{}]interface{}))
			case "exprom":
				slotDef.ExpROM = parseRom(value.(map[interface{}]interface{}))
			case "options":
				slotDef.Options = parseOptions(value.(map[interface{}]interface{}))
			}

		}

		return slotDef
	}

	if options != "" {
		parts := strings.Split(options, ",")
		for _, flag := range parts {
			switch flag {
			case "swalt":
				this.DisableSWAlt = true
			case "sw80":
				this.DisableSW80 = true
			case "swaux":
				this.DisableSWAux = true
			case "swdbl":
				this.DisableSWDbl = true
			}
		}
	}

	// init slot cards
	if misc != nil {
		slotmap, smok := misc["slots"]
		if smok {
			for slot, tmp := range slotmap {

				slotNum := slot.(int)
				slotDefinition := parseSlotDef(tmp.(map[interface{}]interface{}))
				//fmt.Printf("*** Slot %d: %v\n", slotNum, slotDefinition)

				var f func(b bool)

				// We want to be able to engage warp speed for diskaccess
				if slotDefinition.Driver == "diskiicard" {
					f = func(b bool) {
						cpu := apple2helpers.GetCPU(this.e)
						if b && !settings.NoDiskWarp[ent.GetMemIndex()] {
							//fmt.Println("Engaging warp mode for disk access")
							settings.MuteCPU = true
							cpu.SetWarp(8)
						} else {
							//fmt.Println("Disengaging warp mode for disk access")
							userwarp, userspeed := cpu.HasUserWarp()
							if userwarp {
								//fmt.Printf("Returning to %.2f%%\n", userspeed*100)
								cpu.SetWarpUser(userspeed)
							} else {
								//fmt.Println("Returning to 100%")
								cpu.SetWarp(1)
							}
							settings.MuteCPU = false
						}
					}
				} else {
					f = nil
				}

				settings.SetGlobalOverrides(ent.GetMemIndex(), slotDefinition.Options)

				card := CardFactory(slotDefinition.Driver, mm, ent.GetMemIndex(), f, ent)
				if card == nil {
					panic("Failed to create card " + slotDefinition.Driver)
				}

				// try load rom from a file if not a shim...
				if slotDefinition.StdROM != nil {
					stdrom, e := common.LoadData(slotDefinition.StdROM.Source, slotDefinition.StdROM.Base, slotDefinition.StdROM.Size)
					if e != nil {
						panic(e)
					}
					card.SetROM(stdrom)
				}

				if slotDefinition.ExpROM != nil {
					exprom, e := common.LoadData(slotDefinition.ExpROM.Source, slotDefinition.ExpROM.Base, slotDefinition.ExpROM.Size)
					if e != nil {
						panic(e)
					}
					card.SetC8ROM(exprom)
				}

				if card.IsFW() {
					card.LoadFW(ent, 0xc000+slotNum*256, card)
				} else {
					card.LoadROM(ent, 0xc000+256*slotNum)
				}
				card.LoadExpROM(ent, 0xc800, slotNum)
				card.Init(slotNum)
				this.cards[slotNum] = card

				//log.Printf("Created card %s in slot %d", slotDefinition.Driver, slotNum)
			}
		}

	}

	// Let's ask to know about some stuff so we can handle it
	mbm := mm.BlockMapper[this.e.GetMemIndex()]
	mbm.RegisterListener(
		&memory.MemoryListener{
			Label:  "0xCXXX read",
			Start:  0xC100,
			End:    0xCFFF,
			Type:   memory.MA_READ,
			Target: this,
		},
	)
	// mbm.RegisterListener(
	// 	&memory.MemoryListener{
	// 		Label:  "0x400",
	// 		Start:  0x400,
	// 		End:    0xBFF,
	// 		Type:   memory.MA_WRITE,
	// 		Target: this,
	// 	},
	// )
	// mbm.RegisterListener(
	// 	&memory.MemoryListener{
	// 		Label:  "0x2000",
	// 		Start:  0x2000,
	// 		End:    0x5fff,
	// 		Type:   memory.MA_WRITE,
	// 		Target: this,
	// 	},
	// )

	// Setup speaker...
	this.SetupSpeaker(ent)

	mbm.ResetFunc = this.ResetMemory
	//mbm.ResetFunc(false)
	this.HQMode = settings.UseHQAudio

	// ent.SetCycleCounter(this)

	cpu := apple2helpers.GetCPU(ent)
	cpu.DoneFunc = this.AfterTask
	cpu.InitFunc = this.BeforeTask
	cpu.IO = this

	if settings.SystemID[ent.GetMemIndex()] == "apple2c-en" {
		this.UseHighROMs = true // boot in high rom mode
	}

	this.UnifiedFrame = NewScanRenderer(ent.GetMemIndex(), ent.GetMemoryMap(), this)

	// and return
	return this
}

func (this *Apple2IOChip) Done() {
	// handler for disposing of things

	//fmt.Println("Apple2IOChip :: decommissioning hardware devices")
	for _, c := range this.cards {
		if c == nil {
			continue
		}
		//fmt.Println("Apple2IOChip :: " + c.CardName())
		c.Done(this.e.GetMemIndex())
	}

}

func (this *Apple2IOChip) CheckAudioMode() {
	if settings.UseHQAudio != this.HQMode {
		log2.Printf(
			"Audio mode change - HQ = %v",
			settings.UseHQAudio,
		)
		clock := this.speaker.CPUSpeed
		this.speaker.Done()
		this.cassette.Done()
		this.SetupSpeaker(this.e)
		this.speaker.AdjustClock(clock)
		this.cassette.AdjustClock(clock)
	}
}

func (this *Apple2IOChip) SetupSpeaker(ent interfaces.Interpretable) {
	if settings.UseHQAudio {
		this.speaker = NewAppleSpeakerHQ(
			ent,
			"OUT",
			0,
			settings.SampleRate,
			1020484,
			64,
			32, //64,
		)
		this.cassette = NewAppleSpeakerHQ(
			ent,
			"CAS",
			1,
			settings.SampleRate,
			1020484,
			64,
			20, //64,
		)

		this.speaker.Bind(
			this.e.PassWaveBuffer,
			this.e.PassWaveBufferNB,
			this.e.PassWaveBufferUncompressed,
		)
		this.cassette.Bind(
			this.e.PassWaveBuffer,
			this.e.PassWaveBufferNB,
			this.e.PassWaveBufferCompressed,
		)
	} else {

		this.speaker = NewAppleSpeaker(
			ent,
			"OUT",
			0,
			22050,
			1020484,
			settings.BPSBufferSize,
		)
		this.cassette = NewAppleSpeaker(
			ent,
			"CAS",
			1,
			22050,
			1020484,
			settings.BPSBufferSize,
		)
		this.speaker.Bind(
			this.e.PassWaveBuffer,
			this.e.PassWaveBufferNB,
			this.e.PassWaveBufferCompressed,
		)
		this.cassette.Bind(
			this.e.PassWaveBuffer,
			this.e.PassWaveBufferNB,
			this.e.PassWaveBufferCompressed,
		)
	}
	this.HQMode = settings.UseHQAudio

	//servicebus.Subscribe(
	//	this.e.GetMemIndex(),
	//	servicebus.UnifiedPlaybackSync,
	//	this,
	//)
	servicebus.Subscribe(
		this.e.GetMemIndex(),
		servicebus.Cycles6502Update,
		this,
	)
	servicebus.Subscribe(
		this.e.GetMemIndex(),
		servicebus.Clocks6502Update,
		this,
	)
}

func (mr *Apple2IOChip) GetMemMode() MemoryFlag {
	return mr.memmode
}

func (mr *Apple2IOChip) GetVidMode() VideoFlag {
	return mr.vidmode
}

func (mr *Apple2IOChip) BeforeTask(cpu *mos6502.Core6502) {

}

func (mr *Apple2IOChip) AfterTask(cpu *mos6502.Core6502) {
	//fmt.Printf("After task...")
	mr.speaker.Flush()
	mr.cassette.Flush()
}

func (mr *Apple2IOChip) ProcessEvent(name string, addr int, value *uint64, action memory.MemoryAction) (bool, bool) {
	//fmt.Printf("Apple2IOChip sees event %s, addr %d, value %d, action %s\n", name, addr, *value, action)

	// if name == "0x400" {
	// 	log2.Printf("ProcessEvent (a2io): %s, $%.4x", name, addr)
	// }

	index := mr.e.GetMemIndex()

	switch name {
	case "0xCXXX read":
		mr.AddressRead_Cxxx(addr, value)
		return false, false
	case "0x400":
		if settings.UnifiedRender[index] {
			// log2.Printf("%s handler: write to $%.4x", name, addr)
			if addr >= 0x400 && addr < 0x800 {
				mr.RegisterScanUpdates(addr-0x400, *value, []string{"TEXT"})
			} else if addr >= 0x800 && addr < 0xc00 {
				mr.RegisterScanUpdates(addr-0x800, *value, []string{"TXT2"})
			}
		}
	case "0x2000":
		if settings.UnifiedRender[index] {
			// log2.Printf("%s handler: write to $%.4x", name, addr)
			if addr >= 0x2000 && addr < 0x4000 {
				mr.RegisterScanUpdates(addr-0x2000, *value, []string{"HGR1"})
			} else if addr >= 0x4000 && addr < 0x6000 {
				mr.RegisterScanUpdates(addr-0x4000, *value, []string{"HGR2"})
			}
		}
	}

	return true, true
}

func (mr *Apple2IOChip) GetCard(index int) SlotCard {
	return mr.cards[index%8]
}

func (mr *Apple2IOChip) Reset() {
	f := apple2helpers.SoftSwitchConfig{}

	// Soft switch defaults - Power on behaviour
	f.SoftSwitch_MIXED = true
	f.SoftSwitch_GRAPHICS = false
	f.SoftSwitch_HIRES = false
	f.SoftSwitch_PAGE2 = false
	f.SoftSwitch_HRAMWRT = true
	f.SoftSwitch_BSRBANK2 = true
	f.SoftSwitch_SLOTC3ROM = false

	mr.ss = f
	mr.ConfigurePaging(true)

	mr.SetMemMode(MF_DEFAULT)
	mr.SetVidModeForce(VF_DEFAULT)
	mr.SetRegKeyVal(0)
}

func (mr *Apple2IOChip) ResetTask(cpu *mos6502.Core6502) {
	mr.Reset()
	mr.ConfigurePaging(true)
}

func (mr *Apple2IOChip) ResetMemory(skip bool) {
	mr.ConfigurePaging(true)

	//settings.PureBootCheck(mr.e.GetMemIndex())

	cpu := apple2helpers.GetCPU(mr.e)
	cpu.DoneFunc = mr.AfterTask
	cpu.InitFunc = mr.BeforeTask
	cpu.ResetFunc = mr.ResetTask

	if !settings.PureBoot(mr.e.GetMemIndex()) {
		return
	}

	cpu.SetWarpUser(1.01)
	cpu.SetWarpUser(1.00)
	cpu.CalcTiming()

	if !skip {
		mr.MemReset()
	}

	//	fmt.Println("==================================================================")
	//	fmt.Println("MEMORY HAS BEEN RESET")
	//	fmt.Println("==================================================================")
}

func (mr *Apple2IOChip) STORE80OFF(mm *memory.MappedRegion, address int, value uint64) {

	e := mr.e
	of := mr.ss
	nf := of // copy

	//apple2helpers.TEXT(e).NormalInterleave()

	nf.SoftSwitch_80STORE = false
	mr.ReconfigureMemoryMap(e, nf)

	mr.ss = nf
}

func (mr *Apple2IOChip) STORE80ON(mm *memory.MappedRegion, address int, value uint64) {

	e := mr.e
	of := mr.ss
	nf := of // copy

	//apple2helpers.TEXT(e).SwitchedInterleave()

	nf.SoftSwitch_80STORE = true
	mr.ReconfigureMemoryMap(e, nf)

	mr.ss = nf

}

func (mr *Apple2IOChip) RAMWRTOFF(mm *memory.MappedRegion, address int, value uint64) {

	e := mr.e
	of := mr.ss
	nf := of // copy

	//e.GetMemoryMap().DisableAux(e.GetMemIndex())

	nf.SoftSwitch_RAMWRT = false
	mr.ss = nf

	mr.ReconfigureMemoryMap(e, nf)

}

func (mr *Apple2IOChip) RAMWRTON(mm *memory.MappedRegion, address int, value uint64) {

	e := mr.e
	of := mr.ss
	nf := of // copy

	//e.GetMemoryMap().EnableAux(e.GetMemIndex())

	nf.SoftSwitch_RAMWRT = true
	mr.ss = nf

	mr.ReconfigureMemoryMap(e, nf)

}

func (mr *Apple2IOChip) RAMRDOFF(mm *memory.MappedRegion, address int, value uint64) {

	e := mr.e
	of := mr.ss
	nf := of // copy

	//e.GetMemoryMap().DisableAuxR(e.GetMemIndex())

	nf.SoftSwitch_RAMRD = false
	mr.ss = nf

	mr.ReconfigureMemoryMap(e, nf)

}

func (mr *Apple2IOChip) RAMRDON(mm *memory.MappedRegion, address int, value uint64) {

	e := mr.e
	of := mr.ss
	nf := of // copy

	//e.GetMemoryMap().EnableAuxR(e.GetMemIndex())

	nf.SoftSwitch_RAMRD = true
	mr.ss = nf

	mr.ReconfigureMemoryMap(e, nf)

}

func (mr *Apple2IOChip) COL80OFF(mm *memory.MappedRegion, address int, value uint64) {

	// removed mutex code

	e := mr.e
	of := mr.ss
	nf := of // copy

	nf.SoftSwitch_80COL = false

	//apple2helpers.ReconfigureSoftSwitches(e, of, nf, false)

	mr.ss = nf
	mr.ss = nf
	if settings.PureBoot(mr.e.GetMemIndex()) {
		apple2helpers.MODE40Preserve(e)
	} else {
		apple2helpers.MODE40(e)
	}

}

func (mr *Apple2IOChip) COL80ON(mm *memory.MappedRegion, address int, value uint64) {

	// removed mutex code

	e := mr.e
	of := mr.ss
	nf := of // copy

	nf.SoftSwitch_80COL = true

	//apple2helpers.ReconfigureSoftSwitches(e, of, nf, false)

	mr.ss = nf
	if settings.PureBoot(mr.e.GetMemIndex()) {
		apple2helpers.MODE80Preserve(e)
		apple2helpers.MODE80PreserveAlt(e)
	} else {
		apple2helpers.MODE80(e)
	}

}

/*
ReadSpeakerHandler -- triggered when the speaker address is peeked
*/
func (mr *Apple2IOChip) ReadSpeakerHandler(mm *memory.MappedRegion, address int) uint64 {

	//v := mm.Global.ReadGlobal(memory.OCTALYZER_SPEAKER_TOGGLE)
	//mm.Global.WriteGlobal(memory.OCTALYZER_SPEAKER_TOGGLE, v+1)

	mm.Global.WriteGlobal(mr.e.GetMemIndex(), memory.OCTALYZER_SPEAKER_MS, 3)
	mm.Global.WriteGlobal(mr.e.GetMemIndex(), memory.OCTALYZER_SPEAKER_FREQ, 86)

	return 0
}

func (mr *Apple2IOChip) ReadKeyHandler(mm *memory.MappedRegion, address int) uint64 {

	// Do something here - modify upper memory space
	// removed mutex code
	e := mr.e
	if e == nil {
		panic("invalid mm -> e lookup")
	}

	//	if mm.Data.Read(0x0000) < 128 {
	if mm.Global.KeyBufferSize(e.GetMemIndex()) > 0 {
		v := mm.Global.KeyBufferGetLatest(e.GetMemIndex())

		switch v {
		case vduconst.CSR_DOWN:
			v = 10
		case vduconst.CSR_UP:
			v = 11
		case vduconst.CSR_LEFT:
			v = 8
		case vduconst.CSR_RIGHT:
			v = 21
		}

		if v <= 127 {
			//mm.Data.Write(0x0000, 128|v)
			mr.SetRegKeyVal(128 | v)
		}
	}
	//	}

	return mr.GetRegKeyVal()
}

func (mr *Apple2IOChip) WriteSpeakerHandler(mm *memory.MappedRegion, address int, value uint64) {

	mm.Global.WriteGlobal(mr.e.GetMemIndex(), memory.OCTALYZER_SPEAKER_MS, 3)
	mm.Global.WriteGlobal(mr.e.GetMemIndex(), memory.OCTALYZER_SPEAKER_FREQ, 86)

}

func (mr *Apple2IOChip) ClearKeyStrobeHandler(mm *memory.MappedRegion, address int, value uint64) {

	// strobe (unlatch) keyboard data
	// In this case we clear but 7 of 0x0000 in direct non mapped mode
	//mm.Data.Write(0x0000, mm.Data.Read(0x0000)&0x7f)

	mr.SetRegKeyVal(mr.GetRegKeyVal() & 0x7f)

}

func (mr *Apple2IOChip) ClearKeyStrobeHandlerR(mm *memory.MappedRegion, address int) uint64 {

	// strobe (unlatch) keyboard data
	// In this case we clear but 7 of 0x0000 in direct non mapped mode

	e := mr.e
	if e == nil {
		panic("invalid mm -> e lookup")
	}

	//mm.Data.Write(0x0000, mm.Data.Read(0x0000)&0x7f)
	mr.SetRegKeyVal(mr.GetRegKeyVal() & 0x7f)

	if mm.Global.KeyBufferPeek(e.GetMemIndex()) != 0 {

		// keypressed

		v := mm.Global.KeyBufferPeek(e.GetMemIndex())

		switch v {
		case vduconst.CSR_DOWN:
			v = 10
		case vduconst.CSR_UP:
			v = 11
		case vduconst.CSR_LEFT:
			v = 8
		case vduconst.CSR_RIGHT:
			v = 21
		}

		if v <= 127 {
			//mm.Data.Write(0x0010, 128|v)
			mr.SetRegKeyVal(128 | v)
		}

		return 128 | v

	} else {
		return mr.GetRegKeyVal() & 0x7f
	}

}

func (mr *Apple2IOChip) PaddleXRead(mm *memory.MappedRegion, address int) uint64 {

	//~ i := 0
	e := mr.e

	pdlnum := (address - 0x64) & 3

	currcycles := apple2helpers.GetCPU(e).GlobalCycles

	//~ if currcycles >= PDL_TARGET[i] {
	//~ pdlval := mm.Global.IntGetPaddleValue(e.GetMemIndex(), pdlnum)
	//~ clocks := int64( float64(pdlval) * PDL_CNTR_INTERVAL )
	//~ PDL_TARGET[pdlnum] = currcycles + clocks
	//~ }

	if currcycles < PDL_TARGET[pdlnum] {
		return 0x80
	}

	return 0x00
}

func (mr *Apple2IOChip) Paddle0Read(mm *memory.MappedRegion, address int) uint64 {

	i := 0
	e := mr.e
	if apple2helpers.GetCPU(e).GlobalCycles >= PDL_TARGET[i] {
		v := mm.Global.IntGetPaddleValue(e.GetMemIndex(), i)
		ticks := PDLTICKS + int64((float64(v)/255)*PDLTIME)
		PDL_TARGET[i] = apple2helpers.GetCPU(e).GlobalCycles + ticks
		return 0x00
	}

	return 0x80
}

func (mr *Apple2IOChip) Paddle1Read(mm *memory.MappedRegion, address int) uint64 {

	i := 1
	e := mr.e
	if apple2helpers.GetCPU(e).GlobalCycles >= PDL_TARGET[i] {
		v := mm.Global.IntGetPaddleValue(e.GetMemIndex(), i)
		ticks := PDLTICKS + int64((float64(v)/255)*PDLTIME)
		PDL_TARGET[i] = apple2helpers.GetCPU(e).GlobalCycles + ticks
		return 0x00
	}

	return 0x80
}

func (mr *Apple2IOChip) Paddle2Read(mm *memory.MappedRegion, address int) uint64 {

	i := 2
	e := mr.e
	if apple2helpers.GetCPU(e).GlobalCycles >= PDL_TARGET[i] {
		v := mm.Global.IntGetPaddleValue(e.GetMemIndex(), i)
		ticks := PDLTICKS + int64((float64(v)/255)*PDLTIME)
		PDL_TARGET[i] = apple2helpers.GetCPU(e).GlobalCycles + ticks
		return 0x00
	}

	return 0x80
}

func (mr *Apple2IOChip) Paddle3Read(mm *memory.MappedRegion, address int) uint64 {

	i := 3
	e := mr.e
	if apple2helpers.GetCPU(e).GlobalCycles >= PDL_TARGET[i] {
		v := mm.Global.IntGetPaddleValue(e.GetMemIndex(), i)
		ticks := PDLTICKS + int64((float64(v)/255)*PDLTIME)
		PDL_TARGET[i] = apple2helpers.GetCPU(e).GlobalCycles + ticks
		return 0x00
	}

	return 0x80
}

func (mr *Apple2IOChip) TriggerPaddlesW(mm *memory.MappedRegion, address int, value uint64) {
	//    //fmt.Println("Trigger PDL read state")
	e := mr.e
	for i := 0; i < 4; i++ {
		v := mm.Global.IntGetPaddleValue(e.GetMemIndex(), i)
		ticks := int64((float64(v) * PDL_CNTR_INTERVAL))
		PDL_TARGET[i] = apple2helpers.GetCPU(e).GlobalCycles + ticks
	}
}

func (mr *Apple2IOChip) TriggerPaddles(mm *memory.MappedRegion, address int) uint64 {
	//    //fmt.Println("Trigger PDL read state")
	e := mr.e
	for i := 0; i < 4; i++ {
		v := mm.Global.IntGetPaddleValue(e.GetMemIndex(), i)
		ticks := int64((float64(v) * PDL_CNTR_INTERVAL))
		PDL_TARGET[i] = apple2helpers.GetCPU(e).GlobalCycles + ticks
	}

	return 0
}

func (mr *Apple2IOChip) ReadPaddleButton0(mm *memory.MappedRegion, address int) uint64 {
	e := mr.e
	if mm.Global.IntGetPaddleButton(e.GetMemIndex(), 0) == 0 {
		return 0
	}
	return 255
}

func (mr *Apple2IOChip) ReadPaddleButton1(mm *memory.MappedRegion, address int) uint64 {
	e := mr.e
	if mm.Global.IntGetPaddleButton(e.GetMemIndex(), 1) == 0 {
		return 0
	}
	return 255
}

func (mr *Apple2IOChip) ReadPaddleButton2(mm *memory.MappedRegion, address int) uint64 {
	e := mr.e
	if mm.Global.IntGetPaddleButton(e.GetMemIndex(), 2) == 0 {
		return 0
	}
	return 255
}

func (mr *Apple2IOChip) ReadPaddleButton3(mm *memory.MappedRegion, address int) uint64 {
	e := mr.e
	if mm.Global.IntGetPaddleButton(e.GetMemIndex(), 3) == 0 {
		return 0
	}
	return 255
}

func (mr *Apple2IOChip) RefreshGraphics(mm *memory.MappedRegion, address int) uint64 {
	e := mr.e
	of := mr.ss
	a := &of
	a.FromUint(e.GetMemory(0xfcff))
	mr.ss = of
	return 0
}

func (mr *Apple2IOChip) ResetSwitches(mm *memory.MappedRegion, address int) uint64 {

	// removed mutex code

	e := mr.e
	of := mr.ss
	nf := of // copy

	nf.SoftSwitch_MIXED = true
	nf.SoftSwitch_GRAPHICS = false
	nf.SoftSwitch_HIRES = false
	nf.SoftSwitch_PAGE2 = false
	nf.SoftSwitch_DoubleRes = false
	nf.SoftSwitch_80STORE = false
	nf.SoftSwitch_RAMWRT = false
	nf.SoftSwitch_RAMRD = false

	apple2helpers.ReconfigureSoftSwitches(e, of, nf, false)
	mr.ReconfigureMemoryMap(e, nf)

	mr.ss = nf

	return 0
}

func (mr *Apple2IOChip) GRAPHICSONR(mm *memory.MappedRegion, address int) uint64 {
	mr.GRAPHICSON(mm, address, 0)
	return 0
}

func (mr *Apple2IOChip) GRAPHICSON(mm *memory.MappedRegion, address int, value uint64) {

	// removed mutex code

	e := mr.e
	of := mr.ss
	nf := of // copy

	nf.SoftSwitch_GRAPHICS = true

	apple2helpers.ReconfigureSoftSwitches(e, of, nf, false)

	mr.ss = nf

}

func (mr *Apple2IOChip) GRAPHICSOFFR(mm *memory.MappedRegion, address int) uint64 {

	// removed mutex code

	e := mr.e
	of := mr.ss
	nf := of // copy

	nf.SoftSwitch_GRAPHICS = false

	apple2helpers.ReconfigureSoftSwitches(e, of, nf, false)

	mr.ss = nf

	//GRAPHICSOFF( mm, address, 0 )
	return 0
}

func (mr *Apple2IOChip) GRAPHICSOFF(mm *memory.MappedRegion, address int, value uint64) {

	// removed mutex code

	e := mr.e
	of := mr.ss
	nf := of // copy

	nf.SoftSwitch_GRAPHICS = false

	apple2helpers.ReconfigureSoftSwitches(e, of, nf, false)

	mr.ss = nf

}

func (mr *Apple2IOChip) MIXEDOFFR(mm *memory.MappedRegion, address int) uint64 {
	mr.MIXEDOFF(mm, address, 0)
	return 0
}

func (mr *Apple2IOChip) MIXEDOFF(mm *memory.MappedRegion, address int, value uint64) {

	// removed mutex code

	e := mr.e
	of := mr.ss
	nf := of // copy

	nf.SoftSwitch_MIXED = false

	apple2helpers.ReconfigureSoftSwitches(e, of, nf, false)

	mr.ss = nf

}

func (mr *Apple2IOChip) MIXEDONR(mm *memory.MappedRegion, address int) uint64 {
	mr.MIXEDON(mm, address, 0)
	return 0
}

func (mr *Apple2IOChip) MIXEDON(mm *memory.MappedRegion, address int, value uint64) {

	// removed mutex code

	e := mr.e
	of := mr.ss
	nf := of // copy

	nf.SoftSwitch_MIXED = true

	apple2helpers.ReconfigureSoftSwitches(e, of, nf, false)

	mr.ss = nf

}

func (mr *Apple2IOChip) PAGE2OFFR(mm *memory.MappedRegion, address int) uint64 {
	mr.PAGE2OFF(mm, address, 0)
	return 0
}

func (mr *Apple2IOChip) PAGE2OFF(mm *memory.MappedRegion, address int, value uint64) {

	// removed mutex code

	e := mr.e
	of := mr.ss
	nf := of // copy

	nf.SoftSwitch_PAGE2 = false
	if of.SoftSwitch_80STORE {
		mr.ReconfigureMemoryMap(e, nf)
		mr.ss = nf
		return
	}

	//apple2helpers.ReconfigureSoftSwitches(e, of, nf, false)

	//if nf.SoftSwitch_HIRES && nf.SoftSwitch_GRAPHICS {
	//// change visibility
	//apple2helpers.ChangeVisibilityGFX(e, "HGR1", true)
	//apple2helpers.ChangeVisibilityGFX(e, "HGR2", false)
	//} else {
	apple2helpers.ReconfigureSoftSwitches(e, of, nf, false)
	//}
	//	fmt.Println("HGR1")

	mr.ss = nf

}

func (mr *Apple2IOChip) PAGE2ONR(mm *memory.MappedRegion, address int) uint64 {
	mr.PAGE2ON(mm, address, 0)
	return 0
}

func (mr *Apple2IOChip) ReadVBLANK(mm *memory.MappedRegion, address int) uint64 {
	vb = ((vb + 1) % 2048)
	return uint64(vb / 8)
}

func (mr *Apple2IOChip) PAGE2ON(mm *memory.MappedRegion, address int, value uint64) {

	// removed mutex code

	e := mr.e
	of := mr.ss
	nf := of // copy

	nf.SoftSwitch_PAGE2 = true
	if of.SoftSwitch_80STORE {
		mr.ReconfigureMemoryMap(e, nf)
		mr.ss = nf
		return
	}

	//apple2helpers.ReconfigureSoftSwitches(e, of, nf, false)

	//if nf.SoftSwitch_HIRES && nf.SoftSwitch_GRAPHICS {
	//apple2helpers.ChangeVisibilityGFX(e, "HGR2", true)
	//apple2helpers.ChangeVisibilityGFX(e, "HGR1", false)
	//} else {
	apple2helpers.ReconfigureSoftSwitches(e, of, nf, false)
	//}

	//	fmt.Println("HGR2")

	mr.ss = nf

}

func (mr *Apple2IOChip) HIRESONR(mm *memory.MappedRegion, address int) uint64 {
	//	fmt.Printf(" Read set hires triggered by $%.4x read\n", address)
	mr.HIRESON(mm, address, 0)
	return 0
}

func (mr *Apple2IOChip) HIRESON(mm *memory.MappedRegion, address int, value uint64) {

	//fmt.Printf(" Write set hires triggered by $%.4x write", address)
	// removed mutex code

	e := mr.e
	of := mr.ss
	nf := of // copy

	nf.SoftSwitch_HIRES = true

	apple2helpers.ReconfigureSoftSwitches(e, of, nf, false)
	mr.ReconfigureMemoryMap(e, nf)

	mr.ss = nf

}

func (mr *Apple2IOChip) HIRESOFFR(mm *memory.MappedRegion, address int) uint64 {
	mr.HIRESOFF(mm, address, 0)
	return 0
}

func (mr *Apple2IOChip) HIRESOFF(mm *memory.MappedRegion, address int, value uint64) {

	// removed mutex code

	e := mr.e
	of := mr.ss
	nf := of // copy

	nf.SoftSwitch_HIRES = false

	apple2helpers.ReconfigureSoftSwitches(e, of, nf, false)
	mr.ReconfigureMemoryMap(e, nf)

	mr.ss = nf

}

func (mr *Apple2IOChip) DOUBLERESON(mm *memory.MappedRegion, address int, value uint64) {

	// removed mutex code

	e := mr.e
	of := mr.ss
	nf := of // copy

	nf.SoftSwitch_DoubleRes = true

	apple2helpers.ReconfigureSoftSwitches(e, of, nf, false)

	mr.ss = nf

}

func (mr *Apple2IOChip) DOUBLERESOFF(mm *memory.MappedRegion, address int, value uint64) {

	// removed mutex code

	e := mr.e
	of := mr.ss
	nf := of // copy

	nf.SoftSwitch_DoubleRes = false

	apple2helpers.ReconfigureSoftSwitches(e, of, nf, false)

	mr.ss = nf

}

func (mr *Apple2IOChip) DOUBLERESONR(mm *memory.MappedRegion, address int) uint64 {

	e := mr.e
	of := mr.ss
	nf := of // copy

	nf.SoftSwitch_DoubleRes = true

	apple2helpers.ReconfigureSoftSwitches(e, of, nf, false)

	mr.ss = nf

	return 0

}

func (mr *Apple2IOChip) DOUBLERESOFFR(mm *memory.MappedRegion, address int) uint64 {

	e := mr.e
	of := mr.ss
	nf := of // copy

	nf.SoftSwitch_DoubleRes = false

	apple2helpers.ReconfigureSoftSwitches(e, of, nf, false)

	mr.ss = nf

	return 0
}

func (mr *Apple2IOChip) GRAPHICS(mm *memory.MappedRegion, address int) uint64 {
	of := mr.ss
	if of.SoftSwitch_GRAPHICS {
		return 0
	} else {
		return 0x80
	}
}

func (mr *Apple2IOChip) ALTZP(mm *memory.MappedRegion, address int) uint64 {
	of := mr.ss
	if of.SoftSwitch_ALTZP {
		return 0x80
	} else {
		return 0x0
	}
}

func (mr *Apple2IOChip) SLOTC3ROM(mm *memory.MappedRegion, address int) uint64 {
	of := mr.ss
	if of.SoftSwitch_SLOTC3ROM {
		return 0x80
	} else {
		return 0x0
	}
}

func (mr *Apple2IOChip) INTCXROM(mm *memory.MappedRegion, address int) uint64 {
	of := mr.ss
	if of.SoftSwitch_INTCXROM {
		return 0x80
	} else {
		return 0x0
	}
}

func (mr *Apple2IOChip) BSRBANK2(mm *memory.MappedRegion, address int) uint64 {
	of := mr.ss
	if of.SoftSwitch_BSRBANK2 {
		return 0x80
	} else {
		return 0x0
	}
}

func (mr *Apple2IOChip) BSRREADRAM(mm *memory.MappedRegion, address int) uint64 {
	of := mr.ss
	if of.SoftSwitch_HRAMRD {
		return 0x80
	} else {
		return 0x0
	}
}

func (mr *Apple2IOChip) BSRWRITERAM(mm *memory.MappedRegion, address int) uint64 {
	of := mr.ss
	if of.SoftSwitch_HRAMWRT {
		return 0x80
	} else {
		return 0x0
	}
}

func (mr *Apple2IOChip) RAMWRT(mm *memory.MappedRegion, address int) uint64 {
	of := mr.ss
	if of.SoftSwitch_RAMWRT {
		return 0x80
	} else {
		return 0x0
	}
}

func (mr *Apple2IOChip) RAMRD(mm *memory.MappedRegion, address int) uint64 {
	of := mr.ss
	if of.SoftSwitch_RAMRD {
		return 0x80
	} else {
		return 0x0
	}
}

func (mr *Apple2IOChip) MIXED(mm *memory.MappedRegion, address int) uint64 {
	of := mr.ss
	if of.SoftSwitch_MIXED {
		return 0x80
	} else {
		return 0x00
	}
}

func (mr *Apple2IOChip) PAGE2(mm *memory.MappedRegion, address int) uint64 {
	of := mr.ss
	if of.SoftSwitch_PAGE2 {
		return 0x80
	} else {
		return 0x00
	}
}

func (mr *Apple2IOChip) HIRES(mm *memory.MappedRegion, address int) uint64 {
	of := mr.ss
	if of.SoftSwitch_HIRES {
		return 0x80
	} else {
		return 0x00
	}
}

func (mr *Apple2IOChip) ALTCHARSETON(mm *memory.MappedRegion, address int, value uint64) {

	// removed mutex code

	e := mr.e
	of := mr.ss
	nf := of // copy

	nf.SoftSwitch_ALTCHARSET = true

	//apple2helpers.ReconfigureSoftSwitches(e, of, nf, false)

	mr.ss = nf

	apple2helpers.UseAltChars(e, true)
	mr.e.GetMemoryMap().IntSetAltChars(mr.e.GetMemIndex(), true)

}

func (mr *Apple2IOChip) ALTCHARSETOFF(mm *memory.MappedRegion, address int, value uint64) {

	// removed mutex code

	e := mr.e
	of := mr.ss
	nf := of // copy

	nf.SoftSwitch_ALTCHARSET = false

	//apple2helpers.ReconfigureSoftSwitches(e, of, nf, false)

	mr.ss = nf

	apple2helpers.UseAltChars(e, false)
	mr.e.GetMemoryMap().IntSetAltChars(mr.e.GetMemIndex(), false)

}

/*
	__________________________    ______

| 80STORE      | OFF | ON  |  |//////|
|______________|_____|_____|  |//////|
| PAGE2        |  X  | OFF |  |//////|
|______________|_____|_____|  |______|
| HIRES        |  X  |  X  |   Active
|______________|_____|_____|   Memory
| RAMRD/RAMWRT | OFF | OFF |
|______________|_____|_____|

Figure 4A - Memory Map One-A
*/
func MemoryMapper1A(addr int, write bool) int {
	return addr
}

func MemoryMapper1B(addr int, write bool) int {
	if addr >= 0x200 && addr < 0xbfff {
		return addr + 65536
	}
	return addr
}

func MemoryMapper1C(addr int, write bool) int {
	if addr >= 0x200 && addr < 0xbfff && write {
		return addr + 65536
	}
	return addr
}

func MemoryMapper1D(addr int, write bool) int {
	if addr >= 0x200 && addr < 0xbfff && !write {
		return addr + 65536
	}
	return addr
}

func MemoryMapper2A(addr int, write bool) int {
	if addr >= 0x400 && addr < 0x7ff {
		return addr + 65536
	}
	return addr
}

func MemoryMapper2B(addr int, write bool) int {
	if addr >= 0x400 && addr < 0x7ff {
		return addr + 65536
	}
	if addr >= 0x2000 && addr < 0x3fff {
		return addr + 65536
	}
	return addr
}

func MemoryMapper3A(addr int, write bool) int {
	if addr >= 0x200 && addr < 0x3ff {
		return addr + 65536
	}
	if addr >= 0x800 && addr < 0xbfff {
		return addr + 65536
	}
	return addr
}

func MemoryMapper3B(addr int, write bool) int {
	if addr >= 0x200 && addr < 0x3ff {
		return addr + 65536
	}
	if addr >= 0x800 && addr < 0x1fff {
		return addr + 65536
	}
	if addr >= 0x4000 && addr < 0xbfff {
		return addr + 65536
	}
	return addr
}

func SetREAD(m map[string]string, name string) {
	v := m[name]
	switch v {
	case "off":
		v = "r"
	case "w":
		v = "rw"
	}
	m[name] = v
}

func SetRW(m map[string]string, name string) {
	SetREAD(m, name)
	SetWRITE(m, name)
}

func ClrRW(m map[string]string, name string) {
	ClrREAD(m, name)
	ClrWRITE(m, name)
}

func ClrREAD(m map[string]string, name string) {
	v := m[name]
	switch v {
	case "rw":
		v = "w"
	case "r":
		v = "off"
	}
	m[name] = v
}

func SetWRITE(m map[string]string, name string) {
	v := m[name]
	switch v {
	case "off":
		v = "w"
	case "r":
		v = "rw"
	}
	m[name] = v
}

func ClrWRITE(m map[string]string, name string) {
	v := m[name]
	switch v {
	case "rw":
		v = "r"
	case "w":
		v = "off"
	}
	m[name] = v
}

func INVERT(m map[string]string, a, b string) {
	aval := m[a]
	bval := m[b]

	m[a] = bval
	m[b] = aval
}

func IfDefinedRead(m map[string]string, testname, a, b string) {
	_, ok := m[testname]
	if ok {
		SetREAD(m, a)
	} else {
		SetREAD(m, b)
	}
}

func (mr *Apple2IOChip) ReconfigureMemoryMap(e interfaces.Interpretable, s apple2helpers.SoftSwitchConfig) {

	//~ fmt.Printf(
	//~ "State Change: INTC8ROM=%v, INTCXROM=%v, SLOTC3ROM=%v, ALTZP=%v, BSRBANK2=%v, HRAMRD=%v, HRAMWRT=%v, 80STORE=%v, 80VID=%v, PAGE2=%v, HIRES=%v\n",
	//~ s.SoftSwitch_INTC8ROM,
	//~ s.SoftSwitch_INTCXROM,
	//~ s.SoftSwitch_SLOTC3ROM,
	//~ s.SoftSwitch_ALTZP,
	//~ s.SoftSwitch_BSRBANK2,
	//~ s.SoftSwitch_HRAMRD,
	//~ s.SoftSwitch_HRAMWRT,
	//~ s.SoftSwitch_80STORE,
	//~ s.SoftSwitch_80COL,
	//~ s.SoftSwitch_PAGE2,
	//~ s.SoftSwitch_HIRES,
	//~ )

	mm := e.GetMemoryMap()
	index := e.GetMemIndex()
	mbm := mm.BlockMapper[index]

	MEM := mbm.GetDefaultMap()
	//~ fmt.Println("Starting with default state:", MEM)

	/*
	   if (SoftSwitches.CXROM.getState()) {
	       // Enable C1-CF to point to rom
	       activeRead.setBanks(0, 0x0F, 0x0C1, cPageRom);
	   } else {
	       // Enable C1-CF to point to slots
	       for (int slot = 1; slot <= 7; slot++) {
	           PagedMemory page = getCard(slot).map(Card::getCxRom).orElse(blank);
	           activeRead.setBanks(0, 1, 0x0c0 + slot, page);
	       }
	       if (getActiveSlot() == 0) {
	           for (int i = 0x0C8; i < 0x0D0; i++) {
	               activeRead.set(i, blank.get(0));
	           }
	       } else {
	           getCard(getActiveSlot()).ifPresent(c -> activeRead.setBanks(0, 8, 0x0c8, c.getC8Rom()));
	       }
	       if (SoftSwitches.SLOTC3ROM.isOff()) {
	           // Enable C3 to point to internal ROM
	           activeRead.setBanks(2, 1, 0x0C3, cPageRom);
	       }
	       if (SoftSwitches.INTC8ROM.isOn()) {
	           // Enable C8-CF to point to internal ROM
	           activeRead.setBanks(7, 8, 0x0C8, cPageRom);
	       }
	   }
	*/

	// Handle C1xx-CFxx memory space
	if s.SoftSwitch_INTCXROM {
		// Enable C1-CF to point to rom
		SetREAD(MEM, "rom.intcxrom")
	} else {
		//SetREAD( MEM, "rom.intcxrom" )

		// Enable C1-CF to point to slots
		IfDefinedRead(MEM, "slot1.rom", "slot1.rom", "rom.slothole1")
		IfDefinedRead(MEM, "slot2.rom", "slot2.rom", "rom.slothole2")
		IfDefinedRead(MEM, "slot3.rom", "slot3.rom", "rom.intc3rom")
		IfDefinedRead(MEM, "slot4.rom", "slot4.rom", "rom.slothole4")
		IfDefinedRead(MEM, "slot5.rom", "slot5.rom", "rom.slothole5")
		IfDefinedRead(MEM, "slot6.rom", "slot6.rom", "rom.slothole6")
		IfDefinedRead(MEM, "slot7.rom", "slot7.rom", "rom.slothole7")

		// Disabble slot 3 if SLOTC3ROM is off
		if !s.SoftSwitch_SLOTC3ROM {
			ClrREAD(MEM, "slot3.rom")
			ClrREAD(MEM, "rom.slothole3")
			SetREAD(MEM, "rom.intc3rom") // make sure we can see rom underneath
		}

		// Enable INTC8ROM if INTC8ROM is on
		if s.SoftSwitch_INTC8ROM {
			SetREAD(MEM, "rom.intc8rom")
		}
	}

	// ALTZP -- note we'll let the BSRREADRAM && BSRBANK2 switches decide the state of upper memory
	// we just turn things on or off depending...
	if s.SoftSwitch_ALTZP {

		SetRW(MEM, "aux.zp")
		SetRW(MEM, "aux.stack")

		ClrRW(MEM, "main.zp")
		ClrRW(MEM, "main.stack")

		// Are we using highram?

		if (s.SoftSwitch_HRAMRD || s.SoftSwitch_HRAMWRT) && !s.SoftSwitch_BSRBANK2 {

			// BANK 1
			switch {
			case s.SoftSwitch_HRAMRD && s.SoftSwitch_HRAMWRT: // RW

				SetRW(MEM, "aux.languagecard")
				ClrREAD(MEM, "rom.applesoft")
				ClrREAD(MEM, "rom.monitor")

			case s.SoftSwitch_HRAMRD && !s.SoftSwitch_HRAMWRT:

				SetREAD(MEM, "aux.languagecard")
				ClrREAD(MEM, "rom.applesoft")
				ClrREAD(MEM, "rom.monitor")

			case !s.SoftSwitch_HRAMRD && s.SoftSwitch_HRAMWRT:

				SetWRITE(MEM, "aux.languagecard")

			}

		} else if (s.SoftSwitch_HRAMRD || s.SoftSwitch_HRAMWRT) && s.SoftSwitch_BSRBANK2 {

			// BANK 2
			switch {
			case s.SoftSwitch_HRAMRD && s.SoftSwitch_HRAMWRT: // RW

				SetRW(MEM, "aux.languagecard")
				SetRW(MEM, "aux.lcbank2")
				ClrREAD(MEM, "rom.applesoft")
				ClrREAD(MEM, "rom.monitor")

			case s.SoftSwitch_HRAMRD && !s.SoftSwitch_HRAMWRT:

				SetREAD(MEM, "aux.languagecard")
				SetREAD(MEM, "aux.lcbank2")
				ClrREAD(MEM, "rom.applesoft")
				ClrREAD(MEM, "rom.monitor")

			case !s.SoftSwitch_HRAMRD && s.SoftSwitch_HRAMWRT:

				SetWRITE(MEM, "aux.languagecard")
				SetWRITE(MEM, "aux.lcbank2")

			}

		}

	} else {

		// Are we using highram?

		if (s.SoftSwitch_HRAMRD || s.SoftSwitch_HRAMWRT) && !s.SoftSwitch_BSRBANK2 {

			// BANK 1
			switch {
			case s.SoftSwitch_HRAMRD && s.SoftSwitch_HRAMWRT: // RW

				SetRW(MEM, "main.languagecard")
				ClrREAD(MEM, "rom.applesoft")
				ClrREAD(MEM, "rom.monitor")

			case s.SoftSwitch_HRAMRD && !s.SoftSwitch_HRAMWRT:

				SetREAD(MEM, "main.languagecard")
				ClrREAD(MEM, "rom.applesoft")
				ClrREAD(MEM, "rom.monitor")

			case !s.SoftSwitch_HRAMRD && s.SoftSwitch_HRAMWRT:

				SetWRITE(MEM, "main.languagecard")

			}

		} else if (s.SoftSwitch_HRAMRD || s.SoftSwitch_HRAMWRT) && s.SoftSwitch_BSRBANK2 {

			// BANK 2
			switch {
			case s.SoftSwitch_HRAMRD && s.SoftSwitch_HRAMWRT: // RW

				SetRW(MEM, "main.languagecard")
				SetRW(MEM, "main.lcbank2")
				ClrREAD(MEM, "rom.applesoft")
				ClrREAD(MEM, "rom.monitor")

			case s.SoftSwitch_HRAMRD && !s.SoftSwitch_HRAMWRT: // RO

				SetREAD(MEM, "main.languagecard")
				SetREAD(MEM, "main.lcbank2")
				ClrREAD(MEM, "rom.applesoft")
				ClrREAD(MEM, "rom.monitor")

			case !s.SoftSwitch_HRAMRD && s.SoftSwitch_HRAMWRT: // WO

				SetWRITE(MEM, "main.languagecard")
				SetWRITE(MEM, "main.lcbank2")

			}

		}

	}

	// RAMRD

	switch {

	case !s.SoftSwitch_RAMRD && !s.SoftSwitch_RAMWRT:

		// RAMRD & RAMWRT are off... defaults apply

	case s.SoftSwitch_RAMRD && s.SoftSwitch_RAMWRT:

		// RAMRD & RAMWRT are on...
		SetRW(MEM, "aux.main.b1")
		SetRW(MEM, "aux.main.text")
		SetRW(MEM, "aux.main.b2")
		SetRW(MEM, "aux.main.hgr1")
		SetRW(MEM, "aux.main.b3")

	case s.SoftSwitch_RAMRD:

		// RAMRD is on

		SetREAD(MEM, "aux.main.b1")
		SetREAD(MEM, "aux.main.text")
		SetREAD(MEM, "aux.main.b2")
		SetREAD(MEM, "aux.main.hgr1")
		SetREAD(MEM, "aux.main.b3")

	case s.SoftSwitch_RAMWRT:

		SetWRITE(MEM, "aux.main.b1")
		SetWRITE(MEM, "aux.main.text")
		SetWRITE(MEM, "aux.main.b2")
		SetWRITE(MEM, "aux.main.hgr1")
		SetWRITE(MEM, "aux.main.b3")

	}

	// handle 80STORE
	if s.SoftSwitch_80STORE {

		INVERT(MEM, "main.main.text", "aux.main.text")

		if s.SoftSwitch_HIRES {

			INVERT(MEM, "main.main.hgr1", "aux.main.hgr1")

		}

		if s.SoftSwitch_PAGE2 {

			INVERT(MEM, "aux.main.b1", "main.main.b1")
			INVERT(MEM, "aux.main.text", "main.main.text")
			INVERT(MEM, "aux.main.b2", "main.main.b2")
			INVERT(MEM, "aux.main.hgr1", "main.main.hgr1")
			INVERT(MEM, "aux.main.b3", "main.main.b3")

		}

	}

	//~ if f != nil {
	//~ e.GetMemoryMap().SetMapper( e.GetMemIndex(), f )
	//~ }

	//fmt.Println("AFTER:", MEM)

	mbm.SetMap(MEM)
	//mbm.DumpMap()

}

func (mr *Apple2IOChip) ExecuteActions(offset int, value *uint64, act *IOChipAction) bool {

	// removed mutex code
	e := mr.e
	index := e.GetMemIndex()
	m := e.GetMemoryMap()

	mbm := m.BlockMapper[index]

	cont := false

	for _, a := range act.Actions {
		switch a.Command {
		case "flag.clr":
			//~ fmt.Printf("CLEAR FLAG %s\n", a.Params[0])
			delete(mr.Flags, a.Params[0])
			//~ fmt.Println(mr.Flags)
		case "flag.set":
			//~ fmt.Printf("SET FLAG %s\n", a.Params[0])
			mr.Flags[a.Params[0]] = true
			//~ fmt.Println(mr.Flags)
		case "continue":
			cont = true
		case "mem.on":
			mbm.EnableBlocks(a.Params)
		case "mem.off":
			mbm.DisableBlocks(a.Params)
		case "mem.setread":
			mbm.EnableRead(a.Params)
		case "mem.clrread":
			mbm.DisableRead(a.Params)
		case "mem.setwrite":
			mbm.EnableWrite(a.Params)
		case "mem.clrwrite":
			mbm.DisableWrite(a.Params)
		case "mem.ison":
			if len(a.Params) == 3 {
				target := a.Params[0]
				tval := uint64(utils.StrToInt(a.Params[1]))
				fval := uint64(utils.StrToInt(a.Params[2]))
				if mbm.IsEnabled(target) {
					*value = tval
				} else {
					*value = fval
				}
			}
		case "flag.isset":
			if len(a.Params) == 3 {
				target := a.Params[0]
				tval := uint64(utils.StrToInt(a.Params[1]))
				fval := uint64(utils.StrToInt(a.Params[2]))

				set, ok := mr.Flags[target]

				if ok && set {
					*value = tval
				} else {
					*value = fval
				}
			}
		case "mem.canread":
			if len(a.Params) == 3 {
				target := a.Params[0]
				tval := uint64(utils.StrToInt(a.Params[1]))
				fval := uint64(utils.StrToInt(a.Params[2]))
				if mbm.IsReadable(target) {
					*value = tval
				} else {
					*value = fval
				}
			}
		case "mem.canwrite":
			if len(a.Params) == 3 {
				target := a.Params[0]
				tval := uint64(utils.StrToInt(a.Params[1]))
				fval := uint64(utils.StrToInt(a.Params[2]))
				if mbm.IsWritable(target) {
					*value = tval
				} else {
					*value = fval
				}
			}
		default:
			panic("unknown IOChip command: " + a.Command)
		}
	}

	return cont

}

func (mr *Apple2IOChip) Read(address int) uint64 {
	return mr.RelativeRead(address - mr.Base)
}

func (mr *Apple2IOChip) Exec(address int) {
	mr.RelativeExec(address - mr.Base)
}

func (mr *Apple2IOChip) Write(address int, value uint64) {
	mr.RelativeWrite(address-mr.Base, value)
}

func (mr *Apple2IOChip) ALTZPOFF(mm *memory.MappedRegion, address int, value uint64) {

	e := mr.e
	of := mr.ss
	nf := of // copy

	//e.GetMemoryMap().DisableAux(e.GetMemIndex())

	nf.SoftSwitch_ALTZP = false
	mr.ss = nf

	mr.ReconfigureMemoryMap(e, nf)

}

func (mr *Apple2IOChip) ALTZPON(mm *memory.MappedRegion, address int, value uint64) {

	e := mr.e
	of := mr.ss
	nf := of // copy

	//e.GetMemoryMap().DisableAux(e.GetMemIndex())

	nf.SoftSwitch_ALTZP = true
	mr.ss = nf

	mr.ReconfigureMemoryMap(e, nf)

}

func (mr *Apple2IOChip) INTCXROMOFF(mm *memory.MappedRegion, address int, value uint64) {

	e := mr.e
	of := mr.ss
	nf := of // copy

	//e.GetMemoryMap().DisableAux(e.GetMemIndex())

	//~ fmt.Println("CX ROM OFF CALLED!!!!")

	nf.SoftSwitch_INTCXROM = false
	mr.ss = nf

	mr.ReconfigureMemoryMap(e, nf)

}

func (mr *Apple2IOChip) INTCXROMON(mm *memory.MappedRegion, address int, value uint64) {

	e := mr.e
	of := mr.ss
	nf := of // copy

	//e.GetMemoryMap().DisableAux(e.GetMemIndex())

	nf.SoftSwitch_INTCXROM = true
	mr.ss = nf

	mr.ReconfigureMemoryMap(e, nf)

}

func (mr *Apple2IOChip) SLOTC3ROMOFF(mm *memory.MappedRegion, address int, value uint64) {

	e := mr.e
	of := mr.ss
	nf := of // copy

	//e.GetMemoryMap().DisableAux(e.GetMemIndex())

	nf.SoftSwitch_SLOTC3ROM = false
	mr.ss = nf

	mr.ReconfigureMemoryMap(e, nf)

}

func (mr *Apple2IOChip) SLOTC3ROMON(mm *memory.MappedRegion, address int, value uint64) {

	e := mr.e
	of := mr.ss
	nf := of // copy

	//e.GetMemoryMap().DisableAux(e.GetMemIndex())

	nf.SoftSwitch_SLOTC3ROM = true
	mr.ss = nf

	mr.ReconfigureMemoryMap(e, nf)

}

func (mr *Apple2IOChip) STORE80(mm *memory.MappedRegion, address int) uint64 {
	of := mr.ss
	if of.SoftSwitch_80STORE {
		return 0x80
	} else {
		return 0x0
	}
}

func (mr *Apple2IOChip) ALTCHARSET(mm *memory.MappedRegion, address int) uint64 {
	of := mr.ss
	if of.SoftSwitch_ALTCHARSET {
		return 0x80
	} else {
		return 0x0
	}
}

func (mr *Apple2IOChip) COL80(mm *memory.MappedRegion, address int) uint64 {
	of := mr.ss
	if of.SoftSwitch_80COL {
		return 0x80
	} else {
		return 0x0
	}
}

func (mr *Apple2IOChip) BSRBANK2OFF(mm *memory.MappedRegion, address int, value uint64) {

	e := mr.e
	of := mr.ss
	nf := of // copy

	//e.GetMemoryMap().DisableAux(e.GetMemIndex())

	nf.SoftSwitch_BSRBANK2 = false
	mr.ss = nf

	mr.ReconfigureMemoryMap(e, nf)

}

func (mr *Apple2IOChip) BSRBANK2ON(mm *memory.MappedRegion, address int, value uint64) {

	e := mr.e
	of := mr.ss
	nf := of // copy

	//e.GetMemoryMap().DisableAux(e.GetMemIndex())

	nf.SoftSwitch_BSRBANK2 = true
	mr.ss = nf

	mr.ReconfigureMemoryMap(e, nf)

}

func (mr *Apple2IOChip) BSRREADRAMOFF(mm *memory.MappedRegion, address int, value uint64) {

	e := mr.e
	of := mr.ss
	nf := of // copy

	//e.GetMemoryMap().DisableAux(e.GetMemIndex())

	nf.SoftSwitch_HRAMRD = false
	mr.ss = nf

	mr.ReconfigureMemoryMap(e, nf)

}

func (mr *Apple2IOChip) BSRREADRAMON(mm *memory.MappedRegion, address int, value uint64) {

	e := mr.e
	of := mr.ss
	nf := of // copy

	//e.GetMemoryMap().DisableAux(e.GetMemIndex())

	nf.SoftSwitch_HRAMRD = true
	mr.ss = nf

	mr.ReconfigureMemoryMap(e, nf)

}

func (mr *Apple2IOChip) READBSR2(mm *memory.MappedRegion, address int) uint64 {

	// removed mutex code

	e := mr.e
	of := mr.ss
	nf := of // copy

	nf.BSR2 = apple2helpers.READBSR
	nf.BSR1 = apple2helpers.OFFBSR

	nf.SoftSwitch_BSRBANK2 = true
	nf.SoftSwitch_HRAMRD = true

	mr.ReconfigureMemoryMap(e, nf)

	mr.ss = nf

	return 0

}

func (mr *Apple2IOChip) READBSR1(mm *memory.MappedRegion, address int) uint64 {

	// removed mutex code

	e := mr.e
	of := mr.ss
	nf := of // copy

	nf.BSR1 = apple2helpers.READBSR
	nf.BSR2 = apple2helpers.OFFBSR

	nf.SoftSwitch_BSRBANK2 = false
	nf.SoftSwitch_HRAMRD = true

	mr.ReconfigureMemoryMap(e, nf)

	mr.ss = nf

	return 0

}

func (mr *Apple2IOChip) WRITEBSR2(mm *memory.MappedRegion, address int) uint64 {

	// removed mutex code

	e := mr.e
	of := mr.ss
	nf := of // copy

	nf.BSR2 = apple2helpers.WRITEBSR
	nf.BSR1 = apple2helpers.OFFBSR

	nf.SoftSwitch_BSRBANK2 = true
	nf.SoftSwitch_HRAMRD = false

	mr.ReconfigureMemoryMap(e, nf)

	mr.ss = nf

	return 0

}

func (mr *Apple2IOChip) WRITEBSR1(mm *memory.MappedRegion, address int) uint64 {

	// removed mutex code

	e := mr.e
	of := mr.ss
	nf := of // copy

	nf.BSR1 = apple2helpers.WRITEBSR
	nf.BSR2 = apple2helpers.OFFBSR

	nf.SoftSwitch_BSRBANK2 = false
	nf.SoftSwitch_HRAMRD = false

	mr.ReconfigureMemoryMap(e, nf)

	mr.ss = nf

	return 0

}

func (mr *Apple2IOChip) RDWRBSR2(mm *memory.MappedRegion, address int) uint64 {

	// removed mutex code

	e := mr.e
	of := mr.ss
	nf := of // copy

	nf.BSR2 = apple2helpers.RDWRBSR
	nf.BSR1 = apple2helpers.OFFBSR

	nf.SoftSwitch_BSRBANK2 = true
	nf.SoftSwitch_HRAMRD = false

	mr.ReconfigureMemoryMap(e, nf)

	mr.ss = nf

	return 0

}

func (mr *Apple2IOChip) RDWRBSR1(mm *memory.MappedRegion, address int) uint64 {

	// removed mutex code

	e := mr.e
	of := mr.ss
	nf := of // copy

	nf.BSR1 = apple2helpers.RDWRBSR
	nf.BSR2 = apple2helpers.OFFBSR

	nf.SoftSwitch_BSRBANK2 = false
	nf.SoftSwitch_HRAMRD = false

	mr.ReconfigureMemoryMap(e, nf)

	mr.ss = nf

	return 0

}

func (mr *Apple2IOChip) OFFBSR2(mm *memory.MappedRegion, address int) uint64 {

	// removed mutex code

	e := mr.e
	of := mr.ss
	nf := of // copy

	nf.BSR2 = apple2helpers.OFFBSR
	nf.BSR1 = apple2helpers.OFFBSR

	nf.SoftSwitch_BSRBANK2 = true
	nf.SoftSwitch_HRAMRD = false

	mr.ReconfigureMemoryMap(e, nf)

	mr.ss = nf

	return 0

}

func (mr *Apple2IOChip) OFFBSR1(mm *memory.MappedRegion, address int) uint64 {

	// removed mutex code

	e := mr.e
	of := mr.ss
	nf := of // copy

	nf.BSR1 = apple2helpers.OFFBSR
	nf.BSR2 = apple2helpers.OFFBSR

	nf.SoftSwitch_BSRBANK2 = false
	nf.SoftSwitch_HRAMRD = false

	mr.ReconfigureMemoryMap(e, nf)

	mr.ss = nf

	return 0

}
