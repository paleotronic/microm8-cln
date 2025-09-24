package main

import (
	"runtime"
	"sync"
	"time"

	"github.com/go-gl/gl/v2.1/gl"
	"paleotronic.com/core/glumbyclient/decalfont"
	"paleotronic.com/core/types"
	"paleotronic.com/core/vduconst"
	"paleotronic.com/core/vduproto"
	"paleotronic.com/utils"
	"paleotronic.com/glumby"
	"paleotronic.com/log"
)

var (
	light        *glumby.LightSource
	cammy, fxcam *glumby.Camera
	distance_z   float32
	vm           types.VideoMode
	cx, cy       int
	fontNormal   decalfont.DecalDuckFont
	fontInverted decalfont.DecalDuckFont
	screen       *decalfont.DecalDuckScreen
	c            *VDUClient
	//	a              *AudioClient
	strChan        chan vduproto.VDUServerEvent
	runeChan       chan vduproto.VDUServerEvent
	modeChan       chan vduproto.VDUServerEvent
	memChan        chan vduproto.VDUServerEvent
	posChan        chan vduproto.VDUServerEvent
	thinChan       chan vduproto.VDUServerEvent
	spkChan        chan vduproto.VDUServerEvent
	scanChan       chan vduproto.VDUServerEvent
	props          *glumby.WindowProperties
	thinEvents     *vduproto.ThinScreenEventBuffer
	scanEvents     []vduproto.ScanLineEvent
	dataMutex      sync.Mutex
	scanMutex      sync.Mutex
	modeChanged    bool
	memUpdates     []vduproto.ScreenMemoryEvent
	lastP0, lastP1 int
)

func init() {
	runtime.LockOSThread()
}

func SetupKeymapper() glumby.KeyTranslationMap {

	km := glumby.NewDefaultMapper()

	km = append(km, glumby.KeyState{Key: glumby.KeyF2, States: []glumby.Action{glumby.Release, glumby.Repeat}, Mapping: vduconst.F2})

	km = append(km, glumby.KeyState{Key: glumby.KeyDown, States: []glumby.Action{glumby.Release, glumby.Repeat}, Mapping: vduconst.CSR_DOWN})
	km = append(km, glumby.KeyState{Key: glumby.KeyUp, States: []glumby.Action{glumby.Release, glumby.Repeat}, Mapping: vduconst.CSR_UP})
	km = append(km, glumby.KeyState{Key: glumby.KeyLeft, States: []glumby.Action{glumby.Release, glumby.Repeat}, Mapping: vduconst.CSR_LEFT})
	km = append(km, glumby.KeyState{Key: glumby.KeyRight, States: []glumby.Action{glumby.Release, glumby.Repeat}, Mapping: vduconst.CSR_RIGHT})

	km = append(km, glumby.KeyState{Key: glumby.KeyPageDown, States: []glumby.Action{glumby.Release, glumby.Repeat}, Mapping: vduconst.PAGE_DOWN})
	km = append(km, glumby.KeyState{Key: glumby.KeyPageUp, States: []glumby.Action{glumby.Release, glumby.Repeat}, Mapping: vduconst.PAGE_UP})
	km = append(km, glumby.KeyState{Key: glumby.KeyHome, States: []glumby.Action{glumby.Release, glumby.Repeat}, Mapping: vduconst.HOME})
	km = append(km, glumby.KeyState{Key: glumby.KeyEnd, States: []glumby.Action{glumby.Release, glumby.Repeat}, Mapping: vduconst.END})

	return km

}

func main() {
	runtime.GOMAXPROCS(6 * runtime.NumCPU())

	props = glumby.NewWindowProperties()

	props.Title = "Super-8 VDU Prototype (golang)"
	props.Width = 960
	props.Height = 720
	props.KeyMapper = SetupKeymapper()
	w := glumby.NewWindow(props)

	w.OnCreate = OnCreateWindow
	w.OnDestroy = OnDestroyWindow
	w.OnRender = OnRenderWindow
	w.OnEvent = OnEventWindow
	w.OnKey = OnKeyEvent
	w.OnChar = OnCharEvent
	w.OnMouseMove = OnMouseMoveEvent
	w.OnMouseButton = OnMouseButtonEvent
	w.OnScroll = OnScrollEvent

	thinEvents = vduproto.NewThinScreenEventBuffer()
	memUpdates = make([]vduproto.ScreenMemoryEvent, 0)

	// Client
	c = NewVDUClient("localhost", ":9988")
	//a = NewAudioClient("localhost", ":9988")
	// event channels
	strChan = make(chan vduproto.VDUServerEvent)
	runeChan = make(chan vduproto.VDUServerEvent)
	modeChan = make(chan vduproto.VDUServerEvent)
	spkChan = make(chan vduproto.VDUServerEvent, 100)
	memChan = make(chan vduproto.VDUServerEvent, 100)
	posChan = make(chan vduproto.VDUServerEvent, 100)
	thinChan = make(chan vduproto.VDUServerEvent, 100)
	scanChan = make(chan vduproto.VDUServerEvent, 100)
	c.RegisterMessageType("SOE", strChan)
	c.RegisterMessageType("COE", runeChan)
	c.RegisterMessageType("SFE", modeChan)
	c.RegisterMessageType("SME", memChan)
	c.RegisterMessageType("SPE", posChan)
	c.RegisterMessageType("TSE", thinChan)
	c.RegisterMessageType("HGS", scanChan)
	//a.RegisterMessageType("SPK", spkChan)

	scanEvents = make([]vduproto.ScanLineEvent, 0)

	c.Connect()
	//a.Connect()

	time.Sleep(time.Millisecond * 100)

	c.SendScreenStateRequest()

	//InitAudio()

	w.Run()
	c.Close()

}

