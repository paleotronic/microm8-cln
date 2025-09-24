package vduproto

import (
	"errors"
	"paleotronic.com/core/types"
)

// ConnectCommand holds instructions for the restalgia system
type ConnectCommand struct {
	Data    []byte // Payload - could be string
}

// MarshalBinary encodes ConnectCommand to bytes
func (ce *ConnectCommand) MarshalBinary() ([]byte, error) {
	buffer := []byte{types.MtConnectCommand}

	buffer = append(buffer, ce.Data...)

	return buffer, nil
}

// UnmarshalBinary unpacks ConnectCommand from bytes
func (ce *ConnectCommand) UnmarshalBinary(data []byte) error {
	if len(data) < 2 {
		return errors.New("not enough data")
	}

	if data[0] != types.MtConnectCommand {
		return errors.New("wrong type")
	}

	ce.Data = data[1:]

	return nil
}
