package vduproto

import (
	"errors"
	"paleotronic.com/core/types"
)

type ColorEvent struct {
	R, G, B, A byte
}

func (me *ColorEvent) MarshalBinary() ([]byte, error) {
	data := make([]byte, 5)
	data[0] = types.MtBGColorEvent
	data[1] = me.R
	data[2] = me.G
	data[3] = me.B
	data[4] = me.A

	return data, nil
}

func (me *ColorEvent) UnmarshalBinary(data []byte) error {
	if len(data) < 5 {
		return errors.New("Not enough data")
	}
	if data[0] != types.MtBGColorEvent {
		return errors.New("Wrong type")
	}
	me.R = data[1]
	me.G = data[2]
	me.B = data[3]
	me.A = data[4]

	return nil
}
