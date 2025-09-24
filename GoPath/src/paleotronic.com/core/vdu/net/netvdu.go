package vdu

import (
	"errors"
	"paleotronic.com/fmt"
	//"os"
	"math"
	"time"

	"paleotronic.com/core/hires"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/core/vduproto"
	"paleotronic.com/core/vduserver"
	"paleotronic.com/files"
	"paleotronic.com/log"
	"paleotronic.com/utils"
	"paleotronic.com/restalgia"
	"paleotronic.com/restalgia/driver"
)

const (
	TEXT_BASE_WIDTH = 80
	TEXT_BASE_HEIGHT = 48
)

type NetVDU struct {
	VDU
	Server        *vduserver.VDUServer
	AServer       *vduserver.AudioServer
	stateChan     chan vduproto.VDUServerEvent
	keyChan       chan vduproto.VDUServerEvent
	paddleChan    chan vduproto.VDUServerEvent
	assChan       chan vduproto.VDUServerEvent
	cpuChan       chan vduproto.VDUServerEvent
	ThinEvents    *vduproto.ThinScreenEventBuffer
	lastSoundTick int64
	Output        driver.Output
	buzzer        *restalgia.Voice
	tone          *restalgia.Voice
	instTone      *restalgia.Instrument
}

func NewNetVDU() *NetVDU {
	this := &NetVDU{}
	this.VDU = *NewVDU(nil)

	this.Server = vduserver.NewVDUServer(":9988", vduproto.VDU_DEFAULT_HOST, this)
	this.AServer = vduserver.NewAudioServer(":9988", vduproto.VDU_DEFAULT_HOST, this)

	// register channels
	this.stateChan = make(chan vduproto.VDUServerEvent)
	this.keyChan = make(chan vduproto.VDUServerEvent)
	this.paddleChan = make(chan vduproto.VDUServerEvent)
	this.cpuChan = make(chan vduproto.VDUServerEvent)
	this.assChan = make(chan vduproto.VDUServerEvent)
	this.Server.RegisterMessageType("KEY", this.keyChan)
	this.Server.RegisterMessageType("SRQ", this.stateChan)
	this.Server.RegisterMessageType("PBE", this.paddleChan)
	this.Server.RegisterMessageType("PVE", this.paddleChan)
	this.Server.RegisterMessageType("PME", this.paddleChan)
	this.ThinEvents = vduproto.NewThinScreenEventBuffer()
	this.Server.RegisterMessageType("YLD", this.cpuChan)
	this.Server.RegisterMessageType("AQT", this.assChan)
	this.Server.RegisterMessageType("AQF", this.assChan)

	this.lastSoundTick = time.Now().UnixNano()

	go this.HandleClientMessages()
	go this.HandleThinMessages()

	return this
}

func (this *NetVDU) Reconnect(ip string) {
	// send message to front-end telling it to reconnect
	this.Server.SendConnectCommand(ip)
}

func (this *NetVDU) CheckAndSend() {
	//	for {
	//		select {
	//		case d := <-buzzer.PacketChan:
	//			// got a packet

	//			var empty bool = true
	//			for _, v := range d {
	//				if v != 0 {
	//					empty = false
	//					break
	//				}
	//			}

	//			if !empty {

	//				ce := &vduproto.ClickEvent{Data: d}
	//				if this.AServer != nil {
	//					this.AServer.SendSpeakerClickEvent(ce)
	//				}

	//			}
	//			//////fmt.Println("send", time.Now().Second())
	//		}
	//	}
}

//func (this *NetVDU) StartAudio() {

//	go this.CheckAndSend()
//	go func() {
//		for {
//			if math.Abs(float64(buzzer.LastPacketSent)-float64(buzzer.WritePacket)) > 1 {
//				buzzer.LastPacketSent = (buzzer.LastPacketSent + 1) % 10
//				buzzer.PacketBuffer[buzzer.LastPacketSent].SendBlock()
//			} else {
//				time.Sleep(time.Microsecond * 500)
//			}
//		}
//	}()

//}

//func (this *NetVDU) StopAudio() {

//}

func (this *NetVDU) StopAudio() {
	//	this.Output.Stop()
}

