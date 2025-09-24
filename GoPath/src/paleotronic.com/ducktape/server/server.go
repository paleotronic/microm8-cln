package server

import (
	//"os"
	//	"paleotronic.com/fmt"
	"crypto/md5"
	"strings"
	"sync"
	"time"

	"paleotronic.com/ducktape"
	"paleotronic.com/ducktape/client"
	"paleotronic.com/panic"
)
import "container/list"
import "net"
import "paleotronic.com/log"

type DuckHandler func(c *ducktape.Client, s *DuckTapeServer, msg *ducktape.DuckTapeBundle) error

type DuckHandlerMap map[string]DuckHandler

var LocalServer *DuckTapeServer

type DuckTapeServer struct {
	ClientList    *list.List
	UClientList   map[string]*ducktape.Client
	InChannel     chan *ducktape.DuckTapeBundle
	LocalConnect  chan *client.DuckTapeClient
	Service       string
	Mappings      DuckHandlerMap
	MessageExpiry int64
	netListen     net.Listener
	Running       bool
	idm, msgm     sync.Mutex
	idlist        map[string]int64
	//messagestore  map[string][]ducktape.DuckTapeBundle
	OnDisconnect       func(s *DuckTapeServer, count int, ocount int)
	OnClientConnect    func(c *ducktape.Client)
	OnClientDisconnect func(c *ducktape.Client)
}

func NewDuckTapeServer(port string, mapping DuckHandlerMap) *DuckTapeServer {
	this := &DuckTapeServer{}
	this.Service = port
	this.Mappings = mapping
	LocalServer = this
	return this
}

func (s *DuckTapeServer) FindClient(name string) *ducktape.Client {

	for e := s.ClientList.Front(); e != nil; e = e.Next() {
		client := e.Value.(ducktape.Client)
		if client.Name == name {
			return &client
		}
	}

	return nil

}

func (s *DuckTapeServer) GetUUIDForChannel(channel string) int64 {
	s.idm.Lock()
	defer s.idm.Unlock()

	// at this point we are exclusive
	if s.idlist == nil {
		s.idlist = make(map[string]int64)
	}

	v, ok := s.idlist[channel]
	if !ok {
		v = 0
	}
	v = v + 1
	s.idlist[channel] = v
	return v
}

func (s *DuckTapeServer) HandleSingleClientMsg(client *ducktape.Client, msg *ducktape.DuckTapeBundle) {
	handled := false

	log.Printf("Msg = %s", msg.ID)

	switch msg.ID {
	case "PIG":
		client.Incoming <- &ducktape.DuckTapeBundle{
			ID:      "POG",
			Payload: []byte("."),
		}
		client.Touch()
	case "BYE":
		client.Close("Remote user requested shutdown")
		return
	case "SND":
		str := string(msg.Payload)
		client.CurrentChannel = str
		client.Incoming <- &ducktape.DuckTapeBundle{
			ID:      "BRD",
			Payload: []byte(str),
		}
		handled = true
	case "SUB":
		str := string(msg.Payload)
		list := strings.Split(str, ",")
		for _, item := range list {
			log.Printf("======================================= SUBSCRIBE CHANNEL = %s", item)
			suberr := client.SubscribeChannel(item)
			if suberr != nil {
				client.Incoming <- &ducktape.DuckTapeBundle{
					ID:      "ERR",
					Payload: []byte(suberr.Error()),
				}
			} else {
				client.Incoming <- &ducktape.DuckTapeBundle{
					ID:      "RCV",
					Payload: []byte(item),
				}

			}
		}
		ncount := s.ClientList.Len()
		if s.OnDisconnect != nil {
			s.OnDisconnect(s, ncount, ncount)
		}
		handled = true
	case "USB":
		str := string(msg.Payload)
		list := strings.Split(str, ",")
		for _, item := range list {
			log.Printf("-------------------------------------- UNSUBSCRIBE CHANNEL = %s", item)
			suberr := client.UnsubscribeChannel(item)
			if suberr != nil {
				client.Incoming <- &ducktape.DuckTapeBundle{
					ID:      "ERR",
					Payload: []byte(suberr.Error()),
				}
			} else {
				client.Incoming <- &ducktape.DuckTapeBundle{
					ID:      "IGN",
					Payload: []byte(item),
				}
			}
		}
		ncount := s.ClientList.Len()
		if s.OnDisconnect != nil {
			s.OnDisconnect(s, ncount, ncount)
		}
		handled = true
	default:
		// default case lets check the handler
		handler, exists := s.Mappings[msg.ID]
		if exists {
			now := time.Now()
			log.Printf("Delegating to defined handler for action [%s]\n", msg.ID)
			handler(client, s, msg)
			snc := time.Since(now)
			log.Printf("Delegated request completed in %v", snc)
			handled = true
		} else {
			// No defined message of this type...
		}
	}
	//log.Printf("ClientReader: From connected client: %v\n", msg)
	if !handled {
		msg.Channel = client.CurrentChannel
		log.Printf("Passing on not handled message %v\n", msg)
		client.Outgoing <- msg
	}
}