func ReceiveVDUEvents() {
	for {
		select {
		case cout := <-scanChan:
			z := cout.Data.(vduproto.ScanLineEvent)
			scanMutex.Lock()
			scanEvents = append(scanEvents, z)
			scanMutex.Unlock()
		case cout := <-posChan:
			z := cout.Data.(vduproto.ScreenPositionEvent)
			screen.CX = float32(z.X)
			screen.CY = float32(z.Y)
		case mout := <-modeChan:
			z := mout.Data.(types.VideoMode)
			vm = z
			modeChanged = true
		case tout := <-thinChan:
			//log.Println("Received thin events")
			z := tout.Data.(vduproto.ThinScreenEventList)
			for _, v := range z {
				thinEvents.Add(v)
				//log.Println(v)
			}
		case mout := <-memChan:
			z := mout.Data.(vduproto.ScreenMemoryEvent)

			// If there is a pending mode change, we wait... it will be processed by the GL thread
			// then we can update the display memory

			//for modeChanged {
			//	time.Sleep(time.Millisecond * 1)
			//}

			dataMutex.Lock()
			memUpdates = append(memUpdates, z)
			dataMutex.Unlock()

			//	screen.RedrawNextFrame = false

			//screen.TextMemory.PurgeLog()
		}
	}
}

func OnCreateWindow(w *glumby.Window) {

	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.LIGHTING)
	//gl.Enable(gl.COLOR_MATERIAL)
	//gl.ColorMaterial(gl.FRONT, gl.AMBIENT)
	gl.ClearColor(0.9, 0.5, 0.5, 0.0)
	gl.ClearDepth(1)
	gl.DepthFunc(gl.LEQUAL)

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	// Lights
	// Add a glumby light
	ambient := []float32{0.5, 0.5, 0.5, 1}
	diffuse := []float32{1, 1, 1, 1}
	lightPosition := []float32{-10, 10, 10, 0}
	light = glumby.NewLightSource(glumby.Light0, ambient, diffuse, lightPosition)
	light.On()

	// Camera
	ww, hh := w.GetGLFWWindow().GetFramebufferSize()
	aspect := float64(ww) / float64(hh)

	cammy = glumby.NewCamera(glumby.Rect{-1 * aspect, 1 * aspect, -1, 1}, 0.1, 1000, true)
	cammy.SetPos(256, 192, 20.3)

	fxcam = glumby.NewCamera(glumby.Rect{-1 * aspect, 1 * aspect, -1, 1}, 0.1, 1000, true)
	fxcam.SetPos(256, 192, 20.3)

	// fonts
	fontNormal := decalfont.LoadNormalFont()
	fontInverted := decalfont.LoadInvertedFont()

	// screen
	s := decalfont.NewDecalDuckScreen(512, 384, 40, 24, fontNormal, fontInverted)
	screen = s

	go ReceiveVDUEvents()
}

func OnDestroyWindow(w *glumby.Window) {
	log.Println("Destroy called")
}

func processThinEvents() {
	evlist := thinEvents.GetEvents()

	if len(evlist) == 0 {
		return
	}

	//log.Printf("Processing thin events [%v]\n", evlist)

	for _, ev := range evlist {
		switch ev.ID {
		case vduproto.CurrentPage:
			screen.CurrentPage = ev.C
		case vduproto.DisplayPage:
			screen.DisplayPage = ev.C
		case vduproto.Fill2D:
			if (screen.Mode == nil) || (screen.Mode.Width < 50) {
				//
			} else {
				screen.HFill2D(ev.C)
			}
		case vduproto.Plot2D:
			if (screen.Mode == nil) || (screen.Mode.Width < 50) {
				screen.Plot2D(ev.X0, ev.Y0, ev.C)
			} else {
				screen.HPlot2D(ev.X0, ev.Y0, ev.C)
			}
			//log.Printf("screen.Plot2D(%d, %d, %d)\n", ev.X0, ev.Y0, ev.C)
		case vduproto.Line2D:
			if (screen.Mode == nil) || (screen.Mode.Width < 50) {
				screen.Line2D(ev.X0, ev.Y0, ev.X1, ev.Y1, ev.C)
			} else {
				screen.HLine2D(ev.X0, ev.Y0, ev.X1, ev.Y1, ev.C)
			}
		}
	}
}

