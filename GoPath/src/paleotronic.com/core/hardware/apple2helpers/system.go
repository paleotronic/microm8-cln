package apple2helpers

/*
	Ancillary dialect specific functions
*/

import (
	"bytes"
	"errors"
	"math"
	"strings"
	"time"

	s8webclient "paleotronic.com/api"
	"paleotronic.com/octalyzer/bus"

	"github.com/mjibson/go-dsp/wav"
	"paleotronic.com/core/hardware/cpu"
	"paleotronic.com/core/hardware/cpu/mos6502"
	"paleotronic.com/core/hardware/cpu/z80"
	"paleotronic.com/core/hardware/restalgia"
	"paleotronic.com/core/hardware/servicebus"
	"paleotronic.com/core/hires"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/core/vduconst"
	"paleotronic.com/decoding" //"runtime"
	"paleotronic.com/files"
	"paleotronic.com/fmt"
	"paleotronic.com/log"
	"paleotronic.com/runestring"
	"paleotronic.com/utils"
)

//import "paleotronic.com/utils"

//import "paleotronic.com/octalyzer/clientperipherals"

type BSRState int

const (
	READBSR BSRState = iota
	WRITEBSR
	OFFBSR
	RDWRBSR
)

type SoftSwitchConfig struct {
	SoftSwitch_GRAPHICS  bool // (false) POKE -16304,0  (enable) POKE -16303,0 (disable)
	SoftSwitch_MIXED     bool // (true)  POKE -16301,0  (enable) POKE -16302,0 (disable)
	SoftSwitch_HIRES     bool // (false) POKE -16297,0  (enable) POKE -16298,0 (disable)
	SoftSwitch_PAGE2     bool // (false) POKE -16299,0  (enable) POKE -16300,0 (disable)
	SoftSwitch_DoubleRes bool
	// Memory related switches.
	SoftSwitch_80STORE    bool
	SoftSwitch_80COL      bool
	SoftSwitch_RAMRD      bool
	SoftSwitch_RAMWRT     bool
	SoftSwitch_SLOTC3ROM  bool
	SoftSwitch_INTCXROM   bool
	SoftSwitch_INTC8ROM   bool
	SoftSwitch_ALTZP      bool
	SoftSwitch_ALTCHARSET bool
	SoftSwitch_BSRBANK2   bool
	SoftSwitch_HRAMRD     bool
	SoftSwitch_HRAMWRT    bool
	SoftSwitch_SHR        bool
	SoftSwitch_SHR_LINEAR bool
	//
	BSR2 BSRState
	BSR1 BSRState

	WRITECOUNT int
}

var MouseKeys bool
var ColorFlip bool
var cpu6502 [memory.OCTALYZER_NUM_INTERPRETERS]*mos6502.Core6502
var cpuZ80 [memory.OCTALYZER_NUM_INTERPRETERS]*z80.CoreZ80
var textselect [memory.OCTALYZER_NUM_INTERPRETERS]string
var softswitch [memory.OCTALYZER_NUM_INTERPRETERS]uint64
var layerstate [memory.OCTALYZER_NUM_INTERPRETERS]*bytes.Buffer

// var turtlepos map[int]*Turtle
var LastX, LastY int

func init() {
	//cpu6502 = make(map[int]*mos6502.Core6502)
	//textselect = make(map[int]string)
	for i := 0; i < memory.OCTALYZER_NUM_INTERPRETERS; i++ {
		textselect[i] = "TEXT"
	}
	//~ softswitch = make(map[int]uint64)
	//~ layerstate = make(map[int]*bytes.Buffer)
}

type Buzzer interface {
	Click()
}

func SwitchToDialect(ent interfaces.Interpretable, d string) {
	// exec := ent.GetState() == types.EXEC6502 || ent.GetState() == types.DIRECTEXEC6502
	// if exec || ent.GetDialect().GetShortName() != d {
	// command := "text"
	// switch d {
	// case "logo":
	// 	command = "cs"
	// }
	path := "/local/"
	if s8webclient.CONN.IsAuthenticated() {
		path = "/"
	}
	CommandResetOverride(
		ent,
		"!"+d,
		path,
		" ",
	)
	ent.Halt()
	//}
}

func SuperHiresEnable(ent interfaces.Interpretable, use640 bool) {
	ent.SetMemory(0xc029, 128|64)
	shr := GETGFX(ent, "SHR1")
	if shr == nil {
		panic("SHR1 is nil")
	}
	control := shr.HControl.(*hires.SuperHiResBuffer)
	if use640 {
		control.Init640()
	} else {
		control.Init320()
	}
}

func SuperHiresDisable(ent interfaces.Interpretable, use640 bool) {
	ent.SetMemory(0xc029, 0)
}

func CommandResetOverride(ent interfaces.Interpretable, dialect string, path string, command string) {
	// s := &settings.RState{
	// 	WorkDir: path,
	// 	Dialect: dialect,
	// 	Command: command,
	// }
	settings.VMLaunch[ent.GetMemIndex()] = &settings.VMLauncherConfig{
		WorkingDir: path,
		Dialect:    dialect,
		RunCommand: command,
	}
	settings.PureBootVolume[ent.GetMemIndex()] = ""
	settings.PureBootVolume2[ent.GetMemIndex()] = ""
	ent.GetMemoryMap().IntSetSlotRestart(ent.GetMemIndex(), true)
}

func SaveVSTATE(ent interfaces.Interpretable) {
	//	 ent.WaitForLayers()
	//	 ent.ReadLayersFromMemory()
	//     data := bytes.NewBuffer([]byte(nil))
	//     ent.FreezeStreamLayers( data )
	//     layerstate[ ent.GetMemIndex() ] = data
	softswitch[ent.GetMemIndex()] = ent.GetMemory(0xfcff)
	nf := &SoftSwitchConfig{}
	nf.FromUint(softswitch[ent.GetMemIndex()])
	//fmt.Printf("Switch state on save is [%v]\n", *nf)
}

func LoadVSTATE(ent interfaces.Interpretable) {
	//	 ent.WaitForLayers()
	//	 data := layerstate[ ent.GetMemIndex() ]
	//     ent.ThawStreamLayers( data )
	//     ent.WriteLayersToMemory()
	u := softswitch[ent.GetMemIndex()]

	/*
	   Poke -16297,0 Hi-res
	   Poke -16298,0 Lo-res
	   Poke -16299,0 Switch from high-res page 1 to page 2
	   Poke -16300,0 Switch from high-res page 2 to page 1
	   Poke -16301,0 Allows graphics and 4 lines of text
	   Poke -16302,0 Full screen graphics - no text
	   Poke -16303,0 Shows text screen
	   Poke -16304,0 Shows graphics screen
	*/

	nf := &SoftSwitchConfig{}
	nf.FromUint(u)

	if nf.SoftSwitch_HIRES {
		ent.SetMemory(65536-16297, 0)
	} else {
		ent.SetMemory(65536-16298, 0)
	}

	if nf.SoftSwitch_PAGE2 {
		ent.SetMemory(65536-16299, 0)
	} else {
		ent.SetMemory(65536-16300, 0)
	}

	if nf.SoftSwitch_MIXED {
		ent.SetMemory(65536-16301, 0)
	} else {
		ent.SetMemory(65536-16302, 0)
	}

	if nf.SoftSwitch_GRAPHICS {
		ent.SetMemory(65536-16304, 0)
	} else {
		ent.SetMemory(65536-16303, 0)
	}

	//fmt.Printf("Switch state on load is [%v]\n", *nf)

}

func IsVectorLayerEnabled(ent interfaces.Interpretable) bool {
	vec := GETGFX(ent, "VCTR")
	return vec.GetActive()
}

func ChangeVisibilityGFX(ent interfaces.Interpretable, layername string, visible bool) {
	vec := GETGFX(ent, layername)
	vec.SetActive(visible)
}

func VectorLayerEnable(ent interfaces.Interpretable, b bool, split bool) {

	if settings.DisableMetaMode[ent.GetMemIndex()] {
		return
	}

	vec := GETGFX(ent, "VCTR")

	ent.WaitForLayers()

	ent.ReadLayersFromMemory()

	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt == nil {
		panic("Expected layer id TEXT not found")
	}

	if b {
		vec.SetActive(true)
		if split {
			txt.SetActive(true)
			txt.SetBounds(0, 40, 79, 47)
		} else {
			txt.SetActive(false)
			txt.SetBounds(80, 48, 80, 48)
		}
	} else {
		vec.SetActive(false)
		txt.SetActive(true)
		txt.SetBounds(0, 0, 79, 47)
	}

	ent.WriteLayersToMemory()
}

func VectorLayerEnableFullText(ent interfaces.Interpretable, b bool) {

	vec := GETGFX(ent, "VCTR")

	ent.WaitForLayers()

	ent.ReadLayersFromMemory()

	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt == nil {
		panic("Expected layer id TEXT not found")
	}

	if b {
		vec.SetActive(true)
		// vec.SetWidth(uint16(width))
		// vec.SetHeight(uint16(height))
		//vec.SetBounds(0, 0, uint16(width-1), uint16(height-1))
	}

	txt.SetActive(true)
	txt.SetBounds(0, 0, 79, 47)

}

func VectorLayerEnableCustom(ent interfaces.Interpretable, b bool, split bool, width, height int) {

	vec := GETGFX(ent, "VCTR")

	ent.WaitForLayers()

	ent.ReadLayersFromMemory()

	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt == nil {
		panic("Expected layer id TEXT not found")
	}

	if b {
		vec.SetActive(true)
		vec.SetWidth(uint16(width))
		vec.SetHeight(uint16(height))
		vec.SetBounds(0, 0, uint16(width-1), uint16(height-1))
		if split {
			txt.SetActive(true)
			txt.SetBounds(0, 40, 79, 47)
		} else {
			txt.SetActive(false)
			txt.SetBounds(80, 48, 80, 48)
		}
	} else {
		vec.SetActive(false)
		txt.SetActive(true)
		txt.SetBounds(0, 0, 79, 47)
	}

	ent.WriteLayersToMemory()
}

func CubeLayerEnable(ent interfaces.Interpretable, b bool, split bool) {

	vec := GETGFX(ent, "CUBE")

	ent.WaitForLayers()

	ent.ReadLayersFromMemory()

	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt == nil {
		panic("Expected layer id TEXT not found")
	}

	if b {
		vec.SetActive(true)
		if split {
			txt.SetActive(true)
			txt.SetBounds(0, 40, 79, 47)
		} else {
			txt.SetActive(false)
			txt.SetBounds(80, 48, 80, 48)
		}
	} else {
		vec.SetActive(false)
		txt.SetActive(true)
		txt.SetBounds(0, 0, 79, 47)
		control := types.NewOrbitController(ent.GetMemoryMap(), ent.GetMemIndex(), 1)
		control.ResetALL()
		ent.GetMemoryMap().SetCameraConfigure(ent.GetMemIndex(), 0)
	}

	ent.WriteLayersToMemory()
}

func CubeLayerEnableCustom(ent interfaces.Interpretable, b bool, split bool, width, height int) {

	vec := GETGFX(ent, "CUBE")

	ent.WaitForLayers()

	ent.ReadLayersFromMemory()

	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt == nil {
		panic("Expected layer id TEXT not found")
	}

	if b {
		vec.SetActive(true)
		vec.SetWidth(uint16(width))
		vec.SetHeight(uint16(height))
		vec.SetBounds(0, 0, uint16(width-1), uint16(height-1))
		if split {
			txt.SetActive(true)
			txt.SetBounds(0, 40, 79, 47)
		} else {
			txt.SetActive(false)
			txt.SetBounds(80, 48, 80, 48)
		}
		vec.SetDirty(true)
		mm := ent.GetMemoryMap()
		mm.SetCameraConfigure(ent.GetMemIndex(), 1)
	} else {
		vec.SetActive(false)
		txt.SetActive(true)
		txt.SetBounds(0, 0, 79, 47)
		mm := ent.GetMemoryMap()
		mm.SetCameraConfigure(ent.GetMemIndex(), 0)
	}

	ent.WriteLayersToMemory()
}

func XGRLayerEnableCustom(ent interfaces.Interpretable, page string, b bool, split bool, width, height int) {

	xgr := GETGFX(ent, page)
	if xgr == nil {
		return
	}

	ent.WaitForLayers()

	ent.ReadLayersFromMemory()

	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt == nil {
		panic("Expected layer id " + page + " not found")
	}

	if b {
		xgr.SetActive(true)
		xgr.SetWidth(uint16(width))
		xgr.SetHeight(uint16(height))
		xgr.SetBounds(0, 0, uint16(width-1), uint16(height-1))
		if split {
			txt.SetActive(true)
			txt.SetBounds(0, 40, 79, 47)
		} else {
			txt.SetActive(false)
			txt.SetBounds(80, 48, 80, 48)
		}
	} else {
		xgr.SetActive(false)
		txt.SetActive(true)
		txt.SetBounds(0, 0, 79, 47)
	}

	ent.WriteLayersToMemory()
}

