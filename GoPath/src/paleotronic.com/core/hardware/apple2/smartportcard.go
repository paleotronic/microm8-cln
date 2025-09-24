package apple2

import (
	//~ "errors"
	"strings"

	"paleotronic.com/log" //"log"

	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/hardware/cpu/mos6502"
	"paleotronic.com/core/hardware/servicebus"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/settings"
	"paleotronic.com/files"
	"paleotronic.com/fmt"
)

const LEDOffCycles = 125000
const WriteOffCycles = 1020484

type SmartPortErrorType int

const (
	spetNoError        SmartPortErrorType = 0x00
	spetInvalidCommand SmartPortErrorType = 0x01
	spetBadParamCount  SmartPortErrorType = 0x04
	spetInvalidUnit    SmartPortErrorType = 0x11
	spetInvalidCode    SmartPortErrorType = 0x21
	spetNoWrite        SmartPortErrorType = 0x2b
	spetBadBlockNumber SmartPortErrorType = 0x2d
	spetOffline        SmartPortErrorType = 0x2f
)

type MLICommandType int

const (
	mliStatus MLICommandType = iota
	mliRead
	mliWrite
	mliFormat
)

type MLIReturnType int

const (
	mliNoError        MLIReturnType = 0x00
	mliIOError        MLIReturnType = 0x27
	mliNoDevice       MLIReturnType = 0x28
	mliWriteProtected MLIReturnType = 0x2B
)

const ddOffset = 0x0a
const maxSmartPortDrives = 0x01
const zpRefSlot16 = 0x2B
const zpRefMLIDDBase = 0x48
const zpRefMLICommand = 0x042
const zpRefMLIUnit = 0x043
const zpRefMLIBufferAddress = 0x044
const zpRefMLIBlockNumber = 0x046
const zpRefMLIReadRoutine = 0x048

var smartportSig = []uint64{
	0xa9, 0x20, 0xa9, 0x00,
	0xa9, 0x03, 0xa9, 0x3c,
	0xd0, 0x07, 0x60, 0xb0,
	0x01, 0x18, 0xb0, 0x5a,
}

type IOCardSmartPort struct {
	IOCard
	Int               interfaces.Interpretable
	Drives            [maxSmartPortDrives]*SmartPortBlockDevice
	CurrentDrive      int
	LEDTurnOffCycles  int64
	WriteUpdateCycles int64
	GlobalCycles      int64
}

func (d *IOCardSmartPort) Init(slot int) {
	d.IOCard.Init(slot)
	d.LEDTurnOffCycles = -1
	d.WriteUpdateCycles = -1
	d.Log("Initialising smartport...")
	servicebus.Subscribe(
		d.Int.GetMemIndex(),
		servicebus.SmartPortEject,
		d,
	)
	servicebus.Subscribe(
		d.Int.GetMemIndex(),
		servicebus.SmartPortInsertFilename,
		d,
	)
	servicebus.Subscribe(
		d.Int.GetMemIndex(),
		servicebus.SmartPortInsertBytes,
		d,
	)
	d.Int.SetCycleCounter(d)
	log.Println("Smart port init")
}

func (d *IOCardSmartPort) DoLED() {
	d.LEDTurnOffCycles = d.GlobalCycles + LEDOffCycles
	switch d.CurrentDrive % 2 {
	case 0:
		d.Int.GetMemoryMap().IntSetLED0(d.Int.GetMemIndex(), 2)
	case 1:
		d.Int.GetMemoryMap().IntSetLED1(d.Int.GetMemIndex(), 2)
	}
}

func (d *IOCardSmartPort) ImA() string {
	return "SmartPort"
}

func (d *IOCardSmartPort) Increment(n int) {
	d.HandleServiceBusInjection(d.HandleServiceBusRequest)
	d.GlobalCycles += int64(n)
	if d.LEDTurnOffCycles != -1 && d.GlobalCycles > d.LEDTurnOffCycles {
		d.LEDTurnOffCycles = -1
		d.Int.GetMemoryMap().IntSetLED0(d.Int.GetMemIndex(), 0)
		d.Int.GetMemoryMap().IntSetLED1(d.Int.GetMemIndex(), 0)
	}
	if d.WriteUpdateCycles != -1 && d.GlobalCycles > d.WriteUpdateCycles {
		d.WriteUpdateCycles = -1
		for _, dev := range d.Drives {
			if dev != nil {
				dev.CheckUpdate()
			}
		}
	}
}

