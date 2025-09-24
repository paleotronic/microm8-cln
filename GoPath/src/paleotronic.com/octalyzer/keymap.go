//go:build !remint
// +build !remint

package main

import (
	fmt2 "fmt"
	"os"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"paleotronic.com/core/editor"
	"paleotronic.com/core/hardware/apple2"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/hardware/servicebus"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/core/vduconst"
	"paleotronic.com/debugger"
	"paleotronic.com/files"
	"paleotronic.com/fmt"
	"paleotronic.com/freeze"
	"paleotronic.com/glumby"
	"paleotronic.com/log"
	"paleotronic.com/octalyzer/backend"
	"paleotronic.com/octalyzer/bus"
	"paleotronic.com/octalyzer/clientperipherals"
	"paleotronic.com/octalyzer/video"
	"paleotronic.com/octalyzer/video/font"
	"paleotronic.com/runestring"
	"paleotronic.com/utils"
)

type MetaInputClosure struct {
	ValidChars []rune
	Delay      time.Duration
	MaxChars   int
}

var inMeta [settings.NUMSLOTS]bool
var closure [settings.NUMSLOTS]MetaInputClosure
var metaStart [settings.NUMSLOTS]time.Time
var collected [settings.NUMSLOTS][]rune
var mouseMoveCamera, mouseMoveCameraAlt bool

func StartMetaMode(index int, vc []rune, msdelay int, triggerchar rune, max int) {

	if settings.DisableMetaMode[index] {
		// Just send key
		RAM.KeyBufferAdd(index, uint64(triggerchar))
		return
	}

	if inMeta[index] {
		EndMetaMode(index)
	}

	collected[index] = make([]rune, 1)
	collected[index][0] = triggerchar
	inMeta[index] = true
	metaStart[index] = time.Now()
	closure[index].ValidChars = vc
	closure[index].MaxChars = max
	closure[index].Delay = time.Millisecond * time.Duration(msdelay)
	clientperipherals.SPEAKER.MakeTone(967, 33)
}

func EndMetaMode(index int) []rune {
	inMeta[index] = false

	// insert to key buffer
	ProcessCollected(index)

	clientperipherals.SPEAKER.MakeTone(488, 33)

	return collected[index]
}

func EndMetaModeQuiet(index int) []rune {
	inMeta[index] = false

	// insert to key buffer
	ProcessCollected(index)

	return collected[index]
}

