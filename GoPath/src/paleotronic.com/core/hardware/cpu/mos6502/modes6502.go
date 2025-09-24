package mos6502

/*

  Addressing mode handlers.

  format:

  func <name>( cpu *Core6502 ) (value int, cycle_penalty int) {

  }

*/

type MODEENUM int

const (
	MODE_IMMEDIATE MODEENUM = iota
	MODE_IMPLIED
	MODE_ABSOLUTE
	MODE_ABSOLUTE_X
	MODE_ABSOLUTE_Y
	MODE_ZEROPAGE
	MODE_ZEROPAGE_X
	MODE_ZEROPAGE_Y
	MODE_INDIRECT
	MODE_INDIRECT_NMOS
	MODE_INDIRECT_X
	MODE_INDIRECT_ZP
	MODE_INDIRECT_ZP_X
	MODE_INDIRECT_ZP_Y
	MODE_RELATIVE
	MODE_RELATIVE_ZP
)

func (m MODEENUM) String() string {
	switch m {
	case MODE_IMMEDIATE:
		return "imm"
	case MODE_IMPLIED:
		return "imp"
	case MODE_ABSOLUTE:
		return "abs"
	case MODE_ABSOLUTE_X:
		return "abs,x"
	case MODE_ABSOLUTE_Y:
		return "abs,y"
	case MODE_ZEROPAGE:
		return "zp"
	case MODE_ZEROPAGE_X:
		return "zp,x"
	case MODE_ZEROPAGE_Y:
		return "zp,y"
	case MODE_INDIRECT:
		return "ind"
	case MODE_INDIRECT_NMOS:
		return "ind"
	case MODE_INDIRECT_X:
		return "ind,x"
	case MODE_INDIRECT_ZP:
		return "ind_zp"
	case MODE_INDIRECT_ZP_X:
		return "ind_zp,x"
	case MODE_INDIRECT_ZP_Y:
		return "ind_zp_y"
	case MODE_RELATIVE:
		return "rel"
	case MODE_RELATIVE_ZP:
		return "rel_zp"
	}
	return "unknown"
}

type MODE func(cpu *Core6502, addr *int) (literal bool, value int, penalty int)

