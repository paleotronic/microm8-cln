package vduproto

import (
	"errors"

	"paleotronic.com/core/types"
)

// ClickEvent holds buzzer information for network transport
type ClickEvent struct {
	Data []byte
}

// MarshalBinary encodes ClickEvent to bytes
func (ce *ClickEvent) MarshalBinary() ([]byte, error) {
	buffer := []byte{types.MtSpeakerClick}

	// ss := StreamPack{}
	// ss.AddSlice(ce.Data)

	buffer = append(buffer, ce.Data...)

	return buffer, nil
}

// UnmarshalBinary unpacks ClickEvent from bytes
func (ce *ClickEvent) UnmarshalBinary(data []byte) error {
	if len(data) < 1 {
		return errors.New("not enough data")
	}

	if data[0] != types.MtSpeakerClick {
		return errors.New("wrong type")
	}

	// ss := StreamPack{}
	// ss.Data = data[1:]

	// ce.Data, _ = ss.Unwind()
	ce.Data = data[1:]

	return nil
}
