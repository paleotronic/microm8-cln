package settings

import (
	"bytes"
	"image"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"paleotronic.com/octalyzer/video/font"
)

type MouseMode int

const (
	MM_MOUSE_JOYSTICK MouseMode = iota
	MM_MOUSE_DPAD
	MM_MOUSE_GEOS
	MM_MOUSE_DDRAW
	MM_MOUSE_CAMERA
	MM_MOUSE_OFF
)

func (mm MouseMode) String() string {
	switch mm {
	case MM_MOUSE_JOYSTICK:
		return "Joystick"
	case MM_MOUSE_DPAD:
		return "D-Pad"
	case MM_MOUSE_OFF:
		return "Off"
	case MM_MOUSE_GEOS:
		return "GEOS stick"
	case MM_MOUSE_CAMERA:
		return "Camera"
	case MM_MOUSE_DDRAW:
		return "Dazzle stick"
	}
	return "Unknown"
}

var mousemodecallback func(m MouseMode)

func SetMouseModeCallback(f func(m MouseMode)) {
	mousemodecallback = f
}

func CheckMouseMode(v MouseMode) {
	if mousemodecallback != nil {
		mousemodecallback(v)
	}
}

func SetMouseMode(v MouseMode) {
	CurrentMouseMode = v
	switch v {
	case MM_MOUSE_DDRAW:
		StickCenter = [2]float64{145, 150}
		MaxStickLevel[StickLEFT] = 80
		MaxStickLevel[StickRIGHT] = 96
		MaxStickLevel[StickUP] = 96
		MaxStickLevel[StickDOWN] = 96
	case MM_MOUSE_GEOS:
		StickCenter = [2]float64{127, 127}
		MaxStickLevel[StickLEFT] = 80
		MaxStickLevel[StickRIGHT] = 96
		MaxStickLevel[StickUP] = 80
		MaxStickLevel[StickDOWN] = 48
	}
	CheckMouseMode(v)
}

func GetMouseMode() MouseMode {
	return CurrentMouseMode
}

type VideoPaletteTint uint64

const (
	VPT_NONE VideoPaletteTint = iota
	VPT_GREY
	VPT_GREEN
	VPT_AMBER
	// max
	VPT_MAX
)

func (v VideoPaletteTint) String() string {
	switch v {
	case VPT_NONE:
		return "Full Colour"
	case VPT_GREY:
		return "Black and White"
	case VPT_GREEN:
		return "Green"
	case VPT_AMBER:
		return "Amber"
	}
	return "Unknown"
}

type VoxelDepth int

const (
	VXD_STANDARD VoxelDepth = iota
	VXD_2_TIMES
	VXD_3_TIMES
	VXD_4_TIMES
	VXD_5_TIMES
	VXD_6_TIMES
	VXD_7_TIMES
	VXD_8_TIMES
	VXD_9_TIMES
	VXD_MAX
)

func (v VoxelDepth) String() string {
	switch v {
	case VXD_STANDARD:
		return "1x"
	case VXD_2_TIMES:
		return "2x"
	case VXD_3_TIMES:
		return "3x"
	case VXD_4_TIMES:
		return "4x"
	case VXD_5_TIMES:
		return "5x"
	case VXD_6_TIMES:
		return "6x"
	case VXD_7_TIMES:
		return "7x"
	case VXD_8_TIMES:
		return "8x"
	case VXD_9_TIMES:
		return "9x"
	}
	return "Unknown"
}

type DHGRHighBitMode int

const (
	DHB_MIXED_AUTO DHGRHighBitMode = iota
	DHB_MIXED_ON
	DHB_MIXED_OFF
)

type VideoMode int

const (
	VM_DOTTY VideoMode = iota
	VM_VOXELS
	VM_FLAT
	//VM_LAYERS
	VM_MONO_DOTTY
	VM_MONO_VOXELS
	VM_MONO_FLAT
	// max mode
	VM_MAX_MODE
)

func (v VideoMode) String() string {
	switch v {
	case VM_DOTTY:
		return "Color Dot Mode"
	case VM_VOXELS:
		return "Color Voxels"
	case VM_FLAT:
		return "Color Raster"
	case VM_MONO_DOTTY:
		return "Mono Dot Mode"
	case VM_MONO_VOXELS:
		return "Mono Voxels"
	case VM_MONO_FLAT:
		return "Mono Raster"
	}
	return "Unknown"
}

