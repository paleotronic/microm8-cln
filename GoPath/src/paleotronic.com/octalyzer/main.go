//go:build !remint
// +build !remint

package main

//  trivial comment

import (
	"bytes"
	"flag"
	fmt2 "paleotronic.com/fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"paleotronic.com/core/interfaces"
	pnc "paleotronic.com/panic"

	//log2 "log"

	s8webclient "paleotronic.com/api"
	"paleotronic.com/core/dialect/plus"
	"paleotronic.com/core/editor"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/hardware/common"
	"paleotronic.com/core/hardware/cpu/mos6502"
	"paleotronic.com/core/hardware/cpu/z80"
	"paleotronic.com/core/hardware/restalgia" //"net/http"
	"paleotronic.com/core/hardware/servicebus"
	"paleotronic.com/core/instrument"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/core/types/glmath"
	"paleotronic.com/core/vduconst"
	"paleotronic.com/debugger"
	"paleotronic.com/files"
	"paleotronic.com/fmt"
	"paleotronic.com/gl"
	"paleotronic.com/glumby"
	"paleotronic.com/log"
	"paleotronic.com/microtracker"
	"paleotronic.com/octalyzer/assets"
	"paleotronic.com/octalyzer/backend"
	"paleotronic.com/octalyzer/bus"
	"paleotronic.com/octalyzer/clientperipherals"
	"paleotronic.com/octalyzer/ui"
	"paleotronic.com/octalyzer/ui/chat"
	"paleotronic.com/octalyzer/video" //"paleotronic.com/core/hardware/servicebus"

	// "net/http"
	"paleotronic.com/restalgia/driver"
	"paleotronic.com/update"
	"paleotronic.com/utils"
)

const (
	PADDLE_BORDER = 16
	//CRTASPECT = types.CWIDTH / types.CHEIGHT
)

var (
	ramActiveState                        [memory.OCTALYZER_NUM_INTERPRETERS]uint64
	MESHZOOM                              float32
	RAM                                   *memory.MemoryMap
	light, light2                         *glumby.LightSource
	osd                                   [memory.OCTALYZER_NUM_INTERPRETERS]*glumby.PerspCamera
	pcam                                  [memory.OCTALYZER_NUM_INTERPRETERS]*glumby.PerspCamera
	fxcam                                 [memory.OCTALYZER_NUM_INTERPRETERS][memory.OCTALYZER_MAPPED_CAM_GFXCOUNT]*glumby.PerspCamera
	distance_z                            float32
	cx, cy                                int
	BGColor                               color.RGBA = color.RGBA{0, 0, 0, 255}
	props                                 *glumby.WindowProperties
	dataMutex                             sync.Mutex
	scanMutex                             sync.Mutex
	modeChanged                           bool
	lastP0, lastP1                        int
	filecache                             map[[16]byte]*files.FilePack
	splash                                *video.Decal
	splashTexture                         *glumby.Texture
	unified                               *video.Decal
	unifiedTexture                        *glumby.Texture
	ovpath                                string
	ovdecal, undecal                      *video.Decal
	ovTexture, uiTexture, unTexture       *glumby.Texture
	musicPath                             [memory.OCTALYZER_NUM_INTERPRETERS]string
	bgpath                                string
	bgcamidx                              int
	bgdecal                               *video.Decal
	bgTexture                             *glumby.Texture
	bgOpacity                             float32 = 0.25
	bgAspect                              float32
	bgZoom                                float64 = 1.0
	bgZoomFactor                          float64
	bgX, bgY, bgZ                         float64
	bgCamTrack                            bool
	w                                     *glumby.Window
	splashdata                            []byte
	splashwidth, splashheight, splashxoff float32
	modifier                              bool
	buffer                                int
	HUDLayers                             map[int][]*video.TextLayer
	GraphicsLayers                        map[int][]*video.GraphicsLayer
	HUDSpecs                              map[int][]*types.LayerSpecMapped
	GFXSpecs                              map[int][]*types.LayerSpecMapped
	FlipCase                              bool
	playfield, playfieldOSD               glumby.PRect
	tspace, tspaceOSD, gspace             glumby.CRect
	ignoreKeyEvents                       bool
	reboot                                bool
	lastbg                                uint64
	lastlm                                uint64
	lastpx                                uint64
	lastInt                               int = 0
	SelectedCamera                        int
	SelectedCameraIndex                   int
	SelectedIndex                         int
	SelectedAudioIndex                    int
	GFXLAYERVAL, HUDLAYERVAL              [8]uint64 // last known active states
	leds                                  []*video.LED
	visiblestate                          [memory.OCTALYZER_NUM_INTERPRETERS][memory.OCTALYZER_MAX_GFX_LAYERS]bool
	visiblestatetext                      [memory.OCTALYZER_NUM_INTERPRETERS][memory.OCTALYZER_MAX_HUD_LAYERS]bool
	ScreenLogging                         bool
	overlay                               *glumby.Texture
	overlayMesh                           *glumby.Mesh
	output                                driver.Output
	diskdrop                              *glumby.Texture
	diskdrop1                             *glumby.Texture
	diskdrop2                             *glumby.Texture
	hdv1                                  *glumby.Texture
	hdv2                                  *glumby.Texture
	pak0                                  *glumby.Texture
	pak1                                  *glumby.Texture
	tape0                                 *glumby.Texture
	tape1                                 *glumby.Texture
	lastmx, lastmy                        int
	lastdpx, lastdpy                      int
	program                               *glumby.Program
	waspect, waspectOSD                   float64
	fov, fovOSD                           float64
	mx, my                                float64
	lastSI                                float32
	lxpc, lypc                            float64
)

var targetHost = flag.String("backend", "localhost", "Host to connect to") // blah
var dataHost = flag.String("server", "microm8.paleotronic.com", "Server to connect to")
var versionDisplay = flag.Bool("version", false, "Show version and exit.")
var noUpdate = flag.Bool("no-update", false, "Dont check for updates.")
var trace6502 = flag.Bool("trace-cpu", false, "Trace 6502/Z80 access")
var breakIll6502 = flag.Bool("stop-ill-65c02", false, "Stop on 65c02 illegal opcode")
var trace6502mem = flag.Bool("trace-mem", false, "Trace memory writes via CPU ST?")
var prof = flag.Bool("profile-cpu", false, "Profile 6502 performance")
var heatmap = flag.Bool("heatmap", false, "Trace ROM access")
var share [memory.OCTALYZER_NUM_INTERPRETERS]*bool
var shareport [memory.OCTALYZER_NUM_INTERPRETERS]*int
var spkmode [memory.OCTALYZER_NUM_INTERPRETERS]uint64
var testing = flag.Bool("testing", false, "Use testing channel instead of stable")
var measureRemote = flag.Bool("measure-remote", false, "Trace remote network incoming")
var arch = flag.Bool("arch", false, "Show architecture and exit")
var bootdisk = flag.String("drive1", "", "Apple // boot volume")
var auxdisk = flag.String("drive2", "", "Apple // second volume")
var ramtest = flag.Int("goto", -1, "Execute from address")
var bankenable = flag.String("bankenable", "", "Enable selected banks before goto")
var debug = flag.Bool("debug", false, "Debug mode - invoke debugger")
var loqaudio = flag.Bool("loqaudio", false, "Use lower quality audio")
var fullscreen = flag.Bool("fullscreen", false, "Use fullscreen mode")
var nolog = flag.Bool("nolog", false, "Turn logging off.")
var fps = flag.Int("fps", 30, "Target frames per second")
var local = flag.Bool("local", false, "Boot to catalog")
var offline = flag.Bool("offline", false, "Boot to basic")
var unifiedTest = flag.Bool("demo-mode", false, "Cycle accurate video using cyanide core")
var parallelPrinter = flag.String("parallel-printer", "", "Parallel line printer device")

// var verbose = flag.Bool("verbose", false, "Print debug messages")
var localhelp = flag.Bool("localhelp", false, "Use local help path ")
var audiobuffer = flag.Int("abuffer", settings.BPSBufferSize, "Audio buffer size (32 bit segments)")
var monoGreen = flag.Bool("monogreen", false, "Default video mode is green mono")
var monoAmber = flag.Bool("monoamber", false, "Default video mode is amber mono")
var bw = flag.Bool("bw", false, "Default video mode is black and white")
var verbose = flag.Bool("verbose", false, "Allow logging to console")

var pprofile = flag.Bool("pprof", false, "Use PPROF profiling")
var aspectRatios = []float64{1.0, 1.33, 1.46, 1.62, 1.78}
var aspectRatioIndex [memory.OCTALYZER_NUM_INTERPRETERS]int
var drive1wp = flag.Bool("drive1wp", false, "Drive 1 in write protect mode")
var drive2wp = flag.Bool("drive2wp", false, "Drive 2 in write protect mode")
var cacheExtract = flag.String("cache-extract", "", "Extract file from cache object file")
var pcmOut = flag.String("pcm-out", "", "Record PCM to file")
var instVM = flag.Bool("inst-vms", false, "Allow instrumenting VMs over http")
var instPort = flag.String("inst-port", ":9502", "Instrumentation http port")
var launch = flag.String("launch", "", "File to launch with microM8")
var launchQuitCPUExit = flag.Bool("launch-cpu-quit", false, "Quit app on CPU exit")
var convertDSK2WOZ = flag.Bool("dsk-to-woz", false, "Convert DSKs to WOZ format")
var liveRecord = flag.Bool("live-record", false, "Enable live recording")
var controlLaunch = flag.String("control", "", "Specify program to launch in slot 2 (control)")
var profiles = flag.Bool("list-profiles", false, "List available machine profiles")
var setProfile = flag.String("profile", "", "Set machine profile")
var jMaxLeft = flag.Float64("jmaxleft", 80, "Maximum calibration left for joystick")
var jMaxRight = flag.Float64("jmaxright", 96, "Maximum calibration right for joystick")
var jMaxUp = flag.Float64("jmaxup", 80, "Maximum calibration up for joystick")
var jMaxDown = flag.Float64("jmaxdown", 48, "Maximum calibration down for joystick")
var nomenu = flag.Bool("no-menu", false, "Disable menu")
var contrast = flag.Bool("contrast", false, "Use high contrast mode for menu / catalog etc")
var controlport = flag.Bool("control-port", false, "Enable control thru ports 38911/38912")
var noborder = flag.Bool("disable-border", false, "Disable OS window border")

var recorderOn = flag.Bool("record", false, "Start recording (apple 2 only)")
var recorderFullCPU = flag.Bool("record-full-cpu", false, "Record all CPU states (-record)")

var port = flag.String("port", "6581", "Default server port")

var tracker = flag.Bool("tracker", false, "Boot to tracker")
var trace6522 = flag.Bool("trace-via", false, "Trace VIA 6522 interactions")
var modemInit = flag.String("modem-init", "", "Set modem init string")
var sscUseHardwarePort = flag.String("ssc-port", "", "Map super serial card to real serial port on host")
var sscListPorts = flag.Bool("ssc-list-ports", false, "List hardware serial ports on host for -ssc-port")
var sscEmulatedImageWriter = flag.Bool("ssc-imagewriter-emu", false, "Emulate an imagewriter attached to SSC")
var sscEmulatedESCP = flag.Bool("ssc-epson-emu", false, "Emulate an epson 9-pin printer attached to SSC")
var mcpMode = flag.Bool("mcp", false, "Run as MCP server")
var mcpTransport = flag.String("mcp-mode", "stdio", "MCP transport mode: stdio, sse, or streaming")
var mcpPort = flag.Int("mcp-port", 1983, "Port for MCP HTTP server (SSE or streaming)")

// const hamburger = `\uE104\uE11B\u0469\u0469\u0469\uE10F\uE110
// \uE10D\uE118\u0479\u0479\u0479\uE10F\uE110
// \uE101\uE11B\u0479\u0479\u0479\uE10F\uE110`

func init() {

	for i := 0; i < memory.OCTALYZER_NUM_INTERPRETERS; i++ {

		share[i] = flag.Bool(fmt.Sprintf("share-%d", i), false, "Share slot enabled")
		shareport[i] = flag.Int(fmt.Sprintf("sport-%d", i), 8580+i, "Share slot port")
		aspectRatioIndex[i] = 2

	}

	runtime.LockOSThread()
}

func SetSlotAspect(index int, aspect float64) {
	pcam[index].SetAspect(aspect)
	osd[index].SetAspect(1.77778)
	for i := 0; i < 8; i++ {
		fxcam[index][i].SetAspect(aspect)
	}
}

func initBackend(r *memory.MemoryMap) {
	go backend.Run(r, nil)
}

func round(f float64) float64 {
	return math.Floor(f + .5)
}

func floor(f float64) float64 {
	return math.Floor(f)
}

func RoundTo5(v int64) int64 {

	z := float64(v)

	z = round(z/5) * 5

	return int64(z)
}

func getSelectedCameraIndex(slotid int) int {
	ent := backend.ProducerMain.GetInterpreter(slotid)
	for ent.GetChild() != nil {
		ent = ent.GetChild()
	}

	mode := GetVideoMode(SelectedIndex)

	if mode != "" {
		spec := apple2helpers.GETGFX(ent, mode)

		if spec != nil {
			format := spec.GetFormat()
			if !settings.PureBoot(slotid) && format == types.LF_VECTOR || format == types.LF_CUBE_PACKED {
				// if fxcam[slotid][1].GetFar() != 30000 {
				// 	fxcam[slotid][1].SetNear(800)
				// 	fxcam[slotid][1].SetFar(30000)
				//log2.Printf("Using camera 1, in mode %s\n", mode)
				// }
				return 1
			}
		}
	}
	return 0
}

func initShaders() {

	vshader, err := glumby.NewShader(vertexSource, gl.VERTEX_SHADER)
	if err != nil {
		panic(err)
	}

	fshader, err := glumby.NewShader(fragmentSource, gl.FRAGMENT_SHADER)
	if err != nil {
		panic(err)
	}

	if fshader == nil {
		panic("Fragment shader did not compile...?")
	}

	program = glumby.NewSimpleProgram(vshader, fshader)

}

func logMonitor() {
	// lm := RAM.ReadGlobal(RAM.MEMBASE(0) + memory.OCTALYZER_INTERPRETER_LOGMODE)
	// if lm != lastlm {

	// 	switch lm {
	// 	case 1:
	// 		ScreenLogging = true
	// 	default:
	// 		ScreenLogging = false
	// 	}

	// 	lastlm = lm
	// }
}

func bgMonitor() {
	bg := RAM.ReadGlobal(SelectedIndex, RAM.MEMBASE(SelectedIndex)+memory.OCTALYZER_BGCOLOR)
	if bg != lastbg {
		fmt.Println("BGColor change")
		BGColor.R = uint8(bg & 0xff)
		BGColor.G = uint8((bg >> 8) & 0xff)
		BGColor.B = uint8((bg >> 16) & 0xff)
		BGColor.A = uint8((bg >> 24) & 0xff)
		lastbg = bg
	}

	px := RAM.ReadGlobal(0, RAM.MEMBASE(0)+memory.OCTALYZER_HGR_SIZE)
	if px != lastpx {
		video.HGRPixelSize = float32(px) / 10
		lastpx = px
	}

	speakerMonitor()
}

func speakerMonitor() {
	for i := 0; i < memory.OCTALYZER_NUM_INTERPRETERS; i++ {
		mode := RAM.IntGetSpeakerMode(i)
		if mode != spkmode[i] {
			spkmode[i] = mode
			clientperipherals.SPEAKER.SetBlockingMode(i, (mode != 0))
		}
	}
}

func startSharing() {

	for i := 0; i < memory.OCTALYZER_NUM_INTERPRETERS; i++ {
		if *share[i] {
			RAM.Share(i, backend.VPP, *shareport[i])
		}
	}

}

func clickMonitor() {
	var freq, ms uint64
	var cmd uint64
	for {

		//var found bool = false

		for index := 0; index < memory.OCTALYZER_NUM_INTERPRETERS; index++ {

			freq = RAM.ReadGlobal(index, RAM.MEMBASE(index)+memory.OCTALYZER_SPEAKER_FREQ)
			ms = RAM.ReadGlobal(index, RAM.MEMBASE(index)+memory.OCTALYZER_SPEAKER_MS)
			if freq > 0 {
				if freq == 999999 {
					//clientperipherals.SPEAKER.ToneInit()
					//clientperipherals.SPEAKER.BeepInit()
					RAM.WriteGlobal(index, RAM.MEMBASE(index)+memory.OCTALYZER_SPEAKER_FREQ, 0)
				} else {

					clientperipherals.SPEAKER.MakeTone(int(freq), int(ms))
					RAM.WriteGlobal(index, RAM.MEMBASE(index)+memory.OCTALYZER_SPEAKER_FREQ, 0)
					//fmt.Printf("Speaker tone in slot %d\n", index)
					//found = true
				}
				continue
			}

			// look for command
			cmd = RAM.ReadGlobal(index, RAM.MEMBASE(index)+memory.OCTALYZER_MUSIC_COMMAND)
			if types.RestalgiaCommand(cmd) != types.RS_None {
				l := int(2 + RAM.ReadGlobal(index, RAM.MEMBASE(index)+memory.OCTALYZER_MUSIC_BUFFER_COUNT))
				buffer := RAM.Data[index][RAM.MEMBASE(index)+memory.OCTALYZER_MUSIC_COMMAND : RAM.MEMBASE(index)+memory.OCTALYZER_MUSIC_COMMAND+l]

				clientperipherals.SPEAKER.Command(buffer)

				// finish up
				RAM.WriteGlobal(index, RAM.MEMBASE(index)+memory.OCTALYZER_MUSIC_COMMAND, uint64(types.RS_None))

				//found = true
			}

		}

		time.Sleep(10 * time.Millisecond)
	}
}

func RestalgiaCommandCallback(index int, s string) {

	clientperipherals.SPEAKER.SendCommands(s)

}

func DecodePackedAudio(bitcount int, buffer []uint64, ampscale float32) []float32 {

	//fmt.Printf("Decoding %d bits\n", bitcount)

	var fdata []float32

	fdata = make([]float32, 1)
	var bitnum int = 31
	var bindex int = 0
	var findex int
	var bitsprocessed int

	for bitsprocessed < bitcount && bindex < len(buffer) {
		bitval := uint64(1 << uint64(bitnum))
		if buffer[bindex]&bitval != 0 {
			// 1: bump amp
			fdata[findex] += ampscale
		} else {
			// 0: move to next
			fdata = append(fdata, 0)
			findex++
		}
		bitsprocessed++
		bitnum--
		if bitnum < 0 {
			// next 'byte'
			bindex++
			bitnum = 31
		}
	}
	return fdata[0:findex]
}

