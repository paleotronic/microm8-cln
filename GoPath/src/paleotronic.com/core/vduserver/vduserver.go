package vduserver

import (
//	"paleotronic.com/fmt"
	"time"

	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/core/vduproto"
	"paleotronic.com/ducktape"
	"paleotronic.com/ducktape/client"
	"paleotronic.com/log"
)

type VDUServer struct {
	Client     *client.DuckTapeClient
	Port       string
	Running    bool
	Address    string
	Display    interfaces.Display
	Dispatcher map[string]chan vduproto.VDUServerEvent
}

func NewVDUServer(port string, bind string, d interfaces.Display) *VDUServer {
	if port == "" {
		port = ":10001"
	}
	if bind == "" {
		bind = "localhost"
	}
	this := &VDUServer{Port: port, Address: bind, Running: false, Display: d}
	this.Dispatcher = make(map[string]chan vduproto.VDUServerEvent)
	return this
}

// Register a channel to receive certain types of messages on
func (this *VDUServer) RegisterMessageType(msgtype string, ch chan vduproto.VDUServerEvent) {
	this.Dispatcher[msgtype] = ch
}

func (this *VDUServer) HandleMessage(msg ducktape.DuckTapeBundle, dispatcher map[string]chan vduproto.VDUServerEvent) {

	log.Printf("VDU Server in HandleMessage() for %v\n", msg)

	switch {
	case msg.ID == "PBE":
		paddleChan, ok := dispatcher[msg.ID]
		if ok {
			var sout vduproto.PaddleButtonEvent
			_ = sout.UnmarshalBinary(msg.Payload)
			paddleChan <- vduproto.VDUServerEvent{Kind: msg.ID, Data: sout}
			log.Println("Handled PBE")
		} else {
			//log.Println("No recipient for this message type!")
		}
	case msg.ID == "PVE":
		paddleChan, ok := dispatcher[msg.ID]
		if ok {
			var sout vduproto.PaddleValueEvent
			_ = sout.UnmarshalBinary(msg.Payload)
			paddleChan <- vduproto.VDUServerEvent{Kind: msg.ID, Data: sout}
			log.Println("Handled PVE")
		} else {
			//log.Println("No recipient for this message type!")
		}
	case msg.ID == "PME":
		paddleChan, ok := dispatcher[msg.ID]
		if ok {
			var sout vduproto.PaddleModifyEvent
			_ = sout.UnmarshalBinary(msg.Payload)
			paddleChan <- vduproto.VDUServerEvent{Kind: msg.ID, Data: sout}
			log.Println("Handled PME")
		} else {
			//log.Println("No recipient for this message type!")
		}
	case msg.ID == "KEY":
		keyChan, ok := dispatcher[msg.ID]
		if ok {
			var sout vduproto.KeyPressEvent
			_ = sout.UnmarshalBinary(msg.Payload)
			keyChan <- vduproto.VDUServerEvent{Kind: msg.ID, Data: sout}
			log.Println("Handled KEY")
		} else {
			//log.Println("No recipient for this message type!")
		}
	case msg.ID == "YLD":
		cpuChan, ok := dispatcher[msg.ID]
		if ok {
			var sout vduproto.EmptyClientRequest
			_ = sout.UnmarshalBinary(msg.Payload)
			cpuChan <- vduproto.VDUServerEvent{Kind: msg.ID, Data: sout}
		} else {
			//log.Println("No recipient for this message type!")
		}
	case msg.ID == "SRQ":
		stateChan, ok := dispatcher[msg.ID]
		if ok {
			var sout vduproto.EmptyClientRequest
			_ = sout.UnmarshalBinary(msg.Payload)
			stateChan <- vduproto.VDUServerEvent{Kind: msg.ID, Data: sout}
		} else {
			//log.Println("No recipient for this message type!")
		}
	case msg.ID == "AQT":
		assChan, ok := dispatcher[msg.ID]
		if ok {
			var sout vduproto.AssetQuery
			_ = sout.UnmarshalBinary(msg.Payload)
			assChan <- vduproto.VDUServerEvent{Kind: msg.ID, Data: sout}
		} else {
			//log.Println("No recipient for this message type!")
		}
	case msg.ID == "AQF":
		assChan, ok := dispatcher[msg.ID]
		if ok {
			var sout vduproto.AssetQuery
			_ = sout.UnmarshalBinary(msg.Payload)
			assChan <- vduproto.VDUServerEvent{Kind: msg.ID, Data: sout}
		} else {
			//log.Println("No recipient for this message type!")
		}
	}

}

