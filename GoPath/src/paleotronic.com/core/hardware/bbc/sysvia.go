package bbc

import (
	"paleotronic.com/core/hardware/common"
	"paleotronic.com/log"
)

type SystemVIA struct {
	*common.MOS6522
	IC32State             int
	KbdState              [10][8]bool
	KbdCol, KbdRow        int
	KeysDown              int
	JoystickButton        bool
	SlowDataBusWriteValue int
	cpu                   Interruptable
}

func NewSystemVIA(cpu Interruptable) *SystemVIA {
	v := &SystemVIA{}
	v.MOS6522 = common.NewMOS6522(
		cpu,
		v.Write,
		v.Read,
	)
	v.Reset()
	return v
}

func (v *SystemVIA) Reset() {
	var row, col int
	v.MOS6522.Reset()
	//vialog=fopen("/via.log","wt");

	/* Make it no keys down and no dip switches set */
	for row = 0; row < 8; row++ {
		for col = 0; col < 10; col++ {
			v.KbdState[col][row] = false
		}
	}

	v.KbdState[6][5] = true
}

// CPU cycle handler
func (v *SystemVIA) DoCycles(ncycles int) {
	v.Increment(ncycles)
	v.CheckKeyboardInterrupt()
}

func (v *SystemVIA) IC32Write(Value int) {
	var bit uint

	bit = uint(Value) & 7

	if Value&8 != 0 {
		v.IC32State |= (1 << bit)
	} else {
		v.IC32State &= 0xff - (1 << bit)
	}

	v.CheckKeyboardInterrupt() /* Should really only if write enable on KBD changes */
}

func (v *SystemVIA) KbdOP() bool {
	if (v.KbdCol > 9) || (v.KbdRow > 7) {
		return false /* Key not down if overrange - perhaps we should do something more? */
	}
	return (v.KbdState[v.KbdCol][v.KbdRow])
}

func (v *SystemVIA) SlowDBWrite(Value int) {
	v.SlowDataBusWriteValue = Value

	if (v.IC32State & 8) == 0 {
		v.KbdRow = (Value >> 4) & 7
		v.KbdCol = (Value & 0xf)
		v.CheckKeyboardInterrupt() /* Should really only if write enable on KBD changes */
	}

}

func (v *SystemVIA) SlowDBRead() int {
	var result int

	result = int(v.ORA & v.DDRA)

	if (v.IC32State & 8) == 0 {
		if v.KbdOP() {
			result |= 128
		}
	}

	if (v.IC32State & 4) == 0 {
		result = 0xff
	}

	return (result)
}

func (v *SystemVIA) Write(Address int, Value byte) {

	switch Address {
	case 0:
		v.IC32Write(int(Value))
	case 1:
		v.SlowDBWrite(int(Value) & 0xff)
	case 2:
	case 3:
	case 4:
	case 6:
	case 5:
	case 7:
	case 8:
	case 9:
	case 10:
	case 11:
	case 12:
	case 13:
	case 14:
	case 15:
		v.SlowDBWrite(int(Value) & 0xff)
	} /* Address switch */
}

func (v *SystemVIA) Read(Address int) byte {

	var tmp byte = 0xff

	switch Address {
	case 0: /* IRB read */
		// Clear bit 4 of IFR from ATOD Conversion
		v.IRB = v.ORB & v.DDRB
		v.IRB |= 32 /* Fire button 2 released */
		if !v.JoystickButton {
			v.IRB |= 16
		}

		v.IRB |= 192 /* Speech system non existant */
	case 2:
	case 3:
	case 4: /* Timer 1 lo counter */
	case 5: /* Timer 1 hi counter */
	case 6: /* Timer 1 lo latch */
	case 7: /* Timer 1 hi latch */
	case 8: /* Timer 2 lo counter */
	case 9: /* Timer 2 hi counter */
	case 10:
	case 11:
	case 12:
	case 13:
	case 14:
	case 1:
	case 15:
		/* slow data bus read */
		v.IRA = byte(v.SlowDBRead())
	} /* Address switch */

	return tmp

}

func (v *SystemVIA) CheckKeyboardInterrupt() {
	if (v.KeysDown > 0) && ((v.PCR & 0xc) == 4) {
		//log.Printf("%d keys are down...", v.KeysDown)
		if (v.IC32State & 8) == 8 {
			//log.Printf("bit3 IC32 is set")
			v.IFR |= 1 /* CA2 */
			v.UpdateInterrupt()
		} else {
			//log.Printf("bit3 IC32 is not set")
			if v.KbdCol < 10 {
				log.Printf("KbdCol is within range")
				var presrow int
				for presrow = 1; presrow < 8; presrow++ {
					if v.KbdState[v.KbdCol][presrow] {
						log.Printf("Key %d, %d is down", presrow, v.KbdCol)
						v.IFR |= 1
						v.UpdateInterrupt()
					}
				} /* presrow */
			} /* KBDCol range */
		} /* WriteEnable on */
	} /* Keys down and CA2 input enabled */
}

func (mr *SystemVIA) KeyRelease(col, row int) {
	if row < 0 || col < 0 || row >= 8 || col >= 10 {
		return
	}

	if (mr.KbdState[col][row]) && (row != 0) {
		mr.KeysDown--
	}

	mr.KbdState[col][row] = false
}

func (mr *SystemVIA) KeyPress(col, row int) {
	if row < 0 || col < 0 || row >= 8 || col >= 10 {
		return
	}

	log.Printf("Keymatrix row=%d, col=%d", row, col)

	if (!mr.KbdState[col][row]) && (row != 0) {
		mr.KeysDown++
	}

	mr.KbdState[col][row] = true

	mr.CheckKeyboardInterrupt()
}

func (v *SystemVIA) KeyPressRune(ch rune) {
	mapping, ok := KeyMap[string(ch)]
	if ok {
		log.Printf("Mapping for %s", string(ch), mapping[1], mapping[0])
		v.KeyPress(mapping[0], mapping[1])
	} else {
		log.Printf("No Mapping for %s", string(ch))
	}

}
