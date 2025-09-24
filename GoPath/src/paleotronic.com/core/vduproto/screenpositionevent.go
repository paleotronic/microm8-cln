package vduproto

import (
	"errors"
	"paleotronic.com/fmt"
	"paleotronic.com/core/types"
)

// ScreenPositionEvent updates the cursor position
type ScreenPositionEvent struct {
	LayerID int
	X int
	Y int
}

func (this ScreenPositionEvent) Identity() byte {
	return types.MtScreenPositionEvent
}

func (this ScreenPositionEvent) MarshalBinary() ([]byte, error) {
	return []byte{byte(this.Identity()), byte(this.LayerID), byte(this.X), byte(this.Y)}, nil
}

func (this *ScreenPositionEvent) UnmarshalBinary(data []byte) error {
	if len(data) < 3 {
		return errors.New(fmt.Sprintf("Incorrect length: Expected %d, got %d", 3, len(data)))
	}

	if data[0] != this.Identity() {
		return errors.New(fmt.Sprintf("Incorrect type: Expected %d, got %d", this.Identity(), data[0]))
	}

	this.LayerID = int(data[1])
	this.X = int(data[2])
	this.Y = int(data[3])

	return nil
}
