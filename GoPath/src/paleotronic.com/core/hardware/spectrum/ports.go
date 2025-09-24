package spectrum

import (
	"paleotronic.com/z80"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
)

type ZXPorts48K struct {
	e           interfaces.Interpretable
	spectrum    *ZXSpectrum
	beeperLevel byte
	color       byte
}

func NewZXPorts48K(e interfaces.Interpretable, spec *ZXSpectrum) *ZXPorts48K {
	return &ZXPorts48K{
		e:        e,
		spectrum: spec,
	}
}

func (p *ZXPorts48K) ReadPort(address uint16) byte {
	return p.ReadPortInternal(address, true)
}

func (p *ZXPorts48K) ReadPortInternal(address uint16, contend bool) byte {
	if contend {
		p.ContendPortPreio(address)
		p.ContendPortPostio(address)
	}

	var result byte = 0xff

	if (address & 0x0001) == 0x0000 {
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

func (p *ZXPorts48K) WritePort(address uint16, b byte) {
	p.WritePortInternal(address, b, true)
}

func (p *ZXPorts48K) WritePortInternal(address uint16, b byte, contend bool) {
	if contend {
		p.ContendPortPreio(address)
	}

	//cpu := apple2helpers.GetZ80CPU(p.e).Z80()
	//log.Printf("port-write: %.2x -> %.4x (cpu at %.4x)", b, address, cpu.PC())

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

func contendPort(p *ZXPorts48K, z80 *z80.Z80, time int) {
	tstates_p := &z80.Tstates
	*tstates_p += int(p.spectrum.delay_table[*tstates_p])
	*tstates_p += time
}

func (p *ZXPorts48K) ContendPortPreio(address uint16) {
	cpu := apple2helpers.GetZ80CPU(p.e).Z80()
	if (address & 0xc000) == 0x4000 {
		contendPort(p, cpu, 1)
	} else {
		cpu.Tstates += 1
	}
}

func (p *ZXPorts48K) ContendPortPostio(address uint16) {
	cpu := apple2helpers.GetZ80CPU(p.e).Z80()
	if (address & 0x0001) == 1 {
		if (address & 0xc000) == 0x4000 {
			contendPort(p, cpu, 1)
			contendPort(p, cpu, 1)
			contendPort(p, cpu, 1)
		} else {
			cpu.Tstates += 3
		}

	} else {
		contendPort(p, cpu, 3)
	}
}