func MonitorPanel(ent interfaces.Interpretable, on bool) {
	//~ ent.ReadLayersFromMemory()
	if on {
		//~ SaveVSTATE(ent)
		ntxt := GETHUD(ent, "MONI")
		if ntxt == nil {
			panic("HUD Layer MONI not found")
		}
		ntxt.SetActive(true)

		if strings.HasPrefix(settings.SpecFile[ent.GetMemIndex()], "apple2") {
			txt := GETHUD(ent, "TEXT")
			if txt == nil {
				panic("HUD Layer TEXT not found")
			}
			txt.SetActive(false)

			txt2 := GETHUD(ent, "TXT2")
			if txt2 == nil {
				panic("HUD Layer TEXT not found")
			}
			txt2.SetActive(false)
		}

		textselect[ent.GetMemIndex()] = "MONI"
		//~ ent.WriteLayersToMemory()
	} else {
		ntxt := GETHUD(ent, "MONI")
		if ntxt == nil {
			panic("HUD Layer MONI not found")
		}
		ntxt.SetActive(false)
		if strings.HasPrefix(settings.SpecFile[ent.GetMemIndex()], "apple2") {
			txt := GETHUD(ent, "TEXT")
			if txt == nil {
				panic("HUD Layer TEXT not found")
			}
			txt.SetActive(true)
		}
		textselect[ent.GetMemIndex()] = "TEXT"
		//~ ent.WriteLayersToMemory()
		//~ LoadVSTATE(ent)
	}
}

func GetSpriteController(e interfaces.Interpretable) *types.SpriteController {
	return types.NewSpriteController(
		e.GetMemIndex(),
		e.GetMemoryMap(),
		memory.MICROM8_SPRITE_CONTROL_BASE,
	)
}

func SpriteReset(ent interfaces.Interpretable) {
	GetSpriteController(ent).Reset()
}

func OSDPanel(ent interfaces.Interpretable, on bool) {

	if settings.MenuActive {
		return
	}

	//~ ent.ReadLayersFromMemory()
	if on {
		//~ SaveVSTATE(ent)
		ntxt := GETHUD(ent, "OOSD")
		if ntxt == nil {
			panic("HUD Layer MONI not found")
		}
		ntxt.SetActive(true)
		ntxt.SetSubFormat(types.LSF_FREEFORM)
		//textselect[ent.GetMemIndex()] = "OOSD"
	} else {
		ntxt := GETHUD(ent, "OOSD")
		if ntxt == nil {
			panic("HUD Layer MONI not found")
		}
		ntxt.SetActive(false)
		//textselect[ent.GetMemIndex()] = "TEXT"
	}
}

func OSDShowProgress(ent interfaces.Interpretable, label string, percent float32) {

	if settings.MenuActive {
		return
	}

	fmt.Println("osd")
	OSDPanel(ent, true)
	osd := GETHUD(ent, "OOSD")
	if osd == nil {
		return
	}
	c := osd.Control
	c.Font = types.W_NORMAL_H_NORMAL
	c.SetWindow(0, 0, 79, 47)
	if settings.HighContrastUI {
		c.FGColor = 15
	} else {
		c.FGColor = 13
	}
	c.BGColor = 0
	c.ClearScreen()
	c.GotoXY(0, 0)
	if settings.HighContrastUI {
		c.FGColor = 0
		c.BGColor = 15
	} else {
		c.FGColor = 13
		c.BGColor = 2
	}
	width := 40 - len(label)
	fillwidth := int(float32(width) * percent)
	emptywidth := width - fillwidth
	//c.BGColor = 2
	c.PutStr(label)
	for i := 0; i < fillwidth; i++ {
		c.Put(1150)
	}
	for i := 0; i < emptywidth; i++ {
		c.Put(1027)
	}
	c.HideCursor()
	c.FullRefresh()
	fmt.Println("Done percent")
}

func OSDShow(ent interfaces.Interpretable, label string) {

	if settings.MenuActive {
		return
	}

	fmt.Printf("OSD Shown on slot %d\n", ent.GetMemIndex())

	fmt.Println("osd")
	OSDPanel(ent, true)
	osd := GETHUD(ent, "OOSD")
	if osd == nil {
		return
	}
	osd.SetSubFormat(types.LSF_FREEFORM)
	c := osd.Control
	c.Font = types.W_NORMAL_H_NORMAL
	c.SetWindow(0, 0, 79, 47)
	if settings.HighContrastUI {
		c.FGColor = 15
	} else {
		c.FGColor = 13
	}
	c.BGColor = 0
	c.ClearScreen()
	c.GotoXY(0, 0)
	if settings.HighContrastUI {
		c.FGColor = 0
		c.BGColor = 15
	} else {
		c.FGColor = 13
		c.BGColor = 2
	}
	c.PutStr(label)
	c.HideCursor()
	c.FGColor = 15
	c.FullRefresh()
	fmt.Println("Done percent")
	time.AfterFunc(3*time.Second, func() {
		c.BGColor = 0
		c.Shade = 0
		c.ClearScreen()
		OSDPanel(ent, false)
		fmt.Printf("OSD Cleared on slot %d\n", ent.GetMemIndex())
	})
}
func OSDShowBlink(ent interfaces.Interpretable, label string, interval int64) {

	if settings.MenuActive {
		return
	}

	fmt.Println("osd")

	vis := time.Now().UnixNano()%interval >= (interval / 2)

	OSDPanel(ent, vis)
	osd := GETHUD(ent, "OOSD")
	if osd == nil {
		return
	}
	c := osd.Control
	c.Font = types.W_NORMAL_H_NORMAL
	c.SetWindow(0, 0, 79, 47)
	if settings.HighContrastUI {
		c.FGColor = 15
	} else {
		c.FGColor = 13
	}
	c.BGColor = 0
	c.ClearScreen()
	c.GotoXY(0, 0)
	if settings.HighContrastUI {
		c.FGColor = 0
		c.BGColor = 15
	} else {
		c.FGColor = 13
		c.BGColor = 2
	}
	c.PutStr(label)

	c.HideCursor()
	c.FullRefresh()
	fmt.Println("Done percent")
}

func SelectHUD(ent interfaces.Interpretable, name string) {

	current := textselect[ent.GetMemIndex()]
	txt, ok := ent.GetHUDLayerByID(current)
	if !ok {
		panic("HUD Layer TEXT not found")
	}
	txt.SetActive(false)
	ntxt, ok := ent.GetHUDLayerByID(name)
	if !ok {
		panic("HUD Layer TEXT not found")
	}
	ntxt.SetActive(true)
	textselect[ent.GetMemIndex()] = name
	ent.WriteLayersToMemory()
}

func TrashCPU(ent interfaces.Interpretable) {

	//debug.PrintStack()

	if cpu6502[ent.GetMemIndex()] != nil {
		servicebus.Unsubscribe(
			ent.GetMemIndex(),
			cpu6502[ent.GetMemIndex()],
		)
	}
	cpu6502[ent.GetMemIndex()] = nil
	_ = GetCPU(ent)

	// Ensure we trash any Z80 too
	TrashZ80CPU(ent)
}

func GetCPU(ent interfaces.Interpretable) *mos6502.Core6502 {
	if cpu6502[ent.GetMemIndex()] == nil {
		//fmt.Println("Fresh cpu")
		//if ent.GetMemoryMap().IntGetProcessorState(ent.GetMemIndex()) == 128 {
		switch settings.CPUModel[ent.GetMemIndex()] {
		case "65C02":
			cpu6502[ent.GetMemIndex()] = mos6502.NewCore65C02(ent, 0, 0, 0, 0xffee, 0, 0x1f8, ent)
		case "6502":
			cpu6502[ent.GetMemIndex()] = mos6502.NewCore6502(ent, 0, 0, 0, 0xffee, 0, 0x1f8, ent)
		default:
			cpu6502[ent.GetMemIndex()] = mos6502.NewCore65C02(ent, 0, 0, 0, 0xffee, 0, 0x1f8, ent)
		}
		cpu6502[ent.GetMemIndex()].BaseSpeed = int64(settings.CPUClock[ent.GetMemIndex()])

		// patch cpu handlers

		if !settings.PureBoot(ent.GetMemIndex()) {
			cpu6502[ent.GetMemIndex()].Inject(0x36, []uint64{0xf0, 0xfd})
			cpu6502[ent.GetMemIndex()].Inject(0xaa52, []uint64{0x02})
			cpu6502[ent.GetMemIndex()].Inject(0x3d0, []uint64{0x4C, 0xBF, 0x9D, 0x4C, 0x84, 0x9D})
			cpu6502[ent.GetMemIndex()].Inject(0x3f2, []uint64{0xBF, 0x9D, 0x38, 0x4C, 0x58, 0xFF, 0x4C, 0x65, 0xFF, 0x4C, 0x65, 0xFF, 0x65, 0xFF})
		}
		cpu6502[ent.GetMemIndex()].MemIndex = ent.GetMemIndex()

		cpu6502[ent.GetMemIndex()].SetFlag(mos6502.F_R, true)
		cpu6502[ent.GetMemIndex()].SetFlag(mos6502.F_B, true)
		cpu6502[ent.GetMemIndex()].SetFlag(mos6502.F_I, true)
	}
	return cpu6502[ent.GetMemIndex()]
}

func TrashZ80CPU(ent interfaces.Interpretable) {
	// if cpuZ80[ent.GetMemIndex()] != nil {
	// 	servicebus.Unsubscribe(
	// 		ent.GetMemIndex(),
	// 		cpuZ80[ent.GetMemIndex()],
	// 	)
	// }
	cpuZ80[ent.GetMemIndex()] = nil
	_ = GetZ80CPU(ent)
}

func GetZ80CPU(ent interfaces.Interpretable) *z80.CoreZ80 {
	if cpuZ80[ent.GetMemIndex()] == nil {
		cpuZ80[ent.GetMemIndex()] = z80.NewCoreZ80(ent, nil, nil)
	}
	return cpuZ80[ent.GetMemIndex()]
}

/* TEXT returns the TextBuffer attached to the TEXT memory layer */
func TEXT(ent interfaces.Interpretable) *types.TextBuffer {
	txt, ok := ent.GetHUDLayerByID(textselect[ent.GetMemIndex()])
	if !ok {
		panic("HUD Layer " + textselect[ent.GetMemIndex()] + " not found")
	}
	if txt.Control == nil {
		panic("HUD Layer " + textselect[ent.GetMemIndex()] + " missing control interface")
	}
	return txt.Control
}

/* VECTOR returns the VectorBuffer attached to the VECTOR memory layer */
func VECTOR(ent interfaces.Interpretable) *types.VectorBuffer {
	txt, ok := ent.GetGFXLayerByID("VCTR")
	if !ok {
		panic("GFX Layer VCTR not found")
	}
	if txt.VControl == nil {
		panic("GFX Layer VCTR missing control interface")
	}
	return txt.VControl
}

/* CUBE returns the VectorBuffer attached to the VECTOR memory layer */
func CUBE(ent interfaces.Interpretable) *types.CubeScreen {
	txt, ok := ent.GetGFXLayerByID("CUBE")
	if !ok {
		panic("GFX Layer CUBE not found")
	}
	if txt.CubeControl == nil {
		txt.CubeControl = types.NewCubeScreen(
			txt.GetBase(),
			0x10000,
			txt.Mm.GetHintedMemorySlice(txt.Index, txt.GetID()),
		)
	}
	return txt.CubeControl
}

func LOGR(ent interfaces.Interpretable) *types.LayerSpecMapped {
	txt, ok := ent.GetGFXLayerByID("LOGR")
	if !ok {
		panic("GFX Layer LOGR not found")
	}
	return txt
}

func LOGR80(ent interfaces.Interpretable) *types.LayerSpecMapped {
	txt, ok := ent.GetGFXLayerByID("GR80")
	if !ok {
		panic("GFX Layer GR80 not found")
	}
	return txt
}

func TEXTLAYER(ent interfaces.Interpretable) *types.LayerSpecMapped {
	txt, ok := ent.GetHUDLayerByID(textselect[ent.GetMemIndex()])
	if !ok {
		panic("HUD Layer TEXT not found")
	}
	return txt
}

func GetSelectedHUD(ent interfaces.Interpretable) *types.LayerSpecMapped {
	return GETHUD(ent, textselect[ent.GetMemIndex()])
}

func GETHUD(ent interfaces.Interpretable, name string) *types.LayerSpecMapped {
	//fmt.Println("hud layer name =", name)
	txt, ok := ent.GetHUDLayerByID(name)
	if !ok {
		fmt.Printf("HUD Layer %s not found in vm %d", name, ent.GetMemIndex())
		//panic("HUD Layer " + name + " not found")
	}
	return txt
}

