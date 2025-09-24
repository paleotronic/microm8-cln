// +build !remint

package apple2helpers

import (
	"paleotronic.com/core/hardware/cpu"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/settings"
)

func Exec6502Code(ent interfaces.Interpretable, a int, x int, y int, pc int, sr int, sp int, mlmode bool) cpu.FEResponse {
	//	if cpu6502[ent.GetMemIndex()] == nil {
	//		cpu6502[ent.GetMemIndex()] = mos6502.NewCore6502(ent, a, x, y, pc, sr, sp, ent)
	//	}

	CPU := GetCPU(ent)

	CPU.ROM = DoCall
	CPU.PC = pc
	CPU.P = sr
	CPU.SP = sp
	CPU.InitialSP = sp
	CPU.A = a
	CPU.X = x
	CPU.Y = y
	CPU.Halted = false
	CPU.BasicMode = !mlmode
	CPU.MemIndex = ent.GetMemIndex()
	memory.Safe = false

	io, ok := ent.GetMemoryMap().InterpreterMappableAtAddress(ent.GetMemIndex(), 0xc000)
	if ok {
		if a2io, ok := io.(interfaces.Countable); ok {
			CPU.IO = a2io
		}
	}

	if ent.GetMemoryMap().IntGetPDState(ent.GetMemIndex())&128 != 0 {
		CPU.UseProDOS = true
	}

	if !settings.PureBoot(CPU.MemIndex) {
		CPU.RegisterCallShim(0xbd00, RWTSInvoker)
		CPU.RegisterCallShim(0x3e3, RWTSLocateParams)
	}

	ent.GetMemoryMap().IntSetSpeakerMode(ent.GetMemIndex(), 1)

	r := CPU.ExecTillHalted()
	memory.Safe = true

	ent.GetMemoryMap().IntSetSpeakerMode(ent.GetMemIndex(), 0)

	// now make sure earth is safe
	if r == cpu.FE_CTRLBREAK {
		TEXT40(ent)
	}
	//ent.Start6502(pc, x, y, a, sr, sp)

	return r
}
