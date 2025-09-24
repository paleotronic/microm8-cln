package core

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"paleotronic.com/update"
	"runtime"
	"strings"
	"sync"
	"time"

	tomb "gopkg.in/tomb.v2"
	s8webclient "paleotronic.com/api"
	"paleotronic.com/core/dialect/applesoft"
	"paleotronic.com/core/dialect/logo"
	"paleotronic.com/core/dialect/plus"
	"paleotronic.com/core/hardware"
	"paleotronic.com/core/hardware/apple2"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/hardware/common"
	"paleotronic.com/core/hardware/cpu"
	"paleotronic.com/core/hardware/cpu/mos6502"
	"paleotronic.com/core/hardware/cpu/z80"
	"paleotronic.com/core/hardware/restalgia"
	"paleotronic.com/core/hardware/servicebus"
	"paleotronic.com/core/hardware/spectrum"
	"paleotronic.com/core/hardware/spectrum/snapshot"
	"paleotronic.com/core/hires"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/interpreter"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/filerecord"
	"paleotronic.com/files"
	"paleotronic.com/freeze"
	"paleotronic.com/octalyzer/bus"
	"paleotronic.com/octalyzer/video/font"
	pnc "paleotronic.com/panic"
	"paleotronic.com/presentation"
	"paleotronic.com/runestring"
	"paleotronic.com/utils"
)

type VMMode int

const (
	VMModeInterpreter VMMode = iota
	VMModeEmulator
)

type handlerFunc func(task *Task) (interface{}, error)

// VM Contains the profile and structure of a VM
type VM struct {
	TaskPerformer // we embed task performer here
	Index         int
	RAM           *memory.MemoryMap
	Mode          VMMode
	SpecName      string
	SpecFile      string
	SpecData      *hardware.MachineSpec
	GFXLayers     []*types.LayerSpecMapped
	HUDLayers     []*types.LayerSpecMapped
	cpu6502       *mos6502.Core6502
	cpuZ80        *z80.CoreZ80
	e             interfaces.Interpretable
	// private
	t                   *tomb.Tomb
	handlers            map[string]handlerFunc
	p                   *Producer
	mustExitImmediately bool
	CoreCPU             string
	Dependants          []*VM
	Parent              *VM
	sync.Mutex
}

// NewVM creates a new VM with attached RAM and hardware
func NewVM(
	index int,
	ram *memory.MemoryMap,
	prod *Producer,
	specfile string,
	mode VMMode,
) (*VM, error) {

	vm := &VM{
		Index:      index,
		RAM:        ram,
		p:          prod,
		handlers:   make(map[string]handlerFunc),
		Dependants: make([]*VM, 0),
	}
	vm.Alloc()

	// bind controlling VM to interpreter
	//vm.e.Bind(vm)
	vm.e = vm.CreateInterpreter(index, "main", applesoft.NewDialectApplesoft(), specfile)

	vm.Logf("====================== VM#%d is starting ====================", vm.Index)
	err := vm.LoadSpec(specfile)
	vm.Logf("====================== VM#%d is spec loaded ====================", vm.Index)

	return vm, err

}

func (vm *VM) Alloc() {
	if vm.RAM.Data[vm.Index] != nil {
		return
	}
	var tmp [memory.OCTALYZER_INTERPRETER_SIZE]uint64
	vm.RAM.Data[vm.Index] = tmp[:]
}

func (vm *VM) DeAlloc() {
	// vm.RAM.Data[vm.Index] = nil
}

func (vm *VM) AddDependant(dvm *VM) {
	vm.Lock()
	defer vm.Unlock()
	dvm.Parent = vm
	vm.Dependants = append(vm.Dependants, dvm)
}

func (vm *VM) StopDependants() {
	vm.Lock()
	defer vm.Unlock()
	for i, dvm := range vm.Dependants {
		vm.Logf("Stopping dependant %d (VM#%d)", i, dvm.Index+1)
		dvm.Parent = nil
		dvm.Stop()
	}
	vm.Dependants = make([]*VM, 0)
}

func (vm *VM) RemoveDependant(dvm *VM) {
	vm.Lock()
	defer vm.Unlock()
	var found = -1
	for i, v := range vm.Dependants {
		if v == dvm {
			found = i
			break
		}
	}
	if found != -1 {
		vm.Dependants = append(vm.Dependants[:found], vm.Dependants[found+1:]...)
	}
}

func (vm *VM) CreateInterpreter(slot int, name string, dia interfaces.Dialecter, spec string) interfaces.Interpretable {

	/* vars */
	var result *interpreter.Interpreter
	if slot >= memory.OCTALYZER_NUM_INTERPRETERS || slot < 0 {
		return nil
	}
	result = interpreter.NewInterpreter(name, dia, nil, vm.RAM, slot, spec, nil, vm)
	result.SetUUID(0)
	vm.p.MasterLayerPos[slot] = types.LayerPosMod{0, 0}
	vm.p.Context = slot
	result.SetProducer(vm.p)
	return result
}

func (vm *VM) Logf(pattern string, args ...interface{}) {
	// msg := fmt.Sprintf(pattern, args...)
	// log.Printf("[vm#%d] %s", vm.Index, msg)
}

func (vm *VM) IsHostingInterpreter() bool {
	return vm.GetInterpreter() != nil
}

func (vm *VM) SetInterpreter(e interfaces.Interpretable) {
	vm.e = e
}

func (vm *VM) GetInterpreter() interfaces.Interpretable {
	i := vm.e
	for i != nil && i.GetChild() != nil {
		i = i.GetChild()
	}
	return i
}

func (vm *VM) EnableHUDLayers(enabled map[string]bool) {

	for _, l := range vm.HUDLayers {
		if l == nil {
			continue
		}
		l.SetActive(enabled[l.GetID()])
	}

}

func (vm *VM) DisableHUDLayers() map[string]bool {
	var enabled = map[string]bool{}
	for _, l := range vm.HUDLayers {
		if l == nil {
			continue
		}
		enabled[l.GetID()] = l.GetActive()
		l.SetActive(false)
	}
	return enabled
}

func (vm *VM) EnableGFXLayers(enabled map[string]bool) {
	for _, l := range vm.GFXLayers {
		if l == nil {
			continue
		}
		l.SetActive(enabled[l.GetID()])
	}
}

func (vm *VM) DisableGFXLayers() map[string]bool {
	var enabled = map[string]bool{}
	for _, l := range vm.GFXLayers {
		if l == nil {
			continue
		}
		enabled[l.GetID()] = l.GetActive()
		l.SetActive(false)
	}
	return enabled
}

func (vm *VM) IsDying() bool {
	return vm.mustExitImmediately
}

func (vm *VM) GetCPU6502() *mos6502.Core6502 {
	return vm.cpu6502
}

func (vm *VM) GetCPUZ80() *z80.CoreZ80 {
	return vm.cpuZ80
}

func (vm *VM) GetMemoryMap() *memory.MemoryMap {
	return vm.RAM
}

func (vm *VM) GetMemIndex() int {
	return vm.Index
}

func (vm *VM) GetMemory(addr int) uint64 {
	return vm.RAM.ReadInterpreterMemory(vm.Index, addr)
}

func (vm *VM) SetMemory(addr int, value uint64) {
	vm.RAM.WriteInterpreterMemory(vm.Index, addr, value)
}

func (vm *VM) KillWatcher() error {
	<-vm.t.Dying()
	vm.mustExitImmediately = true
	return nil
}

