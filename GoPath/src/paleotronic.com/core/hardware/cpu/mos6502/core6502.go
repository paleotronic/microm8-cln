package mos6502

import (
	"encoding/hex"
	"encoding/json"
	fmt2 "fmt"
	"io/ioutil"
	"os"
	"paleotronic.com/core/types"
	"strings"
	"sync"
	"time"

	"paleotronic.com/core/hardware/cpu"
	"paleotronic.com/core/hardware/servicebus"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/settings"
	"paleotronic.com/debugger/debugtypes"
	"paleotronic.com/fmt"
	"paleotronic.com/octalyzer/bus"
)

const (
	//SPEED                 = 1020484
	DEBUG        = false
	REG_A        = 0
	REG_X        = 1
	REG_Y        = 2
	MAXAUDIO     = SAMPLERATE * 3
	F_N      int = 128
	F_V      int = 64
	F_R      int = 32
	F_B      int = 16
	F_D      int = 8
	F_I      int = 4
	F_Z      int = 2
	F_C      int = 1
	//CLOCKS_PER_SAMPLE     = SPEED / SAMPLERATE
	CYCLES_PER_SYNC int = 256
)

var TRACE bool = false
var TRACEMEM bool = false
var PROFILE bool = false
var HEATMAP bool = false
var STOP65C02 bool = false

var count int64
var tickinterval int64 = 5
var msdiv int64

// var chunktime int64
var ticker *time.Ticker

var ff *os.File
var pp *os.File

const SCHEDMS = 3

func init() {
	ticker = time.NewTicker(SCHEDMS * time.Millisecond)
	tickinterval = SCHEDMS
	msdiv = 1000 / tickinterval
	go func() {

		for {
			select {
			case _ = <-ticker.C:
				count++
			}
		}

	}()

}

type CPURunState int

const (
	CrsFreeRun CPURunState = iota // default run state
	CrsPaused
	CrsSingleStep
	CrsHalted
	CrsFreeRunTrace
	CrsStepOver
	CrsStepOut
)

type ROMHandler interface {
	DoCall(addr int, caller interfaces.Interpretable, passtocpu bool, zeropage bool) bool
}

type WaveStreamer interface {
	PassWaveBuffer(channel int, data []float32, loop bool, rate int)
	PassWaveBufferNB(channel int, data []float32, loop bool, rate int)
	PassWaveBufferCompressed(channel int, data []uint64, loop bool, rate int, directSent bool)
}

type MemoryAccess interface {
	SetMemory(address int, value uint64)
	GetMemory(address int) uint64
}

type AddressFlags uint64

const (
	AF_READ_BREAK  AddressFlags = 1 << iota
	AF_WRITE_BREAK AddressFlags = 1 << iota
	AF_EXEC_BREAK  AddressFlags = 1 << iota
	AF_WRITE_LOCK  AddressFlags = 1 << iota
	AF_WRITE_BUMP  AddressFlags = 1 << iota
	AF_READ_BUMP   AddressFlags = 1 << iota
	AF_EXEC_BUMP   AddressFlags = 1 << iota
)

type Registers struct {
	Y            int
	PC           int
	SP           int
	X            int
	P            int
	GlobalCycles int64
	A            int
	RealSpeed    int
}

type Core6502 struct {
	Registers
	RecRegisters     Registers
	RequestInterrupt bool

	RunState     CPURunState
	StepOverSP   int
	StepOverAddr int

	EmuStartTime             int64
	PlayingUntil             int64
	BaseSpeed                int64
	Warp, PrevWarp, UserWarp float64
	TRIMSAMPLES              int
	//CycleCount, CycleInterval int64
	StepOutCycles int64
	SampleCycles  int64
	SAMPLEVALUE   float32
	InitialSP     int
	Opref         [256]*Op6502
	BasicMode     bool

	Halted          bool
	HotStart        bool
	Int             interfaces.Interpretable
	DEBUG           bool
	OperCycles      int
	AllowOperCycles bool
	// SAMPLECOUNT          int
	// SAMPLEDATA           []uint64
	// SAMPLENUM            uint64
	// SAMPLEFLIP           bool
	LastCycleDeficit     int64
	Handler              Event6502
	VDU                  WaveStreamer
	ROM                  func(addr int, caller interfaces.Interpretable, passtocpu bool) bool
	tracenum, profilenum int
	IgnoreBRK            bool
	IgnoreILL            bool
	SpeakerSample        float32
	MemoryTrip           bool
	MemoryTripAddress    int
	HeatMap              map[string]uint64
	RAM                  *memory.MemoryMap
	UseProDOS            bool
	//	SyncCycleCounter          int
	IgnoreStackFallouts bool
	CPS                 float64
	profiledatacum      map[string]int64
	profiledatacount    map[string]int64
	shim                map[int]Func6502
	lastReadAddr        int
	lastReadValue       int
	PauseNextRTS        bool

	Counters []interfaces.Countable

	SpecialFlag     map[int]AddressFlags
	HasSpecialFlags bool
	LockValue       map[int]uint64

	MemIndex            int
	InitTime            time.Time
	Model               string
	InitFunc            func(cpu *Core6502)
	DoneFunc            func(cpu *Core6502)
	ResetFunc           func(cpu *Core6502)
	ResetLine           bool
	WaitingForSync      bool
	chunktime           int64
	UserWarpMode        bool
	WaitCounter         time.Duration
	basecount           int64
	startcycles         int64
	ff                  *os.File
	events              []*servicebus.ServiceBusRequest
	ResetRequested      bool
	ResetRequestedAddr  int
	mmu                 *memory.MemoryManagementUnit
	CycleCounters       []interfaces.Countable
	IO                  interfaces.Countable
	Suspended           bool
	SuspendAcknowledged bool
	sync.Mutex
}

func onecomp(a int) int {
	return 255 - (a & 0xff)
}

func ab(cond bool, a, b int) int {
	if cond {
		return a
	}
	return b
}

func (this *Core6502) PullResetLine() {
	this.ResetLine = true
	this.RequestResume()
}

func (this *Core6502) RequestSuspend() bool {
	if this.Suspended || this.Int.GetState() != types.EXEC6502 || this.RAM.IntGetSlotRestart(this.MemIndex) {
		return true
	}
	this.SuspendAcknowledged = false
	this.Suspended = true
	for this.Suspended && !this.SuspendAcknowledged && !this.RAM.IntGetSlotRestart(this.MemIndex) {
		time.Sleep(5 * time.Millisecond)
	}
	return true
}

func (this *Core6502) RequestResume() {
	this.Suspended = false
	this.SuspendAcknowledged = true
	this.Halted = false
	this.ResetSliced()
}

func (this *Core6502) RequestReset(addr int) {
	this.ResetRequested = true
	this.ResetRequestedAddr = addr
}

func (this *Core6502) testMemFlags(f AddressFlags, mask AddressFlags) bool {
	return f&mask == mask
}

func (this *Core6502) GetMemFlags(addr int) AddressFlags {
	return this.SpecialFlag[addr]
}

func (this *Core6502) AreMemFlagsSet(addr int, f AddressFlags) bool {
	return this.SpecialFlag[addr]&f == f
}

func (this *Core6502) SetMemFlags(addr int, f AddressFlags, on bool) {

	if on {
		this.SpecialFlag[addr] |= f
	} else {
		this.SpecialFlag[addr] &= (0xffff ^ f)
	}

}

func (this *Core6502) Init() {
	this.RequestResume()
	if this.InitFunc != nil {
		this.InitFunc(this)
	}
	// clean mem locks
	this.LockValue = make(map[int]uint64)
}

func (this *Core6502) Done() {
	if this.DoneFunc != nil {
		this.DoneFunc(this)
	} else {
		//fmt.Println("no done handler")
	}
}

func (this *Core6502) GetModel() string {
	return this.Model
}

func (this *Core6502) PendingIRQ() bool {
	return this.RequestInterrupt
}

func (this *Core6502) PullIRQLine() {
	this.RequestInterrupt = true
	this.SetFlag(F_B, false)
}

func (this *Core6502) CheckIRQLine() {
	if !this.RequestInterrupt {
		return
	}

	if !this.TestFlag(F_I) || this.TestFlag(F_B) {
		// we need to clear this now so a secondary chip can trigger the IRQ and it will be serviced immediately afterwards
		// ref: http://6502.org/tutorials/interrupts.html
		this.RequestInterrupt = false

		retaddr := this.PC
		// return - lo, hi
		this.ClockTick() // FIX: Add missing two cycles of interrupt time to clock - IRQ trigger should consume 7 cycles
		this.ClockTick()
		this.Push((retaddr >> 8) & 0xff)
		this.Push(retaddr & 0xff)
		// status
		this.Push(this.P)
		// get vector address
		newaddr := this.FetchByteAddr(0xfffe) | (256 * this.FetchByteAddr(0xffff))
		this.PC = newaddr
		// set i flag
		this.SetFlag(F_I, true)
		if this.Model == "65C02" {
			this.SetFlag(F_D, false) // 65C02 clears D flag when setting I flag
		}

		if TRACE {
			this.ff.WriteString(fmt.Sprintf("IRQ -> $%.4x (I=%v, B=%v)\n", newaddr, this.TestFlag(F_I), this.TestFlag(F_B)))
		}

		if settings.Debug6522 {
			fmt2.Printf("6522TRACE: IRQ triggered\n")
		}
	}

}

func (this *Core6502) CheckWarp() bool {

	if this.Warp != this.PrevWarp {

		//fmt.Printf("Warp changed from %f%% to %f%%\n", this.PrevWarp*100, this.Warp*100)

		// Defer user speed changes when blocked
		if this.UserWarpMode && !settings.CanUserChangeSpeed() {
			return false
		}

		this.PrevWarp = this.Warp
		this.CalcTiming()
		return true
	}
	return false
}

func (this *Core6502) ResetSliced() {
	this.basecount = count
	this.startcycles = this.GlobalCycles
}

func (this *Core6502) IsTracing() bool {
	return this.ff != nil
}

func (this *Core6502) StopTrace() {
	if this.ff != nil {
		this.ff.Close()
		TRACE = false
	}
}

func (this *Core6502) StartTrace(filename string) error {
	this.StopTrace()

	ff, err := os.Create(filename)
	if err == nil {
		TRACE = true
		this.ff = ff
	}
	return err
}

func (this *Core6502) checkAndApplyMemLocks() {
	if m := settings.MemLocks[this.MemIndex]; m != nil && len(m) > 0 && len(this.LockValue) == 0 {
		for a, v := range m {
			this.LockValue[a] = v
		}
	}
}

func (this *Core6502) ExecuteSliced() cpu.FEResponse {

	this.Counters = this.Int.GetCycleCounter()
	this.CycleCounters = this.Counters
	this.checkAndApplyMemLocks()

	// var currpc int
	var r cpu.FEResponse

	entryTime := time.Now()

	var loopcount int

	if this.Halted {
		this.SuspendAcknowledged = true
	}

	for !this.Halted && r == cpu.FE_OK {

		if settings.CleanBootRequested {
			this.Halted = true
			this.SuspendAcknowledged = true
			return cpu.FE_HALTED
		}

		if this.Suspended {
			this.SuspendAcknowledged = true
			this.ResetSliced()
			return cpu.FE_SLEEP
		}

		//now = time.Now()
		os := this.Int.GetState()
		r = this.FetchExecute()
		if this.Int.GetState() != os {
			return cpu.FE_OK
		}
		if this.RunState == CrsPaused {
			return cpu.FE_SLEEP
		}

		// We only sleep here if we are running 'realtime' otherwise we just belt through
		tickdiff := count - this.basecount // # 5ms clock ticks since we started
		if tickdiff > 0 && (this.GlobalCycles-this.startcycles)/tickdiff > this.chunktime {
			return cpu.FE_SLEEP
		}

		// Check if warp changed
		if this.CheckWarp() {
			this.ResetSliced()
			if this.Suspended {
				this.SuspendAcknowledged = true
			}
			return cpu.FE_SLEEP
		}

		// if we've been here too long, perhaps nitro CPU mode and we need to service the loops
		if time.Since(entryTime) > 500*time.Millisecond {
			this.ResetSliced()
			if this.Suspended {
				this.SuspendAcknowledged = true
			}
			return cpu.FE_SLEEP
		}

		loopcount++
		if loopcount > 100 {
			loopcount = 0
			this.Int.PBPaste()
			if this.CheckWarp() {
				this.ResetSliced()
				if this.Suspended {
					this.SuspendAcknowledged = true
				}
				return cpu.FE_SLEEP
			}
		}

	}

	return r
}

