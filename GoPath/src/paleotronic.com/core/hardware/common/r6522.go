package common

import (
	"fmt"

	"paleotronic.com/core/settings"
)

type R6522Register int

const (
	ORB  R6522Register = iota // Output Register B
	ORA                       // Output Register A
	DDRB                      // Data direction reg B
	DDRA                      // Data direction reg A
	T1CL                      // T1 low-order latches (low-order counter for read operations)
	T1CH                      // T1 high-order counter
	T1LL                      // T1 low-order latches
	T1LH                      // T1 high-order latches
	T2CL                      // T2 low-order latches (low-order counter for read operations)
	T2CH                      // T2 high-order counter
	SR                        // Shift register
	ACR                       // Aux control register
	PCR                       // Perripheral control register
	IFR                       // Interrupt flag register
	IER                       // Interrupt enable register
	ORAH                      // Output Register A (no handshake)
)

func (r R6522Register) String() string {
	switch r {
	case ORB:
		return "ORB"
	case ORA:
		return "ORA"
	case DDRB:
		return "DDRB"
	case DDRA:
		return "DDRA"
	case T1CL:
		return "T1CL"
	case T1CH:
		return "T1CH"
	case T1LL:
		return "T1LL"
	case T1LH:
		return "T1LH"
	case T2CL:
		return "T2CL"
	case T2CH:
		return "T2CH"
	case SR:
		return "SR"
	case ACR:
		return "ACR"
	case PCR:
		return "PCR"
	case IFR:
		return "IFR"
	case IER:
		return "IER"
	case ORAH:
		return "ORAH"
	}
	return "INVALID"
}

type Interruptable interface {
	PullIRQLine()
}

type R6522 struct {
	oraReg int
	iraReg int
	orbReg int
	irbReg int

	dataDirectionA int
	dataDirectionB int

	timer1interruptEnabled bool
	timer1IRQ              bool // True if timer interrupt flag is set
	timer1latch            int
	timer1counter          int
	timer1freerun          bool
	timer1running          bool
	timer2interruptEnabled bool
	timer2IRQ              bool // True if timer interrupt flag is set
	timer2latch            int
	timer2counter          int
	timer2running          bool

	sendA func(val int)
	sendB func(val int)

	recvA func() int
	recvB func() int

	label string

	IOBaseAddress int

	cpu Interruptable
}

func boolPack(flags ...bool) byte {
	var b byte
	if len(flags) > 8 {
		flags = flags[:8]
	}
	for _, f := range flags {
		b <<= 1
		if f {
			b |= 1
		}
	}
	return b
}

func boolUnpack(b byte, flags ...*bool) {
	if len(flags) > 8 {
		flags = flags[:8]
	}
	for _, f := range flags {
		*f = (b&1 != 0)
		b >>= 1
	}
}

func (r *R6522) FromBytes(b []byte) {
	if len(b) < 15 {
		return
	}
	r.oraReg = int(b[0])
	r.iraReg = int(b[1])
	r.orbReg = int(b[2])
	r.irbReg = int(b[3])
	r.dataDirectionA = int(b[4])
	r.dataDirectionB = int(b[5])
	boolUnpack(
		b[6],
		&r.timer2running,
		&r.timer2IRQ,
		&r.timer2interruptEnabled,
		&r.timer1running,
		&r.timer1freerun,
		&r.timer1IRQ,
		&r.timer1interruptEnabled,
	)
	r.timer1latch = int(b[7]) | (int(b[8]) << 8)
	r.timer1counter = int(b[9]) | (int(b[10]) << 8)
	r.timer2latch = int(b[11]) | (int(b[12]) << 8)
	r.timer2counter = int(b[13]) | (int(b[14]) << 8)
}

