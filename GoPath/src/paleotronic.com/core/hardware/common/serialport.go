package common

import (
	"fmt"
	"go.bug.st/serial"
	"io"
)

const serialStagingBuffer = 4096

type SerialPortDevice struct {
	RecvData          chan byte
	SendData          chan byte
	Ptr               int
	port              serial.Port
	r                 bool
	mode              *serial.Mode
	xOff              bool
	allowSoftwareFlow bool
}

func parityValue(parity string) (serial.Parity, error) {
	switch parity {
	case "N":
		return serial.NoParity, nil
	case "O":
		return serial.OddParity, nil
	case "E":
		return serial.EvenParity, nil
	case "M":
		return serial.MarkParity, nil
	case "S":
		return serial.SpaceParity, nil
	}
	return serial.NoParity, fmt.Errorf("Invalid parity setting: '%s'", parity)
}

func stopValue(stopBits string) (serial.StopBits, error) {
	switch stopBits {
	case "1":
		return serial.OneStopBit, nil
	case "1.5":
		return serial.OnePointFiveStopBits, nil
	case "2":
		return serial.TwoStopBits, nil
	}
	return serial.StopBits(serial.InvalidStopBits), fmt.Errorf("Invalid stopBits setting: '%s'", stopBits)
}

func NewSerialPortDevice(device string, baudRate int, parity string, dataBits int, stopBits string, allowXOFF bool) (*SerialPortDevice, error) {
	pval, err := parityValue(parity)
	if err != nil {
		return nil, err
	}
	sval, err := stopValue(stopBits)
	if err != nil {
		return nil, err
	}
	port, err := serial.Open(device, &serial.Mode{})
	if err != nil {
		return nil, err
	}
	mode := &serial.Mode{
		BaudRate: baudRate,
		Parity:   pval,
		DataBits: dataBits,
		StopBits: sval,
	}
	if err := port.SetMode(mode); err != nil {
		return nil, err
	}
	fmt.Println("Serial port initialized: " + device)
	spd := &SerialPortDevice{
		port:              port,
		mode:              mode,
		RecvData:          make(chan byte, serialStagingBuffer),
		SendData:          make(chan byte, serialStagingBuffer),
		allowSoftwareFlow: allowXOFF,
	}
	spd.Start()
	return spd, nil
}

func (d *SerialPortDevice) ChangeMode(baudRate int, dataBits int, parity string, stopBits string) error {
	pval, err := parityValue(parity)
	if err != nil {
		return err
	}
	sval, err := stopValue(stopBits)
	if err != nil {
		return err
	}
	d.mode = &serial.Mode{
		BaudRate: baudRate,
		Parity:   pval,
		DataBits: dataBits,
		StopBits: sval,
	}
	return d.updateMode()
}

func (d *SerialPortDevice) SetParity(parity string) error {
	sval, err := parityValue(parity)
	if err != nil {
		return err
	}
	d.mode.Parity = sval
	return d.updateMode()
}

func (d *SerialPortDevice) SetStopBits(stopBits string) error {
	sval, err := stopValue(stopBits)
	if err != nil {
		return err
	}
	d.mode.StopBits = sval
	return d.updateMode()
}

func (d *SerialPortDevice) SetDataBits(dataBits int) error {
	d.mode.DataBits = dataBits
	return d.updateMode()
}

func (d *SerialPortDevice) SetBaudRate(baudRate int) error {
	if baudRate == 19200 {
		baudRate = 9600
	}
	d.mode.BaudRate = baudRate
	return d.updateMode()
}

func (d *SerialPortDevice) updateMode() error {
	return d.port.SetMode(d.mode)
}

func (d *SerialPortDevice) CanSend() bool {
	bits, err := d.port.GetModemStatusBits()
	if err != nil {
		return true
	}
	return bits.DSR && !d.xOff
}

func (d *SerialPortDevice) GetStatusBits() (cts bool, dsr bool, dcd bool, ri bool) {
	bits, err := d.port.GetModemStatusBits()
	if err != nil {
		return false, false, false, false
	}
	return bits.CTS, bits.DSR, bits.DCD, bits.RI
}

func (d *SerialPortDevice) HasInput() bool {
	return len(d.RecvData) > 0
}

func (d *SerialPortDevice) Recv() int {

	if !d.HasInput() {
		return 0
	}

	v := <-d.RecvData

	//fmt.Printf("<- 0x%.2X\n", v)

	return int(v)
}

func (d *SerialPortDevice) Write(data []byte) {
	for _, b := range data {
		d.SendData <- b
	}
}

func (d *SerialPortDevice) SendOutputByte(value int) {

	//log.Printf("RawSerial: -> 0x%.2X\n", value)

	d.Write([]byte{byte(value)})
	//d.port.Write([]byte{byte(value)})

}

func (d *SerialPortDevice) InputAvailable() bool {
	return d.HasInput()
}

func (d *SerialPortDevice) GetInputByte() int {
	return d.Recv()
}

func (d *SerialPortDevice) IsConnected() bool {
	return d.port != nil
}

func (d *SerialPortDevice) Stop() {
	for len(d.RecvData) > 0 {
		<-d.RecvData
	}
	for len(d.SendData) > 0 {
		<-d.SendData
	}
	if d.port != nil {
		d.r = false
		d.port.Close()
	}
}

func (d *SerialPortDevice) Start() {

	if d.r {
		return
	}

	d.r = true

	go func(r io.Reader) {
		buff := make([]byte, 1)

		for d.r {
			n, err := r.Read(buff)
			if err != nil {
				d.r = false
				return
			}
			if n > 0 {
				for _, v := range buff[0:n] {
					if d.allowSoftwareFlow {
						if v == 19 {
							d.xOff = true
							fmt.Println("SerialPort: XOFF")
							continue
						} else if v == 17 {
							d.xOff = false
							fmt.Println("SerialPort: XON")
							continue
						}
					}

					d.RecvData <- v
				}
			}
		}

	}(d.port)

	go func(w io.Writer) {
		for d.r {
			select {
			case b := <-d.SendData:
				//log.Printf("Writing byte")
				_, err := w.Write([]byte{b})
				if err != nil {
					d.r = false
					return
				}
			}
		}

	}(d.port)

}

func EnumerateSerialPorts() ([]string, error) {
	ports, err := serial.GetPortsList()
	return ports, err
}
