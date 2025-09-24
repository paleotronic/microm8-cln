package fastserv

import "paleotronic.com/log"
import "net"
import "time"
import "paleotronic.com/fmt"
import "strings"
import "crypto/md5"
import "bytes"
import "paleotronic.com/encoding/ffpak"

type FSPayloadType byte

const (
	FS_MEMSYNC_REQUEST FSPayloadType = 0x40 + iota
	FS_MEMSYNC_RESPONSE
	FS_BULKMEM
	FS_CLIENT
	FS_CLIENTMEM
	FS_CLIENTAUDIO
	FS_ID_CLIENT
	FS_BYE_CLIENT
	FS_WELCOME_CLIENT
	FS_PING
	FS_PONG
	FS_BYE
	FS_ZZZ
	FS_REQUEST_TRANSFER_OWNERSHIP // tell server to xfer ownership to another connection
	FS_TRANSFER_OWNERSHIP_OK
	FS_ALLOCATE_CONTROL // modify control config.
	FS_ALLOCATE_CONTROL_OK
	FS_REMOTE_PARSE
	FS_REMOTE_PARSE_OK
	FS_REMOTE_EXEC
	FS_REMOTE_EXEC_OK
	FS_RESTALGIA_COMMAND
)

type HandlerMap map[FSPayloadType]func(client *Client, server *Server, msg []byte) error

type Server struct {
	ClientList    map[string]*Client
	InChannel     chan []byte
	Service       string
	MessageExpiry int64
	netListen     net.Listener
	Running       bool
	Mappings      HandlerMap
	OnConnect     func(name string)
	OnDisconnect  func()
	Owner         string // Current service owner
}

func (s *Server) HandleSingleClientMsg(client *Client, msg []byte) {
	handled := false

	if len(msg) == 0 {
		return
	}

	switch FSPayloadType(msg[0]) {
	case FS_PING:
		log.Printf("FS_PING from %s", client.Name)
		client.Incoming <- []byte{byte(FS_PONG)}
		client.Touch()
		return
	case FS_BYE:
		log.Printf("FS_BYE from %s", client.Name)
		client.Close()
		if s.OnDisconnect != nil {
			s.OnDisconnect()
		}
		return
	default:
		// default case lets check the handler
		handler, exists := s.Mappings[FSPayloadType(msg[0])]
		if exists {
			log.Printf("Delegating to defined handler for action [%x]\n", msg[0])
			handler(client, s, msg)
			handled = true
		} else {
			// No defined message of this type...
		}
	}
	//log.Printf("ClientReader: From connected client: %v\n", msg)
	if !handled {
		log.Printf("Passing on not handled message code %x\n", msg[0])
		client.Outgoing <- msg
	}
}

func (s *Server) ClientSpooler(client *Client) {

	if s.OnConnect != nil {
		s.OnConnect(client.Name)
		s.CheckOwnership()
	}

	for client.Connected {

		var didsomething bool

		// anything to read from network?
		ok, data := client.ConnectedAndTalking()
		if ok {
			for _, msg := range data {
				s.HandleSingleClientMsg(client, msg)
				didsomething = true
			}
		} else {
			client.Connected = false
			break
		}

		// anything to send out...
		for len(client.Incoming) > 0 {
			msg := <-client.Incoming
			sbuffer := ffpak.FFPack(msg)
			sbuffer = append(sbuffer, []byte{13, 10}...)
			_, e := client.Conn.Write(sbuffer)
			if e != nil {
				client.Connected = false
				fmt.Printf("Failed sending bytes to client\n")
			} else {
				fmt.Printf("Send bytes client\n")
				didsomething = true
			}
		}

		if !didsomething {
			time.Sleep(250 * time.Microsecond)
		}

	}

	client.Close()

	if s.OnDisconnect != nil {
		s.OnDisconnect()
		s.CheckOwnership()
	}

}

func (s *Server) Run() {

	s.ClientList = make(map[string]*Client)
	s.InChannel = make(chan []byte, 1048576)
	go s.IOHandler(s.InChannel, s.ClientList)
	//go s.IdleHandler(s.InChannel, s.ClientList)

	service := s.Service

	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)

	if err != nil {
		log.Fatalf("Could not resolve service: [%s]: %s\n", service, err.Error())
	} else {
		var err error
		s.netListen, err = net.Listen(tcpAddr.Network(), tcpAddr.String())
		if err != nil {
			log.Fatalf("Could not start service [%s]: %s\n", service, err.Error())
		} else {
			s.Running = true

			defer s.netListen.Close()

			for s.Running {
				log.Println("Waiting for TCP connections...")
				connection, err := s.netListen.Accept()
				if err != nil {
					log.Printf("Client error: %s\n", err.Error())
				} else {
					go s.ClientHandler(connection, s.InChannel, s.ClientList)
				}
			}
		}
	}

}

func (s *Server) GetUsers() []string {
	out := make([]string, 0)
	for _, client := range s.ClientList {
		out = append(out, client.Name)
	}
	return out
}

