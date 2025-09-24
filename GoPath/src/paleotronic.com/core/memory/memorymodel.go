package memory

import (
	"math"
	"sync"
	"time" //"runtime"

	"paleotronic.com/core/settings"
	"paleotronic.com/core/vduconst"
	"paleotronic.com/encoding/mempak"
	"paleotronic.com/fastserv"
	"paleotronic.com/fastserv/client"
	"paleotronic.com/fmt" //"log"
	"paleotronic.com/log" //	"os"
)

/*
	0 - 128k		RAM
	128k - 1024k
*/

const (
	MAX_LOG_EVENTS             = 32
	KILOBYTE                   = 1024
	MEGABYTE                   = 1048576
	OCTALYZER_MEMORY_SIZE      = OCTALYZER_NUM_INTERPRETERS * MEGABYTE
	OCTALYZER_SIM_SIZE         = 128 * KILOBYTE
	OCTALYZER_INTERPRETER_SIZE = MEGABYTE
	OCTALYZER_NUM_INTERPRETERS = settings.NUMSLOTS
	OCTALYZER_INTERPRETER_MAX  = OCTALYZER_NUM_INTERPRETERS * OCTALYZER_INTERPRETER_SIZE
	OCTALYZER_INT_STATE_SIZE   = 16
	OCTALYZER_IO_STATE_SIZE    = 16
	/*
		OCTALYZER_INTERPRETER_STATE_BASE = MEGABYTE - 500000 + 213091 // TODO: Adjust this when you add fields
		OCTALYZER_INTERPRETER_PROFILE    = OCTALYZER_INTERPRETER_STATE_BASE + 5
		OCTALYZER_INTERPRETER_LOGMODE    = OCTALYZER_INTERPRETER_STATE_BASE + 7
		OCTALYZER_CONTROL_BASE           = OCTALYZER_INTERPRETER_STATE_BASE + OCTALYZER_INT_STATE_SIZE
		OCTALYZER_LAYERSPEC_SIZE         = 256
		OCTALYZER_LAYERSTATE             = OCTALYZER_INTERPRETER_STATE_BASE + 1
		OCTALYZER_MAX_HUD_LAYERS         = 16
		OCTALYZER_MAX_GFX_LAYERS         = 16
		OCTALYZER_KEY_BUFFER_BASE        = OCTALYZER_CONTROL_BASE + 1
		OCTALYZER_KEY_BUFFER_SIZE        = 17
		OCTALYZER_MAX_PADDLES            = 4
		OCTALYZER_PADDLE_BASE            = OCTALYZER_KEY_BUFFER_BASE + OCTALYZER_KEY_BUFFER_SIZE
		OCTALYZER_PADDLE_SIZE            = 1
		PDL0                             = OCTALYZER_PADDLE_BASE
		PDL1                             = PDL0 + 1
		PDL2                             = PDL1 + 1
		PDL3                             = PDL2 + 1
		// Active states can be used to modify layer active states without rewriting all the layer data - You're Welcome (AAG)
		OCTALYZER_HUD_ACTIVESTATES    = OCTALYZER_PADDLE_BASE + OCTALYZER_MAX_PADDLES*OCTALYZER_PADDLE_SIZE
		OCTALYZER_GFX_ACTIVESTATES    = OCTALYZER_HUD_ACTIVESTATES + 1
		OCTALYZER_HUD_BASE            = OCTALYZER_GFX_ACTIVESTATES + 1
		OCTALYZER_GFX_BASE            = OCTALYZER_HUD_BASE + OCTALYZER_MAX_HUD_LAYERS*OCTALYZER_LAYERSPEC_SIZE
		OCTALYZER_RESTALGIA_BASE      = OCTALYZER_GFX_BASE + OCTALYZER_MAX_GFX_LAYERS*OCTALYZER_LAYERSPEC_SIZE
		OCTALYZER_SPEAKER_PLAYSTATE   = OCTALYZER_RESTALGIA_BASE
		OCTALYZER_SPEAKER_MODE        = OCTALYZER_SPEAKER_PLAYSTATE + 1
		OCTALYZER_SPEAKER_SAMPLERATE  = OCTALYZER_SPEAKER_MODE + 1
		OCTALYZER_SPEAKER_TOGGLE      = OCTALYZER_SPEAKER_SAMPLERATE + 1
		OCTALYZER_SPEAKER_SAMPLECOUNT = OCTALYZER_SPEAKER_TOGGLE + 1
		OCTALYZER_SPEAKER_BUFFER      = OCTALYZER_SPEAKER_SAMPLECOUNT + 1
		OCTALYZER_SPEAKER_MAX         = 44100 * 5
		OCTALYZER_SPEAKER_FREQ        = OCTALYZER_SPEAKER_BUFFER + OCTALYZER_SPEAKER_MAX
		OCTALYZER_SPEAKER_MS          = OCTALYZER_SPEAKER_FREQ + 1
		OCTALYZER_KEYCASE_STATE       = OCTALYZER_SPEAKER_MS + 1
		//--------------------------------------------------------------------------
		OCTALYZER_CAMERA_BASE = OCTALYZER_KEYCASE_STATE + 2
		//--------------------------------------------------------------------------
		OCTALYZER_CAMERA_BUFFER_SIZE = 512
		OCTALYZER_CAMERA_HUD_INDEX   = OCTALYZER_CAMERA_BASE
		OCTALYZER_CAMERA_HUD_BASE    = OCTALYZER_CAMERA_HUD_INDEX + 1
		OCTALYZER_CAMERA_GFX_INDEX   = OCTALYZER_CAMERA_HUD_BASE + OCTALYZER_CAMERA_BUFFER_SIZE
		OCTALYZER_CAMERA_GFX_BASE    = OCTALYZER_CAMERA_GFX_INDEX + 1
		// -------------------------------------------------------------------------------
		// Music stuff
		// -------------------------------------------------------------------------------
		OCTALYZER_MUSIC_BASE           = OCTALYZER_CAMERA_GFX_BASE + OCTALYZER_CAMERA_BUFFER_SIZE
		OCTALYZER_MUSIC_COMMAND        = OCTALYZER_MUSIC_BASE + 0
		OCTALYZER_MUSIC_COMMAND_LENGTH = 32 * KILOBYTE
		OCTALYZER_MUSIC_BUFFER_COUNT   = OCTALYZER_MUSIC_BASE + 1
		OCTALYZER_MUSIC_BUFFER         = OCTALYZER_MUSIC_BASE + 2
		OCTALYZER_BGCOLOR              = OCTALYZER_MUSIC_BUFFER + OCTALYZER_MUSIC_COMMAND_LENGTH
		OCTALYZER_HGR_SIZE             = OCTALYZER_BGCOLOR + 1

		OCTALYZER_MOUSE_BUTTONS        = OCTALYZER_HGR_SIZE + 1
		OCTALYZER_MOUSE_X			   = OCTALYZER_MOUSE_BUTTONS + 1
		OCTALYZER_MOUSE_Y			   = OCTALYZER_MOUSE_X + 1

		OCTALYZER_MAPPED_CAM_BASE      = OCTALYZER_MOUSE_Y + 1
		OCTALYZER_MAPPED_CAM_SIZE      = 40
		OCTALYZER_MAPPED_CAM_HUDCOUNT  = 1
		OCTALYZER_MAPPED_CAM_GFXCOUNT  = 8
		OCTALYZER_MAPPED_CAM_END       = OCTALYZER_MAPPED_CAM_BASE + (OCTALYZER_MAPPED_CAM_HUDCOUNT  + OCTALYZER_MAPPED_CAM_GFXCOUNT) * OCTALYZER_MAPPED_CAM_SIZE

		OCTALYZER_MAPPED_CAM_VIEW      = OCTALYZER_MAPPED_CAM_END
		OCTALYZER_MAPPED_CAM_CONTROL   = OCTALYZER_MAPPED_CAM_VIEW + 1

		OCTALYZER_SELECT_ENABLE        = OCTALYZER_MAPPED_CAM_CONTROL+1
		OCTALYZER_SELECT_SX            = OCTALYZER_SELECT_ENABLE + 1
		OCTALYZER_SELECT_SY            = OCTALYZER_SELECT_ENABLE + 2
		OCTALYZER_SELECT_EX            = OCTALYZER_SELECT_ENABLE + 3
		OCTALYZER_SELECT_EY            = OCTALYZER_SELECT_ENABLE + 4

		GFX_CONFIG_LO = OCTALYZER_HUD_BASE
		GFX_CONFIG_HI = OCTALYZER_RESTALGIA_BASE
	*/

	/* NEW DEFINITIONS GO PRIOR TO THIS */
	MICROM8_SPRITE_CONTROL_BASE = 655085
	MICROM8_SPRITE_ENABLE_0     = 655085 // 2 bytes
	MICROM8_SPRITE_ENABLE_1     = 655086
	MICROM8_SPRITE_CONFIG       = 655087 // 128 bytes
	MICROM8_SPRITE_DATA         = 655215 // 6144 bytes
	MICROM8_SPRITE_BACKING      = 661359 // 6144 bytes
	MICROM8_MAX_SPRITES         = 128

	MICROM8_2ND_DISKII_BASE = 667503
	MICROM8_2ND_DISKII_SIZE = 15000

	MICROM8_R6522_BASE = 697503
	MICROM8_MB_R6522_1 = 697503
	MICROM8_MB_R6522_2 = 697519

	MICROM8_RESTALGIA_PATH_LOOP = 697759
	MICROM8_RESTALGIA_PATH_SIZE = 697760
	MICROM8_RESTALGIA_PATH_BASE = 697761

	MICROM8_VOICE_PORT_BASE = 697889
	MICROM8_VOICE_PORT_SIZE = 2
	MICROM8_VOICE_COUNT     = 64

	MICROM8_RESTALGIA_CMD_LEN    = 698017
	MICROM8_RESTALGIA_CMD_BUFFER = 698018

	OCTALYZER_LOWEST_STATE = 699042

	OCTALYZER_OSD_CAM = 699043

	OCTALYZER_SLOT_HALT     = 699123
	OCTALYZER_LIGHT_DIFFUSE = 699124
	OCTALYZER_LIGHT_AMBIENT = 699125

	OCTALYZER_UPPERCASE_FORCE   = 699126
	OCTALYZER_OVERLAY_PATH_SIZE = 699127
	OCTALYZER_OVERLAY_PATH      = 699128

	OCTALYZER_BACKDROP_XPOS = 699256
	OCTALYZER_BACKDROP_YPOS = 699257
	OCTALYZER_BACKDROP_ZPOS = 699258

	OCTALYZER_BACKDROP_CAMTRACK  = 699259
	OCTALYZER_BACKDROP_ZRAT      = 699260
	OCTALYZER_BACKDROP_ZOOM      = 699261
	OCTALYZER_BACKDROP_OPACITY   = 699262
	OCTALYZER_BACKDROP_CAM       = 699263
	OCTALYZER_BACKDROP_PATH_SIZE = 699264
	OCTALYZER_BACKDROP_PATH      = 699265

	OCTALYZER_MAPPED_CAM_BASE     = 699393
	OCTALYZER_MAPPED_CAM_SIZE     = 80
	OCTALYZER_MAPPED_CAM_HUDCOUNT = 1
	OCTALYZER_MAPPED_CAM_GFXCOUNT = 8
	OCTALYZER_MAPPED_CAM_END      = 700113
	OCTALYZER_MAPPED_CAM_VIEW     = 700113
	OCTALYZER_MAPPED_CAM_CONTROL  = 700114

	OCTALYZER_TARGET_SLOT = 700115

	OCTALYZER_IO_BASE = 700116

	OCTALYZER_DISKII_SIZE       = 29995
	OCTALYZER_DISKII_BASE       = 700132
	OCTALYZER_DISKII_CARD_STATE = OCTALYZER_DISKII_BASE + 2*OCTALYZER_DISKII_SIZE

	OCTALYZER_VOXEL_DEPTH = 760132

	OCTALYZER_MAPPED_RESTALGIA_BASE = 760133
	OCTALYZER_MAPPED_VOICE_COUNT    = 16
	OCTALYZER_MAPPED_VOICE_SIZE     = 32

	OCTALYZER_SPECTRUM_RENDERMODE = 760642
	OCTALYZER_SHR_RENDERMODE      = 760643
	OCTALYZER_GR_RENDERMODE       = 760644
	OCTALYZER_HGR_RENDERMODE      = 760645
	OCTALYZER_DHGR_RENDERMODE     = 760646
	OCTALYZER_VIDEO_TINT          = 760647

	OCTALYZER_HELP_INTERRUPT = 760648
	OCTALYZER_ALT_CHARSET    = 760649

	OCTALYZER_SLOT_INTERRUPT = 760650
	OCTALYZER_SLOT_RESTART   = 760651
	OCTALYZER_CPU_HALT       = 760652
	OCTALYZER_CPU_BREAK      = 760653

	OCTALYZER_LED_0 = 760654
	OCTALYZER_LED_1 = 760655

	OCTALYZER_MAP_FILENAME_COUNT = 760656 // number of mapped images (1-10)

	OCTALYZER_MAP_FILENAME_SIZE = 100 // max image filename length

	OCTALYZER_MAP_FILENAME_0 = 760657 // filename buffers
	OCTALYZER_MAP_FILENAME_1 = 760757
	OCTALYZER_MAP_FILENAME_2 = 760857
	OCTALYZER_MAP_FILENAME_3 = 760957
	OCTALYZER_MAP_FILENAME_4 = 761057
	OCTALYZER_MAP_FILENAME_5 = 761157
	OCTALYZER_MAP_FILENAME_6 = 761257
	OCTALYZER_MAP_FILENAME_7 = 761357
	OCTALYZER_MAP_FILENAME_8 = 761457
	OCTALYZER_MAP_FILENAME_9 = 761557

	OCTALYZER_MAP_CHAR_0 = 761657 // mapped characters
	OCTALYZER_MAP_CHAR_1 = 761658
	OCTALYZER_MAP_CHAR_2 = 761659
	OCTALYZER_MAP_CHAR_3 = 761660
	OCTALYZER_MAP_CHAR_4 = 761661
	OCTALYZER_MAP_CHAR_5 = 761662
	OCTALYZER_MAP_CHAR_6 = 761663
	OCTALYZER_MAP_CHAR_7 = 761664
	OCTALYZER_MAP_CHAR_8 = 761665
	OCTALYZER_MAP_CHAR_9 = 761666

	OCTALYZER_INTERPRETER_STATE_BASE = 761667
	OCTALYZER_INTERPRETER_PROFILE    = 761672
	OCTALYZER_INTERPRETER_LOGMODE    = 761674
	OCTALYZER_CONTROL_BASE           = 761683
	OCTALYZER_LAYERSPEC_SIZE         = 256
	OCTALYZER_LAYERSTATE             = 761668
	OCTALYZER_MAX_HUD_LAYERS         = 16
	OCTALYZER_MAX_GFX_LAYERS         = 16
	OCTALYZER_KEY_BUFFER_BASE        = 761684
	OCTALYZER_KEY_BUFFER_SIZE        = 17
	OCTALYZER_MAX_PADDLES            = 4
	OCTALYZER_PADDLE_BASE            = 761701
	OCTALYZER_PADDLE_SIZE            = 1
	PDL0                             = 761701
	PDL1                             = 761702
	PDL2                             = 761703
	PDL3                             = 761704
	OCTALYZER_HUD_ACTIVESTATES       = 761705
	OCTALYZER_GFX_ACTIVESTATES       = 761706
	OCTALYZER_HUD_BASE               = 761707
	OCTALYZER_GFX_BASE               = 765803
	OCTALYZER_RESTALGIA_BASE         = 769899
	OCTALYZER_SPEAKER_PLAYSTATE      = 769899
	OCTALYZER_SPEAKER_MODE           = 769900 // speaker output stuff
	OCTALYZER_SPEAKER_SAMPLERATE     = 769901
	OCTALYZER_SPEAKER_TOGGLE         = 769902
	OCTALYZER_SPEAKER_SAMPLECOUNT    = 769903
	OCTALYZER_SPEAKER_BUFFER         = 769904
	OCTALYZER_SPEAKER_MAX            = 48000
	OCTALYZER_CASSETTE_PLAYSTATE     = 817904 // cassette output stuff
	OCTALYZER_CASSETTE_MODE          = 817905
	OCTALYZER_CASSETTE_SAMPLERATE    = 817906
	OCTALYZER_CASSETTE_TOGGLE        = 817907
	OCTALYZER_CASSETTE_SAMPLECOUNT   = 817908
	OCTALYZER_CASSETTE_BUFFER        = 817909

	OCTALYZER_DIGI_MUSIC_BASE  = 880154
	OCTALYZER_DIGI_PLAYSTATE   = 880154
	OCTALYZER_DIGI_CHANNELS    = 880155
	OCTALYZER_DIGI_SAMPLERATE  = 880156
	OCTALYZER_DIGI_TOGGLE      = 880157
	OCTALYZER_DIGI_SAMPLECOUNT = 880158
	OCTALYZER_DIGI_ATTENUATION = 880159 // 0 = full volume, 100 = silent
	OCTALYZER_DIGI_BUFFER      = 880160
	OCTALYZER_DIGI_MAX         = 110200

	OCTALYZER_SPEAKER_FREQ         = 990404
	OCTALYZER_SPEAKER_MS           = 990405
	OCTALYZER_KEYCASE_STATE        = 990406
	OCTALYZER_CAMERA_BASE          = 990408
	OCTALYZER_CAMERA_BUFFER_SIZE   = 512
	OCTALYZER_CAMERA_HUD_INDEX     = 990408
	OCTALYZER_CAMERA_HUD_BASE      = 990409
	OCTALYZER_CAMERA_GFX_INDEX     = 990921
	OCTALYZER_CAMERA_GFX_BASE      = 990922
	OCTALYZER_MUSIC_BASE           = 991434
	OCTALYZER_MUSIC_COMMAND        = 991434
	OCTALYZER_MUSIC_COMMAND_LENGTH = 32768
	OCTALYZER_MUSIC_BUFFER_COUNT   = 991435
	OCTALYZER_MUSIC_BUFFER         = 991436
	OCTALYZER_BGCOLOR              = 1024204
	OCTALYZER_HGR_SIZE             = 1024205
	OCTALYZER_MOUSE_BUTTONS        = 1024206
	OCTALYZER_MOUSE_X              = 1024207
	OCTALYZER_MOUSE_Y              = 1024208

	OCTALYZER_META_KEYINSERT = 1024209

	// OCTALYZER_MAPPED_CAM_BASE        = 1024209
	// OCTALYZER_MAPPED_CAM_SIZE        = 40
	// OCTALYZER_MAPPED_CAM_HUDCOUNT    = 1
	// OCTALYZER_MAPPED_CAM_GFXCOUNT    = 8
	// OCTALYZER_MAPPED_CAM_END         = 1024569
	// OCTALYZER_MAPPED_CAM_VIEW        = 1024569
	// OCTALYZER_MAPPED_CAM_CONTROL     = 1024570

	// FREESPACE 1024209 -> 1024570 (361 bytes)

	OCTALYZER_MEMORY_TRIGGER_BASE = 1024210
	OCTALYZER_MEMORY_TRIGGER_MAX  = 256

	OCTALYZER_MENU_TRIGGER = 1024266

	MICROM8_KEYCODE         = 1024570 // new place for keycode
	OCTALYZER_SELECT_ENABLE = 1024571
	OCTALYZER_SELECT_SX     = 1024572
	OCTALYZER_SELECT_SY     = 1024573
	OCTALYZER_SELECT_EX     = 1024574
	OCTALYZER_SELECT_EY     = 1024575
)

