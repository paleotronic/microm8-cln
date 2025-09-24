package interpreter

import (
	"bytes"
	"errors"
	log2 "log"
	"math"
	"sync"
	"time"

	"paleotronic.com/core/editor"
	"paleotronic.com/core/hardware/apple2"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/hardware/servicebus"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/debugger/debugtypes" //	"io"
	"paleotronic.com/ducktape"
	"paleotronic.com/encoding/mempak"
	"paleotronic.com/filerecord"
	"paleotronic.com/files"
	"paleotronic.com/fmt"
	"paleotronic.com/freeze"
	"paleotronic.com/log"
	"paleotronic.com/octalyzer/bus"
	"paleotronic.com/utils"
)

const HGR_MONITOR = 8192 + 85

type PlayerPos struct {
	Filename string
	Position int
}

type PlayerState struct {
	current      filerecord.FileRecord
	filenum      int
	backwards    bool
	fileptr      int
	Pos          int
	timeFactor   float64
	cpuSyncCount int
	debugSeek    bool
	lastSeekAddr int
	seekDelta    int
}

type Player struct {
	PlayerState
	Destination         *Interpreter
	running             bool
	start               time.Time
	pathname            string
	files               []files.FileDef
	startMS             int
	endMS               int
	currentMS           int
	sliceMode           bool
	r                   *Recorder
	rtNextSync          bool
	rtNextSyncRolloff   bool
	stopNextVSync       bool
	usemem              bool
	exitmode            interfaces.PlayerExitMode
	memblocks           []*bytes.Buffer
	paused              bool
	noResume            bool
	injectedBusRequests []*servicebus.ServiceBusRequest
	lastFile            bool
	m                   sync.Mutex
	pauseOnFUL          bool
	active              bool
}

func NewPlayer(dest *Interpreter, source string, blocks []*bytes.Buffer, backwards bool, backJumpMS int) (*Player, error) {

	rate := settings.GetRewindSpeeds()[0]

	this := &Player{
		Destination: dest,
		pathname:    source,
		running:     true,
		usemem:      (source == ""),
		memblocks:   blocks,
		PlayerState: PlayerState{
			backwards:  backwards,
			timeFactor: rate,
			seekDelta:  backJumpMS * 1000,
		},
		injectedBusRequests: make([]*servicebus.ServiceBusRequest, 0),
	}

	//mr, ok := this.Destination.Memory.InterpreterMappableAtAddress(this.Destination.MemIndex, 0xc000)
	//if ok {
	//	servicebus.Subscribe(this.Destination.MemIndex, servicebus.UnifiedPlaybackSync, mr.(*apple2.Apple2IOChip))
	//}

	return this, this.Reset(source, blocks)

}

func (p *Player) ResetToStart() error {
	p.backwards = false
	p.timeFactor = 0
	p.debugSeek = false
	return p.Reset(p.pathname, p.memblocks)
}

func (this *Player) Reset(source string, blocks []*bytes.Buffer) error {
	if !this.usemem {
		_, f, e := files.ReadDirViaProvider(source+"/", "*.s")
		if e != nil {
			fmt.Println(e)
			//rlog.Printf("Result of openining file recording = %v", e)
			return e
		}
		//rlog.Printf("Got %d files...", len(f))
		this.files = f
		this.filenum = -1
		if this.backwards {
			this.filenum = len(f)
		}
	} else {
		this.filenum = len(blocks)
	}
	this.fileptr = 0

	//fmt.Println(f)
	var e error

	if this.backwards {
		e = this.LoadPrevFile()
	} else {
		e = this.LoadNextFile()
	}

	if e != nil {
		fmt.Println(e)
		return e
	}

	// d := this.CreateRefTable()
	// fmt.Println(d, len(d))

	this.exitmode = interfaces.PEM_NONE

	this.ShowState()

	return nil
}

func (p *Player) SetNoResume(b bool) {
	p.noResume = b
}

func (p *Player) GetTimeShift() float64 {
	return p.timeFactor
}

func (p *Player) SetTimeShift(f float64) {
	p.timeFactor = f
	p.ShowState()
}

func fIndexOf(f float64, list []float64) int {
	for i, v := range list {
		if v == f {
			return i
		}
	}
	return -1
}

func (p *Player) Faster() {

	speeds := settings.GetRewindSpeeds()
	i := fIndexOf(p.timeFactor, speeds)

	if i < len(speeds)-1 {
		i++
	}

	p.timeFactor = speeds[i]

	blocknum := p.filenum
	blocks := len(p.memblocks)
	if len(p.files) != 0 {
		blocks = len(p.files)
	}

	servicebus.SendServiceBusMessage(
		p.Destination.GetMemIndex(),
		servicebus.LiveRewindStateUpdate,
		&debugtypes.LiveRewindState{
			Enabled:     p.IsPlaying(),
			CanBack:     true,
			CanForward:  true,
			CanResume:   true,
			Backwards:   p.backwards,
			TimeFactor:  p.timeFactor,
			Block:       blocknum,
			TotalBlocks: blocks,
		},
	)

	p.ShowState()
}

func (p *Player) Slower() {
	speeds := settings.GetRewindSpeeds()
	i := fIndexOf(p.timeFactor, speeds)

	if i > 0 {
		i--
		p.timeFactor = speeds[i]
	} else {
		p.timeFactor = 0
	}

	blocknum := p.filenum
	blocks := len(p.memblocks)
	if len(p.files) != 0 {
		blocks = len(p.files)
	}

	servicebus.SendServiceBusMessage(
		p.Destination.GetMemIndex(),
		servicebus.LiveRewindStateUpdate,
		&debugtypes.LiveRewindState{
			Enabled:     p.IsPlaying(),
			CanBack:     true,
			CanForward:  true,
			CanResume:   true,
			Backwards:   p.backwards,
			TimeFactor:  p.timeFactor,
			Block:       blocknum,
			TotalBlocks: blocks,
		},
	)

	p.ShowState()
}

func (p *Player) Reverse(ms int) {
	p.backwards = true
	var deltaNeeded = ms * 1000
	delta, _, err := p.Prev()
	for err == nil && deltaNeeded > 0 {
		deltaNeeded -= delta
		delta, _, err = p.Prev()
	}
	p.RealTimeStop()
}

func (p *Player) Pause() {
	p.paused = true
}

func (p *Player) IsPaused() bool {
	return p.paused
}