func GETGFX(ent interfaces.Interpretable, name string) *types.LayerSpecMapped {
	txt, ok := ent.GetGFXLayerByID(name)
	if !ok {
		//panic("GFX Layer " + name + " not found")
		fmt.Printf("GFX Layer %s not found in vm %d", name, ent.GetMemIndex())
	}
	return txt
}

func GFXDisable(ent interfaces.Interpretable) {
	layers := ent.GetGFXLayerSet()
	for _, l := range layers {
		if l != nil && l.GetActive() {
			l.SetActive(false)
		}
	}
}

func SetColorOffset(ent interfaces.Interpretable, layer string, c int, o int) {
	//fmt.Printf("SetColorOffset(%s, %d, %d)\n", layer, c, o)
	ent.ReadLayersFromMemory()
	l, ok := ent.GetGFXLayerByID(layer)
	if !ok {
		return
	}
	col := l.GetPaletteColor(c % l.GetPaletteSize())
	if col != nil {
		col.Offset = int8(o)
		l.SetPaletteColor(c%l.GetPaletteSize(), col)
		l.SetRefresh(true)
	} else {
		//fmt.Printf("Color index %d yields a NULL color record!\n", c)
	}
}

func SetColorDepth(ent interfaces.Interpretable, layer string, c int, o int) {
	//fmt.Printf("SetColorOffset(%s, %d, %d)\n", layer, c, o)
	ent.ReadLayersFromMemory()
	l, ok := ent.GetGFXLayerByID(layer)
	if !ok {
		return
	}
	col := l.GetPaletteColor(c % l.GetPaletteSize())
	if col != nil {
		col.Depth = uint8(o)
		l.SetPaletteColor(c%l.GetPaletteSize(), col)
		l.SetRefresh(true)
	} else {
		//fmt.Printf("Color index %d yields a NULL color record!\n", c)
	}
}

func TextMode(caller interfaces.Interpretable) {
	cy := GetCursorY(caller)

	if GetRows(caller) == 0 {
		cy = 23
	}

	caller.SetMemory(65536-16303, 0)
	caller.SetMemory(65536-16300, 0)
	caller.SetMemory(65536-16301, 0)
	caller.SetMemory(65536-16298, 0)

	// turn off XGR modes if needed
	XGRLayerEnableCustom(caller, "XGR1", false, true, 280, 160)
	XGRLayerEnableCustom(caller, "XGR2", false, false, 280, 192)

	SetRealWindow(caller, 0, 0, 79, 47)

	//PutStr(caller, "\r") // horiz home cursor
	cx := GetCursorX(caller)
	Gotoxy(caller, cx, cy)
	Attribute(caller, types.VA_NORMAL)
}

func SetColorRGBA(ent interfaces.Interpretable, layer string, c int, r, g, b, a int) {
	ent.ReadLayersFromMemory()
	l, ok := ent.GetGFXLayerByID(layer)
	if !ok {
		return
	}
	col := l.GetPaletteColor(c % l.GetPaletteSize())
	if col != nil {
		col.Red = uint8(r)
		col.Green = uint8(g)
		col.Blue = uint8(b)
		col.Alpha = uint8(a)
		l.SetPaletteColor(c%l.GetPaletteSize(), col)
		l.SetRefresh(true)
	} else {
		//fmt.Printf("Color index %d yields a NULL color record!\n", c)
	}
}

func SetTextColorRGBA(ent interfaces.Interpretable, layer string, c int, r, g, b, a int) {
	ent.ReadLayersFromMemory()
	l, ok := ent.GetHUDLayerByID(layer)
	if !ok {
		return
	}
	col := l.GetPaletteColor(c % l.GetPaletteSize())
	if col != nil {
		//fmt.Printf("Update %s color %d to #%.2d%.2d%.2d\n", layer, c, r, g, b, a)
		col.Red = uint8(r)
		col.Green = uint8(g)
		col.Blue = uint8(b)
		col.Alpha = uint8(a)
		l.SetPaletteColor(c%l.GetPaletteSize(), col)
		l.SetRefresh(true)
	} else {
		//fmt.Printf("Color index %d yields a NULL color record!\n", c)
	}
}

func ResetColorOffsets(ent interfaces.Interpretable, layer string) {
	ent.ReadLayersFromMemory()
	l, ok := ent.GetGFXLayerByID(layer)
	if !ok {
		return
	}
	for i := 0; i < l.GetPaletteSize(); i++ {
		col := l.GetPaletteColor(i)
		if col != nil {
			col.Offset = 0
			l.SetPaletteColor(i, col)
			l.SetRefresh(true)
		} else {
			//fmt.Printf("Color index %d yields a NULL color record!\n", i)
		}
	}
}

func ResetColorDepths(ent interfaces.Interpretable, layer string) {
	ent.ReadLayersFromMemory()
	l, ok := ent.GetGFXLayerByID(layer)
	if !ok {
		return
	}
	for i := 0; i < l.GetPaletteSize(); i++ {
		col := l.GetPaletteColor(i)
		if col != nil {
			col.Depth = 0
			l.SetPaletteColor(i, col)
			l.SetRefresh(true)
		} else {
			//fmt.Printf("Color index %d yields a NULL color record!\n", i)
		}
	}
}

func SetActiveLayers(ent interfaces.Interpretable, gfx map[string]bool, hud map[string]bool) {

	for name, enabled := range gfx {
		l := GETGFX(ent, name)
		if l != nil {
			l.SetActive(enabled)
		}
	}

	for name, enabled := range hud {
		l := GETHUD(ent, name)
		if l != nil {
			l.SetActive(enabled)
		}
	}

}

func GetActiveLayers(ent interfaces.Interpretable) (map[string]bool, map[string]bool) {
	layers := ent.GetGFXLayerSet()
	out := make(map[string]bool)
	for _, l := range layers {
		if l == nil {
			continue
		}
		mode := l.GetID()
		out[mode] = l.GetActive()
	}
	layers = ent.GetHUDLayerSet()
	outhud := make(map[string]bool)
	for _, l := range layers {
		if l == nil {
			continue
		}
		mode := l.GetID()
		outhud[mode] = l.GetActive()
	}
	return out, outhud
}

func GetActiveVideoModes(ent interfaces.Interpretable) []string {
	layers := ent.GetGFXLayerSet()
	out := make([]string, 0)
	for _, l := range layers {
		if l == nil {
			continue
		}
		if !l.GetActive() {
			continue
		}
		mode := l.GetID()
		out = append(out, l.GetID())
		// make sure paired modes are updated
		switch mode {
		case "HGR1":
			out = append(out, "HGR2")
		case "HGR2":
			out = append(out, "HGR1")
		case "XGR1":
			out = append(out, "XGR2")
		case "XGR2":
			out = append(out, "XGR1")
		case "DHR1":
			out = append(out, "DHR2")
		case "DHR2":
			out = append(out, "DHR1")
		}
	}
	return out
}

func GetActiveTextModes(ent interfaces.Interpretable) []string {
	layers := ent.GetHUDLayerSet()
	out := make([]string, 0)
	for _, l := range layers {
		if l == nil {
			continue
		}
		if !l.GetActive() {
			continue
		}
		mode := l.GetID()
		out = append(out, mode)
	}
	return out
}

func Put(ent interfaces.Interpretable, ch rune) {

	switch ch {
	case 10:
		TEXT(ent).LF()
	case 13:
		TEXT(ent).CR()
	default:
		TEXT(ent).Put(ch)
	}
}

func TextAddWindow(ent interfaces.Interpretable, s string, sx, sy, ex, ey int) {

	TEXT(ent).AddNamedWindow(s, sx, sy, ex, ey)

}

func TextDrawBox(ent interfaces.Interpretable, x, y, w, h int, content string, shadow, window bool) {

	TEXT(ent).DrawTextBox(x, y, w, h, content, shadow, window)

}

func TextUseWindow(ent interfaces.Interpretable, s string) {

	TEXT(ent).SetNamedWindow(s)

}

func TextPushCursor(ent interfaces.Interpretable) {

	TEXT(ent).PushCursor()

}

func TextPopCursor(ent interfaces.Interpretable) {

	TEXT(ent).PopCursor()

}

func TextHideCursor(ent interfaces.Interpretable) {

	TEXT(ent).HideCursor()

}

func TextSaveScreen(ent interfaces.Interpretable) {

	TEXT(ent).SaveState()

}

func TextRestoreScreen(ent interfaces.Interpretable) {

	TEXT(ent).RestoreState()

}

func TextShowCursor(ent interfaces.Interpretable) {

	TEXT(ent).ShowCursor()

}

func UseAltChars(ent interfaces.Interpretable, b bool) {

	if b {
		TEXT(ent).AltChars()
	} else {
		TEXT(ent).NormalChars()
	}

}

func PutStr(ent interfaces.Interpretable, s string) {

	for _, ch := range s {
		RealPut(ent, ch)
		//if delay > 0 {
		//   time.Sleep(time.Duration(delay)*time.Microsecond)
		//}
	}
}

func Clearscreen(ent interfaces.Interpretable) {
	txt := TEXT(ent)
	txt.ClearScreenWindow()
	txt.CX, txt.CY = txt.SX, txt.SY
}

func ClearToEOL(ent interfaces.Interpretable) {
	txt := TEXT(ent)
	txt.ClearToEOLWindow()
}

func ClearToBottom(ent interfaces.Interpretable) {
	txt := TEXT(ent)
	txt.ClearToBottomWindow()
}

func Gotoxy(ent interfaces.Interpretable, x, y int) {
	txt := TEXT(ent)
	txt.GotoXY(x, y)
}

func Home(ent interfaces.Interpretable) {
	txt := TEXT(ent)
	txt.GotoXY(txt.SX, txt.SY)
}

func SetFGColor(ent interfaces.Interpretable, c uint64) {
	txt := TEXT(ent)
	txt.FGColor = c
}

func SetFGColorXY(ent interfaces.Interpretable, x, y int, c uint64) {
	txt := TEXT(ent)
	txt.SetFGColorXY(x, y, c)
}

func SetBGColorXY(ent interfaces.Interpretable, x, y int, c uint64) {
	txt := TEXT(ent)
	txt.SetBGColorXY(x, y, c)
}

func SetTextSize(ent interfaces.Interpretable, c uint64) {
	txt := TEXT(ent)
	txt.Font = types.TextSize(c)
}

func SetBGColor(ent interfaces.Interpretable, c uint64) {
	txt := TEXT(ent)
	txt.BGColor = c
}

func SetShade(ent interfaces.Interpretable, c uint64) {
	txt := TEXT(ent)
	txt.Shade = c
}

func GetFGColor(ent interfaces.Interpretable) uint64 {
	txt := TEXT(ent)
	return txt.FGColor
}

func GetBGColor(ent interfaces.Interpretable) uint64 {
	txt := TEXT(ent)
	return txt.BGColor
}

func SetTextSplit(ent interfaces.Interpretable, lines int) {
	ent.ReadLayersFromMemory()
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	txt.SetBounds(0, uint16(48-lines), 79, 47)
	txt.Control.SY = 48 - lines
	txt.Control.SX = 0
	txt.Control.EX = 79
	txt.Control.EY = 47
	ent.WriteLayersToMemory()
}

func SetGRSplit(ent interfaces.Interpretable, lines int) {
	ent.ReadLayersFromMemory()
	gr := GETGFX(ent, "LOGR")
	gr.SetBounds(0, 0, 39, uint16(47-lines))
	ent.WriteLayersToMemory()
}

func SetHGRSplit(ent interfaces.Interpretable, lines int) {
	ent.ReadLayersFromMemory()
	hgr1 := GETGFX(ent, "HGR1")
	hgr1.SetBounds(0, 0, 279, uint16(191-lines*4))
	hgr2 := GETGFX(ent, "HGR2")
	hgr2.SetBounds(0, 0, 279, uint16(191-lines*4))
	ent.WriteLayersToMemory()
}

func SetGFXLayerOff(ent interfaces.Interpretable, layer string) {
	ent.ReadLayersFromMemory()
	hgr1 := GETGFX(ent, layer)
	hgr1.SetActive(false)
	ent.WriteLayersToMemory()
}

func SetGFXLayerOn(ent interfaces.Interpretable, layer string) {
	ent.ReadLayersFromMemory()
	hgr1 := GETGFX(ent, layer)
	hgr1.SetActive(true)
	ent.WriteLayersToMemory()
}

func SetHUDLayerOff(ent interfaces.Interpretable, layer string) {
	ent.ReadLayersFromMemory()
	hgr1 := GETHUD(ent, layer)
	hgr1.SetActive(false)
	ent.WriteLayersToMemory()
}

func SetHUDLayerOn(ent interfaces.Interpretable, layer string) {
	ent.ReadLayersFromMemory()
	hgr1 := GETHUD(ent, layer)
	hgr1.SetActive(true)
	ent.WriteLayersToMemory()
}

