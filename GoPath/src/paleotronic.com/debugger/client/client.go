package debugclient

import (
	"encoding/json"
	"net/url"
	"time"

	"paleotronic.com/log"

	"paleotronic.com/core/settings"

	"paleotronic.com/utils"

	"paleotronic.com/debugger/debugtypes"

	"github.com/gorilla/websocket"
)

type DebugHandlerFunc func(msg *debugtypes.WebSocketMessage)

type DebugClient struct {
	Host      string
	Port      string
	Path      string
	conn      *websocket.Conn
	running   bool
	functions map[string]DebugHandlerFunc
	slotid    int
}

func NewDebugClient(slotid int, host, port, path string) *DebugClient {
	return &DebugClient{
		Host:      host,
		Port:      port,
		Path:      path,
		functions: map[string]DebugHandlerFunc{},
		slotid:    slotid,
	}
}

func (c *DebugClient) Handle(kind string, f DebugHandlerFunc) {
	c.functions[kind] = f
}

func (c *DebugClient) Connect() error {
	addr := c.Host + ":" + c.Port
	u := url.URL{Scheme: "ws", Host: addr, Path: c.Path}
	log.Printf("connecting to %s", u.String())

	cn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	log.Printf("Connected")
	c.conn = cn

	c.SendCommand("attach", []string{utils.IntToStr(c.slotid)})

	settings.DebuggerAttachSlot = c.slotid + 1

	go c.Process()

	return nil
}

func (c *DebugClient) Process() {
	c.running = true
	for c.running && c.conn != nil {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		//log.Printf("recv: %s", message)
		c.handleMessage(message)
	}
}

func (c *DebugClient) Disconnect() error {
	c.SendCommand("detach", nil)
	time.Sleep(time.Second)
	c.running = false
	c.conn.Close()
	//debug.PrintStack()
	settings.DebuggerAttachSlot = -1
	return nil
}

func (c *DebugClient) handleMessage(msg []byte) {

	m := debugtypes.WebSocketMessage{}

	err := json.Unmarshal(msg, &m)
	if err != nil {
		log.Printf("bad message: %v", msg)
		return
	}

	log.Printf("received message: %s (ok=%v) (payload=%+v)", m.Type, m.Ok, m.Payload)
	switch m.Type {
	default:
		if f, ok := c.functions[m.Type]; ok {
			f(&m)
		} else if f, ok := c.functions["*"]; ok {
			f(&m)
		}
	}

}

func (c *DebugClient) SendCommand(command string, args []string) error {
	msg := &debugtypes.WebSocketCommand{
		Type: command,
		Args: args,
	}
	j, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	log.Printf("sending command: %s", string(j))
	return c.conn.WriteMessage(1, j)
}
