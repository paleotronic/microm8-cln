package fastserv

import (
	"bytes"
	"errors"
	"net"
	"time"

	"paleotronic.com/fmt"

	"paleotronic.com/encoding/ffpak"
	"paleotronic.com/log"
)

const IDLE_TIMEOUT = 60000

type Client struct {
	Name         string
	Conn         net.Conn
	Connected    bool
	RecvBuffer   []byte
	Incoming     chan []byte
	Outgoing     chan []byte
	Quit         chan bool
	Subscription map[string]int64
	ClientList   map[string]*Client
	Token        [16]byte
	LastActive   time.Time
	RIndex       int
}

func (c *Client) IsIdle() bool {

	return (time.Since(c.LastActive) > IDLE_TIMEOUT*time.Millisecond)

}

// Attempt to read data from the connection
func (c *Client) ConnectedAndTalking() (bool, [][]byte) {
	var hasData [][]byte

	chunk, remaining, error := ReadLineWithTimeout(c.Conn, 5000, c.RecvBuffer, []byte{13, 10})
	c.RecvBuffer = remaining
	if error != nil {
		log.Printf("ReadLineWithTimeout() -> %v\n", error)
		return false, hasData
	}

	// got chunk... try unbundle into a message
	hasData = append(hasData, ffpak.FFUnpack(chunk))

	c.Touch()

	return true, hasData
}

// Cleanly close client
func (c *Client) Close() {
	//c.Quit <- true
	c.Conn.Close()
	c.RemoveMe()
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
	delete(c.ClientList, c.Name)
}

func (c *Client) Touch() {
	c.LastActive = time.Now()
}

func (client *Client) SendMessage(msg []byte) error {
	//~ log.Printf("Send %v\n", msg)

	//~ sbuffer := ffpak.FFPack(msg)
	//~ sbuffer = append(sbuffer, []byte{13, 10}...)
	//~ log.Printf("ClientSender sending %x to %s\n", msg[0], client.Name)
	//~ log.Println("Send size: ", len(sbuffer))
	//~ _, e := client.Conn.Write(sbuffer)
	//~ log.Println("Send response:", e)
	//~ if e != nil {
	//~ // client conn failed...
	//~ //fmt.Printf("*** Client [%v] has left us...\n", client.Conn.RemoteAddr())
	//~ client.Close()
	//~ return e
	//~ }
	if len(client.Incoming) > 9000 {
		return errors.New("buffer full")
	}

	fmt.Println("msg on client queue")

	client.Incoming <- msg

	return nil
}