var (
	IMMEDIATE = func(cpu *Core6502, addr *int) (literal bool, value int, penalty int) {
		literal = true
		value = cpu.FetchBytePC(addr)

		return literal, value, penalty
	}
	IMPLIED = func(cpu *Core6502, addr *int) (literal bool, value int, penalty int) {
		cpu.FetchBytePCNOP(addr)
		literal = true
		return literal, value, penalty
	}
	ABSOLUTE = func(cpu *Core6502, addr *int) (literal bool, value int, penalty int) {
		literal, value, penalty = false, cpu.FetchWordPC(addr), 0
		return literal, value, penalty
	}
	ABSOLUTE_X = func(cpu *Core6502, addr *int) (literal bool, value int, penalty int) {
		//format = "$oper,X"
		address := cpu.FetchWordPC(addr)
		value = (address + cpu.X) & 0xffff
		if address&0xff00 != value&0xff00 {
			penalty = 1
			cpu.ClockTick()
		}
		return literal, value, penalty
	}
	ABSOLUTE_X_WRITE = func(cpu *Core6502, addr *int) (literal bool, value int, penalty int) {
		//format = "$oper,X"
		address := cpu.FetchWordPC(addr)
		value = (address + cpu.X) & 0xffff
		return literal, value, penalty
	}
	ABSOLUTE_Y = func(cpu *Core6502, addr *int) (literal bool, value int, penalty int) {
		//format = "$oper,Y"
		address := cpu.FetchWordPC(addr)
		value = (address + cpu.Y) & 0xffff
		if address&0xff00 != value&0xff00 {
			penalty = 1
			cpu.ClockTick()
		}
		return literal, value, penalty
	}
	ABSOLUTE_Y_WRITE = func(cpu *Core6502, addr *int) (literal bool, value int, penalty int) {
		//format = "$oper,Y"
		address := cpu.FetchWordPC(addr)
		value = (address + cpu.Y) & 0xffff
		return literal, value, penalty
	}
	ZEROPAGE = func(cpu *Core6502, addr *int) (literal bool, value int, penalty int) {
		//format = "$oper"
		value = cpu.FetchBytePC(addr)

		return literal, value, penalty
	}
	ZEROPAGE_X = func(cpu *Core6502, addr *int) (literal bool, value int, penalty int) {
		//format = "$oper,X"
		value = cpu.FetchBytePC(addr)
		cpu.FetchByteAddr(value)
		value = (value + cpu.X) & 0xff

		return literal, value, penalty
	}
	ZEROPAGE_Y = func(cpu *Core6502, addr *int) (literal bool, value int, penalty int) {
		//format = "$oper,Y"
		value = (cpu.FetchBytePC(addr) + cpu.Y) & 0xff

		return literal, value, penalty
	}
	INDIRECT = func(cpu *Core6502, addr *int) (literal bool, value int, penalty int) {
		//format = "$(oper)"
		address := cpu.FetchWordPC(addr)
		value = cpu.FetchWordAddr(address)

		return literal, value, penalty
	}
	// NMOS indirect with 0x00ff bug
	INDIRECT_NMOS = func(cpu *Core6502, addr *int) (literal bool, value int, penalty int) {
		//format = "$(oper)"
		address := cpu.FetchWordPC(addr)
		value = cpu.FetchWordAddrNMOS(address)

		return literal, value, penalty
	}
	INDIRECT_X = func(cpu *Core6502, addr *int) (literal bool, value int, penalty int) {
		//format = "$(oper,X)"
		address := cpu.FetchWordPC(addr) + cpu.X
		value = cpu.FetchWordAddr(address & 0xffff)

		return literal, value, penalty
	}
	INDIRECT_ZP = func(cpu *Core6502, addr *int) (literal bool, value int, penalty int) {
		//format = "$(oper)"
		address := cpu.FetchBytePC(addr)
		value = cpu.FetchWordAddr(address)

		return literal, value, penalty
	}
	INDIRECT_ZP_X = func(cpu *Core6502, addr *int) (literal bool, value int, penalty int) {
		//format = "$(oper,X)"
		address := cpu.FetchBytePC(addr)
		cpu.FetchByteAddr(address)
		value = cpu.FetchWordAddr((address + cpu.X) & 0xff)

		return literal, value, penalty
	}
	INDIRECT_ZP_Y = func(cpu *Core6502, addr *int) (literal bool, value int, penalty int) {
		//format = "$(oper),Y"
		address := cpu.FetchBytePC(addr) & 0xff
		address = cpu.FetchWordAddr(address)
		value = address + cpu.Y
		if (address & 0xff00) != (value & 0xff00) {
			penalty = 1
			cpu.ClockTick()
		}

		return literal, value, penalty
	}
	INDIRECT_ZP_Y_WRITE = func(cpu *Core6502, addr *int) (literal bool, value int, penalty int) {
		//format = "$(oper),Y"
		address := cpu.FetchBytePC(addr) & 0xff
		address = cpu.FetchWordAddr(address)
		value = address + cpu.Y

		return literal, value, penalty
	}
	RELATIVE = func(cpu *Core6502, addr *int) (literal bool, value int, penalty int) {
		//format = "$oper"
		diff := int8(cpu.FetchBytePC(addr))
		value = cpu.PC + int(diff)

		return literal, value, penalty
	}
	ZP_RELATIVE = func(cpu *Core6502, addr *int) (newaddr int, value int, penalty int) {
		//format = "$oper"
		newaddr = cpu.FetchBytePC(addr)
		diff := int8(cpu.FetchBytePC(addr))
		value = cpu.PC + int(diff)

		return newaddr, value, penalty
	}
)
