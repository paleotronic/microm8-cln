package interpreter

import (
	"bytes"
	"io/ioutil"
	log2 "log"
	"sync"
	"time"

	"paleotronic.com/core/hardware/apple2"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/hardware/cpu/mos6502"
	"paleotronic.com/core/hardware/servicebus"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/settings"
	"paleotronic.com/debugger/debugtypes"
	"paleotronic.com/ducktape"
	"paleotronic.com/encoding/mempak"
	"paleotronic.com/files"
	"paleotronic.com/fmt"
	"paleotronic.com/freeze"
	"paleotronic.com/log"
	"paleotronic.com/utils"
)

/*
	Memory stream format.

	Description		Bytes		Format
	delta ms 		2			H, L
	chunksize		3			HH, H, L
	chunk			<chunksize>	Ducktape binary encoded

	Possible current message types:

	FUL:			Must be the first message, GZIPPED memory freeze (initial state)
	BMU:			Bulk Memory Update (delta changes to memory locations)
	DR0:			Disk 0 change (will follow a FUL, or precede a CPU if there was a change)
	DR1:			Disk 1 change (will follow a FUL, or precede a CPU if there was a change)
	CPU:			CPU state event (typically once every speed / 10 cycles)
*/

const MAX_BLOCK = 1 << 17
const CLOCKS_PER_REGISTER_SYNC = 17030
const MaxStreamSize = 24576

type CPUSwitchState struct {
	CPU     freeze.CPURegs
	vidmode apple2.VideoFlag
	memmode apple2.MemoryFlag
	drives  *RecordDriveEvent
}

type CSSFlag int

const (
	CSS_CPU_PC CSSFlag = iota
	CSS_CPU_P
	CSS_CPU_SP
	CSS_CPU_A
	CSS_CPU_X
	CSS_CPU_Y
	CSS_MMU_VIDMODE
	CSS_MMU_MEMMODE
	CSS_CPU_SPEED
	CSS_CPU_SCANOFFSET
	// this should always be last
	CSS_SIZE
)

const MAX_MEM_BLOCKS = 1000

type Recorder6522State struct {
	controllers [2][]byte
}

type RecorderMemPageEvent struct {
	lastmode int
	newmode  int
}

type RecorderVidPageEvent struct {
	lastmode int
	newmode  int
}

type RecorderAudioEvent struct {
	count      int
	rate       int
	bytepacked bool
	data       []uint64
}

type RecorderJumpEvent struct {
	isJump bool
	from   int
	to     int
}

type RecordDriveEvent struct {
	bitptr0      int
	bitptr1      int
	clock        int
	state        int
	dataRegister byte
	pendingWrite byte
	writeMode    bool
	current      int
}

type RecorderEvent struct {
	cpu      *CPUSwitchState
	ncpu     *CPUDelta // this is the new improved way
	mem      []memory.MemoryChange
	ful      bool
	done     chan bool
	audio    RecorderAudioEvent
	memmode  *RecorderMemPageEvent
	vidmode  *RecorderVidPageEvent
	jmp      *RecorderJumpEvent
	drives   []byte
	mock     *Recorder6522State
	vblank   *int
	scandata []byte
	seq      int
	full     []byte
	stop     bool
}

type RecorderEventStream struct {
	buffer  chan *RecorderEvent
	seq     int
	lastPut time.Time
}

func NewRecorderEventStream() *RecorderEventStream {
	return &RecorderEventStream{
		buffer: make(chan *RecorderEvent, MaxStreamSize),
		seq:    1,
	}
}

func (res *RecorderEventStream) Put(re *RecorderEvent) {
	re.seq = res.seq
	res.buffer <- re
	res.seq++
	res.lastPut = time.Now()
}

func (res *RecorderEventStream) Empty() {
	for len(res.buffer) > 0 {
		<-res.buffer
	}
}

func (res *RecorderEventStream) Get() *RecorderEvent {
	if len(res.buffer) == 0 {
		return nil
	}
	re := <-res.buffer
	return re
}

func (res *RecorderEventStream) CanGet() bool {
	return len(res.buffer) > 0
}

