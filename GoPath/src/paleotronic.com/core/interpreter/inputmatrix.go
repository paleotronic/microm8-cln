package interpreter

import "errors"

type InputEventType byte

const (
	ET_NONE InputEventType = iota
	ET_KEYPRESS
	ET_PADDLE_BUTTON_DOWN
	ET_PADDLE_BUTTON_UP
	ET_PADDLE_VALUE_CHANGE
)

// Event wrapper for locally detected events
type InputEvent struct {
	Kind  InputEventType
	ID    byte
	Value int
}

type InputAction struct {
	Kind InputEventType
	ID   byte
}

type InputMatrix struct {
	Data map[InputEventType]map[byte]InputAction
}

// NewStdInputMatrix creates a standard input matrix
func NewStdInputMatrix() *InputMatrix {

	this := &InputMatrix{}

	this.Data = make(map[InputEventType]map[byte]InputAction)

	return this

}

func (this *InputMatrix) FilterEvent(e InputEvent) InputEvent {

	etmap, ok := this.Data[e.Kind]
	if !ok {
		return e
	}

	action, ok := etmap[e.ID]
	if !ok {
		return e
	}

	e.Kind = action.Kind
	e.ID = action.ID

	return e

}

func (this *InputEvent) MarshalBinary() ([]byte, error) {

	data := make([]byte, 0)
	data = append(data, byte(this.Kind))
	data = append(data, byte(this.ID))
	data = append(data, byte((this.Value>>24)&0xff))
	data = append(data, byte((this.Value>>16)&0xff))
	data = append(data, byte((this.Value>>8)&0xff))
	data = append(data, byte((this.Value>>0)&0xff))

	return data, nil
}

func (this *InputEvent) UnmarshalBinary(data []byte) error {

	if len(data) != 6 {
		return errors.New("Not enough data")
	}

	this.Kind = InputEventType(data[0])
	this.ID = data[1]
	this.Value = int(data[2])<<24 | int(data[3])<<16 | int(data[4])<<8 | int(data[5])

	return nil
}
