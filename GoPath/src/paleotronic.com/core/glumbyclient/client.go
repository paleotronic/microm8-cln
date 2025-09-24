package main

import (
	"paleotronic.com/fmt"
	"paleotronic.com/core/types"
	"paleotronic.com/core/vduproto"
	"paleotronic.com/ducktape"
	"paleotronic.com/ducktape/client"
	"paleotronic.com/log"
	"time"
	//	"time"
)

type VDUClient struct {
	Client     *client.DuckTapeClient
	Port       string
	Running    bool
	Address    string
	Connected  bool
	Dispatcher map[string]chan vduproto.VDUServerEvent
}

func NewVDUClient(addr string, port string) *VDUClient {
	if port == "" {
		port = ":10001"
	}
	if addr == "" {
		addr = "localhost"
	}
	this := &VDUClient{Port: port, Address: addr, Connected: false}
	this.Dispatcher = make(map[string]chan vduproto.VDUServerEvent)
	return this
}

func (this *VDUClient) Connect() {
	if this.Connected {
		return
	}

	s := client.NewDuckTapeClient(this.Address, this.Port, "glclient", "tcp")
	err := s.Connect()
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(500 * time.Millisecond)
	this.Connected = true
	this.Client = s
	this.Client.SendMessage("SND", []byte("input-channel"), false)
	this.Client.SendMessage("SUB", []byte("display-channel"), false)

	log.Printf("client: connected to %s:%s\n", s.Host, s.Service)

	go this.StartMessageMuncher()

}

func (this *VDUClient) StartMessageMuncher() {

	go func() {
		for {
			select {
			case msg := <-this.Client.Incoming:
				this.HandleMessages(msg, this.Dispatcher)
			}
		}
	}()

}

// Register a channel to receive certain types of messages on
func (this *VDUClient) RegisterMessageType(msgtype string, ch chan vduproto.VDUServerEvent) {
	this.Dispatcher[msgtype] = ch
}

func (this *VDUClient) HandleMessages(msg ducktape.DuckTapeBundle, dispatcher map[string]chan vduproto.VDUServerEvent) {

	switch {
	case msg.ID == "HGS":
		scanChan, ok := dispatcher[msg.ID]
		if ok {
			var sout vduproto.ScanLineEvent
			_ = sout.UnmarshalBinary(msg.Payload)
			////fmt.Printf("Payload packet for HGS is %d bytes, binary %v\n", len(msg.Payload), msg.Binary)
			scanChan <- vduproto.VDUServerEvent{Kind: msg.ID, Data: sout}
		} else {
			//log.Println("No recipient for this message type!")
		}
	case msg.ID == "SPK":
		spkChan, ok := dispatcher[msg.ID]
		if ok {
			var sout vduproto.ClickEvent
			_ = sout.UnmarshalBinary(msg.Payload)
			spkChan <- vduproto.VDUServerEvent{Kind: msg.ID, Data: sout}
		} else {
			//log.Println("No recipient for this message type!")
		}
	case msg.ID == "SOE":
		strChan, ok := dispatcher[msg.ID]
		if ok {
			var sout vduproto.StringOutEvent
			_ = sout.UnmarshalBinary(msg.Payload)
			strChan <- vduproto.VDUServerEvent{Kind: msg.ID, Data: sout}
		} else {
			//log.Println("No recipient for this message type!")
		}
	case msg.ID == "SME":
		memChan, ok := dispatcher[msg.ID]
		if ok {
			var cout vduproto.ScreenMemoryEvent
			_ = cout.UnmarshalBinary(msg.Payload)
			memChan <- vduproto.VDUServerEvent{Kind: msg.ID, Data: cout}
		} else {
			//log.Println("No recipient for this message type!")
		}
	case msg.ID == "SFE":
		modeChan, ok := dispatcher[msg.ID]
		if ok {
			var cout types.VideoMode
			_ = cout.UnmarshalBinary(msg.Payload)
			modeChan <- vduproto.VDUServerEvent{Kind: msg.ID, Data: cout}
		} else {
			//log.Println("No recipient for this message type!")
		}
	case msg.ID == "SPE":
		posChan, ok := dispatcher[msg.ID]
		if ok {
			var cout vduproto.ScreenPositionEvent
			_ = cout.UnmarshalBinary(msg.Payload)
			posChan <- vduproto.VDUServerEvent{Kind: msg.ID, Data: cout}
		} else {
			//log.Println("No recipient for this message type!")
		}
	case msg.ID == "TSE":
		thinChan, ok := dispatcher[msg.ID]
		if ok {
			var cout vduproto.ThinScreenEventList
			_ = cout.UnmarshalBinary(msg.Payload)
			thinChan <- vduproto.VDUServerEvent{Kind: msg.ID, Data: cout}
		} else {
			//log.Println("No recipient for this message type!")
		}
	}

}

func (this *VDUClient) SendPaddleButton(id byte, state byte) {
	if !this.Connected {
		this.Connect()
	}

	chunk := vduproto.PaddleButtonEvent{PaddleID: id, ButtonState: state}
	b, _ := chunk.MarshalBinary()
	this.Client.SendMessage("PBE", []byte(b), true)

	log.Println("Sent paddle button to client...")
}

func (this *VDUClient) SendPaddleValue(id byte, value byte) {
	if !this.Connected {
		this.Connect()
	}

	chunk := vduproto.PaddleValueEvent{PaddleID: id, PaddleValue: value}
	b, _ := chunk.MarshalBinary()
	this.Client.SendMessage("PVE", []byte(b), true)

	log.Println("Sent paddle value to client...")
}

func (this *VDUClient) SendPaddleModify(id byte, value int8) {
	if !this.Connected {
		this.Connect()
	}

	chunk := vduproto.PaddleModifyEvent{PaddleID: id, Difference: value}
	b, _ := chunk.MarshalBinary()
	this.Client.SendMessage("PME", []byte(b), true)

	log.Println("Sent paddle modify to client...")
}

func (this *VDUClient) SendKeyPress(ch rune) {
	if !this.Connected {
		this.Connect()
	}

	chunk := vduproto.KeyPressEvent{Character: ch}
	b, _ := chunk.MarshalBinary()
	this.Client.SendMessage("KEY", []byte(b), true)

	log.Println("Sent keypress to client...")
}

func (this *VDUClient) SendScreenStateRequest() {
	if !this.Connected {
		this.Connect()
	}

	chunk := vduproto.EmptyClientRequest{}
	b, _ := chunk.MarshalBinary()
	this.Client.SendMessage("SRQ", []byte(b), true)
}

func (this *VDUClient) Close() {
	if !this.Connected {
		return
	}
	this.Client.Close()
}