func SetupKeymapper() glumby.KeyTranslationMap {

	km := glumby.NewDefaultMapper()

	km = append(km, glumby.KeyState{Key: glumby.KeyDown, States: []glumby.Action{glumby.Press, glumby.Repeat, glumby.Release}, Mapping: vduconst.CSR_DOWN})
	km = append(km, glumby.KeyState{Key: glumby.KeyUp, States: []glumby.Action{glumby.Press, glumby.Repeat, glumby.Release}, Mapping: vduconst.CSR_UP})
	km = append(km, glumby.KeyState{Key: glumby.KeyLeft, States: []glumby.Action{glumby.Press, glumby.Repeat, glumby.Release}, Mapping: vduconst.CSR_LEFT})
	km = append(km, glumby.KeyState{Key: glumby.KeyRight, States: []glumby.Action{glumby.Press, glumby.Repeat, glumby.Release}, Mapping: vduconst.CSR_RIGHT})

	km = append(km, glumby.KeyState{Key: glumby.KeyLeft, Modifier: glumby.ModShift, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: vduconst.SHIFT_CSR_LEFT})
	km = append(km, glumby.KeyState{Key: glumby.KeyRight, Modifier: glumby.ModShift, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: vduconst.SHIFT_CSR_RIGHT})
	km = append(km, glumby.KeyState{Key: glumby.KeyUp, Modifier: glumby.ModShift, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: vduconst.SHIFT_CSR_UP})
	km = append(km, glumby.KeyState{Key: glumby.KeyDown, Modifier: glumby.ModShift, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: vduconst.SHIFT_CSR_DOWN})

	km = append(km, glumby.KeyState{Key: glumby.KeyLeft, Modifier: glumby.ModAlt, States: []glumby.Action{glumby.Release}, Mapping: vduconst.CSR_LEFT})
	km = append(km, glumby.KeyState{Key: glumby.KeyRight, Modifier: glumby.ModAlt, States: []glumby.Action{glumby.Release}, Mapping: vduconst.CSR_RIGHT})
	km = append(km, glumby.KeyState{Key: glumby.KeyUp, Modifier: glumby.ModAlt, States: []glumby.Action{glumby.Release}, Mapping: vduconst.CSR_UP})
	km = append(km, glumby.KeyState{Key: glumby.KeyDown, Modifier: glumby.ModAlt, States: []glumby.Action{glumby.Release}, Mapping: vduconst.CSR_DOWN})

	// km = append(km, glumby.KeyState{Key: glumby.KeyLeft, Modifier: glumby.ModControl | glumby.ModShift, States: []glumby.Action{glumby.Press, glumby.Release}, Mapping: vduconst.SHIFT_CTRL_CSR_LEFT})
	// km = append(km, glumby.KeyState{Key: glumby.KeyRight, Modifier: glumby.ModControl | glumby.ModShift, States: []glumby.Action{glumby.Press, glumby.Release}, Mapping: vduconst.SHIFT_CTRL_CSR_RIGHT})
	// km = append(km, glumby.KeyState{Key: glumby.KeyUp, Modifier: glumby.ModControl | glumby.ModShift, States: []glumby.Action{glumby.Press, glumby.Release}, Mapping: vduconst.SHIFT_CTRL_CSR_UP})
	// km = append(km, glumby.KeyState{Key: glumby.KeyDown, Modifier: glumby.ModControl | glumby.ModShift, States: []glumby.Action{glumby.Press, glumby.Release}, Mapping: vduconst.SHIFT_CTRL_CSR_DOWN})

	km = append(km, glumby.KeyState{Key: 32, Modifier: glumby.ModShift, States: []glumby.Action{glumby.Press, glumby.Release}, Mapping: 32})

	//	km = append(km, glumby.KeyState{Key: glumby.KeyLeftShift, Modifier: glumby.ModShift, States: []glumby.Action{glumby.Press, glumby.Release}, Mapping: vduconst.CTRL})
	//km = append(km, glumby.KeyState{Key: glumby.KeyX, Modifier: glumby.ModControl, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: vduconst.CTRL_X})
	//km = append(km, glumby.KeyState{Key: glumby.KeyD, Modifier: glumby.ModControl, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: vduconst.CTRL_D})
	//km = append(km, glumby.KeyState{Key: glumby.KeyV, Modifier: glumby.ModControl, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: vduconst.CTRL_V})

	km = append(km, glumby.KeyState{Key: glumby.KeyEnter, Modifier: glumby.ModShift | glumby.ModControl, States: []glumby.Action{glumby.Press}, Mapping: vduconst.SHIFT_CTRL_ENTER})

	km = append(km, glumby.KeyState{Key: glumby.KeySpace, Modifier: glumby.ModControl, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: vduconst.CTRL_SPACE})
	km = append(km, glumby.KeyState{Key: glumby.KeySpace, Modifier: glumby.ModShift, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: vduconst.SHIFT_SPACE})

	km = append(km, glumby.KeyState{Key: glumby.KeySpace, Modifier: glumby.ModControl | glumby.ModShift, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: vduconst.SHIFT_CTRL_SPACE})
	km = append(km, glumby.KeyState{Key: glumby.KeyEqual, Modifier: glumby.ModControl | glumby.ModShift, States: []glumby.Action{glumby.Press}, Mapping: vduconst.SHIFT_CTRL_PLUS})
	km = append(km, glumby.KeyState{Key: glumby.KeyMinus, Modifier: glumby.ModControl | glumby.ModShift, States: []glumby.Action{glumby.Press}, Mapping: vduconst.SHIFT_CTRL_MINUS})

	// SHIFT_CTRL_OSB & SHIFT_CTRL_CSB
	km = append(km, glumby.KeyState{Key: '[', Modifier: glumby.ModControl | glumby.ModShift, States: []glumby.Action{glumby.Press, glumby.Release}, Mapping: vduconst.SHIFT_CTRL_OSB})
	km = append(km, glumby.KeyState{Key: ']', Modifier: glumby.ModControl | glumby.ModShift, States: []glumby.Action{glumby.Press, glumby.Release}, Mapping: vduconst.SHIFT_CTRL_CSB})
	km = append(km, glumby.KeyState{Key: '\\', Modifier: glumby.ModControl | glumby.ModShift, States: []glumby.Action{glumby.Press, glumby.Release}, Mapping: vduconst.SHIFT_CTRL_BACKSLASH})

	km = append(km, glumby.KeyState{Key: glumby.KeyBackspace, Modifier: glumby.ModControl | glumby.ModShift, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: vduconst.SHIFT_CTRL_BACKSPACE})
	km = append(km, glumby.KeyState{Key: glumby.KeyBackspace, Modifier: glumby.ModControl, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: vduconst.CTRL_BACKSPACE})

	km = append(km, glumby.KeyState{Key: glumby.KeyGraveAccent, Modifier: glumby.ModControl | glumby.ModShift, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: vduconst.CTRL_TWIDDLE})

	km = append(km, glumby.KeyState{Key: glumby.KeyPeriod, Modifier: glumby.ModControl | glumby.ModShift, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: vduconst.SHIFT_CTRL_PERIOD})

	km = append(km, glumby.KeyState{Key: glumby.Key1, Modifier: glumby.ModControl, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: vduconst.CTRL_1})
	km = append(km, glumby.KeyState{Key: glumby.Key2, Modifier: glumby.ModControl, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: vduconst.CTRL_2})
	km = append(km, glumby.KeyState{Key: glumby.Key3, Modifier: glumby.ModControl, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: vduconst.CTRL_3})
	km = append(km, glumby.KeyState{Key: glumby.Key4, Modifier: glumby.ModControl, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: vduconst.CTRL_4})
	km = append(km, glumby.KeyState{Key: glumby.Key5, Modifier: glumby.ModControl, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: vduconst.CTRL_5})
	km = append(km, glumby.KeyState{Key: glumby.Key6, Modifier: glumby.ModControl, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: vduconst.CTRL_6})
	km = append(km, glumby.KeyState{Key: glumby.Key7, Modifier: glumby.ModControl, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: vduconst.CTRL_7})
	km = append(km, glumby.KeyState{Key: glumby.Key8, Modifier: glumby.ModControl, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: vduconst.CTRL_8})

	km = append(km, glumby.KeyState{Key: glumby.KeyUp, Modifier: glumby.ModShift, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: vduconst.PAGE_UP})
	km = append(km, glumby.KeyState{Key: glumby.KeyDown, Modifier: glumby.ModShift, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: vduconst.PAGE_DOWN})
	//~ km = append(km, glumby.KeyState{Key: glumby.KeyLeft, Modifier: glumby.ModShift, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: vduconst.HOME})
	//~ km = append(km, glumby.KeyState{Key: glumby.KeyRight, Modifier: glumby.ModShift, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: vduconst.END})

	km = append(km, glumby.KeyState{Key: glumby.KeyTab, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: 9})

	km = append(km, glumby.KeyState{Key: glumby.Key('['), Modifier: glumby.ModControl, States: []glumby.Action{glumby.Press}, Mapping: vduconst.CTRL_F1})
	km = append(km, glumby.KeyState{Key: glumby.Key(']'), Modifier: glumby.ModControl, States: []glumby.Action{glumby.Press}, Mapping: vduconst.CTRL_F2})

	km = append(km, glumby.KeyState{Key: glumby.KeyF2, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: vduconst.F2})
	km = append(km, glumby.KeyState{Key: glumby.KeyF5, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: vduconst.F5})
	km = append(km, glumby.KeyState{Key: glumby.KeyF6, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: vduconst.F6})
	km = append(km, glumby.KeyState{Key: glumby.KeyF3, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: vduconst.F3})
	km = append(km, glumby.KeyState{Key: glumby.KeyF7, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: vduconst.F7})
	km = append(km, glumby.KeyState{Key: glumby.KeyF8, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: vduconst.F8})
	km = append(km, glumby.KeyState{Key: glumby.KeyF9, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: vduconst.F9})

	//km = append(km, glumby.KeyState{Key: glumby.KeyF, Modifier: glumby.ModAlt, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: vduconst.FREEZE})
	//km = append(km, glumby.KeyState{Key: glumby.KeyT, Modifier: glumby.ModAlt, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: vduconst.THAW})

	km = append(km, glumby.KeyState{Key: glumby.KeySpace, Modifier: glumby.ModAlt | glumby.ModControl, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: vduconst.INVERSE_ON})

	km = append(km, glumby.KeyState{Key: glumby.KeyPageDown, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: vduconst.PAGE_DOWN})
	km = append(km, glumby.KeyState{Key: glumby.KeyPageUp, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: vduconst.PAGE_UP})
	km = append(km, glumby.KeyState{Key: glumby.KeyHome, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: vduconst.HOME})
	km = append(km, glumby.KeyState{Key: glumby.KeyEnd, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: vduconst.END})

	//km = append(km, glumby.KeyState{Key: glumby.KeyV, Modifier: glumby.ModControl, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: vduconst.PASTE})
	km = append(km, glumby.KeyState{Key: glumby.KeyInsert, Modifier: glumby.ModShift, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: vduconst.PASTE})

	km = append(km, glumby.KeyState{Key: glumby.KeyInsert, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: vduconst.INSERT})
	km = append(km, glumby.KeyState{Key: glumby.KeyDelete, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: vduconst.DELETE})

	km = append(km, glumby.KeyState{Key: glumby.KeyEscape, Modifier: glumby.ModShift, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: vduconst.INPUTSWITCH})
	km = append(km, glumby.KeyState{Key: glumby.KeyEscape, Modifier: glumby.ModAlt, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: 27})

	km = append(km, glumby.KeyState{Key: glumby.KeyF1, Modifier: glumby.ModControl, States: []glumby.Action{glumby.Press}, Mapping: vduconst.CTRL_F1})
	km = append(km, glumby.KeyState{Key: glumby.KeyF2, Modifier: glumby.ModControl, States: []glumby.Action{glumby.Press}, Mapping: vduconst.CTRL_F2})

	for i := int(glumby.KeyA); i <= int(glumby.KeyZ); i++ {
		km = append(
			km,
			glumby.KeyState{
				Key:      glumby.Key(i),
				Modifier: glumby.ModAlt | glumby.ModControl | glumby.ModShift,
				States:   []glumby.Action{glumby.Press, glumby.Repeat},
				Mapping:  rune(65 + 1024 + (i - int(glumby.KeyA))),
			})
		// lower
		km = append(
			km,
			glumby.KeyState{
				Key:      glumby.Key(i),
				Modifier: glumby.ModAlt | glumby.ModControl,
				States:   []glumby.Action{glumby.Press, glumby.Repeat},
				Mapping:  rune(97 + 1024 + (i - int(glumby.KeyA))),
			})
	}

	for i := int(glumby.Key0); i <= int(glumby.Key9); i++ {
		// lower
		km = append(
			km,
			glumby.KeyState{
				Key:      glumby.Key(i),
				Modifier: glumby.ModAlt | glumby.ModControl,
				States:   []glumby.Action{glumby.Press, glumby.Repeat},
				Mapping:  rune(48 + 1024 + (i - int(glumby.Key0))),
			})
	}

	// Keypad numbers
	for i := int(glumby.KeyKP0); i <= int(glumby.KeyKP9); i++ {
		// lower
		km = append(
			km,
			glumby.KeyState{
				Key:     glumby.Key(i),
				States:  []glumby.Action{glumby.Press, glumby.Repeat},
				Mapping: rune(48 + (i - int(glumby.KeyKP0))),
			})
	}

	// KP Enter
	km = append(
		km,
		glumby.KeyState{
			Key:     glumby.KeyKPEnter,
			States:  []glumby.Action{glumby.Press, glumby.Repeat},
			Mapping: rune(13),
		})
	// KP +
	km = append(
		km,
		glumby.KeyState{
			Key:     glumby.KeyKPAdd,
			States:  []glumby.Action{glumby.Press, glumby.Repeat},
			Mapping: '+',
		})
	// KP -
	km = append(
		km,
		glumby.KeyState{
			Key:     glumby.KeyKPSubtract,
			States:  []glumby.Action{glumby.Press, glumby.Repeat},
			Mapping: '-',
		})
	// KP *
	km = append(
		km,
		glumby.KeyState{
			Key:     glumby.KeyKPMultiply,
			States:  []glumby.Action{glumby.Press, glumby.Repeat},
			Mapping: '*',
		})
	// KP /
	km = append(
		km,
		glumby.KeyState{
			Key:     glumby.KeyKPDivide,
			States:  []glumby.Action{glumby.Press, glumby.Repeat},
			Mapping: '/',
		})
	// KP .
	km = append(
		km,
		glumby.KeyState{
			Key:     glumby.KeyKPDecimal,
			States:  []glumby.Action{glumby.Press, glumby.Repeat},
			Mapping: '.',
		})
	// KP =
	km = append(
		km,
		glumby.KeyState{
			Key:     glumby.KeyKPEqual,
			States:  []glumby.Action{glumby.Press, glumby.Repeat},
			Mapping: '=',
		})

	for i := int(glumby.KeyA); i <= int(glumby.KeyZ); i++ {
		// lower
		km = append(
			km,
			glumby.KeyState{
				Key:      glumby.Key(i),
				Modifier: glumby.ModControl,
				States:   []glumby.Action{glumby.Press, glumby.Repeat},
				Mapping:  rune(1 + (i - int(glumby.KeyA))),
			})
	}

	for i := int(glumby.KeyA); i <= int(glumby.KeyZ); i++ {
		// lower
		km = append(
			km,
			glumby.KeyState{
				Key:     glumby.Key(i),
				States:  []glumby.Action{glumby.Press, glumby.Repeat},
				Mapping: rune(65 + (i - int(glumby.KeyA))),
			})
	}

	for i := int(glumby.KeyA); i <= int(glumby.KeyZ); i++ {
		// lower
		km = append(
			km,
			glumby.KeyState{
				Key:      glumby.Key(i),
				Modifier: glumby.ModShift,
				States:   []glumby.Action{glumby.Press, glumby.Repeat},
				Mapping:  rune(97 + (i - int(glumby.KeyA))),
			})
	}

	// CTRL+X
	for i := int(glumby.KeyA); i <= int(glumby.KeyZ); i++ {
		// lower
		km = append(
			km,
			glumby.KeyState{
				Key:      glumby.Key(i),
				States:   []glumby.Action{glumby.Press, glumby.Repeat},
				Modifier: glumby.ModControl,
				Mapping:  rune(i - int(glumby.KeyA) + 1),
			})
	}

	for i := int(glumby.KeyA); i <= int(glumby.KeyZ); i++ {
		// lower
		km = append(
			km,
			glumby.KeyState{
				Key:      glumby.Key(i),
				States:   []glumby.Action{glumby.Press, glumby.Repeat},
				Modifier: glumby.ModAlt,
				Mapping:  65 + rune(i-int(glumby.KeyA)),
			})
	}

	//for i := int(glumby.KeyA); i <= int(glumby.KeyZ); i++ {
	//	// lower
	//	km = append(
	//		km,
	//		glumby.KeyState{
	//			Key:      glumby.Key(i),
	//			States:   []glumby.Action{glumby.Press, glumby.Repeat},
	//			Modifier: glumby.ModAlt,
	//			Mapping:  vduconst.OA_A + rune(i-int(glumby.KeyA)),
	//		})
	//}

	km = append(
		km,
		glumby.KeyState{
			Key:      glumby.KeySlash,
			States:   []glumby.Action{glumby.Press, glumby.Repeat},
			Modifier: glumby.ModAlt | glumby.ModShift,
			Mapping:  '?',
		})

	// SHIFT+CTRL+X
	for i := int(glumby.KeyA); i <= int(glumby.KeyZ); i++ {
		// lower
		km = append(
			km,
			glumby.KeyState{
				Key:      glumby.Key(i),
				States:   []glumby.Action{glumby.Press, glumby.Repeat},
				Modifier: glumby.ModControl | glumby.ModShift,
				Mapping:  rune(vduconst.SHIFT_CTRL_A + (i - int(glumby.KeyA))),
			})
	}

	km = append(
		km,
		glumby.KeyState{
			Key:     glumby.KeyCapsLock,
			States:  []glumby.Action{glumby.Press},
			Mapping: vduconst.CAPS_LOCK_ON,
		})

	// Shift+1
	km = append(
		km,
		glumby.KeyState{
			Key:      glumby.Key1,
			Modifier: glumby.ModShift,
			States:   []glumby.Action{glumby.Press, glumby.Repeat},
			Mapping:  '!',
		})
	// Shift+2
	km = append(
		km,
		glumby.KeyState{
			Key:      glumby.Key2,
			Modifier: glumby.ModShift,
			States:   []glumby.Action{glumby.Press, glumby.Repeat},
			Mapping:  '@',
		})
	// Shift+3
	km = append(
		km,
		glumby.KeyState{
			Key:      glumby.Key3,
			Modifier: glumby.ModShift,
			States:   []glumby.Action{glumby.Press, glumby.Repeat},
			Mapping:  '#',
		})
	// Shift+4
	km = append(
		km,
		glumby.KeyState{
			Key:      glumby.Key4,
			Modifier: glumby.ModShift,
			States:   []glumby.Action{glumby.Press, glumby.Repeat},
			Mapping:  '$',
		})
	// Shift+5
	km = append(
		km,
		glumby.KeyState{
			Key:      glumby.Key5,
			Modifier: glumby.ModShift,
			States:   []glumby.Action{glumby.Press, glumby.Repeat},
			Mapping:  '%',
		})
	// Shift+6
	km = append(
		km,
		glumby.KeyState{
			Key:      glumby.Key6,
			Modifier: glumby.ModShift,
			States:   []glumby.Action{glumby.Press, glumby.Repeat},
			Mapping:  '^',
		})
	// Shift+7
	km = append(
		km,
		glumby.KeyState{
			Key:      glumby.Key7,
			Modifier: glumby.ModShift,
			States:   []glumby.Action{glumby.Press, glumby.Repeat},
			Mapping:  '&',
		})
	// Shift+8
	km = append(
		km,
		glumby.KeyState{
			Key:      glumby.Key8,
			Modifier: glumby.ModShift,
			States:   []glumby.Action{glumby.Press, glumby.Repeat},
			Mapping:  '*',
		})
	// Shift+9
	km = append(
		km,
		glumby.KeyState{
			Key:      glumby.Key9,
			Modifier: glumby.ModShift,
			States:   []glumby.Action{glumby.Press, glumby.Repeat},
			Mapping:  '(',
		})
	// Shift+0
	km = append(
		km,
		glumby.KeyState{
			Key:      glumby.Key0,
			Modifier: glumby.ModShift,
			States:   []glumby.Action{glumby.Press, glumby.Repeat},
			Mapping:  ')',
		})

	km = append(
		km,
		glumby.KeyState{
			Key:      glumby.KeyGraveAccent,
			Modifier: glumby.ModShift,
			States:   []glumby.Action{glumby.Press, glumby.Repeat},
			Mapping:  '~',
		})

	km = append(
		km,
		glumby.KeyState{
			Key:      glumby.KeyPeriod,
			Modifier: glumby.ModShift,
			States:   []glumby.Action{glumby.Press, glumby.Repeat},
			Mapping:  '>',
		})

	km = append(
		km,
		glumby.KeyState{
			Key:      glumby.KeyComma,
			Modifier: glumby.ModShift,
			States:   []glumby.Action{glumby.Press, glumby.Repeat},
			Mapping:  '<',
		})

	km = append(
		km,
		glumby.KeyState{
			Key:      glumby.KeySlash,
			Modifier: glumby.ModShift,
			States:   []glumby.Action{glumby.Press, glumby.Repeat},
			Mapping:  '?',
		})

	km = append(
		km,
		glumby.KeyState{
			Key:      glumby.KeySemicolon,
			Modifier: glumby.ModShift,
			States:   []glumby.Action{glumby.Press, glumby.Repeat},
			Mapping:  ':',
		})

	km = append(
		km,
		glumby.KeyState{
			Key:      glumby.KeyApostrophe,
			Modifier: glumby.ModShift,
			States:   []glumby.Action{glumby.Press, glumby.Repeat},
			Mapping:  '"',
		})

	km = append(
		km,
		glumby.KeyState{
			Key:      glumby.Key('['),
			Modifier: glumby.ModShift,
			States:   []glumby.Action{glumby.Press, glumby.Repeat},
			Mapping:  '{',
		})

	km = append(
		km,
		glumby.KeyState{
			Key:      glumby.Key(']'),
			Modifier: glumby.ModShift,
			States:   []glumby.Action{glumby.Press, glumby.Repeat},
			Mapping:  '}',
		})

	km = append(
		km,
		glumby.KeyState{
			Key:      glumby.Key('\\'),
			Modifier: glumby.ModShift,
			States:   []glumby.Action{glumby.Press, glumby.Repeat},
			Mapping:  '|',
		})

	km = append(
		km,
		glumby.KeyState{
			Key:      glumby.Key('-'),
			Modifier: glumby.ModShift,
			States:   []glumby.Action{glumby.Press, glumby.Repeat},
			Mapping:  '_',
		})

	km = append(
		km,
		glumby.KeyState{
			Key:      glumby.Key('='),
			Modifier: glumby.ModShift,
			States:   []glumby.Action{glumby.Press, glumby.Repeat},
			Mapping:  '+',
		})

	// ALT Arrow
	km = append(
		km,
		glumby.KeyState{
			Key:      glumby.KeyLeft,
			Modifier: glumby.ModAlt | glumby.ModControl | glumby.ModShift,
			States:   []glumby.Action{glumby.Press, glumby.Repeat},
			Mapping:  vduconst.ALT_LEFT,
		})

	km = append(
		km,
		glumby.KeyState{
			Key:      glumby.KeyRight,
			Modifier: glumby.ModAlt | glumby.ModControl | glumby.ModShift,
			States:   []glumby.Action{glumby.Press, glumby.Repeat},
			Mapping:  vduconst.ALT_RIGHT,
		})

	km = append(
		km,
		glumby.KeyState{
			Key:      glumby.KeyUp,
			Modifier: glumby.ModAlt | glumby.ModControl | glumby.ModShift,
			States:   []glumby.Action{glumby.Press, glumby.Repeat},
			Mapping:  vduconst.ALT_UP,
		})

	km = append(
		km,
		glumby.KeyState{
			Key:      glumby.KeyDown,
			Modifier: glumby.ModAlt | glumby.ModControl | glumby.ModShift,
			States:   []glumby.Action{glumby.Press, glumby.Repeat},
			Mapping:  vduconst.ALT_DOWN,
		})

	// ALT Arrow
	km = append(
		km,
		glumby.KeyState{
			Key:      glumby.KeyLeft,
			Modifier: glumby.ModAlt,
			States:   []glumby.Action{glumby.Press, glumby.Repeat},
			Mapping:  vduconst.CSR_LEFT,
		})

	km = append(
		km,
		glumby.KeyState{
			Key:      glumby.KeyRight,
			Modifier: glumby.ModAlt,
			States:   []glumby.Action{glumby.Press, glumby.Repeat},
			Mapping:  vduconst.CSR_RIGHT,
		})

	km = append(
		km,
		glumby.KeyState{
			Key:      glumby.KeyUp,
			Modifier: glumby.ModAlt,
			States:   []glumby.Action{glumby.Press, glumby.Repeat},
			Mapping:  vduconst.CSR_UP,
		})

	km = append(
		km,
		glumby.KeyState{
			Key:      glumby.KeyDown,
			Modifier: glumby.ModAlt,
			States:   []glumby.Action{glumby.Press, glumby.Repeat},
			Mapping:  vduconst.CSR_DOWN,
		})

	km = append(
		km,
		glumby.KeyState{
			Key:      glumby.KeyHome,
			Modifier: glumby.ModAlt | glumby.ModControl | glumby.ModShift,
			States:   []glumby.Action{glumby.Press, glumby.Repeat},
			Mapping:  vduconst.ALT_HOME,
		})

	km = append(
		km,
		glumby.KeyState{
			Key:      glumby.KeyLeft,
			Modifier: glumby.ModAlt | glumby.ModShift,
			States:   []glumby.Action{glumby.Press, glumby.Repeat},
			Mapping:  vduconst.CTRL_LEFT,
		})

	km = append(
		km,
		glumby.KeyState{
			Key:      glumby.KeyRight,
			Modifier: glumby.ModAlt | glumby.ModShift,
			States:   []glumby.Action{glumby.Press, glumby.Repeat},
			Mapping:  vduconst.CTRL_RIGHT,
		})

	km = append(
		km,
		glumby.KeyState{
			Key:      glumby.KeyUp,
			Modifier: glumby.ModAlt | glumby.ModShift,
			States:   []glumby.Action{glumby.Press, glumby.Repeat},
			Mapping:  vduconst.CTRL_UP,
		})

	km = append(
		km,
		glumby.KeyState{
			Key:      glumby.KeyDown,
			Modifier: glumby.ModAlt | glumby.ModShift,
			States:   []glumby.Action{glumby.Press, glumby.Repeat},
			Mapping:  vduconst.CTRL_DOWN,
		})

	km = append(
		km,
		glumby.KeyState{
			Key:      glumby.KeyEnd,
			Modifier: glumby.ModAlt | glumby.ModControl,
			States:   []glumby.Action{glumby.Press},
			Mapping:  vduconst.ENDREMOTE,
		})

	km = append(
		km,
		glumby.KeyState{
			Key:     glumby.KeyEscape,
			States:  []glumby.Action{glumby.Press, glumby.Repeat},
			Mapping: rune(27),
		})

	km = append(
		km,
		glumby.KeyState{
			Key:     glumby.KeyEnter,
			States:  []glumby.Action{glumby.Press, glumby.Repeat},
			Mapping: rune(13),
		})

	km = append(
		km,
		glumby.KeyState{
			Key:      glumby.KeyEnter,
			Modifier: glumby.ModAlt,
			States:   []glumby.Action{glumby.Press},
			Mapping:  vduconst.ALT_ENTER,
		})

	km = append(
		km,
		glumby.KeyState{
			Key:     glumby.KeyBackspace,
			States:  []glumby.Action{glumby.Press, glumby.Repeat},
			Mapping: rune(127),
		})

	km = append(
		km,
		glumby.KeyState{
			Key:      glumby.KeyBackspace,
			Modifier: glumby.ModShift,
			States:   []glumby.Action{glumby.Press, glumby.Repeat},
			Mapping:  rune(127),
		})

	km = append(
		km,
		glumby.KeyState{
			Key:      glumby.KeyLeft,
			Modifier: glumby.ModControl | glumby.ModAlt,
			States:   []glumby.Action{glumby.Press},
			Mapping:  vduconst.CTRL_ALT_LEFT,
		})

	km = append(
		km,
		glumby.KeyState{
			Key:      glumby.KeyRight,
			Modifier: glumby.ModControl | glumby.ModAlt,
			States:   []glumby.Action{glumby.Press},
			Mapping:  vduconst.CTRL_ALT_RIGHT,
		})

	km = append(
		km,
		glumby.KeyState{
			Key:      glumby.KeyUp,
			Modifier: glumby.ModControl | glumby.ModAlt,
			States:   []glumby.Action{glumby.Press},
			Mapping:  vduconst.CTRL_ALT_UP,
		})

	km = append(
		km,
		glumby.KeyState{
			Key:      glumby.KeyDown,
			Modifier: glumby.ModControl | glumby.ModAlt,
			States:   []glumby.Action{glumby.Press},
			Mapping:  vduconst.CTRL_ALT_DOWN,
		})

	km = append(
		km,
		glumby.KeyState{
			Key:      glumby.KeyPageUp,
			Modifier: glumby.ModControl | glumby.ModAlt,
			States:   []glumby.Action{glumby.Press},
			Mapping:  vduconst.INPUTSWITCH,
		})

	km = append(km, glumby.KeyState{Key: glumby.Key('['), Modifier: glumby.ModAlt | glumby.ModControl, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: 1024 + '['})
	km = append(km, glumby.KeyState{Key: glumby.Key(']'), Modifier: glumby.ModAlt | glumby.ModControl, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: 1024 + ']'})
	km = append(km, glumby.KeyState{Key: glumby.Key('\\'), Modifier: glumby.ModAlt | glumby.ModControl, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: 1024 + '\\'})
	km = append(km, glumby.KeyState{Key: glumby.Key('`'), Modifier: glumby.ModAlt | glumby.ModControl, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: 1024 + '`'})
	km = append(km, glumby.KeyState{Key: glumby.Key('-'), Modifier: glumby.ModAlt | glumby.ModControl, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: 1024 + '-'})
	km = append(km, glumby.KeyState{Key: glumby.Key('='), Modifier: glumby.ModAlt | glumby.ModControl, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: 1024 + '='})
	km = append(km, glumby.KeyState{Key: glumby.Key(';'), Modifier: glumby.ModAlt | glumby.ModControl, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: 1024 + ';'})
	km = append(km, glumby.KeyState{Key: glumby.Key('\''), Modifier: glumby.ModAlt | glumby.ModControl, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: 1024 + '\''})
	km = append(km, glumby.KeyState{Key: glumby.Key('/'), Modifier: glumby.ModAlt | glumby.ModControl, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: 1024 + '/'})
	km = append(km, glumby.KeyState{Key: glumby.Key(','), Modifier: glumby.ModAlt | glumby.ModControl, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: 1024 + ','})
	km = append(km, glumby.KeyState{Key: glumby.Key('.'), Modifier: glumby.ModAlt | glumby.ModControl, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: 1024 + '.'})

	km = append(km, glumby.KeyState{Key: glumby.Key('1'), Modifier: glumby.ModAlt | glumby.ModControl | glumby.ModShift, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: 1024 + '!'})
	km = append(km, glumby.KeyState{Key: glumby.Key('2'), Modifier: glumby.ModAlt | glumby.ModControl | glumby.ModShift, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: 1024 + '@'})
	km = append(km, glumby.KeyState{Key: glumby.Key('3'), Modifier: glumby.ModAlt | glumby.ModControl | glumby.ModShift, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: 1024 + '#'})
	km = append(km, glumby.KeyState{Key: glumby.Key('4'), Modifier: glumby.ModAlt | glumby.ModControl | glumby.ModShift, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: 1024 + '$'})
	km = append(km, glumby.KeyState{Key: glumby.Key('5'), Modifier: glumby.ModAlt | glumby.ModControl | glumby.ModShift, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: 1024 + '%'})
	km = append(km, glumby.KeyState{Key: glumby.Key('6'), Modifier: glumby.ModAlt | glumby.ModControl | glumby.ModShift, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: 1024 + '^'})
	km = append(km, glumby.KeyState{Key: glumby.Key('7'), Modifier: glumby.ModAlt | glumby.ModControl | glumby.ModShift, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: 1024 + '&'})
	km = append(km, glumby.KeyState{Key: glumby.Key('8'), Modifier: glumby.ModAlt | glumby.ModControl | glumby.ModShift, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: 1024 + '*'})
	km = append(km, glumby.KeyState{Key: glumby.Key('9'), Modifier: glumby.ModAlt | glumby.ModControl | glumby.ModShift, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: 1024 + '('})
	km = append(km, glumby.KeyState{Key: glumby.Key('0'), Modifier: glumby.ModAlt | glumby.ModControl | glumby.ModShift, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: 1024 + ')'})

	km = append(km, glumby.KeyState{Key: glumby.Key('['), Modifier: glumby.ModAlt | glumby.ModControl | glumby.ModShift, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: 1024 + '{'})
	km = append(km, glumby.KeyState{Key: glumby.Key(']'), Modifier: glumby.ModAlt | glumby.ModControl | glumby.ModShift, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: 1024 + '}'})
	km = append(km, glumby.KeyState{Key: glumby.Key('\\'), Modifier: glumby.ModAlt | glumby.ModControl | glumby.ModShift, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: 1024 + '|'})
	km = append(km, glumby.KeyState{Key: glumby.Key('`'), Modifier: glumby.ModAlt | glumby.ModControl | glumby.ModShift, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: 1024 + '~'})
	km = append(km, glumby.KeyState{Key: glumby.Key('-'), Modifier: glumby.ModAlt | glumby.ModControl | glumby.ModShift, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: 1024 + '_'})
	km = append(km, glumby.KeyState{Key: glumby.Key('='), Modifier: glumby.ModAlt | glumby.ModControl | glumby.ModShift, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: 1024 + '+'})
	km = append(km, glumby.KeyState{Key: glumby.Key(';'), Modifier: glumby.ModAlt | glumby.ModControl | glumby.ModShift, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: 1024 + ':'})
	km = append(km, glumby.KeyState{Key: glumby.Key('\''), Modifier: glumby.ModAlt | glumby.ModControl | glumby.ModShift, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: 1024 + '"'})
	km = append(km, glumby.KeyState{Key: glumby.Key('/'), Modifier: glumby.ModAlt | glumby.ModControl | glumby.ModShift, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: 1024 + '?'})
	km = append(km, glumby.KeyState{Key: glumby.Key(','), Modifier: glumby.ModAlt | glumby.ModControl | glumby.ModShift, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: 1024 + '<'})
	km = append(km, glumby.KeyState{Key: glumby.Key('.'), Modifier: glumby.ModAlt | glumby.ModControl | glumby.ModShift, States: []glumby.Action{glumby.Press, glumby.Repeat}, Mapping: 1024 + '>'})

	km = append(km, glumby.KeyState{Key: glumby.Key('1'), Modifier: glumby.ModControl, States: []glumby.Action{glumby.Press}, Mapping: vduconst.CTRL1})
	km = append(km, glumby.KeyState{Key: glumby.Key('2'), Modifier: glumby.ModControl, States: []glumby.Action{glumby.Press}, Mapping: vduconst.CTRL2})
	km = append(km, glumby.KeyState{Key: glumby.Key('3'), Modifier: glumby.ModControl, States: []glumby.Action{glumby.Press}, Mapping: vduconst.CTRL3})
	km = append(km, glumby.KeyState{Key: glumby.Key('5'), Modifier: glumby.ModControl, States: []glumby.Action{glumby.Press}, Mapping: vduconst.CTRL5})
	km = append(km, glumby.KeyState{Key: glumby.Key('6'), Modifier: glumby.ModControl, States: []glumby.Action{glumby.Press}, Mapping: vduconst.CTRL6})
	km = append(km, glumby.KeyState{Key: glumby.Key('7'), Modifier: glumby.ModControl, States: []glumby.Action{glumby.Press}, Mapping: vduconst.CTRL7})
	km = append(km, glumby.KeyState{Key: glumby.Key('8'), Modifier: glumby.ModControl, States: []glumby.Action{glumby.Press}, Mapping: vduconst.CTRL8})
	km = append(km, glumby.KeyState{Key: glumby.Key('9'), Modifier: glumby.ModControl, States: []glumby.Action{glumby.Press}, Mapping: vduconst.CTRL9})
	km = append(km, glumby.KeyState{Key: glumby.Key('0'), Modifier: glumby.ModControl, States: []glumby.Action{glumby.Press}, Mapping: vduconst.CTRL0})
	km = append(km, glumby.KeyState{Key: glumby.Key('1'), Modifier: glumby.ModControl | glumby.ModShift, States: []glumby.Action{glumby.Press}, Mapping: vduconst.SCTRL1})
	km = append(km, glumby.KeyState{Key: glumby.Key('2'), Modifier: glumby.ModControl | glumby.ModShift, States: []glumby.Action{glumby.Press}, Mapping: vduconst.SCTRL2})
	km = append(km, glumby.KeyState{Key: glumby.Key('3'), Modifier: glumby.ModControl | glumby.ModShift, States: []glumby.Action{glumby.Press}, Mapping: vduconst.SCTRL3})
	km = append(km, glumby.KeyState{Key: glumby.Key('4'), Modifier: glumby.ModControl | glumby.ModShift, States: []glumby.Action{glumby.Press}, Mapping: vduconst.SCTRL4})
	km = append(km, glumby.KeyState{Key: glumby.Key('5'), Modifier: glumby.ModControl | glumby.ModShift, States: []glumby.Action{glumby.Press}, Mapping: vduconst.SCTRL5})
	km = append(km, glumby.KeyState{Key: glumby.Key('6'), Modifier: glumby.ModControl | glumby.ModShift, States: []glumby.Action{glumby.Press}, Mapping: vduconst.SCTRL6})
	km = append(km, glumby.KeyState{Key: glumby.Key('7'), Modifier: glumby.ModControl | glumby.ModShift, States: []glumby.Action{glumby.Press}, Mapping: vduconst.SCTRL7})
	km = append(km, glumby.KeyState{Key: glumby.Key('8'), Modifier: glumby.ModControl | glumby.ModShift, States: []glumby.Action{glumby.Press}, Mapping: vduconst.SCTRL8})
	km = append(km, glumby.KeyState{Key: glumby.Key('9'), Modifier: glumby.ModControl | glumby.ModShift, States: []glumby.Action{glumby.Press}, Mapping: vduconst.SCTRL9})
	km = append(km, glumby.KeyState{Key: glumby.Key('0'), Modifier: glumby.ModControl | glumby.ModShift, States: []glumby.Action{glumby.Press}, Mapping: vduconst.SCTRL0})

	// alt + number
	km = append(km, glumby.KeyState{Key: glumby.Key('1'), Modifier: glumby.ModAlt, States: []glumby.Action{glumby.Press}, Mapping: vduconst.ALT1})
	km = append(km, glumby.KeyState{Key: glumby.Key('2'), Modifier: glumby.ModAlt, States: []glumby.Action{glumby.Press}, Mapping: vduconst.ALT2})
	km = append(km, glumby.KeyState{Key: glumby.Key('3'), Modifier: glumby.ModAlt, States: []glumby.Action{glumby.Press}, Mapping: vduconst.ALT3})
	km = append(km, glumby.KeyState{Key: glumby.Key('4'), Modifier: glumby.ModAlt, States: []glumby.Action{glumby.Press}, Mapping: vduconst.ALT4})
	km = append(km, glumby.KeyState{Key: glumby.Key('5'), Modifier: glumby.ModAlt, States: []glumby.Action{glumby.Press}, Mapping: vduconst.ALT5})
	km = append(km, glumby.KeyState{Key: glumby.Key('6'), Modifier: glumby.ModAlt, States: []glumby.Action{glumby.Press}, Mapping: vduconst.ALT6})
	km = append(km, glumby.KeyState{Key: glumby.Key('7'), Modifier: glumby.ModAlt, States: []glumby.Action{glumby.Press}, Mapping: vduconst.ALT7})
	km = append(km, glumby.KeyState{Key: glumby.Key('8'), Modifier: glumby.ModAlt, States: []glumby.Action{glumby.Press}, Mapping: vduconst.ALT8})

	km = append(km, glumby.KeyState{Key: glumby.KeyLeftAlt, Modifier: glumby.ModAlt, States: []glumby.Action{glumby.Press, glumby.Release}, Mapping: vduconst.OPEN_APPLE})
	km = append(km, glumby.KeyState{Key: glumby.KeyRightAlt, Modifier: glumby.ModAlt, States: []glumby.Action{glumby.Press, glumby.Release}, Mapping: vduconst.CLOSE_APPLE})

	km = append(km, glumby.KeyState{Key: 342, States: []glumby.Action{glumby.Press, glumby.Release}, Mapping: vduconst.OPEN_APPLE})
	km = append(km, glumby.KeyState{Key: 346, States: []glumby.Action{glumby.Press, glumby.Release}, Mapping: vduconst.CLOSE_APPLE})

	km = append(km, glumby.KeyState{Key: 0, Modifier: glumby.ModShift | glumby.ModControl, States: []glumby.Action{glumby.Press, glumby.Release}, Mapping: vduconst.SHIFT_CTRL})

	return km

}