func RestOpCallback(index int, voice int, opcode int, value uint64) uint64 {
	fmt.Printf("Client recv callback: i=%d, v=%d, oc=%d, v=%d\n", index, voice, opcode, value)
	clientperipherals.SPEAKER.Mixer.Slots[index].ExecuteOpcode(voice, opcode, value)
	return 0
}

// WaveMonkey is a callback...
func WaveMonkey(index int, channel int, indata []uint64, rate int, bytepacked bool) {

	//log2.Printf("block from slot %d, rate = %d, ch = %d, samples = %d", index, rate, channel, len(indata))

	//dbg.PrintStack()

	// for originslot, targetslot := range settings.SpeakerRedirects {
	// 	if originslot != index && targetslot != nil && targetslot.VM == index {
	// 		return
	// 	}
	// }

	if r := settings.SpeakerRedirects[index]; r != nil {
		if channel == 0 {
			index = r.VM
			channel = r.Channel
		}
	}

	if settings.SpeakerVolume[index] == 0 {
		return
	}

	if SelectedAudioIndex != index {
		//fmt.Printf("Ignoring audio channel %d\n", index)
		return // ignore this channel
	}

	var data []float32
	if bytepacked {

		bitcount := int(indata[0])
		buffer := indata[1:]
		data = DecodePackedAudio(bitcount, buffer, 0.25)

	} else {
		data = memory.UintSlice2Float(indata)
	}

	if settings.AudioPacketReverse[index] {
		tmp := make([]float32, len(data))
		l := len(data) - 1
		for i, v := range data {
			tmp[l-i] = v
		}
		data = tmp
	}

	// are we recording cassette
	// if channel == 1 && len(data) > 0 && settings.RecordC020[index] {
	// 	settings.RecordC020Buffer[index] = append(settings.RecordC020Buffer[index], data...)
	// }

	// end recording

	clientperipherals.SPEAKER.PassWaveBuffer(index, channel, data, false, rate)

}

func DancingMonkey(index int, indata []uint64, rate int, channels int, bytepacked bool) {

	if bytepacked && SelectedAudioIndex != index {
		fmt.Printf("Ignoring audio channel %d\n", index)
		return // ignore this channel
	}

	var data []float32
	if bytepacked {

		bitcount := int(indata[0])
		buffer := indata[1:]
		data = DecodePackedAudio(bitcount, buffer, 0.25)

	} else {
		data = memory.UintSlice2Float(indata)
	}

	if settings.AudioPacketReverse[index] {
		tmp := make([]float32, len(data))
		l := len(data) - 1
		for i, v := range data {
			tmp[l-i] = v
		}
		data = tmp
	}

	clientperipherals.SPEAKER.PassMusicBuffer(index, data, false, rate, channels)

}

func initLayerPointers() {

	HUDSpecs = make(map[int][]*types.LayerSpecMapped)
	GFXSpecs = make(map[int][]*types.LayerSpecMapped)
	HUDLayers = make(map[int][]*video.TextLayer)
	GraphicsLayers = make(map[int][]*video.GraphicsLayer)

	for index := 0; index < memory.OCTALYZER_NUM_INTERPRETERS; index++ {

		HUDSpecs[index] = make([]*types.LayerSpecMapped, memory.OCTALYZER_MAX_HUD_LAYERS)
		HUDLayers[index] = make([]*video.TextLayer, memory.OCTALYZER_MAX_HUD_LAYERS)

		for i := 0; i < memory.OCTALYZER_MAX_HUD_LAYERS; i++ {
			haddr := memory.OCTALYZER_HUD_BASE + i*memory.OCTALYZER_LAYERSPEC_SIZE
			HUDSpecs[index][i] = &types.LayerSpecMapped{Index: index, Base: haddr, Mm: RAM}
			HUDLayers[index][i] = nil
		}

		GFXSpecs[index] = make([]*types.LayerSpecMapped, memory.OCTALYZER_MAX_GFX_LAYERS)
		GraphicsLayers[index] = make([]*video.GraphicsLayer, memory.OCTALYZER_MAX_GFX_LAYERS)

		for i := 0; i < memory.OCTALYZER_MAX_GFX_LAYERS; i++ {
			haddr := memory.OCTALYZER_GFX_BASE + i*memory.OCTALYZER_LAYERSPEC_SIZE
			GFXSpecs[index][i] = &types.LayerSpecMapped{Index: index, Base: haddr, Mm: RAM}
			GraphicsLayers[index][i] = nil
		}

	}

}

func ilp(index int) {
	HUDSpecs[index] = make([]*types.LayerSpecMapped, memory.OCTALYZER_MAX_HUD_LAYERS)
	HUDLayers[index] = make([]*video.TextLayer, memory.OCTALYZER_MAX_HUD_LAYERS)

	for i := 0; i < memory.OCTALYZER_MAX_HUD_LAYERS; i++ {
		haddr := memory.OCTALYZER_HUD_BASE + i*memory.OCTALYZER_LAYERSPEC_SIZE
		HUDSpecs[index][i] = &types.LayerSpecMapped{Index: index, Base: haddr, Mm: RAM}
		HUDLayers[index][i] = nil
	}

	GFXSpecs[index] = make([]*types.LayerSpecMapped, memory.OCTALYZER_MAX_GFX_LAYERS)
	GraphicsLayers[index] = make([]*video.GraphicsLayer, memory.OCTALYZER_MAX_GFX_LAYERS)

	for i := 0; i < memory.OCTALYZER_MAX_GFX_LAYERS; i++ {
		haddr := memory.OCTALYZER_GFX_BASE + i*memory.OCTALYZER_LAYERSPEC_SIZE
		GFXSpecs[index][i] = &types.LayerSpecMapped{Index: index, Base: haddr, Mm: RAM}
		GraphicsLayers[index][i] = nil
	}
}

func stopBackend() {

}

func CheckSHRState() {
	SHRActive = false

	var vs [memory.OCTALYZER_NUM_INTERPRETERS][memory.OCTALYZER_MAX_GFX_LAYERS]bool

	for index, layerset := range GraphicsLayers {

		if ramActiveState[index] == 0 {
			continue
		}

		for lindex, layer := range layerset {
			if layer != nil && layer.Spec.GetActive() {
				if layer.Spec.GetID() == "SHR1" && index == SelectedIndex {
					SHRActive = true
					if !visiblestate[index][lindex] {
						mode := layer.Spec.GetID()
						if strings.HasPrefix(mode, "HGR") {
							UpdateLighting(settings.LastRenderModeHGR[index])
						} else if strings.HasPrefix(mode, "DHR") {
							UpdateLighting(settings.LastRenderModeDHGR[index])
						}
					}
					vs[index][lindex] = true
					layer.Fetch()
				}
			}
		}
	}

	visiblestate = vs
}

func UpdateGraphicsLayers() {

	var vs [memory.OCTALYZER_NUM_INTERPRETERS][memory.OCTALYZER_MAX_GFX_LAYERS]bool

	for index, layerset := range GraphicsLayers {

		if ramActiveState[index] == 0 {
			continue
		}

		for lindex, layer := range layerset {
			if layer != nil && layer.Spec.GetActive() {

				if !visiblestate[index][lindex] {
					mode := layer.Spec.GetID()
					if strings.HasPrefix(mode, "HGR") {
						UpdateLighting(settings.LastRenderModeHGR[index])
					} else if strings.HasPrefix(mode, "DHR") {
						UpdateLighting(settings.LastRenderModeDHGR[index])
					}
					// } else {
					// 	UpdateLighting(settings.VM_VOXELS)
					// }
				}

				vs[index][lindex] = true
				layer.Fetch()
			}
		}
	}

	visiblestate = vs

}

func UpdateTextLayers() {

	var vs [memory.OCTALYZER_NUM_INTERPRETERS][memory.OCTALYZER_MAX_HUD_LAYERS]bool

	for index, layerset := range HUDLayers {

		if ramActiveState[index] == 0 {
			continue
		}

		for lindex, layer := range layerset {
			if layer != nil && layer.Spec.GetActive() {
				vs[index][lindex] = true
				if settings.ForceTextVideoRefresh {
					layer.NeedRefresh = true
				}
				layer.Fetch()

			}
		}
	}

	visiblestatetext = vs
	settings.ForceTextVideoRefresh = false
}

var SHRActive bool

func RenderGraphicsLayers() {
	//gl.Disable(gl.ALPHA_TEST)

	SelectedCameraIndex = getSelectedCameraIndex(SelectedIndex)

	//log2.Printf("*** SelectedCameraIndex = %d", SelectedCameraIndex)

	for index, layerset := range GraphicsLayers {

		if ramActiveState[index] == 0 {
			//fmt.Printf("%d=off ", index)
			continue
		}

		gl.PushMatrix()
		lp := backend.ProducerMain.MasterLayerPos[index]
		//cindex := RAM.ReadGlobal(RAM.MEMBASE(index) + memory.OCTALYZER_MAPPED_CAM_VIEW)
		fxcam[index][SelectedCameraIndex].ApplyWindow(&playfield, lp, tspace, waspect, fov, types.GFXMULT)
		//log2.Printf(">>> Near plane is %f, far = %f\n", fxcam[index][SelectedCameraIndex].GetNear(), fxcam[index][SelectedCameraIndex].GetFar())
		//layerlist := ""
		for lindex, layer := range layerset {
			if layer != nil && visiblestate[index][lindex] {

				if layer.Spec.GetRefresh() {
					v := settings.LastTintMode[index]
					switch v {
					case settings.VPT_NONE:
						layer.Tint = nil
					case settings.VPT_AMBER:
						layer.Tint = types.NewVideoColor(255, 115, 0, 255)
					case settings.VPT_GREEN:
						layer.Tint = types.NewVideoColor(103, 253, 146, 255)
					case settings.VPT_GREY:
						layer.Tint = types.NewVideoColor(255, 255, 255, 255)
					}
					layer.TintChanged = true
					layer.Spec.SetRefresh(false)
					layer.DepthChanged = true
				}

				layer.Update()
				layer.Render()
				//log2.Printf("Rendering GFX: %s", layer.Spec.GetID())
				//layerlist += layer.Spec.GetID() + " "
			}
		}
		// if layerlist != "" {
		// 	log.Println(layerlist)
		// }
		gl.PopMatrix()
	}
}

func SnapLayers() string {
	ww, hh := w.GetGLFWWindow().GetFramebufferSize()
	filepath := files.GetUserDirectory(files.BASEDIR + "/MyScreenshots")
	os.MkdirAll(filepath, 0755)
	t := time.Now()
	filename := fmt.Sprintf("%s/screenshot-%.4d-%s-%.2d-%.2d-%.2d-%.2d.png", filepath, t.Year(), t.Month().String(), t.Day(), t.Hour(), t.Minute(), t.Second())
	fmt.Println(filename)
	ScreenShotPNG(0, 0, ww, hh, filename)
	return filename
}

// func RenderUnified() {
// 	gl.Enable(gl.TEXTURE_2D)
// 	t.ScreenTex.Bind()
// 	// t.MBO.Begin(gl.TRIANGLES)
// 	glumby.MeshBuffer_Begin(gl.TRIANGLES)
// 	t.d.Mesh.DrawWithMeshBuffer(t.BitmapPosX+float32(sx), t.BitmapPosY+float32(sy), t.BitmapPosZ+float32(sz))
// 	glumby.MeshBuffer_End()
// 	// t.MBO.Draw(t.BitmapPosX, t.BitmapPosY, t.BitmapPosZ, t.d.Mesh)
// 	// t.MBO.Send(true)
// 	gl.Disable(gl.TEXTURE_2D)
// }

func RenderTextLayers() {

	UpdateTextLayers()

	gl.Enable(gl.ALPHA_TEST)
	for index, layerset := range HUDLayers {

		if ramActiveState[index] == 0 {
			continue
		}

		////fmt.Printf("Hud layer %d\n", i)
		gl.PushMatrix()
		lp := backend.ProducerMain.MasterLayerPos[index]

		if pcam[index] == nil {
			initCameraInSlot(index) // we need to init the camera
		}

		pcam[index].ApplyWindow(&playfield, lp, tspace, waspect, fov, 1)
		changed := backend.ProducerMain.ForceUpdate[index]

		for lindex, layer := range layerset {

			if layer != nil && visiblestatetext[index][lindex] {

				if layer.Spec.GetID() == "OOSD" || layer.Spec.GetID() == "MONI" {
					continue
				}

				if layer.Spec.GetRefresh() {
					v := settings.LastTintMode[index]
					switch v {
					case settings.VPT_NONE:
						layer.Tint = nil
					case settings.VPT_AMBER:
						layer.Tint = types.NewVideoColor(255, 115, 0, 255)
					case settings.VPT_GREEN:
						layer.Tint = types.NewVideoColor(103, 253, 146, 255)
					case settings.VPT_GREY:
						layer.Tint = types.NewVideoColor(255, 255, 255, 255)
					}
					layer.TintChanged = true
					layer.Spec.SetRefresh(false)
				}

				// if bounds changed then force refresh
				if !layer.Spec.GetBoundsRect().Equals(layer.LastBounds) {
					fmt.Println("TEXT bounds changed")
					layer.NeedRefresh = true
					layer.Buffer.Write(4089, layer.Buffer.Read(4089)|256)
				}

				layer.PosChanged = changed
				layer.Changed = true
				layer.Update()
				layer.Render()
				//fmt.Println(layer.Spec.GetID(), layer.Spec.GetActive())
			}
		}
		//fmt.Println()
		backend.ProducerMain.ForceUpdate[index] = false
		////fmt.Printf("Hud layer %d - done\n", i)
		gl.PopMatrix()
	}
	//~ gl.Enable(gl.DEPTH_TEST)
	gl.Disable(gl.ALPHA_TEST)
}

func RenderTextLayersOSD() {

	gl.Enable(gl.ALPHA_TEST)
	for index, layerset := range HUDLayers {

		if ramActiveState[index] == 0 {
			//if index == SelectedIndex {
			//	log2.Printf("Skipping render due to ram active state being off")
			//}
			continue
		}

		gl.PushMatrix()
		lp := types.LayerPosMod{
			XPercent: 0,
			YPercent: 0,
		}

		osd[index].SetAspect(1.788889)
		osd[index].ApplyWindow(&playfieldOSD, lp, tspaceOSD, waspectOSD, fovOSD, 1)
		changed := backend.ProducerMain.ForceUpdate[index]

		for lindex, layer := range layerset {
			if layer != nil {
				if layer.Spec.GetID() != "OOSD" && layer.Spec.GetID() != "MONI" {
					continue
				}
				if !visiblestatetext[index][lindex] {
					//log2.Printf("OSD visibility flag not set")
					continue
				}
				if !layer.Spec.GetBoundsRect().Equals(layer.LastBounds) {
					fmt.Println("TEXT bounds changed")
					layer.NeedRefresh = true
					layer.Buffer.Write(4089, layer.Buffer.Read(4089)|256)
				}
				layer.PosChanged = changed
				layer.Changed = true
				layer.Fetch()
				layer.Update()
				layer.Render()
				//log2.Printf("Rendering OSD in slot %d", index)
			}
		}
		backend.ProducerMain.ForceUpdate[index] = false
		gl.PopMatrix()
	}

	gl.Disable(gl.ALPHA_TEST)
}

func restalgiaInit() {
	clientperipherals.SPEAKER = &clientperipherals.SoundPod{}
	clientperipherals.SPEAKER.StartAudio(&SelectedAudioIndex)
	restalgia.SetCreateVoiceCallback(RestCreateVoice)
	restalgia.SetRemoveVoiceCallback(RestRemoveVoice)
}

func RestCreateVoice(index int, port int, label string, inst string) {
	clientperipherals.SPEAKER.Mixer.SetupVoice(index, port, label, inst)
}

func RestRemoveVoice(index int, port int, label string) {
	clientperipherals.SPEAKER.Mixer.DestroyVoice(index, port, label)
}

func audioInit() {

	go clickMonitor()
	go checkAudioPlay()

	for i := 0; i < memory.OCTALYZER_NUM_INTERPRETERS; i++ {
		RAM.SetRestCallback(i, RestalgiaCommandCallback)
		RAM.SetWaveCallback(i, WaveMonkey)
		RAM.SetMusicCallback(i, DancingMonkey)
		RAM.SetRestOpCallback(i, RestOpCallback)
	}

}

func checkFullscreen() {
	if w.GetFullscreen() != (!settings.Windowed) {
		w.SetFullscreen(!settings.Windowed)
	}
}

func memoryInit() {
	if RAM == nil {
		RAM = memory.NewMemoryMap()
	}
	for i := 0; i < memory.OCTALYZER_NUM_INTERPRETERS; i++ {
		//RAM.SetWaveCallback(i, WaveMonkey)
		RAM.SlotReset(i)
	}

	initBackend(RAM)

	RAM.InputToggle(0)

	startSharing()

}

func main() {
	pnc.Do(
		maininner,
		func(r interface{}) {
			// r is an exception...
			b := make([]byte, 8192)
			i := runtime.Stack(b, false)

			filename := fmt.Sprintf("%s-app-crash.log", time.Now().Format("2006-01-02-15-04-05"))

			bb := bytes.NewBuffer([]byte(fmt.Sprintf("\n%v\n\n", r)))
			bb.WriteString(fmt.Sprintf("Build: %s\nGit  : %s\nBuilt: %s\nLevel: application\n\n", update.GetBuildNumber(), update.GetBuildHash(), update.GetBuildDate()))
			bb.Write(b[0:i])

			files.WriteBytesViaProvider("/local/logs", filename, bb.Bytes())
		},
	)
}