/******************************************************************************\
First Megabyte (Low memory):
0000k - 0128K	Super Eight interpreter 0
0128k - 0256K	Super Eight interpreter 1
0256k - 0384k   Super Eight interpreter 2
0384k - 0512k	Super Eight interpreter 3
0512k - 0640k	Super Eight interpreter 4
0640k - 0768k   Super Eight interpreter 5
0768k - 0896k   Super Eight interpreter 6
0896k - 1024k   Super Eight interpreter 7

Second, Third, Forth megabytes:
Super Eight control spaces.

4Mb -> 8Mb:
Dynamic resource loading.

All memory accesses propagate top down.
\******************************************************************************/

var WarmStart bool
var Debug bool
var Safe bool = true

type Mappable interface {
	ClaimsAddress(address int) bool
	IsDirty() bool
	SetDirty(b bool)
	RelativeWrite(offset int, value uint64)
	RelativeRead(offset int) uint64
	GetBase() int
	GetSize() int
	GetLabel() string
	Write(address int, value uint64)
	Read(address int) uint64
	ReadData(offset int) uint64
	WriteData(offset int, value uint64)
	SubscribeWriteHandler(localaddress int, handler WriteSubscriptionHandler)
	SubscribeReadHandler(localaddress int, handler ReadSubscriptionHandler)
	IsMaskEnabled() bool
	SetMaskEnabled(v bool)
	IsEnabled() bool
	SetEnabled(v bool)
	Done()
}

type WaveCallback func(index int, channel int, data []uint64, rate int, cpu bool)
type RestCallback func(index int, s string)
type MusicCallback func(index int, data []uint64, rate int, channels int, cpu bool)

type LoggerFunc func(int, []MemoryChange)

type MapperFunc func(addr int, write bool) int

type MemoryRange struct {
	Base int
	Size int
}

type DumpRequest struct {
	c     chan []MemoryChange
	Index int
}

type MapList map[MemoryRange]Mappable

var sno uint64

type MemoryChange struct {
	Sequence uint64
	Index    int
	Global   int
	Value    []uint64
	Delta    time.Duration
	When     time.Time
}

type MemoryRead struct {
	Global int
	Count  int
	C      chan []uint64
}

type MemoryWrite struct {
	Global   int
	Value    []uint64
	KeepHigh bool
	C        chan uint64
}

type MemoryMapCallback func(index int, addr int, value uint64)

const MEMCAP_NONE = 0
const MEMCAP_RECORD = 1
const MEMCAP_REMOTE = 2
const MEMCAP_CUSTOM = 4

const MAX_WRITE_MIRRORS = 2

type WriteMirror struct {
	base         int
	size         int
	destinations []int
	f            func(mm *MemoryMap, index int, address int, value uint64)
}

type MemoryMap struct {
	WriteMirrors [OCTALYZER_NUM_INTERPRETERS][MAX_WRITE_MIRRORS]*WriteMirror

	rRequest     chan MemoryRead
	wRequest     chan MemoryWrite
	lRequest     chan MemoryChange
	lDumpRequest chan DumpRequest
	mutex        sync.RWMutex
	hmutex       sync.Mutex
	logmutex     [OCTALYZER_NUM_INTERPRETERS]sync.RWMutex
	LockedVideo  bool

	LastLog time.Time

	Data                     [OCTALYZER_NUM_INTERPRETERS][]uint64 // backing is a physical array
	InterpreterMappings      [OCTALYZER_NUM_INTERPRETERS]MapList
	GlobalMappings           MapList
	MemoryHints              [OCTALYZER_NUM_INTERPRETERS]map[string][]MemoryRange
	BlockMapper              [OCTALYZER_NUM_INTERPRETERS]*MemoryManagementUnit
	RemoteLog                [OCTALYZER_NUM_INTERPRETERS][]MemoryChange
	RemoteLogCount           [OCTALYZER_NUM_INTERPRETERS]int
	RecordLog                [OCTALYZER_NUM_INTERPRETERS][]MemoryChange
	IncomingLog              [OCTALYZER_NUM_INTERPRETERS][]MemoryChange
	RecordLogCount           [OCTALYZER_NUM_INTERPRETERS]int
	Track                    [OCTALYZER_NUM_INTERPRETERS]bool
	SlotMemoryViolation      [OCTALYZER_NUM_INTERPRETERS]bool
	TrackCallback            [OCTALYZER_NUM_INTERPRETERS]LoggerFunc
	AudioCallback            [OCTALYZER_NUM_INTERPRETERS]WaveCallback
	RestalgiaCallback        [OCTALYZER_NUM_INTERPRETERS]RestCallback
	DigitalMusicCallback     [OCTALYZER_NUM_INTERPRETERS]MusicCallback
	Mutex                    [OCTALYZER_NUM_INTERPRETERS]sync.Mutex
	IncMutex                 [OCTALYZER_NUM_INTERPRETERS]sync.Mutex
	IMCache                  [OCTALYZER_NUM_INTERPRETERS][]Mappable
	CustomLogger             [OCTALYZER_NUM_INTERPRETERS]func(mc *MemoryChange)
	CustomAudioLogger        [OCTALYZER_NUM_INTERPRETERS]func(c int, rate int, bytepacked bool, indata []uint64)
	CustomAudioLoggerF       [OCTALYZER_NUM_INTERPRETERS]func(c int, rate int, bytepacked bool, indata []float32)
	CustomDigitalMusicLogger [OCTALYZER_NUM_INTERPRETERS]func(c int, rate int, bytepacked bool, indata []uint64)
	PendingLog               [OCTALYZER_NUM_INTERPRETERS]*MemoryChange
	RestOpCallback           [OCTALYZER_NUM_INTERPRETERS]func(index int, voice int, opcode int, value uint64) uint64
	WriteListener4K          [OCTALYZER_NUM_INTERPRETERS][256]func(addr int, value *uint64)

	// MemCapMode controls how / if memory capture occurs
	MemCapMode [OCTALYZER_NUM_INTERPRETERS]int
	SlotShares [OCTALYZER_NUM_INTERPRETERS]*ShareService

	spklaston [OCTALYZER_NUM_INTERPRETERS]bool
	spklevel  [OCTALYZER_NUM_INTERPRETERS]int
	spkdiff   [OCTALYZER_NUM_INTERPRETERS]int

	CallBack            [OCTALYZER_NUM_INTERPRETERS]MemoryMapCallback
	CallBackFilterIndex int
	PaddleMap           map[int]map[int]int
	KeyMap              map[int]map[uint64]uint64
	//KeyIndex            map[int]int          // reroute index
	KeyIndex          [OCTALYZER_NUM_INTERPRETERS]int
	KeyIndexWrite     [OCTALYZER_NUM_INTERPRETERS]int
	KeyIndexWriteOnly [OCTALYZER_NUM_INTERPRETERS]bool
	GlobalFunc        map[int]func(w bool)
	//AuxMap     map[int]bool
	//AuxMapR    map[int]bool
	RemoteSync map[int]*client.FSClient

	WaveValue [OCTALYZER_NUM_INTERPRETERS]float32

	Mappers [OCTALYZER_NUM_INTERPRETERS]MapperFunc
}