func CheckMetaMode(index int) {
	if !inMeta[index] {
		return
	}
	if time.Since(metaStart[index]) > closure[index].Delay {
		EndMetaMode(index)
		return
	}
	if len(collected[index]) >= closure[index].MaxChars {
		EndMetaMode(index)
		return
	}
}

func inRunes(rlist []rune, r rune) bool {
	for _, cr := range rlist {
		if cr == r {
			return true
		}
	}
	return false
}

func GetVideoMode(slotid int) string {

	for _, layer := range GFXSpecs[slotid] {
		if layer.GetActive() {
			id := layer.GetID()
			// return id
			if strings.HasPrefix(id, "HGR") {
				return "HGR"
			} else if strings.HasPrefix(id, "DHR") {
				return "DHGR"
			} else if strings.HasPrefix(id, "XGR") {
				return "XGR"
			} else if strings.HasPrefix(id, "LOGR") || strings.HasPrefix(id, "LGR2") {
				return "LOGR"
			} else if strings.HasPrefix(id, "DLG") {
				return "DLGR"
			} else {
				return id
			}
		}
	}

	return "NONE"

}

var openApplePressed time.Time
var closeApplePressed time.Time

func CheckKeyInserts() {
	for index := 0; index < settings.NUMSLOTS; index++ {

		keycode, subkey := RAM.MetaKeyGet(index)
		if keycode != 0 {
			// Meta key insert
			inMeta[index] = true
			collected[index] = []rune{keycode, subkey}
			metaStart[index] = time.Now()
			EndMetaModeQuiet(index)
		}

	}
}

