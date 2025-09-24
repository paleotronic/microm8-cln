package woz2

import (
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"io/ioutil"

	"paleotronic.com/core/memory"
	"paleotronic.com/log"
)

const crcPos = 8
const chunkBase = 12

var wozMagic = []byte{0x57, 0x4f, 0x5a, 0x32, 0xff, 0x0a, 0x0d, 0x0a}
var ErrNotValid = errors.New("Not a valid WOZ2 image")
var ErrOutOfData = errors.New("Out of data")

type ChunkID string

const (
	INFOChunk ChunkID = "INFO"
	TMAPChunk ChunkID = "TMAP"
	TRKSChunk ChunkID = "TRKS"
	METAChunk ChunkID = "META"
)

const forceINFOSize = 60
const trackLength = 6656
const bitstreamLength = 6646
const defaultTrackBits = 51200
const defaultTrackBytes = 6400

type WOZ2Image struct {
	Data   memory.MemBytes
	Ptr    int
	Chunks map[ChunkID]*WOZ2Chunk
	// Recognized Chunks
	INFO *WOZ2INFOChunk
	TMAP *WOZ2TMAPChunk
	TRKS *WOZ2TRKSChunk
	// Currently mapped track
	Track         *WOZ2Track
	BitPtr        int
	Modified      bool
	Size          int
	EssentialSize int
}

const trimWoz = false

func NewWOZ2Image(r io.Reader, buffer memory.MemBytes) (*WOZ2Image, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	if len(data) > buffer.Len() {
		//fmt.Println("Using dynamic buffer for WOZ2 as mapped buffer is too small")
		buffer = memory.NewMemByteSlice(len(data))
	}

	length := len(data)
	log.Printf("Initial size = %d bytes", length)
	if trimWoz {
		tmp := memory.NewMemByteSlice(length)
		tmp.Write(0, data)
		img := &WOZ2Image{
			Data: tmp,
			Ptr:  0,
			Size: len(data),
		}
		if !img.IsValid() {
			return nil, ErrNotValid
		}
		img.LoadChunks()
		length = img.EssentialSize
		log.Printf("Essential size = %d bytes", length)
	}

	buffer.Write(0, data[:length])
	img := &WOZ2Image{
		Data: buffer,
		Ptr:  0,
		Size: length,
	}

	if trimWoz {
		// need to update crc32
		img.UpdateCRC32()
	}

	if !img.IsValid() {
		return nil, ErrNotValid
	}
	img.LoadChunks()

	//fmt.Printf("woz2: boot format flag: %s\n", img.INFO.BootSectorFormat())

	//img.DumpGeometry()

	return img, err
}

func (w *WOZ2Image) GetOptimalBitTiming() int {
	return w.INFO.OptimalBitTiming()
}

func (w *WOZ2Image) SetModified(b bool) {
	w.Modified = b
}

func (w *WOZ2Image) IsModified() bool {
	return w.Modified
}

func (w *WOZ2Image) GetData() memory.MemBytes {
	return w.Data
}

func (w *WOZ2Image) GetSize() int {
	return w.Size
}

func (w *WOZ2Image) SetWriteProtected(b bool) {
	w.INFO.SetWriteProtected(b)
}

func (w *WOZ2Image) WriteProtected() bool {
	return w.INFO.WriteProtected()
}

func (w *WOZ2Image) BitCount() uint16 {
	return w.Track.BitCount()
}

func (w *WOZ2Image) ReadBit(ptr int) byte {
	return w.Track.ReadBit(ptr)
}

func (w *WOZ2Image) WriteBit(ptr int, bit byte) {
	w.Track.WriteBit(ptr, bit)
}

func (w *WOZ2Image) TrackOK() bool {
	return w.Track != nil
}

func (w *WOZ2Image) IsValid() bool {
	for i, v := range wozMagic {
		if w.Data.Get(i) != v {
			return false
		}
	}

	// check crc
	currCRC := w.GetEmbeddedCRC()
	newCRC := crc32.ChecksumIEEE(w.Data.ByteSlice(12, w.Size))
	if currCRC != newCRC {
		log.Printf("Checksums differ! header CRC32=%.8x, data CRC32=%.8x")
		return false
	}

	log.Println("Volume OK")

	return true
}

