package freeze

import (
	"io/ioutil"
	"os"
	"time"

	"paleotronic.com/core/hardware"
	"paleotronic.com/core/hardware/apple2"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/fmt"
	"paleotronic.com/log"
	"paleotronic.com/utils"

	yaml "gopkg.in/yaml.v2"
)

/*
This package contains code to freeze a running instance of a CPU in a slot
and store the data.

*/

type CPURegs struct {
	A         int
	X         int
	Y         int
	SP        int
	P         int
	PC        int
	SPEED     int
	ScanCycle int
}

type IntState struct {
	PC        *types.CodeRef
	LPC       *types.CodeRef
	Alg       *types.Algorithm
	DAlg      *types.Algorithm
	CallStack *interfaces.CallStack
	State     types.EntityState
	SubState  types.EntitySubState
	Dialect   string
}

type FreezeState struct {
	Created          time.Time
	OMode            bool
	OState           IntState
	SkipRAM          bool
	Profile          string
	CPU              CPURegs
	MemMode          int
	VidMode          int
	ZMem             []byte
	Disk0, Disk1     string
	Disk0WP, Disk1WP bool
	Blobs            map[string][]byte
	ActiveHUD        map[string]bool
	ActiveGFX        map[string]bool
	Compressor       string
}

func NewEmptyState(ent interfaces.Interpretable) *FreezeState {

	f := &FreezeState{
		Created: time.Now(),
		Blobs:   make(map[string][]byte),
	}

	return f
}

func NewFreezeState(ent interfaces.Interpretable, skipram bool) *FreezeState {

	for ent.GetChild() != nil {
		ent = ent.GetChild()
	}

	ent.StopTheWorld()

	f := &FreezeState{
		Created: time.Now(),
		Blobs:   make(map[string][]byte),
		SkipRAM: skipram,
	}

	// CPU Registers
	cpu := apple2helpers.GetCPU(ent)
	f.CPU.A = cpu.A
	f.CPU.X = cpu.X
	f.CPU.Y = cpu.Y
	f.CPU.SP = cpu.SP
	f.CPU.P = cpu.P
	f.CPU.PC = cpu.PC

	if cpu.Halted {
		f.OMode = true
		f.OState = IntState{
			PC:        ent.GetPC(),
			LPC:       ent.GetLPC(),
			Alg:       ent.GetCode(),
			DAlg:      ent.GetDirectAlgorithm(),
			CallStack: ent.GetStack(),
			State:     ent.GetState(),
			SubState:  ent.GetSubState(),
			Dialect:   ent.GetDialect().GetShortName(),
		}
	}

	a2io, ok := ent.GetMemoryMap().InterpreterMappableAtAddress(ent.GetMemIndex(), 0xc000)
	if ok {
		f.MemMode = int(a2io.(*apple2.Apple2IOChip).GetMemMode())
		f.VidMode = int(a2io.(*apple2.Apple2IOChip).GetVidMode())

		log.Printf("f.VidMode = %s\n", a2io.(*apple2.Apple2IOChip).GetVidMode().String())

		for i := 1; i < 7; i++ {
			card := a2io.(*apple2.Apple2IOChip).GetCard(i)
			if card != nil {
				log.Printf("Saving card state: %s", card.CardName())
				f.Blobs[card.CardName()] = card.GetYAML()
			}
		}
	}

	if !skipram {
		b, _ := ent.FreezeBytes()
		f.ZMem = utils.XZBytes(b)
		f.Compressor = "xz"
	}

	f.Profile = ent.GetSpec()
	f.Disk0 = settings.PureBootVolume[ent.GetMemIndex()]
	f.Disk1 = settings.PureBootVolume2[ent.GetMemIndex()]
	f.Disk0WP = settings.PureBootVolumeWP[ent.GetMemIndex()]
	f.Disk1WP = settings.PureBootVolumeWP2[ent.GetMemIndex()]

	// woz
	if dsk := apple2.GetDisk(ent, 0); dsk != nil {
		log.Printf("Disk %d is %s:%s", 0, "WOZ", f.Disk0)
		f.Blobs["nib0"] = dsk.GetData().ByteSlice(0, dsk.GetSize())
	}
	if dsk := apple2.GetDisk(ent, 1); dsk != nil {
		log.Printf("Disk %d is %s:%s", 1, "WOZ", f.Disk1)
		f.Blobs["nib1"] = dsk.GetData().ByteSlice(0, dsk.GetSize())
	}
	// nib
	if dsk := apple2.GetDiskNib(ent, 0); dsk != nil {
		log.Printf("Disk %d is %s:%s", 0, "NIB", f.Disk0)
		f.Blobs["nib2"] = dsk.GetNibbles()
	}
	if dsk := apple2.GetDiskNib(ent, 1); dsk != nil {
		log.Printf("Disk %d is %s:%s", 1, "NIB", f.Disk1)
		f.Blobs["nib3"] = dsk.GetNibbles()
	}

	f.ActiveGFX, f.ActiveHUD = apple2helpers.GetActiveLayers(ent)

	ent.ResumeTheWorld()

	return f
}

