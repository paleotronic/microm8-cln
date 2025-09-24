package vduproto

import (
	"errors"
	"paleotronic.com/core/types"
)

// RestalgiaCommand holds instructions for the restalgia system
type RestalgiaCommand struct {
	SubType byte   // Type of event
	Data    []byte // Payload - could be string
}

// MarshalBinary encodes RestalgiaCommand to bytes
func (ce *RestalgiaCommand) MarshalBinary() ([]byte, error) {
	buffer := []byte{types.MtRestalgiaCommand}

	buffer = append(buffer, ce.SubType)
	buffer = append(buffer, ce.Data...)

	return buffer, nil
}

// UnmarshalBinary unpacks RestalgiaCommand from bytes
func (ce *RestalgiaCommand) UnmarshalBinary(data []byte) error {
	if len(data) < 2 {
		return errors.New("not enough data")
	}

	if data[0] != types.MtRestalgiaCommand {
		return errors.New("wrong type")
	}

	ce.SubType = data[1]
	ce.Data = data[2:]

	return nil
}
