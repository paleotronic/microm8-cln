package clientperipherals

import (
	"paleotronic.com/glumby"
	//"paleotronic.com/fmt"
)

type ControlAction int

const (
	CaNone                 ControlAction = 1 + iota
	CaPaddleButtonPress0   ControlAction = 1 + iota
	CaPaddleButtonPress1   ControlAction = 1 + iota
	CaPaddleButtonPress2   ControlAction = 1 + iota
	CaPaddleButtonPress3   ControlAction = 1 + iota
	CaPaddleButtonRelease0 ControlAction = 1 + iota
	CaPaddleButtonRelease1 ControlAction = 1 + iota
	CaPaddleButtonRelease2 ControlAction = 1 + iota
	CaPaddleButtonRelease3 ControlAction = 1 + iota
	CaPaddleDecrease0      ControlAction = 1 + iota
	CaPaddleDecrease1      ControlAction = 1 + iota
	CaPaddleDecrease2      ControlAction = 1 + iota
	CaPaddleDecrease3      ControlAction = 1 + iota
	CaPaddleIncrease0      ControlAction = 1 + iota
	CaPaddleIncrease1      ControlAction = 1 + iota
	CaPaddleIncrease2      ControlAction = 1 + iota
	CaPaddleIncrease3      ControlAction = 1 + iota
	CaPaddleValue0         ControlAction = 1 + iota
	CaPaddleValue1         ControlAction = 1 + iota
	CaPaddleValue2         ControlAction = 1 + iota
	CaPaddleValue3         ControlAction = 1 + iota
	CaPaddleModValue0      ControlAction = 1 + iota
	CaPaddleModValue1      ControlAction = 1 + iota
	CaPaddleModValue2      ControlAction = 1 + iota
	CaPaddleModValue3      ControlAction = 1 + iota
	CaGameStart            ControlAction = 1 + iota
	CaGameSelect           ControlAction = 1 + iota
	CaGameBack             ControlAction = 1 + iota
	CaGameAccept           ControlAction = 1 + iota
)

func (c ControlAction) String() string {
	switch c {
	case CaNone:
		return "None"
	case CaPaddleButtonPress0:
		return "PressPaddleButton0"
	case CaPaddleButtonPress1:
		return "PressPaddleButton1"
	case CaPaddleButtonPress2:
		return "PressPaddleButton2"
	case CaPaddleButtonPress3:
		return "PressPaddleButton3"
	case CaPaddleButtonRelease0:
		return "ReleasePaddleButton0"
	case CaPaddleButtonRelease1:
		return "ReleasePaddleButton1"
	case CaPaddleButtonRelease2:
		return "ReleasePaddleButton2"
	case CaPaddleButtonRelease3:
		return "ReleasePaddleButton3"
	case CaPaddleDecrease0:
		return "DecreasePaddle0"
	case CaPaddleDecrease1:
		return "DecreasePaddle1"
	case CaPaddleDecrease2:
		return "DecreasePaddle2"
	case CaPaddleDecrease3:
		return "DecreasePaddle3"
	case CaPaddleIncrease0:
		return "IncreasePaddle0"
	case CaPaddleIncrease1:
		return "IncreasePaddle1"
	case CaPaddleIncrease2:
		return "IncreasePaddle2"
	case CaPaddleIncrease3:
		return "IncreasePaddle3"
	case CaPaddleValue0:
		return "ValuePaddle0"
	case CaPaddleValue1:
		return "ValuePaddle1"
	case CaPaddleValue2:
		return "ValuePaddle2"
	case CaPaddleValue3:
		return "ValuePaddle3"
	case CaPaddleModValue0:
		return "ModValuePaddle0"
	case CaPaddleModValue1:
		return "ModValuePaddle1"
	case CaPaddleModValue2:
		return "ModValuePaddle2"
	case CaPaddleModValue3:
		return "ModValuePaddle3"
	case CaGameStart:
		return "GameStart"
	case CaGameSelect:
		return "GameSelect"
	case CaGameBack:
		return "GameBack"
	case CaGameAccept:
		return "GameAccept"
	}

	return "Unknown"
}

