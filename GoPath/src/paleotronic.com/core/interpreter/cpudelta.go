package interpreter

import (
	"errors"

	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/freeze"
	"paleotronic.com/utils"
)

type CPUDelta struct {
	A  *byte
	X  *byte
	Y  *byte
	PC *uint16
	SP *byte
	P  *byte
}

func (d *CPUDelta) FromBytes(data []byte) {
	//log.Printf("Req to decode CPD packet: %+v", data)
	ptr := 0
	for ptr < len(data) {
		t := CSSFlag(data[ptr])
		ptr++
		switch t {
		case CSS_CPU_PC:
			a := uint16(data[ptr+0]) + 256*uint16(data[ptr+1])
			ptr += 2
			d.PC = &a
		case CSS_CPU_A:
			a := data[ptr+0]
			ptr++
			d.A = &a
		case CSS_CPU_X:
			a := data[ptr+0]
			ptr++
			d.X = &a
		case CSS_CPU_Y:
			a := data[ptr+0]
			ptr++
			d.Y = &a
		case CSS_CPU_P:
			a := data[ptr+0]
			ptr++
			d.P = &a
		case CSS_CPU_SP:
			a := data[ptr+0]
			ptr++
			d.SP = &a
		default:
			panic(errors.New("bad id " + utils.IntToStr(int(t))))
		}
	}
}

func (d *CPUDelta) ToBytes() []byte {
	regs := []CSSFlag{CSS_CPU_PC, CSS_CPU_A, CSS_CPU_X, CSS_CPU_Y, CSS_CPU_SP, CSS_CPU_P}
	data := make([]byte, 13)
	ptr := 0
	for _, r := range regs {
		switch r {
		case CSS_CPU_PC:
			if d.PC != nil {
				data[ptr+0] = byte(CSS_CPU_PC)
				data[ptr+1] = byte(*d.PC & 0xff)
				data[ptr+2] = byte((*d.PC >> 8) & 0xff)
				ptr += 3
			}
		case CSS_CPU_A:
			if d.A != nil {
				data[ptr+0] = byte(CSS_CPU_A)
				data[ptr+1] = byte(*d.A)
				ptr += 2
			}
		case CSS_CPU_X:
			if d.X != nil {
				data[ptr+0] = byte(CSS_CPU_X)
				data[ptr+1] = byte(*d.X)
				ptr += 2
			}
		case CSS_CPU_Y:
			if d.Y != nil {
				data[ptr+0] = byte(CSS_CPU_Y)
				data[ptr+1] = byte(*d.Y)
				ptr += 2
			}
		case CSS_CPU_P:
			if d.P != nil {
				data[ptr+0] = byte(CSS_CPU_P)
				data[ptr+1] = byte(*d.P)
				ptr += 2
			}
		case CSS_CPU_SP:
			if d.SP != nil {
				data[ptr+0] = byte(CSS_CPU_SP)
				data[ptr+1] = byte(*d.SP)
				ptr += 2
			}
		}
	}
	return data[:ptr]
}

func getDelta(orig, new *freeze.CPURegs) *CPUDelta {
	d := &CPUDelta{}

	if orig.A != new.A {
		a := byte(new.A)
		d.A = &a
	}
	if orig.X != new.X {
		a := byte(new.X)
		d.X = &a
	}
	if orig.Y != new.Y {
		a := byte(new.Y)
		d.Y = &a
	}
	if orig.P != new.P {
		a := byte(new.P)
		d.P = &a
	}
	if orig.SP != new.SP {
		a := byte(new.SP & 0xff)
		d.SP = &a
	}
	if orig.PC != new.PC {
		a := uint16(new.PC & 0xffff)
		d.PC = &a
	}

	return d
}

func (r *Recorder) LogCPUDelta() {

	if r.cpu == nil {
		r.cpu = apple2helpers.GetCPU(r.Source)
	}

	if r.lastCPU == nil {
		return // need LogCPU call to have happened first
	}

	if r.cpu.RecRegisters.GlobalCycles == r.lastGlobalCycles {
		return
	}

	r.lastGlobalCycles = r.cpu.RecRegisters.GlobalCycles

	c := freeze.CPURegs{
		A:         r.cpu.RecRegisters.A,
		X:         r.cpu.RecRegisters.X,
		Y:         r.cpu.RecRegisters.Y,
		PC:        r.cpu.RecRegisters.PC,
		SP:        r.cpu.RecRegisters.SP,
		P:         r.cpu.RecRegisters.P,
		SPEED:     int(r.cpu.RecRegisters.RealSpeed),
		ScanCycle: int(r.cpu.RecRegisters.GlobalCycles % 17030),
	}

	// delta
	d := getDelta(r.lastCPU, &c)

	// update state
	r.lastCPU = &c

	r.muChan.Put(&RecorderEvent{
		ncpu: d,
	})

}
