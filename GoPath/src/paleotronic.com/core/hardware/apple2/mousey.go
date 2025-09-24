package apple2

import (
	"math"
	"sync"
	"strings"

	"gopkg.in/yaml.v2"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/hardware/cpu/mos6502"
	"paleotronic.com/core/hardware/servicebus"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/memory"
	"paleotronic.com/log"
	"paleotronic.com/core/settings"
)

const MouseUpdateTicks = 1020484 / 60

type MouseState struct {
	Button0Pressed  bool
	PButton0Pressed bool
	Button1Pressed  bool
	PButton1Pressed bool
	Position        Point
	PPosition       Point
	ClampWindow     Rectangle
	PClampWindow    Rectangle
	SavedPoint      Point
}

type MouseCardState struct {
	Active             bool
	Mode               int
	Status             int
	IRQOnMove          bool
	IRQOnButton        bool
	IRQOnVBlank        bool
	InIRQ              bool
	InVBlank           bool
	MovedSinceLastRead bool
	MovedSinceLastTick bool
	Mouse              MouseState
}

type IOCardMouse struct {
	IOCard
	MouseCardState
	Int        interfaces.Interpretable
	bus        []*servicebus.ServiceBusRequest
	busm       sync.Mutex
	cyclecount int
	//terminate chan bool
}

func (d *IOCardMouse) ImA() string {
	return "mousecard"
}

func (d *IOCardMouse) Increment(n int) {
	//
	d.HandleServiceBusInjection(d.HandleServiceBusRequest)
	d.cyclecount += n
	if d.cyclecount >= MouseUpdateTicks {
		d.cyclecount -= MouseUpdateTicks
		// if !d.IRQOnButton && !d.IRQOnMove {
		// 	return
		// }
		// if d.IRQOnButton {
		// 	if d.Mouse.Button0Pressed != d.Mouse.PButton0Pressed || d.Mouse.Button1Pressed != d.Mouse.PButton1Pressed {
		// 		d.InIRQ = true
		// 		cpu := apple2helpers.GetCPU(d.Int)
		// 		cpu.RequestInterrupt = true
		// 		return
		// 	}
		// }
		// if d.IRQOnMove {
		// 	if d.MovedSinceLastTick {
		// 		d.InIRQ = true
		// 		cpu := apple2helpers.GetCPU(d.Int)
		// 		cpu.RequestInterrupt = true
		// 	}
		// }
		// d.MovedSinceLastTick = false
	}
}

func (d *IOCardMouse) Decrement(n int) {
	//
}

func (d *IOCardMouse) AdjustClock(n int) {
	//
}

func (d *IOCardMouse) GetYAML() []byte {
	data, _ := yaml.Marshal(d.MouseCardState)
	return data
}

func (d *IOCardMouse) SetYAML(b []byte) {
	var s MouseCardState
	if yaml.Unmarshal(b, &s) == nil {
		d.MouseCardState = s
	}
}

func (d *IOCardMouse) HandleServiceBusRequest(r *servicebus.ServiceBusRequest) (*servicebus.ServiceBusResponse, bool) {

	//log.Printf("Got ServiceBusRequest(%s)", r)

	//index := d.e.GetMemIndex()

	switch r.Type {
	case servicebus.MouseButton:
		ev := r.Payload.(*servicebus.MouseButtonState)
		//log.Printf("Got mouse button event: %+v", ev)
		switch ev.Index {
		case 0:
			d.Mouse.Button0Pressed = ev.Pressed
		case 1:
			d.Mouse.Button1Pressed = ev.Pressed
		}
	case servicebus.MousePosition:
		ev := r.Payload.(*servicebus.MousePositionState)
		//log.Printf("Got mouse position event: %+v", ev)
		d.Mouse.PPosition.X = ev.X
		d.Mouse.PPosition.Y = ev.Y
		d.MovedSinceLastRead = true
		d.MovedSinceLastTick = true
	case servicebus.BeginVBLANK:
		d.VBlank()
	}

	return &servicebus.ServiceBusResponse{}, true

}

