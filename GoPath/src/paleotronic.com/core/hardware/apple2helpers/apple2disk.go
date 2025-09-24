package apple2helpers

/*
Param List structure: --

	Offset 	Size	Desc
	0	1	IOB Type
	1	1	SLOT * 16 (60)
	2	1	DRIVE (0/1)
	3	1	VOLUME (0 = any)
	4	1	TRACK (0-34)
	5	1	SECTOR (0-15)
	6	2	ADDRESS OF DCT
	8	2	ADDRESS OF BUFFER
	10	2	SECTOR SIZE (0100)

	12	1	COMMAND
			0x00 NULL
			0x01 READ
			0x02 WRITE
			0x04 FORMAT

	13	1	RETURN CODE
			0x10 WRITE PROTECTED
			0x20 VOLUME MISMATCH
			0x40 DRIVE ERROR
			0x80 READ ERROR

	14	1	TRUE VOLUME (from RWTS)
	15	1	PREV SLOT (from last RWTS call)
	16	1	PREV DRIVE (from last RWTS call)
*/

import "paleotronic.com/core/hardware/cpu/mos6502"
import "paleotronic.com/files"
import "paleotronic.com/disk"
import "paleotronic.com/core/memory"

import "paleotronic.com/log"

type RWTSCommand byte

const (
	RWTS_None   RWTSCommand = 0x00
	RWTS_Read   RWTSCommand = 0x01
	RWTS_Write  RWTSCommand = 0x02
	RWTS_Format RWTSCommand = 0x04
)

func (r RWTSCommand) String() string {
	switch r {
	case RWTS_None:
		return "None"
	case RWTS_Read:
		return "Read"
	case RWTS_Write:
		return "Write"
	case RWTS_Format:
		return "Format"
	}
	return "Unknown"
}

type RWTSReturnCode byte

const (
	RWTS_Ok             RWTSReturnCode = 0x00
	RWTS_InitError      RWTSReturnCode = 0x08
	RWTS_WriteProtect   RWTSReturnCode = 0x10
	RWTS_VolumeMismatch RWTSReturnCode = 0x20
	RWTS_DriveError     RWTSReturnCode = 0x40
	RWTS_ReadError      RWTSReturnCode = 0x80
)

func (r RWTSReturnCode) String() string {
	switch r {
	case RWTS_Ok:
		return "Ok"
	case RWTS_InitError:
		return "Init Error"
	case RWTS_WriteProtect:
		return "Write Protect Error"
	case RWTS_VolumeMismatch:
		return "Drive Error"
	case RWTS_ReadError:
		return "Read Error"
	}
	return "Unknown"
}

const RWTS_PARM_ADDR = 0x9000

var RWTSLocateParams = func(cpu *mos6502.Core6502) int64 {
	cpu.Y = RWTS_PARM_ADDR % 256
	cpu.A = RWTS_PARM_ADDR / 256

	log.Println("Passing custom buffer")

	return 0
}

var RWTSInvoker = func(cpu *mos6502.Core6502) int64 {

	// STore IOB
	cpu.Int.SetMemory(0x48, uint64(cpu.Y))
	cpu.Int.SetMemory(0x49, uint64(cpu.A))

	paddr := cpu.Y + 256*cpu.A

	rwts := &RWTSParams{
		index: cpu.Int.GetMemIndex(),
		addr:  paddr,
		mm:    cpu.Int.GetMemoryMap(),
	}

	rwts.Dump()

	var dsk *disk.DSKWrapper

	cpu.SetFlag(mos6502.F_C, false)

	switch rwts.GetDriveNo() {
	case 0x01:
		dsk = files.GetDisk(0)
	case 0x02:
		dsk = files.GetDisk(1)
	}

	if dsk == nil {
		// device error
		rwts.SetReturnCode(RWTS_DriveError)
		cpu.SetFlag(mos6502.F_C, true)
		return 100
	}

	switch rwts.GetCommandCode() {
	case RWTS_Format:
		rwts.SetReturnCode(RWTS_InitError)
		cpu.SetFlag(mos6502.F_C, true)
		return 100
	case RWTS_Write:
		rwts.SetReturnCode(RWTS_Ok)
		cpu.SetFlag(mos6502.F_C, false)
		return 100
	case RWTS_Read:

		e := dsk.Seek(rwts.GetTrackNumber(), rwts.GetSectorNumber())

		if e != nil {
			rwts.SetReturnCode(RWTS_ReadError)
			cpu.SetFlag(mos6502.F_C, true)
			return 1000
		}

		data := dsk.Read()

		base := rwts.GetBufferAddress()
		for i, v := range data {
			cpu.Int.SetMemory(base+i, uint64(v))
		}

		rwts.SetReturnCode(RWTS_Ok)
		rwts.SetPreviousDrive(rwts.GetDriveNo())
		rwts.SetPreviousSlot(rwts.GetSlot())
		rwts.SetTrueVolume(254)
		return 100
	}

	return 0
}

