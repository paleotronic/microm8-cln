package s8webclient

import (
	"encoding/json"
	"errors"
	"paleotronic.com/log"
	"strings"

	"paleotronic.com/fmt"
	"runtime"
	"time"

	"paleotronic.com/core/settings"
	"paleotronic.com/ducktape/client"
	"paleotronic.com/ducktape/crypt"
	"paleotronic.com/filerecord"
	"paleotronic.com/utils"
)

const (
	LOGMAX = 10000
)

// Client defines the API client for the octalyzer server
type LogMessage struct {
	req     string
	message string
}

type Client struct {
	Hostname      string
	Port          string
	Username      string
	Password      string
	Session       string
	Authenticated bool
	c             *client.DuckTapeClient
	MOTD          string
	Messages      chan LogMessage
}

// GetDT returns the current active client
func (c *Client) GetDT() *client.DuckTapeClient {
	return c.c
}

func (c *Client) Done() {
	c.c.SendMessage("BYE", []byte(nil), false)
	time.Sleep(500 * time.Millisecond)
	c.c.Close()
}

// CONN is the global connection object
var CONN *Client

// NetworkInit is a public function used by a mainline to init the network
const PersistentConnect = false

var RetryCounter = 1

func MonitorNetwork() {
	for {
		time.Sleep(10 * time.Second)
		if !CONN.IsConnectedPoll() {
			e := CONN.c.ConnectTCP()

			if e == nil {
				go CONN.ProcessLogs()
			}
		}
	}
}

func NetworkInit(host string) error {
	log.Printf("[microM8] starting network layer... attempt %d", RetryCounter)
	CONN = New(host, settings.ServerPort)
	CONN.c = client.NewDuckTapeClient(CONN.Hostname, CONN.Port, utils.GetUniqueName(), "tcp")
	e := CONN.c.ConnectTCP()

	if e == nil {
		go CONN.ProcessLogs()
	}

	go MonitorNetwork()

	return e
}

// New creates a new network client
func New(host, port string) *Client {

	c := &Client{
		Hostname: host,
		Port:     port,
		Messages: make(chan LogMessage, LOGMAX),
		Username: "system",
		Session:  "deadbeefdeadbeefdeadbeefdeadbeef",
	}

	return c
}

// MOTD holds and stores a server message of the day
var MOTD string

func (c *Client) IsAuthenticated() bool {
	if c == nil {
		return false
	}
	return c.Authenticated
}

func (c *Client) IsConnectedPoll() bool {
	if c == nil || c.c == nil {
		return false
	}
	return CONN.c.Connected && CONN.c.OK
}

func (c *Client) IsConnected() bool {

	if c == nil || c.c == nil {
		return false
	}

	if !c.c.Connected {
		settings.EBOOT = true
		e := NetworkInit("paleotronic.com")
		if e == nil {
			settings.EBOOT = false
		}
	}

	return CONN.c.Connected
}

// Login attempts to log a user in with there password.
func (c *Client) Login(username, password string) (int, error) {

	var err error

	if c.c == nil {
		c.c = client.NewDuckTapeClient(c.Hostname, c.Port, username, "tcp")
		err := c.c.Connect()
		if err != nil {
			return 500, err
		}
	}

	kp := c.c.Token[0:]

	////fmt.Println("LOCAL ADDR IS ",c.c.Conn.LocalAddr().String())

	// Now do the connection
	//	c.c.SendMessage("LGN", []byte(username+":"+password), false)
	c.c.SendMessage("LGN", crypt.YogoCrypt([]byte(username+":"+password), kp, 3), true)

	// get response
	code := 500
	c.Session = ""
	tochan := time.After(time.Second * 20)
	select {
	case _ = <-tochan:
		code = 500
	case msg := <-c.c.Incoming:
		fmt.Printf("in _login() %s: %s\n", msg.ID, string(msg.Payload))
		if msg.ID == "LOK" {
			// Login OK
			code = 200
			c.Session = string(msg.Payload[:32])
			MOTD = string(msg.Payload[33:])
			c.Username = username
			c.Password = password

			platform := runtime.GOOS + "/" + runtime.GOARCH

			settings.DONTLOG = settings.DONTLOGDEFAULT

			c.LogMessage("UCL", "User has authenticated successfully from "+platform)

			c.Authenticated = true

		} else if msg.ID == "LFL" {
			code = 403
		}
	}

	CONN = c

	return code, err

}

