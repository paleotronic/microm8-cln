package buzzer

import (
	"paleotronic.com/log"
	"sync"
	"time"
)

const (
	TimingWindow = 1000 // uS = 1ms
	MaxToggles   = 45
	MaxPacks     = 10
)

type BuzzerPacket struct {
	Toggles      []byte
	Needle       int
	m            sync.Mutex
	ch           chan []byte
	LastTouched  int64
	SentTime     int64
	DispatchTime int64
}

var PacketChan chan []byte = make(chan []byte)
var PacketBuffer [10]*BuzzerPacket = [10]*BuzzerPacket{
	NewBuzzerPacket(PacketChan),
	NewBuzzerPacket(PacketChan),
	NewBuzzerPacket(PacketChan),
	NewBuzzerPacket(PacketChan),
	NewBuzzerPacket(PacketChan),
	NewBuzzerPacket(PacketChan),
	NewBuzzerPacket(PacketChan),
	NewBuzzerPacket(PacketChan),
	NewBuzzerPacket(PacketChan),
	NewBuzzerPacket(PacketChan),
}
var WritePacket int
var LastPacketSent int

var Packet *BuzzerPacket = NewBuzzerPacket(PacketChan)
var StartTime int64 = time.Now().UnixNano()

// NewBuzzerPacket creates a buzzer packet
func NewBuzzerPacket(ch chan []byte) *BuzzerPacket {
	this := &BuzzerPacket{Toggles: make([]byte, MaxToggles), ch: ch}
	return this
}

func (this *BuzzerPacket) Pluck() {

	this.m.Lock()
	defer this.m.Unlock()

	v := time.Now().UnixNano()

	diff := v - StartTime
	this.Needle = int(diff%1000000) / 22800 // uS

	this.Toggles[this.Needle] = 1
	this.LastTouched = v
}

func (this *BuzzerPacket) SendBlock() {
	this.m.Lock()
	defer this.m.Unlock()

	var empty bool = true
	for _, v := range this.Toggles {
		if v != 0 {
			empty = false
			break
		}
	}

	if this.LastTouched != 0 {
		log.Println("Last pluck to send", (time.Now().UnixNano()-this.LastTouched)/1000000, "ms")
	}

	if !empty {
		this.ch <- this.Toggles
	}
	this.Toggles = make([]byte, MaxToggles)
	this.Needle = 0
	this.LastTouched = 0
}

func Pluck() {
	diff := time.Now().UnixNano() - StartTime
	WritePacket = int((diff / 1000000) % 10) // 10ms increment

	PacketBuffer[WritePacket].Pluck()
}
