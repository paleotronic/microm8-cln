package spectrum

import (
	"paleotronic.com/z80"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/memory"
	"paleotronic.com/octalyzer/bus"
)

type ZXMemory48K struct {
	RAM         *memory.MemoryMap
	e           interfaces.Interpretable
	spectrum    *ZXSpectrum
	MemIndex    int
	lastTstates int
	syncDone    bool
}

func NewZXMemory48K(ram *memory.MemoryMap, index int, e interfaces.Interpretable, spec *ZXSpectrum) *ZXMemory48K {
	return &ZXMemory48K{
		RAM:      ram,
		MemIndex: index,
		e:        e,
		spectrum: spec,
	}
}

func (mem *ZXMemory48K) reset() {
	var val uint64
	for i := 0; i < 0x10000; i++ {
		mem.RAM.BlockMapper[mem.MemIndex].Do(i, memory.MA_WRITE, &val)
	}
}

func (mem *ZXMemory48K) ReadByteInternal(address uint16) byte {
	var val uint64
	mem.RAM.BlockMapper[mem.MemIndex].Do(int(address), memory.MA_READ, &val)
	return byte(val)
}

func (mem *ZXMemory48K) updateTstates() {
	cpu := apple2helpers.GetZ80CPU(mem.e).Z80()
	diff := cpu.Tstates - mem.lastTstates
	if diff > 0 {
		mem.spectrum.beeper.Increment(diff)
	}
	mem.lastTstates = cpu.Tstates
}

func (mem *ZXMemory48K) SetMemory128K(data []byte) {

	// mm := mem.e.GetMemoryMap()
	// mmu := mm.BlockMapper[mem.e.GetMemIndex()]
	// var v uint64
	// for bank := 0; bank < 8; bank++ {
	// 	log.Printf("updating 128K bank %d", bank)
	// 	base := 0x4000 * bank
	// 	bank := mmu.Get(fmt.Sprintf("bank.%d", bank))
	// 	if bank == nil {
	// 		panic("missing memory bank")
	// 	}
	// 	for i := 0; i < 0x4000; i++ {
	// 		v = uint64(data[base+i])
	// 		bank.Do(0xc000+i, memory.MA_WRITE, &v)
	// 	}
	// }

	for i, v := range data {
		if i < 0x20000 {
			mem.RAM.WriteInterpreterMemorySilent(mem.MemIndex, i, uint64(v))
		}
	}
}

func (mem *ZXMemory48K) SetMemory(data []byte) {
	for i, v := range data {
		if i < 0xc000 {
			mem.RAM.WriteInterpreterMemorySilent(mem.MemIndex, 0x4000+i, uint64(v))
		}
	}
}

func (mem *ZXMemory48K) WriteByteInternal(address uint16, b byte) {
	// TODO: bind ula here
	// if (address >= SCREEN_BASE_ADDR) && (address < ATTR_BASE_ADDR) {
	// 	memory.speccy.ula.screenBitmapWrite(address, memory.data[address], b)
	// } else if (address >= ATTR_BASE_ADDR) && (address < 0x5b00) {
	// 	memory.speccy.ula.screenAttrWrite(address, memory.data[address], b)
	// }
	var val = uint64(b)

	if address >= 0x4000 {
		mem.RAM.BlockMapper[mem.MemIndex].Do(int(address), memory.MA_WRITE, &val)
	}
}

func (mem *ZXMemory48K) ReadByte(address uint16) byte {
	cpu := apple2helpers.GetZ80CPU(mem.e).Z80()
	mem.contendZXMemory48K(cpu, address, 3)
	return mem.ReadByteInternal(address)
}

func (mem *ZXMemory48K) WriteByte(address uint16, b byte) {
	cpu := apple2helpers.GetZ80CPU(mem.e).Z80()
	mem.contendZXMemory48K(cpu, address, 3)
	mem.WriteByteInternal(address, b)
}

func (mem *ZXMemory48K) checkTstates(tstates int) int {
	if tstates > mem.spectrum.config.TStatesPerFrame {

		tstates -= mem.spectrum.config.TStatesPerFrame
		mem.lastTstates -= mem.spectrum.config.TStatesPerFrame
		mem.spectrum.framecounter++
		//bus.Sync()
		cpu := apple2helpers.GetZ80CPU(mem.e)
		cpu.PullIRQLine()
		mem.syncDone = false // ready for next sync
		//log.Println("irq")
	} else if !mem.syncDone && tstates > mem.spectrum.config.VSyncTiming {
		mem.syncDone = true
		bus.Sync()
	}

	mem.spectrum.HandleServiceBusInjection(mem.spectrum.HandleServiceBusRequest)

	return tstates
}

