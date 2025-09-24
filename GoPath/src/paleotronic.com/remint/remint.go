package main

import (
	"errors"
	"flag"
	"os"
	"runtime"
	"strings"
	"sync"
	"time" //	"net/http"
	//	_ "net/http/pprof"

	"paleotronic.com/api"
	"paleotronic.com/core"
	"paleotronic.com/core/dialect/appleinteger"
	"paleotronic.com/core/dialect/applesoft"
	"paleotronic.com/core/dialect/shell"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/interpreter"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/settings"
	"paleotronic.com/ducktape"
	"paleotronic.com/ducktape/client"
	"paleotronic.com/ducktape/server"
	"paleotronic.com/encoding/mempak"
	"paleotronic.com/files"
	"paleotronic.com/fmt"
	"paleotronic.com/log" //"paleotronic.com/fmt"
	"paleotronic.com/utils"
	//"github.com/pkg/profile"
	//_ "net/http/pprof"
)

const REMINT_SLOTS = memory.OCTALYZER_NUM_INTERPRETERS

var RAM *memory.MemoryMap
var mutex sync.Mutex
var p *core.Producer
var useport = flag.Int("port", 8580, "Port to run interpreter on")
var usedialect = flag.String("dialect", "fp", "Default dialect to run")
var bootstrap = flag.String("bootstrap", "_login@ = \"sandbox\", \"sandbow\", 5, \"SESS$\"", "Session start command")
var specfile = flag.String("specfile", settings.DefaultProfile, "Spec to start with")
var owner = flag.String("owner", "system", "Current process owner")
var uuid = flag.String("uuid", "", "UUID of filesystem mount")
var max = flag.Int("max-slots", 1, "Max remint slots")
var servername = flag.String("server", "paleotronic.com", "Address of server")
var s *server.DuckTapeServer
var c [memory.OCTALYZER_NUM_INTERPRETERS]*client.DuckTapeClient
var m server.DuckHandlerMap
var dia interfaces.Dialecter
var exec = flag.String("exec", "print \"hello\"", "Command to execute")
var BlockUnmapped bool
var matrix map[string]*interpreter.InputMatrix
var presets map[string]string
var changes [memory.OCTALYZER_NUM_INTERPRETERS]chan []memory.MemoryChange

var VALID_CONTROL map[string]map[int]int

func DefinePresets() {
	presets = make(map[string]string)
	presets["table-left"] = P1_PONG_PERSPECTIVE
	presets["table-right"] = P2_PONG_PERSPECTIVE
}

func DefineMatrix() {
	VALID_CONTROL = make(map[string]map[int]int)

	VALID_CONTROL["april"] = make(map[int]int)
	VALID_CONTROL["april"][memory.PDL0] = memory.PDL0
	VALID_CONTROL["april"][memory.PDL1] = -1
	VALID_CONTROL["melody"] = make(map[int]int)
	VALID_CONTROL["melody"][memory.PDL0] = memory.PDL1
	VALID_CONTROL["melody"][memory.PDL1] = -1
}

func init() {
	DefinePresets()
	DefineMatrix()
}

func ProcessIncoming(slotid int) {

	c[slotid].ClientReaderSingle()

	select {
	case msg := <-c[slotid].Incoming:
		switch msg.ID {

		case "FUL":
			// request for full update

			data, _ := p.GetInterpreters()[slotid].FreezeBytes()
			cdata := utils.GZIPBytes(data)
			c[slotid].SendMessage("RAM", cdata, true)
			log.Printf("Sending RAM snapshot of %d bytes\n", len(cdata))
		case "IEV":
			ie := &interpreter.InputEvent{}
			e := ie.UnmarshalBinary(msg.Payload)
			log.Println(ie)

			user := c[slotid].Name

			im, ok := matrix[user]
			if ok {
				// matrix exists
				*ie = im.FilterEvent(*ie)
				log.Printf("--> XLATED message IEV from %s: %d / %d / %d\n", user, ie.Kind, ie.ID, ie.Value)
			}

			if e == nil {
				// process
				switch ie.Kind {
				case interpreter.ET_KEYPRESS:
					log.Printf("Keypress %d\n", ie.Value)
					mutex.Lock()
					RAM.KeyBufferAdd(slotid, uint(ie.Value))
					mutex.Unlock()
				case interpreter.ET_PADDLE_BUTTON_DOWN:
					mutex.Lock()
					RAM.IntSetPaddleButton(slotid, int(ie.ID), uint(ie.Value))
					mutex.Unlock()
				case interpreter.ET_PADDLE_BUTTON_UP:
					mutex.Lock()
					RAM.IntSetPaddleButton(slotid, int(ie.ID), uint(ie.Value))
					mutex.Unlock()
				case interpreter.ET_PADDLE_VALUE_CHANGE:
					mutex.Lock()
					RAM.IntSetPaddleValue(slotid, int(ie.ID), uint(ie.Value))
					mutex.Unlock()
				}
			} else {
				log.Println(e)
			}
		}
	default:
		break
	}

	if c[slotid].NeedReconnect {
		fmt.Print("E")
		c[slotid].ConnectSingle()
		return
	}

	fmt.Print("i")

}