// Register creates a user account on the server.
func (c *Client) Register(username, password, firstname, gender, birthdate, cell string, location string) error {

	var err error

	if c.c == nil {
		c.c = client.NewDuckTapeClient(c.Hostname, c.Port, username, "tcp")
		err := c.c.Connect()
		if err != nil {
			return err
		}
	}

	// Now do the connection
	c.c.SendMessage("REG", crypt.YogoCrypt([]byte(username+":"+password+":"+firstname+":"+gender+":"+birthdate+":"+cell+":"+location), c.c.Token[0:], 3), false)

	// get response
	c.Session = ""
	tochan := time.After(time.Second * 20)
	select {
	case _ = <-tochan:
		err = errors.New("timeout")
	case msg := <-c.c.Incoming:
		////fmt.Printf("in _login() %s\n", msg.ID)
		if msg.ID == "LOK" {
			// Login OK
			c.Session = string(msg.Payload[:32])
			c.MOTD = string(msg.Payload[33:])

			c.LogMessage("UCL", "User has registered successfully")

		} else if msg.ID == "ERR" {
			err = errors.New("registration failed: " + string(msg.Payload))
			fmt.Println(err)
		}
	}

	return err
}

// GRCustomGroup checks if a user has a specified group or not
func (c *Client) GRCustomGroup(req string, username string, groupname string) error {

	var err error

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

	// Now do the connection
	r := []byte(c.Session)
	r = append(r, []byte(username)...)
	r = append(r, []byte(":"+groupname)...)
	c.c.SendMessage(req, r, true)

	// get response
	tochan := time.After(time.Second * 20)
	select {
	case _ = <-tochan:
		err = errors.New("timeout")
	case msg := <-c.c.Incoming:
		//fmt.Printf("in GRCustomGroup() %s, %s\n", msg.ID, string(msg.Payload))
		if msg.ID == "POK" {
			err = nil
		} else if msg.ID == "ERR" {
			err = errors.New("i/o error")
		}
	}

	return err
}

// GrantReadGroup gives read access to a user
func (c *Client) GrantReadGroup(username string, groupname string) error {
	return c.GRCustomGroup("GRG", username, groupname)
}

// GrantWriteGroup gives write access to a user
func (c *Client) GrantWriteGroup(username string, groupname string) error {
	return c.GRCustomGroup("GWG", username, groupname)
}

// RevokeReadGroup revokes read access from a user
func (c *Client) RevokeReadGroup(username string, groupname string) error {
	return c.GRCustomGroup("RRG", username, groupname)
}

// RevokeWriteGroup revokes write access from a user
func (c *Client) RevokeWriteGroup(username string, groupname string) error {
	return c.GRCustomGroup("RWG", username, groupname)
}

// ChangePassword changes a user password
func (c *Client) ChangePassword(password string, newpassword string) error {

	var err error

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

	// Now do the connection
	r := []byte(c.Session)
	r = append(r, []byte(password)...)
	r = append(r, []byte(":"+newpassword)...)
	c.c.SendMessage("CPR", crypt.YogoCrypt(r, c.c.Token[0:], 3), true)

	// get response
	tochan := time.After(time.Second * 20)
	select {
	case _ = <-tochan:
		err = errors.New("timeout")
	case msg := <-c.c.Incoming:
		//fmt.Printf("in ChangePassword() %s, %s\n", msg.ID, string(msg.Payload))
		if msg.ID == "CPO" {
			err = nil
		} else if msg.ID == "ERR" {
			err = errors.New(string(msg.Payload))
		}
	}

	return err
}

