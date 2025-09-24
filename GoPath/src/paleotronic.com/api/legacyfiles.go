package s8webclient

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	//	"strings"
	//	"paleotronic.com/fmt"
	"errors"
	"strings"
	"time"

	"paleotronic.com/filerecord"
)

func (c *Client) httpGetBytes(url string) ([]byte, error) {

	client := &http.Client{}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode/100 != 2 {
		return nil, errors.New("Bad status")
	}
	return ioutil.ReadAll(resp.Body)

}

// FetchLegacyFile returns the file record
func (c *Client) FetchLegacyFile(filepath string, filename string) (filerecord.FileRecord, error) {

	var fr filerecord.FileRecord

	fullpath := filepath + "/" + filename

	base := "http://" + CONN.Hostname + ":6582/"

	metapath := base + "meta/legacy/" + strings.Trim(fullpath, "/")
	datapath := base + "files/legacy/" + strings.Trim(fullpath, "/")

	// log.Printf("Using meta path: %s", metapath)
	// log.Printf("Using file path: %s", datapath)

	metadata, err := c.httpGetBytes(metapath)
	if err != nil {
		return fr, err
	}
	err = json.Unmarshal(metadata, &fr)
	if err != nil {
		return fr, err
	}
	filedata, err := c.httpGetBytes(datapath)
	if err != nil {
		return fr, err
	}
	fr.Content = filedata

	return fr, nil
}

// FetchLegacyFile returns the file record
func (c *Client) FetchLegacyFileOld(filepath string, filename string) (filerecord.FileRecord, error) {

	fullpath := filepath + "/" + filename

	var err error

	if c.c == nil {
		return filerecord.FileRecord{}, errors.New("Not connected")
	}

	// Now do the connection
	r := []byte(c.Session + fullpath)
	c.c.SendMessage("LLD", r, true)

	// get response
	tochan := time.After(time.Second * 20)
	bb := &filerecord.FileRecord{}
	select {
	case _ = <-tochan:
		err = errors.New("timeout")
	case msg := <-c.c.Incoming:
		////fmt.Printf("in _FetchLegacyFile() %s\n", msg.ID)
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

// FetchLegacyDir returns the dir
func (c *Client) FetchLegacyDir(filepath string, filespec string) ([]byte, error) {
	fullpath := filepath

	var err error

	if c.c == nil {
		return []byte(nil), errors.New("Not connected")
	}

	// Now do the connection
	r := []byte(c.Session + fullpath + ":" + filespec)
	c.c.SendMessage("LDR", r, true)

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

// StoreLegacyFile stores a file in the legacy space
func (c *Client) StoreLegacyFile(filepath string, filename string, data []byte) error {

	fullpath := filepath + "/" + filename

	var err error

	if c.c == nil {
		return errors.New("Not connected")
	}

	// Now do the connection
	r := []byte(c.Session + fullpath + ":")
	r = append(r, data...)
	c.c.SendMessage("LSV", r, true)

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
			err = errors.New("failed")
		}
	}

	return err
}
