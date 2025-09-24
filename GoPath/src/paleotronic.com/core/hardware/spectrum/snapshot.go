package spectrum

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/hardware/spectrum/snapshot"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/settings"
	"paleotronic.com/files"
)

func ReadSnapshot(filename string) (*snapshot.Z80, error) {
	snapshot, err := snapshot.ReadProgram(filename)
	if err != nil {
		return nil, err
	}
	return snapshot, nil
}

func (spectrum *ZXSpectrum) LoadFile(filename string) error {
	snapshot, err := ReadSnapshot(filename)
	if err != nil {
		return err
	}
	err = spectrum.LoadSnapshot(snapshot)
	if err == nil {
		spectrum.loadedFile = filename
	}
	return err
}

var reTimeStamp = regexp.MustCompile("^(.+)(_([0-9]{14}))?.z80$")
var reNonTimeStamp = regexp.MustCompile("^(.+).z80$")

func (spectrum *ZXSpectrum) CreateSnapshot(filename string) error {

	log.Printf("CreateSnapshot called")

	// sort out filename if not supplied
	if filename == "" {
		now := time.Now().Format("20060102150405")
		f := settings.SnapshotFile[spectrum.e.GetMemIndex()]
		if f != "" {
			path := "/local/MySaves/"
			base := files.GetFilename(f)
			if reTimeStamp.MatchString(base) {
				m := reTimeStamp.FindAllStringSubmatch(base, -1)
				filename = path + "/" + m[0][1] + "_" + now + ".z80"
			} else if reNonTimeStamp.MatchString(base) {
				m := reNonTimeStamp.FindAllStringSubmatch(base, -1)
				filename = path + "/" + m[0][1] + "_" + now + ".z80"
			}
		} else {
			var model string
			switch spectrum.model {
			case model128k:
				model = "128k"
			case model48k:
				model = "48k"
			default:
				model = "unknown"
			}
			filename = "/local/MySaves/spectrum_" + model + "_" + now + ".z80"
		}
	}

	log.Printf("creating %s", filename)
	err := spectrum.SaveSnapshot(filename)
	settings.SnapshotFile[spectrum.e.GetMemIndex()] = filename

	if err == nil {
		apple2helpers.OSDShow(spectrum.e, "Z80: Created "+filename)
	}

	return err

}

func (spectrum *ZXSpectrum) LoadSnapshot(s *snapshot.Z80) error {

	if s.Is128K {
		spectrum.model = model128k
		spectrum.ports = NewZXPorts128K(spectrum.e, spectrum)
	} else {
		spectrum.model = model48k
		spectrum.ports = NewZXPorts48K(spectrum.e, spectrum)
	}

	spectrum.ResetMemory(true)

	cpu := s.CPU
	//ula := s.UlaState() -- ignoring border for now
	mem := s.Memory()

	Z80 := apple2helpers.GetZ80CPU(spectrum.e).Z80()

	// Populate registers
	Z80.A = cpu.RegA
	Z80.F = cpu.RegF
	Z80.B = cpu.RegB
	Z80.C = cpu.RegC
	Z80.D = cpu.RegD
	Z80.E = cpu.RegE
	Z80.H = cpu.RegH
	Z80.L = cpu.RegL
	Z80.A_ = cpu.RegAx
	Z80.F_ = cpu.RegFx
	Z80.B_ = cpu.RegBx
	Z80.C_ = cpu.RegCx
	Z80.D_ = cpu.RegDx
	Z80.E_ = cpu.RegEx
	Z80.H_ = cpu.RegHx
	Z80.L_ = cpu.RegLx
	Z80.IXL = byte(cpu.RegIX & 0xff)
	Z80.IXH = byte(cpu.RegIX >> 8)
	Z80.IYL = byte(cpu.RegIY & 0xff)
	Z80.IYH = byte(cpu.RegIY >> 8)

	Z80.I = cpu.RegI
	Z80.IFF1 = cpu.IFF1
	Z80.IFF2 = cpu.IFF2
	Z80.IM = cpu.ModeINT

	Z80.R = uint16(cpu.RegR & 0x7f)
	Z80.R7 = cpu.RegR & 0x80

	Z80.SetPC(cpu.RegPC)
	Z80.SetSP(cpu.RegSP)

	// Border color
	spectrum.ports.WritePortInternal(0xfe, byte(s.Border), false /*contend*/)

	// Populate memory
	if s.Is128K {
		spectrum.memory.SetMemory128K(mem[:])
		spectrum.ConfigureMemory128K(s.Port7ffd) // setup correct memory paging

		if len(s.AYRegs) == 16 {
			log.Printf("Restoring AY state from snapshot")
			p := spectrum.ports.(*ZXPorts128K)
			for reg, val := range s.AYRegs {
				p.AYSelect(0, reg)
				p.AYSelect(1, reg)
				p.AYWrite(0, int(val))
				p.AYWrite(1, int(val))
			}
		}

	} else {
		spectrum.memory.SetMemory(mem[:])
	}

	Z80.Tstates = int(cpu.Tstates)

	log.Printf("CPU is ready to resume at %.4x", Z80.PC())

	//spectrum.SaveSnapshot("/local/z80test.z80")

	return nil
}