func (this *Core6502) ExecNCycles(n int64) cpu.FEResponse {

	this.CheckWarp()

	cyclesToNS := 1000000000 / this.ActualSpeed()

	if this.Int.IsPaused() {
		return cpu.FE_OK
	}

	currcycles := this.GlobalCycles
	targetcount := currcycles + n

	var r cpu.FEResponse = cpu.FE_OK
	var c, a int64
	var d time.Duration
	for r == cpu.FE_OK && !this.Halted && this.GlobalCycles < targetcount {
		c = this.GlobalCycles
		r = this.FetchExecute()
		a = this.GlobalCycles - c
		d = time.Duration(a*cyclesToNS) * time.Nanosecond
		this.WaitCounter += d
		if this.WaitCounter > 1050*time.Microsecond {
			time.Sleep(time.Millisecond)
			this.WaitCounter -= 1050 * time.Microsecond
		} else if this.GetWarp() > 1 {
			if this.GlobalCycles%int64(this.GetWarp()*10000) == 0 {
				return r
			}
		}
	}

	return r
}

func (this *Core6502) HandleReset() {
	// readdr := this.PC

	// this.Push(readdr / 256)
	// this.Push(readdr % 256)

	// p := this.P | F_B

	// this.Push(p)

	//target := this.FetchByteAddr(0xfffe) + 256*this.FetchByteAddr(0xffff)
	target := this.FetchByteAddr(0xfffc) + 256*this.FetchByteAddr(0xfffd)

	this.PC = target
}

func (this *Core6502) Break() {
	readdr := this.PC

	this.Push(readdr / 256)
	this.Push(readdr % 256)

	p := this.P | F_B

	this.Push(p)

	//target := this.FetchByteAddr(0xfffe) + 256*this.FetchByteAddr(0xffff)
	target := this.FetchByteAddr(0xfffc) + 256*this.FetchByteAddr(0xfffd)

	this.PC = target
}

func (this *Core6502) SetWarp(w float64) {
	this.Warp = w
	//this.CalcTiming()
}

func (this *Core6502) HasUserWarp() (bool, float64) {
	return this.UserWarpMode, this.UserWarp
}

func (this *Core6502) SetWarpUser(w float64) {

	settings.MuteCPU = (w >= 8)

	this.SetWarp(w)
	this.UserWarp = w
	//this.PrevWarp = w
	this.UserWarpMode = (w != 1)
	//this.CalcTiming()
}

func (this *Core6502) GetWarp() float64 {
	return this.Warp
}

func (this *Core6502) ActualSpeed() int64 {
	return int64(float64(this.BaseSpeed) * this.Warp)
}

func (this *Core6502) ExecTillHalted() cpu.FEResponse {

	this.Counters = this.Int.GetCycleCounter()
	this.CycleCounters = this.Counters

	this.SetWarp(1)
	this.CheckWarp()

	vm := this.Int.VM()

	//fmt.Printf("Basic mode = %v\n", this.BasicMode)

	this.WaitingForSync = false

	this.Init()

	//go this.AudioFunnel()
	this.RAM = this.Int.GetMemoryMap()
	this.GlobalCycles = 0

	if !settings.PureBoot(this.Int.GetMemIndex()) {
		this.SP = this.InitialSP

		this.SetFlag(F_R, false)
		this.SetFlag(F_B, false)
		this.SetFlag(F_I, true)
	}

	if settings.PureBoot(this.Int.GetMemIndex()) || settings.RealTimeBasicMode {
		this.BasicMode = false
	}

	if !this.BasicMode {
		bus.StopClock()
	}

	defer func() {
		if HEATMAP {
			////fmt.Println(this.HeatMap)
			ndata, e := json.Marshal(this.HeatMap)
			////fmt.Println(string(ndata))
			if e == nil {
				e = ioutil.WriteFile("romheatmap.json", ndata, 0755)
			}
		}
	}()

	if TRACE || TRACEMEM {
		ff, _ = os.Create(fmt.Sprintf("trace_6502_%d.txt", this.tracenum))
		this.tracenum++
		defer ff.Close()
	}

	if PROFILE {
		pp, _ = os.Create(fmt.Sprintf("profile_6502_%d.csv", this.profilenum))
		this.profilenum++
		defer pp.Close()
	}

	//fmt.RPrintf("<***> START CPU execution at 0x%.4x\n", this.PC)

	//if !this.BasicMode {
	//this.CycleInterval = 0 // cycles between speaker clicks
	//this.CycleCount = 0    // total cycles
	//this.SyncCycleCounter = 0
	this.GlobalCycles = 0
	// this.SAMPLECOUNT = 0
	// this.SAMPLEDATA = make([]uint64, MAXAUDIO)
	// this.SAMPLENUM = 31
	//}

	this.Halted = false
	this.EmuStartTime = time.Now().UnixNano() // time allotted to start
	var r cpu.FEResponse = cpu.FE_OK
	//var q int64
	var currpc int
	//~ var now time.Time
	//~ var durns int64

	basecount := count

	startcycles := this.GlobalCycles
	//s := time.Now()

	for !this.Halted && r == cpu.FE_OK {

		if this.GlobalCycles%1000 == 0 {
			vm.ExecutePendingTasks()
		}

		if this.Int.VM().IsDying() {
			this.Halted = true
			return cpu.FE_HALTED
		}

		if this.RAM.IntGetSlotHalt(this.MemIndex) {
			this.RAM.IntSetSlotHalt(this.MemIndex, false)
			this.Halted = true
			return cpu.FE_HALTED
		}

		if this.RAM.IntGetCPUBreak(this.MemIndex) {
			this.RAM.IntSetCPUBreak(this.MemIndex, false)
			this.Break()
		}

		if settings.CleanBootRequested {
			this.Halted = true
			return cpu.FE_HALTED
		}

		oc := count
		if this.Int.WaitForWorld() {
			diff := count - oc
			basecount += diff
			if this.RAM.IntGetSlotRestart(this.MemIndex) {
				return cpu.FE_HALTED
			}
		}

		if this.Halted {
			//fmt.RPrintln("CPU is halted")
			return cpu.FE_HALTED
		}

		currpc = this.PC

		if TRACE {
			txt, _ := this.DecodeTrace(currpc)
			ff.WriteString(txt + "\r\n")
		}

		//now = time.Now()
		r = this.FetchExecute()

		// We only sleep here if we are running 'realtime' otherwise we just belt through
		tickdiff := count - basecount // # 5ms clock ticks since we started
		if !this.BasicMode && tickdiff > 0 && (this.GlobalCycles-startcycles)/tickdiff > this.chunktime {
			time.Sleep(time.Millisecond)
			//fmt.RPrintf("z")
		}

		if this.Halted {
			this.DoneFunc(this)
			//worktime := int64(time.Since(s) / time.Microsecond)
			basiccycles := this.GlobalCycles - startcycles
			if this.BasicMode {
				this.SetWait(basiccycles)
			}
			return cpu.FE_OK
		}

		// Check if warp changed
		if !this.BasicMode && this.CheckWarp() {
			basecount = count
			startcycles = this.GlobalCycles
		}

	}

	basiccycles := this.GlobalCycles - startcycles
	//	if !this.BasicMode {
	// send samples out

	if this.BasicMode {
		this.SetWait(basiccycles)
	}
	//for this.PlayingUntil > time.Now().UnixNano() {
	//time.Sleep(1*time.Millisecond)
	//}

	//speaker.SetVolume(0);
	//	}

	//    this.Int.GetMemoryMap().PaddleMap[this.Int.GetMemIndex()][0] = 0
	//    this.Int.GetMemoryMap().PaddleMap[this.Int.GetMemIndex()][1] = 1

	if PROFILE {
		pp.WriteString("INSTRUCTION,COUNT,CUMULATIVE_NS,AVERAGE_NS\r\n")
		for k, _ := range this.profiledatacum {
			count := this.profiledatacount[k]
			ns := this.profiledatacum[k]
			avg := float64(ns) / float64(count)
			s := fmt.Sprintf("\"%s\",%d,%d,%f\r\n", k, count, ns, avg)
			pp.WriteString(s)
		}
		pp.Close()
	}

	//fmt.Println("ETH - end")
	this.Done()

	if !this.BasicMode {
		bus.StartDefault()
	}

	if settings.LaunchQuitCPUExit {
		os.Exit(0)
	}

	return r
}

func (this *Core6502) INT8(v int) int {
	return (int)((byte)(v & 0xff))
}

func (this *Core6502) CalcTiming() {
	this.chunktime = int64(int64(float64(this.BaseSpeed)*this.Warp) / msdiv)
	this.RealSpeed = int(float64(this.BaseSpeed) * this.Warp)
	// for _, k := range this.Counters {
	// 	k.AdjustClock(this.RealSpeed)
	// }
	if this.IO != nil {
		this.IO.AdjustClock(int(this.RealSpeed))
	}
	// servicebus.SendServiceBusMessage(
	// 	this.MemIndex,
	// 	servicebus.Clocks6502Update,
	// 	int(this.RealSpeed),
	// )
}

func (this *Core6502) SetFlag(val int, condition bool) {

	//    fn := "unknown"
	//	switch val {
	//    case F_Z: fn = "ZERO"
	//    case F_C: fn = "CARRY"
	//    case F_I: fn = "IRQ"
	//    case F_D: fn = "DECIMAL"
	//    case F_B: fn = "BREAK"
	//    case F_V: fn = "OVERFLOW"
	//    case F_N: fn = "NEGATIVE"
	//    }

	if condition {
		this.P |= (val & 0xff)
		//        //fmt.Printf("SET flag %s\n", fn)
	} else {
		this.P &= (255 - val)
		//        //fmt.Printf("CLEAR flag %s\n", fn)

	}
}

func (this *Core6502) Nmi() {
	this.PC = this.FetchWordAddr(0xfffa)
}

func (this *Core6502) Page_changing(addr int, offset int) bool {
	vaddr := addr + offset
	return ((addr / 256) != (vaddr))
}

func (this *Core6502) Flag(f int) int {
	if (this.P & f) != 0 {
		return 1
	} else {
		return 0
	}
}

func (this *Core6502) Test(f int) bool {
	return ((this.P & f) != 0)
}

func (this *Core6502) Decode(addr int) (string, int) {
	opcode := this.FetchByteAddr(addr)
	info := this.GetOpDesc(opcode)

	if info == nil {
		return fmt.Sprintf("%.4X- %-10s %s\r\n", addr, fmt.Sprintf("%.2X", opcode), "???"), 1
	}

	extra := info.GetBytes() - 1

	value := 0
	opvalue := 0
	switch extra {
	case 1:
		value = this.FetchByteAddr(addr + 1)
		opvalue = value
		if strings.HasPrefix(info.Description, "B") {
			opvalue = addr + 2 + int(int8(value&0xff))
		}

		break
	case 2:
		value = this.FetchWordAddr(addr + 1)
		break
	}

	desc := info.Description
	//desc = strings.Replace(desc, "oper", this.Pad(hex.EncodeToString([]byte{byte(value / 256), byte(value % 256)}), info.GetBytes()-1), -1)

	switch info.GetBytes() {
	case 1:
		return fmt.Sprintf(
			"%.4X- %-10s %s\r\n",
			addr,
			fmt.Sprintf("%.2X", opcode),
			desc), info.GetBytes()
	case 2:
		desc = strings.Replace(
			desc,
			"oper",
			fmt.Sprintf("$%.2X", opvalue),
			-1,
		)
		return fmt.Sprintf(
			"%.4X- %-10s %s\r\n",
			addr,
			fmt.Sprintf("%.2X %.2X", opcode, value%256),
			desc), info.GetBytes()
	default:
		desc = strings.Replace(
			desc,
			"oper",
			fmt.Sprintf("$%.4X", value),
			-1,
		)
		return fmt.Sprintf(
			"%.4X- %-10s %s\r\n",
			addr,
			fmt.Sprintf("%.2X %.2X %.2X", opcode, value%256, value/256),
			desc), info.GetBytes()
	}

}

