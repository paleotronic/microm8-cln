package vduproto

import (
	"errors"
	"paleotronic.com/fmt"
	"paleotronic.com/core/types"
)

type PaddleButtonEvent struct {
	PaddleID    byte // 0 - 4
	ButtonState byte // 0 / != 0
}

type PaddleValueEvent struct {
	PaddleID    byte
	PaddleValue byte
}

type PaddleModifyEvent struct {
	PaddleID   byte
	Difference int8
}

func (this PaddleButtonEvent) MarshalBinary() ([]byte, error) {
	b := make([]byte, 3)
	b[0] = byte(types.MtPaddleButtonEvent)
	b[1] = byte(this.PaddleID)
	b[2] = byte(this.ButtonState)

	return b, nil
}

func (this *PaddleButtonEvent) UnmarshalBinary(data []byte) error {
	if len(data) < 3 {
		return errors.New(fmt.Sprintf("Incorrect length: Expected %d, got %d", 3, len(data)))
	}
	if data[0] != types.MtPaddleButtonEvent {
		return errors.New(fmt.Sprintf("Incorrect type: Expected %d, got %d", types.MtPaddleButtonEvent, data[0]))
	}
	this.PaddleID = data[1]
	this.ButtonState = data[2]
	return nil
}

// PaddleValueEvent

func (this PaddleValueEvent) MarshalBinary() ([]byte, error) {
	b := make([]byte, 3)
	b[0] = byte(types.MtPaddleValueEvent)
	b[1] = byte(this.PaddleID)
	b[2] = byte(this.PaddleValue)

	return b, nil
}

func (this *PaddleValueEvent) UnmarshalBinary(data []byte) error {
	if len(data) < 3 {
		return errors.New(fmt.Sprintf("Incorrect length: Expected %d, got %d", 3, len(data)))
	}
	if data[0] != types.MtPaddleValueEvent {
		return errors.New(fmt.Sprintf("Incorrect type: Expected %d, got %d", types.MtPaddleValueEvent, data[0]))
	}
	this.PaddleID = data[1]
	this.PaddleValue = data[2]
	return nil
}

// PaddleModifyEvent

func (this PaddleModifyEvent) MarshalBinary() ([]byte, error) {
	b := make([]byte, 3)
	b[0] = byte(types.MtPaddleModifyEvent)
	b[1] = byte(this.PaddleID)
	b[2] = byte(this.Difference)

	return b, nil
}

func (this *PaddleModifyEvent) UnmarshalBinary(data []byte) error {
	if len(data) < 3 {
		return errors.New(fmt.Sprintf("Incorrect length: Expected %d, got %d", 3, len(data)))
	}
	if data[0] != types.MtPaddleModifyEvent {
		return errors.New(fmt.Sprintf("Incorrect type: Expected %d, got %d", types.MtPaddleModifyEvent, data[0]))
	}
	this.PaddleID = data[1]
	this.Difference = int8(data[2])
	return nil
}
