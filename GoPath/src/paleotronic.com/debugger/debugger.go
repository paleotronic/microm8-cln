package debugger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"paleotronic.com/core/hardware/apple2"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/hardware/cpu/mos6502"
	"paleotronic.com/core/hardware/servicebus"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/settings"
	"paleotronic.com/debugger/debugtypes"
	"paleotronic.com/files"
	"paleotronic.com/log"
	"paleotronic.com/octalyzer/backend"
)

const cpuHistory = 10

type DebuggerConfig struct {
	CPUInstructionBacklog   int // number of instructions in backlog
	CPUInstructionLookahead int
	CPUStateInterval        int  // ms between CPU updates in run mode
	ScreenRefreshMS         int  // ms between Screen image buffers
	CPURecordTiming         int  // number of timing points per second
	FullCPURecord           bool // record every instruction in between normal syncs
	BreakOnIllegalOp        bool
	BreakOnBRK              bool
	SaveSettingsOnSubmit    bool
}

type DebuggerState struct {
	Config        *DebuggerConfig
	CPU           *debugtypes.CPUState
	SWMEM         []debugtypes.SoftSwitchInfo
	SWVID         []debugtypes.SoftSwitchInfo
	Breakpoints   []*debugtypes.CPUBreakpoint
	LastPC        []int // previous 5 pc...
	LastBPMessage string
}

type Debugger struct {
	DebuggerState
	events   []*servicebus.ServiceBusRequest
	slotid   int
	running  bool
	attached bool
	changed  bool
	//ent              interfaces.Interpretable
	socket           *websocket.Conn
	lastCPUState     time.Time
	lastScreenState  time.Time
	sendQueue        chan *debugtypes.WebSocketMessage
	senderRunning    bool
	LastScreenBuffer []byte
	sync.Mutex
}

const dbgConfig = "/local/settings/debugger.json"

func NewDebugger() *Debugger {

	dc := &DebuggerConfig{
		CPUInstructionBacklog:   10,
		CPUInstructionLookahead: 10,
		CPURecordTiming:         10,
		CPUStateInterval:        1000,
		ScreenRefreshMS:         1000,
		FullCPURecord:           false,
		BreakOnBRK:              false,
		BreakOnIllegalOp:        true,
	}

	d := &Debugger{
		DebuggerState: DebuggerState{
			Config:      dc,
			Breakpoints: make([]*debugtypes.CPUBreakpoint, 0, 16),
			LastPC:      make([]int, dc.CPUInstructionBacklog),
		},
		events:    make([]*servicebus.ServiceBusRequest, 0, 16),
		slotid:    -1,
		sendQueue: make(chan *debugtypes.WebSocketMessage, 1024),
	}

	d.LoadConfig(dbgConfig)

	return d
}

func (d *Debugger) SaveConfig(path string) error {
	j, e := json.Marshal(d.Config)
	if e != nil {
		return e
	}
	return files.WriteBytesViaProvider(files.GetPath(path), files.GetFilename(path), j)
}

func (d *Debugger) LoadConfig(path string) error {
	j, e := files.ReadBytesViaProvider(files.GetPath(path), files.GetFilename(path))
	if e != nil {
		return e
	}
	return json.Unmarshal(j.Content, d.Config)
}

func (d *Debugger) ent() interfaces.Interpretable {
	return backend.ProducerMain.GetInterpreter(d.slotid)
}

func (d *Debugger) ResetCounters() {
	for _, b := range d.Breakpoints {
		b.Counter = 0
	}
}

func (d *Debugger) UpdateConfig() {
	if len(d.LastPC) != d.Config.CPUInstructionBacklog {
		n := make([]int, d.Config.CPUInstructionBacklog)
		for i := 0; i < len(d.LastPC); i++ {
			index := len(d.LastPC) - 1 - i
			v := d.LastPC[index]
			nindex := len(n) - 1 - i
			if nindex >= 0 {
				n[nindex] = v
			}
		}
		d.LastPC = n
	}
	settings.CPURecordTimingPoints = d.Config.CPURecordTiming
}

