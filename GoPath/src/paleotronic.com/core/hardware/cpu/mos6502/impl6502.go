package mos6502

//	"paleotronic.com/fmt"

/*

Operation implementations...

*/

import (
	fmt2 "fmt"
	"os"

	"paleotronic.com/core/settings"
	"paleotronic.com/fmt"
)

type OP func(cpu *Core6502, oper *Op6502) int

// LDA Impl
func LDA(cpu *Core6502, oper *Op6502) int {

	imm, value, penalty := oper.Fetch(cpu, &cpu.PC)

	if imm {
		cpu.A = value & 0xff
	} else {
		cpu.A = 0xff & cpu.FetchByteAddr(value)
	}

	cpu.Set_nz(cpu.A)

	return oper.Cycles + penalty
}

func STA(cpu *Core6502, oper *Op6502) int {

	_, value, penalty := oper.Fetch(cpu, &cpu.PC)

	switch oper.FetchMode {
	case MODE_IMPLIED, MODE_ABSOLUTE, MODE_ZEROPAGE, MODE_ZEROPAGE_X, MODE_INDIRECT_ZP_X, MODE_INDIRECT_ZP:
	default:
		cpu.FetchByteAddr(value)
	}
	cpu.StoreByteAddr(value, cpu.A)

	return oper.Cycles + penalty

}

func STZ(cpu *Core6502, oper *Op6502) int {

	_, value, penalty := oper.Fetch(cpu, &cpu.PC)

	switch oper.FetchMode {
	case MODE_IMPLIED, MODE_ABSOLUTE, MODE_ZEROPAGE, MODE_ZEROPAGE_X, MODE_ABSOLUTE_X:
	default:
		cpu.FetchByteAddr(value)
	}
	cpu.StoreByteAddr(value, 0)

	return oper.Cycles + penalty

}

func LDX(cpu *Core6502, oper *Op6502) int {

	imm, value, penalty := oper.Fetch(cpu, &cpu.PC)

	if imm {
		cpu.X = value & 0xff
	} else {
		cpu.X = 0xff & cpu.FetchByteAddr(value)
	}

	cpu.Set_nz(cpu.X)

	return oper.Cycles + penalty
}

func STX(cpu *Core6502, oper *Op6502) int {

	_, value, penalty := oper.Fetch(cpu, &cpu.PC)

	switch oper.FetchMode {
	case MODE_IMPLIED, MODE_ABSOLUTE, MODE_ZEROPAGE, MODE_ZEROPAGE_X:
	default:
		cpu.FetchByteAddr(value)
	}
	cpu.StoreByteAddr(value, cpu.X)

	return oper.Cycles + penalty
}

func LDY(cpu *Core6502, oper *Op6502) int {

	imm, value, penalty := oper.Fetch(cpu, &cpu.PC)

	if imm {
		cpu.Y = value & 0xff
	} else {
		switch oper.FetchMode {
		case MODE_ZEROPAGE_Y, MODE_INDIRECT_ZP_X:
			cpu.FetchByteAddr(value)
		}
		cpu.Y = 0xff & cpu.FetchByteAddr(value)
	}

	cpu.Set_nz(cpu.Y)

	return oper.Cycles + penalty
}

func STY(cpu *Core6502, oper *Op6502) int {

	_, value, penalty := oper.Fetch(cpu, &cpu.PC)

	switch oper.FetchMode {
	case MODE_IMPLIED, MODE_ABSOLUTE, MODE_ZEROPAGE, MODE_ZEROPAGE_X:
	default:
		cpu.FetchByteAddr(value)
	}
	cpu.StoreByteAddr(value, cpu.Y)

	return oper.Cycles + penalty
}

func PHP(cpu *Core6502, oper *Op6502) int {

	_, _, penalty := oper.Fetch(cpu, &cpu.PC)

	cpu.Push(cpu.P | F_B | F_R)

	return oper.Cycles + penalty
}

func PLP(cpu *Core6502, oper *Op6502) int {

	_, _, penalty := oper.Fetch(cpu, &cpu.PC)

	cpu.FetchByteAddr(0x100 + cpu.SP)
	cpu.P = cpu.Pop() & (F_Z | F_V | F_C | F_I | F_N | F_D)

	return oper.Cycles + penalty
}

func PHA(cpu *Core6502, oper *Op6502) int {

	_, _, penalty := oper.Fetch(cpu, &cpu.PC)

	//cpu.FetchByteAddr(0x100 + cpu.SP)
	cpu.Push(cpu.A)

	return oper.Cycles + penalty
}

func PLA(cpu *Core6502, oper *Op6502) int {

	_, _, penalty := oper.Fetch(cpu, &cpu.PC)

	cpu.FetchByteAddr(0x100 + cpu.SP)
	cpu.A = cpu.Pop()
	cpu.Set_nz(cpu.A)

	return oper.Cycles + penalty
}

func PHX(cpu *Core6502, oper *Op6502) int {

	_, _, penalty := oper.Fetch(cpu, &cpu.PC)

	//cpu.FetchByteAddr(0x100 + cpu.SP)
	cpu.Push(cpu.X)

	return oper.Cycles + penalty
}

func PLX(cpu *Core6502, oper *Op6502) int {

	_, _, penalty := oper.Fetch(cpu, &cpu.PC)

	cpu.FetchByteAddr(0x100 + cpu.SP)
	cpu.X = cpu.Pop()
	cpu.Set_nz(cpu.X)

	return oper.Cycles + penalty
}

func PHY(cpu *Core6502, oper *Op6502) int {

	_, _, penalty := oper.Fetch(cpu, &cpu.PC)

	//cpu.FetchByteAddr(0x100 + cpu.SP)
	cpu.Push(cpu.Y)

	return oper.Cycles + penalty
}

func PLY(cpu *Core6502, oper *Op6502) int {

	_, _, penalty := oper.Fetch(cpu, &cpu.PC)

	cpu.FetchByteAddr(0x100 + cpu.SP)
	cpu.Y = cpu.Pop()
	cpu.Set_nz(cpu.Y)

	return oper.Cycles + penalty
}