func OnRawKeyEvent(w *glumby.Window, ch glumby.Key, scancode int, mod glumby.ModifierKey, state glumby.Action) {
	//log2.Printf("Keyevent: ch = %d, scancode = %d, mod = %d, state = %d\n", ch, scancode, mod, state)

	if settings.BlueScreen {
		return
	}

	if ignoreKeyEvents {
		return
	}

	servicebus.SendServiceBusMessage(
		SelectedIndex,
		servicebus.KeyEvent,
		&servicebus.KeyEventData{
			Key:      rune(ch),
			ScanCode: scancode,
			Modifier: servicebus.KeyMod(mod),
			Action:   servicebus.KeyAction(state),
		},
	)
}

var KeyUp, KeyDown, KeyLeft, KeyRight bool

func OnKeyEvent(w *glumby.Window, ch glumby.Key, mod glumby.ModifierKey, state glumby.Action) {

	if settings.BlueScreen || backend.ProducerMain == nil {
		return
	}

	if ignoreKeyEvents {
		return
	}

	if state == glumby.Press || state == glumby.Repeat || state == glumby.Release {

		if inMeta[SelectedIndex] {

			if state != glumby.Release {
				fmt.Printf("ValidChars: %v <- %d\n", closure[SelectedIndex].ValidChars, rune(ch))
				if !inRunes(closure[SelectedIndex].ValidChars, rune(ch)) {
					EndMetaMode(SelectedIndex)
					return
				}
				CheckMetaMode(SelectedIndex)
				if inMeta[SelectedIndex] {
					fmt.Println("end mode")
					collected[SelectedIndex] = append(collected[SelectedIndex], rune(ch))
					metaStart[SelectedIndex] = time.Now() // reset so we can keep entering
					clientperipherals.SPEAKER.MakeTone(768, 16)
					return // we need to exit now so as not to buffer the key
				}
			}
			return

		}

		//log2.Printf("Keyevent: ch = %d, mod = %d, state = %d\n", ch, mod, state)

		// Open Apple + ? keys
		//if ch >= vduconst.OA_A && ch <= vduconst.OA_Z {
		//	if state == glumby.Press {
		//		for x := 0; x < memory.OCTALYZER_NUM_INTERPRETERS; x++ {
		//			RAM.IntSetPaddleButton(x, 0, 1)
		//		}
		//	} else if state == glumby.Release {
		//		for x := 0; x < memory.OCTALYZER_NUM_INTERPRETERS; x++ {
		//			RAM.IntSetPaddleButton(x, 0, 0)
		//		}
		//	}
		//	time.AfterFunc(
		//		500*time.Millisecond,
		//		func() {
		//			for x := 0; x < memory.OCTALYZER_NUM_INTERPRETERS; x++ {
		//				RAM.IntSetPaddleButton(x, 0, 0)
		//			}
		//			//fmt.Println("OPEN APPLE RELEASED")
		//		},
		//	)
		//	ch = ch - vduconst.OA_A + 65
		//}

		if settings.ArrowKeyPaddles && !settings.DisableMetaMode[SelectedIndex] {

			var changed bool

			//log2.Printf("keycode = %d, state = %d", ch, state)

			switch {
			case ch == vduconst.CSR_LEFT && (state == glumby.Press || state == glumby.Repeat):
				KeyLeft = true
				changed = true
			case ch == vduconst.CSR_RIGHT && (state == glumby.Press || state == glumby.Repeat):
				KeyRight = true
				changed = true
			case ch == vduconst.CSR_UP && (state == glumby.Press || state == glumby.Repeat):
				KeyUp = true
				changed = true
			case ch == vduconst.CSR_DOWN && (state == glumby.Press || state == glumby.Repeat):
				KeyDown = true
				changed = true
			case ch == vduconst.CSR_LEFT && state == glumby.Release:
				KeyLeft = false
				changed = true
			case ch == vduconst.CSR_RIGHT && state == glumby.Release:
				KeyRight = false
				changed = true
			case ch == vduconst.CSR_UP && state == glumby.Release:
				KeyUp = false
				changed = true
			case ch == vduconst.CSR_DOWN && state == glumby.Release:
				KeyDown = false
				changed = true
			}

			if changed {
				// set paddles
				var ax, ay float32
				if KeyLeft {
					ax -= 1
				}
				if KeyRight {
					ax += 1
				}
				if KeyUp {
					ay -= 1
				}
				if KeyDown {
					ay += 1
				}
				SendPaddleValue(0, ax)
				SendPaddleValue(1, ay)
				// exit
				return
			}

		}

		if ch == vduconst.OPEN_APPLE {
			//fmt.Println("OPEN APPLE PRESSED")
			var s uint64
			if state == glumby.Press || state == glumby.Repeat {
				s = 1
			} else {
				s = 0
			}
			for x := 0; x < memory.OCTALYZER_NUM_INTERPRETERS; x++ {
				RAM.IntSetPaddleButton(x, 0, s)
			}
			// time.AfterFunc(
			// 	500*time.Millisecond,
			// 	func() {
			// 		for x := 0; x < memory.OCTALYZER_NUM_INTERPRETERS; x++ {
			// 			RAM.IntSetPaddleButton(x, 0, 0)
			// 		}
			// 		//fmt.Println("OPEN APPLE RELEASED")
			// 	},
			// )
			if !settings.PreventSuppressAlt[SelectedIndex] {
				return
			}
		}

		if ch == vduconst.CLOSE_APPLE {
			//fmt.Println("CLOSE APPLE PRESSED")
			var s uint64
			if state == glumby.Press || state == glumby.Repeat {
				s = 1
			} else {
				s = 0
			}
			for x := 0; x < memory.OCTALYZER_NUM_INTERPRETERS; x++ {
				RAM.IntSetPaddleButton(x, 1, s)
			}
			// time.AfterFunc(
			// 	500*time.Millisecond,
			// 	func() {
			// 		for x := 0; x < memory.OCTALYZER_NUM_INTERPRETERS; x++ {
			// 			RAM.IntSetPaddleButton(x, 1, 0)
			// 		}
			// 		//fmt.Println("CLOSE APPLE RELEASED")
			// 	},
			// )
			if !settings.PreventSuppressAlt[SelectedIndex] {
				return
			}
		}

		if state == glumby.Release {
			return
		}

		if ch == vduconst.SHIFT_CTRL_PERIOD && !RAM.IntGetSlotMenu(SelectedIndex) {
			RAM.IntSetSlotMenu(SelectedIndex, true)
			return
		}

		if ch == vduconst.SHIFT_CTRL_P && !settings.DisableMetaMode[SelectedIndex] {

			e := backend.ProducerMain.GetInterpreter(SelectedIndex)
			octafile, err := editor.CreateOctafileFromState(e, "/local")
			if err == nil {
				fmt.Printf("Created %s\n", octafile)
				apple2helpers.OSDShow(e, "microPAK created "+octafile)
			} else {
				fmt.Printf("Failed creating octafile %s: %s\n", octafile, err.Error())
			}
			return

		}

		if ch == vduconst.SHIFT_CTRL_Z && !settings.DisableMetaMode[SelectedIndex] {

			servicebus.InjectServiceBusMessage(
				SelectedIndex,
				servicebus.SpectrumSaveSnapshot,
				"",
			)

			return

		}

		if ch == vduconst.SHIFT_CTRL_E && !settings.DisableMetaMode[SelectedIndex] {
			settings.ArrowKeyPaddles = !settings.ArrowKeyPaddles
			e := backend.ProducerMain.GetInterpreter(SelectedIndex)
			if settings.ArrowKeyPaddles {
				apple2helpers.OSDShow(e, "Arrow keys: Paddles")
			} else {
				apple2helpers.OSDShow(e, "Arrow keys: Arrows")
			}
			return
		}

		if ch == vduconst.SHIFT_CTRL_SPACE || (ch == vduconst.DPAD_START && !settings.DisableMetaMode[SelectedIndex]) {
			e := backend.ProducerMain.GetInterpreter(SelectedIndex)
			for e.GetChild() != nil {
				e = e.GetChild()
			}
			//fmt.Println("PAUSE/UNPAUSE", e.IsPaused())
			if e.IsWaitingForWorld() {
				e.ResumeTheWorld()
				bus.Sync()
			} else {
				e.StopTheWorld()
				bus.Sync()
			}
			return
		}

		if ch == vduconst.SHIFT_CTRL_PLUS {
			fmt.Println("vol +")
			e := backend.ProducerMain.GetInterpreter(SelectedIndex)
			clientperipherals.SPEAKER.Mixer.VolumeUp()
			apple2helpers.OSDShowProgress(e, "VOLUME ", clientperipherals.SPEAKER.Mixer.GetVolume())
			time.AfterFunc(3*time.Second,
				func() {
					apple2helpers.OSDPanel(e, false)
				},
			)
			return
		}

		if ch == vduconst.SHIFT_CTRL_MINUS {
			fmt.Println("vol -")
			e := backend.ProducerMain.GetInterpreter(SelectedIndex)
			clientperipherals.SPEAKER.Mixer.VolumeDown()
			apple2helpers.OSDShowProgress(e, "VOLUME ", clientperipherals.SPEAKER.Mixer.GetVolume())
			time.AfterFunc(3*time.Second,
				func() {
					apple2helpers.OSDPanel(e, false)
				},
			)
			return
		}

		if ch == vduconst.SHIFT_CTRL_Y && !settings.DisableMetaMode[SelectedIndex] {
			clientperipherals.SPEAKER.Mixer.DumpState()
		}

		if ch == vduconst.SHIFT_CTRL_C && !settings.DisableMetaMode[SelectedIndex] {
			settings.HasCPUBreak[SelectedIndex] = true
			return
		}

		if ch == vduconst.SHIFT_CSR_LEFT && !settings.DisableMetaMode[SelectedIndex] {
			go func() {
				e := backend.ProducerMain.GetInterpreter(SelectedIndex)
				//apple2helpers.GetCPU(e).Halted = true
				if servicebus.HasReceiver(SelectedIndex, servicebus.PlayerBackstep) {
					servicebus.SendServiceBusMessage(
						SelectedIndex,
						servicebus.PlayerBackstep,
						5000,
					)
				} else if e.IsRecordingVideo() {
					apple2helpers.OSDShow(e, "Flashback 5 seconds")
					e.BackstepVideo(5000)
				} else {
					apple2helpers.OSDShow(e, "Live rewind not enabled")
				}
			}()
			return
		}

		if ch == vduconst.SHIFT_CTRL_OSB {
			e := backend.ProducerMain.GetInterpreter(SelectedIndex)
			//apple2helpers.GetCPU(e).Halted = true
			e.BackVideo()
			return
		}

		if ch == vduconst.SHIFT_CTRL_CSB {
			e := backend.ProducerMain.GetInterpreter(SelectedIndex)
			//apple2helpers.GetCPU(e).Halted = true
			e.ForwardVideo()
			return
		}

		if settings.PureBoot(SelectedIndex) && ch == vduconst.SHIFT_CTRL_V && !settings.DisableMetaMode[SelectedIndex] {
			data, err := clipboard.ReadAll()
			fmt.Printf("Data=%v, err=%v\n", data, err)
			if err == nil {
				r := runestring.Cast(string(data))
				ee := backend.ProducerMain.GetInterpreter(SelectedIndex)
				for ee.GetChild() != nil {
					ee = ee.GetChild()
				}
				ee.SetPasteBuffer(r)
				fmt.Println("Pasting some data into pureboot...")
			}
			return
		}

		if ch == vduconst.SHIFT_CTRL_BACKSLASH {
			SnapLayers()
			fmt.Println("Layer snaps taken")
			return
		}

		if ch == vduconst.DPAD_SELECT && !RAM.IntGetSlotInterrupt(SelectedAudioIndex) {
			ch = vduconst.CTRL_TWIDDLE
		}

		//fmt.Printf("Slot interrupt state for %d: %v\n", SelectedAudioIndex, RAM.IntGetSlotInterrupt(SelectedAudioIndex))

		if ch == vduconst.CTRL_TWIDDLE && !RAM.IntGetSlotInterrupt(SelectedAudioIndex) {

			//hardware.DiskSwap(backend.ProducerMain.GetInterpreter(SelectedIndex))
			//control.CatalogPresent(backend.ProducerMain.GetInterpreter(SelectedIndex))
			RAM.IntSetSlotInterrupt(SelectedAudioIndex, true)

			fmt.Printf("SLOT INTERRUPT SENT TO SLOT #%d\n", SelectedAudioIndex)

			return

		}

		if ch == vduconst.SHIFT_CTRL_H && !RAM.IntGetHelpInterrupt(SelectedIndex) && !settings.DisableMetaMode[SelectedIndex] {

			//hardware.DiskSwap(backend.ProducerMain.GetInterpreter(SelectedIndex))
			//control.CatalogPresent(backend.ProducerMain.GetInterpreter(SelectedIndex))
			RAM.IntSetHelpInterrupt(SelectedIndex, true)

			fmt.Printf("HELP INTERRUPT SENT TO SLOT #%d\n", SelectedIndex)

			return

		}

		if ch == vduconst.SHIFT_CTRL_X && !settings.DisableMetaMode[SelectedIndex] {
			mouseMoveCamera = !mouseMoveCamera
			s := "Enabled"
			if !mouseMoveCamera {
				s = "Disabled"
			}
			e := backend.ProducerMain.GetInterpreter(SelectedIndex)
			apple2helpers.OSDShow(e, "Mouse camera control: "+s)
			return
		}

		if ch == vduconst.OPEN_APPLE && mouseMoveCamera {
			mouseMoveCameraAlt = !mouseMoveCameraAlt
			s := "Orbit"
			if mouseMoveCameraAlt {
				s = "Zoom and Rotate"
			}
			e := backend.ProducerMain.GetInterpreter(SelectedIndex)
			apple2helpers.OSDShow(e, "Mouse camera control: "+s)
			return
		}

		if ch == vduconst.SHIFT_CTRL_B && !settings.DisableMetaMode[SelectedIndex] && (settings.PureBoot(SelectedIndex) || backend.ProducerMain.GetInterpreter(SelectedIndex).IsPlayingVideo()) {
			if debugger.DebuggerInstance.IsAttached() {
				//dbg.PrintStack()
				settings.DebuggerAttachSlot = -1
				debugger.DebuggerInstance.ContinueCPU()
				debugger.DebuggerInstance.Detach()
			} else {
				settings.DebuggerAttachSlot = SelectedIndex + 1
				utils.OpenURL(fmt.Sprintf("http://localhost:%d/?attach=%d", settings.DebuggerPort, settings.DebuggerAttachSlot))
			}
			return
		}

		// if ch == vduconst.F7 {
		// 	settings.OptimalBitTiming = 28
		// 	log2.Println("timing 3.5")
		// 	return
		// }

		// if ch == vduconst.F8 {
		// 	settings.OptimalBitTiming = 32
		// 	log2.Println("timing 4.0")
		// 	return
		// }

		// if ch == vduconst.F9 {
		// 	filepath := files.GetUserDirectory(files.BASEDIR) + "/Saves"
		// 	os.MkdirAll(filepath, 0755)
		// 	filename := filepath + "/octalyzer.frz"
		// 	fmt.Printf("Saving STATE to %s\n", filename)
		// 	ent := backend.ProducerMain.GetInterpreter(SelectedIndex)
		// 	f := freeze.NewFreezeState(ent)
		// 	f.SaveToFile(filename)
		// 	return
		// }

		if ch == vduconst.SHIFT_CTRL_BACKSPACE && !settings.BlockCSR[SelectedIndex] {
			//RAM.SlotReset(SelectedIndex)
			servicebus.SendServiceBusMessage(
				SelectedIndex,
				servicebus.RecorderTerminate,
				0,
			)

			if settings.Pakfile[SelectedIndex] != "" {
				settings.MicroPakPath = settings.Pakfile[SelectedIndex]
			}
			RAM.IntSetSlotRestart(SelectedIndex, true)
			time.Sleep(2 * time.Millisecond)
			//go backend.ProducerMain.RebootVM(SelectedIndex)
			return
		}

		if ch == vduconst.CTRL_BACKSPACE {
			//RAM.IntSetCPUBreak(SelectedIndex, true)
			e := backend.ProducerMain.GetInterpreter(SelectedIndex)
			cpu := apple2helpers.GetCPU(e)
			cpu.PullResetLine()
			return
		}

		if ch >= glumby.KeyKP0 && ch <= glumby.KeyKP9 {
			buffer = (10 * buffer) + int(ch) - int(glumby.KeyKP0)
			//fmt.Printf("Buffer is %d\n", buffer)
		} else if ch == glumby.KeyKPEnter {
			if buffer >= 1056 && buffer <= 1056+127 {
				ch = glumby.Key(buffer)
			}
			buffer = 0
		}

		// if ch >= vduconst.F5 && ch <= vduconst.F8 {

		// if ch == vduconst.F5 {
		// 	settings.RotationOrder -= 1
		// 	if settings.RotationOrder < 0 {
		// 		settings.RotationOrder += 1
		// 	}
		// 	log2.Printf("rorder at %s", settings.RotationOrder)
		// 	return
		// }

		// if ch == vduconst.F6 {
		// 	settings.RotationOrder += 1
		// 	if settings.RotationOrder > 11 {
		// 		settings.RotationOrder -= 1
		// 	}
		// 	log2.Printf("rorder at %s", settings.RotationOrder)
		// 	return
		// }

		// 	if ch == vduconst.F6 {
		// 		fxcam[0][0].SetNear(fxcam[0][0].GetNear() + 1)
		// 	}

		// 	if ch == vduconst.F7 {
		// 		fxcam[0][0].SetNear(fxcam[0][0].GetNear() - 10)
		// 	}

		// 	if ch == vduconst.F8 {
		// 		fxcam[0][0].SetNear(fxcam[0][0].GetNear() + 10)
		// 	}

		// 	fmt.Printf("Near = %f\n", fxcam[0][0].GetNear())

		// 	return

		// }

		// this can be changed using the camera.select{n} function
		index := 0
		//i := int(RAM.ReadGlobal(RAM.MEMBASE(index) + memory.OCTALYZER_CAMERA_GFX_INDEX))

		if ch >= vduconst.SCTRL1 && ch <= vduconst.SCTRL0 {
			//if s8webclient.CONN.IsAuthenticated() {
			slotid := int(ch - vduconst.SCTRL1)
			backend.ProducerMain.Select(slotid)
			SelectedCamera = slotid
			SelectedIndex = slotid
			SelectedAudioIndex = slotid
			clientperipherals.Context = slotid
			clientperipherals.SPEAKER.SelectChannel(SelectedAudioIndex)
			//}
			return
		} else if ch == vduconst.CTRL_ALT_UP {
			fxcam[SelectedCamera][SelectedCameraIndex].Ascend(8)
		} else if ch == vduconst.CTRL_ALT_DOWN {
			fxcam[SelectedCamera][SelectedCameraIndex].Ascend(-8)
		} else if ch == vduconst.CTRL_ALT_RIGHT {
			fxcam[SelectedCamera][SelectedCameraIndex].Strafe(8)
		} else if ch == vduconst.CTRL_ALT_LEFT {
			fxcam[SelectedCamera][SelectedCameraIndex].Strafe(-8)
		} else if ch == vduconst.ALT_HOME {
			fxcam[SelectedCamera][SelectedCameraIndex].ResetALL()
		} else if ch == vduconst.CTRL_UP {
			fxcam[SelectedCamera][SelectedCameraIndex].SetZoom(fxcam[SelectedCamera][SelectedCameraIndex].GetZoom() * 1.1)
			//fxcam[i].Ascend( 8 )
		} else if ch == vduconst.CTRL_DOWN {
			fxcam[SelectedCamera][SelectedCameraIndex].SetZoom(fxcam[SelectedCamera][SelectedCameraIndex].GetZoom() / 1.1)
			//fxcam[i].Ascend( -8 )
		} else if ch == vduconst.CTRL_RIGHT {
			fxcam[SelectedCamera][SelectedCameraIndex].Rotate3DZ(1)
			//fxcam[i].Strafe(8)
		} else if ch == vduconst.CTRL_LEFT {
			fxcam[SelectedCamera][SelectedCameraIndex].Rotate3DZ(-1)
			//fxcam[i].Strafe(-8)
		} else if ch == vduconst.ALT_RIGHT {
			fxcam[SelectedCamera][SelectedCameraIndex].Orbit(1, 0)
		} else if ch == vduconst.ALT_LEFT {
			fxcam[SelectedCamera][SelectedCameraIndex].Orbit(-1, 0)
		} else if ch == vduconst.ALT_DOWN {
			fxcam[SelectedCamera][SelectedCameraIndex].Orbit(0, 1)
		} else if ch == vduconst.ALT_UP {
			fxcam[SelectedCamera][SelectedCameraIndex].Orbit(0, -1)
		} else if ch == vduconst.CAPS_LOCK_ON {
			fmt.Println("Toggle case")
			FlipCase = !FlipCase
			if FlipCase {
				RAM.WriteGlobal(index, RAM.MEMBASE(index)+memory.OCTALYZER_KEYCASE_STATE, 1)
			} else {
				RAM.WriteGlobal(index, RAM.MEMBASE(index)+memory.OCTALYZER_KEYCASE_STATE, 0)
			}
		} else if ch == vduconst.CTRL_F1 {
			//backend.ProducerMain.SetInputContext(0)
			//SelectedCamera = 0
			apple2.PDLTIME += 50
			fmt.Printf("Paddle timing value = %f\n", apple2.PDLTIME)
		} else if ch == vduconst.CTRL_F2 {
			//backend.ProducerMain.SetInputContext(1)
			//SelectedCamera = 1
			apple2.PDLTIME -= 50
			fmt.Printf("Paddle timing value = %f\n", apple2.PDLTIME)
		} else if ch == vduconst.ALT_ENTER {

			//w.SetFullscreen(!w.GetFullscreen())
			settings.Windowed = w.GetFullscreen()
			UpdatePixelSize(w)
			settings.FrameSkip = settings.DefaultFrameSkip

		} else if ch == vduconst.ENDREMOTE {
			backend.ProducerMain.EndRemotesNeeded = true
			for i := 1; i < memory.OCTALYZER_NUM_INTERPRETERS; i++ {
				GraphicsLayers[i] = make([]*video.GraphicsLayer, 0)
				HUDLayers[i] = make([]*video.TextLayer, 0)
				GFXSpecs[i] = make([]*types.LayerSpecMapped, 0)
				HUDSpecs[i] = make([]*types.LayerSpecMapped, 0)
			}
		} else if ch == vduconst.SHIFT_CTRL_Y {
			if strings.HasPrefix(settings.SpecFile[SelectedIndex], "apple2") {
				servicebus.SendServiceBusMessage(
					SelectedIndex,
					servicebus.DiskIIExchangeDisks,
					"Disk swap",
				)
				e := backend.ProducerMain.GetInterpreter(SelectedIndex)
				apple2helpers.OSDShow(e, "Apple II swap disks")
			}
		} else if ch == vduconst.SHIFT_CTRL_I {
			StartMetaMode(SelectedIndex, []rune{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}, 1000, rune(ch), 2)
		} else if ch == vduconst.SHIFT_CTRL_F {
			StartMetaMode(SelectedIndex, []rune{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}, 1000, rune(ch), 2)
			//		} else if ch == vduconst.SHIFT_CTRL_B {
			//			StartMetaMode(SelectedIndex, []rune{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}, 1000, rune(ch), 3)
		} else if ch == vduconst.SHIFT_CTRL_S {
			StartMetaMode(SelectedIndex, []rune{'1', '2', '3', '4', '5', '6', '7', '8'}, 1000, rune(ch), 2)
		} else if ch == vduconst.SHIFT_CTRL_L {
			StartMetaMode(SelectedIndex, []rune{'1', '2', '3', '4', '5', '6', '7', '8'}, 1000, rune(ch), 2)
		} else if ch == vduconst.SHIFT_CTRL_D {
			StartMetaMode(SelectedIndex, []rune{'1', '2', '3', '4', '5', '6', '7', '8', '9'}, 1000, rune(ch), 2)
		} else if ch == vduconst.SHIFT_CTRL_J {
			StartMetaMode(SelectedIndex, []rune{'1', '2'}, 1000, rune(ch), 2)
		} else if ch == vduconst.SHIFT_CTRL_R {
			StartMetaMode(SelectedIndex, []rune{'W', 'R', 'F', 'f', 'B', 'S', 'w', 'r', 'b', 's', ' ', 'A', 'a', 'P', 'p'}, 1000, rune(ch), 2)
		} else if ch == vduconst.SHIFT_CTRL_N {
			StartMetaMode(SelectedIndex, []rune{27}, 1000, rune(ch), 2)
		} else if ch == vduconst.SHIFT_CTRL_W {
			StartMetaMode(SelectedIndex, []rune{'1', '2', '3', '4', '5', '6'}, 1000, rune(ch), 2)
		} else if ch == vduconst.SHIFT_CTRL_G {
			StartMetaMode(SelectedIndex, []rune{'1', '2', '3', '4', '5', '6'}, 1000, rune(ch), 2)
		} else if ch == vduconst.SHIFT_CTRL_T {
			StartMetaMode(SelectedIndex, []rune{'1', '2', '3', '4'}, 1000, rune(ch), 2)
		} else if ch == vduconst.SHIFT_CTRL_A {
			StartMetaMode(SelectedIndex, []rune{'1', '2', '3', '4', '5'}, 1000, rune(ch), 2)
		} else if ch == vduconst.SHIFT_CTRL_M {
			StartMetaMode(SelectedIndex, []rune{'1', '2', '3', '4', '5', '6'}, 1000, rune(ch), 2)
		} else {

			e := backend.ProducerMain.GetInterpreter(SelectedIndex)
			if e.IsPlayingVideo() {
				p := e.GetPlayer()
				switch ch {
				case '1':
					p.SetTimeShift(4)
					p.SetBackwards(true)
				case '2':
					p.SetTimeShift(2)
					p.SetBackwards(true)
				case '3':
					p.SetTimeShift(1)
					p.SetBackwards(true)
				case '4':
					p.SetTimeShift(0.5)
					p.SetBackwards(true)
				case '5':
					p.SetTimeShift(0.25)
					p.SetBackwards(true)
				case '6':
					p.SetTimeShift(0.25)
					p.SetBackwards(false)
				case '7':
					p.SetTimeShift(0.5)
					p.SetBackwards(false)
				case '8':
					p.SetTimeShift(1)
					p.SetBackwards(false)
				case '9':
					p.SetTimeShift(2)
					p.SetBackwards(false)
				case '0':
					p.SetTimeShift(4)
					p.SetBackwards(false)
				case vduconst.SHIFT_CTRL_B:
					// do nothing here
				default:
					fmt.Println("Attempting break into video...")
					e.BreakIntoVideo()
				}

				return
			}

			//			if FlipCase {
			//				switch {
			//				case ch >= 'a' && ch <= 'z':
			//					ch -= 32
			//				case ch >= 'A' && ch <= 'Z':
			//					ch += 32
			//				}
			//			}

			if settings.MonitorActive[SelectedIndex] {

				settings.MonitorKeyAdd(SelectedIndex, rune(ch))

			} else {

				if RAM.IntGetUppercaseOnly(SelectedIndex) && ch >= 'a' && ch <= 'z' {
					ch -= 32
				}

				//for x := 0; x < memory.OCTALYZER_NUM_INTERPRETERS; x++ {
				if ch == vduconst.CTRL_C {
					RAM.KeyBufferEmpty(SelectedIndex)
				}

				if ch >= 'a' && ch <= 'z' && (RAM.IntGetPaddleButton(SelectedIndex, 0) != 0 || RAM.IntGetPaddleButton(SelectedIndex, 1) != 0) {
					ch -= 32
				}

				for x := 0; x < memory.OCTALYZER_NUM_INTERPRETERS; x++ {
					RAM.KeyBufferAdd(x, uint64(ch))
				}
				//pcam[0].ShakeFrames = 12
				//pcam[0].ShakeMax = 25
				//}
			}

		}

	}
}