func maininner() {

	chat.SetMenuHook(ui.TestMenu)
	editor.SetMenuHook(ui.TestMenu)
	microtracker.SetMenuHook(ui.TestMenu)

	// fmt.Printf("MeshBufferSize = %d bytes\n", unsafe.Sizeof(glumby.MeshBufferObject{}))
	// os.Exit(1)

	// f, _ := os.Create("trace.out")
	// trace.Start(f)
	// defer func() {
	// 	trace.Stop()
	// 	f.Close()
	// }()

	MESHZOOM = 223

	// defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()
	//runtime.GOMAXPROCS(16)
	//debug.SetGCPercent(-1)
	// if runtime.GOOS == "darwin" {
	// 	runtime.GOMAXPROCS(runtime.NumCPU() * 2)
	// }

	if !CapsPressed() {
		// Enable caps key
		toggleCaps()
	}

	FlipCase = false

	flag.Parse()

	// Handle MCP mode
	if *sscListPorts {
		fmt2.Println("Available hardware serial ports:")
		ports, err := common.EnumerateSerialPorts()
		if err == nil {
			for i, port := range ports {
				fmt2.Printf("  %d) %s\n", i+1, port)
			}
		}
		os.Exit(0)
	}

	settings.SSCHardwarePort = *sscUseHardwarePort
	if settings.SSCHardwarePort != "" {
		settings.SSCCardMode[0] = settings.SSCModeSerialRaw // force raw serial mode
	}
	if *sscEmulatedImageWriter {
		settings.SSCCardMode[0] = settings.SSCModeEmulatedImageWriter
		settings.SSCHardwarePort = ""
	}
	if *sscEmulatedESCP {
		settings.SSCCardMode[0] = settings.SSCModeEmulatedESCP
		settings.SSCHardwarePort = ""
	}
	settings.DefModemInitString = *modemInit

	settings.NoUpdates = *noUpdate

	settings.Verbose = *verbose

	if *instVM && settings.SystemType != "nox" {
		instrument.StartInstServer(*instPort)
	}

	settings.SetAutoLiveRecording(*liveRecord)

	settings.UnifiedRenderGlobal = *unifiedTest

	settings.ParallelLinePrinter = *parallelPrinter
	settings.ParallelPassThrough = *parallelPrinter != ""

	settings.ShowHamburger = !*nomenu
	settings.HighContrastUI = *contrast

	//debugger.Start()
	if *profiles {
		pr := ui.GetMachineList()
		settings.Verbose = true
		fmt.RPrintln("Machine types (use with -profile):")
		for _, p := range pr {
			fmt.RPrintf(" %-20s %s\r\n", strings.Replace(p.Filename, ".yaml", "", -1), p.Name)
		}
		settings.Verbose = *verbose
		os.Exit(0)
	}

	if *setProfile != "" {
		if !strings.HasSuffix(*setProfile, ".yaml") {
			*setProfile = *setProfile + ".yaml"
		}
		settings.SpecFile[0] = *setProfile
	}

	if *monoGreen {
		settings.DefaultRenderModeDHGR = settings.VM_MONO_FLAT
		settings.DefaultRenderModeHGR = settings.VM_MONO_FLAT
		settings.DefaultTintMode = settings.VPT_GREEN
	}

	if *monoAmber {
		settings.DefaultRenderModeDHGR = settings.VM_MONO_FLAT
		settings.DefaultRenderModeHGR = settings.VM_MONO_FLAT
		settings.DefaultTintMode = settings.VPT_AMBER
	}

	if *bw {
		settings.DefaultRenderModeDHGR = settings.VM_FLAT
		settings.DefaultRenderModeHGR = settings.VM_FLAT
		settings.DefaultTintMode = settings.VPT_GREY
	}

	settings.MaxStickLevel[settings.StickLEFT] = *jMaxLeft
	settings.MaxStickLevel[settings.StickRIGHT] = *jMaxRight
	settings.MaxStickLevel[settings.StickUP] = *jMaxUp
	settings.MaxStickLevel[settings.StickDOWN] = *jMaxDown

	if *cacheExtract != "" {
		e := files.ExtractFromCache(*cacheExtract)
		if e != nil {
			fmt.Println(e)
		} else {
			fmt.Println("Ok")
		}
		os.Exit(0)
	}

	settings.ServerPort = ":" + *port

	settings.BPSBufferSize = *audiobuffer

	if *pprofile && settings.SystemType != "nox" {
		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}

	if *testing {
		update.CHANNEL = "testing"
		update.SetupChannels()
	}

	settings.LocalBoot = *local
	settings.Offline = *offline
	settings.TrackerMode = *tracker
	settings.RecordMix = *pcmOut

	settings.PureBootVolumeWP[0] = *drive1wp
	settings.PureBootVolumeWP2[0] = *drive2wp

	if *debug {
		settings.DebuggerAttachSlot = 1
	}

	if *loqaudio {
		settings.UseHQAudio = false
	}

	mos6502.STOP65C02 = *breakIll6502
	mos6502.TRACE = *trace6502
	mos6502.TRACEMEM = *trace6502mem
	mos6502.PROFILE = *prof
	mos6502.HEATMAP = *heatmap
	z80.TRACE = *trace6502

	settings.Debug6522 = *trace6522

	if *versionDisplay {
		fmt.Println(update.GetHumanVersion())
		os.Exit(0)
	}

	if *localhelp {
		settings.HelpBase = "/local/help"
	}

	if *dataHost != "" && settings.SystemType != "nox" {
		s8webclient.CONN = s8webclient.New(*dataHost, ":6581")
	}

	if *measureRemote {
		settings.TRACENET = true
	}

	if *arch {
		archstring := runtime.GOOS + "/" + runtime.GOARCH
		fmt.Printf("Platform: %s\n", archstring)
		os.Exit(0)
	}

	if *bootdisk != "" {
		//settings.SplashDisk = "local:" + *bootdisk
		//settings.PureBoot = true
		settings.PureBootVolume[0] = "local:" + *bootdisk
	}

	if *auxdisk != "" {
		settings.PureBootVolume2[0] = "local:" + *auxdisk
		//settings.PureBoot = true
	}

	if *ramtest != -1 {
		//settings.PureBoot = true
		settings.PureBootVolume[0] = fmt.Sprintf("ram:0x%.4x", *ramtest)
		fmt.Printf("Boot address =  0x%.4x\n", *ramtest)
	}

	if *bankenable != "" {
		settings.PureBootBanks = strings.Split(*bankenable, ",")
	}

	settings.DONTLOGDEFAULT = *nolog

	RAM = memory.NewMemoryMap()

	go func() {

		fmt.Println("waiting for window")

		// Window is created
		for w == nil {
			time.Sleep(100 * time.Millisecond)
		}

		if *controlport && settings.SystemType != "nox" {
			settings.SuppressWindowedMenu = true
			go StartControlServer(":38911", w)
			//go StartNotifyServer(":38912")
			// go StartMCPServer(1979) // Start MCP server on port 1979
		}

		if *mcpMode {
			// Run as MCP server
			go func() {
				if err := StartMCPServerSDK(); err != nil {
					log.Printf("MCP server error: %v", err)
					os.Exit(1)
				}
			}()
			// os.Exit(0)
		}

		// Framecount is 5
		for w.FrameCount < 20 {
			time.Sleep(100 * time.Millisecond)
		}

		fmt.Println("got window")

		settings.DONTLOG = true

		tries := 0
		var e error
		if settings.SystemType != "nox" {
			e = networkInit(*dataHost)
			for e != nil && tries < 3 {
				tries++
				//fmt.Println("Trying again")
				e = networkInit(*dataHost)
			}

			if e != nil {
				fmt.Println("Going to fallback mode as we don't have network...")
				settings.EBOOT = true
				files.System = false
				settings.DONTLOG = true
			}
		}

		restalgiaInit()

		memoryInit()

		fmt.Println("meminit")

		for RAM == nil {
			time.Sleep(50 * time.Millisecond)
		}

		time.Sleep(50 * time.Millisecond)
		audioInit()

		fmt.Println("audioinit")

		fmt.Println("Letting gfx core take over...")

		for backend.ProducerMain == nil || backend.ProducerMain.GetInterpreter(0) == nil {
			time.Sleep(time.Millisecond)
		}
		//backend.ProducerMain.RebootVM(0) // force a reboot
		//RAM.IntSetBackdrop(0, "/local/galaxy.png", 7, 0.8)

		settings.PreserveDSK = !*convertDSK2WOZ

		settings.DiskRecordStart = *recorderOn
		settings.FileFullCPURecord = *recorderFullCPU

		if *launch != "" {
			e := backend.ProducerMain.GetInterpreter(0)
			servicebus.Unsubscribe(0, e)
			servicebus.Subscribe(
				0,
				servicebus.LaunchEmulator,
				e,
			)
			OnDropFiles(
				w,
				[]string{
					*launch,
				},
				true,
			)
			settings.LaunchQuitCPUExit = true
		}

		if *controlLaunch != "" {

			//time.AfterFunc(2*time.Second, func() {
			backend.ProducerMain.Activate(1)
			e := backend.ProducerMain.GetInterpreter(1)
			e.SetMicroControl(true)
			//
			servicebus.Unsubscribe(1, e)
			servicebus.Subscribe(
				e.GetMemIndex(),
				servicebus.LaunchEmulator,
				e,
			)
			servicebus.SendServiceBusMessage(
				1,
				servicebus.LaunchEmulator,
				&servicebus.LaunchEmulatorTarget{
					Filename:  *controlLaunch,
					Drive:     0,
					IsControl: true,
				},
			)

			settings.LaunchQuitCPUExit = true
			//})

		}

		for !settings.BootCheckDone {
			time.Sleep(1 * time.Millisecond)
		}
		debugger.Start()
		plus.StateLoadFunc = debugger.DebuggerInstance.LoadState

		for true {
			time.Sleep(1 * time.Second)
		}

	}()

	settings.Args = os.Args

	settings.Windowed = !*fullscreen
	settings.UseFullScreen = *fullscreen

	props = glumby.NewWindowProperties()
	props.Title = "microM8"
	if settings.SystemType == "nox" {
		props.Title = "Nox Archaist"
	}
	props.Width = 1024
	props.Height = 576
	props.KeyMapper = SetupKeymapper()
	props.Resizable = true
	props.Fullscreen = *fullscreen
	props.VersionMajor = 2
	props.VersionMinor = 0
	props.FPS = *fps
	props.CursorHidden = false
	props.NoBorder = *noborder
	w = glumby.NewWindow(props)
	loadSplash()
	loadUnified()
	loadDiskOverlays()
	checkBG()
	checkOverlay()

	ts := glumby.GetMaxTextureSize()
	fmt.Printf("Max texture size = %d\n", ts)

	if w != nil {
		fmt.Println("i have a window")
	}

	w.OnCreate = OnCreateWindow
	w.OnDestroy = OnDestroyWindow
	w.OnEvent = OnEventWindow
	w.OnKey = OnKeyEvent
	w.OnRawKey = OnRawKeyEvent
	w.OnChar = OnCharEvent
	w.OnMouseMove = OnMouseMoveEvent
	w.OnMouseButton = OnMouseButtonEvent
	w.OnResize = OnResizeEvent
	w.OnScroll = OnMouseScroll
	w.OnDrop = OnDropFiles

	w.OnRender = OnRenderWindow
	w.Run()
	fmt.Println("GFX loop has exited Run()")

	// flush disks
	for i := 0; i < settings.NUMSLOTS; i++ {
		servicebus.SendServiceBusMessage(
			i,
			servicebus.DiskIIFlush,
			0,
		)
		servicebus.SendServiceBusMessage(
			i,
			servicebus.DiskIIFlush,
			1,
		)
		servicebus.SendServiceBusMessage(
			i,
			servicebus.SmartPortFlush,
			0,
		)
	}

	clientperipherals.SPEAKER.StopAudio()

}

func networkInit(host string) error {
	if settings.SystemType == "nox" {
		return nil
	}
	return s8webclient.NetworkInit(host)
}

func CalcPlayfieldOld(w *glumby.Window) {
	width, height := w.GetGLFWWindow().GetFramebufferSize()
	aspect := float64(width) / float64(height)

	nw := width
	nh := height

	if aspect < types.PASPECT {
		nh = int(float64(nw) * (1 / types.PASPECT))
	} else {
		nw = int(types.PASPECT * float64(nh))
	}

	playfield.X0, playfield.Y0, playfield.X1, playfield.Y1 = float64(width-nw)/2, float64(height-nh)/2, float64(width-nw)/2+float64(nw), float64(height-nh)/2+float64(nh)
}

func calcPlayfield(w *glumby.Window) {
	width, height := w.GetGLFWWindow().GetSize()
	aspect := float64(width) / float64(height)
	waspect = aspect

	nw := width
	nh := height

	if aspect < types.PASPECT {
		// Vertical should be based on width
		nh = int(float64(nw) * (1 / types.PASPECT))
	} else {
		// Horizontal should be based on height
		nw = int(types.PASPECT * float64(nh))
	}

	playfield.X0, playfield.Y0, playfield.X1, playfield.Y1 = float64(width-nw)/2, float64(height-nh)/2, float64(width-nw)/2+float64(nw), float64(height-nh)/2+float64(nh)

	// ory := float64(height) / float64(nh)
	// orx := float64(width) / float64(nw)

	if aspect <= types.PASPECT {
		fov = VFOVFromHFOV(90, float64(width), float64(height))
	} else {
		fov = VFOVFromHFOV(90, float64(nw), float64(height))
	}

	if pcam[0] != nil {

		topleft, _ := pcam[0].ScreenPosToWorldPos(&glmath.Vector3{playfield.X0, playfield.Y0, 0.09}, waspect, fov, 0, 0, width, height)
		botright, _ := pcam[0].ScreenPosToWorldPos(&glmath.Vector3{playfield.X1, playfield.Y1, 0.09}, waspect, fov, 0, 0, width, height)

		//fmt.Printf("topleft = %v\n", topleft)

		tspace.L = float64(topleft.X())
		tspace.R = float64(botright.X())
		tspace.T = float64(topleft.Y())
		tspace.B = float64(botright.Y())

	}

	//fmt.Printf("F.L=%f, F.R=%f, F.T=%f, F.B=%f\n", tspace.L, tspace.R, tspace.T, tspace.B)

}

func calcPlayfieldOSD(w *glumby.Window) {
	width, height := w.GetGLFWWindow().GetSize()
	aspect := float64(width) / float64(height)
	waspectOSD = aspect

	nw := width
	nh := height

	if aspect < types.PASPECT {
		// Vertical should be based on width
		nh = int(float64(nw) * (1 / types.PASPECT))
	} else {
		// Horizontal should be based on height
		nw = int(types.PASPECT * float64(nh))
	}

	playfieldOSD.X0, playfieldOSD.Y0, playfieldOSD.X1, playfieldOSD.Y1 = float64(width-nw)/2, float64(height-nh)/2, float64(width-nw)/2+float64(nw), float64(height-nh)/2+float64(nh)

	// ory := float64(height) / float64(nh)
	// orx := float64(width) / float64(nw)

	if aspect <= types.PASPECT {
		fovOSD = VFOVFromHFOV(90, float64(width), float64(height))
	} else {
		fovOSD = VFOVFromHFOV(90, float64(nw), float64(height))
	}

	if pcam[0] != nil {

		topleft, _ := osd[0].ScreenPosToWorldPos(&glmath.Vector3{playfieldOSD.X0, playfieldOSD.Y0, 0.09}, waspectOSD, fovOSD, 0, 0, width, height)
		botright, _ := osd[0].ScreenPosToWorldPos(&glmath.Vector3{playfieldOSD.X1, playfieldOSD.Y1, 0.09}, waspectOSD, fovOSD, 0, 0, width, height)

		//fmt.Printf("topleft = %v\n", topleft)

		tspaceOSD.L = float64(topleft.X())
		tspaceOSD.R = float64(botright.X())
		tspaceOSD.T = float64(topleft.Y())
		tspaceOSD.B = float64(botright.Y())

	}

	//fmt.Printf("F.L=%f, F.R=%f, F.T=%f, F.B=%f\n", tspace.L, tspace.R, tspace.T, tspace.B)

}

func led0() bool {
	on := (RAM.IntGetLED0(SelectedIndex) == 1)
	return on
}

func led1() bool {
	on := (RAM.IntGetLED1(SelectedIndex) == 1)
	return on
}

func led0sp() bool {
	on := (RAM.IntGetLED0(SelectedIndex) == 2)
	return on
}

func led1sp() bool {
	on := (RAM.IntGetLED1(SelectedIndex) == 2)
	return on
}

func initLED() {

	fmt.Println("begin init led layer")

	leds = make([]*video.LED, 0)

	//if settings.PureBoot {

	leds = append(leds,
		video.NewLED(
			"images/ledalt.png",
			"",
			0, 7, 0,
			24, 24,
			true,
			led1,
		),
	)

	leds = append(leds,
		video.NewLED(
			"images/ledalt.png",
			"",
			32, 7, 0,
			24, 24,
			true,
			led0,
		),
	)

	leds = append(leds,
		video.NewLED(
			"images/ledgreen.png",
			"",
			0, 7, 0,
			24, 24,
			true,
			led1sp,
		),
	)

	leds = append(leds,
		video.NewLED(
			"images/ledgreen.png",
			"",
			32, 7, 0,
			24, 24,
			true,
			led0sp,
		),
	)

	leds = append(leds,
		video.NewLED(
			"images/rewindicon.png",
			"",
			96, 7, 0,
			48, 48,
			true,
			func() bool {
				e := backend.ProducerMain.GetInterpreter(SelectedIndex)
				return e != nil && e.IsPlayingVideo() && e.GetPlayer().IsBackwards()
			},
		),
	)

	leds = append(leds,
		video.NewLED(
			"images/playicon.png",
			"",
			128, 7, 0,
			48, 48,
			true,
			func() bool {
				e := backend.ProducerMain.GetInterpreter(SelectedIndex)
				return e != nil && e.IsPlayingVideo() && !e.GetPlayer().IsBackwards()
			},
		),
	)

	leds = append(leds,
		video.NewLED(
			"images/pauseicon.png",
			"",
			160, 7, 0,
			48, 48,
			true,
			func() bool {
				e := backend.ProducerMain.GetInterpreter(SelectedIndex)
				return e != nil && e.IsWaitingForWorld()
			},
		),
	)

	leds = append(leds,
		video.NewLED(
			"images/recordicon.png",
			"",
			64, 7, 0,
			48, 48,
			true,
			func() bool {
				e := backend.ProducerMain.GetInterpreter(SelectedIndex)
				return e != nil && e.IsRecordingDiscVideo()
			},
		),
	)

	//}

	fmt.Println("end init led layer, got", len(leds))

}

func drawLEDs(x, y float32) {

	//fmt.Println("DrawLEDs", len(leds))
	gl.Enable(gl.TEXTURE_2D)
	//gl.Disable(gl.ALPHA_TEST)
	for _, l := range leds {
		//fmt.Println("LED ---------------------------------")
		l.Draw(x, y)
	}
	//gl.Disable(gl.TEXTURE_2D)
	gl.Enable(gl.ALPHA_TEST)

}