func CMP(cpu *Core6502, oper *Op6502) int {

	imm, addr, penalty := oper.Fetch(cpu, &cpu.PC)

	value := addr
	if !imm {
		value = cpu.FetchByteAddr(addr)
	}

	val := cpu.A - value

	cpu.SetFlag(F_C, (val >= 0))
	cpu.Set_nz(val)

	return oper.Cycles + penalty
}

func CPX(cpu *Core6502, oper *Op6502) int {

	imm, addr, penalty := oper.Fetch(cpu, &cpu.PC)

	value := addr
	if !imm {
		value = cpu.FetchByteAddr(addr)
	}

	val := cpu.X - value

	cpu.SetFlag(F_C, (val >= 0))
	cpu.Set_nz(val)

	return oper.Cycles + penalty
}

func CPY(cpu *Core6502, oper *Op6502) int {

	imm, addr, penalty := oper.Fetch(cpu, &cpu.PC)

	value := addr
	if !imm {
		value = cpu.FetchByteAddr(addr)
	}

	val := cpu.Y - value

	cpu.SetFlag(F_C, (val >= 0))
	cpu.Set_nz(val)

	return oper.Cycles + penalty
}

func SED(cpu *Core6502, oper *Op6502) int {

	_, _, penalty := oper.Fetch(cpu, &cpu.PC)

	cpu.SetFlag(F_D, true)

	return oper.Cycles + penalty
}

func CLD(cpu *Core6502, oper *Op6502) int {

	_, _, penalty := oper.Fetch(cpu, &cpu.PC)

	cpu.SetFlag(F_D, false)

	return oper.Cycles + penalty
}

func SEI(cpu *Core6502, oper *Op6502) int {

	_, _, penalty := oper.Fetch(cpu, &cpu.PC)

	cpu.SetFlag(F_I, true)

	if settings.Debug6522 {
		fmt2.Printf("6522TRACE: SEI instruction\n")
	}

	return oper.Cycles + penalty
}

func CLI(cpu *Core6502, oper *Op6502) int {

	_, _, penalty := oper.Fetch(cpu, &cpu.PC)

	cpu.SetFlag(F_I, false)

	if settings.Debug6522 {
		fmt2.Printf("6522TRACE: CLI instruction\n")
	}

	return oper.Cycles + penalty
}

func SEC(cpu *Core6502, oper *Op6502) int {

	_, _, penalty := oper.Fetch(cpu, &cpu.PC)

	cpu.SetFlag(F_C, true)

	return oper.Cycles + penalty
}

func CLC(cpu *Core6502, oper *Op6502) int {

	_, _, penalty := oper.Fetch(cpu, &cpu.PC)

	cpu.SetFlag(F_C, false)

	return oper.Cycles + penalty
}

func CLV(cpu *Core6502, oper *Op6502) int {

	_, _, penalty := oper.Fetch(cpu, &cpu.PC)

	cpu.SetFlag(F_V, false)

	return oper.Cycles + penalty
}

func AND(cpu *Core6502, oper *Op6502) int {

	imm, addr, penalty := oper.Fetch(cpu, &cpu.PC)

	value := addr
	if !imm {
		value = cpu.FetchByteAddr(addr)
	}

	cpu.A &= value
	cpu.Set_nz(cpu.A)

	return oper.Cycles + penalty
}

func ORA(cpu *Core6502, oper *Op6502) int {

	imm, addr, penalty := oper.Fetch(cpu, &cpu.PC)

	value := addr
	if !imm {
		value = cpu.FetchByteAddr(addr)
	}

	cpu.A = 0xff & (cpu.A | value)
	cpu.Set_nz(cpu.A)

	return oper.Cycles + penalty
}

func EOR(cpu *Core6502, oper *Op6502) int {

	imm, addr, penalty := oper.Fetch(cpu, &cpu.PC)

	value := addr
	if !imm {
		value = cpu.FetchByteAddr(addr)
	}

	cpu.A = 0xff & (cpu.A ^ value)
	cpu.Set_nz(cpu.A)

	return oper.Cycles + penalty
}

func ASL(cpu *Core6502, oper *Op6502) int {

	imm, addr, penalty := oper.Fetch(cpu, &cpu.PC)

	value := cpu.A
	if !imm {
		if oper.FetchMode == MODE_ABSOLUTE_X {
			cpu.FetchByteAddr(addr)
		}
		value = cpu.FetchByteAddr(addr)
		cpu.StoreByteAddr(addr, value) // IMPLIED WRITE-OP
	}

	cpu.SetFlag(F_C, (value&0x80 != 0))
	value = 0x0FE & (value << 1)
	cpu.Set_nz(value)

	if !imm {
		cpu.StoreByteAddr(addr, value)
	} else {
		cpu.A = value
	}

	return oper.Cycles + penalty
}

func LSR(cpu *Core6502, oper *Op6502) int {

	imm, addr, penalty := oper.Fetch(cpu, &cpu.PC)

	value := cpu.A
	if !imm {
		if oper.FetchMode == MODE_ABSOLUTE_X {
			cpu.FetchByteAddr(addr)
		}
		value = cpu.FetchByteAddr(addr)
		cpu.StoreByteAddr(addr, value) // IMPLIED WRITE-OP
	}

	cpu.SetFlag(F_C, (value&0x01 != 0))
	value = 0x07F & (value >> 1)
	cpu.Set_nz(value)

	if !imm {
		cpu.StoreByteAddr(addr, value)
	} else {
		cpu.A = value
	}

	return oper.Cycles + penalty
}

func ROL(cpu *Core6502, oper *Op6502) int {

	imm, addr, penalty := oper.Fetch(cpu, &cpu.PC)

	value := cpu.A
	if !imm {
		if oper.FetchMode == MODE_ABSOLUTE_X {
			cpu.FetchByteAddr(addr)
		}
		value = cpu.FetchByteAddr(addr)
		cpu.StoreByteAddr(addr, value) // IMPLIED WRITE-OP
	}

	oldC := ab(cpu.TestFlag(F_C), 1, 0)

	cpu.SetFlag(F_C, (value&0x80 != 0))
	value = 0x0ff & ((value << 1) | oldC)
	cpu.Set_nz(value)

	if !imm {
		cpu.StoreByteAddr(addr, value)
	} else {
		cpu.A = value
	}

	return oper.Cycles + penalty
}