func processScanEvents() {

	scanMutex.Lock()
	defer scanMutex.Unlock()

	for _, sle := range scanEvents {
		offs := screen.WozHGR[screen.CurrentPage].XYToOffset(0, int(sle.Y))

		if len(sle.Data) < 40 {
			panic("Not enough scanline " + utils.IntToStr(len(sle.Data)))
		}

		for idx, b := range sle.Data {
			screen.WozHGR[screen.CurrentPage].Data[offs+idx] = b
		}
	}

	scanEvents = make([]vduproto.ScanLineEvent, 0)

}

func OnRenderWindow(w *glumby.Window) {

	if modeChanged {
		log.Printf("Changing mode %v\n", vm)
		modeChanged = false
		screen.ApplyMode(&vm)
	}

	processThinEvents()

	processScanEvents()

	for _, z := range memUpdates {
		screen.TextMemory.SetValues(z.Offset, z.Content)
		screen.RedrawNextFrame = true
		screen.CX = float32(z.X)
		screen.CY = float32(z.Y)
		dataMutex.Lock()
		memUpdates = make([]vduproto.ScreenMemoryEvent, 0)
		dataMutex.Unlock()
	}

	//gl.CullFace(gl.BACK)

	gl.ClearColor(0, 0, 0, 1)
	gl.ClearDepth(1)

	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	//cammy.RotateZ(0.02)
	//cammy.RotateX(0.02)

	gl.Color4f(1, 1, 1, 0.9)

	//now := time.Now().UnixNano()

	// screen render here
	//screen.Render()

	gl.PushMatrix()
	//fxcam.RotateZ(0.02)
	fxcam.Apply()
	if (screen.Mode != nil) && (screen.Mode.ActualRows != screen.Mode.Rows) {
		if screen.Mode.Width > 50 {
			if decalfont.USEWOZHGR {
				screen.RenderAppleIIHGRWoz(screen.DisplayPage)
			} else {
				screen.RenderAppleIIHGR(screen.DisplayPage)
			}
		} else {
			screen.RenderAppleIIGR()
		}
	}
	gl.PopMatrix()

	gl.PushMatrix()
	cammy.Apply()
	screen.RedrawIfRequired()
	screen.RenderCursor()
	gl.PopMatrix()

	//after := time.Now().UnixNano()
	//log.Printf("Cubez drawn in %v ms\n", ((after - now) / 1000000))

	w.SwapBuffers()
}

func OnEventWindow(w *glumby.Window) {
	log.Println("Event called")
}

func OnKeyEvent(w *glumby.Window, ch glumby.Key, mod glumby.ModifierKey, state glumby.Action) {
	log.Printf("Keyevent: ch = %d, mod = %d, state = %d\n", ch, mod, state)
}

func OnCharEvent(w *glumby.Window, ch rune) {
	log.Printf("Keyevent: ch = %d\n", ch)

	//w.SetTitle(fmt.Sprintf("You have pressed key [%d]", ch))
	c.SendKeyPress(ch)
}

func OnMouseMoveEvent(w *glumby.Window, x, y float64) {
	//	log.Printf("Mouse at (%f, %f)\n", x, y)

	// x,y are pixels in window / screen size
	ww, hh := w.GetGLFWWindow().GetSize()
	p1 := int((x / float64(ww)) * 255)
	p0 := int((y / float64(hh)) * 255)

	if p0 != lastP0 {
		c.SendPaddleValue(0, byte(p0))
	}
	if p1 != lastP1 {
		c.SendPaddleValue(1, byte(p1))
	}

	lastP0 = p0
	lastP1 = p1

}

func OnMouseButtonEvent(w *glumby.Window, button glumby.MouseButton, action glumby.Action, mod glumby.ModifierKey) {
	log.Printf("Mouse button %d, action = %d, mod = %d\n", button, action, mod)

	if button == glumby.MouseButtonLeft {
		switch action {
		case glumby.Press:
			c.SendPaddleButton(0, 1)
		case glumby.Repeat:
			c.SendPaddleButton(0, 1)
		case glumby.Release:
			c.SendPaddleButton(0, 0)
		}
	}

	if button == glumby.MouseButtonRight {
		switch action {
		case glumby.Press:
			c.SendPaddleButton(1, 1)
		case glumby.Repeat:
			c.SendPaddleButton(1, 1)
		case glumby.Release:
			c.SendPaddleButton(1, 0)
		}
	}
}

func OnScrollEvent(w *glumby.Window, xoff, yoff float64) {
	log.Printf("Scroll x = %f, y = %f\n", xoff, yoff)
	distance_z += float32(yoff)

	c.SendPaddleModify(0, int8(yoff*8))
}