func GFXEnable(ent interfaces.Interpretable) {
	ent.ReadLayersFromMemory()
	ent.WriteLayersToMemory()
}

func SetDisplayPage(ent interfaces.Interpretable, page int) {
	ent.ReadLayersFromMemory()
	switch page % 2 {
	case 0:
		hgr1 := GETGFX(ent, "HGR1")
		hgr1.SetActive(true)
		hgr2 := GETGFX(ent, "HGR2")
		hgr2.SetActive(false)
	case 1:
		hgr1 := GETGFX(ent, "HGR1")
		hgr1.SetActive(false)
		hgr2 := GETGFX(ent, "HGR2")
		hgr2.SetActive(true)
	}
	ent.WriteLayersToMemory()
}

func LOGRActivePage(ent interfaces.Interpretable) string {

	// modes := GetActiveVideoModes(ent)

	// if len(modes) > 0 {
	// 	switch modes[0] {
	// 	case "LOGR", "HGR1", "DHR1":
	// 		return "LOGR"
	// 	case "LGR2", "HGR2", "DHR2":
	// 		return "LGR2"
	// 	}
	// }

	// // still here then try text modes
	// modes = GetActiveTextModes(ent)

	// if len(modes) > 0 {
	// 	switch modes[0] {
	// 	case "TEXT":
	// 		return "LOGR"
	// 	case "TXT2":
	// 		return "LGR2"
	// 	}
	// }

	return "LOGR"

}

func LOGRPlot40(ent interfaces.Interpretable, x, y, c uint64) {
	page := LOGRActivePage(ent)
	txt := GETGFX(ent, page)
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}

	px := int((x * 2) % 80)
	py := int(((y / 2) * 2) % 48)

	v := txt.Control.GetValueXY(px, py)
	c0 := v & 0xf
	c1 := (v & 0xf0) >> 4
	switch y % 2 {
	case 0:
		c0 = c
	case 1:
		c1 = c
	}
	v = (v & 0xffff0000) | (c1 << 4) | c0
	txt.Control.PutValueXY(px, py, v)
}

func LOGRPlot80(ent interfaces.Interpretable, x, y, c uint64) {
	txt := GETGFX(ent, "DLGR")
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	txt.Control.SwitchedInterleave()

	px := int(x % 80)
	py := int(((y / 2) * 2) % 48)

	v := txt.Control.GetValueXY(px, py)
	cv := c
	if (px % 2) == 0 {
		cv = uint64(ror4bit(int(cv)))
	}
	c0 := v & 0xf
	c1 := (v & 0xf0) >> 4
	switch y % 2 {
	case 0:
		c0 = cv
	case 1:
		c1 = cv
	}
	v = (v & 0xffff0000) | (c1 << 4) | c0
	txt.Control.PutValueXY(px, py, v)
}

func LOGRGet40(ent interfaces.Interpretable, x, y int) uint64 {
	page := LOGRActivePage(ent)
	txt := GETGFX(ent, page)
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}

	px := int((x * 2) % 80)
	py := int(((y / 2) * 2) % 48)

	v := txt.Control.GetValueXY(px, py)
	c0 := v & 0xf
	c1 := (v & 0xf0) >> 4
	switch y % 2 {
	case 0:
		return c0
	case 1:
		return c1
	}
	return 0
}

func LOGRGet80(ent interfaces.Interpretable, x, y int) uint64 {
	txt := GETGFX(ent, "DLGR")
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	txt.Control.SwitchedInterleave()

	px := int(x % 80)
	py := int(((y / 2) * 2) % 48)

	v := txt.Control.GetValueXY(px, py)
	c0 := v & 0xf
	c1 := (v & 0xf0) >> 4
	switch y % 2 {
	case 0:
		return c0
	case 1:
		return c1
	}
	return 0
}

func LOGRHLine40(ent interfaces.Interpretable, x0, x1, y0, c uint64) {
	page := LOGRActivePage(ent)
	//log2.Printf("page = %s", page)
	txt := GETGFX(ent, page)
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}

	y := y0
	for x := x0; x <= x1; x++ {

		px := int((x * 2) % 80)
		py := int(((y / 2) * 2) % 48)

		v := txt.Control.GetValueXY(px, py)
		c0 := v & 0xf
		c1 := (v & 0xf0) >> 4
		switch y % 2 {
		case 0:
			c0 = c
		case 1:
			c1 = c
		}
		v = (v & 0xffff0000) | (c1 << 4) | c0
		txt.Control.PutValueXY(px, py, v)

	}

}

func LOGRHLine80(ent interfaces.Interpretable, x0, x1, y0, c uint64) {
	txt := GETGFX(ent, "DLGR")
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	txt.Control.SwitchedInterleave()

	y := y0
	for x := x0; x <= x1; x++ {

		px := int(x % 80)
		py := int(((y / 2) * 2) % 48)

		v := txt.Control.GetValueXY(px, py)
		c0 := v & 0xf
		c1 := (v & 0xf0) >> 4

		cv := c
		if (x % 2) == 0 {
			cv = uint64(ror4bit(int(cv)))
		}

		switch y % 2 {
		case 0:
			c0 = cv
		case 1:
			c1 = cv
		}
		v = (v & 0xffff0000) | (c1 << 4) | c0
		txt.Control.PutValueXY(px, py, v)

	}

}

func LOGRClear(ent interfaces.Interpretable, c uint64, lines int) {
	page := LOGRActivePage(ent)
	txt := GETGFX(ent, page)
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}

	for y := 0; y < lines; y++ {
		for x := 0; x <= 40; x++ {

			px := int((x * 2) % 80)
			py := int(((y / 2) * 2) % 48)

			v := txt.Control.GetValueXY(px, py)
			c0 := v & 0xf
			c1 := (v & 0xf0) >> 4
			switch y % 2 {
			case 0:
				c0 = c
			case 1:
				c1 = c
			}
			v = (v & 0xffff0000) | (c1 << 4) | c0
			txt.Control.PutValueXY(px, py, v)

		}
	}

}

func ror4bit(c int) int {
	return (c >> 1) | ((c & 1) << 3)
}

func LOGRClear80(ent interfaces.Interpretable, c uint64, lines int) {
	txt := GETGFX(ent, "DLGR")
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	txt.Control.SwitchedInterleave()

	for y := 0; y < lines; y++ {
		for x := 0; x <= 80; x++ {

			px := int(x % 80)
			py := int(((y / 2) * 2) % 48)

			v := txt.Control.GetValueXY(px, py)
			c0 := v & 0xf
			c1 := (v & 0xf0) >> 4

			cv := c
			if (x % 2) == 0 {
				cv = uint64(ror4bit(int(cv)))
			}

			switch y % 2 {
			case 0:
				c0 = cv
			case 1:
				c1 = cv
			}
			v = (v & 0xffff0000) | (c1 << 4) | c0
			txt.Control.PutValueXY(px, py, v)

		}
	}

}

func LOGRVLine40(ent interfaces.Interpretable, y0, y1, x0, c uint64) {
	page := LOGRActivePage(ent)
	//log2.Printf("page = %s", page)
	txt := GETGFX(ent, page)
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}

	x := x0
	for y := y0; y <= y1; y++ {

		px := int((x * 2) % 80)
		py := int(((y / 2) * 2) % 48)

		v := txt.Control.GetValueXY(px, py)
		c0 := v & 0xf
		c1 := (v & 0xf0) >> 4
		switch y % 2 {
		case 0:
			c0 = c
		case 1:
			c1 = c
		}
		v = (v & 0xffff0000) | (c1 << 4) | c0
		txt.Control.PutValueXY(px, py, v)

	}

}

func LOGRVLine80(ent interfaces.Interpretable, y0, y1, x0, c uint64) {
	txt := GETGFX(ent, "DLGR")
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	txt.Control.SwitchedInterleave()

	x := x0
	for y := y0; y <= y1; y++ {

		px := int(x % 80)
		py := int(((y / 2) * 2) % 48)

		v := txt.Control.GetValueXY(px, py)
		c0 := v & 0xf
		c1 := (v & 0xf0) >> 4

		cv := c
		if (x % 2) == 0 {
			cv = uint64(ror4bit(int(cv)))
		}

		switch y % 2 {
		case 0:
			c0 = cv
		case 1:
			c1 = cv
		}
		v = (v & 0xffff0000) | (c1 << 4) | c0
		txt.Control.PutValueXY(px, py, v)

	}

}

func SetTextFull(ent interfaces.Interpretable) {
	SetHUDLayerOn(ent, textselect[ent.GetMemIndex()])
	SetTextSplit(ent, 48) // full rows
	SetHGRSplit(ent, 48)
	SetGRSplit(ent, 48)
}

func SetGraphicsFull(ent interfaces.Interpretable) {
	SetHUDLayerOff(ent, textselect[ent.GetMemIndex()])
	SetTextSplit(ent, 48) // full rows
	SetHGRSplit(ent, 0)
	SetGRSplit(ent, 0)
}

func SetGraphicsOn(ent interfaces.Interpretable) {
	GFXEnable(ent)
}

func NLIN(ent interfaces.Interpretable) {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	if txt.Control.CX > txt.Control.SX {
		txt.Control.CR()
		txt.Control.LF()
	}
}

func HomeLeft(ent interfaces.Interpretable) {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	txt.Control.CX = txt.Control.SX
}

func TEXT40(ent interfaces.Interpretable) {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	XGRLayerEnableCustom(ent, "XGR1", false, true, 280, 160)
	XGRLayerEnableCustom(ent, "XGR2", false, false, 280, 192)
	_ = ent.GetMemory(49152 + 89)
	txt.Control.SetSizeAndClear(types.W_NORMAL_H_NORMAL)
	txt.Control.NormalInterleave()
	txt.Control.NormalChars()
	txt.Control.FullRefresh()
}

func TEXT80(ent interfaces.Interpretable) {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	XGRLayerEnableCustom(ent, "XGR1", false, true, 280, 160)
	XGRLayerEnableCustom(ent, "XGR2", false, false, 280, 192)
	_ = ent.GetMemory(49152 + 89)
	txt.Control.SetSizeAndClear(types.W_HALF_H_NORMAL)
	txt.Control.SwitchedInterleave()
	txt.Control.AltChars()
	txt.Control.FullRefresh()
}

func TEXT40Preserve(ent interfaces.Interpretable) {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	XGRLayerEnableCustom(ent, "XGR1", false, true, 280, 160)
	XGRLayerEnableCustom(ent, "XGR2", false, false, 280, 192)
	_ = ent.GetMemory(49152 + 89)
	txt.Control.SetSizeAndClearAttr(types.W_NORMAL_H_NORMAL)
	txt.Control.NormalInterleave()
	txt.Control.NormalChars()
	txt.Control.FullRefresh()
}

func TEXT80Preserve(ent interfaces.Interpretable) {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	XGRLayerEnableCustom(ent, "XGR1", false, true, 280, 160)
	XGRLayerEnableCustom(ent, "XGR2", false, false, 280, 192)
	//_ = ent.GetMemory(49152 + 89)
	txt.Control.SwitchedInterleave()
	txt.Control.AltChars()
	txt.Control.SetSizeAndClearAttr(types.W_HALF_H_NORMAL)
	//txt.Control.ClearScreenWindow()
	txt.Control.FullRefresh()
}

func SwitchCPU(ent interfaces.Interpretable, cpu *mos6502.Core6502) {
	o := GetCPU(ent)
	o.Opref = cpu.Opref
	o.Model = cpu.Model
}

func New65C02(ent interfaces.Interpretable) *mos6502.Core6502 {
	return mos6502.NewCore65C02(
		ent,
		0, 0, 0,
		0xfa62,
		0,
		0x1fe,
		ent,
	)
}

func New6502(ent interfaces.Interpretable) *mos6502.Core6502 {
	return mos6502.NewCore6502(
		ent,
		0, 0, 0,
		0xfa62,
		0,
		0x1fe,
		ent,
	)
}

func RotatePalette(ent interfaces.Interpretable, min, max int, change int) {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	_ = ent.GetMemory(49152 + 89)
	txt.Control.RotatePalette(min, max, change)
}

func TEXTMAX(ent interfaces.Interpretable) {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	txt.Control.SetSizeAndClear(types.W_HALF_H_HALF)
	txt.Control.SwitchedInterleave()
	txt.Control.AltChars()
	txt.Control.FullRefresh()
}

func MODE40(ent interfaces.Interpretable) {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	//_ = ent.GetMemory(49152 + 89)
	txt.SetActive(true)
	txt.Control.NormalInterleave()
	txt.Control.SetSizeAndClear(types.W_NORMAL_H_NORMAL)
	txt.Control.NormalChars()
	txt.Control.FullRefresh()
	txt.Control.NormalInterleave()
	txt.SetSubFormat(0)
}