func ROR(cpu *Core6502, oper *Op6502) int {

	imm, addr, penalty := oper.Fetch(cpu, &cpu.PC)

	value := cpu.A
	if !imm {
		if oper.FetchMode == MODE_ABSOLUTE_X {
			cpu.FetchByteAddr(addr)
		}
		value = cpu.FetchByteAddr(addr)
		cpu.StoreByteAddr(addr, value) // IMPLIED WRITE-OP
	}

	oldC := ab(cpu.TestFlag(F_C), 0x80, 0)

	cpu.SetFlag(F_C, (value&0x1 != 0))
	value = 0x0ff & ((value >> 1) | oldC)
	cpu.Set_nz(value)

	if !imm {
		cpu.StoreByteAddr(addr, value)
	} else {
		cpu.A = value
	}

	return oper.Cycles + penalty
}

func ADC(cpu *Core6502, oper *Op6502) int {

	imm, addr, penalty := oper.Fetch(cpu, &cpu.PC)

	value := addr
	if !imm {
		value = cpu.FetchByteAddr(addr)
	}

	// impl begin
	var w int
	cpu.SetFlag(F_V, ((cpu.A^value)&0x080) == 0)
	if cpu.TestFlag(F_D) {
		// Decimal Mode
		w = (cpu.A & 0x0f) + (value & 0x0f) + ab(cpu.TestFlag(F_C), 1, 0)
		if w >= 10 {
			w = 0x010 | ((w + 6) & 0x0f)
		}
		w += (cpu.A & 0x0f0) + (value & 0x00f0)
		if w >= 0x0A0 {
			cpu.SetFlag(F_C, true)
			if w >= 0x0180 {
				cpu.SetFlag(F_V, false)
			}
			w += 0x060
		} else {
			cpu.SetFlag(F_C, false)
			if w < 0x080 {
				cpu.SetFlag(F_V, false)
			}
		}
	} else {
		// Binary Mode
		w = cpu.A + value + ab(cpu.TestFlag(F_C), 1, 0)
		if w >= 0x0100 {
			cpu.SetFlag(F_C, true)
			if w >= 0x0180 {
				cpu.SetFlag(F_V, false)
			}
		} else {
			cpu.SetFlag(F_C, false)
			if w < 0x080 {
				cpu.SetFlag(F_V, false)
			}
		}
	}
	cpu.A = w & 0x0ff
	cpu.Set_nz(cpu.A)
	// impl end

	return oper.Cycles + penalty
}

func SBC(cpu *Core6502, oper *Op6502) int {

	imm, addr, penalty := oper.Fetch(cpu, &cpu.PC)

	value := addr
	if !imm {
		value = cpu.FetchByteAddr(addr)
	}

	// impl begin
	cpu.SetFlag(F_V, ((cpu.A^value)&0x080) != 0)
	var w int
	if cpu.TestFlag(F_D) {
		temp := 0x0f + (cpu.A & 0x0f) - (value & 0x0f) + ab(cpu.TestFlag(F_C), 1, 0)
		if temp < 0x10 {
			w = 0
			temp -= 6
		} else {
			w = 0x10
			temp -= 0x10
		}
		w += 0x00f0 + (cpu.A & 0x00f0) - (value & 0x00f0)
		if w < 0x100 {
			cpu.SetFlag(F_C, false)
			if w < 0x080 {
				cpu.SetFlag(F_V, false)
			}
			w -= 0x60
		} else {
			cpu.SetFlag(F_C, true)
			if w >= 0x180 {
				cpu.SetFlag(F_V, false)
			}
		}
		w += temp
	} else {
		w = 0x0ff + cpu.A - value + ab(cpu.TestFlag(F_C), 1, 0)
		if w < 0x100 {
			cpu.SetFlag(F_C, false)
			if w < 0x080 {
				cpu.SetFlag(F_V, false)
			}
		} else {
			cpu.SetFlag(F_C, true)
			if w >= 0x180 {
				cpu.SetFlag(F_V, false)
			}
		}
	}
	cpu.A = w & 0x0ff
	cpu.Set_nz(cpu.A)
	// impl end

	return oper.Cycles + penalty
}

func BCC(cpu *Core6502, oper *Op6502) int {

	_, addr, penalty := oper.Fetch(cpu, &cpu.PC)

	if !cpu.TestFlag(F_C) {
		cpu.PC = addr
		penalty += 1
		cpu.ClockTick()
	}

	return oper.Cycles + penalty
}

func BCS(cpu *Core6502, oper *Op6502) int {

	_, addr, penalty := oper.Fetch(cpu, &cpu.PC)

	if cpu.TestFlag(F_C) {
		cpu.PC = addr
		penalty += 1
		cpu.ClockTick()
	}

	return oper.Cycles + penalty
}

func BNE(cpu *Core6502, oper *Op6502) int {

	_, addr, penalty := oper.Fetch(cpu, &cpu.PC)

	if !cpu.TestFlag(F_Z) {
		cpu.PC = addr
		penalty += 1
		cpu.ClockTick()
	}

	return oper.Cycles + penalty
}

func BEQ(cpu *Core6502, oper *Op6502) int {

	_, addr, penalty := oper.Fetch(cpu, &cpu.PC)

	if cpu.TestFlag(F_Z) {
		cpu.PC = addr
		penalty += 1
		cpu.ClockTick()

	}

	return oper.Cycles + penalty
}

func BRA(cpu *Core6502, oper *Op6502) int {

	_, addr, penalty := oper.Fetch(cpu, &cpu.PC)

	cpu.PC = addr

	return oper.Cycles + penalty
}

