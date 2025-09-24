package z80

import (
	"fmt"
	"os"
	"strings"
	"time"

	"paleotronic.com/z80"
	"paleotronic.com/core/hardware/cpu"
	"paleotronic.com/core/hardware/servicebus"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/settings"
)

const SCHEDMS = 1

var count int64
var tickinterval int64 = 5
var msdiv int64

//var chunktime int64
var ticker *time.Ticker

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

var TRACE bool

type CoreZ80 struct {
	p                  *z80.Z80
	RAM                *memory.MemoryMap
	MemIndex           int
	Int                interfaces.Interpretable
	Halted             bool
	BaseSpeed          int64
	RealSpeed          int
	GlobalCycles       int64
	startcycles        int64
	basecount          int64
	EmuStartTime       int64
	LinearMemory       bool
	PendingMaskableIRQ bool
	chunktime          int64
	Warp               float64
	f                  *os.File
	tracenum           int
	reader             *MemReader
	PrevWarp           float64
	UserWarp           float64
	UserWarpMode       bool
}

type MemReader struct {
	RAM   *memory.MemoryMap
	Index int
}

func (mr *MemReader) ReadByte(address uint16) byte {
	return byte(mr.RAM.ReadInterpreterMemory(mr.Index, int(address)))
}

func NewCoreZ80(ent interfaces.Interpretable, ma z80.MemoryAccessor, pa z80.PortAccessor) *CoreZ80 {
	z := &CoreZ80{
		Int:          ent,
		RAM:          ent.GetMemoryMap(),
		MemIndex:     ent.GetMemIndex(),
		BaseSpeed:    3500000,
		Warp:         1,
		UserWarpMode: false,
		PrevWarp:     1,
		UserWarp:     1,
		reader:       &MemReader{RAM: ent.GetMemoryMap(), Index: ent.GetMemIndex()},
	}
	if ma != nil {
		z.p = z80.NewZ80(ma, pa)
	} else {
		z.p = z80.NewZ80(z, z)
	}
	z.p.Reset()
	z.p.Halted = false
	z.chunktime = int64(int64(float64(z.BaseSpeed)*z.Warp) / msdiv)
	return z
}

func (this *CoreZ80) CalcTiming() {
	this.chunktime = int64(int64(float64(this.BaseSpeed)*this.Warp) / msdiv)
	this.RealSpeed = int(float64(this.BaseSpeed) * this.Warp)
	servicebus.SendServiceBusMessage(
		this.MemIndex,
		servicebus.Z80SpeedChange,
		this.RealSpeed,
	)
}

func (this *CoreZ80) CheckWarp() bool {
	if this.Warp != this.PrevWarp {
		if this.UserWarpMode && !settings.CanUserChangeSpeed() {
			return false
		}

		this.PrevWarp = this.Warp
		this.CalcTiming()
		return true
	}
	return false
}

func (z *CoreZ80) Z80() *z80.Z80 {
	return z.p
}

func (z *CoreZ80) PullIRQLine() {
	z.PendingMaskableIRQ = true
}

func (z *CoreZ80) Reset() {
	z.p.Reset()
}

func (z *CoreZ80) DecodeInstruction(pc int) ([]int, string, int) {

	mn, next, _ := z80.Disassemble(z.reader, z.p.PC(), 0)
	bytes := int(next) - int(pc)
	if bytes < 0 {
		bytes += 65536
	}
	opcodes := []int{}
	for i := 0; i < bytes; i++ {
		opcodes = append(opcodes, int(z.reader.ReadByte(uint16((int(pc)+i)%65536))))
	}

	return opcodes, mn, 0

}

func (z *CoreZ80) Flags() string {
	out := ""
	if z.p.F&0x80 != 0 {
		out += "S"
	} else {
		out += "-"
	}
	if z.p.F&0x40 != 0 {
		out += "Z"
	} else {
		out += "-"
	}
	if z.p.F&0x20 != 0 {
		out += "5"
	} else {
		out += "-"
	}
	if z.p.F&0x10 != 0 {
		out += "H"
	} else {
		out += "-"
	}
	if z.p.F&0x08 != 0 {
		out += "3"
	} else {
		out += "-"
	}
	if z.p.F&0x04 != 0 {
		out += "V"
	} else {
		out += "-"
	}
	if z.p.F&0x02 != 0 {
		out += "N"
	} else {
		out += "-"
	}
	if z.p.F&0x01 != 0 {
		out += "C"
	} else {
		out += "-"
	}
	return out
}

func (z *CoreZ80) TraceEvent(cat string, msg string) {
	if !TRACE {
		return
	}
	pc := z.p.PC()
	z.f.WriteString(
		fmt.Sprintf(
			"pc=%.4x: a=%.2x bc=%.4x de=%.4x f=%s hl=%.4x ix=%.4x iy=%.4x sp=%.4x : %-9s : %-20s\n",
			pc,
			z.p.A_,
			z.p.BC(),
			z.p.DE(),
			z.Flags(),
			z.p.HL(),
			z.p.IX(),
			z.p.IY(),
			z.p.SP(),
			cat,
			msg,
		),
	)
}

func (z *CoreZ80) Disassemble() string {
	pc := z.p.PC()
	mn, next, sh := z80.Disassemble(z.reader, z.p.PC(), 0)
	bytes := int(next) - int(pc)
	if bytes < 0 {
		bytes += 65536
	}
	if sh > 0 {
		bytes += 1
	}
	bstrs := []string{}
	for i := 0; i < bytes; i++ {
		bstrs = append(bstrs, fmt.Sprintf("%.2x", z.reader.ReadByte(uint16((int(pc)+i)%65536))))
	}
	opcodes := strings.Join(bstrs, " ")

	return fmt.Sprintf(
		"pc=%.4x: a=%.2x bc=%.4x de=%.4x f=%s hl=%.4x ix=%.4x iy=%.4x sp=%.4x : %-9s : %-20s\n",
		pc,
		z.p.A_,
		z.p.BC(),
		z.p.DE(),
		z.Flags(),
		z.p.HL(),
		z.p.IX(),
		z.p.IY(),
		z.p.SP(),
		opcodes,
		mn,
	)

}