func (s *DuckTapeServer) ClientReader(client *ducktape.Client) {

	//	buffer := make([]byte, ducktape.MaxBufferSize)
	ok, data := client.ConnectedAndTalking()
	for ok {
		for _, msg := range data {
			s.HandleSingleClientMsg(client, msg)
		}
		time.Sleep(1 * time.Millisecond)
		ok, data = client.ConnectedAndTalking()
	}

	client.Outgoing <- &ducktape.DuckTapeBundle{
		ID:      "CLS",
		Payload: []byte("Connection ended"),
	}

	//fmt.Println("ClientReader stopped for ", client.Name)
	client.RemoveMe()
}

func (s *DuckTapeServer) ClientSender(client *ducktape.Client) {
	for {
		select {
		case msg := <-client.Incoming:
			sbuffer, _ := msg.MarshalBinary()
			log.Printf("ClientSender sending %s %v to %s\n", msg.ID, sbuffer, client.Name)
			log.Println("Send size: ", len(sbuffer))
			_, e := client.Conn.Write(sbuffer)
			log.Println("Send response:", e)
			if e != nil {
				// client conn failed...
				log.Printf("*** Client [%v] has left us...\n", client.Conn.RemoteAddr())
				client.Close("Connection has been dropped")
			}
		case <-client.Quit:
			log.Println("Client ", client.Name, " quitting")
			client.Close("Client quit")
			break
		default:
			time.Sleep(1 * time.Millisecond)
		}
	}
}

func (s *DuckTapeServer) Close(client *ducktape.Client, reason string) {
	client.Close(reason)
}

func (s *DuckTapeServer) ClientSpooler(client *ducktape.Client) {

	panic.Do(

		func() {

			unit := 300 * time.Second
			client.DeadTime = time.Now().Add(unit)
			log.Printf("Dead time @ %v\n", client.DeadTime)

			log.Printf("Starting Client Spooler for %s\n", client.Conn.RemoteAddr().String())

			client.Connected = true

			if s.OnClientConnect != nil {
				s.OnClientConnect(client)
			}

			for client.Connected {

				var didsomething bool

				// anything to read from network?
				ok, data := client.ConnectedAndTalking()
				if ok {
					for _, msg := range data {
						log.Printf("Mesg = %s\n", msg.ID)
						s.HandleSingleClientMsg(client, msg)
						didsomething = true
					}
				} else {
					log.Println("Connected and talking failed")
					client.Connected = false
					break
				}

				// anything to send out...
				for len(client.Incoming) > 0 {
					msg := <-client.Incoming
					sbuffer, _ := msg.MarshalBinary()
					sbuffer = append(sbuffer, []byte{13, 10}...)
					// client.Conn.SetWriteDeadline(time.Now().Add(15000 * time.Millisecond))
					// _, e := client.Conn.Write(sbuffer)

					sbuffer, e := ducktape.WriteLineWithTimeout(client.Conn, 5000*time.Millisecond, 5, sbuffer)

					if e != nil {
						log.Println("Send failed", e)
						client.Connected = false
						break
					} else {
						didsomething = true
					}
				}

				if !didsomething {
					time.Sleep(250 * time.Microsecond)
					if time.Since(client.DeadTime) > 0 {
						client.SendMessageEx("CLS", []byte(nil), false)
						break
					}
				} else {
					client.DeadTime = time.Now().Add(unit)
					log.Printf("Dead time @ %v\n", client.DeadTime)
				}

			}

			go client.Close("Spooler informed not connected anymore")

			if s.OnClientDisconnect != nil {
				s.OnClientDisconnect(client)
			}

			log.Printf("Ending Client Spooler for %s\n", client.Conn.RemoteAddr().String())

		},
		func(r interface{}) {
			log.Printf("Client service connection %v has encountered an error (%s) and is closing", client, r)
			client.SendMessage(
				&ducktape.DuckTapeBundle{
					ID:      "ERR",
					Payload: []byte("internal server error - a bug has been logged and we are on it"),
					Binary:  false,
				},
			)
			//client.Close("Server error")
			go s.ClientSpooler(client)
		},
	)

}