func (d *IOCardMouse) VBlank() {
	if d.IRQOnVBlank && d.Active {
		d.InVBlank = true
		d.InIRQ = true
		cpu := apple2helpers.GetCPU(d.Int)
		cpu.PullIRQLine()
	}
}

func (d *IOCardMouse) InjectServiceBusRequest(r *servicebus.ServiceBusRequest) {
	d.busm.Lock()
	defer d.busm.Unlock()
	d.bus = append(d.bus, r)
}

func (d *IOCardMouse) HandleServiceBusInjection(handler func(r *servicebus.ServiceBusRequest) (*servicebus.ServiceBusResponse, bool)) {
	if len(d.bus) > 0 {
		d.busm.Lock()
		defer d.busm.Unlock()
		for _, r := range d.bus {
			handler(r)
		}
		d.bus = d.bus[:0]
	}
}

func (d *IOCardMouse) Reset() {
	d.MouseCardState.Mode = 0
	d.MouseCardState.Status = 0
	d.Mouse.ClampWindow = Rect(0x000, 0x000, 0x3ff, 0x3ff)
}

func (d *IOCardMouse) Init(slot int) {
	d.IOCard.Init(slot)
	log.Println("Initialising mousecard...")
	d.Reset()
	d.Int.SetCycleCounter(d)
	servicebus.Subscribe(
		d.Int.GetMemIndex(),
		servicebus.MouseButton,
		d,
	)
	servicebus.Subscribe(
		d.Int.GetMemIndex(),
		servicebus.MousePosition,
		d,
	)
	servicebus.Subscribe(
		d.Int.GetMemIndex(),
		servicebus.BeginVBLANK,
		d,
	)
	d.bus = make([]*servicebus.ServiceBusRequest, 0, 16)
	//d.terminate = make(chan bool)
}

func (d *IOCardMouse) Done(slot int) {
	//d.terminate <- true
	servicebus.Unsubscribe(slot, d)
}

func (d *IOCardMouse) HandleIO(register int, value *uint64, eventType IOType) {

	log.Printf("%s of register %.2x (%.2x:%s)\n", eventType.String(), register, *value, string(rune(*value-128)))

}

func (d *IOCardMouse) FirmwareRead(offset int) uint64 {
	//log.Printf("IOCardMouse firmware read @ %.2x", offset)
	
	hasVIDHD := strings.Contains( settings.SpecFile[d.Int.GetMemIndex()], "apple2e" )
	//~ hasVIDHD = true
	
	if hasVIDHD {
		switch offset {
		case 0x00:
			return 0x24
		case 0x01:
			return 0xea
		case 0x02:
			return 0x4c
		}
	} else {
		switch offset {
		case 0x00:
			return 0x2c
		case 0x01:
			return 0x58
		case 0x02:
			return 0xff
		}		
	}

	switch offset {
	case 0x05:
		return 0x38
	case 0x07:
		return 0x18
	case 0x0B:
		return 0x01
	case 0x0C:
		return 0x20
	case 0x0FB:
		return 0xD6
	case 0x08:
		return 0x00
	case 0x011:
		return 0x00
	case 0x12:
		return 0x80
	case 0x13:
		return 0x81
	case 0x14:
		return 0x82
	case 0x15:
		return 0x83
	case 0x16:
		return 0x84
	case 0x17:
		return 0x85
	case 0x18:
		return 0x86
	case 0x19:
		return 0x87
	case 0x1A:
		return 0x88
	default:
		return 0x60
	}

}

func (d *IOCardMouse) FirmwareWrite(offset int, value uint64) {
	//
	log.Printf("IOCardMouse firmware write @ 0x%.2x < 0x%.2x", offset, value)
}

