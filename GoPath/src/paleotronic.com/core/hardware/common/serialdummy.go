package common

import "paleotronic.com/fmt"

type SerialDevice interface {
	IsConnected() bool
	InputAvailable() bool
	CanSend() bool
	GetInputByte() int
	SendOutputByte(value int)
	Stop()
}

type SerialDummyDevice struct {
	Data   []byte
	Ptr    int
	inText bool
	inHead bool
	text   string
}

func (d *SerialDummyDevice) CanSend() bool {
	return true
}

func (d *SerialDummyDevice) HasInput() bool {
	return d.Ptr < len(d.Data)
}

func (d *SerialDummyDevice) Recv() int {

	if !d.HasInput() {
		return 0
	}

	v := d.Data[d.Ptr]
	d.Ptr++

	fmt.Printf("<- 0x%.2X\n", v)

	return int(v)
}

func (d *SerialDummyDevice) Write(data []byte) {
	d.Data = data
	d.Ptr = 0
}

func (d *SerialDummyDevice) SendOutputByte(value int) {

	fmt.Printf("-> 0x%.2X\n", value)

	d.Write([]byte{byte(value)})

}
