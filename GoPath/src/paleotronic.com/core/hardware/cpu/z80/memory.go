package z80

func z80ReadToA2(Addr uint16) int {
	var addr int

	switch Addr / 0x1000 {
	case 0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xA:
		addr = int(Addr) + 0x1000
		return addr
	case 0xB, 0xC, 0xD:
		addr = int(Addr) + 0x2000
		return addr
	case 0xE:
		addr = int(Addr) - 0x2000
		return addr
	case 0xF:
		addr = int(Addr) - 0xF000
		return addr
	}
	return -1
}

func z80WriteToA2(Addr uint16) int {
	var addr int
	var laddr int
	laddr = int(Addr) & 0xFFF
	switch Addr & 0xF000 {
	case 0x0000:
		addr = laddr + 0x1000
		break
	case 0x1000:
		addr = laddr + 0x2000
		break
	case 0x2000:
		addr = laddr + 0x3000
		break
	case 0x3000:
		addr = laddr + 0x4000
		break
	case 0x4000:
		addr = laddr + 0x5000
		break
	case 0x5000:
		addr = laddr + 0x6000
		break
	case 0x6000:
		addr = laddr + 0x7000
		break
	case 0x7000:
		addr = laddr + 0x8000
		break
	case 0x8000:
		addr = laddr + 0x9000
		break
	case 0x9000:
		addr = laddr + 0xA000
		break
	case 0xA000:
		addr = laddr + 0xB000
		break
	case 0xB000:
		addr = laddr + 0xD000
		break
	case 0xC000:
		addr = laddr + 0xE000
		break
	case 0xD000:
		addr = laddr + 0xF000
		break
	case 0xE000:
		addr = laddr + 0xC000
		break
	case 0xF000:
		addr = laddr + 0x0000
		break
	}
	return addr
}

// ReadByte reads a byte from address taking into account
// contention.
func (z *CoreZ80) ReadByte(address uint16) byte {

	// if address >= 0xc000 && address <= 0xc07f {
	// 	log.Printf("Softswitch zone READ 0x%.4x", address)
	// }

	// if z.LinearMemory {
	// 	v := z.RAM.ReadInterpreterMemory(z.MemIndex, int(address))
	// 	return byte(v)
	// }

	return byte(z.RAM.ReadInterpreterMemory(z.MemIndex, z80ReadToA2(address)))
}

// ReadByteInternal reads a byte from address without taking
// into account contention.
func (z *CoreZ80) ReadByteInternal(address uint16) byte {
	// if address >= 0xc000 && address <= 0xc07f {
	// 	log.Printf("Softswitch zone READ 0x%.4x", address)
	// }
	// if z.LinearMemory {
	// 	v := z.RAM.ReadInterpreterMemory(z.MemIndex, int(address))
	// 	return byte(v)
	// }

	return byte(z.RAM.ReadInterpreterMemory(z.MemIndex, z80ReadToA2(address)))
}

// WriteByte writes a byte at address taking into account
// contention.
func (z *CoreZ80) WriteByte(address uint16, value byte) {
	// if address >= 0xc000 && address <= 0xc07f {
	// 	log.Printf("Softswitch zone WRITE 0x%.2x -> 0x%.4x", value, address)
	// }
	// if z.LinearMemory {
	// 	log.Printf("Write of %.2x to %.4x", value, address)
	// 	z.RAM.WriteInterpreterMemory(z.MemIndex, int(address), uint64(value))
	// 	return
	// }
	z.RAM.WriteInterpreterMemory(z.MemIndex, z80WriteToA2(address), uint64(value))
}

// WriteByteInternal writes a byte at address without taking
// into account contention.
func (z *CoreZ80) WriteByteInternal(address uint16, value byte) {
	// if address >= 0xc000 && address <= 0xc07f {
	// 	log.Printf("Softswitch zone WRITE 0x%.2x -> 0x%.4x", value, address)
	// }
	// if z.LinearMemory {
	// 	z.RAM.WriteInterpreterMemory(z.MemIndex, int(address), uint64(value))
	// 	return
	// }
	z.RAM.WriteInterpreterMemory(z.MemIndex, z80WriteToA2(address), uint64(value))
}

// Follow contention methods. Leave unimplemented if you don't
// care about memory contention.

// ContendRead increments the Tstates counter by time as a
// result of a memory read at the given address.
func (z *CoreZ80) ContendRead(address uint16, time int) {

}

func (z *CoreZ80) ContendReadNoMreq(address uint16, time int) {

}

func (z *CoreZ80) ContendReadNoMreq_loop(address uint16, time int, count uint) {

}

func (z *CoreZ80) ContendWriteNoMreq(address uint16, time int) {

}

func (z *CoreZ80) ContendWriteNoMreq_loop(address uint16, time int, count uint) {

}

func (z *CoreZ80) Read(address uint16) byte {
	return z.ReadByte(address)
}

func (z *CoreZ80) Write(address uint16, value byte, protectROM bool) {
	z.WriteByte(address, value)
}

// Data returns the memory content.
func (z *CoreZ80) Data() []byte {
	chunk := z.RAM.BlockRead(z.MemIndex, z.RAM.MEMBASE(z.MemIndex), 65536)
	out := make([]byte, len(chunk))
	for i, v := range chunk {
		out[i] = byte(v)
	}
	return out
}