func (p *Player) Resume() {
	p.paused = false
}

func (p *Player) LoadNextFile() error {
	p.filenum++

	if p.usemem {
		if p.filenum >= len(p.memblocks) {
			//p.running = false
			return errors.New("EOF")
		}
	} else {
		if p.filenum >= len(p.files) {
			//p.running = false
			return errors.New("EOF")
		}
	}

	var data filerecord.FileRecord
	var e error

	if p.usemem {
		fmt.Printf("---> Buffering from memblock %d\n", p.filenum)
		data.Content = p.memblocks[p.filenum].Bytes()
		data.ContentSize = len(data.Content)
	} else {
		fmt.Printf("---> Buffering from %s/%s\n", p.pathname, p.files[p.filenum].Name+"."+p.files[p.filenum].Extension)

		data, e = files.ReadBytesViaProvider(p.pathname, p.files[p.filenum].Name+"."+p.files[p.filenum].Extension)
		if e != nil {
			return e
		}
	}

	p.current = data
	p.fileptr = 0

	return nil
}

func (p *Player) LoadPrevFile() error {
	p.filenum--

	if p.filenum < 0 {
		//p.running = false
		return errors.New("EOF BEGIN")
	}

	var data filerecord.FileRecord
	var e error

	if p.usemem {
		data.Content = p.memblocks[p.filenum].Bytes()
		data.ContentSize = len(data.Content)
	} else {

		if p.filenum < 0 || p.filenum >= len(p.files) {
			//p.running = false
			return errors.New("EOF")
		}

		fmt.Printf("---> Buffering from %s/%s\n", p.pathname, p.files[p.filenum].Name+"."+p.files[p.filenum].Extension)

		data, e = files.ReadBytesViaProvider(p.pathname, p.files[p.filenum].Name+"."+p.files[p.filenum].Extension)
		if e != nil {
			return e
		}
	}

	p.current = data
	p.current.PositionRead = len(p.current.Content)
	p.fileptr = len(p.current.Content)

	// file is last file
	p.lastFile = p.filenum == len(p.files)-1

	return nil
}

func (p *Player) extractAddressFromSync(msg *ducktape.DuckTapeBundle) int {
	if msg == nil {
		return -1
	}
	switch msg.ID {
	case "CPD":
		delta := &CPUDelta{}
		delta.FromBytes(msg.Payload)
		if delta.PC != nil {
			return int(*delta.PC)
		}
		return -1
	case "CPU":
		return (int(msg.Payload[0]) << 24) | (int(msg.Payload[1]) << 16) | (int(msg.Payload[2]) << 8) | int(msg.Payload[3])
	}
	return -1
}

func (p *Player) GetLastNSyncs(count int, current int) []int {

	if p.filenum < 0 {
		return []int{0}
	}
	if p.usemem && p.filenum >= len(p.memblocks) {
		return []int{0}
	}
	if !p.usemem && p.filenum >= len(p.files) {
		return []int{0}
	}

	// 1. save current state
	ostate := p.PlayerState
	defer func() {
		p.PlayerState = ostate
	}()

	var needed = count
	var bundle *ducktape.DuckTapeBundle
	var err error

	var out = make([]int, count)

	//var s = time.Now()
	var addr, lastAddr int

	lastAddr = current

	for needed > 0 {
		// go back 1 packet
		_, bundle, err = p.Prev()
		if err == nil {
			// got packet
			if bundle.ID == "CPU" || bundle.ID == "CPD" {
				addr = p.extractAddressFromSync(bundle)
				// skip repeats from x-boundary sync packets
				if addr != lastAddr {
					out[needed-1] = addr
					needed--
					lastAddr = addr
				}
			}
		} else {
			break
		}
	}

	//d := time.Since(s)
	//rlog.Printf("<<< Found %d reverse syncs in %v: %v", count-needed, d, out)

	return out

}

func (p *Player) IsNearEnd() bool {
	cpu := apple2helpers.GetCPU(p.Destination)
	data := p.GetNextNSyncs(3, cpu.PC)

	return data[2] == 0
}

func (p *Player) GetNextNSyncs(count int, current int) []int {

	var out = make([]int, count)

	// if p.filenum < 0 {
	// 	return out
	// }
	if p.usemem && p.filenum >= len(p.memblocks) {
		return out
	}
	if !p.usemem && p.filenum >= len(p.files) {
		return out
	}

	// 1. save current state
	ostate := p.PlayerState
	defer func() {
		p.PlayerState = ostate
	}()

	var needed = count
	var bundle *ducktape.DuckTapeBundle
	var err error

	//var s = time.Now()
	var addr, lastAddr int

	lastAddr = current

	for needed > 0 {
		// go back 1 packet
		_, bundle, err = p.Next()
		if err == nil {
			// got packet
			if bundle.ID == "CPU" || bundle.ID == "CPD" {
				addr = p.extractAddressFromSync(bundle)
				// skip repeats from x-boundary sync packets
				if addr != lastAddr {
					out[count-needed] = addr
					needed--
					lastAddr = addr
				}
			}
		} else {
			break
		}
	}

	//d := time.Since(s)
	//rlog.Printf(">>> Found %d forward syncs in %v: %v", count-needed, d, out)

	return out

}

func (p *Player) SetBackwards(b bool) {
	p.backwards = b
	settings.AudioPacketReverse[p.Destination.MemIndex] = b
	p.ShowState()
}

func (p *Player) IsPlaying() bool {
	return p.running
}

func (p *Player) RealTimeStop() {
	//if settings.UnifiedRender[p.Destination.MemIndex] {
	//	p.stopNextVSync = true
	//	return
	//}
	p.rtNextSync = true
}

