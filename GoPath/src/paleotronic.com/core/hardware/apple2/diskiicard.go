package apple2

import (
	"bytes"
	"encoding/binary"
	//rlog "log"
	//log2 "log"
	"math"
	"os"
	"strings"
	"sync"

	"paleotronic.com/core/hardware/apple2/woz"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/hardware/servicebus"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/settings"
	"paleotronic.com/disk"
	"paleotronic.com/files"
	"paleotronic.com/fmt"
	"paleotronic.com/log"
)

const DRIVE_PHASE_CYCLES = 510

var romP6 = [256]byte{
	0x18, 0xd8, 0x18, 0x08, 0x0a, 0x0a, 0x0a, 0x0a,
	0x18, 0x39, 0x18, 0x39, 0x18, 0x3b, 0x18, 0x3b,
	0x18, 0x38, 0x18, 0x28, 0x0a, 0x0a, 0x0a, 0x0a,
	0x18, 0x39, 0x18, 0x39, 0x18, 0x3b, 0x18, 0x3b,
	0x2d, 0xd8, 0x38, 0x48, 0x0a, 0x0a, 0x0a, 0x0a,
	0x28, 0x48, 0x28, 0x48, 0x28, 0x48, 0x28, 0x48,
	0x2d, 0x48, 0x38, 0x48, 0x0a, 0x0a, 0x0a, 0x0a,
	0x28, 0x48, 0x28, 0x48, 0x28, 0x48, 0x28, 0x48,
	0xd8, 0xd8, 0xd8, 0xd8, 0x0a, 0x0a, 0x0a, 0x0a,
	0x58, 0x78, 0x58, 0x78, 0x58, 0x78, 0x58, 0x78,
	0x58, 0x78, 0x58, 0x78, 0x0a, 0x0a, 0x0a, 0x0a,
	0x58, 0x78, 0x58, 0x78, 0x58, 0x78, 0x58, 0x78,
	0xd8, 0xd8, 0xd8, 0xd8, 0x0a, 0x0a, 0x0a, 0x0a,
	0x68, 0x08, 0x68, 0x88, 0x68, 0x08, 0x68, 0x88,
	0x68, 0x88, 0x68, 0x88, 0x0a, 0x0a, 0x0a, 0x0a,
	0x68, 0x08, 0x68, 0x88, 0x68, 0x08, 0x68, 0x88,
	0xd8, 0xcd, 0xd8, 0xd8, 0x0a, 0x0a, 0x0a, 0x0a,
	0x98, 0xb9, 0x98, 0xb9, 0x98, 0xbb, 0x98, 0xbb,
	0x98, 0xbd, 0x98, 0xb8, 0x0a, 0x0a, 0x0a, 0x0a,
	0x98, 0xb9, 0x98, 0xb9, 0x98, 0xbb, 0x98, 0xbb,
	0xd8, 0xd9, 0xd8, 0xd8, 0x0a, 0x0a, 0x0a, 0x0a,
	0xa8, 0xc8, 0xa8, 0xc8, 0xa8, 0xc8, 0xa8, 0xc8,
	0x29, 0x59, 0xa8, 0xc8, 0x0a, 0x0a, 0x0a, 0x0a,
	0xa8, 0xc8, 0xa8, 0xc8, 0xa8, 0xc8, 0xa8, 0xc8,
	0xd9, 0xfd, 0xd8, 0xf8, 0x0a, 0x0a, 0x0a, 0x0a,
	0xd8, 0xf8, 0xd8, 0xf8, 0xd8, 0xf8, 0xd8, 0xf8,
	0xd9, 0xfd, 0xa0, 0xf8, 0x0a, 0x0a, 0x0a, 0x0a,
	0xd8, 0xf8, 0xd8, 0xf8, 0xd8, 0xf8, 0xd8, 0xf8,
	0xd8, 0xdd, 0xe8, 0xe0, 0x0a, 0x0a, 0x0a, 0x0a,
	0xe8, 0x88, 0xe8, 0x08, 0xe8, 0x88, 0xe8, 0x08,
	0x08, 0x4d, 0xe8, 0xe0, 0x0a, 0x0a, 0x0a, 0x0a,
	0xe8, 0x88, 0xe8, 0x08, 0xe8, 0x88, 0xe8, 0x08,
}

const (
	DISKII_PENDING_WRITE = 0 //
	DISKII_LATCH         = 1 //
	DISKII_WRITE_ENABLED = 2 //
	DISKII_STATE         = 3 //
	DISKII_Q6            = 4 //
	DISKII_CURRENT       = 5 //
	DISKII_CLOCK         = 6
	DISKII_DRIVE_ENABLED = 7
)

func bool2byte(b bool) byte {
	if b {
		return 1
	}
	return 0
}

func byte2bool(b byte) bool {
	if b != 0 {
		return true
	}
	return false
}

func int2bytes(i int) []byte {
	b := bytes.NewBuffer(nil)
	if binary.Write(b, binary.LittleEndian, i) == nil {
		return b.Bytes()
	}
	return []byte{}
}

func int642bytes(i int64) []byte {
	b := bytes.NewBuffer(nil)
	if binary.Write(b, binary.LittleEndian, i) == nil {
		return b.Bytes()
	}
	return []byte{}
}

func bytes2int(b []byte) int {
	var i int
	binary.Read(bytes.NewBuffer(b), binary.LittleEndian, &i)
	return i
}

func bytes2int64(b []byte) int64 {
	var i int64
	binary.Read(bytes.NewBuffer(b), binary.LittleEndian, &i)
	return i
}

func (io *IOCardDiskII) ToBytes() []byte {
	return append([]byte{
		io.GetPendingWrite(),
		io.GetDataRegister(),
		bool2byte(io.GetWriteMode()),
		byte(io.GetState()),
		byte(io.GetQ6()),
		byte(io.GetCurrent()),
		bool2byte(io.GetDriveEnabled())},
		int642bytes(io.currentCycles)...)
}

func (io *IOCardDiskII) FromBytes(b []byte) {
	io.SetPendingWrite(b[0])
	io.SetDataRegister(b[1])
	io.SetWriteMode(byte2bool(b[2]))
	io.SetState(int(b[3]))
	io.SetQ6(int(b[4]))
	io.SetCurrent(int(b[5]))
	io.SetDriveEnabled(byte2bool(b[6]))
	c := bytes2int64(b[7:])
	io.currentCycles = c
	if io.GetDriveEnabled() {
		io.driveOffCycles = io.currentCycles + 2*1020484
		io.driveSlowdownCycles = io.currentCycles + 250000
	}
}