func (d *IOCardSmartPort) Decrement(n int) {

}

func (d *IOCardSmartPort) AdjustClock(n int) {

}

func (d *IOCardSmartPort) Done(slot int) {
	servicebus.Unsubscribe(
		d.Int.GetMemIndex(),
		d,
	)
	d.Int.ClearCycleCounter(d)

	for _, dev := range d.Drives {
		if dev != nil {
			log.Printf("*** checking smartport volume for update")
			if dev.Updated {
				log.Printf("Need to update disk image %s...", dev.Filename)
				dev.CheckUpdate()
			}
		}
	}

}

func (d *IOCardSmartPort) HandleIO(register int, value *uint64, eventType IOType) {
	switch eventType {
	case IOT_READ:
		switch register {
		default:
			fmt.RPrintf("%s of register %.2x (%.2x:%s)\n", eventType.String(), register, *value, string(rune(*value-128)))
		}
	case IOT_WRITE:
		switch register {
		default:
			fmt.RPrintf("%s of register %.2x (%.2x:%s)\n", eventType.String(), register, *value, string(rune(*value-128)))
		}
	}
}

func NewIOCardSmartPort(mm *memory.MemoryMap, index int, ent interfaces.Interpretable) *IOCardSmartPort {
	this := &IOCardSmartPort{}
	this.SetMemory(mm, index)
	this.Int = ent
	this.Name = "IOCardSmartPort"
	this.IsFWHandler = true

	return this
}

func (d *IOCardSmartPort) FirmwareRead(offset int) uint64 {
	log.Printf("Smartport firmware read @ %.2x", offset)

	if d.Drives[0] == nil {
		return 0x00 // let's pretend we aren't here unless there is a disk mounted
	}

	if offset < 16 {
		return smartportSig[offset]
	}

	switch offset {
	case 0xFC:
		return 0xff
	case 0xFD:
		return 0x7f
	case 0x0FE:
		return 0xd7
	case 0x0FF:
		return ddOffset
	}

	return 0x00
}

func (d *IOCardSmartPort) FirmwareWrite(offset int, value uint64) {
	//
	log.Printf("Smartport firmware write @ 0x%.2x < 0x%.2x", offset, value)
}

func (d *IOCardSmartPort) FirmwareExec(
	offset int,
	PC, A, X, Y, SP, P *int,
) int64 {
	log.Printf("Smartport firmware exec @ %.2x", offset)

	d.DoLED()

	switch offset {
	case 0x00, 0x40:
		if d.Drives[d.CurrentDrive] != nil {
			// can we boot this volume??
			d.BootStrap(d.Drives[d.CurrentDrive])
			return int64(7)
		} else {
			// RTS ...
			cpu := apple2helpers.GetCPU(d.Int)
			rts := cpu.Opref[0x60]
			p := int64(rts.Do(cpu))
			log.Printf("Return to PC = 0x%.4x", cpu.PC)
			return p
		}
	default:
		var p int64
		if offset == ddOffset {
			p = d.handleMLI()
		} else if offset == ddOffset+3 {
			d.handleSmartport()
		} else {
			//panic(errors.New(fmt.Sprintf("Call to unrecognised handler at offset %d", offset)))
		}
		cpu := apple2helpers.GetCPU(d.Int)
		rts := cpu.Opref[0x60]
		rts.Do(cpu)
		log.Printf("*************************** Return to PC = 0x%.4x", cpu.PC)
		return p
	}

	return 0
}

func (d *IOCardSmartPort) Attach(drive int, s *SmartPortBlockDevice) {
	s.e = d.Int
	d.Drives[drive] = s
}

