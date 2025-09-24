package client

import (
	"errors"
	"net"
	"strings"
	"time"

	"paleotronic.com/ducktape"
	"paleotronic.com/fmt"
	"paleotronic.com/log"
)

/* Duck tape client interface */

type LocalConnector interface {
	ConnectMe(dct *DuckTapeClient)
}

type DuckTapeClient struct {
	Host             string
	Service          string
	Name             string
	SendChannel      string
	Conn             net.Conn
	UConn            net.UDPConn
	Proto            string
	RecvBuffer       []byte
	Incoming         chan *ducktape.DuckTapeBundle
	Outgoing         chan *ducktape.DuckTapeBundle
	Quit             chan bool
	Subscription     map[string]int64
	Connected        bool
	LocalSend        chan *ducktape.DuckTapeBundle
	LocalRecv        chan *ducktape.DuckTapeBundle
	LastMesg         time.Time
	LastHeartbeat    time.Time
	Token            [16]byte
	OK               bool
	RECV_OK, SEND_OK bool
	Pending          []byte
	NeedReconnect    bool
	RIndex           int
	DeadTime         time.Time
	CustomHandler    map[string]func(c *DuckTapeClient, msg *ducktape.DuckTapeBundle)
	Distributor      map[string]chan *ducktape.DuckTapeBundle
}

func NewDuckTapeClient(host, service string, name string, proto string) *DuckTapeClient {

	this := &DuckTapeClient{}
	this.Host = host
	this.Name = name
	this.Service = service
	this.RecvBuffer = make([]byte, 0)
	this.Incoming = make(chan *ducktape.DuckTapeBundle, 10000)
	this.Outgoing = make(chan *ducktape.DuckTapeBundle, 10000)
	this.LocalSend = make(chan *ducktape.DuckTapeBundle, 100)
	this.LocalRecv = make(chan *ducktape.DuckTapeBundle, 100)
	this.Quit = make(chan bool)
	this.Subscription = make(map[string]int64)
	this.Proto = proto
	this.CustomHandler = make(map[string]func(c *DuckTapeClient, msg *ducktape.DuckTapeBundle))
	this.Distributor = map[string]chan *ducktape.DuckTapeBundle{}

	return this
}

func (c *DuckTapeClient) Connect() error {
	return c.ConnectTCP()
}

func (c *DuckTapeClient) ConnectSingle() error {
	c.NeedReconnect = false
	return c.ConnectTCPSingle()
}

// Connect and process
func (c *DuckTapeClient) ConnectTCP() error {
	conn, err := net.DialTimeout("tcp", c.Host+c.Service, 10*time.Second)

	if err != nil {
		log.Printf("Connect failed to: %s - %s", c.Host+c.Service, err.Error())
		return err
	}

	c.OK = true

	// Connected
	c.Conn = conn

	log.Println("Starting read/writers and monitor")
	// start senders
	c.OK = true
	//~ go c.ClientReader()
	//~ go c.ClientSender()

	go c.ClientSpooler()

	log.Println("Sending QCK and waiting")

	// send connect message
	c.Outgoing <- &ducktape.DuckTapeBundle{
		ID:      "QCK",
		Payload: []byte(c.Name),
	}

	cutoff := time.Now().Add(5 * time.Second)

	for !c.Connected && time.Now().Before(cutoff) {
		//log.Println("Awaiting QCK response")
		time.Sleep(50 * time.Millisecond)
	}

	if c.Connected {
		return nil
	}

	return errors.New("Timeout")
}

func (c *DuckTapeClient) ConnectTCPSingle() error {
	conn, err := net.Dial("tcp", c.Host+c.Service)

	c.OK = true

	if err != nil {
		return err
	}

	// Connected
	c.Conn = conn

	// start senders
	//~ go c.ClientReader()
	//~ go c.ClientSender()
	//~ go c.Monitor()

	// send connect message
	c.Outgoing <- &ducktape.DuckTapeBundle{
		ID:      "QCK",
		Payload: []byte(c.Name),
	}

	for !c.Connected {
		c.ClientSenderSingle()
		c.ClientReaderSingle()
		time.Sleep(5 * time.Millisecond)
	}

	return nil
}

func (c *DuckTapeClient) ConnectedAndTalking() (bool, []*ducktape.DuckTapeBundle) {
	var hasData []*ducktape.DuckTapeBundle

	chunk, remaining, error := ducktape.ReadLineWithTimeout(c.Conn, 5000, c.RecvBuffer, []byte{13, 10})
	c.RecvBuffer = remaining
	if error != nil {
		log.Printf("ReadLineWithTimeout() -> %v %s:%s\n", error, c.Host, c.Service)
		c.OK = false
		return false, hasData
	}

	// got chunk... try unbundle into a message
	var msg *ducktape.DuckTapeBundle = &ducktape.DuckTapeBundle{}
	err := msg.UnmarshalBinary(chunk)
	if err == nil {
		hasData = append(hasData, msg)
	}

	return true, hasData
}

