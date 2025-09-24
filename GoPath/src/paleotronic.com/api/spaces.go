package s8webclient

import (
	//	"paleotronic.com/fmt"
	"strings"

	//	"strings"
	//	"paleotronic.com/fmt"
	"errors"
	"time"

	"paleotronic.com/filerecord"
)

// StoreRemIntFile saves a remote file into the database backend
func (c *Client) StoreRemIntFile(project string, filepath string, filename string, data []byte) error {

	//log.Printf("StoreProjectFile() called %s")

	fullpath := filepath + "/" + filename

	var err error

	if c.c == nil {
		return errors.New("Not connected")
	}

	// Now do the connection
	r := []byte(c.Session + fullpath + ":" + project + ":")
	r = append(r, data...)
	c.c.SendMessage("ISV", r, true)

	// get response
	tochan := time.After(time.Second * 20)
	select {
	case _ = <-tochan:
		err = errors.New("timeout")
	case msg := <-c.c.Incoming:
		////fmt.Printf("in _StoreUserFile() %s\n", msg.ID)
		if msg.ID == "SOK" {
			// Login OK
			err = nil
		} else if msg.ID == "ERR" {
			err = errors.New(string(msg.Payload))
		}
	}

	return err
}

// StoreProjectFile saves a project file into the database backend
func (c *Client) StoreProjectFile(project string, filepath string, filename string, data []byte) error {

	//log.Printf("StoreProjectFile() called %s")

	fullpath := filepath + "/" + filename

	var err error

	if c.c == nil {
		return errors.New("Not connected")
	}

	// Now do the connection
	r := []byte(c.Session + fullpath + ":" + project + ":")
	r = append(r, data...)
	c.c.SendMessage("PSV", r, true)

	// get response
	tochan := time.After(time.Second * 20)
	select {
	case _ = <-tochan:
		err = errors.New("timeout")
	case msg := <-c.c.Incoming:
		////fmt.Printf("in _StoreUserFile() %s\n", msg.ID)
		if msg.ID == "SOK" {
			// Login OK
			err = nil
		} else if msg.ID == "ERR" {
			err = errors.New(string(msg.Payload))
		}
	}

	return err
}

// FetchRemIntFile fetches a remote file
func (c *Client) FetchRemIntFile(project string, filepath string, filename string) (filerecord.FileRecord, error) {

	fullpath := filepath + "/" + filename

	var err error

	if c.c == nil {
		return filerecord.FileRecord{}, errors.New("Not connected")
	}

	// Now do the connection
	r := []byte(c.Session + fullpath + ":" + project)
	c.c.SendMessage("ILD", r, true)

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

// FetchProjectFile fetches a project file
func (c *Client) FetchProjectFile(project string, filepath string, filename string) (filerecord.FileRecord, error) {

	fullpath := filepath + "/" + filename

	var err error

	if c.c == nil {
		return filerecord.FileRecord{}, errors.New("Not connected")
	}

	// Now do the connection
	r := []byte(c.Session + fullpath + ":" + project)
	c.c.SendMessage("PLD", r, true)

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

// FetchProjectDir retrieves a project dir
func (c *Client) FetchProjectDir(project string, filepath string, filespec string) ([]byte, error) {
	fullpath := filepath

	var err error

	if c.c == nil {
		return []byte(nil), errors.New("Not connected")
	}

	// Now do the connection
	r := []byte(c.Session + fullpath + ":" + filespec + ":" + project)
	c.c.SendMessage("PDR", r, true)

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

// FetchRemIntDir retrieves a remote dir
func (c *Client) FetchRemIntDir(project string, filepath string, filespec string) ([]byte, error) {
	fullpath := filepath

	var err error

	if c.c == nil {
		return []byte(nil), errors.New("Not connected")
	}

	// Now do the connection
	r := []byte(c.Session + fullpath + ":" + filespec + ":" + project)
	c.c.SendMessage("IDR", r, true)

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

// CreateRemIntDir creates a remote dir
func (c *Client) CreateRemIntDir(project string, filepath string, filename string) error {

	return c.CreateCustomDir("ICD", project, filepath, filename)

}

// CreateProjectDir creates a project dir
func (c *Client) CreateProjectDir(project string) error {

	var err error

	if c.c == nil {
		return errors.New("Not connected")
	}

	// Now do the connection
	r := []byte(c.Session + project)
	c.c.SendMessage("PCR", r, true)

	// get response
	tochan := time.After(time.Second * 20)
	select {
	case _ = <-tochan:
		err = errors.New("timeout")
	case msg := <-c.c.Incoming:
		////fmt.Printf("in _StoreUserFile() %s\n", msg.ID)
		if msg.ID == "SOK" {
			// Login OK
			err = nil
		} else if msg.ID == "ERR" {
			err = errors.New("Project create failed")
		}
	}

	return err
}

// ProjectStatus retrieves project status
func (c *Client) ProjectStatus(project string) (bool, bool, bool, error) {

	var err error

	if c.c == nil {
		return false, false, false, errors.New("Not connected")
	}

	// Now do the connection
	r := []byte(c.Session + project)
	c.c.SendMessage("PST", r, true)

	var ex bool
	var ow bool
	var ed bool

	// get response
	tochan := time.After(time.Second * 20)
	select {
	case _ = <-tochan:
		err = errors.New("timeout")
	case msg := <-c.c.Incoming:
		////fmt.Printf("in _StoreUserFile() %s\n", msg.ID)
		if msg.ID == "PSR" {
			// Login OK
			ex = (rune(msg.Payload[0]) == 'Y')
			ow = (rune(msg.Payload[1]) == 'Y')
			ed = (rune(msg.Payload[2]) == 'Y')

			err = nil
		} else if msg.ID == "ERR" {
			err = errors.New("Project status failed")
		}
	}

	return ex, ow, ed, err
}

// FetchProjectList retrieves a list of projects available to the user.
func (c *Client) FetchProjectList() ([]string, error) {

	var err error
	out := make([]string, 0)

	if c.c == nil {
		return out, errors.New("Not connected")
	}

	// Now do the connection
	r := []byte(c.Session)
	c.c.SendMessage("FPL", r, true)

	// get response
	tochan := time.After(time.Second * 20)
	select {
	case _ = <-tochan:
		err = errors.New("timeout")
	case msg := <-c.c.Incoming:
		//fmt.Printf("in _FetchProjectList() %s\n", msg.ID)
		if msg.ID == "APL" {
			// Login OK
			err = nil
			//fmt.Println(string(msg.Payload))
			out = strings.Split(string(msg.Payload), ":")
		} else if msg.ID == "ERR" {
			err = errors.New("Project fetch list failed")
		}
	}

	return out, err
}