func (w *WOZ2Image) getLE32(index int) uint32 {
	if w.Data.Len()-index < 4 {
		return 0
	}
	return uint32(w.Data.Get(index+0)) |
		(uint32(w.Data.Get(index+1)) << 8) |
		(uint32(w.Data.Get(index+2)) << 16) |
		(uint32(w.Data.Get(index+3)) << 24)
}

func (w *WOZ2Image) GetEmbeddedCRC() uint32 {
	return w.getLE32(crcPos)
}

func (w *WOZ2Image) ResetChunkPtr() {
	w.Ptr = chunkBase
}

func (w *WOZ2Image) movePtr(change int) error {
	n := w.Ptr + change
	if n < 0 || n >= w.Data.Len() {
		return ErrOutOfData
	}
	w.Ptr = n
	return nil
}

func (w *WOZ2Image) hasBytes(count int) bool {
	return w.Data.Len()-w.Ptr >= count
}

func (w *WOZ2Image) ReadChunk() (*WOZ2Chunk, error) {
	if !w.hasBytes(8) {
		return nil, ErrOutOfData
	}
	id := ChunkID(w.Data.ByteSlice(w.Ptr, w.Ptr+4))
	w.movePtr(4)
	chunkSize := int(w.getLE32(w.Ptr))
	w.movePtr(4)
	if !w.hasBytes(chunkSize) {
		return nil, ErrOutOfData
	}
	data := w.Data.Slice(w.Ptr, w.Ptr+chunkSize)
	w.movePtr(chunkSize)
	log.Printf("After chunk ptr %s is %d", id, w.Ptr)
	return &WOZ2Chunk{
		ID:   id,
		Size: chunkSize,
		Data: data,
		w:    w,
	}, nil
}

func (w *WOZ2Image) LoadChunks() {
	w.Chunks = make(map[ChunkID]*WOZ2Chunk)
	w.ResetChunkPtr()
	var err error
	var chunk *WOZ2Chunk
	for err == nil {
		chunk, err = w.ReadChunk()
		if err == nil && chunk.Size > 0 {
			log.Printf("Read %d byte chunk with ID [%s]", chunk.Size, chunk.ID)
			switch chunk.ID {
			case INFOChunk:
				w.INFO = &WOZ2INFOChunk{WOZ2Chunk: chunk}
			case TMAPChunk:
				w.TMAP = &WOZ2TMAPChunk{WOZ2Chunk: chunk}
			case TRKSChunk:
				w.TRKS = &WOZ2TRKSChunk{WOZ2Chunk: chunk}
				w.EssentialSize = w.Ptr + chunk.Size
			default:
				w.Chunks[chunk.ID] = chunk
			}
		}
	}
}

func (w *WOZ2Image) WriteByte(byteIdx int, b byte) {
	if w.Track != nil {
		w.Track.WriteByte(byteIdx, b)
	}
}

// func (w *WOZ2Image) Load525QTrack(q int) {
// 	log.Printf("Quarter track request: %d", q)

// 	p := w.TMAP.GetTrackPointer(q)

// 	if p == 0xff {
// 		log.Printf("Supplying empty track...")
// 		track := NewEmptyTrack()
// 		w.Track = track
// 		return
// 	}

// 	track := w.TRKS.GetTrack(p)
// 	if track != nil {
// 		w.Track = track
// 		log.Printf("Track bitcount = %d", track.BitCount())
// 		//fmt.Printf("Q Track = %d\n", q)
// 		//w.Track.ExtractNibbles()
// 	}
// }