func (this *NetVDU) StartAudio() {

	//	output, err := driver.Get(44100, 1)
	//	if err != nil {
	//		panic(err)
	//	}

	//	this.Output = output

	//	// Now lets add a tone generator
	//	this.instTone = restalgia.NewInstrument("WAVE=PULSE:VOLUME=1.0:ADSR=0,0,100,0")
	//	this.tone = restalgia.NewVoice(44100, restalgia.SINE, 1.0)
	//	this.instTone.Apply(this.tone)
	//	this.tone.SetVolume(1.0)
	//	this.tone.SetFrequency(1000)

	//	// Now lets add a custom wave generator
	//	fred := restalgia.NewInstrument("WAVE=CUSTOM:VOLUME=1.0:ADSR=0,0,1000,0")
	//	this.buzzer = restalgia.NewVoice(44100, restalgia.CUSTOM, 1.0)
	//	fred.Apply(this.buzzer)
	//	this.buzzer.SetVolume(1.0)
	//	this.buzzer.SetFrequency(1000)

	//	this.Output.Start()

	//	go func() {
	//		// feed tones
	//		for {
	//			this.CheckToneLevel()

	//			fragment := make([]float32, 128)
	//			for i := 0; i < len(fragment); i++ {
	//				v := (this.tone.GetAmplitude() + this.buzzer.GetAmplitude()) / 2
	//				fragment[i] = v
	//			}
	//			// pass fragment
	//			this.Output.Push(fragment)
	//			//////fmt.Println("tick")
	//		}
	//	}()

}

// calculate and supply the changed screen data portion
/*func (this *NetVDU) GetChangedBounds(ox, oy, nx, ny int) (int, []int) {
	if ((ny * this.VideoMode.Columns) + nx) < ((oy * this.VideoMode.Columns) + ox) {
		ox = 0
		oy = 0
		nx = this.VideoMode.Columns - 1
		ny = this.VideoMode.Rows - 1
	}

	addr1 := this.XYToOffset(ox, oy)
	addr2 := this.XYToOffset(nx, ny)

	if addr1 < addr2 {
		return addr1, this.TextMemory[addr1 : addr2+1]
	} else {
		return addr2, this.TextMemory[addr2 : addr1+1]
	}
}
*/

func (this *NetVDU) RealPut(ch rune) {
	// call super method
	//ox := this.GetCursorX()
	//	oy := this.GetCursorY()
	if ch == 7 {
		this.Beep()
	}

	this.VDUCore.RealPut(ch)

	//	nx := this.GetCursorX()
	//	ny := this.GetCursorY()
	//log.Printf("Cursor was at %d, %d then moved to %d, %d\n", ox, oy, nx, ny)
	//this.Server.SendScreenPositionUpdate(this.VDUCore.CursorX, this.VDUCore.CursorY)
}

func (this *NetVDU) TAB() {
	this.VDUCore.TAB()
	this.Server.SendScreenPositionUpdate(this.VDUCore.CursorX, this.VDUCore.CursorY)
}

func (this *NetVDU) LF() {
	this.VDUCore.LF()
	this.Server.SendScreenPositionUpdate(this.VDUCore.CursorX, this.VDUCore.CursorY)
}

func (this *NetVDU) CR() {
	this.VDUCore.CR()
	this.Server.SendScreenPositionUpdate(this.VDUCore.CursorX, this.VDUCore.CursorY)
}

func (this *NetVDU) Backspace() {
	this.VDUCore.Backspace()
	this.Server.SendScreenPositionUpdate(this.VDUCore.CursorX, this.VDUCore.CursorY)
}

/*

func (this *NetVDU) Home() {
	this.VDUCore.Home()
	//	this.Server.SendScreenMemoryChange(0, this.VDUCore.TextMemory, this.VDUCore.CursorX, this.VDUCore.CursorY)
}

func (this *NetVDU) ClrHome() {
	this.VDUCore.Home()
	this.VDUCore.Clear()
	//	this.Server.SendScreenMemoryChange(0, this.VDUCore.TextMemory, this.VDUCore.CursorX, this.VDUCore.CursorY)
}

func (this *NetVDU) Scroll() {
	this.VDUCore.Scroll()
	//	this.Server.SendScreenMemoryChange(0, this.VDUCore.TextMemory, this.VDUCore.CursorX, this.VDUCore.CursorY)
}

func (this *NetVDU) Clear() {
	this.VDUCore.Clear()
	//	this.Server.SendScreenMemoryChange(0, this.VDUCore.TextMemory, this.VDUCore.CursorX, this.VDUCore.CursorY)
}
*/

