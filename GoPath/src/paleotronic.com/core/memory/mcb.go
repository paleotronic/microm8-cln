package memory

import (
	//	"paleotronic.com/fmt"
	"sync"
)

type MemoryControlBlock struct {
	Data   [][]uint64
	GStart []int
	Size   int
	mm     *MemoryMap
	Index  int
	mutex  sync.Mutex
	UseMM  bool
}

type MemoryEventProcessor interface {
	ProcessEvent(name string, addr int, value *uint64, action MemoryAction) (bool, bool)
}

func NewMemoryControlBlock(m *MemoryMap, index int, useMM bool) *MemoryControlBlock {
	this := &MemoryControlBlock{Data: make([][]uint64, 0), Size: 0, mm: m, UseMM: useMM, Index: index}
	return this
}

func (mcb *MemoryControlBlock) GetMM() *MemoryMap {

	return mcb.mm

}

func (mcb *MemoryControlBlock) Add(chunk []uint64, gstart int) {
	gstart = gstart % OCTALYZER_INTERPRETER_SIZE
	mcb.Data = append(mcb.Data, chunk)
	mcb.GStart = append(mcb.GStart, gstart)
	mcb.Size += len(chunk)
}

func (mr *MemoryControlBlock) ProcessEvent(name string, addr int, action MemoryAction) {
	panic("this should be overridden")
}

func (mcb *MemoryControlBlock) GetRef(offset int) (int, int) {

	block := 0

	if offset >= mcb.Size {
		return -1, -1
	}

	for offset >= len(mcb.Data[block]) {
		offset -= len(mcb.Data[block])
		block++
	}

	return block, offset
}

func (mcb *MemoryControlBlock) Write(offset int, value uint64) {

	if offset >= mcb.Size || offset < 0 {
		return
	}

	bidx, oidx := mcb.GetRef(offset)
	if bidx == -1 {
		return
	}

	//mcb.mm.mutex.Lock()
	//	mcb.mutex.Lock()
	//	defer mcb.mutex.Unlock()
	ov := mcb.Data[bidx][oidx]

	a := mcb.GStart[bidx] + oidx

	if ov == value && (a%OCTALYZER_INTERPRETER_SIZE)/256 != 192 {
		return
	}

	if mcb.UseMM {
		addr := mcb.GStart[bidx] % OCTALYZER_INTERPRETER_SIZE
		index := mcb.GStart[bidx] / OCTALYZER_INTERPRETER_SIZE
		mcb.mm.WriteInterpreterMemory(index, addr+oidx, value)
		return
	}

	pvalue := mcb.Data[bidx][oidx]
	mcb.Data[bidx][oidx] = value

	//mcb.mm.mutex.Unlock()

	//if mcb.mm.Track[mcb.GStart[0]/OCTALYZER_INTERPRETER_SIZE] {
	mcb.mm.LogMCBWrite(mcb.Index, a, value, pvalue)
	//}

}

func (mcb *MemoryControlBlock) Read(offset int) uint64 {

	if offset >= mcb.Size || offset < 0 {
		return 0
	}

	bidx, oidx := mcb.GetRef(offset)
	if bidx == -1 {
		return 0
	}

	return mcb.Data[bidx][oidx]

}

func (mcb *MemoryControlBlock) ReadSlice(start, end int) []uint64 {

	end = end - 1

	if end >= mcb.Size {
		end = mcb.Size - 1
	}

	sb, so := mcb.GetRef(start)
	eb, eo := mcb.GetRef(end)

	if sb == -1 || eb == -1 {
		return []uint64(nil)
	}

	chunk := make([]uint64, 0)
	//mcb.mm.mutex.Lock()
	//defer mcb.mm.mutex.Unlock()
	//	mcb.mutex.Lock()
	//	defer mcb.mutex.Unlock()
	for block := sb; block <= eb; block++ {
		// Default to take whole block
		a := 0
		b := len(mcb.Data[block])
		//
		if block == sb {
			a = so
		}
		if block == eb {
			b = eo + 1
		}

		chunk = append(chunk, mcb.Data[block][a:b]...)
	}

	return chunk

}

func (mcb *MemoryControlBlock) ReadSliceCopy(start, end int) []uint64 {

	end = end - 1

	if end >= mcb.Size {
		end = mcb.Size - 1
	}

	sb, so := mcb.GetRef(start)
	eb, eo := mcb.GetRef(end)

	if sb == -1 || eb == -1 {
		return []uint64(nil)
	}

	chunk := make([]uint64, 0)
	//mcb.mm.mutex.Lock()
	//defer mcb.mm.mutex.Unlock()
	//	mcb.mutex.Lock()
	//	defer mcb.mutex.Unlock()
	for block := sb; block <= eb; block++ {
		// Default to take whole block
		a := 0
		b := len(mcb.Data[block])
		//
		if block == sb {
			a = so
		}
		if block == eb {
			b = eo + 1
		}

		var tmp = make([]uint64, b-a)
		copy(tmp, mcb.Data[block][a:b])

		chunk = append(chunk, tmp...)
	}

	return chunk

}