// CustomUserQuery queries a users meta data
func (c *Client) CustomUserQuery(req string, targetuser string) (string, error) {

	var err error

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

	// Now do the connection
	r := []byte(c.Session)
	r = append(r, []byte(targetuser)...)
	c.c.SendMessage(req, r, true)

	var result string

	// get response
	tochan := time.After(time.Second * 20)
	select {
	case _ = <-tochan:
		err = errors.New("timeout")
	case msg := <-c.c.Incoming:
		//fmt.Printf("in ChangePassword() %s, %s\n", msg.ID, string(msg.Payload))
		if msg.ID == "UIO" {
			err = nil
			result = string(msg.Payload)
		} else if msg.ID == "ERR" {
			err = errors.New(string(msg.Payload))
		}
	}

	return result, err
}

// GetUserFirstName returns users first name
func (c *Client) GetUserFirstName() (string, error) {
	return c.CustomUserQuery("GFN", c.Username)
}

// GetUserDOB returns user DOB
func (c *Client) GetUserDOB() (string, error) {
	return c.CustomUserQuery("GDB", c.Username)
}

// GetUserGender retrieves a users gender
func (c *Client) GetUserGender() (string, error) {
	return c.CustomUserQuery("GGD", c.Username)
}

// GetRemoteInstances retrieves remote instances from the server
func (c *Client) GetRemoteInstances(filter string) ([]filerecord.RemoteInstance, error) {

	var err error
	var result []filerecord.RemoteInstance

	if c.c == nil {
		c.Username = "system"
		c.c = client.NewDuckTapeClient(c.Hostname, c.Port, c.Username, "tcp")
		err := c.c.Connect()
		if err != nil {
			//			//fmt.Println("Failed connect")
			return result, err
		}
		//		//fmt.Println("Connected")
	}

	if c.Session == "" {
		c.Session = "12345678123456781234567812345678"
	}

	// Now do the connection
	r := []byte(c.Session)
	r = append(r, []byte(filter)...)
	c.c.SendMessage("GRI", r, true)

	// get response
	tochan := time.After(time.Second * 20)
	select {
	case _ = <-tochan:
		err = errors.New("timeout")
	case msg := <-c.c.Incoming:
		//fmt.Printf("in GetRemoteInstances() %s, %s\n", msg.ID, string(msg.Payload))
		if msg.ID == "GRO" {
			err = nil
			_ = json.Unmarshal(msg.Payload, &result)
		} else if msg.ID == "ERR" {
			err = errors.New(string(msg.Payload))
		}
	}

	return result, err

}

// CreateUpdateBug lodges a bug report
func (c *Client) CreateUpdateBug(bug filerecord.BugReport) error {

	var err error

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

	// Now do the connection
	r := []byte(c.Session)
	r = append(r, bug.BSON()...)

	code := "BGC"
	if bug.DefectID > 0 {
		code = "BGU"
	}

	c.c.SendMessage(code, r, true)

	// get response
	tochan := time.After(time.Second * 20)
	select {
	case _ = <-tochan:
		err = errors.New("timeout")
	case msg := <-c.c.Incoming:
		//fmt.Printf("in CreateUpdateBug() %s, %s\n", msg.ID, string(msg.Payload))
		if msg.ID == "ERR" {
			err = errors.New(string(msg.Payload))
		} else {
			err = nil
		}
	}

	return err

}

// GetBugList returns the current visible bugs
func (c *Client) GetBugList(t filerecord.BugType) ([]filerecord.BugReport, error) {

	var err error
	var result []filerecord.BugReport

	if c.c == nil {
		c.Username = "system"
		c.c = client.NewDuckTapeClient(c.Hostname, c.Port, c.Username, "tcp")
		err := c.c.Connect()
		if err != nil {
			//			//fmt.Println("Failed connect")
			return result, err
		}
		//		//fmt.Println("Connected")
	}

	if c.Session == "" {
		c.Session = "12345678123456781234567812345678"
	}

	// Now do the connection
	r := []byte(c.Session)
	r = append(r, byte(t))
	c.c.SendMessage("BGL", r, true)

	// get response
	tochan := time.After(time.Second * 20)
	select {
	case _ = <-tochan:
		err = errors.New("timeout")
	case msg := <-c.c.Incoming:
		//fmt.Printf("in GetBugList() %s, %s\n", msg.ID, string(msg.Payload))
		if msg.ID == "BLO" {
			err = nil
			_ = json.Unmarshal(msg.Payload, &result)
		} else if msg.ID == "ERR" {
			err = errors.New(string(msg.Payload))
		}
	}

	return result, err

}