func ProcessOutgoing(slotid int) {

	// Send whatever is buffered
	c[slotid].ClientSenderSingle()

	if c[slotid].NeedReconnect {
		fmt.Print("E")
		c[slotid].ConnectSingle()
		return
	}

	fmt.Print("o")

}

func SetupClient(slotid int) {
	c[slotid] = client.NewDuckTapeClient("localhost", ":"+utils.IntToStr(*useport), "ORAC"+utils.IntToStr(slotid), "tcp")
	e := c[slotid].ConnectSingle()
	if e != nil {
		panic(e)
	}
	log.Printf("RAM slave %s connected to port %d\n", c[slotid].Name, *useport)
	c[slotid].SendMessage("SND", []byte("memupd"+utils.IntToStr(slotid)), false)
	c[slotid].SendMessage("SUB", []byte("cliupd"+utils.IntToStr(slotid)), false)

	ProcessOutgoing(slotid)
	//time.Sleep(10*time.Millisecond)
	ProcessIncoming(slotid)
}

func HandleRAMUpdate(index int, addr int, value uint) {
	//log.Printf("%d/%d ---> %d\n", index, addr, value)

	b := make([]byte, 0)
	b = append(b, byte(index))
	b = append(b, byte((addr/65536)%256))
	b = append(b, byte((addr/256)%256))
	b = append(b, byte(addr%256))
	// data
	b = append(b, byte((value>>24)&255))
	b = append(b, byte((value>>16)&255))
	b = append(b, byte((value>>8)&255))
	b = append(b, byte((value>>0)&255))

	c[index].SendMessage("BYT", b, true)
}

func ChangeSender(slotid int) {

	for {
		if !PostCycleCallback(slotid) {
			time.Sleep(1 * time.Millisecond)
		}
	}

}

func PostCycleCallback(slotid int) bool {
	data := RAM.GetRemoteLoggedChanges(slotid)
	count := 0
	if len(data) > 0 {

		payload := make([]byte, 3)

		for _, mc := range data {

			if len(mc.Value) == 0 {
				payload = append(payload, mempak.Encode(0, mc.Global, 0, true)...)
				count++
			} else {

				if mc.Global >= memory.OCTALYZER_SIM_SIZE || RAM.IsMappedAddress(mc.Global) {

					for i, v := range mc.Value {
						payload = append(payload, mempak.Encode(0, mc.Global+i, v, false)...)
						count++
					}

				}
			}

		}

		if count > 0 {
			payload[0] = byte((count / 65536) % 256)
			payload[1] = byte((count / 256) % 256)
			payload[2] = byte(count % 256)
			c[slotid].SendMessage("BMU", payload, true)

			//fmt.Printf("*** %d memchanges shipped for slot %d\n", count, slotid)
			ProcessOutgoing(slotid)
			ProcessIncoming(slotid)
			return true
		}
	}

	ProcessOutgoing(slotid)
	ProcessIncoming(slotid)

	return false
}