func (d *IOCardMouse) FirmwareExec(
	offset int,
	PC, A, X, Y, SP, P *int,
) int64 {
	//log.Printf("IOCardMouse firmware exec @ %.2x", offset)

	defer func() {
		cpu := apple2helpers.GetCPU(d.Int)
		rts := cpu.Opref[0x60]
		rts.Do(cpu)
		//log.Printf("Return to PC = 0x%.4x", cpu.PC)
	}()

	switch offset - 0x80 {
	case 0:
		return d.setMouseMode()
	case 1:
		return d.mouseIRQ()
	case 2:
		return d.readMouse()
	case 3:
		return d.clearMouse()
	case 4:
		return d.posMouse()
	case 5:
		return d.clampMouse()
	case 6:
		return d.homeMouse()
	case 7:
		return d.initMouse()
	case 8:
		return d.getMouseClamp()
	}

	return 0
}

func (d *IOCardMouse) setActive(active bool) {
	d.Active = active
	log.Printf("Set mouse active state to %v", active)
	if !active {
		d.Mode = 0x00
		d.IRQOnMove = false
		d.IRQOnButton = false
	}
}

func (d *IOCardMouse) setMouseMode() int64 {
	cpu := apple2helpers.GetCPU(d.Int)

	mode := cpu.A
	if mode > 0x0f {
		cpu.SetFlag(mos6502.F_C, true)
		return 0
	}
	cpu.SetFlag(mos6502.F_C, false)

	d.IRQOnVBlank = (mode & 0x08) != 0

	if (mode & 0x01) == 0 {
		d.setActive(false)
		d.IRQOnMove = false
		d.IRQOnButton = false
		return 0
	}

	d.IRQOnMove = (mode & 0x02) != 0
	d.IRQOnButton = (mode & 0x04) != 0

	d.setActive(true)

	return 0
}

func (d *IOCardMouse) mouseIRQ() int64 {
	cpu := apple2helpers.GetCPU(d.Int)
	if d.InIRQ {
		d.publishMouseState()
	}
	cpu.SetFlag(mos6502.F_C, !d.InIRQ)
	return 0
}

func (d *IOCardMouse) readMouse() int64 {
	d.publishMouseState()
	d.InIRQ = false
	d.InVBlank = false
	cpu := apple2helpers.GetCPU(d.Int)
	cpu.SetFlag(mos6502.F_C, false)
	return 0
}

func (d *IOCardMouse) clearMouse() int64 {
	d.InVBlank = false
	d.InIRQ = false
	d.Mouse.Button0Pressed = false
	d.Mouse.Button1Pressed = false
	d.Mouse.PButton0Pressed = false
	d.Mouse.PButton1Pressed = false
	d.homeMouse()
	return 0
}

func (d *IOCardMouse) posMouse() int64 {
	cpu := apple2helpers.GetCPU(d.Int)

	x := float64(d.Int.GetMemory(0x0478) + 256*d.Int.GetMemory(0x0578))
	y := float64(d.Int.GetMemory(0x04f8) + 256*d.Int.GetMemory(0x05f8))

	log.Printf("Pos requested %f, %f", x, y)

	if y >= 32000 || x >= 32000 {
		// restore last point
		x := d.Mouse.SavedPoint.X
		y := d.Mouse.SavedPoint.Y

		d.Int.SetMemory(0x0478+d.Slot, uint64(x)&0x0ff)
		d.Int.SetMemory(0x04F8+d.Slot, uint64(y)&0x0ff)
		d.Int.SetMemory(0x0578+d.Slot, (uint64(x)&0x0ff00)>>8)
		d.Int.SetMemory(0x05F8+d.Slot, (uint64(y)&0x0ff00)>>8)
	}

	cpu.SetFlag(mos6502.F_C, false)
	return 0
}

/*
 * $Cn17 CLAMPMOUSE Sets mouse bounds in a window
 *      Sets up clamping window for mouse user
 *      Power up defaults are 0 - 1023 (0 - 3ff)
 *      Caller sets:
 *      A = 0 if setting X, 1 if setting Y
 *      $0478 = low byte of low clamp.
 *      $04F8 = low byte of high clamp.
 *      $0578 = high byte of low clamp.
 *      $05F8 = high byte of high clamp.
 *      //gs homes mouse to low address, but //c and //e do not
 */