type Recorder struct {
	Source              *Interpreter
	Buffer              []memory.MemoryChange
	start, lastupdate   time.Time
	running             bool
	pathname            string
	filename            string
	blocknum            int
	blockbuffer         *bytes.Buffer
	muChan              *RecorderEventStream
	accCycles           int
	d0, d1              string
	usemem              bool
	memblocks           []*bytes.Buffer
	lastmemmode         int
	lastvidmode         int
	globalCycles        int64
	lastCycles          int64
	lastCyclesPerSecond int64
	stopNextSync        bool
	cpu                 *mos6502.Core6502
	allCPUStates        bool // true if we need to capture all cpu states
	lastCPU             *freeze.CPURegs
	lastGlobalCycles    int64
	injectedBusRequests []*servicebus.ServiceBusRequest
	m                   sync.Mutex
	unified             bool
	// EPS               int
	// counters          map[string]int
	// mcounters         map[int]int
	// last camera state
	//camera [memory.OCTALYZER_MAPPED_CAM_SIZE]uint64
}

func bfn(n int) string {
	return fmt.Sprintf("block%.6d.s", n)
}

func NewRecorder(source *Interpreter, p string, blocks []*bytes.Buffer, debugMode bool) (*Recorder, error) {
	this := &Recorder{
		Source:       source,
		Buffer:       make([]memory.MemoryChange, 0),
		running:      false,
		pathname:     p,
		usemem:       (p == ""),
		memblocks:    blocks,
		allCPUStates: debugMode && (p != ""),
		unified:      settings.UnifiedRender[source.MemIndex],
		//camera:    source.Memory.GetGFXCameraData(source.MemIndex, 0),
	}

	if this.memblocks == nil {
		this.memblocks = make([]*bytes.Buffer, 0)
	}
	if this.usemem {
		this.blocknum = len(this.memblocks)
	}

	// Create base dir
	if !this.usemem {
		e := files.MkdirViaProvider(p)
		if e != nil {
			return nil, e
		}
	}

	// are there any files here...?
	_, f, e := files.ReadDirViaProvider(p+"/", "*.s")
	if e != nil {
		fmt.Println(e)
		return nil, e
	}

	if len(f) > 0 {
		this.blocknum = len(f) + 1
		log.Printf("*** Continuing recording from block %d", this.blocknum)
	} else {
		this.blocknum = 1
	}

	this.blockbuffer = bytes.NewBuffer([]byte(nil))
	this.muChan = NewRecorderEventStream()
	// Enable tracking on stuff
	//
	//source.GetMemoryMap().Track[source.GetMemIndex()] = true

	if len(this.memblocks) > 0 {
		// resume recording
		this.blockbuffer = this.memblocks[len(this.memblocks)-1]
		this.memblocks = this.memblocks[0 : len(this.memblocks)-1]
	}

	// Start tracking CPU clock cycles
	ClocksPerSync = CLOCKS_PER_REGISTER_SYNC

	return this, nil
}

func (r *Recorder) RecordJump(from, to int, context string) {

	// don't record jumps when doing live record.
	if r.filename == "" {
		return
	}

	j := &RecorderJumpEvent{
		from:   from,
		to:     to,
		isJump: (context == "JMP"),
	}
	r.muChan.Put(&RecorderEvent{
		jmp: j,
	})
}

func (r *Recorder) GetDeltauS() int64 {
	v := (r.globalCycles - r.lastCycles)
	r.lastCycles = r.globalCycles
	return v
}

func (r *Recorder) AdjustClock(addCycles int) {

}

func (r *Recorder) Decrement(cycles int) {

}

func (r *Recorder) ImA() string {
	return "RECORDER"
}

func (r *Recorder) IsRecording() bool {
	return r.running
}

func (r *Recorder) IsLiveRecording() bool {
	return r.running && r.pathname == ""
}

func (r *Recorder) IsDiscRecording() bool {
	return r.running && r.pathname != ""
}

func (r *Recorder) SetVidMode(mode int) {
	me := &RecorderEvent{
		vidmode: &RecorderVidPageEvent{
			lastmode: r.lastvidmode,
			newmode:  mode,
		},
	}
	r.muChan.Put(me)
	r.lastvidmode = mode
}

func (r *Recorder) SetMemMode(mode int) {

	if r.lastmemmode == mode {
		return
	}

	me := &RecorderEvent{
		memmode: &RecorderMemPageEvent{
			lastmode: r.lastmemmode,
			newmode:  mode,
		},
	}
	r.muChan.Put(me)

	//fmt.RPrintf("--> mode from %s to %s\n", apple2.MemoryFlag(r.lastmemmode), apple2.MemoryFlag(mode))

	r.lastmemmode = mode
}

