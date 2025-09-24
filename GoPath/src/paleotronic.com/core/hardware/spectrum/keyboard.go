/*

Copyright (c) 2010 Andrea Fazzi

Permission is hereby granted, free of charge, to any person obtaining
a copy of this software and associated documentation files (the
"Software"), to deal in the Software without restriction, including
without limitation the rights to use, copy, modify, merge, publish,
distribute, sublicense, and/or sell copies of the Software, and to
permit persons to whom the Software is furnished to do so, subject to
the following conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

*/

package spectrum

import (
	"sync"
	"time"

	"paleotronic.com/core/hardware/servicebus"
	"paleotronic.com/core/interfaces"
)

type rowState struct {
	row, state byte
}

type Cmd_KeyPress struct {
	logicalKeyCode KeyMapping
	done           chan bool
}

type Cmd_SendLoad struct {
	romType int
}

type Keyboard struct {
	e         interfaces.Interpretable
	keyStates [8]byte
	mutex     sync.RWMutex

	CommandChannel chan interface{}
	WaitingForRead bool

	spectrum *ZXSpectrum
}

func NewKeyboard(e interfaces.Interpretable, spec *ZXSpectrum) *Keyboard {
	keyboard := &Keyboard{
		e:        e,
		spectrum: spec,
	}
	keyboard.reset()
	keyboard.init()

	keyboard.CommandChannel = make(chan interface{})

	return keyboard
}

func (keyboard *Keyboard) init( /*speccy *Spectrum48k*/ ) {
	//keyboard.speccy = speccy
	go keyboard.commandLoop()
	go keyboard.keyFeeder()
}

func (keyboard *Keyboard) waitFrames(n uint64) {
	// until := n + keyboard.spectrum.framecounter
	// for keyboard.spectrum.framecounter < until {
	// 	time.Sleep(time.Millisecond)
	// }
	// for i := 0; i < int(n); i++ {
	// 	n := keyboard.spectrum.framecounter
	// 	for n == keyboard.spectrum.framecounter {
	// 		time.Sleep(time.Millisecond)
	// 	}
	// }
	time.Sleep(time.Duration(n) * 20 * time.Millisecond)
}

func (keyboard *Keyboard) delayAfterKeyDown() {
	//keyboard.waitFrames(2)
	keyboard.WaitingForRead = true
	for keyboard.WaitingForRead {
		time.Sleep(time.Millisecond)
	}
}

func (keyboard *Keyboard) delayAfterKeyUp() {
	//keyboard.waitFrames(10)
	keyboard.WaitingForRead = true
	for keyboard.WaitingForRead {
		time.Sleep(time.Millisecond)
	}
	time.Sleep(20 * time.Millisecond)
}

func (keyboard *Keyboard) ProcessKeyEvent(ev *servicebus.KeyEventData) {

	if ev.Key >= 'A' && ev.Key <= 'Z' {
		ev.Key += 32
	}

	lookup := KeyEvent{
		ev.Key,
		ev.Modifier,
	}

	if logicalKeyCodes, ok := GlumbyKeymap[lookup]; ok {
		switch ev.Action {
		case servicebus.ActionPress, servicebus.ActionRepeat:
			if logicalKeyCodes.Caps {
				keyboard.KeyDown(KEY_CapsShift)
			}
			if logicalKeyCodes.Sym {
				keyboard.KeyDown(KEY_SymbolShift)
			}
			if logicalKeyCodes.Code != KEY_None {
				keyboard.KeyDown(logicalKeyCodes.Code)
			}
		case servicebus.ActionRelease:
			if logicalKeyCodes.Code != KEY_None {
				keyboard.KeyUp(logicalKeyCodes.Code)
			}
			if logicalKeyCodes.Sym {
				keyboard.KeyUp(KEY_SymbolShift)
			}
			if logicalKeyCodes.Caps {
				keyboard.KeyUp(KEY_CapsShift)
			}
		}
	}

}

func (keyboard *Keyboard) keyFeeder() {

	return

	// vm := keyboard.e.VM()
	// for !vm.IsDying() {
	// 	if ch := keyboard.e.GetMemoryMap().KeyBufferGetLatest(keyboard.e.GetMemIndex()); ch != 0 {
	// 		log.Printf("Got keycode %d", ch)
	// 		if logicalKeyCodes, ok := GlumbyKeymap[rune(ch)]; ok {
	// 			log.Printf("Sending keystroke for glumby code %d -> %+v", ch, logicalKeyCodes)
	// 			<-keyboard.KeyPressSequence(logicalKeyCodes)
	// 		}
	// 	} else {
	// 		time.Sleep(time.Millisecond)
	// 	}
	// }
}