func (this *VDUServer) Run() {

	s := client.NewDuckTapeClient(this.Address, this.Port, "backend", "tcp")
	err := s.Connect()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Started VDU backend service on %s:%s\n", s.Host, s.Service)
	this.Client = s
	this.Running = true

	time.Sleep(500 * time.Millisecond)
	this.Client.SendMessage("SND", []byte("display-channel"), false)
	this.Client.SendMessage("SUB", []byte("input-channel"), false)

	go this.Watch()

	for this.Client.Connected {
		select {
		case msg := <-this.Client.Incoming:
			this.HandleMessage(msg, this.Dispatcher)
		}
	}
}

func (this *VDUServer) Watch() {

	for {

		offset, data := this.Display.GetShadowTextMemory().GetChangedData()
		if len(data) > 0 {
			// update
			this.SendScreenMemoryChange(offset, data, this.Display.GetCursorX(), this.Display.GetCursorY())
		}

		time.Sleep(10 * time.Millisecond)

	}
}

func (this *VDUServer) StringOut(s string) {

	if this.Client == nil {
		log.Fatal("There is no connected VDU client to send to :(")
	}

	//err := this.Client.Put("display-channel", mqclient.Wrap(&vduproto.StringOutEvent{Content: s, X: 0, Y: 0, Attribute: 'N'}), "StringOutEvent", "")

	chunk := vduproto.StringOutEvent{Content: s, X: 0, Y: 0, Attribute: 'N'}
	b, err := chunk.MarshalBinary()
	this.Client.SendMessage("SOE", b, true)

	log.Println("Sending string:", s)

	if err != nil {
		log.Println(err)
	}
}

func (this *VDUServer) SendScreenMemoryChange(offset int, data []uint, nx int, ny int) {
	if this.Client == nil {
		log.Fatal("There is no connected VDU client to send to :(")
	}

	if len(data) == 0 {
		log.Println("Dropping zero length message - no change")
		return
	}

	//err := this.Client.Put("display-channel", mqclient.Wrap(&vduproto.ScreenMemoryEvent{Content: data, X: nx, Y: ny, Offset: offset}), "ScreenMemoryEvent", "")

	chunk := vduproto.ScreenMemoryEvent{Content: data, X: nx, Y: ny, Offset: offset}
	b, err := chunk.MarshalBinary()
	this.Client.SendMessage("SME", b, true)

	log.Printf("Sending screen memory @%d (%d long):", offset, len(data))

	if err != nil {
		log.Println(err)
	}
}

func (this *VDUServer) SendScreenSpecification(mode types.VideoMode) {
	if this.Client == nil {
		log.Fatal("There is no connected VDU client to send to :(")
	}

	//err := this.Client.Put("display-channel", mqclient.Wrap(&mode), "ScreenFormatEvent", "")

	b, err := mode.MarshalBinary()
	this.Client.SendMessage("SFE", b, true)

	if err != nil {
		log.Println(err)
	}
}

func (this *VDUServer) SendScreenPositionUpdate(x, y int) {
	if this.Client == nil {
		log.Fatal("There is no connected VDU client to send to :(")
	}

	//	err := this.Client.Put("display-channel", mqclient.Wrap(&vduproto.ScreenPositionEvent{X: x, Y: y}), "ScreenPositionEvent", "")

	chunk := vduproto.ScreenPositionEvent{X: x, Y: y}
	b, err := chunk.MarshalBinary()
	this.Client.SendMessage("SPE", b, true)

	if err != nil {
		log.Println(err)
	}
}

func (this *VDUServer) SendThinScreenMessages(evlist vduproto.ThinScreenEventList) {

	if this.Client == nil {
		log.Fatal("There is no connected VDU client to send to :(")
	}

	log.Printf("Sending %d thin events\n", len(evlist))

	//err := this.Client.Put("display-channel", mqclient.Wrap(&evlist), "ThinScreenEvent", "")

	b, err := evlist.MarshalBinary()
	this.Client.SendMessage("TSE", b, true)

	if err != nil {
		log.Println(err)
	}
}

func (this *VDUServer) SendScanLine(i int, bytes []byte) {

	if this.Client == nil {
		log.Fatal("There is no connected VDU client to send to :(")
	}

	//err := this.Client.Put("display-channel", mqclient.Wrap(&evlist), "ThinScreenEvent", "")

	ce := vduproto.ScanLineEvent{Y: byte(i), Data: bytes}

	b, err := ce.MarshalBinary()
	////fmt.Printf("Sending HGS packet %d bytes\n", len(b))

	this.Client.SendMessage("HGS", b, true)

	if err != nil {
		log.Println(err)
	}
}