func (d *IOCardSmartPort) BootStrap(s *SmartPortBlockDevice) {
	log.Printf("Attempting to bootstrap smartport")
	loadAddress := 0x800

	cpu := apple2helpers.GetCPU(d.Int)
	slot16 := (d.Slot << 4)
	cpu.X = slot16

	mm := d.Int.GetMemoryMap()
	index := d.Int.GetMemIndex()

	mm.WriteInterpreterMemory(index, zpRefSlot16, uint64(slot16))
	mm.WriteInterpreterMemory(index, zpRefMLICommand, uint64(mliRead))
	mm.WriteInterpreterMemory(index, zpRefMLIUnit, uint64(slot16))

	read := 0xc000 + ddOffset + (d.Slot * 0x0100)

	// Write location to block read routine to zero page
	mm.WriteInterpreterMemory(index, zpRefMLIReadRoutine+0, uint64(read&0xff))
	mm.WriteInterpreterMemory(index, zpRefMLIReadRoutine+1, uint64((read>>8)&0xff))

	b0, err := d.Drives[d.CurrentDrive].GetRawBlock(0)
	if err == nil {
		for i, v := range b0 {
			d.Int.GetMemoryMap().WriteInterpreterMemory(d.Int.GetMemIndex(), loadAddress+i, uint64(v))
		}
		b1, err := d.Drives[d.CurrentDrive].GetRawBlock(1)
		if err == nil {
			for i, v := range b1 {
				d.Int.GetMemoryMap().WriteInterpreterMemory(d.Int.GetMemIndex(), loadAddress+512+i, uint64(v))
			}
		}
	}

	cpu.PC = loadAddress + 1
}

func (d *IOCardSmartPort) handleMLI() int64 {
	d.DoLED()
	cpu := apple2helpers.GetCPU(d.Int)
	r, p := d.prodosMLI()
	cpu.A = int(r)
	if r == mliNoError {
		cpu.SetFlag(mos6502.F_C, false)
	} else {
		cpu.SetFlag(mos6502.F_C, true)
	}
	log.Printf("handleMLI() returns %d", r)
	return p
}

func (d *IOCardSmartPort) changeUnit(unit int) bool {
	if unit < 0 || unit >= maxSmartPortDrives {
		return false
	}
	if d.Drives[unit] == nil {
		return false
	}
	d.CurrentDrive = unit
	return true
}

func (d *IOCardSmartPort) prodosMLI() (MLIReturnType, int64) {
	d.DoLED()
	mm := d.Int.GetMemoryMap()
	index := d.Int.GetMemIndex()

	command := MLICommandType(mm.ReadInterpreterMemory(index, zpRefMLICommand))
	log.Printf("Got MLI command %d", command)
	unit := 0
	if mm.ReadInterpreterMemory(index, zpRefMLIUnit)&0x80 != 0 {
		unit = 1
	}

	if d.changeUnit(unit) == false {
		return mliNoDevice, 0
	}

	log.Printf("Got unit number: %d", unit)

	block := int(mm.ReadInterpreterMemory(index, zpRefMLIBlockNumber)) +
		256*int(mm.ReadInterpreterMemory(index, zpRefMLIBlockNumber+1))

	log.Printf("Got block %d", block)

	bufferAddress := int(mm.ReadInterpreterMemory(index, zpRefMLIBufferAddress)) +
		256*int(mm.ReadInterpreterMemory(index, zpRefMLIBufferAddress+1))

	log.Printf("Got buffer address 0x%.4x", bufferAddress)

	var p int64 = 7

	switch command {
	case mliStatus:
		blocks := d.Drives[d.CurrentDrive].GetBlockCount()
		cpu := apple2helpers.GetCPU(d.Int)
		cpu.X = blocks & 0x0ff
		cpu.Y = (blocks >> 8) & 0x0ff
		locked, _, _ := d.Drives[d.CurrentDrive].GetFlags()
		if locked {
			return mliWriteProtected, p
		} else {
			return mliNoError, p
		}
		break
	case mliFormat:
		d.mliFormat()
	case mliRead:
		d.mliRead(block, bufferAddress)
		p = 3064
		break
	case mliWrite:
		d.mliWrite(block, bufferAddress)
		p = 3064
		break
	default:
		return mliIOError, p
	}

	return mliNoError, p
}

func (d *IOCardSmartPort) mliFormat() {
	// nothing
}

func (d *IOCardSmartPort) mliRead(block int, bufferAddress int) {

	//if bufferAddress > 0xd000 {
	//log.Printf("read of block to %.4x", bufferAddress)
	//}

	d.DoLED()
	data, err := d.Drives[d.CurrentDrive].GetRawBlock(block)
	if err == nil {

		log.Printf("Block leader: %.2x %.2x %.2x %.2x %.2x %.2x %.2x %.2x",
			data[0], data[1], data[2], data[3], data[4], data[5], data[6], data[7])

		mm := d.Int.GetMemoryMap()
		index := d.Int.GetMemIndex()
		for i, v := range data {
			mm.WriteInterpreterMemory(index, bufferAddress+i, uint64(v))
		}
		log.Printf("Fetched block %d -> 0x%.4x", block, bufferAddress)
	} else {
		log.Printf("Failed to fetch block %d", block)
	}
}

