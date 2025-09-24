package woz

import (
	"bytes"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"io/ioutil"
	"os"

	"paleotronic.com/log"

	"paleotronic.com/core/memory"
	"paleotronic.com/disk"
)

const crcPos = 8
const chunkBase = 12

var wozMagic = []byte{0x57, 0x4f, 0x5a, 0x31, 0xff, 0x0a, 0x0d, 0x0a}
var ErrNotValid = errors.New("Not a valid WOZ image")
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

type WOZImage struct {
	Data   memory.MemBytes
	Ptr    int
	Chunks map[ChunkID]*WOZChunk
	// Recognized Chunks
	INFO *WOZINFOChunk
	TMAP *WOZTMAPChunk
	TRKS *WOZTRKSChunk
	// Currently mapped track
	Track         *WOZTrack
	BitPtr        int
	Modified      bool
	Size          int
	EssentialSize int
}

const trimWoz = false

func NewWOZImage(r io.Reader, buffer memory.MemBytes) (*WOZImage, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	if len(data) > buffer.Len() {
		//fmt.Println("Using dynamic buffer for WOZ as mapped buffer is too small")
		buffer = memory.NewMemByteSlice(len(data))
	}

	length := len(data)
	log.Printf("Initial size = %d bytes", length)
	if trimWoz {
		tmp := memory.NewMemByteSlice(length)
		tmp.Write(0, data)
		img := &WOZImage{
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
	img := &WOZImage{
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

	//img.TMAP.Dump()

	//img.DumpGeometry()

	return img, err
}

func (w *WOZImage) GetOptimalBitTiming() int {
	return 32
}

func (w *WOZImage) SetWriteProtected(b bool) {
	w.INFO.SetWriteProtected(b)
}

func (w *WOZImage) WriteProtected() bool {
	return w.INFO.WriteProtected()
}

func (w *WOZImage) SetModified(b bool) {
	w.Modified = b
}

func (w *WOZImage) IsModified() bool {
	return w.Modified
}

func (w *WOZImage) GetData() memory.MemBytes {
	return w.Data
}

func (w *WOZImage) GetSize() int {
	return w.Size
}

func (w *WOZImage) BitCount() uint16 {
	return w.Track.BitCount()
}

func (w *WOZImage) ReadBit(ptr int) byte {
	return w.Track.ReadBit(ptr)
}

func (w *WOZImage) WriteBit(ptr int, bit byte) {
	w.Track.WriteBit(ptr, bit)
}

func (w *WOZImage) TrackOK() bool {
	return w.Track != nil
}

func (w *WOZImage) IsValid() bool {
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

	return true
}

func (w *WOZImage) getLE32(index int) uint32 {
	if w.Data.Len()-index < 4 {
		return 0
	}
	return uint32(w.Data.Get(index+0)) |
		(uint32(w.Data.Get(index+1)) << 8) |
		(uint32(w.Data.Get(index+2)) << 16) |
		(uint32(w.Data.Get(index+3)) << 24)
}

func (w *WOZImage) GetEmbeddedCRC() uint32 {
	return w.getLE32(crcPos)
}

func (w *WOZImage) ResetChunkPtr() {
	w.Ptr = chunkBase
}

func (w *WOZImage) movePtr(change int) error {
	n := w.Ptr + change
	if n < 0 || n >= w.Data.Len() {
		return ErrOutOfData
	}
	w.Ptr = n
	return nil
}

func (w *WOZImage) hasBytes(count int) bool {
	return w.Data.Len()-w.Ptr >= count
}

func (w *WOZImage) ReadChunk() (*WOZChunk, error) {
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
	log.Printf("After chunk ptr is %d", w.Ptr)
	return &WOZChunk{
		ID:   id,
		Size: chunkSize,
		Data: data,
	}, nil
}

func (w *WOZImage) LoadChunks() {
	w.Chunks = make(map[ChunkID]*WOZChunk)
	w.ResetChunkPtr()
	var err error
	var chunk *WOZChunk
	for err == nil {
		chunk, err = w.ReadChunk()
		if err == nil && chunk.Size > 0 {
			log.Printf("Read %d byte chunk with ID [%s]", chunk.Size, chunk.ID)
			switch chunk.ID {
			case INFOChunk:
				w.INFO = &WOZINFOChunk{WOZChunk: chunk}
			case TMAPChunk:
				w.TMAP = &WOZTMAPChunk{WOZChunk: chunk}
			case TRKSChunk:
				w.TRKS = &WOZTRKSChunk{WOZChunk: chunk}
				w.EssentialSize = w.Ptr + chunk.Size
			default:
				w.Chunks[chunk.ID] = chunk
			}
		}
	}
}

func (w *WOZImage) Load525QTrack(q int) ([]byte, uint16) {
	log.Printf("Quarter track request: %d", q)

	p := w.TMAP.GetTrackPointer(q)

	if p == 0xff {
		log.Printf("Supplying empty track...")
		track := NewEmptyTrack()
		w.Track = track
		return track.BitStream(), track.BitCount()
	}

	log.Printf("track pointer is %d", p)

	track := w.TRKS.GetTrack(p)
	if track != nil {
		w.Track = track
		log.Printf("Track bitcount = %d", track.BitCount())
		if track.BitCount() == 0 {
			log.Printf("Supplying empty track because it seems corrupt...")
			track := NewEmptyTrack()
			w.Track = track
			return track.BitStream(), track.BitCount()
		}
	}

	return track.BitStream(), track.BitCount()
}

func (w *WOZImage) Load35SideTrack(s int, t int) {
	p := w.TMAP.GetTrackPointer(80*s + t)
	track := w.TRKS.GetTrack(p)
	if track != nil {
		w.Track = track
	}
}

func (w *WOZImage) Advance() {
	if w.Track == nil {
		return
	}
	w.BitPtr = (w.BitPtr + 1) % int(w.Track.BitCount())
}

func (w *WOZImage) AdvanceN(count int) {
	if w.Track == nil {
		return
	}
	w.BitPtr = (w.BitPtr + count) % int(w.Track.BitCount())
}

func (w *WOZImage) UpdateCRC32() {
	crc := crc32.ChecksumIEEE(w.Data.ByteSlice(12, w.Size))
	setLE32(w.Data, 8, crc)
}

func (w *WOZImage) DumpGeometry() {

	for i := 0; i < 160; i++ {
		tnum := w.TMAP.GetTrackPointer(i)
		if tnum != 0xff {
			fmt.Printf("Stream %d -> %d\n", i, tnum)
		}
	}

}

func (w *WOZImage) WriteByte(byteIdx int, b byte) {
	if w.Track != nil {
		w.Track.WriteByte(byteIdx, b)
	}
}

type WOZChunk struct {
	ID   ChunkID
	Size int
	Data memory.MemBytes
}

type WOZINFOChunk struct {
	*WOZChunk
}

type WOZDiskType byte

const (
	WOZDiskType525 WOZDiskType = 1
	WOZDiskType35  WOZDiskType = 2
)

func (t WOZDiskType) String() string {
	switch t {
	case WOZDiskType35:
		return "3.5 inch floppy"
	case WOZDiskType525:
		return "5.25 inch floppy"
	}
	return "Unknown"
}

func (w *WOZINFOChunk) Version() byte {
	return w.Data.Get(0)
}

func (w *WOZINFOChunk) SetVersion(v byte) {
	w.Data.Set(0, v)
}

func (w *WOZINFOChunk) DiskType() WOZDiskType {
	return WOZDiskType(w.Data.Get(1))
}

func (w *WOZINFOChunk) SetDiskType(d WOZDiskType) {
	w.Data.Set(1, byte(d))
}

func (w *WOZINFOChunk) WriteProtected() bool {
	return w.Data.Get(2) == 1
}

func (w *WOZINFOChunk) SetWriteProtected(b bool) {
	if b {
		w.Data.Set(2, 1)
	} else {
		w.Data.Set(2, 0)
	}
}

func (w *WOZINFOChunk) Synchronized() bool {
	return w.Data.Get(3) == 1
}

func (w *WOZINFOChunk) SetSynchronized(b bool) {
	if b {
		w.Data.Set(3, 1)
	} else {
		w.Data.Set(3, 0)
	}
}

func (w *WOZINFOChunk) Cleaned() bool {
	return w.Data.Get(4) == 1
}

func (w *WOZINFOChunk) SetCleaned(b bool) {
	if b {
		w.Data.Set(4, 1)
	} else {
		w.Data.Set(4, 0)
	}
}

func (w *WOZINFOChunk) Creator() string {
	return string(w.Data.ByteSlice(5, 37))
}

func (w *WOZINFOChunk) SetCreator(c string) {
	for i := 0; i < 32; i++ {
		w.Data.Set(5+i, 32)
	}
	for i, ch := range c {
		if i < 32 {
			w.Data.Set(5+i, byte(ch))
		}
	}
}

func (w *WOZINFOChunk) String() string {
	return w.DiskType().String() + " created by " + w.Creator()
}

type WOZTMAPChunk struct {
	*WOZChunk
}

func (w *WOZTMAPChunk) GetTrackPointer(i int) int {
	if i >= 160 || i < 0 {
		return 0
	}
	return int(w.Data.Get(i))
}

func (w *WOZTMAPChunk) SetTrackPointer(i int, p int) {
	if i >= 160 || i < 0 {
		return
	}
	w.Data.Set(i, byte(p))
}

func (w *WOZTMAPChunk) Dump() {
	log.Printf("  | %.2x %.2x %.2x %.2x %.2x %.2x %.2x %.2x %.2x %.2x %.2x %.2x %.2x %.2x %.2x %.2x ",
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15)
	for i := 0; i < 160; i += 16 {
		log.Printf(
			"%.2x| %.2x %.2x %.2x %.2x %.2x %.2x %.2x %.2x %.2x %.2x %.2x %.2x %.2x %.2x %.2x %.2x ",
			i,
			w.Data.Get(i+0),
			w.Data.Get(i+1),
			w.Data.Get(i+2),
			w.Data.Get(i+3),
			w.Data.Get(i+4),
			w.Data.Get(i+5),
			w.Data.Get(i+6),
			w.Data.Get(i+7),
			w.Data.Get(i+8),
			w.Data.Get(i+9),
			w.Data.Get(i+10),
			w.Data.Get(i+11),
			w.Data.Get(i+12),
			w.Data.Get(i+13),
			w.Data.Get(i+14),
			w.Data.Get(i+15),
		)
	}
	os.Exit(0)
}

type WOZTRKSChunk struct {
	*WOZChunk
}

func (w *WOZTRKSChunk) GetTrack(index int) *WOZTrack {
	offset := trackLength * index
	log.Printf("track is at offset %d", offset)
	if w.Data.Len()-offset < trackLength {
		return nil
	}
	return &WOZTrack{Data: w.Data.Slice(offset, offset+trackLength)}
}

type WOZTrack struct {
	Data       memory.MemBytes
	Modified   bool
	mMin, mMax int
}

func (w *WOZTrack) Copy(s *WOZTrack) {
	for i := 0; i < s.Data.Len(); i++ {
		w.Data.Set(i, s.Data.Get(i))
	}
}

func (w *WOZTrack) WriteBit(ptr int, value byte) {
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

func (w *WOZTrack) WriteByte(byteIdx int, b byte) {
	w.Data.Set(byteIdx, b)
}

func (w *WOZTrack) ReadBit(ptr int) byte {
	i := ptr % int(w.BitCount())
	byteIdx := i / 8
	bitIdx := uint(7 - (i % 8))
	return (w.Data.Get(byteIdx) >> bitIdx) & 1
}

func (w *WOZTrack) BitStream() []byte {
	c := int(w.BytesUsed())
	return w.Data.ByteSlice(0, c)
}

func (w *WOZTrack) BytesUsed() uint16 {
	offset := bitstreamLength
	return getLE16(w.Data, offset)
}

func (w *WOZTrack) SetBytesUsed(b uint16) {
	offset := bitstreamLength
	setLE16(w.Data, offset, b)
}

func (w *WOZTrack) BitCount() uint16 {
	offset := bitstreamLength + 2
	return getLE16(w.Data, offset)
}

func (w *WOZTrack) SetBitCount(b uint16) {
	offset := bitstreamLength + 2
	setLE16(w.Data, offset, b)
}

func (w *WOZTrack) SplicePoint() uint16 {
	offset := bitstreamLength + 4
	return getLE16(w.Data, offset)
}

func (w *WOZTrack) SetSplicePoint(b uint16) {
	offset := bitstreamLength + 4
	setLE16(w.Data, offset, b)
}

func (w *WOZTrack) SpliceNibble() byte {
	offset := bitstreamLength + 6
	return w.Data.Get(offset)
}

func (w *WOZTrack) SetSpliceNibble(b byte) {
	offset := bitstreamLength + 6
	w.Data.Set(offset, b)
}

func (w *WOZTrack) SpliceBitCount() byte {
	offset := bitstreamLength + 7
	return w.Data.Get(offset)
}

func (w *WOZTrack) SetSpliceBitCount(b byte) {
	offset := bitstreamLength + 7
	w.Data.Set(offset, b)
}

const minWOZSize = 12 + 8 + 60 + 8 + 160 + 8 + 232960

func (w *WOZImage) NewChunk(index int, size uint32, id ChunkID) *WOZChunk {
	asize := 8 + size // factor header
	for i := 0; i < int(asize); i++ {
		w.Data.Set(index+i, 0)
	}
	for i, ch := range id[:] {
		w.Data.Set(index+i, byte(ch))
	}
	setLE32(w.Data, index+4, size)
	return &WOZChunk{
		Data: w.Data.Slice(index+8, index+8+int(size)),
		ID:   id,
		Size: int(size),
	}
}

func CreateWOZEmpty(data memory.MemBytes) *WOZImage {
	w := &WOZImage{
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
	w.INFO = &WOZINFOChunk{
		WOZChunk: w.NewChunk(12, 60, INFOChunk),
	}
	w.INFO.SetCleaned(true)
	w.INFO.SetWriteProtected(false)
	w.INFO.SetDiskType(WOZDiskType525)
	w.INFO.SetSynchronized(true)
	w.INFO.SetVersion(1)
	w.INFO.SetCreator("MicroM8 (c) Paleotronic.com")
	// TMAP Chunk
	w.TMAP = &WOZTMAPChunk{
		WOZChunk: w.NewChunk(80, 160, TMAPChunk),
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
	w.TRKS = &WOZTRKSChunk{
		WOZChunk: w.NewChunk(248, 232960, TRKSChunk),
	}

	for t := 0; t < 35; t++ {
		track := &WOZTrack{Data: w.Data.Slice(256+t*6656, 256+(t+1)*6656)}
		tr := NewEmptyTrack()
		track.Copy(tr)
	}

	return w
}

func NewEmptyTrack() *WOZTrack {
	data := memory.NewMemByteSlice(trackLength)
	track := &WOZTrack{
		Data: data,
	}
	track.SetBitCount(defaultTrackBits)
	track.SetBytesUsed(defaultTrackBytes)
	track.SetSpliceBitCount(0x00)
	track.SetSpliceNibble(0x00)
	track.SetSplicePoint(0x00)
	return track
}

func NewTrackFromNibbles(nibbles []byte, syncs []byte) *WOZTrack {
	log.Printf("Len syncs = %d", len(syncs))
	if len(syncs) == 0 {
		log.Println("Need syncs")
		syncs = CalcSyncsForNibbles(nibbles)
	}
	data := memory.NewMemByteSlice(trackLength)
	track := &WOZTrack{
		Data: data,
	}
	b, bytecount, bitcount := NibblesToBitstream(nibbles, syncs)
	for i, bb := range b {
		track.Data.Set(i, bb)
	}
	track.SetBitCount(bitcount)
	track.SetBytesUsed(bytecount)
	track.SetSpliceBitCount(0x00)
	track.SetSpliceNibble(0x00)
	track.SetSplicePoint(0x00)
	return track
}

func CreateWOZFromDSK(dsk *disk.DSKWrapper, data memory.MemBytes) *WOZImage {
	w := CreateWOZEmpty(data)
	for t := 0; t < 35; t++ {
		nibbles, syncs := dsk.NibblizeTrack(t)
		strack := NewTrackFromNibbles(nibbles, syncs)
		track := &WOZTrack{Data: w.Data.Slice(256+t*6656, 256+(t+1)*6656)}
		track.Copy(strack)
	}
	w.UpdateCRC32()

	return w
}

var badMMHeader = []byte{
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xd5,
	0xaa, 0x96,
}

func CreateWOZFromNIB(data []byte, buffer memory.MemBytes) *WOZImage {

	// Fix: Bad nibble encoding of previous microM8 disks
	if bytes.HasPrefix(data, badMMHeader) {
		dskdata, err := DeNibblizeImage(data, nil)
		if err != nil {
			log.Printf("Failed to convert nib: %v", err)
			os.Exit(1)
			return CreateWOZEmpty(buffer)
		}
		dsk, err := disk.NewDSKWrapperBin(nil, dskdata, "fixed.dsk")
		if err != nil {
			return CreateWOZEmpty(buffer)
		}
		return CreateWOZFromDSK(dsk, buffer)
	}

	log.Println("===================FROM NIB===================")
	w := CreateWOZEmpty(buffer)
	out := make([]byte, 0, 232960)
	for t := 0; t < 35; t++ {
		offset := 6656 * t
		raw := data[offset : offset+6566]
		nibbles := StandardizeNibbleSpacing(raw) // fix: syncs and leaders for dodgy nib files :)
		out = append(out, nibbles...)
		strack := NewTrackFromNibbles(nibbles[:6544], []byte(nil))
		track := &WOZTrack{Data: w.Data.Slice(256+t*6656, 256+(t+1)*6656)}
		track.Copy(strack)
	}
	w.UpdateCRC32()

	f, _ := os.Create("fixed.nib")
	f.Write(out)
	f.Close()
	return w
}

func (w *WOZTrack) ExtractTrackNibbles() ([]byte, error) {
	var reg byte
	var out = make([]byte, 0, std525NibblesPerTrack)

	bp := 0
	for bp < int(w.BitCount()) {
		reg = (reg << 1) | w.ReadBit(bp)
		if reg&0x80 != 0 {
			out = append(out, reg)
			reg = 0x00
		}
		bp++
	}

	if len(out) > std525NibblesPerTrack {
		return out, errors.New("nibbles exceeds standard length bytes")
	}

	if len(out) < std525NibblesPerTrack {
		// We might need to pad here..
		index := bytes.Index(out, addrPrologue)
		if index == -1 {
			return out, errors.New("could not sync to standard length bytes")
		}
		needed := std525NibblesPerTrack - len(out)
		syncs := make([]byte, needed)
		for i, _ := range syncs {
			syncs[i] = 0xFF
		}
		head := out[:index]
		tail := out[index:]
		out = append(head, syncs...)
		out = append(out, tail...)
	}

	return out, nil
}

func (w *WOZImage) ConvertToDSK() (*disk.DSKWrapper, error) {

	var nib = make([]byte, 0, 232960)

	if w.INFO.DiskType() != WOZDiskType525 {
		return nil, errors.New("Wrong disk image type")
	}

	if w.TMAP.GetTrackPointer(36*4) != 0xff {
		return nil, errors.New("Disk has non-standard number of tracks")
	}

	for t := 0; t < 35; t++ {
		qt := 4 * t
		w.Load525QTrack(qt)
		nibbles, err := w.Track.ExtractTrackNibbles()
		if err != nil {
			return nil, err
		}
		nib = append(nib, nibbles...)
	}

	f, _ := os.Create("out.nib")
	f.Write(nib)
	f.Close()

	d, err := DeNibblizeImage(nib, nil)

	if err != nil {
		return nil, err
	}

	return disk.NewDSKWrapperBin(nil, d, "wozdisk.dsk")
}
