package spectrum

import (
	"paleotronic.com/core/hardware/servicebus"
)

// Logical key codes
const (
	KEY_1 = iota
	KEY_2
	KEY_3
	KEY_4
	KEY_5
	KEY_6
	KEY_7
	KEY_8
	KEY_9
	KEY_0

	KEY_Q
	KEY_W
	KEY_E
	KEY_R
	KEY_T
	KEY_Y
	KEY_U
	KEY_I
	KEY_O
	KEY_P

	KEY_A
	KEY_S
	KEY_D
	KEY_F
	KEY_G
	KEY_H
	KEY_J
	KEY_K
	KEY_L
	KEY_Enter

	KEY_CapsShift
	KEY_Z
	KEY_X
	KEY_C
	KEY_V
	KEY_B
	KEY_N
	KEY_M
	KEY_SymbolShift
	KEY_Space
	KEY_None
)

type keyCell struct {
	row, mask byte
}

var keyCodes = map[uint]keyCell{
	KEY_1: keyCell{row: 3, mask: 0x01},
	KEY_2: keyCell{row: 3, mask: 0x02},
	KEY_3: keyCell{row: 3, mask: 0x04},
	KEY_4: keyCell{row: 3, mask: 0x08},
	KEY_5: keyCell{row: 3, mask: 0x10},
	KEY_6: keyCell{row: 4, mask: 0x10},
	KEY_7: keyCell{row: 4, mask: 0x08},
	KEY_8: keyCell{row: 4, mask: 0x04},
	KEY_9: keyCell{row: 4, mask: 0x02},
	KEY_0: keyCell{row: 4, mask: 0x01},

	KEY_Q: keyCell{row: 2, mask: 0x01},
	KEY_W: keyCell{row: 2, mask: 0x02},
	KEY_E: keyCell{row: 2, mask: 0x04},
	KEY_R: keyCell{row: 2, mask: 0x08},
	KEY_T: keyCell{row: 2, mask: 0x10},
	KEY_Y: keyCell{row: 5, mask: 0x10},
	KEY_U: keyCell{row: 5, mask: 0x08},
	KEY_I: keyCell{row: 5, mask: 0x04},
	KEY_O: keyCell{row: 5, mask: 0x02},
	KEY_P: keyCell{row: 5, mask: 0x01},

	KEY_A:     keyCell{row: 1, mask: 0x01},
	KEY_S:     keyCell{row: 1, mask: 0x02},
	KEY_D:     keyCell{row: 1, mask: 0x04},
	KEY_F:     keyCell{row: 1, mask: 0x08},
	KEY_G:     keyCell{row: 1, mask: 0x10},
	KEY_H:     keyCell{row: 6, mask: 0x10},
	KEY_J:     keyCell{row: 6, mask: 0x08},
	KEY_K:     keyCell{row: 6, mask: 0x04},
	KEY_L:     keyCell{row: 6, mask: 0x02},
	KEY_Enter: keyCell{row: 6, mask: 0x01},

	KEY_CapsShift:   keyCell{row: 0, mask: 0x01},
	KEY_Z:           keyCell{row: 0, mask: 0x02},
	KEY_X:           keyCell{row: 0, mask: 0x04},
	KEY_C:           keyCell{row: 0, mask: 0x08},
	KEY_V:           keyCell{row: 0, mask: 0x10},
	KEY_B:           keyCell{row: 7, mask: 0x10},
	KEY_N:           keyCell{row: 7, mask: 0x08},
	KEY_M:           keyCell{row: 7, mask: 0x04},
	KEY_SymbolShift: keyCell{row: 7, mask: 0x02},
	KEY_Space:       keyCell{row: 7, mask: 0x01},
}

type KeyMode int

const (
	kmDown KeyMode = 1 << iota
	kmUp
	kmBoth
)

type KeyMapping struct {
	Sym  bool
	Caps bool
	Code uint
}

func CapSym(c uint) KeyMapping {
	return KeyMapping{Caps: true, Sym: true, Code: c}
}

func Cap(c uint) KeyMapping {
	return KeyMapping{Caps: true, Sym: false, Code: c}
}

func Sym(c uint) KeyMapping {
	return KeyMapping{Caps: false, Sym: true, Code: c}
}