func (vm *VM) LoadSpec(specfile string) error {

	vm.Logf("Commencing load of spec %s", specfile)

	ms, err := hardware.LoadSpec(specfile)
	if err != nil {
		return err
	}

	if ms.CPU.VerticalRetraceCycles == 0 {
		ms.CPU.VerticalRetraceCycles = hardware.DefaultMachineSpec.VerticalRetraceCycles
	}

	if ms.CPU.VBlankCycles == 0 {
		ms.CPU.VBlankCycles = hardware.DefaultMachineSpec.VBlankCycles
	}

	if ms.CPU.FPS == 0 {
		ms.CPU.FPS = hardware.DefaultMachineSpec.FPS
	}

	if ms.CPU.ScanCycles == 0 {
		ms.CPU.ScanCycles = hardware.DefaultMachineSpec.ScanCycles
	}

	vm.SpecFile = specfile
	settings.SystemID[vm.Index] = ms.ID

	// clear overrides
	settings.GlobalOverrides[vm.Index] = map[string]interface{}{}
	settings.ForcePureBoot[vm.Index] = ms.AllowDisklessBoot
	settings.PreventSuppressAlt[vm.Index] = ms.AltPassThrough

	// caps
	vm.RAM.IntSetUppercaseOnly(vm.Index, ms.CapsOnly)

	if ms.Font != "" && settings.FirstBoot[vm.Index] {
		vm.Logf("Loading default font: %s", ms.Font)
		f, err := font.LoadFromFile(ms.Font)
		if err != nil {
			panic(err)
		}
		settings.DefaultFont[vm.Index] = f
		settings.Font[vm.Index] = ms.Font
	}

	vm.Logf("Storing auxilary fonts: %+v", ms.AuxFonts)
	settings.AuxFonts[vm.Index] = append([]string{ms.Font}, ms.AuxFonts...)

	vm.Logf("CPU model is: %s", ms.CPU.Model)
	settings.CPUModel[vm.Index] = ms.CPU.Model
	vm.Logf("CPU Clocks are %d", ms.CPU.Clocks)
	settings.CPUClock[vm.Index] = ms.CPU.Clocks

	vm.Logf("constructing processors")
	apple2helpers.TrashCPU(vm.e)
	vm.cpu6502 = apple2helpers.GetCPU(vm.e)
	vm.cpuZ80 = apple2helpers.GetZ80CPU(vm.e)

	vm.BuildHardware(ms, false)

	settings.SpecName[vm.Index] = ms.Name
	settings.SetSubtitle(settings.SpecName[vm.Index])

	// Make structures
	vm.Logf("Constructing video layers")

	hudlayers := make([]*types.LayerSpecMapped, memory.OCTALYZER_MAX_HUD_LAYERS)
	gfxlayers := make([]*types.LayerSpecMapped, memory.OCTALYZER_MAX_GFX_LAYERS)

	var rsh [256]memory.ReadSubscriptionHandler
	var esh [256]memory.ExecSubscriptionHandler
	var wsh [256]memory.WriteSubscriptionHandler

	for _, l := range ms.Layers {

		// vm.Logf("Processing layer %s", l.ID)

		if l == nil {
			l = &types.LayerSpec{}
		}

		//fmt.RPrintf("[spec] vm#%d, layer: %d, id: %s\n", vm.Index, lno, l.ID)

		offset := memory.OCTALYZER_LAYERSPEC_SIZE * int(l.Index)
		gbase := vm.RAM.MEMBASE(vm.Index) + memory.OCTALYZER_GFX_BASE + offset
		hbase := vm.RAM.MEMBASE(vm.Index) + memory.OCTALYZER_HUD_BASE + offset

		if len(l.Blocks) > 0 {
			vm.RAM.CreateMemoryHint(vm.Index, l.ID, l.Blocks)
		}
		switch l.Type {
		case 0: // hud
			var tb *types.TextBuffer
			if l.Format == types.LF_TEXT_WOZ {
				memory.WarmStart = true
				tb = types.NewTextBufferMapped(
					false,
					int(l.Base),
					types.W_NORMAL_H_NORMAL,
					memory.NewMappedRegionFromHint(
						vm.RAM,
						vm.Index*memory.OCTALYZER_INTERPRETER_SIZE,
						int(l.Base),
						4096,
						l.Format.String(),
						l.ID,
						rsh,
						esh,
						wsh,
					),
					vm.e,
				)
				memory.WarmStart = false
			} else {
				tb = types.NewTextBufferMapped(
					true,
					int(l.Base),
					types.W_NORMAL_H_NORMAL,
					memory.NewMappedRegionFromHint(
						vm.RAM,
						vm.Index*memory.OCTALYZER_INTERPRETER_SIZE,
						int(l.Base),
						4096,
						l.Format.String(),
						l.ID,
						rsh,
						esh,
						wsh,
					),
					vm.e,
				)
			}
			l.Control = tb

			hudlayers[l.Index%memory.OCTALYZER_MAX_HUD_LAYERS] = types.NewLayerSpecMapped(vm.RAM, l, vm.Index, hbase)
			hudlayers[l.Index%memory.OCTALYZER_MAX_HUD_LAYERS].SetDirty(true) // force rebuild
		case 1: // gfx
			switch l.Format {
			case types.LF_DHGR_WOZ:
				mcb := vm.RAM.GetHintedMemorySlice(vm.Index, l.ID)
				mcb.UseMM = true
				var woz *hires.DHGRScreen = hires.NewDHGRScreen(
					mcb,
				)
				l.HControl = woz
				gfxlayers[l.Index%memory.OCTALYZER_MAX_GFX_LAYERS] = types.NewLayerSpecMapped(vm.RAM, l, vm.Index, gbase)
			case types.LF_HGR_WOZ:
				mcb := vm.RAM.GetHintedMemorySlice(vm.Index, l.ID)
				mcb.UseMM = true
				var woz *hires.HGRScreen = hires.NewHGRScreen(
					mcb,
				)
				l.HControl = woz
				gfxlayers[l.Index%memory.OCTALYZER_MAX_GFX_LAYERS] = types.NewLayerSpecMapped(vm.RAM, l, vm.Index, gbase)
				gfxlayers[l.Index%memory.OCTALYZER_MAX_GFX_LAYERS].SetDirty(true)
			case types.LF_HGR_X:
				var woz *hires.IndexedVideoBuffer = hires.NewIndexedVideoBuffer(
					280,
					192,
					vm.RAM.GetHintedMemorySlice(vm.Index, l.ID),
				)
				l.HControl = woz
				gfxlayers[l.Index%memory.OCTALYZER_MAX_GFX_LAYERS] = types.NewLayerSpecMapped(vm.RAM, l, vm.Index, gbase)
				gfxlayers[l.Index%memory.OCTALYZER_MAX_GFX_LAYERS].SetDirty(true)
			case types.LF_SUPER_HIRES:
				mcb := vm.RAM.GetHintedMemorySlice(vm.Index, l.ID)
				//log.Printf("super hires slice at %.6x, size %.4x", mcb.GStart[0], mcb.Size)
				mcb.UseMM = true
				var woz *hires.SuperHiResBuffer = hires.NewSuperHiResBuffer(
					mcb,
				)
				l.HControl = woz
				gfxlayers[l.Index%memory.OCTALYZER_MAX_GFX_LAYERS] = types.NewLayerSpecMapped(vm.RAM, l, vm.Index, gbase)
				gfxlayers[l.Index%memory.OCTALYZER_MAX_GFX_LAYERS].SetDirty(true)
			case types.LF_LOWRES_WOZ:
				var tb *types.TextBuffer
				memory.WarmStart = true
				tb = types.NewTextBufferMapped(
					false,
					int(l.Base),
					types.W_NORMAL_H_NORMAL,
					memory.NewMappedRegionFromHint(
						vm.RAM,
						vm.Index*memory.OCTALYZER_INTERPRETER_SIZE,
						int(l.Base),
						4096,
						l.Format.String(),
						l.ID,
						rsh,
						esh,
						wsh,
					),
					vm.e,
				)
				memory.WarmStart = false
				l.Control = tb
				gfxlayers[l.Index%memory.OCTALYZER_MAX_GFX_LAYERS] = types.NewLayerSpecMapped(vm.RAM, l, vm.Index, gbase)
			case types.LF_LOWRES_LINEAR:
				var tb *types.TextBuffer
				tb = types.NewTextBufferMapped(
					true,
					int(l.Base),
					types.W_NORMAL_H_NORMAL,
					memory.NewMappedRegionFromHint(
						vm.RAM,
						vm.Index*memory.OCTALYZER_INTERPRETER_SIZE,
						int(l.Base),
						4096,
						l.Format.String(),
						l.ID,
						rsh,
						esh,
						wsh,
					),
					vm.e,
				)
				l.Control = tb
				gfxlayers[l.Index%memory.OCTALYZER_MAX_GFX_LAYERS] = types.NewLayerSpecMapped(vm.RAM, l, vm.Index, gbase)
			case types.LF_VECTOR:
				var vb *types.VectorBuffer
				vb = types.NewVectorBufferMapped(
					int(l.Base),
					0x10000,
					memory.NewMappedRegionFromHint(
						vm.RAM,
						vm.Index*memory.OCTALYZER_INTERPRETER_SIZE,
						int(l.Base),
						0x10000,
						l.Format.String(),
						l.ID,
						rsh,
						esh,
						wsh,
					),
				)
				l.VControl = vb
				gfxlayers[l.Index%memory.OCTALYZER_MAX_GFX_LAYERS] = types.NewLayerSpecMapped(vm.RAM, l, vm.Index, gbase)
			case types.LF_CUBE_PACKED:
				cb := types.NewCubeScreen(
					int(l.Base),
					0x10000,
					vm.RAM.GetHintedMemorySlice(vm.Index, l.ID),
				)
				l.CubeControl = cb
				//log.Printf("Added Layer %s with CubeControl = %v", l.ID, l.CubeControl)
				gfxlayers[l.Index%memory.OCTALYZER_MAX_GFX_LAYERS] = types.NewLayerSpecMapped(vm.RAM, l, vm.Index, gbase)
			default:
				gfxlayers[l.Index%memory.OCTALYZER_MAX_GFX_LAYERS] = types.NewLayerSpecMapped(vm.RAM, l, vm.Index, gbase)
			}
		}
	}

	time.Sleep(10 * time.Millisecond)

	vm.HUDLayers = hudlayers
	vm.GFXLayers = gfxlayers

	return nil

}