type IOCardDiskII struct {
	IOCard
	//current             int
	disks               [2]*DiskIIDrive
	e                   interfaces.Interpretable
	lastLSSCycles       float64
	lastCycles          int64
	currentCycles       int64
	currentLSSCycles    float64
	driveOffCycles      int64
	driveSlowdownCycles int64
	driveBlinkCycles    int64
	//dataRegister        byte
	driveEnabled bool
	//clock        int
	//state        int
	//q6 int
	//writeMode bool
	readMode bool
	//pendingWrite        byte
	timerCycles   int64
	timerFunction func()
	timerParams   []interface{}
	tm            sync.Mutex
	lssPulse      byte
	lssCycleRatio float64
}

func NewIOCardDiskII(mm *memory.MemoryMap, e interfaces.Interpretable, index int, acb func(b bool)) *IOCardDiskII {
	this := &IOCardDiskII{
		e: e,
	}
	this.SetMemory(mm, index)
	this.Name = "IOCardDiskII"
	this.ROM = [256]uint64{
		0xa2, 0x20, 0xa0, 0x00, 0xa2, 0x03, 0x86, 0x3c,
		0x8a, 0x0a, 0x24, 0x3c, 0xf0, 0x10, 0x05, 0x3c,
		0x49, 0xff, 0x29, 0x7e, 0xb0, 0x08, 0x4a, 0xd0,
		0xfb, 0x98, 0x9d, 0x56, 0x03, 0xc8, 0xe8, 0x10,
		0xe5, 0x20, 0x58, 0xff, 0xba, 0xbd, 0x00, 0x01,
		0x0a, 0x0a, 0x0a, 0x0a, 0x85, 0x2b, 0xaa, 0xbd,
		0x8e, 0xc0, 0xbd, 0x8c, 0xc0, 0xbd, 0x8a, 0xc0,
		0xbd, 0x89, 0xc0, 0xa0, 0x50, 0xbd, 0x80, 0xc0,
		0x98, 0x29, 0x03, 0x0a, 0x05, 0x2b, 0xaa, 0xbd,
		0x81, 0xc0, 0xa9, 0x56, 0xa9, 0x00, 0xea, 0x88,
		0x10, 0xeb, 0x85, 0x26, 0x85, 0x3d, 0x85, 0x41,
		0xa9, 0x08, 0x85, 0x27, 0x18, 0x08, 0xbd, 0x8c,
		0xc0, 0x10, 0xfb, 0x49, 0xd5, 0xd0, 0xf7, 0xbd,
		0x8c, 0xc0, 0x10, 0xfb, 0xc9, 0xaa, 0xd0, 0xf3,
		0xea, 0xbd, 0x8c, 0xc0, 0x10, 0xfb, 0xc9, 0x96,
		0xf0, 0x09, 0x28, 0x90, 0xdf, 0x49, 0xad, 0xf0,
		0x25, 0xd0, 0xd9, 0xa0, 0x03, 0x85, 0x40, 0xbd,
		0x8c, 0xc0, 0x10, 0xfb, 0x2a, 0x85, 0x3c, 0xbd,
		0x8c, 0xc0, 0x10, 0xfb, 0x25, 0x3c, 0x88, 0xd0,
		0xec, 0x28, 0xc5, 0x3d, 0xd0, 0xbe, 0xa5, 0x40,
		0xc5, 0x41, 0xd0, 0xb8, 0xb0, 0xb7, 0xa0, 0x56,
		0x84, 0x3c, 0xbc, 0x8c, 0xc0, 0x10, 0xfb, 0x59,
		0xd6, 0x02, 0xa4, 0x3c, 0x88, 0x99, 0x00, 0x03,
		0xd0, 0xee, 0x84, 0x3c, 0xbc, 0x8c, 0xc0, 0x10,
		0xfb, 0x59, 0xd6, 0x02, 0xa4, 0x3c, 0x91, 0x26,
		0xc8, 0xd0, 0xef, 0xbc, 0x8c, 0xc0, 0x10, 0xfb,
		0x59, 0xd6, 0x02, 0xd0, 0x87, 0xa0, 0x00, 0xa2,
		0x56, 0xca, 0x30, 0xfb, 0xb1, 0x26, 0x5e, 0x00,
		0x03, 0x2a, 0x5e, 0x00, 0x03, 0x2a, 0x91, 0x26,
		0xc8, 0xd0, 0xee, 0xe6, 0x27, 0xe6, 0x3d, 0xa5,
		0x3d, 0xcd, 0x00, 0x08, 0xa6, 0x2b, 0x90, 0xdb,
		0x4c, 0x01, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00,
	}

	for i, _ := range this.disks {
		this.disks[i] = NewDiskIIDrive(acb, mm, mm.MEMBASE(index)+memory.OCTALYZER_DISKII_BASE+(i*memory.OCTALYZER_DISKII_SIZE), i, this)
	}
	this.lastCycles = this.getCycles()

	// ServiceBus subscriptions
	servicebus.Subscribe(this.e.GetMemIndex(), servicebus.DiskIIExchangeDisks, this)
	servicebus.Subscribe(this.e.GetMemIndex(), servicebus.DiskIIEject, this)
	servicebus.Subscribe(this.e.GetMemIndex(), servicebus.DiskIIInsertBlank, this)
	servicebus.Subscribe(this.e.GetMemIndex(), servicebus.DiskIIInsertBytes, this)
	servicebus.Subscribe(this.e.GetMemIndex(), servicebus.DiskIIInsertFilename, this)
	servicebus.Subscribe(this.e.GetMemIndex(), servicebus.DiskIIToggleWriteProtect, this)
	servicebus.Subscribe(this.e.GetMemIndex(), servicebus.DiskIIQueryWriteProtect, this)

	return this
}

func (d *IOCardDiskII) IsBitstream() bool {
	return d.disks[0].IsBitstream() || d.disks[1].IsBitstream()
}

func (d *IOCardDiskII) SetLSSRatio(r float64) {
	d.lssCycleRatio = r
}

func (d *IOCardDiskII) ClearTimer() {
	d.tm.Lock()
	defer d.tm.Unlock()
	d.timerCycles = 0
	d.timerFunction = nil
	d.timerParams = []interface{}{}
}

func (d *IOCardDiskII) SetTimer(cycles int64, f func(), values ...interface{}) {
	d.tm.Lock()
	defer d.tm.Unlock()

	//rlog.Printf("Setting timer")

	d.timerCycles = cycles
	d.timerFunction = f
	d.timerParams = values
}