func (d *Debugger) CheckBreakpoints(f func(b *debugtypes.CPUBreakpoint), WA, WV, RA, RV int, memonly bool) {
	cpu := apple2helpers.GetCPU(d.ent())

	main, aux := -1, -1
	if WA != -1 {
		main, aux = getMemoryBanks(WA, memory.MA_WRITE)
	} else if RA != -1 {
		main, aux = getMemoryBanks(WA, memory.MA_READ)
	}

	for idx, v := range d.Breakpoints {

		if memonly && v.ReadAddress == nil && v.WriteAddress == nil && v.ReadValue == nil && v.WriteValue == nil {
			continue
		}

		if v.ShouldBreak(
			cpu.PC,
			cpu.A,
			cpu.X,
			cpu.Y,
			cpu.SP,
			cpu.P,
			WA,
			WV,
			RA,
			RV,
			main,
			aux,
		) {
			if v.Ephemeral {
				v.Disabled = true
				v.Ephemeral = false
				d.Breakpoints[idx] = v
			}
			if v.Action == nil {
				// if cpu.RunState != mos6502.CrsSingleStep && cpu.RunState != mos6502.CrsStepOut && cpu.RunState != mos6502.CrsStepOver {
				f(v)
				// }
			} else {
				switch v.Action.Type {
				case debugtypes.BABreak:
					f(v)
				case debugtypes.BAText:
					msg, ok := v.Action.Arg0.(string)
					if ok {
						//d.SendMessage("debug-message", msg, true)
						d.LastBPMessage = msg
					}
				case debugtypes.BAChime:
					apple2helpers.Beep(d.ent())
				case debugtypes.BATraceOff:
					d.Trace("off")
				case debugtypes.BATraceOn:
					d.Trace("on")
				case debugtypes.BAJump:
					addr, ok := v.Action.Arg0.(int)
					if ok {
						cpu := apple2helpers.GetCPU(DebuggerInstance.ent())

						caddr := cpu.PC
						info := cpu.Opref[cpu.FetchByteAddr(caddr)]

						if info != nil {
							readdr := cpu.PC + info.GetBytes()

							//log.Printf("setting a return address of %.4x", readdr)

							cpu.Push(readdr / 256)
							cpu.Push(readdr % 256)

							cpu.PC = addr
						}
					}
				case debugtypes.BACount:
					v.Counter++
					d.Breakpoints[idx] = v

				case debugtypes.BASpeed:
					speed, ok := v.Action.Arg0.(float64)
					if ok {
						cpu := apple2helpers.GetCPU(DebuggerInstance.ent())
						cpu.SetWarpUser(speed)
						cpu.CalcTiming()
						log.Printf("Setting CPU warp to %f", speed)
						d.SendMessage(
							"set-warp-response",
							"Warp set",
							true,
						)
					}

				case debugtypes.BALogToTrace:
					cpu := apple2helpers.GetCPU(DebuggerInstance.ent())
					var msg string
					var ok bool
					if v.Action.Arg0 != nil {
						msg, ok = v.Action.Arg0.(string)
						if !ok {
							msg = fmt.Sprintf("breakpoint reached at $%4x", cpu.PC)
						}
					} else {
						msg = fmt.Sprintf("breakpoint reached at $%4x", cpu.PC)
					}
					cpu.LogToTrace(
						fmt.Sprintf(
							"debugger: %s",
							msg,
						),
					)

				case debugtypes.BARecordOn:

					if d.ent().IsPlayingVideo() {
						d.ent().BreakIntoVideo()
					}
					if !d.ent().IsRecordingDiscVideo() {
						d.ent().StopRecording()
						d.ent().RecordToggle(d.Config.FullCPURecord)
					}

				case debugtypes.BARecordOff:
					if d.ent().IsPlayingVideo() {
						p := d.ent().GetPlayer()
						p.SetNoResume(true)
						d.ent().BreakIntoVideo()
					}
					d.ent().StopRecording()
					d.SendMessage(
						"live-rewind-response",
						&debugtypes.LiveRewindState{
							CanBack:    false,
							CanForward: false,
							CanResume:  false,
							Enabled:    false,
						},
						true,
					)
				}
			}
		}
	}
}

func (d *Debugger) GetBreakpoints() []*debugtypes.CPUBreakpoint {
	return d.Breakpoints
}

func (d *Debugger) AddBreakpoint(b *debugtypes.CPUBreakpoint) {
	d.Breakpoints = append(d.Breakpoints, b)
}

func (d *Debugger) SetBreakpointCounter(idx int, value int) {
	d.Breakpoints[idx].Counter = value
}

func (d *Debugger) UpdateBreakpoint(idx int, b *debugtypes.CPUBreakpoint) {
	b.Disabled = d.Breakpoints[idx].Disabled
	b.Counter = d.Breakpoints[idx].Counter
	d.Breakpoints[idx] = b
}

func (d *Debugger) RemoveBreakpoint(idx int) {
	if idx < 0 || idx >= len(d.Breakpoints) {
		return
	}
	d.Breakpoints = append(d.Breakpoints[:idx], d.Breakpoints[idx+1:]...)
}