func (w *WOZ2Image) Load525QTrack(q int) ([]byte, uint16) {
	log.Printf("Quarter track request: %d", q)

	p := w.TMAP.GetTrackPointer(q)

	if p == 0xff {
		log.Printf("Supplying empty track...")
		track := NewEmptyTrack()
		w.Track = track
		return track.BitStream(), track.BitCount()
	}

	track := w.TRKS.GetTrack(p)
	if track != nil {
		w.Track = track
		log.Printf("Track bitcount = %d", track.BitCount())
	}

	return track.BitStream(), track.BitCount()
}

func (w *WOZ2Image) Load35SideTrack(s int, t int) {
	p := w.TMAP.GetTrackPointer(80*s + t)
	track := w.TRKS.GetTrack(p)
	if track != nil {
		w.Track = track
	}
}

func (w *WOZ2Image) Advance() {
	if w.Track == nil {
		return
	}
	w.BitPtr = (w.BitPtr + 1) % int(w.Track.BitCount())
}

func (w *WOZ2Image) AdvanceN(count int) {
	//log.Println("adv")
	if w.Track == nil {
		return
	}
	w.BitPtr = (w.BitPtr + count) % int(w.Track.BitCount())
	//log.Printf("bitptr = %d", w.BitPtr)
}

func (w *WOZ2Image) UpdateCRC32() {
	crc := crc32.ChecksumIEEE(w.Data.ByteSlice(12, w.Size))
	setLE32(w.Data, 8, crc)
}

func (w *WOZ2Image) DumpGeometry() {

	for i := 0; i < 160; i++ {
		tnum := w.TMAP.GetTrackPointer(i)
		if tnum != 0xff {
			fmt.Printf("Stream %d -> %d\n", i, tnum)
		}
	}

}

type WOZ2Chunk struct {
	ID   ChunkID
	Size int
	Data memory.MemBytes
	w    *WOZ2Image
}

type WOZ2INFOChunk struct {
	*WOZ2Chunk
}

type WOZ2DiskType byte

const (
	WOZ2DiskType525 WOZ2DiskType = 1
	WOZ2DiskType35  WOZ2DiskType = 2
)

func (t WOZ2DiskType) String() string {
	switch t {
	case WOZ2DiskType35:
		return "3.5 inch floppy"
	case WOZ2DiskType525:
		return "5.25 inch floppy"
	}
	return "Unknown"
}

func (w *WOZ2INFOChunk) Version() byte {
	return w.Data.Get(0)
}

func (w *WOZ2INFOChunk) SetVersion(v byte) {
	w.Data.Set(0, v)
}

func (w *WOZ2INFOChunk) DiskType() WOZ2DiskType {
	return WOZ2DiskType(w.Data.Get(1))
}

func (w *WOZ2INFOChunk) SetDiskType(d WOZ2DiskType) {
	w.Data.Set(1, byte(d))
}

func (w *WOZ2INFOChunk) WriteProtected() bool {
	return w.Data.Get(2) == 1
}

func (w *WOZ2INFOChunk) SetWriteProtected(b bool) {
	if b {
		w.Data.Set(2, 1)
	} else {
		w.Data.Set(2, 0)
	}
}

func (w *WOZ2INFOChunk) Synchronized() bool {
	return w.Data.Get(3) == 1
}

func (w *WOZ2INFOChunk) SetSynchronized(b bool) {
	if b {
		w.Data.Set(3, 1)
	} else {
		w.Data.Set(3, 0)
	}
}

func (w *WOZ2INFOChunk) Cleaned() bool {
	return w.Data.Get(4) == 1
}

func (w *WOZ2INFOChunk) SetCleaned(b bool) {
	if b {
		w.Data.Set(4, 1)
	} else {
		w.Data.Set(4, 0)
	}
}

func (w *WOZ2INFOChunk) Creator() string {
	return string(w.Data.ByteSlice(5, 37))
}

func (w *WOZ2INFOChunk) SetCreator(c string) {
	for i := 0; i < 32; i++ {
		w.Data.Set(5+i, 32)
	}
	for i, ch := range c {
		if i < 32 {
			w.Data.Set(5+i, byte(ch))
		}
	}
}