func (this *Core6502) DecodeInstruction(addr int) ([]int, string, int) {
	// 36597: e577 BIT & 25f         [ 25f]  A: e0 X:  0 Y:  a SP: f1 P: 00110001 ROM: 15 +4
	// "%.d: %.4x %-18s [%4x]  A: %2x X: %2x Y: %2x SP: %2x P: %.8b"
	opcode := this.FetchByteAddr(addr)
	info := this.GetOpDesc(opcode)
	if info == nil {
		return []int{opcode}, "???", 2
	}
	extra := info.GetBytes() - 1
	value := 0
	opvalue := 0
	switch extra {
	case 1:
		value = this.FetchByteAddr(addr + 1)
		opvalue = value
		if strings.HasPrefix(info.Description, "B") {
			opvalue = addr + 2 + int(int8(value&0xff))
		}
		break
	case 2:
		value = this.FetchWordAddr(addr + 1)
		break
	}
	desc := info.Description
	addr++
	_, _, penalty := info.Fetch(this, &addr)
	//desc = strings.Replace(desc, "oper", this.Pad(hex.EncodeToString([]byte{byte(value / 256), byte(value % 256)}), info.GetBytes()-1), -1)
	switch info.GetBytes() {
	case 1:
		return []int{opcode}, desc, info.GetCycles() + penalty
	case 2:
		desc = strings.Replace(
			desc,
			"oper",
			fmt.Sprintf("$%.2X", opvalue),
			-1,
		)
		return []int{opcode, value}, desc, info.GetCycles() + penalty
	default:
		desc = strings.Replace(
			desc,
			"oper",
			fmt.Sprintf("$%.4X", value),
			-1,
		)
		return []int{opcode, value % 256, value / 256}, desc, info.GetCycles() + penalty
	}
}

func (this *Core6502) DecodeTraceAlt(addr int) (string, int) {
	// 36597: e577 BIT & 25f         [ 25f]  A: e0 X:  0 Y:  a SP: f1 P: 00110001 ROM: 15 +4
	// "%.d: %.4x %-18s [%4x]  A: %2x X: %2x Y: %2x SP: %2x P: %.8b"

	// Save and restore PC on return

	f := "%.d: %.4x %-18s [%4x]  A: %2x X: %2x Y: %2x SP: %2x P: %.8b +%d"
	opcode := this.FetchByteAddr(addr)
	info := this.GetOpDesc(opcode)
	if info == nil {
		return fmt.Sprintf(
			f,
			this.GlobalCycles,
			addr,
			"???",
			0,
			this.A,
			this.X,
			this.Y,
			this.SP&0xff,
			this.P,
			2,
		), 1
		//return fmt.Sprintf( "%.4x- %-10s %s\r\n", addr, fmt.Sprintf("%.2x", opcode), "???" ), 1
	}
	extra := info.GetBytes() - 1
	value := 0
	opvalue := 0
	switch extra {
	case 1:
		value = this.FetchByteAddr(addr + 1)
		opvalue = value
		if strings.HasPrefix(info.Description, "B") {
			opvalue = addr + 2 + int(int8(value&0xff))
		}
		break
	case 2:
		value = this.FetchWordAddr(addr + 1)
		break
	}
	desc := info.Description
	addr++
	_, actual, penalty := info.Fetch(this, &addr)
	//desc = strings.Replace(desc, "oper", this.Pad(hex.EncodeToString([]byte{byte(value / 256), byte(value % 256)}), info.GetBytes()-1), -1)
	switch info.GetBytes() {
	case 1:
		return fmt.Sprintf(
			f,
			this.GlobalCycles,
			addr,
			desc,
			actual,
			this.A,
			this.X,
			this.Y,
			this.SP&0xff,
			this.P,
			info.GetCycles()+penalty,
		), info.GetBytes()
	case 2:
		desc = strings.Replace(
			desc,
			"oper",
			fmt.Sprintf("$%.2X", opvalue),
			-1,
		)
		return fmt.Sprintf(
			f,
			this.GlobalCycles,
			addr,
			desc,
			actual,
			this.A,
			this.X,
			this.Y,
			this.SP&0xff,
			this.P,
			info.GetCycles()+penalty,
		), info.GetBytes()
	default:
		desc = strings.Replace(
			desc,
			"oper",
			fmt.Sprintf("$%.4X", value),
			-1,
		)
		return fmt.Sprintf(
			f,
			this.GlobalCycles,
			addr,
			desc,
			actual,
			this.A,
			this.X,
			this.Y,
			this.SP&0xff,
			this.P,
			info.GetCycles()+penalty,
		), info.GetBytes()
	}
}

func (this *Core6502) DecodeTrace(addr int) (string, int) {

	//   A: X: Y: SP:  Flags     Addr:Opcode    Mnemonic
	f := "%.2X %.2X %.2X %.4X %8s %.4X:%-8s  %s"
	opcode := this.FetchByteAddr(addr)
	info := this.GetOpDesc(opcode)
	if info == nil {
		return fmt.Sprintf(
			f,
			this.A,
			this.X,
			this.Y,
			this.SP,
			this.FlagString(),
			addr,
			//fmt.Sprintf("%.2X", opcode),
			"???",
		), 1
		//return fmt.Sprintf( "%.4x- %-10s %s\r\n", addr, fmt.Sprintf("%.2x", opcode), "???" ), 1
	}
	extra := info.GetBytes() - 1
	value := 0
	opvalue := 0
	switch extra {
	case 1:
		value = this.FetchByteAddr(addr + 1)
		opvalue = value
		if strings.HasPrefix(info.Description, "B") {
			opvalue = addr + 2 + int(int8(value&0xff))
		}
		break
	case 2:
		value = this.FetchWordAddr(addr + 1)
		break
	}
	desc := info.Description
	//desc = strings.Replace(desc, "oper", this.Pad(hex.EncodeToString([]byte{byte(value / 256), byte(value % 256)}), info.GetBytes()-1), -1)
	switch info.GetBytes() {
	case 1:
		return fmt.Sprintf(
			f,
			this.A,
			this.X,
			this.Y,
			this.SP,
			this.FlagString(),
			addr,
			fmt.Sprintf("%.2X", opcode),
			desc,
		), info.GetBytes()
	case 2:
		desc = strings.Replace(
			desc,
			"oper",
			fmt.Sprintf("$%.2X", opvalue),
			-1,
		)
		return fmt.Sprintf(
			f,
			this.A,
			this.X,
			this.Y,
			this.SP,
			this.FlagString(),
			addr,
			fmt.Sprintf("%.2X %.2X", opcode, value%256),
			desc,
		), info.GetBytes()
	default:
		desc = strings.Replace(
			desc,
			"oper",
			fmt.Sprintf("$%.4X", value),
			-1,
		)
		return fmt.Sprintf(
			f,
			this.A,
			this.X,
			this.Y,
			this.SP,
			this.FlagString(),
			addr,
			fmt.Sprintf("%.2X %.2X %.2X", opcode, value%256, value/256),
			desc,
		), info.GetBytes()
	}
}

func (this *Core6502) DecodeProfileCSV(addr int, costns int64) (string, int) {

	//   A: X: Y: SP:  Flags     Addr:Opcode    Mnemonic
	f := "%s"
	opcode := this.FetchByteAddr(addr)
	info := this.GetOpDesc(opcode)

	if info == nil {

		return fmt.Sprintf(
			f,
			"???",
		), 1

		//return fmt.Sprintf( "%.4x- %-10s %s\r\n", addr, fmt.Sprintf("%.2x", opcode), "???" ), 1
	}

	extra := info.GetBytes() - 1

	value := 0
	opvalue := 0
	switch extra {
	case 1:
		value = this.FetchByteAddr(addr + 1)
		opvalue = value
		if strings.HasPrefix(info.Description, "B") {
			opvalue = addr + 2 + int(int8(value&0xff))
		}

		break
	case 2:
		value = this.FetchWordAddr(addr + 1)
		break
	}

	desc := info.Description
	//desc = strings.Replace(desc, "oper", this.Pad(hex.EncodeToString([]byte{byte(value / 256), byte(value % 256)}), info.GetBytes()-1), -1)

	switch info.GetBytes() {
	case 1:
		return fmt.Sprintf(
			f,
			desc,
		), info.GetBytes()
	case 2:
		desc = strings.Replace(
			desc,
			"oper",
			fmt.Sprintf("$%.2X", opvalue),
			-1,
		)
		return fmt.Sprintf(
			f,
			desc,
		), info.GetBytes()
	default:
		desc = strings.Replace(
			desc,
			"oper",
			fmt.Sprintf("$%.4X", value),
			-1,
		)
		return fmt.Sprintf(
			f,
			desc,
		), info.GetBytes()
	}

}

func (this *Core6502) TriggerVideo() {
	if !settings.IsRemInt {
		if this.RAM.IntGetLayerState(this.MemIndex) != 0 {
			bus.Sync()
		}
	}
}

func (this *Core6502) LogToTrace(msg string) {
	if !TRACE {
		return
	}
	if this.ff == nil {
		return
	}
	ff.WriteString(msg + "\r\n")
}