func (d *Debugger) DisableBreakpoint(idx int) {
	if idx < 0 || idx >= len(d.Breakpoints) {
		return
	}
	d.Breakpoints[idx].Disabled = true
}

func (d *Debugger) EnableBreakpoint(idx int) {
	if idx < 0 || idx >= len(d.Breakpoints) {
		return
	}
	d.Breakpoints[idx].Disabled = false
}

func (d *Debugger) Detach() {
	if d.slotid != -1 {
		cpu := apple2helpers.GetCPU(d.ent())
		servicebus.Unsubscribe(d.slotid, cpu)
		servicebus.Unsubscribe(d.slotid, d)
		log.Printf("Unsubscribing from events for slot %d", d.slotid)
		d.slotid = -1
	}
	d.attached = false
	close(d.sendQueue)
	for d.senderRunning {
		time.Sleep(1 * time.Millisecond)
	}
	d.sendQueue = make(chan *debugtypes.WebSocketMessage, 1024)
	settings.DebuggerActiveSlot = -1
}

func (d *Debugger) AttachSlot(slotid int) {
	for backend.ProducerMain == nil {
		time.Sleep(time.Millisecond)
	}
	d.Detach()
	d.slotid = slotid

	//d.ent() = backend.ProducerMain.GetInterpreter(d.slotid)
	cpu := apple2helpers.GetCPU(d.ent())
	servicebus.Subscribe(
		slotid,
		servicebus.CPUState,
		d,
	)
	servicebus.Subscribe(
		slotid,
		servicebus.MemorySoftSwitchState,
		d,
	)
	servicebus.Subscribe(
		slotid,
		servicebus.VideoSoftSwitchState,
		d,
	)
	servicebus.Subscribe(
		slotid,
		servicebus.CPUReadMem,
		d,
	)
	servicebus.Subscribe(
		slotid,
		servicebus.CPUWriteMem,
		d,
	)
	servicebus.Subscribe(
		slotid,
		servicebus.LiveRewindStateUpdate,
		d,
	)
	servicebus.Subscribe(slotid, servicebus.CPUControl, cpu)
	log.Printf("Subcribing to debug events from slot %d", d.slotid)
	d.attached = true
	go d.sender()
	go d.ScreenServer()
	//d.ClearAllBreakpoints()
	if !d.ent().IsPlayingVideo() {
		d.PauseCPU()
		if d.ent().IsRecordingDiscVideo() {
			time.AfterFunc(2*time.Second, func() {
				d.SendMessage(
					"live-rewind-response",
					liveRewindState(false),
					true,
				)
			})
		}
	} else {
		d.ent().GetPlayer().SetTimeShift(0)
		time.AfterFunc(2*time.Second, func() {
			d.SendMessage(
				"live-rewind-response",
				liveRewindState(false),
				true,
			)
		})
	}
	d.RequestSwitchStates()
	//d.SetBreakpoint(0xc28e)
	settings.DebuggerActiveSlot = slotid
}

func (d *Debugger) sender() {
	d.senderRunning = true
	defer func() {
		d.senderRunning = false
	}()
	for d.attached {
		select {
		case msg := <-d.sendQueue:
			// send it
			d.sendMessage(msg)
		}
	}
}

func CopyFrame(src *image.RGBA) *image.RGBA {
	f := image.NewRGBA(image.Rect(0, 0, 560, 384))
	var r, g, b, a uint8
	var line = 560 * 4
	for y := 0; y < 192; y++ {
		var yBaseSrc = y * line
		var yBaseTgt = y * line * 2
		var yBaseTgt2 = yBaseTgt + line
		for x := 0; x < 560; x++ {
			var xBase = x * 4

			r = uint8(float32(src.Pix[xBase+yBaseSrc+0]) * 1)
			g = uint8(float32(src.Pix[xBase+yBaseSrc+1]) * 1)
			b = uint8(float32(src.Pix[xBase+yBaseSrc+2]) * 1)
			a = src.Pix[xBase+yBaseSrc+3]

			f.Pix[xBase+yBaseTgt+0] = r
			f.Pix[xBase+yBaseTgt+1] = g
			f.Pix[xBase+yBaseTgt+2] = b
			f.Pix[xBase+yBaseTgt+3] = a

			f.Pix[xBase+yBaseTgt2+0] = uint8(float32(r) * settings.ScanLineIntensity)
			f.Pix[xBase+yBaseTgt2+1] = uint8(float32(g) * settings.ScanLineIntensity)
			f.Pix[xBase+yBaseTgt2+2] = uint8(float32(b) * settings.ScanLineIntensity)
			f.Pix[xBase+yBaseTgt2+3] = a
		}
	}
	return f
}

