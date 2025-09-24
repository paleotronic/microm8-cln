// +build: !remint

package main

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"paleotronic.com/log"

	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/gorilla/mux"
	"paleotronic.com/core/hardware"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/hardware/servicebus"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/core/vduconst"
	"paleotronic.com/files"
	"paleotronic.com/fmt"
	"paleotronic.com/freeze"
	"paleotronic.com/glumby"
	"paleotronic.com/octalyzer/backend"
	"paleotronic.com/octalyzer/bus"
	"paleotronic.com/octalyzer/clientperipherals"
	"paleotronic.com/runestring"
	"paleotronic.com/utils"
)

type WindowState struct {
	Visbile bool
}

type AssemblerAPIRequest struct {
	Code string `json:"code"`
}

type MouseEventData struct {
	Button glumby.MouseButton
	Action glumby.Action
}

type SyncWindowRequest struct {
	Pos              *SetWindowRequest
	KeyEvent         *KeyRequest
	MouseButtonEvent *MouseEventData
	SnapShot         bool
	WindowState      *WindowState
}

var syncWindow = make(chan SyncWindowRequest, 1024)

var win *glumby.Window
var producer interfaces.Producable
var hidden bool

func CheckAndProcessSyncRequests() {
	for len(syncWindow) > 0 {
		r := <-syncWindow
		if r.Pos != nil {
			win.GetGLFWWindow().SetPos(r.Pos.X, r.Pos.Y)
			win.GetGLFWWindow().SetSize(r.Pos.W, r.Pos.H)
		}
		if r.KeyEvent != nil {
			req := r.KeyEvent
			// go func() {
				for backend.ProducerMain.AddressSpace.KeyBufferSize(SelectedIndex) > 0 {
					time.Sleep(time.Millisecond)
				}
				OnKeyEvent(
					win,
					glumby.Key(req.Key),
					glumby.ModifierKey(req.Modifiers),
					glumby.Action(req.Action),
				)
			// }()
			//win.GetGLFWWindow().Focus()
		}
		// simulate a mouse event
		if r.MouseButtonEvent != nil {
			//log.Printf("Synthetic mouse button event: %v", *r.MouseButtonEvent)
			OnMouseButtonEvent(win, r.MouseButtonEvent.Button, r.MouseButtonEvent.Action, 0)
		}
		// snapshot
		if r.SnapShot {
			SnapLayers()
		}
		// window state
		if r.WindowState != nil {
			if r.WindowState.Visbile {
				win.GetGLFWWindow().Show()
				hidden = false
			} else {
				// hide
				win.GetGLFWWindow().Hide()
				hidden = true
			}
		}
	}
}

func SetMainWindow(w *glumby.Window) {
	win = w
}

func UnpackRequest(w http.ResponseWriter, r *http.Request, req interface{}) error {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		w.WriteHeader(400)
		return err
	}

	//log.Printf("JSON: %s", string(b))

	err = json.Unmarshal(b, req)
	if err != nil {
		log.Printf("Error unmarshalling payload: %v", err)
		w.WriteHeader(400)
		return err
	}

	return nil
}

func HandleKeyRequest(w http.ResponseWriter, r *http.Request) {

	var req KeyRequest
	err := UnpackRequest(w, r, &req)
	if err != nil {
		return
	}

	//log.Printf("Got key request: %+v", req)

	syncWindow <- SyncWindowRequest{
		KeyEvent: &req,
	}

}

func HandleRebootRequest(w http.ResponseWriter, r *http.Request) {
	backend.ProducerMain.RebootVM(SelectedIndex)
}

func HandleCatalogRequest(w http.ResponseWriter, r *http.Request) {
	RAM.IntSetSlotInterrupt(SelectedIndex, true)
}

func HandleHelpRequest(w http.ResponseWriter, r *http.Request) {
	if !RAM.IntGetHelpInterrupt(SelectedIndex) && !settings.DisableMetaMode[SelectedIndex] {

		RAM.IntSetHelpInterrupt(SelectedIndex, true)

		fmt.Printf("HELP INTERRUPT SENT TO SLOT #%d\n", SelectedIndex)

		return

	}
}