type VMLauncherConfig struct {
	WorkingDir  string
	Disks       []string
	Pakfile     string
	SmartPort   string
	RunFile     string
	RunCommand  string
	Dialect     string
	ZXState     string
	ForceSplash bool
}

type RState struct {
	Command   string
	Dialect   string
	WorkDir   string
	IsControl bool
}

func GetRewindSpeeds() []float64 {
	if DisableFractionalRewindSpeeds {
		return RewindSpeedsNonFractional
	}
	return RewindSpeedsAll
}

var fSetTitle func(s string)

func SetSubtitle(s string) {
	// if fSetTitle != nil {
	// 	fSetTitle(s)
	// }
}

func SetTitleFunc(f func(s string)) {
	fSetTitle = f
}

func SetGlobalOverrides(slot int, o map[string]interface{}) {
	if GlobalOverrides[slot] == nil {
		GlobalOverrides[slot] = map[string]interface{}{}
	}
	for k, v := range o {
		GlobalOverrides[slot][k] = v
	}
}

func GetStringOverride(slot int, key string) string {
	v, ok := GlobalOverrides[slot][key]
	if !ok {
		return ""
	}
	return v.(string)
}

func IsSetBoolOverride(slot int, key string) bool {
	v, ok := GlobalOverrides[slot][key]
	if !ok {
		return false
	}
	return v.(bool)
}

var DefaultSHRDitherPalette = [16]int{
	0x000, // black
	0x00b, // blue
	0xbb0, // yellow
	0xddd, // grey
	0xb00, // red
	0xb0b, // purple
	0xb80, // orange
	0xf88, // ltred
	0x0b0, // green
	0x0aa, // aqua
	0x8a0, // lime
	0x8f8, // ltgrn
	0xddd, // grey2
	0x88f, // ltblue
	0xff8, // ltyellow
	0xfff, // white
}

var DefaultSHR320Palette = [16]int{
	0x000,
	0x777,
	0x841,
	0x82d,
	0x00f,
	0x090,
	0xf70,
	0xe00,
	0xfba,
	0xff0,
	0x1f0,
	0x4ef,
	0xebf,
	0x78f,
	0xccc,
	0xfff,
}

const StickUP = 0
const StickDOWN = 1
const StickLEFT = 2
const StickRIGHT = 3

type SpeakerRedirect struct {
	VM      int
	Channel int
}

type HeatMapMode int

const (
	HMOff HeatMapMode = iota
	HMMain
	HMAux
	HMMainBank
	HMAuxBank
	HMExecCombined
)

var PDFSpool = false
var ParallelLinePrinter = "/dev/usb/lp1"
var ParallelPassThrough = false
var DemoModeEnabled bool
var CommandLineTelnetAddress = "localhost:5555"
var UserWarpOverride [NUMSLOTS]bool
var RecordC020 [NUMSLOTS]bool
var RecordC020Rate [NUMSLOTS]int
var RecordC020Buffer [NUMSLOTS][]float32
var SHRFrameForce [NUMSLOTS]bool
var ArrowKeyPaddles bool
var UseDHGRForHGR [NUMSLOTS]bool
var UseVerticalBlend [NUMSLOTS]bool

var LogoCameraControl [NUMSLOTS]bool // enable logo camera control
var DisableJoystick [NUMSLOTS]bool   // disable joystick control

var RotationOrder mgl32.RotationOrder
var ImageDrawRect [NUMSLOTS]*image.Rectangle
var HeatMap [NUMSLOTS]HeatMapMode
var HeatMapBank [NUMSLOTS]int
var HeatMapCPU [NUMSLOTS]bool
var DebugFullCPURecord bool
var FileFullCPURecord bool
var DiskRecordStart bool

type SSCMode int

const (
	SSCModeVirtualModem        SSCMode = iota
	SSCModeTelnetServer        SSCMode = iota
	SSCModeEmulatedImageWriter SSCMode = iota
	SSCModeEmulatedESCP        SSCMode = iota
	SSCModeSerialRaw           SSCMode = iota
)

