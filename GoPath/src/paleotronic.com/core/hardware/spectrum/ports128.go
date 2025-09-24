package spectrum

import (
	"fmt"

	"paleotronic.com/z80"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"

	corez80 "paleotronic.com/core/hardware/cpu/z80"
)

type ZXPorts128K struct {
	e           interfaces.Interpretable
	spectrum    *ZXSpectrum
	beeperLevel byte
	color       byte
}

func NewZXPorts128K(e interfaces.Interpretable, spec *ZXSpectrum) *ZXPorts128K {
	return &ZXPorts128K{
		e:        e,
		spectrum: spec,
	}
}

func (p *ZXPorts128K) ReadPort(address uint16) byte {
	return p.ReadPortInternal(address, true)
}

func (p *ZXPorts128K) ReadPortInternal(address uint16, contend bool) byte {
	if contend {
		p.ContendPortPreio(address)
		p.ContendPortPostio(address)
	}

	var result byte = 0xff

	if address == 0xfffd {
		result = byte(p.AYRead(0))
		//result = byte(p.spectrum.ay38910[0].ReadRegister(common.AYRegister(p.spectrum.ayReg)))
		//byte(p.spectrum.ayReg)
	} else if address == 0x7ffd {

		// memory map
		result = p.spectrum.pageState

	} else if (address & 0x0001) == 0x0000 {
		//log.Printf("might be reading keyboard")
		// Read keyboard
		var row uint
		for row = 0; row < 8; row++ {
			if (address & (1 << (uint16(row) + 8))) == 0 { // bit held low, so scan this row
				result &= p.spectrum.keyboard.GetKeyState(row)
			}
		}

		// // Read tape
		// if p.speccy.readFromTape && (address == 0x7ffe) {
		// 	p.tapeReadCount++
		// 	earBit := p.speccy.tapeDrive.getEarBit()
		// 	result &= earBit
		// }
		p.spectrum.keyboard.WaitingForRead = false
	} else if (address & 0x00e0) == 0x0000 {
		result &= byte(p.spectrum.joystickState)
	} else {
		// Unassigned port
		result = 0xff
	}

	return result
}

func (p *ZXPorts128K) WritePort(address uint16, b byte) {
	p.WritePortInternal(address, b, true)
}

func (p *ZXPorts128K) AYSelect(chip int, reg int) {
	p.spectrum.ay38910[chip].SetBus(reg)
	p.spectrum.ay38910[chip].SetControl(7) // select reg
}

func (p *ZXPorts128K) AYWrite(chip int, val int) {
	p.spectrum.ay38910[chip].SetBus(val)
	p.spectrum.ay38910[chip].SetControl(6) // write reg
}

func (p *ZXPorts128K) AYRead(chip int) int {
	p.spectrum.ay38910[chip].SetControl(5) // read reg
	return p.spectrum.ay38910[chip].GetBus()
}

func (p *ZXPorts128K) WritePortInternal(address uint16, b byte, contend bool) {
	if contend {
		p.ContendPortPreio(address)
	}

	cpu := apple2helpers.GetZ80CPU(p.e)
	if corez80.TRACE {
		cpu.TraceEvent(
			"WPORT",
			fmt.Sprintf("write %.2x to port %.4x", b, address),
		)
	}
	// fmt.Printf("PORT WRITE: %.2x -> %.4x (cpu at %.4x)", b, address, cpu.PC())

	// AY handling
	if (address & 0x8002) == 0x8000 {
		if (address & 0x4000) != 0 {
			p.spectrum.ayReg = int(b)
			p.AYSelect(0, int(b&15))
			p.AYSelect(1, int(b&15))
		} else {
			p.AYWrite(0, int(b))
			p.AYWrite(1, int(b))
			// p.spectrum.ay38910[0].WriteReg(common.AYRegister(p.spectrum.ayReg), int(b))
			// p.spectrum.ay38910[1].WriteReg(common.AYRegister(p.spectrum.ayReg), int(b))
		}
	}

	// Paging
	if /*address == 0x7ffd*/ (address & 0x8002) == 0 {

		// memory map
		p.spectrum.ConfigureMemory128K(b)

	}

	// Screen border and beeper
	if (address & 0x0001) == 0 {
		color := (b & 0x07)

		// Modify the border only if it really changed
		if p.color != color {
			//log.Printf("border color changed from %d -> %d", p.color, color)
			p.color = color
			p.spectrum.border = int(color)
		}

		// EAR(bit 4) and MIC(bit 3) output
		newBeeperLevel := (b & 0x18) >> 3
		// if p.speccy.readFromTape && !p.speccy.tapeDrive.AcceleratedLoad {
		// 	if p.speccy.tapeDrive.earBit == 0xff {
		// 		newBeeperLevel |= 2
		// 	} else {
		// 		newBeeperLevel &^= 2
		// 	}
		// }
		if p.beeperLevel != newBeeperLevel {
			//log.Printf("beeper level changed from %d -> %d", p.beeperLevel, newBeeperLevel)
			p.spectrum.beeper.ToggleSpeaker(true)
			p.beeperLevel = newBeeperLevel
		}
		// 	p.beeperLevel = newBeeperLevel

		// 	last := len(p.beeperEvents) - 1
		// 	if p.beeperEvents[last].TState == p.speccy.Cpu.Tstates {
		// 		p.beeperEvents[last].Level = newBeeperLevel
		// 	} else {
		// 		p.beeperEvents = append(p.beeperEvents, BeeperEvent{p.speccy.Cpu.Tstates, newBeeperLevel})
		// 	}
		// }
	}

	if contend {
		p.ContendPortPostio(address)
	}
}

func contendPort128(p *ZXPorts128K, z80 *z80.Z80, time int) {
	tstates_p := &z80.Tstates
	*tstates_p += int(p.spectrum.delay_table[*tstates_p])
	*tstates_p += time
}

func (p *ZXPorts128K) ContendPortPreio(address uint16) {
	cpu := apple2helpers.GetZ80CPU(p.e).Z80()
	if (address & 0xc000) == 0x4000 {
		contendPort128(p, cpu, 1)
	} else {
		cpu.Tstates += 1
	}
}

func (p *ZXPorts128K) ContendPortPostio(address uint16) {
	cpu := apple2helpers.GetZ80CPU(p.e).Z80()
	if (address & 0x0001) == 1 {
		if (address & 0xc000) == 0x4000 {
			contendPort128(p, cpu, 1)
			contendPort128(p, cpu, 1)
			contendPort128(p, cpu, 1)
		} else {
			cpu.Tstates += 3
		}

	} else {
		contendPort128(p, cpu, 3)
	}
}