func (r *Recorder) LogCPU(force bool) {

	mr, _ := r.Source.Memory.InterpreterMappableAtAddress(r.Source.MemIndex, 0xc000)

	if r.cpu == nil {
		r.cpu = apple2helpers.GetCPU(r.Source)
	}

	if r.cpu.RecRegisters.GlobalCycles == r.lastGlobalCycles && !force {
		return
	}

	r.lastGlobalCycles = r.cpu.RecRegisters.GlobalCycles

	io := mr.(*apple2.Apple2IOChip)

	c := freeze.CPURegs{
		A:         r.cpu.RecRegisters.A,
		X:         r.cpu.RecRegisters.X,
		Y:         r.cpu.RecRegisters.Y,
		PC:        r.cpu.RecRegisters.PC,
		SP:        r.cpu.RecRegisters.SP,
		P:         r.cpu.RecRegisters.P,
		SPEED:     int(r.cpu.RecRegisters.RealSpeed),
		ScanCycle: int(io.UnifiedFrame.Clock % 17030), // always from the scanner
	}

	r.lastCPU = &c

	var r6522 = &Recorder6522State{}
	m := io.GetCard(4)
	if m != nil && m.CardName() == "mockingboard" {
		card := m.(*apple2.IOCardMockingBoard)
		r6522.controllers[0] = card.GetChip(0).Bytes()
		r6522.controllers[1] = card.GetChip(1).Bytes()
	}

	var drives []byte
	m = io.GetCard(6)
	if m.CardName() == "IOCardDiskII" {
		drives = m.(*apple2.IOCardDiskII).ToBytes()
	}

	cs := CPUSwitchState{
		CPU:     c,
		vidmode: io.GetVidMode(),
		memmode: io.GetMemMode(),
	}

	// put cpu state on the channel
	r.muChan.Put(
		&RecorderEvent{
			drives: drives,
		},
	)
	r.muChan.Put(
		&RecorderEvent{
			mock: r6522,
		},
	)
	r.muChan.Put(&RecorderEvent{
		cpu: &cs,
	})
	r.muChan.Put(
		&RecorderEvent{
			mock: r6522,
		},
	)
	r.muChan.Put(
		&RecorderEvent{
			drives: drives,
		},
	)
}

var ClocksPerSync int64 = CLOCKS_PER_REGISTER_SYNC

func (r *Recorder) Increment(cycles int) {
	r.globalCycles += int64(cycles)
	r.accCycles += cycles

	if r.accCycles >= int(ClocksPerSync) {
		r.LogCPU(false)

		// deduct cycles
		r.accCycles -= int(ClocksPerSync)

		// check and update
		ClocksPerSync = 1020480 / int64(settings.CPURecordTimingPoints)
	} else if r.allCPUStates {
		r.LogCPUDelta()
	}
}

// Start a recording
func (r *Recorder) Start() {

	//return
	//os.Exit(1)
	settings.RecordIgnoreIO[r.Source.GetMemIndex()] = true

	r.running = true

	r.Source.StopTheWorld()

	time.Sleep(5 * time.Millisecond)

	// always make CPU first event in a block
	r.LogCPU(false)
	//if !r.usemem {
	r.FreezeInitial()
	//}

	r.start = time.Now()
	r.lastupdate = r.start

	// enable push based log tracking
	//r.Source.Memory.EnableLogTracking(r.Source.MemIndex, r.Receive)
	r.Source.Memory.SetRecordLogger(r.Source.MemIndex, r.Receive, r.ReceiveRawAudio)

	r.Source.ResumeTheWorld()

	r.Source.SetCycleCounter(r)

	r.SetInitialFlags()

	go r.Poll()

	// We register here so if unified changes during recording we don't care
	if settings.UnifiedRender[r.Source.MemIndex] {
		servicebus.UnsubscribeType(r.Source.MemIndex, servicebus.UnifiedScanUpdate)
		servicebus.Subscribe(r.Source.MemIndex, servicebus.UnifiedScanUpdate, r)
		servicebus.UnsubscribeType(r.Source.MemIndex, servicebus.UnifiedVBLANK)
		servicebus.Subscribe(r.Source.MemIndex, servicebus.UnifiedVBLANK, r)
	}

	servicebus.UnsubscribeType(r.Source.MemIndex, servicebus.RecorderTerminate)
	servicebus.Subscribe(r.Source.MemIndex, servicebus.RecorderTerminate, r)

	servicebus.SendServiceBusMessage(
		r.Source.GetMemIndex(),
		servicebus.LiveRewindStateUpdate,
		&debugtypes.LiveRewindState{
			Enabled:    true,
			CanBack:    true,
			CanForward: false,
			CanResume:  false,
		},
	)
}