func (d *IOCardDiskII) CheckTimer() {
	if d.timerCycles == 0 || d.timerFunction == nil {
		return
	}
	if d.timerCycles > d.currentCycles {
		return
	}
	d.timerCycles = 0
	//rlog.Printf("Triggering timer")
	d.timerFunction()
	d.timerFunction = nil
}

func (d *IOCardDiskII) UpdatePhase() {
	reg := d.timerParams[0].(int)
	current := d.GetCurrent()
	d.GetDrive(current).StepQ(reg)
}

func (d *IOCardDiskII) HandleServiceBusRequest(r *servicebus.ServiceBusRequest) (*servicebus.ServiceBusResponse, bool) {

	log.Printf("Got ServiceBusRequest(%s)", r)

	index := d.e.GetMemIndex()

	switch r.Type {
	case servicebus.Clocks6502Update:
		d.AdjustClock(r.Payload.(int))
		return nil, false

	case servicebus.Cycles6502Update:
		d.Increment(r.Payload.(int))

	case servicebus.DiskIIFlush:
		d.CheckUpdate()

	case servicebus.DiskIIEject:
		drive := r.Payload.(int)
		d.CheckUpdate()
		d.GetDrive(drive % 2).Eject()
		switch drive {
		case 0:
			settings.PureBootVolume[index] = ""
		case 1:
			settings.PureBootVolume2[index] = ""
		}

	case servicebus.DiskIIExchangeDisks:
		log.Printf("Swapping disks")
		d.CheckUpdate()
		v0 := settings.PureBootVolume[index]
		d0 := d.GetDisk(0)
		n0 := d.disks[0].NibImage
		un0 := d.disks[0].UseNib()
		wp0 := d.disks[0].IsWriteProtected()
		v1 := settings.PureBootVolume2[index]
		d1 := d.GetDisk(1)
		n1 := d.disks[1].NibImage
		un1 := d.disks[1].UseNib()
		wp1 := d.disks[1].IsWriteProtected()

		d.SetDisk(0, d1)
		d.disks[0].NibImage = n1
		d.disks[0].SetUseNib(un1)
		settings.PureBootVolume[index] = v1
		d.SetDisk(1, d0)
		d.disks[1].NibImage = n0
		d.disks[1].SetUseNib(un0)
		settings.PureBootVolume2[index] = v0

		d.disks[0].InsertDisk(v1, wp1)
		d.disks[1].InsertDisk(v0, wp0)

	case servicebus.DiskIIInsertBlank:
		// TODO: Need to maybe make a WOZ2 default here in future
		d.CheckUpdate()
		t := r.Payload.(servicebus.DiskTargetString)
		d.GetDrive(t.Drive % 2).Eject()
		w := woz.CreateWOZEmpty(d.GetDrive(t.Drive % 2).GetWBuffer())
		d.SetDisk(t.Drive%2, w)
		switch t.Drive {
		case 0:
			settings.PureBootVolume[index] = t.Filename
		case 1:
			settings.PureBootVolume2[index] = t.Filename
		}

	case servicebus.DiskIIInsertFilename:
		d.CheckUpdate()
		t := r.Payload.(servicebus.DiskTargetString)
		drive := t.Drive
		log.Printf("Insert disk to drive %d", t.Drive)
		d.GetDrive(drive % 2).Eject()
		d.GetDrive(drive%2).InsertDisk(t.Filename, false)
		switch t.Drive {
			case 0:
				settings.PureBootVolume[index] = t.Filename
			case 1:
				settings.PureBootVolume2[index] = t.Filename
		}

	case servicebus.DiskIIInsertBytes:
		d.CheckUpdate()
		t := r.Payload.(servicebus.DiskTargetBytes)
		drive := t.Drive
		log.Printf("Insert disk to drive %d", t.Drive)
		d.GetDrive(drive % 2).Eject()
		d.GetDrive(drive%2).InsertDiskBin(t.Bytes, t.Filename, false)
		switch t.Drive {
			case 0:
				settings.PureBootVolume[index] = t.Filename
			case 1:
				settings.PureBootVolume2[index] = t.Filename
		}

	case servicebus.DiskIIToggleWriteProtect:
		drive := r.Payload.(int)
		disk := d.GetDrive(drive).Disk
		if disk != nil {
			disk.SetWriteProtected(!disk.WriteProtected())
		}
		nib := d.GetDrive(drive).NibImage
		if nib != nil {
			nib.WriteProtected = !nib.WriteProtected
		}

	case servicebus.DiskIIQueryWriteProtect:

		dr := d.GetDrive(r.Payload.(int) % 2)
		var state bool
		if dr.UseNib() {
			state = dr.NibImage != nil && dr.NibImage.WriteProtected
		} else {
			state = dr.Disk != nil && dr.Disk.WriteProtected()
		}

		return &servicebus.ServiceBusResponse{
			Payload: state,
		}, true

	}

	return &servicebus.ServiceBusResponse{
		Payload: true,
	}, true

}

func (d *IOCardDiskII) GetDrive(i int) *DiskIIDrive {
	return d.disks[i%2]
}

func (d *IOCardDiskII) SetCurrent(c int) {
	//log.Printf("Set current drive %d", c)
	if settings.RecordIgnoreIO[d.index] {
		d.mm.WriteInterpreterMemorySilent(d.index, memory.OCTALYZER_DISKII_CARD_STATE+DISKII_CURRENT, uint64(c))
		return
	}
	d.mm.WriteInterpreterMemory(d.index, memory.OCTALYZER_DISKII_CARD_STATE+DISKII_CURRENT, uint64(c))
}

func (d *IOCardDiskII) GetCurrent() int {
	return int(d.mm.ReadInterpreterMemorySilent(d.index, memory.OCTALYZER_DISKII_CARD_STATE+DISKII_CURRENT))
}

func (d *IOCardDiskII) SetPendingWrite(c byte) {
	if settings.RecordIgnoreIO[d.index] {
		d.mm.WriteInterpreterMemorySilent(d.index, memory.OCTALYZER_DISKII_CARD_STATE+DISKII_PENDING_WRITE, uint64(c))
		return
	}
	d.mm.WriteInterpreterMemory(d.index, memory.OCTALYZER_DISKII_CARD_STATE+DISKII_PENDING_WRITE, uint64(c))
}

func (d *IOCardDiskII) GetPendingWrite() byte {
	return byte(d.mm.ReadInterpreterMemorySilent(d.index, memory.OCTALYZER_DISKII_CARD_STATE+DISKII_PENDING_WRITE))
}

