package core

import (
	"time"

	s8webclient "paleotronic.com/api"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/files"
	"paleotronic.com/fmt"
	"paleotronic.com/runestring"
)

var producerIsReady bool

// func (this *Producer) Reboot() {
// 	// sleep 200 ms
// 	time.Sleep(200)

// 	// this.DiaAS = applesoft.NewDialectApplesoft()
// 	// this.DiaINT = appleinteger.NewDialectAppleInteger()
// 	// this.DiaShell = shell.NewDialectShell()

// 	for i := 0; i < memory.OCTALYZER_NUM_INTERPRETERS; i++ {
// 		this.StopInterpreter(i)
// 		this.CreateInterpreter(i, "main", shell.NewDialectShell(), settings.DefaultProfile, GetUUID())
// 		if i > 0 {
// 			this.VM[i].GetInterpreter().SetDisabled(true)
// 		}
// 	}
// 	//this.VM[0].Parse("run \"system/autoexec\"")
// 	this.BootTime = time.Now().Unix()
// }

func NewProducer(r *memory.MemoryMap, bootstrap string) *Producer {

	//fmt.Printf("BOOTSTRAP is %s\n", bootstrap)

	this := &Producer{}
	this.MasterLayerPos = make([]types.LayerPosMod, memory.OCTALYZER_NUM_INTERPRETERS)
	this.ForceUpdate = make([]bool, memory.OCTALYZER_NUM_INTERPRETERS)

	if r == nil {
		r = memory.NewMemoryMap()
	}

	this.AddressSpace = r

	count := 1
	if settings.IsRemInt {
		count = memory.OCTALYZER_NUM_INTERPRETERS
	}

	for i := 0; i < count; i++ {

		this.Quit[i] = make(chan bool)

		err := this.CreateVM(i, this.BuildBootConfig(i), nil)
		if err != nil {
			panic(err)
		}

	}

	this.LastExecute = time.Now().Unix()
	this.History = make([]runestring.RuneString, 0)

	this.RebootVM(0)

	settings.BlueScreen = false

	go this.RebootService()
	go this.MusicService()

	producerIsReady = true

	return this
}

func (this *Producer) MusicService() {
	for {
		for i, vm := range this.VM {
			if vm == nil {
				continue
			}
			if settings.MusicTrack[i] != "" {
				vm.ExecuteRequest("vm.music.stop")
				_, err := vm.ExecuteRequest("vm.music.play", settings.MusicTrack[i], settings.MusicLeadin[i], settings.MusicFadein[i])
				settings.MusicTrack[i] = ""
				if err != nil {
					return
				}
			}

		}
		time.Sleep(time.Millisecond * 100)
	}
}

func (this *Producer) RebootService() {
	for {
		time.Sleep(500 * time.Millisecond)
		if settings.CleanBootRequested {
			settings.CleanBootRequested = false
			settings.MicroPakPath = ""
			for i := 0; i < settings.NUMSLOTS; i++ {
				settings.FirstBoot[i] = true
				settings.PureBootVolume[i] = ""
				settings.PureBootVolume2[i] = ""
				settings.PureBootSmartVolume[i] = ""
				if this.VM[i] == nil {
					continue
				}
				if i == 0 {
					go this.RebootVM(i)
				} else {
					this.VM[i].Stop()
					this.VM[i] = nil
				}

			}
		} else {
			for i := 0; i < settings.NUMSLOTS; i++ {
				if this.AddressSpace.IntGetSlotRestart(i) {
					this.AddressSpace.IntSetSlotRestart(i, false)
					go func(slot int) {
						// if this.VM[slot] != nil {
						// 	this.VM[slot].Stop()
						// }
						this.RebootVM(slot)
					}(i)
				}
			}
		}

	}
}

func (this *Producer) TerminateSlot(i int) {

	if this.ThreadActive[i] == 0 {
		return
	}

	// slot is active...
	this.ThreadActive[i] = 0
	<-this.Quit[i]

	// when we get here it is no more...

}

func (this *Producer) RestartSlot(slotid int) {
	this.AddressSpace.IntSetSlotRestart(slotid, true)
}

func (this *Producer) ResetSlot(slotid int) {

	ent := this.VM[slotid].GetInterpreter()
	for ent != nil && ent.GetChild() != nil {
		ent = ent.GetChild()
	}

	for ent != nil {

		// Stop recording and halt CPU
		ent.StopRecording()
		ent.Halt()
		ent.Halt6502(0)

		ent.SetChild(nil)
		ent = ent.GetParent()

	}

	ent = this.VM[slotid].GetInterpreter()
	if ent != nil {
		w, uw := apple2helpers.GetCPU(ent).HasUserWarp()
		fmt.Printf("User warp = %v, speed mult = %f\n", w, uw)
		apple2helpers.TrashCPU(ent)

		if w {
			apple2helpers.GetCPU(ent).SetWarpUser(uw)
		}

		ent.SetSpec("")
		ent.LoadSpec(settings.DefaultProfile)
		ent.PeripheralReset()

		//this.CreateInterpreterInSlot(slotid, "run /boot/boot", true)
	}

}

// func (this *Producer) CreateInterpreterInSlot(i int, bootstrap string, resetonly bool) {
// 	if !resetonly {
// 		this.CreateInterpreter(i, "main", shell.NewDialectShell(), settings.SpecFile[i], GetUUID())
// 	}

// 	if !strings.HasPrefix(settings.SpecFile[i], "apple2") {
// 		return
// 	}