func main() {

	//go http.ListenAndServe(":6569", http.DefaultServeMux)
	runtime.GOMAXPROCS(96)

	//defer profile.Start().Stop()

	BlockUnmapped = false

	RAM = memory.NewMemoryMap()
	//RAM.SetCallback(HandleRAMUpdate)
	//RAM.SetCallback(HandleMemThings, 0)
	//matrix = make(map[string]*interpreter.InputMatrix)

	flag.Parse()

	if *max == 0 {
		os.Exit(1)
	}

	if *max > REMINT_SLOTS {
		*max = REMINT_SLOTS
	}

	s8webclient.CONN = s8webclient.New(*servername, ":6581")

	*bootstrap = "_login@ = \"" + *uuid + "\", \"sandbow\", 5, \"SESS$\""

	m = make(server.DuckHandlerMap)
	m["TOS"] = TransferHandler
	m["TRM"] = TerminateHandler
	m["SCM"] = ControlsHandler
	m["CAM"] = CameraHandler
	m["QIM"] = QueryInputHandler
	m["CBY"] = MemHandler
	m["STK"] = STHandler
	//
	s = server.NewDuckTapeServer(":"+utils.IntToStr(*useport), m)
	s.OnDisconnect = OnChange
	go s.Run()

	fmt.Println(*usedialect)

	switch *usedialect {
	case "fp":
		dia = applesoft.NewDialectApplesoft()
	case "int":
		dia = appleinteger.NewDialectAppleInteger()
	case "shell":
		dia = shell.NewDialectShell()
	}

	fmt.Println("DIALECT LOADED", dia.GetTitle())

	p = core.NewProducerWithParams(RAM, *bootstrap, *exec, dia, *specfile)

	fmt.Println("CRAP LOADED")

	//p.SetPostCallback(PostCycleCallback)

	for i := 0; i < *max; i++ {
		changes[i] = make(chan []memory.MemoryChange, 10000)
		RAM.Track[i] = true
		RAM.MemCapMode[i] = memory.MEMCAP_REMOTE // <--- remote capture mode
		p.GetInterpreters()[i].SetClientSync(false)
		RAM.SetWaveCallback(i, AudioSink)
		time.Sleep(50 * time.Millisecond)
		SetupClient(i)
		//time.Sleep(50 * time.Millisecond)
		//ChangeSender(i)
	}

	if *uuid != "" {
		files.SetRemIntUUID(*uuid)
	}

	// spawn memory monitor...
	p.Run()
}

func TransferHandler(c *ducktape.Client, s *server.DuckTapeServer, msg *ducktape.DuckTapeBundle) error {

	if c.Name != *owner {
		return errors.New("Not owner")
	}

	*owner = string(msg.Payload)

	return nil

}

func TerminateHandler(c *ducktape.Client, s *server.DuckTapeServer, msg *ducktape.DuckTapeBundle) error {

	if c.Name != *owner {
		return errors.New("Not owner")
	}

	os.Exit(0)

	return nil

}

func CameraHandler(c *ducktape.Client, s *server.DuckTapeServer, msg *ducktape.DuckTapeBundle) error {

	/*
		<username>: <json> || <profile>
		<username>: <json> || <profile>
	*/

	log.Printf("Message %s : %s from %s\n", msg.ID, string(msg.Payload), c.Name)

	if c.Name != *owner {
		return errors.New("Not owner")
	}

	// user is owner
	str := string(msg.Payload)

	parts := strings.SplitN(str, ":", 2)

	name := parts[0]
	body := strings.Trim(parts[1], " ")

	// if the body is a single word, it should be assumed to contain
	// a named preset.
	preset, ok := presets[body]
	if !ok && rune(body[0]) == '{' {
		// got json
		preset = body
	} else if !ok {
		log.Printf("inavlid cam: %s\n", body)
		return errors.New("Invalid camera postion")
	}

	// relay camera change to desired player
	cc := s.FindClient(name + "-control")
	if cc == nil {
		log.Printf("Client not found %s\n", name)
		return errors.New("Invalid target")
	}

	log.Printf("camera = %s\n", preset)

	cc.Incoming <- ducktape.DuckTapeBundle{
		ID:      "FXC",
		Payload: []byte(preset),
		Binary:  true,
	}

	return nil
}