func (d *IOCardMouse) clampMouse() int64 {
	cpu := apple2helpers.GetCPU(d.Int)
	setX := cpu.A == 0
	setY := cpu.A == 1

	min := float64(d.Int.GetMemory(0x0478) + 256*d.Int.GetMemory(0x0578))
	max := float64(d.Int.GetMemory(0x04f8) + 256*d.Int.GetMemory(0x05f8))

	if min >= 32768 {
		min -= 65536
	}
	if max >= 32768 {
		max -= 65536
	}

	if setX {
		if max == 32767 {
			max = 560
		}
		log.Printf("Set Clamp-X to min: %f, max: %f", min, max)
		d.Mouse.ClampWindow.X0 = min
		d.Mouse.ClampWindow.X1 = max
	}
	if setY {
		if max == 32767 {
			max = 192
		}
		log.Printf("Set Clamp-Y to min: %f, max: %f", min, max)
		d.Mouse.ClampWindow.Y0 = min
		d.Mouse.ClampWindow.Y1 = max
	}

	return 0
}

func (d *IOCardMouse) homeMouse() int64 {
	d.Mouse.PPosition = Pt(0.5, 0.5)
	d.publishMouseState()
	cpu := apple2helpers.GetCPU(d.Int)
	cpu.SetFlag(mos6502.F_C, false)
	return 0
}

func (d *IOCardMouse) initMouse() int64 {
	d.Mouse.ClampWindow = Rect(0, 0, 0x3ff, 0x3ff)
	d.clearMouse()
	return 0
}

/**
 * Described in Apple Mouse technical note #7
 * Cn1A: Read mouse clamping values
 * Register number is stored in $478 and ranges from x47 to x4e
 * Return value should be stored in $5782
 * Values should be returned in this order:
 * MinXH, MinYH, MinXL, MinYL, MaxXH, MaxYH, MaxXL, MaxYL
 */
func (d *IOCardMouse) getMouseClamp() int64 {
	reg := int(d.Int.GetMemory(0x478)) - 0x047
	var val uint64 = 0

	switch reg {
	case 0:
		val = (uint64(d.Mouse.ClampWindow.MinX()) >> 8)
	case 1:
		val = (uint64(d.Mouse.ClampWindow.MinY()) >> 8)
	case 2:
		val = (uint64(d.Mouse.ClampWindow.MinX()) & 255)
	case 3:
		val = (uint64(d.Mouse.ClampWindow.MinY()) & 255)
	case 4:
		val = (uint64(d.Mouse.ClampWindow.MaxX()) >> 8)
	case 5:
		val = (uint64(d.Mouse.ClampWindow.MaxY()) >> 8)
	case 6:
		val = (uint64(d.Mouse.ClampWindow.MaxX()) & 255)
	case 7:
		val = (uint64(d.Mouse.ClampWindow.MaxY()) & 255)
	}

	d.Int.SetMemory(0x578, val)

	return 0
}