func (p *Player) NextMessagePos() (string, int, error) {
	msg := &ducktape.DuckTapeBundle{}
	delta := 0
	size := 0
	bsize := 6
	header := make([]byte, 0)
	chunk := make([]byte, bsize)

	var e error
	var count int

	cpos := p.fileptr

	for e == nil && len(header) < 6 {
		count, e = p.in(chunk)
		header = append(header, chunk[0:count]...)
		bsize = 6 - len(header)
		chunk = make([]byte, bsize)
	}

	if e != nil {
		return p.current.FileName, cpos, e
	}

	p.Pos += 6

	// got a header, let's decode it...
	delta = int(header[0])*65536 + int(header[1])*256 + int(header[2])
	size = int(header[3])*65536 + int(header[4])*256 + int(header[5])

	p.currentMS = delta

	bsize = 4096
	if size < bsize {
		bsize = size
	}

	chunk = make([]byte, bsize)
	buffer := make([]byte, 0)

	for e == nil && len(buffer) < size {
		count, e = p.in(chunk)
		buffer = append(buffer, chunk[0:count]...)
		if size-len(buffer) < 4096 {
			bsize = size - len(buffer)
		}
		chunk = make([]byte, bsize)
	}

	if len(buffer) != size {
		return p.current.FileName, cpos, errors.New(fmt.Sprintf("Player: incorrect chunk size got %d, expected %d", len(buffer), size))
	}

	if size == 0 {
		return p.current.FileName, cpos, nil
	}

	p.Pos += len(buffer)

	//fmt.Printf("Stream position is %d...\n", p.Pos)

	// got correct chunk size
	e = msg.UnmarshalBinary(buffer)
	return p.current.FileName, cpos, e
}

func (p *Player) Next() (int, *ducktape.DuckTapeBundle, error) {
	msg := &ducktape.DuckTapeBundle{}
	delta := 0
	size := 0
	bsize := 6
	header := make([]byte, 0)
	chunk := make([]byte, bsize)

	var e error
	var count int

	for e == nil && len(header) < 6 {
		count, e = p.in(chunk)
		header = append(header, chunk[0:count]...)
		bsize = 6 - len(header)
		chunk = make([]byte, bsize)
	}

	if e != nil {
		return delta, msg, e
	}

	p.Pos += 6

	// got a header, let's decode it...
	delta = int(header[0])*65536 + int(header[1])*256 + int(header[2])
	size = int(header[3])*65536 + int(header[4])*256 + int(header[5])

	p.currentMS = delta

	bsize = 4096
	if size < bsize {
		bsize = size
	}

	chunk = make([]byte, bsize)
	buffer := make([]byte, 0)

	for e == nil && len(buffer) < size {
		count, e = p.in(chunk)
		buffer = append(buffer, chunk[0:count]...)
		if size-len(buffer) < 4096 {
			bsize = size - len(buffer)
		}
		chunk = make([]byte, bsize)
	}

	if len(buffer) != size {
		return delta, msg, errors.New(fmt.Sprintf("Player: incorrect chunk size got %d, expected %d", len(buffer), size))
	}

	if size == 0 {
		return 0, msg, nil
	}

	// read trailing size
	trailer := make([]byte, 3)
	count, e = p.in(trailer)
	nsize := int(trailer[0])*65536 + int(trailer[1])*256 + int(trailer[2])

	if nsize != size {
		log2.Fatalf("Trailing size (%d) != Header packet size (%d)", nsize, size)
		panic("Trailing size != Header packet size")
	}

	p.Pos += len(buffer) + 3

	//fmt.Printf("Stream position is %d...\n", p.Pos)

	// got correct chunk size
	e = msg.UnmarshalBinary(buffer)
	return delta, msg, e
}

func (p *Player) Prev() (int, *ducktape.DuckTapeBundle, error) {
	msg := &ducktape.DuckTapeBundle{}
	delta := 0
	size := 0
	bsize := 6
	header := make([]byte, 0)
	chunk := make([]byte, bsize)

	if p.current.PositionRead == 0 {
		e := p.LoadPrevFile()
		if e != nil {
			return 0, msg, e
		}
	}

	//fmt.Printf("Prev called at pos=%d, file=%s\n", p.current.PositionRead, p.current.FileName)

	// We need to find the previous packet...
	// Assume we are at the end of a packet...
	// <DELTA><SIZE><PACKET><SIZE>

	var e error
	var count int

	var prevsize = make([]byte, 3)
	count, e = p.inB(prevsize)

	p.Pos -= 3

	// read size
	size = int(prevsize[0])*65536 + int(prevsize[1])*256 + int(prevsize[2])
	bsize = 4096
	if size < bsize {
		bsize = size
	}

	//fmt.Printf("Prev packet size = %d\n", size)

	// read data chunk
	chunk = make([]byte, bsize)
	buffer := make([]byte, 0)

	for e == nil && len(buffer) < size {
		count, e = p.inB(chunk)
		buffer = append(chunk[0:count], buffer...)
		if size-len(buffer) < 4096 {
			bsize = size - len(buffer)
		}
		chunk = make([]byte, bsize)
	}

	//fmt.Printf("Buffer length = %d\n", len(buffer))

	if e != nil {
		return delta, msg, e
	}

	p.Pos -= len(buffer)

	// read header
	header = make([]byte, 6)
	count, e = p.inB(header)

	p.Pos -= 6

	// got a header, let's decode it...
	delta = int(header[0])*65536 + int(header[1])*256 + int(header[2])

	//fmt.Printf("delta is %d ms\n", delta)

	size = int(header[3])*65536 + int(header[4])*256 + int(header[5])

	//fmt.Printf("Pre size = %d, delta = %d\n", size, delta)

	//fmt.Printf("Packet=[%s]\n", string(buffer))

	p.currentMS = delta

	if len(buffer) != size {
		return delta, msg, errors.New(fmt.Sprintf("Player: incorrect chunk size got %d, expected %d", len(buffer), size))
	}

	if size == 0 {
		return 0, msg, nil
	}

	// got correct chunk size
	e = msg.UnmarshalBinary(buffer)

	//fmt.Printf("MESG=%s\n", msg.ID)

	return delta, msg, e
}

func (p *Player) in(buffer []byte) (int, error) {

	count, e := p.current.ReadBytes(buffer)

	if e != nil && e.Error() == "EOF" {
		return count, p.LoadNextFile()
	}

	return count, e
}

func (p *Player) inB(buffer []byte) (int, error) {

	// try backup by len(buffer)

	if p.current.PositionRead >= len(buffer) {

		p.current.PositionRead -= len(buffer)
		//fmt.Printf("Read %d bytes from pos %d\n", len(buffer), p.current.PositionRead)

		count, e := p.current.ReadBytes(buffer)
		p.current.PositionRead -= len(buffer)
		return count, e
	}

	tmpchunk := make([]byte, p.current.PositionRead)
	p.current.PositionRead -= len(tmpchunk)
	count, _ := p.current.ReadBytes(tmpchunk)
	p.current.PositionRead -= len(tmpchunk)

	return count, p.LoadPrevFile()
}