func BPL(cpu *Core6502, oper *Op6502) int {

	_, addr, penalty := oper.Fetch(cpu, &cpu.PC)

	if !cpu.TestFlag(F_N) {
		cpu.PC = addr
		penalty += 1
		cpu.ClockTick()
	}

	return oper.Cycles + penalty
}

func BMI(cpu *Core6502, oper *Op6502) int {

	_, addr, penalty := oper.Fetch(cpu, &cpu.PC)

	if cpu.TestFlag(F_N) {
		cpu.PC = addr
		penalty += 1
		cpu.ClockTick()
	}

	return oper.Cycles + penalty
}

func BVC(cpu *Core6502, oper *Op6502) int {

	_, addr, penalty := oper.Fetch(cpu, &cpu.PC)

	if !cpu.TestFlag(F_V) {
		cpu.PC = addr
		penalty += 1
		cpu.ClockTick()
	}

	return oper.Cycles + penalty
}

func BVS(cpu *Core6502, oper *Op6502) int {

	_, addr, penalty := oper.Fetch(cpu, &cpu.PC)

	if cpu.TestFlag(F_V) {
		cpu.PC = addr
		penalty += 1
		cpu.ClockTick()
	}

	return oper.Cycles + penalty
}

func BIT(cpu *Core6502, oper *Op6502) int {

	imm, addr, penalty := oper.Fetch(cpu, &cpu.PC)

	value := addr
	if !imm {
		value = cpu.FetchByteAddr(addr)
	}

	result := (cpu.A & value)
	cpu.SetFlag(F_Z, result == 0)
	cpu.SetFlag(F_N, (value&0x080) != 0)
	// As per http://www.6502.org/tutorials/vflag.html
	if !imm {
		cpu.SetFlag(F_V, (value&0x040) != 0)
	}

	return oper.Cycles + penalty
}

func BRK(cpu *Core6502, oper *Op6502) int {

	_, _, penalty := oper.Fetch(cpu, &cpu.PC)

	// <BRK> <padbyte> <NEXT>
	//       ^PC here

	if settings.PureBoot(cpu.Int.GetMemIndex()) {

		readdr := cpu.PC + 1

		cpu.Push(readdr / 256)
		cpu.Push(readdr % 256)

		p := cpu.P | F_B // ensure "B" is set on P pushed to stack

		cpu.Push(p)

		target := cpu.FetchByteAddr(0xfffe) + 256*cpu.FetchByteAddr(0xffff)

		cpu.PC = target

	} else {
		cpu.Halted = true
	}

	return oper.Cycles + penalty
}

func NOP(cpu *Core6502, oper *Op6502) int {

	_, _, penalty := oper.Fetch(cpu, &cpu.PC)

	return oper.Cycles + penalty
}

func DEC(cpu *Core6502, oper *Op6502) int {

	_, addr, penalty := oper.Fetch(cpu, &cpu.PC)

	if oper.FetchMode == MODE_ABSOLUTE_X {
		cpu.FetchByteAddr(addr)
	}
	value := cpu.FetchByteAddr(addr)
	cpu.StoreByteAddr(addr, value)

	value = 0xff & (value - 1)
	cpu.Set_nz(value)

	cpu.StoreByteAddr(addr, value)

	return oper.Cycles + penalty
}

func INC(cpu *Core6502, oper *Op6502) int {

	// fetch op    1
	// fetch addr  2
	// fetch byte  1
	// store byte  1
	// ..
	// store byte  1

	_, addr, penalty := oper.Fetch(cpu, &cpu.PC)

	if oper.FetchMode == MODE_ABSOLUTE_X {
		cpu.FetchByteAddr(addr)
	}
	value := cpu.FetchByteAddr(addr)
	cpu.StoreByteAddr(addr, value)

	value = 0xff & (value + 1)
	cpu.Set_nz(value)

	cpu.StoreByteAddr(addr, value)

	return oper.Cycles + penalty
}

func DEA(cpu *Core6502, oper *Op6502) int {

	_, _, penalty := oper.Fetch(cpu, &cpu.PC)

	cpu.A = 0xff & (cpu.A - 1)
	cpu.Set_nz(cpu.A)

	return oper.Cycles + penalty
}

func DEX(cpu *Core6502, oper *Op6502) int {

	_, _, penalty := oper.Fetch(cpu, &cpu.PC)

	cpu.X = 0xff & (cpu.X - 1)
	cpu.Set_nz(cpu.X)

	return oper.Cycles + penalty
}

func INA(cpu *Core6502, oper *Op6502) int {

	_, _, penalty := oper.Fetch(cpu, &cpu.PC)

	cpu.A = 0xff & (cpu.A + 1)
	cpu.Set_nz(cpu.A)

	return oper.Cycles + penalty
}

func INX(cpu *Core6502, oper *Op6502) int {

	_, _, penalty := oper.Fetch(cpu, &cpu.PC)

	cpu.X = 0xff & (cpu.X + 1)
	cpu.Set_nz(cpu.X)

	return oper.Cycles + penalty
}

func DEY(cpu *Core6502, oper *Op6502) int {

	_, _, penalty := oper.Fetch(cpu, &cpu.PC)

	cpu.Y = 0xff & (cpu.Y - 1)
	cpu.Set_nz(cpu.Y)

	return oper.Cycles + penalty
}

func INY(cpu *Core6502, oper *Op6502) int {

	_, _, penalty := oper.Fetch(cpu, &cpu.PC)

	cpu.Y = 0xff & (cpu.Y + 1)
	cpu.Set_nz(cpu.Y)

	return oper.Cycles + penalty
}

func JMP(cpu *Core6502, oper *Op6502) int {

	_, addr, penalty := oper.Fetch(cpu, &cpu.PC)

	cpu.PostJumpEvent(cpu.PC-3, addr, "JMP")

	cpu.PC = addr

	return oper.Cycles + penalty
}

func TAX(cpu *Core6502, oper *Op6502) int {

	_, _, penalty := oper.Fetch(cpu, &cpu.PC)

	cpu.X = cpu.A
	cpu.Set_nz(cpu.X)

	return oper.Cycles + penalty
}

