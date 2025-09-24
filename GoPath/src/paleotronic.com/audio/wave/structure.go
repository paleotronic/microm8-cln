package audio

type RIFFChunk struct {
	ChunkID   [4]byte // RIFF
	ChunkSize uint32  // Size 4+n
}
