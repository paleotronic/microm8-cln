package vduproto

import (
	"errors"

	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

// ClickEvent holds buzzer information for network transport
type ScanLineEvent struct {
	Y    byte
	Data []byte
}

// MarshalBinary encodes ClickEvent to bytes
func (ce *ScanLineEvent) MarshalBinary() ([]byte, error) {
	buffer := []byte{types.MtScanLineEvent, ce.Y}

	if len(ce.Data) < 40 {
		panic("Not enough scanline " + utils.IntToStr(len(ce.Data)))
	}

	ss := StreamPack{}
	ss.AddSlice(ce.Data)

	buffer = append(buffer, ss.Data...)

	return buffer, nil
}

// UnmarshalBinary unpacks ClickEvent from bytes
func (ce *ScanLineEvent) UnmarshalBinary(data []byte) error {
	if len(data) < 2 {
		return errors.New("not enough data")
	}

	if data[0] != types.MtScanLineEvent {
		return errors.New("wrong type")
	}

	ce.Y = data[1]

	ss := StreamPack{}
	ss.Data = data[2:]

	ce.Data, _ = ss.Unwind()

	return nil
}
