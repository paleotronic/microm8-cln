package core

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	tomb "gopkg.in/tomb.v2"
	s8webclient "paleotronic.com/api"
	"paleotronic.com/core/editor"
	"paleotronic.com/core/hardware/apple2"
	"paleotronic.com/core/hardware/apple2/woz2"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/hardware/control"
	"paleotronic.com/core/hardware/cpu"
	"paleotronic.com/core/hardware/cpu/mos6502"
	"paleotronic.com/core/hardware/servicebus"
	"paleotronic.com/core/hardware/spectrum"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/core/vduconst"
	"paleotronic.com/disk"
	"paleotronic.com/filerecord"
	"paleotronic.com/files"
	"paleotronic.com/fmt"
	"paleotronic.com/freeze"
	"paleotronic.com/log"
	"paleotronic.com/octalyzer/bus"
	"paleotronic.com/octalyzer/video/font"
	"paleotronic.com/panic"
	"paleotronic.com/runestring"
	"paleotronic.com/utils"

	log2 "log"
)

var interp *Producer

type Producer struct {
	Context      int
	InputContext int
	//Global  types.VarMap
	Command                      runestring.RuneString
	History                      []runestring.RuneString
	HistIndex, InsertPos, Px, Py int
	TimeBetweenInstructionsNS    int64
	LastExecute                  int64
	BootTime                     int64
	//DiaINT                       *appleinteger.DialectAppleInteger
	NeedsPrompt  bool
	RebootNeeded bool
	//DiaAS                        *applesoft.DialectApplesoft
	//DiaShell                     *shell.DialectShell
	AddressSpace *memory.MemoryMap
	VM           [memory.OCTALYZER_NUM_INTERPRETERS]*VM
	ThreadActive [memory.OCTALYZER_NUM_INTERPRETERS]uint64
	//NumInterpreters              int
	PostCallback     [memory.OCTALYZER_NUM_INTERPRETERS]func(index int)
	MasterLayerPos   []types.LayerPosMod
	ForceUpdate      []bool
	EndRemotesNeeded bool
	Paused           bool
	Quit             [memory.OCTALYZER_NUM_INTERPRETERS]chan bool
	PStates          [memory.OCTALYZER_NUM_INTERPRETERS]interfaces.Presentable
	LastPStateFile   [memory.OCTALYZER_NUM_INTERPRETERS]string

	// control programs
	pendingControlPath     string
	pendingControlPrograms []string
}

func (this *Producer) StartPendingControlPrograms() {
	if len(this.pendingControlPrograms) == 0 {
		return
	}
	// spin up controls here
	//log.Printf("Starting %d control programs...", len(this.pendingControlPrograms))

	for i, control := range this.pendingControlPrograms {
		if files.ExistsViaProvider(this.pendingControlPath, control+".apl") {
			//log.Printf("... starting %s.apl", control)
			// load control program
			//go func(i int, ent interfaces.Interpretable, control string) {
			//time.Sleep(3 * time.Second)
			myslot := 0
			cslot := myslot + 1 + i
			if cslot >= memory.OCTALYZER_NUM_INTERPRETERS {
				//log.Println("Failed to find free slot for control program!")
				this.pendingControlPrograms = []string{}
				this.pendingControlPath = ""
				return
			}
			//log.Printf("Chose slot %d for control program (control slot %d)\n", cslot, myslot)
			this.Activate(cslot)
			cent := this.GetInterpreter(cslot)
			for cent.GetChild() != nil {
				cent = cent.GetChild()
			}

			apple2helpers.CommandResetOverride(
				cent,
				"fp",
				"/"+strings.Trim(this.pendingControlPath, "/")+"/",
				"run \""+control+".apl\"",
			)
			settings.ResetState[cent.GetMemIndex()].IsControl = true

			//cent.SetMicroControl(true)
		}
	}
	this.pendingControlPrograms = []string{}
	this.pendingControlPath = ""
}

func (this *Producer) ClearPState(index int) {
	this.PStates[index] = nil
}

func (this *Producer) SetPState(index int, p interfaces.Presentable, source string) {
	this.PStates[index] = p
	this.LastPStateFile[index] = source
}

func (this *Producer) SetPresentationSource(index int, source string) {
	this.LastPStateFile[index] = source
}

func (this *Producer) GetPresentationSource(index int) string {
	return this.LastPStateFile[index]
}

func (this *Producer) GetPState(index int) interfaces.Presentable {
	return this.PStates[index]
}

var um sync.Mutex
var uuid uint64 = 0x2a82

func GetUUID() uint64 {
	um.Lock()
	defer um.Unlock()
	uuid++
	return uuid
}

func SetInstance(i *Producer) {
	interp = i
}

func GetInstance() interfaces.Producable {
	return interp
}

func (this *Producer) GetContext() int {
	return this.Context
}

func (this *Producer) GetInterpreterList() [memory.OCTALYZER_NUM_INTERPRETERS]interfaces.Interpretable {
	var out [memory.OCTALYZER_NUM_INTERPRETERS]interfaces.Interpretable
	for i, vm := range this.VM {
		if vm == nil {
			continue
		}
		if vm.GetInterpreter() != nil {
			out[i] = vm.GetInterpreter()
		}
	}
	return out
}

func (this *Producer) GetInterpreter(slot int) interfaces.Interpretable {
	if slot < 0 {
		slot = 0
	} else if slot >= memory.OCTALYZER_NUM_INTERPRETERS {
		slot = memory.OCTALYZER_NUM_INTERPRETERS - 1
	}
	if this.VM[slot] == nil {
		this.CreateVM(slot, nil, nil)
		//this.CreateInterpreterInSlot(slot, "fp", false)
	}
	e := this.VM[slot].GetInterpreter()
	return e
}

func (this *Producer) Stop() {

	for _, vm := range this.VM {
		if vm != nil {
			vm.Stop()
		}
	}

}

func (this *Producer) AmIActive(ent interfaces.Interpretable) bool {
	return ent.GetUUID() == this.ThreadActive[ent.GetMemIndex()]
}

func (this *Producer) StopMicroControls() {

	for _, vm := range this.VM {
		if vm == nil {
			continue
		}
		ent := vm.GetInterpreter()
		if ent != nil && ent.IsMicroControl() {
			vm.Stop()
		}
	}

}

func (this *Producer) PauseMicroControls() {

	for _, vm := range this.VM {
		if vm == nil {
			continue
		}
		ent := vm.GetInterpreter()
		if ent != nil && (ent.IsMicroControl() || ent.GetMemIndex() == 1) {
			fmt.Printf("Stop slot %d\n", ent.GetMemIndex())
			ent.StopTheWorld()
		}
	}

}

func (this *Producer) ResumeMicroControls() {

	for _, vm := range this.VM {

		if vm == nil {
			continue
		}

		ent := vm.GetInterpreter()

		if ent != nil && (ent.IsMicroControl() || ent.GetMemIndex() == 1) {
			fmt.Printf("Resume slot %d\n", ent.GetMemIndex())
			ent.ResumeTheWorld()
		}
	}

}

func (this *Producer) StartMicroControls() {

	for i, vm := range this.VM {

		if vm == nil {
			continue
		}

		ent := vm.GetInterpreter()
		if ent != nil && ent.IsMicroControl() {
			fmt.Printf(">>> Resuming control program in slot %d\n", i)
			ent.SetDisabled(false)
			ent.ResumeTheWorld()
		}
	}

}

func (this *Producer) HasRunningInterpreters() bool {

	result := false

	for _, vm := range this.VM {

		if vm == nil {
			continue
		}

		ent := vm.GetInterpreter()
		if ent == nil {
			continue
		}

		// If an entity is childless && running - TRUE
		// If an entity has a child and it is running - TRUE
		if ent.GetChild() == nil {
			if ent.IsRunning() || ent.IsRunningDirect() {
				result = true
			}
		} else {
			if ent.GetChild().IsRunning() || ent.GetChild().IsRunningDirect() {
				result = true
			}
		}
	}

	/* enforce non void return */
	return result

}

var doers map[int]interfaces.Interpretable

func (this *Producer) GetMemoryCallback(index int) func(index int) {
	return this.PostCallback[index]
}

func (this *Producer) DropInterpreter(slotid int) {
	if slotid == 0 || slotid >= len(this.VM) {
		return
	}
	this.VM[slotid].Stop()
}

func (this *Producer) EndRemotes() {

	this.SetContext(0)
	this.SetInputContext(0)

	for i, vm := range this.VM {
		if vm == nil {
			continue
		}

		// Delegate to nested subentity
		ent := vm.GetInterpreter()
		if ent == nil {
			continue
		}

		if ent.IsRemote() {
			ent.EndRemote()
		}

		this.MasterLayerPos[i] = types.LayerPosMod{0, 0}
	}

	//~ this.VM = this.VM[:1]
	//~ this.NumInterpreters = 1

	for i := 1; i < 8; i++ {
		this.AddressSpace.IntSetActiveState(i, 0)
	}

	e := this.VM[0].GetInterpreter()

	if e != nil {
		e.EndProgram()
		apple2helpers.TEXT40(e)
		apple2helpers.Clearscreen(e)
	}
}

func (this *Producer) Post(index int) {
	if this.PostCallback[index] != nil {
		this.PostCallback[index](index)
	}
}