func TAY(cpu *Core6502, oper *Op6502) int {

	_, _, penalty := oper.Fetch(cpu, &cpu.PC)

	cpu.Y = cpu.A
	cpu.Set_nz(cpu.Y)

	return oper.Cycles + penalty
}

func TXA(cpu *Core6502, oper *Op6502) int {

	_, _, penalty := oper.Fetch(cpu, &cpu.PC)

	cpu.A = cpu.X
	cpu.Set_nz(cpu.A)

	return oper.Cycles + penalty
}

func TYA(cpu *Core6502, oper *Op6502) int {

	_, _, penalty := oper.Fetch(cpu, &cpu.PC)

	cpu.A = cpu.Y
	cpu.Set_nz(cpu.A)

	return oper.Cycles + penalty
}

func TSX(cpu *Core6502, oper *Op6502) int {

	_, _, penalty := oper.Fetch(cpu, &cpu.PC)

	cpu.X = cpu.SP & 0xff
	cpu.Set_nz(cpu.X)

	return oper.Cycles + penalty
}

func TXS(cpu *Core6502, oper *Op6502) int {

	_, _, penalty := oper.Fetch(cpu, &cpu.PC)

	cpu.SP = (cpu.X & 0xff) | 0x100
	if cpu.SP > cpu.InitialSP || cpu.IgnoreStackFallouts {
		cpu.InitialSP = cpu.SP
	}

	return oper.Cycles + penalty
}

func JSR(cpu *Core6502, oper *Op6502) int {

	_, addr, penalty := oper.Fetch(cpu, &cpu.PC)

	cpu.PostJumpEvent(cpu.PC-3, addr, "JSR")

	// Modern callback for
	if ok, addpen := cpu.IsCallShimmable(addr); ok {
		return oper.Cycles + penalty + int(addpen)
	}

	if cpu.ROM != nil && cpu.ROM(addr, cpu.Int, false) {
		//fmt.Printf("Using rom shim @%.4x\n", addr)
		return oper.Cycles + penalty
	}

	readdr := cpu.PC - 1

	if addr == 0xbf00 && cpu.UseProDOS {
		cpu.PC = readdr + 4 // skip prodos command tables
		return oper.Cycles + penalty
	}

	cpu.FetchByteAddr(0x100 + cpu.SP)

	cpu.Push(readdr / 256)
	cpu.Push(readdr % 256)

	cpu.PC = addr

	return oper.Cycles + penalty
}

func RTS(cpu *Core6502, oper *Op6502) int {

	// 1, 1, 1, 1, 1

	_, _, penalty := oper.Fetch(cpu, &cpu.PC)

	if cpu.SP == cpu.InitialSP && !settings.PureBoot(cpu.Int.GetMemIndex()) {
		cpu.Halted = true
		return oper.Cycles + penalty
	}

	cpu.FetchByteAddr(0x100 + cpu.SP)
	readdr := cpu.Pop() + 256*cpu.Pop() + 1

	cpu.FetchByteAddr(readdr)

	cpu.PC = readdr

	return oper.Cycles + penalty
}

func RTI(cpu *Core6502, oper *Op6502) int {

	_, _, penalty := oper.Fetch(cpu, &cpu.PC)

	if cpu.SP > cpu.InitialSP && !settings.PureBoot(cpu.Int.GetMemIndex()) {
		cpu.Halted = true
		return oper.Cycles + penalty
	}

	cpu.P = cpu.Pop() & (F_Z | F_V | F_C | F_I | F_N | F_D)

	readdr := cpu.Pop() + 256*cpu.Pop()

	fmt.Printf("*** CPU: return from IRQ to 0x%.4x\n", readdr)

	if settings.Debug6522 {
		fmt2.Printf("6522TRACE: RTI instruction\n")
	}

	cpu.FetchByteAddr(readdr) // we pre-read addr

	cpu.PC = readdr

	return oper.Cycles + penalty
}

/* ============================================================================ *\
|  UNDOCUMENTED 6502 NMOS OP-CODES - HERE BE DRAGONS
\* ============================================================================ */

/*
+-----------------------------------------------------------------------------+
|SAX -     Store Accumulator "AND"              | [4] (A "AND" (MSB(adr)+1)   |
|      [3]   (MSB(Address)+1) "AND"             |       "AND" X) -> M         |
|            Index X in Memory                  |                             |
+-----------------------------------------------------------------------------+
*/

func SAX(cpu *Core6502, oper *Op6502) int {

	_, addr, penalty := oper.Fetch(cpu, &cpu.PC)

	value := cpu.A & cpu.X & ((addr/256 + 1) & 0xff)

	cpu.StoreByteAddr(addr, value)

	return oper.Cycles + penalty
}

/*
+-----------------------------------------------------------------------------+
|AXA -     "AND" Memory with Index X            |     (M "AND" X) -> A        |
|            into Accumulator                   |     Z,N                     |
+-----------------------------------------------------------------------------+
*/

func AXA(cpu *Core6502, oper *Op6502) int {

	imm, addr, penalty := oper.Fetch(cpu, &cpu.PC)

	value := addr
	if !imm {
		value = cpu.FetchByteAddr(addr)
	}

	cpu.A = 0xff & (value & cpu.X)

	cpu.Set_nz(cpu.A)

	return oper.Cycles + penalty
}

/*
+-----------------------------------------------------------------------------+
|SXA -     Store Index X "AND"                  |     (X "AND" A) -> M        |
|            Accumulator in Memory              |                             |
+-----------------------------------------------------------------------------+
*/

func SXA(cpu *Core6502, oper *Op6502) int {

	_, addr, penalty := oper.Fetch(cpu, &cpu.PC)

	value := 0xff & (cpu.A & cpu.X)

	cpu.StoreByteAddr(addr, value)

	return oper.Cycles + penalty
}