type RWTSParams struct {
	index int
	addr  int
	mm    *memory.MemoryMap
}

func (r *RWTSParams) GetIOBType() byte {
	return byte(r.mm.ReadInterpreterMemory(r.index, r.addr+0))
}

func (r *RWTSParams) GetSlot() byte {
	return byte(r.mm.ReadInterpreterMemory(r.index, r.addr+1))
}

func (r *RWTSParams) GetDriveNo() byte {
	return byte(r.mm.ReadInterpreterMemory(r.index, r.addr+2))
}

func (r *RWTSParams) GetVolumeID() byte {
	return byte(r.mm.ReadInterpreterMemory(r.index, r.addr+3))
}

func (r *RWTSParams) GetTrackNumber() int {
	return int(r.mm.ReadInterpreterMemory(r.index, r.addr+4))
}

func (r *RWTSParams) GetSectorNumber() int {
	return int(r.mm.ReadInterpreterMemory(r.index, r.addr+5))
}

func (r *RWTSParams) GetDCTAddress() int {
	return int(r.mm.ReadInterpreterMemory(r.index, r.addr+6)) + 256*int(r.mm.ReadInterpreterMemory(r.index, r.addr+7))
}

func (r *RWTSParams) GetBufferAddress() int {
	return int(r.mm.ReadInterpreterMemory(r.index, r.addr+8)) + 256*int(r.mm.ReadInterpreterMemory(r.index, r.addr+9))
}

func (r *RWTSParams) GetSectorSize() int {
	return int(r.mm.ReadInterpreterMemory(r.index, r.addr+10)) + 256*int(r.mm.ReadInterpreterMemory(r.index, r.addr+11))
}

func (r *RWTSParams) GetCommandCode() RWTSCommand {
	return RWTSCommand(r.mm.ReadInterpreterMemory(r.index, r.addr+12))
}

func (r *RWTSParams) GetReturnCode() RWTSReturnCode {
	return RWTSReturnCode(r.mm.ReadInterpreterMemory(r.index, r.addr+13))
}

func (r *RWTSParams) SetReturnCode(v RWTSReturnCode) {
	r.mm.WriteInterpreterMemory(r.index, r.addr+13, uint64(v))
}

func (r *RWTSParams) GetTrueVolume() byte {
	return byte(r.mm.ReadInterpreterMemory(r.index, r.addr+14))
}

func (r *RWTSParams) SetTrueVolume(b byte) {
	r.mm.WriteInterpreterMemory(r.index, r.addr+14, uint64(b))
}

func (r *RWTSParams) GetPreviousSlot() byte {
	return byte(r.mm.ReadInterpreterMemory(r.index, r.addr+15))
}

func (r *RWTSParams) SetPreviousSlot(b byte) {
	r.mm.WriteInterpreterMemory(r.index, r.addr+15, uint64(b))
}

func (r *RWTSParams) GetPreviousDrive() byte {
	return byte(r.mm.ReadInterpreterMemory(r.index, r.addr+16))
}

func (r *RWTSParams) SetPreviousDrive(b byte) {
	r.mm.WriteInterpreterMemory(r.index, r.addr+16, uint64(b))
}

func (r *RWTSParams) Dump() {
	log.Println("RWTS REQUEST")
	log.Println("============")
	log.Printf(" IOB Type: %d\n", r.GetIOBType())
	log.Printf(" SLOT #  : %d\n", r.GetSlot()/16)
	log.Printf(" Drive # : %d\n", r.GetDriveNo())
	log.Printf(" Volume  : %d\n", r.GetVolumeID())
	log.Printf(" Track   : %d\n", r.GetTrackNumber())
	log.Printf(" Sector  : %d\n", r.GetSectorNumber())
	log.Printf(" DCT Addr: %d\n", r.GetDCTAddress())
	log.Printf(" Buf addr: %d\n", r.GetBufferAddress())
	log.Printf(" SSize   : %d\n", r.GetSectorSize())
	log.Printf(" Command : %d\n", r.GetCommandCode())
	log.Println("------------")
}