func (this *Core6502) FetchExecute() cpu.FEResponse {

	this.OperCycles = 0
	this.AllowOperCycles = true
	defer func() {
		this.AllowOperCycles = false
	}()

	this.lastReadAddr = -1

	if this.ResetLine {
		if this.ResetFunc != nil {
			this.ResetFunc(this)
		}
		this.ResetLine = false
		this.HandleReset()
	}

	if this.ResetRequested {
		//this.Reset()
		this.ResetRequested = false
		this.GlobalCycles = 0
		if this.ResetRequestedAddr >= 0 {
			this.PC = this.ResetRequestedAddr
		}
	}

	this.CheckIRQLine() // check if IRQ line has been triggered

	// Handle some special cases
	if settings.DebuggerOn && settings.DebuggerAttachSlot == this.MemIndex {
		this.ServiceBusProcessPending()
	}

	if this.RunState == CrsPaused {
		return cpu.FE_SLEEP
	}

	this.MemoryTrip = false
	this.MemoryTripAddress = this.PC

	// flags := this.GetMemFlags(this.PC)

	// if this.testMemFlags(flags, AF_EXEC_BREAK) {
	// 	this.MemoryTrip = true
	// 	this.Halted = true
	// 	return cpu.FE_BREAKPOINT
	// }

	// if this.testMemFlags(flags, AF_EXEC_BUMP) {
	// 	counter := int(flags>>32) & 0xff
	// 	this.RAM.IntBumpCounter(this.MemIndex, counter)
	// }

	frpc := this.PC
	opcode := this.FetchBytePC(&this.PC)

	if this.RunState == CrsStepOver && frpc == this.StepOverAddr && opcode != 0x20 {
		this.RunState = CrsSingleStep
	}

	// is a firmware shim set in the blockmapper?
	fr := this.mmu.FirmwareLastRead
	if fr != nil {
		clocks := fr.FirmwareExec(
			frpc%256,
			&this.PC,
			&this.A,
			&this.X,
			&this.Y,
			&this.SP,
			&this.P,
		)
		// for _, k := range this.CycleCounters {
		// 	k.Increment(int(clocks)) // roll the clock
		// }
		//this.SyncCycleCounter += int(clocks)

		if this.IO != nil {
			this.IO.Increment(int(clocks))
		}

		this.GlobalCycles += int64(clocks)
		//this.CycleInterval += int64(clocks)
		return cpu.FE_OK
	}

	info := this.GetOpDesc(opcode)

	skipahead := false

	if info == nil {

		if !this.IgnoreILL {
			this.Halted = true
			this.PC--
			return cpu.FE_ILLEGALOPCODE
		} else {
			return cpu.FE_OK
		}

	}

	var r cpu.FEResponse = cpu.FE_OK

	//var clocks int = info.Cycles - info.Bytes
	//var preClocks int = info.Bytes

	if skipahead {
		this.PC = this.PC + info.GetBytes() - 1
	} else {

		//if this.IO != nil {
		//	this.IO.Increment(int(preClocks))
		//}
		//
		////this.SyncCycleCounter += clocks
		//this.GlobalCycles += int64(preClocks)

		// advance clocks for fetch stage prior to execute

		info.Do(this) //- preClocks // total clocks

		if this.lastReadAddr != -1 {
			if settings.DebuggerOn && settings.DebuggerAttachSlot == this.MemIndex+1 {
				servicebus.SendServiceBusMessage(
					this.MemIndex,
					servicebus.CPUReadMem,
					&debugtypes.CPUMemoryRead{
						Address: this.lastReadAddr,
						Value:   this.lastReadValue,
					},
				)
			}
		}

		// if this.RAM.SlotMemoryViolation[this.MemIndex] {
		// 	this.MemoryTrip = true
		// 	this.RAM.SlotMemoryViolation[this.MemIndex] = false
		// }

		// if this.MemoryTrip {
		// 	this.Halted = true
		// 	r = cpu.FE_BREAKPOINT_MEM
		// 	this.PC = this.MemoryTripAddress
		// 	this.Halted = true
		// 	return r
		// }

		if settings.DebuggerOn && settings.DebuggerAttachSlot == this.MemIndex+1 {
			if this.RunState == CrsSingleStep {
				this.RunState = CrsPaused
			} else if this.RunState == CrsStepOver {
				if this.StepOverSP == this.SP {
					this.RunState = CrsPaused
				}
			} else if this.RunState == CrsFreeRun && this.PauseNextRTS {
				if opcode == 0x60 {
					// RTS
					this.RunState = CrsPaused
					this.PauseNextRTS = false
					// We have paused here
					if settings.DebuggerAttachSlot == this.MemIndex+1 {
						servicebus.SendServiceBusMessage(
							this.MemIndex,
							servicebus.CPUState,
							&debugtypes.CPUState{
								PC:    this.PC,
								A:     this.A,
								X:     this.X,
								Y:     this.Y,
								P:     this.P,
								SP:    this.SP,
								CC:    this.GlobalCycles,
								Speed: this.UserWarp,
							},
						)
					}
				}
			}
		}
	}

	//if info != nil {
	//	if this.OperCycles != info.Cycles && info.Opcode&0xf != 0 {
	//		log2.Printf("6502: unexpected cycles (op: %.2x - '%s') - expected %d clocks, oper cycles gave %d", opcode, info.Description, info.Cycles, this.OperCycles)
	//	}
	//	//this.CycleInterval += int64(clocks)
	//}

	// ServiceBus message
	if settings.DebuggerOn && settings.DebuggerAttachSlot == this.MemIndex+1 {

		rb, _ := servicebus.SendServiceBusMessage(
			this.MemIndex,
			servicebus.CPUState,
			&debugtypes.CPUState{
				PC:    this.PC,
				A:     this.A,
				X:     this.X,
				Y:     this.Y,
				P:     this.P,
				SP:    this.SP,
				CC:    this.GlobalCycles,
				Speed: this.UserWarp,
			},
		)

		if len(rb) > 0 {
			if rb[0].Type == servicebus.CPUControl {
				this.HandleServiceBusRequest(
					&servicebus.ServiceBusRequest{
						Type:    rb[0].Type,
						Payload: rb[0].Payload,
					},
				)
			}
		}

	}

	if this.RunState == CrsPaused {
		return cpu.FE_SLEEP
	}

	// TODO: move this outta here
	if !this.Halted && settings.HasCPUBreak[this.MemIndex] {
		settings.HasCPUBreak[this.MemIndex] = false
		return cpu.FE_CTRLBREAK
	}

	if r != cpu.FE_OK {
		return r
	}

	if this.Halted {
		return cpu.FE_HALTED
	}

	this.RecRegisters = this.Registers // after successful opcodes only

	return cpu.FE_OK
}

func (this *Core6502) PostJumpEvent(from, to int, context string) {
	this.Int.PostJumpEvent(from, to, context)
}

func (this *Core6502) GetOpDesc(opcode int) *Op6502 {
	return this.Opref[opcode]
}

func (this *Core6502) NotFlag(f int) int {
	if (this.P & f) != 0 {
		return 0
	} else {
		return 1
	}
}

func (this *Core6502) UINT8(v int) int {
	return (v & 0xff)
}

// func (this *Core6502) Click() {
// 	this.SAMPLEVALUE = -this.SAMPLEVALUE
// 	this.CycleCount = 0
// }

// func (this *Core6502) InjectROM(name string, base int) error {

// 	data, e := assets.Asset(name)
// 	if e != nil {
// 		return e
// 	}

// 	rawdata := make([]uint64, len(data))
// 	for i, v := range data {
// 		rawdata[i] = uint64(v)
// 	}

// 	this.RAM.BlockWrite(this.RAM.MEMBASE(this.MemIndex)+base, rawdata)

// 	return nil

// }

func (this *Core6502) Inject(i int, js []uint64) {

	for z := 0; z < len(js); z++ {
		//this.Int.SetMemory(i+z, js[z])

		this.RAM.WriteInterpreterMemory(this.MemIndex, i+z, js[z])
	}

}

func (this *Core6502) SetListener(list Event6502) {
	this.Handler = list
}

func (this *Core6502) Dec_SP() {
	this.SP = 0x100 | ((this.SP - 1) & 0xff)
}

func (this *Core6502) Push(value int) {
	//this.Int.SetMemory(this.SP, uint64(value))

	var v = uint64(value)
	this.mmu.Do(this.SP, memory.MA_WRITE, &v)
	this.ClockTick()
	// this.RAM.WriteInterpreterMemory(this.MemIndex, this.SP, uint64(value))

	this.Dec_SP()
}

func (this *Core6502) GetStack(max int) []int {
	var out = make([]int, 0, max)
	for i := 0; i < max && this.SP+1+i < 0x200; i++ {
		out = append(out, int(this.RAM.ReadInterpreterMemory(this.MemIndex, this.SP+1+i)))
	}
	return out
}

func (this *Core6502) Pop() int {
	this.Inc_SP()
	var tmp uint64
	this.mmu.Do(this.SP, memory.MA_READ, &tmp)
	this.ClockTick()
	//return int(this.RAM.ReadInterpreterMemory(this.MemIndex, this.SP))
	return int(tmp)
}

func (this *Core6502) Wait(cycles int) {
	// pause execution for n cycles - simulate real speed
	var INTERVAL int64 = int64(cycles*1000) + this.LastCycleDeficit

	var start int64 = time.Now().UnixNano()
	var end int64 = 0
	for {
		end = time.Now().UnixNano()
		if !(start+INTERVAL >= end) {
			break
		}
	}

	//System.Out.Println(end - start);
	this.LastCycleDeficit = INTERVAL - (end - start)
}

// func (this *Core6502) FetchMemoryByteAbsoluteX(addr *int) int {
// 	return this.FetchByteAddr(this.FetchWordPC(addr) + this.X)
// }

func (this *Core6502) Inc_SP() {
	this.SP = 0x100 | ((this.SP + 1) & 0xff)
}

// func (this *Core6502) FetchMemoryByteIndirectZeroPageY(addr *int) int {
// 	return this.FetchByteAddr(this.FetchWordAddr(this.FetchBytePC(addr)) + this.Y)
// }

func (this *Core6502) SetMapper(m *memory.MemoryManagementUnit) {
	this.mmu = m
}

func (this *Core6502) FetchByteAddr(addr int) int {

	if this.HasSpecialFlags {
		flags := this.GetMemFlags(addr)

		if this.testMemFlags(flags, AF_READ_BREAK) {
			this.MemoryTrip = true
		}

		if this.testMemFlags(flags, AF_READ_BUMP) {
			counter := int(flags>>32) & 0xff
			this.RAM.IntBumpCounter(this.MemIndex, counter)
		}
	}

	this.lastReadAddr = addr

	// v := int(this.RAM.ReadInterpreterMemory(this.MemIndex, addr)) & 0xff

	var tmp uint64
	this.mmu.Do(addr&0xffff, memory.MA_READ, &tmp)
	v := int(tmp) & 0xff

	this.lastReadValue = v

	this.ClockTick() // add for a read

	return v
}

func (this *Core6502) ClockTick() {
	if !this.AllowOperCycles {
		return // prevent IO clocks
	}
	this.OperCycles++
	this.GlobalCycles++
	if this.IO != nil {
		this.IO.Increment(1)
	}
}

func (this *Core6502) FetchBytePCNOP(addr *int) int {
	this.mmu.CPURead = true
	var z uint64
	this.mmu.Do(*addr, memory.MA_READ, &z)
	this.mmu.CPURead = false
	this.ClockTick()
	return int(z) & 0xff
}

func (this *Core6502) FetchBytePC(addr *int) int {
	//this.mmu.CPURead = true
	//var z uint64
	//this.mmu.Do(*addr, memory.MA_READ, &z)
	this.mmu.CPURead = true
	z := this.RAM.ReadInterpreterMemory(this.MemIndex, *addr)
	this.mmu.CPURead = false
	this.ClockTick()

	// if HEATMAP && (*addr >= 0x9600 && *addr <= 0xffff) {
	// 	this.HeatMap[fmt.Sprintf("%x", *addr)] = z // store value
	// }

	*addr++
	return int(z) & 0xff
}

// func (this *Core6502) FetchMemoryByteAbsoluteY(addr *int) int {
// 	return this.FetchByteAddr(this.FetchWordPC(addr) + this.Y)
// }

func (this *Core6502) Set_nz(v int) {
	this.P &= ^(F_Z | F_N)
	if (v & 0x80) != 0 {
		this.P |= F_N
	}
	if v == 0 {
		this.P |= F_Z
	}
	//this.FetchBytePCNOP(&this.PC)
}

// func (this *Core6502) FetchMemoryByteAbsolute() int {
// 	return this.FetchByteAddr(this.FetchWordPC())
// }

func (this *Core6502) Reset() {
	this.A = 0
	this.X = 0
	this.Y = 0
	this.PC = this.FetchWordAddr(0xfffc) // cold start
	this.SP = 0x01ff
	this.SetFlag(F_R, false)
	this.SetFlag(F_B, false)
	this.SetFlag(F_I, true)
	this.SetFlag(F_D, false)
	this.SetFlag(F_C, false)
	this.SetFlag(F_V, false)
	this.SetFlag(F_N, false)
	this.GlobalCycles = 0
	this.RunState = CrsFreeRun
	this.ResetSliced()
	//this.OpTableTest()
}

// func (this *Core6502) FetchMemoryByteZeroPage() int {
// 	return this.FetchByteAddr(this.FetchBytePC())
// }

// func (this *Core6502) FetchMemoryByteZeroPageX() int {
// 	return this.FetchByteAddr((this.FetchBytePC() + this.X) % 256)
// }

func (this *Core6502) StoreByteAddr(addr int, v int) {

	// if this.HasSpecialFlags {
	flags := this.GetMemFlags(addr)

	// if addr < 0 || addr >= 65536 || this.testMemFlags(flags, AF_WRITE_BREAK) {
	// 	this.MemoryTrip = true
	// 	return
	// }

	if this.testMemFlags(flags, AF_WRITE_BUMP) {
		counter := int(flags>>32) & 0xff
		this.RAM.IntBumpCounter(this.MemIndex, counter)
	}

	if this.testMemFlags(flags, AF_WRITE_LOCK) {
		return
	}

	if lv, ok := this.LockValue[addr]; ok {
		if lv == uint64(v) {
			this.SetMemFlags(addr, AF_WRITE_LOCK, true)
			delete(this.LockValue, addr)
		}
	}
	// }

	// if TRACEMEM {
	// 	ff.WriteString(fmt.Sprintf("memwrite: Stored 0x%.2x -> 0x%.4x\n", v, addr))
	// }

	if settings.DebuggerOn && settings.DebuggerAttachSlot == this.MemIndex+1 {
		servicebus.SendServiceBusMessage(
			this.MemIndex,
			servicebus.CPUWriteMem,
			&debugtypes.CPUMemoryWrite{
				Address: addr,
				Value:   v,
			},
		)
	}

	if addr >= 1024 && addr < 2048 {
		v = int(this.RAM.ReadInterpreterMemorySilent(this.MemIndex, addr)&0xffffff00) | (v & 0xff)
	}

	//this.Int.SetMemory(addr, uint64(v))
	this.RAM.WriteInterpreterMemory(this.MemIndex, addr, uint64(v))
	//tmp := uint64(v)
	//this.mmu.Do(addr, memory.MA_WRITE, &tmp)
	this.ClockTick()
}