func (r *Recorder) SetInitialFlags() {

	mr, ok := r.Source.Memory.InterpreterMappableAtAddress(r.Source.MemIndex, 0xc000)
	if ok {
		r.lastmemmode = int(mr.(*apple2.Apple2IOChip).GetMemMode())
		r.lastvidmode = int(mr.(*apple2.Apple2IOChip).GetVidMode())
	}

}

// WriteBlock encodes a block for output
func (r *Recorder) WriteBlock(delta int, msg *ducktape.DuckTapeBundle) {

	if delta == 0 {
		delta = int(r.GetDeltauS())
	}

	// if r.counters == nil {
	// 	r.counters = make(map[string]int)
	// }
	// r.counters[msg.ID]++

	rawmsg, _ := msg.MarshalBinary()

	sz := len(rawmsg) - 2 // deduct 13/10 -- not needed here

	if sz == 0 {
		panic("Zero size detected...")
	}

	// we are about to commit the message -- check if we want a boundary
	if msg.ID == "CPU" && r.blockbuffer.Len() > MAX_BLOCK {
		r.CommitBlock()
	}

	// Why?
	r.Out(
		[]byte{
			byte((delta >> 16) & 0xff),
			byte((delta >> 8) & 0xff),
			byte(delta & 0xff),
		},
	)

	// Length for forward seeking
	r.Out(
		[]byte{
			byte((sz >> 16) & 0xff),
			byte((sz >> 8) & 0xff),
			byte(sz & 0xff),
		},
	)

	// Message data
	r.Out(rawmsg[:sz])

	// Length again... for backseeking
	r.Out(
		[]byte{
			byte((sz >> 16) & 0xff),
			byte((sz >> 8) & 0xff),
			byte(sz & 0xff),
		},
	)

	// if len(r.blockbuffer) > MAX_BLOCK {
	// 	r.CommitBlock()
	// }

}

func (r *Recorder) ReceiveRawAudio(c int, rate int, bytepacked bool, indata []uint64) {
	if !settings.CanUserChangeSpeed() {
		return
	}
	//log2.Printf("recording %d samples", c)
	r.muChan.Put(&RecorderEvent{
		audio: RecorderAudioEvent{
			count:      c,
			rate:       rate,
			bytepacked: bytepacked,
			data:       indata,
		},
	})
}

func (r *Recorder) Receive(block *memory.MemoryChange) {

	//fmt.Printf("Memory update @ %d\n", block.Global)

	r.muChan.Put(&RecorderEvent{
		mem: []memory.MemoryChange{*block},
	})

}

// Stop sets a signal to tell the recorder to
func (r *Recorder) Stop() {

	if !r.running {
		return
	}

	// Ask CPU to stop
	if r.cpu == nil {
		r.cpu = apple2helpers.GetCPU(r.Source)
	}
	r.cpu.RequestSuspend()

	for time.Since(r.muChan.lastPut) < 5*time.Millisecond {
		time.Sleep(1 * time.Millisecond)
	}

	r.LogCPU(true) // force the CPU position

	r.muChan.Put(&RecorderEvent{
		stop: true,
	})

	r.Source.ClearCycleCounter(r)
	r.Source.Memory.SetRecordLogger(r.Source.MemIndex, nil, nil)
	r.Source.Memory.DisableLogTracking(r.Source.MemIndex)
	settings.RecordIgnoreIO[r.Source.GetMemIndex()] = false
	servicebus.Unsubscribe(r.Source.MemIndex, r)

	//r.muChan.Put(&RecorderEvent{ful: true})

	if settings.DebuggerAttachSlot-1 == r.Source.GetMemIndex() {
		cpu := apple2helpers.GetCPU(r.Source)
		if cpu.RunState != mos6502.CrsFreeRun {
			r.LogCPU(false)
		}
	} else {
		for r.running && !r.Source.GetMemoryMap().IntGetSlotRestart(r.Source.MemIndex) {
			time.Sleep(5 * time.Millisecond)
		}
	}

	servicebus.SendServiceBusMessage(
		r.Source.GetMemIndex(),
		servicebus.LiveRewindStateUpdate,
		&debugtypes.LiveRewindState{
			Enabled:    false,
			CanBack:    false,
			CanForward: false,
			CanResume:  false,
		},
	)

	r.Source.ResumeTheWorld()

}