const bgdist = 8

func applyMult(v float32, slice []float32) []float32 {
	out := make([]float32, len(slice))
	for i, sv := range slice {
		sv = sv * v
		if sv > 1 {
			sv = 1
		} else if sv < 0 {
			sv = 0
		}
		out[i] = sv
	}
	return out
}

func initCameraInSlot(i int) {
	osd[i] = glumby.NewPerspCameraWithConfig(
		RAM,
		i,
		memory.OCTALYZER_OSD_CAM,
		settings.CameraOSDDefaults,
	)

	pcam[i] = glumby.NewPerspCameraWithConfig(
		RAM,
		i,
		memory.OCTALYZER_MAPPED_CAM_BASE,
		settings.CameraTextDefaults,
	)

	for z := 0; z < memory.OCTALYZER_MAPPED_CAM_GFXCOUNT; z++ {

		fxcam[i][z] = glumby.NewPerspCameraWithConfig(
			RAM,
			i,
			memory.OCTALYZER_MAPPED_CAM_BASE+(z+1)*memory.OCTALYZER_MAPPED_CAM_SIZE,
			settings.GFXCameraDefaults[z],
		)
	}
}

func initCameras() {

	initCameraInSlot(0)

	// for i := 0; i < memory.OCTALYZER_NUM_INTERPRETERS; i++ {

	// 	osd[i] = glumby.NewPerspCameraWithConfig(
	// 		RAM,
	// 		i,
	// 		memory.OCTALYZER_OSD_CAM,
	// 		settings.CameraOSDDefaults,
	// 	)

	// 	pcam[i] = glumby.NewPerspCameraWithConfig(
	// 		RAM,
	// 		i,
	// 		memory.OCTALYZER_MAPPED_CAM_BASE,
	// 		settings.CameraTextDefaults,
	// 	)

	// 	for z := 0; z < memory.OCTALYZER_MAPPED_CAM_GFXCOUNT; z++ {

	// 		fxcam[i][z] = glumby.NewPerspCameraWithConfig(
	// 			RAM,
	// 			i,
	// 			memory.OCTALYZER_MAPPED_CAM_BASE+(z+1)*memory.OCTALYZER_MAPPED_CAM_SIZE,
	// 			settings.GFXCameraDefaults[z],
	// 		)
	// 	}

	// }

	calcPlayfield(w)
	calcPlayfieldOSD(w)
}

func OnVideoSync() {
	UpdateTextLayers()
	UpdateGraphicsLayers()
	w.Synced()
}

func loadDiskOverlays() {
	b, err := assets.Asset("images/dropdiskoverlay.png")
	if err != nil {
		panic("burn it to the ground")
	}
	bb := bytes.NewBuffer(b)

	diskdrop, err = glumby.NewTextureFromBytes(bb)
	if err != nil {
		panic("failed to parse image from bytes :(")
	}

	b, err = assets.Asset("images/dropdiskoverlay1.png")
	if err != nil {
		panic("burn it to the ground")
	}
	bb = bytes.NewBuffer(b)

	diskdrop1, err = glumby.NewTextureFromBytes(bb)
	if err != nil {
		panic("failed to parse image from bytes :(")
	}

	b, err = assets.Asset("images/dropdiskoverlay2.png")
	if err != nil {
		panic("burn it to the ground")
	}
	bb = bytes.NewBuffer(b)

	diskdrop2, err = glumby.NewTextureFromBytes(bb)
	if err != nil {
		panic("failed to parse image from bytes :(")
	}

	b, err = assets.Asset("images/hdv1.png")
	if err != nil {
		panic("burn it to the ground")
	}
	bb = bytes.NewBuffer(b)

	hdv1, err = glumby.NewTextureFromBytes(bb)
	if err != nil {
		panic("failed to parse image from bytes :(")
	}

	b, err = assets.Asset("images/hdv2.png")
	if err != nil {
		panic("burn it to the ground")
	}
	bb = bytes.NewBuffer(b)

	hdv2, err = glumby.NewTextureFromBytes(bb)
	if err != nil {
		panic("failed to parse image from bytes :(")
	}

	// pak
	b, err = assets.Asset("images/pak0.png")
	if err != nil {
		panic("burn it to the ground")
	}
	bb = bytes.NewBuffer(b)

	pak0, err = glumby.NewTextureFromBytes(bb)
	if err != nil {
		panic("failed to parse image from bytes :(")
	}

	b, err = assets.Asset("images/pak1.png")
	if err != nil {
		panic("burn it to the ground")
	}
	bb = bytes.NewBuffer(b)

	pak1, err = glumby.NewTextureFromBytes(bb)
	if err != nil {
		panic("failed to parse image from bytes :(")
	}

	// tape
	b, err = assets.Asset("images/tape1.png")
	if err != nil {
		panic("burn it to the ground")
	}
	bb = bytes.NewBuffer(b)

	tape0, err = glumby.NewTextureFromBytes(bb)
	if err != nil {
		panic("failed to parse image from bytes :(")
	}

	b, err = assets.Asset("images/tape2.png")
	if err != nil {
		panic("burn it to the ground")
	}
	bb = bytes.NewBuffer(b)

	tape1, err = glumby.NewTextureFromBytes(bb)
	if err != nil {
		panic("failed to parse image from bytes :(")
	}
}

func loadSplash() {
	splash = video.NewDecal(types.CWIDTH, types.CHEIGHT)

	fname := "images/octasplash.png"
	if settings.SystemType == "nox" {
		fname = "images/noxsplash.png"
	}

	b, err := assets.Asset(fname)
	if err != nil {
		panic("burn it to the ground")
	}
	bb := bytes.NewBuffer(b)

	splashTexture, err = glumby.NewTextureFromBytes(bb)
	if err != nil {
		panic("failed to parse image from bytes :(")
	}

	splash.Texture = splashTexture
}

func loadUnified() {
	unified = video.NewDecal(1.3333*types.CHEIGHT, types.CHEIGHT)

	// bb := bytes.NewBuffer(b)
	img := image.NewRGBA(image.Rect(0, 0, 560, 384))
	draw.Draw(img, img.Bounds(), image.NewUniform(color.RGBA{0, 0, 0, 255}), image.Point{}, draw.Src)
	unifiedTexture = glumby.NewTextureFromRGBA(img)
	// if err != nil {
	// 	panic("failed to parse image from bytes :(")
	// }

	unified.Texture = unifiedTexture
}

func blendable(r, g, b uint8) bool {
	return (r+g+b) > 0 && (r < 255 || g < 255 || b < 255)
}

func CopyFrame(src *image.RGBA) *image.RGBA {
	f := image.NewRGBA(image.Rect(0, 0, 560, 384))
	var r, g, b, a uint8
	var lr, lg, lb uint8
	var nr, ng, nb uint8
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

			lr, lg, lb = r, g, b
			nr, ng, nb = r, g, b

			if y > 0 {
				lr = uint8(float32(src.Pix[xBase+(y-1)*line+0]) * 1)
				lg = uint8(float32(src.Pix[xBase+(y-1)*line+1]) * 1)
				lb = uint8(float32(src.Pix[xBase+(y-1)*line+2]) * 1)
			}
			if y < 191 {
				nr = uint8(float32(src.Pix[xBase+(y+1)*line+0]) * 1)
				ng = uint8(float32(src.Pix[xBase+(y+1)*line+1]) * 1)
				nb = uint8(float32(src.Pix[xBase+(y+1)*line+2]) * 1)
			}

			if settings.UseVerticalBlend[SelectedIndex] {
				if blendable(r, g, b) && blendable(lr, lg, lb) && blendable(nr, ng, nb) {
					f.Pix[xBase+yBaseTgt+0] = uint8((float32(r) + float32(lr) + float32(nr)) / 3)
					f.Pix[xBase+yBaseTgt+1] = uint8((float32(g) + float32(lg) + float32(ng)) / 3)
					f.Pix[xBase+yBaseTgt+2] = uint8((float32(b) + float32(lb) + float32(nb)) / 3)
					f.Pix[xBase+yBaseTgt+3] = src.Pix[xBase+yBaseSrc+3]

					f.Pix[xBase+yBaseTgt2+0] = uint8((float32(r) + float32(lr) + float32(nr)) / 3 * settings.ScanLineIntensity)
					f.Pix[xBase+yBaseTgt2+1] = uint8((float32(g) + float32(lg) + float32(ng)) / 3 * settings.ScanLineIntensity)
					f.Pix[xBase+yBaseTgt2+2] = uint8((float32(b) + float32(lb) + float32(nb)) / 3 * settings.ScanLineIntensity)
					f.Pix[xBase+yBaseTgt2+3] = a
				} else if blendable(r, g, b) && blendable(lr, lg, lb) {
					f.Pix[xBase+yBaseTgt+0] = uint8((float32(r) + float32(lr)) / 2)
					f.Pix[xBase+yBaseTgt+1] = uint8((float32(g) + float32(lg)) / 2)
					f.Pix[xBase+yBaseTgt+2] = uint8((float32(b) + float32(lb)) / 2)
					f.Pix[xBase+yBaseTgt+3] = src.Pix[xBase+yBaseSrc+3]

					f.Pix[xBase+yBaseTgt2+0] = uint8((float32(r) + float32(lr)) / 2 * settings.ScanLineIntensity)
					f.Pix[xBase+yBaseTgt2+1] = uint8((float32(g) + float32(lg)) / 2 * settings.ScanLineIntensity)
					f.Pix[xBase+yBaseTgt2+2] = uint8((float32(b) + float32(lb)) / 2 * settings.ScanLineIntensity)
					f.Pix[xBase+yBaseTgt2+3] = a
				} else if blendable(r, g, b) && blendable(nr, ng, nb) {
					f.Pix[xBase+yBaseTgt+0] = uint8((float32(r) + float32(nr)) / 2)
					f.Pix[xBase+yBaseTgt+1] = uint8((float32(g) + float32(ng)) / 2)
					f.Pix[xBase+yBaseTgt+2] = uint8((float32(b) + float32(nb)) / 2)
					f.Pix[xBase+yBaseTgt+3] = src.Pix[xBase+yBaseSrc+3]

					f.Pix[xBase+yBaseTgt2+0] = uint8((float32(r) + float32(nr)) / 2 * settings.ScanLineIntensity)
					f.Pix[xBase+yBaseTgt2+1] = uint8((float32(g) + float32(ng)) / 2 * settings.ScanLineIntensity)
					f.Pix[xBase+yBaseTgt2+2] = uint8((float32(b) + float32(nb)) / 2 * settings.ScanLineIntensity)
					f.Pix[xBase+yBaseTgt2+3] = a
				} else {
					f.Pix[xBase+yBaseTgt+0] = uint8(float32(src.Pix[xBase+yBaseSrc+0]) * 1)
					f.Pix[xBase+yBaseTgt+1] = uint8(float32(src.Pix[xBase+yBaseSrc+1]) * 1)
					f.Pix[xBase+yBaseTgt+2] = uint8(float32(src.Pix[xBase+yBaseSrc+2]) * 1)
					f.Pix[xBase+yBaseTgt+3] = src.Pix[xBase+yBaseSrc+3]

					f.Pix[xBase+yBaseTgt2+0] = uint8(float32(r) * settings.ScanLineIntensity)
					f.Pix[xBase+yBaseTgt2+1] = uint8(float32(g) * settings.ScanLineIntensity)
					f.Pix[xBase+yBaseTgt2+2] = uint8(float32(b) * settings.ScanLineIntensity)
					f.Pix[xBase+yBaseTgt2+3] = a
				}
			} else {
				f.Pix[xBase+yBaseTgt+0] = uint8(float32(src.Pix[xBase+yBaseSrc+0]) * 1)
				f.Pix[xBase+yBaseTgt+1] = uint8(float32(src.Pix[xBase+yBaseSrc+1]) * 1)
				f.Pix[xBase+yBaseTgt+2] = uint8(float32(src.Pix[xBase+yBaseSrc+2]) * 1)
				f.Pix[xBase+yBaseTgt+3] = src.Pix[xBase+yBaseSrc+3]

				f.Pix[xBase+yBaseTgt2+0] = uint8(float32(r) * settings.ScanLineIntensity)
				f.Pix[xBase+yBaseTgt2+1] = uint8(float32(g) * settings.ScanLineIntensity)
				f.Pix[xBase+yBaseTgt2+2] = uint8(float32(b) * settings.ScanLineIntensity)
				f.Pix[xBase+yBaseTgt2+3] = a
			}
		}
	}
	return f
}

func updateUnified() {
	if !settings.UnifiedRenderChanged[SelectedIndex] || strings.Contains(settings.SpecFile[SelectedIndex], "spectrum") {
		return // skip updates
	}
	if settings.UnifiedRenderFrame[SelectedIndex] != nil {
		unifiedTexture.SetSourceSame(CopyFrame(settings.UnifiedRenderFrame[SelectedIndex]))
	}
}

func checkAudioPlay() {

	for {

		time.Sleep(250 * time.Millisecond)

		if clientperipherals.SPEAKER == nil || clientperipherals.SPEAKER.Mixer == nil {
			continue
		}

		enabled, path, loop := RAM.IntGetRestalgiaPath(SelectedIndex)

		if fmt.Sprintf("%s:%v", path, loop) == musicPath[SelectedIndex] {
			continue
		}

		//log2.Printf(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>> Got music request: %s -> %d", path, SelectedIndex)

		rm := clientperipherals.SPEAKER.Mixer.Slots[clientperipherals.SPEAKER.Mixer.SlotSelect]

		if !enabled || path == "" {
			musicPath[SelectedIndex] = fmt.Sprintf("%s:%v", path, loop)
			rm.StopPlaying()
			continue
		}

		musicPath[SelectedIndex] = fmt.Sprintf("%s:%v", path, loop)

		rm.StartPlayback(path)

	}
}

func checkBG() {

	// isNew := RAM.IntGetBackdropIsNew(SelectedIndex)
	// if !isNew {
	// 	return
	// }

	// see if the selected background is set
	bgX, bgY, bgZ = RAM.IntGetBackdropPos(SelectedIndex)
	enabled, path, camidx, opacity, zoom, zoomfactor, camtrack := RAM.IntGetBackdrop(SelectedIndex)
	if !enabled || path == "" {
		bgTexture = nil
		bgpath = ""
		bgAspect = 0
		bgZoom = 0
		bgCamTrack = false
		// RAM.IntClearBackdropIsNew(SelectedIndex) // clear flag
		return
	}

	if fmt.Sprintf("%s:%f", path, opacity) == bgpath {
		bgZoom = float64(zoom)
		bgZoomFactor = float64(zoomfactor)
		bgCamTrack = camtrack
		// RAM.IntClearBackdropIsNew(SelectedIndex) // clear flag
		return
	}

	parts := strings.Split(bgpath, ":")
	if parts[0] == path {
		bgOpacity = opacity
		bgCamTrack = camtrack
		// RAM.IntClearBackdropIsNew(SelectedIndex) // clear flag
		return
	}

	a := float32(fxcam[SelectedIndex][camidx].GetAspect())

	bgdecal = video.NewDecal(types.CWIDTH, types.CHEIGHT)
	bgdecal.Mesh = video.GetPlaneAsTrianglesInv(types.CHEIGHT*1.05*a, types.CHEIGHT*1.05)

	// Different file, let's load it up
	fmt.Printf("Loading backdrop from path: %s\n", path)
	fr, err := files.ReadBytesViaProvider(files.GetPath(path), files.GetFilename(path))
	if err != nil {
		bgTexture = nil
		bgpath = ""
		bgZoom = 0
		bgCamTrack = false
		RAM.IntSetBackdrop(SelectedIndex, "", 7, 1, 16, 0, false)
		// RAM.IntClearBackdropIsNew(SelectedIndex) // clear fla
		return
	}
	bb := bytes.NewBuffer(fr.Content)

	bgTexture, err = glumby.NewTextureFromBytes(bb)
	if err != nil {
		return
	}

	bgdecal.Texture = bgTexture
	bgcamidx = 7 //camidx
	bgpath = fmt.Sprintf("%s:%f", path, opacity)
	bgOpacity = opacity
	bgAspect = a
	bgZoom = float64(zoom)
	bgZoomFactor = float64(zoomfactor)
	bgCamTrack = camtrack

	// RAM.IntClearBackdropIsNew(SelectedIndex) // clear flag
}

func checkOverlay() {

	// see if the selected background is set
	enabled, path := RAM.IntGetOverlay(SelectedIndex)
	if !enabled || path == "" {
		ovTexture = nil
		ovpath = ""
		return
	}

	if path == ovpath {
		return
	}

	if ovdecal == nil || ovdecal.Mesh == nil {
		ovdecal = video.NewDecal(types.CWIDTH, types.CHEIGHT)
		ovdecal.Mesh = video.GetPlaneAsTrianglesInv(types.CHEIGHT*16/9, types.CHEIGHT)
	}

	// Different file, let's load it up
	//log2.Printf("load overlay: %s", path)
	fr, err := files.ReadBytesViaProvider(files.GetPath(path), files.GetFilename(path))
	if err != nil {
		ovTexture = nil
		ovpath = ""
		RAM.IntSetOverlay(SelectedIndex, "")
		return
	}
	bb := bytes.NewBuffer(fr.Content)

	ovTexture, err = glumby.NewTextureFromBytes(bb)
	if err != nil {
		return
	}

	ovdecal.Texture = ovTexture
	ovpath = path
}

func OnCreateWindow(w *glumby.Window) {

	calcPlayfield(w)

	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.LIGHTING)
	gl.ClearColor(0, 0, 0, 0)
	gl.ClearDepth(1)
	gl.DepthFunc(gl.LESS)

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	// Lights
	// Add a glumby light
	ambient := []float32{0.0, 0.0, 0.0, 1}
	diffuse := []float32{0.8, 0.8, 0.8, 1}
	lightPosition := []float32{types.CWIDTH / 2, types.CHEIGHT / 2, types.CDIST * types.GFXMULT * 1.5, 0}
	light = glumby.NewLightSource(glumby.Light0, ambient, diffuse, lightPosition)
	light.On()

	ambient2 := []float32{0.0, 0.0, 0.0, 1}
	diffuse2 := []float32{1, 1, 1, 1}
	light2 = glumby.NewLightSource(glumby.Light1, ambient2, diffuse2, lightPosition)
	light2.On()

	//gl.Viewport(int32(width-nw)/2, int32(height-nh)/2, int32(nw), int32(nh))
	fmt.Println("before init cameras")
	initCameras()
	fmt.Println("after init cameras")
	// screen

	// initShaders()
	// program.Use()

	// reinit layers
	fmt.Println("before init layers")
	initLayerPointers()
	fmt.Println("after init layers")

	// leds
	initLED()

	//	go ReceiveVDUEvents()
	glumby.EnumerateControllers()

	settings.SetTitleFunc(w.SetSubtitle)

	bus.SetCallback(OnVideoSync)

	//go SampleMouse(w)
	settings.SetMouseModeCallback(OnChangeMouseMode)
}

