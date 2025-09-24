package mos6502

type Op6502 struct {
	Description    string
	AddressingMode string
	Opcode         int
	Bytes          int
	Fetch          MODE
	FetchMode      MODEENUM
	Execute        OP
	Cycles         int
}

func NewOp6502(desc string, addressingMode string, opcode int, bytes int, cycles int, fetch MODE, fenum MODEENUM, execute OP) *Op6502 {
	this := &Op6502{}
	this.Description = desc
	this.AddressingMode = addressingMode
	this.Opcode = opcode
	this.Bytes = bytes
	this.Cycles = cycles
	this.Fetch = fetch
	this.FetchMode = fenum
	this.Execute = execute
	return this
}

func (this *Op6502) Do(cpu *Core6502) int {
	// add a read cycle here
	//cpu.FetchBytePCNOP(&cpu.PC)

	return this.Execute(cpu, this)
}

func (this *Op6502) SetDescription(s string) {
	this.Description = s
}

func (this *Op6502) SetOpCode(v int) {
	this.Opcode = v
}

func (this *Op6502) GetDescription() string {
	return this.Description
}

func (this *Op6502) GetOpCode() int {
	return this.Opcode
}

func (this *Op6502) SetAddressingMode(s string) {
	this.AddressingMode = s
}

func (this *Op6502) SetBytes(v int) {
	this.Bytes = v
}

func (this *Op6502) SetCycles(v int) {
	this.Cycles = v
}

func (this *Op6502) GetAddressingMode() string {
	return this.AddressingMode
}

func (this *Op6502) GetBytes() int {
	return this.Bytes
}

func (this *Op6502) GetCycles() int {
	return this.Cycles
}