func (s *ZXSpectrum) SaveSnapshot(filename string) error {

	var z80HeaderV3 = make([]byte, 87)
	var data = []byte{}

	z80 := apple2helpers.GetZ80CPU(s.e).Z80()

	// cpu state stuff
	z80HeaderV3[0] = z80.A
	z80HeaderV3[1] = z80.F
	z80HeaderV3[2] = z80.C
	z80HeaderV3[3] = z80.B
	z80HeaderV3[4] = z80.L
	z80HeaderV3[5] = z80.H
	// skip bytes 6/7 in v3 as they should be zero
	z80HeaderV3[8] = byte(z80.SP() & 0xff)
	z80HeaderV3[9] = byte(z80.SP() >> 8)
	z80HeaderV3[10] = z80.I
	z80HeaderV3[11] = byte(z80.R & 0x7f)
	z80HeaderV3[12] = byte(s.border << 1)
	if z80.R&0x80 != 0 {
		z80HeaderV3[12] |= 0x01
	}
	z80HeaderV3[13] = z80.E
	z80HeaderV3[14] = z80.D
	z80HeaderV3[15] = z80.C_
	z80HeaderV3[16] = z80.B_
	z80HeaderV3[17] = z80.E_
	z80HeaderV3[18] = z80.D_
	z80HeaderV3[19] = z80.L_
	z80HeaderV3[20] = z80.H_
	z80HeaderV3[21] = z80.A_
	z80HeaderV3[22] = z80.F_
	z80HeaderV3[23] = byte(z80.IY() & 0xff)
	z80HeaderV3[24] = byte(z80.IY() >> 8)
	z80HeaderV3[25] = byte(z80.IX() & 0xff)
	z80HeaderV3[26] = byte(z80.IX() >> 8)
	z80HeaderV3[27] = z80.IFF1
	z80HeaderV3[28] = z80.IFF2
	z80HeaderV3[29] = z80.IM

	// Kempston
	z80HeaderV3[29] |= 0x40

	// header length
	z80HeaderV3[30] = 55
	z80HeaderV3[31] = 0

	// pc
	z80HeaderV3[32] = byte(z80.PC() & 0xff)
	z80HeaderV3[33] = byte(z80.PC() >> 8)

	// model
	switch s.model {
	case model48k:
		z80HeaderV3[34] = 0
	case model128k:
		z80HeaderV3[34] = 4
	}

	// 128 specific registers
	if s.model != model128k {
		// page state
		z80HeaderV3[35] = s.pageState
		// ay
		z80HeaderV3[37] |= 0x04
		z80HeaderV3[38] = byte(s.ayReg)
		regAY := s.ay38910[0].State()
		for reg, val := range regAY {
			z80HeaderV3[39+reg] = byte(val)
		}
	}

	// built header, now append it
	data = append(data, z80HeaderV3...)

	mmu := s.e.GetMemoryMap().BlockMapper[s.e.GetMemIndex()]

	// now for the memory
	if s.model == model48k {

		m := s.e.GetMemoryMap()
		i := s.e.GetMemIndex()

		var raw []uint64
		var buffer []byte

		// page 5

		raw = m.BlockRead(s.e.GetMemIndex(), m.MEMBASE(i)+0x4000, 0x4000)
		buffer = make([]byte, len(raw))
		for i, v := range raw {
			buffer[i] = byte(v)
		}
		data = append(data, 0xff, 0xff)
		data = append(data, 8)
		data = append(data, buffer...)

		// page 2

		raw = m.BlockRead(s.e.GetMemIndex(), m.MEMBASE(i)+0x8000, 0x4000)
		buffer = make([]byte, len(raw))
		for i, v := range raw {
			buffer[i] = byte(v)
		}
		data = append(data, 0xff, 0xff)
		data = append(data, 4)
		data = append(data, buffer...)

		// page 0

		raw = m.BlockRead(s.e.GetMemIndex(), m.MEMBASE(i)+0xC000, 0x4000)
		buffer = make([]byte, len(raw))
		for i, v := range raw {
			buffer[i] = byte(v)
		}
		data = append(data, 0xff, 0xff)
		data = append(data, 5)
		data = append(data, buffer...)

	} else { // Mode 128k
		for page := 0; page < 8; page++ {
			buffer := compressPageZ80(mmu.Get(fmt.Sprintf("bank.%d", page)))
			if len(buffer) == 0x4000 {
				data = append(data, 0xff, 0xff)
			} else {
				data = append(
					data,
					byte(len(buffer)&0xff),
					byte((len(buffer)>>8)&0xff),
				)
			}
			data = append(data, byte(page+3))
			data = append(data, buffer...)
		}
	}

	// later we save it...
	return files.WriteBytesViaProvider(
		files.GetPath(filename),
		files.GetFilename(filename),
		data,
	)

}