// CheckOwnership checks that the current server owner is still valid...
// If not it will allocate a new owner
func (s *Server) CheckOwnership() {

	// 1. Does current owner still exist?
	var found bool
	var c *Client = &Client{}
	for _, client := range s.ClientList {
		if client.Name == s.Owner {
			found = true
			break
		}
		if c.Name == "" {
			c = client // get first user in connection order just in case we need to change it
		}
	}

	// if not found, allocate
	if !found {
		s.Owner = c.Name
		fmt.Printf("-> Allocated ownership to %s as there was no valid owner...\n", s.Owner)
	}

}

// IsOwner returns true if the specified client is the owner...
func (s *Server) IsOwner(c *Client) bool {
	return (s.Owner == c.Name)
}

func (s *Server) TransferOwnership(c *Client, target string) bool {

	if c.Name != s.Owner {
		return false
	}

	// Ok 'c' is the owner, they can transfer to a valid connection
	for _, client := range s.ClientList {
		if client.Name == target {
			s.Owner = target
			fmt.Printf("-> Transferred ownership from %s to %s...\n", c.Name, target)
			return true
		}
	}

	// if we got here we didn't match a known connection...
	return false
}

func (s *Server) Find(target string) *Client {

	// Ok 'c' is the owner, they can transfer to a valid connection
	for _, client := range s.ClientList {
		if client.Name == target {
			return client
		}
	}

	// if we got here we didn't match a known connection...
	return nil
}

func (s *Server) ClientHandler(conn net.Conn, ch chan []byte, clientList map[string]*Client) {
	buffer := make([]byte, 0)

	chunk, buffer, error := ReadLineWithTimeout(conn, 30000, buffer, []byte{13, 10})

	if error != nil {
		log.Println("Client connection error: ", error)
	}

	// Expect name
	str := string(chunk)

	log.Printf("Client sent: [%s]\n", str)

	var name string

	if len(str) < 2 {
		_ = conn.Close()
		return
	}

	if FSPayloadType(str[0]) == FS_ID_CLIENT {
		name = str[1:]
	} else {
		return
	}

	// Create unique token here
	tt := md5.Sum([]byte(name + conn.RemoteAddr().String() + conn.LocalAddr().String()))

	newClient := &Client{
		Name:       name,
		Incoming:   make(chan []byte, 10000),
		Outgoing:   ch,
		Conn:       conn,
		Quit:       make(chan bool),
		RecvBuffer: buffer, // just in case the client is still sending something
		ClientList: clientList,
		Connected:  true,
		Token:      tt,
	}

	newClient.LastActive = time.Now()

	go s.ClientSpooler(newClient)

	// Add client to be tracked for message forwarding
	clientList[newClient.Name] = newClient

	// Send an initial message
	newClient.Incoming <- []byte{byte(FS_WELCOME_CLIENT)}
}

// UTILS
func ReadLineWithTimeout(conn net.Conn, ms time.Duration, buffer []byte, sep []byte) ([]byte, []byte, error) {

	// rules... see if we have a line in the buffer supplied already
	// if not read more with timeout
	// if yes return line

	idx := bytes.Index(buffer, sep)

	if idx > -1 {
		// line in existing buffer
		str := buffer[0:idx]           // up to pos of sep
		buffer = buffer[idx+len(sep):] // chop off sep as well
		return str, buffer, nil        // OK
	} else {
		// no line in existing buffer - try to read more with timeout

		if ms == 0 {
			ms = 1000 // 1 seconds default
		}

		tmp := make([]byte, 2048)
		conn.SetReadDeadline(time.Now().Add(50 * time.Millisecond)) // read timeout for this call so we can vary as needed
		bytesRead, err := conn.Read(tmp)
		if (err != nil) && (!strings.Contains(err.Error(), "timeout")) {
			// something bad happened pass it on
			return []byte(nil), buffer, err
		}
		if bytesRead > 0 {
			// got more data, let's munge it and try get a chunk of bytes again
			buffer = append(buffer, tmp[0:bytesRead]...)
			idx = bytes.Index(buffer, sep)
			if idx > -1 {
				// line in existing buffer
				str := buffer[0:idx]           // up to pos of sep
				buffer = buffer[idx+len(sep):] // chop off sep as well
				return str, buffer, nil
			}
		}
	}

	// remaining case - nothing read
	return []byte(nil), buffer, nil
}

func (s *Server) IOHandler(Incoming <-chan []byte, clientList map[string]*Client) {
	for {
		log.Println("IOHandler: Waiting for input")
		input := <-Incoming
		for _, client := range clientList {
			err := client.SendMessage(input)
			if err != nil {
				fmt.Println("Dropping client due to " + err.Error())
				client.Close()
				s.CheckOwnership()
			}
		}
	}
}

func (s *Server) Stop() {
	s.Running = false
	time.Sleep(50 * time.Millisecond)
}

func NewServer(port string, mapping HandlerMap) *Server {
	this := &Server{}
	this.Service = port
	this.Mappings = mapping
	return this
}