func (this *Producer) SetContext(v int) error {

	/* vars */
	v = v % memory.OCTALYZER_NUM_INTERPRETERS
	ent := this.VM[v]

	if ent == nil {
		v = 0
		ent = this.VM[v]
		//return errors.New("Attempt to switch to a non-existent interpreter")
	}

	this.Context = v

	return nil
}

func (this *Producer) SetInputContext(v int) error {

	if v == -1 {
		this.AddressSpace.InputToggle(v)
		return nil
	}

	/* vars */
	v = v % memory.OCTALYZER_NUM_INTERPRETERS
	ent := this.VM[v]

	if ent == nil {
		v = 0
		ent = this.VM[v]
		//return errors.New("Attempt to switch to a non-existent interpreter")
	}

	this.InputContext = v

	//fmt.Printf("---> Updating input context to %d\n", this.InputContext)

	// switch input too
	this.AddressSpace.InputToggle(this.InputContext)

	return nil
}

func (this *Producer) Parse(s string) {

	/* vars */
	ent := this.VM[this.Context].GetInterpreter()

	if ent != nil {
		log.Printf("Delegating [%s] to %s\n", s, ent.GetName())
		ent.Parse(s)
	} else {
		log.Printf("Invalid interpreter context")
	}

}

func (this *Producer) Broadcast(msg types.InfernalMessage) {

	/* vars */
}

func (this *Producer) SetNeedsPrompt(v bool) {
	this.NeedsPrompt = v
}

func (this *Producer) DropVM(slot int) {
	this.VM[slot] = nil
}

func (this *Producer) CreateVM(slot int, config *VMLauncherConfig, t *tomb.Tomb) error {

	// log2.Printf("===========================================================> In create VM for slot %d", slot)

	// Block system from three finger salute whilst building VM
	settings.BlockCSR[slot] = true
	defer func() {
		settings.BlockCSR[slot] = false
	}()

	if this.VM[slot] != nil {
		// log2.Printf("calling stop for slot... %d", slot)
		this.VM[slot].Stop()
	}

	this.AddressSpace.Zero(slot)

	settings.FirstBoot[slot] = true

	this.AddressSpace.IntSetActiveState(slot, 0)
	this.AddressSpace.IntSetLayerState(slot, 0)

	settings.VideoSuspended = true
	defer func() {
		settings.VideoSuspended = false
	}()

	specfile := settings.SpecFile[slot]
	if config != nil && config.Pakfile != nil {
		// extract machine configuration preference from pakfile
		specname := config.Pakfile.GetProfile("apple2e-en")
		if specname != "" {
			if !strings.HasSuffix(specname, ".yaml") {
				specfile = specname + ".yaml"
			}
		}
	}

	// if this.VM[0] != nil && slot > 0 {
	// 	log2.Printf("In create VM slot %d, prod slot 0 = %v", slot, this.VM[0].p)
	// }

	vm, err := NewVM(slot, this.AddressSpace, this, specfile, 0)
	if err != nil {
		return err
	}
	vm.p = this
	this.VM[slot] = vm
	this.VM[slot].Start(t)
	this.PreInit(slot)

	// log2.Printf("have called preinit for slot %d, slot 0 prod = %v", slot, this.VM[0].p)

	p, err := files.LoadDefaultState(vm.e)
	if err == nil && p != nil && this.GetPState(slot) == nil {
		p.Apply("init", vm.e)
	}

	// log2.Printf("have have applied state for slot %d, slot 0 prod = %v", slot, this.VM[0].p)

	return vm.ApplyLaunchConfig(config)
}

// func (this *Producer) CreateInterpreter(slot int, name string, dia interfaces.Dialecter, spec string, uuid uint64) (interfaces.Interpretable, error) {

// 	/* vars */
// 	var result *interpreter.Interpreter

// 	if slot >= memory.OCTALYZER_NUM_INTERPRETERS || slot < 0 {
// 		return result, errors.New("Bad slot!")
// 	}

// 	//fmt.Println("pre CreateInterpreter()")
// 	result = interpreter.NewInterpreter(name, dia, nil, this.AddressSpace, slot, spec, nil)
// 	result.SetUUID(uuid)
// 	//fmt.Println("post CreateInterpreter()")

// 	//this.VM[slot].SetInterpreter(result)
// 	this.MasterLayerPos[slot] = types.LayerPosMod{0, 0}
// 	this.Context = slot
// 	result.SetProducer(this)

// 	//fmt.Println("Status = ", result.GetState())

// 	/* enforce non void return */
// 	return result, nil

// }

func (this *Producer) StopInterpreter(slotid int) {
	this.ThreadActive[slotid] = 0

	this.AddressSpace.IntSetCPUHalt(slotid, true)
	//settings.CPUForceHalt[slotid] = true
	time.Sleep(1 * time.Millisecond)
}

func (this *Producer) RestartInterpreter(slotid int) {

	fmt.Printf("########### SLOT %d RESTART IN PROGRESS ###########\n", slotid)
	//this.StopInterpreter(slotid)

	ent := this.VM[slotid].GetInterpreter()

	if settings.PureBootRestoreStateBin[slotid] == nil {
		this.AddressSpace.Zero(slotid)

		if !settings.SlotRestartContinueCPU[slotid] {
			//apple2helpers.TrashCPU(this.VM[slotid])
			ent.SetSpec("")
			ent.LoadSpec(settings.SpecFile[slotid])
			cpu := apple2helpers.GetCPU(ent)
			cpu.Reset()
			cpu.ResetSliced()
			ent.GetMemoryMap().BlockMapper[ent.GetMemIndex()].Reset(false)
			ent.GetMemoryMap().IntSetLED0(ent.GetMemIndex(), 0)
			ent.GetMemoryMap().IntSetLED1(ent.GetMemIndex(), 0)

			//for i := 0; i < 131072; i++ {
			//	ent.SetMemorySilent(i, 0)
			//}
			apple2helpers.MonitorPanel(ent, false)
			apple2helpers.TextHideCursor(ent)
			apple2helpers.SetFGColor(ent, 15)
			apple2helpers.SetBGColor(ent, 0)
			apple2helpers.Clearscreen(ent)
		}
	}

	//ent, _ := this.CreateInterpreter(slotid, "main", this.DiaAS, settings.DefaultProfile)

	this.AddressSpace.IntSetPaddleValue(slotid, 0, 127)
	this.AddressSpace.IntSetPaddleValue(slotid, 1, 127)
	this.AddressSpace.IntSetPaddleValue(slotid, 2, 127)
	this.AddressSpace.IntSetPaddleValue(slotid, 3, 127)

	this.VM[slotid] = ent.VM().(*VM)
	ent.SetDisabled(false)
	this.AddressSpace.IntSetActiveState(slotid, 1)
	this.AddressSpace.IntSetLayerState(slotid, 1)
	ent.ResumeTheWorld()
	settings.FrameSkip = settings.DefaultFrameSkip
	//go this.Executor(slotid)
	fmt.Printf("########### SLOT %d RESTART IS COMPLETE ###########\n", slotid)

	fmt.Printf("State before = %s\n", ent.GetState())
	this.ExecutorInit(slotid)
	fmt.Printf("State after = %s\n", ent.GetState())

	if slotid == 0 && !settings.PureBoot(slotid) && !settings.Offline && ent.GetState() != types.DIRECTRUNNING {
		apple2helpers.MODE40(ent)
		ent.Bootstrap("fp", true)
		if !settings.IsRemInt {
			settings.PureBootVolume[slotid] = settings.SplashDisk
			settings.PureBootVolume2[slotid] = settings.SplashDisk2
			settings.SplashDisk = ""
			if settings.TrackerMode {
				ent.Parse("tracker")
			} else {
				ent.Parse("run \"/boot/boot\"")
			}
			fmt.Println("STARTING MENU", ent.GetState())
		}
	}

}

func (this *Producer) Reset() {

	if !settings.CleanBootRequested {
		return
	}

	settings.CleanBootRequested = false
	settings.BlueScreen = true

	for i := 0; i < settings.NUMSLOTS; i++ {
		servicebus.SendServiceBusMessage(
			i,
			servicebus.DiskIIEject,
			0,
		)
		servicebus.SendServiceBusMessage(
			i,
			servicebus.DiskIIEject,
			1,
		)
		servicebus.SendServiceBusMessage(
			i,
			servicebus.SmartPortEject,
			0,
		)
		settings.PureBootSmartVolume[i] = ""
		settings.PureBootVolume[i] = ""
		settings.PureBootVolume2[i] = ""
	}

	this.Stop()
	//this.StopMicroControls()

	// for i, _ := range this.ThreadActive {
	// 	this.ThreadActive[i] = 0
	// }

	settings.Offline = false
	settings.LocalBoot = false
	settings.FirstBoot[0] = true

	// for i, _ := range this.VM {
	// 	this.VM[i] = nil
	// 	this.AddressSpace.Zero(i)
	// }

	time.Sleep(500 * time.Millisecond)

	//bootstrap := "run \"/boot/boot\""

	count := 1
	if settings.IsRemInt {
		count = memory.OCTALYZER_NUM_INTERPRETERS
	}

	for i := 0; i < count; i++ {

		this.Quit[i] = make(chan bool)

		this.CreateVM(i, nil, nil)
		this.ThreadActive[i] = this.VM[i].GetInterpreter().GetUUID()
		//log.Printf("Set threadactive[%d] to %d", i, this.ThreadActive[i])

		if settings.IsRemInt || i == 0 {
			z := i
			apple2helpers.MODE40(this.VM[z].GetInterpreter())
			apple2helpers.Clearscreen(this.VM[z].GetInterpreter())
			go this.Executor(z)
			time.Sleep(50 * time.Millisecond)
			this.VM[z].GetInterpreter().SetDisabled(false)
			this.VM[z].GetInterpreter().ResumeTheWorld()
		}

	}

	this.Select(0)
	settings.BlueScreen = false

	//fmt.Println("Just after create interpreter")
	this.LastExecute = time.Now().Unix()
	this.History = make([]runestring.RuneString, 0)

	bus.StartDefault()
}