var AutosaveFilename [NUMSLOTS]string
var SnapshotFile [NUMSLOTS]string
var SpectrumVSync int = 37264
var PreventSuppressAlt [NUMSLOTS]bool
var DisableTextSelect [NUMSLOTS]bool
var ForcePureBoot [NUMSLOTS]bool
var DiskIIUse13Sectors [NUMSLOTS]bool
var SSCHardwarePort = ""
var SSCDipSwitch1 = 0xe9
var SSCDipSwitch2 = 0x04
var DefModemInitString = ""
var SuppressATIResponse [NUMSLOTS]bool
var DisableOverlays bool
var SpeakerRedirects [NUMSLOTS]*SpeakerRedirect
var LeftButton, RightButton, MiddleButton bool
var TakeScreenshot bool
var DisableScanlines bool
var LastDisableScanlines bool
var SuppressWindowedMenu bool
var HighContrastUI bool = false
var LogoFastDraw [NUMSLOTS]bool
var LogoSuppressDefines [NUMSLOTS]bool
var GlobalOverrides [NUMSLOTS]map[string]interface{}
var SSCCardMode [NUMSLOTS]SSCMode
var BlockCSR [NUMSLOTS]bool
var VMLaunch [NUMSLOTS]*VMLauncherConfig
var OptimalBitTiming = 32
var UseFullScreen bool
var CleanBootRequested bool
var MicroTrackerEnabled [NUMSLOTS]bool
var Debug6522 bool
var CurrentMouseMode MouseMode = MM_MOUSE_JOYSTICK
var StickCenter [2]float64 = [2]float64{127, 127}
var MaxStickLevel [4]float64 = [4]float64{64, 64, 64, 64}
var MouseStickInterval = 7 * time.Millisecond
var MouseXWindowPC, MouseYWindowPC float64
var DisableFractionalRewindSpeeds bool
var RewindSpeedsAll = []float64{0.25, 0.5, 1, 2, 4}
var RewindSpeedsNonFractional = []float64{1, 2, 4}
var CPURecordTimingPoints int = 10
var PrintToPDFTimeoutSec int = 15
var DebuggerActiveSlot int = -1
var PreserveDSK bool = true
var IgnoreSoftSwitches [NUMSLOTS]bool
var ScreenShotNeeded bool
var ScreenShotJPEGData []byte
var BinaryPath = GetBinaryPath()
var DebuggerPort = 9502
var DebuggerOn = true
var DebuggerAttachSlot = 0
var SkipCameraOnSave bool
var IsPakBoot bool
var Pakfile [NUMSLOTS]string
var FirstBoot [NUMSLOTS]bool
var MixerVolume = 0.5
var SpeakerVolume [NUMSLOTS]float64
var MockingBoardPSG0Bal = -0.5
var MockingBoardPSG1Bal = 0.5
var TemporaryMute bool
var WindowApplied bool
var Verbose bool
var NoUpdates bool
var MenuActive bool
var SampleRate int = 48000
var TRACENET bool = false
var IsRemInt bool = false
var RecordIgnoreIO [NUMSLOTS]bool
var RecordIgnoreAudio [NUMSLOTS]bool
var NoDiskWarp [NUMSLOTS]bool
var PBState [NUMSLOTS]bool
var ResetState [NUMSLOTS]*RState
var PureBootSmartVolume [NUMSLOTS]string
var PureBootVolume [NUMSLOTS]string
var PureBootVolumeWP [NUMSLOTS]bool
var PureBootVolume2 [NUMSLOTS]string
var PureBootVolumeWP2 [NUMSLOTS]bool
var PureBootRestoreState [NUMSLOTS]string
var PureBootRestoreStateBin [NUMSLOTS][]byte
var SlotZPEmu [NUMSLOTS]bool
var CPUModel [NUMSLOTS]string
var CPUClock [NUMSLOTS]int
var PureBootBanks []string
var PureBootDebugCommand string = "G"
var AudioUsesLeapTicks = false

// var UseHQAudio bool = true
var BootMenuMS time.Duration = 5000 * time.Millisecond
var UseBootMenu bool = true
var DisableMetaMode [NUMSLOTS]bool
var SpeakerBitstreamDiv int = 8
var SpeakerBitstreamPsuedoLevels int = 32
var RealTimeBasicMode bool = false
var TrackerMode bool = false

