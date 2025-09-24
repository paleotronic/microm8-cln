package s8webclient

import (
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	"paleotronic.com/fmt"

	"regexp"

	"paleotronic.com/ducktape"
	"paleotronic.com/ducktape/client"
	//	"paleotronic.com/filerecord"
)

var reTable = regexp.MustCompile("(?i)(FROM|INTO) ([a-z][a-z0-9]+)")
var dbhandle int
var handlemap map[int]string
var stmthandle int
var stmtmap map[int][]map[string]interface{}
var appnametostmtid map[string]int
var stm sync.Mutex

func init() {
	handlemap = make(map[int]string)
	stmtmap = make(map[int][]map[string]interface{})
	appnametostmtid = make(map[string]int)
	//Do("insert into frogs (name, age) values ('kermit', 24);")
}

// AppDatabaseConnect tries to connect to the specified database
func (c *Client) AppDatabaseConnect(req string, appname string, appsig string) (int, error) {

	var err error

	if c.c == nil {
		c.Username = "system"
		c.c = client.NewDuckTapeClient(c.Hostname, c.Port, c.Username, "tcp")
		err := c.c.Connect()
		if err != nil {
			//			//fmt.Println("Failed connect")
			return -1, err
		}
		//		//fmt.Println("Connected")
	}

	if c.Session == "" {
		c.Session = "12345678123456781234567812345678"
	}

	// Now do the connection
	r := []byte(c.Session)
	r = append(r, byte(':'))
	r = append(r, []byte(appname)...)
	r = append(r, byte(':'))
	r = append(r, []byte(appsig)...)
	c.c.SendMessage(req, r, true)

	// get response
	tochan := time.After(time.Second * 20)
	select {
	case _ = <-tochan:
		err = errors.New("timeout")
	case msg := <-c.c.Incoming:
		fmt.Printf("in AppDatabaseConnect() %s, %s\n", msg.ID, string(msg.Payload))
		if msg.ID == "DCO" {
			dbhandle++
			handlemap[dbhandle] = appname
		} else if msg.ID == "ERR" {
			err = errors.New("i/o error")
		}
	}

	return dbhandle, err
}

// AppDatabaseQuery runs a query on the remote database
func (c *Client) AppDatabaseQuery(req string, handle int, appsig string, query string) (int, error) {

	c.c.CustomHandler["STI"] = HandleSTI

	var tablename string

	if reTable.MatchString(query) {
		m := reTable.FindAllStringSubmatch(query, -1)
		tablename = m[0][2]
	}

	stmthandle++

	// setup handle
	appname, ok := handlemap[handle]
	if !ok {
		return -1, errors.New("Invalid DB handle")
	}
	appnametostmtid[appname+"."+tablename] = stmthandle

	var err error

	if c.c == nil {
		c.Username = "system"
		c.c = client.NewDuckTapeClient(c.Hostname, c.Port, c.Username, "tcp")
		err := c.c.Connect()
		if err != nil {
			//			//fmt.Println("Failed connect")
			return -1, err
		}
		//		//fmt.Println("Connected")
	}

	if c.Session == "" {
		c.Session = "12345678123456781234567812345678"
	}

	// Now do the connection
	r := []byte(c.Session)
	r = append(r, byte(':'))
	r = append(r, []byte(appname)...)
	r = append(r, byte(':'))
	r = append(r, []byte(appsig)...)
	r = append(r, byte(':'))
	r = append(r, []byte(query)...)
	c.c.SendMessage(req, r, true)

	// get response
	tochan := time.After(time.Second * 20)
	select {
	case _ = <-tochan:
		err = errors.New("timeout")
	case msg := <-c.c.Incoming:
		fmt.Printf("in AppDatabaseQuery() %s, %s\n", msg.ID, string(msg.Payload))
		if msg.ID == "DQO" {

			var data []map[string]interface{}

			e := json.Unmarshal(msg.Payload, &data)

			m, ok := stmtmap[stmthandle]
			if ok {
				m = append(m, data...)
				stmtmap[stmthandle] = m
			} else {
				stmtmap[stmthandle] = data
			}

			return stmthandle, e
		} else if msg.ID == "ERR" {
			err = errors.New(string(msg.Payload))
		}
	}

	return -1, err
}

// DBResultCount returns the number of pending results on a handle
func (c *Client) DBResultCount(sthandle int) (int, error) {

	//fmt.Printf("Checking buffer for stmtid = %d\n", sthandle)

	stm.Lock()
	defer stm.Unlock()
	m, ok := stmtmap[sthandle]
	if !ok {
		return 0, errors.New("No such statement handle")
	}
	return len(m), nil
}

// DBResultFetch fetches one record from a statement handle
func (c *Client) DBResultFetch(sthandle int) (map[string]interface{}, error) {

	fmt.Printf("Fetching records for stmtid = %d\n", sthandle)

	stm.Lock()
	defer stm.Unlock()

	m, ok := stmtmap[sthandle]
	if !ok {
		fmt.Println("no handle", sthandle)
		return nil, errors.New("No such statement handle")
	}

	fmt.Println(m)

	if len(m) == 0 {
		fmt.Println("no results for handle", sthandle)
		return nil, errors.New("No more results")
	}

	v := m[0]
	m = m[1:]

	stmtmap[sthandle] = m

	fmt.Printf("Fetched: %v\n", v)

	return v, nil
}

// DBResultDone closes an open statement handle
func (c *Client) DBResultDone(sthandle int) {
	stm.Lock()
	defer stm.Unlock()
	delete(stmtmap, sthandle)
}