func MODE80(ent interfaces.Interpretable) {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	//_ = ent.GetMemory(49152 + 89)
	txt.Control.SetSizeAndClear(types.W_HALF_H_NORMAL)
	//txt.Control.SwitchedInterleave()
	txt.Control.AltChars()
	txt.Control.FullRefresh()
}

func MODE80PreserveAlt(ent interfaces.Interpretable) {
	txt := GETHUD(ent, "TXT2")
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	//_ = ent.GetMemory(49152 + 89)
	//txt.Control.SetSizeAndClear(types.W_HALF_H_NORMAL)
	//txt.SetActive(true)
	txt.SetSubFormat(types.LSF_FIXED_80_24)
	//txt.Control.AltChars()
	txt.Control.FullRefresh()
}

func MODE80Preserve(ent interfaces.Interpretable) {
	// if textselect[ent.GetMemIndex()] != "TEXT" {
	// 	return
	// }
	txt := GETHUD(ent, "TEXT")
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	//_ = ent.GetMemory(49152 + 89)
	//txt.Control.SetSizeAndClear(types.W_HALF_H_NORMAL)
	txt.SetActive(true)
	txt.SetSubFormat(types.LSF_FIXED_80_24)
	//txt.Control.AltChars()
	txt.Control.FullRefresh()
}

func MODE40Preserve(ent interfaces.Interpretable) {
	if textselect[ent.GetMemIndex()] != "TEXT" {
		return
	}
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	//_ = ent.GetMemory(49152 + 89)
	txt.SetActive(true)
	txt.SetSubFormat(types.LSF_FIXED_40_24)
	txt.Control.NormalChars()
	txt.Control.FullRefresh()
}

func GetCursorX(ent interfaces.Interpretable) int {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	return txt.Control.CX
}

func SetCursorX(ent interfaces.Interpretable, x int) {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	txt.Control.CX = x
	txt.Control.Repos()
}

func SetCursorY(ent interfaces.Interpretable, y int) {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	txt.Control.CY = y
	txt.Control.Repos()
}

func ResyncCursor(ent interfaces.Interpretable) {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	txt.Control.Sync()
}

func GetCursorRelativeX(ent interfaces.Interpretable) int {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	return txt.Control.CX - txt.Control.SX
}

func GetCursorY(ent interfaces.Interpretable) int {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	return txt.Control.CY
}

func GetCursorRelativeY(ent interfaces.Interpretable) int {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	return txt.Control.CY - txt.Control.SY
}

func GetFullColumns(ent interfaces.Interpretable) int {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	r := 80
	switch txt.Control.Font / 4 {
	case 1:
		r = r / 2
	case 2:
		r = r / 4
	case 3:
		r = r / 8
	}
	return r
}

func GetFullRows(ent interfaces.Interpretable) int {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	r := 48
	switch txt.Control.Font % 4 {
	case 1:
		r = r / 2
	case 2:
		r = r / 4
	case 3:
		r = r / 8
	}
	return r
}

func GetColumns(ent interfaces.Interpretable) int {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	r := txt.Control.EX - txt.Control.SX + 1
	switch txt.Control.Font / 4 {
	case 1:
		r = r / 2
	case 2:
		r = r / 4
	case 3:
		r = r / 8
	}
	return r
}

func GetRows(ent interfaces.Interpretable) int {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	r := txt.Control.EY - txt.Control.SY + 1
	switch txt.Control.Font % 4 {
	case 1:
		r = r / 2
	case 2:
		r = r / 4
	case 3:
		r = r / 8
	}
	return r
}

func GetActualRows(ent interfaces.Interpretable) int {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	r := (txt.Control.EY - txt.Control.SY + 1)
	switch txt.Control.Font % 4 {
	case 1:
		r = r / 2
	case 2:
		r = r / 4
	case 3:
		r = r / 8
	}
	return r
}

func Backspace(ent interfaces.Interpretable) {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	txt.Control.Backspace()
}

func HTab(ent interfaces.Interpretable, v int) {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	txt.Control.HorizontalTab(v)
}

func VTab(ent interfaces.Interpretable, v int) {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	txt.Control.VerticalTab(v)
}

func Attribute(ent interfaces.Interpretable, va types.VideoAttribute) {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	txt.Control.Attribute = va
}

func BeepReset(ent interfaces.Interpretable) {
	base := ent.GetMemoryMap().MEMBASE(ent.GetMemIndex())

	ent.GetMemoryMap().WriteGlobal(ent.GetMemIndex(), base+memory.OCTALYZER_SPEAKER_MS, 99)
	ent.GetMemoryMap().WriteGlobal(ent.GetMemIndex(), base+memory.OCTALYZER_SPEAKER_FREQ, 999999)
}

func Beep(ent interfaces.Interpretable) {

	// base := ent.GetMemoryMap().MEMBASE(ent.GetMemIndex())

	// ent.GetMemoryMap().WriteGlobal(ent.GetMemIndex(), base+memory.OCTALYZER_SPEAKER_MS, 100)
	// ent.GetMemoryMap().WriteGlobal(ent.GetMemIndex(), base+memory.OCTALYZER_SPEAKER_FREQ, 933)
	// time.Sleep(100 * time.Millisecond)

	// new world
	tone := ent.GetAudioPort("beep")
	restalgia.CommandF(ent, tone, restalgia.RF_setVolume, 0)
	restalgia.CommandI(ent, tone, restalgia.RF_setEnvelopeAttack, 0)
	restalgia.CommandI(ent, tone, restalgia.RF_setEnvelopeDecay, 0)
	restalgia.CommandI(ent, tone, restalgia.RF_setEnvelopeSustain, 100)
	restalgia.CommandI(ent, tone, restalgia.RF_setEnvelopeRelease, 1)
	restalgia.CommandF(ent, tone, restalgia.RF_setFrequency, 933)
	restalgia.CommandF(ent, tone, restalgia.RF_setVolume, 1)

	time.Sleep(100 * time.Millisecond)
}

func BeepError(ent interfaces.Interpretable) {

	base := ent.GetMemoryMap().MEMBASE(ent.GetMemIndex())

	ent.GetMemoryMap().WriteGlobal(ent.GetMemIndex(), base+memory.OCTALYZER_SPEAKER_MS, 200)
	ent.GetMemoryMap().WriteGlobal(ent.GetMemIndex(), base+memory.OCTALYZER_SPEAKER_FREQ, 200)

	time.Sleep(200 * time.Millisecond)

}

func Click(ent interfaces.Interpretable) {

}

func GR40At(ent interfaces.Interpretable, x, y uint64) uint64 {
	return LOGRGet40(ent, int(x), int(y))
}

func GR80At(ent interfaces.Interpretable, x, y uint64) uint64 {
	return LOGRGet80(ent, int(x), int(y))
}

func HGRAt(ent interfaces.Interpretable, x, y uint64) uint64 {
	cp := ent.GetCurrentPage()
	page := GETGFX(ent, cp)
	if page == nil || page.HControl == nil {
		panic("missing control interface")
	}
	return uint64(page.HControl.ColorAt(int(x), int(y)))
}

func HGRFill(ent interfaces.Interpretable, c uint64) {
	cp := ent.GetCurrentPage()
	page := GETGFX(ent, cp)
	if page == nil || page.HControl == nil {
		panic("missing control interface")
	}
	page.HControl.Fill(int(c))
}

func HGRClear(ent interfaces.Interpretable) {
	cp := ent.GetCurrentPage()
	page := GETGFX(ent, cp)
	if page == nil || page.HControl == nil {
		panic("missing control interface")
	}
	page.HControl.Clear(0)
}

func HGRPlot(ent interfaces.Interpretable, x, y, c uint64) {
	cp := ent.GetCurrentPage()
	page := GETGFX(ent, cp)
	if page == nil || page.HControl == nil {
		panic("missing control interface")
	}
	page.HControl.Plot(int(x), int(y), int(c))
	LastX = int(x)
	LastY = int(y)
	ent.SetMemory(224, uint64(LastX%256))
	ent.SetMemory(225, uint64(LastX/256))
	ent.SetMemory(226, uint64(LastY%256))
}

func HGRShape(ent interfaces.Interpretable, shape hires.ShapeEntry, x int, y int, scl int, rot float64, c int, usecol bool) {
	cp := ent.GetCurrentPage()
	page := GETGFX(ent, cp)
	if page == nil || page.HControl == nil {
		panic("missing control interface")
	}

	var div float64
	switch scl {
	case 1:
		div = 16
	case 2:
		div = 8
	case 3:
		div = 4
	case 4:
		div = 2
	default:
		div = 1
	}

	rot = math.Floor((rot+(div/2))/div) * div

	deg := (rot / 64) * 360

	page.HControl.Shape(shape, x, y, scl, deg, c, usecol)
	LastX, LastY = page.HControl.GetLastXY()
	ent.SetMemory(224, uint64(LastX%256))
	ent.SetMemory(225, uint64(LastX/256))
	ent.SetMemory(226, uint64(LastY%256))
}

func Circle(ent interfaces.Interpretable, radius, x, y, c int, plot func(ent interfaces.Interpretable, x, y, c uint64)) {
	var lcx, lcy float64 = -1, -1
	var sx, sy float64 = -1, -1
	for angle := 0; angle < 360; angle++ {
		cx := float64(x) + float64(radius)*math.Sin(float64(angle)*0.0174532925)
		cy := float64(y) + float64(radius)*math.Cos(float64(angle)*0.0174532925)

		if lcx != -1 {
			BrenshamLine(ent, int(lcx), int(lcy), int(cx), int(cy), c, plot)
		} else {
			sx, sy = cx, cy
		}

		lcx, lcy = cx, cy
	}
	// close it up
	BrenshamLine(ent, int(lcx), int(lcy), int(sx), int(sy), c, plot)
}

func Poly(ent interfaces.Interpretable, radius, sides, x, y, c int, plot func(ent interfaces.Interpretable, x, y, c uint64)) {
	var lcx, lcy float64 = -1, -1
	var sx, sy float64 = -1, -1

	step := 360 / sides

	//fmt.Printf("Poly: sides=%d, step=%d\n", sides, step)

	for angle := 0; angle < 360; angle += step {

		cx := float64(x) + float64(radius)*math.Sin(float64(angle)*0.0174532925)
		cy := float64(y) + float64(radius)*math.Cos(float64(angle)*0.0174532925)

		//fmt.Printf("angle=%d, cx=%f, cy=%f\n", angle, cx, cy)

		if lcx != -1 {
			BrenshamLine(ent, int(lcx), int(lcy), int(cx), int(cy), c, plot)
		} else {
			sx, sy = cx, cy
		}

		lcx, lcy = cx, cy
	}
	// close it up
	BrenshamLine(ent, int(lcx), int(lcy), int(sx), int(sy), c, plot)
}

func GR40Fill(ent interfaces.Interpretable, c uint64) {
	rows := GetRows(ent) * 2
	LOGRClear(ent, c, 48-rows)
}

func GR80Fill(ent interfaces.Interpretable, c uint64) {
	rows := GetRows(ent) * 2
	LOGRClear80(ent, c, 48-rows)
}

func Fill(ent interfaces.Interpretable, c int, fill func(ent interfaces.Interpretable, c uint64)) {
	if fill != nil {
		fill(ent, uint64(c))
	}
}

func FloodFill(ent interfaces.Interpretable, x, y, c int, maxx, maxy int, get func(ent interfaces.Interpretable, x, y uint64) uint64, plot func(ent interfaces.Interpretable, x, y, c uint64)) {
	targ := get(ent, uint64(x), uint64(y))
	var ff func(x, y int)
	ff = func(x, y int) {
		if x < 0 || x > maxx || y < 0 || y > maxy {
			return
		}
		p := get(ent, uint64(x), uint64(y))
		if p == targ {
			plot(ent, uint64(x), uint64(y), uint64(c))
			ff(x-1, y)
			ff(x+1, y)
			ff(x, y-1)
			ff(x, y+1)
		}
	}
	ff(x, y)
}

func Arc(ent interfaces.Interpretable, radius, start, end, x, y, c int, plot func(ent interfaces.Interpretable, x, y, c uint64)) {
	var lcx, lcy float64 = -1, -1
	for angle := start; angle <= end; angle++ {
		cx := float64(x) + float64(radius)*math.Sin(float64(angle)*0.0174532925)
		cy := float64(y) + float64(radius)*math.Cos(float64(angle)*0.0174532925)

		if lcx != -1 {
			BrenshamLine(ent, int(lcx), int(lcy), int(cx), int(cy), c, plot)
		}

		lcx, lcy = cx, cy
	}
}

