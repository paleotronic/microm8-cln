package mos6502

import (
	"paleotronic.com/core/interfaces"
)

func NewCore65C02(mem interfaces.Interpretable, a int, x int, y int, pc int, sr int, sp int, vdu WaveStreamer) *Core6502 {
	this := NewCore6502(mem, a, x, y, pc, sr, sp, vdu)
	this.InitOpListCMOS()
	this.IgnoreILL = !STOP65C02

	return this
}

func (this *Core6502) InitOpListCMOS() {
	this.Model = "65C02"
	//this.Opref = make(map[int]*Op6502)
	this.Opref[0x69] = NewOp6502("ADC #oper", "immidiate", 0x69, 2, 2, IMMEDIATE, MODE_IMMEDIATE, ADC)
	this.Opref[0x65] = NewOp6502("ADC oper", "zeropage", 0x65, 2, 3, ZEROPAGE, MODE_ZEROPAGE, ADC)
	this.Opref[0x75] = NewOp6502("ADC oper,X", "zeropage,X", 0x75, 2, 4, ZEROPAGE_X, MODE_ZEROPAGE_X, ADC)
	this.Opref[0x6D] = NewOp6502("ADC oper", "absolute", 0x6D, 3, 4, ABSOLUTE, MODE_ABSOLUTE, ADC)
	this.Opref[0x7D] = NewOp6502("ADC oper,X", "absolute,X", 0x7D, 3, 4, ABSOLUTE_X, MODE_ABSOLUTE_X, ADC)
	this.Opref[0x79] = NewOp6502("ADC oper,Y", "absolute,Y", 0x79, 3, 4, ABSOLUTE_Y, MODE_ABSOLUTE_Y, ADC)
	this.Opref[0x61] = NewOp6502("ADC (oper,X)", "(indirect,X)", 0x61, 2, 6, INDIRECT_ZP_X, MODE_INDIRECT_ZP_X, ADC)
	this.Opref[0x71] = NewOp6502("ADC (oper),Y", "(indirect),Y", 0x71, 2, 5, INDIRECT_ZP_Y, MODE_INDIRECT_ZP_Y, ADC)
	this.Opref[0x29] = NewOp6502("AND #oper", "immidiate", 0x29, 2, 2, IMMEDIATE, MODE_IMMEDIATE, AND)
	this.Opref[0x25] = NewOp6502("AND oper", "zeropage", 0x25, 2, 3, ZEROPAGE, MODE_ZEROPAGE, AND)
	this.Opref[0x35] = NewOp6502("AND oper,X", "zeropage,X", 0x35, 2, 4, ZEROPAGE_X, MODE_ZEROPAGE_X, AND)
	this.Opref[0x2D] = NewOp6502("AND oper", "absolute", 0x2D, 3, 4, ABSOLUTE, MODE_ABSOLUTE, AND)
	this.Opref[0x3D] = NewOp6502("AND oper,X", "absolute,X", 0x3D, 3, 4, ABSOLUTE_X, MODE_ABSOLUTE_X, AND)
	this.Opref[0x39] = NewOp6502("AND oper,Y", "absolute,Y", 0x39, 3, 4, ABSOLUTE_Y, MODE_ABSOLUTE_Y, AND)
	this.Opref[0x21] = NewOp6502("AND (oper,X)", "(indirect,X)", 0x21, 2, 6, INDIRECT_ZP_X, MODE_INDIRECT_ZP_X, AND)
	this.Opref[0x31] = NewOp6502("AND (oper),Y", "(indirect),Y", 0x31, 2, 5, INDIRECT_ZP_Y, MODE_INDIRECT_ZP_Y, AND)
	this.Opref[0x0A] = NewOp6502("ASL A", "accumulator", 0x0A, 1, 2, IMPLIED, MODE_IMPLIED, ASL)
	this.Opref[0x06] = NewOp6502("ASL oper", "zeropage", 0x06, 2, 5, ZEROPAGE, MODE_ZEROPAGE, ASL)
	this.Opref[0x16] = NewOp6502("ASL oper,X", "zeropage,X", 0x16, 2, 6, ZEROPAGE_X, MODE_ZEROPAGE_X, ASL)
	this.Opref[0x0E] = NewOp6502("ASL oper", "absolute", 0x0E, 3, 6, ABSOLUTE, MODE_ABSOLUTE, ASL)
	this.Opref[0x1E] = NewOp6502("ASL oper,X", "absolute,X", 0x1E, 3, 7, ABSOLUTE_X_WRITE, MODE_ABSOLUTE_X, ASL)
	this.Opref[0x90] = NewOp6502("BCC oper", "relative", 0x90, 2, 2, RELATIVE, MODE_RELATIVE, BCC)
	this.Opref[0xB0] = NewOp6502("BCS oper", "relative", 0xB0, 2, 2, RELATIVE, MODE_RELATIVE, BCS)
	this.Opref[0xF0] = NewOp6502("BEQ oper", "relative", 0xF0, 2, 2, RELATIVE, MODE_RELATIVE, BEQ)
	this.Opref[0x24] = NewOp6502("BIT oper", "zeropage", 0x24, 2, 3, ZEROPAGE, MODE_ZEROPAGE, BIT)
	this.Opref[0x2C] = NewOp6502("BIT oper", "absolute", 0x2C, 3, 4, ABSOLUTE, MODE_ABSOLUTE, BIT)
	this.Opref[0x30] = NewOp6502("BMI oper", "relative", 0x30, 2, 2, RELATIVE, MODE_RELATIVE, BMI)
	this.Opref[0xD0] = NewOp6502("BNE oper", "relative", 0xD0, 2, 2, RELATIVE, MODE_RELATIVE, BNE)
	this.Opref[0x10] = NewOp6502("BPL oper", "relative", 0x10, 2, 2, RELATIVE, MODE_RELATIVE, BPL)
	this.Opref[0x00] = NewOp6502("BRK", "implied", 0x00, 1, 7, IMPLIED, MODE_IMPLIED, BRK)
	this.Opref[0x50] = NewOp6502("BVC oper", "relative", 0x50, 2, 2, RELATIVE, MODE_RELATIVE, BVC)
	this.Opref[0x70] = NewOp6502("BVS oper", "relative", 0x70, 2, 2, RELATIVE, MODE_RELATIVE, BVS)
	this.Opref[0x18] = NewOp6502("CLC", "implied", 0x18, 1, 2, IMPLIED, MODE_IMPLIED, CLC)
	this.Opref[0xD8] = NewOp6502("CLD", "implied", 0xD8, 1, 2, IMPLIED, MODE_IMPLIED, CLD)
	this.Opref[0x58] = NewOp6502("CLI", "implied", 0x58, 1, 2, IMPLIED, MODE_IMPLIED, CLI)
	this.Opref[0xB8] = NewOp6502("CLV", "implied", 0xB8, 1, 2, IMPLIED, MODE_IMPLIED, CLV)
	this.Opref[0xC9] = NewOp6502("CMP #oper", "immidiate", 0xC9, 2, 2, IMMEDIATE, MODE_IMMEDIATE, CMP)
	this.Opref[0xC5] = NewOp6502("CMP oper", "zeropage", 0xC5, 2, 3, ZEROPAGE, MODE_ZEROPAGE, CMP)
	this.Opref[0xD5] = NewOp6502("CMP oper,X", "zeropage,X", 0xD5, 2, 4, ZEROPAGE_X, MODE_ZEROPAGE_X, CMP)
	this.Opref[0xCD] = NewOp6502("CMP oper", "absolute", 0xCD, 3, 4, ABSOLUTE, MODE_ABSOLUTE, CMP)
	this.Opref[0xDD] = NewOp6502("CMP oper,X", "absolute,X", 0xDD, 3, 4, ABSOLUTE_X, MODE_ABSOLUTE_X, CMP)
	this.Opref[0xD9] = NewOp6502("CMP oper,Y", "absolute,Y", 0xD9, 3, 4, ABSOLUTE_Y, MODE_ABSOLUTE_Y, CMP)
	this.Opref[0xC1] = NewOp6502("CMP (oper,X)", "(indirect,X)", 0xC1, 2, 6, INDIRECT_ZP_X, MODE_INDIRECT_ZP_X, CMP)
	this.Opref[0xD1] = NewOp6502("CMP (oper),Y", "(indirect),Y", 0xD1, 2, 5, INDIRECT_ZP_Y, MODE_INDIRECT_ZP_Y, CMP)
	this.Opref[0xE0] = NewOp6502("CPX #oper", "immidiate", 0xE0, 2, 2, IMMEDIATE, MODE_IMMEDIATE, CPX)
	this.Opref[0xE4] = NewOp6502("CPX oper", "zeropage", 0xE4, 2, 3, ZEROPAGE, MODE_ZEROPAGE, CPX)
	this.Opref[0xEC] = NewOp6502("CPX oper", "absolute", 0xEC, 3, 4, ABSOLUTE, MODE_ABSOLUTE, CPX)
	this.Opref[0xC0] = NewOp6502("CPY #oper", "immidiate", 0xC0, 2, 2, IMMEDIATE, MODE_IMMEDIATE, CPY)
	this.Opref[0xC4] = NewOp6502("CPY oper", "zeropage", 0xC4, 2, 3, ZEROPAGE, MODE_ZEROPAGE, CPY)
	this.Opref[0xCC] = NewOp6502("CPY oper", "absolute", 0xCC, 3, 4, ABSOLUTE, MODE_ABSOLUTE, CPY)
	this.Opref[0xC6] = NewOp6502("DEC oper", "zeropage", 0xC6, 2, 5, ZEROPAGE, MODE_ZEROPAGE, DEC)
	this.Opref[0xD6] = NewOp6502("DEC oper,X", "zeropage,X", 0xD6, 2, 6, ZEROPAGE_X, MODE_ZEROPAGE_X, DEC)
	this.Opref[0xCE] = NewOp6502("DEC oper", "absolute", 0xCE, 3, 6, ABSOLUTE, MODE_ABSOLUTE, DEC)
	this.Opref[0xDE] = NewOp6502("DEC oper,X", "absolute,X", 0xDE, 3, 7, ABSOLUTE_X_WRITE, MODE_ABSOLUTE_X, DEC)
	this.Opref[0xCA] = NewOp6502("DEX", "implied", 0xCA, 1, 2, IMPLIED, MODE_IMPLIED, DEX)
	this.Opref[0x88] = NewOp6502("DEY", "implied", 0x88, 1, 2, IMPLIED, MODE_IMPLIED, DEY)
	this.Opref[0x49] = NewOp6502("EOR #oper", "immidiate", 0x49, 2, 2, IMMEDIATE, MODE_IMMEDIATE, EOR)
	this.Opref[0x45] = NewOp6502("EOR oper", "zeropage", 0x45, 2, 3, ZEROPAGE, MODE_ZEROPAGE, EOR)
	this.Opref[0x55] = NewOp6502("EOR oper,X", "zeropage,X", 0x55, 2, 4, ZEROPAGE_X, MODE_ZEROPAGE_X, EOR)
	this.Opref[0x4D] = NewOp6502("EOR oper", "absolute", 0x4D, 3, 4, ABSOLUTE, MODE_ABSOLUTE, EOR)
	this.Opref[0x5D] = NewOp6502("EOR oper,X", "absolute,X", 0x5D, 3, 4, ABSOLUTE_X, MODE_ABSOLUTE_X, EOR)
	this.Opref[0x59] = NewOp6502("EOR oper,Y", "absolute,Y", 0x59, 3, 4, ABSOLUTE_Y, MODE_ABSOLUTE_Y, EOR)
	this.Opref[0x41] = NewOp6502("EOR (oper,X)", "(indirect,X)", 0x41, 2, 6, INDIRECT_ZP_X, MODE_INDIRECT_ZP_X, EOR)
	this.Opref[0x51] = NewOp6502("EOR (oper),Y", "(indirect),Y", 0x51, 2, 5, INDIRECT_ZP_Y, MODE_INDIRECT_ZP_Y, EOR)
	this.Opref[0xE6] = NewOp6502("INC oper", "zeropage", 0xE6, 2, 5, ZEROPAGE, MODE_ZEROPAGE, INC)
	this.Opref[0xF6] = NewOp6502("INC oper,X", "zeropage,X", 0xF6, 2, 6, ZEROPAGE_X, MODE_ZEROPAGE_X, INC)
	this.Opref[0xEE] = NewOp6502("INC oper", "absolute", 0xEE, 3, 6, ABSOLUTE, MODE_ABSOLUTE, INC)
	this.Opref[0xFE] = NewOp6502("INC oper,X", "absolute,X", 0xFE, 3, 7, ABSOLUTE_X_WRITE, MODE_ABSOLUTE_X, INC)
	this.Opref[0xE8] = NewOp6502("INX", "implied", 0xE8, 1, 2, IMPLIED, MODE_IMPLIED, INX)
	this.Opref[0xC8] = NewOp6502("INY", "implied", 0xC8, 1, 2, IMPLIED, MODE_IMPLIED, INY)
	this.Opref[0x4C] = NewOp6502("JMP oper", "absolute", 0x4C, 3, 3, ABSOLUTE, MODE_ABSOLUTE, JMP)
	this.Opref[0x6C] = NewOp6502("JMP (oper)", "indirect", 0x6C, 3, 5, INDIRECT, MODE_INDIRECT, JMP)
	this.Opref[0x20] = NewOp6502("JSR oper", "absolute", 0x20, 3, 6, ABSOLUTE, MODE_ABSOLUTE, JSR)
	this.Opref[0xA9] = NewOp6502("LDA #oper", "immidiate", 0xA9, 2, 2, IMMEDIATE, MODE_IMMEDIATE, LDA)
	this.Opref[0xA5] = NewOp6502("LDA oper", "zeropage", 0xA5, 2, 3, ZEROPAGE, MODE_ZEROPAGE, LDA)
	this.Opref[0xB5] = NewOp6502("LDA oper,X", "zeropage,X", 0xB5, 2, 4, ZEROPAGE_X, MODE_ZEROPAGE_X, LDA)
	this.Opref[0xAD] = NewOp6502("LDA oper", "absolute", 0xAD, 3, 4, ABSOLUTE, MODE_ABSOLUTE, LDA)
	this.Opref[0xBD] = NewOp6502("LDA oper,X", "absolute,X", 0xBD, 3, 4, ABSOLUTE_X, MODE_ABSOLUTE_X, LDA)
	this.Opref[0xB9] = NewOp6502("LDA oper,Y", "absolute,Y", 0xB9, 3, 4, ABSOLUTE_Y, MODE_ABSOLUTE_Y, LDA)
	this.Opref[0xA1] = NewOp6502("LDA (oper,X)", "(indirect,X)", 0xA1, 2, 6, INDIRECT_ZP_X, MODE_INDIRECT_ZP_X, LDA)
	this.Opref[0xB1] = NewOp6502("LDA (oper),Y", "(indirect),Y", 0xB1, 2, 5, INDIRECT_ZP_Y, MODE_INDIRECT_ZP_Y, LDA)
	this.Opref[0xA2] = NewOp6502("LDX #oper", "immidiate", 0xA2, 2, 2, IMMEDIATE, MODE_IMMEDIATE, LDX)
	this.Opref[0xA6] = NewOp6502("LDX oper", "zeropage", 0xA6, 2, 3, ZEROPAGE, MODE_ZEROPAGE, LDX)
	this.Opref[0xB6] = NewOp6502("LDX oper,Y", "zeropage,Y", 0xB6, 2, 4, ZEROPAGE_Y, MODE_ZEROPAGE_Y, LDX)
	this.Opref[0xAE] = NewOp6502("LDX oper", "absolute", 0xAE, 3, 4, ABSOLUTE, MODE_ABSOLUTE, LDX)
	this.Opref[0xBE] = NewOp6502("LDX oper,Y", "absolute,Y", 0xBE, 3, 4, ABSOLUTE_Y, MODE_ABSOLUTE_Y, LDX)
	this.Opref[0xA0] = NewOp6502("LDY #oper", "immidiate", 0xA0, 2, 2, IMMEDIATE, MODE_IMMEDIATE, LDY)
	this.Opref[0xA4] = NewOp6502("LDY oper", "zeropage", 0xA4, 2, 3, ZEROPAGE, MODE_ZEROPAGE, LDY)
	this.Opref[0xB4] = NewOp6502("LDY oper,X", "zeropage,X", 0xB4, 2, 4, ZEROPAGE_X, MODE_ZEROPAGE_X, LDY)
	this.Opref[0xAC] = NewOp6502("LDY oper", "absolute", 0xAC, 3, 4, ABSOLUTE, MODE_ABSOLUTE, LDY)
	this.Opref[0xBC] = NewOp6502("LDY oper,X", "absolute,X", 0xBC, 3, 4, ABSOLUTE_X, MODE_ABSOLUTE_X, LDY)
	this.Opref[0x4A] = NewOp6502("LSR A", "accumulator", 0x4A, 1, 2, IMPLIED, MODE_IMPLIED, LSR)
	this.Opref[0x46] = NewOp6502("LSR oper", "zeropage", 0x46, 2, 5, ZEROPAGE, MODE_ZEROPAGE, LSR)
	this.Opref[0x56] = NewOp6502("LSR oper,X", "zeropage,X", 0x56, 2, 6, ZEROPAGE_X, MODE_ZEROPAGE_X, LSR)
	this.Opref[0x4E] = NewOp6502("LSR oper", "absolute", 0x4E, 3, 6, ABSOLUTE, MODE_ABSOLUTE, LSR)
	this.Opref[0x5E] = NewOp6502("LSR oper,X", "absolute,X", 0x5E, 3, 7, ABSOLUTE_X_WRITE, MODE_ABSOLUTE_X, LSR)
	this.Opref[0xEA] = NewOp6502("NOP", "implied", 0xEA, 1, 2, IMPLIED, MODE_IMPLIED, NOP)
	this.Opref[0x09] = NewOp6502("ORA #oper", "immidiate", 0x09, 2, 2, IMMEDIATE, MODE_IMMEDIATE, ORA)
	this.Opref[0x05] = NewOp6502("ORA oper", "zeropage", 0x05, 2, 3, ZEROPAGE, MODE_ZEROPAGE, ORA)
	this.Opref[0x15] = NewOp6502("ORA oper,X", "zeropage,X", 0x15, 2, 4, ZEROPAGE_X, MODE_ZEROPAGE_X, ORA)
	this.Opref[0x0D] = NewOp6502("ORA oper", "absolute", 0x0D, 3, 4, ABSOLUTE, MODE_ABSOLUTE, ORA)
	this.Opref[0x1D] = NewOp6502("ORA oper,X", "absolute,X", 0x1D, 3, 4, ABSOLUTE_X, MODE_ABSOLUTE_X, ORA)
	this.Opref[0x19] = NewOp6502("ORA oper,Y", "absolute,Y", 0x19, 3, 4, ABSOLUTE_Y, MODE_ABSOLUTE_Y, ORA)
	this.Opref[0x01] = NewOp6502("ORA (oper,X)", "(indirect,X)", 0x01, 2, 6, INDIRECT_ZP_X, MODE_INDIRECT_ZP_X, ORA)
	this.Opref[0x11] = NewOp6502("ORA (oper),Y", "(indirect),Y", 0x11, 2, 5, INDIRECT_ZP_Y, MODE_INDIRECT_ZP_Y, ORA)
	this.Opref[0x48] = NewOp6502("PHA", "implied", 0x48, 1, 3, IMPLIED, MODE_IMPLIED, PHA)
	this.Opref[0x08] = NewOp6502("PHP", "implied", 0x08, 1, 3, IMPLIED, MODE_IMPLIED, PHP)
	this.Opref[0x68] = NewOp6502("PLA", "implied", 0x68, 1, 4, IMPLIED, MODE_IMPLIED, PLA)
	this.Opref[0x28] = NewOp6502("PLP", "implied", 0x28, 1, 4, IMPLIED, MODE_IMPLIED, PLP)
	this.Opref[0x2A] = NewOp6502("ROL A", "accumulator", 0x2A, 1, 2, IMPLIED, MODE_IMPLIED, ROL)
	this.Opref[0x26] = NewOp6502("ROL oper", "zeropage", 0x26, 2, 5, ZEROPAGE, MODE_ZEROPAGE, ROL)
	this.Opref[0x36] = NewOp6502("ROL oper,X", "zeropage,X", 0x36, 2, 6, ZEROPAGE_X, MODE_ZEROPAGE_X, ROL)
	this.Opref[0x2E] = NewOp6502("ROL oper", "absolute", 0x2E, 3, 6, ABSOLUTE, MODE_ABSOLUTE, ROL)
	this.Opref[0x3E] = NewOp6502("ROL oper,X", "absolute,X", 0x3E, 3, 7, ABSOLUTE_X_WRITE, MODE_ABSOLUTE_X, ROL)
	this.Opref[0x6A] = NewOp6502("ROR A", "accumulator", 0x6A, 1, 2, IMPLIED, MODE_IMPLIED, ROR)
	this.Opref[0x66] = NewOp6502("ROR oper", "zeropage", 0x66, 2, 5, ZEROPAGE, MODE_ZEROPAGE, ROR)
	this.Opref[0x76] = NewOp6502("ROR oper,X", "zeropage,X", 0x76, 2, 6, ZEROPAGE_X, MODE_ZEROPAGE_X, ROR)
	this.Opref[0x6E] = NewOp6502("ROR oper", "absolute", 0x6E, 3, 6, ABSOLUTE, MODE_ABSOLUTE, ROR)
	this.Opref[0x7E] = NewOp6502("ROR oper,X", "absolute,X", 0x7E, 3, 7, ABSOLUTE_X_WRITE, MODE_ABSOLUTE_X, ROR)
	this.Opref[0x40] = NewOp6502("RTI", "implied", 0x40, 1, 6, IMPLIED, MODE_IMPLIED, RTI)
	this.Opref[0x60] = NewOp6502("RTS", "implied", 0x60, 1, 6, IMPLIED, MODE_IMPLIED, RTS)
	this.Opref[0xE9] = NewOp6502("SBC #oper", "immidiate", 0xE9, 2, 2, IMMEDIATE, MODE_IMMEDIATE, SBC)
	this.Opref[0xE5] = NewOp6502("SBC oper", "zeropage", 0xE5, 2, 3, ZEROPAGE, MODE_ZEROPAGE, SBC)
	this.Opref[0xF5] = NewOp6502("SBC oper,X", "zeropage,X", 0xF5, 2, 4, ZEROPAGE_X, MODE_ZEROPAGE_X, SBC)
	this.Opref[0xED] = NewOp6502("SBC oper", "absolute", 0xED, 3, 4, ABSOLUTE, MODE_ABSOLUTE, SBC)
	this.Opref[0xFD] = NewOp6502("SBC oper,X", "absolute,X", 0xFD, 3, 4, ABSOLUTE_X, MODE_ABSOLUTE_X, SBC)
	this.Opref[0xF9] = NewOp6502("SBC oper,Y", "absolute,Y", 0xF9, 3, 4, ABSOLUTE_Y, MODE_ABSOLUTE_Y, SBC)
	this.Opref[0xE1] = NewOp6502("SBC (oper,X)", "(indirect,X)", 0xE1, 2, 6, INDIRECT_ZP_X, MODE_INDIRECT_ZP_X, SBC)
	this.Opref[0xF1] = NewOp6502("SBC (oper),Y", "(indirect),Y", 0xF1, 2, 5, INDIRECT_ZP_Y, MODE_INDIRECT_ZP_Y, SBC)
	this.Opref[0x38] = NewOp6502("SEC", "implied", 0x38, 1, 2, IMPLIED, MODE_IMPLIED, SEC)
	this.Opref[0xF8] = NewOp6502("SED", "implied", 0xF8, 1, 2, IMPLIED, MODE_IMPLIED, SED)
	this.Opref[0x78] = NewOp6502("SEI", "implied", 0x78, 1, 2, IMPLIED, MODE_IMPLIED, SEI)
	this.Opref[0x85] = NewOp6502("STA oper", "zeropage", 0x85, 2, 3, ZEROPAGE, MODE_ZEROPAGE, STA)
	this.Opref[0x95] = NewOp6502("STA oper,X", "zeropage,X", 0x95, 2, 4, ZEROPAGE_X, MODE_ZEROPAGE_X, STA)
	this.Opref[0x8D] = NewOp6502("STA oper", "absolute", 0x8D, 3, 4, ABSOLUTE, MODE_ABSOLUTE, STA)
	this.Opref[0x9D] = NewOp6502("STA oper,X", "absolute,X", 0x9D, 3, 5, ABSOLUTE_X_WRITE, MODE_ABSOLUTE_X, STA)
	this.Opref[0x99] = NewOp6502("STA oper,Y", "absolute,Y", 0x99, 3, 5, ABSOLUTE_Y_WRITE, MODE_ABSOLUTE_Y, STA)
	this.Opref[0x81] = NewOp6502("STA (oper,X)", "(indirect,X)", 0x81, 2, 6, INDIRECT_ZP_X, MODE_INDIRECT_ZP_X, STA)
	this.Opref[0x91] = NewOp6502("STA (oper),Y", "(indirect),Y", 0x91, 2, 6, INDIRECT_ZP_Y_WRITE, MODE_INDIRECT_ZP_Y, STA)
	this.Opref[0x86] = NewOp6502("STX oper", "zeropage", 0x86, 2, 3, ZEROPAGE, MODE_ZEROPAGE, STX)
	this.Opref[0x96] = NewOp6502("STX oper,Y", "zeropage,Y", 0x96, 2, 4, ZEROPAGE_Y, MODE_ZEROPAGE_Y, STX)
	this.Opref[0x8E] = NewOp6502("STX oper", "absolute", 0x8E, 3, 4, ABSOLUTE, MODE_ABSOLUTE, STX)
	this.Opref[0x84] = NewOp6502("STY oper", "zeropage", 0x84, 2, 3, ZEROPAGE, MODE_ZEROPAGE, STY)
	this.Opref[0x94] = NewOp6502("STY oper,X", "zeropage,X", 0x94, 2, 4, ZEROPAGE_X, MODE_ZEROPAGE_X, STY)
	this.Opref[0x8C] = NewOp6502("STY oper", "absolute", 0x8C, 3, 4, ABSOLUTE, MODE_ABSOLUTE, STY)
	this.Opref[0xAA] = NewOp6502("TAX", "implied", 0xAA, 1, 2, IMPLIED, MODE_IMPLIED, TAX)
	this.Opref[0xA8] = NewOp6502("TAY", "implied", 0xA8, 1, 2, IMPLIED, MODE_IMPLIED, TAY)
	this.Opref[0xBA] = NewOp6502("TSX", "implied", 0xBA, 1, 2, IMPLIED, MODE_IMPLIED, TSX)
	this.Opref[0x8A] = NewOp6502("TXA", "implied", 0x8A, 1, 2, IMPLIED, MODE_IMPLIED, TXA)
	this.Opref[0x9A] = NewOp6502("TXS", "implied", 0x9A, 1, 2, IMPLIED, MODE_IMPLIED, TXS)
	this.Opref[0x98] = NewOp6502("TYA", "implied", 0x98, 1, 2, IMPLIED, MODE_IMPLIED, TYA)

	// Undocumented opcode - DOP (double NOP)
	this.Opref[0x04] = NewOp6502("dop #oper", "immediate", 0x04, 2, 3, IMMEDIATE, MODE_IMMEDIATE, NOP)

	this.Opref[0x44] = NewOp6502("dop #oper", "immediate", 0x44, 2, 3, IMMEDIATE, MODE_IMMEDIATE, NOP)
	this.Opref[0x54] = NewOp6502("dop #oper", "immediate", 0x54, 2, 4, IMMEDIATE, MODE_IMMEDIATE, NOP)
	this.Opref[0x82] = NewOp6502("dop #oper", "immediate", 0x82, 2, 2, IMMEDIATE, MODE_IMMEDIATE, NOP)

	this.Opref[0xc2] = NewOp6502("dop #oper", "immediate", 0xc2, 2, 2, IMMEDIATE, MODE_IMMEDIATE, NOP)
	this.Opref[0xd4] = NewOp6502("dop #oper", "immediate", 0xd4, 2, 4, IMMEDIATE, MODE_IMMEDIATE, NOP)
	this.Opref[0xe2] = NewOp6502("dop #oper", "immediate", 0xe2, 2, 2, IMMEDIATE, MODE_IMMEDIATE, NOP)
	this.Opref[0xf4] = NewOp6502("dop #oper", "immediate", 0xf4, 2, 4, IMMEDIATE, MODE_IMMEDIATE, NOP)

	// SAX
	this.Opref[0x9F] = NewOp6502("sax oper,Y", "oper,Y", 0x9F, 3, 5, ABSOLUTE_Y, MODE_ABSOLUTE_Y, SAX)
	this.Opref[0x93] = NewOp6502("sax (oper),Y", "(oper),Y", 0x93, 2, 6, INDIRECT_ZP_Y, MODE_INDIRECT_ZP_Y, SAX)

	// AXA
	this.Opref[0x8B] = NewOp6502("axa #oper", "#oper", 0x8B, 2, 2, IMMEDIATE, MODE_IMMEDIATE, AXA)

	// SBX
	this.Opref[0xCB] = NewOp6502("sbx #oper", "#oper", 0xCB, 2, 2, IMMEDIATE, MODE_IMMEDIATE, SBX)

	// SXA
	this.Opref[0x83] = NewOp6502("sxa (oper,X)", "(oper,X)", 0x83, 2, 6, INDIRECT_ZP_X, MODE_INDIRECT_ZP_X, SXA)
	this.Opref[0x87] = NewOp6502("sxa oper", "oper", 0x87, 2, 3, ZEROPAGE, MODE_ZEROPAGE, SXA)
	this.Opref[0x8F] = NewOp6502("sxa oper", "oper", 0x8F, 3, 4, ZEROPAGE, MODE_ZEROPAGE, SXA)
	this.Opref[0x93] = NewOp6502("sxa (oper),Y", "(oper),Y", 0x93, 2, 6, INDIRECT_ZP_Y, MODE_INDIRECT_ZP_Y, SXA)
	this.Opref[0x97] = NewOp6502("sxa oper,Y", "oper,Y", 0x97, 2, 4, ZEROPAGE_Y, MODE_ZEROPAGE_Y, SXA)

	// NOP 1A, 3A, 5A, 7A, DA, FA.

	// NOP 1A, 3A, 5A, 7A, DA, FA.
	//this.Opref[0x0C] = NewOp6502("nop oper", "oper", 0x0C, 3, 4, ABSOLUTE, MODE_ABSOLUTE, NOP)
	this.Opref[0x5C] = NewOp6502("nop oper", "oper", 0x5C, 3, 8, ABSOLUTE, MODE_ABSOLUTE, NOP)
	this.Opref[0xDC] = NewOp6502("nop oper", "oper", 0xDC, 3, 4, ABSOLUTE, MODE_ABSOLUTE, NOP)
	this.Opref[0xFC] = NewOp6502("nop oper", "oper", 0xFC, 3, 4, ABSOLUTE, MODE_ABSOLUTE, NOP)

	// SLO
	this.Opref[0x0F] = NewOp6502("slo oper", "oper", 0x0F, 3, 6, ABSOLUTE, MODE_ABSOLUTE, SLO)
	this.Opref[0x1F] = NewOp6502("slo oper,X", "oper,X", 0x1F, 3, 7, ABSOLUTE_X, MODE_ABSOLUTE_X, SLO)
	this.Opref[0x1B] = NewOp6502("slo oper,Y", "oper,Y", 0x1B, 3, 7, ABSOLUTE_Y, MODE_ABSOLUTE_Y, SLO)
	this.Opref[0x07] = NewOp6502("slo oper", "oper", 0x07, 2, 5, ZEROPAGE, MODE_ZEROPAGE, SLO)
	this.Opref[0x17] = NewOp6502("slo oper,X", "oper,X", 0x17, 2, 6, ZEROPAGE_X, MODE_ZEROPAGE_X, SLO)
	this.Opref[0x03] = NewOp6502("slo (oper,X)", "(oper,X)", 0x03, 2, 8, INDIRECT_ZP_X, MODE_INDIRECT_ZP_X, SLO)
	this.Opref[0x13] = NewOp6502("slo (oper),Y", "(oper),Y", 0x13, 2, 8, INDIRECT_ZP_Y, MODE_INDIRECT_ZP_Y, SLO)

	// LXA
	this.Opref[0xAF] = NewOp6502("lxa oper", "oper", 0xAF, 3, 4, ABSOLUTE, MODE_ABSOLUTE, LXA)
	this.Opref[0xBF] = NewOp6502("lxa oper,Y", "oper,Y", 0xBF, 2, 4, ABSOLUTE_Y, MODE_ABSOLUTE_Y, LXA)
	this.Opref[0xA7] = NewOp6502("lxa oper", "oper", 0xA7, 2, 3, ZEROPAGE, MODE_ZEROPAGE, LXA)
	this.Opref[0xB7] = NewOp6502("lxa oper,Y", "oper,Y", 0xB7, 2, 4, ZEROPAGE_Y, MODE_ZEROPAGE_Y, LXA)
	this.Opref[0xA3] = NewOp6502("lxa (oper,X)", "(oper,X)", 0xA3, 2, 6, INDIRECT_ZP_X, MODE_INDIRECT_ZP_X, LXA)
	this.Opref[0xB3] = NewOp6502("lxa (oper),Y", "(oper),Y", 0xB3, 2, 5, INDIRECT_ZP_Y, MODE_INDIRECT_ZP_Y, LXA)

	// SRE
	this.Opref[0x4F] = NewOp6502("sre oper", "oper", 0x4F, 3, 6, ABSOLUTE, MODE_ABSOLUTE, SRE)
	this.Opref[0x5F] = NewOp6502("sre oper,X", "oper,X", 0x5F, 3, 7, ABSOLUTE_X, MODE_ABSOLUTE_X, SRE)
	this.Opref[0x5B] = NewOp6502("sre oper,Y", "oper,Y", 0x5B, 3, 7, ABSOLUTE_Y, MODE_ABSOLUTE_Y, SRE)
	this.Opref[0x47] = NewOp6502("sre oper", "oper", 0x47, 2, 5, ZEROPAGE, MODE_ZEROPAGE, SRE)
	this.Opref[0x57] = NewOp6502("sre oper,X", "oper,X", 0x57, 2, 6, ZEROPAGE_X, MODE_ZEROPAGE_X, SRE)
	this.Opref[0x43] = NewOp6502("sre (oper,X)", "(oper,X)", 0x43, 2, 8, INDIRECT_ZP_X, MODE_INDIRECT_ZP_X, SRE)
	this.Opref[0x53] = NewOp6502("sre (oper),Y", "(oper),Y", 0x53, 2, 8, INDIRECT_ZP_Y, MODE_INDIRECT_ZP_Y, SRE)

	// RRA
	this.Opref[0x6F] = NewOp6502("rra oper", "oper", 0x6F, 3, 6, ABSOLUTE, MODE_ABSOLUTE, RRA)
	this.Opref[0x7F] = NewOp6502("rra oper,X", "oper,X", 0x7F, 3, 7, ABSOLUTE_X, MODE_ABSOLUTE_X, RRA)
	this.Opref[0x7B] = NewOp6502("rra oper,Y", "oper,Y", 0x7B, 3, 7, ABSOLUTE_Y, MODE_ABSOLUTE_Y, RRA)
	this.Opref[0x67] = NewOp6502("rra oper", "oper", 0x67, 2, 5, ZEROPAGE, MODE_ZEROPAGE, RRA)
	this.Opref[0x77] = NewOp6502("rra oper,X", "oper,X", 0x77, 2, 6, ZEROPAGE_X, MODE_ZEROPAGE_X, RRA)
	this.Opref[0x63] = NewOp6502("rra (oper,X)", "(oper,X)", 0x63, 2, 8, INDIRECT_ZP_X, MODE_INDIRECT_ZP_X, RRA)
	this.Opref[0x73] = NewOp6502("rra (oper),Y", "(oper),Y", 0x73, 2, 8, INDIRECT_ZP_Y, MODE_INDIRECT_ZP_Y, RRA)

	// RLA
	this.Opref[0x2F] = NewOp6502("rla oper", "oper", 0x2F, 3, 6, ABSOLUTE, MODE_ABSOLUTE, RLA)
	this.Opref[0x3F] = NewOp6502("rla oper,X", "oper,X", 0x3F, 3, 7, ABSOLUTE_X, MODE_ABSOLUTE_X, RLA)
	this.Opref[0x3B] = NewOp6502("rla oper,Y", "oper,Y", 0x3B, 3, 7, ABSOLUTE_Y, MODE_ABSOLUTE_Y, RLA)
	this.Opref[0x27] = NewOp6502("rla oper", "oper", 0x27, 2, 5, ZEROPAGE, MODE_ZEROPAGE, RLA)
	this.Opref[0x37] = NewOp6502("rla oper,X", "oper,X", 0x37, 2, 6, ZEROPAGE_X, MODE_ZEROPAGE_X, RLA)
	this.Opref[0x23] = NewOp6502("rla (oper,X)", "(oper,X)", 0x23, 2, 8, INDIRECT_ZP_X, MODE_INDIRECT_ZP_X, RLA)
	this.Opref[0x33] = NewOp6502("rla (oper),Y", "(oper),Y", 0x33, 2, 8, INDIRECT_ZP_Y, MODE_INDIRECT_ZP_Y, RLA)

	// INS
	this.Opref[0xEF] = NewOp6502("ins oper", "oper", 0xEF, 3, 6, ABSOLUTE, MODE_ABSOLUTE, INS)
	this.Opref[0xFF] = NewOp6502("ins oper,X", "oper,X", 0xFF, 3, 7, ABSOLUTE_X, MODE_ABSOLUTE_X, INS)
	this.Opref[0xFB] = NewOp6502("ins oper,Y", "oper,Y", 0xFB, 3, 7, ABSOLUTE_Y, MODE_ABSOLUTE_Y, INS)
	this.Opref[0xE7] = NewOp6502("ins oper", "oper", 0xE7, 2, 5, ZEROPAGE, MODE_ZEROPAGE, INS)
	this.Opref[0xF7] = NewOp6502("ins oper,X", "oper,X", 0xF7, 2, 6, ZEROPAGE_X, MODE_ZEROPAGE_X, INS)
	this.Opref[0xE3] = NewOp6502("ins (oper,X)", "(oper,X)", 0xE3, 2, 8, INDIRECT_ZP_X, MODE_INDIRECT_ZP_X, INS)
	this.Opref[0xF3] = NewOp6502("ins (oper),Y", "(oper),Y", 0xF3, 2, 8, INDIRECT_ZP_Y, MODE_INDIRECT_ZP_Y, INS)

	// ANC
	this.Opref[0x0B] = NewOp6502("anc #oper", "#oper", 0x0B, 2, 2, IMMEDIATE, MODE_IMMEDIATE, ANC)
	this.Opref[0x2B] = NewOp6502("anc #oper", "#oper", 0x2B, 2, 2, IMMEDIATE, MODE_IMMEDIATE, ANC)

	// 02, 12, 22, 32, 42, 52, 62, 72, 92, B2, D2, F2 -> NOP
	this.Opref[0x02] = NewOp6502("hlt", "implied", 0x02, 2, 2, IMPLIED, MODE_IMPLIED, NOP)
	this.Opref[0x12] = NewOp6502("hlt", "implied", 0x12, 1, 2, IMPLIED, MODE_IMPLIED, NOP)
	this.Opref[0x22] = NewOp6502("hlt", "implied", 0x22, 2, 2, IMPLIED, MODE_IMPLIED, NOP)
	this.Opref[0x32] = NewOp6502("hlt", "implied", 0x32, 1, 2, IMPLIED, MODE_IMPLIED, NOP)
	this.Opref[0x42] = NewOp6502("hlt", "implied", 0x42, 2, 2, IMPLIED, MODE_IMPLIED, NOP)
	this.Opref[0x52] = NewOp6502("hlt", "implied", 0x52, 1, 2, IMPLIED, MODE_IMPLIED, NOP)
	this.Opref[0x62] = NewOp6502("hlt", "implied", 0x62, 2, 2, IMPLIED, MODE_IMPLIED, NOP)
	this.Opref[0x72] = NewOp6502("hlt", "implied", 0x72, 1, 2, IMPLIED, MODE_IMPLIED, NOP)
	this.Opref[0x92] = NewOp6502("hlt", "implied", 0x92, 1, 2, IMPLIED, MODE_IMPLIED, NOP)
	this.Opref[0xB2] = NewOp6502("hlt", "implied", 0xB2, 1, 2, IMPLIED, MODE_IMPLIED, NOP)
	this.Opref[0xD2] = NewOp6502("hlt", "implied", 0xD2, 1, 2, IMPLIED, MODE_IMPLIED, NOP)
	this.Opref[0xF2] = NewOp6502("hlt", "implied", 0xF2, 1, 2, IMPLIED, MODE_IMPLIED, NOP)

	// DCP
	this.Opref[0xCF] = NewOp6502("dcp oper", "oper", 0xCF, 3, 6, ABSOLUTE, MODE_ABSOLUTE, DCP)
	this.Opref[0xDF] = NewOp6502("dcp oper,X", "oper,X", 0xDF, 3, 7, ABSOLUTE_X, MODE_ABSOLUTE_X, DCP)
	this.Opref[0xDB] = NewOp6502("dcp oper,Y", "oper,Y", 0xDB, 3, 7, ABSOLUTE_Y, MODE_ABSOLUTE_Y, DCP)
	this.Opref[0xC7] = NewOp6502("dcp oper", "oper", 0xC7, 2, 5, ZEROPAGE, MODE_ZEROPAGE, DCP)
	this.Opref[0xD7] = NewOp6502("dcp oper,X", "oper,X", 0xD7, 2, 6, ZEROPAGE_X, MODE_ZEROPAGE_X, DCP)
	this.Opref[0xC3] = NewOp6502("dcp (oper,X)", "(oper,X)", 0xC3, 2, 8, INDIRECT_ZP_X, MODE_INDIRECT_ZP_X, DCP)
	this.Opref[0xD3] = NewOp6502("dcp (oper),Y", "(oper),Y", 0xD3, 2, 8, INDIRECT_ZP_Y, MODE_INDIRECT_ZP_Y, DCP)

	// 65C02 addressing mode INDIRECT ZEROPAGE
	/*
		OP LEN CYC MODE FLAGS    SYNTAX
		-- --- --- ---- ------   ------
		72 2   5 a (zp) NV....ZC ADC ($12)
		32 2   5   (zp) N.....Z. AND ($12)
		D2 2   5   (zp) N.....ZC CMP ($12)
		52 2   5   (zp) N.....Z. EOR ($12)
		B2 2   5   (zp) N.....Z. LDA ($12)
		12 2   5   (zp) N.....Z. ORA ($12)
		F2 2   5 a (zp) NV....ZC SBC ($12)
		92 2   5   (zp) ........ STA ($12)
	*/
	this.Opref[0x72] = NewOp6502("ADC (oper)", "oper", 0x72, 2, 5, INDIRECT_ZP, MODE_INDIRECT_ZP, ADC)
	this.Opref[0x32] = NewOp6502("AND (oper)", "oper", 0x32, 2, 5, INDIRECT_ZP, MODE_INDIRECT_ZP, AND)
	this.Opref[0xD2] = NewOp6502("CMP (oper)", "oper", 0xD2, 2, 5, INDIRECT_ZP, MODE_INDIRECT_ZP, CMP)
	this.Opref[0x52] = NewOp6502("EOR (oper)", "oper", 0x52, 2, 5, INDIRECT_ZP, MODE_INDIRECT_ZP, EOR)
	this.Opref[0xB2] = NewOp6502("LDA (oper)", "oper", 0xB2, 2, 5, INDIRECT_ZP, MODE_INDIRECT_ZP, LDA)
	this.Opref[0x12] = NewOp6502("ORA (oper)", "oper", 0x12, 2, 5, INDIRECT_ZP, MODE_INDIRECT_ZP, ORA)
	this.Opref[0xF2] = NewOp6502("SBC (oper)", "oper", 0xF2, 2, 5, INDIRECT_ZP, MODE_INDIRECT_ZP, SBC)
	this.Opref[0x92] = NewOp6502("STA (oper)", "oper", 0x92, 2, 5, INDIRECT_ZP, MODE_INDIRECT_ZP, STA)

	// BIT additional
	this.Opref[0x89] = NewOp6502("BIT #oper", "immediate", 0x89, 2, 2, IMMEDIATE, MODE_IMMEDIATE, BIT)
	this.Opref[0x34] = NewOp6502("BIT oper,X", "immediate", 0x34, 2, 4, ZEROPAGE_X, MODE_ZEROPAGE_X, BIT)
	this.Opref[0x3C] = NewOp6502("BIT oper,X", "oper", 0x3C, 3, 4, ABSOLUTE_X, MODE_ABSOLUTE_X, BIT)

	// DEA, INA
	this.Opref[0x1A] = NewOp6502("INA", "implied", 0x1A, 1, 2, IMPLIED, MODE_IMPLIED, INA)
	this.Opref[0x3A] = NewOp6502("DEA", "implied", 0x3A, 1, 2, IMPLIED, MODE_IMPLIED, DEA)

	// JMP (aaaaa,X)
	this.Opref[0x7C] = NewOp6502("JMP (oper,X)", "oper", 0x7C, 3, 6, INDIRECT_X, MODE_INDIRECT_X, JMP)

	// BRA - branch always
	this.Opref[0x80] = NewOp6502("BRA oper", "immediate", 0x80, 2, 2, RELATIVE, MODE_RELATIVE, BRA)

	/*
		OP LEN CYC MODE FLAGS    SYNTAX
		-- --- --- ---- -----    ------
		DA 1   3   imp  ........ PHX
		5A 1   3   imp  ........ PHY
		FA 1   4   imp  N.....Z. PLX
		7A 1   4   imp  N.....Z. PLY
	*/
	this.Opref[0x5A] = NewOp6502("PHY", "implied", 0x5A, 1, 3, IMPLIED, MODE_IMPLIED, PHY)
	this.Opref[0x7A] = NewOp6502("PLY", "implied", 0x7A, 1, 4, IMPLIED, MODE_IMPLIED, PLY)
	this.Opref[0xDA] = NewOp6502("PHX", "implied", 0xDA, 1, 3, IMPLIED, MODE_IMPLIED, PHX)
	this.Opref[0xFA] = NewOp6502("PLX", "implied", 0xFA, 1, 4, IMPLIED, MODE_IMPLIED, PLX)

	/*
		OP LEN CYC MODE  FLAGS    SYNTAX
		-- --- --- ----  -----    ------
		64 2   3   zp    ........ STZ $12
		74 2   4   zp,X  ........ STZ $12,X
		9C 3   4   abs   ........ STZ $3456
		9E 3   5   abs,X ........ STZ $3456,X
	*/
	this.Opref[0x64] = NewOp6502("STZ oper", "immediate", 0x64, 2, 3, ZEROPAGE, MODE_ZEROPAGE, STZ)
	this.Opref[0x74] = NewOp6502("STZ oper,X", "immediate", 0x74, 2, 4, ZEROPAGE_X, MODE_ZEROPAGE_X, STZ)
	this.Opref[0x9c] = NewOp6502("STZ oper", "immediate", 0x9c, 2, 4, ABSOLUTE, MODE_ABSOLUTE, STZ)
	this.Opref[0x9e] = NewOp6502("STZ oper,X", "immediate", 0x9e, 2, 5, ABSOLUTE_X_WRITE, MODE_ABSOLUTE_X, STZ)

	// TRB, TSB
	/*
		OP LEN CYC MODE FLAGS    SYNTAX
		-- --- --- ---- -----    ------
		14 2   5   zp   ......Z. TRB $12
		1C 3   6   abs  ......Z. TRB $3456
	*/
	this.Opref[0x14] = NewOp6502("TRB oper", "immediate", 0x14, 2, 5, ZEROPAGE, MODE_ZEROPAGE, TRB)
	this.Opref[0x1C] = NewOp6502("TRB oper", "oper", 0x1C, 3, 6, ABSOLUTE, MODE_ABSOLUTE, TRB)
	this.Opref[0x04] = NewOp6502("TSB oper", "immediate", 0x04, 2, 5, ZEROPAGE, MODE_ZEROPAGE, TSB)
	this.Opref[0x0C] = NewOp6502("TSB oper", "oper", 0x0C, 3, 6, ABSOLUTE, MODE_ABSOLUTE, TSB)

	// BBRn
	this.Opref[0x0F] = NewOp6502("BBR0 oper", "oper", 0x0F, 3, 5, ABSOLUTE, MODE_ABSOLUTE, BBR0)
	this.Opref[0x1F] = NewOp6502("BBR1 oper", "oper", 0x1F, 3, 5, ABSOLUTE, MODE_ABSOLUTE, BBR1)
	this.Opref[0x2F] = NewOp6502("BBR2 oper", "oper", 0x2F, 3, 5, ABSOLUTE, MODE_ABSOLUTE, BBR2)
	this.Opref[0x3F] = NewOp6502("BBR3 oper", "oper", 0x3F, 3, 5, ABSOLUTE, MODE_ABSOLUTE, BBR3)
	this.Opref[0x4F] = NewOp6502("BBR4 oper", "oper", 0x4F, 3, 5, ABSOLUTE, MODE_ABSOLUTE, BBR4)
	this.Opref[0x5F] = NewOp6502("BBR5 oper", "oper", 0x5F, 3, 5, ABSOLUTE, MODE_ABSOLUTE, BBR5)
	this.Opref[0x6F] = NewOp6502("BBR6 oper", "oper", 0x6F, 3, 5, ABSOLUTE, MODE_ABSOLUTE, BBR6)
	this.Opref[0x7F] = NewOp6502("BBR7 oper", "oper", 0x7F, 3, 5, ABSOLUTE, MODE_ABSOLUTE, BBR7)

	// BBSn
	this.Opref[0x8F] = NewOp6502("BBS0 oper", "oper", 0x8F, 3, 5, ABSOLUTE, MODE_ABSOLUTE, BBS0)
	this.Opref[0x9F] = NewOp6502("BBS1 oper", "oper", 0x9F, 3, 5, ABSOLUTE, MODE_ABSOLUTE, BBS1)
	this.Opref[0xAF] = NewOp6502("BBS2 oper", "oper", 0xAF, 3, 5, ABSOLUTE, MODE_ABSOLUTE, BBS2)
	this.Opref[0xBF] = NewOp6502("BBS3 oper", "oper", 0xBF, 3, 5, ABSOLUTE, MODE_ABSOLUTE, BBS3)
	this.Opref[0xCF] = NewOp6502("BBS4 oper", "oper", 0xCF, 3, 5, ABSOLUTE, MODE_ABSOLUTE, BBS4)
	this.Opref[0xDF] = NewOp6502("BBS5 oper", "oper", 0xDF, 3, 5, ABSOLUTE, MODE_ABSOLUTE, BBS5)
	this.Opref[0xEF] = NewOp6502("BBS6 oper", "oper", 0xEF, 3, 5, ABSOLUTE, MODE_ABSOLUTE, BBS6)
	this.Opref[0xFF] = NewOp6502("BBS7 oper", "oper", 0xFF, 3, 5, ABSOLUTE, MODE_ABSOLUTE, BBS7)

	// RMBn
	this.Opref[0x07] = NewOp6502("RMB0 oper", "oper", 0x07, 2, 5, ZEROPAGE, MODE_ZEROPAGE, RMB0)
	this.Opref[0x17] = NewOp6502("RMB1 oper", "oper", 0x17, 2, 5, ZEROPAGE, MODE_ZEROPAGE, RMB1)
	this.Opref[0x27] = NewOp6502("RMB2 oper", "oper", 0x27, 2, 5, ZEROPAGE, MODE_ZEROPAGE, RMB2)
	this.Opref[0x37] = NewOp6502("RMB3 oper", "oper", 0x37, 2, 5, ZEROPAGE, MODE_ZEROPAGE, RMB3)
	this.Opref[0x47] = NewOp6502("RMB4 oper", "oper", 0x47, 2, 5, ZEROPAGE, MODE_ZEROPAGE, RMB4)
	this.Opref[0x57] = NewOp6502("RMB5 oper", "oper", 0x57, 2, 5, ZEROPAGE, MODE_ZEROPAGE, RMB5)
	this.Opref[0x67] = NewOp6502("RMB6 oper", "oper", 0x67, 2, 5, ZEROPAGE, MODE_ZEROPAGE, RMB6)
	this.Opref[0x77] = NewOp6502("RMB7 oper", "oper", 0x77, 2, 5, ZEROPAGE, MODE_ZEROPAGE, RMB7)

	// SMBn
	this.Opref[0x87] = NewOp6502("SMB0 oper", "oper", 0x87, 2, 5, ZEROPAGE, MODE_ZEROPAGE, SMB0)
	this.Opref[0x97] = NewOp6502("SMB1 oper", "oper", 0x97, 2, 5, ZEROPAGE, MODE_ZEROPAGE, SMB1)
	this.Opref[0xA7] = NewOp6502("SMB2 oper", "oper", 0xA7, 2, 5, ZEROPAGE, MODE_ZEROPAGE, SMB2)
	this.Opref[0xB7] = NewOp6502("SMB3 oper", "oper", 0xB7, 2, 5, ZEROPAGE, MODE_ZEROPAGE, SMB3)
	this.Opref[0xC7] = NewOp6502("SMB4 oper", "oper", 0xC7, 2, 5, ZEROPAGE, MODE_ZEROPAGE, SMB4)
	this.Opref[0xD7] = NewOp6502("SMB5 oper", "oper", 0xD7, 2, 5, ZEROPAGE, MODE_ZEROPAGE, SMB5)
	this.Opref[0xE7] = NewOp6502("SMB6 oper", "oper", 0xE7, 2, 5, ZEROPAGE, MODE_ZEROPAGE, SMB6)
	this.Opref[0xF7] = NewOp6502("SMB7 oper", "oper", 0xF7, 2, 5, ZEROPAGE, MODE_ZEROPAGE, SMB7)

	this.Opref[0x6B] = NewOp6502("ARR #oper", "oper", 0x6B, 2, 2, IMMEDIATE, MODE_IMMEDIATE, ARR)
}