func OnChangeMouseMode(m settings.MouseMode) {
	if m == settings.MM_MOUSE_GEOS || m == settings.MM_MOUSE_DDRAW {
		w.SetCursorDisabled(true)
	} else {
		w.SetCursorDisabled(false)
	}
}

// Handle application teardown
func OnDestroyWindow(w *glumby.Window) {

	// flush slot 0 disks
	e := backend.ProducerMain.GetInterpreter(SelectedIndex)
	e.VM().Teardown()

	if s8webclient.CONN != nil && s8webclient.CONN.IsConnected() {
		s8webclient.CONN.Done()
	}

}

func dropAnimation(drive int) {
	old := uiTexture
	fmt.Printf("Drive %d\n", drive)
	var overlays []*glumby.Texture
	switch drive {
	case 1:
		overlays = []*glumby.Texture{diskdrop, diskdrop1, diskdrop, diskdrop1, diskdrop, diskdrop1, old}
	case 2:
		overlays = []*glumby.Texture{diskdrop, diskdrop2, diskdrop, diskdrop2, diskdrop, diskdrop2, old}
	case 3:
		overlays = []*glumby.Texture{hdv1, hdv2, hdv1, hdv2, hdv1, hdv2, old}
	case 4:
		overlays = []*glumby.Texture{pak0, pak1, pak0, pak1, pak0, pak1, old}
	case 5:
		overlays = []*glumby.Texture{tape0, tape1, tape0, tape1, tape0, tape1, old}
	}
	for _, v := range overlays {
		uiTexture = v
		time.Sleep(333 * time.Millisecond)
	}
}

func remapHGR(c int) int {
	if c >= 8 {
		c = c - 8
	} else {
		c = (c & 7) | 16
	}
	return c
}

func ReconfigureHUD(index int, count int, ls *types.LayerSpecMapped) {

	// Nuke this layer its zero width
	if ls.GetWidth() == 0 {
		HUDLayers[index][count] = nil
		return
	}

	// Ok configure it then
	var useTex *glumby.Texture
	var useBM *image.RGBA
	if HUDLayers[index][count] != nil {
		useTex = HUDLayers[index][count].ScreenTex
	}
	HUDLayers[index][count] = video.NewTextLayer(
		int(ls.GetWidth()),
		int(ls.GetHeight()),
		types.CWIDTH,
		types.CHEIGHT,
		RAM.GetHintedMemorySlice(index, ls.GetID()),
		ls,
		useTex,
		useBM,
	)

	HUDLayers[index][count].Format = ls.GetFormat()

	HUDLayers[index][count].PosChanged = true

}

func CheckGFXCamera() {

	return

	for index := 0; index < memory.OCTALYZER_NUM_INTERPRETERS; index++ {
		if RAM.ReadGlobal(index, RAM.MEMBASE(index)+memory.OCTALYZER_CAMERA_GFX_BASE) != 0 {
			cindex := RAM.ReadGlobal(index, RAM.MEMBASE(index)+memory.OCTALYZER_MAPPED_CAM_CONTROL)
			data := RAM.BlockRead(index, RAM.MEMBASE(index)+memory.OCTALYZER_CAMERA_GFX_BASE, memory.OCTALYZER_CAMERA_BUFFER_SIZE)
			//fmt.Printf("*** Camera reconfig: slot %d, code %d\n", index, data[0])
			fxcam[index][int(cindex)].Command(data, index, RAM)
			//os.Exit(1)
			if data[0] != uint64(types.CC_GetJSONR) {
				RAM.WriteGlobalSilent(index, RAM.MEMBASE(index)+memory.OCTALYZER_CAMERA_GFX_BASE, 0)
			}
		}
	}
}

func CheckHUDCamera() {
	for index := 0; index < memory.OCTALYZER_NUM_INTERPRETERS; index++ {
		if RAM.ReadGlobal(index, RAM.MEMBASE(index)+memory.OCTALYZER_CAMERA_HUD_BASE) != 0 {
			//cindex := RAM.ReadGlobal(RAM.MEMBASE(index) + memory.OCTALYZER_CAMERA_HUD_INDEX)
			//fmt.Println("cindex", cindex, index)
			data := RAM.BlockRead(index, RAM.MEMBASE(index)+memory.OCTALYZER_CAMERA_HUD_BASE, memory.OCTALYZER_CAMERA_BUFFER_SIZE)
			pcam[index].Command(data, index, RAM)
			RAM.WriteGlobal(index, RAM.MEMBASE(index)+memory.OCTALYZER_CAMERA_HUD_BASE, 0)
		}
	}
}

func CheckDPAD() {
	if settings.CurrentMouseMode == settings.MM_MOUSE_DPAD {

		//w.SampleDPAD()
		px, py := w.GetPADValues()

		//fmt.Printf("Debug: px=%d, py=%d\n", px, py)

		var p0 int = lastP0
		var p1 int = lastP1

		dv := 32

		if px == 0 {
			p0 = 127
		} else {
			p0 = lastP0 + dv*px
			if p0 < 0 {
				p0 = 0
			} else if p0 > 255 {
				p0 = 255
			}
		}

		if py == 0 {
			p1 = 127
		} else {
			p1 = lastP1 + dv*py
			if p1 < 0 {
				p1 = 0
			} else if p1 > 255 {
				p1 = 255
			}
		}

		//fmt.Printf("PDL0=%d, PDL1=%d\n", p0, p1)

		for i := 0; i < memory.OCTALYZER_NUM_INTERPRETERS; i++ {
			//			if p0 != lastP0 {

			if settings.LogoCameraControl[i] {
				CameraAxis0 = p0
				CameraAxis1 = p1
			}

			RAM.IntSetPaddleValue(i, 0, uint64(p0))
			//			}
			//			if p1 != lastP1 {
			RAM.IntSetPaddleValue(i, 1, uint64(p1))

			//			}
		}

		lastP0 = p0
		lastP1 = p1

		if px == lastdpx && py == lastdpy {
			return
		}

		lastdpx, lastdpy = px, py

	}
}

func ReconfigureGFX(index int, count int, ls *types.LayerSpecMapped) {

	if ls.GetWidth() == 0 {
		GraphicsLayers[index][count] = nil
		return
	}

	if GraphicsLayers[index][count] != nil {
		GraphicsLayers[index][count].Free()
	}

	if GraphicsLayers[index][count] != nil {
		GraphicsLayers[index][count].Done()
	}

	GraphicsLayers[index][count] = video.NewGraphicsLayer(
		int(ls.GetWidth()),
		int(ls.GetHeight()),
		types.CWIDTH,
		types.CHEIGHT,
		ls.GetFormat(),
		RAM.GetHintedMemorySlice(index, ls.GetID()),
		ls,
	)

	id := ls.GetID()
	if strings.HasPrefix(id, "HGR") {
		settings.LastRenderModeHGR[index] = 0
	} else if strings.HasPrefix(id, "DHR") {
		settings.LastRenderModeDHGR[index] = 0
	} else if id == "LOGR" || id == "DLGR" || id == "DLG2" || id == "LGR2" {
		settings.LastRenderModeGR[index] = 0
	} else if strings.HasPrefix(id, "SHR") {
		settings.LastRenderModeSHR[index] = 0
	} else if strings.HasPrefix(id, "SCR") {
		settings.LastRenderModeSpectrum[index] = 0
	}

	GraphicsLayers[index][count].TransparentIndex = []int{0}

	if ls.GetFormat() == types.LF_HGR_WOZ {
		GraphicsLayers[index][count].TransparentIndex = []int{0, 4}
	} else if ls.GetFormat() == types.LF_HGR_LINEAR {
		GraphicsLayers[index][count].TransparentIndex = []int{0, 16, 20}
	} else if ls.GetFormat() == types.LF_DHGR_WOZ {
		GraphicsLayers[index][count].TransparentIndex = []int{0}
	}

}

// CheckLayersForDirtyAllocs looks for any layer where the dirty flag has been set ...
// The only time this will be set is if the Layers underlying memory map has changed ...
// Since all other info for the layers comes directly from memory then its the *ONLY*
// time we need to touch anything regarding a layer config.
func CheckLayersForDirtyAllocs() {

	for index := 0; index < memory.OCTALYZER_NUM_INTERPRETERS; index++ {

		// skip if slot not active...
		if RAM.IntGetActiveState(index) == 0 && RAM.IntGetLayerForceState(index) == 0 {
			ramActiveState[index] = 0
			continue
		} else {
			ramActiveState[index] = 1
		}

		//fmt.Printf("S%d ", index)

		var ls *types.LayerSpecMapped

		// Now slot is active, might be dirty HUD layers
		for i := 0; i < memory.OCTALYZER_MAX_HUD_LAYERS; i++ {
			ls = HUDSpecs[index][i]
			if !ls.GetDirty() {
				continue // skip this layer, no change
			}
			//fmt.Println("Lets reconfigure a HUD layer, yall:", ls.String())
			// if we are here, we need to redo the mmap

			ReconfigureHUD(index, i, ls)
			// settings.LastTintMode[index] = RAM.IntGetVideoTint(i) + 1
			// if settings.LastTintMode[index] >= settings.VPT_MAX {
			// 	settings.LastTintMode[index] = settings.VPT_NONE
			// }
			// CheckTintMode()
			ls.SetDirty(false)
		}

		// Now slot is active, might be dirty GFX layers
		for i := 0; i < memory.OCTALYZER_MAX_GFX_LAYERS; i++ {
			ls = GFXSpecs[index][i]
			if !ls.GetDirty() {
				continue // skip this layer, no change
			}
			//fmt.Println("Lets reconfigure a GFX layer, yall:", ls.String())

			// if we are here, we need to redo the mmap
			ReconfigureGFX(index, i, ls)
			// settings.LastTintMode[index] = RAM.IntGetVideoTint(i) + 1
			// if settings.LastTintMode[index] >= settings.VPT_MAX {
			// 	settings.LastTintMode[index] = settings.VPT_NONE
			// }
			// CheckTintMode()
			ls.SetDirty(false)
		}

		//fmt.Println()

	}

}

func CheckTintMode() {

	if settings.UnifiedRender[SelectedIndex] {
		return
	}

	v := RAM.IntGetVideoTint(SelectedIndex)

	// Tint is disabled for Unified Render
	if (settings.UnifiedRender[SelectedIndex] || settings.UnifiedRenderGlobal) && ! strings.Contains(settings.SpecFile[SelectedIndex], "spectrum") {
		v = settings.VPT_NONE
	}

	if v != settings.LastTintMode[SelectedIndex] || settings.LastScanLineIntensity != settings.ScanLineIntensity ||
		settings.DisableScanlines != settings.LastDisableScanlines {

		//log2.Printf("last tint = %d, new mode = %d", settings.LastTintMode[SelectedIndex], v)

		settings.LastTintMode[SelectedIndex] = v
		settings.LastScanLineIntensity = settings.ScanLineIntensity
		settings.LastDisableScanlines = settings.DisableScanlines

		//log2.Printf("after last tint = %d, new mode = %d", settings.LastTintMode[SelectedIndex], v)

		layerset := GraphicsLayers[SelectedIndex]
		for _, layer := range layerset {
			if layer != nil {
				switch v {
				case settings.VPT_NONE:
					layer.Tint = nil
				case settings.VPT_AMBER:
					layer.Tint = types.NewVideoColor(255, 115, 0, 255)
				case settings.VPT_GREEN:
					layer.Tint = types.NewVideoColor(103, 253, 146, 255)
				case settings.VPT_GREY:
					layer.Tint = types.NewVideoColor(255, 255, 255, 255)
				default:
					r, g, b, a := RAM.IntGetVideoTintRGBA(SelectedIndex)
					layer.Tint = types.NewVideoColor(r, g, b, a)
				}
				layer.TintChanged = true
			}
		}
		tlayerset := HUDLayers[SelectedIndex]
		for _, layer := range tlayerset {
			if layer != nil {
				switch v {
				case settings.VPT_NONE:
					layer.Tint = nil
				case settings.VPT_AMBER:
					layer.Tint = types.NewVideoColor(255, 115, 0, 255)
				case settings.VPT_GREEN:
					layer.Tint = types.NewVideoColor(103, 253, 146, 255)
				case settings.VPT_GREY:
					layer.Tint = types.NewVideoColor(255, 255, 255, 255)
				default:
					r, g, b, a := RAM.IntGetVideoTintRGBA(SelectedIndex)
					layer.Tint = types.NewVideoColor(r, g, b, a)
				}
				layer.TintChanged = true
			}
		}

	}
}

func CheckRenderModeLORES() {

	CheckRenderModeGR()

	mode := GetVideoMode(SelectedIndex)

	if mode != "LOGR" && mode != "DLGR" && mode != "DLG2" && mode == "LGR2" {
		return
	}

	if light.Diffuse[0] != 0.5 {

		light.Diffuse = []float32{0.5, 0.5, 0.5, 1}
		light.Ambient = []float32{0.5, 0.5, 0.5, 1}
		light.Update()

	}

}

func CheckRenderModeHGR() {

	var forceTint bool

	for index := 0; index < memory.OCTALYZER_NUM_INTERPRETERS; index++ {

		v := RAM.IntGetHGRRender(index)

		var changed bool

		if v != settings.LastRenderModeHGR[index] {
			fmt.Printf("HGR: Render mode has changed %s -> %s...\n", settings.LastRenderModeHGR[index].String(), v.String())

			layerset := GraphicsLayers[index]
			for _, layer := range layerset {

				if layer != nil {

					changed = true

					id := layer.Spec.GetID()
					if id != "HGR1" && id != "HGR2" && id != "XGR1" && id != "XGR2" {
						continue
					}

					fmt.Println(layer.Spec.GetID())

					switch v {
					case settings.VM_DOTTY:
						layer.Spec.SetMono(false)
						layer.SetSubFormat(types.LSF_FREEFORM, true)
						if layer.Spec.GetActive() {
							light.Diffuse = []float32{0.5, 0.5, 0.5, 1}
							light.Ambient = []float32{0.5, 0.5, 0.5, 1}
							light.Update()
							fmt.Println("HGR dotty")
						}

					case settings.VM_VOXELS:
						layer.Spec.SetMono(false)
						layer.SetSubFormat(types.LSF_VOXELS, true)
						if layer.Spec.GetActive() {
							light.Diffuse = []float32{0.5, 0.5, 0.5, 1}
							light.Ambient = []float32{0.5, 0.5, 0.5, 1}
							light.Update()
							fmt.Println("HGR voxels")
						}

					case settings.VM_MONO_VOXELS:
						layer.Spec.SetMono(true)
						layer.SetSubFormat(types.LSF_VOXELS, true)
						if layer.Spec.GetActive() {
							light.Diffuse = []float32{0.8, 0.8, 0.8, 1}
							light.Ambient = []float32{0.1, 0.1, 0.1, 1}
							light.Update()
							fmt.Println("HGR mono voxels")
						}

					case settings.VM_MONO_DOTTY:
						layer.Spec.SetMono(true)
						layer.SetSubFormat(types.LSF_FREEFORM, true)
						if layer.Spec.GetActive() {
							light.Diffuse = []float32{0.5, 0.5, 0.5, 1}
							light.Ambient = []float32{0.5, 0.5, 0.5, 1}
							light.Update()
							fmt.Println("HGR mono dotty")
						}

					case settings.VM_FLAT:
						fmt.Println("HGR flat")
						layer.Spec.SetMono(false)
						layer.SetSubFormat(types.LSF_SINGLE_LAYER, true)
						if layer.Spec.GetActive() {
							light.Diffuse = []float32{1, 1, 1, 1}
							light.Ambient = []float32{0.7, 0.7, 0.7, 1}
							light.Update()
							fmt.Println("HGR flat")
						}

					case settings.VM_MONO_FLAT:
						fmt.Println("HGR flat mono")
						layer.Spec.SetMono(true)
						layer.SetSubFormat(types.LSF_SINGLE_LAYER, true)
						if layer.Spec.GetActive() {
							light.Diffuse = []float32{1, 1, 1, 1}
							light.Ambient = []float32{0.7, 0.7, 0.7, 1}
							light.Update()
							fmt.Println("HGR mono flat")
						}
					default:
						fmt.Println("No match")
					}
				}
			}

			if changed {
				// force retint
				settings.LastTintMode[index] = settings.VPT_NONE
				forceTint = true
			}

			//fmt.Printf("mono = %v\n", settings.DHGRMono[index])

			settings.LastRenderModeHGR[index] = v
		}

	}

	if forceTint {
		CheckTintMode()
	}

}

func CheckRenderModeSHR() {

	var forceTint bool

	for index := 0; index < memory.OCTALYZER_NUM_INTERPRETERS; index++ {

		v := RAM.IntGetSHRRender(index)

		var changed bool

		if v != settings.LastRenderModeSHR[index] {
			//log2.Printf("SHR: Render mode has changed %s -> %s...\n", settings.LastRenderModeSHR[index].String(), v.String())

			layerset := GraphicsLayers[index]
			for _, layer := range layerset {

				if layer != nil {

					changed = true

					id := layer.Spec.GetID()
					if id != "SHR1" {
						continue
					}

					fmt.Println(layer.Spec.GetID())

					switch v {
					case settings.VM_DOTTY:
						layer.Spec.SetMono(false)
						layer.SetSubFormat(types.LSF_FREEFORM, true)
						if layer.Spec.GetActive() {
							light.Diffuse = []float32{0.5, 0.5, 0.5, 1}
							light.Ambient = []float32{0.5, 0.5, 0.5, 1}
							light.Update()
							fmt.Println("HGR dotty")
						}

					case settings.VM_VOXELS:
						layer.Spec.SetMono(false)
						layer.SetSubFormat(types.LSF_VOXELS, true)
						if layer.Spec.GetActive() {
							light.Diffuse = []float32{0.8, 0.8, 0.8, 1}
							light.Ambient = []float32{0.1, 0.1, 0.1, 1}
							light.Update()
							fmt.Println("HGR voxels")
						}

					case settings.VM_FLAT:
						fmt.Println("HGR flat")
						layer.Spec.SetMono(false)
						layer.SetSubFormat(types.LSF_SINGLE_LAYER, true)
						if layer.Spec.GetActive() {
							light.Diffuse = []float32{1, 1, 1, 1}
							light.Ambient = []float32{0.7, 0.7, 0.7, 1}
							light.Update()
							fmt.Println("HGR flat")
						}

					default:
						fmt.Println("No match")
					}
				}
			}

			if changed {
				// force retint
				settings.LastTintMode[index] = settings.VPT_NONE
				forceTint = true
			}

			//fmt.Printf("mono = %v\n", settings.DHGRMono[index])

			settings.LastRenderModeSHR[index] = v
		}

	}

	if forceTint {
		CheckTintMode()
	}

}