type ControlMapping map[*glumby.ControllerEvent]ControlAction

var ControlMap ControlMapping
var UseMapA bool

func MapA() {
	ControlMap = make(ControlMapping)
	ControlMap[&glumby.ControllerEvent{Type: glumby.ControllerButtonEvent, ID: 0, Index: 0, Pressed: true}] = CaPaddleButtonPress0
	ControlMap[&glumby.ControllerEvent{Type: glumby.ControllerButtonEvent, ID: 0, Index: 0, Pressed: false}] = CaPaddleButtonRelease0
	ControlMap[&glumby.ControllerEvent{Type: glumby.ControllerButtonEvent, ID: 0, Index: 1, Pressed: true}] = CaPaddleButtonPress1
	ControlMap[&glumby.ControllerEvent{Type: glumby.ControllerButtonEvent, ID: 0, Index: 1, Pressed: false}] = CaPaddleButtonRelease1

	ControlMap[&glumby.ControllerEvent{Type: glumby.ControllerButtonEvent, ID: 0, Index: 11, Pressed: false}] = CaPaddleIncrease0
	ControlMap[&glumby.ControllerEvent{Type: glumby.ControllerButtonEvent, ID: 0, Index: 10, Pressed: false}] = CaPaddleDecrease1
	ControlMap[&glumby.ControllerEvent{Type: glumby.ControllerButtonEvent, ID: 0, Index: 13, Pressed: false}] = CaPaddleDecrease0
	ControlMap[&glumby.ControllerEvent{Type: glumby.ControllerButtonEvent, ID: 0, Index: 12, Pressed: false}] = CaPaddleIncrease1

	ControlMap[&glumby.ControllerEvent{Type: glumby.ControllerAxisEvent, ID: 0, Index: 0}] = CaPaddleModValue0
	ControlMap[&glumby.ControllerEvent{Type: glumby.ControllerAxisEvent, ID: 0, Index: 1}] = CaPaddleModValue1

	ControlMap[&glumby.ControllerEvent{Type: glumby.ControllerAxisEvent, ID: 0, Index: 2}] = CaPaddleModValue2
	ControlMap[&glumby.ControllerEvent{Type: glumby.ControllerAxisEvent, ID: 0, Index: 3}] = CaPaddleModValue3

	ControlMap[&glumby.ControllerEvent{Type: glumby.ControllerButtonEvent, ID: 0, Index: 8, Pressed: true}] = CaGameSelect
	ControlMap[&glumby.ControllerEvent{Type: glumby.ControllerButtonEvent, ID: 0, Index: 9, Pressed: true}] = CaGameStart
	ControlMap[&glumby.ControllerEvent{Type: glumby.ControllerButtonEvent, ID: 0, Index: 8, Pressed: false}] = CaGameSelect
	ControlMap[&glumby.ControllerEvent{Type: glumby.ControllerButtonEvent, ID: 0, Index: 9, Pressed: false}] = CaGameStart
}

