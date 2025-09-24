package types

import (
	//	"paleotronic.com/fmt"
	"paleotronic.com/core/memory"
	"paleotronic.com/utils"
)

type StringPtr3b struct {
	length byte
	hi, lo byte
}

func NewStringPtr3b(l byte, addr int) *StringPtr3b {
	this := &StringPtr3b{}
	this.SetPointer(addr)
	this.SetLength(l)

	return this
}

func (i *StringPtr3b) SetPointer(v int) {
	i.lo = byte(v % 256)
	i.hi = byte(v / 256)
}

func (i *StringPtr3b) GetPointer() int {
	v := int(i.hi)*256 + int(i.lo)
	return v
}

func (i *StringPtr3b) GetLength() byte {

	return i.length

}

func (i *StringPtr3b) SetLength(l byte) {

	i.length = l

}

func (i *StringPtr3b) String() string {

	return utils.IntToStr(int(i.GetPointer()))

}

func (i *StringPtr3b) WriteMemory(mm *memory.MemoryMap, index int, address int) {
	mm.WriteInterpreterMemory(index, address+0, uint64(i.length))
	mm.WriteInterpreterMemory(index, address+1, uint64(i.hi))
	mm.WriteInterpreterMemory(index, address+2, uint64(i.lo))
}

func (i *StringPtr3b) ReadMemory(mm *memory.MemoryMap, index int, address int) {
	i.length = byte(mm.ReadInterpreterMemory(index, address+0))
	i.hi = byte(mm.ReadInterpreterMemory(index, address+1))
	i.lo = byte(mm.ReadInterpreterMemory(index, address+2))
}

// Retrieve pointed string
func (i *StringPtr3b) FetchString(mm *memory.MemoryMap, index int) string {

	addr := i.GetPointer()
	size := int(i.GetLength())

	out := ""

	for i := 0; i < size; i++ {
		out += string(rune(mm.ReadInterpreterMemory(index, addr+i)))
	}

	////fmt.Println([]byte(out))

	return out
}

func (i *StringPtr3b) Top() int {
	return i.GetPointer() + int(i.GetLength())
}