func NewWriteMirror(base int, size int, destinations []int, f func(mm *MemoryMap, index int, address int, value uint64)) *WriteMirror {
	return &WriteMirror{
		base:         base,
		size:         size,
		destinations: destinations,
		f:            f,
	}
}

/* Init and return a new global memory map */
func NewMemoryMap() *MemoryMap {

	var ddd [OCTALYZER_INTERPRETER_SIZE]uint64

	this := &MemoryMap{}

	// precreate slice for slot zero
	this.Data[0] = ddd[:]

	// for i, _ := range this.Data {
	// 	this.Data[i] = 0
	// }

	this.GlobalMappings = make(MapList, 0)
	this.GlobalFunc = make(map[int]func(w bool))
	//	this.MemoryHints = make(map[int]map[string][]MemoryRange)
	this.PaddleMap = make(map[int]map[int]int)
	this.KeyMap = make(map[int]map[uint64]uint64)
	this.RemoteSync = make(map[int]*client.FSClient)
	//this.KeyIndex    = make(map[int]int)

	for i := 0; i < OCTALYZER_NUM_INTERPRETERS; i++ {
		this.InterpreterMappings[i] = make(MapList, 0)
		this.PaddleMap[i] = make(map[int]int)
		for p := 0; p < 4; p++ {
			this.PaddleMap[i][p] = p
		}
		this.KeyMap[i] = make(map[uint64]uint64)
		this.KeyIndex[i] = -1 // no mapping
		this.IMCache[i] = make([]Mappable, 256)
		this.RecordLog[i] = make([]MemoryChange, 0)
		this.RemoteLog[i] = make([]MemoryChange, 0)
		this.IncomingLog[i] = make([]MemoryChange, 0)
		this.MemoryHints[i] = make(map[string][]MemoryRange)
		this.WaveValue[i] = -1
		this.BlockMapper[i] = NewMemoryManagementUnit()
	}

	return this
}

func SuperAreaSize() int {

	return OCTALYZER_HGR_SIZE + 1 - OCTALYZER_INTERPRETER_STATE_BASE

}

func (m *MemoryMap) RegisterWriteListener(index int, addr int, f func(addr int, value *uint64)) {
	m.WriteListener4K[index][addr/4096] = f
}

func (m *MemoryMap) SetRestOpCallback(index int, f func(index int, voice int, opcode int, value uint64) uint64) {
	m.RestOpCallback[index] = f
}

func (m *MemoryMap) WriteMirrorsClear(index int) {
	for i, _ := range m.WriteMirrors[index] {
		m.WriteMirrors[index][i] = nil
	}
}

func (m *MemoryMap) WriteMirrorRegister(index int, wm *WriteMirror) bool {
	for i, _ := range m.WriteMirrors[index] {
		if m.WriteMirrors[index][i] == nil {
			m.WriteMirrors[index][i] = wm
			return true
		} else if m.WriteMirrors[index][i].base == wm.base && m.WriteMirrors[index][i].size == wm.size {
			m.WriteMirrors[index][i] = wm
			return true
		}
	}
	return false
}

func (m *MemoryMap) WriteMirrorGet(index int, address int) *WriteMirror {
	if address >= 131072 {
		return nil
	}
	for i, _ := range m.WriteMirrors[index] {
		wm := m.WriteMirrors[index][i]
		if wm == nil {
			return nil
		}
		if address >= wm.base && address < wm.base+wm.size {
			return wm
		}
	}
	return nil
}

func (m *MemoryMap) SetCameraConfigure(index int, cam int) {

	m.WriteInterpreterMemory(index, OCTALYZER_MAPPED_CAM_CONTROL, uint64(cam%8))

}

func (m *MemoryMap) SetCameraView(index int, cam int) {

	m.WriteInterpreterMemory(index, OCTALYZER_MAPPED_CAM_VIEW, uint64(cam%8))

}

func (m *MemoryMap) GetCameraConfigure(index int) int {

	return int(m.ReadInterpreterMemory(index, OCTALYZER_MAPPED_CAM_CONTROL))

}

func (m *MemoryMap) GetCameraView(index int) int {

	return int(m.ReadInterpreterMemory(index, OCTALYZER_MAPPED_CAM_VIEW))

}

func (m *MemoryMap) AddIncoming(index int, addr int, value uint64) {

	m.IncMutex[index].Lock()
	defer m.IncMutex[index].Unlock()

	m.IncomingLog[index] = append(m.IncomingLog[index], MemoryChange{Global: addr, Index: index, Value: []uint64{value}})

}

func (m *MemoryMap) GetIncoming(index int) []MemoryChange {

	if len(m.IncomingLog[index]) == 0 {
		return []MemoryChange(nil)
	}

	m.IncMutex[index].Lock()
	defer m.IncMutex[index].Unlock()

	d := m.IncomingLog[index]
	m.IncomingLog[index] = make([]MemoryChange, 0)

	return d

}

func (m *MemoryMap) SlotReset(i int) {
	if settings.FirstBoot[i] {
		m.IntSetDHGRRender(i, settings.DefaultRenderModeDHGR)
		m.IntSetHGRRender(i, settings.DefaultRenderModeHGR)
		m.IntSetGRRender(i, settings.DefaultRenderModeGR)
		m.IntSetVideoTint(i, settings.DefaultTintMode)
		m.IntSetVoxelDepth(i, settings.VXD_2_TIMES)
	}
	settings.LastRenderModeDHGR[i] = settings.VM_DOTTY
	settings.LastRenderModeHGR[i] = settings.VM_DOTTY
	settings.LastTintMode[i] = settings.VPT_GREY
	settings.LastVoxelDepth[i] = settings.VXD_9_TIMES
	settings.DisableMetaMode[i] = false
	settings.CPURecordTimingPoints = 10
	m.IntSetOverlay(i, "")
	m.IntSetBackdrop(i, "", 7, 1, 16, 0, false)
	m.IntSetAmbientLevel(i, 1)
	m.IntSetDiffuseLevel(i, 1)
	m.IntSetUppercaseOnly(i, false)
	m.IntSetRestalgiaPath(i, "", false)
	m.SetBGColor(i, 0, 0, 0, 0)
	// m.IntSetCPUHalt(i, false)
	// m.IntSetSlotHalt(i, false)
	// m.IntSetSlotRestart(i, false)
	// m.IntSetSlotInterrupt(i, false)
}

func (m *MemoryMap) MEMBASE(index int) int {

	//	fmt.Println("In MEMBASE for slot", index)
	return 0
}

// LogProcessor tracks memory updates
//func (m *MemoryMap) LogProcessor() {

//	for {

//		select {
//		case l := <-m.lRequest:
//			// log something
//			m.Log[l.Index] = append(m.Log[l.Index], l)

//			if len(m.Log[l.Index]) > MAX_LOG_EVENTS && m.TrackCallback[l.Index] != nil {

//				// ship logs
//				m.TrackCallback[l.Index](m.Log[l.Index])
//				m.Log[l.Index] = make([]MemoryChange, 0)

//			}

//		case d := <-m.lDumpRequest:
//			// empty logs
//			d.c <- m.Log[d.Index]
//			m.Log[d.Index] = make([]MemoryChange, 0)

//		default:
//			time.Sleep(1 * time.Microsecond)
//		}

//	}

//}

// ActivityProcessor keeps our reads and writes in sync
//func (m *MemoryMap) ActivityProcessor() {

//	for {

//		select {
//		/* Memory read */
//		case r := <-m.rRequest:
//			chunk := make([]uint64, r.Count)
//			for i, _ := range chunk {
//				if r.Global < 0 {
//					chunk[i] = 0
//				} else {
//					chunk[i] = m.Data[r.Global+i]
//				}

//			}
//			r.C <- chunk
//		/* Memory Write */
//		case w := <-m.wRequest:
//			count := len(w.Value)
//			for i := 0; i < count; i++ {
//				if w.KeepHigh {
//					m.Data[w.Global+i] = (m.Data[w.Global+i] & 0xffffff00) | (w.Value[i] & 0xff)
//				} else {
//					m.Data[w.Global+i] = w.Value[i]
//				}
//			}
//			w.C <- uint64(count)
//		default:
//			time.Sleep(1 * time.Microsecond)
//		}

//	}

//}

func (m *MemoryMap) R(index int, addr int) uint64 {

	//if false && Safe {
	//	c := make(chan []uint64)
	//	m.rRequest <- MemoryRead{Global: addr, C: c, Count: 1}
	//	v := <-c
	//	return v[0]
	//}
	if m.Data[index] == nil {
		return 0
	}

	return m.Data[index][addr]

}

func (m *MemoryMap) AddHandler(addr int, f func(w bool)) {
	m.GlobalFunc[addr] = f
}

func (m *MemoryMap) W(index int, addr int, value uint64) bool {

	if m.Data[index] == nil {
		return false
	}

	//if false && Safe {
	//	c := make(chan uint64)
	//	m.wRequest <- MemoryWrite{Global: addr, Value: []uint64{value}, C: c}
	//	<-c
	//	return true
	//}

	i := m.GetIncoming(index)
	if len(i) > 0 {
		for _, mc := range i {
			m.Data[index][mc.Global] = mc.Value[0]
		}
	}

	if m.Data[index][addr] == value && (addr%OCTALYZER_INTERPRETER_SIZE) >= OCTALYZER_SIM_SIZE {
		return false
	}

	//	fmt.Printf("W(%d, %d)\n", addr, value)

	m.Data[index][addr] = value

	return true

}

func (m *MemoryMap) SetCallback(cb MemoryMapCallback, index int) {
	m.CallBack[index] = cb
}

/* MapInterpreterRegion sets up a mapped region */
func (m *MemoryMap) MapInterpreterRegion(index int, scope MemoryRange, mapping Mappable) {

	if index < 0 || index >= OCTALYZER_NUM_INTERPRETERS {
		panic("Invalid interpreter map access")
	}

	m.logmutex[index].Lock()
	defer m.logmutex[index].Unlock()

	m.InterpreterMappings[index][scope] = mapping

	for i := scope.Base; i < scope.Base+scope.Size; i++ {
		block := (i / 256) % 256
		m.IMCache[index][block] = mapping
	}

}

func (m *MemoryMap) InterpreterMappableAtAddress(index, localaddress int) (Mappable, bool) {

	if localaddress > 64*KILOBYTE {
		return nil, false
	}

	//log.Printf("Mappings for slot #%d: %v\n",  index, m.InterpreterMappings[index])

	//for memrange, mp := range m.InterpreterMappings[index] {
	//	if localaddress >= memrange.Base && localaddress < memrange.Base+memrange.Size && mp.IsEnabled() {
	//		return mp, true
	//	}
	//}
	mci := m.IMCache[index][(localaddress%65536)/256]
	return mci, (mci != nil && mci.IsEnabled())

	return nil, false
}

func (m *MemoryMap) WriteInterpreterMemory(index int, address int, value uint64) {

	if m.Data[index] == nil {
		return
	}

	index = index % OCTALYZER_NUM_INTERPRETERS
	address = address % OCTALYZER_INTERPRETER_SIZE

	remint, ok := m.RemoteSync[index]
	if ok {
		if (address < OCTALYZER_INTERPRETER_STATE_BASE || address > OCTALYZER_INTERPRETER_STATE_BASE+1) && !(address >= 8192 && address < 24576) {
			b := mempak.Encode(remint.RIndex, address, value, false)
			remint.SendMessage(fastserv.FS_CLIENTMEM, b)
		}
	}

	gaddr, ok := m.BlockMapper[index].Absolute(address, MA_WRITE)
	gaddr = gaddr % OCTALYZER_INTERPRETER_SIZE

	if ok {

		//fmt.Printf("index = %d, gaddr = 0x%.6x\n", index, gaddr)

		pvalue := m.Data[index][gaddr]

		m.BlockMapper[index].Do(address, MA_WRITE, &value)

		nvalue := m.Data[index][gaddr]
		if nvalue != value && (address/256 != 0xc0) && (address/256 != 0x00) {
			//fmt.RPrintf("Warning: Possible misdirect: original address %d - wrote %d, but readback gave %d\n", address, value, nvalue)
		}

		// check for mirrors
		if address < 131072 {
			wm := m.WriteMirrorGet(index, address)
			if wm != nil {
				for _, d := range wm.destinations {
					wm.f(m, index, d, value)
				}
			}
		}

		m.LogMCBWrite(index, gaddr, value, pvalue)

		f := m.WriteListener4K[index][(gaddr%OCTALYZER_INTERPRETER_SIZE)/4096]
		if f != nil {
			f(gaddr%OCTALYZER_INTERPRETER_SIZE, &value)
		}

	} else {

		//		fmt.Printf("about to write to slot index = %d, mb = %d, address = %d, value  = %d\n", index, m.MEMBASE(index), address, value)
		if address < 0x10000 {
			return
		}

		// fix for listeners above 64kb
		// if m.BlockMapper[index].HasListeners(address, nil, MA_WRITE) {
		// 	m.BlockMapper[index].ProcessListeners(address, &value, MA_WRITE)
		// }

		pvalue := m.Data[index][address]

		if m.W(index, address, value) {
			m.LogMCBWrite(index, address, value, pvalue)
		}

		f := m.WriteListener4K[index][address/4096]
		if f != nil {
			f(address, &value)
		}

		if address%OCTALYZER_INTERPRETER_SIZE == OCTALYZER_SPEAKER_PLAYSTATE && value != 0 {
			m.HandleAudio(index, 0)
		}

	}

}