// GetBugByID retrieves a bug report from the server
func (c *Client) GetBugByID(id int64) (*filerecord.BugReport, error) {

	var err error
	var bug *filerecord.BugReport

	if c.c == nil {
		c.Username = "system"
		c.c = client.NewDuckTapeClient(c.Hostname, c.Port, c.Username, "tcp")
		err := c.c.Connect()
		if err != nil {
			//			//fmt.Println("Failed connect")
			return bug, err
		}
		//		//fmt.Println("Connected")
	}

	if c.Session == "" {
		c.Session = "12345678123456781234567812345678"
	}

	// Now do the connection
	r := []byte(c.Session)

	bug = &filerecord.BugReport{}
	bug.DefectID = id

	r = append(r, bug.BSON()...)

	code := "BGF"

	c.c.SendMessage(code, r, true)

	// get response
	tochan := time.After(time.Second * 20)
	select {
	case _ = <-tochan:
		err = errors.New("timeout")
	case msg := <-c.c.Incoming:
		//fmt.Printf("in CreateUpdateBug() %s, %s\n", msg.ID, string(msg.Payload))
		if msg.ID == "ERR" {
			err = errors.New(string(msg.Payload))
		} else {
			err = nil
			bug.UnBSON(msg.Payload)
		}
	}

	return bug, err

}

// GetRemoteToken retrieves a remote access token
func (c *Client) GetRemoteToken(hostname string, port string) ([]byte, error) {

	if hostname == "localhost" || hostname == "127.0.0.1" || strings.HasPrefix(hostname, "192.168.") || strings.HasPrefix(hostname, "10.") {
		fmt.Println("====================== LOCAL CONNECT =====================")
		return []byte("c47b33f54a73c007"), nil
	}

	var err error

	var rr []byte

	if c.c == nil {
		c.Username = "system"
		c.c = client.NewDuckTapeClient(c.Hostname, c.Port, c.Username, "tcp")
		err := c.c.Connect()
		if err != nil {
			//			//fmt.Println("Failed connect")
			return rr, err
		}
		//		//fmt.Println("Connected")
	}

	if c.Session == "" {
		c.Session = "12345678123456781234567812345678"
	}

	// Now do the connection
	r := []byte(c.Session)
	r = append(r, []byte(hostname)...)
	r = append(r, []byte(":"+port)...)
	c.c.SendMessage("RAR", r, true)

	// get response
	tochan := time.After(time.Second * 20)
	select {
	case _ = <-tochan:
		err = errors.New("timeout")
	case msg := <-c.c.Incoming:
		//fmt.Printf("in GRCustomGroup() %s, %s\n", msg.ID, string(msg.Payload))
		rr = msg.Payload
		if msg.ID == "RAO" {
			err = nil
		} else if msg.ID == "ERR" {
			err = errors.New(string(msg.Payload))
		}
	}

	return rr, err
}

// GetMOTD will return MOTD to user
func (c *Client) GetMOTD() string {
	return MOTD
}

// AddMOTD will add a line to MOTD
func (c *Client) AddMOTD(motd string) error {

	var err error

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

	// Now do the connection
	r := []byte(c.Session)
	r = append(r, []byte(motd)...)
	c.c.SendMessage("MTD", r, true)

	// get response
	tochan := time.After(time.Second * 20)
	select {
	case _ = <-tochan:
		err = errors.New("timeout")
	case msg := <-c.c.Incoming:
		fmt.Printf("in AddMOTD() %s, %s\n", msg.ID, string(msg.Payload))
		if msg.ID == "MTO" {
			err = nil
		} else if msg.ID == "ERR" {
			err = errors.New(string(msg.Payload))
		}
	}

	return err
}
