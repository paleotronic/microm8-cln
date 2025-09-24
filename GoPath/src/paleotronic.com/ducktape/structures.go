package ducktape

//import "paleotronic.com/fmt"
import (
	"bytes"
	"container/list" //	"paleotronic.com/fmt"
	"net"
	"time"

	"paleotronic.com/log"
)

//import "bytes"

/* DuckTape is a TCP/IP protocol layer for exchanging messages between
   clients and a server

   Packet format: -

   Offset	Size	Description
   ------	----	-----------
	0		8		'DuCkTaPe'
	8		3		24 bit data size (N)
	11		N		N bytes of data

*/

const (
	MaxBufferSize = 2048
	DUCKMAGIC     = "DuCkTaPe"
	IDLE_TIMEOUT  = 60000
)

type Client struct {
	Name           string
	CurrentChannel string
	Conn           net.Conn
	UConn          *net.UDPConn
	Proto          string // "udp" for udp connections
	RecvBuffer     []byte
	Incoming       chan *DuckTapeBundle
	Outgoing       chan *DuckTapeBundle
	Quit           chan bool
	UDPAddr        *net.UDPAddr
	Subscription   map[string]int64
	ClientList     *list.List
	Token          [16]byte
	LastActive     time.Time
	Connected      bool
	DeadTime       time.Time
	RIndex         int
}

func (c *Client) SubscribeChannel(channel string) error {

	log.Printf("Subscribing to channel %s", channel)

	_, ex := c.Subscription[channel]
	if ex {
		return nil
	}

	c.Subscription[channel] = 0 // no msgs recved
	return nil

}

func (c *Client) IsIdle() bool {

	return (time.Since(c.LastActive) > IDLE_TIMEOUT*time.Millisecond)

}

func (c *Client) UnsubscribeChannel(channel string) error {

	log.Printf("Unsubscribing from channel %s", channel)

	_, ex := c.Subscription[channel]

	if ex {
		delete(c.Subscription, channel)
		return nil
	}

	return nil

}

// Attempt to read data from the connection
func (c *Client) ConnectedAndTalking() (bool, []*DuckTapeBundle) {
	var hasData []*DuckTapeBundle

	chunk, remaining, error := ReadLineWithTimeout(c.Conn, 5000*time.Millisecond, c.RecvBuffer, []byte{13, 10})
	c.RecvBuffer = remaining
	if error != nil {
		log.Printf("ReadLineWithTimeout() -> %v\n", error)
		return false, hasData
	}

	// got chunk... try unbundle into a message
	var msg *DuckTapeBundle = &DuckTapeBundle{}
	err := msg.UnmarshalBinary(chunk)
	if err == nil {
		hasData = append(hasData, msg)
	}

	c.Touch()

	return true, hasData
}

// Cleanly close client
func (c *Client) Close(reason string) {
	c.Connected = false
	time.Sleep(1 * time.Second)
	c.Conn.Close()
	c.RemoveMe()
}

func (c *Client) IsDead() bool {
	return time.Since(c.DeadTime) > 0
}

// Test if clients are equal based on name and TCP connection
func (c *Client) Equal(other *Client) bool {
	if bytes.Equal([]byte(c.Name), []byte(other.Name)) {
		if c.Conn == other.Conn {
			return true
		}
	}
	return false
}

// Remove this client from list of clients
func (c *Client) RemoveMe() {
	for entry := c.ClientList.Front(); entry != nil; entry = entry.Next() {
		client := entry.Value.(Client)
		if c.Equal(&client) {
			c.ClientList.Remove(entry)

		}
	}
	//	//fmt.Printf("There are now %d users connected.\n", c.ClientList.Len())
}

func (c *Client) Touch() {
	//	//fmt.Printf("Client %s (%v) is alive, got a PIG from them\n", c.Name, c.Conn.RemoteAddr())
	c.LastActive = time.Now()
}

func (c *Client) HasSubscription(channel string) bool {
	_, ex := c.Subscription[channel]
	return ex
}

func (c *Client) SendMessage(msg *DuckTapeBundle) {
	log.Printf("Send %v\n", msg)
	c.Incoming <- msg
}

func (c *Client) SendMessageEx(ID string, payload []byte, binary bool) {

	msg := &DuckTapeBundle{
		ID:      ID,
		Payload: payload,
		Binary:  binary,
	}

	log.Printf("Send %v\n", msg)
	c.Incoming <- msg
}