func ControlsHandler(c *ducktape.Client, s *server.DuckTapeServer, msg *ducktape.DuckTapeBundle) error {

	log.Printf("Message %s from %s\n", msg.ID, c.Name)

	//if c.Name != *owner {
	//	return errors.New("Not owner")
	//}

	// user is owner
	str := string(msg.Payload)

	/*
		april: p0=p0, p1=off, p2=off, p3=off, keys=on
		melody: p0=p1, p1=off, p2=off, p3=off, keys=on
	*/

	var defmap = map[string]string{}

	parts := strings.SplitN(str, ":", 2)

	if len(parts) < 2 {
		return errors.New("error in command")
	}

	name := strings.Trim(parts[0], " ")
	parts = strings.Split(parts[1], " ")

	for _, v := range parts {
		if v == "" {
			continue
		}
		pp := strings.SplitN(v, "=", 2)
		if len(pp) < 2 {
			return errors.New("error in command")
		}
		nn := strings.Trim(pp[0], " ")
		vv := strings.Trim(pp[1], " ")

		defmap[nn] = vv
	}

	log.Println(name, defmap)

	i, hasmap := matrix[name]
	if !hasmap {
		i = interpreter.NewStdInputMatrix()
	}

	for n, v := range defmap {
		if n == v {
			continue
		}
		if v == "on" {
			continue
		}
		if v == "off" {
			switch rune(n[0]) {
			case 'p':
				index := byte(utils.StrToInt(string(n[1])))

				m1, ok1 := i.Data[interpreter.ET_PADDLE_VALUE_CHANGE]
				m2, ok2 := i.Data[interpreter.ET_PADDLE_BUTTON_DOWN]
				m3, ok3 := i.Data[interpreter.ET_PADDLE_BUTTON_UP]
				if !ok1 {
					m1 = make(map[byte]interpreter.InputAction)
					i.Data[interpreter.ET_PADDLE_VALUE_CHANGE] = m1
				}
				if !ok2 {
					m2 = make(map[byte]interpreter.InputAction)
					i.Data[interpreter.ET_PADDLE_BUTTON_DOWN] = m1
				}
				if !ok3 {
					m3 = make(map[byte]interpreter.InputAction)
					i.Data[interpreter.ET_PADDLE_BUTTON_UP] = m1
				}
				m1[index] = interpreter.InputAction{Kind: interpreter.ET_NONE}
				m2[index] = interpreter.InputAction{Kind: interpreter.ET_NONE}
				m3[index] = interpreter.InputAction{Kind: interpreter.ET_NONE}
			case 'k':
				m1, ok1 := i.Data[interpreter.ET_KEYPRESS]
				if !ok1 {
					m1 = make(map[byte]interpreter.InputAction)
					i.Data[interpreter.ET_KEYPRESS] = m1
				}
				m1[0] = interpreter.InputAction{Kind: interpreter.ET_NONE}
			}
		}
		if rune(n[0]) == 'p' {
			switch rune(n[0]) {
			case 'p':
				index := byte(utils.StrToInt(string(n[1])))
				tindex := byte(utils.StrToInt(string(v[1])))

				m1, ok1 := i.Data[interpreter.ET_PADDLE_VALUE_CHANGE]
				m2, ok2 := i.Data[interpreter.ET_PADDLE_BUTTON_DOWN]
				m3, ok3 := i.Data[interpreter.ET_PADDLE_BUTTON_UP]
				if !ok1 {
					m1 = make(map[byte]interpreter.InputAction)
					i.Data[interpreter.ET_PADDLE_VALUE_CHANGE] = m1
				}
				if !ok2 {
					m2 = make(map[byte]interpreter.InputAction)
					i.Data[interpreter.ET_PADDLE_BUTTON_DOWN] = m1
				}
				if !ok3 {
					m3 = make(map[byte]interpreter.InputAction)
					i.Data[interpreter.ET_PADDLE_BUTTON_UP] = m1
				}
				m1[index] = interpreter.InputAction{Kind: interpreter.ET_PADDLE_VALUE_CHANGE, ID: tindex}
				m2[index] = interpreter.InputAction{Kind: interpreter.ET_PADDLE_BUTTON_DOWN, ID: tindex}
				m3[index] = interpreter.InputAction{Kind: interpreter.ET_PADDLE_BUTTON_UP, ID: tindex}
			}
		}
	}

	log.Printf("mapping for %s:\n%v\n", name, *i)

	matrix[name] = i

	return nil

}