/*
+-----------------------------------------------------------------------------+
|SLO -     Shift Memory One Bit Left            |     ASL M                   |
|          THEN "OR" Memory with Accumulator    |     THEN (M "OR" A) -> A,M  |
|            into Accumulator and Memory        |                             |
+-----------------------------------------------------------------------------+
*/

func SLO(cpu *Core6502, oper *Op6502) int {

	imm, addr, penalty := oper.Fetch(cpu, &cpu.PC)

	value := cpu.A
	if !imm {
		value = cpu.FetchByteAddr(addr)
	}

	cpu.SetFlag(F_C, (value&0x80 != 0))
	value = 0x0FE & (value << 1)

	value = value | cpu.A

	cpu.Set_nz(value)

	cpu.StoreByteAddr(addr, value)
	cpu.A = value

	return oper.Cycles + penalty
}

/*
+-----------------------------------------------------------------------------+
|LXA -     Load Index X and Accumulator         |     M -> X,A                |
|            with Memory                        |                             |
+-----------------------------------------------------------------------------+
*/
func LXA(cpu *Core6502, oper *Op6502) int {

	imm, value, penalty := oper.Fetch(cpu, &cpu.PC)

	if imm {
		cpu.A = value & 0xff
		cpu.X = value & 0xff
	} else {
		cpu.A = 0xff & cpu.FetchByteAddr(value)
		cpu.X = 0xff & cpu.FetchByteAddr(value)
	}

	cpu.Set_nz(cpu.A)

	return oper.Cycles + penalty
}

/*
+-----------------------------------------------------------------------------+
|SRE -     Shift Memory One Bit Right           |     LSR M                   |
|          THEN "Exclusive OR" Memory           |     THEN (M "EOR" A) -> A   |
|            with Accumulator                   |                             |
+-----------------------------------------------------------------------------+
*/

func SRE(cpu *Core6502, oper *Op6502) int {

	_, addr, penalty := oper.Fetch(cpu, &cpu.PC)

	value := cpu.FetchByteAddr(addr)

	cpu.SetFlag(F_C, (value&0x01 != 0))
	value = 0x07F & (value >> 1)

	cpu.StoreByteAddr(addr, value)

	cpu.A = 0xff & (value ^ cpu.A)
	cpu.Set_nz(cpu.A)

	return oper.Cycles + penalty
}

/*
+-----------------------------------------------------------------------------+
|RRA -     Rotate Memory One Bit Right          |     ROR M                   |
|          THEN Add Memory to Accumulator       |     THEN (A + M + C) -> A   |
|            with Carry                         |                             |
+-----------------------------------------------------------------------------+
*/

func RRA(cpu *Core6502, oper *Op6502) int {

	_, addr, penalty := oper.Fetch(cpu, &cpu.PC)

	// ROR M

	value := cpu.FetchByteAddr(addr)

	oldC := ab(cpu.TestFlag(F_C), 0x80, 0)

	cpu.SetFlag(F_C, (value&0x1 != 0))
	value = 0x0ff & ((value >> 1) | oldC)
	cpu.Set_nz(value)

	cpu.StoreByteAddr(addr, value)

	// A + M + C -> A
	var w int
	cpu.SetFlag(F_V, ((cpu.A^value)&0x080) == 0)
	if cpu.TestFlag(F_D) {
		// Decimal Mode
		w = (cpu.A & 0x0f) + (value & 0x0f) + ab(cpu.TestFlag(F_C), 1, 0)
		if w >= 10 {
			w = 0x010 | ((w + 6) & 0x0f)
		}
		w += (cpu.A & 0x0f0) + (value & 0x00f0)
		if w >= 0x0A0 {
			cpu.SetFlag(F_C, true)
			if cpu.TestFlag(F_V) && w >= 0x0180 {
				cpu.SetFlag(F_V, false)
			}
			w += 0x060
		} else {
			cpu.SetFlag(F_C, false)
			if cpu.TestFlag(F_V) && w < 0x080 {
				cpu.SetFlag(F_V, false)
			}
		}
	} else {
		// Binary Mode
		w = cpu.A + value + ab(cpu.TestFlag(F_C), 1, 0)
		if w >= 0x0100 {
			cpu.SetFlag(F_C, true)
			if cpu.TestFlag(F_V) && w >= 0x0180 {
				cpu.SetFlag(F_V, false)
			}
		} else {
			cpu.SetFlag(F_C, false)
			if cpu.TestFlag(F_V) && w < 0x080 {
				cpu.SetFlag(F_V, false)
			}
		}
	}
	cpu.A = w & 0x0ff
	cpu.Set_nz(cpu.A)

	return oper.Cycles + penalty
}

/*
+-----------------------------------------------------------------------------+
|INS -     Increment Memory by One              |     (M + 1) -> M            |
|          THEN Subtract Memory from            |     THEN (A - M - ~C) -> A  |
|            Accumulator with Borrow            |                             |
+-----------------------------------------------------------------------------+
*/

func INS(cpu *Core6502, oper *Op6502) int {

	_, addr, penalty := oper.Fetch(cpu, &cpu.PC)

	// (M + 1) -> M

	value := cpu.FetchByteAddr(addr)

	value = 0xff & (value + 1)
	cpu.Set_nz(value)

	cpu.StoreByteAddr(addr, value)

	// THEN (A - M - ~C) -> A
	cpu.SetFlag(F_V, ((cpu.A^value)&0x080) != 0)
	var w int
	if cpu.TestFlag(F_D) {
		temp := 0x0f + (cpu.A & 0x0f) - (value & 0x0f) + ab(cpu.TestFlag(F_C), 1, 0)
		if temp < 0x10 {
			w = 0
			temp -= 6
		} else {
			w = 0x10
			temp -= 0x10
		}
		w += 0x00f0 + (cpu.A & 0x00f0) - (value & 0x00f0)
		if w < 0x100 {
			cpu.SetFlag(F_C, false)
			if cpu.TestFlag(F_V) && w < 0x080 {
				cpu.SetFlag(F_V, false)
			}
			w -= 0x60
		} else {
			cpu.SetFlag(F_C, true)
			if cpu.TestFlag(F_V) && w >= 0x180 {
				cpu.SetFlag(F_V, false)
			}
		}
		w += temp
	} else {
		w = 0x0ff + cpu.A - value + ab(cpu.TestFlag(F_C), 1, 0)
		if w < 0x100 {
			cpu.SetFlag(F_C, false)
			if cpu.TestFlag(F_V) && (w < 0x080) {
				cpu.SetFlag(F_V, false)
			}
		} else {
			cpu.SetFlag(F_C, true)
			if cpu.TestFlag(F_V) && (w >= 0x180) {
				cpu.SetFlag(F_V, false)
			}
		}
	}
	cpu.A = w & 0x0ff
	cpu.Set_nz(cpu.A)

	return oper.Cycles + penalty
}