func (mem *ZXMemory48K) contendZXMemory48K(z80 *z80.Z80, address uint16, time int) {
	//tstates_p := &z80.Tstates
	tstates := z80.Tstates //*tstates_p

	if (address & 0xc000) == 0x4000 {
		tstates += int(mem.spectrum.delay_table[tstates])
	}

	tstates += time

	tstates = mem.checkTstates(tstates)

	//*tstates_p = tstates
	z80.Tstates = tstates

	mem.updateTstates() // update beeper timing
}

// Equivalent to executing "contendZXMemory48K(z80, address, time)" count times
func (mem *ZXMemory48K) contendZXMemory48K_loop(z80 *z80.Z80, address uint16, time int, count uint) {
	//tstates_p := &z80.Tstates
	tstates := z80.Tstates //*tstates_p

	if (address & 0xc000) == 0x4000 {
		for i := uint(0); i < count; i++ {
			tstates += int(mem.spectrum.delay_table[tstates])
			tstates += time
		}
	} else {
		tstates += time * int(count)
	}

	tstates = mem.checkTstates(tstates)

	//*tstates_p = tstates
	z80.Tstates = tstates
}

func (mem *ZXMemory48K) ContendRead(address uint16, time int) {
	cpu := apple2helpers.GetZ80CPU(mem.e).Z80()
	mem.contendZXMemory48K(cpu, address, time)
}

func (mem *ZXMemory48K) ContendReadNoMreq(address uint16, time int) {
	cpu := apple2helpers.GetZ80CPU(mem.e).Z80()
	mem.contendZXMemory48K(cpu, address, time)
}

func (mem *ZXMemory48K) ContendReadNoMreq_loop(address uint16, time int, count uint) {
	cpu := apple2helpers.GetZ80CPU(mem.e).Z80()
	mem.contendZXMemory48K_loop(cpu, address, time, count)
}

func (mem *ZXMemory48K) ContendWriteNoMreq(address uint16, time int) {
	cpu := apple2helpers.GetZ80CPU(mem.e).Z80()
	mem.contendZXMemory48K(cpu, address, time)
}

func (mem *ZXMemory48K) ContendWriteNoMreq_loop(address uint16, time int, count uint) {
	cpu := apple2helpers.GetZ80CPU(mem.e).Z80()
	mem.contendZXMemory48K_loop(cpu, address, time, count)
}

func (mem *ZXMemory48K) Read(address uint16) byte {
	var val uint64
	mem.RAM.BlockMapper[mem.MemIndex].Do(int(address), memory.MA_READ, &val)
	return byte(val)
}

func (mem *ZXMemory48K) Write(address uint16, value byte, protectROM bool) {
	if address >= 0x4000 {
		var val = uint64(value)

		if address >= 0x4000 {
			mem.RAM.BlockMapper[mem.MemIndex].Do(int(address), memory.MA_WRITE, &val)
		}
	}
}

func (mem *ZXMemory48K) Data() []byte {
	chunk := mem.RAM.BlockRead(mem.MemIndex, mem.RAM.MEMBASE(mem.MemIndex), 65536)
	out := make([]byte, len(chunk))
	for i, v := range chunk {
		out[i] = byte(v)
	}
	return out
}

// Number of T-states to delay, for each possible T-state within a frame.
// The array is extended at the end - this covers the case when the emulator
// begins to execute an instruction at Tstate=(TStatesPerFrame-1). Such an
// instruction will finish at (TStatesPerFrame-1+4) or later.
// var delay_table [TStatesPerFrame + 100]byte

// // Initialize 'delay_table' at program startup
// func init() {
// 	// Note: The language automatically initialized all values
// 	//       of the 'delay_table' array to zeroes. So, we only
// 	//       have to modify the non-zero elements.

// 	tstate := FIRST_SCREEN_BYTE - 1
// 	for y := 0; y < ScreenHeight; y++ {
// 		for x := 0; x < ScreenWidth; x += 16 {
// 			tstate_x := x / PIXELS_PER_TSTATE
// 			delay_table[tstate+tstate_x+0] = 6
// 			delay_table[tstate+tstate_x+1] = 5
// 			delay_table[tstate+tstate_x+2] = 4
// 			delay_table[tstate+tstate_x+3] = 3
// 			delay_table[tstate+tstate_x+4] = 2
// 			delay_table[tstate+tstate_x+5] = 1
// 		}
// 		tstate += TSTATES_PER_LINE
// 	}
// }