func (d *IOCardDiskII) SetDataRegister(c byte) {
	if settings.RecordIgnoreIO[d.index] {
		d.mm.WriteInterpreterMemorySilent(d.index, memory.OCTALYZER_DISKII_CARD_STATE+DISKII_LATCH, uint64(c))
		return
	}
	d.mm.WriteInterpreterMemory(d.index, memory.OCTALYZER_DISKII_CARD_STATE+DISKII_LATCH, uint64(c))
}

func (d *IOCardDiskII) GetDataRegister() byte {
	return byte(d.mm.ReadInterpreterMemorySilent(d.index, memory.OCTALYZER_DISKII_CARD_STATE+DISKII_LATCH))
}

func (d *IOCardDiskII) SetState(c int) {
	if settings.RecordIgnoreIO[d.index] {
		d.mm.WriteInterpreterMemorySilent(d.index, memory.OCTALYZER_DISKII_CARD_STATE+DISKII_STATE, uint64(c))
		return
	}
	d.mm.WriteInterpreterMemory(d.index, memory.OCTALYZER_DISKII_CARD_STATE+DISKII_STATE, uint64(c))
}

func (d *IOCardDiskII) GetState() int {
	return int(d.mm.ReadInterpreterMemorySilent(d.index, memory.OCTALYZER_DISKII_CARD_STATE+DISKII_STATE))
}

func (d *IOCardDiskII) SetWriteMode(c bool) {
	v := 0
	if c {
		v = 1
	}
	if settings.RecordIgnoreIO[d.index] {
		d.mm.WriteInterpreterMemorySilent(d.index, memory.OCTALYZER_DISKII_CARD_STATE+DISKII_WRITE_ENABLED, uint64(v))
		return
	}
	d.mm.WriteInterpreterMemory(d.index, memory.OCTALYZER_DISKII_CARD_STATE+DISKII_WRITE_ENABLED, uint64(v))
}

func (d *IOCardDiskII) GetWriteMode() bool {
	return d.mm.ReadInterpreterMemorySilent(d.index, memory.OCTALYZER_DISKII_CARD_STATE+DISKII_WRITE_ENABLED) != 0
}

func (d *IOCardDiskII) SetDriveEnabled(c bool) {
	v := 0
	if c {
		v = 1
	}
	if settings.RecordIgnoreIO[d.index] {
		d.mm.WriteInterpreterMemorySilent(d.index, memory.OCTALYZER_DISKII_CARD_STATE+DISKII_DRIVE_ENABLED, uint64(v))
		return
	}
	d.mm.WriteInterpreterMemory(d.index, memory.OCTALYZER_DISKII_CARD_STATE+DISKII_DRIVE_ENABLED, uint64(v))
}

func (d *IOCardDiskII) GetDriveEnabled() bool {
	return d.mm.ReadInterpreterMemorySilent(d.index, memory.OCTALYZER_DISKII_CARD_STATE+DISKII_DRIVE_ENABLED) != 0
}

func (d *IOCardDiskII) SetQ6(c int) {
	if settings.RecordIgnoreIO[d.index] {
		d.mm.WriteInterpreterMemorySilent(d.index, memory.OCTALYZER_DISKII_CARD_STATE+DISKII_Q6, uint64(c))
		return
	}
	d.mm.WriteInterpreterMemory(d.index, memory.OCTALYZER_DISKII_CARD_STATE+DISKII_Q6, uint64(c))
}

func (d *IOCardDiskII) GetQ6() int {
	return int(d.mm.ReadInterpreterMemorySilent(d.index, memory.OCTALYZER_DISKII_CARD_STATE+DISKII_Q6))
}

func (d *IOCardDiskII) SetClock(c int) {
	if settings.RecordIgnoreIO[d.index] {
		d.mm.WriteInterpreterMemorySilent(d.index, memory.OCTALYZER_DISKII_CARD_STATE+DISKII_CLOCK, uint64(c))
		return
	}
	d.mm.WriteInterpreterMemory(d.index, memory.OCTALYZER_DISKII_CARD_STATE+DISKII_CLOCK, uint64(c))
}

func (d *IOCardDiskII) GetClock() int {
	return int(d.mm.ReadInterpreterMemorySilent(d.index, memory.OCTALYZER_DISKII_CARD_STATE+DISKII_CLOCK))
}

func (d *IOCardDiskII) Init(slot int) {
	d.IOCard.Init(slot)
	//d.Log("Initialising drives...")
	d.SetDataRegister(0)
	d.SetCurrent(0)
	d.SetClock(0)
	d.SetQ6(0)
	d.SetState(0)

	// d.e.SetCycleCounter(d)
	// servicebus.Subscribe(
	// 	d.index,
	// 	servicebus.Cycles6502Update,
	// 	d,
	// )
	// servicebus.Subscribe(
	// 	d.index,
	// 	servicebus.Clocks6502Update,
	// 	d,
	// )

	//d.SetLSSRatio(2.28571428571)
	d.SetLSSRatio(2 * 32 / float64(settings.OptimalBitTiming))
}