func HandleUnfreezeRequest(w http.ResponseWriter, r *http.Request) {

}

func HandleFreezeRequest(w http.ResponseWriter, r *http.Request) {

}

func HandleFocusedRequest(w http.ResponseWriter, r *http.Request) {
	if win.IsFocused() {
		w.Write([]byte("1"))
	} else {
		w.Write([]byte("0"))
	}
}

func HandleScreenRequest(w http.ResponseWriter, r *http.Request) {
	var req ScreenRequest
	err := UnpackRequest(w, r, &req)
	if err != nil {
		return
	}

	settings.ScreenShotNeeded = true
	for settings.ScreenShotNeeded {
		time.Sleep(1 * time.Millisecond)
	}

	f, err := os.Create(req.Path)
	if err != nil {
		return
	}
	defer f.Close()

	f.Write(settings.ScreenShotJPEGData)
}

func HandleHideRequest(w http.ResponseWriter, r *http.Request) {

	e := backend.ProducerMain.GetInterpreter(SelectedIndex)
	e.StopTheWorld()
	bus.Sync()

	// win.GetGLFWWindow().Hide()

	syncWindow <- SyncWindowRequest{
		WindowState: &WindowState{
			Visbile: false,
		},
	}
}

func HandleShowRequest(w http.ResponseWriter, r *http.Request) {

	e := backend.ProducerMain.GetInterpreter(SelectedIndex)
	e.ResumeTheWorld()
	bus.Sync()

	// win.GetGLFWWindow().Show()

	syncWindow <- SyncWindowRequest{
		WindowState: &WindowState{
			Visbile: true,
		},
	}
}

func HandlePositionRequest(w http.ResponseWriter, r *http.Request) {

	if !settings.Windowed {
		return
	}

	var req SetWindowRequest
	err := UnpackRequest(w, r, &req)
	if err != nil {
		return
	}

	//log.Printf("Got pos request: %+v", req)

	syncWindow <- SyncWindowRequest{Pos: &req}

	//win.GetGLFWWindow().SetPos(req.X, req.Y)
	//win.GetGLFWWindow().SetSize(req.W, req.H)
	//win.GetGLFWWindow().Focus()

}

func HandleDiskBlankRequest(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	//log.Printf("Got disk blank request: %+v", vars)

	drive := utils.StrToInt(vars["drive"])
	now := time.Now()
	filename := fmt.Sprintf(
		"/local/MyDisks/blank_%.4d%.2d%.2d_%.2d%.2d%.2d.woz",
		now.Year(),
		now.Month(),
		now.Day(),
		now.Hour(),
		now.Minute(),
		now.Second(),
	)

	servicebus.SendServiceBusMessage(SelectedIndex, servicebus.DiskIIInsertBlank, servicebus.DiskTargetString{
		Drive:    drive,
		Filename: filename,
	})

	w.Write([]byte(filename))

}

func HandleDiskInsertRequest(w http.ResponseWriter, r *http.Request) {

	var req InsertDiskRequest
	err := UnpackRequest(w, r, &req)
	if err != nil {
		return
	}

	req.Filename = strings.Replace(req.Filename, "\\", "/", -1)

	log.Printf("Got disk insert request: %+v", req)

	switch req.Drive {
	case 0, 1:
		if req.Filename != "" {
			servicebus.SendServiceBusMessage(
				SelectedIndex,
				servicebus.DiskIIInsertFilename,
				servicebus.DiskTargetString{
					Filename: req.Filename,
					Drive:    req.Drive,
				},
			)
		} else {
			servicebus.SendServiceBusMessage(
				SelectedIndex,
				servicebus.DiskIIEject,
				req.Drive,
			)
		}
	case 2:
		if req.Filename != "" {
			servicebus.SendServiceBusMessage(
				SelectedIndex,
				servicebus.SmartPortInsertFilename,
				servicebus.DiskTargetString{
					Filename: req.Filename,
					Drive:    req.Drive,
				},
			)
		} else {
			servicebus.SendServiceBusMessage(
				SelectedIndex,
				servicebus.SmartPortEject,
				0,
			)
		}
	}

	settings.MicroPakPath = ""
	settings.Pakfile[SelectedIndex] = ""

	log.Printf("micropakpath = %s, settings.pakfile[0] = %s", settings.MicroPakPath, settings.Pakfile[SelectedIndex])

}