func InputHandler(c *ducktape.Client, s *server.DuckTapeServer, msg *ducktape.DuckTapeBundle) error {

	user := c.Name

	ie := &interpreter.InputEvent{}
	e := ie.UnmarshalBinary(msg.Payload)
	if e != nil {
		return e
	}

	log.Printf("*** message IEV from %s: %d / %d / %d\n", user, ie.Kind, ie.ID, ie.Value)
	im, ok := matrix[user]
	if ok {
		// matrix exists
		*ie = im.FilterEvent(*ie)
		log.Printf("--> XLATED message IEV from %s: %d / %d / %d\n", user, ie.Kind, ie.ID, ie.Value)
	}

	if ie.Kind != interpreter.ET_NONE {

		switch ie.Kind {
		case interpreter.ET_KEYPRESS:
			log.Printf("Keypress %d\n", ie.Value)
			mutex.Lock()
			RAM.KeyBufferAdd(0, uint(ie.Value))
			mutex.Unlock()
		case interpreter.ET_PADDLE_BUTTON_DOWN:
			mutex.Lock()
			RAM.IntSetPaddleButton(0, int(ie.ID), uint(ie.Value))
			mutex.Unlock()
		case interpreter.ET_PADDLE_BUTTON_UP:
			mutex.Lock()
			RAM.IntSetPaddleButton(0, int(ie.ID), uint(ie.Value))
			mutex.Unlock()
		case interpreter.ET_PADDLE_VALUE_CHANGE:
			mutex.Lock()
			RAM.IntSetPaddleValue(0, int(ie.ID), uint(ie.Value))
			mutex.Unlock()
		}

	}

	return nil

}

// QueryInputHandler is called when a user requests their control mappings
func QueryInputHandler(c *ducktape.Client, s *server.DuckTapeServer, msg *ducktape.DuckTapeBundle) error {

	name := c.Name

	name = strings.Replace(name, "-control", "", -1)
	query := string(msg.Payload)

	log.Println(name)

	// valid props are: p0 -> p3, keys

	result := ""

	switch {
	case query == "p0":
		ie := interpreter.InputEvent{
			Kind: interpreter.ET_PADDLE_BUTTON_DOWN,
			ID:   0,
		}
		im, ok := matrix[name]
		if ok {
			ie = im.FilterEvent(ie)
		}
		if ie.Kind == interpreter.ET_NONE {
			result = "off"
		} else if ie.Kind == interpreter.ET_PADDLE_BUTTON_DOWN {
			result = "p" + utils.IntToStr(int(ie.ID))
		}
	case query == "p1":
		ie := interpreter.InputEvent{
			Kind: interpreter.ET_PADDLE_BUTTON_DOWN,
			ID:   1,
		}
		im, ok := matrix[name]
		if ok {
			ie = im.FilterEvent(ie)
		}
		if ie.Kind == interpreter.ET_NONE {
			result = "off"
		} else if ie.Kind == interpreter.ET_PADDLE_BUTTON_DOWN {
			result = "p" + utils.IntToStr(int(ie.ID))
		}
	case query == "p2":
		ie := interpreter.InputEvent{
			Kind: interpreter.ET_PADDLE_BUTTON_DOWN,
			ID:   2,
		}
		im, ok := matrix[name]
		if ok {
			ie = im.FilterEvent(ie)
		}
		if ie.Kind == interpreter.ET_NONE {
			result = "off"
		} else if ie.Kind == interpreter.ET_PADDLE_BUTTON_DOWN {
			result = "p" + utils.IntToStr(int(ie.ID))
		}
	case query == "p3":
		ie := interpreter.InputEvent{
			Kind: interpreter.ET_PADDLE_BUTTON_DOWN,
			ID:   3,
		}
		im, ok := matrix[name]
		if ok {
			ie = im.FilterEvent(ie)
		}
		if ie.Kind == interpreter.ET_NONE {
			result = "off"
		} else if ie.Kind == interpreter.ET_PADDLE_BUTTON_DOWN {
			result = "p" + utils.IntToStr(int(ie.ID))
		}
	case query == "keys":
		ie := interpreter.InputEvent{
			Kind:  interpreter.ET_KEYPRESS,
			ID:    0,
			Value: 32,
		}
		im, ok := matrix[name]
		if ok {
			ie = im.FilterEvent(ie)
		}
		if ie.Kind == interpreter.ET_NONE {
			result = "off"
		} else {
			result = "on"
		}
	}

	if result == "" {
		c.SendMessage(
			ducktape.DuckTapeBundle{
				ID:      "QIE",
				Payload: []byte("Bad query: " + query),
				Binary:  true,
			},
		)
		return nil
	}

	c.SendMessage(
		ducktape.DuckTapeBundle{
			ID:      "QIS",
			Payload: []byte(result),
			Binary:  true,
		},
	)

	return nil
}

