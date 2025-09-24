package types

import (
	"paleotronic.com/core/memory"
	"paleotronic.com/utils"
)

type Integer2b struct {
	hi, lo byte
}

func NewInteger2b(v int16) *Integer2b {
	this := &Integer2b{}
	this.SetValue(v)

	return this
}

func (i *Integer2b) SetValue(v int16) {
	u16 := uint16(v)
	i.lo = byte(u16 & 0xff)
	i.hi = byte((u16 >> 8) & 0xff)
}

func (i *Integer2b) GetValue() int16 {
	u16 := (uint16(i.hi) << 8) | uint16(i.lo)
	return int16(u16)
}

func (i Integer2b) String() string {

	return utils.IntToStr(int(i.GetValue()))

}

func (i *Integer2b) WriteMemory(mm *memory.MemoryMap, index int, address int) {
	mm.WriteInterpreterMemory(index, address+0, uint64(i.hi))
	mm.WriteInterpreterMemory(index, address+1, uint64(i.lo))
}

func (i *Integer2b) ReadMemory(mm *memory.MemoryMap, index int, address int) {
	i.hi = byte(mm.ReadInterpreterMemory(index, address+0))
	i.lo = byte(mm.ReadInterpreterMemory(index, address+1))
}
