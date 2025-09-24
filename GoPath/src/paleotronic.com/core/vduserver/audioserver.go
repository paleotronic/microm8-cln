package vduserver

import (
	"time"

	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/vduproto"
	"paleotronic.com/ducktape"
	"paleotronic.com/ducktape/client"
	"paleotronic.com/log"
)

type AudioServer struct {
	Client     *client.DuckTapeClient
	Port       string
	Running    bool
	Address    string
	Display    interfaces.Display
	Dispatcher map[string]chan vduproto.VDUServerEvent
}

func NewAudioServer(port string, bind string, d interfaces.Display) *AudioServer {
	if port == "" {
		port = ":10001"
	}
	if bind == "" {
		bind = "localhost"
	}
	this := &AudioServer{Port: port, Address: bind, Running: false, Display: d}
	this.Dispatcher = make(map[string]chan vduproto.VDUServerEvent)
	return this
}

// Register a channel to receive certain types of messages on
func (this *AudioServer) RegisterMessageType(msgtype string, ch chan vduproto.VDUServerEvent) {
	this.Dispatcher[msgtype] = ch
}

func (this *AudioServer) HandleMessage(msg ducktape.DuckTapeBundle, dispatcher map[string]chan vduproto.VDUServerEvent) {

	log.Printf("VDU Server in HandleMessage() for %v\n", msg)

}

func (this *AudioServer) Run() {

	s := client.NewDuckTapeClient(this.Address, this.Port, "backend-audio", "tcp")
	err := s.Connect()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Started VDU backend service on %s:%s\n", s.Host, s.Service)
	this.Client = s
	this.Running = true

	time.Sleep(500 * time.Millisecond)
	this.Client.SendMessage("SND", []byte("audio-channel"), false)
	this.Client.SendMessage("SUB", []byte("dummy-channel"), false)

	for this.Client.Connected {
		select {
		case msg := <-this.Client.Incoming:
			this.HandleMessage(msg, this.Dispatcher)
		}
	}
}

func (this *AudioServer) SendSpeakerClickEvent(usefreq int) {
	if this.Client == nil {
		log.Fatal("There is no connected VDU client to send to :(")
	}

	//	err := this.Client.Put("display-channel", mqclient.Wrap(&vduproto.ScreenPositionEvent{X: x, Y: y}), "ScreenPositionEvent", "")

	chunk := &vduproto.CPUEvent{AL: byte(usefreq % 256), AH: byte(usefreq / 256)}

	b, err := chunk.MarshalBinary()
	this.Client.SendMessage("SPK", b, true)

	if err != nil {
		log.Println(err)
	}
}