func (this *NetVDU) ConfigSpecification( vm types.VideoMode ) {

	this.Specification.HUDLayers = make([]vduproto.LayerSpec, 0)
	this.Specification.GFXLayers = make([]vduproto.LayerSpec, 0)

	//fmt.Printf("Debug: VideoMode actual rows = %d, virtual = %d\n", vm.ActualRows, vm.Rows)

	// TEXT LAYER ZERO
//	xft := vm.Columns / TEXT_BASE_WIDTH
	yft := TEXT_BASE_HEIGHT / vm.Rows
	var r vduproto.LayerRect
	if vm.ActualRows != vm.Rows {
		//fmt.Println("Partial text mode")
		r = vduproto.LayerRect{
			0, uint16(vm.Rows-vm.ActualRows)*uint16(yft),
			TEXT_BASE_WIDTH-1, TEXT_BASE_HEIGHT-1,
		}
	} else {
		//fmt.Println("FULL text mode")
		r = vduproto.LayerRect{
			0, 0,
			TEXT_BASE_WIDTH-1, TEXT_BASE_HEIGHT-1,
		}
	}
	this.Specification.HUDLayers = append( this.Specification.HUDLayers, vduproto.LayerConfigText(
		vduproto.APPLE_TEXT_PAGE_0,
		(vm.ActualRows > 0),
		TEXT_BASE_WIDTH, TEXT_BASE_HEIGHT,
		vm.Columns, vm.Rows,
		vm.Palette,
		r,
	) )

	// LOWRES LAYER ZERO
	this.Specification.GFXLayers = append( this.Specification.GFXLayers, vduproto.LayerConfigGFX(
		vduproto.APPLE_LORES_PAGE_0,
		(vm.ActualRows < vm.Rows && vm.Width < 50),
		40, 48,
		vm.Palette,
		vduproto.LayerRect{
			0, 0, 39, (47 - (uint16(vm.ActualRows)*2)),
		},
	) )

	// HIRES LAYER ZERO
	this.Specification.GFXLayers = append( this.Specification.GFXLayers, vduproto.LayerConfigGFX(
		vduproto.APPLE_HIRES_PAGE_0,
		(vm.ActualRows < vm.Rows && vm.Width > 50 && this.DisplayPage == 0),
		280, 192,
		vm.Palette,
		vduproto.LayerRect{
			0, 0, 279, (191 - (uint16(vm.ActualRows)*8)),
		},
	) )

	// HIRES LAYER ONE
	this.Specification.GFXLayers = append( this.Specification.GFXLayers, vduproto.LayerConfigGFX(
		vduproto.APPLE_HIRES_PAGE_1,
		(vm.ActualRows < vm.Rows && vm.Width > 50 && this.DisplayPage == 1),
		280, 192,
		vm.Palette,
		vduproto.LayerRect{
			0, 0, 279, (191 - (uint16(vm.ActualRows)*8)),
		},
	) )

	//fmt.Println( this.Specification.String() )

}

func (this *NetVDU) SetVideoMode(vm types.VideoMode) {
	this.VDUCore.SetVideoMode(vm)

	// Text Layer spec
	// LayerSendText( index byte,  width, height int, vwidth, vheight int, palette types.VideoPalette, bounds LayerRect)
	this.ConfigSpecification(vm)
	this.Server.SendLayerBundle(this.Specification)

	//this.Server.SendScreenSpecification(vm)
}

func (this *NetVDU) Put(ch rune) {
	this.RealPut(ch)
}

func (this *NetVDU) PutStr(s string) error {

	for _, ch := range s {
		this.Put(ch)
	}

	return nil
}

func (this *NetVDU) DoPrompt() {
	_ = this.PutStr(this.Prompt)
}

