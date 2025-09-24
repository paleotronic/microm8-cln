package s8webclient

import (
	"errors"
	"time"

	"paleotronic.com/ducktape"
)

// SendAndWait sends a network request with a timeout
func (c *Client) SendAndWait(id string, payload []byte, valid []string) (*ducktape.DuckTapeBundle, error) {

	//	var err error

	if c.c == nil {
		return &ducktape.DuckTapeBundle{}, errors.New("Not connected")
	}

	// Now do the connection
	c.c.SendMessage(id, payload, true)

	// get response
	tochan := time.After(time.Second * 20)
	//var bb []byte
	select {
	case _ = <-tochan:
		return &ducktape.DuckTapeBundle{}, errors.New("timeout error")
	case msg := <-c.c.Incoming:

		for _, i := range valid {
			if i == msg.ID {
				return msg, nil
			}
		}

		return msg, errors.New("Unexpected message")
	}

	//return &ducktape.DuckTapeBundle{}, errors.New("Unexpected message")
}