func HandleDiskEjectRequest(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	//log.Printf("Got disk eject request: %+v", vars)

	drive := utils.StrToInt(vars["drive"])

	switch drive {
	case 0, 1:
		servicebus.SendServiceBusMessage(
			SelectedIndex,
			servicebus.DiskIIEject,
			drive,
		)
	case 2:
		servicebus.SendServiceBusMessage(
			SelectedIndex,
			servicebus.SmartPortEject,
			0,
		)
	}

	w.Write([]byte(settings.PureBootVolume[SelectedIndex]))

}

func HandleDiskSwapRequest(w http.ResponseWriter, r *http.Request) {

	servicebus.SendServiceBusMessage(
		SelectedIndex,
		servicebus.DiskIIExchangeDisks,
		"",
	)

}

func HandleInterpRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	switch vars["name"] {
	case "fp", "int", "logo":
		e := backend.ProducerMain.GetInterpreter(SelectedIndex)
		apple2helpers.SwitchToDialect(e, vars["name"])
	}
}

func HandleMetaKeyRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	value := vars["value"]
	RAM.MetaKeySet(SelectedIndex, vduconst.SHIFT_CTRL_A+(rune(key[0])-'a'), rune(value[0]))
}

func handleFocusChange(ww *glumby.Window, focus bool) {
	//log.Printf("Callback for focus: %v", focus)
	// if notifyConn != nil && focus {
	// 	log.Printf("Sending socket notify")
	// 	notifyConn.Write([]byte("focused\r\n"))
	// }
}

func HandleSettingsUpdateRequest(w http.ResponseWriter, r *http.Request) {

	var req SettingsUpdateRequest
	err := UnpackRequest(w, r, &req)
	if err != nil {
		return
	}

	//log.Printf("Got settings update request: %+v", req)

	if SetConfig(SelectedIndex, req.Path, req.Value, req.Persist) {
		log.Printf("Successful")
	} else {
		log.Printf("Failed")
	}

}

func HandleMouseRequest(w http.ResponseWriter, r *http.Request) {

	var req MouseRequest
	err := UnpackRequest(w, r, &req)
	if err != nil {
		return
	}

	//log.Printf("Got mouse update request: %+v", req)

	win.OnMouseMove(win, float64(req.X), float64(req.Y))

}

func HandleSettingsFetchRequest(w http.ResponseWriter, r *http.Request) {

	var req SettingsUpdateRequest
	err := UnpackRequest(w, r, &req)
	if err != nil {
		return
	}

	//log.Printf("Got settings fetch request: %+v", req)

	value, _, _ := GetConfig(SelectedIndex, req.Path)
	w.Write([]byte(value))

}

func HandleProfileSetRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	profile := vars["profile"]
	e := backend.ProducerMain.GetInterpreter(SelectedIndex)
	settings.SpecFile[SelectedIndex] = profile + ".yaml"
	settings.ForcePureBoot[e.GetMemIndex()] = false
	if settings.MicroPakPath != "" {
		settings.MicroPakPath = ""
		settings.PureBootVolume[e.GetMemIndex()] = ""
		settings.PureBootVolume2[e.GetMemIndex()] = ""
		settings.PureBootSmartVolume[e.GetMemIndex()] = ""
	}
	settings.DiskIIUse13Sectors[e.GetMemIndex()] = false
	if ms, err := hardware.LoadSpec(settings.SpecFile[SelectedIndex]); err == nil {
		settings.ForcePureBoot[e.GetMemIndex()] = ms.AllowDisklessBoot
	}
	RAM.IntSetSlotRestart(SelectedIndex, true)
}

func HandleProfileGetRequest(w http.ResponseWriter, r *http.Request) {
	profile := strings.Replace(settings.SpecFile[SelectedIndex], ".yaml", "", -1)
	w.Write([]byte(profile))
}