func (this *VDUServer) SendMemoryUpdate(addr, value int) {
	if this.Client == nil {
		log.Fatal("There is no connected VDU client to send to :(")
	}

	//err := this.Client.Put("display-channel", mqclient.Wrap(&evlist), "ThinScreenEvent", "")

	ce := vduproto.MemEvent{AL: byte(addr % 256), AH: byte(addr / 256), Value: byte(value & 0xff)}

	b, err := ce.MarshalBinary()

	this.Client.SendMessage("MEM", b, true)

	if err != nil {
		log.Println(err)
	}
}

func (this *VDUServer) SendBGColorUpdate(rr, gg, bb, aa byte) {
	if this.Client == nil {
		log.Fatal("There is no connected VDU client to send to :(")
	}

	//err := this.Client.Put("display-channel", mqclient.Wrap(&evlist), "ThinScreenEvent", "")

	ce := vduproto.ColorEvent{R: rr, G: gg, B: bb, A: aa}

	b, err := ce.MarshalBinary()

	this.Client.SendMessage("BGC", b, true)

	if err != nil {
		log.Println(err)
	}
}

func (this *VDUServer) SendCPUEvent(addr int) {
	if this.Client == nil {
		log.Fatal("There is no connected VDU client to send to :(")
	}

	//err := this.Client.Put("display-channel", mqclient.Wrap(&evlist), "ThinScreenEvent", "")

	ce := vduproto.CPUEvent{AL: byte(addr % 256), AH: byte(addr / 256)}

	b, err := ce.MarshalBinary()

	this.Client.SendMessage("CPU", b, true)

	if err != nil {
		log.Println(err)
	}
}

func (this *VDUServer) SendRestalgiaEvent(kind byte, content string) {
	if this.Client == nil {
		log.Fatal("There is no connected VDU client to send to :(")
	}

	ce := vduproto.RestalgiaCommand{SubType: kind, Data: []byte(content)}

	b, err := ce.MarshalBinary()

	this.Client.SendMessage("RST", b, true)

	if err != nil {
		log.Println(err)
	}
}

func (this *VDUServer) SendConnectCommand(content string) {
	if this.Client == nil {
		log.Fatal("There is no connected VDU client to send to :(")
	}

	ce := vduproto.ConnectCommand{Data: []byte(content)}

	b, err := ce.MarshalBinary()

	this.Client.SendMessage("RCN", b, true)

	if err != nil {
		log.Println(err)
	}
}

func (this *VDUServer) SendAssetQuery(aq *vduproto.AssetQuery) {
	if this.Client == nil {
		log.Fatal("There is no connected VDU client to send to :(")
	}

	//err := this.Client.Put("display-channel", mqclient.Wrap(&evlist), "ThinScreenEvent", "")

	b, err := aq.MarshalBinary()

	this.Client.SendMessage("AQY", b, true)

	if err != nil {
		log.Println(err)
	}
}

func (this *VDUServer) SendAssetBlock(ab *vduproto.AssetBlock) {
	if this.Client == nil {
		log.Fatal("There is no connected VDU client to send to :(")
	}

	//err := this.Client.Put("display-channel", mqclient.Wrap(&evlist), "ThinScreenEvent", "")

	b, err := ab.MarshalBinary()

	this.Client.SendMessage("ABK", b, true)

	if err != nil {
		log.Println(err)
	}
}

func (this *VDUServer) SendAssetPlayback(ab *vduproto.AssetAction) {
	if this.Client == nil {
		log.Fatal("There is no connected VDU client to send to :(")
	}

	//err := this.Client.Put("display-channel", mqclient.Wrap(&evlist), "ThinScreenEvent", "")

	b, err := ab.MarshalBinary()

	this.Client.SendMessage("ASA", b, true)

	if err != nil {
		log.Println(err)
	}
}

func (this *VDUServer) SendLayerSpec(ab vduproto.LayerSpec) {
	if this.Client == nil {
		log.Fatal("There is no connected VDU client to send to :(")
	}

	b, _ := ab.MarshalBinary()

	this.Client.SendMessage("LSD", b, true)
}

func (this *VDUServer) SendLayerBundle(ab vduproto.LayerBundle) {
	if this.Client == nil {
		log.Fatal("There is no connected VDU client to send to :(")
	}

	b, _ := ab.MarshalBinary()

	this.Client.SendMessage("LSB", b, true)
}