func (this *Producer) BootStrap(path string) error {
	fmt.Printf("BOOTSTRAP for %s\n", path)

	ent := this.GetInterpreter(0)
	fp, err := files.ReadBytesViaProvider(files.GetPath(path), files.GetFilename(path))
	if err != nil {
		return err
	}
	zp, err := files.NewOctContainer(&fp, path)
	if err != nil {
		return err
	}

	settings.PureBootSmartVolume[0] = ""

	cfiles := zp.GetControlFiles()
	cFilesValid := false
	for _, control := range cfiles {
		if zp.Exists("", control+".apl") {
			cFilesValid = true
		}
	}

	this.Stop()
	this.StopMicroControls()

	cpu := apple2helpers.GetCPU(ent)
	cpu.Halted = true
	cpu.ResetSliced()

	if ent.GetMemIndex() != 0 && cFilesValid {
		ent.GetMemoryMap().MetaKeySet(ent.GetMemIndex(), vduconst.SCTRL1, ' ')
		ent = this.GetInterpreter(0)
	}

	startfile := zp.GetStartup()

	fmt.Printf("Startup file is %s\n", startfile)

	if startfile == "" {
		return errors.New("Invalid MicroPak")
	}

	fmt.Printf("Starting %s\n", startfile)
	if startfile != "" {

		profile := zp.GetProfile("apple2e-en")
		settings.SpecName[ent.GetMemIndex()] = profile
		settings.SpecFile[ent.GetMemIndex()] = profile + ".yaml"

		backdropfile, opacity, zoomratio, zoom, camtrack := zp.GetBackdrop()
		if backdropfile != "" && !strings.HasPrefix(backdropfile, "/") {
			backdropfile = strings.Trim(path, "/") + "/" + backdropfile
		}

		musicfile, leadin, fadein := zp.GetMusicTrack()
		if musicfile != "" && !strings.HasPrefix(musicfile, "/") {
			musicfile = strings.Trim(path, "/") + "/" + musicfile
		}

		ext := files.GetExt(startfile)
		if files.IsBootable(ext) {

			ent.SetWorkDir(path)
			settings.PureBootVolume[ent.GetMemIndex()] = path + "/" + startfile

			if ad := zp.GetAux(); ad != "" {
				settings.PureBootVolume2[ent.GetMemIndex()] = path + "/" + ad
			}

			ent.GetMemoryMap().IntSetSlotRestart(ent.GetMemIndex(), true)

		} else if ok, dialect, command := files.IsLaunchable(ext); ok {

			apple2helpers.MonitorPanel(ent, false)

			ent.Halt()
			ent.Halt6502(0)
			ent.Bootstrap(dialect, true)
			ent.SetWorkDir(path)

			command = strings.Replace(command, "%f", startfile, -1)
			fmt.Printf("Launching command: %s\n", command)

			s := &settings.RState{
				Dialect: dialect,
				Command: command,
				WorkDir: path,
			}
			settings.ResetState[ent.GetMemIndex()] = s
			ent.GetMemoryMap().IntSetSlotRestart(ent.GetMemIndex(), true)
			fmt.Printf("Requesting controlled restart of slot %d\n", ent.GetMemIndex())

		} else if ok, dialect := files.IsRunnable(ext); ok {

			apple2helpers.MonitorPanel(ent, false)

			apple2helpers.CommandResetOverride(
				ent,
				dialect,
				"/"+strings.Trim(path, "/")+"/",
				"RUN "+startfile[0:len(startfile)-len(ext)-1],
			)

		} else if files.IsBinary(ext) {

			apple2helpers.CommandResetOverride(
				ent,
				"fp",
				"/"+strings.Trim(path, "/")+"/",
				"PRINT CHR$(4);\"BRUN "+startfile[0:len(startfile)-len(ext)-1]+"\"",
			)

		}

		p, e := files.OpenPresentationState(path)
		if e == nil {
			this.SetPState(ent.GetMemIndex(), p, path)
		}

		//cfiles := []string{"control", "control2", "control3", "control4"}

		//this.StopMicroControls()

		this.pendingControlPrograms = cfiles
		this.pendingControlPath = path

		if musicfile != "" {
			//ent.PlayMusic(files.GetPath(musicfile), files.GetFilename(musicfile), leadin, fadein)
			settings.MusicLeadin[ent.GetMemIndex()] = leadin
			settings.MusicFadein[ent.GetMemIndex()] = fadein
			settings.MusicTrack[ent.GetMemIndex()] = musicfile
		}

		if backdropfile != "" {
			settings.BackdropFile[ent.GetMemIndex()] = backdropfile
			settings.BackdropOpacity[ent.GetMemIndex()] = opacity
			settings.BackdropZoom[ent.GetMemIndex()] = zoom
			settings.BackdropZoomRatio[ent.GetMemIndex()] = zoomratio
			settings.BackdropTrack[ent.GetMemIndex()] = camtrack
		}

	}

	//settings.FirstBoot[ent.GetMemIndex()] = true
	settings.IsPakBoot = true
	settings.Pakfile[ent.GetMemIndex()] = path

	return nil
}

func (this *Producer) SetPostCallback(index int, f func(index int)) {
	this.PostCallback[index] = f
}

func (this *Producer) SetMasterLayerPos(index int, x, y float64) {
	this.MasterLayerPos[index] = types.LayerPosMod{x, y}
	this.ForceUpdate[index] = true
}

func (this *Producer) GetMasterLayerPos(index int) (float64, float64) {
	return this.MasterLayerPos[index].XPercent, this.MasterLayerPos[index].YPercent
}

func (this *Producer) Select(slot int) {
	slot = slot % memory.OCTALYZER_NUM_INTERPRETERS

	settings.VideoSuspended = true
	defer func() {
		settings.VideoSuspended = false
		// Finish with the input context
		this.SetInputContext(slot)
		this.SetContext(slot)
	}()
	time.Sleep(5 * time.Millisecond)

	for i := 0; i < memory.OCTALYZER_NUM_INTERPRETERS; i++ {

		if slot == i {

			//fmt.Printf("Activate slot %d\n", i)

			if this.VM[i] == nil || this.VM[i].IsDying() {
				// Go routine is not running for this slot
				fmt.Printf("Creating an interpreter in slot %d as it was not running yet.\n", i)
				this.CreateVM(i, this.BuildBootConfig(i), nil)
			}

			this.VM[i].GetInterpreter().SetDisabled(false)
			this.AddressSpace.IntSetActiveState(i, 1)
			this.AddressSpace.IntSetLayerState(i, 1)
			this.VM[i].GetInterpreter().ResumeTheWorld()

			if this.VM[i].GetInterpreter().GetState() == types.EXEC6502 || this.VM[i].GetInterpreter().GetState() == types.DIRECTEXEC6502 {
				bus.StopClock()
			} else {
				bus.StartDefault()
			}

			settings.SetSubtitle(settings.SpecName[i])

		} else {
			this.AddressSpace.IntSetActiveState(i, 0)
			this.AddressSpace.IntSetLayerState(i, 0)
		}

	}

}

func (this *Producer) Activate(slot int) {
	slot = slot % memory.OCTALYZER_NUM_INTERPRETERS

	for i := 0; i < memory.OCTALYZER_NUM_INTERPRETERS; i++ {

		if slot == i {

			if this.VM[i] == nil || this.ThreadActive[i] != this.VM[i].GetInterpreter().GetUUID() {
				// Go routine is not running for this slot
				fmt.Printf("Creating an interpreter in slot %d as it was not running yet.\n", i)
				this.CreateVM(i, nil, nil)
				//go this.Executor(i)
			}

			this.VM[i].GetInterpreter().SetDisabled(false)
			this.AddressSpace.IntSetActiveState(i, 1)
			this.AddressSpace.IntSetLayerState(i, 1)
			this.VM[i].GetInterpreter().ResumeTheWorld()
		}

	}
}

func (this *Producer) ExecutorInit(slotid int) {
	switch settings.SpecFile[slotid] {
	case "apple2e-en.yaml", "apple2e-en-cpm.yaml", "apple2-plus.yaml", "apple2e.yaml":
		this.ExecutorInitApple2(slotid)
	case "bbcb.yaml":
		this.ExecutorInitBBC(slotid)
	}
}