func (m *MemoryMap) WriteInterpreterMemorySilent(index int, address int, value uint64) {

	//~ m.Mutex[index].Lock()
	//~ defer m.Mutex[index].Unlock()
	if m.Data[index] == nil {
		return
	}

	index = index % OCTALYZER_NUM_INTERPRETERS

	address = address % OCTALYZER_INTERPRETER_SIZE

	m.W(index, address, value)

}

func (m *MemoryMap) ReadInterpreterMemory(index int, address int) uint64 {

	if m.Data[index] == nil {
		return 0
	}

	if address < 65536 {
		var value uint64
		m.BlockMapper[index].Do(address, MA_READ, &value)
		return value
	}

	return m.Data[index][address]

}

func (m *MemoryMap) ReadInterpreterMemorySilent(index int, address int) uint64 {
	//index = index % OCTALYZER_NUM_INTERPRETERS
	if m.Data[index] == nil {
		return 0
	}
	return m.Data[index][address]
}

func (m *MemoryMap) InterpreterMappableByLabel(index int, name string) (Mappable, bool) {

	for _, mp := range m.InterpreterMappings[index] {
		if mp.GetLabel() == name {
			return mp, true
		}
	}

	return nil, false
}

func (m *MemoryMap) WriteGlobal(index int, addr int, value uint64) {
	if addr < 0 || addr >= OCTALYZER_MEMORY_SIZE {
		return
	}

	if m.Data[index] == nil {
		return
	}

	// handle if it maps to bottom 1mb -- delegate to the handler for the interpreter
	if (addr % OCTALYZER_INTERPRETER_SIZE) < OCTALYZER_SIM_SIZE {
		index := addr / OCTALYZER_INTERPRETER_SIZE
		m.WriteInterpreterMemory(index, addr%OCTALYZER_INTERPRETER_SIZE, value)
	} else {

		pvalue := m.Data[index][addr]

		if m.W(index, addr, value) {
			m.LogMCBWrite(index, addr, value, pvalue)
		}

		address := addr % OCTALYZER_INTERPRETER_SIZE

		f := m.WriteListener4K[index][address/4096]
		if f != nil {
			f(address, &value)
		}

		if addr%OCTALYZER_INTERPRETER_SIZE == OCTALYZER_SPEAKER_PLAYSTATE && value != 0 {
			m.HandleAudio(index, 0)
		} else if address >= MICROM8_VOICE_PORT_BASE && address < MICROM8_VOICE_PORT_BASE+128 {
			if (address-MICROM8_VOICE_PORT_BASE)%2 == 0 {

				voice := (address - MICROM8_VOICE_PORT_BASE) / 2
				memval := m.ReadInterpreterMemorySilent(index, address+1)
				opcode := int(value)

				_ = m.RestalgiaOpCode(
					index,
					voice,
					opcode,
					memval,
				)
				//m.WriteInterpreterMemorySilent(index, address+1, nmemval)
			}
		}

		remint, ok := m.RemoteSync[index]
		if ok {

			addr = addr % OCTALYZER_INTERPRETER_SIZE

			if (address < OCTALYZER_INTERPRETER_STATE_BASE || address > OCTALYZER_INTERPRETER_STATE_BASE+1) && !(address >= 8192 && address < 24576) {
				b := mempak.Encode(remint.RIndex, address, value, false)
				remint.SendMessage(fastserv.FS_CLIENTMEM, b)
			}
		}
	}
}

func (m *MemoryMap) WriteGlobalSilent(index int, addr int, value uint64) {

	if m.Data[index] == nil {
		return
	}

	if addr < 0 || addr >= OCTALYZER_MEMORY_SIZE {
		return
	}
	// handle if it maps to bottom 1mb -- delegate to the handler for the interpreter
	//if (addr % OCTALYZER_INTERPRETER_SIZE) < OCTALYZER_SIM_SIZE {
	m.WriteInterpreterMemorySilent(index, addr%OCTALYZER_INTERPRETER_SIZE, value)
	//} else {
	//	// above 1mb
	//	//		m.mutex.Lock()
	//	//		m.Data[addr] = value
	//	//		m.mutex.Unlock()

	//	m.W(addr, value)
	//	//m.LogMCBWrite(addr, value)
	//}
}

func (m *MemoryMap) ReadGlobal(index int, addr int) uint64 {
	if addr < 0 || addr >= OCTALYZER_MEMORY_SIZE {
		return 0
	}

	if m.Data[index] == nil {
		return 0
	}

	// handle if it maps to bottom 1mb -- delegate to the handler for the interpreter
	//if (addr % OCTALYZER_INTERPRETER_SIZE) < OCTALYZER_SIM_SIZE {
	return m.ReadInterpreterMemory(index, addr%OCTALYZER_INTERPRETER_SIZE)
	//} else {
	// above 1mb
	//		m.mutex.Lock()
	//		z := m.Data[addr]
	//		m.mutex.Unlock()

	//	z := m.R(addr)
	//	return z
	//}
}

func (m *MemoryMap) KeyBufferSize(index int) int {

	ni, ex := m.KeyIndex[index], (m.KeyIndex[index] != -1)
	if ex {
		index = ni
		if index >= OCTALYZER_NUM_INTERPRETERS {
			return 0
		}
	}

	address := OCTALYZER_KEY_BUFFER_BASE
	return int(m.ReadGlobal(index, address))
}

func (m *MemoryMap) KeyBufferHasBreak(index int) bool {

	ni, ex := m.KeyIndex[index], (m.KeyIndex[index] != -1)
	if ex {
		index = ni
		if index >= OCTALYZER_NUM_INTERPRETERS {
			return false
		}
	}

	address := OCTALYZER_KEY_BUFFER_BASE

	count := int(m.ReadGlobal(index, address))

	for i := 0; i < count; i++ {
		if m.ReadGlobal(index, address+1+i)&0xffff == 3 {
			return true
		}
	}

	return false
}

func (m *MemoryMap) KeyBufferHasSpecialBreak(index int) bool {

	ni, ex := m.KeyIndex[index], (m.KeyIndex[index] != -1)
	if ex {
		index = ni
		if index >= OCTALYZER_NUM_INTERPRETERS {
			return false
		}
	}

	address := m.MEMBASE(index) + OCTALYZER_KEY_BUFFER_BASE

	count := int(m.ReadGlobal(index, address))

	for i := 0; i < count; i++ {
		if m.ReadGlobal(index, address+1+i)&0xffff == vduconst.SHIFT_CTRL_C {
			return true
		}
	}

	return false
}

func (m *MemoryMap) KeyBufferAdd(index int, value uint64) {

	ni, ex := m.KeyIndex[index], (m.KeyIndex[index] != -1)
	if ex {
		//log2.Printf("redirecting key to slot %d", ni)
		index = ni
		if index >= OCTALYZER_NUM_INTERPRETERS {
			return
		}
	}

	wi, ex := m.KeyIndexWrite[index], (m.KeyIndexWrite[index] != -1)
	if ex {
		index = wi
		if index >= OCTALYZER_NUM_INTERPRETERS {
			return
		}
	}

	nvalue, ex := m.KeyMap[index][value]
	if ex {
		if nvalue == 0 {
			return
		} else {
			value = nvalue
		}
	}

	address := OCTALYZER_KEY_BUFFER_BASE
	//ptr := int(m.ReadGlobal(index,address))

	//if ptr < OCTALYZER_KEY_BUFFER_SIZE-1 {
	//	m.WriteGlobal(address+1+ptr, value)
	//	ptr++
	//	m.WriteGlobal(address, uint64(ptr))
	//} else {
	// m.KeyBufferEmpty(index)
	ptr := int(m.ReadGlobal(index, address))
	if ptr >= OCTALYZER_KEY_BUFFER_SIZE {
		return // prevent buffer overrun
	}
	m.WriteGlobal(index, address+1+ptr, value)
	ptr++
	m.WriteGlobal(index, address, uint64(ptr))
	//}

	//log.Printf("Scancode %d sent to buffer %d\n", value, index)
}

func (m *MemoryMap) KeyBufferAddNoRedirect(index int, value uint64) {
	address := OCTALYZER_KEY_BUFFER_BASE
	m.KeyBufferEmpty(index)
	ptr := int(m.ReadGlobal(index, address))
	m.WriteGlobal(index, address+1+ptr, value)
	ptr++
	m.WriteGlobal(index, address, uint64(ptr))
}

func (m *MemoryMap) KeyBufferEmpty(index int) {

	ni, ex := m.KeyIndex[index], (m.KeyIndex[index] != -1)
	if ex {
		index = ni
		if index >= OCTALYZER_NUM_INTERPRETERS {
			return
		}
	}

	address := m.MEMBASE(index) + OCTALYZER_KEY_BUFFER_BASE
	m.WriteGlobal(index, address, 0)
}

func (m *MemoryMap) MetaKeyPeek(index int) (rune, rune) {

	code := m.ReadInterpreterMemory(index, OCTALYZER_META_KEYINSERT)
	if code != 0 {
		return rune(code & 0xFFFF), rune(code >> 16)
	}

	return 0, 0
}

func (m *MemoryMap) MetaKeyGet(index int) (rune, rune) {

	code := m.ReadInterpreterMemory(index, OCTALYZER_META_KEYINSERT)
	if code != 0 {
		m.WriteInterpreterMemorySilent(index, OCTALYZER_META_KEYINSERT, 0)
		return rune(code & 0xFFFF), rune(code >> 16)
	}

	return 0, 0
}

func (m *MemoryMap) MetaKeySet(index int, key, subkey rune) {
	code := (uint64(subkey) << 16) | uint64(key)
	m.WriteInterpreterMemorySilent(index, OCTALYZER_META_KEYINSERT, code)
}

func (m *MemoryMap) KeyBufferGet(index int) uint64 {

	ni, ex := m.KeyIndex[index], (m.KeyIndex[index] != -1)
	if ex && !m.KeyIndexWriteOnly[index] {
		index = ni
		if index >= OCTALYZER_NUM_INTERPRETERS {
			return 0
		}
	}

	address := m.MEMBASE(index) + OCTALYZER_KEY_BUFFER_BASE
	ptr := int(m.ReadGlobal(index, address))

	if ptr > 0 {
		v := m.ReadGlobal(index, address+1)
		for x := 1; x <= OCTALYZER_KEY_BUFFER_SIZE-1; x++ {
			m.WriteGlobal(index, address+x, m.ReadGlobal(index, address+1+x))
		}
		ptr--
		m.WriteGlobal(index, address, uint64(ptr))
		return v
	}

	return 0
}

func (m *MemoryMap) KeyBufferGetNoRedirect(index int) uint64 {
	address := m.MEMBASE(index) + OCTALYZER_KEY_BUFFER_BASE
	ptr := int(m.ReadGlobal(index, address))

	if ptr > 0 {
		v := m.ReadGlobal(index, address+1)
		for x := 1; x <= OCTALYZER_KEY_BUFFER_SIZE-1; x++ {
			m.WriteGlobal(index, address+x, m.ReadGlobal(index, address+1+x))
		}
		ptr--
		m.WriteGlobal(index, address, uint64(ptr))
		return v
	}

	return 0
}

func (m *MemoryMap) KeyBufferHasRune(index int, target rune) bool {

	ni, ex := m.KeyIndex[index], (m.KeyIndex[index] != -1)
	if ex {
		index = ni
		if index >= OCTALYZER_NUM_INTERPRETERS {
			return false
		}
	}

	address := m.MEMBASE(index) + OCTALYZER_KEY_BUFFER_BASE
	ptr := int(m.ReadGlobal(index, address))

	if ptr > 0 {
		for i := 0; i < ptr; i++ {
			v := m.ReadGlobal(index, address+1+i)
			if rune(v) == target {
				m.WriteGlobal(index, address+1+i, 0)
				return true
			}
		}
	}

	return false
}

func (m *MemoryMap) KeyBufferPeek(index int) uint64 {

	ni, ex := m.KeyIndex[index], (m.KeyIndex[index] != -1)
	if ex {
		index = ni
		if index >= OCTALYZER_NUM_INTERPRETERS {
			return 0
		}
	}

	address := m.MEMBASE(index) + OCTALYZER_KEY_BUFFER_BASE
	ptr := int(m.ReadGlobal(index, address))

	if ptr > 0 {
		v := m.ReadGlobal(index, address+1)
		return v
	}

	return 0
}

func (m *MemoryMap) KeyBufferGetLatestNoRedirect(index int) uint64 {
	address := m.MEMBASE(index) + OCTALYZER_KEY_BUFFER_BASE
	ptr := int(m.ReadGlobal(index, address))

	if ptr > 0 {
		v := m.ReadGlobal(index, address+ptr)
		ptr = 0
		m.WriteGlobal(index, address, uint64(ptr))
		return v
	}

	return 0
}

func (m *MemoryMap) KeyBufferGetLatest(index int) uint64 {

	ni, ex := m.KeyIndex[index], (m.KeyIndex[index] != -1)
	if ex && !m.KeyIndexWriteOnly[index] {
		index = ni
		if index >= OCTALYZER_NUM_INTERPRETERS {
			return 0
		}
	}

	address := m.MEMBASE(index) + OCTALYZER_KEY_BUFFER_BASE
	ptr := int(m.ReadGlobal(index, address))

	if ptr > 0 {
		v := m.ReadGlobal(index, address+ptr)
		ptr = 0
		m.WriteGlobal(index, address, uint64(ptr))
		return v
	}

	return 0
}