func (s *DuckTapeServer) ClientHandler(conn net.Conn, ch chan *ducktape.DuckTapeBundle, clientList *list.List) {
	buffer := make([]byte, 0)

	var str string
	var tries int = 0
	var err error
	var chunk []byte

	for str == "" && tries < 10 {
		chunk, buffer, err = ducktape.ReadLineWithTimeout(conn, 30000, buffer, []byte{13, 10})

		if err != nil {
			log.Println("Client connection error: ", err)
			conn.Close()
			return
		}

		if len(chunk) > 0 {
			str = string(chunk)
		} else {
			time.Sleep(1 * time.Millisecond)
			tries++
		}
	}

	log.Printf("Client sent: [%s]\n", str)

	var name string

	if len(str) < 5 {
		log.Println("Closing due to short name")
		_ = conn.Close()
		return
	}

	if str[0:4] == "QCK " {
		name = str[4:]
	} else {
		return
	}

	// Create unique token here
	tt := md5.Sum([]byte(name + conn.RemoteAddr().String() + conn.LocalAddr().String()))

	newClient := &ducktape.Client{
		Name:         name,
		Incoming:     make(chan *ducktape.DuckTapeBundle, 10000),
		Outgoing:     ch,
		Conn:         conn,
		Quit:         make(chan bool),
		RecvBuffer:   buffer, // just in case the client is still sending something
		ClientList:   clientList,
		Subscription: make(map[string]int64),
		Token:        tt,
		DeadTime:     time.Now().Add(time.Second * 60),
	}

	newClient.LastActive = time.Now()

	// Start client send/receiver go-routines
	//go s.ClientSender(newClient)
	//go s.ClientReader(newClient)

	go s.ClientSpooler(newClient)

	// Add client to be tracked for message forwarding
	clientList.PushBack(*newClient)

	// Send an initial message - note: message with ClientName and no channel is protocol message
	log.Println("Sending WCM message")
	newClient.Incoming <- &ducktape.DuckTapeBundle{
		ID:      "WCM",
		Payload: newClient.Token[0:],
		Binary:  true,
	}
}

func (s *DuckTapeServer) IOHandler(Incoming <-chan *ducktape.DuckTapeBundle, clientList *list.List) {
	for {
		log.Println("IOHandler: Waiting for input")
		input := <-Incoming
		//log.Println("IOHandler: Handling ", input)

		//		if input.Channel != "" {
		//			s.QueueMessage(input)
		//		}

		for e := clientList.Front(); e != nil; e = e.Next() {
			client := e.Value.(ducktape.Client)
			if (input.Channel != "") && (client.HasSubscription(input.Channel)) {
				log.Printf("Client %s has a subscription to %s, passing message %v\n", client.Name, input.Channel, input)
				client.Incoming <- input
			}
		}
	}
}

func (s *DuckTapeServer) IdleHandler(Incoming <-chan *ducktape.DuckTapeBundle, clientList *list.List) {

	count := s.ClientList.Len()
	tick := 0

	for {

		ncount := s.ClientList.Len()
		if ncount != count {
			if s.OnDisconnect != nil {
				s.OnDisconnect(s, ncount, count)
			}
			count = ncount
		}

		tick++
		if tick%1000 == 0 {

			for e := clientList.Front(); e != nil; e = e.Next() {
				client := e.Value.(ducktape.Client)

				if client.IsIdle() {
					client.Incoming <- &ducktape.DuckTapeBundle{
						ID:      "ZZZ",
						Payload: []byte("."),
						Binary:  false,
					}
				}

			}

		}

		// now sleep
		time.Sleep(time.Millisecond * 50)
	}
}

func (s *DuckTapeServer) Run() {

	s.ClientList = list.New()
	s.InChannel = make(chan *ducktape.DuckTapeBundle, 1048576)
	go s.IOHandler(s.InChannel, s.ClientList)
	go s.IdleHandler(s.InChannel, s.ClientList)
	//go s.CatchupHandler(s.ClientList)
	//go s.CleanUpHandler(s.ClientList)

	service := s.Service

	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)

	////fmt.Printntln(tcpAddr.String())

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
					log.Printf("Connecting from %v", connection)
					go s.ClientHandler(connection, s.InChannel, s.ClientList)
				}
			}
		}
	}

}

func (s *DuckTapeServer) RunALL() {

	s.Run()

}

func (s *DuckTapeServer) ConnectMe(dtc *client.DuckTapeClient) {
	s.LocalConnect <- dtc
}

// ChannelCounts returns a mapping of channel name to active users
func (s *DuckTapeServer) ChannelCounts() map[string]int {

	clientList := s.ClientList

	m := make(map[string]int)

	for e := clientList.Front(); e != nil; e = e.Next() {
		client := e.Value.(ducktape.Client)
		for cn, _ := range client.Subscription {
			m[cn] = m[cn] + 1
		}
	}

	return m
}

// DirectSend allows the server to send a targeted message...
func (s *DuckTapeServer) DirectSend(ID string, payload []byte, binary bool, targetchannel string) {

	input := &ducktape.DuckTapeBundle{
		ID:      ID,
		Payload: payload,
		Binary:  binary,
		Channel: targetchannel,
	}

	if s.ClientList == nil {
		return
	}

	for e := s.ClientList.Front(); e != nil; e = e.Next() {
		client := e.Value.(ducktape.Client)
		log.Printf("??? Client %s, Channel %s, Subscribed? %v", client.Name, input.Channel, client.HasSubscription(input.Channel))
		if (input.Channel != "") && (client.HasSubscription(input.Channel)) {
			log.Printf("Client %s has a subscription to %s, passing message %v\n", client.Name, input.Channel, input)
			client.Incoming <- input
		}
	}

}

func (s *DuckTapeServer) Stop() {
	s.Running = false
	s.netListen.Close()
}