func (d *IOCardSmartPort) mliWrite(block int, bufferAddress int) {
	d.WriteUpdateCycles = d.GlobalCycles + WriteOffCycles // set a pending flush
	d.DoLED()
	data := make([]byte, 512)
	mm := d.Int.GetMemoryMap()
	index := d.Int.GetMemIndex()
	for i, _ := range data {
		data[i] = byte(mm.ReadInterpreterMemory(index, bufferAddress+i))
	}
	if d.Drives[d.CurrentDrive].SetRawBlock(block, data) == nil {
		log.Printf("Updated block %d from 0x%.4x", block, bufferAddress)
	} else {
		log.Printf("Failed to update block %d", block)
	}
}

func (d *IOCardSmartPort) handleSmartport() {
	cpu := apple2helpers.GetCPU(d.Int)
	r := d.callSmartPort()
	cpu.A = int(r)
	if r == spetNoError {
		log.Printf("Smartpoint returned ok")
		cpu.SetFlag(mos6502.F_C, false)
	} else {
		cpu.SetFlag(mos6502.F_C, true)
	}
}

func (d *IOCardSmartPort) callSmartPort() SmartPortErrorType {
	d.DoLED()
	cpu := apple2helpers.GetCPU(d.Int)
	mm := d.Int.GetMemoryMap()
	index := d.Int.GetMemIndex()

	callAddress := cpu.Pop() + 256*cpu.Pop() + 1
	command := mm.ReadInterpreterMemory(index, callAddress)
	extendedCall := command >= 0x040

	log.Printf("==========================================================SMARTPORT CALL")
	log.Printf("Got smartport call address of %.4x", callAddress)
	log.Printf("Command is %.2x", command)
	log.Printf("Call extended? %v", extendedCall)

	var retAddr int
	if extendedCall {
		retAddr = callAddress + 5
	} else {
		retAddr = callAddress + 3
	}

	log.Printf("return address will be %.4x", retAddr)
	cpu.Push((retAddr - 1) / 256)
	cpu.Push((retAddr - 1) % 256)

	// Calculate parameter address block
	var parmAddr int
	parmAddr = int(mm.ReadInterpreterMemory(index, callAddress+1) + 256*mm.ReadInterpreterMemory(index, callAddress+2))
	var paramCount = int(mm.ReadInterpreterMemory(index, parmAddr))

	// Now process command
	log.Printf("Call with command %.2x with param block at address %.4x (length %d)", command, parmAddr, paramCount)

	var params [16]int
	for i := 0; i < len(params); i++ {
		value := int(0x0ff & mm.ReadInterpreterMemory(index, parmAddr+i))
		params[i] = value
		log.Printf("Param #%d: %.2x", i, value)
	}

	unitNumber := params[1] - 1 // unitNumber is 1 based for SP, 0 based internally us
	if !d.changeUnit(unitNumber) {
		log.Printf("Invalid unit: %d", unitNumber)
		return spetInvalidUnit
	}
	dataBuffer := params[2] | (params[3] << 8)

	switch command {
	case 0:
		return spetNoError
	case 1:
		blockNum := params[4] | (params[5] << 8) | (params[6] << 16)
		d.read(blockNum, dataBuffer)
		log.Printf("after read call")
		return spetNoError
	case 2:
		blockNum := params[4] | (params[5] << 8) | (params[6] << 16)
		d.write(blockNum, dataBuffer)
		return spetNoError
	case 3:
		log.Printf("Unimplemented command %d", command)
		return spetInvalidCommand
	case 4:
		log.Printf("Unimplemented command %d", command)
		return spetInvalidCommand
	case 5:
		log.Printf("Unimplemented command %d", command)
		return spetInvalidCommand
	case 6:
		log.Printf("Unimplemented command %d", command)
		return spetInvalidCommand
	case 7:
		log.Printf("Unimplemented command %d", command)
		return spetInvalidCommand
	case 8:
		log.Printf("Unimplemented command %d", command)
		return spetInvalidCommand
	case 9:
		log.Printf("Unimplemented command %d", command)
		return spetInvalidCommand
	default:
		log.Printf("Unimplemented command %d", command)
		return spetInvalidCommand
	}

}