func (r *R6522) Bytes() []byte {
	if r == nil {
		return []byte(nil)
	}
	return []byte{
		byte(r.oraReg),
		byte(r.iraReg),
		byte(r.orbReg),
		byte(r.irbReg),
		byte(r.dataDirectionA),
		byte(r.dataDirectionB),
		boolPack(
			r.timer1interruptEnabled,
			r.timer1IRQ,
			r.timer1freerun,
			r.timer1running,
			r.timer2interruptEnabled,
			r.timer2IRQ,
			r.timer2running,
		),
		byte(r.timer1latch & 0xff),
		byte((r.timer1latch >> 8) & 0xff),
		byte(r.timer1counter & 0xff),
		byte((r.timer1counter >> 8) & 0xff),
		byte(r.timer2latch & 0xff),
		byte((r.timer2latch >> 8) & 0xff),
		byte(r.timer2counter & 0xff),
		byte((r.timer2counter >> 8) & 0xff),
	}
}

func (r *R6522) Reset() {
	r.oraReg = 0
	r.iraReg = 0
	r.orbReg = 0
	r.irbReg = 0
	r.dataDirectionA = 0
	r.dataDirectionB = 0
	r.timer1interruptEnabled = true
	r.timer1IRQ = false
	r.timer1latch = 0
	r.timer1counter = 0
	r.timer1freerun = false
	r.timer1running = false
	r.timer2interruptEnabled = true
	r.timer2IRQ = false
	r.timer2latch = 0
	r.timer2counter = 0
	r.timer2running = false
}

func NewR6522(label string, i Interruptable) *R6522 {
	r := &R6522{label: label}
	r.Reset()
	r.timer1freerun = true
	r.timer1latch = 0x1fff
	r.timer1interruptEnabled = false
	r.setRun(true)
	r.cpu = i
	return r
}

func (r *R6522) setRun(b bool) {
	r.timer1running = b
}

func (r *R6522) DoCycle() {
	if r.timer1running {
		r.timer1counter--
		if r.timer1counter < 0 {
			r.timer1counter = r.timer1latch
			if !r.timer1freerun {
				r.timer1running = false
			}
			if r.timer1interruptEnabled {
				r.timer1IRQ = true
				r.cpu.PullIRQLine()
			}
		}
	}
	if r.timer2running {
		r.timer2counter--
		if r.timer2counter < 0 {
			r.timer2running = false
			r.timer2counter = r.timer2latch
			if r.timer2interruptEnabled {
				r.timer2IRQ = true
				r.cpu.PullIRQLine()
			}
		}
	}
	if !r.timer1running && !r.timer2running {
		r.setRun(false)
	}
}

func (r *R6522) sendOutputA(val int) {
	if r.sendA != nil {
		r.sendA(val)
	}
}

func (r *R6522) sendOutputB(val int) {
	if r.sendB != nil {
		r.sendB(val)
	}
}