func (client *DuckTapeClient) Reconnect() error {

	log.Println("Waiting for RECV to shutdown...")
	for client.RECV_OK {
		time.Sleep(50 * time.Millisecond)
	}
	log.Println("Waiting for SEND to shutdown...")
	for client.SEND_OK {
		time.Sleep(50 * time.Millisecond)
	}
	client.Conn.Close()

	e := client.ConnectTCP()
	if e == nil {
		// resubscribe
		for s, _ := range client.Subscription {
			_ = client.SubscribeChannel(s)
		}
		client.SendMessage("SND", []byte(client.SendChannel), false)
		//		if len(client.Pending) > 0 {
		//			client.Conn.Write(client.Pending)
		//			client.Pending = []byte(nil)
		//		}
	}

	return e
}

func (client *DuckTapeClient) sendBytes() error {
	return nil
}

func (client *DuckTapeClient) ClientSender() {
	for client.OK {
		select {
		case msg := <-client.Outgoing:
			sbuffer, _ := msg.MarshalBinary()
			log.Printf("%s: ClientSender sending %v to %v\n", client.Name, msg.ID, client.Conn.RemoteAddr())
			//log.Println("Send size: ", len(sbuffer))
			client.Conn.SetWriteDeadline(time.Now().Add(500 * time.Millisecond))
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
			client.OK = false
			log.Println("Client ", client.Name, " quitting")
			client.Conn.Close()
			break
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}
	log.Println("SENDER EXITING")
}

func (client *DuckTapeClient) ClientSenderSingle() bool {
	//client.SEND_OK = true
	//for client.OK {
	var didsomething bool

	for len(client.Outgoing) > 0 {
		didsomething = true
		select {
		case msg := <-client.Outgoing:
			if msg.ID == "PIG" {
				log.Println("Sending heartbeat")
			}
			sbuffer, _ := msg.MarshalBinary()
			//log.Printf("%s: ClientSender sending %v\n", client.Name, msg.String())
			//log.Println("Send size: ", len(sbuffer))
			//client.Conn.SetWriteDeadline(time.Now().Add(500 * time.Millisecond))

			// Use WriteLineWithTimeout
			sbuffer, e := ducktape.WriteLineWithTimeout(client.Conn, 100*time.Millisecond, 1, sbuffer)

			//_, e := client.Conn.Write(sbuffer)
			if e != nil {
				client.OK = false
				client.SEND_OK = false
				client.Pending = sbuffer
				client.Connected = false
				client.NeedReconnect = true
				log.Printf("Reconnecting to the server due to %s\n", e.Error())
				return false
			}
		case <-client.Quit:
			log.Println("Client ", client.Name, " quitting")
			client.Conn.Close()
			break
		default:
		}

	}

	return didsomething
}

func (client *DuckTapeClient) HandleMsg(msg *ducktape.DuckTapeBundle) {
	log.Println(client.Name+": RAW", msg.String())
	//client.LastMesg = time.Now()
	switch msg.ID {
	case "CLS":
		client.Connected = false
		client.Close()
		return
	case "POG":
		client.Connected = true
	case "ZZZ":
		client.SendMessage("PIG", []byte("."), false)
	case "WCM":
		client.Connected = true
		// Store session key
		if len(msg.Payload) == 16 {
			for i, b := range msg.Payload {
				client.Token[i] = b
			}
		}
		//handled = true
	case "RCV":
		str := string(msg.Payload)
		list := strings.Split(str, ",")
		for _, item := range list {
			_ = client.SubscribeChannel(item)
		}
		//log.Println(client.Name + ": receiving " + str)
		//handled = true
	case "IGN":
		str := string(msg.Payload)
		list := strings.Split(str, ",")
		for _, item := range list {
			_ = client.UnsubscribeChannel(item)
		}
		//handled = true
	default:
		if dchan, ok := client.Distributor[msg.ID]; ok {
			dchan <- msg
			return
		}
		// default case lets check the handler
		//////fmt.Println("ooooo "+msg.ID)
		if h, ok := client.CustomHandler[msg.ID]; ok {
			h(client, msg)
		} else {
			client.Incoming <- msg
		}
	}
}

// Handles receiving and queueing messages
func (client *DuckTapeClient) ClientReader() {

	client.LastMesg = time.Now()

	//	buffer := make([]byte, ducktape.MaxBufferSize)
	var ok bool
	var data []*ducktape.DuckTapeBundle
	ok, data = client.ConnectedAndTalking()

	for ok && client.OK {
		if len(data) > 0 {
			//log.Printf("Got messages: %v\n", data)
		}
		for _, msg := range data {
			//handled := false
			log.Printf("Msg [%s] from server\n", msg.ID)
			client.HandleMsg(msg)
			client.LastMesg = time.Now()
		}

		if time.Since(client.LastMesg) > 15*time.Second {
			// send something
			//client.LastMesg = time.Now()
			client.Outgoing <- &ducktape.DuckTapeBundle{
				ID:      "PIG",
				Payload: []byte("."),
			}
		}

		if len(data) == 0 {
			//log.Println("Nothing coming back yet")
			time.Sleep(1 * time.Millisecond)
		}

		ok, data = client.ConnectedAndTalking()

	}

	client.OK = false

	log.Println("READER EXITING")
}

func (client *DuckTapeClient) ClientReaderSingle() bool {

	var didsomething bool

	//	buffer := make([]byte, ducktape.MaxBufferSize)
	var ok bool
	var data []*ducktape.DuckTapeBundle
	ok, data = client.ConnectedAndTalking()
	if ok {
		didsomething = true
		if len(data) > 0 {
			//log.Printf("Got messages: %v\n", data)
		}
		for _, msg := range data {
			//handled := false
			client.LastMesg = time.Now()
			client.HandleMsg(msg)
		}

		if time.Since(client.LastMesg) > 15*time.Second && time.Since(client.LastHeartbeat) > 15*time.Second {
			// send something
			client.LastHeartbeat = time.Now()
			client.Outgoing <- &ducktape.DuckTapeBundle{
				ID:      "PIG",
				Payload: []byte("."),
			}
		}
	} else {
		client.Connected = false
		client.NeedReconnect = true
	}

	return didsomething
}

func (c *DuckTapeClient) Reestablish() {
	err := c.Reconnect()
	if err != nil {
		log.Printf("Trying to reconnect...")
		time.AfterFunc(15*time.Second, c.Reestablish)
		return
	}
	log.Println("SUCCESS!")
}

func (c *DuckTapeClient) ClientSpooler() {
	log.Printf("Client spooler started\n")
	// swdur := 10 * time.Second
	// c.DeadTime = time.Now().Add(swdur)
	c.LastMesg = time.Now()
	for c.OK {
		sent := c.ClientSenderSingle()
		recv := c.ClientReaderSingle()

		if !sent && !recv {
			time.Sleep(250 * time.Millisecond)
		}
		if time.Since(c.LastMesg) > 20*time.Second {
			c.OK = false
		}
	}

	//if c.NeedReconnect {
	//	c.Reestablish()
	//}

	log.Printf("Client spooler ended\n")
}

func (c *DuckTapeClient) Close() {
	c.Connected = false
	time.Sleep(1 * time.Second)

	if c.Proto == "udp" {
		c.UConn.Close()
	} else {
		c.Conn.Close()
	}
}

func (c *DuckTapeClient) HasSubscription(channel string) bool {
	_, ex := c.Subscription[channel]
	return ex
}

func (c *DuckTapeClient) SubscribeChannel(channel string) error {

	_, ex := c.Subscription[channel]
	if ex {
		return errors.New("Already subscribed")
	}

	c.Subscription[channel] = 0 // no msgs recved
	return nil

}

func (c *DuckTapeClient) SendMessageAndCatchResponses(
	id string,
	payload []byte,
	binary bool,
	valid string,
	err string,
	timeout time.Duration,
) (*ducktape.DuckTapeBundle, error) {
	temp := make(map[string]chan *ducktape.DuckTapeBundle)
	if ch, ok := c.Distributor[valid]; ok {
		temp[valid] = ch
	}
	c.Distributor[valid] = make(chan *ducktape.DuckTapeBundle)
	if ch, ok := c.Distributor[err]; ok {
		temp[err] = ch
	}
	c.Distributor[err] = make(chan *ducktape.DuckTapeBundle)
	defer func() {
		// for k, v := range temp {
		// 	c.Distributor[k] = v
		// }
		delete(c.Distributor, valid)
		delete(c.Distributor, err)
	}()

	fmt.Println("before send " + id)
	c.SendMessage(id, payload, binary)
	fmt.Println("after send " + id)

	t := time.After(timeout)
	select {
	case <-t:
		fmt.Println("timeout for " + id)
		return nil, errors.New("timeout")
	case msg := <-c.Distributor[valid]:
		fmt.Println("get valid response " + valid)
		return msg, nil
	case msg := <-c.Distributor[err]:
		fmt.Println("got invalid response " + err)
		return msg, errors.New(string(msg.Payload))
	}
}

func (c *DuckTapeClient) SendMessage(id string, payload []byte, binary bool) {

	var tries int
	for !c.OK && tries < 10 {
		log.Println("Trying to send a message but client needs reconnection...")
		log.Println("Draining channels")
		for len(c.Outgoing) > 0 {
			_ = <-c.Outgoing
		}
		for len(c.Incoming) > 0 {
			_ = <-c.Incoming
		}
		e := c.ConnectTCP()
		if e != nil {
			//return
			time.Sleep(10 * time.Second)
			tries++
		} else {
			log.Println("Client reconnected OK")
			time.Sleep(100 * time.Millisecond)
			break
		}
	}

	if !c.OK {
		return
	}

	c.Outgoing <- &ducktape.DuckTapeBundle{
		ID:      id,
		Payload: payload,
		Binary:  binary,
	}
}

func (c *DuckTapeClient) UnsubscribeChannel(channel string) error {

	_, ex := c.Subscription[channel]

	if ex {
		delete(c.Subscription, channel)
		return nil
	}

	return errors.New("Not subscribed")

}