// -- non zero active state means interpreter should be rendered

func (m *MemoryMap) IntEnableKeyRedirect(index int, target int, writeOnly bool) {
	m.KeyIndexWrite[index] = target
	m.KeyIndexWriteOnly[index] = writeOnly
}

func (m *MemoryMap) IntSetUppercaseOnly(index int, on bool) {
	base := m.MEMBASE(index) + OCTALYZER_UPPERCASE_FORCE
	v := 0
	if on {
		v = 1
	}
	m.WriteGlobal(index, base, uint64(v))
}

func (m *MemoryMap) IntGetUppercaseOnly(index int) bool {
	base := m.MEMBASE(index) + OCTALYZER_UPPERCASE_FORCE
	return m.ReadGlobal(index, base) != 0
}

func (m *MemoryMap) IntSetTargetSlot(index int, state int) {
	base := m.MEMBASE(index) + OCTALYZER_TARGET_SLOT
	v := uint64(state)
	m.WriteGlobal(index, base, v)
}

func (m *MemoryMap) IntGetTargetSlot(index int) int {
	base := m.MEMBASE(index) + OCTALYZER_TARGET_SLOT
	return int(m.ReadGlobal(index, base))
}

func (m *MemoryMap) IntSetSlotRestart(index int, state bool) {
	base := m.MEMBASE(index) + OCTALYZER_SLOT_RESTART
	v := uint64(0)
	if state {
		v = 1
	}
	m.WriteGlobal(index, base, v)
}

func (m *MemoryMap) IntGetSlotRestart(index int) bool {
	base := m.MEMBASE(index) + OCTALYZER_SLOT_RESTART
	return m.ReadGlobal(index, base) != 0
}

func (m *MemoryMap) IntSetSlotMenu(index int, state bool) {
	base := m.MEMBASE(index) + OCTALYZER_MENU_TRIGGER
	v := uint64(0)
	if state {
		v = 1
	}
	m.WriteGlobal(index, base, v)
}

func (m *MemoryMap) IntGetSlotMenu(index int) bool {
	base := m.MEMBASE(index) + OCTALYZER_MENU_TRIGGER
	return m.ReadGlobal(index, base) != 0
}

func (m *MemoryMap) IntSetSlotHalt(index int, state bool) {
	base := m.MEMBASE(index) + OCTALYZER_SLOT_HALT
	v := uint64(0)
	if state {
		v = 1
	}
	m.WriteGlobal(index, base, v)
}

func (m *MemoryMap) IntGetSlotHalt(index int) bool {
	base := m.MEMBASE(index) + OCTALYZER_SLOT_HALT
	return m.ReadGlobal(index, base) != 0
}

func (m *MemoryMap) IntSetAltChars(index int, state bool) {
	base := m.MEMBASE(index) + OCTALYZER_ALT_CHARSET
	v := uint64(0)
	if state {
		v = 1
	}
	m.WriteGlobal(index, base, v)
}

func (m *MemoryMap) IntGetAltChars(index int) bool {
	base := m.MEMBASE(index) + OCTALYZER_ALT_CHARSET
	return m.ReadGlobal(index, base) != 0
}

func (m *MemoryMap) IntSetSlotInterrupt(index int, state bool) {
	base := m.MEMBASE(index) + OCTALYZER_SLOT_INTERRUPT
	v := uint64(0)
	if state {
		v = 1
	}
	m.WriteGlobal(index, base, v)
}

func (m *MemoryMap) IntGetHelpInterrupt(index int) bool {
	base := m.MEMBASE(index) + OCTALYZER_HELP_INTERRUPT
	return m.ReadGlobal(index, base) != 0
}

func (m *MemoryMap) IntSetHelpInterrupt(index int, state bool) {
	base := m.MEMBASE(index) + OCTALYZER_HELP_INTERRUPT
	v := uint64(0)
	if state {
		v = 1
	}
	m.WriteGlobal(index, base, v)
}

func (m *MemoryMap) IntGetSlotInterrupt(index int) bool {
	base := m.MEMBASE(index) + OCTALYZER_SLOT_INTERRUPT
	return m.ReadGlobal(index, base) != 0
}

// --------------------------------------------------------
func (m *MemoryMap) IntBumpCounter(index int, counter int) {
	base := m.MEMBASE(index) + OCTALYZER_MEMORY_TRIGGER_BASE + (counter % 256)
	v := m.ReadGlobal(index, base)
	m.WriteGlobal(index, base, v+1)
}

func (m *MemoryMap) IntGetCounter(index int, counter int) uint64 {
	base := m.MEMBASE(index) + OCTALYZER_MEMORY_TRIGGER_BASE + (counter % 256)
	v := m.ReadGlobal(index, base)
	m.WriteGlobal(index, base, 0)
	return v
}

// --------------------------------------------------------

func (m *MemoryMap) IntSetCPUBreak(index int, state bool) {
	base := m.MEMBASE(index) + OCTALYZER_CPU_BREAK
	v := uint64(0)
	if state {
		v = 1
	}
	m.WriteGlobal(index, base, v)
}

func (m *MemoryMap) IntGetCPUBreak(index int) bool {
	base := m.MEMBASE(index) + OCTALYZER_CPU_BREAK
	return m.ReadGlobal(index, base) != 0
}

func (m *MemoryMap) IntSetCPUHalt(index int, state bool) {
	base := m.MEMBASE(index) + OCTALYZER_CPU_HALT
	v := uint64(0)
	if state {
		v = 1
	}
	m.WriteGlobal(index, base, v)
}

func (m *MemoryMap) IntGetCPUHalt(index int) bool {
	base := m.MEMBASE(index) + OCTALYZER_CPU_HALT
	return m.ReadGlobal(index, base) != 0
}

/****/
func (m *MemoryMap) IntSetVoxelDepth(index int, state settings.VoxelDepth) {
	base := m.MEMBASE(index) + OCTALYZER_VOXEL_DEPTH
	m.WriteGlobal(index, base+0, uint64(state))
}

func (m *MemoryMap) IntGetVoxelDepth(index int) settings.VoxelDepth {
	base := m.MEMBASE(index) + OCTALYZER_VOXEL_DEPTH
	return settings.VoxelDepth(m.ReadGlobal(index, base+0))
}

func (m *MemoryMap) IntSetGRRender(index int, state settings.VideoMode) {
	base := m.MEMBASE(index) + OCTALYZER_GR_RENDERMODE
	m.WriteGlobal(index, base+0, uint64(state))
}

func (m *MemoryMap) IntGetGRRender(index int) settings.VideoMode {
	base := m.MEMBASE(index) + OCTALYZER_GR_RENDERMODE
	return settings.VideoMode(m.ReadGlobal(index, base+0))
}

func (m *MemoryMap) IntSetHGRRender(index int, state settings.VideoMode) {
	base := m.MEMBASE(index) + OCTALYZER_HGR_RENDERMODE
	m.WriteGlobal(index, base+0, uint64(state))
}

func (m *MemoryMap) IntGetHGRRender(index int) settings.VideoMode {
	base := m.MEMBASE(index) + OCTALYZER_HGR_RENDERMODE
	return settings.VideoMode(m.ReadGlobal(index, base+0))
}

func (m *MemoryMap) IntSetSpectrumRender(index int, state settings.VideoMode) {
	base := m.MEMBASE(index) + OCTALYZER_SPECTRUM_RENDERMODE
	m.WriteGlobal(index, base+0, uint64(state))
}

func (m *MemoryMap) IntGetSpectrumRender(index int) settings.VideoMode {
	base := m.MEMBASE(index) + OCTALYZER_SPECTRUM_RENDERMODE
	return settings.VideoMode(m.ReadGlobal(index, base+0))
}

func (m *MemoryMap) IntSetSHRRender(index int, state settings.VideoMode) {
	base := m.MEMBASE(index) + OCTALYZER_SHR_RENDERMODE
	m.WriteGlobal(index, base+0, uint64(state))
}

func (m *MemoryMap) IntGetSHRRender(index int) settings.VideoMode {
	base := m.MEMBASE(index) + OCTALYZER_SHR_RENDERMODE
	return settings.VideoMode(m.ReadGlobal(index, base+0))
}

func (m *MemoryMap) IntSetDHGRRender(index int, state settings.VideoMode) {
	base := m.MEMBASE(index) + OCTALYZER_DHGR_RENDERMODE
	m.WriteGlobal(index, base+0, uint64(state))
}

func (m *MemoryMap) IntGetDHGRRender(index int) settings.VideoMode {
	base := m.MEMBASE(index) + OCTALYZER_DHGR_RENDERMODE
	return settings.VideoMode(m.ReadGlobal(index, base+0))
}

func (m *MemoryMap) RestalgiaOpCode(index int, voice int, opcode int, value uint64) uint64 {
	// fmt.RPrintf(
	// 	"-> RestOpCode(%d, %d, %d, %d)\n",
	// 	index, voice, opcode, value,
	// )

	if m.RestOpCallback[index] != nil {
		return m.RestOpCallback[index](index, voice, opcode, value)
	}
	return 0
}

const BackdropNew = 1 << 31

func (m *MemoryMap) IntSetBackdrop(index int, backdrop string, camidx int, opacity float32, zoom float32, zoomfactor float32, camtrack bool) {

	//fmt.Printf("IntSetBackdrop(%d, %s, %d, %f, %f, %f)\n", index, backdrop, camidx, opacity, zoom, zoomfactor)
	// fmt2.Printf("Set backdrop called: %s\n", backdrop)
	// debug.PrintStack()

	if len(backdrop) > 128 {
		backdrop = ""
	}

	if camidx != -1 && camidx >= 1 && camidx <= 7 {
		m.WriteInterpreterMemory(index, OCTALYZER_BACKDROP_CAM, uint64(camidx))
	}

	if opacity > 1 {
		opacity = 1
	}
	if opacity < 0 {
		opacity = 0
	}
	if zoomfactor < 0 {
		zoomfactor = 0
	}

	m.WriteInterpreterMemory(index, OCTALYZER_BACKDROP_ZOOM, uint64(zoom*1000000))

	m.WriteInterpreterMemory(index, OCTALYZER_BACKDROP_ZRAT, uint64(zoomfactor*1000000))

	m.WriteInterpreterMemory(index, OCTALYZER_BACKDROP_OPACITY, uint64(opacity*100))

	if camtrack {
		m.WriteInterpreterMemory(index, OCTALYZER_BACKDROP_CAMTRACK, 1)
	} else {
		m.WriteInterpreterMemory(index, OCTALYZER_BACKDROP_CAMTRACK, 0)
	}

	// path component
	for i, ch := range backdrop {
		m.WriteInterpreterMemory(index, OCTALYZER_BACKDROP_PATH+i, uint64(ch))
	}

	// size
	m.WriteInterpreterMemory(index, OCTALYZER_BACKDROP_PATH_SIZE, uint64(len(backdrop))|BackdropNew)

}

func (m *MemoryMap) IntGetBackdropIsNew(index int) bool {
	return m.ReadInterpreterMemorySilent(index, OCTALYZER_BACKDROP_PATH_SIZE)&BackdropNew != 0
}

func (m *MemoryMap) IntClearBackdropIsNew(index int) {
	m.WriteInterpreterMemorySilent(index, OCTALYZER_BACKDROP_PATH_SIZE, m.ReadInterpreterMemorySilent(index, OCTALYZER_BACKDROP_PATH_SIZE)&0x1fffffff)
}

func (m *MemoryMap) IntGetAmbientLevel(index int) float32 {
	x := float32(math.Float64frombits(m.ReadInterpreterMemorySilent(index, OCTALYZER_LIGHT_AMBIENT)))
	if x <= 0 {
		x = 0
	} else if x > 5 {
		x = 5
	}
	return x
}

func (m *MemoryMap) IntSetAmbientLevel(index int, x float32) {
	if x <= 0 {
		x = 0
	} else if x > 5 {
		x = 5
	}
	m.WriteInterpreterMemorySilent(index, OCTALYZER_LIGHT_AMBIENT, math.Float64bits(float64(x)))
}

func (m *MemoryMap) IntGetDiffuseLevel(index int) float32 {
	x := float32(math.Float64frombits(m.ReadInterpreterMemorySilent(index, OCTALYZER_LIGHT_DIFFUSE)))
	if x <= 0 {
		x = 0
	} else if x > 5 {
		x = 5
	}
	return x
}

func (m *MemoryMap) IntSetDiffuseLevel(index int, x float32) {
	if x <= 0 {
		x = 0
	} else if x > 5 {
		x = 5
	}
	m.WriteInterpreterMemorySilent(index, OCTALYZER_LIGHT_DIFFUSE, math.Float64bits(float64(x)))
}

func (m *MemoryMap) IntGetBackdropPos(index int) (float64, float64, float64) {
	x := math.Float64frombits(m.ReadInterpreterMemorySilent(index, OCTALYZER_BACKDROP_XPOS))
	y := math.Float64frombits(m.ReadInterpreterMemorySilent(index, OCTALYZER_BACKDROP_YPOS))
	z := math.Float64frombits(m.ReadInterpreterMemorySilent(index, OCTALYZER_BACKDROP_ZPOS))
	return x, y, z
}

func (m *MemoryMap) IntSetBackdropPos(index int, x, y, z float64) {
	m.WriteInterpreterMemorySilent(index, OCTALYZER_BACKDROP_XPOS, math.Float64bits(x))
	m.WriteInterpreterMemorySilent(index, OCTALYZER_BACKDROP_YPOS, math.Float64bits(y))
	m.WriteInterpreterMemorySilent(index, OCTALYZER_BACKDROP_ZPOS, math.Float64bits(z))
}

