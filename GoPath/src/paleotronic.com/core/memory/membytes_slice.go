package memory

// MemByteSlice is a slice based implementation of MemBytes
type MemByteSlice struct {
	data []byte
}

func (m *MemByteSlice) Set(index int, b byte) {
	m.data[index] = b
}

func (m *MemByteSlice) Get(index int) byte {
	return m.data[index]
}

func (m *MemByteSlice) Slice(start, end int) MemBytes {
	if end == -1 {
		end = m.Len()
	}
	if start < 0 {
		start = 0
	}
	return &MemByteSlice{data: m.data[start:end]}
}

func (m *MemByteSlice) Bytes() []byte {
	return m.data
}

func (m *MemByteSlice) ByteSlice(start, end int) []byte {
	if end == -1 {
		end = m.Len()
	}
	if start < 0 {
		start = 0
	}
	return m.data[start:end]
}

func (m *MemByteSlice) Len() int {
	return len(m.data)
}

func (m *MemByteSlice) Write(index int, data []byte) {
	for i, v := range data {
		if index+i < len(m.data) {
			m.data[index+i] = v
		}
	}
}

func NewMemByteSlice(size int) *MemByteSlice {
	return &MemByteSlice{
		data: make([]byte, size),
	}
}
