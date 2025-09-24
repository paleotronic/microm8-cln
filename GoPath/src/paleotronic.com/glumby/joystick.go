package glumby

import (
	"math"

	"paleotronic.com/log"

	"github.com/go-gl/glfw/v3.2/glfw"
	//"math"
)

type ControllerEventType int

const (
	ControllerButtonEvent   ControllerEventType = 1 << iota
	ControllerAxisEvent     ControllerEventType = 1 << iota
	ControllerAxisPullEvent ControllerEventType = 1 << iota
)

type ControllerEvent struct {
	Type    ControllerEventType
	ID      Joystick
	Index   int
	Value   float32
	Pressed bool
}

type Controller struct {
	ID             Joystick
	Name           string
	lastButton     []byte
	lastAxisValues []float32
}

// EnumerateControllers returns active controllers on the system
func EnumerateControllers() []*Controller {
	joy := make([]*Controller, 0)
	for j := Joystick1; j < JoystickLast; j++ {
		if glfw.JoystickPresent(glfw.Joystick(j)) {
			con := &Controller{ID: j, Name: glfw.GetJoystickName(glfw.Joystick(j))}
			joy = append(joy, con)
			log.Printf("Controller: %v\n", con)
		}
	}
	return joy
}

func (c *Controller) ButtonPressed(id int) bool {
	c.lastButton = glfw.GetJoystickButtons(glfw.Joystick(c.ID))
	if id < 0 || id >= len(c.lastButton) {
		return false
	}
	return (c.lastButton[id] == byte(Press))
}

// nudgeAxis helps for logitech controllers that don't quite ever get to 1/-1
func nudgeAxis(v float32) float32 {
	sgn := float32(1)
	if v < 0 {
		sgn = -1
	}
	a := float32(math.Abs(float64(v)))
	if a <= 0.010 {
		a = 0
	} else if a >= 0.990 {
		a = 1
	}
	return sgn * a
}

func (c *Controller) AxisValue(id int) float32 {
	c.lastAxisValues = glfw.GetJoystickAxes(glfw.Joystick(c.ID))
	if id < 0 || id >= len(c.lastAxisValues) {
		return 0
	}
	return nudgeAxis(c.lastAxisValues[id])
}

func (c *Controller) GetEvents() []*ControllerEvent {
	ev := make([]*ControllerEvent, 0)

	bb := glfw.GetJoystickButtons(glfw.Joystick(c.ID))
	aa := glfw.GetJoystickAxes(glfw.Joystick(c.ID))

	if len(bb) != len(c.lastButton) || len(aa) != len(c.lastAxisValues) {
		c.lastAxisValues = aa
		c.lastButton = bb
		return ev
	}

	for i, axis := range aa {
		if axis != c.lastAxisValues[i] {
			ev = append(ev, &ControllerEvent{ID: c.ID, Type: ControllerAxisEvent, Value: axis, Index: i})
			//log.Println(ControllerEvent{ID: c.ID, Type: ControllerAxisEvent, Value: axis, Index: i})
		}
		//if math.Abs(float64(axis)) > 0.01 {
		//	ev = append(ev, &ControllerEvent{ID: c.ID, Type: ControllerAxisEvent, Value: axis, Index: i})
		//	log.Println( ControllerEvent{ID: c.ID, Type: ControllerAxisEvent, Value: axis, Index: i} )
		//}
	}

	for i, button := range bb {
		if button != c.lastButton[i] {
			ev = append(ev, &ControllerEvent{ID: c.ID, Type: ControllerButtonEvent, Pressed: (button == byte(Press)), Index: i})
		}
	}

	c.lastAxisValues = aa
	c.lastButton = bb

	return ev
}
