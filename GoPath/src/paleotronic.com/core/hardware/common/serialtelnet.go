package common

import (
	"io"
	"net"
	"time"
)

type SerialTelnetDevice struct {
	host, port string
	conn       net.Conn
	fromConn   chan byte
	toConn     chan byte
	r          bool
}

func (d *SerialTelnetDevice) IsConnected() bool {
	return d.conn != nil && d.r
}

func (d *SerialTelnetDevice) InputAvailable() bool {
	return len(d.fromConn) > 0
}

func (d *SerialTelnetDevice) GetInputByte() int {

	// if !d.InputAvailable() {
	// 	return 0
	// }

	v := <-d.fromConn

	//log.Printf("GetInputByte <- 0x%.2X\n", v)

	return int(v)
}

func (d *SerialTelnetDevice) CanSend() bool {
	return d.IsConnected()
}

func (d *SerialTelnetDevice) SendOutputByte(value int) {

	//fmt.Printf("-> 0x%.2X\n", value)

	d.toConn <- byte(value)
	if value == 0xff {
		d.toConn <- byte(value) // repeat for proper encoding
	}

}

func (d *SerialTelnetDevice) Stop() {
	if d.conn != nil {
		d.r = false
	}
}

func (d *SerialTelnetDevice) Start() {

	d.Stop()
	// tcpAddr, err := net.ResolveTCPAddr("tcp", d.host+":"+d.port)
	// if err != nil {
	// 	return
	// }
	var err error
	d.conn, err = net.DialTimeout("tcp", d.host+":"+d.port, 3*time.Second)
	if err != nil {
		return
	}
	d.r = true

	go func(r io.Reader) {

		var commandMode bool

		buff := make([]byte, 1)

		for d.r {
			n, err := r.Read(buff)
			if err != nil {
				d.fromConn <- byte('E')
				d.fromConn <- byte('O')
				d.fromConn <- byte('F')
				d.fromConn <- byte('\r')
				d.fromConn <- byte('\n')
				d.r = false
				return
			}
			if n > 0 {
				for _, v := range buff[0:n] {
					if commandMode {
						switch v {
						case 0xff:
							d.fromConn <- v
							commandMode = false
						default:
							commandMode = false
						}
					} else {
						switch v {
						case 0xff:
							commandMode = true
						default:
							d.fromConn <- v
						}
					}
				}
			}
		}

	}(d.conn)

	go func(w io.Writer) {

		for d.r {
			select {
			case b := <-d.toConn:
				_, err := w.Write([]byte{b})
				if err != nil {
					d.fromConn <- byte('E')
					d.fromConn <- byte('O')
					d.fromConn <- byte('F')
					d.fromConn <- byte('\r')
					d.fromConn <- byte('\n')
					d.r = false
					return
				}
			}
		}

	}(d.conn)

}

func NewSerialTelnetDevice(host, port string) *SerialTelnetDevice {
	if port == "" {
		port = "23"
	}
	if host == "" {
		host = "localhost"
	}
	c := &SerialTelnetDevice{
		host:     host,
		port:     port,
		fromConn: make(chan byte, 24576),
		toConn:   make(chan byte, 1024),
		r:        false,
	}
	c.Start()
	return c

}