func (this *Core6502) ToString() string {
	return "* PC = $" + hex.EncodeToString([]byte{byte(this.PC / 256), byte(this.PC % 256)}) + ", SP = $" + hex.EncodeToString([]byte{byte(this.SP & 0xff)}) + ", this.A = $" + hex.EncodeToString([]byte{byte(this.A & 0xff)}) + ", X = $" + hex.EncodeToString([]byte{byte(this.X & 0xff)}) + ", this.Y = $" + hex.EncodeToString([]byte{byte(this.Y & 0xff)}) + ", this.P = " + this.FlagString()
}

// func (this *Core6502) FetchMemoryByteXIndirect() int {
// 	return this.FetchByteAddr(this.FetchWordAddr(this.FetchBytePC() + this.X))
// }

func (this *Core6502) Pad(s string, bytes int) string {
	c := bytes * 2
	for len(s) < c {
		s = "0" + s
	}
	return s
}

func (this *Core6502) FlagString() string {
	out := ""
	if this.Test(F_N) {
		out += "N"
	} else {
		out += "."
	}
	if this.Test(F_V) {
		out += "V"
	} else {
		out += "."
	}
	if this.Test(F_R) {
		out += "R"
	} else {
		out += "."
	}
	if this.Test(F_B) {
		out += "B"
	} else {
		out += "."
	}
	if this.Test(F_D) {
		out += "D"
	} else {
		out += "."
	}
	if this.Test(F_I) {
		out += "I"
	} else {
		out += "."
	}
	if this.Test(F_Z) {
		out += "Z"
	} else {
		out += "."
	}
	if this.Test(F_C) {
		out += "C"
	} else {
		out += "."
	}
	return out
}

func (this *Core6502) FetchWordAddr(addr int) int {
	return this.FetchByteAddr(addr) + 256*this.FetchByteAddr(addr+1)
}

func NewCore6502(mem interfaces.Interpretable, a int, x int, y int, pc int, sr int, sp int, vdu WaveStreamer) *Core6502 {
	this := &Core6502{}
	// this.SAMPLEDATA = make([]uint64, MAXAUDIO)
	// this.SAMPLEVALUE = -1.0
	//this.MEM = make([]int, 65536)
	this.A = a
	this.X = x
	this.Y = y
	this.PC = pc
	this.P = sr
	this.SP = sp
	this.InitialSP = sp
	this.Int = mem
	this.Halted = true
	this.InitOpList()
	this.VDU = vdu
	this.BasicMode = true
	this.HeatMap = make(map[string]uint64)
	this.shim = make(map[int]Func6502)
	if this.Int != nil {
		this.MemIndex = this.Int.GetMemIndex()
		this.RAM = this.Int.GetMemoryMap()
		this.mmu = this.RAM.BlockMapper[this.MemIndex]
	}
	this.InitTime = time.Now()

	this.profiledatacum = make(map[string]int64)
	this.profiledatacount = make(map[string]int64)

	this.SpecialFlag = make(map[int]AddressFlags)
	this.LockValue = make(map[int]uint64)

	this.BaseSpeed = 1020484
	this.Warp = 1.0

	return this
}

func (this *Core6502) RegisterCallShim(address int, f Func6502) {

	this.shim[address] = f

}

func (this *Core6502) IsCallShimmable(address int) (bool, int64) {

	f, ok := this.shim[address]
	if !ok {
		return ok, 0
	}

	return ok, f(this)

}

func (this *Core6502) SetWait(uS int64) {

	// if basic mode, we need to update the wait time
	if this.BasicMode {
		microseconds := time.Duration(uS) * time.Microsecond
		this.Int.WaitAdd(microseconds)
		//fmt.Printf("Setting up basic mode wait of %v\n", microseconds)
	}

}

// func (this *Core6502) FetchMemoryByteIndirect() int {
// 	return this.FetchByteAddr(this.FetchWordAddr(this.FetchWordPC()))
// }

func (this *Core6502) Write(addr int, value int) {
	//this.Int.SetMemory(addr%65536, uint64(value&0xff))

	this.RAM.WriteInterpreterMemory(this.MemIndex, addr%65536, uint64(value&0xff))
}

func (this *Core6502) FetchWordPC(addr *int) int {
	return this.FetchBytePC(addr) + 256*this.FetchBytePC(addr)
}

func (this *Core6502) FetchWordAddrNMOS(addr int) int {

	var b1, b2 int

	if addr%256 == 255 {
		b1 = this.FetchByteAddr(addr)
		b2 = this.FetchByteAddr(addr & 0xff00)
	} else {
		b1 = this.FetchByteAddr(addr)
		b2 = this.FetchByteAddr(addr + 1)
	}

	return b1 + 256*b2
}

func (this *Core6502) TestFlag(f int) bool {
	return ((this.P & f) != 0)
}