func (w *WOZ2INFOChunk) DiskSides() int {
	return int(w.Data.Get(37))
}

func (w *WOZ2INFOChunk) SetDiskSides(n int) {
	w.Data.Set(37, byte(n))
}

type WOZ2BootSectorFormat byte

const (
	WOZ2BSFUnknown  WOZ2BootSectorFormat = 0
	WOZ2BSF16Sector WOZ2BootSectorFormat = 1
	WOZ2BSF13Sector WOZ2BootSectorFormat = 2
	WOZ2BSFBoth     WOZ2BootSectorFormat = 3
)

func (t WOZ2BootSectorFormat) String() string {
	switch t {
	case WOZ2BSF16Sector:
		return "16 Sector"
	case WOZ2BSF13Sector:
		return "13 Sector"
	case WOZ2BSFBoth:
		return "13 and 16 Sector"
	}
	return "Unknown"
}

func (w *WOZ2INFOChunk) BootSectorFormat() WOZ2BootSectorFormat {
	return WOZ2BootSectorFormat(w.Data.Get(38))
}

func (w *WOZ2INFOChunk) SetBootSectorFormat(n WOZ2BootSectorFormat) {
	w.Data.Set(38, byte(n))
}

func (w *WOZ2INFOChunk) OptimalBitTiming() int {
	return int(w.Data.Get(39))
}

func (w *WOZ2INFOChunk) SetOptimalBitTiming(n int) {
	w.Data.Set(39, byte(n))
}

type WOZ2CompatibleHardware uint16

const (
	WOZ2CompatApple2 WOZ2CompatibleHardware = 1 << iota
	WOZ2CompatApple2Plus
	WOZ2CompatApple2e
	WOZ2CompatApple2c
	WOZ2CompatApple2eEnhanced
	WOZ2CompatApple2gs
	WOZ2CompatApple2cPlus
	WOZ2CompatApple3
	WOZ2CompatApple3Plus
)

func (w *WOZ2INFOChunk) CompatibleHardware() WOZ2CompatibleHardware {
	return WOZ2CompatibleHardware(getLE16(w.Data, 40))
}

func (w *WOZ2INFOChunk) SetCompatibleHardware(n WOZ2CompatibleHardware) {
	setLE16(w.Data, 40, uint16(n))
}

func (w *WOZ2INFOChunk) RequiredRAM() int {
	return int(getLE16(w.Data, 42))
}

func (w *WOZ2INFOChunk) SetRequiredRAM(n int) {
	setLE16(w.Data, 42, uint16(n))
}

func (w *WOZ2INFOChunk) LargestTrack() int {
	return int(getLE16(w.Data, 44))
}

func (w *WOZ2INFOChunk) SetLargestTrack(n int) {
	setLE16(w.Data, 44, uint16(n))
}

func (w *WOZ2INFOChunk) String() string {
	return w.DiskType().String() + " created by " + w.Creator()
}

type WOZ2TMAPChunk struct {
	*WOZ2Chunk
}

func (w *WOZ2TMAPChunk) GetTrackPointer(i int) int {
	if i >= 160 || i < 0 {
		return 0
	}
	return int(w.Data.Get(i))
}

func (w *WOZ2TMAPChunk) SetTrackPointer(i int, p int) {
	if i >= 160 || i < 0 {
		return
	}
	w.Data.Set(i, byte(p))
}

/*

 */
const TRKHeaderLength = 8

type WOZ2TRKSChunk struct {
	*WOZ2Chunk
}

func (w *WOZ2TRKSChunk) GetTrack(index int) *WOZ2Track {
	offset := TRKHeaderLength * index
	if w.Data.Len()-offset < TRKHeaderLength {
		return nil
	}
	t := &WOZ2Track{
		Header: w.Data.Slice(offset, offset+TRKHeaderLength),
	}
	if w.w == nil {
		log.Printf("w.w is nil")
	}
	t.loadTrackData(w.w.Data)
	return t
}

type WOZ2Track struct {
	Header       memory.MemBytes
	Data         memory.MemBytes
	Modified     bool
	mMin, mMax   int
	realBitCount int
}