func HandleFreezeRestoreRequest(w http.ResponseWriter, r *http.Request) {

	var req FreezeRequest
	err := UnpackRequest(w, r, &req)
	if err != nil {
		return
	}

	req.Path = strings.Replace(req.Path, "\\", "/", -1)

	log.Printf("Got freeze restore request: %+v", req)

	data, err := ioutil.ReadFile(req.Path)
	if err != nil {
		fp, err := files.ReadBytesViaProvider(files.GetPath(req.Path), files.GetFilename(req.Path))
		if err != nil {
			return
		}
		data = fp.Content
	}
	settings.PureBootRestoreStateBin[SelectedIndex] = data
	RAM.IntSetSlotRestart(SelectedIndex, true)

	// e := backend.ProducerMain.GetInterpreter(SelectedIndex)
	// f := freeze.NewEmptyState(e)
	// _ = f.LoadFromFile(req.Path)
	// f.Apply(e)

}

func HandleFreezeSaveRequest(w http.ResponseWriter, r *http.Request) {

	var req FreezeRequest
	err := UnpackRequest(w, r, &req)
	if err != nil {
		return
	}

	req.Path = strings.Replace(req.Path, "\\", "/", -1)

	log.Printf("Got freeze save request: %+v", req)

	e := backend.ProducerMain.GetInterpreter(SelectedIndex)
	f := freeze.NewFreezeState(e, false)
	_ = f.SaveToFile(req.Path)

}

func HandleRecordingRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	action := vars["action"]

	e := backend.ProducerMain.GetInterpreter(SelectedIndex)

	switch action {
	case "start-file-recording":
		e.StopRecording()
		e.RecordToggle(false)
	case "stop-recording":
		e.StopRecording()
	case "rewind":
		e.BackVideo()
	case "resume":
		if e.IsPlayingVideo() {
			servicebus.InjectServiceBusMessage(
				SelectedIndex,
				servicebus.PlayerResume,
				"",
			)
		}
	case "play":
		e.ForwardVideo()
	case "start-live-recording":
		e.StartRecording("", false)
	default:
		state := 0
		if e.IsRecordingVideo() {
			state = 1
		}
		if e.IsRecordingDiscVideo() {
			state = 2
		}
		w.Write([]byte(utils.IntToStr(state)))
	}
}

func HandleAudioRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	action := vars["action"]
	channel := vars["channel"]

	var level float64

	switch channel {
	case "master":
		level = settings.MixerVolume
	case "speaker":
		level = settings.SpeakerVolume[SelectedIndex]
	}

	switch action {
	case "up":
		level += 0.1
		if level > 1 {
			level = 1
		}
	case "down":
		level -= 0.1
		if level < 0 {
			level = 0
		}
	}

	switch channel {
	case "master":
		settings.MixerVolume = level
	case "speaker":
		settings.SpeakerVolume[SelectedIndex] = level
	}
}

func HandleCPURequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	action := vars["action"]

	e := backend.ProducerMain.GetInterpreter(SelectedIndex)
	cpu := apple2helpers.GetCPU(e)

	var level float64 = cpu.GetWarp()

	switch action {
	case "up":
		level *= 2
		if level > 4 {
			level = 4
		}
	case "down":
		level /= 2
		if level < 0.25 {
			level = 0.25
		}
	}

	cpu.SetWarpUser(level)

}

func HandlePauseRequest(w http.ResponseWriter, r *http.Request) {

	e := backend.ProducerMain.GetInterpreter(SelectedIndex)

	if e.IsWaitingForWorld() {
		e.ResumeTheWorld()
		bus.Sync()
	} else {
		e.StopTheWorld()
		bus.Sync()
	}

}