func (d *Debugger) UpdateScreenshot() {

	if settings.UnifiedRender[d.slotid] {
		if settings.UnifiedRenderChanged[d.slotid] && settings.UnifiedRenderFrame[d.slotid] != nil {
			b := bytes.NewBuffer(nil)
			if err := png.Encode(b, CopyFrame(settings.UnifiedRenderFrame[d.slotid])); err == nil {
				settings.ScreenShotJPEGData = b.Bytes()
			}
		}
		return
	}

	settings.ScreenShotNeeded = true
	s := time.Now()
	for settings.ScreenShotNeeded && time.Since(s) < time.Millisecond*250 {
		time.Sleep(1 * time.Millisecond)
	}
}

func (d *Debugger) RequestSwitchStates() {
	mr, ok := d.ent().GetMemoryMap().InterpreterMappableAtAddress(d.ent().GetMemIndex(), 0xc000)
	if ok {
		d.SWVID = mr.(*apple2.Apple2IOChip).GetVideoSwitchInfo()
		d.SWMEM = mr.(*apple2.Apple2IOChip).GetMemorySwitchInfo()
		d.SendMessage("switch-video-response",
			&debugtypes.VideoSoftSwitches{
				Switches: append(d.SWMEM, d.SWVID...),
			}, true)
	}
}

func (d *Debugger) PauseCPU() {
	servicebus.SendServiceBusMessage(
		d.slotid,
		servicebus.CPUControl,
		&servicebus.CPUControlData{
			Action: "pause",
		},
	)
}

func (d *Debugger) StepCPU() {
	servicebus.SendServiceBusMessage(
		d.slotid,
		servicebus.CPUControl,
		&servicebus.CPUControlData{
			Action: "step",
		},
	)
}

func (d *Debugger) StepCPUOut() {
	servicebus.SendServiceBusMessage(
		d.slotid,
		servicebus.CPUControl,
		&servicebus.CPUControlData{
			Action: "step-out",
		},
	)
}

func (d *Debugger) StepCPUOver() {
	cpu := apple2helpers.GetCPU(d.ent())
	servicebus.SendServiceBusMessage(
		d.slotid,
		servicebus.CPUControl,
		&servicebus.CPUControlData{
			Action: "step-over",
			Data: map[string]interface{}{
				"sp-level": cpu.SP,
				"pc":       cpu.PC,
			},
		},
	)
}

func (d *Debugger) WriteBlob(addr int, data []byte) {
	cpu := apple2helpers.GetCPU(d.ent())
	for i, v := range data {
		cpu.StoreByteAddr((addr+i)%65536, int(v))
	}
}

func (d *Debugger) ReadBlob(addr int, count int) []byte {
	out := make([]byte, count)
	cpu := apple2helpers.GetCPU(d.ent())
	for i, _ := range out {
		out[i] = byte(cpu.FetchByteAddr((addr + i) % 65536))
	}
	return out
}

func (d *Debugger) SetVal(name string, value int) {
	//rlog.Printf("Setting name = %s, value = %d", name, value)
	cpu := apple2helpers.GetCPU(d.ent())
	switch name {
	case "6502.a":
		cpu.A = value
	case "6502.x":
		cpu.X = value
	case "6502.y":
		cpu.Y = value
	case "6502.sp":
		cpu.SP = value
	case "6502.pc":
		cpu.PC = value
	case "6502.p":
		cpu.P = value
	case "save-on-submit":
		d.Config.SaveSettingsOnSubmit = (value != 0)
		if d.Config.SaveSettingsOnSubmit {
			d.SaveConfig(dbgConfig)
		}
	case "cpu-break-ill":
		d.Config.BreakOnIllegalOp = (value != 0)
		if d.Config.SaveSettingsOnSubmit {
			d.SaveConfig(dbgConfig)
		}
	case "cpu-break-brk":
		d.Config.BreakOnBRK = (value != 0)
		if d.Config.SaveSettingsOnSubmit {
			d.SaveConfig(dbgConfig)
		}
	case "cpu-full-record":
		d.Config.FullCPURecord = (value != 0)
		settings.DebugFullCPURecord = d.Config.FullCPURecord
		if d.Config.SaveSettingsOnSubmit {
			d.SaveConfig(dbgConfig)
		}
	case "cpu-update-ms":
		d.Config.CPUStateInterval = value
		if d.Config.SaveSettingsOnSubmit {
			d.SaveConfig(dbgConfig)
		}
	case "cpu-backlog-lines":
		d.Config.CPUInstructionBacklog = value
		d.UpdateConfig()
		if d.Config.SaveSettingsOnSubmit {
			d.SaveConfig(dbgConfig)
		}
	case "cpu-lookahead-lines":
		d.Config.CPUInstructionLookahead = value
		d.UpdateConfig()
	case "cpu-record-timing":
		if value == 0 {
			value = 10
		}
		d.Config.CPURecordTiming = value
		d.UpdateConfig()
		if d.Config.SaveSettingsOnSubmit {
			d.SaveConfig(dbgConfig)
		}
	case "screen-refresh-ms":
		d.Config.ScreenRefreshMS = value
		if d.Config.SaveSettingsOnSubmit {
			d.SaveConfig(dbgConfig)
		}
	}
}

