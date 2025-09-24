package vduproto

import (
	"errors"
	"paleotronic.com/fmt"
	"paleotronic.com/core/types"
)

type KeyPressEvent struct {
	Streamable
	Character rune
}

func (this KeyPressEvent) Identity() byte {
	return types.MtKeyPressEvent
}

func (this KeyPressEvent) MarshalBinary() ([]byte, error) {
	b := make([]byte, 3)
	b[0] = byte(this.Identity())
	b[1] = byte(this.Character % 256)
	b[2] = byte(this.Character / 256)

	return b, nil
}

func (this *KeyPressEvent) UnmarshalBinary(data []byte) error {

	//////fmt.Println("UnmarshalBinary()")

	if len(data) < 3 {
		return errors.New(fmt.Sprintf("Incorrect length: Expected %d, got %d", 3, len(data)))
	}

	if data[0] != this.Identity() {
		return errors.New(fmt.Sprintf("Incorrect type: Expected %d, got %d", this.Identity(), data[0]))
	}

	this.Character = rune(int(data[1]) + 256*int(data[2]))

	return nil
}