func HandleLaunchRequest(w http.ResponseWriter, r *http.Request) {

	var req LaunchRequest
	err := UnpackRequest(w, r, &req)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	if req.Pakfile != "" {
		req.Pakfile = strings.Replace(req.Pakfile, "\\", "/", -1)
		if rune(req.Pakfile[1]) == ':' {
			req.Pakfile = "/fs" + req.Pakfile[2:]
		} else if rune(req.Pakfile[0]) != '/' {
			req.Pakfile = "/fs" + req.Pakfile
		}
		w.Write([]byte("path: " + req.Pakfile))
	}

	settings.VMLaunch[SelectedIndex] = &settings.VMLauncherConfig{
		req.WorkingDir,
		req.Disks,
		req.Pakfile,
		req.SmartPort,
		req.RunFile,
		req.RunCommand,
		req.Dialect,
		"",
		false,
	}
	go backend.ProducerMain.RebootVM(SelectedIndex)

}

func HandleDebugLaunch(w http.ResponseWriter, r *http.Request) {
	settings.DebuggerOn = true
	settings.DebuggerAttachSlot = SelectedIndex + 1
	settings.PureBootCheck(settings.DebuggerAttachSlot - 1)
	if settings.PureBoot(settings.DebuggerAttachSlot - 1) {
		utils.OpenURL(fmt.Sprintf("http://localhost:%d/?attach=%d", settings.DebuggerPort, settings.DebuggerAttachSlot))
	}
}

func HandleQuit(w http.ResponseWriter, r *http.Request) {
	os.Exit(0)
}

func HandleAlive(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

func HandleMBRequest(w http.ResponseWriter, r *http.Request) {

	state := 0

	vars := mux.Vars(r)
	sentstate := vars["state"]
	if sentstate != "" {
		var button glumby.MouseButton
		var action glumby.Action = glumby.Press
		state = utils.StrToInt(sentstate)
		settings.LeftButton = (state & 1) != 0
		settings.RightButton = (state & 2) != 0
		settings.MiddleButton = (state & 4) != 0
		if settings.LeftButton {
			button = glumby.MouseButtonLeft
		}
		if settings.RightButton {
			button = glumby.MouseButtonRight
		}

		if state == 0 {
			action = glumby.Release
			button = glumby.MouseButtonLeft
		}

		syncWindow <- SyncWindowRequest{
			MouseButtonEvent: &MouseEventData{
				Button: button,
				Action: action,
			},
		}
	}

	if settings.LeftButton {
		state |= 1
	}

	if settings.RightButton {
		state |= 2
	}

	if settings.MiddleButton {
		state |= 4
	}

	// log.Printf("mouse button state = %d", state)

	//w.Write([]byte(utils.IntToStr(state)))
	w.Write([]byte(utils.IntToStr(state)))

}

func HandleVMRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vm := vars["vm"]

	if vm != "" {
		// switch
		slotid := utils.StrToInt(vm) - 1
		backend.ProducerMain.Select(slotid)
		SelectedCamera = slotid
		SelectedIndex = slotid
		SelectedAudioIndex = slotid
		clientperipherals.Context = slotid
		clientperipherals.SPEAKER.SelectChannel(SelectedAudioIndex)
	}

	w.Write([]byte(utils.IntToStr(SelectedIndex + 1)))
}

func HandleShotRequest(w http.ResponseWriter, r *http.Request) {
	syncWindow <- SyncWindowRequest{
		SnapShot: true,
	}
}

func HandleDiskWPRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	drive := utils.StrToInt(vars["drive"])
	verb := vars["verb"]

	if verb != "" {
		switch verb {
		case "toggle":
			servicebus.SendServiceBusMessage(
				SelectedIndex,
				servicebus.DiskIIToggleWriteProtect,
				drive&1,
			)
		}
	}

	// return state
	resp, _ := servicebus.SendServiceBusMessage(
		SelectedIndex,
		servicebus.DiskIIQueryWriteProtect,
		drive&1,
	)
	var rs = "0"
	if resp[0].Payload.(bool) {
		rs = "1"
	}

	w.Write([]byte(rs))

}

var lastHeartbeat = time.Now()