func BrenshamLine(ent interfaces.Interpretable, x0, y0, x1, y1 int, c int, plot func(ent interfaces.Interpretable, x, y, c uint64)) {
	// implemented straight from WP pseudocode
	dx := x1 - x0
	if dx < 0 {
		dx = -dx
	}
	dy := y1 - y0
	if dy < 0 {
		dy = -dy
	}
	var sx, sy int
	if x0 < x1 {
		sx = 1
	} else {
		sx = -1
	}
	if y0 < y1 {
		sy = 1
	} else {
		sy = -1
	}
	err := dx - dy

	for {
		//out = append(out, [2]int{x0, y0})
		plot(ent, uint64(x0), uint64(y0), uint64(c))
		//if p & 3 == 3 {
		//plot(ent, uint64(x0+1), uint64(y0), uint64(c))
		// }
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

func HGRLine(ent interfaces.Interpretable, x0, y0, x1, y1, c uint64) {
	cp := ent.GetCurrentPage()
	page := GETGFX(ent, cp)
	if page == nil || page.HControl == nil {
		panic("missing control interface")
	}
	hires.BrenshamLine(int(x0), int(y0), int(x1), int(y1), int(c), page.HControl)
	LastX = int(x1)
	LastY = int(y1)
	ent.SetMemory(224, uint64(LastX%256))
	ent.SetMemory(225, uint64(LastX/256))
	ent.SetMemory(226, uint64(LastY%256))
}

func GetHGRCollisionCount(ent interfaces.Interpretable) uint64 {
	cp := ent.GetCurrentPage()
	page := GETGFX(ent, cp)
	if page == nil || page.HControl == nil {
		panic("missing control interface")
	}
	return page.HControl.GetCollisionCount()
}

func SetHGRCollisionCount(ent interfaces.Interpretable, v uint64) {
	cp := ent.GetCurrentPage()
	page := GETGFX(ent, cp)
	if page == nil || page.HControl == nil {
		panic("missing control interface")
	}
	page.HControl.SetCollisionCount(v)
}

func ScrollWindow(ent interfaces.Interpretable) {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	txt.Control.ScrollBy(txt.Control.FontH())
}

func ScrollWindowBy(ent interfaces.Interpretable, lines int) {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	txt.Control.ScrollByWindow(lines)
}

func LOGRScroll40(ent interfaces.Interpretable) {
	txt := GETGFX(ent, "LOGR")
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}

	for y := 1; y < 48; y++ {
		for x := 0; x <= 40; x++ {

			px := int((x * 2) % 80)
			py0 := int((((y - 1) / 2) * 2) % 48)
			py1 := int(((y / 2) * 2) % 48)

			v := txt.Control.GetValueXY(px, py1)
			txt.Control.PutValueXY(px, py0, v)

		}

		if y == 47 {
			for x := 0; x <= 40; x++ {
				px := int((x * 2) % 80)
				py := int(((y / 2) * 2) % 48)
				v := txt.Control.GetValueXY(px, py)
				txt.Control.PutValueXY(px, py, v&0xffff0000)
			}
		}
	}

}

func LOGRScroll80(ent interfaces.Interpretable) {
	txt := GETGFX(ent, "LOGR")
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}

	for y := 1; y < 48; y++ {
		for x := 0; x <= 80; x++ {

			px := int(x % 80)
			py0 := int((((y - 1) / 2) * 2) % 48)
			py1 := int(((y / 2) * 2) % 48)

			v := txt.Control.GetValueXY(px, py1)
			txt.Control.PutValueXY(px, py0, v)

		}

		if y == 47 {
			for x := 0; x <= 80; x++ {
				px := int(x % 80)
				py := int(((y / 2) * 2) % 48)
				v := txt.Control.GetValueXY(px, py)
				txt.Control.PutValueXY(px, py, v&0xffff0000)
			}
		}
	}

}

func CursorLeft(ent interfaces.Interpretable) {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	txt.Control.CursorLeft()
}

func CursorRight(ent interfaces.Interpretable) {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	txt.Control.CursorRight()
}

func CursorUp(ent interfaces.Interpretable) {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	txt.Control.CursorUp()
}

func CursorDown(ent interfaces.Interpretable) {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	txt.Control.CursorDown()
}

func SetTextLeftMargin(ent interfaces.Interpretable, value int) {
	// log2.Printf("settextleftmargin for slot %d / %s", ent.GetMemIndex(), textselect[ent.GetMemIndex()])
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	r := 1
	switch txt.Control.Font / 4 {
	case 1:
		r = 2
	case 2:
		r = 4
	case 3:
		r = 8
	}

	w := txt.Control.EX - txt.Control.SX

	txt.Control.SX = value * r
	txt.Control.EX = txt.Control.SX + w
	if txt.Control.EX > 79 {
		txt.Control.EX = 79
	}

	if txt.Control.CX < txt.Control.SX {
		txt.Control.CX = txt.Control.SX
	}

	log.Printf("*** Left Margin: %d, Right Margin: %d\n", txt.Control.SX, txt.Control.EX)
}

func SetTextWidth(ent interfaces.Interpretable, value int) {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	r := 1
	switch txt.Control.Font / 4 {
	case 1:
		r = 2
	case 2:
		r = 4
	case 3:
		r = 8
	}
	txt.Control.EX = txt.Control.SX + (value * r) - 1
}

func SetTextTopMargin(ent interfaces.Interpretable, value int) {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	r := 1
	switch txt.Control.Font % 4 {
	case 1:
		r = 2
	case 2:
		r = 4
	case 3:
		r = 8
	}
	txt.Control.SY = value * r
}

func SetTextBottomMargin(ent interfaces.Interpretable, value int) {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	r := 1
	switch txt.Control.Font % 4 {
	case 1:
		r = 2
	case 2:
		r = 4
	case 3:
		r = 8
	}
	txt.Control.EY = value*r - 1
}

func SetClientCopy(ent interfaces.Interpretable, b bool) {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	if b {
		txt.Control.CopyOn()
	} else {
		txt.Control.CopyOff()
	}
}

func ClipCursor(ent interfaces.Interpretable) {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}

	x := txt.Control.CX
	y := txt.Control.CY

	if x < txt.Control.SX {
		x = txt.Control.SX
	}

	if x > txt.Control.EX {
		x = txt.Control.EX
	}

	if y < txt.Control.SY {
		y = txt.Control.SY
	}

	if y > txt.Control.EY {
		y = txt.Control.EX
	}

	txt.Control.CX, txt.Control.CY = x, y

}

func GetTextLeftMargin(ent interfaces.Interpretable) int {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	r := 1
	switch txt.Control.Font / 4 {
	case 1:
		r = 2
	case 2:
		r = 4
	case 3:
		r = 8
	}
	return txt.Control.SX / r
}

func GetTextWidth(ent interfaces.Interpretable) int {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	r := 1
	switch txt.Control.Font / 4 {
	case 1:
		r = 2
	case 2:
		r = 4
	case 3:
		r = 8
	}
	return (txt.Control.EX - txt.Control.SX + 1) / r
}

func GetTextTopMargin(ent interfaces.Interpretable) int {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	r := 1
	switch txt.Control.Font % 4 {
	case 1:
		r = 2
	case 2:
		r = 4
	case 3:
		r = 8
	}
	return txt.Control.SY / r
}

func GetTextBottomMargin(ent interfaces.Interpretable) int {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	r := 1
	switch txt.Control.Font % 4 {
	case 1:
		r = 2
	case 2:
		r = 4
	case 3:
		r = 8
	}
	return (txt.Control.EY + 1) / r
}

func GetAttribute(ent interfaces.Interpretable) types.VideoAttribute {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	return txt.Control.Attribute
}

func GetWozOffsetLine(ent interfaces.Interpretable) int {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt.Control == nil {
		panic("layer has a nil control surface")
	}
	return txt.Control.GetOffsetXY(0, txt.Control.CY)
}

func (f *SoftSwitchConfig) AsUint() uint64 {
	var u uint64
	/*
		if f.SoftSwitch_80STORE {
			u = u | 128
		}
		if f.SoftSwitch_RAMWRT {
			u = u | 64
		}
		if f.SoftSwitch_RAMRD {
			u = u | 32
		}
	*/
	if f.SoftSwitch_DoubleRes {
		u = u | 16
	}
	if f.SoftSwitch_GRAPHICS {
		u = u | 8
	}
	if f.SoftSwitch_MIXED {
		u = u | 4
	}
	if f.SoftSwitch_HIRES {
		u = u | 2
	}
	if f.SoftSwitch_PAGE2 {
		u = u | 1
	}
	return u
}

func (f *SoftSwitchConfig) FromUint(u uint64) {
	/*
		f.SoftSwitch_80STORE = (u & 128) != 0
		f.SoftSwitch_RAMWRT = (u & 64) != 0
		f.SoftSwitch_RAMRD = (u & 32) != 0
	*/
	f.SoftSwitch_DoubleRes = (u & 16) != 0
	f.SoftSwitch_GRAPHICS = (u & 8) != 0
	f.SoftSwitch_MIXED = (u & 4) != 0
	f.SoftSwitch_HIRES = (u & 2) != 0
	f.SoftSwitch_PAGE2 = (u & 1) != 0
}

func RestoreSoftSwitches(ent interfaces.Interpretable) {
	of := SoftSwitchConfig{
		SoftSwitch_HIRES:     false,
		SoftSwitch_PAGE2:     false,
		SoftSwitch_GRAPHICS:  false,
		SoftSwitch_MIXED:     false,
		SoftSwitch_DoubleRes: false,
	}
	nf := &SoftSwitchConfig{}
	nf.FromUint(ent.GetMemory(0xfcff))
	ReconfigureSoftSwitches(ent, of, *nf, true)
}

func ResetSoftSwitches(ent interfaces.Interpretable) {
	f := SoftSwitchConfig{}
	o := SoftSwitchConfig{}
	f.SoftSwitch_DoubleRes = false
	f.SoftSwitch_MIXED = false
	f.SoftSwitch_GRAPHICS = false
	f.SoftSwitch_HIRES = false
	f.SoftSwitch_PAGE2 = false
	f.SoftSwitch_80STORE = false
	f.SoftSwitch_RAMRD = false
	f.SoftSwitch_RAMWRT = false
	ReconfigureSoftSwitches(ent, o, f, true)
}

func GetVideoMode(ent interfaces.Interpretable) string {

	result := "TEXT"

	var l *types.LayerSpecMapped

	modes := []string{"LOGR", "DLGR", "HGR1", "HGR2", "DHR1", "DHR2", "VCTR", "CUBE", "SHR1"}

	for _, mn := range modes {
		l = GETGFX(ent, mn)
		if l != nil && l.GetActive() {
			return mn
		}
	}

	return result
}

