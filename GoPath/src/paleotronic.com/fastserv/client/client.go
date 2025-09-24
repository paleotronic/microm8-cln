package client

import (
	"errors"
	//"paleotronic.com/fmt"
	"net"
	"time"

	"paleotronic.com/log"

	"paleotronic.com/fastserv"

	"paleotronic.com/encoding/ffpak"
)

/* Duck tape client interface */

type LocalConnector interface {
	ConnectMe(dct *FSClient)
}

const MAX_TRIES = 10

type ClientState int

const (
	CS_NEW ClientState = iota
	CS_CREATE_SOCKET
	CS_WAIT_RETRY
	CS_SOCKET_FAILURE
	CS_HANDSHAKE_SEND
	CS_HANDSHAKE_WELCOME
	CS_CONNECTED
	CS_ERROR
	CS_SHUTDOWN
	CS_DISCONNECTED
	CS_PROTOCOL_ERROR
)

func (cs ClientState) String() string {
	switch cs {
	case CS_NEW:
		return "New Connection"
	case CS_CREATE_SOCKET:
		return "Creating Socket"
	case CS_WAIT_RETRY:
		return "Waiting for retry"
	case CS_SOCKET_FAILURE:
		return "Socket failure"
	case CS_HANDSHAKE_SEND:
		return "Sending handshake"
	case CS_HANDSHAKE_WELCOME:
		return "Waiting for welcome"
	case CS_CONNECTED:
		return "Connected"
	case CS_ERROR:
		return "Error"
	case CS_SHUTDOWN:
		return "Shutdown requested"
	case CS_DISCONNECTED:
		return "Disconnected by request"
	case CS_PROTOCOL_ERROR:
		return "Protocol error"
	}
	return "Unknown"
}

type FSClient struct {
	Host             string
	Service          string
	Name             string
	Conn             net.Conn
	RecvBuffer       []byte
	Incoming         chan []byte
	Outgoing         chan []byte
	Quit             chan bool
	ReConnect        chan bool
	Connected        bool
	LastMesg         int64
	Token            [16]byte
	OK               bool
	Tries            int
	Done             bool
	RECV_OK, SEND_OK bool
	OnConnect        func()
	Pending          []byte
	NeedReconnect    bool
	State            ClientState
	RIndex           int
	CustomHandler    map[fastserv.FSPayloadType]func(c *FSClient, msg []byte)
}

func NewFSClient(host, service string, name string, proto string) *FSClient {

	this := &FSClient{}
	this.Host = host
	this.Name = name
	this.Service = service
	this.RecvBuffer = make([]byte, 0)
	this.Incoming = make(chan []byte, 10000)
	this.Outgoing = make(chan []byte, 10000)
	this.Quit = make(chan bool)
	this.ReConnect = make(chan bool)
	this.CustomHandler = make(map[fastserv.FSPayloadType]func(c *FSClient, msg []byte))

	this.State = CS_NEW

	return this
}

// Connect and process
func (c *FSClient) Connect() error {
	conn, err := net.Dial("tcp", c.Host+c.Service)

	c.OK = true

	if err != nil {
		log.Printf("Connect failed to: %s - %s", c.Host+c.Service, err.Error())
		return err
	}

	// Connected
	c.Conn = conn

	// send connect message
	chunk := []byte{byte(fastserv.FS_ID_CLIENT)}
	chunk = append(chunk, []byte(c.Name)...)

	c.Outgoing <- chunk

	for !c.Connected && c.Do() {
		time.Sleep(1 * time.Millisecond)
	}

	c.Tries = 0 // reset

	if !c.Connected {
		return errors.New("Connection failed")
	}

	return nil
}