func (m *MemoryMap) IntGetBackdrop(index int) (bool, string, int, float32, float32, float32, bool) {

	l := int(m.ReadInterpreterMemorySilent(index, OCTALYZER_BACKDROP_PATH_SIZE) & 0xff)
	if l > 128 {
		l = 0
	}

	backdrop := ""
	for i := 0; i < l; i++ {
		backdrop += string(rune(m.ReadInterpreterMemorySilent(index, OCTALYZER_BACKDROP_PATH+i)))
	}

	opacity := float32(m.ReadInterpreterMemorySilent(index, OCTALYZER_BACKDROP_OPACITY)) / 100
	zoom := float32(m.ReadInterpreterMemorySilent(index, OCTALYZER_BACKDROP_ZOOM)) / 1000000
	zoomfactor := float32(m.ReadInterpreterMemorySilent(index, OCTALYZER_BACKDROP_ZRAT)) / 1000000

	camidx := int(m.ReadInterpreterMemorySilent(index, OCTALYZER_BACKDROP_CAM))

	camtrack := m.ReadInterpreterMemorySilent(index, OCTALYZER_BACKDROP_CAMTRACK) != 0

	return (backdrop != ""), backdrop, camidx, opacity, zoom, zoomfactor, camtrack
}

func (m *MemoryMap) IntGetRestalgiaPath(index int) (bool, string, bool) {

	l := int(m.ReadInterpreterMemorySilent(index, MICROM8_RESTALGIA_PATH_SIZE))
	if l > 128 {
		l = 0
	}

	backdrop := ""
	for i := 0; i < l; i++ {
		backdrop += string(rune(m.ReadInterpreterMemorySilent(index, MICROM8_RESTALGIA_PATH_BASE+i)))
	}

	loop := m.ReadInterpreterMemorySilent(index, MICROM8_RESTALGIA_PATH_LOOP) != 0

	return (backdrop != ""), backdrop, loop

}

func (m *MemoryMap) IntSetRestalgiaPath(index int, backdrop string, loop bool) {

	//fmt.Printf("IntSetBackdrop(%d, %s, %d, %f, %f, %f)\n", index, backdrop, camidx, opacity, zoom, zoomfactor)

	if len(backdrop) > 128 {
		backdrop = ""
	}

	// path component
	for i, ch := range backdrop {
		m.WriteInterpreterMemory(index, MICROM8_RESTALGIA_PATH_BASE+i, uint64(ch))
	}

	// size
	m.WriteInterpreterMemory(index, MICROM8_RESTALGIA_PATH_SIZE, uint64(len(backdrop)))
	var vloop uint64
	if loop {
		vloop = 1
	}
	m.WriteInterpreterMemory(index, MICROM8_RESTALGIA_PATH_LOOP, vloop)

}

func (m *MemoryMap) IntSetOverlay(index int, backdrop string) {

	//log2.Printf("setting overlay: %d -> %s", index, backdrop)

	//debug.PrintStack()

	//fmt.Printf("IntSetBackdrop(%d, %s, %d, %f, %f, %f)\n", index, backdrop, camidx, opacity, zoom, zoomfactor)

	if len(backdrop) > 128 {
		backdrop = ""
	}

	// path component
	for i, ch := range backdrop {
		m.WriteInterpreterMemory(index, OCTALYZER_OVERLAY_PATH+i, uint64(ch))
	}

	// size
	m.WriteInterpreterMemory(index, OCTALYZER_OVERLAY_PATH_SIZE, uint64(len(backdrop)))

}

func (m *MemoryMap) IntGetOverlay(index int) (bool, string) {

	l := int(m.ReadInterpreterMemorySilent(index, OCTALYZER_OVERLAY_PATH_SIZE))
	if l > 128 {
		l = 0
	}

	backdrop := ""
	for i := 0; i < l; i++ {
		backdrop += string(rune(m.ReadInterpreterMemorySilent(index, OCTALYZER_OVERLAY_PATH+i)))
	}

	return len(backdrop) > 0, backdrop
}

func (m *MemoryMap) IntSetVideoTint(index int, state settings.VideoPaletteTint) {
	base := m.MEMBASE(index) + OCTALYZER_VIDEO_TINT
	m.WriteGlobal(index, base+0, uint64(state))
}

func (m *MemoryMap) IntGetVideoTint(index int) settings.VideoPaletteTint {
	base := m.MEMBASE(index) + OCTALYZER_VIDEO_TINT
	return settings.VideoPaletteTint(m.ReadGlobal(index, base+0))
}

func (m *MemoryMap) IntSetVideoTintRGBA(index int, r, g, b, a uint8) {
	state := uint64(r)<<24 | uint64(g)<<16 | uint64(b)<<8 | uint64(a) | 1<<32
	base := m.MEMBASE(index) + OCTALYZER_VIDEO_TINT
	m.WriteGlobal(index, base+0, uint64(state))
}

func (m *MemoryMap) IntGetVideoTintRGBA(index int) (uint8, uint8, uint8, uint8) {
	base := m.MEMBASE(index) + OCTALYZER_VIDEO_TINT
	v := settings.VideoPaletteTint(m.ReadGlobal(index, base+0))

	if v&(1<<32) == 0 {
		return 0, 0, 0, 0
	}

	v = v & 0xffffffff

	return uint8(v >> 24), uint8(v >> 16), uint8(v >> 8), uint8(v)
}

/****/

func (m *MemoryMap) IntSetActiveState(index int, state uint64) {
	base := m.MEMBASE(index) + OCTALYZER_INTERPRETER_STATE_BASE
	m.WriteGlobal(index, base+0, state)
}

func (m *MemoryMap) IntGetActiveState(index int) uint64 {
	base := m.MEMBASE(index) + OCTALYZER_INTERPRETER_STATE_BASE
	return m.ReadGlobal(index, base+0)
}

func (m *MemoryMap) IntSetLED0(index int, state uint64) {
	//	fmt.Printf("LED 0: %d\n", state)
	base := m.MEMBASE(index) + OCTALYZER_LED_0
	m.WriteGlobal(index, base+0, state)
}

func (m *MemoryMap) Zero(index int) {

	if m.Data[index] == nil {
		return
	}

	log.Println("Zero")
	size := 131072
	base := m.MEMBASE(index)
	for i := 0; i < size; i++ {
		m.Data[index][base+i] = m.Data[index][base+i] & 0xffffffffffffff00
	}
	// for i := size; i < OCTALYZER_INTERPRETER_SIZE; i++ {
	// 	m.Data[base+i] = m.Data[base+i] & 0xffffffffffffff00
	// }
}

func (m *MemoryMap) ZeroAll(index int) {
	log.Println("Zero")
	size := OCTALYZER_INTERPRETER_SIZE
	base := m.MEMBASE(index)
	for i := 0; i < size; i++ {
		m.Data[index][base+i] = 0
	}
}

func (m *MemoryMap) IntGetLED0(index int) uint64 {
	base := m.MEMBASE(index) + OCTALYZER_LED_0
	return m.ReadGlobal(index, base+0)
}

func (m *MemoryMap) IntSetLED1(index int, state uint64) {
	//	fmt.Printf("LED 1: %d\n", state)
	base := m.MEMBASE(index) + OCTALYZER_LED_1
	m.WriteGlobal(index, base+0, state)
}

func (m *MemoryMap) IntGetLED1(index int) uint64 {
	base := m.MEMBASE(index) + OCTALYZER_LED_1
	return m.ReadGlobal(index, base+0)
}

// -- non zero active state means interpreter should be rendered

func (m *MemoryMap) IntSetProcessorState(index int, state uint64) {
	base := m.MEMBASE(index) + OCTALYZER_INTERPRETER_STATE_BASE
	m.WriteGlobal(index, base+2, state)
}

func (m *MemoryMap) IntGetProcessorState(index int) uint64 {
	base := m.MEMBASE(index) + OCTALYZER_INTERPRETER_STATE_BASE
	return m.ReadGlobal(index, base+2)
}

func (m *MemoryMap) IntSetZeroPageState(index int, state uint64) {
	base := m.MEMBASE(index) + OCTALYZER_INTERPRETER_STATE_BASE
	m.WriteGlobal(index, base+3, state)
}

func (m *MemoryMap) IntGetZeroPageState(index int) uint64 {
	base := m.MEMBASE(index) + OCTALYZER_INTERPRETER_STATE_BASE
	return m.ReadGlobal(index, base+3)
}

func (m *MemoryMap) IntSetPDState(index int, state uint64) {
	base := m.MEMBASE(index) + OCTALYZER_INTERPRETER_STATE_BASE
	m.WriteGlobal(index, base+4, state)
}

func (m *MemoryMap) IntGetPDState(index int) uint64 {
	base := m.MEMBASE(index) + OCTALYZER_INTERPRETER_STATE_BASE
	return m.ReadGlobal(index, base+4)
}

func (m *MemoryMap) IntSetSpeakerMode(index int, state uint64) {
	base := m.MEMBASE(index) + OCTALYZER_SPEAKER_MODE
	m.WriteGlobal(index, base, state)
}

func (m *MemoryMap) IntGetSpeakerMode(index int) uint64 {
	base := m.MEMBASE(index) + OCTALYZER_SPEAKER_MODE
	return m.ReadGlobal(index, base)
}

func (m *MemoryMap) IntSetLayerState(index int, state uint64) {
	base := m.MEMBASE(index) + OCTALYZER_INTERPRETER_STATE_BASE
	m.WriteGlobalSilent(index, base+1, state)
}

func (m *MemoryMap) IntGetLayerState(index int) uint64 {
	base := m.MEMBASE(index) + OCTALYZER_INTERPRETER_STATE_BASE
	return m.ReadGlobal(index, base+1)
}

func (m *MemoryMap) GetPaddleAddress(index int, pi int) int {
	ni, ex := m.KeyIndex[index], (m.KeyIndex[index] != -1)
	if ex {
		index = ni
		if index >= OCTALYZER_NUM_INTERPRETERS {
			return 0
		}
	}

	return m.MEMBASE(index) + OCTALYZER_PADDLE_BASE + (pi % OCTALYZER_MAX_PADDLES)
}

func (m *MemoryMap) IntSetPaddleButton(index int, pi int, state uint64) {
	ni, ex := m.KeyIndex[index], (m.KeyIndex[index] != -1)
	if ex {
		index = ni
		if index >= OCTALYZER_NUM_INTERPRETERS {
			return
		}
	}

	//fmt.RPrintf("map: %v\n", m.PaddleMap)

	npi, ex := m.PaddleMap[index][pi]
	if ex {
		pi = npi
		if npi >= 4 {
			return
		}
	}

	base := m.MEMBASE(index) + OCTALYZER_PADDLE_BASE + (pi % OCTALYZER_MAX_PADDLES)
	v := m.ReadGlobal(index, base)
	m.WriteGlobal(index, base, (v&0xffffff00)|(state&0xff))
}

func (m *MemoryMap) IntGetPaddleButton(index int, pi int) uint64 {
	ni, ex := m.KeyIndex[index], (m.KeyIndex[index] != -1)
	if ex {
		index = ni
		if index >= OCTALYZER_NUM_INTERPRETERS {
			return 0
		}
	}

	base := m.MEMBASE(index) + OCTALYZER_PADDLE_BASE + (pi % OCTALYZER_MAX_PADDLES)
	return m.ReadGlobal(index, base) & 0xff
}

func (m *MemoryMap) IntSetPaddleValue(index int, pi int, state uint64) {

	ni, ex := m.KeyIndex[index], (m.KeyIndex[index] != -1)
	if ex {
		index = ni
		if index >= OCTALYZER_NUM_INTERPRETERS {
			return
		}
	}

	npi, ex := m.PaddleMap[index][pi]
	if ex {
		//fmt.RPrintf("slot %d: pdl%d -> pdl%d\n", index, pi, npi)
		pi = npi
		if npi >= 4 {
			return
		}
	}

	// reverse paddle values if needed
	if (pi == 0 && settings.JoystickReverseX[index]) || (pi == 1 && settings.JoystickReverseY[index]) {
		state = 255 - state
	}

	base := m.MEMBASE(index) + OCTALYZER_PADDLE_BASE + (pi % OCTALYZER_MAX_PADDLES)
	v := m.ReadGlobal(index, base)
	m.WriteGlobal(index, base, (v&0xffff00ff)|((state&0xff)<<8))
}

func (m *MemoryMap) IntSetMouseX(index int, value uint64) {
	m.WriteInterpreterMemory(index, OCTALYZER_MOUSE_X, value)
}

func (m *MemoryMap) IntSetMouseY(index int, value uint64) {
	m.WriteInterpreterMemory(index, OCTALYZER_MOUSE_Y, value)
}

func (m *MemoryMap) IntGetMousePos(index int) (uint64, uint64) {

	return m.ReadInterpreterMemory(index, OCTALYZER_MOUSE_X), m.ReadInterpreterMemory(index, OCTALYZER_MOUSE_Y)

}

func (m *MemoryMap) IntSetMouseButtons(index int, left, right bool) {

	var v uint64
	if left {
		v |= 1
	}
	if right {
		v |= 2
	}

	m.WriteInterpreterMemory(index, OCTALYZER_MOUSE_BUTTONS, v)

}