func (d *IOCardDiskII) CheckUpdate() {

	var denibbleOnSave = true

	for disk := 0; disk < 2; disk++ {

		if d.disks[disk].UseNib() {

			if d.disks[disk].NibImage == nil || !d.disks[disk].GetDiskUpdatePending() {
				continue
			}

			var volume string
			switch disk {
			case 0:
				volume = settings.PureBootVolume[d.e.GetMemIndex()]
			case 1:
				volume = settings.PureBootVolume2[d.e.GetMemIndex()]
			}

			log.Printf("Current volume: %s", volume)

			if !settings.PreserveDSK || !denibbleOnSave {
				if !strings.HasSuffix(volume, ".nib") {
					p := strings.Split(volume, ".")
					p[len(p)-1] = "nib"
					volume = strings.Join(p, ".")

					log.Printf("Updating volume to %s...", volume)
					switch disk {
					case 0:
						settings.PureBootVolume[d.e.GetMemIndex()] = volume
					case 1:
						settings.PureBootVolume2[d.e.GetMemIndex()] = volume
					}
				}
			}

			if strings.HasPrefix(volume, "/appleii") || strings.HasPrefix(volume, "appleii") {
				volume = "/local/mydisks/" + files.GetFilename(volume)
				switch disk {
				case 0:
					settings.PureBootVolume[d.e.GetMemIndex()] = volume
				case 1:
					settings.PureBootVolume2[d.e.GetMemIndex()] = volume
				}
			}

			if strings.HasPrefix(volume, "local:") {
				fn := strings.Replace(volume, "local:", "", -1)
				if settings.PreserveDSK && denibbleOnSave {
					var buffer []byte
					var err error
					switch d.disks[disk].NibImage.Format.SPT() {
					case 16:
						buffer, err = woz.DeNibblizeImage(d.disks[disk].NibImage.GetNibbles(), d.disks[disk].NibImage.CurrentSectorOrder)
					case 13:
						buffer, err = woz.DeNibblizeImage13(d.disks[disk].NibImage.GetNibbles(), d.disks[disk].NibImage.CurrentSectorOrder)
					}
					if err == nil {
						f, err := os.Create(fn)
						if err == nil {
							log.Printf("DiskII: saving nib -> DSK: %s", volume)
							f.Write(buffer)
							f.Close()
						}
					} else {
						log.Printf("DiskII: saving nib -> DSK failed: %v", err)
					}
				} else {
					f, err := os.Create(fn)
					if err == nil {
						f.Write(d.disks[disk].NibImage.GetNibbles())
						f.Close()
					}
				}

			} else {
				if settings.PreserveDSK && denibbleOnSave {
					var buffer []byte
					var err error
					switch d.disks[disk].NibImage.Format.SPT() {
					case 16:
						buffer, err = woz.DeNibblizeImage(d.disks[disk].NibImage.GetNibbles(), d.disks[disk].NibImage.CurrentSectorOrder)
					case 13:
						buffer, err = woz.DeNibblizeImage13(d.disks[disk].NibImage.GetNibbles(), d.disks[disk].NibImage.CurrentSectorOrder)
					}
					if err == nil {
						log.Printf("DiskII: saving nib -> DSK: %s", volume)
						err := files.WriteBytesViaProvider(files.GetPath(volume), files.GetFilename(volume), buffer)
						log.Printf("error = %v", err)
					} else {
						log.Printf("DiskII: saving nib -> DSK failed: %v", err)
					}
				} else {
					files.WriteBytesViaProvider(files.GetPath(volume), files.GetFilename(volume), d.disks[disk].NibImage.GetNibbles())
				}
			}

			d.disks[disk].SetDiskUpdatePending(false)

		} else {

			if d.disks[disk].Disk == nil || !d.disks[disk].Disk.IsModified() {
				continue
			}
			var volume string
			switch disk {
			case 0:
				volume = settings.PureBootVolume[d.e.GetMemIndex()]
			case 1:
				volume = settings.PureBootVolume2[d.e.GetMemIndex()]
			}

			//log.Printf("Current volume: %s", volume)

			if !strings.HasSuffix(volume, ".woz") || (strings.HasPrefix(volume, "/appleii") || strings.HasPrefix(volume, "appleii")) {
				p := strings.Split(volume, ".")
				p[len(p)-1] = "woz"
				volume = strings.Join(p, ".")

				if strings.HasPrefix(volume, "/appleii") || strings.HasPrefix(volume, "appleii") {
					volume = "/local/mydisks/" + files.GetFilename(volume)
				}

				//log.Printf("Updating volume to %s...", volume)
				switch disk {
				case 0:
					settings.PureBootVolume[d.e.GetMemIndex()] = volume
				case 1:
					settings.PureBootVolume2[d.e.GetMemIndex()] = volume
				}
			}

			d.disks[disk].Disk.UpdateCRC32()
			if strings.HasPrefix(volume, "local:") {
				fn := strings.Replace(volume, "local:", "", -1)
				f, err := os.Create(fn)
				if err == nil {
					f.Write(d.disks[disk].Disk.GetData().ByteSlice(0, d.disks[disk].Disk.GetSize()))
					f.Close()
				}
			} else {
				files.WriteBytesViaProvider(files.GetPath(volume), files.GetFilename(volume), d.disks[disk].Disk.GetData().ByteSlice(0, d.disks[disk].Disk.GetSize()))
			}

			d.disks[disk].Disk.SetModified(false)

		}
	}

}

func (d *IOCardDiskII) AdjustClock(n int) {

}

func (d *IOCardDiskII) Increment(n int) {

	current := d.GetCurrent()

	var isBitstream = d.disks[current].Disk != nil

	d.currentCycles += int64(n)
	d.currentLSSCycles += float64(n) * d.lssCycleRatio

	if d.driveBlinkCycles != 0 && d.driveBlinkCycles < d.currentCycles {
		if current == 0 {
			d.mm.IntSetLED0(d.index, 0)
		} else {
			d.mm.IntSetLED1(d.index, 0)
		}
	}

	// out of warp 0.25s after last i/o call
	if d.driveSlowdownCycles != 0 && d.driveSlowdownCycles < d.currentCycles {
		d.disks[current].ActivationCallback(false)
		d.driveSlowdownCycles = 0
	}

	if d.driveOffCycles != 0 && d.driveOffCycles < d.currentCycles {
		//log.Printf("Drive %d spindown complete", current)
		d.driveOffCycles = 0
		d.disks[current].SetOn(false)
		settings.SetUserCanChangeSpeed(true)
		d.CheckUpdate()
	}

	if isBitstream {
		// get clocks from current volume
		d.SetLSSRatio(2 * 32 / float64(d.disks[current].Disk.GetOptimalBitTiming()))

		// disk processing
		d.updateLogicStateSequencer()

	}

	d.CheckTimer()

	if !d.GetDriveEnabled() {
		// handle injected service bus requests inline
		d.HandleServiceBusInjection(d.HandleServiceBusRequest)
	}

}

func (d *IOCardDiskII) Decrement(n int) {
	d.currentCycles -= int64(n)
}

func (d *IOCardDiskII) ImA() string {
	return "DISKII"
}

func (d *IOCardDiskII) Reset() {
	d.disks[0].SetOn(false)
	d.disks[1].SetOn(false)
}

func (d *IOCardDiskII) Done(slot int) {
	// done
	servicebus.Unsubscribe(d.e.GetMemIndex(), d)
}

func (d *IOCardDiskII) HandleIO(register int, value *uint64, eventType IOType) {

	d.readMode = (eventType == IOT_READ)
	current := d.GetCurrent()

	switch d.disks[current].UseNib() {
	case true:
		d.HandleIONib(register, value, eventType)
	case false:
		d.HandleIOBitStream(register, value, eventType)
	}
}