func (d *IOCardSmartPort) read(block int, buffer int) {
	log.Printf("Read of block %d into buffer at %d", block, buffer)
	data, err := d.Drives[d.CurrentDrive].GetRawBlock(block)
	if err == nil {
		for i, v := range data {
			d.Int.GetMemoryMap().WriteInterpreterMemory(d.Int.GetMemIndex(), buffer+i, uint64(v))
		}
	}
}

func (d *IOCardSmartPort) write(block int, buffer int) {
	log.Printf("Write to block %d from buffer at %d", block, buffer)
	data := make([]byte, 512)
	for i, _ := range data {
		data[i] = byte(d.Int.GetMemoryMap().ReadInterpreterMemorySilent(d.Int.GetMemIndex(), buffer+i))
	}
	d.Drives[d.CurrentDrive].SetRawBlock(block, data)
}

func (d *IOCardSmartPort) HandleServiceBusRequest(r *servicebus.ServiceBusRequest) (*servicebus.ServiceBusResponse, bool) {

	log.Printf("Got ServiceBusRequest(%+v)", r.Type)

	switch r.Type {
	case servicebus.SmartPortFlush:
		for _, v := range d.Drives {
			if v != nil {
				v.CheckUpdate()
			}
		}

	case servicebus.SmartPortEject:
		for i, v := range d.Drives {
			if v != nil {
				v.CheckUpdate()
				d.Drives[i] = nil
			}
		}
		settings.PureBootSmartVolume[d.Int.GetMemIndex()] = ""
	case servicebus.SmartPortInsertBytes:
		t := r.Payload.(servicebus.DiskTargetBytes)
		var s *SmartPortBlockDevice
		var err error
		data := t.Bytes

		t.Filename = strings.Replace(t.Filename, "//", "/", -1)
		if !strings.HasPrefix(t.Filename, "local:") {
			t.Filename = "/" + strings.Trim(t.Filename, "/")
		}

		if string(data[:4]) == "2IMG" {
			log.Println("2img")
			s, err = NewSmartPortBlockDevice(data, t.Filename)
		} else {
			log.Println("not 2img")
			s, err = NewSmartPortBlockDeviceNoHeader(data, t.Filename)
		}
		if err == nil {
			d.Attach(0, s)
			log.Printf("Connect smartport disk (bytes): %+v", t.Filename)
			settings.PureBootSmartVolume[d.Int.GetMemIndex()] = t.Filename
		}
	case servicebus.SmartPortInsertFilename:
		t := r.Payload.(servicebus.DiskTargetString)
		log.Printf("Mount smartport volume: %d:%s", t.Drive, t.Filename)
		t.Filename = strings.Replace(t.Filename, "//", "/", -1)
		if !strings.HasPrefix(t.Filename, "local:") {
			t.Filename = "/" + strings.Trim(t.Filename, "/")
		}
		fn := t.Filename
		if strings.HasPrefix(fn, "local:") {
			fn = fn[6:]
			data, err := files.ReadBytes(fn)
			if err == nil {
				var s *SmartPortBlockDevice
				var err error
				if string(data[:4]) == "2IMG" {
					log.Println("2img")
					s, err = NewSmartPortBlockDevice(data, t.Filename)
				} else {
					log.Println("not 2img")
					s, err = NewSmartPortBlockDeviceNoHeader(data, t.Filename)
				}
				if err == nil {
					d.Attach(0, s)
					log.Printf("Connect smartport disk (bytes): %+v", t.Filename)
					settings.PureBootSmartVolume[d.Int.GetMemIndex()] = t.Filename
				}
			}
		} else {
			data, err := files.ReadBytesViaProvider(files.GetPath(fn), files.GetFilename(fn))
			if err == nil {
				var s *SmartPortBlockDevice
				var err error
				if string(data.Content[:4]) == "2IMG" {
					s, err = NewSmartPortBlockDevice(data.Content, t.Filename)
				} else {
					s, err = NewSmartPortBlockDeviceNoHeader(data.Content, t.Filename)
				}
				if err == nil {
					d.Attach(0, s)
					log.Printf("Connect smartport disk (bytes): %+v", t.Filename)
					settings.PureBootSmartVolume[d.Int.GetMemIndex()] = t.Filename
				}
			}
		}
	}

	return &servicebus.ServiceBusResponse{
		Payload: true,
	}, true

}