// Stop recording
//func (r *Recorder) stop() {
//
//	//return
//
//	if !r.running {
//		return
//	}
//
//	// Generate Freeze packet again
//
//	r.Source.StopTheWorld()
//	// wait for it to drain
//	log.Println("Waiting for record buffer to drain...")
//	for r.muChan.CanGet() {
//		time.Sleep(1 * time.Millisecond)
//	}
//
//	if !r.unified {
//		r.LogCPU()
//	}
//	//r.FreezeInitial()
//
//	// wait for it to drain
//	log.Println("Waiting for record buffer to drain...")
//	for r.muChan.CanGet() {
//		time.Sleep(1 * time.Millisecond)
//	}
//
//	// Do the actual stop
//	r.running = false
//	r.Source.ClearCycleCounter(r)
//	r.Source.Memory.SetRecordLogger(r.Source.MemIndex, nil, nil)
//	r.Source.Memory.DisableLogTracking(r.Source.MemIndex)
//	r.CommitBlock()
//
//	r.Source.ResumeTheWorld()
//
//	settings.RecordIgnoreIO[r.Source.GetMemIndex()] = false
//
//	servicebus.SendServiceBusMessage(
//		r.Source.GetMemIndex(),
//		servicebus.LiveRewindStateUpdate,
//		&debugtypes.LiveRewindState{
//			Enabled:    false,
//			CanBack:    false,
//			CanForward: false,
//			CanResume:  false,
//		},
//	)
//}

func (r *Recorder) CommitBlock() {

	// fmt.RPrintf("Commit block of %d bytes: %v\n", r.blockbuffer.Len(), r.counters)
	// if r.counters["BMU"] > 50000 {
	// 	fmt.RPrintf("BMU stats = %v\n", r.mcounters)
	// }
	// r.counters = nil
	// r.mcounters = nil
	var newbuffer *bytes.Buffer

	if r.usemem {

		if len(r.memblocks) < MAX_MEM_BLOCKS {
			r.memblocks = append(r.memblocks, r.blockbuffer)
		} else {
			newbuffer = r.memblocks[0]
			for i := 1; i < len(r.memblocks); i++ {
				r.memblocks[i-1] = r.memblocks[i]
			}
			r.memblocks[MAX_MEM_BLOCKS-1] = r.blockbuffer
		}

	} else {
		//fmt.Printf("---> Saving block %s/%s\n", r.pathname, bfn(r.blocknum))
		//go func(path, fn string, data []byte) {
		// e := files.WriteBytesViaProvider(r.pathname, bfn(r.blocknum), r.blockbuffer.Bytes())
		// if e != nil {
		// 	fmt.Println(e)
		// 	os.Exit(1)
		// }
		//}(r.pathname, bfn(r.blocknum), r.blockbuffer.Bytes())
		go func(fn string, data []byte) {
			// write the block
			path := files.GetUserPath(files.BASEDIR, []string{"MyRecordings", files.GetFilename(r.pathname), fn})
			//log2.Printf("writing file: %s", path)
			ioutil.WriteFile(path, data, 0755)
		}(bfn(r.blocknum), r.blockbuffer.Bytes())
	}
	r.blocknum++
	if newbuffer != nil {
		r.blockbuffer = newbuffer
		r.blockbuffer.Reset()
	} else {
		r.blockbuffer = bytes.NewBuffer([]byte(nil))
	}
}

func (r *Recorder) Out(d []byte) {
	r.blockbuffer.Write(d)
}

// func (r *Recorder) CheckDisk(drive int) {
// 	var prev, current, evt string
// 	var idx int
// 	if drive == 0 {
// 		evt = "DR0"
// 		prev = r.d0
// 		current = settings.PureBootVolume[r.Source.MemIndex]
// 	} else {
// 		idx = 1
// 		evt = "DR1"
// 		prev = r.d1
// 		current = settings.PureBootVolume2[r.Source.MemIndex]
// 	}
// 	if prev != current {
// 		dsk := apple2.GetDisk(r.Source, idx)
// 		var payload []byte
// 		if dsk != nil {
// 			payload = dsk.GetNibbles()
// 		} else {
// 			payload = make([]byte, disk.DISK_NIBBLE_LENGTH)
// 		}
// 		msg := ducktape.DuckTapeBundle{
// 			ID:      evt,
// 			Binary:  true,
// 			Payload: utils.GZIPBytes(payload),
// 		}

// 		switch idx {
// 		case 0:
// 			r.d0 = current
// 		case 1:
// 			r.d1 = current
// 		}

// 		r.WriteBlock(0, &msg)

// 	}
// }

