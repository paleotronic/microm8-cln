package s8webclient

import (
	"paleotronic.com/utils"
	//	"strings"
	//	"paleotronic.com/fmt"
	"errors"
	"time"

	"paleotronic.com/filerecord"
)

// StoreUserFile saves a user file into the database backend
func (c *Client) StoreUserFile(filepath string, filename string, data []byte) error {

	fullpath := filepath + "/" + filename

	var err error

	if c.c == nil {
		return errors.New("Not connected")
	}

	// Now do the connection
	r := []byte(c.Session + fullpath + ":")
	r = append(r, data...)
	c.c.SendMessage("USV", r, true)

	// get response
	select {
	case msg := <-c.c.Incoming:
		////fmt.Printf("in _StoreUserFile() %s\n", msg.ID)
		if msg.ID == "SOK" {
			// Login OK
			err = nil
		} else if msg.ID == "ERR" {
			err = errors.New("registration failed")
		}
	}

	return err
}

// FetchUserFile retrieves a user file.
func (c *Client) FetchUserFile(filepath string, filename string) (filerecord.FileRecord, error) {

	fullpath := filepath + "/" + filename

	var err error

	if c.c == nil {
		return filerecord.FileRecord{}, errors.New("Not connected")
	}

	// Now do the connection
	r := []byte(c.Session + fullpath)
	c.c.SendMessage("ULD", r, true)

	// get response
	tochan := time.After(time.Second * 20)
	bb := &filerecord.FileRecord{}
	select {
	case _ = <-tochan:
		err = errors.New("timeout")
	case msg := <-c.c.Incoming:
		////fmt.Printf("in _StoreUserFile() %s\n", msg.ID)
		if msg.ID == "FIL" {
			// Login OK
			err = nil
			bb.UnJSON(msg.Payload)
		} else if msg.ID == "ERR" {
			err = errors.New("registration failed")
		}
	}

	return *bb, err
}

// FetchUserDir retrieves a user dir
func (c *Client) FetchUserDir(filepath string, filespec string) ([]byte, error) {
	fullpath := filepath

	var err error

	if c.c == nil {
		return []byte(nil), errors.New("Not connected")
	}

	// Now do the connection
	r := []byte(c.Session + fullpath + ":" + filespec)
	c.c.SendMessage("UDR", r, true)

	// get response
	tochan := time.After(time.Second * 20)
	var bb []byte
	select {
	case _ = <-tochan:
		err = errors.New("timeout")
	case msg := <-c.c.Incoming:
		////fmt.Printf("in _StoreUserFile() %s\n", msg.ID)
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

func (c *Client) AppendSystemFile(filepath, filename string, data []byte) error {
	return c.AppendCustomFile("ASF", "", filepath, filename, data)
}

func (c *Client) AppendUserFile(filepath, filename string, data []byte) error {
	return c.AppendCustomFile("AUF", "", filepath, filename, data)
}

func (c *Client) AppendProjectFile(project, filepath, filename string, data []byte) error {
	return c.AppendCustomFile("APF", project, filepath, filename, data)
}

func (c *Client) AppendLegacyFile(filepath, filename string, data []byte) error {
	return c.AppendCustomFile("ALF", "", filepath, filename, data)
}

// AppendCustomFile saves a user file into the database backend
func (c *Client) AppendCustomFile(req string, additional string, filepath string, filename string, data []byte) error {

	fullpath := filepath + "/" + filename

	var err error

	if c.c == nil {
		return errors.New("Not connected")
	}

	// Now do the connection
	r := []byte(c.Session + fullpath + ":")
	r = append(r, []byte(additional+":")...)
	r = append(r, data...)
	c.c.SendMessage(req, r, true)

	// get response
	tochan := time.After(time.Second * 20)
	select {
	case _ = <-tochan:
		err = errors.New("timeout")
	case msg := <-c.c.Incoming:
		//				//fmt.Printf("in AppendCustomFile() %s\n", msg.ID)
		if msg.ID == "AOK" {
			// Login OK
			err = nil
		} else if msg.ID == "ERR" {
			err = errors.New(string(msg.Payload))
		}
	}

	return err
}

func (c *Client) CustomLoadAddress(req string, additional string, filepath string, filename string, address int) error {

	fullpath := filepath + "/" + filename

	var err error

	if c.c == nil {
		return errors.New("Not connected")
	}

	sa := utils.IntToStr(address)

	// Now do the connection
	r := []byte(c.Session + fullpath + ":")
	r = append(r, []byte(additional+":")...)
	r = append(r, []byte(sa)...)
	c.c.SendMessage(req, r, true)

	// get response
	tochan := time.After(time.Second * 20)
	select {
	case _ = <-tochan:
		err = errors.New("timeout")
	case msg := <-c.c.Incoming:
		//				//fmt.Printf("in AppendCustomFile() %s\n", msg.ID)
		if msg.ID == "SOK" {
			// Login OK
			err = nil
		} else if msg.ID == "ERR" {
			err = errors.New(string(msg.Payload))
		}
	}

	return err
}

func (c *Client) SetUserLoadAddress(p, f string, address int) error {
	return c.CustomLoadAddress("USA", "", p, f, address)
}

func (c *Client) SetSystemLoadAddress(p, f string, address int) error {
	return c.CustomLoadAddress("SSA", "", p, f, address)
}

func (c *Client) SetLegacyLoadAddress(p, f string, address int) error {
	return c.CustomLoadAddress("LSA", "", p, f, address)
}

func (c *Client) SetProjectLoadAddress(project, p, f string, address int) error {
	return c.CustomLoadAddress("PSA", project, p, f, address)
}
