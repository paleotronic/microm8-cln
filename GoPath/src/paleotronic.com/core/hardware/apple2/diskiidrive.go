package apple2

import (
	"bytes"
	"io/ioutil"
	"math/rand"
	"strings"
	"time"

	"paleotronic.com/core/hardware/apple2/woz"
	"paleotronic.com/core/hardware/apple2/woz2"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/settings"
	"paleotronic.com/disk"
	"paleotronic.com/filerecord"
	"paleotronic.com/files"
	"paleotronic.com/log"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var driveHeadStepDelta [][]int = [][]int{
	{0, 0, 1, 1, 0, 0, 1, 1, -1, -1, 0, 0, -1, -1, 0, 0}, // phase 0
	{0, -1, 0, -1, 1, 0, 1, 0, 0, -1, 0, -1, 1, 0, 1, 0}, // phase 1
	{0, 0, -1, -1, 0, 0, -1, -1, 1, 1, 0, 0, 1, 1, 0, 0}, // phase 2
	{0, 1, 0, 1, -1, 0, -1, 0, 0, 1, 0, 1, -1, 0, -1, 0}, // phase 3
}

const (
	DISKII_HALFTRACK        = 0
	DISKII_TRACKSTARTOFFSET = 1
	DISKII_NIBBLEOFFSET     = 2
	DISKII_WRITEMODE        = 3
	DISKII_DRIVEON          = 4
	DISKII_MAGNETS          = 5
	DISKII_BITPTR           = 6
	DISKII_SPINCOUNT        = 7
	DISKII_DIRTY_TRACKS     = 8
	DISKII_UPDATE_PENDING   = 9
	DISKII_DRIVE_LATCH      = 10
	DISKII_USE_NIB          = 11
	DISKII_PHASECYCLES      = 12
	DISKII_PHASECONTROL     = 13
	DISKII_PHASEDIRECTION   = 14
	DISKII_PHASELASTBUMP    = 15
	DISKII_PHASESTOP        = 16
	DISKII_PHASEALIGN       = 17
	DISKII_TRACKINDEX       = 18
	DISKII_TRACKPHASE       = 19
	DISKII_NIBBLE_BASE      = 64
	DISKII_NIBBLE_SIZE      = 29120
	DISKII_UNIT_SIZE        = 30000

	halfTrackCount = 80
)

type WOZImg interface {
	Load525QTrack(t int) ([]byte, uint16)
	WriteProtected() bool
	SetWriteProtected(b bool)
	ReadBit(ptr int) byte
	WriteBit(ptr int, bit byte)
	BitCount() uint16
	TrackOK() bool
	SetModified(b bool)
	IsModified() bool
	UpdateCRC32()
	GetData() memory.MemBytes
	GetSize() int
	AdvanceN(count int)
	WriteByte(byteIdx int, b byte)
	GetOptimalBitTiming() int
}

type DiskIIDrive struct {
	Disk               WOZImg //*woz.WOZImage // Actual file representation of an Apple Disk
	NibImage           *disk.DSKWrapper
	VolumeName         string
	LastWriteTime      time.Time
	ActivationCallback func(b bool)
	slotid             int
	drive              int
	baseaddr           int
	mcb                *memory.MemoryControlBlock
	zeroCount          int
	lastQTrack         int
	parent             *IOCardDiskII
	currentBitCount    uint16
	currentBitStream   []byte
}

func (d *DiskIIDrive) checkQuarterTrack() {
	if d.Disk == nil {
		return
	}

	currQTrack := d.GetTrackIndex()
	if d.lastQTrack != currQTrack {
		d.lastQTrack = currQTrack
		d.currentBitStream, d.currentBitCount = d.Disk.Load525QTrack(currQTrack)
		if d.currentBitCount == 0 {
			log.Printf("Warning switched to q-track %d and it has a bitcount of zero!", currQTrack)
		}
	}
}

func (d *DiskIIDrive) getDriveMemoryBlocks() [][2]int {
	index := d.parent.e.GetMemIndex()
	mm := d.parent.e.GetMemoryMap()
	blocks := make([][2]int, 0)
	// 1st block - same as before
	base := d.baseaddr + DISKII_NIBBLE_BASE
	end := base + DISKII_UNIT_SIZE - DISKII_NIBBLE_BASE
	blocks = append(blocks, [2]int{base, end})
	// second block
	base2 := mm.MEMBASE(index) + memory.MICROM8_2ND_DISKII_BASE + (memory.MICROM8_2ND_DISKII_SIZE * d.drive)
	end2 := base2 + memory.MICROM8_2ND_DISKII_SIZE
	blocks = append(blocks, [2]int{base2, end2})
	return blocks
}

func (d *DiskIIDrive) GetWBuffer() memory.MemBytes {

	index := d.parent.e.GetMemIndex()
	mm := d.parent.e.GetMemoryMap()
	mcb := memory.NewMemoryControlBlock(mm, index, false)
	blocks := d.getDriveMemoryBlocks()
	for i, b := range blocks {
		base := b[0]
		end := b[1]
		mcb.Add(
			mm.Data[index][base:end],
			base,
		)
		log.Printf("Buffer %d for vm#%d drive %d runs from 0x%.6x to 0x%.6x (size %d bytes)", i, index, d.drive, base, end, (end-base)*8)
	}
	log.Printf("DiskII buffer capacity is %d bytes", mcb.Size*8)

	return memory.NewMemMicroM8(
		d.mcb.GetMM(),
		d.slotid,
		0,
		mcb,
	)

}

func (d *DiskIIDrive) GetNibble(offset int) byte {
	idx := offset / 8
	mod := offset % 8
	bs := uint64(mod * 8)
	mask := uint64(0xff) << bs

	val := d.mcb.Read(DISKII_NIBBLE_BASE + idx)
	return byte((val & mask) >> bs)
}

func (d *DiskIIDrive) SetNibble(offset int, value byte) {
	idx := offset / 8
	mod := offset % 8
	bs := uint64(mod * 8)
	mask := uint64(0xff) << bs

	val := d.mcb.Read(DISKII_NIBBLE_BASE + idx)
	clrmask := uint64(0xffffffffffffffff) ^ mask
	val = val&clrmask | (uint64(value) << bs)
	d.mcb.Write(DISKII_NIBBLE_BASE+idx, val)
}

func (d *DiskIIDrive) SetPhaseDirection(v int) {
	d.mcb.Write(DISKII_PHASEDIRECTION, uint64(v))
}

func (d *DiskIIDrive) GetPhaseDirection() int {
	return int(d.mcb.Read(DISKII_PHASEDIRECTION))
}

func (d *DiskIIDrive) SetPhaseCycles(v int) {
	d.mcb.Write(DISKII_PHASECYCLES, uint64(v))
}

func (d *DiskIIDrive) GetPhaseCycles() int {
	return int(d.mcb.Read(DISKII_PHASECYCLES))
}

func (d *DiskIIDrive) SetPhaseControl(v int) {
	d.mcb.Write(DISKII_PHASECONTROL, uint64(v))
}

func (d *DiskIIDrive) GetPhaseControl() int {
	return int(d.mcb.Read(DISKII_PHASECONTROL))
}

func (d *DiskIIDrive) SetTrackIndex(v int) {
	d.mcb.Write(DISKII_TRACKINDEX, uint64(v))
}

func (d *DiskIIDrive) GetTrackIndex() int {
	return int(d.mcb.Read(DISKII_TRACKINDEX))
}

func (d *DiskIIDrive) SetTrackPhase(v int) {
	d.mcb.Write(DISKII_TRACKPHASE, uint64(v))
}

func (d *DiskIIDrive) GetTrackPhase() int {
	return int(d.mcb.Read(DISKII_TRACKPHASE))
}

func (d *DiskIIDrive) SetHalfTrack(v int) {
	d.mcb.Write(DISKII_HALFTRACK, uint64(v))
}

func (d *DiskIIDrive) GetHalfTrack() int {
	return int(d.mcb.Read(DISKII_HALFTRACK))
}

func (d *DiskIIDrive) SetTrackStartOffset(v int) {
	d.mcb.Write(DISKII_TRACKSTARTOFFSET, uint64(v))
}

func (d *DiskIIDrive) GetTrackStartOffset() int {
	return int(d.mcb.Read(DISKII_TRACKSTARTOFFSET))
}

func (d *DiskIIDrive) SetNibbleOffset(v int) {
	d.mcb.Write(DISKII_NIBBLEOFFSET, uint64(v))
}

func (d *DiskIIDrive) GetNibbleOffset() int {
	return int(d.mcb.Read(DISKII_NIBBLEOFFSET))
}

func (d *DiskIIDrive) SetMagnets(v int) {
	d.mcb.Write(DISKII_MAGNETS, uint64(v))
}

func (d *DiskIIDrive) GetMagnets() int {
	return int(d.mcb.Read(DISKII_MAGNETS))
}

func (d *DiskIIDrive) SetSpinCount(v int) {
	d.mcb.Write(DISKII_SPINCOUNT, uint64(v))
}

func (d *DiskIIDrive) GetSpinCount() int {
	return int(d.mcb.Read(DISKII_SPINCOUNT))
}

func (d *DiskIIDrive) SetUseNib(v bool) {
	if v {
		d.mcb.Write(DISKII_USE_NIB, 1)
	} else {
		d.mcb.Write(DISKII_USE_NIB, 0)
	}
}

func (d *DiskIIDrive) UseNib() bool {
	return d.mcb.Read(DISKII_USE_NIB) != 0
}

func (d *DiskIIDrive) SetPhaseLastBump(v bool) {
	if v {
		d.mcb.Write(DISKII_PHASELASTBUMP, 1)
	} else {
		d.mcb.Write(DISKII_PHASELASTBUMP, 0)
	}
}

func (d *DiskIIDrive) GetPhaseLastBump() bool {
	return d.mcb.Read(DISKII_PHASELASTBUMP) != 0
}

func (d *DiskIIDrive) SetPhaseAlign(v bool) {
	if v {
		d.mcb.Write(DISKII_PHASEALIGN, 1)
	} else {
		d.mcb.Write(DISKII_PHASEALIGN, 0)
	}
}

func (d *DiskIIDrive) GetPhaseAlign() bool {
	return d.mcb.Read(DISKII_PHASEALIGN) != 0
}

func (d *DiskIIDrive) SetPhaseStop(v bool) {
	if v {
		d.mcb.Write(DISKII_PHASESTOP, 1)
	} else {
		d.mcb.Write(DISKII_PHASESTOP, 0)
	}
}

func (d *DiskIIDrive) GetPhaseStop() bool {
	return d.mcb.Read(DISKII_PHASESTOP) != 0
}

func (d *DiskIIDrive) SetBitPtr(v int) {

	// silent io
	if settings.RecordIgnoreIO[d.slotid] {
		d.mcb.GetMM().WriteInterpreterMemorySilent(d.slotid, d.baseaddr+DISKII_BITPTR, uint64(v))
		return
	}

	d.mcb.Write(DISKII_BITPTR, uint64(v))
}

func (d *DiskIIDrive) GetBitPtr() int {
	return int(d.mcb.Read(DISKII_BITPTR))
}

func (d *DiskIIDrive) SetWriteMode(v bool) {
	if v {
		d.mcb.Write(DISKII_WRITEMODE, 1)
	} else {
		d.mcb.Write(DISKII_WRITEMODE, 0)
	}
}

func (d *DiskIIDrive) GetWriteMode() bool {
	if d.mcb.Read(DISKII_WRITEMODE) != 0 {
		return true
	}
	return false
}

func (d *DiskIIDrive) SetDriveOn(v bool) {
	if v {
		d.mcb.Write(DISKII_DRIVEON, 1)
	} else {
		d.mcb.Write(DISKII_DRIVEON, 0)
	}
	d.parent.SetDriveEnabled(v)
}

func (d *DiskIIDrive) GetDriveOn() bool {
	return d.parent.GetDriveEnabled()
}

func (d *DiskIIDrive) SetDiskUpdatePending(v bool) {
	if v {
		d.mcb.Write(DISKII_UPDATE_PENDING, 1)
	} else {
		d.mcb.Write(DISKII_UPDATE_PENDING, 0)
	}
}

func (d *DiskIIDrive) GetDiskUpdatePending() bool {
	if d.mcb.Read(DISKII_UPDATE_PENDING) != 0 {
		return true
	}
	return false
}

func (d *DiskIIDrive) SetDirtyTracks(v [35]bool) {
	val := uint64(0)
	for i, dirty := range v {
		if dirty {
			val = val | (1 << uint(i))
		}
	}
	d.mcb.Write(DISKII_DIRTY_TRACKS, val)
}

func (d *DiskIIDrive) GetDirtyTracks() [35]bool {
	val := d.mcb.Read(DISKII_DIRTY_TRACKS)
	var out [35]bool
	for i, _ := range out {
		out[i] = (val & (1 << uint(i))) != 0
	}
	return out
}

func (d *DiskIIDrive) SetDirty(track int, v bool) {
	b := d.GetDirtyTracks()
	b[track] = v
	d.SetDirtyTracks(b)
}

func (d *DiskIIDrive) GetDirty(track int) bool {
	b := d.GetDirtyTracks()
	return b[track]
}

func NewDiskIIDrive(activationCallback func(b bool), mm *memory.MemoryMap, baseaddr int, drive int, parent *IOCardDiskII) *DiskIIDrive {
	index := parent.e.GetMemIndex()
	mcb := memory.NewMemoryControlBlock(mm, index, false)
	mcb.Add(
		mm.Data[index][baseaddr:baseaddr+DISKII_UNIT_SIZE],
		baseaddr,
	)

	return &DiskIIDrive{
		VolumeName:         "Empty",
		ActivationCallback: activationCallback,
		drive:              drive,
		baseaddr:           baseaddr,
		mcb:                mcb,
		lastQTrack:         -1,
		parent:             parent,
		slotid:             parent.e.GetMemIndex(),
	}

}

func (d *DiskIIDrive) ClearTrackState() {
	for i := 0; i < 35; i++ {
		d.SetDirty(i, false)
	}
}

func (d *DiskIIDrive) Reset() {

	//debug.PrintStack()

	d.SetDriveOn(false)
	d.SetMagnets(0)
	d.ClearTrackState()
	d.SetDiskUpdatePending(false)
	d.SetNibbleOffset(0)
	d.SetTrackStartOffset(0)
	d.VolumeName = "Empty"
	d.SetTrackIndex(0)
	d.SetTrackPhase(0)
	d.SetPhaseAlign(false)
	d.SetPhaseLastBump(false)
	d.SetPhaseControl(0)
	d.SetPhaseCycles(0)
}

func (d *DiskIIDrive) SetPendingDeactivation() {
	if d.ActivationCallback != nil {
		d.ActivationCallback(false)
	}
}

func (d *DiskIIDrive) SetOn(b bool) {
	if d.ActivationCallback != nil {
		d.ActivationCallback(b)
	}
	d.SetDriveOn(b)
}

func (d *DiskIIDrive) IsOn() bool {
	return d.GetDriveOn()
}

func (d *DiskIIDrive) SetReadMode() {
	d.SetWriteMode(false)
}

func (d *DiskIIDrive) SetWrite() {
	d.SetWriteMode(true)
}

func clrbit(value *int, index uint64) {
	bitmask := int(0xff ^ (1 << index))
	*value = *value & bitmask
}

func setbit(value *int, index uint64) {
	bitmask := int(1 << index)
	*value = *value | bitmask
}

func (d *DiskIIDrive) calcStepperDelta(phase int, phaseControl int) int {
	var nextPhase int

	switch phaseControl {
	case 0x1, 0xb:
		nextPhase = 0

		break

	case 0x3:
		nextPhase = 1

		break

	case 0x2, 0x7:
		nextPhase = 2

		break

	case 0x6:
		nextPhase = 3

		break

	case 0x4, 0xe:
		nextPhase = 4

		break

	case 0xc:
		nextPhase = 5

		break

	case 0x8, 0xd:
		nextPhase = 6

		break

	case 0x9:
		nextPhase = 7

		break
	default:
		nextPhase = phase
	}

	return ((nextPhase - phase + 4) & 0x7) - 4
}

func (d *DiskIIDrive) UpdatePhaseControl(register int) bool {
	//log.Printf("StepQ(%d)", register)

	pc := d.GetPhaseControl()
	opc := pc

	switch register & 1 {
	case 0:
		clrbit(&pc, uint64(register/2))
	case 1:
		setbit(&pc, uint64(register/2))
	}

	if opc != pc {
		d.SetPhaseControl(pc)
		return true
	}

	return false
}

func (d *DiskIIDrive) StepQ(register int) {

	if d.IsOn() {
		// drive is on at the moment...
		QTrack := d.GetTrackIndex()
		//log.Printf("Current Q-Track = %d", QTrack)
		delta := d.calcStepperDelta(QTrack&0x7, d.GetPhaseControl())
		//log.Printf("Stepper Delta = %d")
		newQTrackIndex := QTrack + delta

		if newQTrackIndex < 0 {
			newQTrackIndex = 0
		}
		if newQTrackIndex > 159 {
			newQTrackIndex = 159
		}

		//log.Printf("New Q-Track Index = %d", newQTrackIndex)

		if newQTrackIndex != QTrack {
			d.SetTrackIndex(newQTrackIndex)
			if d.Disk != nil {
				d.currentBitStream, d.currentBitCount = d.Disk.Load525QTrack(newQTrackIndex)
				//log.Printf("Move head from %d to %d", QTrack, newQTrackIndex)
				if !d.Disk.TrackOK() {
					//fmt.Println("Warning Track is nil")
				}
			}
		}
	}
	// }

}

func (d *DiskIIDrive) Step(register int) {
	// switch drive head stepper motor magnets on/off
	var magnet uint64 = uint64(register>>1) & 0x3
	d.SetMagnets(d.GetMagnets() & ^(1 << magnet))
	d.SetMagnets(d.GetMagnets() | ((register & 0x1) << magnet))

	// step the drive head according to stepper magnet changes
	if d.GetDriveOn() {
		delta := driveHeadStepDelta[d.GetHalfTrack()&0x3][d.GetMagnets()]
		if delta != 0 {
			newHalfTrack := d.GetHalfTrack() + delta
			if newHalfTrack < 0 {
				newHalfTrack = 0
				//fmt.Println("CHOCK!")
			} else if newHalfTrack > halfTrackCount {
				newHalfTrack = halfTrackCount
			}
			if newHalfTrack != d.GetHalfTrack() {
				if d.UseNib() {
					d.SetHalfTrack(newHalfTrack)
					d.SetTrackStartOffset((d.GetHalfTrack() >> 1) * disk.TRACK_NIBBLE_LENGTH)
					if d.GetTrackStartOffset() >= disk.DISK_NIBBLE_LENGTH {
						d.SetTrackStartOffset(disk.DISK_NIBBLE_LENGTH - disk.TRACK_NIBBLE_LENGTH)
					}
					d.SetNibbleOffset(0)
				} else {
					d.SetHalfTrack(newHalfTrack)
					if d.Disk != nil {
						d.currentBitStream, d.currentBitCount = d.Disk.Load525QTrack(2 * newHalfTrack)
						if !d.Disk.TrackOK() {
							//fmt.Println("Warning Track is nil")
						}
					}
				}
			}
		}
	}
}

func (d *DiskIIDrive) IsWriteProtected() bool {
	if d.UseNib() {
		if d.NibImage != nil {
			return d.NibImage.WriteProtected
		}
		return false
	}

	if d.Disk != nil {
		return d.Disk.WriteProtected()
	}
	return false
}

func (d *DiskIIDrive) SkipBits(count int64) {
	if d.Disk != nil && d.Disk.TrackOK() {
		d.Disk.AdvanceN(int(count))
	}
}

func (d *DiskIIDrive) IsBitstream() bool {
	return d.Disk != nil
}

func (d *DiskIIDrive) ReadBit() byte {

	if d.Disk == nil {
		return 1
	}

	if !d.Disk.TrackOK() {
		log.Printf("Track invalid resetting to 0")
		d.SetTrackIndex(0)
		d.currentBitStream, d.currentBitCount = d.Disk.Load525QTrack(0)
	}

	d.checkQuarterTrack()

	bitptr := d.GetBitPtr()

	// Move read bitstream inline for speed
	i := bitptr % int(d.currentBitCount)
	byteIdx := i / 8
	bitIdx := uint(7 - (i % 8))
	value := (d.currentBitStream[byteIdx] >> bitIdx) & 1

	bitptr++
	if bitptr >= int(d.currentBitCount) {
		bitptr = 0
	}
	d.SetBitPtr(bitptr)

	if value == 1 {
		d.zeroCount = 0
	} else {
		d.zeroCount++
		if d.zeroCount > 3 {
			if rand.Intn(255) >= 192 {
				value = 1
			}
		}
	}

	return value
}

func (d *DiskIIDrive) SetLatch(v byte) {
	d.mcb.Write(DISKII_DRIVE_LATCH, uint64(v))
}

func (d *DiskIIDrive) GetLatch() byte {
	return byte(d.mcb.Read(DISKII_DRIVE_LATCH))
}

func (d *DiskIIDrive) SetLatchValue(v byte) {
	if d.GetWriteMode() {
		d.SetLatch(v)
	} else {
		d.SetLatch(0xff)
	}
}

func (d *DiskIIDrive) ReadLatch() byte {
	var result byte = 0x07f
	if !d.GetWriteMode() {
		d.SetSpinCount((d.GetSpinCount() + 1) & 0x0F)
		if d.GetSpinCount() > 0 {
			if d.NibImage != nil {

				//t, s := d.NibImage.NibbleOffsetToTS(d.GetTrackStartOffset() + d.GetNibbleOffset())
				//log.Printf("Read nibble at 0x%x value 0x%.2x from track %d, sector %d", d.GetTrackStartOffset()+d.GetNibbleOffset(), result, t, s)

				result = d.GetNibble(d.GetTrackStartOffset() + d.GetNibbleOffset())

				if d.IsOn() {
					d.SetNibbleOffset(d.GetNibbleOffset() + 1)
					if d.GetNibbleOffset() >= disk.TRACK_NIBBLE_LENGTH {
						d.SetNibbleOffset(0)
					}
				}
			} else {
				result = 0x0ff
			}
		}
	} else {
		d.SetSpinCount((d.GetSpinCount() + 1) & 0x0F)
		if d.GetSpinCount() > 0 {
			result = 0x080
		}
	}
	return result
}

func (d *DiskIIDrive) Write() {

	if d.GetWriteMode() {
		// for d.GetDiskUpdatePending() {
		// 	// If another thread requested writes to block (e.g. because of disk activity), wait for it to finish!
		// 	time.Sleep(1 * time.Millisecond)
		// }
		if d.NibImage != nil {
			// Do nothing if write-protection is enabled!
			if !d.NibImage.WriteProtected {
				d.SetDirty(d.GetTrackStartOffset()/disk.TRACK_NIBBLE_LENGTH, true)

				//d.DirtyTracks[d.TrackStartOffset/disk.TRACK_NIBBLE_LENGTH] = true
				d.SetNibble(d.GetTrackStartOffset()+d.GetNibbleOffset(), d.GetLatch())
				d.SetDiskUpdatePending(true)

				//fmt.Printf("WRITE nibble 0x%.2x at position %d\n", d.Latch, d.TrackStartOffset+d.NibbleOffset)

				d.SetNibbleOffset(d.GetNibbleOffset() + 1)
				//d.TriggerDiskUpdate()
				//StateManager.markDirtyValue(disk.nibbles, computer);
			}
		}

		if d.GetNibbleOffset() >= disk.TRACK_NIBBLE_LENGTH {
			d.SetNibbleOffset(0)
		}
	}
}

func (d *DiskIIDrive) WriteBit(bit byte) {

	if d.Disk == nil || !d.Disk.TrackOK() {
		return
	}

	d.checkQuarterTrack()

	//ptr := (int(d.Disk.Track.BitCount()) + d.Disk.BitPtr - 3) % int(d.Disk.Track.BitCount())
	ptr := d.GetBitPtr()

	// update local copy as well (makes write slightly more costly)
	i := ptr % int(d.currentBitCount)
	byteIdx := i / 8
	bitIdx := uint(7 - (i % 8))
	v := d.currentBitStream[byteIdx]
	if bit == 0 {
		v = v & ^(1 << bitIdx)
	} else {
		v = v | (1 << bitIdx)
	}
	d.currentBitStream[byteIdx] = v

	// Write through to memory backed storage for real time sync
	d.Disk.WriteByte(byteIdx, v)
	d.Disk.SetModified(true)
}

func (d *DiskIIDrive) InsertDisk(filename string, wp bool) {

	var fp filerecord.FileRecord
	var e error

	if filename == "" {
		switch d.drive {
		case 0:
			settings.PureBootVolume[d.slotid] = ""
		case 1:
			settings.PureBootVolume2[d.slotid] = ""
		}
		d.Disk = woz.CreateWOZEmpty(d.GetWBuffer())
		return
	}

	if strings.HasPrefix(filename, "local:") {
		lfilename := strings.Replace(filename, "local:", "", -1)
		fp.Content, e = ioutil.ReadFile(lfilename)
	} else {
		fp, e = files.ReadBytesViaProvider(files.GetPath(filename), files.GetFilename(filename))
	}

	if e != nil {
		//fmt.Println("Disk insert FAILED: " + e.Error())
		return
	}

	buffer := d.GetWBuffer() // disk buffer

	var needReboot = false

	d.SetUseNib(false)
	dsk, e := woz.NewWOZImage(bytes.NewBuffer(fp.Content), buffer)
	if e != nil {

		// try woz2
		dsk2, e := woz2.NewWOZ2Image(bytes.NewBuffer(fp.Content), buffer)
		if e != nil {

			if strings.HasSuffix(filename, ".nib") {
				// dsk = woz.CreateWOZFromNIB(data, buffer)
				// fmt.Println("WOZ from NIB: " + dsk.INFO.String())
				nib, err := disk.NewDSKWrapperBin(d, fp.Content, filename)
				if err != nil {
					return
				}
				d.NibImage = nib
				d.SetUseNib(true)
			} else {
				size := len(fp.Content)
				size = (size / 256) * 256
				//log.Println("Disk insert FAILED: " + e.Error())
				tmp, e := disk.NewDSKWrapperBin(d, fp.Content[:size], fp.FileName)
				if e == nil {

					log.Printf("settings.PreserveDSK = %v", settings.PreserveDSK)

					if settings.PreserveDSK {
						log.Printf("DiskII: using nib for DSK file")
						d.NibImage = tmp
						d.SetUseNib(true)
					} else {
						log.Printf("DiskII: using woz for DSK file")
						dsk = woz.CreateWOZFromDSK(tmp, buffer)
						d.Disk = dsk
					}

					//fmt.Println("WOZ from DSK: " + dsk.INFO.String())
				} else {
					return
				}
			}

		} else {
			d.Disk = dsk2
		}
	} else {
		d.Disk = dsk
	}

	parts := strings.Split(filename, "/")

	d.VolumeName = parts[len(parts)-1]
	d.ClearTrackState()
	//fmt.Println("Disk insert OK")
	switch d.drive {
	case 0:
		settings.PureBootVolume[d.slotid] = filename
	case 1:
		settings.PureBootVolume2[d.slotid] = filename
	}

	if needReboot && d.drive == 0 {
		d.parent.e.GetMemoryMap().IntSetSlotRestart(d.parent.e.GetMemIndex(), true)
	}
}

func (d *DiskIIDrive) InsertDiskWrapper(dsk *disk.DSKWrapper) {

	panic("InsertDiskWrapper needs to be implemented")

}

func (d *DiskIIDrive) InsertDiskBin(data []byte, filename string, wp bool) {

	buffer := d.GetWBuffer() // disk buffer

	var needReboot = false

	d.SetUseNib(false)
	dsk, e := woz.NewWOZImage(bytes.NewBuffer(data), buffer)
	if e != nil {

		dsk2, e := woz2.NewWOZ2Image(bytes.NewBuffer(data), buffer)

		if e != nil {

			if strings.HasSuffix(filename, ".nib") {
				// dsk = woz.CreateWOZFromNIB(data, buffer)
				// fmt.Println("WOZ from NIB: " + dsk.INFO.String())
				nib, err := disk.NewDSKWrapperBin(d, data, filename)
				if err != nil {
					return
				}
				d.NibImage = nib
				d.SetUseNib(true)
			} else {
				//log.Println("Disk insert FAILED: " + e.Error())
				tmp, e := disk.NewDSKWrapperBin(d, data, filename)
				if e == nil {
					log.Printf("settings.PreserveDSK = %v", settings.PreserveDSK)
					if settings.PreserveDSK {
						log.Printf("DiskII: using nib for DSK bytes")
						d.NibImage = tmp
						d.SetUseNib(true)
					} else {
						log.Printf("DiskII: using WOZ for disk bytes")
						dsk = woz.CreateWOZFromDSK(tmp, buffer)
						d.Disk = dsk
					}
					//fmt.Println("WOZ from nibbles: " + dsk.INFO.String())
				} else if strings.HasSuffix(filename, ".nib") {
					//dsk = woz.CreateWOZFromNIB(data, buffer)
					//fmt.Println("WOZ from NIB: " + dsk.INFO.String())
					d.NibImage = tmp
					d.SetUseNib(true)
				} else {
					return
				}
			}
		} else {
			d.Disk = dsk2
		}
	} else {
		d.Disk = dsk
	}

	parts := strings.Split(filename, "/")

	d.VolumeName = parts[len(parts)-1]
	d.ClearTrackState()
	//fmt.Println("Disk insert OK")

	switch d.drive {
	case 0:
		settings.PureBootVolume[d.slotid] = filename
	case 1:
		settings.PureBootVolume2[d.slotid] = filename
	}

	if needReboot && d.drive == 0 {
		d.parent.e.GetMemoryMap().IntSetSlotRestart(d.parent.e.GetMemIndex(), true)
	}
}

func (d *DiskIIDrive) HasDisk() bool {
	return (d.Disk != nil)
}

func (d *DiskIIDrive) SetWriteProtect(b bool) {
	if d.Disk != nil {
		d.Disk.SetWriteProtected(b)
	}
}

func (d *DiskIIDrive) Eject() {
	d.Disk = nil
	d.NibImage = nil
	d.SetUseNib(false)
	d.VolumeName = "Empty"
}