func CheckRenderModeSPECCY() {

	var forceTint bool

	for index := 0; index < memory.OCTALYZER_NUM_INTERPRETERS; index++ {

		v := RAM.IntGetSpectrumRender(index)

		var changed bool

		if v != settings.LastRenderModeSpectrum[index] {
			//log2.Printf("SHR: Render mode has changed %s -> %s...\n", settings.LastRenderModeSHR[index].String(), v.String())

			layerset := GraphicsLayers[index]
			for _, layer := range layerset {

				if layer != nil {

					changed = true

					id := layer.Spec.GetID()
					if id != "SCRN" && id != "SCR2" {
						continue
					}

					//fmt.Println(layer.Spec.GetID())

					switch v {
					case settings.VM_DOTTY:
						layer.Spec.SetMono(false)
						layer.SetSubFormat(types.LSF_FREEFORM, true)
						if layer.Spec.GetActive() {
							light.Diffuse = []float32{0.5, 0.5, 0.5, 1}
							light.Ambient = []float32{0.5, 0.5, 0.5, 1}
							light.Update()
							fmt.Println("SCRN dotty")
						}

					case settings.VM_VOXELS:
						layer.Spec.SetMono(false)
						layer.SetSubFormat(types.LSF_VOXELS, true)
						if layer.Spec.GetActive() {
							light.Diffuse = []float32{0.8, 0.8, 0.8, 1}
							light.Ambient = []float32{0.1, 0.1, 0.1, 1}
							light.Update()
							fmt.Println("SCRN voxels")
						}

					case settings.VM_FLAT:
						fmt.Println("SCRN flat")
						layer.Spec.SetMono(false)
						layer.SetSubFormat(types.LSF_SINGLE_LAYER, true)
						if layer.Spec.GetActive() {
							light.Diffuse = []float32{1, 1, 1, 1}
							light.Ambient = []float32{0.7, 0.7, 0.7, 1}
							light.Update()
							fmt.Println("SCRN flat")
						}

					default:
						fmt.Println("No match")
					}
				}
			}

			if changed {
				// force retint
				settings.LastTintMode[index] = settings.VPT_NONE
				forceTint = true
			}

			//fmt.Printf("mono = %v\n", settings.DHGRMono[index])

			settings.LastRenderModeSpectrum[index] = v
		}

	}

	if forceTint {
		CheckTintMode()
	}

}

func CheckRenderModeGR() {

	var forceTint bool

	for index := 0; index < memory.OCTALYZER_NUM_INTERPRETERS; index++ {

		v := RAM.IntGetGRRender(index)

		var changed bool

		if v != settings.LastRenderModeGR[index] {
			fmt.Printf("GR: Render mode has changed %s -> %s...\n", settings.LastRenderModeGR[index].String(), v.String())

			layerset := GraphicsLayers[index]
			for _, layer := range layerset {

				if layer != nil {

					changed = true

					id := layer.Spec.GetID()
					if id != "LOGR" && id != "DLGR" && id != "DLG2" && id != "LGR2" {
						continue
					}

					fmt.Println(layer.Spec.GetID())

					switch v {
					case settings.VM_VOXELS:
						layer.Spec.SetMono(false)
						layer.SetSubFormat(types.LSF_VOXELS, true)
						if layer.Spec.GetActive() {
							light.Diffuse = []float32{0.8, 0.8, 0.8, 1}
							light.Ambient = []float32{0.1, 0.1, 0.1, 1}
							light.Update()
							fmt.Println("HGR voxels")
						}
					case settings.VM_FLAT:
						fmt.Println("HGR flat")
						layer.Spec.SetMono(false)
						layer.SetSubFormat(types.LSF_SINGLE_LAYER, true)
						if layer.Spec.GetActive() {
							light.Diffuse = []float32{1, 1, 1, 1}
							light.Ambient = []float32{0.7, 0.7, 0.7, 1}
							light.Update()
							fmt.Println("HGR flat")
						}
					default:
						layer.Spec.SetMono(false)
						layer.SetSubFormat(types.LSF_VOXELS, true)
						if layer.Spec.GetActive() {
							light.Diffuse = []float32{0.8, 0.8, 0.8, 1}
							light.Ambient = []float32{0.1, 0.1, 0.1, 1}
							light.Update()
							fmt.Println("HGR voxels")
						}
					}
				}
			}

			if changed {
				// force retint
				settings.LastTintMode[index] = settings.VPT_NONE
				forceTint = true
			}

			//fmt.Printf("mono = %v\n", settings.DHGRMono[index])

			settings.LastRenderModeGR[index] = v
		}

	}

	if forceTint {
		CheckTintMode()
	}

}

func CheckVoxelDepth() {

	for index := 0; index < memory.OCTALYZER_NUM_INTERPRETERS; index++ {

		v := RAM.IntGetVoxelDepth(index)

		if v != settings.LastVoxelDepth[index] {
			fmt.Printf("Voxel depth has changed %d -> %d...\n", settings.LastVoxelDepth[index], v)

			layerset := GraphicsLayers[index]
			for _, layer := range layerset {

				if layer != nil {

					id := layer.Spec.GetID()
					if id != "SCRN" && id != "SCR2" && id != "SHR1" && id != "HGR1" && id != "HGR2" && id != "XGR1" && id != "XGR2" && id != "LOGR" && id != "LGR2" && id != "DLGR" && id != "DLG2" && id != "DHR1" && id != "DHR2" {
						continue
					}

					layer.SetVoxelDepth(v)
				}
			}

			settings.LastVoxelDepth[index] = v
		}

	}

}

func UpdateLighting(v settings.VideoMode) {
	switch v {
	case settings.VM_DOTTY:
		light.Diffuse = []float32{0.5, 0.5, 0.5, 1}
		light.Ambient = []float32{0.5, 0.5, 0.5, 1}
		//light.Update()
	case settings.VM_VOXELS:
		light.Diffuse = []float32{0.8, 0.8, 0.8, 1}
		light.Ambient = []float32{0.1, 0.1, 0.1, 1}
		//light.Update()
	case settings.VM_MONO_VOXELS:
		light.Diffuse = []float32{0.8, 0.8, 0.8, 1}
		light.Ambient = []float32{0.1, 0.1, 0.1, 1}
		//light.Update()
	case settings.VM_MONO_DOTTY:
		light.Diffuse = []float32{0.5, 0.5, 0.5, 1}
		light.Ambient = []float32{0.5, 0.5, 0.5, 1}
		//light.Update()
	case settings.VM_FLAT:
		light.Diffuse = []float32{1, 1, 1, 1}
		light.Ambient = []float32{0.7, 0.7, 0.7, 1}
		//light.Update()
	case settings.VM_MONO_FLAT:
		light.Diffuse = []float32{1, 1, 1, 1}
		light.Ambient = []float32{0.7, 0.7, 0.7, 1}
		//light.Update()
	}
}

func CheckRenderModeDHGR() {

	var forceTint bool

	for index := 0; index < memory.OCTALYZER_NUM_INTERPRETERS; index++ {

		v := RAM.IntGetDHGRRender(index)
		var changed bool

		if v != settings.LastRenderModeDHGR[index] {
			//fmt.Printf("DHGR: Render mode has changed %d -> %d...\n", settings.LastRenderModeDHGR[index], v)

			changed = true

			layerset := GraphicsLayers[index]
			for _, layer := range layerset {

				if layer != nil {

					id := layer.Spec.GetID()
					//fmt.Println(id)
					if id != "DHR1" && id != "DHR2" {
						continue
					}

					switch v {
					case settings.VM_DOTTY:
						layer.Spec.SetMono(false)
						layer.SetSubFormat(types.LSF_FREEFORM, true)
						if layer.Spec.GetActive() {
							light.Diffuse = []float32{0.5, 0.5, 0.5, 1}
							light.Ambient = []float32{0.5, 0.5, 0.5, 1}
							light.Update()
							//fmt.Println("DHGR dotty")
						}

					case settings.VM_VOXELS:
						layer.Spec.SetMono(false)
						layer.SetSubFormat(types.LSF_VOXELS, true)
						if layer.Spec.GetActive() {
							light.Diffuse = []float32{0.8, 0.8, 0.8, 1}
							light.Ambient = []float32{0.1, 0.1, 0.1, 1}
							light.Update()
							//fmt.Println("DHGR voxels")
						}

					case settings.VM_MONO_VOXELS:
						layer.Spec.SetMono(true)
						layer.SetSubFormat(types.LSF_VOXELS, true)
						if layer.Spec.GetActive() {
							light.Diffuse = []float32{0.8, 0.8, 0.8, 1}
							light.Ambient = []float32{0.1, 0.1, 0.1, 1}
							light.Update()
							//fmt.Println("DHGR mono voxels")
						}

					case settings.VM_MONO_DOTTY:
						layer.Spec.SetMono(true)
						layer.SetSubFormat(types.LSF_FREEFORM, true)
						if layer.Spec.GetActive() {
							light.Diffuse = []float32{0.5, 0.5, 0.5, 1}
							light.Ambient = []float32{0.5, 0.5, 0.5, 1}
							light.Update()
							//fmt.Println("DHGR mono dotty")
						}

					case settings.VM_FLAT:
						layer.Spec.SetMono(false)
						layer.SetSubFormat(types.LSF_SINGLE_LAYER, true)
						if layer.Spec.GetActive() {
							light.Diffuse = []float32{1, 1, 1, 1}
							light.Ambient = []float32{0.7, 0.7, 0.7, 1}
							light.Update()
							//fmt.Println("DHGR flat")
						}

					case settings.VM_MONO_FLAT:
						layer.Spec.SetMono(true)
						layer.SetSubFormat(types.LSF_SINGLE_LAYER, true)
						if layer.Spec.GetActive() {
							light.Diffuse = []float32{1, 1, 1, 1}
							light.Ambient = []float32{0.7, 0.7, 0.7, 1}
							light.Update()
							//fmt.Println("DHGR mono flat")
						}

						// case settings.VM_LAYERS:
						//      settings.DHGRMono = false
						//      layer.SetSubFormat(types.LSF_COLOR_LAYER, true)
						//                              case settings.VM_GREY_LAYERS:
						//                                      settings.PaletteTint = settings.VPT_GREY
						//                                      settings.DHGRMono = false
						//                                      layer.SetSubFormat(types.LSF_GREY_LAYER, true)
						//                              case settings.VM_GREEN_LAYERS:
						//                                      settings.PaletteTint = settings.VPT_GREEN
						//                                      settings.DHGRMono = false
						//                                      layer.SetSubFormat(types.LSF_GREEN_LAYER, true)
						//                              case settings.VM_AMBER_LAYERS:
						//                                      settings.PaletteTint = settings.VPT_AMBER
						//                                      settings.DHGRMono = false
						//                                      layer.SetSubFormat(types.LSF_AMBER_LAYER, true)
					}
				}
			}

			if changed {
				// force retint
				settings.LastTintMode[index] = settings.VPT_NONE
				forceTint = true
			}

			settings.LastRenderModeDHGR[index] = v

		}

	}

	if forceTint {
		CheckTintMode()
	}

}

func renderBG(w *glumby.Window) {
	if bgTexture == nil {
		return
	}

	a := float32(fxcam[SelectedIndex][bgcamidx].GetAspect())

	if a != bgAspect {
		bgdecal = video.NewDecal(types.CWIDTH, types.CHEIGHT)
		bgdecal.Mesh = video.GetPlaneAsTrianglesInv(types.CHEIGHT*1.05*a, types.CHEIGHT*1.05)
		bgAspect = a
	}

	//fmt.Printf("%v :: Rendering bg with slot %d, cam %d\n", time.Now(), SelectedIndex, bgcamidx)
	//gl.Disable(gl.ALPHA_TEST)
	gl.PushMatrix()
	width, height := w.GetGLFWWindow().GetFramebufferSize()
	gl.Viewport(0, 0, int32(width), int32(height))
	lp := types.LayerPosMod{}
	if bgZoomFactor != 0 {
		fxcam[SelectedIndex][bgcamidx].SetZoom(
			(fxcam[SelectedIndex][0].GetZoom()-types.GFXMULT)*bgZoomFactor + bgZoom,
		)
	} else {
		fxcam[SelectedIndex][bgcamidx].SetZoom(bgZoom)
	}

	if bgCamTrack {
		angle := fxcam[SelectedIndex][0].GetAngle()
		pos := fxcam[SelectedIndex][0].GetPosition()
		fxcam[SelectedIndex][bgcamidx].SetPosition(pos)
		fxcam[SelectedIndex][bgcamidx].SetRotation(angle)
	}

	fxcam[SelectedIndex][bgcamidx].ApplyWindow(&playfield, lp, tspace, waspect, fov, 1)
	gl.Enable(gl.TEXTURE_2D)
	bgTexture.Bind()
	glumby.MeshBuffer_Begin(gl.TRIANGLES)
	bgdecal.Mesh.SetColor(1, 1, 1, bgOpacity)
	//fmt.Printf("bgX=%f, bgY=%f, bgZ=%f\n", bgX, bgY, bgZ)
	bgdecal.Mesh.DrawWithMeshBuffer(float32(bgX), float32(bgY), float32(bgZ))
	glumby.MeshBuffer_End()
	bgTexture.Unbind()
	gl.Disable(gl.TEXTURE_2D)
	//gl.Flush()
	gl.PopMatrix()
}

func renderOverlay(w *glumby.Window, t *glumby.Texture) {
	if t == nil {
		return
	}

	if ovdecal == nil || ovdecal.Mesh == nil {
		ovdecal = video.NewDecal(types.CWIDTH, types.CHEIGHT)
		ovdecal.Mesh = video.GetPlaneAsTrianglesInv(types.CHEIGHT*16/9, types.CHEIGHT)
	}

	//fmt.Printf("%v :: Rendering overlay with slot %d\n", time.Now(), SelectedIndex)
	gl.PushMatrix()
	width, height := w.GetGLFWWindow().GetFramebufferSize()
	gl.Viewport(0, 0, int32(width), int32(height))

	lp := types.LayerPosMod{}
	//pcam[SelectedIndex].ApplyWindow(&playfield, lp, tspace, waspect, fov, 1)
	osd[SelectedIndex].SetAspect(1.788889)
	osd[SelectedIndex].ApplyWindow(&playfieldOSD, lp, tspaceOSD, waspectOSD, fovOSD, 1)

	gl.Enable(gl.ALPHA_TEST)
	gl.Enable(gl.TEXTURE_2D)
	t.Bind()
	glumby.MeshBuffer_Begin(gl.TRIANGLES)
	ovdecal.Mesh.SetColor(1, 1, 1, 1)
	ovdecal.Mesh.DrawWithMeshBuffer(types.CWIDTH/2, types.CHEIGHT/2, -1)
	glumby.MeshBuffer_End()
	t.Unbind()
	gl.Disable(gl.TEXTURE_2D)
	gl.Disable(gl.ALPHA_TEST)
	gl.Flush()
	gl.PopMatrix()
}

func renderUnified(w *glumby.Window, t *glumby.Texture) {
	if t == nil {
		return
	}

	if undecal == nil || undecal.Mesh == nil {
		undecal = video.NewDecal(types.CWIDTH, types.CHEIGHT)
	}

	undecal.Mesh = video.GetPlaneAsTrianglesInv(types.CHEIGHT*float32(aspectRatios[aspectRatioIndex[SelectedIndex]]), types.CHEIGHT)

	gl.PushMatrix()
	width, height := w.GetGLFWWindow().GetFramebufferSize()
	gl.Viewport(0, 0, int32(width), int32(height))

	lp := types.LayerPosMod{}
	pcam[SelectedIndex].ApplyWindow(&playfield, lp, tspace, waspect, fov, 1)
	gl.Enable(gl.ALPHA_TEST)
	gl.Enable(gl.TEXTURE_2D)
	t.Bind()
	glumby.MeshBuffer_Begin(gl.TRIANGLES)
	undecal.Mesh.SetColor(1, 1, 1, 1)
	undecal.Mesh.DrawWithMeshBuffer(types.CWIDTH/2, types.CHEIGHT/2, -1)
	glumby.MeshBuffer_End()
	t.Unbind()
	gl.Disable(gl.TEXTURE_2D)
	gl.Disable(gl.ALPHA_TEST)
	gl.Flush()
	gl.PopMatrix()
}