// HandleSTI is a callback for handling a push database update from the server
func HandleSTI(c *client.DuckTapeClient, msg *ducktape.DuckTapeBundle) {
	fmt.Println("Got a table update")

	parts := strings.SplitN(string(msg.Payload), ":", 3)

	appname := parts[0]
	tablename := parts[1]
	jb := []byte(parts[2])

	var data map[string]interface{}
	e := json.Unmarshal(jb, &data)
	if e != nil {
		fmt.Println(e)
		return
	}

	fmt.Println(data)

	fmt.Println("Check for update to " + appname + "." + tablename)
	stmtid, ok := appnametostmtid[appname+"."+tablename]
	if ok {
		stm.Lock()
		defer stm.Unlock()

		m, ok := stmtmap[stmtid]
		if !ok {
			m = make([]map[string]interface{}, 0)
		}

		m = append(m, data)

		stmtmap[stmtid] = m

		fmt.Printf("Added record to stmt handle %d\n", stmtid)
		fmt.Printf("Blah: %v\n", stmtmap[stmtid])

	}

}

// DBUnsubscribeALL removes all subscriptions from table updates
func (c *Client) DBUnsubscribeALL() {
	stm.Lock()
	defer stm.Unlock()

	for k := range appnametostmtid {
		ch := "#" + k
		c.c.UnsubscribeChannel(ch)
	}

	appnametostmtid = make(map[string]int)

}

func (c *Client) ProcessLogs() {

	for c.c.Connected {

		lm := <-c.Messages

		req := lm.req
		message := lm.message

		//~ LogMessages = append(LogMessages, message)

		if c.c == nil {
			c.Username = "system"
			c.c = client.NewDuckTapeClient(c.Hostname, c.Port, c.Username, "tcp")
			err := c.c.Connect()
			if err != nil {
				//			//fmt.Println("Failed connect")
				continue
			}
			//		//fmt.Println("Connected")
		}

		if c.Session == "" {
			c.Session = "12345678123456781234567812345678"
		}

		if !c.c.Connected {
			continue
		}

		//for _, lm := range LogMessages {

		// Now do the connection
		r := []byte(c.Session)
		r = append(r, byte(':'))
		r = append(r, []byte(message)...)
		c.c.SendMessage(req, r, true)

		tochan := time.After(time.Second * 20)
		select {
		case _ = <-tochan:
			continue
		case msg := <-c.c.Incoming:
			fmt.Printf("in LogMessage() %s, %s\n", msg.ID, string(msg.Payload))
			if msg.ID == "ULO" {
				continue
			} else if msg.ID == "ERR" {
				continue
			}
		}

	}
}

// LogMessage sends a log event to the server to record in a users log table
func (c *Client) LogMessage(req string, message string) error {

	return nil

	c.Messages <- LogMessage{
		req:     req,
		message: message,
	}

	return nil
}

func (c *Client) GetKeyValue(req string, keyname string) (string, error) {

	var err error

	//~ LogMessages = append(LogMessages, message)

	if c.c == nil {
		c.Username = "system"
		c.c = client.NewDuckTapeClient(c.Hostname, c.Port, c.Username, "tcp")
		err := c.c.Connect()
		if err != nil {
			//			//fmt.Println("Failed connect")
			return "", err
		}
		//		//fmt.Println("Connected")
	}

	if c.Session == "" {
		c.Session = "12345678123456781234567812345678"
	}

	if !c.c.Connected {
		return "", nil
	}

	//for _, lm := range LogMessages {

	// Now do the connection
	r := []byte(c.Session)
	r = append(r, byte(':'))
	r = append(r, []byte(keyname)...)
	c.c.SendMessage(req, r, true)

	tochan := time.After(time.Second * 20)
	select {
	case _ = <-tochan:
		err = errors.New("timeout")
		fmt.Println("timeout")
	case msg := <-c.c.Incoming:
		fmt.Printf("in SetKeyValue() %s, %s\n", msg.ID, string(msg.Payload))
		if msg.ID == "GKO" {

			s := string(msg.Payload)
			return s, nil
		} else if msg.ID == "ERR" {
			err = errors.New(string(msg.Payload))
			fmt.Println(err)
		}
	}

	return "", err
}

func (c *Client) SetKeyValue(req string, keyname string, value string) error {

	var err error

	//~ LogMessages = append(LogMessages, message)

	if c.c == nil {
		c.Username = "system"
		c.c = client.NewDuckTapeClient(c.Hostname, c.Port, c.Username, "tcp")
		err := c.c.Connect()
		if err != nil {
			//			//fmt.Println("Failed connect")
			return err
		}
		//		//fmt.Println("Connected")
	}

	if c.Session == "" {
		c.Session = "12345678123456781234567812345678"
	}

	if !c.c.Connected {
		return nil
	}

	//for _, lm := range LogMessages {

	// Now do the connection
	r := []byte(c.Session)
	r = append(r, byte(':'))
	r = append(r, []byte(keyname)...)
	r = append(r, byte(':'))
	r = append(r, []byte(value)...)
	fmt.Printf("Sending message: %s %s\n", req, string(r))
	c.c.SendMessage(req, r, true)

	tochan := time.After(time.Second * 20)
	select {
	case _ = <-tochan:
		err = errors.New("timeout")
	case msg := <-c.c.Incoming:
		fmt.Printf("in SetKeyValue() %s, %s\n", msg.ID, string(msg.Payload))
		if msg.ID == "KVO" {
			return nil
		} else if msg.ID == "ERR" {
			err = errors.New(string(msg.Payload))
		}
	}

	return err
}