func (d *IOCardDiskII) HandleIONib(register int, value *uint64, eventType IOType) {

	d.readMode = (eventType == IOT_READ)
	current := d.GetCurrent()
	//log.Printf("Drive %d using nibble driver...", current)
	// d.driveSlowdownCycles = d.currentCycles + 250000 // 0.25s
	// d.disks[current].ActivationCallback(true)

	switch register {
	case 0x0:
		//d.Log("Step")
		d.disks[current].Step(register)
		break
	case 0x1:
		//d.Log("Step")
		d.disks[current].Step(register)
		break
	case 0x2:
		//d.Log("Step")
		d.disks[current].Step(register)
		break
	case 0x3:
		//d.Log("Step")
		d.disks[current].Step(register)
		break
	case 0x4:
		//d.Log("Step")
		d.disks[current].Step(register)
		break
	case 0x5:
		//d.Log("Step")
		d.disks[current].Step(register)
		break
	case 0x6:
		//d.Log("Step")
		d.disks[current].Step(register)
		break
	case 0x7:
		//d.Log("Step")
		d.disks[current].Step(register)
		break

	case 0x8:
		// drive off
		if current == 0 {
			d.mm.IntSetLED0(d.index, 0)
		} else {
			d.mm.IntSetLED1(d.index, 0)
		}

		//d.Log("Drive Off")

		d.driveOffCycles = d.currentCycles + 2*1020484
		d.driveSlowdownCycles = d.currentCycles + 250000

		break

	case 0x9:
		// drive on

		if current == 0 {
			d.mm.IntSetLED0(d.index, 1)
		} else {
			d.mm.IntSetLED1(d.index, 1)
		}

		//d.Log("Drive on")
		d.driveOffCycles = 0
		d.driveSlowdownCycles = 0
		d.disks[current].SetOn(true)
		settings.SetUserCanChangeSpeed(false)
		break

	case 0xA:
		// drive 1
		//d.Log("Select Disk 0")
		current = 0

		if current == 0 {
			d.mm.IntSetLED1(d.index, 0)
		} else {
			d.mm.IntSetLED0(d.index, 0)
		}

		d.SetCurrent(current)

		break

	case 0xB:
		// drive 2
		//d.Log("Select Disk 1")
		current = 1

		if current == 0 {
			d.mm.IntSetLED1(d.index, 0)
		} else {
			d.mm.IntSetLED0(d.index, 0)
		}

		d.SetCurrent(current)

		break

	case 0xC:
		// read/write latch
		//d.Log("Write")
		d.disks[current].Write()
		*value = uint64(d.disks[current].ReadLatch())
		if d.disks[current].NibImage == nil {
			*value = 0xff
		}
		d.driveBlinkCycles = d.currentCycles + 100000
		if current == 0 {
			d.mm.IntSetLED0(d.index, 1)
		} else {
			d.mm.IntSetLED1(d.index, 1)
		}
		break
	case 0xF:
		// write mode
		//d.Log("Set Write Mode")
		d.disks[current].SetWrite()
	case 0xD:
		//d.Log("Set Latch %d", *value)
		// set latch
		if eventType == IOT_WRITE {
			d.disks[current].SetLatchValue(byte(*value))
		}
		*value = uint64(d.disks[current].ReadLatch())
		d.driveBlinkCycles = d.currentCycles + 100000
		if current == 0 {
			d.mm.IntSetLED0(d.index, 1)
		} else {
			d.mm.IntSetLED1(d.index, 1)
		}
		break

	case 0xE:
		// read mode
		//d.Log("Set Read Mode")
		d.disks[current].SetReadMode()
		if d.disks[current].NibImage != nil && d.disks[current].NibImage.WriteProtected {
			*value = 0x080
		} else {
			*value = 0
		}
		break
	}

	// populate
	if d.GetWriteMode() {
		if register&0x01 == 1 {
			d.disks[current].SetLatchValue(byte(*value))
		}
	}

}

func (d *IOCardDiskII) HandleIOBitStream(register int, value *uint64, eventType IOType) {

	d.readMode = (eventType == IOT_READ)
	current := d.GetCurrent()

	// d.driveSlowdownCycles = d.currentCycles + 250000 // 0.25s
	// d.disks[current].ActivationCallback(true)

	switch register {
	case 0, 1, 2, 3, 4, 5, 6, 7:
		//d.Log("Step")
		//d.disks[current].StepQ(register)
		// set timer to step disk
		if d.GetDrive(current).UpdatePhaseControl(register) {
			d.SetTimer(d.currentCycles+DRIVE_PHASE_CYCLES, d.UpdatePhase, register)
		}

		break

	case 0x8:
		if current == 0 {
			d.mm.IntSetLED0(d.index, 0)
		} else {
			d.mm.IntSetLED1(d.index, 0)
		}

		//d.Log("Drive Off")

		d.driveOffCycles = d.currentCycles + 2*1020484
		d.driveSlowdownCycles = d.currentCycles + 250000 // 0.25s
		//d.disks[d.current].SetPendingDeactivation()

		break

	case 0x9:
		if current == 0 {
			d.mm.IntSetLED0(d.index, 1)
		} else {
			d.mm.IntSetLED1(d.index, 1)
		}

		//d.Log("Drive on")
		d.driveOffCycles = 0
		d.driveSlowdownCycles = 0
		d.disks[current].SetOn(true)
		settings.SetUserCanChangeSpeed(false)
		break

	case 0xA:
		//d.Log("Select Disk 0")
		current = 0

		if current == 0 {
			d.mm.IntSetLED1(d.index, 0)
		} else {
			d.mm.IntSetLED0(d.index, 0)
		}

		d.SetCurrent(current)

		break

	case 0xB:
		//d.Log("Select Disk 1")
		current = 1

		if current == 0 {
			d.mm.IntSetLED1(d.index, 0)
		} else {
			d.mm.IntSetLED0(d.index, 0)
		}

		d.SetCurrent(current)

		break

	case 0xC:
		//d.Log("Read")
		d.SetQ6(0)

		state := d.GetState()
		state = (state & ^0x04) & 0xff
		d.SetState(state)

		d.updateLogicStateSequencer()
		d.driveBlinkCycles = d.currentCycles + 100000
		if current == 0 {
			d.mm.IntSetLED0(d.index, 1)
		} else {
			d.mm.IntSetLED1(d.index, 1)
		}
		break

	case 0xD:
		d.SetQ6(1)

		state := d.GetState()
		state = (state | 0x04) & 0xff
		d.SetState(state)

		d.driveBlinkCycles = d.currentCycles + 100000
		if current == 0 {
			d.mm.IntSetLED0(d.index, 1)
		} else {
			d.mm.IntSetLED1(d.index, 1)
		}
		break

	case 0xE:
		state := d.GetState()
		state = (state & ^0x08) & 0xff
		d.SetState(state)
		d.SetWriteMode(false)
		break

	case 0xF:
		state := d.GetState()
		state = (state | 0x08) & 0xff
		d.SetState(state)
		d.SetWriteMode(true)
		break
	}

	// populate
	if d.readMode {
		if (register & 0x01) == 0 {
			if d.disks[current].Disk == nil {
				*value = 0xff
			} else {
				*value = uint64(d.GetDataRegister())
			}
		} else {
			*value = 0
		}
	} else if d.GetWriteMode() {
		if register&0x01 == 1 {
			d.SetPendingWrite(byte(*value))
		}
	}
}