func OnRenderWindow(w *glumby.Window) {

	//return

	CheckAndProcessSyncRequests()

	if settings.BlueScreen {
		// we do nothing here
		if splash.Mesh == nil {
			splash.Mesh = video.GetPlaneAsTrianglesInv(types.CWIDTH, types.CHEIGHT)
		}
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		gl.ClearDepth(1)
		gl.ClearColor(0, 0, 0, 1)
		gl.Enable(gl.BLEND)
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
		gl.Enable(gl.DEPTH_TEST)
		gl.DepthFunc(gl.LEQUAL)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		gl.AlphaFunc(gl.GREATER, 0.3)
		//gl.Disable(gl.ALPHA_TEST)
		gl.PushMatrix()
		width, height := w.GetGLFWWindow().GetFramebufferSize()
		gl.Viewport(0, 0, int32(width), int32(height))
		lp := types.LayerPosMod{}
		pcam[0].ApplyWindow(&playfield, lp, tspace, waspect, fov, 1)
		gl.Enable(gl.TEXTURE_2D)
		splash.Texture.Bind()
		glumby.MeshBuffer_Begin(gl.TRIANGLES)
		splash.Mesh.DrawWithMeshBuffer(types.CWIDTH/2, types.CHEIGHT/2, 0)
		glumby.MeshBuffer_End()
		splash.Texture.Unbind()
		gl.Disable(gl.TEXTURE_2D)
		gl.PopMatrix()
		return
	}

	if backend.ProducerMain == nil {
		return
	}

	SelectedCameraIndex = getSelectedCameraIndex(SelectedIndex)

	//RAM.WriteInterpreterMemory(0, 0xc019, 0)

	if backend.REBOOT_NEEDED {
		backend.REBOOT_NEEDED = false
		initCameras()

		return // no frame draw
	}

	UpdateCameraControlState()

	//clickMonitor()
	SampleMouse(w)
	checkFullscreen()
	bgMonitor()
	logMonitor()
	CheckKeyInserts()
	CheckMetaMode(SelectedIndex)
	//checkAudioPlay()
	checkBG()
	checkOverlay()

	////fmt.Println("Render")
	PollKeyCase()
	processControllerEvents()

	//CheckDPAD()
	CheckHUDCamera()
	CheckGFXCamera()
	CheckRenderModeHGR()
	CheckRenderModeDHGR()
	CheckRenderModeLORES()
	CheckRenderModeSHR()
	CheckRenderModeSPECCY()
	CheckVoxelDepth()
	CheckTintMode()
	CheckLayersForDirtyAllocs()

	if fxcam[SelectedCamera][SelectedCameraIndex] == nil {
		initCameraInSlot(SelectedCamera)
	}

	pp := fxcam[SelectedCamera][SelectedCameraIndex].GetPosition()

	light.Position = []float32{0, float32(pp[0]), float32(pp[1]), float32(pp[2]) / types.GFXMULT}
	light.Update()
	light.Off()

	// light2.Position = []float32{0, float32(pp[0]), float32(pp[1]), float32(pp[2]) / types.GFXMULT}
	// light2.Update()
	light2.Off()

	r, g, b := float32(BGColor.R)/255, float32(BGColor.G)/255, float32(BGColor.B)/255 //, float32(BGColor.A)/255
	gl.ClearColor(r, g, b, 1)

	gl.ClearDepth(1)

	gl.ShadeModel(gl.SMOOTH)
	gl.Hint(gl.PERSPECTIVE_CORRECTION_HINT, gl.NICEST)
	//gl.TexEnvf(gl.TEXTURE_ENV, gl.TEXTURE_ENV_MODE, gl.MODULATE)

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LEQUAL)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.AlphaFunc(gl.GREATER, 0.3)
	//gl.Disable(gl.ALPHA_TEST)

	gl.Color4f(1, 1, 1, 1)

	width, height := w.GetGLFWWindow().GetFramebufferSize()
	gl.Viewport(0, 0, int32(width), int32(height))

	RAM.LockForVideo(true)

	if backend.ProducerMain == nil {
		RAM.LockForVideo(false)
		return
	}

	if !settings.HamburgerOnHover && !RAM.IntGetSlotMenu(SelectedIndex) {
		drawHamburger(backend.ProducerMain.GetInterpreter(SelectedIndex))
	}

	// FX layers
	//overlay.Unbind()
	//~ gl.Disable(gl.ALPHA_TEST)

	if !settings.UnifiedRender[SelectedIndex] || strings.Contains(settings.SpecFile[SelectedIndex], "spectrum"){
		light.Off()
		light2.On()
		// light2.SetAmbientLevel(RAM.IntGetAmbientLevel(SelectedIndex))
		// light2.SetDiffuseLevel(RAM.IntGetDiffuseLevel(SelectedIndex))
		RenderTextLayers()
		light2.Off()
		light.On()
		light.SetAmbientLevel(RAM.IntGetAmbientLevel(SelectedIndex))
		light.SetDiffuseLevel(RAM.IntGetDiffuseLevel(SelectedIndex))
		// light.SetAmbientLevel(0.3)
		// light.SetDiffuseLevel(0.8)
		renderBG(w)
		RenderGraphicsLayers()
	} else {

		if settings.DisableMetaMode[SelectedIndex] {
			light.Off()
			light2.On()
			// light2.SetAmbientLevel(RAM.IntGetAmbientLevel(SelectedIndex))
			// light2.SetDiffuseLevel(RAM.IntGetDiffuseLevel(SelectedIndex))
			RenderTextLayers()
		} else {
			CheckSHRState()

			if SHRActive {
				//log2.Printf("SHR is visible!")
				light2.Off()
				light.On()
				light.SetAmbientLevel(RAM.IntGetAmbientLevel(SelectedIndex))
				light.SetDiffuseLevel(RAM.IntGetDiffuseLevel(SelectedIndex))
				RenderGraphicsLayers()
			} else {
				light2.On()
				updateUnified()
				renderUnified(w, unifiedTexture)
			}
		}
	}

	//gl.Enable(gl.ALPHA_TEST)
	gl.PushMatrix()
	index := 0
	lp := backend.ProducerMain.MasterLayerPos[index]
	pcam[index].ApplyWindow(&playfield, lp, tspace, waspect, fov, 1)
	drawLEDs(types.CWIDTH-48, types.CHEIGHT-35)
	gl.PopMatrix()
	//gl.Disable(gl.ALPHA_TEST)

	if !settings.DisableOverlays {
		renderOverlay(w, ovTexture)
		renderOverlay(w, uiTexture)
	}

	RenderTextLayersOSD()

	RAM.LockForVideo(false)

	if video.SnapShot && ScreenLogging {
		video.SnapCount++
		ww, hh := w.GetGLFWWindow().GetFramebufferSize()
		ScreenShotPNG(0, 0, ww, hh, fmt.Sprintf("snapshot-%d.png", video.SnapCount))
	}

	if settings.TakeScreenshot {
		settings.TakeScreenshot = false
		SnapLayers()
	}

	if settings.ScreenShotNeeded {
		ww, hh := w.GetGLFWWindow().GetFramebufferSize()
		d := ScreenShotPNGBytes(0, 0, ww, hh)
		if len(d) > 0 {
			settings.ScreenShotJPEGData = d
		}
		settings.ScreenShotNeeded = false
	}

	video.SnapShot = false

}

func OnEventWindow(w *glumby.Window) {
	//	log.Println("Event called")
}

func OnCharEvent(w *glumby.Window, ch rune) {

}

var inBurger bool
var burgerRect = types.LayerRect{1, 1, 7, 7}

func SampleMouse(w *glumby.Window) {

	var stickx, sticky float64
	var dx, dy float64

	//for {

	if settings.CurrentMouseMode == settings.MM_MOUSE_GEOS || settings.CurrentMouseMode == settings.MM_MOUSE_DDRAW {
		dy = my - lypc
		dx = mx - lxpc

		stickx = 0
		sticky = 0

		var maxx, maxy float64

		if math.Abs(dy) > math.Abs(dx) {
			// y greater
			if dy > 0 {
				sticky = 1
				maxy = settings.MaxStickLevel[settings.StickDOWN]
			} else {
				sticky = -1
				maxy = settings.MaxStickLevel[settings.StickUP]
			}
			if dx > 0 {
				stickx = math.Abs(dx) / math.Abs(dy)
				maxx = settings.MaxStickLevel[settings.StickRIGHT]
			} else {
				stickx = -math.Abs(dx) / math.Abs(dy)
				maxx = settings.MaxStickLevel[settings.StickLEFT]
			}
		} else if math.Abs(dx) > math.Abs(dy) {
			// x greater
			if dx > 0 {
				stickx = 1
				maxx = settings.MaxStickLevel[settings.StickRIGHT]
			} else {
				stickx = -1
				maxx = settings.MaxStickLevel[settings.StickLEFT]
			}
			if dy > 0 {
				sticky = math.Abs(dy) / math.Abs(dx)
				maxy = settings.MaxStickLevel[settings.StickDOWN]
			} else {
				sticky = -math.Abs(dy) / math.Abs(dx)
				maxy = settings.MaxStickLevel[settings.StickUP]
			}
		} else if math.Abs(dx) > 0 && math.Abs(dy) > 0 {
			if dx > 0 {
				stickx = 1
			} else {
				stickx = -1
			}
			if dy > 0 {
				sticky = 1
				maxy = settings.MaxStickLevel[settings.StickDOWN]
			} else {
				sticky = -1
				maxy = settings.MaxStickLevel[settings.StickUP]
			}
		}

		// update paddles
		RAM.IntSetPaddleValue(SelectedIndex, 0, uint64(settings.StickCenter[0]+maxx*stickx))
		RAM.IntSetPaddleValue(SelectedIndex, 1, uint64(settings.StickCenter[1]+maxy*sticky))

		if settings.LogoCameraControl[SelectedIndex] {
			CameraAxis0 = int(settings.StickCenter[0] + maxx*stickx)
			CameraAxis1 = int(settings.StickCenter[1] + maxy*sticky)
		}
	}

	// sleep for sample interval
	lxpc, lypc = mx, my
	//time.Sleep(settings.MouseStickInterval)
	//}

}

func OnMouseMoveEvent(w *glumby.Window, x, y float64) {

	if settings.BlueScreen || backend.ProducerMain == nil {
		return
	}

	deltaX, deltaY := x-mx, y-my
	mx, my = x, y

	oxpc, oypc := WindowXYToIndexOSD(SelectedIndex, x, y)

	cx, cy := uint16(oxpc*80), uint16(oypc*48)

	in := burgerRect.Contains(cx, cy)
	if (in != inBurger || !settings.HamburgerOnHover) && !RAM.IntGetSlotMenu(SelectedIndex) && (settings.ShowHamburger && !(settings.Windowed && settings.SuppressWindowedMenu)) {
		inBurger = in
		e := backend.ProducerMain.GetInterpreter(SelectedIndex)
		if inBurger || !settings.HamburgerOnHover {
			drawHamburger(e)
		} else {
			apple2helpers.OSDPanel(e, false)
		}
	}

	idx, xpc, ypc, pxpc, pypc := WindowXYToIndex(x, y)
	settings.MouseXWindowPC, settings.MouseYWindowPC = pxpc, pypc
	////fmt.Printf("(x=%f, y=%f) -> slot #%d\n", float32(x), float32(y), idx)
	if idx != -1 && idx != lastInt {
		lastInt = idx
		fmt.Printf("Switched context to slot %d\n", idx)
		backend.ProducerMain.SetInputContext(idx)
		// camera control will follow
		SelectedCamera = idx
		SelectedIndex = idx
		ent := backend.ProducerMain.GetInterpreter(SelectedIndex)
		if !ent.IgnoreMyAudio() {
			fmt.Printf("Setting selection to %d\n", idx)
			SelectedAudioIndex = idx
			clientperipherals.SPEAKER.SelectChannel(idx)
		}
	}

	if idx != -1 {
		servicebus.InjectServiceBusMessage(
			idx,
			servicebus.MousePosition,
			&servicebus.MousePositionState{
				X:   pxpc,
				Y:   pypc,
				WX0: 0,
				WY0: 0,
				WX1: 100,
				WY1: 100,
			},
		)
	}

	// x,y are pixels in window / screen size

	if idx != -1 {

		if mouseMoveCamera {
			fmt.Println(fxcam[SelectedCamera][SelectedCameraIndex].GetZoom())
			if mouseMoveCameraAlt {
				angle := fxcam[SelectedCamera][SelectedCameraIndex].GetAngle()
				angle.SetZ(xpc*360 - 180)
				fxcam[SelectedCamera][SelectedCameraIndex].SetRotation(angle)
				fxcam[SelectedCamera][SelectedCameraIndex].Update()
				z := (ypc * 32)
				if z <= 0.1 {
					z = 0.1
				}
				fxcam[SelectedCamera][SelectedCameraIndex].SetZoom(z)
			} else {
				angle := fxcam[SelectedCamera][SelectedCameraIndex].GetAngle()
				angle.SetX(ypc*360 - 180)
				angle.SetY(xpc*360 - 180)
				fxcam[SelectedCamera][SelectedCameraIndex].SetRotation(angle)
				fxcam[SelectedCamera][SelectedCameraIndex].Update()
			}
			return
		}

		var p0, p1 int

		if settings.CurrentMouseMode == settings.MM_MOUSE_CAMERA {

			if (deltaX != 0 || deltaY != 0) && (settings.LeftButton || settings.RightButton || settings.MiddleButton) {

				switch {
				case settings.LeftButton && !settings.RightButton && !settings.MiddleButton:

					fxcam[SelectedCamera][SelectedCameraIndex].Orbit(deltaX, 0)
					fxcam[SelectedCamera][SelectedCameraIndex].Orbit(0, deltaY)

				case !settings.LeftButton && settings.RightButton && !settings.MiddleButton:

					oldZoom := fxcam[SelectedCamera][SelectedCameraIndex].GetZoom()
					newZoom := oldZoom + deltaY*0.1
					if newZoom <= 0 {
						newZoom = 0.1
					}

					fxcam[SelectedCamera][SelectedCameraIndex].SetZoom(newZoom)
					fxcam[SelectedCamera][SelectedCameraIndex].Rotate(0, 0, deltaX)

				case settings.MiddleButton || (settings.LeftButton && settings.RightButton):

					fxcam[SelectedCamera][SelectedCameraIndex].Ascend(deltaY)
					fxcam[SelectedCamera][SelectedCameraIndex].Strafe(-deltaX)

				}

			}

		} else if settings.CurrentMouseMode == settings.MM_MOUSE_JOYSTICK {

			p0 = int(xpc * 255)
			p1 = int(ypc * 255)

			//log2.Printf("MouseEvent is updating paddle positions p0=%d, p1=%d", p0, p1)

			for i := 0; i < memory.OCTALYZER_NUM_INTERPRETERS; i++ {
				if p0 != lastP0 {
					if settings.LogoCameraControl[i] {
						CameraAxis0 = p0
					}
					RAM.IntSetPaddleValue(i, 0, uint64(p0))

				}
				if p1 != lastP1 {
					if settings.LogoCameraControl[i] {
						CameraAxis1 = p1
					}
					RAM.IntSetPaddleValue(i, 1, uint64(p1))
				}
			}

		} else if settings.CurrentMouseMode == settings.MM_MOUSE_DPAD {
			// 0 - 35%, 35-65, 65-100
			xvpc := float64(0)
			yvpc := float64(0)

			if xpc < 0.35 {
				xvpc = 0
			} else if xpc > 0.65 {
				xvpc = 1
			} else {
				xvpc = (xpc - 0.35) / 0.30
			}

			if ypc < 0.35 {
				yvpc = 0
			} else if ypc > 0.65 {
				yvpc = 1
			} else {
				yvpc = (ypc - 0.35) / 0.30
			}

			p0 = int(xvpc * 255)
			p1 = int(yvpc * 255)

			for i := 0; i < memory.OCTALYZER_NUM_INTERPRETERS; i++ {
				//if p0 != lastP0 {
				if settings.LogoCameraControl[i] {
					CameraAxis0 = p0
					CameraAxis1 = p1
				}
				RAM.IntSetPaddleValue(i, 0, uint64(p0))

				//}
				//if p1 != lastP1 {
				RAM.IntSetPaddleValue(i, 1, uint64(p1))

				//}
			}

		}

		for i := 0; i < memory.OCTALYZER_NUM_INTERPRETERS; i++ {
			if RAM.IntGetSlotMenu(i) {
				RAM.IntSetMouseX(i, uint64(floor(float64(oxpc)*80)))
				RAM.IntSetMouseY(i, uint64(floor(float64(oypc)*48)))
			} else {
				RAM.IntSetMouseX(i, uint64(floor(float64(pxpc)*80)))
				RAM.IntSetMouseY(i, uint64(floor(float64(pypc)*48)))
			}
		}

		text := getActiveTextLayer(lastInt)
		//text := HUDLayers[lastInt][0]
		if !settings.DisableTextSelect[SelectedIndex] && text != nil && text.SelActive {
			text.DragSelect(int(floor(float64(pxpc)*80)), int(floor(float64(pypc)*48)))
		}

		lastmx = int(floor(float64(pxpc) * 80))
		lastmy = int(floor(float64(pypc) * 48))

		//if settings.MouseAsJoystick && !settings.MousePad {
		lastP0 = p0
		lastP1 = p1
		//}

		UpdateJoystickState()

	} else {
		if RAM.IntGetSlotMenu(SelectedIndex) {
			RAM.IntSetMouseX(SelectedIndex, uint64(floor(float64(oxpc)*80)))
			RAM.IntSetMouseY(SelectedIndex, uint64(floor(float64(oypc)*48)))
		}
	}

}

func drawHamburger(e interfaces.Interpretable) {
	// turn LED on
	apple2helpers.OSDPanel(e, true)
	txt := apple2helpers.GETHUD(e, "OOSD")
	txt.Control.FGColor = 15
	txt.Control.BGColor = 0
	txt.Control.Font = 0
	txt.Control.SetWindow(0, 0, 79, 47)
	//txt.Control.ClearScreen()
	txt.Control.FGColor = 15
	txt.Control.BGColor = 0
	txt.Control.Shade = 0
	//txt.Control.ClearScreen()
	txt.Control.GotoXY(1, 1)
	txt.Control.SetWindow(0, 0, 79, 47)
	// for y := 0; y < 3; y++ {
	// 	txt.Control.GotoXY(1, 1+y)
	// 	for x := 0; x < 3; x++ {
	// 		txt.Control.PutStr(string(rune(1028)))
	// 	}
	// 	txt.Control.PutStr("\r\n")
	// }
	hamburger := hamburgerColor
	if settings.HighContrastUI {
		hamburger = hamburgerMono
	}

	lines := strings.Split(utils.Unescape(strings.Replace(hamburger, "\n", "\r\n", -1)), "\r\n")
	for y, l := range lines {
		txt.Control.GotoXY(1, 1+y)
		txt.Control.PutStr(l)
	}

	txt.Control.HideCursor()
	txt.SetActive(true)
	txt.Control.Shade = 0
}

func getActiveTextLayer(index int) *video.TextLayer {

	if settings.BlueScreen {
		return nil
	}

	for _, l := range HUDLayers[index] {
		if l != nil && l.Spec.GetActive() {
			return l
		}
	}

	return nil

}