/*
+-----------------------------------------------------------------------------+
|ANC -     "AND" Memory with Accumulator        |     (M "AND" A) -> A        |
|      [1] THEN Copy Bit 7 of Result into Carry |     THEN msb(A) -> C        |
+-----------------------------------------------------------------------------+
*/

func ANC(cpu *Core6502, oper *Op6502) int {

	imm, addr, penalty := oper.Fetch(cpu, &cpu.PC)

	value := addr
	if !imm {
		value = cpu.FetchByteAddr(addr)
	}

	cpu.A &= value
	cpu.Set_nz(cpu.A)

	cpu.SetFlag(F_C, cpu.TestFlag(F_N))

	return oper.Cycles + penalty
}

/*
+-----------------------------------------------------------------------------+
|RLA -     Rotate Memory One Bit Left           |     ROL M                   |
|          THEN "AND" Memory with Accumulator   |     THEN (M "AND" A) -> A   |
+-----------------------------------------------------------------------------+
*/

func RLA(cpu *Core6502, oper *Op6502) int {

	_, addr, penalty := oper.Fetch(cpu, &cpu.PC)

	value := cpu.FetchByteAddr(addr)

	oldC := ab(cpu.TestFlag(F_C), 1, 0)

	cpu.SetFlag(F_C, (value&0x80 != 0))
	value = 0x0ff & ((value << 1) | oldC)

	cpu.StoreByteAddr(addr, value)
	cpu.A = value & cpu.A

	cpu.Set_nz(cpu.A)

	return oper.Cycles + penalty
}

/*
+-----------------------------------------------------------------------------+
|DCP -     Decrement Memory by One              |     (M - 1) -> M            |
|          THEN Compare Memory with Accumulator |     THEN CMP M              |
+-----------------------------------------------------------------------------+
*/

func DCP(cpu *Core6502, oper *Op6502) int {

	_, addr, penalty := oper.Fetch(cpu, &cpu.PC)

	value := cpu.FetchByteAddr(addr)

	value = 0xff & (value - 1)
	cpu.Set_nz(value)

	cpu.StoreByteAddr(addr, value)

	// cmp
	val := cpu.A - value

	cpu.SetFlag(F_C, (val >= 0))
	cpu.Set_nz(val)

	return oper.Cycles + penalty
}

/*
+-----------------------------------------------------------------------------+
|SBX -     Subtract Memory from Index X         | [5] (X - M) -> X            |
|            _without_ Borrow or Decimal Mode   |                             |
+-----------------------------------------------------------------------------+
*/

func SBX(cpu *Core6502, oper *Op6502) int {

	imm, addr, penalty := oper.Fetch(cpu, &cpu.PC)

	value := addr

	if !imm {
		value = cpu.FetchByteAddr(addr)
	}

	cpu.SetFlag(F_V, ((cpu.X^value)&0x080) != 0)
	var w int
	w = 0x0ff + cpu.X - value
	if w < 0x100 {
		cpu.SetFlag(F_C, false)
		if cpu.TestFlag(F_V) && (w < 0x080) {
			cpu.SetFlag(F_V, false)
		}
	} else {
		cpu.SetFlag(F_C, true)
		if cpu.TestFlag(F_V) && (w >= 0x180) {
			cpu.SetFlag(F_V, false)
		}
	}
	cpu.X = w & 0x0ff
	cpu.Set_nz(cpu.X)

	return oper.Cycles + penalty
}

func TRB(cpu *Core6502, oper *Op6502) int {

	_, addr, penalty := oper.Fetch(cpu, &cpu.PC)

	value := cpu.FetchByteAddr(addr)

	cpu.SetFlag(F_C, (value&cpu.A) != 0)

	cpu.StoreByteAddr(addr, value&(255-cpu.A))

	return oper.Cycles + penalty
}

func TSB(cpu *Core6502, oper *Op6502) int {

	_, addr, penalty := oper.Fetch(cpu, &cpu.PC)

	value := cpu.FetchByteAddr(addr)

	cpu.SetFlag(F_C, (value&cpu.A) != 0)

	cpu.StoreByteAddr(addr, value|cpu.A)

	return oper.Cycles + penalty
}

func BBSn(cpu *Core6502, oper *Op6502, bit uint) int {

	addr, jump, penalty := ZP_RELATIVE(cpu, &cpu.PC)

	value := cpu.FetchByteAddr(addr)
	cpu.FetchByteAddr(addr)
	if value&(1<<bit) != 0 {
		cpu.PC = jump
	}

	return oper.Cycles + penalty
}

func BBRn(cpu *Core6502, oper *Op6502, bit uint) int {

	addr, jump, penalty := ZP_RELATIVE(cpu, &cpu.PC)

	value := cpu.FetchByteAddr(addr)
	cpu.FetchByteAddr(addr)
	if value&(1<<bit) == 0 {
		cpu.PC = jump
	}

	return oper.Cycles + penalty
}

func BBS0(cpu *Core6502, oper *Op6502) int {
	return BBSn(cpu, oper, 0)
}

func BBS1(cpu *Core6502, oper *Op6502) int {
	return BBSn(cpu, oper, 1)
}

func BBS2(cpu *Core6502, oper *Op6502) int {
	return BBSn(cpu, oper, 2)
}