func (d *IOCardMouse) publishMouseState() {

	if !d.Active {
		return
	}

	var x = d.Mouse.PPosition.X
	x *= d.Mouse.ClampWindow.Width()
	x += d.Mouse.ClampWindow.MinX()
	x = math.Min(math.Max(x, d.Mouse.ClampWindow.MinX()), d.Mouse.ClampWindow.MaxX())

	var y = d.Mouse.PPosition.Y
	y *= d.Mouse.ClampWindow.Height()
	y += d.Mouse.ClampWindow.MinY()
	y = math.Min(math.Max(y, d.Mouse.ClampWindow.MinY()), d.Mouse.ClampWindow.MaxY())

	d.Mouse.SavedPoint = Pt(x, y)

	//log.Printf("publish: Mouse X, Y = %f,%f", x, y)

	/*
	 * $0478 + slot Low byte of absolute X position
	 * $04F8 + slot Low byte of absolute Y position
	 */
	d.Int.SetMemory(0x0478+d.Slot, uint64(x)&0x0ff)
	d.Int.SetMemory(0x04F8+d.Slot, uint64(y)&0x0ff)
	/*
	 * $0578 + slot High byte of absolute X position
	 * $05F8 + slot High byte of absolute Y position
	 */
	d.Int.SetMemory(0x0578+d.Slot, (uint64(x)&0x0ff00)>>8)
	d.Int.SetMemory(0x05F8+d.Slot, (uint64(y)&0x0ff00)>>8)
	/*
	 * $0678 + slot Reserved and used by the firmware
	 * $06F8 + slot Reserved and used by the firmware
	 *
	 * Interrupt status byte:
	 * Set by READMOUSE
	 * Bit 7 6 5 4 3 2 1 0
	 *     | | | | | | | |
	 *     | | | | | | | `--- Previously, button 1 was up (0) or down (1)
	 *     | | | | | | `----- Movement interrupt
	 *     | | | | | `------- Button 0/1 interrupt
	 *     | | | | `--------- VBL interrupt
	 *     | | | `----------- Currently, button 1 is up (0) or down (1)
	 *     | | `------------- X/Y moved since last READMOUSE
	 *     | `--------------- Previously, button 0 was up (0) or down (1)
	 *     `----------------- Currently, button 0 is up (0) or down (1)
	 */
	var status = 0
	if d.Mouse.PButton1Pressed {
		status |= 1
	}
	if d.IRQOnMove && d.MovedSinceLastRead {
		status |= 2
	}
	if d.IRQOnButton && (d.Mouse.Button0Pressed != d.Mouse.PButton0Pressed || d.Mouse.Button1Pressed != d.Mouse.PButton1Pressed) {
		status |= 4
	}
	if d.InVBlank {
		status |= 8
	}
	if d.Mouse.Button1Pressed {
		status |= 16
	}
	if d.MovedSinceLastRead {
		status |= 32
	}
	if d.Mouse.PButton0Pressed {
		status |= 64
	}
	if d.Mouse.Button0Pressed {
		status |= 128
	}
	/*
	 * $0778 + slot Button 0/1 interrupt status byte
	 */
	d.Int.SetMemory(0x0778+d.Slot, uint64(status))

	/*
	 * $07F8 + slot Mode byte
	 */
	d.Int.SetMemory(0x07F8+d.Slot, uint64(d.Mode))

	d.Mouse.PButton0Pressed = d.Mouse.Button0Pressed
	d.Mouse.PButton1Pressed = d.Mouse.Button1Pressed
	d.MovedSinceLastRead = false
}

func NewIOCardMouse(mm *memory.MemoryMap, index int, ent interfaces.Interpretable) *IOCardMouse {
	this := &IOCardMouse{}
	this.SetMemory(mm, index)
	this.Int = ent
	this.Name = "IOCardMouse"
	this.IsFWHandler = true
	return this
}

/* Point and Rectangle types */
type Point struct {
	X, Y float64
}

type Rectangle struct {
	X0, Y0 float64
	X1, Y1 float64
}

func Pt(x, y float64) Point {
	return Point{x, y}
}

func Rect(x0, y0, x1, y1 float64) Rectangle {
	return Rectangle{x0, y0, x1, y1}
}

func (r Rectangle) MinX() float64 {
	if r.X0 < r.X1 {
		return r.X0
	}
	return r.X1
}

func (r Rectangle) MinY() float64 {
	if r.Y0 < r.Y1 {
		return r.Y0
	}
	return r.Y1
}

func (r Rectangle) MaxX() float64 {
	if r.X0 < r.X1 {
		return r.X1
	}
	return r.X0
}

func (r Rectangle) MaxY() float64 {
	if r.Y0 < r.Y1 {
		return r.Y1
	}
	return r.Y0
}

func (r Rectangle) Width() float64 {
	return r.MaxX() - r.MinX() + 1
}

func (r Rectangle) Height() float64 {
	return r.MaxY() - r.MinY() + 1
}

func (r Rectangle) ContainsPoint(p Point) bool {
	return p.X >= r.MinX() && p.X <= r.MaxX() && p.Y >= r.MinY() && p.Y <= r.MaxY()
}
