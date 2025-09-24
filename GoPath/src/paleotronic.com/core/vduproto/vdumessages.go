package vduproto

import (
	"errors"
	"paleotronic.com/core/types"
)

const (
	VDU_SERVICE      = "9001"
	VDU_DEFAULT_HOST = "localhost"
)

type ClientEventType int

type StringOutEvent struct {
	Content   string
	X         int
	Y         int
	Attribute rune
	FGColor   int
	BGColor   int
}

func (this *StringOutEvent) MarshalBinary() ([]byte, error) {
	data := []byte{byte(types.MtStringOutEvent)}
	data = append(data, byte(this.X), byte(this.Y), byte(this.Attribute), byte(this.FGColor), byte(this.BGColor))
	data = append(data, []byte(this.Content)...)
	return data, nil
}

func (this *StringOutEvent) UnmarshalBinary(data []byte) error {
	if data[0] == byte(types.MtStringOutEvent) {
		return errors.New("Wrong type")
	}
	if len(data) < 7 {
		return errors.New("Not enough data")
	}
	this.X = int(data[1])
	this.Y = int(data[2])
	this.Attribute = rune(data[3])
	this.FGColor = int(data[4])
	this.BGColor = int(data[5])
	this.Content = string(data[6:])
	return nil
}

type CharOutEvent struct {
	Content   rune
	X         int
	Y         int
	Attribute rune
	FGColor   int
	BGColor   int
}

// Specifies a change to the screen contents
type EmptyClientRequest struct {
}

func (this EmptyClientRequest) MarshalBinary() ([]byte, error) {
	return []byte{byte(types.MtEmptyClientRequest)}, nil
}

func (this *EmptyClientRequest) UnmarshalBinary(data []byte) error {
	return nil
}

type ScreenFormatEvent struct {
	Content types.VideoMode
}

type ServerResponse struct {
	Status int
}

type ClientResponse struct {
	Status int
}

type VDUClientEventResponse int

type VDUServerEvent struct {
	Kind string
	Data interface{}
}

type VDUClientEvent struct {
	Kind string
	Data interface{}
}

type VDUServerEventResponse int

// ThinScreenEvent