func (this *Producer) ExecutorInitBBC(slotid int) {
	ent := this.VM[slotid].GetInterpreter()
	for ent.GetChild() != nil {
		ent = ent.GetChild()
	}

	this.StopMicroControls()

	if settings.MicroPakPath != "" {

		fmt.Printf("LOADING MICROPAK [%s]\n", settings.MicroPakPath)
		path := settings.MicroPakPath
		fp, err := files.ReadBytesViaProvider(files.GetPath(path), files.GetFilename(path))
		if err == nil {
			zp, err := files.NewOctContainer(&fp, path)
			if err == nil {
				editor.BootStrap(path, zp, ent)
				this.AddressSpace.IntSetSlotRestart(slotid, false)
			}
		}
		settings.MicroPakPath = ""

	}

	ent.Start6502(0xfa62, 0, 0, 0, 0x80, 0x1ff)

	// force CPU to bootstrap..
	cpu := apple2helpers.GetCPU(ent)
	cpu.Reset()
	cpu.SetWarpUser(1.0)
	ent.SetState(types.EXEC6502)

	fmt.Printf("CPU address @ 0x%.4x\n", cpu.PC)
}

func (this *Producer) CheckSystemCompat(slotid int, filename string) bool {

	ext := strings.ToLower(files.GetExt(filename))

	var data []byte
	var err error
	var osdMessage string

	// handy helper to send an osd if we set one up
	defer func() {
		if osdMessage != "" {
			time.AfterFunc(2*time.Second, func() {
				apple2helpers.OSDShow(this.GetInterpreter(slotid), osdMessage)
			})
		}
	}()

	var inlineLoad = func() error {
		if strings.HasPrefix(filename, "local:") {
			filename = filename[6:]
			data, err = ioutil.ReadFile(filename)
			if err != nil {
				return err
			}
		} else {
			fr, err := files.ReadBytesViaProvider(files.GetPath(filename), files.GetFilename(filename))
			if err != nil {
				return err
			}
			data = fr.Content
		}
		return nil
	}

	var profile = settings.SpecFile[slotid]

	switch ext {
	case "hdv", "2mg":
		if !strings.HasPrefix(profile, "apple") {
			profile = "apple2e-en.yaml"
			settings.SpecFile[slotid] = profile
		}
	case "dsk":
		if !strings.HasPrefix(profile, "apple") {
			profile = "apple2e-en.yaml"
			settings.SpecFile[slotid] = profile
		}
		// An inline load might id this disk as 13 sectors
		if inlineLoad() != nil {
			return false
		}
		dsk, err := disk.NewDSKWrapperBin(nil, data, filename)
		if err != nil {
			return false
		}
		if dsk.Format.ID == disk.DF_DOS_SECTORS_13 && !settings.DiskIIUse13Sectors[slotid] {
			settings.DiskIIUse13Sectors[slotid] = true
			osdMessage = "Switched to 13 Sector mode. (DSK/13S)"
			return true
		} else if settings.DiskIIUse13Sectors[slotid] { // switch back...
			settings.DiskIIUse13Sectors[slotid] = false
			osdMessage = "Switched to 16 Sector mode. (DSK/16S)"
			return true
		}
	case "d13":
		if !strings.HasPrefix(profile, "apple") {
			profile = "apple2e-en.yaml"
			settings.SpecFile[slotid] = profile
		}
		// If we have a d13 file, switch on 13 sector firmware and request a reconfigure
		if !settings.DiskIIUse13Sectors[slotid] {
			settings.DiskIIUse13Sectors[slotid] = true
			osdMessage = "Switched to 13 Sector mode. (D13)"
			return true
		}
	case "woz":
		if !strings.HasPrefix(profile, "apple") {
			profile = "apple2e-en.yaml"
			settings.SpecFile[slotid] = profile
		}

		if inlineLoad() != nil {
			return false
		}

		// try load woz2
		dsk2, err := woz2.NewWOZ2Image(bytes.NewBuffer(data), memory.NewMemByteSlice(len(data)))
		if err != nil {
			return false
		}

		hwcompat := dsk2.INFO.CompatibleHardware()
		var current woz2.WOZ2CompatibleHardware
		//log.Printf("Profile is %s", profile)
		switch profile {
		case "apple2-dsys.yaml", "apple2.yaml":
			current = 0x0001
		case "apple2e-en.yaml", "apple2e-en-cpm.yaml":
			current = 0x0010
		case "apple2-plus.yaml":
			current = 0x0002
		case "apple2e.yaml":
			current = 0x0004
		default:
			current = 0x0010
		}
		if hwcompat == 0 {
			hwcompat = current // 2e
		}

		var needReconfig bool

		//log.Printf("hwcompat=%x, current=%x", hwcompat, current)
		if current&hwcompat == 0 {
			switch {
			case hwcompat&woz2.WOZ2CompatApple2eEnhanced != 0:
				profile = "apple2e-en.yaml"
			case hwcompat&woz2.WOZ2CompatApple2Plus != 0:
				profile = "apple2-plus.yaml"
			case hwcompat&woz2.WOZ2CompatApple2 != 0:
				profile = "apple2-dsys.yaml"
			default:
				profile = "apple2e-en.yaml"
			}
			settings.SpecFile[slotid] = profile
			osdMessage += "Profile: " + strings.TrimSuffix(profile, ".yaml") + ". "
			//this.AddressSpace.IntSetSlotRestart(slotid, true)
			needReconfig = true
		}

		bsf := dsk2.INFO.BootSectorFormat()
		switch {
		case bsf == woz2.WOZ2BSF13Sector && !settings.DiskIIUse13Sectors[slotid]:
			settings.DiskIIUse13Sectors[slotid] = true
			osdMessage += "Switched to 13 Sector mode. (WOZ2/BSF)"
			needReconfig = true
		case bsf == woz2.WOZ2BSF16Sector && settings.DiskIIUse13Sectors[slotid]:
			settings.DiskIIUse13Sectors[slotid] = false
			osdMessage += "Switched to 16 Sector mode. (WOZ2/BSF)"
			needReconfig = true
		}

		return needReconfig
	}

	return false

}

func (this *Producer) IsReconfigureNeeded(slotid int, filename string, profile string) (bool, string) {

	ext := strings.ToLower(files.GetExt(filename))

	var data []byte
	var err error
	var osdMessage string

	// handy helper to send an osd if we set one up
	defer func() {
		if osdMessage != "" {
			time.AfterFunc(2*time.Second, func() {
				apple2helpers.OSDShow(this.GetInterpreter(slotid), osdMessage)
			})
		}
	}()

	var inlineLoad = func() error {
		if strings.HasPrefix(filename, "local:") {
			filename = filename[6:]
			data, err = ioutil.ReadFile(filename)
			if err != nil {
				return err
			}
		} else {
			fr, err := files.ReadBytesViaProvider(files.GetPath(filename), files.GetFilename(filename))
			if err != nil {
				return err
			}
			data = fr.Content
		}
		return nil
	}

	switch ext {
	case "2mg", "hdv":
		if !strings.HasPrefix(profile, "apple") {
			profile = "apple2e-en.yaml"
		}
	case "dsk", "po", "do":
		if !strings.HasPrefix(profile, "apple") {
			profile = "apple2e-en.yaml"
		}
		// An inline load might id this disk as 13 sectors
		if inlineLoad() != nil {
			return false, profile
		}
		dsk, err := disk.NewDSKWrapperBin(nil, data, filename)
		if err != nil {
			return false, profile
		}
		if dsk.Format.ID == disk.DF_DOS_SECTORS_13 && !settings.DiskIIUse13Sectors[slotid] {
			settings.DiskIIUse13Sectors[slotid] = true
			osdMessage = "Switched to 13 Sector mode. (DSK/13S)"
			return true, profile
		} else if settings.DiskIIUse13Sectors[slotid] { // switch back...
			settings.DiskIIUse13Sectors[slotid] = false
			osdMessage = "Switched to 16 Sector mode. (DSK/16S)"
			return true, profile
		}
	case "d13":
		if !strings.HasPrefix(profile, "apple") {
			profile = "apple2e-en.yaml"
		}
		// If we have a d13 file, switch on 13 sector firmware and request a reconfigure
		if !settings.DiskIIUse13Sectors[slotid] {
			settings.DiskIIUse13Sectors[slotid] = true
			osdMessage = "Switched to 13 Sector mode. (D13)"
			return true, profile
		}
	case "woz":
		if !strings.HasPrefix(profile, "apple") {
			profile = "apple2e-en.yaml"
		}
		if inlineLoad() != nil {
			return false, profile
		}

		// try load woz2
		dsk2, err := woz2.NewWOZ2Image(bytes.NewBuffer(data), memory.NewMemByteSlice(len(data)))
		if err != nil {
			return false, profile
		}

		hwcompat := dsk2.INFO.CompatibleHardware()
		var current woz2.WOZ2CompatibleHardware
		var profile = settings.SpecFile[slotid]
		//log.Printf("Profile is %s", profile)
		switch profile {
		case "apple2-dsys.yaml", "apple2.yaml":
			current = 0x0001
		case "apple2e-en.yaml", "apple2e-en-cpm.yaml":
			current = 0x0010
		case "apple2-plus.yaml":
			current = 0x0002
		case "apple2e.yaml":
			current = 0x0004
		default:
			current = 0x0010
		}
		if hwcompat == 0 {
			hwcompat = current // 2e
		}

		var needReconfig bool

		//log.Printf("hwcompat=%x, current=%x", hwcompat, current)
		if current&hwcompat == 0 {
			switch {
			case hwcompat&woz2.WOZ2CompatApple2eEnhanced != 0:
				profile = "apple2e-en.yaml"
			case hwcompat&woz2.WOZ2CompatApple2Plus != 0:
				profile = "apple2-plus.yaml"
			case hwcompat&woz2.WOZ2CompatApple2 != 0:
				profile = "apple2-dsys.yaml"
			default:
				profile = "apple2e-en.yaml"
			}
			settings.SpecFile[slotid] = profile
			osdMessage += "Profile: " + strings.TrimSuffix(profile, ".yaml") + ". "
			//this.AddressSpace.IntSetSlotRestart(slotid, true)
			needReconfig = true
		}

		bsf := dsk2.INFO.BootSectorFormat()
		switch {
		case bsf == woz2.WOZ2BSF13Sector && !settings.DiskIIUse13Sectors[slotid]:
			settings.DiskIIUse13Sectors[slotid] = true
			osdMessage += "Switched to 13 Sector mode. (WOZ2/BSF)"
			needReconfig = true
		case bsf == woz2.WOZ2BSF16Sector && settings.DiskIIUse13Sectors[slotid]:
			settings.DiskIIUse13Sectors[slotid] = false
			osdMessage += "Switched to 16 Sector mode. (WOZ2/BSF)"
			needReconfig = true
		}

		return needReconfig, profile
	}

	return false, profile

}

