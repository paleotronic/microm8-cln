package vduproto

import (
	"errors"
	"paleotronic.com/core/types"
)

type MemEvent struct {
	AL, AH byte
	Value  byte
}

func (me *MemEvent) MarshalBinary() ([]byte, error) {
	data := make([]byte, 4)
	data[0] = types.MtMemoryEvent
	data[1] = me.AL
	data[2] = me.AH
	data[3] = me.Value

	return data, nil
}

func (me *MemEvent) UnmarshalBinary(data []byte) error {
	if len(data) < 4 {
		return errors.New("Not enough data")
	}
	if data[0] != types.MtMemoryEvent {
		return errors.New("Wrong type")
	}
	me.AL = data[1]
	me.AH = data[2]
	me.Value = data[3]

	return nil
}

// -------------------------

type CPUEvent struct {
	AL, AH byte
}

func (me *CPUEvent) MarshalBinary() ([]byte, error) {
	data := make([]byte, 3)
	data[0] = types.MtCallEvent
	data[1] = me.AL
	data[2] = me.AH

	return data, nil
}

func (me *CPUEvent) UnmarshalBinary(data []byte) error {
	if len(data) < 3 {
		return errors.New("Not enough data")
	}
	if data[0] != types.MtCallEvent {
		return errors.New("Wrong type")
	}
	me.AL = data[1]
	me.AH = data[2]

	return nil
}