func (m *MemoryMap) IntGetMouseButtons(index int) (bool, bool) {

	v := m.ReadInterpreterMemory(index, OCTALYZER_MOUSE_BUTTONS)

	return (v&1 != 0), (v&2 != 0)

}

func (m *MemoryMap) IntGetPaddleValue(index int, pi int) uint64 {
	ni, ex := m.KeyIndex[index], (m.KeyIndex[index] != -1)
	if ex {
		index = ni
		if index >= OCTALYZER_NUM_INTERPRETERS {
			return 0
		}
	}

	base := m.MEMBASE(index) + OCTALYZER_PADDLE_BASE + (pi % OCTALYZER_MAX_PADDLES)
	return (m.ReadGlobal(index, base) & 0x0000ff00) >> 8
}

func (m *MemoryMap) CreateMemoryHint(index int, name string, data []MemoryRange) {
	m.hmutex.Lock()
	defer m.hmutex.Unlock()
	m.MemoryHints[index][name] = data
}

func (m *MemoryMap) IsMappedAddress(index int, addr int) bool {

	m.hmutex.Lock()
	defer m.hmutex.Unlock()

	for _, mrs := range m.MemoryHints[index] {

		for _, mr := range mrs {

			if mr.Base <= addr && mr.Base+mr.Size > addr {
				return true
			}

		}

	}

	return false

}

/* Build a composite memory slice */
func (m *MemoryMap) GetHintedMemorySlice(index int, name string) *MemoryControlBlock {

	m.hmutex.Lock()
	defer m.hmutex.Unlock()

	s := NewMemoryControlBlock(m, index, false)

	data, ok := m.MemoryHints[index][name]
	if !ok {
		return s
	}

	base := m.MEMBASE(index)

	for _, mr := range data {
		s.Add(m.Data[index][base+mr.Base:base+mr.Base+mr.Size], base+mr.Base)

		//fmt.Printf("Add memory block %s.%d (%d @ %d)\n", name, i,  mr.Size, base+mr.Base)
	}

	return s

}

func (m *MemoryMap) SetCustomLogger(index int, f func(mc *MemoryChange)) {
	m.CustomLogger[index] = f
	if f != nil {
		m.MemCapMode[index] = MEMCAP_CUSTOM
		m.Track[index] = true
		m.PendingLog[index] = nil // clear it :)
	} else {
		m.MemCapMode[index] = 0
		m.Track[index] = false
		m.PendingLog[index] = nil // clear it
	}
}

func (m *MemoryMap) SetRecordLogger(
	index int,
	f func(mc *MemoryChange),
	fa func(c int, rate int, bytepacked bool, indata []uint64),
) {
	m.CustomLogger[index] = f
	m.CustomAudioLogger[index] = fa
	if f != nil {
		m.MemCapMode[index] = MEMCAP_RECORD
		m.Track[index] = true
		m.PendingLog[index] = nil // clear it :)
	} else {
		m.MemCapMode[index] = MEMCAP_NONE
		m.Track[index] = false
		m.PendingLog[index] = nil // clear it
	}
}

func (m *MemoryMap) DoCustomLog(mc MemoryChange) {

	// tag here with current sno
	mc.Sequence = sno
	sno++

	a := mc.Global % OCTALYZER_INTERPRETER_SIZE

	//~ if a >= OCTALYZER_RESTALGIA_BASE && a < OCTALYZER_SPEAKER_FREQ {
	//~ return
	//~ }

	b := a / 256
	if b == 0x00 || b == 0x01 || b == 0x9a || b == 0x9b || b == 0xc0 {
		return
	}

	if m.CustomLogger[mc.Index] != nil {
		m.CustomLogger[mc.Index](&mc)
	}

}

func (m *MemoryMap) DoRecordLog(mc MemoryChange) {

	// tag here with current sno
	mc.Sequence = sno
	sno++

	if m.CustomLogger[mc.Index] != nil {
		m.CustomLogger[mc.Index](&mc)
	}

}

func (m *MemoryMap) DoRemoteLog(mc MemoryChange) {
	m.logmutex[mc.Index].Lock()
	defer m.logmutex[mc.Index].Unlock()

	m.RemoteLog[mc.Index] = append(m.RemoteLog[mc.Index], mc)
}

func (m *MemoryMap) CheckRecordLog(index int) {

	if len(m.RecordLog[index]) > 0 && time.Since(m.LastLog) > 5*time.Millisecond {
		m.logmutex[index].Lock()
		defer m.logmutex[index].Unlock()
		// ship them if we can
		if m.TrackCallback[index] != nil {
			m.TrackCallback[index](index, m.RecordLog[index])
		}
		// reset counter
		m.RecordLog[index] = make([]MemoryChange, 0)
	}

}

func (m *MemoryMap) IsSlotShared(index int) bool {
	return m.Track[index] && m.MemCapMode[index]&MEMCAP_CUSTOM != 0
}

func (m *MemoryMap) LogMCBRead(addr int) {

	return

	now := time.Now()

	index := addr / OCTALYZER_INTERPRETER_SIZE
	if m.Track[index] && m.MemCapMode[index]&MEMCAP_RECORD != 0 {
		m.DoRecordLog(MemoryChange{Global: addr, Value: []uint64(nil), Index: index, Delta: time.Since(m.LastLog)})
	} else if m.Track[index] && m.MemCapMode[index]&MEMCAP_REMOTE != 0 {
		m.DoRemoteLog(MemoryChange{Global: addr, Value: []uint64(nil), Index: index, Delta: time.Since(m.LastLog)})
	} else if m.Track[index] && m.MemCapMode[index]&MEMCAP_CUSTOM != 0 {
		m.DoCustomLog(MemoryChange{Global: addr, Value: []uint64(nil), Index: index, Delta: time.Since(m.LastLog)})
	}

	m.LastLog = now

	return

}

func (m *MemoryMap) LogMCBWrite(index int, addr int, value, pvalue uint64) {

	// Filter out slot change stuff - remote should not care
	if addr >= OCTALYZER_INTERPRETER_STATE_BASE && addr <= OCTALYZER_INTERPRETER_STATE_BASE+1 {
		return
	}

	now := time.Now()

	if m.Track[index] && m.MemCapMode[index]&MEMCAP_RECORD != 0 {
		//mask := pvalue ^ value
		a := addr % OCTALYZER_INTERPRETER_SIZE
		ignore := (a == OCTALYZER_SPEAKER_SAMPLERATE) ||
			(a == OCTALYZER_SPEAKER_SAMPLECOUNT) ||
			(a == OCTALYZER_SPEAKER_PLAYSTATE) ||
			(a >= OCTALYZER_SPEAKER_BUFFER && a <= OCTALYZER_SPEAKER_BUFFER+OCTALYZER_SPEAKER_MAX)
		if !ignore || m.CustomAudioLogger[index] == nil {
			m.DoRecordLog(MemoryChange{Global: addr, Value: []uint64{value, pvalue}, Index: index, Delta: time.Since(m.LastLog)})
		}
	} else if m.Track[index] && m.MemCapMode[index]&MEMCAP_REMOTE != 0 {
		m.DoRemoteLog(MemoryChange{Global: addr, Value: []uint64{value}, Index: index, Delta: time.Since(m.LastLog)})
	} else if m.Track[index] && m.MemCapMode[index]&MEMCAP_CUSTOM != 0 {
		m.DoCustomLog(MemoryChange{Global: addr, Value: []uint64{value}, Index: index, Delta: time.Since(m.LastLog), When: now})
	}
	if m.CallBack[index] != nil {
		m.CallBack[index](index, addr, value)
	}

	m.LastLog = now

	return

}

func (m *MemoryMap) LogMCBWriteBlock(index int, addr int, values, pvalues []uint64) {

	now := time.Now()

	if m.Track[index] && m.MemCapMode[index]&MEMCAP_RECORD != 0 {
		m.DoRecordLog(MemoryChange{Global: addr, Value: append(values, pvalues...), Index: index, Delta: time.Since(m.LastLog)})
	} else if m.Track[index] && m.MemCapMode[index]&MEMCAP_REMOTE != 0 {
		m.DoRemoteLog(MemoryChange{Global: addr, Value: values, Index: index, Delta: time.Since(m.LastLog)})
	} else if m.Track[index] && m.MemCapMode[index]&MEMCAP_CUSTOM != 0 {
		m.DoCustomLog(MemoryChange{Global: addr, Value: values, Index: index, Delta: time.Since(m.LastLog), When: now})
	}

	m.LastLog = now

	return

}

func (m *MemoryMap) GetRemoteLoggedChanges(index int) []MemoryChange {

	m.logmutex[index].Lock()
	defer m.logmutex[index].Unlock()

	//d := m.Log[index]
	//m.Log[index] = make([]MemoryChange, 0)

	d := m.RemoteLog[index]
	m.RemoteLog[index] = make([]MemoryChange, 0)

	return d

}

func (m *MemoryMap) InputToggle(index int) {
	for i := 0; i < OCTALYZER_NUM_INTERPRETERS; i++ {
		if i == index || index == -1 {
			m.KeyIndex[i] = i
		} else {
			m.KeyIndex[i] = 99
		}
	}
}

func (m *MemoryMap) BlockWrite(index int, addr int, data []uint64) {

	if m.Data[index] == nil {
		return
	}

	//if addr + len(data) > OCTALYZER_MEMORY_SIZE || addr < 0 {
	//	panic("Tried to write outside of memory bounds!")
	//}
	//	fmt.Println(index)
	address := addr % OCTALYZER_INTERPRETER_SIZE

	//if Safe {
	//	c := make(chan uint64)
	//	m.wRequest <- MemoryWrite{Global: addr, Value: data, C: c}
	//	<-c
	//	m.LogMCBWriteBlock(addr, data)
	//	return
	//}

	odata := make([]uint64, len(data))

	if address < 65536 {

		// if its 8-bit mapped, send thru the blockmapper
		gaddr, _ := m.BlockMapper[index].Absolute(address, MA_WRITE)

		for i, v := range data {

			odata[i] = m.Data[index][gaddr+i]

			m.BlockMapper[index].Do(address+i, MA_WRITE, &v)

			//fmt.Printf("BW( %d, %d )\n", (addr+i) % OCTALYZER_INTERPRETER_SIZE, v )
		}

	} else {

		for i, v := range data {

			odata[i] = m.Data[index][addr+i]
			m.Data[index][addr+i] = v

			f := m.WriteListener4K[index][(address+i)/4096]
			if f != nil {
				f(address+i, &v)
			}

			//fmt.Printf("BW( %d, %d )\n", (addr+i) % OCTALYZER_INTERPRETER_SIZE, v )
		}

	}

	m.LogMCBWriteBlock(index, addr, data, odata)

}

func (m *MemoryMap) MemDump(index int, base int, count int) {
	for i := 0; i < count; i++ {
		a := base + i
		v := m.ReadInterpreterMemorySilent(index, a)
		if a%16 == 0 {
			fmt.Println()
			fmt.Printf("0x%.4x: ", a)
		}
		fmt.Printf("0x%.2x ", v&0xff)
	}
	fmt.Println()
}

func (m *MemoryMap) BlockWritePr(index int, addr int, data []uint64) {

	if m.Data[index] == nil {
		return
	}

	// index := addr / OCTALYZER_INTERPRETER_SIZE
	address := addr % OCTALYZER_INTERPRETER_SIZE
	if mapper := m.Mappers[index]; mapper != nil {
		address = mapper(address, true)
		addr = index*OCTALYZER_INTERPRETER_SIZE + address
	}

	//if Safe {
	//	c := make(chan uint64)
	//	m.wRequest <- MemoryWrite{Global: addr, Value: data, C: c, KeepHigh: true}
	//	<-c
	//	m.LogMCBWriteBlock(addr, data)
	//	return
	//}

	odata := make([]uint64, len(data))

	for i, v := range data {
		odata[i] = m.Data[index][addr+i]
		m.Data[index][addr+i] = (m.Data[index][addr+i] & 0xffffff00) | (v & 0xff)
	}

	m.LogMCBWriteBlock(index, addr, data, odata)

}

func (m *MemoryMap) BlockRead(index int, addr int, count int) []uint64 {

	return m.Data[index][addr : addr+count]

}

func (m *MemoryMap) GetGFXCameraData(slotid int, camidx int) [OCTALYZER_MAPPED_CAM_SIZE]uint64 {
	var data [OCTALYZER_MAPPED_CAM_SIZE]uint64
	base := m.MEMBASE(slotid) + OCTALYZER_MAPPED_CAM_BASE + (camidx+1)*OCTALYZER_MAPPED_CAM_SIZE
	end := base + OCTALYZER_MAPPED_CAM_SIZE
	copy(data[:], m.Data[slotid][base:end])
	return data
}

func (m *MemoryMap) SetGFXCameraData(slotid int, camidx int, data [OCTALYZER_MAPPED_CAM_SIZE]uint64) {
	base := m.MEMBASE(slotid) + OCTALYZER_MAPPED_CAM_BASE + (camidx+1)*OCTALYZER_MAPPED_CAM_SIZE
	m.BlockWritePr(slotid, base, data[:])
}

func (m *MemoryMap) SetBGColor(index int, r, g, b, a uint8) {

	var v uint64
	v = (uint64(a) << 24) | (uint64(b) << 16) | (uint64(g) << 8) | uint64(r)
	m.W(index, m.MEMBASE(index)+OCTALYZER_BGCOLOR, v)

}

