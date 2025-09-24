package glumby

import (
	"paleotronic.com/log"
	"math" //	"sync"
	"time"

	//	log2 "log"

	"paleotronic.com/fmt"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl64"
	"paleotronic.com/core/settings"
	"paleotronic.com/gl"
	"paleotronic.com/octalyzer/bus" //"os"
	//"runtime"
)

const (
	VERSION_MAJOR = 2
	VERSION_MINOR = 0
	USE_TICKER    = true
)

// Represents Window properties
type WindowProperties struct {
	Resizable     bool
	Fullscreen    bool
	CursorHidden  bool
	VersionMinor  int
	VersionMajor  int
	Width, Height int
	Title         string
	Subtitle      string
	FPS           int
	KeyMapper     KeyTranslationMap
	NoBorder      bool
}

// Represents a window with an opengl context
type Window struct {
	desktopWidth, desktopHeight int
	oX, oY, oW, oH              int
	window                      *glfw.Window
	properties                  *WindowProperties
	Controllers                 []*Controller
	// events
	OnVideoSyncRequested  func(w *Window)
	OnCreate              func(w *Window)
	OnRender              func(w *Window)
	OnEvent               func(w *Window)
	OnDestroy             func(w *Window)
	OnKey                 func(w *Window, scancode Key, mod ModifierKey, state Action)
	OnRawKey              func(w *Window, key Key, scancode int, mod ModifierKey, state Action)
	OnChar                func(w *Window, ch rune)
	OnDrop                func(w *Window, names []string, fake bool)
	OnMouseMove           func(w *Window, xpos, ypos float64)
	OnMouseButton         func(w *Window, button MouseButton, action Action, mod ModifierKey)
	OnScroll              func(w *Window, xoff, yoff float64)
	OnResize              func(w *Window, width, height int)
	OnFocusChanged        func(w *Window, focused bool)
	isFullScreen          bool
	syncCount             int
	lastMx, lastMy        float64
	mx, my                float64
	cx, cy                float64
	lastDPadSample        time.Time
	lastMouseMove         time.Time
	lastVector, padVector *mgl64.Vec2
	haveSamples           bool
	hasFocus              bool

	FrameCount int
}

func (w *Window) KeyPressed(key Key) bool {
	return w.GetGLFWWindow().GetKey(glfw.Key(key)) == glfw.Press
}

func (w *Window) GetMouseButton(mb MouseButton) Action {
	state := Action(w.GetGLFWWindow().GetMouseButton(glfw.MouseButton(mb)))
	return state
}

func (w *Window) GetDesktopSize() (int, int) {

	//w.desktopWidth, w.desktopHeight = glfw.G

	m := glfw.GetPrimaryMonitor().GetVideoMode()

	w.desktopWidth, w.desktopHeight = m.Width, m.Height

	return w.desktopWidth, w.desktopHeight

}

func (w *Window) IsFullscreen() bool {
	ww, hh := w.GetDesktopSize()

	fw, fh := w.GetGLFWWindow().GetSize()

	return (ww == fw && hh == fh)
}

func (w *Window) Save() {
	w.oX, w.oY = w.GetGLFWWindow().GetPos()
	w.oW, w.oH = w.GetGLFWWindow().GetSize()
}

func (w *Window) GetFullscreen() bool {
	return w.isFullScreen
}

func (w *Window) SetFullscreen(fullscreen bool) {

	if w.isFullScreen == fullscreen {
		return
	}

	if fullscreen {
		monitors := glfw.GetMonitors()

		// current glfw window pos
		winPosX, winPosY := w.GetGLFWWindow().GetPos()

		// save old positions
		w.Save()

		var closestIndex int = -1
		var monitorIndex int = -1
		var closestDist float64 = 99999999

		for index, mon := range monitors {

			if mon == nil {
				continue
			}

			vidmode := mon.GetVideoMode()
			X1, Y1 := mon.GetPos()
			X2, Y2 := X1+vidmode.Width-1, Y1+vidmode.Height-1

			dist := math.Sqrt(float64((X1-winPosX)*(X1-winPosX)) + float64((Y1-winPosY)*(Y1-winPosY)))
			if dist < closestDist {
				closestDist = dist
				closestIndex = -1
			}

			if winPosX >= X1 && winPosX <= X2 && winPosY >= Y1 && winPosY <= Y2 {
				monitorIndex = index
				break
			}
		}

		// Should have index set now...
		if monitorIndex == -1 {
			monitorIndex = closestIndex
		}

		if monitorIndex >= 0 {
			vidmode := monitors[monitorIndex].GetVideoMode()
			w.GetGLFWWindow().SetMonitor(monitors[monitorIndex], 0, 0, vidmode.Width, vidmode.Height, vidmode.RefreshRate)
		}

	} else {
		w.GetGLFWWindow().SetMonitor(nil, w.oX, w.oY, w.oW, w.oH, 0)
		w.SetTitle(w.properties.Title)
	}

	w.GetGLFWWindow().Focus()

	w.isFullScreen = fullscreen
}