func (vm *VM) BuildHardware(ms hardware.MachineSpec, skipreset bool) {

	//fmt.Println("NOT SKIPPING MACHINE")
	vm.Logf("Starting hardware build")

	// RAM mappings
	mbm := memory.NewMemoryManagementUnit()

	if vm.RAM.BlockMapper[vm.Index] != nil {
		vm.RAM.BlockMapper[vm.Index].Done()
	}

	vm.RAM.BlockMapper[vm.Index] = mbm
	vm.cpu6502.SetMapper(mbm)

	for _, rb := range ms.RAM {
		vm.Logf("Processing memory allocation %s", rb.Name)
		baseaddr := rb.Base
		for _, r := range rb.Regions {
			rname := rb.Name + "." + r.Name
			vm.Logf("Handling region %s @ 0x%.4x - 0x%.4x", rname, r.Base, r.End)

			size := r.End - r.Base + 1

			if rb.Type == 0 || rb.Type == 1 {
				vm.Logf("Type is RAM")
				mb := memory.NewMemoryBlockRAM(vm.RAM, vm.Index, baseaddr, r.Base, size, r.Active, rname, r.Mux, r.ForceMux, rb.Type)
				if r.Mode != "" {
					mb.SetState(r.Mode)
				}
				mbm.Register(mb)
			} else {
				if r.Blank {
					data := make([]uint64, size)
					for i, _ := range data {
						data[i] = r.Fill
					}
					mb := memory.NewMemoryBlockROM(vm.RAM, vm.Index, baseaddr, r.Base, size, r.Active, rname, data)
					mbm.Register(mb)
				} else {
					data, e := common.LoadData(r.Src, r.SrcBegin, r.SrcLength)
					if e != nil {
						panic(e)
					}
					mb := memory.NewMemoryBlockROM(vm.RAM, vm.Index, baseaddr, r.Base, size, r.Active, rname, data)
					mbm.Register(mb)
				}
			}
		}
	}

	vm.RAM.InterpreterMappings[vm.Index] = make(memory.MapList)

	vm.Logf("Initialising audio ports")
	vm.e.ClearAudioPorts()
	for _, vc := range ms.Restalgia {
		restalgia.CreateVoice(vm.e, vc.Port, vc.Name, vc.Inst)
		vm.e.SetAudioPort(vc.Name, vc.Port)
	}

	for _, dm := range ms.Components {
		vm.Logf("Constructing overlay: %s", dm.Type)
		device := hardware.FactoryProduce(vm.RAM, 0, dm.Base, dm.Type, vm.e, dm.Misc, dm.Options, ms)
		if device == nil {
			continue
		}
		vm.RAM.MapInterpreterRegion(vm.Index, memory.MemoryRange{
			Base: device.GetBase(),
			Size: device.GetSize()},
			device)

		mb := memory.NewMemoryBlockIO(vm.RAM, vm.Index, vm.RAM.MEMBASE(vm.Index), device.GetBase(), device.GetSize(), true, strings.ToLower(device.GetLabel()), device)
		mbm.Register(mb)
	}

	// save default
	mbm.Reset(skipreset)
}

func (vm *VM) findLayer(layer string, layerset []*types.LayerSpecMapped) *types.LayerSpecMapped {
	for _, l := range layerset {
		if l.GetID() == layer {
			return l
		}
	}
	return nil
}

func (vm *VM) WithHUDLayer(layer string, f func(l *types.LayerSpecMapped)) {
	l := vm.findLayer(layer, vm.HUDLayers)
	if l != nil && f != nil {
		f(l)
	}
}

func (vm *VM) WithGFXLayer(layer string, f func(l *types.LayerSpecMapped)) {
	l := vm.findLayer(layer, vm.GFXLayers)
	if l != nil && f != nil {
		f(l)
	}
}

func (vm *VM) GetGFXLayerByID(name string) (*types.LayerSpecMapped, bool) {
	l := vm.findLayer(name, vm.GFXLayers)
	return l, l != nil
}

func (vm *VM) GetHUDLayerByID(name string) (*types.LayerSpecMapped, bool) {
	l := vm.findLayer(name, vm.HUDLayers)
	return l, l != nil
}

func (vm *VM) GetGFXLayerSet() []*types.LayerSpecMapped {
	return vm.GFXLayers
}

func (vm *VM) GetHUDLayerSet() []*types.LayerSpecMapped {
	return vm.HUDLayers
}

// Stop kills execution of the goroutine, sending it the Dying channel
func (vm *VM) Stop() error {
	if vm.t != nil {
		vm.t.Kill(nil)
		return vm.t.Wait()
	}
	if vm.p != nil {
		vm.p.DropVM(vm.Index)
	}
	return nil
}

func (vm *VM) Start(t *tomb.Tomb) {
	if t == nil {
		t = &tomb.Tomb{}
	}
	vm.t = t
	vm.Logf("Starting mainloop")
	vm.t.Go(vm.KillWatcher)
	vm.t.Go(vm.MainLoop)
}

func (vm *VM) ExecutePendingTasks() {
	select {
	case task := <-vm.IncomingTasks:
		vm.Logf("Received task request: %s", task.Action)
		value, err := vm.handleTask(task)
		r := &TaskResponse{
			Err:   err,
			Value: value,
		}
		task.Response <- r
		vm.Logf("Responded to task: %+v", r)
	default:
		// drop out
	}
}