func (c *FSClient) ConnectedAndTalking() (bool, [][]byte) {
	var hasData [][]byte

	chunk, remaining, error := fastserv.ReadLineWithTimeout(c.Conn, 5000, c.RecvBuffer, []byte{13, 10})
	c.RecvBuffer = remaining
	if error != nil {
		log.Printf("ReadLineWithTimeout() -> %v %s:%s\n", error, c.Host, c.Service)
		c.OK = false
		return false, hasData
	}

	// got chunk... try unbundle into a message
	//~ var msg []byte = []byte(nil)
	//~ err := msg.UnmarshalBinary(chunk)
	//if err == nil {
	hasData = append(hasData, ffpak.FFUnpack(chunk))
	//}

	return true, hasData
}

func (client *FSClient) ClientSender() {
	client.SEND_OK = true
	for client.OK {
		select {
		case msg := <-client.Outgoing:
			sbuffer := ffpak.FFPack(msg)
			sbuffer = append(sbuffer, 13, 10)
			//log.Printf("%s: ClientSender sending %v to %v\n", client.Name, msg.ID, client.Conn.RemoteAddr())
			//log.Println("Send size: ", len(sbuffer))
			_, e := client.Conn.Write(sbuffer)
			if e != nil {
				client.OK = false
				client.SEND_OK = false
				client.Pending = sbuffer
				log.Println("SENDER EXITING", e.Error())
				log.Printf("Reconnecting to the server due to %s\n", e.Error())
				return
				log.Printf("Reconnecting to the server due to %s\n", e.Error())
			}
		case <-client.Quit:
			log.Println("Client ", client.Name, " quitting")
			client.Conn.Close()
			break
		}
	}
	client.SEND_OK = false
	log.Println("SENDER EXITING")
}

// Do() acts as a run loop of sorts...
func (c *FSClient) Do() bool {

	//fmt.Printf("FSClient (%s%s): %s\n", c.Host, c.Service, c.State.String())

	switch c.State {

	case CS_NEW:
		// Brand new connection
		c.State = CS_CREATE_SOCKET
		c.Tries = 0

	case CS_CREATE_SOCKET:

		c.Connected = false
		c.RecvBuffer = make([]byte, 0)
		c.Incoming = make(chan []byte, 10000)
		c.Outgoing = make(chan []byte, 10000)

		// close existing socket if needed
		if c.Conn != nil {
			c.Conn.Close()
			c.Conn = nil
		}
		// establish a tcp connection
		conn, err := net.Dial("tcp", c.Host+c.Service)

		if err != nil {
			c.Tries++
			if c.Tries < MAX_TRIES {
				c.State = CS_WAIT_RETRY //
			} else {
				c.State = CS_SOCKET_FAILURE
			}
		} else {

			// do id
			c.Conn = conn
			c.State = CS_HANDSHAKE_SEND

		}

	case CS_HANDSHAKE_SEND:
		// send our id hand shake
		msg := []byte{byte(fastserv.FS_ID_CLIENT)}
		msg = append(msg, []byte(c.Name)...)

		sbuffer := ffpak.FFPack(msg)
		sbuffer = append(sbuffer, 13, 10)
		//log.Printf("%s: ClientSender sending %v to %v\n", client.Name, msg.ID, client.Conn.RemoteAddr())
		//log.Println("Send size: ", len(sbuffer))
		_, e := c.Conn.Write(sbuffer)
		if e != nil {
			c.State = CS_WAIT_RETRY
		} else {
			c.State = CS_HANDSHAKE_WELCOME
		}

	case CS_HANDSHAKE_WELCOME:
		// recv welcome
		chunk, remaining, error := fastserv.ReadLineWithTimeout(c.Conn, 5000, c.RecvBuffer, []byte{13, 10})
		c.RecvBuffer = remaining
		if error != nil {
			c.Tries++
			if c.Tries < MAX_TRIES {
				c.State = CS_WAIT_RETRY //
			} else {
				c.State = CS_SOCKET_FAILURE
			}
		} else if len(chunk) > 0 {
			// read something
			if fastserv.FSPayloadType(chunk[0]) == fastserv.FS_WELCOME_CLIENT {
				c.State = CS_CONNECTED
				c.Connected = true

				// We call our call back if its defined
				if c.OnConnect != nil {
					c.OnConnect()
				}

			} else {
				c.State = CS_PROTOCOL_ERROR
			}
		}

	case CS_PROTOCOL_ERROR:
		c.State = CS_DISCONNECTED

	case CS_SOCKET_FAILURE:
		c.State = CS_DISCONNECTED

	case CS_WAIT_RETRY:
		time.Sleep(500 * time.Millisecond)
		c.State = CS_CREATE_SOCKET

	case CS_CONNECTED:
		// recv
		var ok bool
		var data [][]byte
		ok, data = c.ConnectedAndTalking()

		if ok {
			for _, msg := range data {
				//handled := false

				c.HandleMsg(msg)
				c.LastMesg = time.Now().UnixNano()
			}
		} else {

			// read failed, try reconnect
			c.State = CS_WAIT_RETRY
			return true

		}

		// if we are here lets send anything we have
		for len(c.Outgoing) > 0 {

			msg := <-c.Outgoing
			sbuffer := ffpak.FFPack(msg)
			sbuffer = append(sbuffer, 13, 10)
			_, e := c.Conn.Write(sbuffer)
			if e != nil {
				c.State = CS_WAIT_RETRY
				return true
			}

		}

	case CS_SHUTDOWN:
		c.Conn.Close()
		c.Conn = nil
		c.State = CS_DISCONNECTED

	case CS_DISCONNECTED:
		c.Connected = false
		return false

	}

	return true
}