// var DHGRMono [NUMSLOTS]bool
var DONTLOG bool = true
var DONTLOGDEFAULT bool

// var MouseGEOS bool = false
// var MouseAsJoystick bool = true
var FrameSkip int = 0
var DefaultFrameSkip int = 0
var FPSClock int = 60
var FSVoteThreshold int = 80
var FSVoteUp int = 10
var FSVoteDown int = 1
var LocalBoot bool
var Offline bool
var Windowed bool
var Args []string
var EBOOT bool = false

// var MousePad bool = false
var HamburgerOnHover = false
var JoystickReverseX [NUMSLOTS]bool
var JoystickReverseY [NUMSLOTS]bool
var BlueScreen bool = true
var HelpBase string = "/boot/help"
var SplashDisk string
var SplashDisk2 string
var AllowPerspectiveChanges = true
var MusicTrack [NUMSLOTS]string
var MusicFadein [NUMSLOTS]int
var MusicLeadin [NUMSLOTS]int
var BackdropFile [NUMSLOTS]string
var BackdropZoom [NUMSLOTS]float64
var BackdropOpacity [NUMSLOTS]float64
var BackdropZoomRatio [NUMSLOTS]float64
var BackdropTrack [NUMSLOTS]bool
var AudioPacketReverse [NUMSLOTS]bool
var VideoPlayFrames [NUMSLOTS][]*bytes.Buffer
var VideoRecordFrames [NUMSLOTS][]*bytes.Buffer
var VideoPlayBackwards [NUMSLOTS]bool
var VideoBackSeekMS [NUMSLOTS]int
var VideoRecordFile [NUMSLOTS]string
var VideoPlaybackFile [NUMSLOTS]string
var VideoPlaybackPauseOnFUL [NUMSLOTS]bool
var SlotRestartContinueCPU [NUMSLOTS]bool
var VideoSuspended bool
var BPSBufferSize int = 384
var RAWBufferSize int = 4096
var BPSBufferSizeDefault int = 384
var BPSBufferIncrement int = 64
var LateAudio time.Duration
var autoLiveRecording bool = false
var userCanChangeSpeed bool = true
var speedMutex sync.Mutex
var MuteCPU bool = false
var ServerPort string = ":6581"
var LastCubePlotTime [NUMSLOTS]time.Time
var LastRenderModeGR [NUMSLOTS]VideoMode
var LastRenderModeHGR [NUMSLOTS]VideoMode
var LastRenderModeDHGR [NUMSLOTS]VideoMode
var LastRenderModeSHR [NUMSLOTS]VideoMode
var LastRenderModeSpectrum [NUMSLOTS]VideoMode
var LastTintMode [NUMSLOTS]VideoPaletteTint
var LastVoxelDepth [NUMSLOTS]VoxelDepth
var DHGRHighBit [NUMSLOTS]DHGRHighBitMode
var LastDHGRHighBit [NUMSLOTS]DHGRHighBitMode
var DHGRMode3Detected [NUMSLOTS]bool     //
var LastDHGRMode3Detected [NUMSLOTS]bool //
var RecordMix string
var PasteCPS = 10
var PasteWarp = false
var Pasting [NUMSLOTS]bool
var FlushPDF [NUMSLOTS]bool
var LaunchQuitCPUExit bool
var DefaultLineWidth = float32(2)
var LineWidth [NUMSLOTS]float32
var UnifiedRenderGlobal bool
var UnifiedRender [NUMSLOTS]bool
var UnifiedRenderFrame [NUMSLOTS]*image.RGBA // 560x192
var UnifiedRenderChanged [NUMSLOTS]bool

var DefaultRenderModeGR VideoMode = VM_VOXELS
var DefaultRenderModeHGR VideoMode = VM_FLAT
var DefaultRenderModeDHGR VideoMode = VM_FLAT
var DefaultRenderModeSHR VideoMode = VM_FLAT
var DefaultRenderModeSpectrum VideoMode = VM_FLAT
var DefaultTintMode VideoPaletteTint = VPT_NONE
var ShowHamburger = true

var MicroPakPath string

var brightness float32 = 1.5
var LastScanLineIntensity float32 = 0.88
var ScanLineIntensity float32 = 0.88