func (f *FreezeState) SaveToFile(filename string) error {
	data, _ := yaml.Marshal(f)
	ff, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer ff.Close()

	ff.Write(utils.XZBytes(data))
	return nil
}

func (f *FreezeState) SaveToBytes() []byte {
	data, _ := yaml.Marshal(f)
	return utils.XZBytes(data)
}

func (f *FreezeState) LoadFromBytes(data []byte) error {

	if string(data[0:7]) != "created" {
		data = utils.UnXZBytes(data)
	}

	err := yaml.Unmarshal(data, f)
	return err
}

func (f *FreezeState) LoadFromFile(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return f.LoadFromBytes(data)
}

// Apply freeze state to interpreter slot
func (f *FreezeState) Apply(ent interfaces.Interpretable) error {

	for ent.GetChild() != nil {
		ent = ent.GetChild()
	}

	settings.VideoSuspended = true
	defer func() {
		settings.VideoSuspended = false
	}()

	ent.StopTheWorld()

	// Restore Spec
	if !f.SkipRAM {
		ent.SetSpec("")
		ent.LoadSpec(f.Profile)
	}

	// Restore CPU State
	cpu := apple2helpers.GetCPU(ent)
	cpu.A = f.CPU.A
	cpu.X = f.CPU.X
	cpu.Y = f.CPU.Y
	cpu.PC = f.CPU.PC
	cpu.SP = f.CPU.SP
	cpu.P = f.CPU.P

	cpu.Halted = f.OMode

	if f.OMode {
		ent.Bootstrap(f.OState.Dialect, true)
		ent.SetDirectAlgorithm(f.OState.DAlg)
		ent.SetCode(f.OState.Alg)
		ent.SetPC(f.OState.PC)
		ent.SetLPC(f.OState.LPC)
		ent.SetStack(f.OState.CallStack)
		ent.SetState(f.OState.State)
		ent.SetSubState(f.OState.SubState)
	}

	// Restore RAM State
	if !f.SkipRAM {
		oldls := ent.GetMemoryMap().IntGetLayerState(ent.GetMemIndex())
		if f.Compressor == "xz" {
			ent.ThawBytesNoPost(utils.UnXZBytes(f.ZMem))
		} else {
			ent.ThawBytesNoPost(utils.UnGZIPBytes(f.ZMem))
		}
		ent.GetMemoryMap().IntSetLayerState(ent.GetMemIndex(), oldls)
	}

	// Restore IO/Video state
	a2io, ok := ent.GetMemoryMap().InterpreterMappableAtAddress(ent.GetMemIndex(), 0xc000)
	if ok {
		fmt.Printf("Softswitches: %s\n", apple2.VideoFlag(f.VidMode).String())
		a2io.(*apple2.Apple2IOChip).SetMemModeForce(apple2.MemoryFlag(f.MemMode))
		a2io.(*apple2.Apple2IOChip).SetVidModeForce(apple2.VideoFlag(f.VidMode))
		log.Printf("Unpack f.VidMode = %s\n", apple2.VideoFlag(f.VidMode).String())
		log.Printf("f.VidMode after setting = %s\n", a2io.(*apple2.Apple2IOChip).GetVidMode().String())
		a2io.(*apple2.Apple2IOChip).ConfigureVideo()

		for i := 1; i < 7; i++ {
			card := a2io.(*apple2.Apple2IOChip).GetCard(i)
			if card != nil {
				data := f.Blobs[card.CardName()]
				if len(data) > 0 {
					log.Printf("Restoring card state: %s", card.CardName())
					card.SetYAML(data)
				}
			}
		}
	}

	// Restore Disks
	settings.PureBootVolume[ent.GetMemIndex()] = f.Disk0
	settings.PureBootVolume2[ent.GetMemIndex()] = f.Disk1

	if data, ok := f.Blobs["nib0"]; ok {
		log.Printf("Disk %d is %s:%s", 0, "WOZ", f.Disk0)
		apple2.DiskInsertBin(ent, 0, data, f.Disk0, f.Disk0WP)
	}
	if data, ok := f.Blobs["nib1"]; ok {
		log.Printf("Disk %d is %s:%s", 1, "WOZ", f.Disk1)
		apple2.DiskInsertBin(ent, 1, data, f.Disk1, f.Disk1WP)
	}
	if data, ok := f.Blobs["nib2"]; ok {
		log.Printf("Disk %d is %s:%s", 0, "NIB", f.Disk0)
		apple2.DiskInsertBin(ent, 0, data, f.Disk0, f.Disk0WP)
	}
	if data, ok := f.Blobs["nib3"]; ok {
		log.Printf("Disk %d is %s:%s", 1, "NIB", f.Disk1)
		apple2.DiskInsertBin(ent, 1, data, f.Disk1, f.Disk1WP)
	}

	settings.PureBootVolumeWP[ent.GetMemIndex()] = f.Disk0WP
	settings.PureBootVolumeWP2[ent.GetMemIndex()] = f.Disk1WP

	if f.ActiveGFX != nil {
		apple2helpers.SetActiveLayers(ent, f.ActiveGFX, f.ActiveHUD)
	}

	ent.GetMemoryMap().KeyBufferEmpty(ent.GetMemIndex())

	ent.ResumeTheWorld()

	return nil

}

