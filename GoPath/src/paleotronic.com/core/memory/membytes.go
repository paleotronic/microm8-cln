package memory

// MemBytes represents a slice of bytes, with an abstract backing store
type MemBytes interface {
	Set(index int, b byte)
	Get(index int) byte
	Slice(start, end int) MemBytes
	Bytes() []byte
	ByteSlice(start, end int) []byte
	Len() int
	Write(index int, data []byte)
}
