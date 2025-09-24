package restalgia

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"time"

	"github.com/ulikunitz/xz"
	"paleotronic.com/core/settings"
	"paleotronic.com/files"
)

//var msTicks uint64
var ticker *time.Ticker
var xzMagic = []byte{0xFD, 0x37, 0x7A, 0x58, 0x5A, 0x00}
var startTime = time.Now()

// func init() {
// 	ticker = time.NewTicker(time.Millisecond)
// 	go func() {
// 		for {
// 			select {
// 			case _ = <-ticker.C:
// 				msTicks++
// 				if msTicks%1000 == 0 {
// 					fmt.Println(time.Now())
// 				}
// 			}
// 		}
// 	}()
// }
func resetClock() {
	startTime = time.Now()
}

func msTicks() uint64 {
	return uint64(time.Since(startTime) / time.Millisecond)
}

type RMachineRecorder struct {
	f        *os.File
	b        *xz.Writer
	running  bool
	start    uint64
	events   chan *RMachineEvent
	filename string
}

type RMachineEvent struct {
	Since  uint32
	Voice  byte
	OpCode byte
	Value  uint64
}

func (e *RMachineEvent) Bytes() []byte {
	return []byte{
		byte(e.Since & 0xff),
		byte((e.Since >> 8) & 0xff),
		byte((e.Since >> 16) & 0xff),
		e.Voice,
		e.OpCode,
		byte((e.Value) & 0xff),
		byte((e.Value >> 8) & 0xff),
		byte((e.Value >> 16) & 0xff),
		byte((e.Value >> 24) & 0xff),
		byte((e.Value >> 32) & 0xff),
		byte((e.Value >> 40) & 0xff),
		byte((e.Value >> 48) & 0xff),
		byte((e.Value >> 56) & 0xff),
	}
}

func (e *RMachineEvent) FromBytes(d []byte) error {
	if len(d) != 13 {
		return fmt.Errorf("Expected 13 bytes, got %d", len(d))
	}
	e.Since = uint32(d[0]) | (uint32(d[1]) << 8) | (uint32(d[2]) << 16)
	e.Voice = d[3]
	e.OpCode = d[4]
	e.Value = uint64(d[5]) |
		(uint64(d[6]) << 8) |
		(uint64(d[7]) << 16) |
		(uint64(d[8]) << 24) |
		(uint64(d[9]) << 32) |
		(uint64(d[10]) << 40) |
		(uint64(d[11]) << 48) |
		(uint64(d[12]) << 56)
	return nil
}

func NewRMachineRecorder() *RMachineRecorder {
	r := &RMachineRecorder{
		start:   msTicks(),
		running: false,
		events:  make(chan *RMachineEvent, 1024),
	}
	return r
}

func (r *RMachineRecorder) LogEvent(voice int, opcode int, value uint64) {
	if !r.running {
		return
	}
	r.events <- &RMachineEvent{
		Since:  r.GetDiffMS(),
		Voice:  byte(voice),
		OpCode: byte(opcode),
		Value:  value,
	}

}

func (r *RMachineRecorder) GetDiffMS() uint32 {
	diff := msTicks() - r.start
	return uint32(diff)
}

func (r *RMachineRecorder) Start(filename string) error {
	r.Stop()

	var err error
	r.f, err = os.Create(filename)
	if err != nil {
		return err
	}
	r.b, _ = xz.NewWriter(r.f)

	//resetClock()
	r.filename = filename
	r.start = msTicks()

	go func() {
		r.running = true
		for r.running {
			select {
			case ev := <-r.events:
				// Process event
				b := ev.Bytes()
				r.b.Write(b)
			default:
				time.Sleep(1 * time.Millisecond)
			}
		}
	}()

	return nil
}

func (r *RMachineRecorder) Stop() {
	if r.running {
		r.running = false
		time.Sleep(50 * time.Millisecond)
		r.b.Close()
		r.f.Close()
	}
}

// -- Player code
type RMachinePlayer struct {
	running  bool
	events   chan *RMachineEvent
	buffer   *bytes.Buffer
	rm       *RMachine
	start    uint64
	filename string
	used     [64]bool
	loop     bool
}

func NewRMachinePlayer(rm *RMachine) *RMachinePlayer {
	p := &RMachinePlayer{
		rm:     rm,
		events: make(chan *RMachineEvent, 1024),
		loop:   true,
	}
	return p
}

func (p *RMachinePlayer) Silence() {
	for i, v := range p.used {
		if v {
			// zero voice
			p.rm.ExecuteOpcode(i, 0x81, math.Float64bits(0))
			p.used[i] = false
		}
	}
}

func (p *RMachinePlayer) Stop() {
	if p.running {
		p.running = false
		time.Sleep(50 * time.Millisecond)
		p.Silence()
	}
}

func (p *RMachinePlayer) Start(filename string) error {
	p.Stop()

	for i, _ := range p.used {
		p.used[i] = false
	}

	fr, err := files.ReadBytesViaProvider(files.GetPath(filename), files.GetFilename(filename))
	if err != nil {
		return err
	}

	// check for xz
	var isXZ = true
	for i, v := range xzMagic {
		if v != fr.Content[i] {
			isXZ = false
			break
		}
	}
	if isXZ {
		tmpr, _ := xz.NewReader(bytes.NewBuffer(fr.Content))
		fr.Content, _ = ioutil.ReadAll(tmpr)
	}

	p.buffer = bytes.NewBuffer(fr.Content)

	p.running = true

	// buffer reader
	go func() {
		var chunk [13]byte
		n, err := p.buffer.Read(chunk[:])
		for p.running && n == 13 && err == nil {
			ev := &RMachineEvent{}
			err = ev.FromBytes(chunk[:])
			p.events <- ev

			n, err = p.buffer.Read(chunk[:])

			if err != nil && err.Error() == "EOF" {
				//fmt.Println("TIME TO LOOP")
				p.buffer = bytes.NewBuffer(fr.Content)
				n, err = p.buffer.Read(chunk[:])
				p.events <- &RMachineEvent{
					OpCode: 0xff,
					Since:  0,
				}
			}
		}
	}()

	// player
	go func() {
		time.Sleep(time.Duration(settings.MusicLeadin[p.rm.SlotId]) * time.Millisecond)
		p.start = msTicks()
		for p.running && p.rm != nil {
			select {
			case ev := <-p.events:
				rel := msTicks() - p.start
				for uint32(rel) < ev.Since {
					time.Sleep(500 * time.Microsecond)
					rel = msTicks() - p.start
				}
				if ev.OpCode == 0xff {
					p.start = msTicks()
				} else {
					p.rm.ExecuteOpcode(int(ev.Voice), int(ev.OpCode), ev.Value)
					p.used[int(ev.Voice)] = true
				}
			default:
				time.Sleep(1 * time.Millisecond)
			}
		}
	}()

	return nil
}