func (r *Recorder) UnifiedScanState() {
	if settings.UnifiedRender[r.Source.MemIndex] {
		// We should encode an initial state of the scan memory
		mr, ok := r.Source.Memory.InterpreterMappableAtAddress(r.Source.MemIndex, 0xc000)
		if ok {
			data := mr.(*apple2.Apple2IOChip).UnifiedFrame.SaveState()
			msg := ducktape.DuckTapeBundle{
				ID:      "USS",
				Binary:  true,
				Payload: data,
			}

			r.WriteBlock(1, &msg)
		}

	}
}

// FreezeInitial takes a snapshot of the start state of the interpreter
func (r *Recorder) FreezeInitial() error {

	//fmt.Println("freeze state")

	if r.blocknum > 1 {
		return nil
	}

	fr := freeze.NewFreezeState(r.Source, false)

	payload := fr.SaveToBytes()

	msg := ducktape.DuckTapeBundle{
		ID:      "FUL",
		Binary:  true,
		Payload: utils.GZIPBytes(payload),
	}

	r.WriteBlock(0, &msg)

	r.UnifiedScanState()

	// r.CheckDisk(0)
	// r.CheckDisk(1)
	//r.CheckCamera()

	return nil
}

// Poll loops, polling for recorded memory changes
func (r *Recorder) Poll() {

	defer func() {
		r.running = false
	}()

	var timeSinceVBL int

	for r.running && !r.Source.GetMemoryMap().IntGetSlotRestart(r.Source.GetMemIndex()) {
		if r.muChan.CanGet() {
			// Handle CPU register state sync point (includes video / mmu states)
			if r.handleMessage(timeSinceVBL) {
				r.CommitBlock()
				r.running = false
				return
			}
		} else {
			time.Sleep(1 * time.Millisecond)
		}
	}

}