func (r *R6522) WriteRegister(reg int, val int) {
	if settings.Debug6522 {
		fmt.Printf("6522TRACE: %s: WriteRegister $%.2x -> %s ($%.4x+%d)\n", r.label, val, R6522Register(reg), r.IOBaseAddress, reg)
	}
	value := val & 0x0ff
	rr := R6522Register(reg)
	switch rr {
	case ORB:
		if r.dataDirectionB == 0 {
			//log.Printf("Not sending B - DDRB == 0")
			break
		}
		r.sendOutputB(value & r.dataDirectionB)
		break
	case ORA:
		if r.dataDirectionA == 0 {
			//log.Printf("Not sending A - DDRA == 0")
			break
		}
		//log.Printf("Sending A, r.ddra=%x", r.dataDirectionA)
		r.sendOutputA(value & r.dataDirectionA)
		break
	case ORAH:
		if r.dataDirectionA == 0 {
			//log.Printf("Not sending A - DDRA == 0")
			break
		}
		r.sendOutputA(value & r.dataDirectionA)
		break
	case DDRB:
		r.dataDirectionB = value
		break
	case DDRA:
		r.dataDirectionA = value
		break
	case T1CL:
	case T1LL:
		r.timer1latch = (r.timer1latch & 0x0ff00) | value
		break
	case T1CH:
		r.timer1latch = (r.timer1latch & 0x0ff) | (value << 8)
		r.timer1IRQ = false
		r.timer1counter = r.timer1latch
		r.timer1running = true
		r.setRun(true)
		break
	case T1LH:
		r.timer1latch = (r.timer1latch & 0x0ff) | (value << 8)
		r.timer1IRQ = false
		break
	case T2CL:
		r.timer2latch = (r.timer2latch & 0x0ff00) | value
		break
	case T2CH:
		r.timer2latch = (r.timer2latch & 0x0ff) | (value << 8)
		r.timer2IRQ = false
		r.timer2counter = r.timer2latch
		r.timer2running = true
		r.setRun(true)
		break
	case SR:
		break
	case ACR:
		r.timer1freerun = (value & 64) != 0
		if r.timer1freerun {
			r.timer1running = true
			r.setRun(true)
		}
		break
	case PCR:
		break
	case IFR:
		if (value & 64) != 0 {
			r.timer1IRQ = false
		}
		if (value & 32) != 0 {
			r.timer2IRQ = false
		}
		break
	case IER:
		enable := (value & 128) != 0
		if (value & 64) != 0 {
			r.timer1interruptEnabled = enable
		}
		if (value & 32) != 0 {
			r.timer2interruptEnabled = enable
		}
		break
	default:
	}
}

func (r *R6522) receiveOutputA() int {
	if r.recvA != nil {
		return r.recvA()
	}
	return 0
}

func (r *R6522) receiveOutputB() int {
	if r.recvB != nil {
		return r.recvB()
	}
	return 0
}

func (r *R6522) ReadRegister(reg int) int {
	rr := R6522Register(reg)

	var val int

	if settings.Debug6522 {
		defer func() {
			fmt.Printf("6522TRACE: %s: ReadRegister %s ($%.4x+%d) -> $%.2x\n", r.label, rr, r.IOBaseAddress, reg, val)
		}()
	}

	switch rr {
	case ORB:
		if r.dataDirectionB == 0x0ff {
			break
		}
		val = r.receiveOutputB() & (r.dataDirectionB ^ 0x0ff)
	case ORA:
		if r.dataDirectionA == 0x0ff {
			break
		}
		val = r.receiveOutputA() & (r.dataDirectionA ^ 0x0ff)
	case ORAH:
		if r.dataDirectionA == 0x0ff {
			break
		}
		val = r.receiveOutputA() & (r.dataDirectionA ^ 0x0ff)
	case DDRB:
		val = r.dataDirectionB
	case DDRA:
		val = r.dataDirectionA
	case T1CL:
		r.timer1IRQ = false
		val = r.timer1counter & 0x0ff
	case T1CH:
		val = (r.timer1counter & 0x0ff00) >> 8
	case T1LL:
		val = r.timer1latch & 0x0ff
	case T1LH:
		val = (r.timer1latch & 0x0ff00) >> 8
	case T2CL:
		r.timer2IRQ = false
		val = r.timer2counter & 0x0ff
	case T2CH:
		val = (r.timer2counter & 0x0ff00) >> 8
	case SR:
		val = 0
	case ACR:
		if r.timer1freerun {
			val = 64
			break
		}
		val = 0
	case PCR:
		break
	case IFR:
		val = 0
		if r.timer1IRQ {
			val |= 64
		}
		if r.timer2IRQ {
			val |= 32
		}
		if val != 0 {
			val |= 128
		}
	case IER:
		val = 128
		if r.timer1interruptEnabled {
			val |= 64
		}
		if r.timer2interruptEnabled {
			val |= 32
		}
	}
	return val
}

func (r *R6522) SetBindings(
	sendA, sendB func(val int),
	recvA, recvB func() int,
) {
	r.sendA = sendA
	r.sendB = sendB
	r.recvA = recvA
	r.recvB = recvB
}