func (keyboard *Keyboard) commandLoop() {

	// evtLoop := keyboard.speccy.app.NewEventLoop()
	vm := keyboard.e.VM()

	for !vm.IsDying() {
		select {
		case untyped_cmd := <-keyboard.CommandChannel:
			switch cmd := untyped_cmd.(type) {
			case Cmd_KeyPress:

				if cmd.logicalKeyCode.Sym {
					keyboard.KeyDown(KEY_SymbolShift)
					keyboard.delayAfterKeyDown()
				}
				if cmd.logicalKeyCode.Caps {
					keyboard.KeyDown(KEY_CapsShift)
					keyboard.delayAfterKeyDown()
				}

				if cmd.logicalKeyCode.Code != KEY_None {
					keyboard.KeyDown(cmd.logicalKeyCode.Code)
					keyboard.delayAfterKeyDown()
					keyboard.KeyUp(cmd.logicalKeyCode.Code)
					keyboard.delayAfterKeyUp()
				}

				if cmd.logicalKeyCode.Caps {
					keyboard.KeyUp(KEY_CapsShift)
					keyboard.delayAfterKeyUp()
				}
				if cmd.logicalKeyCode.Sym {
					keyboard.KeyUp(KEY_SymbolShift)
					keyboard.delayAfterKeyUp()
				}

				time.Sleep(2 * time.Millisecond)

				cmd.done <- true

			case Cmd_SendLoad:

				// LOAD
				keyboard.KeyDown(KEY_J)
				keyboard.delayAfterKeyDown()
				keyboard.KeyUp(KEY_J)
				keyboard.delayAfterKeyUp()

				// " "
				keyboard.KeyDown(KEY_SymbolShift)
				{
					keyboard.KeyDown(KEY_P)
					keyboard.delayAfterKeyDown()
					keyboard.KeyUp(KEY_P)
					keyboard.delayAfterKeyUp()

					keyboard.KeyDown(KEY_P)
					keyboard.delayAfterKeyDown()
					keyboard.KeyUp(KEY_P)
					keyboard.delayAfterKeyUp()
				}
				keyboard.KeyUp(KEY_SymbolShift)

				keyboard.KeyDown(KEY_Enter)
				keyboard.delayAfterKeyDown()
				keyboard.KeyUp(KEY_Enter)

			}
		default:
			time.Sleep(time.Millisecond)
		}
	}

}

func (k *Keyboard) reset() {
	// Initialize 'k.keyStates'
	for row := uint(0); row < 8; row++ {
		k.SetKeyState(row, 0xff)
	}
}

func (keyboard *Keyboard) GetKeyState(row uint) byte {
	keyboard.mutex.RLock()
	keyState := keyboard.keyStates[row]
	keyboard.WaitingForRead = false
	keyboard.mutex.RUnlock()
	return keyState
}

func (keyboard *Keyboard) SetKeyState(row uint, state byte) {
	keyboard.mutex.Lock()
	keyboard.keyStates[row] = state
	keyboard.mutex.Unlock()
}

func (keyboard *Keyboard) KeyDown(logicalKeyCode uint) {
	keyCode, ok := keyCodes[logicalKeyCode]

	if ok {
		keyboard.mutex.Lock()
		keyboard.keyStates[keyCode.row] &= ^(keyCode.mask)
		keyboard.mutex.Unlock()
	}
}

func (keyboard *Keyboard) KeyUp(logicalKeyCode uint) {
	keyCode, ok := keyCodes[logicalKeyCode]

	if ok {
		keyboard.mutex.Lock()
		keyboard.keyStates[keyCode.row] |= (keyCode.mask)
		keyboard.mutex.Unlock()
	}
}

func (keyboard *Keyboard) KeyPress(logicalKeyCode KeyMapping) chan bool {
	done := make(chan bool)
	keyboard.CommandChannel <- Cmd_KeyPress{logicalKeyCode, done}
	return done
}

func (keyboard *Keyboard) KeyPressSequence(logicalKeyCodes KeyMapping) chan bool {
	done := make(chan bool, 1)
	keyboard.CommandChannel <- Cmd_KeyPress{logicalKeyCodes, done}
	return done
}
