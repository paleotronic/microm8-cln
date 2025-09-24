package types

import (
	"paleotronic.com/core/memory"
	"paleotronic.com/utils"
)

type FuncPtr5b struct {
	hi, lo   byte
	vhi, vlo byte
	fb       byte
}

func NewFuncPtr5b(fb byte, addr int, vaddr int) *FuncPtr5b {
	this := &FuncPtr5b{}
	this.SetPointer(addr)
	this.SetArgPointer(addr)
	this.fb = fb

	return this
}

func (i *FuncPtr5b) SetPointer(v int) {
	u16 := uint16(v)
	i.lo = byte(u16 & 0xff)
	i.hi = byte((u16 >> 8) & 0xff)
}

func (i *FuncPtr5b) GetPointer() int {
	v := int(i.hi<<8) | int(i.lo)
	return v
}

func (i *FuncPtr5b) SetArgPointer(v int) {
	u16 := uint16(v)
	i.vlo = byte(u16 & 0xff)
	i.vhi = byte((u16 >> 8) & 0xff)
}

func (i *FuncPtr5b) GetArgPointer() int {
	v := int(i.vhi<<8) | int(i.vlo)
	return v
}

func (i FuncPtr5b) GetFirstByte() byte {

	return i.fb

}

func (i FuncPtr5b) SetFirstByte(l byte) {

	i.fb = l

}

func (i FuncPtr5b) String() string {

	return utils.IntToStr(int(i.GetPointer()))

}

func (i *FuncPtr5b) WriteMemory(mm *memory.MemoryMap, index int, address int) {
	mm.WriteInterpreterMemory(index, address+0, uint64(i.hi))
	mm.WriteInterpreterMemory(index, address+1, uint64(i.lo))
	mm.WriteInterpreterMemory(index, address+2, uint64(i.vhi))
	mm.WriteInterpreterMemory(index, address+3, uint64(i.vlo))
	mm.WriteInterpreterMemory(index, address+4, uint64(i.fb))
}

func (i *FuncPtr5b) ReadMemory(mm *memory.MemoryMap, index int, address int) {
	i.hi = byte(mm.ReadInterpreterMemory(index, address+0))
	i.lo = byte(mm.ReadInterpreterMemory(index, address+1))
	i.vhi = byte(mm.ReadInterpreterMemory(index, address+2))
	i.vlo = byte(mm.ReadInterpreterMemory(index, address+3))
	i.fb = byte(mm.ReadInterpreterMemory(index, address+4))
}