func BBS3(cpu *Core6502, oper *Op6502) int {
	return BBSn(cpu, oper, 3)
}

func BBS4(cpu *Core6502, oper *Op6502) int {
	return BBSn(cpu, oper, 4)
}

func BBS5(cpu *Core6502, oper *Op6502) int {
	return BBSn(cpu, oper, 5)
}

func BBS6(cpu *Core6502, oper *Op6502) int {
	return BBSn(cpu, oper, 6)
}

func BBS7(cpu *Core6502, oper *Op6502) int {
	return BBSn(cpu, oper, 7)
}

func BBR0(cpu *Core6502, oper *Op6502) int {
	return BBRn(cpu, oper, 0)
}

func BBR1(cpu *Core6502, oper *Op6502) int {
	return BBRn(cpu, oper, 1)
}

func BBR2(cpu *Core6502, oper *Op6502) int {
	return BBRn(cpu, oper, 2)
}

func BBR3(cpu *Core6502, oper *Op6502) int {
	return BBRn(cpu, oper, 3)
}

func BBR4(cpu *Core6502, oper *Op6502) int {
	return BBRn(cpu, oper, 4)
}

func BBR5(cpu *Core6502, oper *Op6502) int {
	return BBRn(cpu, oper, 5)
}

func BBR6(cpu *Core6502, oper *Op6502) int {
	return BBRn(cpu, oper, 6)
}

func BBR7(cpu *Core6502, oper *Op6502) int {
	return BBRn(cpu, oper, 7)
}

func SMBn(cpu *Core6502, oper *Op6502, bit uint) int {

	_, addr, penalty := oper.Fetch(cpu, &cpu.PC)

	value := cpu.FetchByteAddr(addr)

	value = value | (1 << bit)

	cpu.StoreByteAddr(addr, value)

	return oper.Cycles + penalty
}

func RMBn(cpu *Core6502, oper *Op6502, bit uint) int {

	_, addr, penalty := oper.Fetch(cpu, &cpu.PC)

	value := cpu.FetchByteAddr(addr)

	value = value & (255 - (1 << bit))

	cpu.StoreByteAddr(addr, value)

	return oper.Cycles + penalty
}

func RMB0(cpu *Core6502, oper *Op6502) int {
	return RMBn(cpu, oper, 0)
}

func RMB1(cpu *Core6502, oper *Op6502) int {
	return RMBn(cpu, oper, 1)
}

func RMB2(cpu *Core6502, oper *Op6502) int {
	return RMBn(cpu, oper, 2)
}

func RMB3(cpu *Core6502, oper *Op6502) int {
	return RMBn(cpu, oper, 3)
}

func RMB4(cpu *Core6502, oper *Op6502) int {
	return RMBn(cpu, oper, 4)
}

func RMB5(cpu *Core6502, oper *Op6502) int {
	return RMBn(cpu, oper, 5)
}

func RMB6(cpu *Core6502, oper *Op6502) int {
	return RMBn(cpu, oper, 6)
}

func RMB7(cpu *Core6502, oper *Op6502) int {
	return RMBn(cpu, oper, 7)
}

func SMB0(cpu *Core6502, oper *Op6502) int {
	return SMBn(cpu, oper, 0)
}

func SMB1(cpu *Core6502, oper *Op6502) int {
	return SMBn(cpu, oper, 1)
}

func SMB2(cpu *Core6502, oper *Op6502) int {
	return SMBn(cpu, oper, 2)
}

func SMB3(cpu *Core6502, oper *Op6502) int {
	return SMBn(cpu, oper, 3)
}

func SMB4(cpu *Core6502, oper *Op6502) int {
	return SMBn(cpu, oper, 4)
}

func SMB5(cpu *Core6502, oper *Op6502) int {
	return SMBn(cpu, oper, 5)
}

func SMB6(cpu *Core6502, oper *Op6502) int {
	return SMBn(cpu, oper, 6)
}

func SMB7(cpu *Core6502, oper *Op6502) int {
	return SMBn(cpu, oper, 7)
}

/*
ARR (ARR) [ARR]
~~~~~~~~~~~~~~~
AND byte with accumulator, then rotate one bit right in accu-
mulator and check bit 5 and 6:
If both bits are 1: set C, clear V.
If both bits are 0: clear C and V.
If only bit 5 is 1: set V, clear C.
If only bit 6 is 1: set C and V.
Status flags: N,V,Z,C

Addressing  |Mnemonics  |Opc|Sz | n
------------|-----------|---|---|---
Immediate   |ARR #arg   |$6B| 2 | 2
*/

func ARR(cpu *Core6502, oper *Op6502) int {

	imm, addr, penalty := oper.Fetch(cpu, &cpu.PC)

	value := addr
	if !imm {
		value = cpu.FetchByteAddr(addr)
	}

	cpu.A &= value

	// ROR
	oldC := ab(cpu.TestFlag(F_C), 0x80, 0)

	cpu.SetFlag(F_C, (cpu.A&0x1 != 0))
	cpu.A = 0x0ff & ((cpu.A >> 1) | oldC)
	cpu.Set_nz(cpu.A)

	// Check
	switch cpu.A & 96 {
	case 96:
		cpu.SetFlag(F_C, true)
		cpu.SetFlag(F_V, false)
	case 0:
		cpu.SetFlag(F_C, false)
		cpu.SetFlag(F_V, false)
	case 32:
		cpu.SetFlag(F_C, false)
		cpu.SetFlag(F_V, true)
	case 64:
		cpu.SetFlag(F_C, true)
		cpu.SetFlag(F_V, true)
	}

	return oper.Cycles + penalty
}

func (this *Core6502) OpTableTest() {
	for opcode := 0; opcode < 256; opcode++ {
		if this.Opref[opcode] == nil {
			continue
		}
		this.PC = 0x2000
		this.StoreByteAddr(this.PC, opcode)
		this.StoreByteAddr(this.PC+1, 0x12)
		this.StoreByteAddr(this.PC+2, 0x34)
		this.FetchExecute()
	}
	os.Exit(0)
}