// 	if i > 0 {
// 		this.VM[i].GetInterpreter().ParseImm("display \"Type cat for files, or logo, fp or int\"")
// 		this.VM[i].GetInterpreter().SetDisabled(true)
// 		this.VM[i].GetInterpreter().SaveCPOS()
// 		this.VM[i].GetInterpreter().SetNeedsPrompt(true)
// 		if settings.IsRemInt {
// 			this.AddressSpace.KeyBufferAdd(i, 13)
// 			this.AddressSpace.KeyBufferAdd(i, 13)
// 			this.VM[i].GetInterpreter().SetDisabled(false)
// 		}
// 	} else {
// 		if settings.IsRemInt {
// 			s8webclient.CONN.Login("remint", "remint")
// 			this.VM[0].GetInterpreter().Parse("")
// 			this.VM[i].GetInterpreter().SetDisabled(false)
// 			this.VM[i].GetInterpreter().SaveCPOS()
// 			this.VM[i].GetInterpreter().SetNeedsPrompt(true)
// 			this.AddressSpace.KeyBufferAdd(i, 13)
// 			this.AddressSpace.KeyBufferAdd(i, 13)
// 		} else {
// 			this.VM[0].GetInterpreter().Parse(bootstrap)
// 			this.VM[i].GetInterpreter().SetDisabled(false)
// 			this.VM[i].GetInterpreter().SetBreakable(false)
// 			this.VM[i].GetInterpreter().StopTheWorld() // stop here because we need the update check to complete
// 			this.VM[i].GetInterpreter().SaveCPOS()
// 			this.VM[i].GetInterpreter().SetNeedsPrompt(true)
// 		}

// 	}
// }

func (this *Producer) GetNumInterpreters() int {
	return memory.OCTALYZER_NUM_INTERPRETERS
}

func NewProducerWithParams(r *memory.MemoryMap, bootstrap string, exec string, dia interfaces.Dialecter, specfile string) *Producer {
	this := &Producer{}
	//	this.Global = *types.NewVarMap(-1, nil)
	//~ this.VM = make([]interfaces.Interpretable, memory.OCTALYZER_NUM_INTERPRETERS)
	this.MasterLayerPos = make([]types.LayerPosMod, memory.OCTALYZER_NUM_INTERPRETERS)
	this.ForceUpdate = make([]bool, memory.OCTALYZER_NUM_INTERPRETERS)
	if r == nil {
		r = memory.NewMemoryMap()
	}
	this.AddressSpace = r
	//this.DiaAS = applesoft.NewDialectApplesoft()
	//fmt.Println("Just before create interpreter")
	for i := 0; i < memory.OCTALYZER_NUM_INTERPRETERS; i++ {
		this.CreateVM(i, nil, nil)
		//this.CreateInterpreter(i, "main", shell.NewDialectShell(), settings.DefaultProfile, GetUUID())
		if i > 0 {
			this.VM[i].GetInterpreter().Parse("lang fp")
			this.VM[i].GetInterpreter().SetDisabled(true)
		} else {
			this.VM[0].GetInterpreter().Parse(bootstrap)
			this.VM[i].GetInterpreter().SetDisabled(false)
		}
		z := i
		go this.Executor(z)
		time.Sleep(50 * time.Millisecond)
	}
	//fmt.Println("Just after create interpreter")
	this.LastExecute = time.Now().Unix()
	this.History = make([]runestring.RuneString, 0)

	if exec != "" {
		this.VM[0].GetInterpreter().SetPasteBuffer(runestring.Cast(exec + "\r"))
	}

	return this
}

func (this *Producer) CallEndRemotes() {
	this.EndRemotesNeeded = true
}

func (this *Producer) Run() {

	for {

		if this.Paused {
			time.Sleep(1 * time.Millisecond)
			continue
		}

		// reboot
		if this.RebootNeeded {
			//this.Reboot()
			this.RebootNeeded = false
		}

		if settings.SystemType != "nox" {
			files.System = (s8webclient.CONN.Session == "")
		} else {
			files.System = true
		}

		// for i := 0; i < memory.OCTALYZER_NUM_INTERPRETERS; i++ {
		// 	if !this.ThreadActive[i] {
		// 		// Start slotid
		// 		z := i
		// 		ent := this.GetInterpreters()[i]
		// 		ent.CheckProfile(true)
		// 		ent.ScreenReset()
		// 		ent.SetChild(nil)
		// 		ent.Zero(true)
		// 		ent.Bootstrap("fp", false)
		// 		go this.Executor(z)
		// 	}
		// }

		time.Sleep(500 * time.Millisecond)

		if this.EndRemotesNeeded {
			this.EndRemotes()
			this.EndRemotesNeeded = false
		}

	}

}

func (this *Producer) GetMinWait() (int, time.Duration) {

	now := time.Now()

	var r time.Duration = 1 * time.Hour
	var index int = -1
	var count int
	for i, vm := range this.VM {

		if vm == nil {
			continue
		}
		ent := vm.GetInterpreter()

		if ent == nil || ent.IsDisabled() || ent.IsWaitingForWorld() {
			continue
		}
		count++
		v := ent.GetWaitUntil().Sub(now)
		if (v > 0) && (v < r) {
			r = v
			index = i
		}
	}

	if r == time.Hour {
		if count > 0 {
			r = 0
		} else {
			r = time.Second
		}
	}

	return index, r
}