func (d *Debugger) Trace(verb string) string {
	cpu := apple2helpers.GetCPU(d.ent())
	cpu.StopTrace()
	if verb == "off" {
		return "Trace disabled."
	}
	if cpu.IsTracing() {
		return "Already tracing..."
	}
	fn := fmt.Sprintf("microm8-cpu-trace-%d.log", time.Now().Unix())
	path := files.GetUserPath(files.BASEDIR, []string{"traces"})
	os.MkdirAll(path, 0755)

	cpu.StartTrace(path + "/" + fn)
	return "Tracing to " + path + "/" + fn
}

func (d *Debugger) ToggleCPUFlag(name string) {
	cpu := apple2helpers.GetCPU(d.ent())
	switch name {
	case "N":
		cpu.P ^= mos6502.F_N
	case "V":
		cpu.P ^= mos6502.F_V
	case "I":
		cpu.P ^= mos6502.F_I
	case "D":
		cpu.P ^= mos6502.F_D
	case "C":
		cpu.P ^= mos6502.F_C
	case "B":
		cpu.P ^= mos6502.F_B
	case "Z":
		cpu.P ^= mos6502.F_Z
	}
	d.CPU.P = cpu.P
	d.SendMessage(
		"cpu-flag-response",
		d.CPU,
		true,
	)
}

func (d *Debugger) ToggleSoftSwitch(name string) {
	mr, ok := d.ent().GetMemoryMap().InterpreterMappableAtAddress(d.ent().GetMemIndex(), 0xc000)
	if ok {
		mr.(*apple2.Apple2IOChip).ToggleMemorySwitch(name)
		mr.(*apple2.Apple2IOChip).ToggleVideoSwitch(name)
	}
}

func (d *Debugger) GetSlot() int {
	return d.slotid
}

func (d *Debugger) IsAttached() bool {
	return d.attached
}

func (d *Debugger) SendKey(keycode int) {
	mm := d.ent().GetMemoryMap()
	index := d.ent().GetMemIndex()
	mm.KeyBufferAdd(index, uint64(keycode))
}

func (d *Debugger) ContinueCPUOut() {
	servicebus.SendServiceBusMessage(
		d.slotid,
		servicebus.CPUControl,
		&servicebus.CPUControlData{
			Action: "continue-out",
		},
	)
}

func (d *Debugger) ContinueCPU() {
	servicebus.SendServiceBusMessage(
		d.slotid,
		servicebus.CPUControl,
		&servicebus.CPUControlData{
			Action: "continue",
		},
	)
}

func (d *Debugger) ScreenServer() {
	var lastSc = time.Now()
	d.UpdateScreenshot()
	d.SendMessage(
		"screen-update-response",
		"/api/debug/screen/",
		true,
	)
	for d.attached {
		cpu := apple2helpers.GetCPU(d.ent())
		if (cpu.RunState == mos6502.CrsFreeRun || d.ent().IsPlayingVideo()) && time.Since(lastSc) > time.Duration(d.Config.ScreenRefreshMS)*time.Millisecond {
			d.UpdateScreenshot()
			d.SendMessage(
				"screen-update-response",
				"/api/debug/screen/",
				true,
			)
			lastSc = time.Now()
		} else if settings.UnifiedRender[d.slotid] && settings.UnifiedRenderChanged[d.slotid] && settings.UnifiedRenderFrame[d.slotid] != nil {
			d.UpdateScreenshot()
			d.SendMessage(
				"screen-update-response",
				"/api/debug/screen/",
				true,
			)
			lastSc = time.Now()
		}
		time.Sleep(time.Millisecond)
	}
}