func MemHandler(c *ducktape.Client, s *server.DuckTapeServer, msg *ducktape.DuckTapeBundle) error {

	//_ := int(msg.Payload[0])
	//addr := int(msg.Payload[1])<<16 | int(msg.Payload[2])<<8 | int(msg.Payload[3])
	//value := uint(msg.Payload[4])<<24 | uint(msg.Payload[5])<<16 | uint(msg.Payload[6])<<8 | uint(msg.Payload[7])

	slotid, addr, value, read, _, _ := mempak.Decode(msg.Payload)

	//if addr < memory.OCTALYZER_PADDLE_BASE || addr >= memory.OCTALYZER_PADDLE_BASE +  memory.OCTALYZER_PADDLE_SIZE {
	//	RAM.WriteGlobalSilent(addr, value)
	//}
	user := c.Name

	//fmt.Println("<MEM> Mem event from", c.Name)
	if read {
		return nil
	}

	index := int(slotid)

	redirector, ok := VALID_CONTROL[user]
	if ok {
		raddr, exists := redirector[addr]
		if exists {
			addr = raddr
		}
	}

	if addr != -1 {
		RAM.WriteInterpreterMemorySilent(index, addr, value)
	}

	return nil
}

func STHandler(c *ducktape.Client, s *server.DuckTapeServer, msg *ducktape.DuckTapeBundle) error {

	// goal here is to return a stack trace to the user...
	buffer := make([]byte, 16384)

	count := runtime.Stack(buffer, true)
	ss := string(buffer[0:count])

	lines := strings.Split(ss, "\n")

	for _, l := range lines {

		c.SendMessage(
			ducktape.DuckTapeBundle{
				ID:      "TRC",
				Payload: []byte("trace: " + l),
				Binary:  false,
			},
		)

	}

	return nil
}

func HandleMemThings(index int, address int, value uint) {
	// memory write callback
	HandleRAMUpdate(index, address, value)
}

//~ func Receive(index int, data []memory.MemoryChange) {
//~ changes[index] <- data // queue and continue
//~}

//~ func ChangeSender(slotid int) {

//~ var data []memory.MemoryChange
//~ var count int
//~ var payload []byte
//~ var mc memory.MemoryChange
//~ var i int
//~ var v uint

//~ var lastRecv time.Time

//~ lastRecv = time.Now()

//~ for {

//~ if c[slotid] == nil {
//~ time.Sleep(50 * time.Millisecond)
//~ continue
//~ }

//~ select {
//~ case data = <-changes[slotid]: // this will block until we get something...
//~ lastRecv = time.Now()
//~ default:
//~ if time.Since(lastRecv) > time.Millisecond*50 {
//~ RAM.CheckLog(slotid)
//~ lastRecv = time.Now()
//~ continue
//~ } else {
//~ time.Sleep(1*time.Millisecond)
//~ continue
//~ }
//~ }

//~ count = 0
//~ if len(data) > 0 {

//~ payload = make([]byte, 3)

//~ for _, mc = range data {

//~ if len(mc.Value) == 0 {
//~ payload = append(payload, mempak.Encode(slotid, mc.Global, 0, true)...)
//~ count++
//~ } else {
//~ for i, v = range mc.Value {
//~ payload = append(payload, mempak.Encode(slotid, mc.Global+i, v, false)...)
//~ count++
//~ }
//~ }

//~ }

//~ if count > 0 {
//~ payload[0] = byte((count / 65536) % 256)
//~ payload[1] = byte((count / 256) % 256)
//~ payload[2] = byte(count % 256)
//~ c[slotid].SendMessage("BMU", utils.GZIPBytes(payload), true)
//~ }
//~ }

//~ }
//~ }

func OnChange(s *server.DuckTapeServer, count int, ocount int) {
	fmt.Printf("%d clients left...\n", count)

	counts := s.ChannelCounts()

	for k, v := range counts {
		fmt.Printf("%20s %d\n", k, v)
	}

	for i := 0; i < *max; i++ {

		if counts["memupd"+utils.IntToStr(i)] == 0 {
			if p.GetInterpreters()[i] != nil {
				p.GetInterpreters()[i].StopTheWorld()
				fmt.Println("Suspending", i)
			}
		} else {
			if p.GetInterpreters()[i] != nil {
				p.GetInterpreters()[i].ResumeTheWorld()
				fmt.Println("Resuming", i)
			}
		}

	}
}

func AudioSink(index int, data []float32, rate int) {

	// We need to wait here for the precise amount of gerbils
	//delay := time.Duration(500000*float64(len(data))/float64(rate)) * time.Microsecond
	//time.Sleep(delay)

}