func NewWindowProperties() *WindowProperties {
	this := &WindowProperties{Width: 640, Height: 480, VersionMajor: 2, VersionMinor: 1, Title: "Glumby Window", Resizable: false, KeyMapper: NewDefaultMapper()}
	return this
}

// Accessor for the underlying window data
func (w *Window) GetGLFWWindow() *glfw.Window {
	return w.window
}

// Get Window Properties
func (w *Window) GetProperties() *WindowProperties {
	return w.properties
}

// Cleanup
func (w *Window) Close() {
	if w.OnDestroy != nil {
		w.OnDestroy(w)
	}
	glfw.Terminate()
}

func (w *Window) Create() {
	if w.OnCreate != nil {
		w.OnCreate(w)
	}
	fmt.Println("end of w.Create()")
}

func (w *Window) Render() {

	if settings.VideoSuspended {
		//log.Printf("Window render skipping as video suspended")
		return
	}

	// clear
	w.window.MakeContextCurrent()
	gl.ClearColor(0, 0, 1, 1)
	gl.ClearDepth(32)

	if w.OnRender != nil {
		//log.Printf("Render frame calling")
		w.OnRender(w)
		w.FrameCount++
	}

	w.window.SwapBuffers()
}

// NewWindow creates a new main window
func NewWindow(p *WindowProperties) *Window {

	if p == nil {
		p = &WindowProperties{
			Resizable:    false,
			Width:        800,
			Height:       600,
			Title:        "Glumby Window",
			VersionMinor: VERSION_MINOR,
			VersionMajor: VERSION_MAJOR,
			KeyMapper:    NewDefaultMapper(),
		}
	}

	this := Window{}
	this.properties = p

	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}

	if p.Resizable {
		glfw.WindowHint(glfw.Resizable, glfw.True)
	} else {
		glfw.WindowHint(glfw.Resizable, glfw.False)
	}

	glfw.WindowHint(glfw.ContextVersionMajor, p.VersionMajor)
	glfw.WindowHint(glfw.ContextVersionMinor, p.VersionMinor)
	glfw.WindowHint(glfw.DoubleBuffer, glfw.True)
	if p.NoBorder {
		glfw.WindowHint(glfw.Decorated, glfw.False)
		glfw.WindowHint(glfw.Floating, glfw.True)
	}

	var err error
	this.window, err = glfw.CreateWindow(p.Width, p.Height, p.Title, nil, nil)
	if err != nil {
		panic("Glumby init failed: " + err.Error())
	}

	this.window.MakeContextCurrent()
	glfw.SwapInterval(1)
	this.Save()

	if p.Fullscreen {
		this.SetFullscreen(true)
	}

	this.window.SetKeyCallback(this.handleKey)
	this.window.SetCharCallback(this.handleChar)
	this.window.SetDropCallback(this.handleDrop)
	this.window.SetCursorPosCallback(this.handleMouseMove)
	this.window.SetMouseButtonCallback(this.handleMouseButton)
	this.window.SetScrollCallback(this.handleScroll)
	this.window.SetFramebufferSizeCallback(this.handleSize)
	this.window.SetDropCallback(this.handleDrop)
	this.window.SetFocusCallback(this.handleFocus)

	if err = gl.Init(); err != nil {
		panic("Glumby GL init failed: " + err.Error())
	}

	this.Controllers = EnumerateControllers()

	if this.GetProperties().CursorHidden {
		this.window.SetInputMode(glfw.CursorMode, glfw.CursorHidden)
	}

	return &this
}