func (r *Recorder) handleMessage(timeSinceVBL int) bool {
	ev := r.muChan.Get()
	//r.EPS++

	if ev != nil && !r.running {
		log2.Printf("Processing event %d of max %d", ev.seq, r.muChan.seq-1)
	}

	if ev.ncpu != nil {

		// order PC, A, X, Y, SP, P
		payload := ev.ncpu.ToBytes()
		if len(payload) > 0 {
			msg := ducktape.DuckTapeBundle{
				ID:      "CPD",
				Binary:  true,
				Payload: payload,
			}

			r.WriteBlock(0, &msg)
		}

	} else if ev.stop {
		r.CommitBlock()
		log2.Printf("Stopped recording!")
		return true
	} else if ev.vblank != nil {
		msg := ducktape.DuckTapeBundle{
			ID:      "UVB",
			Binary:  true,
			Payload: []byte{byte(*ev.vblank / 256), byte(*ev.vblank % 256)},
		}
		r.WriteBlock(0, &msg)
		timeSinceVBL = 0

	} else if ev.scandata != nil {
		msg := ducktape.DuckTapeBundle{
			ID:      "USD",
			Binary:  true,
			Payload: ev.scandata,
		}
		r.WriteBlock(0, &msg)
	} else if ev.mock != nil {

		payload := append(ev.mock.controllers[0], ev.mock.controllers[1]...)
		msg := ducktape.DuckTapeBundle{
			ID:      "MCK",
			Binary:  true,
			Payload: payload,
		}

		r.WriteBlock(0, &msg)

	} else if ev.cpu != nil {

		s := ev.cpu

		payload := make([]byte, int(CSS_SIZE)*4)
		var uval, offset int
		var v CSSFlag
		for i := 0; i < int(CSS_SIZE); i++ {
			v = CSSFlag(i)
			switch v {
			case CSS_CPU_SPEED:
				uval = s.CPU.SPEED
			case CSS_CPU_PC:
				uval = s.CPU.PC
			case CSS_CPU_P:
				uval = s.CPU.P
			case CSS_CPU_SP:
				uval = s.CPU.SP
			case CSS_CPU_A:
				uval = s.CPU.A
			case CSS_CPU_X:
				uval = s.CPU.X
			case CSS_CPU_Y:
				uval = s.CPU.Y
			case CSS_MMU_MEMMODE:
				uval = int(s.memmode)
			case CSS_MMU_VIDMODE:
				uval = int(s.vidmode)
			case CSS_CPU_SCANOFFSET:
				uval = int(s.CPU.ScanCycle)
			}
			// Convert to byte payload
			offset = i * 4

			payload[offset+0] = byte((uval >> 24) & 0xff)
			payload[offset+1] = byte((uval >> 16) & 0xff)
			payload[offset+2] = byte((uval >> 8) & 0xff)
			payload[offset+3] = byte(uval & 0xff)
		}

		msg := ducktape.DuckTapeBundle{
			ID:      "CPU",
			Binary:  true,
			Payload: payload,
		}

		r.WriteBlock(0, &msg)

	} else if len(ev.mem) > 0 {

		data := ev.mem

		count := 0

		if len(data) > 0 {

			payload := make([]byte, 3)

			for _, mc := range data {

				a := (mc.Global % memory.OCTALYZER_INTERPRETER_SIZE)
				if a >= 0x41000 && a < 0x42000 || a >= memory.OCTALYZER_HUD_BASE+memory.OCTALYZER_LAYERSPEC_SIZE*3 && a < memory.OCTALYZER_HUD_BASE+memory.OCTALYZER_LAYERSPEC_SIZE*4 {
					continue
				}

				if mc.Value == nil || len(mc.Value) == 0 {
					//payload = append(payload, mempak.Encode(0, mc.Global, 0, true)...)
					//count++
				} else {
					actual := len(mc.Value) / 2

					for i, v := range mc.Value {
						payload = append(payload, mempak.Encode(0, mc.Global+(i%actual), v, false)...)
						count++

						// if r.mcounters == nil {
						// 	r.mcounters = make(map[int]int)
						// }
						// r.mcounters[mc.Global+(i%actual)]++
					}

				}

			}

			if count > 0 {

				// d := time.Since(r.lastupdate) / time.Microsecond
				// r.lastupdate = time.Now()
				d := r.GetDeltauS()

				if r.Source.IsWaitingForWorld() {
					d = 0
				}

				payload[0] = byte((count / 65536) % 256)
				payload[1] = byte((count / 256) % 256)
				payload[2] = byte(count % 256)

				msg := ducktape.DuckTapeBundle{
					ID:      "BMU",
					Binary:  true,
					Payload: payload,
				}

				r.WriteBlock(int(d), &msg)
			}
		}
	} else if ev.memmode != nil {

		mm := *ev.memmode

		//fmt.RPrintf("Memmode: %s\n", apple2.MemoryFlag(mm.newmode).String())

		payload := make([]byte, 8)
		payload[0] = byte((mm.lastmode >> 24) & 0xff)
		payload[1] = byte((mm.lastmode >> 16) & 0xff)
		payload[2] = byte((mm.lastmode >> 8) & 0xff)
		payload[3] = byte(mm.lastmode & 0xff)
		payload[4] = byte((mm.newmode >> 24) & 0xff)
		payload[5] = byte((mm.newmode >> 16) & 0xff)
		payload[6] = byte((mm.newmode >> 8) & 0xff)
		payload[7] = byte(mm.newmode & 0xff)

		msg := ducktape.DuckTapeBundle{
			ID:      "MMD",
			Binary:  true,
			Payload: payload,
		}

		r.WriteBlock(0, &msg)

	} else if ev.drives != nil {

		msg := ducktape.DuckTapeBundle{
			ID:      "DII",
			Binary:  true,
			Payload: ev.drives,
		}

		r.WriteBlock(0, &msg)

	} else if ev.jmp != nil && r.filename != "" {

		mm := *ev.jmp

		payload := make([]byte, 9)
		payload[0] = byte((mm.from >> 24) & 0xff)
		payload[1] = byte((mm.from >> 16) & 0xff)
		payload[2] = byte((mm.from >> 8) & 0xff)
		payload[3] = byte(mm.from & 0xff)
		payload[4] = byte((mm.to >> 24) & 0xff)
		payload[5] = byte((mm.to >> 16) & 0xff)
		payload[6] = byte((mm.to >> 8) & 0xff)
		payload[7] = byte(mm.to & 0xff)
		z := byte(0)
		if mm.isJump {
			z = 1
		}
		payload[8] = byte(z)
		msg := ducktape.DuckTapeBundle{
			ID:      "JMP",
			Binary:  true,
			Payload: payload,
		}

		r.WriteBlock(0, &msg)

		//fmt.RPrintf("Log jump... %d\n", mm.to)

	} else if ev.vidmode != nil {

		mm := *ev.vidmode

		payload := make([]byte, 8)
		payload[0] = byte((mm.lastmode >> 24) & 0xff)
		payload[1] = byte((mm.lastmode >> 16) & 0xff)
		payload[2] = byte((mm.lastmode >> 8) & 0xff)
		payload[3] = byte(mm.lastmode & 0xff)
		payload[4] = byte((mm.newmode >> 24) & 0xff)
		payload[5] = byte((mm.newmode >> 16) & 0xff)
		payload[6] = byte((mm.newmode >> 8) & 0xff)
		payload[7] = byte(mm.newmode & 0xff)

		msg := ducktape.DuckTapeBundle{
			ID:      "VMD",
			Binary:  true,
			Payload: payload,
		}

		r.WriteBlock(0, &msg)

	} else if ev.ful {
		// create a full sync here
		//r.FreezeInitial()
		fr := freeze.NewFreezeState(r.Source, false)

		payload := fr.SaveToBytes()

		msg := ducktape.DuckTapeBundle{
			ID:      "FUL",
			Binary:  true,
			Payload: utils.GZIPBytes(payload),
		}

		r.WriteBlock(0, &msg)
		return true
	} else if ev.audio.count > 0 && !settings.RecordIgnoreAudio[r.Source.GetMemIndex()] {

		//log2.Printf("encode audio block")

		// d := time.Since(r.lastupdate) / time.Microsecond
		// r.lastupdate = time.Now()
		d := r.GetDeltauS()

		if r.Source.IsWaitingForWorld() {
			d = 0
		}

		// count [3]bytes
		// rate  [2]bytes
		// bytepacked [1]bytes
		// data...

		payload := make([]byte, 6+4*ev.audio.count)
		// count
		payload[0] = byte((ev.audio.count / 65536) % 256)
		payload[1] = byte((ev.audio.count / 256) % 256)
		payload[2] = byte(ev.audio.count % 256)
		// rate
		payload[3] = byte((ev.audio.rate / 256) % 256)
		payload[4] = byte(ev.audio.rate % 256)
		// bytepacked
		if ev.audio.bytepacked {
			payload[5] = 1
		}

		// encode the data
		for i, v := range ev.audio.data {
			payload[6+i*4+0] = byte((v >> 24) & 0xff)
			payload[6+i*4+1] = byte((v >> 16) & 0xff)
			payload[6+i*4+2] = byte((v >> 8) & 0xff)
			payload[6+i*4+3] = byte(v & 0xff)
		}

		msg := ducktape.DuckTapeBundle{
			ID:      "SND",
			Binary:  true,
			Payload: payload,
		}

		r.WriteBlock(int(d), &msg)
	}

	if ev.vblank == nil {
		timeSinceVBL++
	}
	return false
}