func (vm *VM) MainLoop() error {
	// main execution thread
	var ent interfaces.Interpretable
	var slotid = vm.Index

	var ostate types.EntityState

	vm.GetInterpreter().GetDialect().InitVDU(vm.GetInterpreter(), true)

	pnc.Do(
		func() {
			for {

				select {
				case task := <-vm.IncomingTasks:
					vm.Logf("Received task request: %s", task.Action)
					value, err := vm.handleTask(task)
					r := &TaskResponse{
						Err:   err,
						Value: value,
					}
					task.Response <- r
					vm.Logf("Responded to task: %+v", r)
				case _ = <-vm.t.Dying():
					//log.Printf("vm %d is dying", vm.Index)
					// TODO: shutdown requested for the loop
					vm.Logf("received kill signal... slot %d", vm.Index)
					vm.Teardown()
					return
				default:

					ent = vm.GetInterpreter()
					if ent == nil {
						time.Sleep(5 * time.Millisecond)
						continue
					}

					if producerIsReady && vm.Index == 0 && !settings.BootCheckDone {
						//ent.StopTheWorld()
						if !settings.NoUpdates {
							vm.Logf("update check executing")
							vm.GetInterpreter().BootCheck()
						} else {
							time.Sleep(500 * time.Millisecond)
						}
						vm.GetInterpreter().ResumeTheWorld()
						settings.BootCheckDone = true
					}

					settings.PureBootCheck(vm.Index)

					if ostate != ent.GetState() {
						//vm.Logf("state = %s (%d), was = %s (%d)", ent.GetState(), ent.GetState(), ostate, ostate)
						ostate = ent.GetState()
						if ent.GetState() == types.STOPPED {
							//ent.SetNeedsPrompt(true)
						}
					}

					// run step code here
					if ent.Is6502Executing() {

						//settings.SlotZPEmu[slotid] = false
						var r int
						if vm.cpu6502.RunState != mos6502.CrsPaused {
							r = int(vm.cpu6502.ExecuteSliced())
						} else {
							r = int(cpu.FE_SLEEP)
						}
						//vm.Logf("Executing 6502 cycles -> %s", r)
						if r == int(cpu.FE_SLEEP) {
							if !ent.WaitForWorld() {
								time.Sleep(1 * time.Millisecond)
							} else {
								runtime.Gosched()
							}
						} else if r == int(cpu.FE_HALTED) && !settings.PureBoot(slotid) {
							// Restore state
							time.Sleep(3 * time.Millisecond)
							ent.RestorePrevState()

							if settings.LaunchQuitCPUExit {
								os.Exit(0)
							}
						} else if r == int(cpu.FE_HALTED) {
							//
						} else if r != int(cpu.FE_OK) {

							m := apple2helpers.NewMonitor(ent)
							if !m.Invoke(cpu.FEResponse(r)) {
								ent.Halt6502(r)
							}

							cpu := apple2helpers.GetCPU(ent)
							cpu.ResetSliced()
						}
					} else if ent.GetState() == types.EMPTY {
						ent.SetState(types.STOPPED)
						time.Sleep(1 * time.Millisecond)
					} else if ent.IsZ80Executing() {
						r := ent.DoCyclesZ80()
						if r == int(cpu.FE_SLEEP) {
							time.Sleep(1 * time.Millisecond)
						}
					} else if ent.GetState() == types.PLAYING {
						if !bus.IsClock() {
							bus.StartDefault()
						}
						if len(settings.VideoPlayFrames[slotid]) > 0 {
							ent.PlayBlocks(settings.VideoPlayFrames[slotid], settings.VideoPlayBackwards[slotid], settings.VideoBackSeekMS[slotid])
							settings.VideoPlayFrames[slotid] = []*bytes.Buffer(nil)
							settings.VideoBackSeekMS[slotid] = 0
						} else if settings.VideoPlaybackFile[slotid] != "" {
							//rlog.Printf("*** Playing file: %s (backwards=%v)", settings.VideoPlaybackFile[slotid], settings.VideoPlayBackwards[slotid])
							ent.PlayRecordingCustom(settings.VideoPlaybackFile[slotid], settings.VideoPlayBackwards[slotid])
							settings.VideoPlaybackFile[slotid] = ""
						}
					} else if ent.IsDisabled() {
						time.Sleep(500 * time.Millisecond)
					} else if ent.IsPaused() || ent.IsWaitingForWorld() {
						time.Sleep(1 * time.Millisecond) // 1 ms
					} else if ent.IsRemote() {
						if !bus.IsClock() {
							bus.StartDefault()
						}
						time.Sleep(2 * time.Millisecond)
					} else if ent.IsRunning() {
						if !bus.IsClock() {
							bus.StartDefault()
						}
						settings.SlotZPEmu[slotid] = true
						ent.RunStatement()
						nexttime := ent.GetWaitUntil()
						for time.Now().Before(nexttime) && !vm.IsDying() {
							time.Sleep(500 * time.Microsecond)
						}
					} else if ent.IsRunningDirect() {
						if !bus.IsClock() {
							bus.StartDefault()
						}
						settings.SlotZPEmu[slotid] = true
						ent.RunStatementDirect()
						nexttime := ent.GetWaitUntil()
						for time.Now().Before(nexttime) && !vm.IsDying() {
							time.Sleep(500 * time.Microsecond)
						}
					} else {
						if !bus.IsClock() {
							bus.StartDefault()
						}
						ent.Interactive()
						time.Sleep(5 * time.Millisecond)
						//vm.Logf("[-] Incoming tasks = %d", len(vm.IncomingTasks))
					}
					// end run step code
				}
			}
		},
		func(r interface{}) {
			vm.Logf("panic trapped: %v", r)
			ent := vm.GetInterpreter()

			servicebus.Unsubscribe(slotid, ent)

			// r is an exception...
			b := make([]byte, 8192)
			i := runtime.Stack(b, false)
			// Stack trace
			stackstr := string(b[0:i])

			t := time.Now()

			filename := fmt.Sprintf("%s-vm-crash.log", t.Format("2006-01-02-15-04-05"))

			bb := bytes.NewBuffer([]byte(fmt.Sprintf("\n%v\n\n", r)))
			bb.WriteString(fmt.Sprintf("Build: %s\nGit  : %s\nBuilt: %s\nLevel: vm\n\n", update.GetBuildNumber(), update.GetBuildHash(), update.GetBuildDate()))
			bb.Write(b[0:i])

			files.WriteBytesViaProvider("/local/logs", filename, bb.Bytes())

			tmp, _ := ent.FreezeBytes() // store memory here

			filenameFreeze := fmt.Sprintf("%s-vm-state.frz", t.Format("2006-01-02-15-04-05"))

			files.WriteBytesViaProvider("/local/logs", filenameFreeze, utils.XZBytes(tmp))

			ent.SystemMessage("An exception has occurred - capturing then restarting")

			// Construct record
			bug := filerecord.BugReport{}
			bug.Summary = "System crashed"
			bug.Body = `
	System crash  occurred in Slot#` + utils.IntToStr(slotid) + `
	` + fmt.Sprintf("%v", r) + `

	Stack trace:
	` + stackstr + `
		`
			// Add compressed stuff
			att := filerecord.BugAttachment{}
			att.Name = "Compressed Runstate"
			att.Created = time.Now()

			att.Content = utils.GZIPBytes(tmp)
			bug.Attachments = []filerecord.BugAttachment{att}
			bug.Comments = []filerecord.BugComment{
				filerecord.BugComment{
					User:    "system",
					Content: "Logged automatically by runtime system.",
					Created: time.Now(),
				},
			}
			bug.Creator = s8webclient.CONN.Username
			bug.Filename = ent.GetFileRecord().FileName
			bug.Filepath = ent.GetFileRecord().FilePath
			bug.Created = time.Now()

			if bug.Filename != "" {
				bug.Summary = "Program crash: " + bug.Filepath + "/" + bug.Filename
			}

			fmt.Println(bug.Body)

			_ = s8webclient.CONN.CreateUpdateBug(bug)
			// cleanup
			settings.FirstBoot[vm.Index] = true
			settings.PureBootVolume[vm.Index] = ""
			settings.PureBootVolume2[vm.Index] = ""
			settings.PureBootSmartVolume[vm.Index] = ""
			settings.MicroPakPath = ""
			vm.RAM.IntSetSlotRestart(vm.Index, true)
		})

	return nil
}

func (vm *VM) Teardown() {
	vm.Logf("Teardown started from dying handler for vm in slot %d", vm.Index)
	// debug.PrintStack()

	if vm.Parent != nil {
		vm.Logf("This vm is a child of VM#%d", vm.Parent.Index)
	}

	if len(vm.Dependants) > 0 {
		vm.Logf("This vm has child VMs... stopping them")
		vm.StopDependants()
	}

	//bus.StartDefault()
	ent := vm.GetInterpreter()

	if ent.GetDialect().GetShortName() == "logo" {
		ent.GetDialect().(*logo.DialectLogo).Driver.KillAllCoroutines()
	}

	ent.StopRecordingHard()
	vm.RAM.IntSetBackdrop(vm.Index, "", 7, 0, 1, 1, false)
	ent.StopMusic()
	vm.RAM.IntSetRestalgiaPath(vm.Index, "", false)
	bus.Sync()

	// := apple2helpers.GetCPU(vm.GetInterpreter())
	vm.GetInterpreter().Halt6502(int(cpu.FE_OK))
	vm.GetInterpreter().HaltZ80(int(cpu.FE_OK))

	vm.DisableGFXLayers()
	vm.DisableHUDLayers()

	if strings.HasPrefix(vm.SpecFile, "apple2") {
		vm.Logf("Shutting down apple2 IO routines")
		if tmp, ok := vm.RAM.InterpreterMappableAtAddress(vm.Index, 0xc000); ok {
			io := tmp.(*apple2.Apple2IOChip)
			io.Done()
		}
	}

	// detach producer reference
	vm.p = nil
	e := vm.GetInterpreter()
	for e != nil {
		e.SetChild(nil)
		e.Bind(nil)
		e.SetProducer(nil)
		e = e.GetParent()
	}

	vm.RAM.IntSetLayerState(vm.Index, 0)
	vm.RAM.IntSetActiveState(vm.Index, 0)

	vm.DeAlloc()

	vm.Logf("====================== VM#%d is now dead ====================", vm.Index)
}

func (vm *VM) GetLayers() ([]*types.LayerSpecMapped, []*types.LayerSpecMapped) {
	return vm.HUDLayers, vm.GFXLayers
}

func (vm *VM) registerHandler(action string, f handlerFunc) {
	vm.handlers[action] = f
}