func HeartBeatReceiver() {

	for backend.ProducerMain == nil {
		time.Sleep(1 * time.Second)
	}

	ticker := time.NewTicker(500 * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			if time.Since(lastHeartbeat) > 1100*time.Millisecond {
				if !hidden {
					e := backend.ProducerMain.GetInterpreter(SelectedIndex)
					e.StopTheWorld()
					bus.Sync()
					syncWindow <- SyncWindowRequest{
						WindowState: &WindowState{
							Visbile: false,
						},
					}
				}
			}
		}
	}
}

func MousePoller() {
	for {
		if settings.LeftButton {
			if win.GetGLFWWindow().GetMouseButton(glfw.MouseButtonLeft) == glfw.Release {
				syncWindow <- SyncWindowRequest{
					MouseButtonEvent: &MouseEventData{
						Button: glumby.MouseButtonLeft,
						Action: glumby.Release,
					},
				}
			}
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func HandleHeartBeatRequest(w http.ResponseWriter, r *http.Request) {

	if hidden {
		e := backend.ProducerMain.GetInterpreter(SelectedIndex)
		e.ResumeTheWorld()
		bus.Sync()
		syncWindow <- SyncWindowRequest{
			WindowState: &WindowState{
				Visbile: true,
			},
		}
	}

	lastHeartbeat = time.Now()
}

func HandlePasteRequest(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}
	e := backend.ProducerMain.GetInterpreter(SelectedIndex)
	e.SetPasteBuffer(runestring.Cast(string(data)))
}

func HandleOSDRequest(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}
	e := backend.ProducerMain.GetInterpreter(SelectedIndex)
	apple2helpers.OSDShow(e, string(data))
}

func HandleMemoryReadRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := utils.StrToInt(vars["address"])
	e := backend.ProducerMain.GetInterpreter(SelectedIndex)
	value := int(e.GetMemory(address)) & 0xff
	w.Write([]byte(utils.IntToStr(value)))
}

func HandleMemoryWriteRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := utils.StrToInt(vars["address"])
	v := uint64(utils.StrToInt(vars["value"]))
	e := backend.ProducerMain.GetInterpreter(SelectedIndex)
	e.SetMemory(address, v)
	w.Write([]byte(utils.IntToStr(int(v))))
}

func HandleTextScreenRequest(w http.ResponseWriter, r *http.Request) {

	for _, l := range HUDLayers[SelectedIndex] {
		if l == nil || l.Spec.GetActive() == false {
			continue
		}
		// found active text layer...
		data := l.GetText(types.LayerRect{0, 0, 79, 47})
		w.Write([]byte(data))
		return
	}

}

func HandleCameraRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	action := vars["action"]
	e := backend.ProducerMain.GetInterpreter(SelectedIndex)
	switch action {
	case "reset":
		index := e.GetMemIndex()
		mm := e.GetMemoryMap()
		cindex := mm.GetCameraConfigure(index)
		control := types.NewOrbitController(mm, index, cindex)
		control.ResetALL()
		control.Update()
	}
}

func HandleButtonClickRequest(w http.ResponseWriter, r *http.Request) {
	syncWindow <- SyncWindowRequest{
		MouseButtonEvent: &MouseEventData{
			glumby.MouseButtonLeft,
			glumby.Press,
		},
	}
	time.Sleep(20 * time.Millisecond)
	syncWindow <- SyncWindowRequest{
		MouseButtonEvent: &MouseEventData{
			glumby.MouseButtonLeft,
			glumby.Release,
		},
	}
}