func (d *Debugger) HandleServiceBusRequest(r *servicebus.ServiceBusRequest) (*servicebus.ServiceBusResponse, bool) {

	switch r.Type {
	case servicebus.LiveRewindStateUpdate:
		rinfo := r.Payload.(*debugtypes.LiveRewindState)
		//fmt.Printf("Got live-rewind status update: %+v\n", rinfo)
		d.SendMessage("live-rewind-response",
			rinfo,
			true,
		)
	case servicebus.CPUReadMem:
		rinfo := r.Payload.(*debugtypes.CPUMemoryRead)
		cpu := apple2helpers.GetCPU(d.ent())
		d.CheckBreakpoints(func(b *debugtypes.CPUBreakpoint) {
			cpu.RunState = mos6502.CrsPaused
			d.SendMessage("pause-response",
				d.CPU,
				true,
			)
			d.SendMessage("break-response", fmt.Sprintf("Stopped at $%.4x (%s)", cpu.PC, b.String()), true)
		}, -1, -1, rinfo.Address, rinfo.Value, true)
	case servicebus.CPUWriteMem:
		winfo := r.Payload.(*debugtypes.CPUMemoryWrite)
		cpu := apple2helpers.GetCPU(d.ent())
		d.CheckBreakpoints(func(b *debugtypes.CPUBreakpoint) {
			cpu.RunState = mos6502.CrsPaused
			d.SendMessage("pause-response",
				d.CPU,
				true,
			)
			d.SendMessage("break-response", fmt.Sprintf("Stopped at $%.4x (%s)", cpu.PC, b.String()), true)
		}, winfo.Address, winfo.Value, -1, -1, true)
		//d.SendMessage("write-response", winfo, true)
	case servicebus.MemorySoftSwitchState:
		//log.Printf("Sending switch states: %+v", r.Payload.([]debugtypes.SoftSwitchInfo))
		d.SWMEM = r.Payload.([]debugtypes.SoftSwitchInfo)
		d.SendMessage("switch-video-response",
			&debugtypes.VideoSoftSwitches{
				Switches: append(d.SWMEM, d.SWVID...),
			}, true)
	case servicebus.VideoSoftSwitchState:
		d.SWVID = r.Payload.([]debugtypes.SoftSwitchInfo)
		d.SendMessage("switch-video-response",
			&debugtypes.VideoSoftSwitches{
				Switches: append(d.SWMEM, d.SWVID...),
			}, true)
	case servicebus.CPUState:

		newstate := r.Payload.(*debugtypes.CPUState)

		// store last pc infos -- if not a recording
		//if !newstate.IsRecording {
		//if d.LastPC[len(d.LastPC)-1] != d.CPU.PC {
		for i := 1; i < len(d.LastPC); i++ {
			d.LastPC[i-1] = d.LastPC[i]
		}
		if d.CPU != nil {
			d.LastPC[len(d.LastPC)-1] = d.CPU.PC
		}
		//}
		//}

		d.CPU = newstate
		d.changed = true
		cpu := apple2helpers.GetCPU(d.ent())

		timeForRedraw := (time.Since(d.lastCPUState) > time.Duration(d.Config.CPUStateInterval)*time.Millisecond)
		cpuRunning := cpu.RunState == mos6502.CrsFreeRun
		playBackRunning := d.ent().IsPlayingVideo() && d.ent().GetPlayer() != nil && d.ent().GetPlayer().GetTimeShift() != 0

		//rlog.Printf("Handling CPU state: running=%v, playing=%v, redraw time=%v", cpuRunning, playBackRunning, timeForRedraw)

		if (newstate.ForceUpdate) || (cpuRunning && timeForRedraw) || (playBackRunning && timeForRedraw) || (!cpuRunning && !playBackRunning) {

			//rlog.Println("Sending state to frontend... ")

			if cpu.RunState == mos6502.CrsPaused && !d.ent().IsPlayingVideo() {
				d.SendMessage("pause-response", d.CPU, true)
			} else {
				d.SendMessage("state-response", d.CPU, true)
			}

			d.lastCPUState = time.Now()
			address := cpu.PC

			count := d.Config.CPUInstructionBacklog + d.Config.CPUInstructionLookahead + 1
			instr := make([]debugtypes.CPUInstructionDecode, int(count))

			if newstate.IsRecording {
				p := d.ent().GetPlayer()
				// backlog
				behind := p.GetLastNSyncs(d.Config.CPUInstructionBacklog, address)
				for i, v := range behind {
					if v == -1 {
						v = 0
					}
					code, desc, cycles := cpu.DecodeInstruction(v)
					instr[i].Address = int(v)
					instr[i].Bytes = code
					instr[i].Instruction = desc
					instr[i].Cycles = cycles
					instr[i].Historic = true
				}
				// current
				code, desc, cycles := cpu.DecodeInstruction(address)
				j := d.Config.CPUInstructionBacklog
				instr[j].Address = address
				instr[j].Bytes = code
				instr[j].Instruction = desc
				instr[j].Cycles = cycles
				instr[j].Historic = false
				// lookahead
				ahead := p.GetNextNSyncs(d.Config.CPUInstructionLookahead, address)
				for i, v := range ahead {
					if v == -1 {
						v = 0
					}
					code, desc, cycles := cpu.DecodeInstruction(v)
					instr[i+j+1].Address = int(v)
					instr[i+j+1].Bytes = code
					instr[i+j+1].Instruction = desc
					instr[i+j+1].Cycles = cycles
					instr[i+j+1].Historic = true
				}
			} else {
				for i, _ := range instr {
					if i < len(d.LastPC) {
						code, desc, cycles := cpu.DecodeInstruction(int(d.LastPC[i]))
						instr[i].Address = int(DebuggerInstance.LastPC[i])
						instr[i].Bytes = code
						instr[i].Instruction = desc
						instr[i].Cycles = cycles
						instr[i].Historic = true
					} else {
						code, desc, cycles := cpu.DecodeInstruction(int(address))
						instr[i].Address = int(address)
						instr[i].Bytes = code
						instr[i].Instruction = desc
						instr[i].Cycles = cycles
						address += len(code) % 65536
					}
				}
			}
			d.SendMessage("decode-response",
				&debugtypes.CPUInstructions{
					Instructions: instr,
				},
				true,
			)
		}

		// before checking breakpoints, check for 0x00
		b := cpu.FetchByteAddr(cpu.PC)

		if b == 0x00 && d.Config.BreakOnBRK {
			if d.ent().IsPlayingVideo() {
				d.ent().GetPlayer().SetTimeShift(0) // pause playback
			} else {
				cpu.RunState = mos6502.CrsPaused
			}
			d.SendMessage("pause-response", d.CPU, true)
			d.SendMessage("break-response", fmt.Sprintf("CPU Paused at $%.4x (Settings/Break on BRK)", cpu.PC), true)
		} else if info := cpu.Opref[b]; (info == nil || info.Description[0] >= 97) && d.Config.BreakOnIllegalOp {
			log.Printf("break at %.4x for illegal opcode %.2x", cpu.PC, b)
			if d.ent().IsPlayingVideo() {
				d.ent().GetPlayer().SetTimeShift(0) // pause playback
			} else {
				cpu.RunState = mos6502.CrsPaused
			}
			d.SendMessage("pause-response", d.CPU, true)
			d.SendMessage("break-response", fmt.Sprintf("CPU Paused at $%.4x (Settings/Break on Illegal opcode: $%.2x)", cpu.PC, b), true)
		} else {

			d.CheckBreakpoints(func(b *debugtypes.CPUBreakpoint) {
				if d.ent().IsPlayingVideo() {
					d.ent().GetPlayer().SetTimeShift(0) // pause playback
					d.SendMessage("pause-response", d.CPU, true)
					d.SendMessage("break-response", fmt.Sprintf("Recording Paused by breakpoint at $%.4x (%s)", cpu.PC, b.String()), true)
				} else {
					cpu.RunState = mos6502.CrsPaused
					d.SendMessage("pause-response", d.CPU, true)
					d.SendMessage("break-response", fmt.Sprintf("CPU Paused by breakpoint at $%.4x (%s)", cpu.PC, b.String()), true)
				}
			}, -1, -1, -1, -1, false)

		}

	}
	return &servicebus.ServiceBusResponse{
		Payload: &debugtypes.CPUAction{
			Type: debugtypes.CatResume,
		},
	}, true
}