func (vm *VM) handleTask(task *Task) (interface{}, error) {

	switch task.Action {
	case "vm.stop":
		vm.Stop()
		return "ok", nil
	case "vm.interpreter.applesoft":
		vm.e.Bootstrap("fp", false)
		vm.e.SetState(types.STOPPED)
		return "ok", nil
	case "vm.interpreter.paste":
		if len(task.Arguments) > 0 {
			text, ok := task.Arguments[0].(string)
			if !ok {
				return nil, errors.New("expected string for " + task.Action)
			}
			vm.e.SetPasteBuffer(runestring.Cast(text))
			return text, nil
		}
	case "vm.memory.set":
		if len(task.Arguments) != 2 {
			return nil, errors.New("expect 2 int arguments")
		}
		addr, ok := task.Arguments[0].(int)
		if !ok {
			return nil, errors.New("expect int value param 1")
		}
		value, ok := task.Arguments[0].(int)
		if !ok {
			return nil, errors.New("expect int value param 2")
		}
		vm.e.SetMemory(addr, uint64(value))
		return vm.e.GetMemory(addr), nil
	case "vm.hud.select":
		if len(task.Arguments) != 1 {
			return nil, errors.New("expect 1 arguments")
		}
		name, ok := task.Arguments[0].(string)
		if !ok {
			return nil, errors.New("expect string")
		}
		apple2helpers.SelectHUD(vm.e, name)
		return name, nil
	case "vm.hud.mode":
		if len(task.Arguments) != 1 {
			return nil, errors.New("expect 1 arguments")
		}
		name, ok := task.Arguments[0].(string)
		if !ok {
			return nil, errors.New("expect string")
		}
		switch name {
		case "mode40.preserve":
			apple2helpers.MODE40Preserve(vm.e)
			return name, nil
		case "mode40":
			apple2helpers.MODE40(vm.e)
			return name, nil
		}
		return name, nil
	case "vm.hud.fgcolor":
		if len(task.Arguments) != 1 {
			return nil, errors.New("expect 1 int arguments")
		}
		color, ok := task.Arguments[0].(int)
		if !ok {
			return nil, errors.New("expect int value param 1")
		}
		apple2helpers.SetFGColor(vm.e, uint64(color))
		return color, nil
	case "vm.hud.bgcolor":
		if len(task.Arguments) != 1 {
			return nil, errors.New("expect 1 int arguments")
		}
		color, ok := task.Arguments[0].(int)
		if !ok {
			return nil, errors.New("expect int value param 1")
		}
		apple2helpers.SetBGColor(vm.e, uint64(color))
		return color, nil
	case "vm.hud.loadfont":
		if len(task.Arguments) != 1 {
			return nil, errors.New("expect 1 int arguments")
		}
		ff, ok := task.Arguments[0].(string)
		if !ok {
			return nil, errors.New("expect layer sub format value param 1")
		}
		f, err := font.LoadFromFile(ff)
		if err == nil {
			settings.DefaultFont[vm.Index] = f
			settings.ForceTextVideoRefresh = true
		}
		return ff, err
	case "vm.audio.setspeakervolume":
		if len(task.Arguments) != 1 {
			return nil, errors.New("expect 1 int arguments")
		}
		v, ok := task.Arguments[0].(float64)
		if !ok {
			return nil, errors.New("expect float64 value param 1")
		}
		settings.SpeakerVolume[vm.Index] = v
		return v, nil
	case "vm.hud.setsubformat":
		if len(task.Arguments) != 1 {
			return nil, errors.New("expect 1 int arguments")
		}
		sf, ok := task.Arguments[0].(types.LayerSubFormat)
		if !ok {
			return nil, errors.New("expect layer sub format value param 1")
		}
		txt := apple2helpers.GetSelectedHUD(vm.e)
		txt.SetSubFormat(sf)
		return sf, nil
	case "vm.hud.clearscreen":
		apple2helpers.Clearscreen(vm.e)
		return "ok", nil
	case "vm.gfx.camera.reset":
		if len(task.Arguments) != 1 {
			return nil, errors.New("expect 1 int arguments")
		}
		sf, ok := task.Arguments[0].(int)
		if !ok {
			return nil, errors.New("expect layer sub format value param 1")
		}
		control := types.NewOrbitController(
			vm.RAM,
			vm.Index,
			sf,
		)
		control.ResetALL()
		// control.SetPivotLock(true)
		// control.SetZoom(types.GFXMULT)
		return sf, nil
	case "vm.gfx.setbackdrop":
		if len(task.Arguments) != 6 {
			return nil, errors.New("expect 6 arguments")
		}
		file, ok := task.Arguments[0].(string)
		if !ok {
			return nil, errors.New("expect arg 1 string")
		}
		a, ok := task.Arguments[1].(int)
		if !ok {
			return nil, errors.New("expect arg 2 int")
		}
		b, ok := task.Arguments[2].(float32)
		if !ok {
			return nil, errors.New("expect arg 3 float32")
		}
		c, ok := task.Arguments[3].(float32)
		if !ok {
			return nil, errors.New("expect arg 4 float32")
		}
		d, ok := task.Arguments[4].(float32)
		if !ok {
			return nil, errors.New("expect arg 5 float32")
		}
		e, ok := task.Arguments[5].(bool)
		if !ok {
			return nil, errors.New("expect arg 6 bool")
		}
		if file != "" && !files.ExistsViaProvider(files.GetPath(file), files.GetFilename(file)) {
			vm.Logf("[backdrop] Backdrop does not exist, so skipping: %s", file)
			return file, nil
		}
		vm.RAM.IntSetBackdrop(vm.Index, file, a, b, c, d, e)
		// count := 0
		// for vm.RAM.IntGetBackdropIsNew(vm.Index) && count < 10 {
		// 	time.Sleep(10 * time.Millisecond)
		// 	count++
		// }
		// if vm.RAM.IntGetBackdropIsNew(vm.Index) {
		// 	return "", errors.New("timeout waiting for backdrop")
		// }
		return file, nil
	case "vm.gfx.setbackdroppos":
		if len(task.Arguments) != 3 {
			return nil, errors.New("expect 3 arguments")
		}
		x, ok := task.Arguments[0].(float64)
		if !ok {
			return nil, errors.New("expect arg 1 float32")
		}
		y, ok := task.Arguments[1].(float64)
		if !ok {
			return nil, errors.New("expect arg 2 float32")
		}
		z, ok := task.Arguments[2].(float64)
		if !ok {
			return nil, errors.New("expect arg 3 float32")
		}
		vm.RAM.IntSetBackdropPos(vm.Index, x, y, z)
		return "", nil
	case "vm.music.stop":
		vm.GetInterpreter().StopMusic()
		return "ok", nil
	case "vm.palette.reload":
		plus.ResetPaletteList(vm.Index)
		return "ok", nil
	case "vm.input.setpaddlevalue":
		if len(task.Arguments) != 2 {
			return nil, errors.New("expect 2 int arguments")
		}
		paddle, ok := task.Arguments[0].(int)
		if !ok {
			return nil, errors.New("expect int value param 1")
		}
		value, ok := task.Arguments[1].(int)
		if !ok {
			return nil, errors.New("expect int value param 2")
		}
		vm.RAM.IntSetPaddleValue(vm.Index, paddle, uint64(value))
		return vm.RAM.IntGetPaddleValue(vm.Index, paddle), nil
	case "vm.input.setpaddlebutton":
		if len(task.Arguments) != 2 {
			return nil, errors.New("expect 2 int arguments")
		}
		paddle, ok := task.Arguments[0].(int)
		if !ok {
			return nil, errors.New("expect int value param 1")
		}
		value, ok := task.Arguments[1].(int)
		if !ok {
			return nil, errors.New("expect int value param 2")
		}
		vm.RAM.IntSetPaddleButton(vm.Index, paddle, uint64(value))
		return vm.RAM.IntGetPaddleButton(vm.Index, paddle), nil
	case "vm.gfx.setoverlay":
		if len(task.Arguments) != 1 {
			return nil, errors.New("expect 1 arguments")
		}
		name, ok := task.Arguments[0].(string)
		if !ok {
			return nil, errors.New("expect string")
		}
		vm.RAM.IntSetOverlay(vm.Index, name)
		return name, nil
	case "vm.interpreter.bootstrapsilent":
		if len(task.Arguments) != 1 {
			return nil, errors.New("expect 1 arguments")
		}
		dia, ok := task.Arguments[0].(string)
		if !ok {
			return nil, errors.New("expect string param 1")
		}
		vm.GetInterpreter().Bootstrap(dia, true)
		vm.GetInterpreter().SetState(types.EMPTY)
		// vm.GetInterpreter().SetWorkDir("/local")
		settings.SpritesEnabled[vm.Index] = true
		return "ok", nil
	case "vm.interpreter.bootstrap":
		if len(task.Arguments) != 1 {
			return nil, errors.New("expect 1 arguments")
		}
		dia, ok := task.Arguments[0].(string)
		if !ok {
			return nil, errors.New("expect string param 1")
		}
		vm.GetInterpreter().Bootstrap(dia, false)
		vm.GetInterpreter().SetState(types.STOPPED)
		//vm.GetInterpreter().SetWorkDir("/local")
		settings.SpritesEnabled[vm.Index] = true
		return "ok", nil
	case "vm.interpreter.command":
		if len(task.Arguments) != 1 {
			return nil, errors.New("expect 1 arguments")
		}
		command, ok := task.Arguments[0].([]string)
		if !ok {
			return nil, errors.New("expect []string param 1")
		}
		for _, cmd := range command {
			vm.e.ParseImm(cmd)
		}
		return "ok", nil
	case "vm.gfx.setrendermode":
		if len(task.Arguments) != 2 {
			return nil, errors.New("expect 1 arguments")
		}
		mode, ok := task.Arguments[0].(string)
		if !ok {
			return nil, errors.New("expect string param 1")
		}
		kind, ok := task.Arguments[1].(settings.VideoMode)
		if !ok {
			return nil, errors.New("expect videomode param 2")
		}
		switch strings.ToLower(mode) {
		case "zx":
			vm.RAM.IntSetSpectrumRender(vm.Index, kind)
		case "shr":
			vm.RAM.IntSetSHRRender(vm.Index, kind)
		case "gr":
			vm.RAM.IntSetGRRender(vm.Index, kind)
		case "hgr":
			vm.RAM.IntSetHGRRender(vm.Index, kind)
		case "dhgr":
			vm.RAM.IntSetDHGRRender(vm.Index, kind)
		default:
			return nil, errors.New("Unknown render mode: " + mode)
		}
		return kind, nil
	case "vm.gfx.setvoxeldepth":
		if len(task.Arguments) != 1 {
			return nil, errors.New("expect 1 arguments")
		}
		value, ok := task.Arguments[0].(settings.VoxelDepth)
		if !ok {
			return nil, errors.New("expect VoxelDepth param 1")
		}
		vm.RAM.IntSetVoxelDepth(vm.Index, value)
		return vm.RAM.IntGetVoxelDepth(vm.Index), nil
	case "vm.gfx.settintmode":
		if len(task.Arguments) != 1 {
			return nil, errors.New("expect 1 arguments")
		}
		value, ok := task.Arguments[0].(settings.VideoPaletteTint)
		if !ok {
			return nil, errors.New("expect VideoPaletteTint param 1")
		}
		vm.RAM.IntSetVideoTint(vm.Index, value)
		return vm.RAM.IntGetVideoTint(vm.Index), nil
	case "vm.gfx.setambientlevel":
		if len(task.Arguments) != 1 {
			return nil, errors.New("expect 1 arguments")
		}
		value, ok := task.Arguments[0].(float32)
		if !ok {
			return nil, errors.New("expect float32 param 1")
		}
		vm.RAM.IntSetAmbientLevel(vm.Index, value)
		return vm.RAM.IntGetAmbientLevel(vm.Index), nil
	case "vm.gfx.setdiffuselevel":
		if len(task.Arguments) != 1 {
			return nil, errors.New("expect 1 arguments")
		}
		value, ok := task.Arguments[0].(float32)
		if !ok {
			return nil, errors.New("expect float32 param 1")
		}
		vm.RAM.IntSetDiffuseLevel(vm.Index, value)
		return vm.RAM.IntGetDiffuseLevel(vm.Index), nil
	case "vm.input.setuppercaseonly":
		if len(task.Arguments) != 1 {
			return nil, errors.New("expect 1 arguments")
		}
		value, ok := task.Arguments[0].(bool)
		if !ok {
			return nil, errors.New("expect bool param 1")
		}
		vm.RAM.IntSetUppercaseOnly(vm.Index, value)
		return vm.RAM.IntGetUppercaseOnly(vm.Index), nil
	case "vm.hardware.setserialmode":
		if len(task.Arguments) != 1 {
			return nil, errors.New("expect 1 arguments")
		}
		value, ok := task.Arguments[0].(int)
		if !ok {
			return nil, errors.New("expect int param 1")
		}
		settings.SSCCardMode[vm.Index] = settings.SSCMode(value)
		return settings.SSCCardMode[vm.Index], nil
	case "vm.hardware.setserialdipsw1":
		if len(task.Arguments) != 1 {
			return nil, errors.New("expect 1 arguments")
		}
		value, ok := task.Arguments[0].(int)
		if !ok {
			return nil, errors.New("expect int param 1")
		}
		settings.SSCDipSwitch1 = value
		return settings.SSCDipSwitch1, nil
	case "vm.hardware.setserialdipsw2":
		if len(task.Arguments) != 1 {
			return nil, errors.New("expect 1 arguments")
		}
		value, ok := task.Arguments[0].(int)
		if !ok {
			return nil, errors.New("expect int param 1")
		}
		settings.SSCDipSwitch2 = value
		return settings.SSCDipSwitch2, nil
	case "vm.music.play":
		if len(task.Arguments) != 3 {
			return nil, errors.New("expect 3 arguments")
		}
		file, ok := task.Arguments[0].(string)
		if !ok {
			return nil, errors.New("expect string param 1")
		}
		leadin, ok := task.Arguments[1].(int)
		if !ok {
			return nil, errors.New("expect int param 2")
		}
		fadein, ok := task.Arguments[2].(int)
		if !ok {
			return nil, errors.New("expect int param 3")
		}
		ent := vm.GetInterpreter()
		if strings.HasSuffix(file, ".rst") {
			ent.GetMemoryMap().IntSetRestalgiaPath(ent.GetMemIndex(), file, true)
		} else {
			go ent.PlayMusic(files.GetPath(file), files.GetFilename(file), leadin, fadein)
		}
		return file, nil
	case "vm.music.setrestalgiapath":
		if len(task.Arguments) != 2 {
			return nil, errors.New("expect 2 arguments")
		}
		file, ok := task.Arguments[0].(string)
		if !ok {
			return nil, errors.New("expect string param 1")
		}
		loop, ok := task.Arguments[1].(bool)
		if !ok {
			return nil, errors.New("expect bool param 2")
		}
		vm.RAM.IntSetRestalgiaPath(vm.Index, file, loop)
		return file, nil
	case "vm.gfx.bgcolor":
		if len(task.Arguments) != 4 {
			return nil, errors.New("expect 4 arguments")
		}
		r, ok := task.Arguments[0].(uint8)
		if !ok {
			return nil, errors.New("expect uint8 param 1")
		}
		g, ok := task.Arguments[1].(uint8)
		if !ok {
			return nil, errors.New("expect uint8 param 2")
		}
		b, ok := task.Arguments[2].(uint8)
		if !ok {
			return nil, errors.New("expect uint8 param 3")
		}
		a, ok := task.Arguments[3].(uint8)
		if !ok {
			return nil, errors.New("expect uint8 param 4")
		}
		vm.RAM.SetBGColor(vm.Index, r, g, b, a)
		return "ok", nil
	case "vm.input.setdisablemetamode":
		if len(task.Arguments) != 1 {
			return nil, errors.New("expect 1 arguments")
		}
		value, ok := task.Arguments[0].(bool)
		if !ok {
			return nil, errors.New("expect bool param 1")
		}
		settings.DisableMetaMode[vm.Index] = value
		return settings.DisableMetaMode[vm.Index], nil
	case "vm.hardware.smartportinsert":
		if len(task.Arguments) != 2 {
			return nil, errors.New("expect 2 arguments")
		}
		drive, ok := task.Arguments[0].(int)
		if !ok {
			return nil, errors.New("expect int param 1")
		}
		disk, ok := task.Arguments[1].(string)
		if !ok {
			return nil, errors.New("expect string param 2")
		}
		servicebus.SendServiceBusMessage(
			vm.Index,
			servicebus.SmartPortInsertFilename,
			servicebus.DiskTargetString{
				Drive:    drive,
				Filename: disk,
			},
		)
		return disk, nil
	case "vm.hardware.floppyinsert":
		if len(task.Arguments) != 2 {
			return nil, errors.New("expect 2 arguments")
		}
		drive, ok := task.Arguments[0].(int)
		if !ok {
			return nil, errors.New("expect int param 1")
		}
		disk, ok := task.Arguments[1].(string)
		if !ok {
			return nil, errors.New("expect string param 2")
		}
		servicebus.SendServiceBusMessage(
			vm.Index,
			servicebus.DiskIIInsertFilename,
			servicebus.DiskTargetString{
				Drive:    drive,
				Filename: disk,
			},
		)
		return disk, nil
	case "vm.hardware.start6502":
		ent := vm.GetInterpreter()
		ent.GetMemoryMap().BlockMapper[ent.GetMemIndex()].DisableBlocks([]string{"apple2iozeropage"})
		if settings.PureBootBanks != nil && len(settings.PureBootBanks) > 0 {
			ent.GetMemoryMap().BlockMapper[ent.GetMemIndex()].EnableBlocks(settings.PureBootBanks)
		}
		ent.GetMemoryMap().IntSetProcessorState(ent.GetMemIndex(), 1) // Force 65C02... we are trying to be a 2e

		// force CPU to bootstrap..
		ent.Start6502(0xfa62, 0, 0, 0, 0x80, 0x1ff)
		cpu := apple2helpers.GetCPU(ent)
		vm.cpu6502 = cpu
		cpu.Reset()
		cpu.ResetSliced()
		cpu.SetWarpUser(1.0)
		cpu.RunState = mos6502.CrsFreeRun

		settings.PureBootCheck(vm.Index)
		ent.GetMemoryMap().BlockMapper[ent.GetMemIndex()].Reset(false)

		ml := settings.MemLocks[ent.GetMemIndex()]
		if ml != nil && len(ml) > 0 {
			for addr, value := range ml {
				cpu.LockValue[addr] = value
			}
		}

		// Attach debugger here if we need it..
		//log.Printf("Debugger On = %v, settings.DebuggerAttachSlot = %d", settings.DebuggerOn, settings.DebuggerAttachSlot)
		if settings.DebuggerOn && settings.DebuggerAttachSlot-1 == vm.Index {
			url := fmt.Sprintf("http://localhost:%d/?attach=%d", settings.DebuggerPort, settings.DebuggerAttachSlot)
			err := utils.OpenURL(url)
			if err != nil {
				log.Printf("Failed to open debugger: %v", err)
			}
			cpu.RunState = mos6502.CrsPaused // start paused .. better for us
		}

		// start recording here if requested
		if settings.DiskRecordStart {
			settings.DiskRecordStart = false
			vm.GetInterpreter().RecordToggle(settings.FileFullCPURecord)
		}

		return "ok", nil

	case "vm.zxstate.restore":

		if len(task.Arguments) != 1 {
			return nil, errors.New("expect 1 arguments")
		}
		zxstate, ok := task.Arguments[0].(*snapshot.Z80)
		if !ok {
			return nil, errors.New("expect snapshot param 1")
		}

		ent := vm.GetInterpreter()
		if settings.PureBootBanks != nil && len(settings.PureBootBanks) > 0 {
			ent.GetMemoryMap().BlockMapper[ent.GetMemIndex()].EnableBlocks(settings.PureBootBanks)
		}
		//ent.GetMemoryMap().IntSetProcessorState(ent.GetMemIndex(), 1) // Force 65C02... we are trying to be a 2e

		// force CPU to bootstrap..
		cpu := apple2helpers.GetZ80CPU(ent)
		cpu.LinearMemory = true
		cpu.Reset()
		vm.cpuZ80 = cpu
		cpu.Reset()
		cpu.ResetSliced()

		settings.PureBootCheck(vm.Index)
		ent.GetMemoryMap().BlockMapper[ent.GetMemIndex()].Reset(false)
		ent.SetState(types.EXECZ80)

		spec, ok := vm.RAM.InterpreterMappableAtAddress(vm.Index, 0xff00)
		if ok {
			s, ok := spec.(*spectrum.ZXSpectrum)
			if ok {
				s.LoadSnapshot(zxstate)
			}
		}

		log.Printf("Starting Z80 CPU...")

		return "ok", nil

	case "vm.hardware.startz80":

		ent := vm.GetInterpreter()
		if settings.PureBootBanks != nil && len(settings.PureBootBanks) > 0 {
			ent.GetMemoryMap().BlockMapper[ent.GetMemIndex()].EnableBlocks(settings.PureBootBanks)
		}
		//ent.GetMemoryMap().IntSetProcessorState(ent.GetMemIndex(), 1) // Force 65C02... we are trying to be a 2e

		// force CPU to bootstrap..
		cpu := apple2helpers.GetZ80CPU(ent)
		cpu.LinearMemory = true
		cpu.Reset()
		vm.cpuZ80 = cpu
		cpu.Reset()
		cpu.ResetSliced()
		//	cpu.SetWarpUser(1.0)
		//	cpu.RunState = mos6502.CrsFreeRun

		settings.PureBootCheck(vm.Index)
		ent.GetMemoryMap().BlockMapper[ent.GetMemIndex()].Reset(false)
		ent.SetState(types.EXECZ80)

		log.Printf("Starting Z80 CPU...")

		return "ok", nil

	case "vm.cpu.resume":
		ent := vm.GetInterpreter()
		slotid := vm.Index
		settings.SlotRestartContinueCPU[slotid] = false
		cpu := apple2helpers.GetCPU(ent)
		cpu.Halted = false
		if !cpu.Halted {
			// start things again recording live mode
			if settings.AutoLiveRecording() && slotid == 0 {
				ent.ResumeRecording(settings.VideoRecordFile[slotid], settings.VideoRecordFrames[slotid], false)
			}
			settings.VideoRecordFrames[slotid] = nil
			ent.SetState(types.EXEC6502)
		}
		return "resume", nil
	case "vm.recording.play":
		if len(task.Arguments) != 2 {
			return nil, errors.New("expect 2 arguments")
		}
		data, ok := task.Arguments[0].([]*bytes.Buffer)
		if !ok {
			return nil, errors.New("expect []*bytes.Buffer param 1")
		}
		backwards, ok := task.Arguments[1].(bool)
		if !ok {
			return nil, errors.New("expect bool param 2")
		}
		ent := vm.GetInterpreter()
		bus.StartDefault()
		ent.PlayBlocks(data, backwards, 0)
		return "playing", nil
	case "vm.freeze.restore":
		if len(task.Arguments) != 1 {
			return nil, errors.New("expect 1 arguments")
		}
		freeze, ok := task.Arguments[0].(*freeze.FreezeState)
		if !ok {
			return nil, errors.New("expect []byte param 1")
		}
		ent := vm.GetInterpreter()
		slotid := vm.Index
		settings.SetPureBoot(slotid, true)
		settings.SlotZPEmu[slotid] = false
		freeze.Apply(ent)
		// start things again recording live mode
		if settings.AutoLiveRecording() && slotid == 0 {
			ent.ResumeRecording(settings.VideoRecordFile[slotid], settings.VideoRecordFrames[slotid], false)
		}
		settings.VideoRecordFrames[slotid] = nil

		//cpu := apple2helpers.GetCPU(ent)
		ent.SetState(types.EXEC6502)

		// Attach debugger here if we need it..
		if settings.DebuggerOn && settings.DebuggerAttachSlot-1 == slotid {
			cpu := apple2helpers.GetCPU(ent)
			url := fmt.Sprintf("http://localhost:%d/?attach=%d", settings.DebuggerPort, settings.DebuggerAttachSlot)
			err := utils.OpenURL(url)
			if err != nil {
				log.Printf("Failed to open debugger: %v", err)
			}
			cpu.RunState = mos6502.CrsPaused // start paused .. better for us
		}
		return "restore", nil
	case "vm.files.setworkdir":
		if len(task.Arguments) != 1 {
			return nil, errors.New("expect 1 arguments")
		}
		value, ok := task.Arguments[0].(string)
		if !ok {
			return nil, errors.New("expect string param 1")
		}
		vm.GetInterpreter().SetWorkDir(value)
		return vm.GetInterpreter().GetWorkDir(), nil
	case "vm.input.setmousemode":
		if len(task.Arguments) != 1 {
			return nil, errors.New("expect 1 arguments")
		}
		value, ok := task.Arguments[0].(settings.MouseMode)
		if !ok {
			return nil, errors.New("expect mouse mode param 1")
		}
		settings.SetMouseMode(value)
		return "ok", nil
	case "vm.input.setjoystickaxis0":
		if len(task.Arguments) != 1 {
			return nil, errors.New("expect 1 arguments")
		}
		value, ok := task.Arguments[0].(int)
		if !ok {
			return nil, errors.New("expect int param 1")
		}
		vm.RAM.PaddleMap[vm.Index][0] = int(value)
		return vm.RAM.PaddleMap[vm.Index][0], nil
	case "vm.input.setjoystickaxis1":
		if len(task.Arguments) != 1 {
			return nil, errors.New("expect 1 arguments")
		}
		value, ok := task.Arguments[0].(int)
		if !ok {
			return nil, errors.New("expect int param 1")
		}
		vm.RAM.PaddleMap[vm.Index][1] = int(value)
		return vm.RAM.PaddleMap[vm.Index][1], nil
	case "vm.hardware.setcpuwarp":
		if len(task.Arguments) != 1 {
			return nil, errors.New("expect 1 arguments")
		}
		value, ok := task.Arguments[0].(float64)
		if !ok {
			return nil, errors.New("expect float64 param 1")
		}
		//log.Printf("Applying cpu speed multiplier %f to slot %d", value, vm.Index)
		cpu := apple2helpers.GetCPU(vm.GetInterpreter())
		cpu.SetWarpUser(value)
		return value, nil
	case "vm.hardware.setdisablewarp":
		if len(task.Arguments) != 1 {
			return nil, errors.New("expect 1 arguments")
		}
		value, ok := task.Arguments[0].(bool)
		if !ok {
			return nil, errors.New("expect bool param 1")
		}
		settings.NoDiskWarp[vm.Index] = value
		return settings.NoDiskWarp[vm.Index], nil
	case "vm.hardware.setpreservedsk":
		if len(task.Arguments) != 1 {
			return nil, errors.New("expect 1 arguments")
		}
		value, ok := task.Arguments[0].(bool)
		if !ok {
			return nil, errors.New("expect bool param 1")
		}
		settings.PreserveDSK = value
		return settings.PreserveDSK, nil
	case "vm.hardware.setliverecording":
		if len(task.Arguments) != 1 {
			return nil, errors.New("expect 1 arguments")
		}
		value, ok := task.Arguments[0].(bool)
		if !ok {
			return nil, errors.New("expect bool param 1")
		}
		settings.SetAutoLiveRecording(value)
		return settings.AutoLiveRecording(), nil
	case "vm.hardware.setdisablefractionalrewind":
		if len(task.Arguments) != 1 {
			return nil, errors.New("expect 1 arguments")
		}
		value, ok := task.Arguments[0].(bool)
		if !ok {
			return nil, errors.New("expect bool param 1")
		}
		settings.DisableFractionalRewindSpeeds = value
		return settings.DisableFractionalRewindSpeeds, nil
	case "vm.hardware.setpdftimeoutsec":
		if len(task.Arguments) != 1 {
			return nil, errors.New("expect 1 arguments")
		}
		value, ok := task.Arguments[0].(int)
		if !ok {
			return nil, errors.New("expect int param 1")
		}
		settings.PrintToPDFTimeoutSec = value
		return settings.PrintToPDFTimeoutSec, nil
	case "vm.presentation.apply":
		if len(task.Arguments) != 1 {
			return nil, errors.New("expect 1 arguments")
		}
		value, ok := task.Arguments[0].(*presentation.Presentation)
		if !ok {
			return nil, errors.New("expect presentation.Presentation param 1")
		}
		ent := vm.GetInterpreter()
		value.Apply("init", ent)
		return "ok", nil
	}

	return nil, errors.New("unrecognised action: " + task.Action)
}

