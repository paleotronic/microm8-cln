package s8webclient

import (
	//	"strings"
	//	"paleotronic.com/fmt"
	"errors"
	"time"

	"paleotronic.com/ducktape/client"
	"paleotronic.com/filerecord"
)

// StoreSystemFile saves a system file into the database backend
func (c *Client) StoreSystemFile(filepath string, filename string, data []byte) error {

	fullpath := filepath + "/" + filename

	var err error

	if c.c == nil {
		c.Username = "system"
		c.c = client.NewDuckTapeClient(c.Hostname, c.Port, c.Username, "tcp")
		err := c.c.Connect()
		if err != nil {
			////fmt.Println("Failed connect")
			return err
		}
		//		//fmt.Println("Connected")
	}

	// Now do the connection
	r := []byte(c.Session + fullpath + ":")
	r = append(r, data...)
	c.c.SendMessage("SSV", r, true)

	// get response
	select {
	case msg := <-c.c.Incoming:
		////fmt.Printf("in _StoreSystemFile() %s\n", msg.ID)
		if msg.ID == "SOK" {
			// Login OK
			err = nil
		} else if msg.ID == "ERR" {
			err = errors.New("registration failed")
		}
	}

	return err
}

// FetchSystemFile retrieves a system file.
func (c *Client) FetchSystemFile(filepath string, filename string) (filerecord.FileRecord, error) {

	if len(filepath) > 1 && rune(filepath[0]) == '/' {
		filepath = filepath[1:]
	}

	////fmt.Printf("FetchSystemFile: path = %s, filename = %s\n", filepath, filename)

	fullpath := filepath + "/" + filename

	var err error

	////fmt.Printf("c = %v, c.c == %v\n", c, c.c)

	if c.c == nil {
		c.Username = "system"
		c.c = client.NewDuckTapeClient(c.Hostname, c.Port, c.Username, "tcp")
		err := c.c.Connect()
		if err != nil {
			////fmt.Println("Failed connect")
			return filerecord.FileRecord{}, err
		}
		//		//fmt.Println("Connected")
	}

	if c.Session == "" {
		c.Session = "12345678123456781234567812345678"
	}

	// Now do the connection
	r := []byte(c.Session + fullpath)
	c.c.SendMessage("SLD", r, true)

	// get response
	tochan := time.After(time.Second * 20)
	bb := &filerecord.FileRecord{}
	select {
	case _ = <-tochan:
		err = errors.New("timeout")
	case msg := <-c.c.Incoming:
		////fmt.Printf("in FetchSystemFile() %s - %s\n", msg.ID, )
		if msg.ID == "FIL" {
			// Login OK
			err = nil
			bb.UnJSON(msg.Payload)
		} else if msg.ID == "ERR" {
			err = errors.New("registration failed")
			////fmt.Println("Failed - "+string(msg.Payload))
		}
	}

	return *bb, err
}

// FetchSystemDir retrieves a system dir.
func (c *Client) FetchSystemDir(filepath string, filespec string) ([]byte, error) {
	fullpath := filepath

	var err error

	if c.c == nil {
		c.c = client.NewDuckTapeClient(c.Hostname, c.Port, c.Username, "tcp")
		err := c.c.Connect()
		if err != nil {
			return []byte(nil), err
		}
	}
	// Now do the connection
	r := []byte(c.Session + fullpath + ":" + filespec)
	c.c.SendMessage("SDR", r, true)

	// get response
	tochan := time.After(time.Second * 20)
	var bb []byte
	select {
	case _ = <-tochan:
		err = errors.New("timeout")
	case msg := <-c.c.Incoming:
		////fmt.Printf("in _StoreUserFile() %s\n", msg.ID)
		////fmt.Printf("in FetchSystemDir() %s\n%s\n", msg.ID, string(msg.Payload))
		if msg.ID == "DIR" {
			// Login OK
			err = nil
			bb = msg.Payload
		} else if msg.ID == "ERR" {
			err = errors.New("registration failed")
		}
	}

	return bb, err
}
