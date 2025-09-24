package servicebus

type DiskTargetString struct {
	Drive    int
	Filename string
}

type DiskTargetBytes struct {
	Drive    int
	Filename string
	Bytes    []byte
}

type LaunchEmulatorTarget struct {
	Filename  string
	Drive     int
	IsControl bool
}

type CPUControlData struct {
	Action string
	Data   map[string]interface{}
}

type MouseButtonState struct {
	Index   int
	Pressed bool
}

type MousePositionState struct {
	X, Y     float64
	WX0, WY0 float64
	WX1, WY1 float64
}

type PlayerJumpCommand struct {
	SyncCount int
}

type KeyAction int

const (
	ActionPress   = 1
	ActionRelease = 0
	ActionRepeat  = 2
)

type KeyMod int

const (
	ModNone  = 0
	ModShift = 1
	ModCtrl  = 2
	ModAlt   = 4
	ModSuper = 8
)

type JoyLine int

const (
	JoystickButton1 JoyLine = 0x20
	JoystickButton0 JoyLine = 0x10
	JoystickUp      JoyLine = 0x08
	JoystickDown    JoyLine = 0x04
	JoystickLeft    JoyLine = 0x02
	JoystickRight   JoyLine = 0x01
)

type JoystickEventData struct {
	Stick int
	Line  JoyLine
}

type KeyEventData struct {
	Key      rune
	ScanCode int
	Action   KeyAction // 0 = release, 1 = press, 2 = repeat
	Modifier KeyMod
}