func (z *CoreZ80) FetchExecute() cpu.FEResponse {

	//fmt.Printf("Z80: PC = 0x%.4x\n", z.p.PC())
	if TRACE && z.f == nil {
		var err error
		z.tracenum++
		z.f, err = os.Create(fmt.Sprintf("z80_trace_%d.out", z.tracenum))
		if err != nil {
			z.f = nil
		}
	}

	if z.p.Halted {
		return cpu.FE_HALTED
	}

	//fmt.Printf("Z80 PC=%.4x, opcode=%.2x\n", z.p.PC(), z.p)
	ots := z.p.Tstates

	if z.PendingMaskableIRQ {
		z.PendingMaskableIRQ = false
		z.p.Interrupt()
		//	fmt.Printf("Z80 PC after irq=%.4x\n", z.p.PC())
	}

	if TRACE && z.f != nil {
		z.f.WriteString(
			z.Disassemble(),
		)
	}

	z.p.DoOpcode()

	nts := z.p.Tstates
	if nts < ots {
		nts += 69888
	}
	dts := nts - ots
	if dts > 0 {
		z.GlobalCycles += int64(dts)
	}

	if z.p.Halted {
		return cpu.FE_HALTED
	}

	return cpu.FE_OK

}

func (z *CoreZ80) ExecuteSliced() cpu.FEResponse {
	// z.Counters = z.Int.GetCycleCounter()
	// z.checkAndApplyMemLocks()

	//var currpc int
	var r cpu.FEResponse

	entryTime := time.Now()

	//log.Printf("Start slicing with Z80.PC=%.4x (TStates = %d)", z.p.PC(), z.p.Tstates)
	// defer func() {

	// }()

	for !z.p.Halted && r == cpu.FE_OK {

		// if z.RAM.IntGetCPUBreak(z.MemIndex) {
		// 	z.RAM.IntSetCPUBreak(z.MemIndex, false)
		// 	z.Break()
		// }

		if z.RAM.IntGetCPUHalt(z.MemIndex) {
			z.RAM.IntSetCPUHalt(z.MemIndex, false)
			return cpu.FE_CTRLBREAK
		}

		z.Int.PBPaste()

		oc := count
		if z.Int.WaitForWorld() {
			diff := count - oc
			z.basecount += diff
		}

		//currpc = int(z.p.PC())

		// if TRACE {

		// 	if z.ff == nil {
		// 		z.tracenum++
		// 		z.ff, _ = os.Create(fmt.Sprintf("mos6502_%d.trace", z.tracenum))
		// 	}

		// 	txt, _ := z.DecodeTrace(currpc)
		// 	z.ff.WriteString(txt + "\r\n")
		// 	//fmt.Println(txt)
		// }

		//now = time.Now()
		os := z.Int.GetState()
		r = z.FetchExecute()
		if z.Int.GetState() != os {
			return cpu.FE_OK
		}

		// if the cpu is halted, we need to spin our wheels until we accumulate enough t-states
		if z.p.Halted {
			//log.Printf("handling halt spin state...")
			var ots, nts, dts int
			for z.p.Halted && !z.PendingMaskableIRQ {
				ots = z.p.Tstates
				z.p.GetMemoryAccessor().ContendRead(z.p.PC(), 4)
				z.p.R = (z.p.R + 1) & 0x7f
				nts = z.p.Tstates
				dts = nts - ots
				if dts > 0 {
					z.GlobalCycles += int64(dts)
				}
			}
			z.p.Interrupt()
		}

		// We only sleep here if we are running 'realtime' otherwise we just belt through
		tickdiff := count - z.basecount // # 5ms clock ticks since we started
		if tickdiff > 0 && (z.GlobalCycles-z.startcycles)/tickdiff > z.chunktime {
			return cpu.FE_SLEEP
		}

		// Check if warp changed
		if z.CheckWarp() {
			z.ResetSliced()
			return cpu.FE_SLEEP
		}

		// if we've been here too long, perhaps nitro CPU mode and we need to service the loops
		if time.Since(entryTime) > 16*time.Millisecond {
			//bus.Sync()
			z.ResetSliced()
			//log.Printf("End slicing with Z80.PC=%.4x (TStates = %d)", z.p.PC(), z.p.Tstates)
			return cpu.FE_OK
		}

	}

	return r
}

func (z *CoreZ80) ResetSliced() {
	// TODO: get slicing working
	z.basecount = count
	z.startcycles = z.GlobalCycles
}

func (z *CoreZ80) Init() {
	// TODO
}

func (this *CoreZ80) SetWarp(w float64) {
	this.Warp = w
	//this.CalcTiming()
}

func (this *CoreZ80) HasUserWarp() (bool, float64) {
	return this.UserWarpMode, this.UserWarp
}

func (this *CoreZ80) SetWarpUser(w float64) {

	settings.MuteCPU = (w >= 8)

	this.SetWarp(w)
	this.UserWarp = w
	//this.PrevWarp = w
	this.UserWarpMode = (w != 1)
	//this.CalcTiming()
}

func (this *CoreZ80) GetWarp() float64 {
	return this.Warp
}

func (this *CoreZ80) ActualSpeed() int64 {
	return int64(float64(this.BaseSpeed) * this.Warp)
}