func (p *Player) ShowProgress() {
	// pc := float32(p.filenum) / float32(len(p.files))
	// // if p.usemem {
	// // 	pc = float32(p.filenum) / float32(len(p.memblocks))
	// // }
	// if p.backwards {
	// 	apple2helpers.OSDShowBlink(p.Destination, "REWIND", 1000000000)
	// } else {
	// 	apple2helpers.OSDShowBlink(p.Destination, "PLAY", 1000000000)
	// }

	blocknum := p.filenum
	blocks := len(p.memblocks)
	if len(p.files) != 0 {
		blocks = len(p.files)
	}

	servicebus.SendServiceBusMessage(
		p.Destination.GetMemIndex(),
		servicebus.LiveRewindStateUpdate,
		&debugtypes.LiveRewindState{
			Enabled:     true,
			CanBack:     true,
			CanForward:  true,
			CanResume:   true,
			Backwards:   p.backwards,
			TimeFactor:  p.timeFactor,
			Block:       blocknum,
			TotalBlocks: blocks,
		},
	)
}

func (p *Player) BusConnect() {
	slotid := p.Destination.GetMemIndex()
	servicebus.Subscribe(slotid, servicebus.PlayerPause, p)
	servicebus.Subscribe(slotid, servicebus.PlayerResume, p)
	servicebus.Subscribe(slotid, servicebus.PlayerFaster, p)
	servicebus.Subscribe(slotid, servicebus.PlayerSlower, p)
	servicebus.Subscribe(slotid, servicebus.PlayerBackstep, p)
}

func (p *Player) BusDisconnect() {
	slotid := p.Destination.GetMemIndex()
	servicebus.Unsubscribe(slotid, p)
}

func (p *Player) AddSeekDelta(ms int) {
	p.seekDelta += 1000 * ms
}

func (p *Player) IsSeeking() bool {
	return p.seekDelta > 0 && p.backwards
}

func (p *Player) IsActive() bool {
	return p.active
}

func (p *Player) Playback() interfaces.PlayerExitMode {
	if p.active {
		return p.exitmode
	}
	p.active = true
	settings.SlotZPEmu[p.Destination.MemIndex] = false
	settings.AudioPacketReverse[p.Destination.MemIndex] = p.backwards

	bus.StartDefault()

	//fmt.Println("In playback")

	var e error
	var delta int
	var msg *ducktape.DuckTapeBundle

	p.start = time.Now()
	vdelta := 0

	var lastOSD = time.Now()
	var lastFrame = time.Now()
	p.running = true
	p.BusConnect()
	//var backStep = p.seekDelta > 0
	defer func() {
		p.running = false
		p.BusDisconnect()
		//log2.Printf("Pausing")
		//time.Sleep(time.Second)
		p.active = false
		p.ResumeRecord()
		cpu := apple2helpers.GetCPU(p.Destination)
		//if backStep {
		//	apple2helpers.OSDShow(p.Destination, "Go!")
		//}
		cpu.RequestResume()
	}()
	p.ShowProgress()

	var tf float64

	for e == nil && p.running {

		p.HandleServiceBusInjection(p.HandleServiceBusRequest)

		if p.Destination.GetMemoryMap().IntGetSlotMenu(p.Destination.GetMemIndex()) {
			editor.TestMenu(p.Destination)
			p.Destination.GetMemoryMap().IntSetSlotMenu(p.Destination.GetMemIndex(), false)
		}

		if p.Destination.GetMemoryMap().IntGetSlotRestart(p.Destination.GetMemIndex()) {
			log.Println("Continue due to restart...")
			p.Continue(false)
			return p.exitmode
		}

		if p.backwards {
			delta, msg, e = p.Prev()
			//log2.Printf("backwards delta %d, factor = %f", delta, p.timeFactor)
			if e != nil {
				if p.timeFactor == 0 {
					log.Println("sync to first")
					p.SetBackwards(false)
					p.cpuSyncCount = 1
					//p.running = true
					delta, msg, e = p.Next()
					goto resume
				}
				p.SetBackwards(false)
				delta, msg, e = p.Next()
				if settings.DebuggerAttachSlot-1 == p.Destination.GetMemIndex() {
					p.timeFactor = 0
					apple2helpers.OSDShow(p.Destination, "At start of recording. Paused.")
					p.debugSeek = false
					p.seekDelta = 0
					servicebus.SendServiceBusMessage(
						p.Destination.GetMemIndex(),
						servicebus.LiveRewindStateUpdate,
						&debugtypes.LiveRewindState{
							Enabled:     true,
							CanBack:     false,
							CanForward:  true,
							CanResume:   true,
							Backwards:   p.backwards,
							TimeFactor:  p.timeFactor,
							Block:       p.filenum,
							TotalBlocks: len(p.files),
						},
					)
				}
				//p.running = true
			}
		} else {
			delta, msg, e = p.Next()
			if e != nil {
				// resume
				log.Println("Continuing as we are out of data in forward direction", p.debugSeek, p.timeFactor)
				if p.timeFactor == 0 {
					log.Println("sync to last")
					p.SetBackwards(true)
					p.cpuSyncCount = 1
					//p.running = true
					delta, msg, e = p.Prev()
					goto resume
				}
				p.Continue(true)
				break
			}
		}

	resume:
		if p.cpuSyncCount > 0 {
			// as fast as possible between sync points
			delta = 0
		}

		if p.seekDelta <= 0 {
			tf = p.timeFactor
			if tf != 0 {
				vdelta += int(float64(delta) / tf)
			} else {
				// only wait if we are not syncing
				if p.cpuSyncCount == 0 {
					s := time.Now()
					for p.timeFactor == 0 && p.cpuSyncCount == 0 && !p.rtNextSync {
						p.HandleServiceBusInjection(p.HandleServiceBusRequest)
						time.Sleep(time.Millisecond)
					}
					us := time.Since(s) / time.Microsecond
					vdelta += int(us)
				}
			}

			since_ms := int(time.Since(p.start) / time.Microsecond)
			d := vdelta - since_ms
			if d > 1000 && !p.stopNextVSync && !p.rtNextSync && p.seekDelta <= 0 {
				time.Sleep(time.Duration(d) * time.Microsecond)
			}
		}

		if e == nil {
			p.HandleServiceBusInjection(p.HandleServiceBusRequest)
			if !p.Do(msg) {
				apple2helpers.OSDPanel(p.Destination, false)
				return p.exitmode
			}
			if time.Since(lastOSD) > 500*time.Millisecond {
				p.ShowProgress()
				lastOSD = time.Now()
			}

			if time.Since(lastFrame) > 17*time.Millisecond {
				if settings.UnifiedRender[p.Destination.MemIndex] {
					servicebus.SendServiceBusMessage(p.Destination.MemIndex, servicebus.UnifiedPlaybackSync, "")
				}
				lastFrame = time.Now()
				time.Sleep(time.Millisecond)
			}

			if p.seekDelta > 0 && p.backwards {
				if msg.ID == "FUL" && p.filenum == 0 {
					log2.Printf("Seeking hit start of stream")
					p.seekDelta = 0
					p.Continue(false)
				}
				p.seekDelta -= delta
				if p.seekDelta <= 0 {
					if msg.ID == "CPU" {
						p.Continue(false)
					} else {
						p.rtNextSync = true
						p.rtNextSyncRolloff = false
					}
				}
			}
		}
	}

	apple2helpers.OSDPanel(p.Destination, false)

	if e != nil {
		fmt.Println(e)
	}

	settings.SlotZPEmu[p.Destination.MemIndex] = true
	settings.AudioPacketReverse[p.Destination.MemIndex] = false

	p.ShowState()

	return p.exitmode

}

