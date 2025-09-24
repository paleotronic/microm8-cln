package main

import (
	"paleotronic.com/core/vduproto"
	"paleotronic.com/log"
	"paleotronic.com/postoffice/client"
	"time"
)

type VDUClient struct {
	Client     *mqclient.MQClient
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

	s, err := mqclient.New(this.Address, this.Port)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("client: connected to %s:%s\n", s.Host, s.Port)

	this.Connected = true
	this.Client = s

	go this.StartMessageMuncher()

}

func (this *VDUClient) StartMessageMuncher() {

	go func() {
		for {
			this.HandleMessages(this.Dispatcher)
			time.Sleep(5 * time.Millisecond)
		}
	}()

}

// Register a channel to receive certain types of messages on
func (this *VDUClient) RegisterMessageType(msgtype string, ch chan vduproto.VDUServerEvent) {
	this.Dispatcher[msgtype] = ch
}

func (this *VDUClient) HandleMessages(dispatcher map[string]chan vduproto.VDUServerEvent) {
	messages, err := this.Client.Lease("display-channel", "", 5, 1)

	if err != nil {
		return
	}

	for _, msg := range messages {

		//log.Printf("--> Got message with type %s\n", msg.Tag)

		switch {
		case msg.Tag == "StringOutEvent":
			strChan, ok := dispatcher[msg.Tag]
			if ok {
				var sout vduproto.StringOutEvent
				err = mqclient.Unwrap(msg.Payload, &sout)
				strChan <- vduproto.VDUServerEvent{Kind: msg.Tag, Data: sout}
			} else {
				//log.Println("No recipient for this message type!")
			}
		case msg.Tag == "CharOutEvent":
			charChan, ok := dispatcher[msg.Tag]
			if ok {
				var cout vduproto.CharOutEvent
				err = mqclient.Unwrap(msg.Payload, &cout)
				charChan <- vduproto.VDUServerEvent{Kind: msg.Tag, Data: cout}
			} else {
				//log.Println("No recipient for this message type!")
			}
		case msg.Tag == "ScreenMemoryEvent":
			memChan, ok := dispatcher[msg.Tag]
			if ok {
				var cout vduproto.ScreenMemoryEvent
				err = mqclient.Unwrap(msg.Payload, &cout)
				memChan <- vduproto.VDUServerEvent{Kind: msg.Tag, Data: cout}
			} else {
				//log.Println("No recipient for this message type!")
			}
		case msg.Tag == "ScreenFormatEvent":
			modeChan, ok := dispatcher[msg.Tag]
			if ok {
				var cout vduproto.ScreenFormatEvent
				err = mqclient.Unwrap(msg.Payload, &cout)
				modeChan <- vduproto.VDUServerEvent{Kind: msg.Tag, Data: cout}
			} else {
				//log.Println("No recipient for this message type!")
			}
		case msg.Tag == "ScreenPositionEvent":
			posChan, ok := dispatcher[msg.Tag]
			if ok {
				var cout vduproto.ScreenPositionEvent
				err = mqclient.Unwrap(msg.Payload, &cout)
				posChan <- vduproto.VDUServerEvent{Kind: msg.Tag, Data: cout}
			} else {
				//log.Println("No recipient for this message type!")
			}
		}
	}

	// remove messages that we just chewed through
	err = this.Client.Remove("display-channel", messages)
}

func (this *VDUClient) SendKeyPress(ch rune) {
	if !this.Connected {
		this.Connect()
	}

	err := this.Client.Put("input-channel", mqclient.Wrap(vduproto.KeyPressEvent{Character: ch}), "KeyPressEvent", "")
	if err != nil {
		//
	}

	log.Println("Sent keypress to client...")
}

func (this *VDUClient) SendScreenStateRequest() {
	if !this.Connected {
		this.Connect()
	}

	err := this.Client.Put("input-channel", mqclient.Wrap(vduproto.EmptyClientRequest{}), "ScreenStateRequest", "")
	if err != nil {
		//
	}
}

func (this *VDUClient) Close() {
	if !this.Connected {
		return
	}
	this.Client.Close()
}