func (client *FSClient) HandleMsg(msg []byte) {

	if len(msg) == 0 {
		return
	}

	log.Printf("Client received binary message %v", msg)

	//log.Println(client.Name+": RAW", msg.String())
	switch fastserv.FSPayloadType(msg[0]) {
	case fastserv.FS_BYE:
		client.Connected = false
		client.Close()
		return
	case fastserv.FS_PONG:
		client.Connected = true
	case fastserv.FS_ZZZ:
		client.SendMessage(fastserv.FS_PING, []byte{0})
	case fastserv.FS_WELCOME_CLIENT:
		client.Connected = true
		//handled = true
	default:
		// default case lets check the handler
		//////fmt.Println("ooooo "+msg.ID)
		if h, ok := client.CustomHandler[fastserv.FSPayloadType(msg[0])]; ok {
			h(client, msg)
		} else {
			client.Incoming <- msg
		}
	}
}

// Handles receiving and queueing messages
func (client *FSClient) ClientReader() {

	client.RECV_OK = true

	// init
	client.LastMesg = time.Now().UnixNano()

	//	buffer := make([]byte, ducktape.MaxBufferSize)
	var ok bool
	var data [][]byte
	ok, data = client.ConnectedAndTalking()

	for ok && client.OK {
		if len(data) > 0 {
			//log.Printf("Got messages: %v\n", data)
		}
		for _, msg := range data {
			//handled := false

			client.HandleMsg(msg)
			client.LastMesg = time.Now().UnixNano()
		}

		if (time.Now().UnixNano()-client.LastMesg)/1000000 > 60000 {
			// send something
			client.Outgoing <- []byte{
				byte(fastserv.FS_PING),
			}
		}

		if len(data) == 0 {
			//log.Println("Nothing coming back yet")
			time.Sleep(1 * time.Millisecond)
		}

		ok, data = client.ConnectedAndTalking()
	}

	client.SendMessage(fastserv.FS_BYE_CLIENT, []byte{0})
	client.RECV_OK = false

	log.Println("READER EXITING")
}

func (c *FSClient) Close() {
	c.State = CS_SHUTDOWN
	//c.Conn.Close()
}

func (c *FSClient) SendMessage(id fastserv.FSPayloadType, payload []byte) {
	if c.State != CS_CONNECTED {
		return
	}

	chunk := []byte{byte(id)}
	chunk = append(chunk, payload...)
	c.Outgoing <- chunk
}