func (w *WOZ2Track) WriteByte(byteIdx int, b byte) {
	w.Data.Set(byteIdx, b)
}

func (w *WOZ2Track) loadTrackData(data memory.MemBytes) {
	startBlock := int(getLE16(w.Header, 0))
	blockCount := int(getLE16(w.Header, 2))
	bitCount := int(getLE32(w.Header, 4))

	s := startBlock * 512
	e := s + blockCount*512
	log.Printf("Track data start: %d, length: %d bytes", s, e-s)

	w.Data = data.Slice(s, e)
	w.realBitCount = bitCount
}

func (w *WOZ2Track) Copy(s *WOZ2Track) {
	for i := 0; i < s.Data.Len(); i++ {
		w.Data.Set(i, s.Data.Get(i))
	}
}

func (w *WOZ2Track) WriteBit(ptr int, value byte) {
	i := ptr % int(w.BitCount())
	byteIdx := i / 8
	bitIdx := uint(7 - (i % 8))
	bitset := byte(1 << bitIdx)
	bitclr := ^bitset

	b := w.Data.Get(byteIdx)
	b &= bitclr
	if value != 0 {
		b |= bitset
	}
	w.Data.Set(byteIdx, b)

	w.Modified = true
	if w.mMin == w.mMin && w.mMax == 0 {
		w.mMax = ptr
		w.mMin = ptr
	} else {
		if ptr < w.mMin {
			w.mMin = ptr
		}
		if ptr > w.mMax {
			w.mMax = ptr
		}
	}
}

func (w *WOZ2Track) ReadBit(ptr int) byte {
	i := ptr % int(w.BitCount())
	byteIdx := i / 8
	bitIdx := uint(7 - (i % 8))
	return (w.Data.Get(byteIdx) >> bitIdx) & 1
}

func (w *WOZ2Track) BitStream() []byte {
	c := int(w.BytesUsed())
	return w.Data.ByteSlice(0, c)
}

func (w *WOZ2Track) BytesUsed() uint16 {
	return uint16(w.realBitCount/8 + 1)
}

func (w *WOZ2Track) SetBytesUsed(b uint16) {
	w.realBitCount = 8 * int(b)
	setLE32(w.Header, 4, uint32(w.realBitCount))
}

func (w *WOZ2Track) BitCount() uint16 {
	return uint16(w.realBitCount)
}

func (w *WOZ2Track) SetBitCount(b uint16) {
	w.realBitCount = int(b)
	setLE32(w.Header, 4, uint32(w.realBitCount))
}

// func (w *WOZ2Track) SplicePoint() uint16 {
// 	offset := bitstreamLength + 4
// 	return getLE16(w.Data, offset)
// }

// func (w *WOZ2Track) SetSplicePoint(b uint16) {
// 	offset := bitstreamLength + 4
// 	setLE16(w.Data, offset, b)
// }

// func (w *WOZ2Track) SpliceNibble() byte {
// 	offset := bitstreamLength + 6
// 	return w.Data.Get(offset)
// }

// func (w *WOZ2Track) SetSpliceNibble(b byte) {
// 	offset := bitstreamLength + 6
// 	w.Data.Set(offset, b)
// }

// func (w *WOZ2Track) SpliceBitCount() byte {
// 	offset := bitstreamLength + 7
// 	return w.Data.Get(offset)
// }

// func (w *WOZ2Track) SetSpliceBitCount(b byte) {
// 	offset := bitstreamLength + 7
// 	w.Data.Set(offset, b)
// }

const minWOZ2Size = 12 + 8 + 60 + 8 + 160 + 8 + 232960

func (w *WOZ2Image) NewChunk(index int, size uint32, id ChunkID) *WOZ2Chunk {
	asize := 8 + size // factor header
	for i := 0; i < int(asize); i++ {
		w.Data.Set(index+i, 0)
	}
	for i, ch := range id[:] {
		w.Data.Set(index+i, byte(ch))
	}
	setLE32(w.Data, index+4, size)
	return &WOZ2Chunk{
		Data: w.Data.Slice(index+8, index+8+int(size)),
		ID:   id,
		Size: int(size),
		w:    w,
	}
}