// Servicebus
func (c *Recorder) HandleServiceBusRequest(r *servicebus.ServiceBusRequest) (*servicebus.ServiceBusResponse, bool) {
	switch r.Type {
	case servicebus.RecorderTerminate:
		go c.Stop()
	case servicebus.UnifiedVBLANK:
		cpu := apple2helpers.GetCPU(c.Source)
		s := int(cpu.RecRegisters.GlobalCycles % 17030)
		c.muChan.Put(&RecorderEvent{
			vblank: &s,
		})
		//log2.Printf("scan offset logged = %d", c.lastGlobalCycles%17030)

	case servicebus.UnifiedScanUpdate:
		if data, ok := r.Payload.([]byte); ok && len(data) == 10 {
			c.muChan.Put(&RecorderEvent{
				scandata: data,
			})
		}
	}

	return &servicebus.ServiceBusResponse{
		Payload: "",
	}, true
}

func (c *Recorder) InjectServiceBusRequest(r *servicebus.ServiceBusRequest) {
	log.Printf("Injecting ServiceBus request: %+v", r)
	c.m.Lock()
	defer c.m.Unlock()
	if c.injectedBusRequests == nil {
		c.injectedBusRequests = make([]*servicebus.ServiceBusRequest, 0, 16)
	}
	c.injectedBusRequests = append(c.injectedBusRequests, r)
}

func (c *Recorder) HandleServiceBusInjection(handler func(r *servicebus.ServiceBusRequest) (*servicebus.ServiceBusResponse, bool)) {
	if c.injectedBusRequests == nil || len(c.injectedBusRequests) == 0 {
		return
	}
	c.m.Lock()
	defer c.m.Unlock()
	for _, r := range c.injectedBusRequests {
		if handler != nil {
			handler(r)
		}
	}
	c.injectedBusRequests = make([]*servicebus.ServiceBusRequest, 0, 16)
}

func (c *Recorder) ServiceBusProcessPending() {
	c.HandleServiceBusInjection(c.HandleServiceBusRequest)
}