func OnMouseScroll(w *glumby.Window, xdiff float64, ydiff float64) {
	if settings.BlueScreen {
		return
	}

	if !settings.LogoCameraControl[SelectedIndex] {

		fmt.Printf("Mouse scroll dx=%f, dy=%f\n", xdiff, ydiff)
		if ydiff < 0 {
			RAM.KeyBufferAdd(SelectedIndex, vduconst.CSR_DOWN)
		} else if ydiff > 0 {
			RAM.KeyBufferAdd(SelectedIndex, vduconst.CSR_UP)
		}

	}

	if settings.LogoCameraControl[SelectedIndex] {
		if ydiff > 0 {
			fxcam[SelectedCamera][SelectedCameraIndex].SetZoom(fxcam[SelectedCamera][SelectedCameraIndex].GetZoom() * 1.1)
		} else {
			fxcam[SelectedCamera][SelectedCameraIndex].SetZoom(fxcam[SelectedCamera][SelectedCameraIndex].GetZoom() / 1.1)
		}
	}
}

func OnMouseButtonEvent(w *glumby.Window, button glumby.MouseButton, action glumby.Action, mod glumby.ModifierKey) {
	//	log.Printf("Mouse button %d, action = %d, mod = %d\n", button, action, mod)

	// notify mouse button event
	servicebus.InjectServiceBusMessage(
		SelectedIndex,
		servicebus.MouseButton,
		&servicebus.MouseButtonState{
			Index:   int(button),
			Pressed: action == glumby.Press,
		},
	)

	switch button {
	case glumby.MouseButtonLeft:
		settings.LeftButton = (action == glumby.Press)
	case glumby.MouseButtonRight:
		settings.RightButton = (action == glumby.Press)
	case glumby.MouseButtonMiddle:
		settings.MiddleButton = (action == glumby.Press)
	}

	//log2.Printf("left=%v, right=%v, middle=%v", settings.LeftButton, settings.RightButton, settings.MiddleButton)

	if settings.BlueScreen || backend.ProducerMain == nil {
		return
	}

	if button == glumby.MouseButtonLeft && inBurger && !RAM.IntGetSlotMenu(SelectedIndex) && settings.ShowHamburger {
		RAM.IntSetSlotMenu(SelectedIndex, true)
		return
	}

	// if settings.GetMouseMode() == settings.MM_MOUSE_OFF {
	// 	return
	// }

	if button == glumby.MouseButtonMiddle && !RAM.IntGetSlotMenu(SelectedIndex) && settings.CurrentMouseMode != settings.MM_MOUSE_CAMERA {
		RAM.IntSetSlotMenu(SelectedIndex, true)
		return
	}

	if settings.BlueScreen {
		return
	}

	if settings.GetMouseMode() != settings.MM_MOUSE_OFF || RAM.IntGetSlotMenu(SelectedIndex) {

		text := getActiveTextLayer(lastInt)

		for index := 0; index < memory.OCTALYZER_NUM_INTERPRETERS; index++ {
			if button == glumby.MouseButtonLeft {
				switch action {
				case glumby.Press:
					if settings.CurrentMouseMode != settings.MM_MOUSE_OFF {
						if !settings.DisableJoystick[SelectedIndex] {
							RAM.IntSetPaddleButton(index, 0, 1)
						}
						if settings.LogoCameraControl[SelectedIndex] {
							CameraButton0 = true
						}
					}
					RAM.IntSetMouseButtons(index, true, false)

					if text != nil {

						bounds := text.Spec.GetBoundsRect()

						if bounds.Contains(uint16(lastmx), uint16(lastmy)) {
							text.StartSelect(lastmx, lastmy)
						}
					}

				case glumby.Repeat:
					if settings.CurrentMouseMode != settings.MM_MOUSE_OFF {
						if !settings.DisableJoystick[SelectedIndex] {
							RAM.IntSetPaddleButton(index, 0, 1)
						}
						if settings.LogoCameraControl[SelectedIndex] {
							CameraButton0 = true
						}
					}
					RAM.IntSetMouseButtons(index, true, false)
				case glumby.Release:
					if settings.CurrentMouseMode != settings.MM_MOUSE_OFF {
						if !settings.DisableJoystick[SelectedIndex] {
							RAM.IntSetPaddleButton(index, 0, 0)
						}
						if settings.LogoCameraControl[SelectedIndex] {
							CameraButton0 = false
						}
					}
					RAM.IntSetMouseButtons(index, false, false)
					if text != nil && text.SelActive {

						text.DoneSelect()

					}
				}

				UpdateJoystickState()
			}

			if button == glumby.MouseButtonRight {
				switch action {
				case glumby.Press:
					if !settings.DisableJoystick[SelectedIndex] {
						RAM.IntSetPaddleButton(index, 1, 1)
					}
					if settings.LogoCameraControl[SelectedIndex] {
						CameraButton1 = true
					}
					RAM.IntSetMouseButtons(index, false, true)
				case glumby.Repeat:
					if !settings.DisableJoystick[SelectedIndex] {
						RAM.IntSetPaddleButton(index, 1, 1)
					}
					if settings.LogoCameraControl[SelectedIndex] {
						CameraButton1 = true
					}
					RAM.IntSetMouseButtons(index, false, true)
				case glumby.Release:
					if !settings.DisableJoystick[SelectedIndex] {
						RAM.IntSetPaddleButton(index, 1, 0)
					}
					if settings.LogoCameraControl[SelectedIndex] {
						CameraButton1 = false
					}
					RAM.IntSetMouseButtons(index, false, false)
				}
			}
		}

	}
}

func UpdateJoystickState() {

	if settings.DisableJoystick[SelectedIndex] {
		return
	}

	var state servicebus.JoyLine

	x := RAM.IntGetPaddleValue(SelectedIndex, 0)
	y := RAM.IntGetPaddleValue(SelectedIndex, 1)
	btn := RAM.IntGetPaddleButton(SelectedIndex, 0) != 0

	if x >= 160 {
		state |= servicebus.JoystickRight
	} else if x < 96 {
		state |= servicebus.JoystickLeft
	}

	if y >= 160 {
		state |= servicebus.JoystickDown
	} else if y < 96 {
		state |= servicebus.JoystickUp
	}

	if btn {
		state |= servicebus.JoystickButton0
	}

	servicebus.SendServiceBusMessage(
		SelectedIndex,
		servicebus.JoyEvent,
		&servicebus.JoystickEventData{
			Stick: 0,
			Line:  state,
		},
	)

}

func UpdatePixelSize(w *glumby.Window) {
	_, h := w.GetGLFWWindow().GetFramebufferSize()
	px := float32(h) / 192
	video.HGRPixelSize = px
}

func OnScrollEvent(w *glumby.Window, xoff, yoff float64) {

	distance_z += float32(yoff)

	for index := 0; index < memory.OCTALYZER_NUM_INTERPRETERS; index++ {
		mv := int(yoff * 4)
		v := int(RAM.IntGetPaddleValue(index, 0)) + mv
		if v < 0 {
			v = 0
		}
		if v > 255 {
			v = 255
		}

		if !settings.DisableJoystick[index] {
			RAM.IntSetPaddleValue(index, 0, uint64(v))
		}
	}
}

func OnResizeEvent(w *glumby.Window, width, height int) {

	settings.FrameSkip = settings.DefaultFrameSkip

	calcPlayfield(w)
	calcPlayfieldOSD(w)

	//log.Printf("Resize: viewport = (%f, %f, %f, %f)\n", playfield.X0, playfield.Y0, playfield.X1, playfield.Y1)

	UpdatePixelSize(w)
}

func SendPaddleButton(pdl int, v uint64) {
	for i := 0; i < memory.OCTALYZER_NUM_INTERPRETERS; i++ {
		if !settings.DisableJoystick[i] {
			RAM.IntSetPaddleButton(i, pdl, v)
		}
	}
}

func SendPaddleModify(pdl int, mv int) {
	for i := 0; i < memory.OCTALYZER_NUM_INTERPRETERS; i++ {
		v := int(RAM.IntGetPaddleValue(i, pdl)) + mv
		if v < 0 {
			v = 0
		}
		if v > 255 {
			v = 255
		}

		if !settings.DisableJoystick[i] {
			RAM.IntSetPaddleValue(i, pdl, uint64(v))
		}
	}
}

var paddlekeys bool = false

func SendPaddleValue(pdl int, f float32) {
	for i := 0; i < memory.OCTALYZER_NUM_INTERPRETERS; i++ {
		z := uint64(127.5 + 127.5*f)

		if settings.LogoCameraControl[i] {
			switch pdl {
			case 0:
				CameraAxis0 = int(z)
			case 1:
				CameraAxis1 = int(z)
			}
		}

		if !settings.DisableJoystick[i] {
			RAM.IntSetPaddleValue(i, pdl, z)
			if paddlekeys {
				switch {
				case pdl == 1 && z < 5:
					//OnCharEvent( w, 'A' )
					RAM.WriteInterpreterMemory(0, 49152, 128|65)
				case pdl == 1 && z >= 250:
					RAM.WriteInterpreterMemory(0, 49152, 128|90)
				case pdl == 0 && z < 5:
					RAM.WriteInterpreterMemory(0, 49152, 128|8)
				case pdl == 0 && z >= 250:
					RAM.WriteInterpreterMemory(0, 49152, 128|21)
				}
			}
		}
	}
}

var CameraButton0 bool
var CameraButton1 bool
var CameraAxis0 int
var CameraAxis1 int

func UpdateCameraControlState() {
	if !settings.LogoCameraControl[SelectedIndex] {
		return
	}
	// updates based states
	if CameraButton0 {

		if CameraAxis0 < 64 {
			// left
			amt := ((64 - float64(CameraAxis0)) / 64) * 5
			fxcam[SelectedIndex][SelectedCameraIndex].Strafe(-amt)
		} else if CameraAxis0 > 192 {
			// right
			amt := ((float64(CameraAxis0) - 192) / 64) * 5
			fxcam[SelectedIndex][SelectedCameraIndex].Strafe(amt)
		}
		if CameraAxis1 < 64 {
			// up
			amt := ((64 - float64(CameraAxis1)) / 64) * 5
			fxcam[SelectedIndex][SelectedCameraIndex].Ascend(amt)
		} else if CameraAxis1 > 192 {
			// down
			amt := ((float64(CameraAxis1) - 192) / 64) * 5
			fxcam[SelectedIndex][SelectedCameraIndex].Ascend(-amt)
		}

	} else if CameraButton1 {

		if CameraAxis0 < 64 {
			// left
			amt := ((64 - float64(CameraAxis0)) / 64) * 5
			fxcam[SelectedCamera][SelectedCameraIndex].Rotate3DZ(amt)
		} else if CameraAxis0 > 192 {
			// right
			amt := ((float64(CameraAxis0) - 192) / 64) * 5
			fxcam[SelectedCamera][SelectedCameraIndex].Rotate3DZ(-amt)
		}
		if CameraAxis1 < 64 {
			// up
			fxcam[SelectedCamera][SelectedCameraIndex].SetZoom(fxcam[SelectedCamera][SelectedCameraIndex].GetZoom() * 1.1)
		} else if CameraAxis1 > 192 {
			// down
			fxcam[SelectedCamera][SelectedCameraIndex].SetZoom(fxcam[SelectedCamera][SelectedCameraIndex].GetZoom() / 1.1)
		}

	} else {
		if CameraAxis0 < 64 {
			// left
			amt := ((64 - float64(CameraAxis0)) / 64) * 5
			fxcam[SelectedIndex][SelectedCameraIndex].Orbit(-amt, 0)
		} else if CameraAxis0 > 192 {
			// right
			amt := ((float64(CameraAxis0) - 192) / 64) * 5
			fxcam[SelectedIndex][SelectedCameraIndex].Orbit(amt, 0)
		}
		if CameraAxis1 < 64 {
			// up
			amt := ((64 - float64(CameraAxis1)) / 64) * 5
			fxcam[SelectedIndex][SelectedCameraIndex].Orbit(0, -amt)
		} else if CameraAxis1 > 192 {
			// down
			amt := ((float64(CameraAxis1) - 192) / 64) * 5
			fxcam[SelectedIndex][SelectedCameraIndex].Orbit(0, amt)
		}
	}
}

func processControllerEvents() {
	for _, con := range w.Controllers {
		evlist := con.GetEvents()
		for _, ev := range evlist {
			a, v := clientperipherals.GetActionForEvent(ev)
			switch a {
			case clientperipherals.CaGameAccept:
				RAM.WriteInterpreterMemory(0, 49152, 128|13)
			case clientperipherals.CaGameBack:
				RAM.WriteInterpreterMemory(0, 49152, 128|27)
			case clientperipherals.CaPaddleButtonPress0:
				//OnKeyEvent(w, vduconst.DPAD_B, 0, glumby.Press)
				SendPaddleButton(0, 1)
			case clientperipherals.CaPaddleButtonRelease0:
				SendPaddleButton(0, 0)
			case clientperipherals.CaPaddleButtonPress1:
				//OnKeyEvent(w, vduconst.DPAD_A, 0, glumby.Press)
				SendPaddleButton(1, 1)
			case clientperipherals.CaPaddleButtonRelease1:
				SendPaddleButton(1, 0)
			case clientperipherals.CaPaddleButtonPress2:
				SendPaddleButton(2, 1)
			case clientperipherals.CaPaddleButtonRelease2:
				SendPaddleButton(2, 0)
			case clientperipherals.CaPaddleButtonPress3:
				SendPaddleButton(3, 1)
			case clientperipherals.CaPaddleButtonRelease3:
				SendPaddleButton(4, 0)
			case clientperipherals.CaPaddleDecrease0:
				SendPaddleModify(0, -8)
			case clientperipherals.CaPaddleIncrease0:
				SendPaddleModify(0, 8)
			case clientperipherals.CaPaddleDecrease1:
				SendPaddleModify(1, -8)
			case clientperipherals.CaPaddleIncrease1:
				SendPaddleModify(1, 8)
			case clientperipherals.CaPaddleDecrease2:
				SendPaddleModify(2, -8)
			case clientperipherals.CaPaddleIncrease2:
				SendPaddleModify(2, 8)
			case clientperipherals.CaPaddleDecrease3:
				SendPaddleModify(3, -8)
			case clientperipherals.CaPaddleIncrease3:
				SendPaddleModify(3, 8)
			case clientperipherals.CaPaddleModValue0:
				// if v <= -1 {
				// 	OnKeyEvent(w, vduconst.DPAD_LEFT, 0, glumby.Press)
				// } else if v >= 1 {
				// 	OnKeyEvent(w, vduconst.DPAD_RIGHT, 0, glumby.Press)
				// }
				SendPaddleValue(0, v)
			case clientperipherals.CaPaddleModValue1:
				// if v <= -1 {
				// 	OnKeyEvent(w, vduconst.DPAD_UP_PRESS, 0, glumby.Press)
				// } else if v >= 1 {
				// 	OnKeyEvent(w, vduconst.DPAD_DOWN_PRESS, 0, glumby.Press)
				// } else {
				// 	OnKeyEvent(w, vduconst.DPAD_UP_RELEASE, 0, glumby.Press)
				// 	OnKeyEvent(w, vduconst.DPAD_DOWN_RELEASE, 0, glumby.Press)
				// }
				SendPaddleValue(1, v)
			case clientperipherals.CaPaddleModValue2:
				SendPaddleValue(2, v)
			case clientperipherals.CaPaddleModValue3:
				SendPaddleValue(3, v)
				// case clientperipherals.CaGameSelect:
				// 	if ev.Pressed {
				// 		OnKeyEvent(w, vduconst.DPAD_SELECT, 0, glumby.Press)
				// 	} else {
				// 		OnKeyEvent(w, vduconst.DPAD_SELECT, 0, glumby.Release)
				// 	}
			}
		}
		if len(evlist) > 0 {
			UpdateJoystickState()
		}
	}
}

// WindowXYToIndex: Based on an XY co-ordinate, determine where the heck the point
// is in relation to an interpreter
func WindowXYToIndex(x, y float64) (int, float64, float64, float64, float64) {

	for i := 0; i < memory.OCTALYZER_NUM_INTERPRETERS; i++ {

		active := (RAM.IntGetActiveState(i) != 0)

		if active && x >= pcam[i].PixelViewPort.X0 && x <= pcam[i].PixelViewPort.X1 && y >= pcam[i].PixelViewPort.Y0 && y <= pcam[i].PixelViewPort.Y1 {

			x1 := float64(pcam[i].PixelViewPort.X1) - PADDLE_BORDER
			y1 := float64(pcam[i].PixelViewPort.Y1) - PADDLE_BORDER
			x0 := float64(pcam[i].PixelViewPort.X0) + PADDLE_BORDER
			y0 := float64(pcam[i].PixelViewPort.Y0) + PADDLE_BORDER

			xpc := (x - x0) / (x1 - x0)
			ypc := (y - y0) / (y1 - y0)

			if xpc < 0 {
				xpc = 0
			}
			if xpc > 1 {
				xpc = 1
			}
			if ypc < 0 {
				ypc = 0
			}
			if ypc > 1 {
				ypc = 1
			}

			px1 := float64(pcam[i].PixelViewPort.X1)
			py1 := float64(pcam[i].PixelViewPort.Y1)
			px0 := float64(pcam[i].PixelViewPort.X0)
			py0 := float64(pcam[i].PixelViewPort.Y0)

			pxpc := (x - px0) / (px1 - px0)
			pypc := (y - py0) / (py1 - py0)

			if pxpc < 0 {
				pxpc = 0
			}
			if pxpc > 1 {
				pxpc = 1
			}
			if pypc < 0 {
				pypc = 0
			}
			if pypc > 1 {
				pypc = 1
			}

			return i, xpc, ypc, pxpc, pypc

		}
	}
	return -1, 0.5, 0.5, 0.5, 0.5

}

// WindowXYToIndex: Based on an XY co-ordinate, determine where the heck the point
// is in relation to an interpreter
func WindowXYToIndexOSD(i int, x, y float64) (float64, float64) {

	x1 := float64(playfieldOSD.X1)
	y1 := float64(playfieldOSD.Y1)
	x0 := float64(playfieldOSD.X0)
	y0 := float64(playfieldOSD.Y0)

	xpc := (x - x0) / (x1 - x0)
	ypc := (y - y0) / (y1 - y0)

	if xpc < 0 {
		xpc = 0
	}
	if xpc > 1 {
		xpc = 1
	}
	if ypc < 0 {
		ypc = 0
	}
	if ypc > 1 {
		ypc = 1
	}

	return xpc, ypc

}