func CreateWOZ2Empty(data memory.MemBytes) *WOZ2Image {
	w := &WOZ2Image{
		Ptr:    0,
		BitPtr: 0,
		Data:   data,
		Size:   233216,
	}
	// Add magic header
	for i, v := range wozMagic {
		w.Data.Set(i, v)
	}

	// INFO chunk
	w.INFO = &WOZ2INFOChunk{
		WOZ2Chunk: w.NewChunk(12, 60, INFOChunk),
	}
	w.INFO.SetCleaned(true)
	w.INFO.SetWriteProtected(false)
	w.INFO.SetDiskType(WOZ2DiskType525)
	w.INFO.SetSynchronized(true)
	w.INFO.SetVersion(1)
	w.INFO.SetCreator("MicroM8 (c) Paleotronic.com")
	// TMAP Chunk
	w.TMAP = &WOZ2TMAPChunk{
		WOZ2Chunk: w.NewChunk(80, 160, TMAPChunk),
	}
	for i := 0; i < 160; i++ {
		w.TMAP.SetTrackPointer(i, 0xff)
	}
	for t := 0; t < 35; t++ {
		for qt := 0; qt < 4; qt++ {
			w.TMAP.SetTrackPointer(t*4+qt, t)
		}
	}
	// TRKS Chunk
	w.TRKS = &WOZ2TRKSChunk{
		WOZ2Chunk: w.NewChunk(248, 232960, TRKSChunk),
	}

	for t := 0; t < 35; t++ {
		track := &WOZ2Track{Data: w.Data.Slice(256+t*6656, 256+(t+1)*6656)}
		tr := NewEmptyTrack()
		track.Copy(tr)
	}

	return w
}

func NewEmptyTrack() *WOZ2Track {
	data := memory.NewMemByteSlice(trackLength)
	header := memory.NewMemByteSlice(8)
	track := &WOZ2Track{
		Data:   data,
		Header: header,
	}
	track.SetBitCount(defaultTrackBits)
	track.SetBytesUsed(defaultTrackBytes)
	track.realBitCount = defaultTrackBits
	// track.SetSpliceBitCount(0x00)
	// track.SetSpliceNibble(0x00)
	// track.SetSplicePoint(0x00)
	return track
}

// func NewTrackFromNibbles(nibbles []byte, syncs []byte) *WOZ2Track {
// 	log.Printf("Len syncs = %d", len(syncs))
// 	if len(syncs) == 0 {
// 		log.Println("Need syncs")
// 		syncs = CalcSyncsForNibbles(nibbles)
// 	}
// 	data := memory.NewMemByteSlice(trackLength)
// 	track := &WOZ2Track{
// 		Data: data,
// 	}
// 	b, bytecount, bitcount := NibblesToBitstream(nibbles, syncs)
// 	for i, bb := range b {
// 		track.Data.Set(i, bb)
// 	}
// 	track.SetBitCount(bitcount)
// 	track.SetBytesUsed(bytecount)
// 	track.SetSpliceBitCount(0x00)
// 	track.SetSpliceNibble(0x00)
// 	track.SetSplicePoint(0x00)
// 	return track
// }

// func CreateWOZ2FromDSK(dsk *disk.DSKWrapper, data memory.MemBytes) *WOZ2Image {
// 	w := CreateWOZ2Empty(data)
// 	for t := 0; t < 35; t++ {
// 		nibbles, syncs := dsk.NibblizeTrack(t)
// 		strack := NewTrackFromNibbles(nibbles, syncs)
// 		track := &WOZ2Track{Data: w.Data.Slice(256+t*6656, 256+(t+1)*6656)}
// 		track.Copy(strack)
// 	}
// 	w.UpdateCRC32()

// 	return w
// }