func MapB() {
	ControlMap = make(ControlMapping)
	ControlMap[&glumby.ControllerEvent{Type: glumby.ControllerButtonEvent, ID: 0, Index: 0, Pressed: true}] = CaPaddleButtonPress0
	ControlMap[&glumby.ControllerEvent{Type: glumby.ControllerButtonEvent, ID: 0, Index: 0, Pressed: false}] = CaPaddleButtonRelease0
	ControlMap[&glumby.ControllerEvent{Type: glumby.ControllerButtonEvent, ID: 0, Index: 1, Pressed: true}] = CaPaddleButtonPress1
	ControlMap[&glumby.ControllerEvent{Type: glumby.ControllerButtonEvent, ID: 0, Index: 1, Pressed: false}] = CaPaddleButtonRelease1

	ControlMap[&glumby.ControllerEvent{Type: glumby.ControllerButtonEvent, ID: 0, Index: 11, Pressed: false}] = CaPaddleIncrease1
	ControlMap[&glumby.ControllerEvent{Type: glumby.ControllerButtonEvent, ID: 0, Index: 10, Pressed: false}] = CaPaddleDecrease0
	ControlMap[&glumby.ControllerEvent{Type: glumby.ControllerButtonEvent, ID: 0, Index: 13, Pressed: false}] = CaPaddleDecrease1
	ControlMap[&glumby.ControllerEvent{Type: glumby.ControllerButtonEvent, ID: 0, Index: 12, Pressed: false}] = CaPaddleIncrease0

	ControlMap[&glumby.ControllerEvent{Type: glumby.ControllerAxisEvent, ID: 0, Index: 0}] = CaPaddleModValue0
	ControlMap[&glumby.ControllerEvent{Type: glumby.ControllerAxisEvent, ID: 0, Index: 1}] = CaPaddleModValue1
	ControlMap[&glumby.ControllerEvent{Type: glumby.ControllerAxisEvent, ID: 0, Index: 2}] = CaPaddleModValue2
	ControlMap[&glumby.ControllerEvent{Type: glumby.ControllerAxisEvent, ID: 0, Index: 3}] = CaPaddleModValue3
	//ControlMap[ &glumby.ControllerEvent{ Type: glumby.ControllerAxisEvent, ID: 0, Index: 0 } ] = CaPaddleModValue1

	ControlMap[&glumby.ControllerEvent{Type: glumby.ControllerAxisEvent, ID: 0, Index: 7}] = CaPaddleModValue1

	ControlMap[&glumby.ControllerEvent{Type: glumby.ControllerButtonEvent, ID: 0, Index: 5, Pressed: true}] = CaGameAccept
	ControlMap[&glumby.ControllerEvent{Type: glumby.ControllerButtonEvent, ID: 0, Index: 6, Pressed: true}] = CaGameBack
	ControlMap[&glumby.ControllerEvent{Type: glumby.ControllerButtonEvent, ID: 0, Index: 7, Pressed: true}] = CaGameStart

	// ControlMap[&glumby.ControllerEvent{Type: glumby.ControllerAxisEvent, ID: 0, Index: 3}] = CaPaddleModValue1
	// ControlMap[&glumby.ControllerEvent{Type: glumby.ControllerAxisEvent, ID: 0, Index: 4}] = CaPaddleModValue0
	//ControlMap[ &glumby.ControllerEvent{ Type: glumby.ControllerAxisEvent, ID: 0, Index: 3 } ] = CaPaddleModValue0

	ControlMap[&glumby.ControllerEvent{Type: glumby.ControllerAxisEvent, ID: 0, Index: 6}] = CaPaddleModValue0

	ControlMap[&glumby.ControllerEvent{Type: glumby.ControllerButtonEvent, ID: 0, Index: 8, Pressed: true}] = CaGameSelect
	ControlMap[&glumby.ControllerEvent{Type: glumby.ControllerButtonEvent, ID: 0, Index: 9, Pressed: true}] = CaGameStart
	ControlMap[&glumby.ControllerEvent{Type: glumby.ControllerButtonEvent, ID: 0, Index: 8, Pressed: false}] = CaGameSelect
	ControlMap[&glumby.ControllerEvent{Type: glumby.ControllerButtonEvent, ID: 0, Index: 9, Pressed: false}] = CaGameStart
}

func ToggleMap() {
	UseMapA = !UseMapA
	if UseMapA {
		MapA()
	} else {
		MapB()
	}
}

func init() {
	MapB()
}

func GetActionForEvent(ev *glumby.ControllerEvent) (ControlAction, float32) {

	//fmt.Printf("Control event %v\n", ev)

	for cev, action := range ControlMap {

		if ev.Type == cev.Type {
			switch ev.Type {
			case glumby.ControllerAxisEvent:
				{
					if ev.Index == cev.Index {
						return action, ev.Value
					}
				}
			case glumby.ControllerButtonEvent:
				{
					if ev.Index == cev.Index && ev.Pressed == cev.Pressed {
						return action, 0
					}
				}
			}
		}

	}

	return CaNone, 0
}