func (p *Player) CreateRefTable() []PlayerPos {

	//fmt.Println("In refbuild")

	var e error
	var pos int
	var filename string
	var out []PlayerPos

	p.start = time.Now()

	for e == nil {
		filename, pos, e = p.NextMessagePos()

		//fmt.Printf( "Playback: Delta %d: %s (payload size %d bytes)\n", delta, msg.ID, len(msg.Payload) )
		if e == nil {
			out = append(out, PlayerPos{Filename: filename, Position: pos})
		}
	}

	if e != nil {
		fmt.Println(e)
	}

	return out

}

func (p *Player) ShowState() {
	if p.seekDelta > 0 {
		return
	}
	if p.timeFactor != 0 && p.backwards {
		apple2helpers.OSDShow(
			p.Destination,
			fmt.Sprintf("Speed: %.2f (rew)", p.timeFactor),
		)
	} else if p.timeFactor != 0 && !p.backwards {
		apple2helpers.OSDShow(
			p.Destination,
			fmt.Sprintf("Speed: %.2f (fwd)", p.timeFactor),
		)
	} else {
		apple2helpers.OSDShow(
			p.Destination,
			"Speed: 0.00 (paused)",
		)
	}
}

func (p *Player) PlaybackToMS(ms int, r *Recorder, deltaoffset int) error {

	//fmt.Printf("In playback to %d ms...\n", ms)

	var e error
	var delta int
	var msg *ducktape.DuckTapeBundle

	p.start = time.Now()

	for e == nil && p.currentMS < ms {
		delta, msg, e = p.Next()

		//fmt.Printf( "Playback: Delta %d: %s (payload size %d bytes)\n", delta, msg.ID, len(msg.Payload) )
		if e == nil {
			p.Do(msg)

			// record packet if r set
			if r != nil {
				d := delta - deltaoffset
				rawmsg, _ := msg.MarshalBinary()
				sz := len(rawmsg)

				// record packet
				r.Out([]byte{byte((d / 65536) & 0xff), byte((d / 256) & 0xff), byte(d % 256)})
				r.Out(
					[]byte{
						byte((sz >> 16) & 0xff),
						byte((sz >> 8) & 0xff),
						byte(sz & 0xff),
					},
				)

				r.Out(rawmsg)
			}
		}

	}

	return e

}

func (p *Player) IsBackwards() bool {
	return p.backwards
}

func (p *Player) ContinueSimple() {
	cpu := apple2helpers.GetCPU(p.Destination)
	cpu.Halted = false

	// if false {
	settings.SetPureBoot(p.Destination.MemIndex, true)
	settings.AudioPacketReverse[p.Destination.MemIndex] = false

	p.Destination.SetState(types.EXEC6502)
	cpu.ResetSliced()
	cpu.CheckWarp()

	bus.StopClock()

	servicebus.SendServiceBusMessage(
		p.Destination.GetMemIndex(),
		servicebus.LiveRewindStateUpdate,
		&debugtypes.LiveRewindState{
			Enabled:    p.Destination.IsRecordingVideo(),
			CanBack:    true,
			CanForward: false,
			CanResume:  false,
		},
	)
}

func (p *Player) ResumeRecord() {
	if p.usemem {
		blocks := p.memblocks[0:p.filenum]
		if p.filenum < len(p.memblocks) {
			b := p.memblocks[p.filenum]
			b.Truncate(p.current.PositionRead)
			blocks = append(blocks, b)
		}

		settings.VideoRecordFrames[p.Destination.MemIndex] = blocks

		if settings.UnifiedRender[p.Destination.MemIndex] {
			mr, ok := p.Destination.Memory.InterpreterMappableAtAddress(p.Destination.MemIndex, 0xc000)
			if ok {
				if io, ok := mr.(*apple2.Apple2IOChip); ok {
					io.UnifiedFrame.RealSync()
					log2.Printf("Resuming live from scan cycle offset %d", io.UnifiedFrame.Clock)
				}
			}
		}

		p.Destination.ResumeRecording(settings.VideoRecordFile[p.Destination.MemIndex], blocks, false)
	} else {
		if settings.DebuggerAttachSlot-1 == p.Destination.GetMemIndex() && !p.noResume {

			log.Printf("p.filenum = %d, p.fileptr = %d, len(p.files) = %d, near end = %v", p.filenum, p.fileptr, len(p.files), p.IsNearEnd())
			log.Printf("full cpu recording is %v", settings.DebugFullCPURecord)

			if p.rtNextSyncRolloff || p.lastFile {
				log.Printf("*** Resuming previous recording: %s", p.pathname)
				p.Destination.ResumeRecording(p.pathname, nil, settings.DebugFullCPURecord)
			} else {
				log.Printf("*** Creating a new recording")
				p.Destination.StopRecording()
				p.Destination.RecordToggle(settings.DebugFullCPURecord)
			}
		}
	}
}

