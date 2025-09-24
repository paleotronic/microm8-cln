package memory

// MemMicroM8 is a slice based implementation of MemBytes
type MemMicroM8 struct {
	mm     *MemoryMap
	slotid int
	base   int // important - this is the 8 bit offset into memory... (address*8)
	size   int
	mcb    *MemoryControlBlock
}

func (m *MemMicroM8) Set(index int, b byte) {
	if index >= m.size || index < 0 {
		return
	}

	apos := m.base + index
	bbIndex := apos / 8
	bbOffs := apos % 8
	bbShift := uint(bbOffs) * 8
	setMask := uint64(0xff) << bbShift
	clrMask := setMask ^ 0xffffffffffffffff
	d := m.mcb.Read(bbIndex) //m.mm.ReadInterpreterMemorySilent(m.slotid, bbIndex)
	d &= clrMask
	d |= (uint64(b) << bbShift)
	m.mcb.Write(bbIndex, d)
}

func (m *MemMicroM8) Get(index int) byte {
	if index >= m.size || index < 0 {
		return 0x00
	}
	apos := m.base + index
	bbIndex := apos / 8
	bbOffs := apos % 8
	bbShift := uint(bbOffs) * 8
	d := m.mcb.Read(bbIndex) //m.mm.ReadInterpreterMemorySilent(m.slotid, bbIndex)
	return byte((d >> bbShift) & 0xff)
}

func (m *MemMicroM8) Slice(start, end int) MemBytes {
	if end == -1 {
		end = m.Len()
	}
	if start < 0 {
		start = 0
	}
	return &MemMicroM8{
		mm:     m.mm,
		slotid: m.slotid,
		base:   m.base + start,
		size:   end - start,
		mcb:    m.mcb,
	}
}

func (m *MemMicroM8) Bytes() []byte {
	return m.ByteSlice(-1, -1)
}

func (m *MemMicroM8) ByteSlice(start, end int) []byte {
	if end == -1 {
		end = m.Len()
	}
	if start < 0 {
		start = 0
	}
	out := make([]byte, end-start)
	for i, _ := range out {
		out[i] = m.Get(start + i)
	}
	return out
}

func (m *MemMicroM8) Len() int {
	return m.mcb.Size*8 - m.base
}

func (m *MemMicroM8) Write(index int, data []byte) {
	for i, v := range data {
		if index+i < m.size {
			m.Set(index+i, v)
		}
	}
}

func NewMemMicroM8(mm *MemoryMap, slotid int, base int, mcb *MemoryControlBlock) *MemMicroM8 {
	//fmt.Printf("Slot Memory buffer @%d, size %d\n", base, size/8+1)
	return &MemMicroM8{
		mm:     mm,
		slotid: slotid,
		base:   base,
		size:   mcb.Size*8 - base, // size in bytes based upon 64 bit mcb size minus base byte offset
		mcb:    mcb,
	}
}