func (this *NetVDU) HandleClientMessages() {
	for {
		select {
		case pdl := <-this.paddleChan:
			log.Printf("Got PDL event [%v]\n", pdl)
			switch {
			case pdl.Kind == "PBE":
				var out vduproto.PaddleButtonEvent
				out = pdl.Data.(vduproto.PaddleButtonEvent)
				this.SetPaddleButtons(int(out.PaddleID), (out.ButtonState != 0))
			case pdl.Kind == "PVE":
				var out vduproto.PaddleValueEvent
				out = pdl.Data.(vduproto.PaddleValueEvent)
				this.SetPaddleValues(int(out.PaddleID), int(out.PaddleValue))
			case pdl.Kind == "PME":
				var out vduproto.PaddleModifyEvent
				out = pdl.Data.(vduproto.PaddleModifyEvent)
				v := this.GetPaddleValues(int(out.PaddleID)) + int(out.Difference)
				if v < 0 {
					v = 0
				} else if v > 255 {
					v = 255
				}
				this.SetPaddleValues(int(out.PaddleID), v)
			}
		case kin := <-this.keyChan:
			var out vduproto.KeyPressEvent
			out = kin.Data.(vduproto.KeyPressEvent)
			this.InsertCharToBuffer(out.Character)
			// do stuff for key
		case _ = <-this.stateChan:
			// send state across
			this.Server.SendScreenSpecification(this.GetVideoMode())
			this.Server.SendScreenMemoryChange(0, this.ShadowTextMemory.GetValues(0, this.ShadowTextMemory.Size()), this.CursorX, this.CursorY)
		}
	}
}

func (this *NetVDU) HandleThinMessages() {
	for {

		bytes, i := this.WozBitmapMemory[this.CurrentPage].GetScanLineChange()

		if i > -1 {
			this.Server.SendScanLine(i, bytes)
			//////fmt.Printf("Sending update for scanline %d\n", i)
		} else {
			time.Sleep(time.Millisecond * 1)
		}

		//		evlist := this.ThinEvents.GetEvents()

		//		if len(evlist) == 0 {
		//			time.Sleep(time.Millisecond * 1)
		//		} else {
		//			this.Server.SendThinScreenMessages(evlist)
		//		}

	}
}

func (this *NetVDU) PassWaveBuffer(data []float32) {
	//	this.buzzer.GetOSC(0).GetWfCUSTOM().Stimulate(data)
	//	this.buzzer.SetVolume(1)
	//	this.buzzer.GetOSC(0).Trigger()
}

func (this *NetVDU) CheckToneLevel() {
	now := time.Now().UnixNano()
	duration := int64(math.Abs(float64(now - this.lastSoundTick)))

	if duration == 0 {
		duration = 1000
	}

	freq := 1000000000 / (duration * 2)
	if (freq < 2) && (this.tone.GetVolume() > 0) {
		this.tone.SetVolume(0)
		//System.err.println("Silence voice");
	}
}

//func (this *NetVDU) Click() {
//	buzzer.Pluck()
//}

func (this *NetVDU) PlayWave(p, f string) (bool, error) {

	// confirm asset available on remote
	exists, err := this.AssetCheck(p, f)
	if err != nil {
		return false, err
	}

	if exists != nil {
		aa := &vduproto.AssetAction{MD5: exists.GetMD5Sum(), Action: vduproto.AT_Audio_WAV}
		this.Server.SendAssetPlayback(aa)
	}

	return true, nil

}

func (this *NetVDU) PNGSplash(p, f string) (bool, error) {

	// confirm asset available on remote
	exists, err := this.AssetCheck(p, f)
	if err != nil {
		return false, err
	}

	if exists != nil {
		aa := &vduproto.AssetAction{MD5: exists.GetMD5Sum(), Action: vduproto.AT_Image_PNG}
		this.Server.SendAssetPlayback(aa)
	}

	return true, nil

}

func (this *NetVDU) PNGBackdrop(p, f string) (bool, error) {

	// confirm asset available on remote
	exists, err := this.AssetCheck(p, f)
	if err != nil {
		return false, err
	}

	if exists != nil {
		aa := &vduproto.AssetAction{MD5: exists.GetMD5Sum(), Action: vduproto.AT_Image_PNG_43}
		this.Server.SendAssetPlayback(aa)
	}

	return true, nil

}

func (this *NetVDU) LoadEightTrack(p, f string) (bool, error) {

	// confirm asset available on remote
	exists, err := this.AssetCheck(p, f)
	if err != nil {
		return false, err
	}

	if exists != nil {
		aa := &vduproto.AssetAction{MD5: exists.GetMD5Sum(), Action: vduproto.AT_Music_8T}
		this.Server.SendAssetPlayback(aa)
	}

	return true, nil

}