/*
ReconfigureSoftSwitches toggles the various layer states based on changes to a flags
collection.  This mimics the behaviour of the Apple2 soft switches
*/
func ReconfigureSoftSwitches(ent interfaces.Interpretable, of, nf SoftSwitchConfig, force bool) {

	// store softswitches
	v := nf.AsUint()
	ent.SetMemory(0xfcff, v)

	ent.WaitForLayers()

	ent.ReadLayersFromMemory()
	// Bulk changes here

	// 1 - TEXT Layer
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt == nil {
		panic("Expected layer id TEXT not found")
	}
	switch {
	case nf.SoftSwitch_GRAPHICS == true && nf.SoftSwitch_MIXED == false:
		txt.SetActive(false)
		txt.SetBounds(80, 48, 80, 48)
	case nf.SoftSwitch_GRAPHICS == true && nf.SoftSwitch_MIXED == true:
		txt.SetActive(true)
		txt.SetBounds(0, 40, 79, 47)
	case nf.SoftSwitch_GRAPHICS == false:
		txt.SetActive(true)
		txt.SetBounds(0, 0, 79, 47)
	}

	// 2 - LOGR Layer
	gr := GETGFX(ent, "LOGR")
	if gr == nil {
		panic("Expected layer id LOGR not found")
	}
	if !nf.SoftSwitch_DoubleRes {
		switch {
		case nf.SoftSwitch_GRAPHICS == true && nf.SoftSwitch_MIXED == false && nf.SoftSwitch_HIRES == false:
			txt.SetActive(false)
			gr.SetActive(true)
			gr.SetBounds(0, 0, 39, 47)
		case nf.SoftSwitch_GRAPHICS == true && nf.SoftSwitch_MIXED == true && nf.SoftSwitch_HIRES == false:
			gr.SetActive(true)
			gr.SetBounds(0, 0, 39, 39)
		default:
			gr.SetActive(false)
		}
	} else {
		gr.SetActive(false)
	}

	// 2 - DLGR Layer
	gr = GETGFX(ent, "DLGR")
	if gr == nil {
		panic("Expected layer id DLGR not found")
	}
	if nf.SoftSwitch_DoubleRes {
		switch {
		case nf.SoftSwitch_GRAPHICS == true && nf.SoftSwitch_MIXED == false && nf.SoftSwitch_HIRES == false:
			txt.SetActive(false)
			gr.SetActive(true)
			gr.SetBounds(0, 0, 79, 47)
		case nf.SoftSwitch_GRAPHICS == true && nf.SoftSwitch_MIXED == true && nf.SoftSwitch_HIRES == false:
			gr.SetActive(true)
			gr.SetBounds(0, 0, 79, 39)
		default:
			gr.SetActive(false)
		}
	} else {
		gr.SetActive(false)
	}

	// 3 - HGR1 Layer
	gr = GETGFX(ent, "HGR1")
	if gr == nil {
		panic("Expected layer id HGR1 not found")
	}
	if !nf.SoftSwitch_DoubleRes {
		switch {
		case nf.SoftSwitch_GRAPHICS == true && nf.SoftSwitch_MIXED == false && nf.SoftSwitch_HIRES == true && nf.SoftSwitch_PAGE2 == false:
			txt.SetActive(false)
			gr.SetActive(true)
			gr.SetBounds(0, 0, 279, 191)
		case nf.SoftSwitch_GRAPHICS == true && nf.SoftSwitch_MIXED == true && nf.SoftSwitch_HIRES == true && nf.SoftSwitch_PAGE2 == false:
			gr.SetActive(true)
			gr.SetBounds(0, 0, 279, 159)
		default:
			gr.SetActive(false)
		}
	} else {
		gr.SetActive(false)
	}

	// 4 - HGR2 Layer
	gr = GETGFX(ent, "HGR2")
	if gr == nil {
		panic("Expected layer id HGR2 not found")
	}
	if !nf.SoftSwitch_DoubleRes {
		switch {
		case nf.SoftSwitch_GRAPHICS == true && nf.SoftSwitch_MIXED == false && nf.SoftSwitch_HIRES == true && nf.SoftSwitch_PAGE2 == true:
			txt.SetActive(false)
			gr.SetActive(true)
			gr.SetBounds(0, 0, 279, 191)
		case nf.SoftSwitch_GRAPHICS == true && nf.SoftSwitch_MIXED == true && nf.SoftSwitch_HIRES == true && nf.SoftSwitch_PAGE2 == true:
			gr.SetActive(true)
			gr.SetBounds(0, 0, 279, 159)
		default:
			gr.SetActive(false)
		}
	} else {
		gr.SetActive(false)
	}

	// 3 - DHR1 Layer
	gr = GETGFX(ent, "DHR1")
	if gr == nil {
		panic("Expected layer id DHR1 not found")
	}
	if nf.SoftSwitch_DoubleRes {
		switch {
		case nf.SoftSwitch_GRAPHICS == true && nf.SoftSwitch_MIXED == false && nf.SoftSwitch_HIRES == true && nf.SoftSwitch_PAGE2 == false:
			txt.SetActive(false)
			gr.SetActive(true)
			gr.SetBounds(0, 0, 559, 191)
		case nf.SoftSwitch_GRAPHICS == true && nf.SoftSwitch_MIXED == true && nf.SoftSwitch_HIRES == true && nf.SoftSwitch_PAGE2 == false:
			txt.SetActive(false)
			gr.SetActive(true)
			gr.SetBounds(0, 0, 559, 159)
		default:
			gr.SetActive(false)
		}
	} else {
		gr.SetActive(false)
	}

	// 3 - DHR1 Layer
	gr = GETGFX(ent, "DHR2")
	if gr == nil {
		panic("Expected layer id DHR2 not found")
	}
	if nf.SoftSwitch_DoubleRes {
		switch {
		case nf.SoftSwitch_GRAPHICS == true && nf.SoftSwitch_MIXED == false && nf.SoftSwitch_HIRES == true && nf.SoftSwitch_PAGE2 == true:
			txt.SetActive(false)
			gr.SetActive(true)
			gr.SetBounds(0, 0, 559, 191)
		case nf.SoftSwitch_GRAPHICS == true && nf.SoftSwitch_MIXED == true && nf.SoftSwitch_HIRES == true && nf.SoftSwitch_PAGE2 == true:
			gr.SetActive(true)
			gr.SetBounds(0, 0, 559, 159)
		default:
			gr.SetActive(false)
		}
	} else {
		gr.SetActive(false)
	}

	// End bulk changes here
	//ent.WriteLayersToMemory()
	//ent.WaitForLayers()
}

func RealPut(ent interfaces.Interpretable, ch rune) {

	this := ent

	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt == nil {
		panic("Expected layer id TEXT not found")
	}

	if runestring.Pos(rune(3), this.GetBuffer()) > 0 && this.IsBreakable() {
		this.SetSpeed(255)
	}

	//System.out.println("Screen memory = $"+Integer.toHexString(addr));

	//ch = (char) translateChar(ch);

	if ch == 10 {

		if (this.IsDosCommand()) && (len(this.GetDosBuffer()) > 0) {
			//doscommand = false;
			this.SetDosCommand(false)

			for ent.GetChild() != nil {
				ent = ent.GetChild()
			}

			//System.out.println("Received dos command: "+dosbuffer);
			//System.out.println("Current work directory: "+ent.WorkDir);
			s := this.GetDosBuffer()
			s = strings.Replace(s, "\r", "", -1)
			this.SetDosBuffer("")
			e := ent.GetDialect().ProcessDynamicCommand(ent, s)
			if e != nil {

				if !ent.HandleError() {
					ent.GetDialect().HandleException(ent, e)
				}

			}
			this.SetDosBuffer("")
			return
		}
	}

	if this.IsNextByteColor() {
		txt.Control.FGColor = (uint64(ch) & 0x0f)
		this.SetNextByteColor(false)
		return
	}

	if this.IsDosCommand() {
		if ch == 10 {
			this.SetDosCommand(false)
			return
		}
		this.SetDosBuffer(this.GetDosBuffer() + string(ch))
		return
	}

	// is this a null.?
	if ch == 0 {
		return
	}

	// handle output
	if this.GetOutChannel() != "" && (!this.IsDosCommand()) {
		// send it out here..
		if ch == 4 {
			this.SetDosCommand(true)
			this.SetDosBuffer("")
			return
		}

		// write out to current file
		if ch == '\t' {
			ch = ','
		}
		// write
		files.DOSPRINT(
			files.GetPath(this.GetOutChannel()),
			files.GetFilename(this.GetOutChannel()),
			[]byte{byte(ch)},
		)

		return
	}

	if files.Dribble != "" {
		files.DOSDRIBBLEDATA([]byte{byte(ch)})
	}

	if ch == 4 && ent.IsIgnoreSpecial() {
		ch = 1058
	}

	//	this.Reposition()
	switch { /* FIXME - Switch statement needs cleanup */
	case ch == 4:
		this.SetDosBuffer("")
		this.SetDosCommand(true)
		break
	case ch == 6:
		this.SetNextByteColor(true)
		break
	case ch == 11:
		txt.Control.ClearToBottom()
		break
	case ch == 12:
		txt.Control.ClearScreen()
		break
	case ch == 13:
		txt.Control.LF()
		txt.Control.CR()
		//System.out.print ( "\r" );
		break
	case ch == 10:
		if this.GetLastChar() != '\r' {
			txt.Control.LF()
		}
		//System.out.print ( "\n" );
		break
	case ch == 14:
		txt.Control.Attribute = types.VA_NORMAL
		break
	case ch == 15:
		txt.Control.Attribute = types.VA_INVERSE
		break
	case ch == 17:
		txt.Control.Font = types.W_NORMAL_H_NORMAL
		txt.Control.ClearScreen()
		break
	case ch == 18:
		txt.Control.Font = types.W_HALF_H_NORMAL
		txt.Control.ClearScreen()
		break
	case ch == 25:
		txt.Control.ClearScreenWindow()
		break
	case ch == 26:
		txt.Control.ClearToEOLWindow()
		break
	case ch == 28:
		txt.Control.CursorRight()
		break
	case ch == 29:
		txt.Control.ClearToEOLWindow()
		break
	case ch == 9:
		txt.Control.TAB(ent.GetTabWidth())
		break
	case ch == 8:
		txt.Control.CursorLeft()
		break
	case ch == 136:
		txt.Control.Backspace()
		break
	case ch == 7:
		Beep(ent)
		break
	case ch == 135:
		Beep(ent)
		break
	case ch == 27:
		{
			if txt.Control.Attribute == types.VA_INVERSE {
				MouseKeys = true
				//writeln("MOUSE KEYS IS == ",this.MouseKeys);
			}
			break
		}
	case ch == 24:
		{
			if txt.Control.Attribute == types.VA_INVERSE {
				MouseKeys = false
			}
			//writeln("MOUSE KEYS IS == ",this.MouseKeys);
			break
		}
	case ch >= vduconst.SHADE0 && ch <= vduconst.SHADE7:
		{
			txt.Control.Shade = uint64(ch - vduconst.SHADE0)
		}
	case ch >= vduconst.COLOR0 && ch <= vduconst.COLOR15:
		{
			txt.Control.FGColor = uint64(ch - vduconst.COLOR0)
		}
	case ch >= vduconst.BGCOLOR0 && ch <= vduconst.BGCOLOR15:
		{
			txt.Control.BGColor = uint64(ch - vduconst.BGCOLOR0)
		}
	case ch == vduconst.INVERSE_ON:
		{
			ColorFlip = ColorFlip
		}
	default:
		{
			txt.Control.Put(ch)
			break
		}
	}

	this.SetLastChar(ch)

}