func (p *Player) Continue(rollout bool) {

	//settings.VideoSuspended = true

	cpu := apple2helpers.GetCPU(p.Destination)

	// if false {
	settings.SetPureBoot(p.Destination.MemIndex, true)
	settings.AudioPacketReverse[p.Destination.MemIndex] = false

	apple2helpers.OSDPanel(p.Destination, false)

	// cleanup
	p.rtNextSync = false

	p.Destination.SetState(types.EXEC6502)
	cpu.ResetSliced()
	cpu.SetWarp(1)
	cpu.SetWarpUser(1)
	cpu.CheckWarp()
	cpu.Halted = false

	cpu.RecRegisters = cpu.Registers

	log2.Printf("CPU State at resume: %+v", cpu.Registers)

	bus.StopClock()

	servicebus.SendServiceBusMessage(
		p.Destination.GetMemIndex(),
		servicebus.LiveRewindStateUpdate,
		&debugtypes.LiveRewindState{
			Enabled:    p.Destination.IsRecordingVideo(),
			CanBack:    true,
			CanForward: false,
			CanResume:  false,
		},
	)

	cpu.RequestResume()
}

func (p *Player) Jump(syncs int) {
	// don't adjust jump whilst seeking
	if p.cpuSyncCount != 0 || p.debugSeek {
		return
	}
	p.cpuSyncCount = int(math.Abs(float64(syncs)))
	p.backwards = syncs < 0
	//p.timeFactor = 0.25
	p.debugSeek = true
}