func Key(c uint) KeyMapping {
	return KeyMapping{Caps: false, Sym: false, Code: c}
}

type KeyEvent struct {
	Key      rune
	Modifier servicebus.KeyMod
}

var GlumbyKeymap = map[KeyEvent]KeyMapping{
	KeyEvent{'0', servicebus.ModNone}: Key(KEY_0),
	KeyEvent{'1', servicebus.ModNone}: Key(KEY_1),
	KeyEvent{'2', servicebus.ModNone}: Key(KEY_2),
	KeyEvent{'3', servicebus.ModNone}: Key(KEY_3),
	KeyEvent{'4', servicebus.ModNone}: Key(KEY_4),
	KeyEvent{'5', servicebus.ModNone}: Key(KEY_5),
	KeyEvent{'6', servicebus.ModNone}: Key(KEY_6),
	KeyEvent{'7', servicebus.ModNone}: Key(KEY_7),
	KeyEvent{'8', servicebus.ModNone}: Key(KEY_8),
	KeyEvent{'9', servicebus.ModNone}: Key(KEY_9),

	KeyEvent{'a', servicebus.ModNone}: Key(KEY_A),
	KeyEvent{'b', servicebus.ModNone}: Key(KEY_B),
	KeyEvent{'c', servicebus.ModNone}: Key(KEY_C),
	KeyEvent{'d', servicebus.ModNone}: Key(KEY_D),
	KeyEvent{'e', servicebus.ModNone}: Key(KEY_E),
	KeyEvent{'f', servicebus.ModNone}: Key(KEY_F),
	KeyEvent{'g', servicebus.ModNone}: Key(KEY_G),
	KeyEvent{'h', servicebus.ModNone}: Key(KEY_H),
	KeyEvent{'i', servicebus.ModNone}: Key(KEY_I),
	KeyEvent{'j', servicebus.ModNone}: Key(KEY_J),
	KeyEvent{'k', servicebus.ModNone}: Key(KEY_K),
	KeyEvent{'l', servicebus.ModNone}: Key(KEY_L),
	KeyEvent{'m', servicebus.ModNone}: Key(KEY_M),
	KeyEvent{'n', servicebus.ModNone}: Key(KEY_N),
	KeyEvent{'o', servicebus.ModNone}: Key(KEY_O),
	KeyEvent{'p', servicebus.ModNone}: Key(KEY_P),
	KeyEvent{'q', servicebus.ModNone}: Key(KEY_Q),
	KeyEvent{'r', servicebus.ModNone}: Key(KEY_R),
	KeyEvent{'s', servicebus.ModNone}: Key(KEY_S),
	KeyEvent{'t', servicebus.ModNone}: Key(KEY_T),
	KeyEvent{'u', servicebus.ModNone}: Key(KEY_U),
	KeyEvent{'v', servicebus.ModNone}: Key(KEY_V),
	KeyEvent{'w', servicebus.ModNone}: Key(KEY_W),
	KeyEvent{'x', servicebus.ModNone}: Key(KEY_X),
	KeyEvent{'y', servicebus.ModNone}: Key(KEY_Y),
	KeyEvent{'z', servicebus.ModNone}: Key(KEY_Z),

	KeyEvent{'a', servicebus.ModShift}: Cap(KEY_A),
	KeyEvent{'b', servicebus.ModShift}: Cap(KEY_B),
	KeyEvent{'c', servicebus.ModShift}: Cap(KEY_C),
	KeyEvent{'d', servicebus.ModShift}: Cap(KEY_D),
	KeyEvent{'e', servicebus.ModShift}: Cap(KEY_E),
	KeyEvent{'f', servicebus.ModShift}: Cap(KEY_F),
	KeyEvent{'g', servicebus.ModShift}: Cap(KEY_G),
	KeyEvent{'h', servicebus.ModShift}: Cap(KEY_H),
	KeyEvent{'i', servicebus.ModShift}: Cap(KEY_I),
	KeyEvent{'j', servicebus.ModShift}: Cap(KEY_J),
	KeyEvent{'k', servicebus.ModShift}: Cap(KEY_K),
	KeyEvent{'l', servicebus.ModShift}: Cap(KEY_L),
	KeyEvent{'m', servicebus.ModShift}: Cap(KEY_M),
	KeyEvent{'n', servicebus.ModShift}: Cap(KEY_N),
	KeyEvent{'o', servicebus.ModShift}: Cap(KEY_O),
	KeyEvent{'p', servicebus.ModShift}: Cap(KEY_P),
	KeyEvent{'q', servicebus.ModShift}: Cap(KEY_Q),
	KeyEvent{'r', servicebus.ModShift}: Cap(KEY_R),
	KeyEvent{'s', servicebus.ModShift}: Cap(KEY_S),
	KeyEvent{'t', servicebus.ModShift}: Cap(KEY_T),
	KeyEvent{'u', servicebus.ModShift}: Cap(KEY_U),
	KeyEvent{'v', servicebus.ModShift}: Cap(KEY_V),
	KeyEvent{'w', servicebus.ModShift}: Cap(KEY_W),
	KeyEvent{'x', servicebus.ModShift}: Cap(KEY_X),
	KeyEvent{'y', servicebus.ModShift}: Cap(KEY_Y),
	KeyEvent{'z', servicebus.ModShift}: Cap(KEY_Z),

	KeyEvent{'a', servicebus.ModCtrl}: Sym(KEY_A),
	KeyEvent{'b', servicebus.ModCtrl}: Sym(KEY_B),
	KeyEvent{'c', servicebus.ModCtrl}: Sym(KEY_C),
	KeyEvent{'d', servicebus.ModCtrl}: Sym(KEY_D),
	KeyEvent{'e', servicebus.ModCtrl}: Sym(KEY_E),
	KeyEvent{'f', servicebus.ModCtrl}: Sym(KEY_F),
	KeyEvent{'g', servicebus.ModCtrl}: Sym(KEY_G),
	KeyEvent{'h', servicebus.ModCtrl}: Sym(KEY_H),
	KeyEvent{'i', servicebus.ModCtrl}: Sym(KEY_I),
	KeyEvent{'j', servicebus.ModCtrl}: Sym(KEY_J),
	KeyEvent{'k', servicebus.ModCtrl}: Sym(KEY_K),
	KeyEvent{'l', servicebus.ModCtrl}: Sym(KEY_L),
	KeyEvent{'m', servicebus.ModCtrl}: Sym(KEY_M),
	KeyEvent{'n', servicebus.ModCtrl}: Sym(KEY_N),
	KeyEvent{'o', servicebus.ModCtrl}: Sym(KEY_O),
	KeyEvent{'p', servicebus.ModCtrl}: Sym(KEY_P),
	KeyEvent{'q', servicebus.ModCtrl}: Sym(KEY_Q),
	KeyEvent{'r', servicebus.ModCtrl}: Sym(KEY_R),
	KeyEvent{'s', servicebus.ModCtrl}: Sym(KEY_S),
	KeyEvent{'t', servicebus.ModCtrl}: Sym(KEY_T),
	KeyEvent{'u', servicebus.ModCtrl}: Sym(KEY_U),
	KeyEvent{'v', servicebus.ModCtrl}: Sym(KEY_V),
	KeyEvent{'w', servicebus.ModCtrl}: Sym(KEY_W),
	KeyEvent{'x', servicebus.ModCtrl}: Sym(KEY_X),
	KeyEvent{'y', servicebus.ModCtrl}: Sym(KEY_Y),
	KeyEvent{'z', servicebus.ModCtrl}: Sym(KEY_Z),

	KeyEvent{'1', servicebus.ModShift}: Cap(KEY_1),
	KeyEvent{'2', servicebus.ModShift}: Cap(KEY_2),
	KeyEvent{'3', servicebus.ModShift}: Cap(KEY_3),
	KeyEvent{'4', servicebus.ModShift}: Cap(KEY_4),
	KeyEvent{'5', servicebus.ModShift}: Cap(KEY_5),
	KeyEvent{'6', servicebus.ModShift}: Cap(KEY_6),
	KeyEvent{'7', servicebus.ModShift}: Cap(KEY_7),
	KeyEvent{'8', servicebus.ModShift}: Cap(KEY_8),
	KeyEvent{'9', servicebus.ModShift}: Cap(KEY_9),
	KeyEvent{'0', servicebus.ModShift}: Cap(KEY_0),

	KeyEvent{257, servicebus.ModNone}:  Key(KEY_Enter),
	KeyEvent{32, servicebus.ModNone}:   Key(KEY_Space),
	KeyEvent{259, servicebus.ModNone}:  Cap(KEY_0),
	KeyEvent{341, servicebus.ModNone}:  Sym(KEY_None),
	KeyEvent{341, servicebus.ModCtrl}:  Sym(KEY_None),
	KeyEvent{340, servicebus.ModNone}:  Cap(KEY_None),
	KeyEvent{340, servicebus.ModShift}: Cap(KEY_None),
	KeyEvent{342, servicebus.ModNone}:  CapSym(KEY_None),
	KeyEvent{342, servicebus.ModAlt}:   CapSym(KEY_None),

	KeyEvent{265, servicebus.ModNone}: Cap(KEY_7),
	KeyEvent{264, servicebus.ModNone}: Cap(KEY_6),
	KeyEvent{263, servicebus.ModNone}: Cap(KEY_5),
	KeyEvent{262, servicebus.ModNone}: Cap(KEY_8),

	KeyEvent{280, servicebus.ModNone}: Cap(KEY_2),

	KeyEvent{32, servicebus.ModShift}: Cap(KEY_Space),

	//"escape":    Key(KEY_CapsShift, KEY_1),
	//"caps lock": Key(KEY_CapsShift, KEY_2), // FIXME: SDL never sends the sdl.KEYUP event

	KeyEvent{'-', servicebus.ModNone}: Sym(KEY_J),
	//"_":  Sym(KEY_0),
	KeyEvent{'=', servicebus.ModNone}: Sym(KEY_L),
	//"+":  Sym(KEY_K),
	KeyEvent{'[', servicebus.ModNone}: Sym(KEY_8), // Maps to "("
	KeyEvent{']', servicebus.ModNone}: Sym(KEY_9), // Maps to ")"
	KeyEvent{';', servicebus.ModNone}: Sym(KEY_O),
	//":":  Sym(KEY_Z),
	KeyEvent{'\'', servicebus.ModNone}: Sym(KEY_7),
	//"\"":  Sym(KEY_P),
	KeyEvent{',', servicebus.ModNone}: Sym(KEY_N),
	KeyEvent{'.', servicebus.ModNone}: Sym(KEY_M),
	KeyEvent{'/', servicebus.ModNone}: Sym(KEY_V),
	KeyEvent{39, servicebus.ModShift}: Sym(KEY_P),
	KeyEvent{39, servicebus.ModNone}:  Sym(KEY_P),
	//"<":  Sym(KEY_R),
	//">":  Sym(KEY_T),
	//"?":  Sym(KEY_C),

	KeyEvent{'1', servicebus.ModCtrl}: Sym(KEY_1),
	KeyEvent{'2', servicebus.ModCtrl}: Sym(KEY_2),
	KeyEvent{'3', servicebus.ModCtrl}: Sym(KEY_3),
	KeyEvent{'4', servicebus.ModCtrl}: Sym(KEY_4),
	KeyEvent{'5', servicebus.ModCtrl}: Sym(KEY_5),
	KeyEvent{'6', servicebus.ModCtrl}: Sym(KEY_6),
	KeyEvent{'7', servicebus.ModCtrl}: Sym(KEY_7),
	KeyEvent{'8', servicebus.ModCtrl}: Sym(KEY_8),
	KeyEvent{'9', servicebus.ModCtrl}: Sym(KEY_9),
	KeyEvent{'0', servicebus.ModCtrl}: Sym(KEY_0),
}

func init() {
	// if len(keyCodes) != 40 {
	// 	panic("invalid keyboard specification")
	// }

	// // Make sure we are able to press every button on the Spectrum keyboard
	// used := make(map[uint]bool)
	// for logicalKeyCode := range keyCodes {
	// 	used[logicalKeyCode] = false
	// }
	// for _, seq := range GlumbyKeymap {
	// 	if len(seq) == 1 {
	// 		used[seq[0]] = true
	// 	}
	// }
	// for _, isUsed := range used {
	// 	if !isUsed {
	// 		panic("some key is missing in the SDL keymap")
	// 	}
	// }
}