// Perform a single LSS cycle
func (mr *IOCardDiskII) updateLogicStateSequencer() {

	enabled := mr.GetDriveEnabled()
	if !enabled {
		return
	}

	current := mr.GetCurrent()
	drive := mr.disks[current]
	workCycles := int(math.Floor(mr.currentLSSCycles - mr.lastLSSCycles))
	mr.lastLSSCycles += float64(workCycles) // catch up whole number of cycles
	dataRegister := mr.GetDataRegister()
	odataRegister := dataRegister
	clock := mr.GetClock()
	oclock := clock
	state := mr.GetState()
	ostate := state
	clocksPerBit := 8
	writeProtected := drive.IsWriteProtected()
	pendingWrite := mr.GetPendingWrite()
	disk := drive.Disk

	for workCycles > 0 {
		workCycles--

		if clock == 1 {
			mr.lssPulse = 0
		}

		// On last clock do read
		if clock >= clocksPerBit-1 {
			if enabled && disk != nil {
				// read bit from stream & store for LSS
				mr.lssPulse = drive.ReadBit()
				// write bit if in write mode
				if (state&0x08 != 0) && !writeProtected {
					drive.WriteBit(dataRegister >> 7)
				}
			}
		}

		//log2.Printf("clock=%d, pulse=%d", clock, mr.lssPulse)

		clock++
		if clock > clocksPerBit-1 {
			clock = 0
		}

		if mr.lssPulse != 0 {
			state = (state & ^0x10) & 0xff
		} else {
			state = (state | 0x10) & 0xff
		}

		if dataRegister&0x80 != 0 {
			state = (state | 0x02) & 0xff
		} else {
			state = (state & ^0x02) & 0xff
		}

		var command = int(romP6[state])

		state = (state & 0x1e) | (command & 0xc0) | ((command & 0x20) >> 5) | ((command & 0x10) << 1)

		// Execute LSS "opcode"
		if command&0xf < 0x8 {
			dataRegister = 0
		} else {
			switch command & 0xf {
			case 0x8, 0xC:
				// NOP
			case 0x9:
				// SL0 *
				dataRegister = (dataRegister << 1) & 0xff
			case 0xA, 0xE:
				// SR *
				dataRegister >>= 1
				if writeProtected {
					dataRegister |= 0x80
				}
			case 0xB, 0xF:
				// LD *
				dataRegister = pendingWrite
			case 0xD:
				// SL1 *
				dataRegister = ((dataRegister << 1) | 0x01) & 0xff
			}
		}

	}

	// Store changed registers to memory
	if odataRegister != dataRegister {
		mr.SetDataRegister(dataRegister)
	}
	if ostate != state {
		mr.SetState(state)
	}
	if oclock != clock {
		mr.SetClock(clock)
	}
}

func (mr *IOCardDiskII) getCycles() int64 {
	cpu := apple2helpers.GetCPU(mr.e)
	return cpu.GlobalCycles
}

func (mr *IOCardDiskII) InsertHelper(drive int, filename string, wp bool) bool {

	fmt.Printf("InsertHelper(%d, \"%s\", %v)\n", drive, filename, wp)

	switch drive {
	case 0:
		mr.disks[drive].InsertDisk(filename, wp)
		mr.Log("Insert volume [%s] into disk %d", drive, filename)
		return true
	case 1:
		mr.disks[drive].InsertDisk(filename, wp)
		mr.Log("Insert volume [%s] into disk %d", drive, filename)
		return true
	}
	return false
}

func (mr *IOCardDiskII) InsertHelperBin(drive int, data []byte, filename string, wp bool) bool {
	switch drive {
	case 0:
		mr.disks[drive].InsertDiskBin(data, filename, wp)
		mr.Log("Insert volume into disk %d", drive)
		return true
	case 1:
		mr.disks[drive].InsertDiskBin(data, filename, wp)
		mr.Log("Insert volume into disk %d", drive)
		return true
	}
	return false
}

func (mr *IOCardDiskII) GetVolumes() []string {
	return []string{mr.disks[0].VolumeName, mr.disks[1].VolumeName}
}

func (mr *IOCardDiskII) HasDisk(drive int) bool {
	switch drive {
	case 0:
		return mr.disks[0].HasDisk()
	case 1:
		return mr.disks[1].HasDisk()
	}
	return false
}

func (mr *IOCardDiskII) EjectHelper(drive int) {
	switch drive {
	case 0:
		mr.Log("Eject disk %d", drive)
		mr.disks[drive].Eject()
	case 1:
		mr.Log("Eject disk %d", drive)
		mr.disks[drive].Eject()
	}
}

func (mr *IOCardDiskII) SwapDisks() {

	// d1 := mr.disks[1].Disk
	// d0 := mr.disks[0].Disk

	// v1 := mr.disks[1].VolumeName
	// v0 := mr.disks[0].VolumeName

	mr.Init(mr.Slot)

	// mr.disks[1].InsertDiskWrapper(d0)
	// mr.disks[1].VolumeName = v0
	// mr.disks[0].InsertDiskWrapper(d1)
	// mr.disks[0].VolumeName = v1
	fmt.Println("SwapDisks needs to be implemented")

	mr.Log("Disks in A/B are now swapped")

}

func (mr *IOCardDiskII) SetWriteProtect(drive int, b bool) {
	switch drive % 2 {
	case 0:
		mr.disks[0].SetWriteProtect(b)
	case 1:
		mr.disks[1].SetWriteProtect(b)
	}
}

func (mr *IOCardDiskII) SetDisk(drive int, img WOZImg) {
	if drive == 0 {
		mr.disks[0].Disk = img
		mr.disks[0].NibImage = nil
		mr.disks[0].SetUseNib(false)
	} else if drive == 1 {
		mr.disks[1].Disk = img
		mr.disks[1].NibImage = nil
		mr.disks[1].SetUseNib(false)
	}
	return
}