// Confirm client has access to file
func (this *NetVDU) AssetCheck(p, f string) (*files.FilePack, error) {

	raw, err := files.ReadBytesViaProvider(p, f)

	if err != nil {
		return nil, err // most likely file does not exist
	}

	// okay file exists, we have the data
	fp := files.NewFilePackFromBytes(raw)
	//fp.CacheData() // store data if it does not exist in user cache

	// Step one query client
	aq := fp.Query()

	if aq == nil {
		return nil, errors.New("AssetQuery generate failed")
	}

	// Dispatch asset query to client
	this.Server.SendAssetQuery(aq)

	var msg vduproto.VDUServerEvent

	timeout := time.After(5 * time.Second)

	select {
	case msg = <-this.assChan:
		// response from client
		////fmt.Printf("Got Client asset response code %s for %v\n", msg.Kind, msg.Data)
	case _ = <-timeout:
		return nil, errors.New("Timeout on AssetQuery response")
	}

	// TODO stream asset content over wire
	if msg.Kind == "AQT" {
		return fp, nil
	}

	if msg.Kind != "AQF" {
		return nil, errors.New("Error in Client AssetQuery response")
	}

	// At this point we can send the asset across to the client
	idx := 0
	assetBlock := fp.GetBlock(idx)
	for assetBlock != nil {
		this.Server.SendAssetBlock(assetBlock)
		idx++
		assetBlock = fp.GetBlock(idx)
	}

	return fp, nil
}

func (this *NetVDU) ExecNative(mem []int, a int, x int, y int, pc int, sr int, sp int, vdu interfaces.Display) {

	this.Server.SendCPUEvent(pc)

	//this.VDUCore.ExecNative(mem, a, x, y, pc, sr, sp, vdu)
	select {
	case done := <-this.cpuChan:
		// control is back :)
		log.Println(done)
	}

}

func (this *NetVDU) SetClassicHGR(v bool) {
	// send ThinScreen event here...
	this.ThinEvents.ToggleHGR( v )
	this.UseClassicHGR = v
	this.Server.SendThinScreenMessages(this.ThinEvents.GetEvents())
}

func (this *NetVDU) IsClassicHGR() bool {
	return this.UseClassicHGR
}

func (this *NetVDU) SetMemoryValue(x, y int) bool {

	ov := this.VDUCore.VideoMode

	mapped := this.VDUCore.SetMemoryValue(x, y)

	nv := this.VDUCore.VideoMode

	if ov.ActualRows != nv.ActualRows || ov.Width != nv.Width || ov.Height != nv.Height {
		this.Server.SendScreenSpecification(nv)
	}

	if !mapped {
		// send update
		this.Server.SendMemoryUpdate(x, y)
	}

	return mapped
}

func (this *NetVDU) Click() {
	// click the speaker
	now := time.Now().UnixNano()
	duration := int64(math.Abs(float64(now - this.lastSoundTick)))

	if duration == 0 {
		this.lastSoundTick = now
		return
	}

	freq := 1000000000 / (duration * 2)

	this.AServer.SendSpeakerClickEvent(int(freq))

	//System.out.println("Approx frequency is "+freq+"Hz, but using median "+usefreq+"Hz");

	//	if this.instTone == nil {
	//		this.instTone = restalgia.NewInstrument("WAVE=PULSE:VOLUME=1.0:ADSR=0,0,100,0")
	//		this.instTone.Apply(this.tone)
	//		this.tone.GetOSC(0).Trigger()
	//	}

	//	if (usefreq > 10) && (float64(usefreq) != this.tone.OSC[0].GetFrequency()) {
	//		this.tone.OSC[0].SetFrequency(float64(usefreq))
	//		if this.tone.GetVolume() == 0 {
	//			this.tone.SetVolume(1)
	//			this.tone.GetOSC(0).Trigger()
	//		}
	//	}

	// TODO Implement sending of frequency across to restalgia

	this.lastSoundTick = now

}

func (this *NetVDU) SetBGColourTriple(rr, gg, bb int) {

	// rr = rr & 0xff
	// gg = gg & 0xff
	// bb = bb & 0xff
	//
	// this.ThinEvents.SetBGColor(rr, gg, bb)
	// this.Server.SendThinScreenMessages(this.ThinEvents.GetEvents())
}