func HandleAssembleRequest(w http.ResponseWriter, r *http.Request) {
	var code string

	// Check Content-Type header
	contentType := r.Header.Get("Content-Type")
	if contentType == "text/plain" || contentType == "text/x-asm" {
		// Read raw text from body
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(400)
			w.Write([]byte(fmt.Sprintf("Error reading body: %v", err)))
			return
		}
		code = string(bodyBytes)
	} else {
		// Default JSON handling
		var req AssemblerAPIRequest
		err := UnpackRequest(w, r, &req)
		if err != nil {
			w.WriteHeader(400)
			w.Write([]byte(err.Error()))
			return
		}
		code = req.Code
	}

	// Base64 encode the assembly code
	encodedCode := base64.StdEncoding.EncodeToString([]byte(code))

	makefile := `
main = "main.s"
	`
	encodedMake := base64.StdEncoding.EncodeToString([]byte(makefile))

	// Prepare the request
	asmReq := AssemblerRequest{
		Files: []AssemblerFile{
			{
				Name:   "main.s",
				Data:   encodedCode,
				Binary: false,
			},
			{
				Name:   "makefile",
				Data:   encodedMake,
				Binary: false,
			},
		},
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(asmReq)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("Failed to marshal request: %v", err)))
		return
	}

	// Create HTTP client that ignores certificate verification
	// WARNING: This is insecure and should only be used temporarily
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	// Make the HTTP request
	resp, err := client.Post("https://turtlespaces.org:6502/api/v1/asm/multifile",
		"application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("Failed to call assembler API: %v", err)))
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("Failed to read response: %v", err)))
		return
	}

	// Now parse the actual response
	var asmResp StructuredASMResponse
	if err := json.Unmarshal(body, &asmResp); err != nil {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("Failed to unmarshal assembler response: %v", err)))
		return
	}

	// Check for errors
	if len(asmResp.Err) > 0 {
		var errorMsg string
		for _, e := range asmResp.Err {
			errorMsg += fmt.Sprintf("Line %d in %s: %s\n", e.Line, e.Filename, e.Message)
		}
		w.WriteHeader(400)
		w.Write([]byte(fmt.Sprintf("Assembly error(s):\n%s", errorMsg)))
		return
	}

	// Success - the data is already in byte array format
	if len(asmResp.Data) == 0 || asmResp.Address == 0 {
		w.WriteHeader(500)
		w.Write([]byte("Unexpected response format: missing data or address"))
		return
	}

	// Write the assembled bytes to memory
	e := backend.ProducerMain.GetInterpreter(SelectedIndex)
	for i, b := range asmResp.Data {
		e.SetMemory(asmResp.Address+i, uint64(b))
	}

	// Prepare success response
	response := map[string]interface{}{
		"success": true,
		"address": asmResp.Address,
		"length":  len(asmResp.Data),
		"message": fmt.Sprintf("%s: %d bytes assembled to memory address $%04X",
			asmResp.Name, len(asmResp.Data), asmResp.Address),
	}

	// If less than 32 bytes, include hex dump
	if len(asmResp.Data) < 32 {
		var hexBytes []string
		for _, b := range asmResp.Data {
			hexBytes = append(hexBytes, fmt.Sprintf("%02X", b))
		}
		response["hex"] = hexBytes
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func StartControlServer(hostport string, w *glumby.Window) error {
	win = w
	r := mux.NewRouter()
	r.HandleFunc("/api/control/window/position", HandlePositionRequest).Methods("POST")
	r.HandleFunc("/api/control/input/keyevent", HandleKeyRequest).Methods("POST")
	r.HandleFunc("/api/control/window/hide", HandleHideRequest).Methods("GET")
	r.HandleFunc("/api/control/window/show", HandleShowRequest).Methods("GET")
	r.HandleFunc("/api/control/system/reboot", HandleRebootRequest).Methods("GET")
	r.HandleFunc("/api/control/window/freeze", HandleFreezeRequest).Methods("GET")
	r.HandleFunc("/api/control/window/unfreeze", HandleUnfreezeRequest).Methods("GET")
	r.HandleFunc("/api/control/window/screen", HandleScreenRequest).Methods("POST")
	r.HandleFunc("/api/control/window/focused", HandleFocusedRequest).Methods("GET")
	r.HandleFunc("/api/control/hardware/disk/insert", HandleDiskInsertRequest).Methods("POST")
	r.HandleFunc("/api/control/hardware/disk/eject/{drive}", HandleDiskEjectRequest).Methods("GET")
	r.HandleFunc("/api/control/hardware/disk/blank/{drive}", HandleDiskBlankRequest).Methods("GET")
	r.HandleFunc("/api/control/system/catalog", HandleCatalogRequest).Methods("GET")
	r.HandleFunc("/api/control/interpreter/{name}", HandleInterpRequest).Methods("GET")
	r.HandleFunc("/api/control/input/meta/key/{key}/value/{value}", HandleMetaKeyRequest).Methods("GET")
	r.HandleFunc("/api/control/settings/update", HandleSettingsUpdateRequest).Methods("POST")
	r.HandleFunc("/api/control/settings/get", HandleSettingsFetchRequest).Methods("POST")
	r.HandleFunc("/api/control/input/mouseevent", HandleMouseRequest).Methods("POST")
	r.HandleFunc("/api/control/hardware/disk/swap", HandleDiskSwapRequest).Methods("GET")
	r.HandleFunc("/api/control/system/profile/set/{profile}", HandleProfileSetRequest).Methods("GET")
	r.HandleFunc("/api/control/system/profile/get", HandleProfileGetRequest).Methods("GET")
	r.HandleFunc("/api/control/system/freeze/restore", HandleFreezeRestoreRequest).Methods("POST")
	r.HandleFunc("/api/control/system/freeze/save", HandleFreezeSaveRequest).Methods("POST")
	r.HandleFunc("/api/control/recorder/{action}", HandleRecordingRequest).Methods("GET")
	r.HandleFunc("/api/control/recorder", HandleRecordingRequest).Methods("GET")
	r.HandleFunc("/api/control/system/launch", HandleLaunchRequest).Methods("POST")
	r.HandleFunc("/api/control/debug/attach", HandleDebugLaunch).Methods("GET")
	r.HandleFunc("/api/control/quit", HandleQuit).Methods("GET")
	r.HandleFunc("/api/control/health", HandleAlive).Methods("GET")
	r.HandleFunc("/api/control/audio/{channel}/{action}", HandleAudioRequest).Methods("GET")
	r.HandleFunc("/api/control/cpu/warp/{action}", HandleCPURequest).Methods("GET")
	r.HandleFunc("/api/control/pause", HandlePauseRequest).Methods("GET")
	r.HandleFunc("/api/control/mouse/buttonstate", HandleMBRequest).Methods("GET")
	r.HandleFunc("/api/control/mouse/buttonstate/{state}", HandleMBRequest).Methods("GET")
	r.HandleFunc("/api/control/vm", HandleVMRequest).Methods("GET")
	r.HandleFunc("/api/control/vm/{vm}", HandleVMRequest).Methods("GET")
	r.HandleFunc("/api/control/system/help", HandleHelpRequest).Methods("GET")
	r.HandleFunc("/api/control/window/screenshot", HandleShotRequest).Methods("GET")
	r.HandleFunc("/api/control/hardware/disk/wp/{drive}", HandleDiskWPRequest).Methods("GET")
	r.HandleFunc("/api/control/hardware/disk/wp/{drive}/{verb}", HandleDiskWPRequest).Methods("GET")
	r.HandleFunc("/api/control/paste", HandlePasteRequest).Methods("POST")
	r.HandleFunc("/api/control/memory/read/{address}", HandleMemoryReadRequest).Methods("GET")
	r.HandleFunc("/api/control/memory/write/{address}/{value}", HandleMemoryWriteRequest).Methods("GET")
	r.HandleFunc("/api/control/memory/screen/text", HandleTextScreenRequest).Methods("GET")
	r.HandleFunc("/api/control/osd/send", HandleOSDRequest).Methods("POST")
	r.HandleFunc("/api/control/system/camera/{action}", HandleCameraRequest).Methods("GET")
	r.HandleFunc("/api/control/mouse/buttonclick", HandleButtonClickRequest).Methods("GET")
	r.HandleFunc("/api/control/system/heartbeat", HandleHeartBeatRequest).Methods("GET")
	r.HandleFunc("/api/control/assembly/assemble", HandleAssembleRequest).Methods("POST")

	w.OnFocusChanged = handleFocusChange

	srv := &http.Server{
		Addr: hostport,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r, // Pass our instance of gorilla/mux in.
	}

	log.Printf("Starting control server on %s", hostport)

	time.AfterFunc(1*time.Second, func() {
		// go MousePoller()
		//go HeartBeatReceiver()
	})

	return srv.ListenAndServe()
}