func (m *MemoryMap) GetBGColor(index int) (uint8, uint8, uint8, uint8) {

	v := m.R(index, m.MEMBASE(index)+OCTALYZER_BGCOLOR)
	r := uint8(v & 0xff)
	g := uint8((v >> 8) & 0xff)
	b := uint8((v >> 16) & 0xff)
	a := uint8((v >> 24) & 0xff)

	return r, g, b, a
}

func (m *MemoryMap) SetPixelSize(v uint64) {

	m.W(0, OCTALYZER_HGR_SIZE, v)

}

func (m *MemoryMap) SetMapper(index int, f MapperFunc) {
	//
	m.Mappers[index] = f
}

func (m *MemoryMap) LockForVideo(b bool) {
	m.LockedVideo = b
}

func (m *MemoryMap) IsLockedForVideo() bool {
	return m.LockedVideo
}

func (m *MemoryMap) SetMemorySync(index int, rindex int, remint *client.FSClient) {
	// add a sync mechanism to processing stuff
	remint.RIndex = rindex
	m.RemoteSync[index%OCTALYZER_NUM_INTERPRETERS] = remint
}

func (m *MemoryMap) GetHUDLayerState(index int) uint64 {
	return m.ReadGlobal(index, m.MEMBASE(index)+OCTALYZER_HUD_ACTIVESTATES)
}

func (m *MemoryMap) SetHUDLayerState(index int, v uint64) {
	m.WriteGlobal(index, m.MEMBASE(index)+OCTALYZER_HUD_ACTIVESTATES, v)
}

func (m *MemoryMap) SetHUDLayerStateSilent(index int, v uint64) {
	m.WriteGlobalSilent(index, m.MEMBASE(index)+OCTALYZER_HUD_ACTIVESTATES, v)
}

func (m *MemoryMap) GetGFXLayerState(index int) uint64 {
	return m.ReadGlobal(index, m.MEMBASE(index)+OCTALYZER_GFX_ACTIVESTATES)
}

func (m *MemoryMap) SetGFXLayerState(index int, v uint64) {
	m.WriteGlobal(index, m.MEMBASE(index)+OCTALYZER_GFX_ACTIVESTATES, v)
}

func (m *MemoryMap) SetGFXLayerStateSilent(index int, v uint64) {
	m.WriteGlobalSilent(index, m.MEMBASE(index)+OCTALYZER_GFX_ACTIVESTATES, v)
}

func (m *MemoryMap) EnableLogTracking(index int, callback LoggerFunc) {

	m.Track[index%OCTALYZER_NUM_INTERPRETERS] = true
	m.TrackCallback[index%OCTALYZER_NUM_INTERPRETERS] = callback

}

func (m *MemoryMap) DisableLogTracking(index int) {

	m.Track[index%OCTALYZER_NUM_INTERPRETERS] = false
	m.TrackCallback[index%OCTALYZER_NUM_INTERPRETERS] = nil

}

func (m *MemoryMap) IntSetProfileState(index int, state uint64) {
	base := m.MEMBASE(index) + OCTALYZER_INTERPRETER_STATE_BASE
	m.WriteGlobal(index, base+5, state)
}

func (m *MemoryMap) IntGetProfileState(index int) uint64 {
	base := m.MEMBASE(index) + OCTALYZER_INTERPRETER_STATE_BASE
	return m.ReadGlobal(index, base+5)
}

func (m *MemoryMap) IntSetLayerForceState(index int, state uint64) {
	base := m.MEMBASE(index) + OCTALYZER_INTERPRETER_STATE_BASE
	m.WriteGlobal(index, base+6, state)
}

func (m *MemoryMap) IntGetLayerForceState(index int) uint64 {
	base := m.MEMBASE(index) + OCTALYZER_INTERPRETER_STATE_BASE
	return m.ReadGlobal(index, base+6)
}

func (m *MemoryMap) IntSetLogMode(index int, state uint64) {
	base := m.MEMBASE(index) + OCTALYZER_INTERPRETER_STATE_BASE
	m.WriteGlobal(index, base+7, state)
}

func (m *MemoryMap) IntGetLogMode(index int) uint64 {
	base := m.MEMBASE(index) + OCTALYZER_INTERPRETER_STATE_BASE
	return m.ReadGlobal(index, base+7)
}

func (m *MemoryMap) GetProfileName(index int) string {
	switch index {
	case 1:
		return "apple2e-en.yaml"
	}

	return "apple2e-en.yaml"
}

func (m *MemoryMap) SetWaveCallback(index int, callback WaveCallback) {

	m.AudioCallback[index%OCTALYZER_NUM_INTERPRETERS] = callback

}

func (m *MemoryMap) SetRestCallback(index int, callback RestCallback) {

	m.RestalgiaCallback[index%OCTALYZER_NUM_INTERPRETERS] = callback

}

func (m *MemoryMap) SetMusicCallback(index int, callback MusicCallback) {

	m.DigitalMusicCallback[index%OCTALYZER_NUM_INTERPRETERS] = callback

}

func (m *MemoryMap) DecodePackedAudio(bitcount int, buffer []uint64, ampscale float32) []float32 {

	//fmt.Printf("Decoding %d bits\n", bitcount)

	var fdata []float32

	fdata = make([]float32, 1)
	var bitnum int = 31
	var bindex int = 0
	var findex int
	var bitsprocessed int

	for bitsprocessed < bitcount {
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

// HandleAudio handles virtual DMA audio
func (m *MemoryMap) HandleAudio(index int, channel int) {
	var c int
	var rate int
	var bytepacked bool
	var directsent bool
	var indata []uint64

	switch channel {
	case 0:
		c = int(m.ReadGlobal(index, m.MEMBASE(index)+OCTALYZER_SPEAKER_SAMPLECOUNT))
		rate = int(m.ReadGlobal(index, m.MEMBASE(index)+OCTALYZER_SPEAKER_SAMPLERATE)) & 0xffff
		bytepacked = (int(m.ReadGlobal(index, m.MEMBASE(index)+OCTALYZER_SPEAKER_SAMPLERATE))&0x010000 != 0)
		directsent = (int(m.ReadGlobal(index, m.MEMBASE(index)+OCTALYZER_SPEAKER_SAMPLERATE))&0x020000 != 0)
		indata = m.BlockRead(index, m.MEMBASE(index)+OCTALYZER_SPEAKER_BUFFER, c)
	case 1:
		c = int(m.ReadGlobal(index, m.MEMBASE(index)+OCTALYZER_CASSETTE_SAMPLECOUNT))
		rate = int(m.ReadGlobal(index, m.MEMBASE(index)+OCTALYZER_CASSETTE_SAMPLERATE)) & 0xffff
		bytepacked = (int(m.ReadGlobal(index, m.MEMBASE(index)+OCTALYZER_CASSETTE_SAMPLERATE))&0x010000 != 0)
		directsent = (int(m.ReadGlobal(index, m.MEMBASE(index)+OCTALYZER_CASSETTE_SAMPLERATE))&0x020000 != 0)
		indata = m.BlockRead(index, m.MEMBASE(index)+OCTALYZER_CASSETTE_BUFFER, c)
	}

	if len(indata) == 0 {
		return
	}

	if m.spkdiff[index] == 0 {
		m.spkdiff[index] = -1
	}

	if m.AudioCallback[index] != nil && !directsent {
		go m.AudioCallback[index](index, channel, indata, rate, bytepacked)
	}

	switch channel {
	case 0:
		m.WriteGlobal(index, m.MEMBASE(index)+OCTALYZER_SPEAKER_SAMPLECOUNT, 0)
		m.WriteGlobal(index, m.MEMBASE(index)+OCTALYZER_SPEAKER_PLAYSTATE, 0)
	case 1:
		m.WriteGlobal(index, m.MEMBASE(index)+OCTALYZER_CASSETTE_SAMPLECOUNT, 0)
		m.WriteGlobal(index, m.MEMBASE(index)+OCTALYZER_CASSETTE_PLAYSTATE, 0)
	}
}

// HandleAudio handles virtual DMA audio
func (m *MemoryMap) HandleDigitalMusic(index int) {
	// m.logmutex[index].Lock()
	// defer m.logmutex[index].Unlock()

	c := int(m.ReadGlobal(index, m.MEMBASE(index)+OCTALYZER_DIGI_SAMPLECOUNT))
	rate := int(m.ReadGlobal(index, m.MEMBASE(index)+OCTALYZER_DIGI_SAMPLERATE)) & 0xffff
	channels := int(m.ReadGlobal(index, m.MEMBASE(index)+OCTALYZER_DIGI_CHANNELS)) & 0xffff
	bytepacked := (int(m.ReadGlobal(index, m.MEMBASE(index)+OCTALYZER_DIGI_SAMPLERATE))&0x010000 != 0)
	directsent := (int(m.ReadGlobal(index, m.MEMBASE(index)+OCTALYZER_DIGI_SAMPLERATE))&0x020000 != 0)
	indata := m.BlockRead(index, m.MEMBASE(index)+OCTALYZER_DIGI_BUFFER, c)

	if len(indata) == 0 {
		return
	}

	if m.spkdiff[index] == 0 {
		m.spkdiff[index] = -1
	}

	if m.DigitalMusicCallback[index] != nil && !directsent {
		go m.DigitalMusicCallback[index](index, indata, rate, channels, bytepacked)
	}

	m.WriteGlobal(index, m.MEMBASE(index)+OCTALYZER_DIGI_SAMPLECOUNT, 0)
	m.WriteGlobal(index, m.MEMBASE(index)+OCTALYZER_DIGI_PLAYSTATE, 0)
}

func (m *MemoryMap) DirectSendAudio(index int, channel int, indata []uint64, rate int) {
	if m.AudioCallback[index] != nil {
		m.AudioCallback[index](index, channel, indata, rate, false)
	}
}

func (m *MemoryMap) DirectSendRestalgiaCmd(index int, s string) {
	if m.RestalgiaCallback[index] != nil {
		m.RestalgiaCallback[index](index, s)
	}
}

func (m *MemoryMap) DirectSendDigitalMusic(index int, indata []uint64, rate int, channels int) {
	if m.DigitalMusicCallback[index] != nil {
		m.DigitalMusicCallback[index](index, indata, rate, channels, false)
	}
}

func (m *MemoryMap) DirectSendAudioPacked(index int, channel int, indata []uint64, rate int) {
	if m.AudioCallback[index] != nil {
		//fmt.Println("callback")
		m.AudioCallback[index](index, channel, indata, rate&0xffff, rate > 65535)
	} else {
		//fmt.Println("nil callback")
	}
}

func (m *MemoryMap) DirectSendDigitalMusicPacked(index int, indata []uint64, rate int, channels int) {
	if m.DigitalMusicCallback[index] != nil {
		//fmt.Println("callback")
		m.DigitalMusicCallback[index](index, indata, rate, channels, true)
	} else {
		//fmt.Println("nil callback")
	}
}

func (m *MemoryMap) RecordSendAudioPacked(index int, channel int, c int, indata []uint64, rate int, bytepacked bool) {
	// TODO: channel handle for handling cassette
	if m.Track[index] && (m.MemCapMode[index] == MEMCAP_RECORD || m.MemCapMode[index] == MEMCAP_CUSTOM) {
		// capture the raw audio packet here
		if m.CustomAudioLogger[index] != nil && channel == 0 {
			//fmt.Println("Custom audio logger called")
			go m.CustomAudioLogger[index](c, rate, bytepacked, indata)
		}
	}
}

func (m *MemoryMap) RecordSendAudioPackedF(index int, channel int, c int, indata []float32, rate int, bytepacked bool) {
	// TODO: channel handle for handling cassette
	if m.Track[index] && (m.MemCapMode[index] == MEMCAP_RECORD || m.MemCapMode[index] == MEMCAP_CUSTOM) {
		// capture the raw audio packet here
		if m.CustomAudioLogger[index] != nil && channel == 0 {
			var tmp = make([]uint64, len(indata))
			for i, v := range indata {
				tmp[i] = uint64(math.Float32bits(v))
			}
			//fmt.Println("Custom audio logger called")
			go m.CustomAudioLogger[index](c, rate, bytepacked, tmp)
		}
	}
}

func (m *MemoryMap) RecordSendDigitalMusicPacked(index int, c int, indata []uint64, rate int, bytepacked bool) {
	if m.Track[index] && (m.MemCapMode[index] == MEMCAP_RECORD || m.MemCapMode[index] == MEMCAP_CUSTOM) {
		// capture the raw audio packet here
		if m.CustomDigitalMusicLogger[index] != nil {
			//fmt.Println("Custom audio logger called")
			go m.CustomDigitalMusicLogger[index](c, rate, bytepacked, indata)
		}
	}
}

func (m *MemoryMap) Share(index int, sc ShareControl, port int) {

	if m.SlotShares[index] == nil {
		m.SlotShares[index] = NewShareService()
	}
	m.SlotShares[index].Start(sc, m, index, port)

}

func UintSlice2Float(u []uint64) []float32 {
	f := make([]float32, len(u))
	for i, v := range u {
		f[i] = math.Float32frombits(uint32(v))
	}
	return f
}

func UintSlice2FloatBP(u []uint64) []float32 {
	f := make([]float32, len(u))
	for i, v := range u {
		//f[i] = math.Float32frombits( uint32(v) )
		switch v {
		case 2:
			f[i] = 1
		case 1:
			f[i] = 0
		case 0:
			f[i] = -1
		}
	}
	return f
}