func (this *Producer) ExecutorInitApple2(slotid int) {

	ent := this.VM[slotid].GetInterpreter()
	for ent.GetChild() != nil {
		ent = ent.GetChild()
	}

	if settings.MicroPakPath != "" {

		fmt.Printf("LOADING MICROPAK [%s]\n", settings.MicroPakPath)
		path := settings.MicroPakPath
		this.BootStrap(path)
		this.AddressSpace.IntSetSlotRestart(slotid, false)
		settings.MicroPakPath = ""

	}

	fn := settings.AuxFonts[slotid][0]
	f, err := font.LoadFromFile(fn)
	if err == nil {
		settings.DefaultFont[slotid] = f
	}

	settings.PureBootCheck(slotid)

	fmt.Println("==============RESET STATE==================")
	fmt.Printf(" Disk 0: %s\n", settings.PureBootVolume[slotid])
	fmt.Printf(" SmartP: %s\n", settings.PureBootSmartVolume[slotid])
	fmt.Printf(" PureBoot = %v\n", settings.PureBoot(slotid))
	fmt.Printf(" Has Reset command = %v\n", settings.ResetState[slotid] != nil)
	fmt.Printf(" Has state resume = %v\n", settings.SlotRestartContinueCPU[slotid])
	fmt.Printf(" Has video playback = %v\n", settings.VideoPlayFrames[slotid] != nil)
	fmt.Printf(" Has PB state restore = %v\n", settings.PureBootRestoreState[slotid] != "" || settings.PureBootRestoreStateBin[slotid] != nil)
	fmt.Println("===========================================")

	if settings.ResetState[slotid] != nil {

		apple2helpers.MODE40(ent)

		s := settings.ResetState[slotid]
		settings.ResetState[slotid] = nil

		ent.Bootstrap(s.Dialect, true)
		ent.SetWorkDir(s.WorkDir)
		log.Printf("Bootstrap command is %s", s.Command)
		ent.SetMicroControl(s.IsControl)
		ent.Parse(s.Command)

		if s.IsControl {
			log.Printf("Hiding display layers in slot %d", slotid)
			this.AddressSpace.IntSetActiveState(slotid, 0)
			this.AddressSpace.IntSetLayerState(slotid, 0)
			if l, ok := ent.GetHUDLayerByID("TEXT"); ok {
				l.SetActive(false)
			}
		}

		bus.StartDefault()

	} else if settings.SlotRestartContinueCPU[slotid] {
		fmt.Println("SOFT RESUME IN PROGRESS")
		settings.SlotRestartContinueCPU[slotid] = false
		// quick resume of CPU
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
	} else if settings.VideoPlayFrames[slotid] != nil {

		bus.StartDefault()

		fmt.Println("Playback...")

		ent.PlayBlocks(settings.VideoPlayFrames[slotid], settings.VideoPlayBackwards[slotid], settings.VideoBackSeekMS[slotid])
		settings.VideoBackSeekMS[slotid] = 0
		settings.VideoPlayFrames[slotid] = nil

	} else if settings.PureBootRestoreState[slotid] != "" || settings.PureBootRestoreStateBin[slotid] != nil {

		fmt.Println("In pureboot restore...")

		apple2helpers.OSDPanel(ent, false)
		settings.AudioPacketReverse[slotid] = false
		f := freeze.NewEmptyState(ent)
		var err error
		if settings.PureBootRestoreState[slotid] != "" {
			fmt.Printf("Restoring STATE from file %s\n", settings.PureBootRestoreState[slotid])
			err = f.LoadFromFile(settings.PureBootRestoreState[slotid])
			settings.PureBootRestoreState[slotid] = ""
		} else {
			fmt.Printf("Restoring STATE from ram\n")
			err = f.LoadFromBytes(settings.PureBootRestoreStateBin[slotid])
			settings.PureBootRestoreStateBin[slotid] = nil
		}

		if err == nil {
			fmt.Println("OK")
			settings.SetPureBoot(slotid, true)
			settings.SlotZPEmu[slotid] = false
			f.Apply(ent)
		} else {
			fmt.Println(err)
		}

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
			url := fmt.Sprintf("http://localhost:%d/?attach=%d", settings.DebuggerPort, settings.DebuggerAttachSlot-1)
			err := utils.OpenURL(url)
			if err != nil {
				log.Printf("Failed to open debugger: %v", err)
			}
			cpu.RunState = mos6502.CrsPaused // start paused .. better for us
		}

	} else if settings.PureBoot(slotid) {

		apple2helpers.MODE40Preserve(ent)

		fmt.Println("In pureboot init...")

		apple2helpers.OSDPanel(ent, false)

		if settings.AutoLiveRecording() && slotid == 0 {
			ent.StopRecording()
			ent.StartRecording("", false) // enable live record
		}

		settings.SlotZPEmu[slotid] = false

		log.Println("Running in a pure boot mode")

		var start int = 64098

		if strings.HasPrefix(settings.PureBootVolume[slotid], "ram:") {

			addrstr := settings.PureBootVolume[slotid][4:]
			start = utils.StrToInt(addrstr)

			fmt.Printf("Using custom boot address $%.4x\n", start)

		} else {

			if settings.PureBootSmartVolume[slotid] != "" {

				log.Printf("Smart volume: %s", settings.PureBootSmartVolume[slotid])

				if strings.HasPrefix(settings.PureBootSmartVolume[slotid], "local:") {
					f, e := os.Open(strings.Replace(settings.PureBootSmartVolume[slotid], "local:", "", -1))
					if e == nil {
						log.Println("read file ok")
						data, e := ioutil.ReadAll(f)
						f.Close()
						if e == nil {
							log.Println("Injecting service bus")
							servicebus.SendServiceBusMessage(
								slotid,
								servicebus.SmartPortInsertBytes,
								servicebus.DiskTargetBytes{
									Drive:    0,
									Bytes:    data,
									Filename: settings.PureBootSmartVolume[slotid],
								},
							)
						}

					} else {
						log.Printf("File read failed: %v", e)
					}
				} else {
					log.Println("Non local smartport file")
					//hardware.DiskInsert(ent, 0, settings.PureBootVolume[slotid], settings.PureBootVolumeWP[slotid])
					servicebus.SendServiceBusMessage(
						slotid,
						servicebus.SmartPortInsertFilename,
						servicebus.DiskTargetString{
							Drive:    0,
							Filename: settings.PureBootSmartVolume[slotid],
						},
					)
				}
			} else {
				log.Println("REBOOT WITHOUT SMARTPORT...")
			}

			if settings.PureBootVolume[slotid] != "" {

				if this.CheckSystemCompat(slotid, settings.PureBootVolume[slotid]) {
					// change needed
					ent.LoadSpec(settings.SpecFile[slotid])
				}

				if strings.HasPrefix(settings.PureBootVolume[slotid], "local:") {
					f, e := os.Open(strings.Replace(settings.PureBootVolume[slotid], "local:", "", -1))
					if e == nil {
						data, e := ioutil.ReadAll(f)
						f.Close()
						if e == nil {
							servicebus.SendServiceBusMessage(
								slotid,
								servicebus.DiskIIInsertBytes,
								servicebus.DiskTargetBytes{
									Drive:    0,
									Bytes:    data,
									Filename: settings.PureBootVolume[slotid],
								},
							)
						}

					}
				} else {
					//hardware.DiskInsert(ent, 0, settings.PureBootVolume[slotid], settings.PureBootVolumeWP[slotid])
					servicebus.SendServiceBusMessage(
						slotid,
						servicebus.DiskIIInsertFilename,
						servicebus.DiskTargetString{
							Drive:    0,
							Filename: settings.PureBootVolume[slotid],
						},
					)
				}
			}

			if settings.PureBootVolume2[slotid] != "" {
				if strings.HasPrefix(settings.PureBootVolume2[slotid], "local:") {
					f, e := os.Open(strings.Replace(settings.PureBootVolume2[slotid], "local:", "", -1))
					if e == nil {
						data, e := ioutil.ReadAll(f)
						f.Close()
						if e == nil {
							servicebus.SendServiceBusMessage(
								slotid,
								servicebus.DiskIIInsertBytes,
								servicebus.DiskTargetBytes{
									Drive:    1,
									Bytes:    data,
									Filename: settings.PureBootVolume2[slotid],
								},
							)
						}

					}
				} else {
					servicebus.SendServiceBusMessage(
						slotid,
						servicebus.DiskIIInsertFilename,
						servicebus.DiskTargetString{
							Drive:    1,
							Filename: settings.PureBootVolume2[slotid],
						},
					)
				}
			}

		}

		// Turn zeropage off
		ent.GetMemoryMap().BlockMapper[ent.GetMemIndex()].DisableBlocks([]string{"apple2iozeropage"})
		if settings.PureBootBanks != nil && len(settings.PureBootBanks) > 0 {
			ent.GetMemoryMap().BlockMapper[ent.GetMemIndex()].EnableBlocks(settings.PureBootBanks)
		}
		ent.GetMemoryMap().IntSetProcessorState(ent.GetMemIndex(), 1) // Force 65C02... we are trying to be a 2e

		ent.Start6502(0xfa62, 0, 0, 0, 0x80, 0x1ff)

		// force CPU to bootstrap..
		cpu := apple2helpers.GetCPU(ent)
		cpu.Reset()
		cpu.SetWarpUser(1.0)

		ml := settings.MemLocks[ent.GetMemIndex()]
		if ml != nil && len(ml) > 0 {
			for addr, value := range ml {
				cpu.LockValue[addr] = value
			}
		}

		// Attach debugger here if we need it..
		if settings.DebuggerOn && settings.DebuggerAttachSlot-1 == slotid {
			url := fmt.Sprintf("http://localhost:%d/?attach=%d", settings.DebuggerPort, settings.DebuggerAttachSlot)
			err := utils.OpenURL(url)
			if err != nil {
				log.Printf("Failed to open debugger: %v", err)
			}
			cpu.RunState = mos6502.CrsPaused // start paused .. better for us
		}

	}

}