var HasCPUBreak [NUMSLOTS]bool
var Font [NUMSLOTS]string
var DefaultFont [NUMSLOTS]*font.DecalFont
var lastFont [NUMSLOTS]string
var SpecFile [NUMSLOTS]string
var SpecName [NUMSLOTS]string
var AuxFonts [NUMSLOTS][]string
var SystemID [NUMSLOTS]string
var ForceTextVideoRefresh bool
var MemLocks [NUMSLOTS]map[int]uint64
var VBLock [NUMSLOTS]sync.Mutex

//var PaletteTint [NUMSLOTS]VideoPaletteTint
//var RenderModeHGR [NUMSLOTS]VideoMode
//var RenderModeDHGR [NUMSLOTS]VideoMode

var MonitorActive [NUMSLOTS]bool
var MonitorKeyBuffer [NUMSLOTS][16]rune
var MonitorKeyCount [NUMSLOTS]int
var monitorKeyMutex sync.Mutex

func MonitorKeyAdd(slotid int, key rune) bool {
	monitorKeyMutex.Lock()
	defer monitorKeyMutex.Unlock()
	if MonitorKeyCount[slotid] >= 16 {
		return false
	}
	MonitorKeyBuffer[slotid][MonitorKeyCount[slotid]] = key
	MonitorKeyCount[slotid]++
	return true
}

func MonitorKeyGet(slotid int) (rune, bool) {
	monitorKeyMutex.Lock()
	defer monitorKeyMutex.Unlock()
	if MonitorKeyCount[slotid] == 0 {
		return 0, false
	}
	var key = MonitorKeyBuffer[slotid][0]
	for i := 0; i < 15; i++ {
		MonitorKeyBuffer[slotid][i] = MonitorKeyBuffer[slotid][i+1]
	}
	MonitorKeyCount[slotid]--
	return key, true
}

func MonitorKeyClear(slotid int) {
	monitorKeyMutex.Lock()
	defer monitorKeyMutex.Unlock()
	MonitorKeyCount[slotid] = 0
}

func MonitorHasKey(slotid int) bool {
	monitorKeyMutex.Lock()
	defer monitorKeyMutex.Unlock()
	return MonitorKeyCount[slotid] > 0
}

var DefaultProfile string = "apple2e-en.yaml"

func AutoLiveRecording() bool {
	return autoLiveRecording
}

func SetAutoLiveRecording(b bool) {
	autoLiveRecording = b
}

func PureBoot(slotid int) bool {
	return PBState[slotid] || ForcePureBoot[slotid]
}

func PureBootCheck(slotid int) {
	PBState[slotid] = (PureBootVolume[slotid] != "") || (PureBootRestoreState[slotid] != "") || (PureBootSmartVolume[slotid] != "")
}

func SetPureBoot(slotid int, b bool) {
	PBState[slotid] = b
}

func SetUserCanChangeSpeed(b bool) {
	speedMutex.Lock()
	defer speedMutex.Unlock()
	userCanChangeSpeed = b
}

func CanUserChangeSpeed() bool {
	return userCanChangeSpeed
}

func init() {
	for i, _ := range FirstBoot {
		FirstBoot[i] = true
		// if i == 3 {
		// 	SpecFile[i] = "bbcb.yaml"
		// } else {
		SpecFile[i] = "apple2e-en.yaml"
		// }
		AuxFonts[i] = make([]string, 0)
		SpeakerVolume[i] = 0.5
	}
	for i, _ := range DHGRHighBit {
		DHGRHighBit[i] = DHB_MIXED_AUTO
		LastDHGRHighBit[i] = DHB_MIXED_OFF
	}
}

func GetBinaryFile() string {
	w := os.Args[0]
	dir, err := filepath.Abs(w)
	if err == nil {
		return dir
	}
	return w
}

func GetBinaryPath() string {
	w := os.Args[0]
	dir, err := filepath.Abs(w)
	if err == nil {
		return filepath.Dir(dir)
	}
	return filepath.Dir(w)
}

func GetModemInitString(slot int) string {
	return DefModemInitString
}

func GetLineWidth(vm int) float32 {
	if LineWidth[vm] != 0 {
		return LineWidth[vm]
	}
	return DefaultLineWidth
}