func (d *Debugger) InjectServiceBusRequest(r *servicebus.ServiceBusRequest) {
	log.Printf("Injecting ServiceBus request: %+v", r)
	d.Lock()
	defer d.Unlock()
	if d.events == nil {
		d.events = make([]*servicebus.ServiceBusRequest, 0, 16)
	}
	d.events = append(d.events, r)
}

func (d *Debugger) HandleServiceBusInjection(handler func(r *servicebus.ServiceBusRequest) (*servicebus.ServiceBusResponse, bool)) {
	if d.events == nil || len(d.events) == 0 {
		return
	}
	d.Lock()
	defer d.Unlock()
	for _, r := range d.events {
		if handler != nil {
			handler(r)
		}
	}
	d.events = make([]*servicebus.ServiceBusRequest, 0, 16)
}

func (d *Debugger) ServiceBusProcessPending() {
	d.HandleServiceBusInjection(d.HandleServiceBusRequest)
}

func (d *Debugger) Stop() {
	if d.running {
		d.running = false
		// wait for goroutine to finish
		for d.attached {
			time.Sleep(time.Millisecond)
		}
	}
}

func (d *Debugger) Run(slotid int, ent interfaces.Interpretable) {
	d.Stop()
	d.AttachSlot(slotid)
	d.running = true
	for d.running {
		d.ServiceBusProcessPending()
		time.Sleep(100 * time.Millisecond)
	}
	d.Detach()
}