func (p *Player) Do(msg *ducktape.DuckTapeBundle) bool {

	//log.Printf("id: %s\n", msg.ID)
	//log.Printf("filenum=%d, blocks=%d, blocklen=%d, fr.ReadPos=%d", p.filenum, len(p.memblocks)-1, len(p.current.Content), p.current.PositionRead)

	if p.usemem && !p.backwards && p.filenum == len(p.memblocks)-1 && p.current.PositionRead >= len(p.current.Content) && !p.debugSeek {
		log.Printf("Setting next sync flag")
		p.rtNextSync = true
	}

	if !p.usemem && !p.backwards && p.filenum == len(p.files)-1 && p.current.PositionRead >= len(p.current.Content) && !p.debugSeek {
		log.Printf("Setting next sync flag")
		p.rtNextSync = true
		p.rtNextSyncRolloff = true
	}

	switch msg.ID {
	case "DR0":
		//fmt.Println("Drive 0 disk changed")
		apple2.DiskInsertBin(p.Destination, 0, msg.Payload, "virtual#0", false)
	case "DR1":
		//fmt.Println("Drive 1 disk changed")
		apple2.DiskInsertBin(p.Destination, 1, msg.Payload, "virtual#1", false)
	case "MCK":
		// handle mockingboard state packets
		mr, ok := p.Destination.Memory.InterpreterMappableAtAddress(p.Destination.MemIndex, 0xc000)
		if ok {
			io := mr.(*apple2.Apple2IOChip)
			card := (io.GetCard(4)).(*apple2.IOCardMockingBoard)
			if len(msg.Payload) >= 30 {
				card.GetChip(0).FromBytes(msg.Payload[:15])
				card.GetChip(1).FromBytes(msg.Payload[15:])
			}
		}
	case "CPD":
		// CPU Delta state unpack

		delta := &CPUDelta{}
		delta.FromBytes(msg.Payload)

		force := false
		if p.cpuSyncCount > 0 && int(*delta.PC) != p.lastSeekAddr {
			p.cpuSyncCount--
			if p.cpuSyncCount == 0 {
				//fmt.Println("sync reached")
				p.debugSeek = false
				force = true
				if p.Destination.GetMemIndex() == settings.DebuggerAttachSlot-1 {
					servicebus.SendServiceBusMessage(
						p.Destination.GetMemIndex(),
						servicebus.LiveRewindStateUpdate,
						&debugtypes.LiveRewindState{
							Enabled:     p.IsPlaying(),
							CanBack:     true,
							CanForward:  true,
							CanResume:   true,
							Backwards:   p.backwards,
							TimeFactor:  p.timeFactor,
							Block:       p.filenum,
							TotalBlocks: len(p.files),
						},
					)
				}
			}
		}
		p.lastSeekAddr = int(*delta.PC)

		settings.SlotZPEmu[p.Destination.MemIndex] = false

		cpu := apple2helpers.GetCPU(p.Destination)
		if delta.A != nil {
			cpu.A = int(*delta.A)
		}
		if delta.X != nil {
			cpu.X = int(*delta.X)
		}
		if delta.Y != nil {
			cpu.Y = int(*delta.Y)
		}
		if delta.PC != nil {
			cpu.PC = int(*delta.PC)
		}
		if delta.P != nil {
			cpu.P = int(*delta.P)
		}
		if delta.SP != nil {
			cpu.SP = 0x100 + int(*delta.SP)
		}

		slot := p.Destination.GetMemIndex()

		if slot == settings.DebuggerAttachSlot-1 {

			servicebus.SendServiceBusMessage(
				slot,
				servicebus.CPUState,
				&debugtypes.CPUState{
					A:           cpu.A,
					X:           cpu.X,
					Y:           cpu.Y,
					PC:          cpu.PC,
					P:           cpu.P,
					SP:          cpu.SP,
					IsRecording: true, // only set this so we know the lookback/ahead decodes need to come from the player
					ForceUpdate: force,
				},
			)

		}

	case "CPU":
		//fmt.Println("CPU syncpoint")

		cs := CPUSwitchState{}
		var offset int
		var v CSSFlag
		var uval int
		for i := 0; i < int(CSS_SIZE); i++ {
			v = CSSFlag(i)
			offset = i * 4
			if offset >= len(msg.Payload) {
				continue
			}
			uval = (int(msg.Payload[offset+0]) << 24) | (int(msg.Payload[offset+1]) << 16) | (int(msg.Payload[offset+2]) << 8) | int(msg.Payload[offset+3])
			switch v {
			case CSS_MMU_MEMMODE:
				cs.memmode = apple2.MemoryFlag(uval)
			case CSS_MMU_VIDMODE:
				cs.vidmode = apple2.VideoFlag(uval)
			case CSS_CPU_PC:
				cs.CPU.PC = uval
			case CSS_CPU_A:
				cs.CPU.A = uval
			case CSS_CPU_X:
				cs.CPU.X = uval
			case CSS_CPU_Y:
				cs.CPU.Y = uval
			case CSS_CPU_P:
				cs.CPU.P = uval
			case CSS_CPU_SP:
				cs.CPU.SP = uval
			case CSS_CPU_SPEED:
				cs.CPU.SPEED = uval
			case CSS_CPU_SCANOFFSET:
				cs.CPU.ScanCycle = uval
			}
		}

		force := false
		if p.cpuSyncCount > 0 && cs.CPU.PC != p.lastSeekAddr {
			p.cpuSyncCount--
			if p.cpuSyncCount == 0 {
				//fmt.Println("sync reached")
				p.debugSeek = false
				force = true
				if p.Destination.GetMemIndex() == settings.DebuggerAttachSlot-1 {
					servicebus.SendServiceBusMessage(
						p.Destination.GetMemIndex(),
						servicebus.LiveRewindStateUpdate,
						&debugtypes.LiveRewindState{
							Enabled:     p.IsPlaying(),
							CanBack:     true,
							CanForward:  true,
							CanResume:   true,
							Backwards:   p.backwards,
							TimeFactor:  p.timeFactor,
							Block:       p.filenum,
							TotalBlocks: len(p.files),
						},
					)
				}
			}
		}
		p.lastSeekAddr = cs.CPU.PC

		settings.SlotZPEmu[p.Destination.MemIndex] = false

		cpu := apple2helpers.GetCPU(p.Destination)
		cpu.A = cs.CPU.A
		cpu.X = cs.CPU.X
		cpu.Y = cs.CPU.Y
		cpu.PC = cs.CPU.PC
		cpu.P = cs.CPU.P
		cpu.SP = cs.CPU.SP
		//cpu.GlobalCycles = int64(cs.CPU.ScanCycle)
		cpu.SetWarp(1)
		cpu.SetWarpUser(1)

		mr, ok := p.Destination.Memory.InterpreterMappableAtAddress(p.Destination.MemIndex, 0xc000)
		if ok {
			io := mr.(*apple2.Apple2IOChip)
			io.SetVidModeForce(cs.vidmode)
			io.SetMemModeForce(cs.memmode)
			//frame := io.VBlankLength + io.VerticalRetrace
			//if settings.UnifiedRender[p.Destination.MemIndex] && cs.CPU.ScanCycle != 0 && cs.CPU.ScanCycle < 17023 {
			//	log2.Printf("Warning spurious CPU sync is outside expected range (%d)", cs.CPU.ScanCycle)
			//}
			io.GlobalCycles = int64(cs.CPU.ScanCycle)
			io.UnifiedFrame.Clock = io.GlobalCycles
		}

		slot := p.Destination.GetMemIndex()

		if slot == settings.DebuggerAttachSlot-1 {
			servicebus.SendServiceBusMessage(
				slot,
				servicebus.CPUState,
				&debugtypes.CPUState{
					A:           cpu.A,
					X:           cpu.X,
					Y:           cpu.Y,
					PC:          cpu.PC,
					P:           cpu.P,
					SP:          cpu.SP,
					IsRecording: true, // only set this so we know the lookback/ahead decodes need to come from the player
					ForceUpdate: force,
				},
			)
		}

		if p.rtNextSync {
			p.rtNextSync = false
			//fmt.RPrintf("Breaking into the real world @ 0x%.4x...\n", cs.CPU.PC)
			log.Println("Continuing now")
			p.Continue(p.rtNextSyncRolloff)

			return false
		}

	case "DII":

		payload := msg.Payload

		mr, ok := p.Destination.Memory.InterpreterMappableAtAddress(p.Destination.MemIndex, 0xc000)
		if ok {

			io := mr.(*apple2.Apple2IOChip)
			card := (io.GetCard(6)).(*apple2.IOCardDiskII)

			card.FromBytes(payload)

			//// Set modified in case we rewind a disk write...
			//// this will be picked up on CPU resume...
			//if card.GetDrive(card.GetCurrent()).GetDiskUpdatePending() {
			//	card.GetDrive(card.GetCurrent()).Disk.SetModified(true)
			//}

		}

	case "FUL", "FLR":
		//fmt.RPrintln("FUL")
		cdata := msg.Payload
		var data []byte
		if msg.ID == "FUL" {
			data = utils.UnGZIPBytes(cdata)
		} else {
			data = cdata
		}
		//_ = p.Destination.ThawBytesNoPost(data)
		fr := freeze.NewEmptyState(p.Destination)
		_ = fr.LoadFromBytes(data)
		fr.Apply(p.Destination)
		if p.seekDelta <= 0 && settings.VideoPlaybackPauseOnFUL[p.Destination.GetMemIndex()] {
			p.timeFactor = 0
			settings.VideoPlaybackPauseOnFUL[p.Destination.GetMemIndex()] = false
		}

	case "MMD":
		var mode int

		if p.backwards {
			mode = (int(msg.Payload[0]) << 24) | (int(msg.Payload[1]) << 16) | (int(msg.Payload[2]) << 8) | int(msg.Payload[3])
		} else {
			mode = (int(msg.Payload[4]) << 24) | (int(msg.Payload[5]) << 16) | (int(msg.Payload[6]) << 8) | int(msg.Payload[7])
		}

		//fmt.RPrintf("--> mode %s\n", apple2.MemoryFlag(mode))

		mr, ok := p.Destination.Memory.InterpreterMappableAtAddress(p.Destination.MemIndex, 0xc000)
		if ok {
			mr.(*apple2.Apple2IOChip).SetMemMode(apple2.MemoryFlag(mode))
			mr.(*apple2.Apple2IOChip).ConfigurePaging(false)
		}

	case "USS":
		if !p.backwards {
			servicebus.UnsubscribeType(p.Destination.MemIndex, servicebus.UnifiedPlaybackSync)
			mr, ok := p.Destination.Memory.InterpreterMappableAtAddress(p.Destination.MemIndex, 0xc000)
			if ok {
				//log2.Printf("Player: Restoring scan state")
				mr.(*apple2.Apple2IOChip).UnifiedFrame.RestoreState(msg.Payload)
			}
		}

	case "UVB":
		// Unified Scan Capture
		if p.seekDelta <= 0 {
			mr, ok := p.Destination.Memory.InterpreterMappableAtAddress(p.Destination.MemIndex, 0xc000)
			if ok {
				//clock := int64(msg.Payload[0])*256 + int64(msg.Payload[1])
				//log2.Printf("Player: Received VBLANK signal")
				io := mr.(*apple2.Apple2IOChip)
				io.UnifiedFrame.RealSync()
				//io.GlobalCycles = clock
				//io.UnifiedFrame.Clock = clock
			}
		}
		//if p.stopNextVSync {
		//	p.Continue(false)
		//}

	case "USD":
		// Unified Scan Delta
		var line = int(msg.Payload[0])
		var xoffs = int(msg.Payload[1])
		var mainO = msg.Payload[2]
		var mainN = msg.Payload[3]
		var auxO = msg.Payload[4]
		var auxN = msg.Payload[5]
		var modeO = msg.Payload[6]
		var modeN = msg.Payload[7]
		//var clock = int64(msg.Payload[5]) + 256*int64(msg.Payload[6])

		mr, ok := p.Destination.Memory.InterpreterMappableAtAddress(p.Destination.MemIndex, 0xc000)
		if ok {
			//log2.Printf("Apply scan delta @ seg = %d, line = %d", xoffs, line)
			if p.backwards {
				mr.(*apple2.Apple2IOChip).UnifiedFrame.ApplyScanDelta(line, xoffs, mainO, auxO, modeO)
			} else {
				mr.(*apple2.Apple2IOChip).UnifiedFrame.ApplyScanDelta(line, xoffs, mainN, auxN, modeN)
			}
			//mr.(*apple2.Apple2IOChip).GlobalCycles = clock
			//mr.(*apple2.Apple2IOChip).UnifiedFrame.Clock = clock
		}

	case "VMD":
		var mode int

		if p.backwards {
			mode = (int(msg.Payload[0]) << 24) | (int(msg.Payload[1]) << 16) | (int(msg.Payload[2]) << 8) | int(msg.Payload[3])
		} else {
			mode = (int(msg.Payload[4]) << 24) | (int(msg.Payload[5]) << 16) | (int(msg.Payload[6]) << 8) | int(msg.Payload[7])
		}

		mr, ok := p.Destination.Memory.InterpreterMappableAtAddress(p.Destination.MemIndex, 0xc000)
		if ok {
			mr.(*apple2.Apple2IOChip).SetVidModeForce(apple2.VideoFlag(mode))
		}

	case "SND":
		if p.seekDelta <= 0 {
			data := msg.Payload
			count := int(data[0])<<16 | int(data[1])<<8 | int(data[2])
			rate := int(data[3])<<8 | int(data[4])
			bytepacked := data[5] != 0
			indata := make([]uint64, count)

			fmt.Printf("SND %d bytes, %d rate, bytepacked = %v\n", count, rate, bytepacked)

			for i, _ := range indata {
				offs := 6 + i*4
				indata[i] = (uint64(data[offs+0]) << 24) | (uint64(data[offs+1]) << 16) | (uint64(data[offs+2]) << 8) | uint64(data[offs+3])
			}

			r := uint64(rate)
			if bytepacked {
				r |= 0xff0000
			}

			p.Destination.Memory.DirectSendAudioPacked(p.Destination.MemIndex, 0, indata, rate)
		}

	case "BMU":
		data := msg.Payload
		fullcount := int(data[0])<<16 | int(data[1])<<8 | int(data[2])
		count := fullcount / 2
		idx := 3

		ss := 0
		ee := count - 1
		if p.backwards {
			ss = count
			ee = fullcount - 1
		}

		for i := 0; i < fullcount; i++ {
			end := idx + 13
			if end >= len(data) {
				end = len(data)
			}
			chunk := data[idx:end]

			_, addr, value, read, size, e := mempak.Decode(chunk)
			if e != nil {
				break
			}

			if i < ss || i > ee {
				idx += size
				continue
			}

			//if i >= ss && i <= ee {

			if !read {

				a := addr % memory.OCTALYZER_INTERPRETER_SIZE

				if a >= 8192 && a <= 24576 {
					p.Destination.Memory.WriteGlobal(
						p.Destination.MemIndex,
						p.Destination.Memory.MEMBASE(p.Destination.MemIndex)+a,
						value)
				} else if a >= memory.MICROM8_VOICE_PORT_BASE && a <= memory.MICROM8_VOICE_PORT_BASE+memory.MICROM8_VOICE_PORT_SIZE*memory.MICROM8_VOICE_COUNT {
					// p.Destination.Memory.WriteGlobal(
					// 	p.Destination.Memory.MEMBASE(p.Destination.MemIndex)+a,
					// 	value)

					offs := a - memory.MICROM8_VOICE_PORT_BASE
					voice := offs / 2
					isOpCode := offs%2 == 0

					p.Destination.Memory.WriteGlobalSilent(
						p.Destination.MemIndex,
						p.Destination.Memory.MEMBASE(p.Destination.MemIndex)+a,
						value)

					if p.backwards {
						if !isOpCode {
							opcode := int(p.Destination.Memory.ReadGlobal(p.Destination.MemIndex, p.Destination.Memory.MEMBASE(p.Destination.MemIndex)+a-1))
							p.Destination.Memory.RestalgiaOpCode(
								p.Destination.MemIndex,
								voice,
								opcode,
								value,
							)
						}
					} else {
						if isOpCode {
							v := p.Destination.Memory.ReadGlobal(p.Destination.MemIndex, p.Destination.Memory.MEMBASE(p.Destination.MemIndex)+a+1)
							p.Destination.Memory.RestalgiaOpCode(
								p.Destination.MemIndex,
								voice,
								int(value),
								v,
							)
						}
					}

				} else {
					if a != memory.OCTALYZER_SLOT_RESTART {
						p.Destination.Memory.WriteGlobalSilent(
							p.Destination.MemIndex,
							p.Destination.Memory.MEMBASE(p.Destination.MemIndex)+a,
							value)
					}
				}

			} else {

				p.Destination.Memory.ReadInterpreterMemory(
					p.Destination.MemIndex,
					addr%memory.OCTALYZER_INTERPRETER_SIZE,
				)

			}

			//}

			idx += size

		}
	}

	return true

}

// Copy records a portion of the stream out to a new stream
func (p *Player) Copy(startMS, endMS int, target string) error {

	r, e := NewRecorder(p.Destination, target, []*bytes.Buffer(nil), false)

	p.Destination.GetMemoryMap().Track[p.Destination.GetMemIndex()] = false // tracking off for now..

	// Step one, advance timecode to start point
	e = p.PlaybackToMS(startMS, nil, 0)
	if e != nil {
		return e
	}

	// Step two, freeze current memory state
	r.FreezeInitial()

	e = p.PlaybackToMS(endMS, r, startMS) // passing r here allows recording of the packets

	return e

}