func InsertKey(ch rune) {

	log.Printf("Meta insert code %x\n", ch)

	for x := 0; x < memory.OCTALYZER_NUM_INTERPRETERS; x++ {
		RAM.KeyBufferAdd(x, uint64(ch))
		servicebus.InjectServiceBusMessage(
			0,
			servicebus.BBCKeyPressed,
			ch,
		)
	}
}

func ProcessCollected(index int) {
	if len(collected[index]) == 0 {
		return
	}

	if len(collected[index]) == 1 {
		InsertKey(collected[index][0])
		collected[index] = make([]rune, 0)
		return
	}

	if collected[index][0] >= vduconst.SCTRL1 && collected[index][0] <= vduconst.SCTRL8 {
		slotid := int(collected[index][0] - vduconst.SCTRL1)
		backend.ProducerMain.Select(slotid)
		SelectedCamera = slotid
		SelectedIndex = slotid
		SelectedAudioIndex = slotid
		clientperipherals.Context = slotid
		clientperipherals.SPEAKER.SelectChannel(SelectedAudioIndex)
		collected[index] = make([]rune, 0)
		return
	}

	// valid meta
	switch collected[index][0] {

	case vduconst.SHIFT_CTRL_I:
		f := float32(100)
		l := utils.StrToInt(string(collected[index][1:]))
		switch l {
		case 0:
			f = 100
		case 1:
			f = 88
		case 2:
			f = 77
		case 3:
			f = 66
		case 4:
			f = 55
		case 5:
			f = 44
		case 6:
			f = 33
		case 7:
			f = 22
		case 8:
			f = 11
		case 9:
			f = 0
		default:
			return
		}
		settings.ScanLineIntensity = f / 100
		e := backend.ProducerMain.GetInterpreter(index)
		apple2helpers.OSDShow(e, fmt.Sprintf("Scanline intensity %d", l))

	case vduconst.SHIFT_CTRL_F:
		s := string(collected[index][1:])
		i := utils.StrToInt(s) % 16
		f := settings.AuxFonts[SelectedIndex]
		if i >= 0 && i < len(f) {
			fn, err := font.LoadFromFile(f[i])
			if err == nil {
				settings.DefaultFont[SelectedIndex] = fn
				settings.ForceTextVideoRefresh = true
				e := backend.ProducerMain.GetInterpreter(index)
				apple2helpers.OSDShow(e, "Switch to font bank #"+s)
			}
		}

	// case vduconst.SHIFT_CTRL_B:
	// 	s := string(collected[index][1:])
	// 	InsertKey(rune(vduconst.BGCOLOR0 + utils.StrToInt(s)%16))

	case vduconst.SHIFT_CTRL_S:
		s := string(collected[index][1:])
		i := utils.StrToInt(s) - 1
		filepath := files.GetDiskSaveDirectory(index)
		os.MkdirAll(filepath, 0755)
		filename := filepath + fmt.Sprintf("/microM8%d.frz", i)
		fmt.Printf("Saving STATE to %s\n", filename)
		ent := backend.ProducerMain.GetInterpreter(index)
		f := freeze.NewFreezeState(ent, false)
		f.SaveToFile(filename)
		return

	case vduconst.SHIFT_CTRL_L:
		s := string(collected[index][1:])
		i := utils.StrToInt(s) - 1
		filepath := files.GetDiskSaveDirectory(index)
		os.MkdirAll(filepath, 0755)
		filename := filepath + fmt.Sprintf("/microM8%d.frz", i)
		if files.Exists(filename) {
			fmt2.Printf("Restoring STATE from %s\n", filename)
			settings.PureBootRestoreState[index] = filename
			RAM.IntSetSlotRestart(index, true)
			return
		} else {
			fmt2.Println("No save in slot")
			return
		}

	case vduconst.SHIFT_CTRL_J:
		s := string(collected[index][1:])
		i := utils.StrToInt(s) - 1
		switch i {
		case 0:
			RAM.PaddleMap[index][1] = 1
			RAM.PaddleMap[index][0] = 0
			e := backend.ProducerMain.GetInterpreter(index)
			apple2helpers.OSDShow(e, "Joystick Axis: Normal")
		case 1:
			RAM.PaddleMap[index][1] = 0
			RAM.PaddleMap[index][0] = 1
			e := backend.ProducerMain.GetInterpreter(index)
			apple2helpers.OSDShow(e, "Joystick Axis: Switch X/Y")
		}

	case vduconst.SHIFT_CTRL_M:

		s := string(collected[index][1:])
		v := settings.MouseMode(utils.StrToInt(s) - 1)

		switch v {
		case settings.MM_MOUSE_JOYSTICK:
		case settings.MM_MOUSE_DPAD:
		case settings.MM_MOUSE_GEOS:
		case settings.MM_MOUSE_DDRAW:
		case settings.MM_MOUSE_CAMERA:
		case settings.MM_MOUSE_OFF:
		default:
			return
		}

		settings.SetMouseMode(v)

		e := backend.ProducerMain.GetInterpreter(index)
		apple2helpers.OSDShow(e, "Mouse Mode: "+v.String())

	case vduconst.SHIFT_CTRL_A:

		s := string(collected[index][1:])
		v := int(utils.StrToInt(s) - 1)

		if settings.AllowPerspectiveChanges {

			if v < len(aspectRatios) && v >= 0 {
				aspectRatioIndex[index] = v
				r := aspectRatios[v]
				SetSlotAspect(index, r)

				e := backend.ProducerMain.GetInterpreter(index)
				apple2helpers.OSDShow(e, "Aspect: "+utils.FloatToStr(r))
			}
		}

	case vduconst.SHIFT_CTRL_T:

		s := string(collected[index][1:])
		v := settings.VideoPaletteTint(utils.StrToInt(s) - 1)
		if v < settings.VPT_MAX {
			RAM.IntSetVideoTint(index, v)

			e := backend.ProducerMain.GetInterpreter(index)
			apple2helpers.OSDShow(e, "Color Mode: "+v.String())
		}

	case vduconst.SHIFT_CTRL_G:

		s := string(collected[index][1:])
		v := settings.VideoMode(utils.StrToInt(s) - 1)

		if v < settings.VM_MAX_MODE {

			mode := GetVideoMode(index)

			if mode == "HGR" || mode == "XGR" {

				RAM.IntSetHGRRender(index, v)

				e := backend.ProducerMain.GetInterpreter(index)
				apple2helpers.OSDShow(e, "Video: "+v.String())

				// force recalibration
				settings.FrameSkip = settings.DefaultFrameSkip
			} else if mode == "SCRN" {

				v = v % 3

				RAM.IntSetSpectrumRender(index, v)

				e := backend.ProducerMain.GetInterpreter(index)
				apple2helpers.OSDShow(e, "Video: "+v.String())

				// force recalibration
				settings.FrameSkip = settings.DefaultFrameSkip

			} else if mode == "SHR1" {

				v = v % 3

				RAM.IntSetSHRRender(index, v)

				e := backend.ProducerMain.GetInterpreter(index)
				apple2helpers.OSDShow(e, "Video: "+v.String())

				// force recalibration
				settings.FrameSkip = settings.DefaultFrameSkip

			} else if mode == "DHGR" {

				RAM.IntSetDHGRRender(index, v)

				e := backend.ProducerMain.GetInterpreter(index)
				apple2helpers.OSDShow(e, "Video: "+v.String())

				// force recalibration
				settings.FrameSkip = settings.DefaultFrameSkip
			} else if mode == "LOGR" || mode == "DLGR" {

				if v == settings.VM_DOTTY || v == settings.VM_MONO_DOTTY {
					v = settings.VM_FLAT
				}

				if v == settings.VM_MONO_FLAT {
					v = settings.VM_FLAT
				}

				if v == settings.VM_MONO_VOXELS {
					v = settings.VM_VOXELS
				}

				if v == settings.VM_FLAT || v == settings.VM_VOXELS {
					RAM.IntSetGRRender(index, v)

					e := backend.ProducerMain.GetInterpreter(index)
					apple2helpers.OSDShow(e, "Video: "+v.String())

					// force recalibration
					settings.FrameSkip = settings.DefaultFrameSkip
				}
			}

		}

		return

	case vduconst.SHIFT_CTRL_D:
		s := string(collected[index][1:])
		v := settings.VoxelDepth(utils.StrToInt(s) - 1)

		if v < settings.VXD_MAX && v >= 0 {

			RAM.IntSetVoxelDepth(index, v)

			e := backend.ProducerMain.GetInterpreter(index)
			apple2helpers.OSDShow(e, "Voxel Depth: "+v.String())

			// force recalibration
			settings.FrameSkip = settings.DefaultFrameSkip
		}

	case vduconst.SHIFT_CTRL_W:
		s := string(collected[index][1:])
		i := utils.StrToInt(s) - 1
		ent := backend.ProducerMain.GetInterpreter(index)
		cpu := apple2helpers.GetCPU(ent)
		zcpu := apple2helpers.GetZ80CPU(ent)
		switch i {
		case 0:
			fmt.Printf("Slot #%d CPU at %.2d\n", index, 25)
			cpu.SetWarpUser(0.25)
			zcpu.SetWarpUser(0.25)
			apple2helpers.OSDShow(ent, "25% Speed")
			settings.UserWarpOverride[SelectedIndex] = true
		case 1:
			fmt.Printf("Slot #%d CPU at %.2d\n", index, 50)
			cpu.SetWarpUser(0.50)
			zcpu.SetWarpUser(0.5)
			apple2helpers.OSDShow(ent, "50% Speed")
			settings.UserWarpOverride[SelectedIndex] = true
		case 2:
			fmt.Printf("Slot #%d CPU at %.2d\n", index, 100)
			cpu.SetWarpUser(1.0)
			zcpu.SetWarpUser(1.0)
			apple2helpers.OSDShow(ent, "100% Speed")
			settings.UserWarpOverride[SelectedIndex] = false
		case 3:
			fmt.Printf("Slot #%d CPU at %.2d\n", index, 200)
			cpu.SetWarpUser(2.0)
			zcpu.SetWarpUser(2.0)
			apple2helpers.OSDShow(ent, "200% Speed")
			settings.UserWarpOverride[SelectedIndex] = true
		case 4:
			fmt.Printf("Slot #%d CPU at %.2d\n", index, 400)
			cpu.SetWarpUser(4.0)
			zcpu.SetWarpUser(4.0)
			apple2helpers.OSDShow(ent, "400% Speed")
			settings.UserWarpOverride[SelectedIndex] = true
		case 5:
			fmt.Printf("Slot #%d CPU at %.2d\n", index, 800)
			cpu.SetWarpUser(8.0)
			zcpu.SetWarpUser(8.0)
			apple2helpers.OSDShow(ent, "800% Speed")
			settings.UserWarpOverride[SelectedIndex] = true
		}

	case vduconst.SHIFT_CTRL_N:
		s := string(collected[index][1:])
		if s[0] == 27 {
			backend.ProducerMain.EndRemotesNeeded = true
			for i := 1; i < memory.OCTALYZER_NUM_INTERPRETERS; i++ {
				GraphicsLayers[i] = make([]*video.GraphicsLayer, 0)
				HUDLayers[i] = make([]*video.TextLayer, 0)
				GFXSpecs[i] = make([]*types.LayerSpecMapped, 0)
				GFXSpecs[i] = make([]*types.LayerSpecMapped, 0)
			}
		}

	case vduconst.SHIFT_CTRL_R:
		e := backend.ProducerMain.GetInterpreter(index)
		s := strings.ToUpper(string(collected[index][1:]))
		switch s {
		case "W":
			apple2helpers.OSDShow(e, "Live Rewind Enabled")
			e.StopRecording()
			e.StartRecording("", false)
			fmt.Println("Live Record enabled")
		case "F":
			apple2helpers.OSDShow(e, "Recording started (Full CPU State)")
			e.StopRecording()
			e.RecordToggle(true)
		case "R":
			apple2helpers.OSDShow(e, "Recording started")
			e.StopRecording()
			e.RecordToggle(false)
		case "B":
			apple2helpers.OSDShow(e, "Recording started")
			e.StopRecording()
			blocks := e.GetLiveBlocks()
			e.StartRecordingWithBlocks(blocks, false)
		case "S":
			apple2helpers.OSDShow(e, "Saved Live recording")
			blocks := e.GetLiveBlocks()
			e.WriteBlocks(blocks)
		case "P":
			apple2helpers.OSDShow(e, "Raw Float PCM recording started")
			path := files.GetUserDirectory(files.BASEDIR + "/MyAudio")
			os.MkdirAll(path, 0755)
			filename := path + "/" + fmt.Sprintf("audio-%d.raw", time.Now().Unix())
			clientperipherals.SPEAKER.Mixer.StartRecording(filename)
		case "A":
			apple2helpers.OSDShow(e, "Restalgia recording started")
			rm := clientperipherals.SPEAKER.Mixer.Slots[clientperipherals.SPEAKER.Mixer.SlotSelect]
			path := files.GetUserDirectory(files.BASEDIR + "/MyAudio")
			os.MkdirAll(path, 0755)
			filename := path + "/" + fmt.Sprintf("restalgia-%d.rst", time.Now().Unix())
			rm.StartRecording(filename)
		case " ":
			apple2helpers.OSDShow(e, "Recording stopped")
			rm := clientperipherals.SPEAKER.Mixer.Slots[clientperipherals.SPEAKER.Mixer.SlotSelect]
			rm.StopRecording()
			e.StopRecording()
			clientperipherals.SPEAKER.Mixer.StopRecording()
		}
	}

	collected[index] = make([]rune, 0)
}