func (ww *Window) MuteKeys() {
	ww.GetGLFWWindow().SetKeyCallback(nil)
}

func (ww *Window) UnmuteKeys() {
	ww.GetGLFWWindow().SetKeyCallback(ww.handleKey)
}

func (ww *Window) handleFocus(w *glfw.Window, f bool) {
	ww.hasFocus = f
	if ww.OnFocusChanged != nil {
		ww.OnFocusChanged(ww, f)
	}
}

func (ww *Window) IsFocused() bool {
	return ww.hasFocus
}

func (ww *Window) handleSize(w *glfw.Window, width int, height int) {

	width, height = w.GetFramebufferSize()

	if ww.OnResize != nil {
		ww.OnResize(ww, width, height)
	}
}

func (ww *Window) handleScroll(w *glfw.Window, xoff float64, yoff float64) {
	if ww.OnScroll != nil {
		ww.OnScroll(ww, xoff, yoff)
	}
}

func (ww *Window) SimulateKey(key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	ww.handleKey(ww.window, key, scancode, action, mods)
}

func (ww *Window) handleKey(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {

	//log2.Printf("check mapping for key=%d, mods=%d, action=%d\n", key, mods, action)

	if ww.OnRawKey != nil {
		ww.OnRawKey(ww, Key(key), scancode, ModifierKey(mods), Action(action))
	}

	if key >= glfw.KeyA && key <= glfw.KeyZ && (mods == 0 || ModifierKey(mods) == ModShift) {
		return
	}

	if ww.OnKey != nil {
		// this is rather cool
		r, ok := ww.properties.KeyMapper.GetMapping(Key(key), ModifierKey(mods), Action(action))
		if ok && (Action(action) == Press || Action(action) == Repeat || Action(action) == Release) {
			ww.OnKey(ww, Key(r), ModifierKey(mods), Action(action))
		} else if int(key) < 256 {
			ww.OnKey(ww, Key(key), ModifierKey(mods), Action(action))
		}
	}

}

func (ww *Window) handleChar(w *glfw.Window, key rune) {

	//log2.Printf("CHAR keycode = %d\n", key)

	if (key >= 'A' && key <= 'Z') || (key >= 'a' && key <= 'z') {

		if ww.OnKey != nil {
			// this is rather cool
			ww.OnKey(ww, Key(key), ModifierKey(0), Action(Press))
		}
	}

}

func (ww *Window) handleDrop(w *glfw.Window, names []string) {
	if ww.OnDrop != nil {
		ww.OnDrop(ww, names, false)
	}
}

func (ww *Window) handleMouseMove(w *glfw.Window, xpos, ypos float64) {
	//ww.lastMx, ww.lastMy = ww.mx, ww.my
	ww.mx, ww.my = xpos, ypos
	ww.lastMouseMove = time.Now()
	if ww.OnMouseMove != nil {
		ww.OnMouseMove(ww, xpos, ypos)
	}
	//	ww.SampleDPAD()
}

func (ww *Window) handleMouseButton(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {

	if ww.OnMouseButton != nil {
		ww.OnMouseButton(ww, MouseButton(button), Action(action), ModifierKey(mod))
	}

}

// Mainloop
func (w *Window) Run() {

	fmt.Println("in w.Run()")

	w.Create()

	defer w.Close()

	fmt.Println("after w.Create()")

	//	var nanosPerFrame int64 = 0
	//	if w.GetProperties().FPS != 0 {
	//		nanosPerFrame = 1000000000 / int64(w.GetProperties().FPS)
	//	}
	//	var lastRender time.Time = time.Now()
	//var diff time.Duration = time.Duration(nanosPerFrame)
	var diff60 time.Duration = time.Duration(time.Second / time.Duration(settings.FPSClock))
	var ticker = time.NewTicker(diff60)
	//var syncCount int
	settings.FrameSkip = settings.FPSClock/w.GetProperties().FPS - 1
	settings.DefaultFrameSkip = settings.FrameSkip

	// We tick 60 times per second...
	bus.StartClock(diff60)
	var startRender time.Time
	var renderTime time.Duration
	var targetDuration time.Duration
	var votes int
	//	var dpcount int

	for !w.window.ShouldClose() {

		targetDuration = time.Second / time.Duration(settings.FPSClock/(settings.FrameSkip+1))

		select {
		case <-ticker.C:
			//			dpcount++
			//			if dpcount > 5 {
			//				dpcount = 0
			//				w.SampleDPAD()
			//			}
			if time.Since(w.lastMouseMove) > 50*time.Millisecond {
				w.SampleDPAD()
				w.lastMouseMove = time.Now()
			}

			if w.syncCount > 0 && (w.syncCount >= (settings.FrameSkip + 1)) {
				//fmt.Printf(".")
				startRender = time.Now()
				bus.SyncDo(w.Render)
				renderTime = time.Since(startRender)
				w.syncCount = 0

				// check whether we are meeting our rendering targets
				if renderTime >= targetDuration {
					votes += settings.FSVoteUp
					//fmt.Printf("fs votes=%d\n", votes)
					//fmt.Printf("Recommend increasing frameskip value... target=%v, actual=%v\n", targetDuration, renderTime)
				} else {
					votes -= settings.FSVoteDown
					if votes < 0 {
						votes = 0
					}
					//fmt.Printf("fs votes=%d\n", votes)
				}

				if votes > settings.FSVoteThreshold {
					settings.FrameSkip++
					// title := fmt.Sprintf(w.properties.Title+" (frameskip=%d)", settings.FrameSkip)
					// w.SetTitle(title)
					votes = 0
				}
			}
			glfw.PollEvents()

		}

	}

	bus.StopClock()

}

// Mainloop
func (w *Window) RunSimple() {

	fmt.Println("in w.Run()")

	w.Create()

	defer w.Close()

	fmt.Println("after w.Create()")

	//	var nanosPerFrame int64 = 0
	//	if w.GetProperties().FPS != 0 {
	//		nanosPerFrame = 1000000000 / int64(w.GetProperties().FPS)
	//	}
	//	var lastRender time.Time = time.Now()
	//var diff time.Duration = time.Duration(nanosPerFrame)
	var diff60 time.Duration = time.Duration(time.Second / time.Duration(settings.FPSClock))
	var ticker = time.NewTicker(diff60)
	//var syncCount int
	settings.FrameSkip = settings.FPSClock/w.GetProperties().FPS - 1
	settings.DefaultFrameSkip = settings.FrameSkip

	// We tick 60 times per second...
	//bus.StartClock(diff60)
	//var startRender time.Time
	//var renderTime time.Duration
	//var targetDuration time.Duration
	//var votes int
	//	var dpcount int

	for !w.window.ShouldClose() {

		//targetDuration = time.Second / time.Duration(settings.FPSClock/(settings.FrameSkip+1))

		select {
		case <-ticker.C:
			//fmt.Println("tick")
			w.Render()
			glfw.PollEvents()
		}

	}

	fmt.Println("window closed")

	//bus.StopClock()

}

func (w *Window) GetPADValues() (int, int) {
	if w.padVector == nil {
		return 0, 0
	}
	return int(w.padVector[0]), int(w.padVector[1])
}

func (w *Window) CenterPAD() {
	w.lastVector = nil
	w.lastMx, w.lastMy = w.mx, w.my
	w.lastDPadSample = time.Now()
	w.padVector = &mgl64.Vec2{0, 0}
}

func normalize(v *mgl64.Vec2) *mgl64.Vec2 {

	//fmt.Printf("normalize(%v)\n", v)

	ax := math.Abs(v[0])
	ay := math.Abs(v[1])

	nv := &mgl64.Vec2{0, 0}
	if ax > 0.6 {
		nv[0] = v[0] / ax
	}
	if ay > 0.6 {
		nv[1] = v[1] / ay
	}

	return nv

}

func (w *Window) SetCursorDisabled(b bool) {
	if b {
		w.window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	} else {
		w.window.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
	}
}

func (w *Window) SampleDPAD() {

	// Border around the window overrides the standard DPAD sampling
	ww, hh := w.GetGLFWWindow().GetFramebufferSize()
	border := int(float64(ww) * 0.03)
	//fmt.Printf("border=%d\n", border)
	if int(w.mx) < border {
		w.padVector = &mgl64.Vec2{-1, 0}
		w.lastMx, w.lastMy = w.mx, w.my
		w.lastVector = nil
		return
	} else if int(w.mx) >= ww-border {
		w.padVector = &mgl64.Vec2{1, 0}
		w.lastMx, w.lastMy = w.mx, w.my
		w.lastVector = nil
		return
	} else if int(w.my) < border {
		w.padVector = &mgl64.Vec2{0, -1}
		w.lastMx, w.lastMy = w.mx, w.my
		w.lastVector = nil
		return
	} else if int(w.my) >= hh-border {
		w.padVector = &mgl64.Vec2{0, 1}
		w.lastMx, w.lastMy = w.mx, w.my
		w.lastVector = nil
		return
	}

	v := &mgl64.Vec2{w.mx - w.lastMx, w.my - w.lastMy}

	if v.Len() < 0.1 {
		return
	}

	tmp := v.Normalize()
	vn := normalize(&tmp)
	if v.Len() > 16 {
		//fmt.Printf("Swipe normal = %v\n", vn)

		if w.padVector == nil {
			w.padVector = &mgl64.Vec2{0, 0}
		}

		if w.lastVector == nil {
			w.lastVector = &mgl64.Vec2{0, 0}
		}

		// Check if we are re-centering the dpad
		if w.padVector.Add(*vn).Len() < 0.1 && v.Len() < math.Max(1.5*w.lastVector.Len(), 100) {
			w.padVector = &mgl64.Vec2{0, 0}
			//			fmt.Println("Center PAD")
		} else {
			w.padVector = vn
			//			if vn[1] < -0.1 {
			//				fmt.Print("UP")
			//			} else if vn[1] > 0.1 {
			//				fmt.Print("DOWN")
			//			}
			//			if vn[0] < -0.1 {
			//				fmt.Print("LEFT")
			//			} else if vn[0] > 0.1 {
			//				fmt.Print("RIGHT")
			//			}
			//			fmt.Println()
		}

	}

	// save current pos as last pos
	w.lastMx, w.lastMy = w.mx, w.my
	w.lastVector = v

}

func (w *Window) SampleDPADOld() {

	if w.mx == w.lastMx && w.my == w.lastMy {
		return
	}

	v := &mgl64.Vec2{w.mx - w.lastMx, w.my - w.lastMy}

	if w.lastVector == nil {
		w.lastVector = v
		return
	}

	diffvec := w.lastVector.Add(*v)
	//fmt.Printf("dv=%v\n", diffvec)

	// dist >= 5
	// get direction of movement...
	//	if time.Since(w.lastDPadSample) > 5*time.Second {
	//		w.CenterPAD()
	//		return
	//	}

	if diffvec.Len() > 16 {
		vv := diffvec.Normalize()
		z := &mgl64.Vec2{0, 0}
		if math.Abs(vv[0]) >= 0.5 {
			z[0] = vv[0] / math.Abs(vv[0])
		}
		if math.Abs(vv[1]) >= 0.5 {
			z[1] = vv[1] / math.Abs(vv[1])
		}
		w.padVector = z
	} else {
		w.padVector = &mgl64.Vec2{0, 0}
		//w.lastVector = nil // force dpad recenter
	}

	w.lastVector = v

	w.lastDPadSample = time.Now()

	fmt.Printf("Pad vector: %v\n", w.padVector)
}

func (w *Window) Synced() {
	w.syncCount++
}

func (w *Window) SwapBuffers() {
	//if runtime.GOOS != "windows"  &&{
	//	w.window.SwapBuffers()
	//}
}

func (w *Window) SetTitle(s string) {
	w.properties.Title = s
	w.updateTitle()
}

func (w *Window) SetSubtitle(s string) {
	w.properties.Subtitle = s
	w.updateTitle()
}

func (w *Window) updateTitle() {
	if w.properties.Title != "" && w.properties.Subtitle != "" {
		w.window.SetTitle(fmt.Sprintf("%s - %s", w.properties.Title, w.properties.Subtitle))
	} else if w.properties.Title != "" && w.properties.Subtitle == "" {
		w.window.SetTitle(fmt.Sprintf("%s", w.properties.Title))
	} else if w.properties.Subtitle != "" {
		w.window.SetTitle(w.properties.Subtitle)
	} else {
		w.window.SetTitle("Untitled")
	}
}