func (mr *IOCardDiskII) SetDiskNib(drive int, img *disk.DSKWrapper) {
	if drive == 0 {
		mr.disks[0].NibImage = img
		mr.disks[0].Disk = nil
		mr.disks[0].SetUseNib(true)
	} else if drive == 1 {
		mr.disks[1].NibImage = img
		mr.disks[1].Disk = nil
		mr.disks[1].SetUseNib(true)
	}
	return
}

func (mr *IOCardDiskII) GetDisk(drive int) WOZImg {
	if drive == 0 {
		if mr.disks[0].UseNib() {
			return nil
		}
		return mr.disks[0].Disk
	} else if drive == 1 {
		if mr.disks[1].UseNib() {
			return nil
		}
		return mr.disks[1].Disk
	}
	return nil
}

func (mr *IOCardDiskII) GetDiskNib(drive int) *disk.DSKWrapper {
	if drive == 0 {
		if !mr.disks[0].UseNib() {
			return nil
		}
		return mr.disks[0].NibImage
	} else if drive == 1 {
		if !mr.disks[1].UseNib() {
			return nil
		}
		return mr.disks[1].NibImage
	}
	return nil
}

func DiskInsert(caller interfaces.Interpretable, drive int, volume string, wp bool) bool {
	mp, ex := caller.GetMemoryMap().InterpreterMappableByLabel(caller.GetMemIndex(), "Apple2IOChip")
	if ex {

		a2io := mp.(*Apple2IOChip)

		for i := 0; i < 8; i++ {

			card := a2io.GetCard(i)
			if card != nil && card.CardName() == "IOCardDiskII" {

				card.(*IOCardDiskII).EjectHelper(drive)
				if volume != "" {
					card.(*IOCardDiskII).InsertHelper(drive, volume, wp)
					card.(*IOCardDiskII).Init(caller.GetMemIndex())
				} else {
					card.(*IOCardDiskII).InsertHelper(drive, "", wp)
				}
				return card.(*IOCardDiskII).HasDisk(drive)

			}

		}
	}
	return false
}

func GetDiskNib(caller interfaces.Interpretable, drive int) *disk.DSKWrapper {
	mp, ex := caller.GetMemoryMap().InterpreterMappableByLabel(caller.GetMemIndex(), "Apple2IOChip")
	if ex {

		a2io := mp.(*Apple2IOChip)

		for i := 0; i < 8; i++ {

			card := a2io.GetCard(i)
			if card != nil && card.CardName() == "IOCardDiskII" {

				dsk := card.(*IOCardDiskII).GetDiskNib(drive)
				return dsk

			}

		}
	}
	return nil
}

func GetDisk(caller interfaces.Interpretable, drive int) WOZImg {
	mp, ex := caller.GetMemoryMap().InterpreterMappableByLabel(caller.GetMemIndex(), "Apple2IOChip")
	if ex {

		a2io := mp.(*Apple2IOChip)

		for i := 0; i < 8; i++ {

			card := a2io.GetCard(i)
			if card != nil && card.CardName() == "IOCardDiskII" {

				dsk := card.(*IOCardDiskII).GetDisk(drive)
				return dsk

			}

		}
	}
	return nil
}

func SetDiskNib(caller interfaces.Interpretable, drive int, img *disk.DSKWrapper) {
	mp, ex := caller.GetMemoryMap().InterpreterMappableByLabel(caller.GetMemIndex(), "Apple2IOChip")
	if ex {

		a2io := mp.(*Apple2IOChip)

		for i := 0; i < 8; i++ {

			card := a2io.GetCard(i)
			if card != nil && card.CardName() == "IOCardDiskII" {

				card.(*IOCardDiskII).SetDiskNib(drive, img)
				return

			}

		}
	}
	return
}

func SetDisk(caller interfaces.Interpretable, drive int, img *woz.WOZImage) {
	mp, ex := caller.GetMemoryMap().InterpreterMappableByLabel(caller.GetMemIndex(), "Apple2IOChip")
	if ex {

		a2io := mp.(*Apple2IOChip)

		for i := 0; i < 8; i++ {

			card := a2io.GetCard(i)
			if card != nil && card.CardName() == "IOCardDiskII" {

				card.(*IOCardDiskII).SetDisk(drive, img)
				return

			}

		}
	}
	return
}

func SetWriteProtect(caller interfaces.Interpretable, drive int, b bool) {
	mp, ex := caller.GetMemoryMap().InterpreterMappableByLabel(caller.GetMemIndex(), "Apple2IOChip")
	if ex {

		a2io := mp.(*Apple2IOChip)

		for i := 0; i < 8; i++ {

			card := a2io.GetCard(i)
			if card != nil && card.CardName() == "IOCardDiskII" {

				card.(*IOCardDiskII).SetWriteProtect(drive, b)

			}

		}
	}
}

func DiskSwap(caller interfaces.Interpretable) {
	mp, ex := caller.GetMemoryMap().InterpreterMappableByLabel(caller.GetMemIndex(), "Apple2IOChip")
	if ex {

		a2io := mp.(*Apple2IOChip)

		for i := 0; i < 8; i++ {

			card := a2io.GetCard(i)
			if card != nil && card.CardName() == "IOCardDiskII" {

				card.(*IOCardDiskII).SwapDisks()
				card.(*IOCardDiskII).Init(caller.GetMemIndex())

			}

		}
	}
}

func DiskInsertBin(caller interfaces.Interpretable, drive int, data []byte, filename string, wp bool) bool {
	mp, ex := caller.GetMemoryMap().InterpreterMappableByLabel(caller.GetMemIndex(), "Apple2IOChip")
	if ex {

		a2io := mp.(*Apple2IOChip)

		for i := 0; i < 8; i++ {

			card := a2io.GetCard(i)
			if card != nil && card.CardName() == "IOCardDiskII" {

				card.(*IOCardDiskII).EjectHelper(drive)
				card.(*IOCardDiskII).InsertHelperBin(drive, data, filename, wp)
				card.(*IOCardDiskII).Init(caller.GetMemIndex())
				return card.(*IOCardDiskII).HasDisk(drive)

			}

		}
	}
	return false
}

func GetDiskVolumes(caller interfaces.Interpretable) []string {
	mp, ex := caller.GetMemoryMap().InterpreterMappableByLabel(caller.GetMemIndex(), "Apple2IOChip")
	if ex {

		a2io := mp.(*Apple2IOChip)

		for i := 0; i < 8; i++ {

			card := a2io.GetCard(i)
			if card != nil && card.CardName() == "IOCardDiskII" {

				return card.(*IOCardDiskII).GetVolumes()

			}

		}
	}
	return []string{"Empty", "Empty"}
}