func (d *Debugger) ClearAllBreakpoints() {
	d.Breakpoints = make([]*debugtypes.CPUBreakpoint, 0, 16)
}

func (d *Debugger) DisableAllBreakpoints() {
	for i, _ := range d.Breakpoints {
		d.Breakpoints[i].Disabled = true
	}
}

func (d *Debugger) EnableAllBreakpoints() {
	for i, _ := range d.Breakpoints {
		d.Breakpoints[i].Disabled = false
	}
}

var DebuggerInstance *Debugger

// GetInstructionTrace returns instruction history and lookahead for the current CPU state
func (d *Debugger) GetInstructionTrace(backlog, lookahead int) *debugtypes.CPUInstructions {
	if d.ent() == nil {
		return nil
	}

	cpu := apple2helpers.GetCPU(d.ent())
	address := cpu.PC

	count := backlog + lookahead + 1
	instr := make([]debugtypes.CPUInstructionDecode, int(count))

	// Check if we're playing back a recording
	if d.CPU != nil && d.CPU.IsRecording && d.ent().IsPlayingVideo() {
		p := d.ent().GetPlayer()
		// backlog
		behind := p.GetLastNSyncs(backlog, address)
		for i, v := range behind {
			if v == -1 {
				v = 0
			}
			code, desc, cycles := cpu.DecodeInstruction(v)
			instr[i].Address = int(v)
			instr[i].Bytes = code
			instr[i].Instruction = desc
			instr[i].Cycles = cycles
			instr[i].Historic = true
		}
		// current
		code, desc, cycles := cpu.DecodeInstruction(address)
		j := backlog
		instr[j].Address = address
		instr[j].Bytes = code
		instr[j].Instruction = desc
		instr[j].Cycles = cycles
		instr[j].Historic = false
		// lookahead
		ahead := p.GetNextNSyncs(lookahead, address)
		for i, v := range ahead {
			if v == -1 {
				v = 0
			}
			code, desc, cycles := cpu.DecodeInstruction(v)
			instr[i+j+1].Address = int(v)
			instr[i+j+1].Bytes = code
			instr[i+j+1].Instruction = desc
			instr[i+j+1].Cycles = cycles
			instr[i+j+1].Historic = true
		}
	} else {
		// Use LastPC for historic data and lookahead for future
		for i := 0; i < count; i++ {
			if i < len(d.LastPC) && i < backlog {
				// Historic instructions from LastPC
				code, desc, cycles := cpu.DecodeInstruction(int(d.LastPC[i]))
				instr[i].Address = int(d.LastPC[i])
				instr[i].Bytes = code
				instr[i].Instruction = desc
				instr[i].Cycles = cycles
				instr[i].Historic = true
			} else if i == backlog {
				// Current instruction
				code, desc, cycles := cpu.DecodeInstruction(int(address))
				instr[i].Address = int(address)
				instr[i].Bytes = code
				instr[i].Instruction = desc
				instr[i].Cycles = cycles
				instr[i].Historic = false
				address += len(code) % 65536
			} else {
				// Lookahead instructions
				code, desc, cycles := cpu.DecodeInstruction(int(address))
				instr[i].Address = int(address)
				instr[i].Bytes = code
				instr[i].Instruction = desc
				instr[i].Cycles = cycles
				instr[i].Historic = false
				address += len(code) % 65536
			}
		}
	}

	return &debugtypes.CPUInstructions{
		Instructions: instr,
	}
}

func Start() {
	if settings.DebuggerOn && settings.SystemType != "nox" {
		DebuggerInstance = NewDebugger()
		DebuggerInstance.Serve(settings.DebuggerPort)
	}
}
