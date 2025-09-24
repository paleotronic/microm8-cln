package mos6502

type Event6502 interface {
	ProcessCPUEvent(MEM []int, opcode int, value int)
}
