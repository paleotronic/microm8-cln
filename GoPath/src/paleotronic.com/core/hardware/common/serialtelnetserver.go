package common

import (
	"io"
	"net"
	"strings"
	"sync"
	"time"

	tomb "gopkg.in/tomb.v2"
	"paleotronic.com/core/settings"
	"paleotronic.com/octalyzer/errorreport"
)

type SerialTelnetServer struct {
	host, port string
	conn       net.Conn
	fromConn   chan byte
	toConn     chan byte
	l          net.Listener
	r          bool
	t          *tomb.Tomb
	slot       int
}

func (d *SerialTelnetServer) IsConnected() bool {
	return d.conn != nil && d.r
}

func (d *SerialTelnetServer) CanSend() bool {
	return d.IsConnected()
}

func (d *SerialTelnetServer) InputAvailable() bool {
	return len(d.fromConn) > 0
}

func (d *SerialTelnetServer) GetInputByte() int {

	if !d.InputAvailable() {
		return 0
	}

	v := <-d.fromConn

	//fmt.Printf("<- 0x%.2X\n", v)

	return int(v)
}

func (d *SerialTelnetServer) SendOutputByte(value int) {

	//fmt.Printf("-> 0x%.2X\n", value)

	d.toConn <- byte(value)

}

func (d *SerialTelnetServer) Stop() {
	if d.t != nil {
		d.t.Kill(nil)
	}
	if d.conn != nil {
		d.conn.Close()
		d.r = false
	}
	if d.l != nil {
		d.l.Close()
	}
}

func (d *SerialTelnetServer) Listen() {

	sermutex[d.slot].Lock()
	defer sermutex[d.slot].Unlock()

	l, err := net.Listen("tcp", d.host+":"+d.port)
	if err != nil {
		errorreport.GracefulErrorReport("Failed to start telnet server on port "+d.port+".", err)
	}
	d.l = l
	defer l.Close()

	d.r = true
	for d.r {
		conn, err := l.Accept()
		if err != nil && strings.Contains(err.Error(), "use of closed network connection") {
			return
		}
		if err != nil {
			errorreport.GracefulErrorReport("Failed to accept incoming telnet connection on port "+d.port+".", err)
		}
		if d.conn != nil {
			d.conn.Close()
		}
		if d.t != nil {
			d.t.Kill(nil)
		}
		d.conn = conn
		d.t = &tomb.Tomb{}
		d.t.Go(func() error {

			var r io.Reader = d.conn

			buff := make([]byte, 8)

			for d.r {
				n, err := r.Read(buff)
				if err != nil {
					d.fromConn <- byte('E')
					d.fromConn <- byte('O')
					d.fromConn <- byte('F')
					d.fromConn <- byte('\r')
					d.fromConn <- byte('\n')
					//d.r = false
					return nil
				}
				if n > 0 {
					for _, v := range buff[0:n] {
						d.fromConn <- v
					}
				} else {
					time.Sleep(25 * time.Millisecond)
				}
				if len(d.t.Dying()) > 0 {
					<-d.t.Dying()
					return nil
				}
			}

			return nil

		})
		d.t.Go(func() error {

			var w io.Writer = d.conn

			for d.r {
				select {
				case <-d.t.Dying():
					return nil
				case b := <-d.toConn:
					_, err := w.Write([]byte{b})
					if err != nil {
						d.fromConn <- byte('E')
						d.fromConn <- byte('O')
						d.fromConn <- byte('F')
						d.fromConn <- byte('\r')
						d.fromConn <- byte('\n')
						//d.r = false
						return nil
					}
				}
			}

			return nil

		})
	}
}

var sermutex [settings.NUMSLOTS]sync.Mutex

func NewSerialTelnetServer(slot int, host, port string) *SerialTelnetServer {
	if port == "" {
		port = "23"
	}
	if host == "" {
		host = "localhost"
	}
	c := &SerialTelnetServer{
		host:     host,
		port:     port,
		fromConn: make(chan byte, 24576),
		toConn:   make(chan byte, 1024),
		r:        false,
		slot:     slot,
	}
	go c.Listen()
	return c

}