// var badMMHeader = []byte{
// 	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
// 	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xd5,
// 	0xaa, 0x96,
// }

// func CreateWOZ2FromNIB(data []byte, buffer memory.MemBytes) *WOZ2Image {

// 	// Fix: Bad nibble encoding of previous microM8 disks
// 	if bytes.HasPrefix(data, badMMHeader) {
// 		dskdata, err := DeNibblizeImage(data, nil)
// 		if err != nil {
// 			log.Printf("Failed to convert nib: %v", err)
// 			os.Exit(1)
// 			return CreateWOZ2Empty(buffer)
// 		}
// 		dsk, err := disk.NewDSKWrapperBin(nil, dskdata, "fixed.dsk")
// 		if err != nil {
// 			return CreateWOZ2Empty(buffer)
// 		}
// 		return CreateWOZ2FromDSK(dsk, buffer)
// 	}

// 	log.Println("===================FROM NIB===================")
// 	w := CreateWOZ2Empty(buffer)
// 	out := make([]byte, 0, 232960)
// 	for t := 0; t < 35; t++ {
// 		offset := 6656 * t
// 		raw := data[offset : offset+6566]
// 		nibbles := StandardizeNibbleSpacing(raw) // fix: syncs and leaders for dodgy nib files :)
// 		out = append(out, nibbles...)
// 		strack := NewTrackFromNibbles(nibbles[:6544], []byte(nil))
// 		track := &WOZ2Track{Data: w.Data.Slice(256+t*6656, 256+(t+1)*6656)}
// 		track.Copy(strack)
// 	}
// 	w.UpdateCRC32()

// 	f, _ := os.Create("fixed.nib")
// 	f.Write(out)
// 	f.Close()
// 	return w
// }

// func (w *WOZ2Track) ExtractTrackNibbles() ([]byte, error) {
// 	var reg byte
// 	var out = make([]byte, 0, std525NibblesPerTrack)

// 	bp := 0
// 	for bp < int(w.BitCount()) {
// 		reg = (reg << 1) | w.ReadBit(bp)
// 		if reg&0x80 != 0 {
// 			out = append(out, reg)
// 			reg = 0x00
// 		}
// 		bp++
// 	}

// 	if len(out) > std525NibblesPerTrack {
// 		return out, errors.New("nibbles exceeds standard length bytes")
// 	}

// 	if len(out) < std525NibblesPerTrack {
// 		// We might need to pad here..
// 		index := bytes.Index(out, addrPrologue)
// 		if index == -1 {
// 			return out, errors.New("could not sync to standard length bytes")
// 		}
// 		needed := std525NibblesPerTrack - len(out)
// 		syncs := make([]byte, needed)
// 		for i, _ := range syncs {
// 			syncs[i] = 0xFF
// 		}
// 		head := out[:index]
// 		tail := out[index:]
// 		out = append(head, syncs...)
// 		out = append(out, tail...)
// 	}

// 	return out, nil
// }

// func (w *WOZ2Image) ConvertToDSK() (*disk.DSKWrapper, error) {

// 	var nib = make([]byte, 0, 232960)

// 	if w.INFO.DiskType() != WOZ2DiskType525 {
// 		return nil, errors.New("Wrong disk image type")
// 	}

// 	if w.TMAP.GetTrackPointer(36*4) != 0xff {
// 		return nil, errors.New("Disk has non-standard number of tracks")
// 	}

// 	for t := 0; t < 35; t++ {
// 		qt := 4 * t
// 		w.Load525QTrack(qt)
// 		nibbles, err := w.Track.ExtractTrackNibbles()
// 		if err != nil {
// 			return nil, err
// 		}
// 		nib = append(nib, nibbles...)
// 	}

// 	f, _ := os.Create("out.nib")
// 	f.Write(nib)
// 	f.Close()

// 	d, err := DeNibblizeImage(nib, nil)

// 	if err != nil {
// 		return nil, err
// 	}

// 	return disk.NewDSKWrapperBin(nil, d, "wozdisk.dsk")
// }