func (this *NetVDU) Beep() {

	pitch := 1000
	duration := 120

	instrument := "WAVE=SQUARE:ADSR=5,10,100,5"

	if instrument != "" {
		this.SendRestalgiaEvent(types.RestalgiaRedefineInstrument, instrument)
	}

	this.SendRestalgiaEvent(types.RestalgiaSoundEffect, utils.IntToStr(pitch))

	time.Sleep(time.Millisecond * time.Duration(duration))

}

func (this *NetVDU) ToggleControls() {
	this.ThinEvents.ToggleControls()
	this.Server.SendThinScreenMessages(this.ThinEvents.GetEvents())
}

func (this *NetVDU) GetCGColour() int {
	return this.CGColour
}

func (this *NetVDU) SetCGColour(c int) {

	// clip to palette
	c = c % this.VideoMode.Palette.Size()
	this.CGColour = c

	//this.VDUCore.SetBGColour(c)
	this.ThinEvents.SetBGColor(c)
	this.Server.SendThinScreenMessages(this.ThinEvents.GetEvents())
}

func (this *NetVDU) SendRestalgiaEvent(kind byte, content string) {
	this.Server.SendRestalgiaEvent(kind, content)
}

func (this *NetVDU) SetCurrentPage(c int) {
	this.VDUCore.SetCurrentPage(c)
	this.ThinEvents.CurrentPage(c)
}

func (this *NetVDU) SetDisplayPage(c int) {
	this.VDUCore.SetDisplayPage(c)
	this.ThinEvents.DisplayPage(c)
	this.Server.SendThinScreenMessages(this.ThinEvents.GetEvents())
}

func (this *NetVDU) CamResetAll() {
	c := 0
	this.ThinEvents.CamReset(c)
	this.Server.SendThinScreenMessages(this.ThinEvents.GetEvents())
}

func (this *NetVDU) CamLock() {
	c := 1
	this.ThinEvents.CamLock(c)
	this.Server.SendThinScreenMessages(this.ThinEvents.GetEvents())
}

func (this *NetVDU) CamUnlock() {
	c := 0
	this.ThinEvents.CamLock(c)
	this.Server.SendThinScreenMessages(this.ThinEvents.GetEvents())
}

func (this *NetVDU) CamResetLoc() {
	c := 1
	this.ThinEvents.CamReset(c)
	this.Server.SendThinScreenMessages(this.ThinEvents.GetEvents())
}

func (this *NetVDU) CamResetAngle() {
	c := 2
	this.ThinEvents.CamReset(c)
	this.Server.SendThinScreenMessages(this.ThinEvents.GetEvents())
}

func (this *NetVDU) CamDolly(f float32) {
	this.ThinEvents.CamDolly(f)
	this.Server.SendThinScreenMessages(this.ThinEvents.GetEvents())
}

func (this *NetVDU) CamMove(x, y, z float32) {
	this.ThinEvents.CamMove(x, y, z)
	this.Server.SendThinScreenMessages(this.ThinEvents.GetEvents())
}

func (this *NetVDU) CamOrbit(x, y float32) {
	this.ThinEvents.CamOrbit(x, y)
	this.Server.SendThinScreenMessages(this.ThinEvents.GetEvents())
}

func (this *NetVDU) CamRot(x, y, z float32) {
	this.ThinEvents.CamRotate(x, y, z)
	this.Server.SendThinScreenMessages(this.ThinEvents.GetEvents())
}


func (this *NetVDU) CamPos(x, y, z float32) {
	this.ThinEvents.CamPos(x, y, z)
	this.Server.SendThinScreenMessages(this.ThinEvents.GetEvents())
}

func (this *NetVDU) CamPivPnt(x, y, z float32) {
	this.ThinEvents.CamPivPnt(x, y, z)
	this.Server.SendThinScreenMessages(this.ThinEvents.GetEvents())
}

func (this *NetVDU) CamZoom(f float32) {
	this.ThinEvents.CamZoom(f)
	this.Server.SendThinScreenMessages(this.ThinEvents.GetEvents())
}