func (this *Producer) VMMediaChange(slotid int, drive int, filename string) error {
	settings.Pakfile[slotid] = ""
	p := &servicebus.LaunchEmulatorTarget{
		Drive:    drive,
		Filename: filename,
	}
	n := p.Filename
	drive = p.Drive
	ext := strings.ToLower(files.GetExt(n))
	if files.IsBootable(ext) {
		settings.MicroPakPath = ""
		if !settings.BlueScreen {
			log.Printf("Got file named %s", n)
			data, err := files.ReadBytes(n)
			if files.Apple2IsHighCapacity(ext, len(data)) {
				log.Println("Smartport handling")
				if err == nil {
					log.Println("No read error")
					disk := "local:" + n
					e := this.VM[slotid].GetInterpreter()
					//go dropAnimation(drive)
					settings.PureBootSmartVolume[e.GetMemIndex()] = disk
					//hardware.DiskInsert(e, 0, settings.PureBootVolume[e.MemIndex], settings.PureBootVolumeWP[e.MemIndex])

					if !strings.HasPrefix(settings.SpecFile[e.GetMemIndex()], "apple2") {
						log.Printf("Rebooting into apple2 mode!!!")
						settings.SpecFile[e.GetMemIndex()] = "apple2e-en.yaml"
						e.GetMemoryMap().IntSetSlotRestart(e.GetMemIndex(), true)
						return nil
					}

					servicebus.SendServiceBusMessage(
						e.GetMemIndex(),
						servicebus.SmartPortInsertFilename,
						servicebus.DiskTargetString{
							Drive:    0,
							Filename: disk,
						},
					)
				} else {
					return err
				}
			} else {
				disk := "local:" + n
				e := this.VM[slotid].GetInterpreter()
				//go dropAnimation(drive)
				switch drive {
				case 1:
					settings.PureBootVolume[e.GetMemIndex()] = disk
					settings.PureBootSmartVolume[e.GetMemIndex()] = ""
				case 2:
					settings.PureBootVolume2[e.GetMemIndex()] = disk
				}
				//hardware.DiskInsert(e, 0, settings.PureBootVolume[e.MemIndex], settings.PureBootVolumeWP[e.MemIndex])
				if !strings.HasPrefix(settings.SpecFile[e.GetMemIndex()], "apple2") {
					log.Printf("Rebooting into apple2 mode!!!")
					settings.SpecFile[e.GetMemIndex()] = "apple2e-en.yaml"
					e.GetMemoryMap().IntSetSlotRestart(e.GetMemIndex(), true)
					return nil
				}

				servicebus.SendServiceBusMessage(
					e.GetMemIndex(),
					servicebus.DiskIIInsertFilename,
					servicebus.DiskTargetString{
						Drive:    drive - 1,
						Filename: disk,
					},
				)
			}
		} else {
			settings.SplashDisk = "local:" + n
		}
	} else {
		// handle other types
		if runtime.GOOS == "windows" && strings.HasPrefix(n, os.Getenv("HOMEDRIVE")) {
			n = n[2:]
		}
		n = strings.Replace(n, "\\", "/", -1)
		n = "/fs/" + strings.Trim(n, "/")
		log.Printf("Bootstrap path: %s", n)
		ent := this.VM[slotid].GetInterpreter()

		if !strings.HasPrefix(settings.SpecFile[ent.GetMemIndex()], "apple2") {
			settings.SpecFile[ent.GetMemIndex()] = "apple2-en.yaml"
		}

		if ext == "wav" {
			fp, err := files.ReadBytesViaProvider(files.GetPath(n), files.GetFilename(n))
			if err == nil {
				mr, ok := ent.GetMemoryMap().InterpreterMappableAtAddress(ent.GetMemIndex(), 0xc000)
				if ok {
					io, ok := mr.(*apple2.Apple2IOChip)
					if ok {
						err := io.TapeAttach(fp.Content)
						log2.Printf("[tape-attach] Tape attach result: %v", err)
					}
				}
			}
		} else if files.IsBinary(ext) {
			fp, err := files.ReadBytesViaProvider(files.GetPath(n), files.GetFilename(n))
			if err == nil {
				startfile := fp.FileName
				path := files.GetPath(n)
				log.Printf("Going to try BRUN of file %s", startfile)
				apple2helpers.CommandResetOverride(
					ent,
					"fp",
					"/"+strings.Trim(path, "/")+"/",
					"PRINT CHR$(4);\"BRUN "+startfile[0:len(startfile)-len(ext)-1]+"\"",
				)
			}
		} else if launchable, dialect, command := files.IsLaunchable(ext); launchable {
			startfile := n
			path := files.GetPath(n)

			// ent.Halt()
			// ent.Halt6502(0)
			// ent.Bootstrap(dialect, true)
			// ent.SetWorkDir(files.GetPath(n))

			command = strings.Replace(command, "%f", startfile, -1)
			// fmt.Printf("Launching command: %s\n", command)
			// s := &settings.RState{
			// 	Dialect:   dialect,
			// 	Command:   command,
			// 	WorkDir:   path,
			// 	IsControl: p.IsControl,
			// }
			// settings.ResetState[ent.GetMemIndex()] = s
			apple2helpers.CommandResetOverride(
				ent,
				dialect,
				path,
				command,
			)
			//settings.ResetState[ent.GetMemIndex()].IsControl = p.IsControl

			ent.GetMemoryMap().IntSetSlotRestart(ent.GetMemIndex(), true)
			fmt.Printf("Requesting controlled restart of slot %d\n", ent.GetMemIndex())
		} else if runnable, d := files.IsRunnable(ext); runnable {
			log.Printf("Going to try running command: %s", n)
			switch d {
			case "fp":
				apple2helpers.CommandResetOverride(
					ent,
					d,
					files.GetPath(n),
					"run "+n,
				)
				//settings.ResetState[ent.GetMemIndex()].IsControl = p.IsControl
			case "int":
				apple2helpers.CommandResetOverride(
					ent,
					d,
					files.GetPath(n),
					"run "+n,
				)
				//settings.ResetState[ent.GetMemIndex()].IsControl = p.IsControl
			case "logo":
				apple2helpers.CommandResetOverride(
					ent,
					d,
					files.GetPath(n),
					"load \""+n,
				)
				//settings.ResetState[ent.GetMemIndex()].IsControl = p.IsControl
			}
		} else if ext == "pak" {

			this.Select(0)
			settings.MicroPakPath = n
			go this.RebootVM(0)

		} else if ext == "z80" || ext == "sna" {
			settings.SpecFile[slotid] = "zx-spectrum-48k.yaml"

			s, e := spectrum.ReadSnapshot(n)
			if e == nil {
				if s.Is128K {
					log.Printf("Switching to 128k mode for this file")
					settings.SpecFile[slotid] = "zx-spectrum-128k.yaml"
				}
				settings.VMLaunch[slotid] = &settings.VMLauncherConfig{
					ZXState: n,
				}

				this.AddressSpace.IntSetSlotRestart(slotid, true)

			} else {
				apple2helpers.OSDShow(ent, "Z80: "+e.Error())
			}

		}
	}

	return nil

}