func (this *Core6502) InitOpList() {
	this.Model = "6502"
	//this.Opref = make(map[int]*Op6502)
	this.Opref[0x69] = NewOp6502("ADC #oper", "immidiate", 0x69, 2, 2, IMMEDIATE, MODE_IMMEDIATE, ADC)
	this.Opref[0x65] = NewOp6502("ADC oper", "zeropage", 0x65, 2, 3, ZEROPAGE, MODE_ZEROPAGE, ADC)
	this.Opref[0x75] = NewOp6502("ADC oper,X", "zeropage,X", 0x75, 2, 4, ZEROPAGE_X, MODE_ZEROPAGE_X, ADC)
	this.Opref[0x6D] = NewOp6502("ADC oper", "absolute", 0x6D, 3, 4, ABSOLUTE, MODE_ABSOLUTE, ADC)
	this.Opref[0x7D] = NewOp6502("ADC oper,X", "absolute,X", 0x7D, 3, 4, ABSOLUTE_X, MODE_ABSOLUTE_X, ADC)
	this.Opref[0x79] = NewOp6502("ADC oper,Y", "absolute,Y", 0x79, 3, 4, ABSOLUTE_Y, MODE_ABSOLUTE_Y, ADC)
	this.Opref[0x61] = NewOp6502("ADC (oper,X)", "(indirect,X)", 0x61, 2, 6, INDIRECT_ZP_X, MODE_INDIRECT_ZP_X, ADC)
	this.Opref[0x71] = NewOp6502("ADC (oper),Y", "(indirect),Y", 0x71, 2, 5, INDIRECT_ZP_Y, MODE_INDIRECT_ZP_Y, ADC)
	this.Opref[0x29] = NewOp6502("AND #oper", "immidiate", 0x29, 2, 2, IMMEDIATE, MODE_IMMEDIATE, AND)
	this.Opref[0x25] = NewOp6502("AND oper", "zeropage", 0x25, 2, 3, ZEROPAGE, MODE_ZEROPAGE, AND)
	this.Opref[0x35] = NewOp6502("AND oper,X", "zeropage,X", 0x35, 2, 4, ZEROPAGE_X, MODE_ZEROPAGE_X, AND)
	this.Opref[0x2D] = NewOp6502("AND oper", "absolute", 0x2D, 3, 4, ABSOLUTE, MODE_ABSOLUTE, AND)
	this.Opref[0x3D] = NewOp6502("AND oper,X", "absolute,X", 0x3D, 3, 4, ABSOLUTE_X, MODE_ABSOLUTE_X, AND)
	this.Opref[0x39] = NewOp6502("AND oper,Y", "absolute,Y", 0x39, 3, 4, ABSOLUTE_Y, MODE_ABSOLUTE_Y, AND)
	this.Opref[0x21] = NewOp6502("AND (oper,X)", "(indirect,X)", 0x21, 2, 6, INDIRECT_ZP_X, MODE_INDIRECT_ZP_X, AND)
	this.Opref[0x31] = NewOp6502("AND (oper),Y", "(indirect),Y", 0x31, 2, 5, INDIRECT_ZP_Y, MODE_INDIRECT_ZP_Y, AND)
	this.Opref[0x0A] = NewOp6502("ASL A", "accumulator", 0x0A, 1, 2, IMPLIED, MODE_IMPLIED, ASL)
	this.Opref[0x06] = NewOp6502("ASL oper", "zeropage", 0x06, 2, 5, ZEROPAGE, MODE_ZEROPAGE, ASL)
	this.Opref[0x16] = NewOp6502("ASL oper,X", "zeropage,X", 0x16, 2, 6, ZEROPAGE_X, MODE_ZEROPAGE_X, ASL)
	this.Opref[0x0E] = NewOp6502("ASL oper", "absolute", 0x0E, 3, 6, ABSOLUTE, MODE_ABSOLUTE, ASL)
	this.Opref[0x1E] = NewOp6502("ASL oper,X", "absolute,X", 0x1E, 3, 7, ABSOLUTE_X, MODE_ABSOLUTE_X, ASL)
	this.Opref[0x90] = NewOp6502("BCC oper", "relative", 0x90, 2, 2, RELATIVE, MODE_RELATIVE, BCC)
	this.Opref[0xB0] = NewOp6502("BCS oper", "relative", 0xB0, 2, 2, RELATIVE, MODE_RELATIVE, BCS)
	this.Opref[0xF0] = NewOp6502("BEQ oper", "relative", 0xF0, 2, 2, RELATIVE, MODE_RELATIVE, BEQ)
	this.Opref[0x24] = NewOp6502("BIT oper", "zeropage", 0x24, 2, 3, ZEROPAGE, MODE_ZEROPAGE, BIT)
	this.Opref[0x2C] = NewOp6502("BIT oper", "absolute", 0x2C, 3, 4, ABSOLUTE, MODE_ABSOLUTE, BIT)
	this.Opref[0x30] = NewOp6502("BMI oper", "relative", 0x30, 2, 2, RELATIVE, MODE_RELATIVE, BMI)
	this.Opref[0xD0] = NewOp6502("BNE oper", "relative", 0xD0, 2, 2, RELATIVE, MODE_RELATIVE, BNE)
	this.Opref[0x10] = NewOp6502("BPL oper", "relative", 0x10, 2, 2, RELATIVE, MODE_RELATIVE, BPL)
	this.Opref[0x00] = NewOp6502("BRK", "implied", 0x00, 1, 7, IMPLIED, MODE_IMPLIED, BRK)
	this.Opref[0x50] = NewOp6502("BVC oper", "relative", 0x50, 2, 2, RELATIVE, MODE_RELATIVE, BVC)
	this.Opref[0x70] = NewOp6502("BVS oper", "relative", 0x70, 2, 2, RELATIVE, MODE_RELATIVE, BVS)
	this.Opref[0x18] = NewOp6502("CLC", "implied", 0x18, 1, 2, IMPLIED, MODE_IMPLIED, CLC)
	this.Opref[0xD8] = NewOp6502("CLD", "implied", 0xD8, 1, 2, IMPLIED, MODE_IMPLIED, CLD)
	this.Opref[0x58] = NewOp6502("CLI", "implied", 0x58, 1, 2, IMPLIED, MODE_IMPLIED, CLI)
	this.Opref[0xB8] = NewOp6502("CLV", "implied", 0xB8, 1, 2, IMPLIED, MODE_IMPLIED, CLV)
	this.Opref[0xC9] = NewOp6502("CMP #oper", "immidiate", 0xC9, 2, 2, IMMEDIATE, MODE_IMMEDIATE, CMP)
	this.Opref[0xC5] = NewOp6502("CMP oper", "zeropage", 0xC5, 2, 3, ZEROPAGE, MODE_ZEROPAGE, CMP)
	this.Opref[0xD5] = NewOp6502("CMP oper,X", "zeropage,X", 0xD5, 2, 4, ZEROPAGE_X, MODE_ZEROPAGE_X, CMP)
	this.Opref[0xCD] = NewOp6502("CMP oper", "absolute", 0xCD, 3, 4, ABSOLUTE, MODE_ABSOLUTE, CMP)
	this.Opref[0xDD] = NewOp6502("CMP oper,X", "absolute,X", 0xDD, 3, 4, ABSOLUTE_X, MODE_ABSOLUTE_X, CMP)
	this.Opref[0xD9] = NewOp6502("CMP oper,Y", "absolute,Y", 0xD9, 3, 4, ABSOLUTE_Y, MODE_ABSOLUTE_Y, CMP)
	this.Opref[0xC1] = NewOp6502("CMP (oper,X)", "(indirect,X)", 0xC1, 2, 6, INDIRECT_ZP_X, MODE_INDIRECT_ZP_X, CMP)
	this.Opref[0xD1] = NewOp6502("CMP (oper),Y", "(indirect),Y", 0xD1, 2, 5, INDIRECT_ZP_Y, MODE_INDIRECT_ZP_Y, CMP)
	this.Opref[0xE0] = NewOp6502("CPX #oper", "immidiate", 0xE0, 2, 2, IMMEDIATE, MODE_IMMEDIATE, CPX)
	this.Opref[0xE4] = NewOp6502("CPX oper", "zeropage", 0xE4, 2, 3, ZEROPAGE, MODE_ZEROPAGE, CPX)
	this.Opref[0xEC] = NewOp6502("CPX oper", "absolute", 0xEC, 3, 4, ABSOLUTE, MODE_ABSOLUTE, CPX)
	this.Opref[0xC0] = NewOp6502("CPY #oper", "immidiate", 0xC0, 2, 2, IMMEDIATE, MODE_IMMEDIATE, CPY)
	this.Opref[0xC4] = NewOp6502("CPY oper", "zeropage", 0xC4, 2, 3, ZEROPAGE, MODE_ZEROPAGE, CPY)
	this.Opref[0xCC] = NewOp6502("CPY oper", "absolute", 0xCC, 3, 4, ABSOLUTE, MODE_ABSOLUTE, CPY)
	this.Opref[0xC6] = NewOp6502("DEC oper", "zeropage", 0xC6, 2, 5, ZEROPAGE, MODE_ZEROPAGE, DEC)
	this.Opref[0xD6] = NewOp6502("DEC oper,X", "zeropage,X", 0xD6, 2, 6, ZEROPAGE_X, MODE_ZEROPAGE_X, DEC)
	this.Opref[0xCE] = NewOp6502("DEC oper", "absolute", 0xCE, 3, 6, ABSOLUTE, MODE_ABSOLUTE, DEC)
	this.Opref[0xDE] = NewOp6502("DEC oper,X", "absolute,X", 0xDE, 3, 7, ABSOLUTE_X, MODE_ABSOLUTE_X, DEC)
	this.Opref[0xCA] = NewOp6502("DEX", "implied", 0xCA, 1, 2, IMPLIED, MODE_IMPLIED, DEX)
	this.Opref[0x88] = NewOp6502("DEY", "implied", 0x88, 1, 2, IMPLIED, MODE_IMPLIED, DEY)
	this.Opref[0x49] = NewOp6502("EOR #oper", "immidiate", 0x49, 2, 2, IMMEDIATE, MODE_IMMEDIATE, EOR)
	this.Opref[0x45] = NewOp6502("EOR oper", "zeropage", 0x45, 2, 3, ZEROPAGE, MODE_ZEROPAGE, EOR)
	this.Opref[0x55] = NewOp6502("EOR oper,X", "zeropage,X", 0x55, 2, 4, ZEROPAGE_X, MODE_ZEROPAGE_X, EOR)
	this.Opref[0x4D] = NewOp6502("EOR oper", "absolute", 0x4D, 3, 4, ABSOLUTE, MODE_ABSOLUTE, EOR)
	this.Opref[0x5D] = NewOp6502("EOR oper,X", "absolute,X", 0x5D, 3, 4, ABSOLUTE_X, MODE_ABSOLUTE_X, EOR)
	this.Opref[0x59] = NewOp6502("EOR oper,Y", "absolute,Y", 0x59, 3, 4, ABSOLUTE_Y, MODE_ABSOLUTE_Y, EOR)
	this.Opref[0x41] = NewOp6502("EOR (oper,X)", "(indirect,X)", 0x41, 2, 6, INDIRECT_ZP_X, MODE_INDIRECT_ZP_X, EOR)
	this.Opref[0x51] = NewOp6502("EOR (oper),Y", "(indirect),Y", 0x51, 2, 5, INDIRECT_ZP_Y, MODE_INDIRECT_ZP_Y, EOR)
	this.Opref[0xE6] = NewOp6502("INC oper", "zeropage", 0xE6, 2, 5, ZEROPAGE, MODE_ZEROPAGE, INC)
	this.Opref[0xF6] = NewOp6502("INC oper,X", "zeropage,X", 0xF6, 2, 6, ZEROPAGE_X, MODE_ZEROPAGE_X, INC)
	this.Opref[0xEE] = NewOp6502("INC oper", "absolute", 0xEE, 3, 6, ABSOLUTE, MODE_ABSOLUTE, INC)
	this.Opref[0xFE] = NewOp6502("INC oper,X", "absolute,X", 0xFE, 3, 7, ABSOLUTE_X, MODE_ABSOLUTE_X, INC)
	this.Opref[0xE8] = NewOp6502("INX", "implied", 0xE8, 1, 2, IMPLIED, MODE_IMPLIED, INX)
	this.Opref[0xC8] = NewOp6502("INY", "implied", 0xC8, 1, 2, IMPLIED, MODE_IMPLIED, INY)
	this.Opref[0x4C] = NewOp6502("JMP oper", "absolute", 0x4C, 3, 3, ABSOLUTE, MODE_ABSOLUTE, JMP)
	this.Opref[0x6C] = NewOp6502("JMP (oper)", "indirect", 0x6C, 3, 5, INDIRECT_NMOS, MODE_INDIRECT_NMOS, JMP)
	this.Opref[0x20] = NewOp6502("JSR oper", "absolute", 0x20, 3, 6, ABSOLUTE, MODE_ABSOLUTE, JSR)
	this.Opref[0xA9] = NewOp6502("LDA #oper", "immidiate", 0xA9, 2, 2, IMMEDIATE, MODE_IMMEDIATE, LDA)
	this.Opref[0xA5] = NewOp6502("LDA oper", "zeropage", 0xA5, 2, 3, ZEROPAGE, MODE_ZEROPAGE, LDA)
	this.Opref[0xB5] = NewOp6502("LDA oper,X", "zeropage,X", 0xB5, 2, 4, ZEROPAGE_X, MODE_ZEROPAGE_X, LDA)
	this.Opref[0xAD] = NewOp6502("LDA oper", "absolute", 0xAD, 3, 4, ABSOLUTE, MODE_ABSOLUTE, LDA)
	this.Opref[0xBD] = NewOp6502("LDA oper,X", "absolute,X", 0xBD, 3, 4, ABSOLUTE_X, MODE_ABSOLUTE_X, LDA)
	this.Opref[0xB9] = NewOp6502("LDA oper,Y", "absolute,Y", 0xB9, 3, 4, ABSOLUTE_Y, MODE_ABSOLUTE_Y, LDA)
	this.Opref[0xA1] = NewOp6502("LDA (oper,X)", "(indirect,X)", 0xA1, 2, 6, INDIRECT_ZP_X, MODE_INDIRECT_ZP_X, LDA)
	this.Opref[0xB1] = NewOp6502("LDA (oper),Y", "(indirect),Y", 0xB1, 2, 5, INDIRECT_ZP_Y, MODE_INDIRECT_ZP_Y, LDA)
	this.Opref[0xA2] = NewOp6502("LDX #oper", "immidiate", 0xA2, 2, 2, IMMEDIATE, MODE_IMMEDIATE, LDX)
	this.Opref[0xA6] = NewOp6502("LDX oper", "zeropage", 0xA6, 2, 3, ZEROPAGE, MODE_ZEROPAGE, LDX)
	this.Opref[0xB6] = NewOp6502("LDX oper,Y", "zeropage,Y", 0xB6, 2, 4, ZEROPAGE_Y, MODE_ZEROPAGE_Y, LDX)
	this.Opref[0xAE] = NewOp6502("LDX oper", "absolute", 0xAE, 3, 4, ABSOLUTE, MODE_ABSOLUTE, LDX)
	this.Opref[0xBE] = NewOp6502("LDX oper,Y", "absolute,Y", 0xBE, 3, 4, ABSOLUTE_Y, MODE_ABSOLUTE_Y, LDX)
	this.Opref[0xA0] = NewOp6502("LDY #oper", "immidiate", 0xA0, 2, 2, IMMEDIATE, MODE_IMMEDIATE, LDY)
	this.Opref[0xA4] = NewOp6502("LDY oper", "zeropage", 0xA4, 2, 3, ZEROPAGE, MODE_ZEROPAGE, LDY)
	this.Opref[0xB4] = NewOp6502("LDY oper,X", "zeropage,X", 0xB4, 2, 4, ZEROPAGE_X, MODE_ZEROPAGE_X, LDY)
	this.Opref[0xAC] = NewOp6502("LDY oper", "absolute", 0xAC, 3, 4, ABSOLUTE, MODE_ABSOLUTE, LDY)
	this.Opref[0xBC] = NewOp6502("LDY oper,X", "absolute,X", 0xBC, 3, 4, ABSOLUTE_X, MODE_ABSOLUTE_X, LDY)
	this.Opref[0x4A] = NewOp6502("LSR A", "accumulator", 0x4A, 1, 2, IMPLIED, MODE_IMPLIED, LSR)
	this.Opref[0x46] = NewOp6502("LSR oper", "zeropage", 0x46, 2, 5, ZEROPAGE, MODE_ZEROPAGE, LSR)
	this.Opref[0x56] = NewOp6502("LSR oper,X", "zeropage,X", 0x56, 2, 6, ZEROPAGE_X, MODE_ZEROPAGE_X, LSR)
	this.Opref[0x4E] = NewOp6502("LSR oper", "absolute", 0x4E, 3, 6, ABSOLUTE, MODE_ABSOLUTE, LSR)
	this.Opref[0x5E] = NewOp6502("LSR oper,X", "absolute,X", 0x5E, 3, 7, ABSOLUTE_X, MODE_ABSOLUTE_X, LSR)
	this.Opref[0xEA] = NewOp6502("NOP", "implied", 0xEA, 1, 2, IMPLIED, MODE_IMPLIED, NOP)
	this.Opref[0x09] = NewOp6502("ORA #oper", "immidiate", 0x09, 2, 2, IMMEDIATE, MODE_IMMEDIATE, ORA)
	this.Opref[0x05] = NewOp6502("ORA oper", "zeropage", 0x05, 2, 3, ZEROPAGE, MODE_ZEROPAGE, ORA)
	this.Opref[0x15] = NewOp6502("ORA oper,X", "zeropage,X", 0x15, 2, 4, ZEROPAGE_X, MODE_ZEROPAGE_X, ORA)
	this.Opref[0x0D] = NewOp6502("ORA oper", "absolute", 0x0D, 3, 4, ABSOLUTE, MODE_ABSOLUTE, ORA)
	this.Opref[0x1D] = NewOp6502("ORA oper,X", "absolute,X", 0x1D, 3, 4, ABSOLUTE_X, MODE_ABSOLUTE_X, ORA)
	this.Opref[0x19] = NewOp6502("ORA oper,Y", "absolute,Y", 0x19, 3, 4, ABSOLUTE_Y, MODE_ABSOLUTE_Y, ORA)
	this.Opref[0x01] = NewOp6502("ORA (oper,X)", "(indirect,X)", 0x01, 2, 6, INDIRECT_ZP_X, MODE_INDIRECT_ZP_X, ORA)
	this.Opref[0x11] = NewOp6502("ORA (oper),Y", "(indirect),Y", 0x11, 2, 5, INDIRECT_ZP_Y, MODE_INDIRECT_ZP_Y, ORA)
	this.Opref[0x48] = NewOp6502("PHA", "implied", 0x48, 1, 3, IMPLIED, MODE_IMPLIED, PHA)
	this.Opref[0x08] = NewOp6502("PHP", "implied", 0x08, 1, 3, IMPLIED, MODE_IMPLIED, PHP)
	this.Opref[0x68] = NewOp6502("PLA", "implied", 0x68, 1, 4, IMPLIED, MODE_IMPLIED, PLA)
	this.Opref[0x28] = NewOp6502("PLP", "implied", 0x28, 1, 4, IMPLIED, MODE_IMPLIED, PLP)
	this.Opref[0x2A] = NewOp6502("ROL A", "accumulator", 0x2A, 1, 2, IMPLIED, MODE_IMPLIED, ROL)
	this.Opref[0x26] = NewOp6502("ROL oper", "zeropage", 0x26, 2, 5, ZEROPAGE, MODE_ZEROPAGE, ROL)
	this.Opref[0x36] = NewOp6502("ROL oper,X", "zeropage,X", 0x36, 2, 6, ZEROPAGE_X, MODE_ZEROPAGE_X, ROL)
	this.Opref[0x2E] = NewOp6502("ROL oper", "absolute", 0x2E, 3, 6, ABSOLUTE, MODE_ABSOLUTE, ROL)
	this.Opref[0x3E] = NewOp6502("ROL oper,X", "absolute,X", 0x3E, 3, 7, ABSOLUTE_X, MODE_ABSOLUTE_X, ROL)
	this.Opref[0x6A] = NewOp6502("ROR A", "accumulator", 0x6A, 1, 2, IMPLIED, MODE_IMPLIED, ROR)
	this.Opref[0x66] = NewOp6502("ROR oper", "zeropage", 0x66, 2, 5, ZEROPAGE, MODE_ZEROPAGE, ROR)
	this.Opref[0x76] = NewOp6502("ROR oper,X", "zeropage,X", 0x76, 2, 6, ZEROPAGE_X, MODE_ZEROPAGE_X, ROR)
	this.Opref[0x6E] = NewOp6502("ROR oper", "absolute", 0x6E, 3, 6, ABSOLUTE, MODE_ABSOLUTE, ROR)
	this.Opref[0x7E] = NewOp6502("ROR oper,X", "absolute,X", 0x7E, 3, 7, ABSOLUTE_X, MODE_ABSOLUTE_X, ROR)
	this.Opref[0x40] = NewOp6502("RTI", "implied", 0x40, 1, 6, IMPLIED, MODE_IMPLIED, RTI)
	this.Opref[0x60] = NewOp6502("RTS", "implied", 0x60, 1, 6, IMPLIED, MODE_IMPLIED, RTS)
	this.Opref[0xE9] = NewOp6502("SBC #oper", "immidiate", 0xE9, 2, 2, IMMEDIATE, MODE_IMMEDIATE, SBC)
	this.Opref[0xE5] = NewOp6502("SBC oper", "zeropage", 0xE5, 2, 3, ZEROPAGE, MODE_ZEROPAGE, SBC)
	this.Opref[0xF5] = NewOp6502("SBC oper,X", "zeropage,X", 0xF5, 2, 4, ZEROPAGE_X, MODE_ZEROPAGE_X, SBC)
	this.Opref[0xED] = NewOp6502("SBC oper", "absolute", 0xED, 3, 4, ABSOLUTE, MODE_ABSOLUTE, SBC)
	this.Opref[0xFD] = NewOp6502("SBC oper,X", "absolute,X", 0xFD, 3, 4, ABSOLUTE_X, MODE_ABSOLUTE_X, SBC)
	this.Opref[0xF9] = NewOp6502("SBC oper,Y", "absolute,Y", 0xF9, 3, 4, ABSOLUTE_Y, MODE_ABSOLUTE_Y, SBC)
	this.Opref[0xE1] = NewOp6502("SBC (oper,X)", "(indirect,X)", 0xE1, 2, 6, INDIRECT_ZP_X, MODE_INDIRECT_ZP_X, SBC)
	this.Opref[0xF1] = NewOp6502("SBC (oper),Y", "(indirect),Y", 0xF1, 2, 5, INDIRECT_ZP_Y, MODE_INDIRECT_ZP_Y, SBC)
	this.Opref[0x38] = NewOp6502("SEC", "implied", 0x38, 1, 2, IMPLIED, MODE_IMPLIED, SEC)
	this.Opref[0xF8] = NewOp6502("SED", "implied", 0xF8, 1, 2, IMPLIED, MODE_IMPLIED, SED)
	this.Opref[0x78] = NewOp6502("SEI", "implied", 0x78, 1, 2, IMPLIED, MODE_IMPLIED, SEI)
	this.Opref[0x85] = NewOp6502("STA oper", "zeropage", 0x85, 2, 3, ZEROPAGE, MODE_ZEROPAGE, STA)
	this.Opref[0x95] = NewOp6502("STA oper,X", "zeropage,X", 0x95, 2, 4, ZEROPAGE_X, MODE_ZEROPAGE_X, STA)
	this.Opref[0x8D] = NewOp6502("STA oper", "absolute", 0x8D, 3, 4, ABSOLUTE, MODE_ABSOLUTE, STA)
	this.Opref[0x9D] = NewOp6502("STA oper,X", "absolute,X", 0x9D, 3, 5, ABSOLUTE_X, MODE_ABSOLUTE_X, STA)
	this.Opref[0x99] = NewOp6502("STA oper,Y", "absolute,Y", 0x99, 3, 5, ABSOLUTE_Y, MODE_ABSOLUTE_Y, STA)
	this.Opref[0x81] = NewOp6502("STA (oper,X)", "(indirect,X)", 0x81, 2, 6, INDIRECT_ZP_X, MODE_INDIRECT_ZP_X, STA)
	this.Opref[0x91] = NewOp6502("STA (oper),Y", "(indirect),Y", 0x91, 2, 6, INDIRECT_ZP_Y, MODE_INDIRECT_ZP_Y, STA)
	this.Opref[0x86] = NewOp6502("STX oper", "zeropage", 0x86, 2, 3, ZEROPAGE, MODE_ZEROPAGE, STX)
	this.Opref[0x96] = NewOp6502("STX oper,Y", "zeropage,Y", 0x96, 2, 4, ZEROPAGE_Y, MODE_ZEROPAGE_Y, STX)
	this.Opref[0x8E] = NewOp6502("STX oper", "absolute", 0x8E, 3, 4, ABSOLUTE, MODE_ABSOLUTE, STX)
	this.Opref[0x84] = NewOp6502("STY oper", "zeropage", 0x84, 2, 3, ZEROPAGE, MODE_ZEROPAGE, STY)
	this.Opref[0x94] = NewOp6502("STY oper,X", "zeropage,X", 0x94, 2, 4, ZEROPAGE_X, MODE_ZEROPAGE_X, STY)
	this.Opref[0x8C] = NewOp6502("STY oper", "absolute", 0x8C, 3, 4, ABSOLUTE, MODE_ABSOLUTE, STY)
	this.Opref[0xAA] = NewOp6502("TAX", "implied", 0xAA, 1, 2, IMPLIED, MODE_IMPLIED, TAX)
	this.Opref[0xA8] = NewOp6502("TAY", "implied", 0xA8, 1, 2, IMPLIED, MODE_IMPLIED, TAY)
	this.Opref[0xBA] = NewOp6502("TSX", "implied", 0xBA, 1, 2, IMPLIED, MODE_IMPLIED, TSX)
	this.Opref[0x8A] = NewOp6502("TXA", "implied", 0x8A, 1, 2, IMPLIED, MODE_IMPLIED, TXA)
	this.Opref[0x9A] = NewOp6502("TXS", "implied", 0x9A, 1, 2, IMPLIED, MODE_IMPLIED, TXS)
	this.Opref[0x98] = NewOp6502("TYA", "implied", 0x98, 1, 2, IMPLIED, MODE_IMPLIED, TYA)

	// Undocumented opcode - DOP (double NOP)
	this.Opref[0x04] = NewOp6502("dop #oper", "immediate", 0x04, 2, 3, IMMEDIATE, MODE_IMMEDIATE, NOP)
	this.Opref[0x14] = NewOp6502("dop #oper", "immediate", 0x14, 2, 4, IMMEDIATE, MODE_IMMEDIATE, NOP)
	this.Opref[0x34] = NewOp6502("dop #oper", "immediate", 0x34, 2, 4, IMMEDIATE, MODE_IMMEDIATE, NOP)
	this.Opref[0x44] = NewOp6502("dop #oper", "immediate", 0x44, 2, 3, IMMEDIATE, MODE_IMMEDIATE, NOP)
	this.Opref[0x54] = NewOp6502("dop #oper", "immediate", 0x54, 2, 4, IMMEDIATE, MODE_IMMEDIATE, NOP)
	this.Opref[0x64] = NewOp6502("dop #oper", "immediate", 0x64, 2, 3, IMMEDIATE, MODE_IMMEDIATE, NOP)
	this.Opref[0x74] = NewOp6502("dop #oper", "immediate", 0x74, 2, 4, IMMEDIATE, MODE_IMMEDIATE, NOP)
	this.Opref[0x80] = NewOp6502("dop #oper", "immediate", 0x80, 2, 2, IMMEDIATE, MODE_IMMEDIATE, NOP)
	this.Opref[0x82] = NewOp6502("dop #oper", "immediate", 0x82, 2, 2, IMMEDIATE, MODE_IMMEDIATE, NOP)
	this.Opref[0x89] = NewOp6502("dop #oper", "immediate", 0x89, 2, 2, IMMEDIATE, MODE_IMMEDIATE, NOP)
	this.Opref[0xc2] = NewOp6502("dop #oper", "immediate", 0xc2, 2, 2, IMMEDIATE, MODE_IMMEDIATE, NOP)
	this.Opref[0xd4] = NewOp6502("dop #oper", "immediate", 0xd4, 2, 4, IMMEDIATE, MODE_IMMEDIATE, NOP)
	this.Opref[0xe2] = NewOp6502("dop #oper", "immediate", 0xe2, 2, 2, IMMEDIATE, MODE_IMMEDIATE, NOP)
	this.Opref[0xf4] = NewOp6502("dop #oper", "immediate", 0xf4, 2, 4, IMMEDIATE, MODE_IMMEDIATE, NOP)

	// SAX
	this.Opref[0x9F] = NewOp6502("sax oper,Y", "oper,Y", 0x9F, 3, 5, ABSOLUTE_Y, MODE_ABSOLUTE_Y, SAX)
	this.Opref[0x93] = NewOp6502("sax (oper),Y", "(oper),Y", 0x93, 2, 6, INDIRECT_ZP_Y, MODE_INDIRECT_ZP_Y, SAX)

	// AXA
	this.Opref[0x8B] = NewOp6502("axa #oper", "#oper", 0x8B, 2, 2, IMMEDIATE, MODE_IMMEDIATE, AXA)

	// SBX
	this.Opref[0xCB] = NewOp6502("sbx #oper", "#oper", 0xCB, 2, 2, IMMEDIATE, MODE_IMMEDIATE, SBX)

	// SXA
	this.Opref[0x83] = NewOp6502("sxa (oper,X)", "(oper,X)", 0x83, 2, 6, INDIRECT_ZP_X, MODE_INDIRECT_ZP_X, SXA)
	this.Opref[0x87] = NewOp6502("sxa oper", "oper", 0x87, 2, 3, ZEROPAGE, MODE_ZEROPAGE, SXA)
	this.Opref[0x8F] = NewOp6502("sxa oper", "oper", 0x8F, 3, 4, ZEROPAGE, MODE_ZEROPAGE, SXA)
	this.Opref[0x93] = NewOp6502("sxa (oper),Y", "(oper),Y", 0x93, 2, 6, INDIRECT_ZP_Y, MODE_INDIRECT_ZP_Y, SXA)
	this.Opref[0x97] = NewOp6502("sxa oper,Y", "oper,Y", 0x97, 2, 4, ZEROPAGE_Y, MODE_ZEROPAGE_Y, SXA)

	// NOP 1A, 3A, 5A, 7A, DA, FA.
	this.Opref[0x1A] = NewOp6502("nop", "implied", 0x1A, 1, 2, IMPLIED, MODE_IMPLIED, NOP)
	this.Opref[0x3A] = NewOp6502("nop", "implied", 0x3A, 1, 2, IMPLIED, MODE_IMPLIED, NOP)
	this.Opref[0x5A] = NewOp6502("nop", "implied", 0x5A, 1, 2, IMPLIED, MODE_IMPLIED, NOP)
	this.Opref[0x7A] = NewOp6502("nop", "implied", 0x7A, 1, 2, IMPLIED, MODE_IMPLIED, NOP)
	this.Opref[0xDA] = NewOp6502("nop", "implied", 0xDA, 1, 2, IMPLIED, MODE_IMPLIED, NOP)
	this.Opref[0xFA] = NewOp6502("nop", "implied", 0xFA, 1, 2, IMPLIED, MODE_IMPLIED, NOP)

	// NOP 1A, 3A, 5A, 7A, DA, FA.
	this.Opref[0x0C] = NewOp6502("nop oper", "oper", 0x0C, 3, 4, ABSOLUTE, MODE_ABSOLUTE, NOP)
	this.Opref[0x1C] = NewOp6502("nop oper", "oper", 0x1C, 3, 4, ABSOLUTE, MODE_ABSOLUTE, NOP)
	this.Opref[0x3C] = NewOp6502("nop oper", "oper", 0x3C, 3, 4, ABSOLUTE, MODE_ABSOLUTE, NOP)
	this.Opref[0x5C] = NewOp6502("nop oper", "oper", 0x5C, 3, 4, ABSOLUTE, MODE_ABSOLUTE, NOP)
	this.Opref[0x7C] = NewOp6502("nop oper", "oper", 0x7C, 3, 4, ABSOLUTE, MODE_ABSOLUTE, NOP)
	this.Opref[0xDC] = NewOp6502("nop oper", "oper", 0xDC, 3, 4, ABSOLUTE, MODE_ABSOLUTE, NOP)
	this.Opref[0xFC] = NewOp6502("nop oper", "oper", 0xFC, 3, 4, ABSOLUTE, MODE_ABSOLUTE, NOP)

	// SLO
	this.Opref[0x0F] = NewOp6502("slo oper", "oper", 0x0F, 3, 6, ABSOLUTE, MODE_ABSOLUTE, SLO)
	this.Opref[0x1F] = NewOp6502("slo oper,X", "oper,X", 0x1F, 3, 7, ABSOLUTE_X, MODE_ABSOLUTE_X, SLO)
	this.Opref[0x1B] = NewOp6502("slo oper,Y", "oper,Y", 0x1B, 3, 7, ABSOLUTE_Y, MODE_ABSOLUTE_Y, SLO)
	this.Opref[0x07] = NewOp6502("slo oper", "oper", 0x07, 2, 5, ZEROPAGE, MODE_ZEROPAGE, SLO)
	this.Opref[0x17] = NewOp6502("slo oper,X", "oper,X", 0x17, 2, 6, ZEROPAGE_X, MODE_ZEROPAGE_X, SLO)
	this.Opref[0x03] = NewOp6502("slo (oper,X)", "(oper,X)", 0x03, 2, 8, INDIRECT_ZP_X, MODE_INDIRECT_ZP_X, SLO)
	this.Opref[0x13] = NewOp6502("slo (oper),Y", "(oper),Y", 0x13, 2, 8, INDIRECT_ZP_Y, MODE_INDIRECT_ZP_Y, SLO)

	// LXA
	this.Opref[0xAF] = NewOp6502("lxa oper", "oper", 0xAF, 3, 4, ABSOLUTE, MODE_ABSOLUTE, LXA)
	this.Opref[0xBF] = NewOp6502("lxa oper,Y", "oper,Y", 0xBF, 2, 4, ABSOLUTE_Y, MODE_ABSOLUTE_Y, LXA)
	this.Opref[0xA7] = NewOp6502("lxa oper", "oper", 0xA7, 2, 3, ZEROPAGE, MODE_ZEROPAGE, LXA)
	this.Opref[0xB7] = NewOp6502("lxa oper,Y", "oper,Y", 0xB7, 2, 4, ZEROPAGE_Y, MODE_ZEROPAGE_Y, LXA)
	this.Opref[0xA3] = NewOp6502("lxa (oper,X)", "(oper,X)", 0xA3, 2, 6, INDIRECT_ZP_X, MODE_INDIRECT_ZP_X, LXA)
	this.Opref[0xB3] = NewOp6502("lxa (oper),Y", "(oper),Y", 0xB3, 2, 5, INDIRECT_ZP_Y, MODE_INDIRECT_ZP_Y, LXA)

	// SRE
	this.Opref[0x4F] = NewOp6502("sre oper", "oper", 0x4F, 3, 6, ABSOLUTE, MODE_ABSOLUTE, SRE)
	this.Opref[0x5F] = NewOp6502("sre oper,X", "oper,X", 0x5F, 3, 7, ABSOLUTE_X, MODE_ABSOLUTE_X, SRE)
	this.Opref[0x5B] = NewOp6502("sre oper,Y", "oper,Y", 0x5B, 3, 7, ABSOLUTE_Y, MODE_ABSOLUTE_Y, SRE)
	this.Opref[0x47] = NewOp6502("sre oper", "oper", 0x47, 2, 5, ZEROPAGE, MODE_ZEROPAGE, SRE)
	this.Opref[0x57] = NewOp6502("sre oper,X", "oper,X", 0x57, 2, 6, ZEROPAGE_X, MODE_ZEROPAGE_X, SRE)
	this.Opref[0x43] = NewOp6502("sre (oper,X)", "(oper,X)", 0x43, 2, 8, INDIRECT_ZP_X, MODE_INDIRECT_ZP_X, SRE)
	this.Opref[0x53] = NewOp6502("sre (oper),Y", "(oper),Y", 0x53, 2, 8, INDIRECT_ZP_Y, MODE_INDIRECT_ZP_Y, SRE)

	// RRA
	this.Opref[0x6F] = NewOp6502("rra oper", "oper", 0x6F, 3, 6, ABSOLUTE, MODE_ABSOLUTE, RRA)
	this.Opref[0x7F] = NewOp6502("rra oper,X", "oper,X", 0x7F, 3, 7, ABSOLUTE_X, MODE_ABSOLUTE_X, RRA)
	this.Opref[0x7B] = NewOp6502("rra oper,Y", "oper,Y", 0x7B, 3, 7, ABSOLUTE_Y, MODE_ABSOLUTE_Y, RRA)
	this.Opref[0x67] = NewOp6502("rra oper", "oper", 0x67, 2, 5, ZEROPAGE, MODE_ZEROPAGE, RRA)
	this.Opref[0x77] = NewOp6502("rra oper,X", "oper,X", 0x77, 2, 6, ZEROPAGE_X, MODE_ZEROPAGE_X, RRA)
	this.Opref[0x63] = NewOp6502("rra (oper,X)", "(oper,X)", 0x63, 2, 8, INDIRECT_ZP_X, MODE_INDIRECT_ZP_X, RRA)
	this.Opref[0x73] = NewOp6502("rra (oper),Y", "(oper),Y", 0x73, 2, 8, INDIRECT_ZP_Y, MODE_INDIRECT_ZP_Y, RRA)

	// RLA
	this.Opref[0x2F] = NewOp6502("rla oper", "oper", 0x2F, 3, 6, ABSOLUTE, MODE_ABSOLUTE, RLA)
	this.Opref[0x3F] = NewOp6502("rla oper,X", "oper,X", 0x3F, 3, 7, ABSOLUTE_X, MODE_ABSOLUTE_X, RLA)
	this.Opref[0x3B] = NewOp6502("rla oper,Y", "oper,Y", 0x3B, 3, 7, ABSOLUTE_Y, MODE_ABSOLUTE_Y, RLA)
	this.Opref[0x27] = NewOp6502("rla oper", "oper", 0x27, 2, 5, ZEROPAGE, MODE_ZEROPAGE, RLA)
	this.Opref[0x37] = NewOp6502("rla oper,X", "oper,X", 0x37, 2, 6, ZEROPAGE_X, MODE_ZEROPAGE_X, RLA)
	this.Opref[0x23] = NewOp6502("rla (oper,X)", "(oper,X)", 0x23, 2, 8, INDIRECT_ZP_X, MODE_INDIRECT_ZP_X, RLA)
	this.Opref[0x33] = NewOp6502("rla (oper),Y", "(oper),Y", 0x33, 2, 8, INDIRECT_ZP_Y, MODE_INDIRECT_ZP_Y, RLA)

	// INS
	this.Opref[0xEF] = NewOp6502("ins oper", "oper", 0xEF, 3, 6, ABSOLUTE, MODE_ABSOLUTE, INS)
	this.Opref[0xFF] = NewOp6502("ins oper,X", "oper,X", 0xFF, 3, 7, ABSOLUTE_X, MODE_ABSOLUTE_X, INS)
	this.Opref[0xFB] = NewOp6502("ins oper,Y", "oper,Y", 0xFB, 3, 7, ABSOLUTE_Y, MODE_ABSOLUTE_Y, INS)
	this.Opref[0xE7] = NewOp6502("ins oper", "oper", 0xE7, 2, 5, ZEROPAGE, MODE_ZEROPAGE, INS)
	this.Opref[0xF7] = NewOp6502("ins oper,X", "oper,X", 0xF7, 2, 6, ZEROPAGE_X, MODE_ZEROPAGE_X, INS)
	this.Opref[0xE3] = NewOp6502("ins (oper,X)", "(oper,X)", 0xE3, 2, 8, INDIRECT_ZP_X, MODE_INDIRECT_ZP_X, INS)
	this.Opref[0xF3] = NewOp6502("ins (oper),Y", "(oper),Y", 0xF3, 2, 8, INDIRECT_ZP_Y, MODE_INDIRECT_ZP_Y, INS)

	// ANC
	this.Opref[0x0B] = NewOp6502("anc #oper", "#oper", 0x0B, 2, 2, IMMEDIATE, MODE_IMMEDIATE, ANC)
	this.Opref[0x2B] = NewOp6502("anc #oper", "#oper", 0x2B, 2, 2, IMMEDIATE, MODE_IMMEDIATE, ANC)

	// 02, 12, 22, 32, 42, 52, 62, 72, 92, B2, D2, F2 -> NOP
	this.Opref[0x02] = NewOp6502("hlt", "implied", 0x02, 1, 2, IMPLIED, MODE_IMPLIED, NOP)
	this.Opref[0x12] = NewOp6502("hlt", "implied", 0x12, 1, 2, IMPLIED, MODE_IMPLIED, NOP)
	this.Opref[0x22] = NewOp6502("hlt", "implied", 0x22, 1, 2, IMPLIED, MODE_IMPLIED, NOP)
	this.Opref[0x32] = NewOp6502("hlt", "implied", 0x32, 1, 2, IMPLIED, MODE_IMPLIED, NOP)
	this.Opref[0x42] = NewOp6502("hlt", "implied", 0x42, 1, 2, IMPLIED, MODE_IMPLIED, NOP)
	this.Opref[0x52] = NewOp6502("hlt", "implied", 0x52, 1, 2, IMPLIED, MODE_IMPLIED, NOP)
	this.Opref[0x62] = NewOp6502("hlt", "implied", 0x62, 1, 2, IMPLIED, MODE_IMPLIED, NOP)
	this.Opref[0x72] = NewOp6502("hlt", "implied", 0x72, 1, 2, IMPLIED, MODE_IMPLIED, NOP)
	this.Opref[0x92] = NewOp6502("hlt", "implied", 0x92, 1, 2, IMPLIED, MODE_IMPLIED, NOP)
	this.Opref[0xB2] = NewOp6502("hlt", "implied", 0xB2, 1, 2, IMPLIED, MODE_IMPLIED, NOP)
	this.Opref[0xD2] = NewOp6502("hlt", "implied", 0xD2, 1, 2, IMPLIED, MODE_IMPLIED, NOP)
	this.Opref[0xF2] = NewOp6502("hlt", "implied", 0xF2, 1, 2, IMPLIED, MODE_IMPLIED, NOP)

	// DCP
	this.Opref[0xCF] = NewOp6502("dcp oper", "oper", 0xCF, 3, 6, ABSOLUTE, MODE_ABSOLUTE, DCP)
	this.Opref[0xDF] = NewOp6502("dcp oper,X", "oper,X", 0xDF, 3, 7, ABSOLUTE_X, MODE_ABSOLUTE_X, DCP)
	this.Opref[0xDB] = NewOp6502("dcp oper,Y", "oper,Y", 0xDB, 3, 7, ABSOLUTE_Y, MODE_ABSOLUTE_Y, DCP)
	this.Opref[0xC7] = NewOp6502("dcp oper", "oper", 0xC7, 2, 5, ZEROPAGE, MODE_ZEROPAGE, DCP)
	this.Opref[0xD7] = NewOp6502("dcp oper,X", "oper,X", 0xD7, 2, 6, ZEROPAGE_X, MODE_ZEROPAGE_X, DCP)
	this.Opref[0xC3] = NewOp6502("dcp (oper,X)", "(oper,X)", 0xC3, 2, 8, INDIRECT_ZP_X, MODE_INDIRECT_ZP_X, DCP)
	this.Opref[0xD3] = NewOp6502("dcp (oper),Y", "(oper),Y", 0xD3, 2, 8, INDIRECT_ZP_Y, MODE_INDIRECT_ZP_Y, DCP)

}