func DoCall(addr int, caller interfaces.Interpretable, passtocpu bool) bool {

	// log2.Printf("docall for $%.4x", addr)

	for caller.GetChild() != nil {
		caller = caller.GetChild()
	}

	match := true
	if settings.PureBoot(caller.GetMemIndex()) {
		match = false
	}

	//if addr >= 0x9600 && addr <= 0xbfff {
	//	// dos ignore
	//	return match
	//}

	found := true

	if !settings.PureBoot(caller.GetMemIndex()) {

		switch addr { /* FIXME - Switch statement needs cleanup */
		case 40278:
			// dos chain entry point
		case 64286:
			// PREAD X=paddlenum, Y=value
			pnum := GetCPU(caller).X
			v := int(caller.GetMemoryMap().IntGetPaddleValue(caller.GetMemIndex(), pnum))
			GetCPU(caller).Y = v
			GetCPU(caller).Set_nz(v)
		case 64347:
			// take processor A value, stick in 37
			caller.SetMemory(37, uint64(GetCPU(caller).A))
			//GetCPU(caller).Rts_imp()
		case 54915:
			// don't do much not needed
		case 653781:
			mm := NewMonitor(caller)
			mm.Manual("")
		case 653851:
			mm := NewMonitor(caller)
			mm.Manual("")
		case 64578:

			ClearToBottom(caller)

		case 64879:

			s := GetCRTLine(caller)

			if len(s) > 255 {
				s = s[0:255]
			}

			for i, ch := range s {
				caller.SetMemory(512+i, uint64(ch))
			}
			v := len(s)
			GetCPU(caller).X = v
			GetCPU(caller).Set_nz(v)

		case 65068:
			os := int(caller.GetMemory(60) + 256*caller.GetMemory(61))
			oe := int(caller.GetMemory(62) + 256*caller.GetMemory(63))
			ns := int(caller.GetMemory(66) + 256*caller.GetMemory(67))

			for idx := os; idx <= oe; idx++ {
				caller.SetMemory(
					ns+(idx-os),
					caller.GetMemory(idx),
				)
			}
		case 64858:
			// wait for ENTER key
			for caller.GetMemory(49152) != 131 {
				time.Sleep(1 * time.Millisecond)
			}
		case 64780:
			// wait for a key
			for caller.GetMemory(49152) < 128 {
				time.Sleep(1 * time.Millisecond)
			}
			caller.SetMemory(49168, 0)
			// Seed key to Accumulator
			v := int(caller.GetMemory(49152) | 128)
			GetCPU(caller).A = v
			GetCPU(caller).Set_nz(v)
		case 64484:
			Beep(caller)
			break
		case 65338:
			Beep(caller)
			break
		case 65152:
			Attribute(caller, types.VA_INVERSE)
			break
		case 65156:
			//caller.GetVDU().SetAttribute(types.VA_NORMAL)
			Attribute(caller, types.VA_NORMAL)
			break
		case 64600:
			Clearscreen(caller)
			break
		case 64580:
			ClearToBottom(caller)
			break
		case 62450:
			//LOGRClear(caller, 0, 48)
			HGRClear(caller)
			break
		case 62454:
			{
				hc := GetHCOLOR(caller)
				HGRFill(caller, uint64(hc))
				break
			}
		case 64624:
			ScrollWindow(caller)
			break
		case 64500:
			CursorRight(caller)
			break
		case 64528:
			CursorLeft(caller)
			break
		case 64538:
			CursorUp(caller)
			break
		case 63538:
			LOGRClear(caller, 0, 48)
			break
		case 63542:
			LOGRClear(caller, 0, 40)
			break
		case 64614:
			CursorDown(caller)
			break
		case 64668:
			//caller.GetVDU().ClearToEOL()
			ClearToEOL(caller)
			break
		case 62923:
			{
				//caller.SetMemory(226, uint64(caller.GetVDU().GetLastY()))
				//caller.SetMemory(224, uint64(caller.GetVDU().GetLastX()%256))
				//caller.SetMemory(225, uint64(caller.GetVDU().GetLastX()/256))
				break
			}
		case 64795:
			{
				// M/L monitor key in
				for caller.GetMemory(49152) < 128 {
					time.Sleep(time.Millisecond * 50)
				}
				caller.SetMemory(49168, 0)
				break
			}
		case 65385:
			mm := NewMonitor(caller)
			mm.Manual("")
		default:
			{
				found = false
				match = false
				break
			}
		}

	}

	if !found || settings.PureBoot(caller.GetMemIndex()) {
		match = false

		if passtocpu {

			var mlmode bool = !GetCPU(caller).BasicMode // default state
			ov := settings.SlotZPEmu[caller.GetMemIndex()]

			if caller.GetMemoryMap().IntGetZeroPageState(caller.GetMemIndex()) != 0 || settings.PureBoot(caller.GetMemIndex()) {
				//fmt.Println("Enabling zero page RAM")
				//fmt.Println("Disable ZP emu for cpu call")

				tmp := make([]uint64, 256)
				for i, _ := range tmp {
					tmp[i] = caller.GetMemory(i)
				}

				settings.SlotZPEmu[caller.GetMemIndex()] = false

				for i, v := range tmp {
					caller.SetMemory(i, v)
				}

				mlmode = true
			}

			if mlmode {
				if GetColumns(caller) == 40 {
					txt := GETHUD(caller, textselect[caller.GetMemIndex()])
					if txt.Control == nil {
						panic("layer has a nil control surface")
					}
					//_ = ent.GetMemory(49152 + 89)
					txt.SetSubFormat(types.LSF_FIXED_40_24)
					txt.Control.FullRefresh()
				} else {
					txt := GETHUD(caller, textselect[caller.GetMemIndex()])
					if txt.Control == nil {
						panic("layer has a nil control surface")
					}
					//_ = ent.GetMemory(49152 + 89)
					txt.SetSubFormat(types.LSF_FIXED_80_24)
					txt.Control.FullRefresh()
				}
				fmt.Println("MLMODE")
			}

			bus.StartClock(time.Second / 60)

			r := Exec6502Code(caller, 0, 0, 0, addr, 0x80, 0x01ff, mlmode)
			settings.SlotZPEmu[caller.GetMemIndex()] = ov
			if mlmode {
				if GetColumns(caller) == 40 {
					txt := GETHUD(caller, textselect[caller.GetMemIndex()])
					if txt.Control == nil {
						panic("layer has a nil control surface")
					}
					//_ = ent.GetMemory(49152 + 89)
					txt.SetSubFormat(types.LSF_FREEFORM)
				} else {
					txt := GETHUD(caller, textselect[caller.GetMemIndex()])
					if txt.Control == nil {
						panic("layer has a nil control surface")
					}
					//_ = ent.GetMemory(49152 + 89)
					txt.SetSubFormat(types.LSF_FREEFORM)
				}
			}

			//			caller.GetMemoryMap().BlockMapper[caller.GetMemIndex()].EnableBlocks([]string{"apple2iozeropage"})
			//caller.GetMemoryMap().IntSetZeroPageState(caller.GetMemIndex(), 0) // turn off masking

			if r == cpu.FE_BREAKPOINT || r == cpu.FE_CTRLBREAK {
				mm := NewMonitor(caller)
				mm.Break()
			}

			if r == cpu.FE_BREAKPOINT_MEM {
				mm := NewMonitor(caller)
				mm.BreakMemory()
			}

			if r == cpu.FE_BREAKINTERRUPT {
				mm := NewMonitor(caller)
				mm.BreakInterrupt()
			}

			if r == cpu.FE_ILLEGALOPCODE {
				mm := NewMonitor(caller)
				mm.IllegalOpcode()
			}

			if r == cpu.FE_MEMORYPROTECT {
				mm := NewMonitor(caller)
				mm.MemoryProtect()
			}

		}

	}

	return match
}

func Exec6502CodeNB(ent interfaces.Interpretable, a int, x int, y int, pc int, sr int, sp int, mlmode bool) {

	fmt.Printf("Inited CPU for exec @ 0x%.4x\n", pc)

	CPU := GetCPU(ent)

	CPU.ROM = DoCall
	CPU.PC = pc
	CPU.P = sr
	CPU.SP = sp
	CPU.InitialSP = sp
	CPU.A = a
	CPU.X = x
	CPU.Y = y
	CPU.BasicMode = !mlmode
	CPU.Halted = false
	memory.Safe = false

	CPU.Init()

	//go this.AudioFunnel()
	CPU.RAM = CPU.Int.GetMemoryMap()
	CPU.GlobalCycles = 0

	CPU.SetFlag(mos6502.F_R, true)
	CPU.SetFlag(mos6502.F_B, true)
	CPU.SetFlag(mos6502.F_I, true)

	CPU.ResetSliced()

	// if ent.GetMemoryMap().IntGetPDState(ent.GetMemIndex())&128 != 0 {
	// 	CPU.UseProDOS = true
	// }

	// CPU.RegisterCallShim(0xbd00, RWTSInvoker)
	// CPU.RegisterCallShim(0x3e3, RWTSLocateParams)

}

func PlayAudio(ent interfaces.Interpretable, p, f string, wait bool) (int, error) {

	fr, e := files.ReadBytesViaProvider(p, f)

	if e != nil {
		return 0, e
	}

	var b bytes.Buffer

	_, _ = b.Write(fr.Content)

	decoder, err := decoding.NewAudio(&b)

	if err != nil {
		return 0, err
	}

	// Got access to a stream

	var raw []float32

	raw, err = decoder.Samples()
	if err != nil {
		//fmt.Println("Decoder error:", err)
		return 0, err
	}

	if len(raw) == 0 {
		//fmt.Println("Zero samples from decoder!")
		return 0, errors.New("Zero samples")
	}

	waittime := decoder.Length()
	channels := decoder.Channels()
	rate := decoder.SampleRate()

	var fl = make([]float32, len(raw)/int(channels))
	for i := 0; i < len(raw); i++ {
		if i%int(channels) == 0 {
			fl[i/int(channels)] = raw[i]
		}
	}

	//fmt.Printf("%d samples sent\n", len(fl))

	ent.PassWaveBufferNB(0, fl, false, rate)
	if wait {
		ent.WaitAdd(waittime)
	}

	return 0, nil
}

func PlayWave(ent interfaces.Interpretable, p, f string) (int, error) {

	fr, e := files.ReadBytesViaProvider(p, f)

	if e != nil {
		return 0, e
	}

	var b bytes.Buffer

	_, _ = b.Write(fr.Content)

	wr, err := wav.New(&b)
	if err != nil {
		return 0, err
	}

	////fmt.Printntf("WAVE %d channels,  sample rate %d, %d samples\n", wr.NumChannels, wr.SampleRate, wr.Samples)

	var raw []float32

	raw, _ = wr.ReadFloats(wr.Samples)
	for len(raw) > 0 {
		var fl = make([]float32, len(raw)/int(wr.NumChannels))
		for i := 0; i < len(raw); i++ {
			if i%int(wr.NumChannels) == 0 {
				fl[i/int(wr.NumChannels)] = raw[i]
			}
		}

		ent.PassWaveBuffer(0, fl, false, int(wr.SampleRate))

		////fmt.Printntf("WAVE decoded %d samples\n", len(fl))

		raw, _ = wr.ReadFloats(44100 * int(wr.NumChannels))
	}

	return 0, nil
}

func GetRealCursorPos(ent interfaces.Interpretable) (int, int) {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt == nil {
		panic("Expected layer id TEXT not found")
	}
	return txt.Control.CX, txt.Control.CY
}

func GetRealWindow(ent interfaces.Interpretable) (int, int, int, int, int, int) {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt == nil {
		panic("Expected layer id TEXT not found")
	}
	return txt.Control.SX, txt.Control.SY, txt.Control.EX, txt.Control.EY, txt.Control.FontW(), txt.Control.FontH()
}

func SetRealWindow(ent interfaces.Interpretable, sx, sy, ex, ey int) {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt == nil {
		panic("Expected layer id TEXT not found")
	}
	txt.Control.SX, txt.Control.SY, txt.Control.EX, txt.Control.EY = sx, sy, ex, ey
}

func SetRealCursorPos(ent interfaces.Interpretable, x, y int) {
	txt := GETHUD(ent, textselect[ent.GetMemIndex()])
	if txt == nil {
		panic("Expected layer id TEXT not found")
	}
	txt.Control.CX, txt.Control.CY = x, y
	txt.Control.Repos()
}

func GetCRTLine(caller interfaces.Interpretable) string {

	command := ""
	collect := true

	caller.SetBuffer(runestring.NewRuneString())

	for collect {

		caller.Post()

		TextShowCursor(caller)

		for caller.GetMemory(49152) < 128 {
			time.Sleep(10 * time.Millisecond)
		}

		TextHideCursor(caller)

		//if len(caller.GetBuffer().Runes) > 0 {
		ch := rune(caller.GetMemory(49152) & 127)
		caller.SetMemory(49168, 0)

		if caller.GetDialect().IsUpperOnly() && ch >= 'a' && ch <= 'z' {
			ch -= 32
		}

		switch ch {
		case 10:
			{
				//display.SetSuppressFormat(true)
				caller.PutStr("\r\n")
				//display.SetSuppressFormat(false)
				return command
			}
		case 13:
			{
				//display.SetSuppressFormat(true)
				caller.PutStr("\r\n")
				//display.SetSuppressFormat(false)
				return command
			}
		case 8:
			{
				if len(command) > 0 {
					command = utils.Copy(command, 1, len(command)-1)
					caller.Backspace()
					//						display.SetSuppressFormat(true)
					caller.PutStr(" ")
					//display.SetSuppressFormat(false)
					caller.Backspace()
				}
				break
			}
		default:
			{

				if !caller.GetDialect().IsUpperOnly() {
					if (ch >= 'a') && (ch <= 'z') {
						ch -= 32
					} else if (ch >= 'A') && (ch <= 'Z') {
						ch += 32
					}
				}

				ch |= 128

				command = command + string(ch)

				caller.PutStr(string(ch))

				break
			}
		}
		//} else {
		//	time.Sleep(50 * time.Millisecond)
		//}
	}

	return command

}

func SetCOLOR(ent interfaces.Interpretable, c int) {
	ent.SetMemory(48, uint64(17*c))
}

func GetCOLOR(ent interfaces.Interpretable) int {
	return int(ent.GetMemory(48) / 17)
}

func SetSPEED(ent interfaces.Interpretable, speed int) {
	ent.SetMemory(241, uint64((256-speed)%256))
}

func GetSPEED(ent interfaces.Interpretable) int {
	return int(256 - ent.GetMemory(241))
}

func SetSCALE(ent interfaces.Interpretable, s int) {
	ent.SetMemory(231, uint64(s%256))
}

func GetSCALE(ent interfaces.Interpretable) int {
	return int(ent.GetMemory(231) & 0xFF)
}

func SetROT(ent interfaces.Interpretable, s int) {
	ent.SetMemory(249, uint64(s%256))
}

func GetROT(ent interfaces.Interpretable) int {
	return int(ent.GetMemory(249) & 0xFF)
}

func SetHCOLOR(ent interfaces.Interpretable, c int) {

	// shove value in memory as close as possible
	v := 0

	if c >= 16 && c < 24 {
		v = v | 0x200
	} else if c >= 8 && c < 16 {
		v = v | 0x100
	}

	switch c & 7 {
	case 1:
		v += 42
	case 2:
		v += 85
	case 3:
		v += 127
	case 4:
		v += 128
	case 5:
		v += 170
	case 6:
		v += 213
	case 7:
		v += 255
	}

	ent.SetMemory(228, uint64(v))

}

func GetHCOLOR(ent interfaces.Interpretable) int {

	v := ent.GetMemory(228)

	var n int

	switch {
	case v&255 == 255:
		n = 7
	case v&213 == 213:
		n = 6
	case v&170 == 170:
		n = 5
	case v&128 == 128:
		n = 4
	case v&127 == 127:
		n = 3
	case v&85 == 85:
		n = 2
	case v&42 == 42:
		n = 1
	default:
		n = 0
	}

	if v&0xf00 == 0x100 {
		n += 8
	} else if v&0xf00 == 0x200 {
		n += 16
	}

	return n
}

func GetHMASK(ent interfaces.Interpretable) int {
	return int(ent.GetMemory(228) & 0xff)
}

func SetHMASK(ent interfaces.Interpretable, v int) {
	ent.SetMemory(228, uint64(v&0xff))
}