func countRepeatedByte(block *memory.MemoryBlock, address int, value uint64) int {
	count := 0

	v := block.DirectRead(address)
	for address < 0x4000 && count < 254 && v == value {
		count++
		address++
		v = block.DirectRead(address)
	}

	return count
}

func compressPageZ80(block *memory.MemoryBlock) []byte {
	address := 0
	addrDst := 0
	var nReps int
	var value uint64
	buffer := make([]byte, 0x4000)

	for address < 0x4000 {
		value = block.DirectRead(address)
		address++
		nReps = countRepeatedByte(block, address, value)
		if value == 0xED {
			if nReps == 0 {
				buffer[addrDst] = 0xED
				addrDst++
				buffer[addrDst] = byte(block.DirectRead(address))
				addrDst++
				address++
			} else {
				buffer[addrDst] = 0xED
				addrDst++
				buffer[addrDst] = 0xED
				addrDst++
				buffer[addrDst] = byte(nReps + 1)
				addrDst++
				buffer[addrDst] = 0xED
				addrDst++
				address += nReps
			}
		} else {
			if nReps < 4 {
				// Si hay menos de 5 valores consecutivos iguales
				// no se comprimen.
				buffer[addrDst] = byte(value)
				addrDst++
			} else {
				buffer[addrDst] = 0xED
				addrDst++
				buffer[addrDst] = 0xED
				addrDst++
				buffer[addrDst] = byte(nReps + 1)
				addrDst++
				buffer[addrDst] = byte(value)
				addrDst++
				address += nReps
			}
		}
	}
	return buffer[:addrDst]
}