func (this *Producer) BuildBootConfig(slotid int) *VMLauncherConfig {

	// if strings.Contains(settings.SpecFile[slotid], "spectrum") {
	// 	return &VMLauncherConfig{
	// 		Forceboot: true,
	// 	}
	// }
	if settings.SystemType == "nox" {
		_, f, _ := files.ReadDirViaProvider("/", "Nox*Archaist*.HDV")
		if len(f) == 0 {
			_, f, _ = files.ReadDirViaProvider("/", "Nox*Archaist*.2MG")
		}
		sort.SliceStable(f, func(i, j int) bool {
			return f[i].Name < f[j].Name
		})
		if len(f) > 0 {
			idx := len(f) - 1
			return &VMLauncherConfig{
				SmartPort: "/" + f[idx].Name + "." + f[idx].Extension,
			}
		} else {

			return &VMLauncherConfig{
				WorkingDir: "/boot",
				RunFile:    "/boot/novolume",
				Dialect:    "fp",
			}
		}
	}

	// is a premade item... takes precedence
	if settings.VMLaunch[slotid] != nil {
		//log.Printf("Received launch config: %+v", settings.VMLaunch[slotid])
		c := &VMLauncherConfig{
			WorkingDir: settings.VMLaunch[slotid].WorkingDir,
			Disks:      settings.VMLaunch[slotid].Disks,
			SmartPort:  settings.VMLaunch[slotid].SmartPort,
			RunFile:    settings.VMLaunch[slotid].RunFile,
			RunCommand: settings.VMLaunch[slotid].RunCommand,
			Dialect:    settings.VMLaunch[slotid].Dialect,
		}
		if len(c.Disks) > 0 {
			if needed, p := this.IsReconfigureNeeded(slotid, c.Disks[0], settings.SpecFile[slotid]); needed {
				settings.SpecFile[slotid] = p
			}
		} else if c.SmartPort != "" {
			if needed, p := this.IsReconfigureNeeded(slotid, c.SmartPort, settings.SpecFile[slotid]); needed {
				settings.SpecFile[slotid] = p
			}
		}
		if settings.VMLaunch[slotid].Pakfile != "" {
			settings.DiskIIUse13Sectors[slotid] = false
			settings.MicroPakPath = settings.VMLaunch[slotid].Pakfile
			data, err := files.ReadBytesViaProvider(files.GetPath(settings.MicroPakPath), files.GetFilename(settings.MicroPakPath))
			if err == nil {
				zp, err := files.NewOctContainer(&data, settings.MicroPakPath)
				if err == nil && zp.GetStartup() != "" {
					c.Pakfile = zp
				}
			}
		}
		if settings.VMLaunch[slotid].ZXState != "" {
			log.Printf("got zx state file: %s\n", settings.VMLaunch[slotid].ZXState)
			sna, err := spectrum.ReadSnapshot(settings.VMLaunch[slotid].ZXState)
			log.Printf("snapshot load status: %v", err)
			if err == nil {
				settings.SnapshotFile[slotid] = settings.VMLaunch[slotid].ZXState
				log.Println("snapshot appears valid")
				c.ZXState = sna
			}
		}
		settings.VMLaunch[slotid] = nil
		return c
	}

	// soft resume
	if settings.SlotRestartContinueCPU[slotid] {
		settings.SlotRestartContinueCPU[slotid] = false
		return &VMLauncherConfig{
			ResumeCPU: true,
		}
	}

	// video playback
	if settings.VideoPlayFrames[slotid] != nil {
		c := &VMLauncherConfig{
			VideoFrames: settings.VideoPlayFrames[slotid],
		}
		settings.VideoPlayFrames[slotid] = nil
		return c
	}

	// freeze state
	if settings.PureBootRestoreState[slotid] != "" || settings.PureBootRestoreStateBin[slotid] != nil {
		apple2helpers.OSDPanel(this.VM[slotid].GetInterpreter(), false)
		settings.AudioPacketReverse[slotid] = false
		f := freeze.NewEmptyState(this.VM[slotid].GetInterpreter())
		var err error
		if settings.PureBootRestoreState[slotid] != "" {
			fmt.Printf("Restoring STATE from file %s\n", settings.PureBootRestoreState[slotid])
			err = f.LoadFromFile(settings.PureBootRestoreState[slotid])
			settings.PureBootRestoreState[slotid] = ""
		} else {
			fmt.Printf("Restoring STATE from ram\n")
			err = f.LoadFromBytes(settings.PureBootRestoreStateBin[slotid])
			settings.PureBootRestoreStateBin[slotid] = nil
		}
		if err == nil {
			return &VMLauncherConfig{
				Freeze: f,
			}
		}
	}

	// Is pak boot scenario
	if slotid == 0 && settings.MicroPakPath != "" {
		settings.FirstBoot[0] = true
		settings.DiskIIUse13Sectors[slotid] = false
		data, err := files.ReadBytesViaProvider(files.GetPath(settings.MicroPakPath), files.GetFilename(settings.MicroPakPath))
		if err == nil {
			zp, err := files.NewOctContainer(&data, settings.MicroPakPath)
			if err == nil && zp.GetStartup() != "" {
				return &VMLauncherConfig{
					Pakfile: zp,
				}
			}
		}
	}

	// Is pure boot scenario
	if settings.PureBootVolume[slotid] != "" || settings.PureBootSmartVolume[slotid] != "" || settings.ForcePureBoot[slotid] {
		if settings.PureBootVolume[slotid] != "" {
			if needed, p := this.IsReconfigureNeeded(slotid, settings.PureBootVolume[slotid], settings.SpecFile[slotid]); needed {
				settings.SpecFile[slotid] = p
			}
		}
		return &VMLauncherConfig{
			Disks:     []string{settings.PureBootVolume[slotid], settings.PureBootVolume2[slotid]},
			SmartPort: settings.PureBootSmartVolume[slotid],
		}
	}

	// fallback for slot 0 to menu
	if slotid == 0 && !settings.IsRemInt {
		return &VMLauncherConfig{
			RunFile: "/boot/boot.apl",
			Dialect: "fp",
		}
	}

	return nil
}

func (this *Producer) RebootVM(slotid int) {

	oldvm := this.VM[slotid]

	if oldvm.e.IsRecordingVideo() {
		oldvm.e.StopRecording()
	}

	isChild := (oldvm != nil) && (oldvm.Parent != nil)
	var parent *VM
	if isChild {
		parent = oldvm.Parent
	}

	oldls := this.AddressSpace.IntGetLayerState(slotid)
	oldas := this.AddressSpace.IntGetActiveState(slotid)
	this.CreateVM(slotid, this.BuildBootConfig(slotid), nil)
	if isChild {
		this.AddressSpace.IntSetActiveState(slotid, oldas)
		this.AddressSpace.IntSetLayerState(slotid, oldls)
		parent.Logf("replacing child vm with new instance...")
		parent.RemoveDependant(oldvm)
		parent.AddDependant(this.VM[slotid])

	}
}

func (this *Producer) PreInit(slotid int) {

	settings.BlueScreen = false

	if !strings.HasPrefix(settings.SpecFile[slotid], "apple2") {
		return
	}

	//ent := this.GetInterpreter(slotid)

	settings.SlotZPEmu[slotid] = false

	vm := this.VM[slotid]
	if vm == nil {
		return
	}

	vm.PreInit()

}

