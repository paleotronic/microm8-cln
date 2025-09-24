package main

import (
	"time"

	"paleotronic.com/core/vduproto"
	"paleotronic.com/ducktape"
	"paleotronic.com/ducktape/client"
	"paleotronic.com/log"
	//	"time"
)

type AudioClient struct {
	Client     *client.DuckTapeClient
	Port       string
	Running    bool
	Address    string
	Connected  bool
	Dispatcher map[string]chan vduproto.VDUServerEvent
}

func NewAudioClient(addr string, port string) *AudioClient {
	if port == "" {
		port = ":10001"
	}
	if addr == "" {
		addr = "localhost"
	}
	this := &AudioClient{Port: port, Address: addr, Connected: false}
	this.Dispatcher = make(map[string]chan vduproto.VDUServerEvent)
	return this
}

func (this *AudioClient) Connect() {
	if this.Connected {
		return
	}

	s := client.NewDuckTapeClient(this.Address, this.Port, "audioclient", "udp")
	err := s.Connect()
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(500 * time.Millisecond)
	this.Connected = true
	this.Client = s
	this.Client.SendMessage("SND", []byte("dummy-channel"), false)
	this.Client.SendMessage("SUB", []byte("audio-channel"), false)

	log.Printf("client: connected to %s:%s\n", s.Host, s.Service)

	go this.StartMessageMuncher()

}

func (this *AudioClient) StartMessageMuncher() {

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
func (this *AudioClient) RegisterMessageType(msgtype string, ch chan vduproto.VDUServerEvent) {
	this.Dispatcher[msgtype] = ch
}

func (this *AudioClient) HandleMessages(msg ducktape.DuckTapeBundle, dispatcher map[string]chan vduproto.VDUServerEvent) {

	switch {
	case msg.ID == "SPK":
		spkChan, ok := dispatcher[msg.ID]
		if ok {
			var sout vduproto.ClickEvent
			_ = sout.UnmarshalBinary(msg.Payload)
			spkChan <- vduproto.VDUServerEvent{Kind: msg.ID, Data: sout}
		} else {
			//log.Println("No recipient for this message type!")
		}
	}

}