func (f *FreezeState) ApplyIO(ent interfaces.Interpretable) error {

	for ent.GetChild() != nil {
		ent = ent.GetChild()
	}

	ent.StopTheWorld()

	// Restore Spec
	ent.SetSpec("")
	//ent.LaadSpec(f.Profile)
	hardware.LoadIOToInterpreter(ent, f.Profile)

	// Restore CPU State
	cpu := apple2helpers.GetCPU(ent)
	cpu.A = f.CPU.A
	cpu.X = f.CPU.X
	cpu.Y = f.CPU.Y
	cpu.PC = f.CPU.PC
	cpu.SP = f.CPU.SP
	cpu.P = f.CPU.P

	cpu.Halted = f.OMode

	if f.OMode {
		ent.Bootstrap(f.OState.Dialect, true)
		ent.SetDirectAlgorithm(f.OState.DAlg)
		ent.SetCode(f.OState.Alg)
		ent.SetPC(f.OState.PC)
		ent.SetLPC(f.OState.LPC)
		ent.SetStack(f.OState.CallStack)
		ent.SetState(f.OState.State)
		ent.SetSubState(f.OState.SubState)
	}

	// Restore RAM State
	if !f.SkipRAM {
		ent.ThawBytesNoPost(utils.UnGZIPBytes(f.ZMem))
	}

	// Restore IO/Video state
	a2io, ok := ent.GetMemoryMap().InterpreterMappableAtAddress(ent.GetMemIndex(), 0xc000)
	if ok {
		fmt.Printf("Softswitches: %s\n", apple2.VideoFlag(f.VidMode).String())
		a2io.(*apple2.Apple2IOChip).SetMemModeForce(apple2.MemoryFlag(f.MemMode))
		a2io.(*apple2.Apple2IOChip).SetVidModeForce(apple2.VideoFlag(f.VidMode))
	}

	// Restore Disks
	settings.PureBootVolume[ent.GetMemIndex()] = f.Disk0
	settings.PureBootVolume2[ent.GetMemIndex()] = f.Disk1

	// apple2.DiskInsertBin(ent, 0, f.Blobs["nib0"], f.Disk0, f.Disk0WP)
	// apple2.DiskInsertBin(ent, 1, f.Blobs["nib1"], f.Disk1, f.Disk1WP)

	settings.PureBootVolumeWP[ent.GetMemIndex()] = f.Disk0WP
	settings.PureBootVolumeWP2[ent.GetMemIndex()] = f.Disk1WP

	ent.GetMemoryMap().KeyBufferEmpty(ent.GetMemIndex())

	ent.ResumeTheWorld()

	return nil

}