func (this *Producer) Executor(slotid int) {

	var ent interfaces.Interpretable
	ent = this.VM[slotid].GetInterpreter()
	for ent.GetChild() != nil {
		ent = ent.GetChild()
	}
	this.ThreadActive[slotid] = ent.GetUUID()

	this.AddressSpace.Zero(slotid)

	fmt.Printf("-> Spawning executor for slot #%d (with id %.8x)\n", slotid, ent.GetUUID())

	// Setup a catch here to recover an exception from
	if slotid == 0 && !settings.BootCheckDone {
		//ent.StopTheWorld()
		if !settings.NoUpdates {
			ent.BootCheck()
		} else {
			time.Sleep(500 * time.Millisecond)
		}
		ent.ResumeTheWorld()
		settings.BootCheckDone = true
	}

	// set state

	panic.Do(
		func() {

			ent = this.VM[slotid].GetInterpreter()
			settings.PureBootCheck(slotid)

			// ServiceBus launch events...
			servicebus.Subscribe(
				slotid,
				servicebus.LaunchEmulator,
				ent,
			)

			// if slotid == 0 {
			// 	debugger.DebuggerInstance.AttachSlot(slotid)
			// 	debugger.DebuggerInstance.Run(slotid, arthurDent)
			// }

			this.PreInit(slotid)

			this.AddressSpace.SlotReset(slotid)

			fmt.Printf("At startslot mark for slot #%d\n", slotid)
			fmt.Printf("Pureboot state = %v\n", settings.PureBoot(slotid))
			fmt.Printf("volume in drive a is %s\n", settings.PureBootVolume[slotid])

			if settings.FirstBoot[slotid] {
				p, e := files.LoadDefaultState(ent)
				if e == nil && p != nil && this.GetPState(slotid) == nil {
					fmt.Println("Applying default state...")
					this.SetPState(slotid, p, p.Filepath)
				}
			}

			// if slotid == 0 {
			// 	this.StartMicroControls()
			// }

			if settings.MusicTrack[slotid] != "" {
				fmt.Printf("Launching music file %s\n", settings.MusicTrack[slotid])
				ent.PlayMusic(files.GetPath(settings.MusicTrack[slotid]), files.GetFilename(settings.MusicTrack[slotid]), settings.MusicLeadin[slotid], settings.MusicFadein[slotid])
				settings.MusicTrack[slotid] = ""
			}

			parent := this.VM[slotid].GetInterpreter()

			for this.ThreadActive[slotid] == parent.GetUUID() {

				ent = parent
				for ent.GetChild() != nil {
					ent = ent.GetChild()
				}

				// check for control programs
				this.StartPendingControlPrograms()

				// ServiceBus processing
				ent.ServiceBusProcessPending()

				if this.AddressSpace.IntGetSlotMenu(slotid) {
					r := bus.IsClock()
					if !r {
						bus.StartDefault() // resume clock
					}
					editor.TestMenu(ent)
					cpu := apple2helpers.GetCPU(ent)
					cpu.ResetSliced()

					if !ent.NeedsClock() {
						bus.StopClock()
					}
					this.AddressSpace.IntSetSlotMenu(slotid, false)
				}

				if this.AddressSpace.IntGetSlotInterrupt(slotid) && !ent.IsRemote() {
					r := bus.IsClock()

					if !r {
						bus.StartDefault() // resume clock
					}
					control.CatalogPresent(ent)
					this.AddressSpace.IntSetSlotInterrupt(slotid, false)
					cpu := apple2helpers.GetCPU(ent)
					cpu.ResetSliced()

					if !ent.NeedsClock() {
						bus.StopClock()
					}
				}

				if this.AddressSpace.IntGetSlotRestart(slotid) && !ent.IsRemote() {

					//for i := 0; i < settings.NUMSLOTS; i++ {
					servicebus.SendServiceBusMessage(
						slotid,
						servicebus.DiskIIFlush,
						0,
					)
					servicebus.SendServiceBusMessage(
						slotid,
						servicebus.DiskIIFlush,
						1,
					)
					servicebus.SendServiceBusMessage(
						slotid,
						servicebus.SmartPortFlush,
						0,
					)
					//}

					parent.SetChild(nil)
					ent = parent

					if slotid == 0 {
						this.StopMicroControls()
					}

					if ent.IsRecordingVideo() {
						ent.StopRecordingHard()
					}
					servicebus.UnsubscribeAll(slotid)
					this.StopAudio()
					apple2helpers.BeepReset(ent)
					settings.VideoSuspended = false
					fmt.Println("restart")
					this.AddressSpace.IntSetSlotRestart(slotid, false)
					apple2helpers.GFXDisable(ent)
					this.ResetSlot(slotid)
					this.RestartInterpreter(slotid)
					ent.PeripheralReset()
					fmt.Printf("After init call state is %s\n", ent.GetState())
					cpu := apple2helpers.GetCPU(ent)
					cpu.ResetSliced()
					this.AddressSpace.SlotReset(slotid)
					apple2helpers.BeepReset(ent)

					p, e := files.LoadDefaultState(ent)
					if e == nil && p != nil && this.GetPState(slotid) == nil {
						//log.Println("Applying default state...")
						this.SetPState(slotid, p, p.Filepath)
					}

					if settings.MusicTrack[slotid] != "" {
						fmt.Printf("Launching music file %s\n", settings.MusicTrack[slotid])
						if strings.HasSuffix(settings.MusicTrack[slotid], ".rst") {
							ent.GetMemoryMap().IntSetRestalgiaPath(ent.GetMemIndex(), settings.MusicTrack[slotid], true)
						} else {
							go ent.PlayMusic(files.GetPath(settings.MusicTrack[slotid]), files.GetFilename(settings.MusicTrack[slotid]), settings.MusicLeadin[slotid], settings.MusicFadein[slotid])
						}
						settings.MusicTrack[slotid] = ""
					}

					if settings.BackdropFile[slotid] != "" {
						this.AddressSpace.IntSetBackdrop(
							slotid,
							settings.BackdropFile[slotid],
							7,
							float32(settings.BackdropOpacity[slotid]),
							float32(settings.BackdropZoom[slotid]),
							float32(settings.BackdropZoomRatio[slotid]),
							settings.BackdropTrack[slotid],
						)
						settings.BackdropFile[slotid] = ""
					}

					servicebus.Subscribe(
						slotid,
						servicebus.LaunchEmulator,
						ent,
					)

					// if slotid == 0 {
					// 	debugger.DebuggerInstance.AttachSlot(slotid)
					// 	debugger.DebuggerInstance.Run(slotid, arthurDent)
					// }
				}

				settings.VideoSuspended = false

				ps := this.GetPState(slotid)
				pspath := this.GetPresentationSource(slotid)
				if ps != nil {
					//log.Printf("applying\n")
					this.ClearPState(slotid)
					owd := ent.GetWorkDir()
					ent.SetWorkDir(pspath)
					ps.Apply("init", ent)
					ent.SetWorkDir(owd)
				}

				// if slotid == 1 {
				// 	fmt.Printf("Current state is %s\n", ent.GetState().String())
				// }

				if ent.IsZ80Executing() {
					r := ent.DoCyclesZ80()
					if r == int(cpu.FE_SLEEP) {
						time.Sleep(1 * time.Millisecond)
					}
				} else if ent.Is6502Executing() {
					//settings.SlotZPEmu[slotid] = false
					r := ent.DoCycles6502()
					if r == int(cpu.FE_SLEEP) {
						time.Sleep(1 * time.Millisecond)
					} else if r == int(cpu.FE_HALTED) && !settings.PureBoot(slotid) {
						// Restore state
						time.Sleep(3 * time.Millisecond)
						ent.RestorePrevState()

						if settings.LaunchQuitCPUExit {
							os.Exit(0)
						}
					} else if r != int(cpu.FE_OK) {

						m := apple2helpers.NewMonitor(ent)
						if !m.Invoke(cpu.FEResponse(r)) {
							ent.Halt6502(r)
						}

						cpu := apple2helpers.GetCPU(ent)
						cpu.ResetSliced()
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
					for time.Now().Before(nexttime) {
						time.Sleep(500 * time.Microsecond)
					}
				} else if ent.IsRunningDirect() {
					if !bus.IsClock() {
						bus.StartDefault()
					}
					settings.SlotZPEmu[slotid] = true
					ent.RunStatementDirect()
					nexttime := ent.GetWaitUntil()
					for time.Now().Before(nexttime) {
						time.Sleep(500 * time.Microsecond)
					}
				} else {
					if !bus.IsClock() {
						bus.StartDefault()
					}
					ent.Interactive()
					time.Sleep(5 * time.Millisecond)
				}

				if slotid == 0 {
					// check and handle reset state
					this.Reset()
				}

			}

			servicebus.Unsubscribe(slotid, ent)
			fmt.Printf("*** Executor stopping for slot %d (with id %.8x)\n", slotid, ent.GetUUID())

			//this.Quit[slotid] <- true

			// check for a restart request

		},

		func(r interface{}) {

			ent := this.GetInterpreter(slotid)
			for ent.GetChild() != nil {
				ent = ent.GetChild()
			}

			servicebus.Unsubscribe(slotid, ent)

			// r is an exception...
			b := make([]byte, 8192)
			i := runtime.Stack(b, false)
			// Stack trace
			stackstr := string(b[0:i])

			f, _ := os.Create("stackdump.txt")
			f.WriteString(r.(error).Error() + "\n\n")
			f.Write(b[0:i])
			f.Close()

			tmp, _ := ent.FreezeBytes() // store memory here

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

			// lodge error report
			this.ThreadActive[slotid] = 0 // signal we died

			settings.PureBootVolume[slotid] = ""
			settings.PureBootVolume2[slotid] = ""
			settings.PureBootSmartVolume[slotid] = ""

			go this.Executor(slotid)
			return

		})

}

func (this *Producer) StopAudio() {
	// return
	// for _, e := range this.VM {
	// 	if e == nil {
	// 		continue
	// 	}
	// 	if e.GetInterpreter() != nil {
	// 		e.GetInterpreter().StopMusic()
	// 	}
	// }
}

func (this *Producer) StopTheWorld(slotid int) {
	this.VM[slotid].GetInterpreter().StopTheWorld()
	this.VM[slotid].GetInterpreter().SetDisabled(true)
	this.VM[slotid].GetInterpreter().GetMemoryMap().IntSetActiveState(slotid, 0)
	this.VM[slotid].GetInterpreter().GetMemoryMap().IntSetLayerState(slotid, 0)
}

func (this *Producer) Freeze(slotid int) []byte {

	f := freeze.NewFreezeState(this.VM[slotid].GetInterpreter(), false)
	return f.SaveToBytes()

}

func (this *Producer) ResumeTheWorld(slotid int) {

	for this.VM[slotid] == nil {
		time.Sleep(50 * time.Millisecond)
	}

	this.VM[slotid].GetInterpreter().ResumeTheWorld()
	this.VM[slotid].GetInterpreter().SetDisabled(false)
	this.VM[slotid].GetInterpreter().GetMemoryMap().IntSetActiveState(slotid, 1)
	this.VM[slotid].GetInterpreter().GetMemoryMap().IntSetLayerState(slotid, 1)
}

func (this *Producer) InjectCommands(slotid int, command string) {

	e := this.VM[slotid]
	//

	for _, ch := range command {
		e.GetMemoryMap().KeyBufferAdd(slotid, uint64(ch))
		time.Sleep(500 * time.Millisecond)
	}

}

func (this *Producer) Halt(slotid int) {
	this.VM[slotid].GetInterpreter().Halt6502(0)
	this.VM[slotid].GetInterpreter().Halt()
}