func (vm *VM) PreInit() {
	var err error

	settings.UserWarpOverride[vm.Index] = false // warp mode off
	settings.UseVerticalBlend[vm.Index] = false
	settings.UseDHGRForHGR[vm.Index] = false
	settings.LogoCameraControl[vm.Index] = false
	settings.DisableJoystick[vm.Index] = false
	settings.ImageDrawRect[vm.Index] = nil
	settings.HeatMap[vm.Index] = settings.HMOff
	settings.HeatMapBank[vm.Index] = 0

	settings.AutosaveFilename[vm.Index] = ""

	settings.DisableTextSelect[vm.Index] = false

	settings.SpritesEnabled[vm.Index] = false

	settings.InitSlotZones(vm.Index)

	settings.SpeakerRedirects[vm.Index] = nil

	vm.RAM.SetCameraConfigure(vm.Index, 0)

	_, err = vm.ExecuteRequest("vm.audio.setspeakervolume", 0.5)
	if err != nil {
		return
	}
	_, err = vm.ExecuteRequest("vm.hud.loadfont", "fonts/appleiifont.yaml")
	if err != nil {
		return
	}
	_, err = vm.ExecuteRequest("vm.memory.set", 50, 255)
	if err != nil {
		return
	}
	_, err = vm.ExecuteRequest("vm.memory.set", 51, 255)
	if err != nil {
		return
	}
	_, err = vm.ExecuteRequest("vm.hud.select", "TEXT")
	if err != nil {
		return
	}
	_, err = vm.ExecuteRequest("vm.hud.mode", "mode40.preserve")
	if err != nil {
		return
	}
	_, err = vm.ExecuteRequest("vm.hud.fgcolor", 15)
	if err != nil {
		return
	}
	_, err = vm.ExecuteRequest("vm.hud.bgcolor", 0)
	if err != nil {
		return
	}
	_, err = vm.ExecuteRequest("vm.hud.clearscreen")
	if err != nil {
		return
	}

	_, err = vm.ExecuteRequest("vm.hud.select", "TXT2")
	if err != nil {
		return
	}
	_, err = vm.ExecuteRequest("vm.hud.fgcolor", 15)
	if err != nil {
		return
	}
	_, err = vm.ExecuteRequest("vm.hud.bgcolor", 0)
	if err != nil {
		return
	}
	_, err = vm.ExecuteRequest("vm.hud.clearscreen")
	if err != nil {
		return
	}
	_, err = vm.ExecuteRequest("vm.hud.setsubformat", types.LSF_FIXED_40_24)
	if err != nil {
		return
	}

	_, err = vm.ExecuteRequest("vm.hud.select", "TEXT")
	if err != nil {
		return
	}

	for i := 0; i < 8; i++ {
		_, err = vm.ExecuteRequest("vm.gfx.camera.reset", i)
		if err != nil {
			return
		}
	}

	_, err = vm.ExecuteRequest("vm.gfx.setbackdrop", "", 7, float32(1), float32(16), float32(0), false)
	if err != nil {
		return
	}
	_, err = vm.ExecuteRequest("vm.gfx.setbackdroppos", float64(types.CWIDTH/2), float64(types.CHEIGHT/2), float64(-types.CWIDTH/2))
	if err != nil {
		return
	}
	_, err = vm.ExecuteRequest("vm.gfx.setoverlay", "")
	if err != nil {
		return
	}

	_, err = vm.ExecuteRequest("vm.music.stop")
	if err != nil {
		return
	}

	_, err = vm.ExecuteRequest("vm.palette.reload")
	if err != nil {
		return
	}

	for i := 0; i < 4; i++ {
		_, err = vm.ExecuteRequest("vm.input.setpaddlevalue", i, 127)
		if err != nil {
			return
		}
	}

	for i := 0; i < 3; i++ {
		_, err = vm.ExecuteRequest("vm.input.setpaddlebutton", i, 0)
		if err != nil {
			return
		}
	}

	// Settings seeding prior state to force change
	slotid := vm.Index
	settings.LastRenderModeSpectrum[slotid] = settings.VM_DOTTY
	settings.LastRenderModeSHR[slotid] = settings.VM_DOTTY
	settings.LastRenderModeDHGR[slotid] = settings.VM_DOTTY
	settings.LastRenderModeHGR[slotid] = settings.VM_DOTTY
	settings.LastRenderModeGR[slotid] = settings.VM_FLAT
	settings.LastTintMode[slotid] = settings.VPT_GREY
	settings.LastVoxelDepth[slotid] = settings.VXD_9_TIMES

	vm.RAM.WriteInterpreterMemorySilent(vm.Index, memory.MICROM8_SPRITE_CONTROL_BASE+0, 0)
	vm.RAM.WriteInterpreterMemorySilent(vm.Index, memory.MICROM8_SPRITE_CONTROL_BASE+1, 0)

	_, err = vm.ExecuteRequest("vm.gfx.setrendermode", "DHGR", settings.DefaultRenderModeDHGR)
	if err != nil {
		return
	}
	_, err = vm.ExecuteRequest("vm.gfx.setrendermode", "HGR", settings.DefaultRenderModeHGR)
	if err != nil {
		return
	}
	_, err = vm.ExecuteRequest("vm.gfx.setrendermode", "GR", settings.DefaultRenderModeGR)
	if err != nil {
		return
	}
	_, err = vm.ExecuteRequest("vm.gfx.setrendermode", "SHR", settings.DefaultRenderModeSHR)
	if err != nil {
		return
	}
	_, err = vm.ExecuteRequest("vm.gfx.setrendermode", "ZX", settings.DefaultRenderModeSpectrum)
	if err != nil {
		return
	}
	_, err = vm.ExecuteRequest("vm.gfx.settintmode", settings.DefaultTintMode)
	if err != nil {
		return
	}
	_, err = vm.ExecuteRequest("vm.gfx.setvoxeldepth", settings.VXD_2_TIMES)
	if err != nil {
		return
	}

	_, err = vm.ExecuteRequest("vm.gfx.setambientlevel", float32(1))
	if err != nil {
		return
	}
	_, err = vm.ExecuteRequest("vm.gfx.setdiffuselevel", float32(1))
	if err != nil {
		return
	}
	_, err = vm.ExecuteRequest("vm.input.setuppercaseonly", false)
	if err != nil {
		return
	}
	_, err = vm.ExecuteRequest("vm.music.setrestalgiapath", "", false)
	if err != nil {
		return
	}
	_, err = vm.ExecuteRequest("vm.gfx.bgcolor", uint8(0), uint8(0), uint8(0), uint8(0))
	if err != nil {
		return
	}

	_, err = vm.ExecuteRequest("vm.input.setdisablemetamode", false)
	if err != nil {
		return
	}

	vm.GetInterpreter().PositionAllLayers(0, 0, 0)

	vm.RAM.IntEnableKeyRedirect(vm.Index, -1, false)

	vm.PreInitOverrides()

}