func (this *NetVDU) HgrShape(shape hires.ShapeEntry, x int, y int, scl int, deg int, c int, usecol bool) {
	hires.GetAppleHiRES().HgrShape(this, this.GetBitmapMemory()[this.GetCurrentPage()%2], shape, x, y, scl, deg, c, usecol)
	this.Server.SendThinScreenMessages(this.ThinEvents.GetEvents())
}

func (this *NetVDU) HgrFill(hc int) {
	hires.GetAppleHiRES().HgrFill(this.GetBitmapMemory()[this.GetCurrentPage()%2], hc)
	page := vduproto.APPLE_HIRES_PAGE_0 + byte(this.GetCurrentPage()%2)
	this.ThinEvents.Fill2DFull(page, hc)
	this.Server.SendThinScreenMessages(this.ThinEvents.GetEvents())
}

func (this *NetVDU) HgrPlotHold(x2, y2 int, hc int) {
	hires.GetAppleHiRES().HgrPlot(this.GetBitmapMemory()[this.GetCurrentPage()%2], x2, y2, hc)
	page := vduproto.APPLE_HIRES_PAGE_0 + byte(this.GetCurrentPage()%2)
	this.ThinEvents.Plot2D(page, x2, y2, hc)
}

func (this *NetVDU) HColorAt( x, y int ) int {
	return hires.GetAppleHiRES().HgrScreen(this.GetBitmapMemory()[this.GetCurrentPage()%2], x, y)
}

func (this *NetVDU) HgrPlot(x2, y2 int, hc int) {
	hires.GetAppleHiRES().HgrPlot(this.GetBitmapMemory()[this.GetCurrentPage()%2], x2, y2, hc)
	page := vduproto.APPLE_HIRES_PAGE_0 + byte(this.GetCurrentPage()%2)
	this.ThinEvents.Plot2D(page, x2, y2, hc)
	this.Server.SendThinScreenMessages(this.ThinEvents.GetEvents())
}

func (this *NetVDU) HgrLine(x1, y1, x2, y2 int, hc int) {
	hires.GetAppleHiRES().HgrLine(this.GetBitmapMemory()[this.GetCurrentPage()%2], x1, y1, x2, y2, hc)
	page := vduproto.APPLE_HIRES_PAGE_0 + byte(this.GetCurrentPage()%2)
	this.ThinEvents.Line2D(page, x1, y1, x2, y2, hc)
	this.Server.SendThinScreenMessages(this.ThinEvents.GetEvents())
}

func (this *NetVDU) GrPlot(x, y, c int) {
	this.ShadowTextMemory.Silent(true)
	this.VDUCore.GrPlot(x, y, c)
	this.ShadowTextMemory.Silent(false)
	this.ThinEvents.Plot2D(vduproto.APPLE_LORES_PAGE_0, x, y, c)
	this.Server.SendThinScreenMessages(this.ThinEvents.GetEvents())
}

func (this *NetVDU) GrVertLine(x, y0, y1, c int) {
	if this.VDUCore.VideoMode.ActualRows != this.VDUCore.VideoMode.Rows {
		this.ShadowTextMemory.Silent(true)
	}
	this.VDUCore.GrVertLine(x, y0, y1, c)
	if this.VDUCore.VideoMode.ActualRows != this.VDUCore.VideoMode.Rows {
		this.ShadowTextMemory.Silent(false)
	}
	this.ThinEvents.Line2D(vduproto.APPLE_LORES_PAGE_0, x, y0, x, y1, c)
	this.Server.SendThinScreenMessages(this.ThinEvents.GetEvents())
}

func (this *NetVDU) GrHorizLine(x0, x1, y, c int) {
	this.ShadowTextMemory.Silent(true)
	this.VDUCore.GrHorizLine(x0, x1, y, c)
	this.ShadowTextMemory.Silent(false)
	this.ThinEvents.Line2D(vduproto.APPLE_LORES_PAGE_0, x0, y, x1, y, c)
	this.Server.SendThinScreenMessages(this.ThinEvents.GetEvents())
}

func (this *NetVDU) RestoreVDUState() {
	this.VDUCore.RestoreVDUState()
	this.Server.SendScreenSpecification(this.GetVideoMode())
	this.Server.SendScreenMemoryChange(0, this.ShadowTextMemory.GetValues(0, this.ShadowTextMemory.Size()), this.CursorX, this.CursorY)
}
