package interpreter

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"sync"
	"time" //"log"

	"github.com/atotto/clipboard"
	s8webclient "paleotronic.com/api"
	"paleotronic.com/core/debug"
	"paleotronic.com/core/dialect/appleinteger"
	"paleotronic.com/core/dialect/applesoft"
	"paleotronic.com/core/dialect/logo"
	"paleotronic.com/core/dialect/plus"
	"paleotronic.com/core/dialect/shell"
	"paleotronic.com/core/editor"
	"paleotronic.com/core/exception"
	"paleotronic.com/core/hardware/apple2"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/hardware/control"
	"paleotronic.com/core/hardware/cpu"
	"paleotronic.com/core/hardware/servicebus"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/syncmanager"
	"paleotronic.com/core/types" //"paleotronic.com/core/types/glmath"
	"paleotronic.com/core/vduconst"
	"paleotronic.com/decoding"
	"paleotronic.com/encoding/mempak"
	"paleotronic.com/fastserv"
	"paleotronic.com/fastserv/client"
	"paleotronic.com/filerecord"
	"paleotronic.com/files"
	"paleotronic.com/fmt"
	"paleotronic.com/freeze" //	"bufio"
	"paleotronic.com/log"
	"paleotronic.com/microtracker/mock"
	"paleotronic.com/microtracker/tracker"
	"paleotronic.com/octalyzer/bus"
	"paleotronic.com/runestring"
	"paleotronic.com/update"
	"paleotronic.com/utils"
)

type ChatMessage struct {
	Sender  string
	Message string
}

type QueryResult struct {
	Result string
	Err    error
}

type Interpreter struct {
	vm interfaces.VM

	UUID                uint64
	isMicroControl      bool
	music               *decoding.AudioPlayer
	CycleCounter        []interfaces.Countable
	Disabled            bool
	clientSync          bool
	InsertPos           int
	Px, Py              int
	LastX               int
	PC                  types.CodeRef
	RefList             types.ReferenceList
	IsolateVars         bool
	MultiArgFunc        interfaces.MafMap
	Name                string
	firstString         string
	Producer            interfaces.Producable
	CreatedTokens       *types.TokenList
	LastZ               int
	Code                *types.Algorithm
	CodeOptimizations   types.Algorithm
	WaitUntil           time.Time
	VarPrefix           string
	LoopStack           types.LoopStack
	WorkDir, ProgramDir string
	DirectAlgorithm     *types.Algorithm
	Dialect             interfaces.Dialecter
	Stack               interfaces.CallStack
	Registers           types.BRegisters
	LoopStep            float64
	LoopVariable        string
	LastY               int
	ignoreMyAudio       bool
	State               types.EntityState
	SubState            types.EntitySubState
	LPC                 types.CodeRef
	Breakpoint          types.CodeRef
	ErrorTrap           types.CodeRef
	OuterVars           bool
	DataMap             *types.TokenMap
	Data                types.CodeRef
	LoopStates          types.LoopStateMap
	Bearing             int
	TokenStack          types.TokenList
	LastExecuted        int64
	OutChannel          string
	InChannel           string
	FirstString         string
	Memory              *memory.MemoryMap
	Parent              interfaces.Interpretable
	Children            interfaces.Interpretable
	History             []runestring.RuneString
	HistIndex           int
	StartTime           time.Time
	c                   *s8webclient.Client
	MemIndex            int
	Params              *types.TokenList
	ExitOnEnd           bool
	DebugMe             bool
	Silent              bool
	RunStart            int64
	HUDLayers           []*types.LayerSpecMapped
	GFXLayers           []*types.LayerSpecMapped
	Prompt              string
	TabWidth            int
	CommandBuffer       runestring.RuneString
	NeedsPrompt         bool
	FeedBuffer          string
	SpecFile            string
	DisplayPage         string
	CurrentPage         string
	Breakable           bool
	DosBuffer           string
	DosCommand          bool
	NextBytecolor       bool
	Speed               int
	CharacterCapture    string
	LastChar            rune
	SuppressFormat      bool
	PasteBuffer         runestring.RuneString
	LastBuffer          runestring.RuneString
	Remote              *client.FSClient
	RemoteIndex         int // remote memindex
	RemIntIndex         int
	ChatMessages        []ChatMessage
	ChatMutex           sync.Mutex
	InputMapper         *InputMatrix
	ResultQueue         chan QueryResult
	MetaData            filerecord.FileRecord
	IgnoreSpecial       bool
	Local               types.VarManager
	CallReturnToken     *types.Token
	HiRemote, LoRemote  int
	r                   *Recorder
	p                   *Player
	lastProfile         int
	Paused              bool
	clist               *types.TokenList
	cprefix             int
	cptr                int
	wantcompletion      bool
	compupper           bool
	PreCPUState         types.EntityState
	CurrentCommandState *interfaces.CommandState
	DiskImage           string
	StatusText          string
	VideoStatusText     string
	audiochannels       map[string]int

	NeedRemoteKill  bool
	SaveRestoreText bool

	Labels map[string]int

	Triggers *TriggerTable

	LastPasteTime time.Time

	CurrentSubroutine string

	injectedBusRequests []*servicebus.ServiceBusRequest
	m                   sync.Mutex

	Semaphore string

	song *tracker.TSong

	UsePromptColor bool
	PromptColor    uint64

	bsMutex sync.Mutex
}

func (this *Interpreter) VM() interfaces.VM {
	if this.Parent != nil {
		return this.Parent.VM()
	}
	return this.vm
}

func (this *Interpreter) Bind(vm interfaces.VM) {
	this.vm = vm
}

func (this *Interpreter) SetSemaphore(s string) {
	this.Semaphore = s
}

func (this *Interpreter) GetSemaphore() string {
	return this.Semaphore
}

func (this *Interpreter) SetCurrentSubroutine(s string) {
	this.CurrentSubroutine = s
}

func (this *Interpreter) GetCurrentSubroutine() string {
	return this.CurrentSubroutine
}

func (this *Interpreter) ClearAudioPorts() {
	this.audiochannels = make(map[string]int)
}

func (this *Interpreter) GetAudioPort(name string) int {
	return this.audiochannels[name]
}

func (this *Interpreter) SetAudioPort(name string, port int) {
	this.audiochannels[name] = port
}

func (this *Interpreter) GetUUID() uint64 {
	return this.UUID
}

func (this *Interpreter) SetUUID(u uint64) {
	this.UUID = u
}

func (this *Interpreter) IsMicroControl() bool {
	return this.isMicroControl
}

func (this *Interpreter) SetMicroControl(b bool) {
	this.isMicroControl = b
}

func (this *Interpreter) PlayMusic(p, f string, leadin int, fadein int) error {
	this.StopMusic()
	fr, err := files.ReadBytesViaProvider(p, f)
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Printf("%s/%s exists. Buffering...\n", p, f)

	b := bytes.NewBuffer(fr.Content)

	switch files.GetExt(f) {

	case "wav", "ogg":
		this.music, err = decoding.NewPlayer(this, b, leadin, fadein)
		if err != nil {
			fmt.Println(err)
			return err
		}
		fmt.Println("Starting playback")
		if fadein != 0 {
			this.SetMusicVolume(0, false)
		} else {
			this.SetMusicVolume(1, true)
		}
		this.music.Start()
	case "sng":
		drv := mock.New(this, 0xc400)
		this.song = tracker.NewSong(120, drv)
		err = this.song.Load("/" + strings.Trim(p+"/"+f, "/"))
		if err != nil {
			return err
		}
		this.song.Start(tracker.PMLoopSong)
		this.song.SetPlayMode(tracker.PMLoopSong)
	}

	return nil
}

func (this *Interpreter) StopMusic() {
	if this.music != nil {
		this.music.Stop()
	}
	if this.song != nil {
		//dbg.PrintStack()
		this.song.Stop()
	}
	if plus.TrackerSong[this.MemIndex] != nil {
		plus.TrackerSong[this.MemIndex].Stop()
	}
}

func (this *Interpreter) SetMusicPaused(paused bool) {
	if this.music != nil {
		this.music.SetPaused(paused)
	}
}

func (this *Interpreter) IgnoreMyAudio() bool {
	if this.Children != nil {
		return this.Children.IgnoreMyAudio()
	}
	return this.ignoreMyAudio
}

func (this *Interpreter) SetIgnoreMyAudio(b bool) {
	if this.Children != nil {
		this.Children.SetIgnoreMyAudio(b)
		return
	}
	this.ignoreMyAudio = b
}

func (this *Interpreter) GetPlayer() interfaces.Playable {
	return this.p
}

func (this *Interpreter) IsPlayingVideo() bool {
	if this.Children != nil {
		return this.Children.IsPlayingVideo()
	}
	return (this.p != nil) && (this.p.IsPlaying())
}

func (this *Interpreter) IsRecordingVideo() bool {
	if this.Children != nil {
		return this.Children.IsRecordingVideo()
	}
	return (this.r != nil) && (this.r.IsRecording())
}

func (this *Interpreter) IsRecordingDiscVideo() bool {
	if this.Children != nil {
		return this.Children.IsRecordingDiscVideo()
	}
	return (this.r != nil) && (this.r.IsDiscRecording())
}

func (this *Interpreter) ReplayVideo() bool {
	if this.Children != nil {
		return this.Children.ReplayVideo()
	}

	if this.IsPlayingVideo() {
		settings.VideoPlaybackPauseOnFUL[this.MemIndex] = true
		this.p.ResetToStart()
	} else if this.IsRecordingDiscVideo() {
		// we are recording, so (1) switch to playback in reverse
		cpu := apple2helpers.GetCPU(this)
		this.r.Stop() // stop implicitly creates a freeze point
		if !settings.PureBoot(this.MemIndex) && !cpu.Halted {
			//fmt.RPrintln("In micromode, so halting CPU and waiting for stop before we rewind")
			cpu.Halted = true
			for this.State == types.EXEC6502 || this.State == types.DIRECTEXEC6502 {
				time.Sleep(5 * time.Millisecond)
			}
			//fmt.RPrintln("CPU has stopped, we can change modes safely...")
		}
		if this.r.usemem {
			blocks := this.r.memblocks
			settings.SetPureBoot(this.MemIndex, false)
			settings.VideoPlayFrames[this.MemIndex] = blocks
			settings.VideoPlayBackwards[this.MemIndex] = false
			settings.VideoPlaybackPauseOnFUL[this.MemIndex] = true
			this.PreCPUState = this.State
			this.SetState(types.PLAYING)
		} else {
			//rlog.Printf("*** Rewinding using file: %s", this.r.pathname)
			settings.SetPureBoot(this.MemIndex, false)
			settings.VideoPlaybackFile[this.MemIndex] = this.r.pathname
			settings.VideoPlayBackwards[this.MemIndex] = false
			settings.VideoPlaybackPauseOnFUL[this.MemIndex] = true
			this.PreCPUState = this.State
			this.SetState(types.PLAYING)
		}
		return true
	}

	return false
}

func (this *Interpreter) BackstepVideo(ms int) {
	if this.Children != nil {
		this.Children.BackstepVideo(ms)
		return
	}

	if this.State == types.PLAYING && this.IsPlayingVideo() && this.GetPlayer().IsSeeking() {
		go servicebus.SendServiceBusMessage(
			this.MemIndex,
			servicebus.PlayerBackstep,
			5000,
		)
		return
	}

	if !this.IsRecordingVideo() {
		return
	}

	if this.IsRecordingVideo() && !this.IsPlayingVideo() && (this.State == types.EXEC6502 || this.State == types.DIRECTEXEC6502) {
		// we are recording, so (1) switch to playback in reverse
		cpu := apple2helpers.GetCPU(this)
		this.r.Stop() // stop implicitly creates a freeze point
		if !settings.PureBoot(this.MemIndex) && !cpu.Halted {
			//fmt.RPrintln("In micromode, so halting CPU and waiting for stop before we rewind")
			cpu.Halted = true
			for this.State == types.EXEC6502 || this.State == types.DIRECTEXEC6502 {
				time.Sleep(5 * time.Millisecond)
			}
			//fmt.RPrintln("CPU has stopped, we can change modes safely...")
		}
		if this.r.usemem {
			blocks := this.r.memblocks
			settings.SetPureBoot(this.MemIndex, false)
			settings.VideoPlayFrames[this.MemIndex] = blocks
			settings.VideoPlayBackwards[this.MemIndex] = true
			settings.VideoBackSeekMS[this.MemIndex] = ms
			this.PreCPUState = this.State
			this.SetState(types.PLAYING)
		} else {
			//rlog.Printf("*** Rewinding using file: %s", this.r.pathname)
			settings.SetPureBoot(this.MemIndex, false)
			settings.VideoPlaybackFile[this.MemIndex] = this.r.pathname
			settings.VideoPlayBackwards[this.MemIndex] = true
			settings.VideoBackSeekMS[this.MemIndex] = ms
			this.PreCPUState = this.State
			this.SetState(types.PLAYING)
		}
		//this.r = nil
	}
}

func (this *Interpreter) BackVideo() bool {

	if this.Children != nil {
		return this.Children.BackVideo()
	}

	if this.IsPlayingVideo() {
		// we are in playback so reverse playback
		if !this.GetPlayer().IsBackwards() {
			// reverse it
			if this.GetPlayer().GetTimeShift() == 0 {
				this.GetPlayer().SetBackwards(true)
				this.GetPlayer().Faster()
			} else {
				this.GetPlayer().Slower()
			}
			return true
		} else {
			this.GetPlayer().Faster()
		}
	} else if this.IsRecordingVideo() {
		// we are recording, so (1) switch to playback in reverse
		cpu := apple2helpers.GetCPU(this)
		this.r.Stop() // stop implicitly creates a freeze point
		if !settings.PureBoot(this.MemIndex) && !cpu.Halted {
			//fmt.RPrintln("In micromode, so halting CPU and waiting for stop before we rewind")
			cpu.Halted = true
			for this.State == types.EXEC6502 || this.State == types.DIRECTEXEC6502 {
				time.Sleep(5 * time.Millisecond)
			}
			//fmt.RPrintln("CPU has stopped, we can change modes safely...")
		}
		if this.r.usemem {
			blocks := this.r.memblocks
			settings.SetPureBoot(this.MemIndex, false)
			settings.VideoPlayFrames[this.MemIndex] = blocks
			settings.VideoPlayBackwards[this.MemIndex] = true
			this.PreCPUState = this.State
			this.SetState(types.PLAYING)
		} else {
			//rlog.Printf("*** Rewinding using file: %s", this.r.pathname)
			settings.SetPureBoot(this.MemIndex, false)
			settings.VideoPlaybackFile[this.MemIndex] = this.r.pathname
			settings.VideoPlayBackwards[this.MemIndex] = true
			this.PreCPUState = this.State
			this.SetState(types.PLAYING)
		}
		return true
	}
	return false
}

func (this *Interpreter) ForwardVideo() bool {

	if this.Children != nil {
		return this.Children.ForwardVideo()
	}

	if this.IsPlayingVideo() {
		// we are in reverse playback so go to forward playback
		if this.GetPlayer().IsBackwards() {
			// reverse it
			if this.GetPlayer().GetTimeShift() == 0 {
				this.GetPlayer().SetBackwards(false)
				this.GetPlayer().Faster()
			} else {
				this.GetPlayer().Slower()
			}
			return true
		} else {
			this.GetPlayer().Faster()
		}
	}
	return false
}

func (this *Interpreter) ForwardVideo1x() bool {

	if this.Children != nil {
		return this.Children.ForwardVideo1x()
	}

	if this.IsPlayingVideo() {
		// we are in reverse playback so go to forward playback
		if this.GetPlayer().IsBackwards() {
			this.GetPlayer().SetBackwards(false)
		}
		// reverse it
		if this.GetPlayer().GetTimeShift() != 1 {
			this.GetPlayer().SetTimeShift(1)
		}

		return true
	}
	return false
}

func (this *Interpreter) SetStatusText(s string) {
	this.StatusText = s
}

func (this *Interpreter) GetStatusText() string {
	return this.StatusText
}

func (this *Interpreter) SetVideoStatusText(s string) {
	this.VideoStatusText = s
}

func (this *Interpreter) GetVideoStatusText() string {
	return this.VideoStatusText
}

func (this *Interpreter) SetPragma(name string) {
	// dummy for pragma set
}

func (this *Interpreter) SetCycleCounter(c interfaces.Countable) {
	// var index = -1
	// for i, v := range this.CycleCounter {
	// 	if v.ImA() == c.ImA() {
	// 		index = i
	// 	}
	// }
	// if index == -1 {
	// 	this.CycleCounter = append(this.CycleCounter, c)
	// } else {
	// 	this.CycleCounter[index] = c
	// }
}

func (this *Interpreter) ClearCycleCounter(c interfaces.Countable) {
	fmt.Println("Removing a cycle counter")
	del := -1
	for i := 0; i < len(this.CycleCounter); i++ {
		if this.CycleCounter[i].ImA() == c.ImA() {
			del = i
			break
		}
	}
	if del != -1 {
		this.CycleCounter = append(this.CycleCounter[:del], this.CycleCounter[del+1:]...)
	}
}

func (this *Interpreter) DeleteCycleCounters() {
	fmt.Println("Removing ALL cycle counters")
	this.CycleCounter = make([]interfaces.Countable, 0)
}

func (this *Interpreter) GetCycleCounter() []interfaces.Countable {
	return this.CycleCounter
}

func (this *Interpreter) SetLabel(name string, line int) {
	this.Labels[strings.ToLower(name)] = line
}

func (this *Interpreter) GetLabel(name string) int {
	return this.Labels[strings.ToLower(name)]
}

func (this *Interpreter) ClearLabels() {
	this.Labels = make(map[string]int)
}

func (this *Interpreter) SetDiskImage(s string) {
	this.DiskImage = s
}

func (this *Interpreter) GetDiskImage() string {
	return this.DiskImage
}

func (this *Interpreter) SetDisabled(b bool) {

	if this.Children != nil {
		this.Children.SetDisabled(b)
		return
	}

	this.Disabled = b

	if b {
		this.Memory.IntSetActiveState(this.MemIndex, 0)
	} else {
		this.Memory.IntSetActiveState(this.MemIndex, 1)
	}

}

func (this *Interpreter) IsDisabled() bool {
	if this.Children != nil {
		return this.Children.IsDisabled()
	}
	return this.Disabled
}

func (this *Interpreter) StopTheWorld() {
	if this.Children != nil {
		this.Children.StopTheWorld()
		return
	}
	this.Paused = true
	time.Sleep(2 * time.Millisecond)
}

func (this *Interpreter) PostJumpEvent(from, to int, context string) {
	if this.r != nil {
		this.r.RecordJump(from, to, context)
	}
}

func (this *Interpreter) ResumeTheWorld() {
	if this.Children != nil {
		this.Children.ResumeTheWorld()
		return
	}
	this.Paused = false
}

func (this *Interpreter) GetPasteBuffer() runestring.RuneString {
	return this.PasteBuffer
}

func (this *Interpreter) GetLastCommand() runestring.RuneString {
	return this.LastBuffer
}

func (this *Interpreter) SetLastCommand(b runestring.RuneString) {
	this.LastBuffer = b
}

func (this *Interpreter) NeedsClock() bool {
	if this.GetChild() != nil {
		return this.GetChild().NeedsClock()
	}
	return (this.State != types.EXEC6502 && this.State != types.DIRECTEXEC6502)
}

func (this *Interpreter) WaitForWorld() bool {

	if this.Children != nil {
		return this.Children.WaitForWorld()
	}

	if this.VM() != nil && this.VM().IsDying() {
		return false
	}

	if this.Memory.IntGetSlotRestart(this.MemIndex) {
		apple2helpers.GetCPU(this).ResetSliced()
		apple2helpers.GetCPU(this).Halted = true
		this.Halt()

		//this.GetProducer().RestartInterpreter(this.MemIndex)
		return false
	}

	if this.Memory.IntGetSlotMenu(this.MemIndex) {

		r := bus.IsClock()

		if !r {
			bus.StartDefault() // resume clock
		}

		editor.TestMenu(this)
		this.Memory.IntSetSlotMenu(this.MemIndex, false)

		if !this.NeedsClock() {
			bus.StopClock()
		}

		return false
	}

	if this.Memory.IntGetSlotInterrupt(this.MemIndex) {

		r := bus.IsClock()

		if !r {
			bus.StartDefault() // resume clock
		}
		sel := control.CatalogPresent(this)
		fmt.Printf("wfw CATALOG SELECTION IS %d\n", sel)
		this.Memory.IntSetSlotInterrupt(this.MemIndex, false)
		apple2helpers.GetCPU(this).ResetSliced()
		if sel > 0 {
			apple2helpers.GetCPU(this).Halted = true
			this.Halt()
			fmt.Println("change in selection should cause cpu boot")
			return true
		}

		if !this.NeedsClock() {
			bus.StopClock()
		}

		return false
	}

	if this.Memory.IntGetHelpInterrupt(this.MemIndex) {
		bus.StartDefault() // resume clock
		control.HelpPresent(this)
		this.Memory.IntSetHelpInterrupt(this.MemIndex, false)
	}

	this.ServiceBusProcessPending()

	this.PBPaste()

	if !this.Paused {
		return false
	}

	isC := bus.IsClock()
	if !isC {
		bus.StartDefault()
	}
	for this.Paused {

		time.Sleep(1 * time.Second)
		//		fmt.Println("Waiting for world")
	}
	if !isC {
		bus.StopClock()
	}

	return true
}

func (this *Interpreter) PBPaste() {
	gap := 1000 * time.Millisecond / time.Duration(settings.PasteCPS)

	if settings.PureBoot(this.MemIndex) && len(this.PasteBuffer.Runes) > 0 && this.GetMemory(0xc000)&128 == 0 && time.Since(this.LastPasteTime) > gap {
		ch := this.PasteBuffer.Runes[0]
		this.PasteBuffer.Runes = this.PasteBuffer.Runes[1:]
		this.LastPasteTime = time.Now()
		switch ch {
		case 10:
			ch = 13
		case 0x201c:
			ch = 34
		case 0x201d:
			ch = 34
		case 0x2018:
			ch = '\''
		case 0x2019:
			ch = '\''
		}
		if ch != 0 {
			this.Memory.KeyBufferAddNoRedirect(this.MemIndex, uint64(ch))
		}
	}
}

func (this *Interpreter) DoCatalog() {
	control.CatalogPresent(this)
}

func (this *Interpreter) CheckProfile(force bool) {
	//	pnum := int(this.Memory.ReadGlobal(this.Memory.MEMBASE(this.MemIndex) + memory.OCTALYZER_INTERPRETER_PROFILE))
	//	if pnum != this.lastProfile || force {
	//		specname := this.Memory.GetProfileName(pnum)
	//		memory.WarmStart = true
	//		this.LoadSpec(specname)
	//		memory.WarmStart = false
	//		this.lastProfile = pnum
	//	}
}

// SliceRecording
func (this *Interpreter) SliceRecording(fn string, newfn string, start int, end int) error {

	var e error
	this.p, e = NewPlayer(this, fn, []*bytes.Buffer(nil), false, 0)
	if e != nil {
		return e
	}

	e = this.p.Copy(start, end, newfn)
	return e
}

func (this *Interpreter) StartRecordingWithBlocks(blocks []*bytes.Buffer, debugMode bool) {
	n := 1
	basename := fmt.Sprintf("video%.3d.rec", n)
	for files.ExistsViaProvider("/local", basename) && n < 1000 {
		n++
		basename = fmt.Sprintf("video%.3d.rec", n)
	}
	p := "/local/" + basename
	files.MkdirViaProvider(p)
	for i := 0; i < len(blocks); i++ {
		fn := fmt.Sprintf("ablock%.5d.s", i)
		e := files.WriteBytesViaProvider(p, fn, blocks[i].Bytes())
		if e != nil {
			return
		}
	}
	this.r.usemem = false
	this.r.pathname = p
	this.r.memblocks = make([]*bytes.Buffer, 0)
	this.r.blockbuffer = bytes.NewBufferString("")
	this.r.blocknum = len(blocks)
	this.r.allCPUStates = debugMode
	this.r.Start()
}

func (this *Interpreter) WriteBlocks(blocks []*bytes.Buffer) {
	n := 1
	basename := fmt.Sprintf("video%.3d.rec", n)
	for files.ExistsViaProvider("/local", basename) && n < 1000 {
		n++
		basename = fmt.Sprintf("video%.3d.rec", n)
	}
	p := "/local/" + basename
	files.MkdirViaProvider(p)
	for i := 0; i < len(blocks); i++ {
		fn := fmt.Sprintf("block%.5d.s", i)
		e := files.WriteBytesViaProvider(p, fn, blocks[i].Bytes())
		if e != nil {
			return
		}
	}
	fmt.Printf("Created live -> disk recording %s\n", p)
}

func (this *Interpreter) RecordToggle(debugMode bool) {
	if this.r != nil {
		this.StopRecording()
		return
	}

	n := 1
	basename := fmt.Sprintf("video%.3d.rec", n)
	for files.ExistsViaProvider("/local/MyRecordings", basename) && n < 1000 {
		n++
		basename = fmt.Sprintf("video%.3d.rec", n)
	}

	this.StartRecording("/local/MyRecordings/"+basename, debugMode)
}

func (this *Interpreter) PlayBlocks(blocks []*bytes.Buffer, backwards bool, backJumpMS int) {

	this.GetProducer().PauseMicroControls()

	isC := bus.IsClock()
	if !isC {
		bus.StartDefault()
	}

	var e error
	this.p, e = NewPlayer(this, "", blocks, backwards, backJumpMS)
	if e != nil {
		this.PutStr(e.Error() + "\r\n")
		return
	}
	r := this.p.Playback()
	fmt.Println("END PLAYBACK")

	if !isC {
		bus.StopClock()
	}

	this.GetProducer().ResumeMicroControls()

	if r == interfaces.PEM_RESUME_CPU {
		acpu := apple2helpers.GetCPU(this)
		mr := acpu.ExecTillHalted()
		if mr != cpu.FE_HALTED && mr != cpu.FE_OK {
			m := apple2helpers.NewMonitor(this)
			m.Break()
		}
	} else if r == interfaces.PEM_RESET_SLOT {
		this.Memory.IntSetSlotRestart(this.MemIndex, true)
	}
}

func (this *Interpreter) PlayRecording(fn string) {
	var e error
	this.p, e = NewPlayer(this, fn, []*bytes.Buffer(nil), false, 0)
	if e != nil {
		this.PutStr(e.Error() + "\r\n")
		return
	}
	this.p.Playback()
	fmt.Println("END PLAYBACK")
}

func (this *Interpreter) PlayRecordingCustom(fn string, backwards bool) {
	var e error
	this.p, e = NewPlayer(this, fn, []*bytes.Buffer(nil), backwards, 0)
	if e != nil {
		this.PutStr(e.Error() + "\r\n")
		return
	}
	this.p.Playback()
	fmt.Println("END PLAYBACK")
}

func (this *Interpreter) AnalyzeRecording(fn string, analyzeMap map[string]interfaces.AnalyzerFunc) bool {
	var e error
	this.p, e = NewPlayer(this, fn, []*bytes.Buffer(nil), false, 0)
	if e != nil {
		this.PutStr(e.Error() + "\r\n")
		return false
	}
	r := this.p.Analyze(analyzeMap)
	fmt.Println("END ANALYSIS")
	return r
}

func (this *Interpreter) StartRecording(fn string, debugMode bool) {
	var e error
	this.r, e = NewRecorder(this, fn, nil, debugMode)
	if e != nil {
		this.PutStr(e.Error() + "\r\n")
		return
	}
	this.r.Start()
}

func (this *Interpreter) IncrementRecording(n int) {
	if this.IsRecordingDiscVideo() || this.IsRecordingVideo() {
		this.r.Increment(n)
	}
}

func (this *Interpreter) ResumeRecording(fn string, blocks []*bytes.Buffer, debugMode bool) {
	var e error
	this.r, e = NewRecorder(this, fn, blocks, debugMode)
	if e != nil {
		this.PutStr(e.Error() + "\r\n")
		return
	}
	this.r.Start()
}

func (this *Interpreter) GetLiveBlocks() []*bytes.Buffer {
	if this.r == nil {
		return nil
	}
	return this.r.memblocks
}

func (this *Interpreter) StopRecordingHard() {
	if this.r != nil {
		this.r.Stop()
		this.r = nil
	}
}

func (this *Interpreter) StopRecording() {
	if this.r != nil {
		this.r.Stop()
		this.r = nil
	}
}

func (this *Interpreter) GetDataNames() []string {
	out := make([]string, 0)
	for k, _ := range this.DataMap.Content {
		out = append(out, k)
	}
	return out
}

func (this *Interpreter) SetCommandState(cs *interfaces.CommandState) {

	if this.Children != nil {
		this.Children.SetCommandState(cs)
		return
	}

	this.CurrentCommandState = cs
}

func (this *Interpreter) GetCommandState() *interfaces.CommandState {

	if this.Children != nil {
		return this.Children.GetCommandState()
	}

	return this.CurrentCommandState
}

func (this *Interpreter) GetPublicDataNames() []string {
	out := make([]string, 0)
	for k, v := range this.DataMap.Content {
		if !v.Hidden {
			out = append(out, k)
		}
	}
	return out
}

func (this *Interpreter) GetCallReturnToken() *types.Token {
	return this.CallReturnToken
}

func (this *Interpreter) SetCallReturnToken(t *types.Token) {
	this.CallReturnToken = t
}

func (this *Interpreter) GetVM() types.VarManager {
	return this.Local
}

func (this *Interpreter) SetVM(vm types.VarManager) {
	this.Local = vm
}

func (this *Interpreter) GetDataKeys() []string {
	return this.DataMap.Keys()
}

func (this *Interpreter) BootCheck() {

	if settings.IsRemInt {
		return
	}

	txt := apple2helpers.TEXT(this)
	apple2helpers.TEXT40(this)
	apple2helpers.Clearscreen(this)

	bus.Sync()

	update.CheckFilename()
	update.CheckAndDownload(txt)
	control.CheckNewReleaseNotes(this, false)

}

func (this *Interpreter) PreOptimizer() {
	// Iterate the code and look for improvement overrides to make
	this.CodeOptimizations = *types.NewAlgorithm()

	if this.IsRunning() {

		for _, lno := range this.Code.GetSortedKeys() {
			ll, _ := this.Code.Get(lno)
			s := ll.String()
			this.Dialect.CheckOptimize(lno, s, this.CodeOptimizations)
		}

	} else if this.IsRunningDirect() {
		for _, lno := range this.DirectAlgorithm.GetSortedKeys() {
			ll, _ := this.DirectAlgorithm.Get(lno)
			s := ll.String()
			//fmt.Println("CheckOptimize:", s)
			this.Dialect.CheckOptimize(lno, s, this.CodeOptimizations)
		}
	}
}

func (this *Interpreter) GetFileRecord() filerecord.FileRecord {
	return this.MetaData
}

func (this *Interpreter) SetFileRecord(fr filerecord.FileRecord) {
	this.MetaData = fr
}

func (this *Interpreter) SetClientSync(b bool) {
	this.clientSync = b
}

//~ func (this *Interpreter) ConnectControl(host, port string, slotid int) error {

//~ return nil
//~ // connect to the control channel
//~ var e error
//~ this.RemIntControl = client.NewDuckTapeClient(host, port, s8webclient.CONN.Username+"-control", "tcp")
//~ e = this.RemIntControl.Connect()
//~ if e != nil {
//~ return e
//~ }

//~ this.ChatMessages = make([]ChatMessage, 0)
//~ this.ResultQueue = make(chan QueryResult)

//~ // Subscribe to chat channel
//~ //this.RemIntControl.SendMessage("SUB", []byte("gamechat"+utils.IntToStr(slotid)), false)
//~ //this.RemIntControl.SendMessage("SND", []byte("gamechat"+utils.IntToStr(slotid)), false)

//~ // Start processing loop
//~ go func() {

//~ for this.RemIntControl != nil && this.RemIntControl.Connected {

//~ select {
//~ case msg := <-this.RemIntControl.Incoming:
//~ //fmt.Printf("FROM REMINT: %s %s\n", msg.ID, string(msg.Payload))
//~ switch msg.ID {
//~ case "CHT":
//~ parts := strings.SplitN(string(msg.Payload), ":", 2)
//~ this.ChatMutex.Lock()
//~ this.ChatMessages = append(this.ChatMessages, ChatMessage{parts[0], parts[1]})
//~ this.ChatMutex.Unlock()
//~ case "QIE":
//~ this.ResultQueue <- QueryResult{Result: "", Err: errors.New(string(msg.Payload))}
//~ case "QIS":
//~ this.ResultQueue <- QueryResult{Result: string(msg.Payload), Err: nil}
//~ case "FXC":
//~ // receive camera dolally blarp
//~ for this.Memory.ReadGlobal(this.Memory.MEMBASE(this.GetMemIndex())+memory.OCTALYZER_CAMERA_GFX_BASE) != 0 {
//~ time.Sleep(1 * time.Millisecond)
//~ }

//~ for i, v := range msg.Payload {
//~ this.Memory.WriteGlobalSilent(this.Memory.MEMBASE(this.GetMemIndex())+memory.OCTALYZER_CAMERA_GFX_BASE+2+i, uint64(v))
//~ }
//~ this.Memory.WriteGlobalSilent(this.Memory.MEMBASE(this.GetMemIndex())+memory.OCTALYZER_CAMERA_GFX_BASE+1, uint64(len(msg.Payload)))
//~ this.Memory.WriteGlobalSilent(this.Memory.MEMBASE(this.GetMemIndex())+memory.OCTALYZER_CAMERA_GFX_INDEX, uint64(this.RemIntIndex))
//~ this.Memory.WriteGlobalSilent(this.Memory.MEMBASE(this.GetMemIndex())+memory.OCTALYZER_CAMERA_GFX_BASE, uint64(types.CC_JSON))
//~ }
//~ default:
//~ time.Sleep(5 * time.Millisecond)
//~ }

//~ }

//~ }()

//~ return nil
//~ }

func (this *Interpreter) GetChatMessages(maxcount int) ([]string, []string, error) {
	names := make([]string, 0)
	msgs := make([]string, 0)
	var err error

	this.ChatMutex.Lock()
	defer this.ChatMutex.Unlock()

	avail := len(this.ChatMessages)
	if avail > maxcount {
		avail = maxcount
	}

	for i := 0; i < avail; i++ {
		names = append(names, this.ChatMessages[i].Sender)
		msgs = append(msgs, this.ChatMessages[i].Message)
	}

	this.ChatMessages = this.ChatMessages[avail:len(this.ChatMessages)]

	return names, msgs, err
}

//~ func (this *Interpreter) SendChatMessage(message string) {
//~ this.SendRemIntMessage("CHT", []byte(s8webclient.CONN.Username+":"+message), true)
//~ }

//~ func (this *Interpreter) GetControlState(control string) (string, error) {

//~ this.SendRemIntMessage("QIM", []byte(control), true)

//~ c := time.NewTimer(10 * time.Second)

//~ var result string
//~ var err error

//~ select {
//~ case _ = <-c.C:
//~ err = errors.New("Timeout")
//~ case r := <-this.ResultQueue:
//~ result = r.Result
//~ err = r.Err
//~ }

//~ return result, err

//~ }

//~ func (this *Interpreter) SendRemIntMessage(id string, payload []byte, binary bool) {
//~ if this.RemIntControl == nil || !this.RemIntControl.Connected {
//~ return
//~ }

//~ this.RemIntControl.SendMessage(id, payload, binary)
//~ }

//~ func (this *Interpreter) GetRemIntControl() *client.DuckTapeClient {
//~ return this.RemIntControl
//~ }

//~ func (this *Interpreter) SetRemIntControl(d *client.DuckTapeClient) {
//~ this.RemIntControl = d
//~ }

func (this *Interpreter) GetRemIntIndex() int {
	return this.RemIntIndex
}

func (this *Interpreter) SetRemIntIndex(i int) {
	this.RemIntIndex = i
}

func (this *Interpreter) IsRemote() bool {
	return (this.Remote != nil && this.Remote.Connected)
}

var instance *Interpreter

func (this *Interpreter) SetPasteBuffer(r runestring.RuneString) {
	if this.Children != nil {
		this.Children.SetPasteBuffer(r)
		return
	}
	this.PasteBuffer = r
}

func (this *Interpreter) SetVidMode(mode int) {
	if this.IsRecordingVideo() {
		this.r.SetVidMode(mode)
	}
}

func (this *Interpreter) SetMemMode(mode int) {
	if this.IsRecordingVideo() {
		this.r.SetMemMode(mode)
	}
}

func (this *Interpreter) GetSpec() string {
	return this.SpecFile
}

func (this *Interpreter) SetSpec(s string) {
	this.SpecFile = s
}

//~ func HandleMemoryEvent(index int, addr int, value uint64) {

//~ //if addr > memory.OCTALYZER_INTERPRETER_MAX && index == instance.MemIndex {
//~ ////fmt.Printf("Would send -- %d, %d\n", addr, value)

//~ if index != instance.MemIndex {
//~ return
//~ }

//~ if index != 0 {
//~ addr = ConvertAddress("local", addr, 0-index)
//~ }

//~ // don't mirror keys
//~ if addr >= memory.OCTALYZER_KEY_BUFFER_BASE && addr < memory.OCTALYZER_KEY_BUFFER_BASE+memory.OCTALYZER_KEY_BUFFER_SIZE {
//~ return
//~ }

//~ if addr >= memory.OCTALYZER_PADDLE_BASE && addr < memory.OCTALYZER_PADDLE_BASE+memory.OCTALYZER_PADDLE_SIZE {
//~ return
//~ }

//~ b := make([]byte, 0)
//~ b = append(b, byte(index))
//~ b = append(b, byte((addr/65536)%256))
//~ b = append(b, byte((addr/256)%256))
//~ b = append(b, byte(addr%256))
//~ // data
//~ b = append(b, byte((value>>24)&255))
//~ b = append(b, byte((value>>16)&255))
//~ b = append(b, byte((value>>8)&255))
//~ b = append(b, byte((value>>0)&255))

//~ instance.Remote.SendMessage("CBY", b, true)

//~ //	log.Printf("--> Instruct remint (%d) to put value %d into address %d\n", instance.MemIndex, value, addr)

//~ //}
//~ }

//~ func (this *Interpreter) ProcessRemote() {

//~ var p0, p1, p2, p3 uint64
//~ var pb0, pb1, pb2, pb3 uint64

//~ for this.Remote != nil && this.Remote.Connected {

//~ // handle any inputs and mirror them to the remote
//~ var ie InputEvent = InputEvent{Kind: ET_NONE}

//~ switch {
//~ case this.Memory.IntGetPaddleButton(this.MemIndex, 0) != pb0:
//~ pb0 = this.Memory.IntGetPaddleButton(this.MemIndex, 0)
//~ pressed := (pb0 != 0)
//~ if pressed {
//~ ie = InputEvent{
//~ Kind:  ET_PADDLE_BUTTON_DOWN,
//~ ID:    0,
//~ Value: int(pb0),
//~ }
//~ } else {
//~ ie = InputEvent{
//~ Kind:  ET_PADDLE_BUTTON_UP,
//~ ID:    0,
//~ Value: int(pb0),
//~ }
//~ }
//~ //ie = this.InputMapper.FilterEvent(ie)
//~ case this.Memory.IntGetPaddleButton(this.MemIndex, 1) != pb1:
//~ pb1 = this.Memory.IntGetPaddleButton(this.MemIndex, 1)
//~ pressed := (pb1 != 0)
//~ if pressed {
//~ ie = InputEvent{
//~ Kind:  ET_PADDLE_BUTTON_DOWN,
//~ ID:    1,
//~ Value: int(pb1),
//~ }
//~ } else {
//~ ie = InputEvent{
//~ Kind:  ET_PADDLE_BUTTON_UP,
//~ ID:    1,
//~ Value: int(pb1),
//~ }
//~ }
//~ //ie = this.InputMapper.FilterEvent(ie)
//~ case this.Memory.IntGetPaddleButton(this.MemIndex, 2) != pb2:
//~ pb2 = this.Memory.IntGetPaddleButton(this.MemIndex, 2)
//~ pressed := (pb2 != 0)
//~ if pressed {
//~ ie = InputEvent{
//~ Kind:  ET_PADDLE_BUTTON_DOWN,
//~ ID:    2,
//~ Value: int(pb2),
//~ }
//~ } else {
//~ ie = InputEvent{
//~ Kind:  ET_PADDLE_BUTTON_UP,
//~ ID:    2,
//~ Value: int(pb2),
//~ }
//~ }
//~ //ie = this.InputMapper.FilterEvent(ie)
//~ case this.Memory.IntGetPaddleButton(this.MemIndex, 3) != pb3:
//~ pb3 = this.Memory.IntGetPaddleButton(this.MemIndex, 3)
//~ pressed := (pb3 != 0)
//~ if pressed {
//~ ie = InputEvent{
//~ Kind:  ET_PADDLE_BUTTON_DOWN,
//~ ID:    3,
//~ Value: int(pb3),
//~ }
//~ } else {
//~ ie = InputEvent{
//~ Kind:  ET_PADDLE_BUTTON_UP,
//~ ID:    3,
//~ Value: int(pb3),
//~ }
//~ }
//~ //ie = this.InputMapper.FilterEvent(ie)
//~ case this.Memory.IntGetPaddleValue(this.MemIndex, 0) != p0:
//~ p0 = this.Memory.IntGetPaddleValue(this.MemIndex, 0)
//~ ie = InputEvent{
//~ Kind:  ET_PADDLE_VALUE_CHANGE,
//~ ID:    0,
//~ Value: int(p0),
//~ }
//~ //ie = this.InputMapper.FilterEvent(ie)
//~ case this.Memory.IntGetPaddleValue(this.MemIndex, 1) != p1:
//~ p1 = this.Memory.IntGetPaddleValue(this.MemIndex, 1)
//~ ie = InputEvent{
//~ Kind:  ET_PADDLE_VALUE_CHANGE,
//~ ID:    1,
//~ Value: int(p1),
//~ }
//~ //ie = this.InputMapper.FilterEvent(ie)
//~ case this.Memory.IntGetPaddleValue(this.MemIndex, 2) != p2:
//~ p2 = this.Memory.IntGetPaddleValue(this.MemIndex, 2)
//~ ie = InputEvent{
//~ Kind:  ET_PADDLE_VALUE_CHANGE,
//~ ID:    2,
//~ Value: int(p2),
//~ }
//~ //ie = this.InputMapper.FilterEvent(ie)
//~ case this.Memory.IntGetPaddleValue(this.MemIndex, 3) != p3:
//~ p3 = this.Memory.IntGetPaddleValue(this.MemIndex, 3)
//~ ie = InputEvent{
//~ Kind:  ET_PADDLE_VALUE_CHANGE,
//~ ID:    3,
//~ Value: int(p3),
//~ }
//~ //ie = this.InputMapper.FilterEvent(ie)
//~ case this.Memory.KeyBufferSize(this.MemIndex) > 0:
//~ {
//~ value := this.Memory.KeyBufferGetLatest(this.MemIndex)
//~ ie = InputEvent{
//~ Kind:  ET_KEYPRESS,
//~ ID:    0,
//~ Value: int(value),
//~ }
//~ //				//fmt.Println(ie)
//~ //ie = this.InputMapper.FilterEvent(ie)
//~ //				//fmt.Println("After filter", ie)
//~ }
//~ default:
//~ time.Sleep(5 * time.Millisecond)
//~ }

//~ if ie.Kind != ET_NONE {

//~ // Send the event
//~ data, _ := ie.MarshalBinary()
//~ if ie.Kind == ET_KEYPRESS {
//~ this.Remote.SendMessage("IEV", data, true)
//~ }

//~ }
//~ }

//~ }

func ConvertAddress(d string, addr int, slotdiff int) int {

	//slotdiff := this.MemIndex - this.RemoteIndex
	//	oaddr := addr

	if slotdiff == 0 {
		//		//fmt.Println("Zero slot diff")
		return addr
	}

	// Interpreter slot

	amod := addr % memory.OCTALYZER_INTERPRETER_SIZE

	if amod >= memory.OCTALYZER_KEY_BUFFER_BASE && amod < memory.OCTALYZER_KEY_BUFFER_BASE+memory.OCTALYZER_KEY_BUFFER_SIZE {
		return 0
	}

	if amod >= memory.OCTALYZER_PADDLE_BASE && amod < memory.OCTALYZER_PADDLE_BASE+memory.OCTALYZER_PADDLE_SIZE {
		return 0
	}

	return addr + memory.OCTALYZER_INTERPRETER_SIZE*slotdiff

}

func (this *Interpreter) TransferRemoteControl(target string) int {

	if this.Remote == nil {
		return 0
	}

	this.Remote.SendMessage(fastserv.FS_REQUEST_TRANSFER_OWNERSHIP, []byte(target))

	return 0

}

func (this *Interpreter) SetControlProfile(target string, profile string) int {
	if this.Remote == nil {
		return 0
	}

	this.Remote.SendMessage(fastserv.FS_ALLOCATE_CONTROL, []byte(target+":"+profile))

	return 0
}

func (this *Interpreter) SendRemoteCommand(command string) int {
	if this.Remote == nil {
		return 0
	}

	ncmd := command + "\r"

	this.Remote.SendMessage(fastserv.FS_REMOTE_EXEC, []byte(ncmd))

	return 0
}

func (this *Interpreter) SendRemoteText(command string) int {
	if this.Remote == nil {
		return 0
	}

	this.Remote.SendMessage(fastserv.FS_REMOTE_PARSE, []byte(command))

	return 0
}

func (this *Interpreter) OnRemoteConnect() {
	this.PutStr("OK.\r\n")
	this.Remote.SendMessage(fastserv.FS_MEMSYNC_REQUEST, []byte{48 + byte(this.RemoteIndex)})
	instance = this
	this.Memory.SetMemorySync(this.MemIndex, this.RemoteIndex, this.Remote)
}

func (this *Interpreter) ConnectRemote(ip, port string, slotid int) {

	if this.Children != nil {
		this.Children.ConnectRemote(ip, port, slotid)
		return
	}

	if this.Remote != nil {
		this.EndRemote()
	}

	//	fmt.Println("CONNECT TO REMOTE")

	this.Remote = client.NewFSClient(ip, port, s8webclient.CONN.Username, "tcp")
	this.Remote.OnConnect = this.OnRemoteConnect // <-- this will always get a nice fresh memsync

	this.RemoteIndex = slotid

	for this.Remote.State != client.CS_CONNECTED && this.Remote.State != client.CS_DISCONNECTED {
		this.Remote.Do()
	}

	var start time.Time = time.Now()
	var byteCount int
	var packetCount int

	this.NeedRemoteKill = false

	//~ this.Remote.SendMessage("SUB", []byte("memupd"), false)
	//~ this.Remote.SendMessage("SND", []byte("cliupd"), false)

	this.PutStr("Connected.\r\n")

	this.Memory.IntSetActiveState(this.MemIndex, 1)
	this.Memory.IntSetLayerState(this.MemIndex, 1)

	//go this.ProcessRemote()
	var blocks [256]int

	this.State = types.REMOTE

	go func() {

		//xx := this.RemoteIndex

		for this.Remote != nil && this.Remote.Do() {

			if this.NeedRemoteKill {
				this.Remote.Close()
				for this.Remote.State != client.CS_DISCONNECTED {
					this.Remote.Do()
				}
				//				fmt.Println("KILLED REMOTE")
				this.NeedRemoteKill = false
				this.Remote = nil
				this.State = types.STOPPED
				return
			}

			if this.Remote.Connected {

				select {
				case msg := <-this.Remote.Incoming:
					packetCount++
					byteCount += len(msg)
					if settings.TRACENET && packetCount%100 == 0 {
						secs := time.Since(start) / time.Second
						if secs > 0 {
							fmt.Printf("** Average throughout = %d Bps, average packet size = %d bytes\n", byteCount/int(secs), byteCount/packetCount)
						}
						if packetCount%10000 == 0 {
							for h := 0; h < 16; h++ {
								fmt.Printf("H %.2x: ", h)
								for l := 0; l < 16; l++ {
									z := 16*h + l
									fmt.Printf("%.6x ", blocks[z])
								}
								fmt.Println()
							}
						}
					}
					switch fastserv.FSPayloadType(msg[0]) {
					case fastserv.FS_RESTALGIA_COMMAND:

						data := string(msg[1:])
						this.PassRestBufferNB(data)

					case fastserv.FS_CLIENTAUDIO:

						//fmt.Println("remote audio received")

						data, err := mempak.UnpackSliceUints(msg[1:])
						if err == nil {
							rate := int(data[0])
							indata := data[1:]
							this.Memory.DirectSendAudioPacked(this.MemIndex, 0, indata, rate)
						}

					case fastserv.FS_MEMSYNC_RESPONSE:
						//						fmt.Println("Got ram sync")
						cdata := msg[1:]
						data := utils.UnGZIPBytes(cdata)
						f := freeze.NewEmptyState(this)
						f.LoadFromBytes(data)
						f.Apply(this)

					case fastserv.FS_BULKMEM:
						fmt.Printf("Got BMU with payload size %d bytes\n", len(msg[1:]))
						data := msg[1:]
						//count := int(data[0])<<16 | int(data[1])<<8 | int(data[2])
						idx := 3

						//fmt.Printf("-> %d memory changes\n", count)

						for idx < len(data) {

							//fmt.Printf("-> idx = %d\n", idx)

							// if data[idx] == 0x18 then we have a special case
							if data[idx] == 0x18 {
								ll := int(data[idx+1])<<16 | int(data[idx+2])<<8 | int(data[idx+3])
								size := 4 + ll
								end := idx + size

								//fmt.Printf("-> Block transfer at %d is %d bytes, msg size = %d, end = %d, len = %d\n", idx+3, ll, size, end, len(data))

								chunk := data[idx+4 : end]
								addr, values, e := mempak.DecodeBlock(chunk)
								if e != nil {
									panic(e)
								}

								this.Memory.BlockWrite(this.MemIndex, this.Memory.MEMBASE(this.MemIndex)+addr, values)

								// done
								idx += size
								continue
							}

							end := idx + 8
							if end >= len(data) {
								end = len(data)
							}
							chunk := data[idx:end]
							//addr := int(chunk[0])<<16 | int(chunk[1])<<8 | int(chunk[2])
							//value := uint64(chunk[3])<<24 | uint64(chunk[4])<<16 | uint64(chunk[5])<<8 | uint64(chunk[6])

							_, addr, value, read, size, e := mempak.Decode(chunk)
							if e != nil {
								break
							}

							a := addr % memory.OCTALYZER_INTERPRETER_SIZE

							if a < this.LoRemote {
								this.LoRemote = a
							}

							if a > this.HiRemote {
								this.HiRemote = a
							}

							if a < 65536 {
								blocks[a/256] = blocks[a/256] + 1
							}

							// if a >= 0x3e000 && a < 0x3f000 {
							// 	fmt.RPrintf("moni update received %d -> %d\n", value, a)
							// }

							////fmt.Printf("Remote - hi sent = %d, lo sent = %d\n", this.HiRemote, this.LoRemote)

							if !read {

								fmt.Printf("Write to address %d\n", a)

								if a >= memory.MICROM8_VOICE_PORT_BASE && a <= memory.MICROM8_VOICE_PORT_BASE+memory.MICROM8_VOICE_PORT_SIZE*memory.MICROM8_VOICE_COUNT {

									offs := a - memory.MICROM8_VOICE_PORT_BASE
									voice := offs / 2
									isOpCode := offs%2 == 0

									if isOpCode {
										v := this.Memory.ReadGlobal(this.MemIndex, this.Memory.MEMBASE(this.MemIndex)+a+1)
										//rlog.Printf("rest opcode: v=%d, opcode=%d, value=%d", voice, value, v)
										this.Memory.RestalgiaOpCode(
											this.MemIndex,
											voice,
											int(value),
											v,
										)
									} else {
										this.Memory.WriteGlobalSilent(
											this.MemIndex,
											this.Memory.MEMBASE(this.MemIndex)+a,
											value)
									}

								} else if addr%memory.OCTALYZER_INTERPRETER_SIZE == memory.OCTALYZER_SPEAKER_PLAYSTATE {

									this.Memory.WriteInterpreterMemory(
										this.MemIndex,
										addr%memory.OCTALYZER_INTERPRETER_SIZE,
										value)

								} else {

									if (a < memory.OCTALYZER_MAPPED_CAM_BASE || a > memory.OCTALYZER_MAPPED_CAM_CONTROL) && a != memory.OCTALYZER_VIDEO_TINT {

										if a >= 8192 && a < 24576 {
											this.Memory.WriteInterpreterMemory(
												this.MemIndex,
												addr%memory.OCTALYZER_INTERPRETER_SIZE,
												value)
										} else {

											this.Memory.WriteInterpreterMemorySilent(
												this.MemIndex,
												addr%memory.OCTALYZER_INTERPRETER_SIZE,
												value)
										}

									}

								}

								if addr%memory.OCTALYZER_INTERPRETER_SIZE == memory.OCTALYZER_LAYERSTATE && value != 0 {
									//									fmt.Println("Waiting for the client to observe a layerchange")

									for this.Memory.ReadGlobal(this.MemIndex, this.Memory.MEMBASE(this.MemIndex)+memory.OCTALYZER_LAYERSTATE) != 0 {
										time.Sleep(1 * time.Millisecond)
									}

									//									fmt.Println("OK...")
								}

							} else {

								this.Memory.ReadInterpreterMemory(
									this.MemIndex,
									addr%memory.OCTALYZER_INTERPRETER_SIZE,
								)

							}

							//if addr % memory.OCTALYZER_INTERPRETER_SIZE == memory.OCTALYZER_SPEAKER_FREQ {
							//	//fmt.Println("SPEAKER tripped")
							//}

							idx += size

						}
					}
				default:
					//bus.Sync()
					time.Sleep(1 * time.Millisecond)
				}

			} else {
				time.Sleep(10 * time.Millisecond)
			}

		}

		//		fmt.Println("ENDED REMOTE")
		this.NeedRemoteKill = false

	}()

}

func (this *Interpreter) SetSuppressFormat(v bool) {
	this.SuppressFormat = v
}

func (this *Interpreter) IsSuppressFormat() bool {
	return this.SuppressFormat
}

func (this *Interpreter) SetFeedBuffer(s string) {
	this.FeedBuffer = s
}

func (this *Interpreter) GetSpeed() int {
	return apple2helpers.GetSPEED(this)
}

func (this *Interpreter) SetSpeed(s int) {
	apple2helpers.SetSPEED(this, s)
}

func (this *Interpreter) GetLastChar() rune {
	return this.LastChar
}

func (this *Interpreter) SetLastChar(s rune) {
	this.LastChar = s
}

func (this *Interpreter) GetCharacterCapture() string {
	return this.CharacterCapture
}

func (this *Interpreter) SetCharacterCapture(s string) {
	this.CharacterCapture = s
}

func (this *Interpreter) GetOutChannel() string {
	return this.OutChannel
}

func (this *Interpreter) SetOutChannel(s string) {
	this.OutChannel = s
}

func (this *Interpreter) GetInChannel() string {
	return this.InChannel
}

func (this *Interpreter) SetInChannel(s string) {
	this.InChannel = s
}

func (this *Interpreter) GetDosBuffer() string {
	return this.DosBuffer
}

func (this *Interpreter) SetDosBuffer(s string) {
	this.DosBuffer = s
}

func (this *Interpreter) SetDosCommand(v bool) {
	this.DosCommand = v
}

func (this *Interpreter) IsDosCommand() bool {
	return this.DosCommand
}

func (this *Interpreter) SetNextByteColor(v bool) {
	this.NextBytecolor = v
}

func (this *Interpreter) IsNextByteColor() bool {
	return this.NextBytecolor
}

func (this *Interpreter) IsBreakable() bool {
	return this.Breakable
}

func (this *Interpreter) SetBreakable(v bool) {
	this.Breakable = v
}

func (this *Interpreter) ShouldSaveAndRestoreText() bool {

	if this.Parent != nil {
		return this.Parent.ShouldSaveAndRestoreText()
	}

	return this.SaveRestoreText
}

func (this *Interpreter) SetSaveAndRestoreText(v bool) {

	if this.Parent != nil {
		this.Parent.SetSaveAndRestoreText(v)
		return
	}

	this.SaveRestoreText = v
}

func (this *Interpreter) SetDisplayPage(s string) {
	this.DisplayPage = s
}

func (this *Interpreter) SetCurrentPage(s string) {
	this.CurrentPage = s
	//	// set memory based off page
	//	if s == "HGR1" || s == "DHR1" {
	//		this.SetMemory(230, 32)
	//	} else if s == "HGR2" || s == "DHR2" {
	//		this.SetMemory(230, 64)
	//	}
}

func (this *Interpreter) GetDisplayPage() string {
	return this.DisplayPage
}

func (this *Interpreter) GetCurrentPage() string {
	return this.CurrentPage
}

func (this *Interpreter) SetPrompt(s string) {
	this.Prompt = s
	//this.SetMemory(51, uint64(s[0]))
}

func (this *Interpreter) GetPrompt() string {
	//this.Prompt = string(rune(this.GetMemory(51)))
	return this.Prompt
}

func (this *Interpreter) SetTabWidth(s int) {
	this.TabWidth = s
}

func (this *Interpreter) SetMemIndex(i int) {
	this.MemIndex = i
}

func (this *Interpreter) GetMemIndex() int {
	return this.MemIndex
}

func (this *Interpreter) Log(component string, message string) {

	if this.IsSilent() {
		return
	}

	msSince := (time.Now().UnixNano() - this.RunStart) / 1000000

	if this.IsDebug() {
		debug.Log(msSince, int64(this.PC.Line), int64(this.PC.Statement), component, message)
	}

}

func (this *Interpreter) IsSilent() bool {
	return false //this.Silent
}

func (this *Interpreter) SetSilent(v bool) {
	this.Silent = v
}

func (this *Interpreter) IsDebug() bool {
	return this.DebugMe
}

func (this *Interpreter) SetDebug(v bool) {
	this.DebugMe = v
}

func (this *Interpreter) IsIgnoreSpecial() bool {
	return this.IgnoreSpecial
}

func (this *Interpreter) SetIgnoreSpecial(v bool) {
	this.IgnoreSpecial = v
}

func (this *Interpreter) SetNeedsPrompt(v bool) {
	this.NeedsPrompt = v
}

func (this *Interpreter) IndexOfLoop(forvarname string) int {
	match := -1
	i := 0
	for _, ls := range this.LoopStack {
		if ls.VarName == forvarname {
			match = i
		}
		i++
	}
	return match
}

func (this *Interpreter) Wait(ms int64) {

	//this.WaitUntil = time.Now().UnixNano() + ms
	if this.Children != nil {
		this.Children.Wait(ms)
		return
	}

	//	fmt.Printf("Wait %d ns\n", ms)

	nwu := time.Now().Add(time.Duration(ms) * time.Nanosecond)

	if nwu.After(this.WaitUntil) {
		this.WaitUntil = nwu
	}

}

func (this *Interpreter) WaitFrom(now time.Time, ms int64) {

	//this.WaitUntil = time.Now().UnixNano() + ms
	if this.Children != nil {
		this.Children.Wait(ms)
		return
	}

	this.WaitUntil = now.Add(time.Duration(ms) * time.Nanosecond)

}

func (this *Interpreter) WaitDeduct(ms int64) {

	//this.WaitUntil = time.Now().UnixNano() + ms
	if this.Children != nil {
		this.Children.WaitDeduct(ms)
		return
	}

	this.WaitUntil = this.WaitUntil.Add(-time.Duration(ms) * time.Nanosecond)

}

func (this *Interpreter) WaitAdd(d time.Duration) {

	//this.WaitUntil = time.Now().UnixNano() + ms
	if this.Children != nil {
		this.Children.WaitAdd(d)
		return
	}

	this.WaitUntil = this.WaitUntil.Add(d)

}

func (this *Interpreter) GetVar(n string) *types.Variable {

	/* vars */
	var result *types.Variable
	var i int

	if this.OuterVars {
		result = this.GetVarLower(n)
		return result
	}

	result = nil
	if this.Local.ContainsKey(strings.ToLower(n)) {
		result = this.Local.Get(strings.ToLower(n))
		return result
	}

	if this.IsolateVars {
		return result
	}

	for i = this.Stack.Size() - 1; i >= 0; i-- {
		if this.Stack.Get(i).Locals.ContainsKey(strings.ToLower(n)) {
			result = this.Stack.Get(i).Locals.Get(strings.ToLower(n))
			return result
		}
	}

	/* enforce non void return */
	return result

}

func (this *Interpreter) PurgeOwnedVariables() {

	if this.Local != nil {
		this.Local.Clear()
	} else {
		//	fmt.Println("POV: before:", this.GetMemory(108)*256+this.GetMemory(107))
		this.Dialect.InitVarmap(this, nil)
		//		fmt.Println("POV: after:", this.GetMemory(108)*256+this.GetMemory(107))
	}
	//this.Local.MaxLength = this.Dialect.GetMaxVariableLength()
	this.FirstString = ""

}

func (this *Interpreter) GetData(n string) *types.Token {

	/* vars */
	var result *types.Token
	var i int

	if n == ":tlevels" {
		this.ShowDef(n)
	}

	result = nil
	if this.DataMap.ContainsKey(strings.ToLower(n)) {
		//log.Printf("Found var %s at current scope", n)
		result = this.DataMap.Get(strings.ToLower(n))
		return result
	}

	for i = this.Stack.Size() - 1; i >= 0; i-- {
		if this.Stack.Get(i).DataMap.ContainsKey(strings.ToLower(n)) {
			//log.Printf("Found var %s at level %d", n, i)
			result = this.Stack.Get(i).DataMap.Get(strings.ToLower(n))
			return result
		}
	}

	/* enforce non void return */
	return result

}

func (this *Interpreter) Bootstrap(n string, silent bool) error {
	preserve := false

	if n == "classic" {
		if this.Dialect.GetTitle() == "INTEGER+" {
			n = "int"
			preserve = true
		}
		if this.Dialect.GetTitle() == "Applesoft+" {
			n = "fp"
			preserve = true
		}
	} else if n == "plus" {
		if this.Dialect.GetTitle() == "INTEGER" {
			n = "int +"
			preserve = true
		}
		if this.Dialect.GetTitle() == "Applesoft" {
			n = "fp +"
			preserve = true
		}
	}

	if n == "logo" {
		this.Dialect = logo.NewDialectLogo()
	} else if n == "int" {
		this.Dialect = appleinteger.NewDialectAppleInteger()
	} else if n == "fp" {
		this.Dialect = applesoft.NewDialectApplesoft()
	} else if n == "shell" {
		this.Dialect = shell.NewDialectShell()
	} else {
		return exception.NewESyntaxError("INVALID DIALECT [" + n + "]")
	}

	//if !silent {
	//	this.Dialect.InitVDU(this, false)
	//}

	this.SetDialect(this.Dialect, preserve, silent)
	this.SaveCPOS()

	if !silent {
		this.NeedsPrompt = true
	}

	return nil
}

func (this *Interpreter) ExistsVarLower(n string) bool {

	/* vars */
	var result bool
	var i int

	result = false

	for i = this.Stack.Size() - 1; i >= 0; i-- {
		if this.Stack.Get(i).Locals.ContainsKey(strings.ToLower(n)) {
			result = true
			return result
		}
	}

	if this.Local.ContainsKey(strings.ToLower(n)) {
		result = true
	}

	/* enforce non void return */
	return result

}

func (this *Interpreter) Call(target types.CodeRef, c *types.Algorithm, s types.EntityState, iso bool, prefix string, stackstate types.TokenList, dia interfaces.Dialecter) {

	/* vars */
	var cr types.CodeRef
	var ostate *interfaces.StackEntry
	var keepregisters bool

	//System.Err.Println("CALL");
	this.StackDump()

	keepregisters = iso

	/* save current next location to stack */
	if this.State == types.RUNNING {
		cr = this.GetNextStatement(this.PC)
		ostate = interfaces.NewStackEntry()
		ostate.PC = &cr
		ostate.State = this.State
		ostate.Code = this.Code
		ostate.Locals = this.Local
		ostate.IsolateVars = this.IsolateVars
		ostate.VarPrefix = this.VarPrefix
		ostate.TokenStack = &this.TokenStack
		ostate.CurrentDialect = this.Dialect
		ostate.Registers = this.Registers
		ostate.DataMap = this.DataMap
		ostate.CreatedTokens = this.CreatedTokens
		ostate.LoopStep = float32(this.LoopStep)
		ostate.LoopVariable = this.LoopVariable
		ostate.LoopStackSize = len(this.LoopStack)
		ostate.CurrentSub = this.CurrentSubroutine
		this.Stack.Push(ostate)

		this.PC = target
		this.PC.SubIndex = -66
		this.Code = c
		this.State = s
		//		this.Local = types.NewVarMap(dia.GetMaxVariableLength(), this)
		this.IsolateVars = false
		this.VarPrefix = prefix
		this.TokenStack = stackstate
		if !keepregisters {
			this.Registers = *types.NewBRegisters(ostate.Registers)
		}
		this.DataMap = types.NewTokenMap()
		//		for k, v := range ostate.DataMap.Content {
		//			this.SetData(k, *v)
		//		}
		this.CreatedTokens = types.NewTokenList()
		this.Dialect = dia
		//System.Out.Println("Dialect == ",dia.Title;
		if this.LoopVariable == "" {
			this.LoopVariable = ""
		}
		this.LoopStep = 0
	} else if this.State == types.DIRECTRUNNING {
		cr = this.GetNextStatement(this.LPC)
		ostate = interfaces.NewStackEntry()
		ostate.PC = &cr
		ostate.State = this.State
		ostate.Code = this.DirectAlgorithm
		ostate.Locals = this.Local
		ostate.IsolateVars = this.IsolateVars
		ostate.VarPrefix = this.VarPrefix
		ostate.TokenStack = &this.TokenStack
		ostate.CurrentDialect = this.Dialect
		ostate.Registers = this.Registers
		ostate.DataMap = this.DataMap
		ostate.CreatedTokens = this.CreatedTokens
		ostate.LoopStep = float32(this.LoopStep)
		ostate.LoopVariable = this.LoopVariable
		ostate.LoopStackSize = len(this.LoopStack)
		ostate.CurrentSub = this.CurrentSubroutine
		this.Stack.Push(ostate)

		this.LPC = target
		this.LPC.SubIndex = -66
		this.DirectAlgorithm = c
		this.State = s
		//this.Local = types.NewVarMap(dia.GetMaxVariableLength(), this)
		this.IsolateVars = false
		this.VarPrefix = prefix
		this.TokenStack = stackstate
		if !keepregisters {
			this.Registers = *types.NewBRegisters(ostate.Registers)
		}
		this.DataMap = types.NewTokenMap()
		//		for k, v := range ostate.DataMap.Content {
		//			//this.DataMap.Content[k] = &types.Token{Content: v.Content, List: v.List, Type: v.Type}
		//			this.SetData(k, *v)
		//		}
		this.CreatedTokens = types.NewTokenList()
		this.Dialect = dia
		//System.Out.Println("Dialect == ",dia.Title;
		if this.LoopVariable == "" {
			this.LoopVariable = ""
		}
		this.LoopStep = 0
	}

}

func (this *Interpreter) CallTrigger(target types.CodeRef, c *types.Algorithm, s types.EntityState, iso bool, prefix string, stackstate types.TokenList, dia interfaces.Dialecter) {

	/* vars */
	var cr types.CodeRef
	var ostate *interfaces.StackEntry
	var keepregisters bool

	//System.Err.Println("CALL");
	this.StackDump()

	keepregisters = iso

	/* save current next location to stack */
	if this.State == types.RUNNING {
		cr = this.PC
		ostate = interfaces.NewStackEntry()
		ostate.PC = &cr
		ostate.State = this.State
		ostate.Code = this.Code
		ostate.Locals = this.Local
		ostate.IsolateVars = this.IsolateVars
		ostate.VarPrefix = this.VarPrefix
		ostate.TokenStack = &this.TokenStack
		ostate.CurrentDialect = this.Dialect
		ostate.Registers = this.Registers
		ostate.DataMap = this.DataMap
		ostate.CreatedTokens = this.CreatedTokens
		ostate.LoopStep = float32(this.LoopStep)
		ostate.LoopVariable = this.LoopVariable
		ostate.LoopStackSize = len(this.LoopStack)
		ostate.CurrentSub = this.CurrentSubroutine
		this.Stack.Push(ostate)

		this.PC = target
		this.PC.SubIndex = -66
		this.Code = c
		this.State = s
		//		this.Local = types.NewVarMap(dia.GetMaxVariableLength(), this)
		this.IsolateVars = false
		this.VarPrefix = prefix
		this.TokenStack = stackstate
		if !keepregisters {
			this.Registers = *types.NewBRegisters(ostate.Registers)
		}
		this.DataMap = types.NewTokenMap()
		//		for k, v := range ostate.DataMap.Content {
		//			this.SetData(k, *v)
		//		}
		this.CreatedTokens = types.NewTokenList()
		this.Dialect = dia
		//System.Out.Println("Dialect == ",dia.Title;
		if this.LoopVariable == "" {
			this.LoopVariable = ""
		}
		this.LoopStep = 0
	} else if this.State == types.DIRECTRUNNING {
		cr = this.LPC
		ostate = interfaces.NewStackEntry()
		ostate.PC = &cr
		ostate.State = this.State
		ostate.Code = this.DirectAlgorithm
		ostate.Locals = this.Local
		ostate.IsolateVars = this.IsolateVars
		ostate.VarPrefix = this.VarPrefix
		ostate.TokenStack = &this.TokenStack
		ostate.CurrentDialect = this.Dialect
		ostate.Registers = this.Registers
		ostate.DataMap = this.DataMap
		ostate.CreatedTokens = this.CreatedTokens
		ostate.LoopStep = float32(this.LoopStep)
		ostate.LoopVariable = this.LoopVariable
		ostate.LoopStackSize = len(this.LoopStack)
		ostate.CurrentSub = this.CurrentSubroutine
		this.Stack.Push(ostate)

		this.LPC = target
		this.LPC.SubIndex = -66
		this.DirectAlgorithm = c
		this.State = s
		//this.Local = types.NewVarMap(dia.GetMaxVariableLength(), this)
		this.IsolateVars = false
		this.VarPrefix = prefix
		this.TokenStack = stackstate
		if !keepregisters {
			this.Registers = *types.NewBRegisters(ostate.Registers)
		}
		this.DataMap = types.NewTokenMap()
		//		for k, v := range ostate.DataMap.Content {
		//			//this.DataMap.Content[k] = &types.Token{Content: v.Content, List: v.List, Type: v.Type}
		//			this.SetData(k, *v)
		//		}
		this.CreatedTokens = types.NewTokenList()
		this.Dialect = dia
		//System.Out.Println("Dialect == ",dia.Title;
		if this.LoopVariable == "" {
			this.LoopVariable = ""
		}
		this.LoopStep = 0
	}

}

func (this *Interpreter) RunStatementFunction() {
}

func (this *Interpreter) Clear() {

	this.State = types.EMPTY
	this.Code = types.NewAlgorithm()
	this.CodeOptimizations = *types.NewAlgorithm()
	syncmanager.Sync.SetSyncKey(this.Code.Checksum())

}

func (this *Interpreter) CallUserFunction(funcname string, inputs types.TokenList) error {

	if this.State == types.FUNCTIONRUNNING {
		return exception.NewESyntaxError("Function in Function Error")
	}

	return nil

}

func (this *Interpreter) SplitOnTokenWithBrackets(tokens types.TokenList, tok types.Token) types.TokenListArray {

	/* vars */
	result := types.NewTokenListArray()
	var idx int
	var bc int
	//Token tt;

	//System.Out.Println("in:");
	//System.Out.Println(this.TokenListAsString(tokens));

	idx = 0
	bc = 0
	//SetLength(result, idx+1);
	result = result.Add(*types.NewTokenList())

	for _, tt := range tokens.Content {
		if (tt.Type == tok.Type) && (strings.ToLower(tt.Content) == strings.ToLower(tok.Content) || tok.Content == "") && (bc == 0) {
			idx++
			result = result.Add(*types.NewTokenList())
		} else {
			if tt.Type == types.OBRACKET || tt.Type == types.FUNCTION || tt.Type == types.PLUSFUNCTION {
				bc = bc + 1
			}
			if tt.Type == types.CBRACKET {
				bc = bc - 1
			}
			tl := result.Get(idx)
			tl.Push(tt)
		}
	}

	//System.Out.Println("out:");
	//for _, tl := range result
	//System.Out.Println(this.TokenListAsString(tl));

	//System.Exit(0);

	/* enforce non void return */
	return result

}

func (this *Interpreter) VariableTypeFromString(s string) types.VariableType {

	/* vars */
	var result types.VariableType

	s = strings.ToLower(s)

	if s == "string" {
		result = types.VT_STRING
	} else if s == "float" {
		result = types.VT_FLOAT
	} else if s == "integer" {
		result = types.VT_INTEGER
	} else if s == "boolean" {
		result = types.VT_BOOLEAN
	} else if s == "expression" {
		result = types.VT_EXPRESSION
	} else {
		result = types.VT_STRING
	}

	/* enforce non void return */
	return result

}

func (this *Interpreter) Halt() error {

	settings.VideoSuspended = false

	p := this.GetProducer()
	if p != nil {
		p.StopAudio()
	}

	this.Triggers.Empty()

	if this.Children != nil {
		return this.Children.Halt()
	}

	/* vars */
	var e error

	cpu := apple2helpers.GetCPU(this)
	if cpu.Halted == false {

		fmt.Println("Attempting to halt CPU")
		cpu.Halted = true

	}

	if (this.State != types.DIRECTRUNNING) && (this.State != types.RUNNING) {
		return nil
	}

	e = nil

	apple2helpers.SetSPEED(this, 255)

	//this.VDU.PutStr("@");
	//for ((this.Stack.Size() > 0) && (this.Stack.Get(this.Stack.Size()-1).Locals != this.Local))
	for (this.Stack.Size() > 0) && (this.Dialect.GetTitle() == "BCODE") {
		//this.VDU.PutStr("~");
		this.Return(true)
	}

	if this.PC.Line == -1 {
		this.PC.Line = this.Code.GetHighIndex()
	}

	if (this.State == types.RUNNING) && (this.PC.Line != 0) {
		e = errors.New("BREAK AT LINE " + utils.IntToStr(this.PC.Line))
	}

	//_ = files.DOSCLOSEALL()

	for this.Stack.Size() > 0 {
		//this.VDU.PutStr("~");
		this.Return(false)
	}

	this.State = types.STOPPED

	//if this.VDU.GetVideoMode().ActualRows == 0 {
	//	this.VDU.SetVideoMode(this.VDU.GetVideoModes()[this.VDU.CurrentMode()-1])
	//}

	if this.ExitOnEnd && this.GetParent() != nil {
		p := this.GetParent()
		p.SaveCPOS()
		p.SetChild(nil)
		this.SetParent(nil)
		p.WriteLayersToMemory()
		if this.ShouldSaveAndRestoreText() {
			apple2helpers.TextRestoreScreen(p)
		}
	}

	if this.DebugMe {
		debug.SetDebug(false)
	}

	this.SetInChannel("")

	this.Memory.KeyBufferEmpty(this.MemIndex)

	if e != nil {
		//this.Dialect.HandleException(this, e);
		return e
	}

	return nil

}

func (this *Interpreter) GetMusicVolume() (float32, bool) {
	ival := this.Memory.ReadInterpreterMemory(this.MemIndex, memory.OCTALYZER_DIGI_ATTENUATION)
	locked := (ival&128 != 0)
	v := 1 - (float32(ival&131072) / 100000)
	return v, locked
}

func (this *Interpreter) SetMusicVolume(v float32, locked bool) {
	ival := uint64((1 - v) * 100000)
	if locked {
		ival |= 131072
	}
	this.Memory.WriteInterpreterMemory(this.MemIndex, memory.OCTALYZER_DIGI_ATTENUATION, ival)
}

func (this *Interpreter) GetCode() *types.Algorithm {
	return this.Code
}

func (this *Interpreter) GetDirectAlgorithm() *types.Algorithm {
	return this.DirectAlgorithm
}

func (this *Interpreter) GetLPC() *types.CodeRef {
	return &this.LPC
}

func (this *Interpreter) GetPC() *types.CodeRef {
	return &this.PC
}

func (this *Interpreter) GetStack() *interfaces.CallStack {
	return &this.Stack
}

func (this *Interpreter) GetState() types.EntityState {

	if this.Children != nil {
		return this.Children.GetState()
	}

	return this.State
}

func (this *Interpreter) GetSubState() types.EntitySubState {

	if this.Children != nil {
		return this.Children.GetSubState()
	}

	return this.SubState
}

func (this *Interpreter) GetTokenStack() *types.TokenList {
	return &this.TokenStack
}

func (this *Interpreter) GetLocal() types.VarManager {
	return this.Local
}

func (this *Interpreter) GetName() string {
	return this.Name
}

func (this *Interpreter) GetVarPrefix() string {
	return this.VarPrefix
}

func (this *Interpreter) GetWaitUntil() time.Time {
	if this.Children != nil {
		return this.Children.GetWaitUntil()
	}
	return this.WaitUntil
}

func (this *Interpreter) SetCode(a *types.Algorithm) {
	this.Code = a
}

func (this *Interpreter) SetDirectAlgorithm(a *types.Algorithm) {
	this.DirectAlgorithm = a
}

func (this *Interpreter) SetLPC(c *types.CodeRef) {
	this.LPC = *c
}

func (this *Interpreter) SetName(s string) {
	this.Name = s
}

func (this *Interpreter) SetStack(s *interfaces.CallStack) {
	this.Stack = *s
}

func (this *Interpreter) SetState(s types.EntityState) {

	if this.Children != nil {
		this.Children.SetState(s)
		return
	}

	this.State = s
}

func (this *Interpreter) SetSubState(s types.EntitySubState) {

	if this.Children != nil {
		this.Children.SetSubState(s)
		return
	}

	this.SubState = s
}

func (this *Interpreter) SetVarPrefix(s string) {
	this.VarPrefix = s
}

func (this *Interpreter) SetWaitUntil(i time.Time) {
	this.WaitUntil = i
}

func (this *Interpreter) GetProducer() interfaces.Producable {
	return this.Producer
}

func (this *Interpreter) SetProducer(v interfaces.Producable) {
	this.Producer = v
}

// func (this *Interpreter) SetVDU(d interfaces.Display) {
// 	this.VDU = d
// }

func (this *Interpreter) GetDialect() interfaces.Dialecter {
	return this.Dialect
}

func (this *Interpreter) GetBreakpoint() *types.CodeRef {
	return &this.Breakpoint
}

func (this *Interpreter) GetDataRef() *types.CodeRef {
	return &this.Data
}

func (this *Interpreter) GetErrorTrap() *types.CodeRef {
	return &this.ErrorTrap
}

func (this *Interpreter) GetFirstString() string {
	return this.FirstString
}

func (this *Interpreter) SetFirstString(s string) {
	this.FirstString = s
}

func (this *Interpreter) SetWorkDir(s string) {
	if this.Children != nil {
		this.Children.SetWorkDir(s)
		return
	}
	this.WorkDir = s
}

func (this *Interpreter) SetProgramDir(s string) {
	if this.Children != nil {
		this.Children.SetProgramDir(s)
		return
	}
	this.ProgramDir = s
}

func (this *Interpreter) GetLoopBase() int {
	return this.LoopBase()
}

func (this *Interpreter) SetLocal(v types.VarManager) {
	this.Local = v
}

func (this *Interpreter) GetLoopStack() *types.LoopStack {
	return &this.LoopStack
}

func (this *Interpreter) GetLoopStates() types.LoopStateMap {
	return this.LoopStates
}

func (this *Interpreter) GetLoopStep() float64 {
	return this.LoopStep
}

func (this *Interpreter) SetLoopStep(v float64) {
	this.LoopStep = v
}

func (this *Interpreter) GetLoopVariable() string {
	return this.LoopVariable
}

func (this *Interpreter) SetOuterVars(b bool) {
	this.OuterVars = b
}

func (this *Interpreter) SetLoopVariable(s string) {
	this.LoopVariable = s
}

func (this *Interpreter) SetPC(c *types.CodeRef) {
	this.PC = *c
}

func (this *Interpreter) GetMemory(address int) uint64 {
	return this.Memory.ReadInterpreterMemory(this.MemIndex, address%memory.OCTALYZER_INTERPRETER_SIZE)
}

func (this *Interpreter) GetMultiArgFunc() interfaces.MafMap {
	return this.MultiArgFunc
}

func (this *Interpreter) SetMultiArgFunc(n string, maf interfaces.MultiArgumentFunction) {
	this.MultiArgFunc[n] = maf
}

func (this *Interpreter) GetWorkDir() string {
	if this.Children != nil {
		return this.Children.GetWorkDir()
	}
	return this.WorkDir
}

func (this *Interpreter) GetProgramDir() string {
	if this.Children != nil {
		return this.Children.GetProgramDir()
	}
	return this.ProgramDir
}

func (this *Interpreter) SetMemory(addr int, v uint64) {

	if addr == 0x36 || addr == 0x37 {
		fmt.Printf("Set %d, %d\n", addr, v)
	}
	// Preserve attribute markers for highermodes
	if (addr >= 1024) && (addr < 3072) {
		v = (v & 0xffff) | (this.GetMemory(addr) & 0xffff0000)
	}

	this.Memory.WriteInterpreterMemory(this.MemIndex, addr%memory.OCTALYZER_INTERPRETER_SIZE, v)

	//this.GetVDU().SetMemoryValue(addr, v)
}

func (this *Interpreter) SetMemorySilent(addr int, v uint64) {

	this.Memory.WriteInterpreterMemory(this.MemIndex, addr%memory.OCTALYZER_INTERPRETER_SIZE, v)

	//this.GetVDU().SetMemoryValue(addr, v)
}

func (this *Interpreter) BackHistory(s runestring.RuneString) runestring.RuneString {
	if this.HistIndex > 0 {
		this.HistIndex--
		s = this.History[this.HistIndex]
	}
	return s
}

func (this *Interpreter) ForwardHistory(s runestring.RuneString) runestring.RuneString {
	if this.HistIndex < len(this.History) {
		this.HistIndex++
		if this.HistIndex < len(this.History) {
			s = this.History[this.HistIndex]
		} else {
			s.Assign("")
		}
	}
	return s
}

func (this *Interpreter) LastHistory() runestring.RuneString {
	if len(this.History) == 0 {
		return runestring.Cast("")
	}
	return this.History[len(this.History)-1]
}

func (this *Interpreter) AddToHistory(cmd runestring.RuneString) {
	if len(this.History) == 0 {
		this.History = append(this.History, cmd)
	} else if string(cmd.Runes) != string(this.History[len(this.History)-1].Runes) {
		this.History = append(this.History, cmd)
	}

	this.HistIndex = len(this.History)
}

func (this *Interpreter) GetVSync() *syncmanager.VariableSyncher {
	return &syncmanager.Sync
}

func (this *Interpreter) ResetView() {
	i := this.GetMemoryMap().GetCameraConfigure(this.GetMemIndex())
	control := types.NewOrbitController(
		this.GetMemoryMap(),
		this.GetMemIndex(),
		i,
	)
	control.ResetALL()
	// control.SetPos(types.CWIDTH/2, types.CHEIGHT/2, types.CDIST*types.GFXMULT)
	// control.SetTarget(&glmath.Vector3{types.CWIDTH / 2, types.CHEIGHT / 2, 0})
	// control.SetPivotLock(true)
	// control.SetZoom(types.GFXMULT)
}

func (this *Interpreter) PeripheralReset() {

	index := this.GetMemIndex()
	mm := this.GetMemoryMap()
	for i := 0; i < 8; i++ {
		control := types.NewOrbitController(
			mm,
			index,
			i,
		)
		control.ResetALL()
		// control.SetPos(types.CWIDTH/2, types.CHEIGHT/2, types.CDIST*types.GFXMULT)
		// control.SetTarget(&glmath.Vector3{types.CWIDTH / 2, types.CHEIGHT / 2, 0})
		// control.SetPivotLock(true)
		// control.SetZoom(types.GFXMULT)
		//control.Update()
	}
	//apple2helpers.MODE40Preserve(this)
	fmt.Printf("Peripheral reset in slot %d\n", index)
	mm.IntSetBackdrop(index, "", 7, 1, 16, 0, false)
	mm.IntSetBackdropPos(index, types.CWIDTH/2, types.CHEIGHT/2, -types.CWIDTH/2)
	this.StopMusic()
	plus.ResetPaletteList(this.MemIndex)
}

func (this *Interpreter) Run(keepvars bool) {

	/* vars */

	if !keepvars {

		this.Zero(false)
		//System.Out.Println("After zero");

		this.PurgeOwnedVariables()
		//System.Out.Println("After Purge");

	}

	this.Reset()
	//System.Out.Println("After Reset");

	/* run pre execute hooks */
	for _, cc := range this.Dialect.GetCommands() {
		//System.Out.Println("Doing beforeRun() on "+name);
		cc.BeforeRun(this)
	}
	//System.Out.Println("After Before Execute");

	if this.PC.Line >= 0 {
		this.State = types.RUNNING
	}

	//this.VDU.SetSpeed(255)

	if s8webclient.CONN != nil && s8webclient.CONN.Session != "" {
		//this.VSync.SetSyncKey( this.Code.Checksum() )
		syncmanager.Sync.SetSyncKey("deadca7bdeadca7bdeadca7bdeadca7b")
	}

	if this.DebugMe {
		debug.SetDebug(true)
	}

	for i := 0; i < memory.OCTALYZER_MAX_PADDLES; i++ {
		this.Memory.IntSetPaddleValue(this.MemIndex, i, 127)
	}

	this.PreOptimizer()

	this.RunStart = time.Now().UnixNano()

	cpu := apple2helpers.GetCPU(this)
	mr, ok := this.GetMemoryMap().InterpreterMappableByLabel(this.GetMemIndex(), "Apple2IOChip")
	if ok {
		z := mr.(*apple2.Apple2IOChip)
		cpu.DoneFunc = z.AfterTask
		cpu.InitFunc = z.BeforeTask
	}

}

func (this *Interpreter) EndRemote() {

	if this.Children != nil {
		this.Children.EndRemote()
		return
	}

	if this.Remote == nil {
		return
	}

	//	fmt.Printf("*** Ending remotes for slot %d\n", this.GetMemIndex())
	this.Memory.SetCallback(nil, this.MemIndex)

	this.NeedRemoteKill = true
	for this.NeedRemoteKill {
		time.Sleep(10 * time.Millisecond)
	}

	this.SetState(types.STOPPED)
	apple2helpers.TEXT40(this)
}

func (this *Interpreter) IsNeedingInit() bool {
	return this.State == types.INITIALIZE
}

func (this *Interpreter) IsRunning() bool {

	/* vars */
	var result bool

	result = (this.State == types.RUNNING) || (this.Children != nil && this.Children.IsRunning())

	/* enforce non void return */
	return result

}

func (this *Interpreter) IsBreak() bool {

	/* vars */
	var result bool

	result = (this.State == types.BREAK)

	/* enforce non void return */
	return result

}

func (this *Interpreter) RunStatementDirect() {

	this.WaitForWorld()

	if this.Children != nil {
		if this.Children.IsRunningDirect() {
			this.Children.RunStatementDirect()
		} else if this.Children.IsRunning() {
			this.Children.RunStatement()
		}
		return
	}

	//	this.VDU.SetMemory(this.Memory)

	/* vars */
	var opc types.CodeRef
	var npc types.CodeRef
	var ln types.Line
	var st types.Statement
	var cc types.TokenList
	var Scope *types.Algorithm
	var n string
	var ok bool
	//	var breakrune rune = 3

	this.CheckProfile(false)

	if this.Dialect.HasCBreak(this) && this.IsBreakable() {
		this.NeedsPrompt = true
		this.SetBuffer(runestring.NewRuneString())
		e := this.Halt()
		if e != nil {
			this.GetDialect().HandleException(this, e)
		}
		return
	}

	if this.State != types.DIRECTRUNNING {
		//		log.Println("return due to incorrect state")
		return
	}

	//Producer.GetInstance().ExecutionContexts.Get(ThreadID) = self;

	/* IPS throttling */
	if this.IsWaiting() {
		//time.Sleep(50*time.Nanosecond)
		return
	}

	pos := -1

	/* Are we in a state machine based function */
	f := this.GetCommandState()
	var e error
	if f != nil {
		ss := this.GetSubState()
		cmd := f.CurrentCommand

		switch ss {
		case types.ESS_INIT:
			_, e = cmd.StateInit(nil, this, f.Params, f.Scope, f.PC)
			if this.GetSubState() == types.ESS_INIT {
				this.SetSubState(types.ESS_EXEC) // so the init runs only once
			}
		case types.ESS_SLEEP:
			if f.SleepCounter > 0 {
				f.SleepCounter-- // tick-tock!
				this.Wait(1000000)
			} else {
				this.SetSubState(f.PostSleepState)
			}
			return
		case types.ESS_EXEC:
			_, e = cmd.StateExec(nil, this, f.Params, f.Scope, f.PC)
		case types.ESS_DONE:
			_, e = cmd.StateDone(nil, this, f.Params, f.Scope, f.PC)
			this.SetCommandState(nil)
			npc = this.GetNextStatement(this.LPC)
			if npc.Line < 0 {
				//npc.Free;
				if this.CreatedTokens.Size() == 0 {
					this.State = types.STOPPED
					return
				}
			}
			this.LPC.Line = npc.Line
			this.LPC.Statement = npc.Statement
			this.LPC.Token = npc.Token
			this.LPC.SubIndex = 0
		case types.ESS_BREAK:
			e := this.Halt()
			if e != nil {
				this.GetDialect().HandleException(this, e)
			}
		}
		if e != nil {
			this.Dialect.HandleException(this, e)
		}
		return
	}

	//	log.Printf("Entered run statement direct, PC = %v\r\n", this.GetLPC())

	Scope = this.DirectAlgorithm

	pos = -1
	if this.CreatedTokens.Size() > 0 {
		kpos := this.CreatedTokens.IndexOfN(0, types.KEYWORD, "")
		dkpos := this.CreatedTokens.IndexOfN(0, types.DYNAMICKEYWORD, "")

		//fmt.Printf("kpos = %d, dpos = %d\n", kpos, dkpos)

		if kpos != -1 {
			pos = kpos
		}
		if dkpos != -1 && (pos == -1 || kpos > dkpos) {
			pos = dkpos
		}

		// Get first command
		if pos == -1 {
			pos = this.CreatedTokens.Size()
		}
		fc := this.CreatedTokens.SubList(0, pos)
		this.CreatedTokens.Content = this.CreatedTokens.Content[pos:]

		log.Printf("Got buffered command to execute: %s\n", this.TokenListAsString(*fc))

		e := this.Dialect.ExecuteDirectCommand(*fc, this, Scope, &this.LPC)
		log.Printf("Buffered command returns: %v", e)
		if e != nil {
			this.Dialect.HandleException(this, e)
			this.State = types.STOPPED
			this.CreatedTokens.Clear()
		}

		return
	}

	/* check pc valid */
	if !this.IsCodeRefValid(*this.GetLPC()) {
		if this.Dialect.GetShortName() == "logo" && this.Stack.Size() > 0 {
			this.Return(false)
			return
		}
		log.Println("return due to invalid code ref")
		log.Println(Scope.String())
		this.State = types.STOPPED
		return
	}

	//	log.Println("Code ref is okay")

	/* get statement */
	opc = *types.NewCodeRef()
	opc.Line = this.LPC.Line
	opc.Statement = this.LPC.Statement
	opc.Token = this.LPC.Token
	opc.SubIndex = 0
	n = this.Dialect.GetTitle()

	/* get appropriate tokens */
	ln, ok = this.CodeOptimizations.Get(this.LPC.Line)
	if !ok {
		ln, _ = Scope.Get(this.LPC.Line)
	}
	st = ln[this.LPC.Statement]

	/* create a copy of the list */
	cc = *st.SubList(0, st.Size())

	//	log.Printf("Executing (%s)\n", this.TokenListAsString(cc))

	/* now handle the statement */
	this.LPC.SubIndex = 0
	//try {
	//	now := time.Now()
	e = this.Dialect.ExecuteDirectCommand(cc, this, Scope, &this.LPC)
	//	diff := time.Since(now)

	//fmt.Printf("[time] Execute direct command %s took %v\n", cc.AsString(), diff)

	//    this.VDU.RegenerateMemory(this)
	//}
	//	log.Println("After execute direct command", e)
	if e != nil {
		this.Dialect.HandleException(this, e)
		this.State = types.STOPPED
	}

	if this.GetCommandState() != nil {
		return
	}

	/* Deassign temp list */
	//cc.Free;

	/* now if (we are here there was no exception maybe */
	if this.State == types.DIRECTEXEC6502 {
		npc = this.GetNextStatement(this.LPC)
		this.LPC.Line = npc.Line
		this.LPC.Statement = npc.Statement
		this.LPC.Token = npc.Token
		this.LPC.SubIndex = 0
		if npc.Line < 0 {
			this.PreCPUState = types.STOPPED
		}
		return
	}

	if this.State != types.DIRECTRUNNING {
		return
	}

	//System.Out.Println( "LPC == \", LPC.Line, \" / ", LPC.Statement, " / ", LPC.SubIndex );
	//System.Out.Println( "OPC == \", opc.Line, \" / ", opc.Statement, " / ", opc.SubIndex );

	if (this.Dialect.GetTitle() == n) && (opc.Line == this.LPC.Line) && (opc.Statement == this.LPC.Statement) && (opc.SubIndex == this.LPC.SubIndex) {
		/* increment program counter */
		//System.Out.Println("LOGO"," - advance pc");
		npc = this.GetNextStatement(this.LPC)
		if npc.Line < 0 {
			//npc.Free;
			if this.CreatedTokens.Size() == 0 {
				this.State = types.STOPPED
				return
			}
		}
		this.LPC.Line = npc.Line
		this.LPC.Statement = npc.Statement
		this.LPC.Token = npc.Token
		this.LPC.SubIndex = 0
		//npc.Free;
	} else {
		this.LPC.Token = 0
	}

	// save runtime state
	//this.Dialect.UpdateRuntimeState(this)

	this.SaveCPOS()

}

func (this *Interpreter) SplitOnToken(tokens types.TokenList, tok types.Token) types.TokenListArray {

	/* vars */
	result := types.NewTokenListArray()
	var idx int
	//Token tt;

	idx = 0
	//SetLength(result, idx+1);
	result = result.Add(*types.NewTokenList())

	for _, tt := range tokens.Content {
		if (tt.Type == tok.Type) && (tt.Content == tok.Content) {
			idx++
			//SetLength(result, idx+1);
			result = result.Add(*types.NewTokenList())
		} else {
			tl := result.Get(idx)
			tl.Push(tt)
		}
	}

	/* enforce non void return */
	return result

}

func (this *Interpreter) SplitOnTokenType(tokens types.TokenList, tok types.TokenType) types.TokenListArray {

	/* vars */
	result := types.NewTokenListArray()
	var idx int
	//Token tt;

	idx = 0
	//SetLength(result, idx+1);
	result = result.Add(*types.NewTokenList())

	for _, tt := range tokens.Content {
		if tt.Type == tok {
			idx++
			//SetLength(result, idx+1);
			result = result.Add(*types.NewTokenList())
		} else {
			tl := result.Get(idx)
			tl.Push(tt)
		}
	}

	/* enforce non void return */
	return result

}

func (this *Interpreter) SplitOnTokenTypeList(tokens types.TokenList, tok []types.TokenType) types.TokenListArray {

	/* vars */
	result := types.NewTokenListArray()
	var idx int
	//Token tt;

	idx = 0
	//SetLength(result, idx+1);
	result = result.Add(*types.NewTokenList())

	for _, tt := range tokens.Content {
		if tt.IsIn(tok) {
			idx++
			//SetLength(result, idx+1);
			result = result.Add(*types.NewTokenList())
			tl := result.Get(idx)
			tl.Push(tt)
		} else {
			tl := result.Get(idx)
			tl.Push(tt)
		}
	}

	/* enforce non void return */
	return result

}

func (this *Interpreter) GetTabWidth() int {
	return this.TabWidth
}

func (this *Interpreter) GetTokenAtCodeRef(current types.CodeRef) types.Token {

	/* vars */
	var result types.Token
	var ln types.Line
	var tl types.Statement
	var lCode *types.Algorithm

	if this.State == types.DIRECTRUNNING {
		lCode = this.DirectAlgorithm
	} else {
		lCode = this.Code
	}

	result = *types.NewToken(types.INVALID, "")
	if !this.IsCodeRefValid(current) {
		return result
	}

	ln, _ = lCode.Get(current.Line)
	tl = ln[current.Statement]

	result.Type = tl.Get(current.Token).Type
	result.Content = tl.Get(current.Token).Content

	/* enforce non void return */
	return result

}

func (this *Interpreter) SeekForwards(current types.CodeRef, eType types.TokenType, eValue string, pairType types.TokenType, pairValue string) types.CodeRef {

	/* vars */
	var result types.CodeRef
	var ptr types.CodeRef
	var notFound bool
	var nestCount int
	var t types.Token

	result = *types.NewCodeRef()
	result.Line = -1

	ptr = *types.NewCodeRefCopy(current)
	notFound = true

	nestCount = 0

	t = this.GetNextToken(&ptr)
	for (ptr.Line != -1) && (notFound) {
		if (pairType != types.NOP) && (t.Type == pairType) && (strings.ToLower(t.Content) == strings.ToLower(pairValue)) {
			nestCount++
		} else if (t.Type == eType) && (nestCount > 0) && (strings.ToLower(t.Content) == strings.ToLower(eValue)) {
			nestCount--
		} else if (t.Type == eType) && (nestCount == 0) && (strings.ToLower(t.Content) == strings.ToLower(eValue)) {
			notFound = false
			result.Line = ptr.Line
			result.Statement = ptr.Statement
			result.Token = ptr.Token
			return result
		}

		/* advance */
		t = this.GetNextToken(&ptr)
	}

	/* enforce non void return */
	return result

}

func (this *Interpreter) AddTrigger(slot int, condition *types.TokenList, line int) {
	this.Triggers.Add(slot, condition, line)
}

func (this *Interpreter) EndProgram() {

	/* vars */
	apple2helpers.SetSPEED(this, 255)

	this.Triggers.Empty()

	if (this.State != types.DIRECTRUNNING) && (this.State != types.RUNNING) {
		return
	}

	for this.Stack.Size() > 0 {
		//this.VDU.PutStr("~");
		this.Return(true)
	}

	this.State = types.STOPPED
	this.PC.Line = -1
	this.LPC.Line = -1

	if this.ExitOnEnd && this.GetParent() != nil {
		p := this.GetParent()
		p.SaveCPOS()
		if this.ShouldSaveAndRestoreText() {
			apple2helpers.TextRestoreScreen(p)
		}
		p.SetChild(nil)
		this.SetParent(nil)
	}

	if this.DebugMe {
		debug.SetDebug(false)
	}

	this.SetInChannel("")

}

func (this *Interpreter) Pop(trimloops bool) error {

	//System.Err.Println("POP");
	this.StackDump()

	/* vars */
	var cr *interfaces.StackEntry
	var save types.TokenList
	//TJanitorBob jb;

	/* is there an address on the stack? */
	if this.Stack.Size() == 0 {
		return errors.New("RETURN WITHOUT GOSUB ERROR")
	}
	//return NewESyntaxError( "Attempt to POP when not in subroutine/function" );

	cr, _ = this.Stack.Pop()

	if trimloops {
		for len(this.LoopStack) > cr.LoopStackSize {
			//this.LoopStack.Remove(len(this.LoopStack) - 1)
			this.LoopStack = this.LoopStack[0 : len(this.LoopStack)-1]
		}
	}

	save = *this.TokenStack.SubList(0, this.TokenStack.Size())
	if this.TokenStack.Size() > 0 {
		//System.Out.Println( "Items on stack after call.Equals("+this.TokenStack.Size() );
		//System.Out.Println( ")Item 0 is [" + this.TokenStack.LPeek().Content + "]" );
	}

	if this.State == types.RUNNING {
		//this.PC = cr.PC;
		//this.Code = cr.Code;
		this.State = cr.State
		if this.Local != cr.Locals {
			// FreeAndNil(Local);
			this.Local = cr.Locals
		}
		//if this.DataMap != cr.DataMap {
		// FreeAndNil(fDataMap);
		this.DataMap = cr.DataMap
		//}
		this.IsolateVars = cr.IsolateVars
		this.VarPrefix = cr.VarPrefix
		//this.TokenStack = cr.TokenStack;

		//if this.TokenStack != cr.TokenStack {
		// FreeAndNil(fTokenStack);
		this.TokenStack = *cr.TokenStack
		//}

		this.TokenStack = save

		this.Dialect = cr.CurrentDialect
		this.Registers = cr.Registers

		/* Handle allocated tokens */
		this.CreatedTokens = cr.CreatedTokens
		this.LoopVariable = cr.LoopVariable
		this.LoopStep = float64(cr.LoopStep)
		this.CurrentSubroutine = cr.CurrentSub
	} else {
		//this.LPC = cr.PC;
		//this.DirectAlgorithm = cr.Code;
		this.State = cr.State
		if this.Local != cr.Locals {
			this.Local = cr.Locals
		}
		//if this.DataMap != cr.DataMap {
		this.DataMap = cr.DataMap
		//}
		this.IsolateVars = cr.IsolateVars
		this.VarPrefix = cr.VarPrefix
		//if this.TokenStack != cr.TokenStack {
		// FreeAndNil(fTokenStack);
		this.TokenStack = *cr.TokenStack
		//}

		this.TokenStack = save

		this.Dialect = cr.CurrentDialect
		this.Registers = cr.Registers
		this.CreatedTokens = cr.CreatedTokens
		this.LoopVariable = cr.LoopVariable
		this.LoopStep = float64(cr.LoopStep)
		this.CurrentSubroutine = cr.CurrentSub
	}

	this.Triggers.Pop()

	// FreeAndNil(cr);
	return nil
}

func (this *Interpreter) ShowDef(n string) {
	log.Printf("===== stack")
	if this.DataMap.ContainsKey(strings.ToLower(n)) {
		log.Printf("[current] %s", n)
	}
	for i := this.Stack.Size() - 1; i >= 0; i-- {
		if this.Stack.Get(i).DataMap.ContainsKey(strings.ToLower(n)) {
			log.Printf("[%.7d] %s", i, n)
		}
	}
}

func (this *Interpreter) SetData(n string, v types.Token, local bool) {

	var found bool
	var foundLocal = this.DataMap.ContainsKey(strings.ToLower(n))
	var i int
	for i = this.Stack.Size() - 1; i >= 0; i-- {
		if this.Stack.Get(i).DataMap.ContainsKey(strings.ToLower(n)) {
			found = true
			break
		}
	}

	if local || foundLocal {
		// current level
		//log.Printf("Updating var %s at current scope", n)

		this.DataMap.Put(strings.ToLower(n), &types.Token{Content: v.Content, Type: v.Type, List: v.List, IsPropList: v.IsPropList})
		log.Printf("DATA = %+v", this.DataMap)
	} else if !found && this.Stack.Size() > 0 {
		//log.Printf("Updating var %s at top level", n)
		this.Stack.Get(0).DataMap.Put(strings.ToLower(n), &types.Token{Content: v.Content, Type: v.Type, List: v.List, IsPropList: v.IsPropList})
		log.Printf("DATA0 = %+v", this.Stack.Get(0).DataMap)
	} else if found {
		//log.Printf("Updating var %s at level %d", n, i)
		this.Stack.Get(i).DataMap.Put(strings.ToLower(n), &types.Token{Content: v.Content, Type: v.Type, List: v.List, IsPropList: v.IsPropList})
		log.Printf("DATA%d = %+v", i, this.Stack.Get(i).DataMap)
	} else {
		//log.Printf("Updating var %s at current scope", n)
		this.DataMap.Put(strings.ToLower(n), &types.Token{Content: v.Content, Type: v.Type, List: v.List, IsPropList: v.IsPropList})
		log.Printf("DATA = %+v", this.DataMap)
	}

}

func (this *Interpreter) HandleEvent(e types.Event) {

	if e.Name == "VARCHANGE" && e.Target == "SPEED" {
		this.SetSpeed(e.IntParam % 256)
	}

	if e.Name == "VARCHANGE" && e.Target == "HCOLOR" {
		apple2helpers.SetHCOLOR(this, e.IntParam)
	}

}

func (this *Interpreter) HandleError() bool {

	/* vars */
	var result bool

	result = false

	if this.State != types.RUNNING {
		return result
	}

	for (this.Stack.Size() > 0) && (this.Dialect.GetTitle() == "BCODE") {
		this.Return(false)
	}

	if this.ErrorTrap.Line == -1 {
		return result
	}

	if !this.Code.ContainsKey(this.ErrorTrap.Line) {
		return result
	}

	result = true
	this.Breakpoint.Line = this.PC.Line
	this.Breakpoint.Statement = this.PC.Statement
	this.Breakpoint.Token = 0
	this.Jump(this.ErrorTrap)

	/* enforce non void return */
	return result

}

func (this *Interpreter) Zero(coldstart bool) {

	this.RefList = types.NewReferenceList()

	//fmt.Println("ZERO: before:", this.GetMemory(108)*256+this.GetMemory(107))
	this.Dialect.InitVarmap(this, this.Local) // init varmap configs var memory in a dialect specific way
	//fmt.Println("ZERO: after:", this.GetMemory(108)*256+this.GetMemory(107))

	this.IsolateVars = false
	this.OuterVars = false

	this.Labels = make(map[string]int)

	this.Registers = *types.NewBRegistersBlank()

	this.State = types.EMPTY

	this.Stack = *interfaces.NewCallStack()

	this.TokenStack = *types.NewTokenList()

	this.DataMap = types.NewTokenMap()

	this.PC = *types.NewCodeRef()

	this.Data = *types.NewCodeRef()

	this.ErrorTrap = *types.NewCodeRef()

	this.Breakpoint = *types.NewCodeRef()

	this.VarPrefix = ""

	this.LoopStates = types.NewLoopStateMap()

	this.MultiArgFunc = interfaces.NewMafMap()

	this.LastExecuted = 0

	this.CreatedTokens = types.NewTokenList()

	this.SetMemory(32, 0)
	this.SetMemory(33, 40)
	this.SetMemory(34, 0)
	this.SetMemory(35, 24)

	this.LoopStep = 1
	this.LoopVariable = ""
	this.LoopStack = make(types.LoopStack, 0)

	// if this.VDU != nil {
	// 	for z := 0; z < 3; z++ {
	// 		this.VDU.SetPaddleValues(z, 127)
	// 		this.VDU.SetPaddleButtons(z, false)
	// 		this.VDU.SetPaddleModifier(z, 0)
	// 	}
	// }

	this.FirstString = ""

	this.CreatedTokens = types.NewTokenList()

}

// Prepends commands into the CreatedTokens structure
func (this *Interpreter) BufferCommands(commands *types.TokenList, times int) {
	for i := 0; i < times; i++ {
		for j := commands.Size() - 1; j >= 0; j-- {
			this.CreatedTokens.UnShift(commands.Get(j))
		}
	}

}

// Prepends commands into the CreatedTokens structure
func (this *Interpreter) BufferEmpty() {
	this.CreatedTokens.Clear()
	fmt.Println("CLEAR")
}

func (this *Interpreter) IsBufferEmpty() bool {
	return this.CreatedTokens.Size() == 0
}

func (this *Interpreter) CreateVarLower(n string, v types.Variable) {

	/* vars */
	if (v.Kind == types.VT_STRING) && (this.FirstString == "") {
		this.FirstString = n
		//System.Err.Println("FIRST STRING = "+n);
	}

	//this.VDU.PutStr("CVL ");
	if this.Stack.Size() > 0 {
		this.Stack.Get(0).Locals.Put(strings.ToLower(n), &v)
		//this.VDU.PutStr("STACK LEVEL 0");
	} else {
		this.Local.Put(strings.ToLower(n), &v)
		//this.VDU.PutStr("TOP");
	}

}

func (this *Interpreter) ReturnFromProc(value *types.Token) error {

	//	origcode := this.DirectAlgorithm

	var e error

	//for this.Stack.Size() > 0 && this.DirectAlgorithm == origcode {
	e = this.Return(true)

	this.CallReturnToken = value
	//}

	return e

}

func (this *Interpreter) Return(trimloops bool) error {

	/* vars */
	var cr *interfaces.StackEntry
	var save types.TokenList

	//System.Err.Println("RETURN");
	this.StackDump()

	/* is there an address on the stack? */
	if this.Stack.Size() == 0 {
		return exception.NewESyntaxError("Attempt to return when not in subroutine/function")
	}

	cr, _ = this.Stack.Pop()
	if trimloops {
		for len(this.LoopStack) > cr.LoopStackSize {
			this.LoopStack = this.LoopStack[0 : len(this.LoopStack)-1]
		}
	}

	save = *this.TokenStack.SubList(0, this.TokenStack.Size())

	if this.State == types.RUNNING {
		this.PC = *cr.PC
		this.Code = cr.Code
		this.State = cr.State
		if this.Local != cr.Locals {
			this.Local = cr.Locals
		}
		//if this.DataMap != cr.DataMap {
		this.DataMap = cr.DataMap
		//}
		this.IsolateVars = cr.IsolateVars
		this.VarPrefix = cr.VarPrefix
		//this.TokenStack = cr.TokenStack;

		//if this.TokenStack != cr.TokenStack {
		// FreeAndNil(fTokenStack);
		this.TokenStack = *cr.TokenStack
		//}

		this.TokenStack = save

		this.Dialect = cr.CurrentDialect
		this.Registers = cr.Registers
		this.CreatedTokens = cr.CreatedTokens
		this.LoopVariable = cr.LoopVariable
		this.LoopStep = float64(cr.LoopStep)
		this.CurrentSubroutine = cr.CurrentSub
	} else {

		//fmt.Printf("Vars before return: %v\n", this.DataMap.Keys())

		this.LPC = *cr.PC
		this.DirectAlgorithm = cr.Code
		this.State = cr.State
		if this.Local != cr.Locals {
			this.Local = cr.Locals
		}
		//if this.DataMap != cr.DataMap {
		this.DataMap = cr.DataMap
		//}
		this.IsolateVars = cr.IsolateVars
		this.VarPrefix = cr.VarPrefix
		//if this.TokenStack != cr.TokenStack {
		// FreeAndNil(fTokenStack);
		this.TokenStack = *cr.TokenStack
		//}

		this.TokenStack = save

		this.Dialect = cr.CurrentDialect
		this.Registers = cr.Registers
		this.CreatedTokens = cr.CreatedTokens
		this.LoopVariable = cr.LoopVariable
		this.LoopStep = float64(cr.LoopStep)
		this.CurrentSubroutine = cr.CurrentSub

		////fmt.Printf("Vars after return: %v\n", this.DataMap.Keys())
	}

	this.LPC.SubIndex = this.Stack.Size()

	// we might get stuck here if there is an invalid code ref, so if it is no good,
	// and we have not got pending repeat tokens, then collapse again?
	if this.CreatedTokens.Size() == 0 && this.LPC.Line == -1 && this.Stack.Size() > 0 {
		////fmt.Println("[info] Stack collapse is necessary to regain the PC")
		return this.Return(trimloops)
	}

	this.Triggers.Return()

	// FreeAndNil(cr);
	return nil
}

func (this *Interpreter) Deliver(im types.InfernalMessage) {

	//System.Out.Println( "entity named "+Name+" received message from "+im.Sender );

}

func (this *Interpreter) ExistsVar(n string) bool {

	/* vars */
	var result bool
	var i int

	if this.OuterVars {
		result = this.ExistsVarLower(n)
		return result
	}

	result = false
	if this.Local.ContainsKey(strings.ToLower(n)) {
		result = true
		return result
	}

	if this.IsolateVars {
		return result
	}

	for i = this.Stack.Size() - 1; i >= 0; i-- {
		if this.Stack.Get(i).Locals.ContainsKey(strings.ToLower(n)) {
			result = true
			return result
		}
	}

	/* enforce non void return */
	return result

}

func (this *Interpreter) IndicesFromTokens(tl types.TokenList, ob string, cb string) ([]int, error) {

	/* vars */
	var result []int = make([]int, 0)
	var front *types.Token
	var back *types.Token
	var dtok types.Token
	var tla types.TokenListArray
	//TokenList itl;
	var idx int

	// System.Err.Println(this.TokenListAsString(tl));

	if tl.Size() < 3 {
		return result, exception.NewESyntaxError("INVALID INDEX EXPRESSION")
	}

	front = tl.Shift()
	back = tl.Pop()

	//this.VDU.PutStr(front.Content+PasUtil.CRLF);
	//this.VDU.PutStr(back.Content+PasUtil.CRLF);

	if (front.Type == types.OBRACKET) && (front.Content == ob) &&
		(back.Type == types.CBRACKET) && (back.Content == cb) {
		tla = this.SplitOnTokenWithBrackets(tl, *types.NewToken(types.SEPARATOR, ","))
		result = make([]int, len(tla))

		idx = 0
		for _, itl := range tla {
			//System.Out.Println("itl: TokenType = "+itl.LPeek().Type+" "+itl.Size());
			if (itl.Size() == 1) && (itl.LPeek().Type == types.VARIABLE) {
				itl.Push(types.NewToken(types.OPERATOR, "+"))
				itl.Push(types.NewToken(types.NUMBER, "0"))
				//System.Out.Println("BEEP");
			}
			dtok = this.ParseTokensForResult(itl)
			//System.Out.Println("Token string = "+dtok.Content);
			// must be ttINTEGER
			if (dtok.Type != types.INTEGER) && (dtok.Type != types.NUMBER) {
				return result, exception.NewESyntaxError("INVALID INDEX EXPRESSION")
			}

			result[idx] = dtok.AsInteger()

			idx++
		}
	} else {
		return result, exception.NewESyntaxError("SYNTAX ERROR")
	}

	/* enforce non void return */
	return result, nil

}

func (this *Interpreter) InFunction() bool {

	return false

}

func (this *Interpreter) SeekBackwards(current types.CodeRef, eType types.TokenType, eValue string, pairType types.TokenType, pairValue string) types.CodeRef {

	/* vars */
	var result types.CodeRef
	var ptr types.CodeRef
	var notFound bool
	var nestCount int
	var t types.Token

	result = *types.NewCodeRef()
	result.Line = -1

	ptr = *types.NewCodeRefCopy(current)
	notFound = true

	nestCount = 0

	t = this.GetPrevToken(&ptr)
	for (ptr.Line != -1) && (notFound) {

		if (pairType != types.NOP) && (t.Type == pairType) && (strings.ToLower(t.Content) == strings.ToLower(pairValue)) {
			nestCount++
		} else if (t.Type == eType) && (nestCount > 0) && (strings.ToLower(t.Content) == strings.ToLower(eValue)) {
			nestCount--
		} else if (t.Type == eType) && (nestCount == 0) && (strings.ToLower(t.Content) == strings.ToLower(eValue)) {
			notFound = false
			result.Line = ptr.Line
			result.Statement = ptr.Statement
			result.Token = ptr.Token
			//System.Out.Println(")* FOUND IT");
			return result
		}

		/* advance */
		t = this.GetPrevToken(&ptr)
	}

	/* enforce non void return */
	return result

}

func (this *Interpreter) IndexOfLoopFromBase(forvarname string) int {
	match := -1
	i := 0
	for _, ls := range this.LoopStack {
		if (ls.VarName == forvarname) && (i > this.LoopBase()) {
			match = i
		}
		i++
	}
	return match
}

func (this *Interpreter) ExistsData(n string) bool {

	/* vars */
	var result bool
	var i int

	result = false
	if this.DataMap.ContainsKey(strings.ToLower(n)) {
		result = true
		return result
	}

	for i = this.Stack.Size() - 1; i >= 0; i-- {
		if this.Stack.Get(i).DataMap.ContainsKey(strings.ToLower(n)) {
			result = true
			return result
		}
	}

	/* enforce non void return */
	return result

}

func (this *Interpreter) Send(t string, content string) {

	/* vars */
	var im types.InfernalMessage

	im = *types.NewInfernalMessage(this.Name, t, content)
	this.Producer.Broadcast(im)

}

func (this *Interpreter) GetVarLower(n string) *types.Variable {

	/* vars */
	var result *types.Variable
	var i int

	result = nil

	for i = this.Stack.Size() - 1; i >= 0; i-- {
		if this.Stack.Get(i).Locals.ContainsKey(strings.ToLower(n)) {
			result = this.Stack.Get(i).Locals.Get(strings.ToLower(n))
			return result
		}
	}

	if this.Local.ContainsKey(strings.ToLower(n)) {
		result = this.Local.Get(strings.ToLower(n))
	}

	/* enforce non void return */
	return result

}

func (this *Interpreter) IsRunningDirect() bool {

	/* vars */
	var result bool

	result = (this.State == types.DIRECTRUNNING) || (this.Children != nil && this.Children.IsRunningDirect())

	/* enforce non void return */
	return result

}

func (this *Interpreter) IsCodeRefValid(current types.CodeRef) bool {

	/* vars */
	var result bool
	var ln types.Line
	var tl types.Statement
	var lCode *types.Algorithm

	//	log.Println(current)

	result = false

	if this.State == types.DIRECTRUNNING {
		lCode = this.DirectAlgorithm
	} else {
		lCode = this.Code
	}

	if !lCode.ContainsKey(current.Line) {
		//		log.Println("Line not exist")
		return result
	}

	ln, _ = lCode.Get(current.Line)

	if current.Statement >= len(ln) {
		//		log.Printf("Statement %d >= %d\r\n", current.Statement, len(ln))
		return result
	}

	tl = ln[current.Statement]

	if current.Token >= tl.Size() {
		//		log.Printf("Token %d >= %d\r\n", current.Token, tl.Size())
		return result
	}

	result = true

	/* enforce non void return */
	return result

}

func (this *Interpreter) RegisterWithParent() {
	if this.Parent != nil {
		this.Parent.SetChild(this)
	}
}

func (this *Interpreter) DeregisterWithParent() {
	if this.Parent != nil {
		this.Parent.SetChild(nil)
		this.Parent = nil
	}
}

func (this *Interpreter) GetChild() interfaces.Interpretable {
	return this.Children
}

func (this *Interpreter) GetParent() interfaces.Interpretable {
	return this.Parent
}

func (this *Interpreter) SetChild(a interfaces.Interpretable) {
	this.Children = a
}

func (this *Interpreter) SetParent(a interfaces.Interpretable) {
	this.Parent = a
}

func (this *Interpreter) NewChild(name string) interfaces.Interpretable {
	e := NewInterpreter(name, this.GetDialect(), this, this.Memory, this.MemIndex, this.SpecFile, this, this.vm)
	//e.SetVDU(this.GetVDU())
	e.Producer = this.Producer
	return e
}

func (this *Interpreter) NewChildWithParamsAndTask(name string, dialectname string, params *types.TokenList, task string) interfaces.Interpretable {
	e := NewInterpreter(name, this.GetDialect(), this, this.Memory, this.MemIndex, this.SpecFile, this, this.vm) // use same specfile
	//e.SetVDU(this.GetVDU())
	e.Producer = this.Producer
	e.SetParams(params)

	this.GetMemoryMap().BlockMapper[this.MemIndex].Reset(true)

	e.Bootstrap(dialectname, true)

	p := *types.NewTokenList()
	p.Add(types.NewToken(types.STRING, task))

	aalg := e.GetDirectAlgorithm()

	owd := this.GetWorkDir()

	_, ee := e.GetDialect().GetCommands()["load"].Execute(nil, e, p, aalg, *e.GetLPC())

	if ee != nil {
		e.EndProgram()
	}

	e.SetWorkDir(owd) // make sure we use the parents workdirectory by default

	// make it terminate on idle
	e.ExitOnEnd = true

	return e
}

func (this *Interpreter) FreezeStream(f io.Writer) {
	var d = make([]byte, 8)
	for i := 0; i < memory.OCTALYZER_INTERPRETER_SIZE; i++ {
		u := this.Memory.ReadInterpreterMemorySilent(this.MemIndex, i)
		d[0] = byte((u & 0xff00000000000000) >> 56)
		d[1] = byte((u & 0x00ff000000000000) >> 48)
		d[2] = byte((u & 0x0000ff0000000000) >> 40)
		d[3] = byte((u & 0x000000ff00000000) >> 32)
		d[4] = byte((u & 0x00000000ff000000) >> 24)
		d[5] = byte((u & 0x0000000000ff0000) >> 16)
		d[6] = byte((u & 0x000000000000ff00) >> 8)
		d[7] = byte(u & 0x00000000000000ff)
		f.Write(d)
	}

}

func (this *Interpreter) FreezeStreamLayers(f io.Writer) {
	// Save Layer state
	//this.ReadLayersFromMemory()
	//for _, l := range this.HUDLayers {
	//b, _ := l.MarshalBinary()
	//for len(b) < memory.OCTALYZER_LAYERSPEC_SIZE {
	//b = append(b, 0)
	//}
	//f.Write(b)
	//}
	//for _, l := range this.GFXLayers {
	//b, _ := l.MarshalBinary()
	//for len(b) < memory.OCTALYZER_LAYERSPEC_SIZE {
	//b = append(b, 0)
	//}
	//f.Write(b)
	//}
}

func (this *Interpreter) FreezeBytes() ([]byte, error) {
	if this.Children != nil {
		return this.Children.FreezeBytes()
	}

	//this.GetDialect().PreFreeze(this)

	f := bytes.NewBuffer(make([]byte, 0, 8*memory.OCTALYZER_INTERPRETER_SIZE))
	this.FreezeStream(f)

	return f.Bytes(), nil
}

func (this *Interpreter) ThawBytes(data []byte) error {
	if this.Children != nil {
		return this.Children.ThawBytes(data)
	}

	//	//fmt.Printf("running thaw in [%s]\n", this.Name)

	f := bytes.NewBuffer(data)

	this.ThawStreamInterpreterMemory(f)

	// Prior to reprocessing layers...
	chunk := make([]uint64, 0)
	for i := 0; i < 8; i++ {
		chunk = append(chunk, this.GetMemory(0x1fff8+i))
	}
	sf := types.UnpackName(chunk)
	if sf != "" {
		//	//fmt.Printf("<<< Reapplying spec %s >>>\n", sf)
		memory.WarmStart = true
		this.LoadSpec(sf)
		memory.WarmStart = false
	}

	// Now do layers
	//this.HUDLayers = make([]*types.LayerSpec, memory.OCTALYZER_MAX_HUD_LAYERS)
	this.ThawStreamLayers(f)

	this.GetDialect().PostThaw(this)

	return nil
}

func (this *Interpreter) ThawBytesNoPost(data []byte) error {
	if this.Children != nil {
		return this.Children.ThawBytesNoPost(data)
	}

	f := bytes.NewBuffer(data)

	this.ThawStreamInterpreterMemory(f)

	return nil
}

func (this *Interpreter) GetGFXLayerState() []bool {
	b := make([]bool, 0)
	for _, l := range this.GFXLayers {
		if l != nil {
			b = append(b, l.GetActive())
		} else {
			b = append(b, false)
		}
	}
	return b
}

func (this *Interpreter) SetGFXLayerState(v []bool) {
	for i, b := range v {
		if this.GFXLayers[i] != nil {
			this.GFXLayers[i].SetActive(b)
		}
	}
}

func (this *Interpreter) Freeze(filename string) error {

	if this.Children != nil {
		return this.Children.Freeze(filename)
	}

	this.GetDialect().PreFreeze(this)

	f, e := os.Create(filename)
	if e != nil {
		return e
	}
	defer f.Close()
	this.FreezeStream(f)

	return nil

}

func (this *Interpreter) ThawStreamInterpreterMemory(f io.Reader) error {
	var d = make([]byte, 8)
	var u uint64

	for i := 0; i < memory.OCTALYZER_INTERPRETER_SIZE; i++ {
		n, e := f.Read(d)
		if e != nil {
			return e
		}
		if n == 8 {

			u = (uint64(d[0]) << 56) | (uint64(d[1]) << 48) | (uint64(d[2]) << 40) | (uint64(d[3]) << 32) | (uint64(d[4]) << 24) | (uint64(d[5]) << 16) | (uint64(d[6]) << 8) | uint64(d[7])

			if (i < memory.OCTALYZER_MAPPED_CAM_BASE || i > memory.OCTALYZER_MAPPED_CAM_CONTROL) && i != memory.OCTALYZER_VIDEO_TINT {

				this.Memory.WriteInterpreterMemorySilent(this.MemIndex, i, u)

				//~ if i >= 8192 && i < 24576 {
				//~ this.Memory.WriteInterpreterMemory(this.MemIndex, i, u)
				//~ } else {
				//~ this.Memory.WriteInterpreterMemorySilent(this.MemIndex, i, u)
				//~ }

			}
		}
	}

	// We want to force the layers in this slot to be dirty
	for i := 0; i < memory.OCTALYZER_MAX_GFX_LAYERS; i++ {
		if this.GFXLayers[i] != nil /*&& this.GFXLayers[i].GetActive()*/ {
			this.GFXLayers[i].SetDirty(true)
		}
	}

	for i := 0; i < memory.OCTALYZER_MAX_HUD_LAYERS; i++ {
		if this.HUDLayers[i] != nil /*&& this.HUDLayers[i].GetActive()*/ {
			this.HUDLayers[i].SetDirty(true)
		}
	}

	return nil
}

func (this *Interpreter) ThawStreamLayers(f io.Reader) error {
	return nil
}

func (this *Interpreter) Thaw(filename string) error {

	if this.Children != nil {
		return this.Children.Thaw(filename)
	}

	f, e := os.Open(filename)
	if e != nil {
		return e
	}
	defer f.Close()

	this.ThawStreamInterpreterMemory(f)

	// Prior to reprocessing layers...
	chunk := make([]uint64, 0)
	for i := 0; i < 8; i++ {
		chunk = append(chunk, this.GetMemory(0x1fff8+i))
	}
	sf := types.UnpackName(chunk)
	if sf != "" {
		//		//fmt.Printf("<<< Reapplying spec %s >>>\n", sf)
		memory.WarmStart = true
		this.LoadSpec(sf)
		memory.WarmStart = false
	}

	// Now do layers
	this.ThawStreamLayers(f)

	this.GetDialect().PostThaw(this)

	return nil

}

func NewInterpreterThaw(name string, dia interfaces.Dialecter, parent interfaces.Interpretable, mm *memory.MemoryMap, memindex int, specfile string, p *Interpreter, image string, vm interfaces.VM) *Interpreter {
	this := NewInterpreter(name, dia, parent, mm, memindex, specfile, p, vm)
	this.Thaw(image)

	return this
}

func NewInterpreter(name string, dia interfaces.Dialecter, parent interfaces.Interpretable, mm *memory.MemoryMap, memindex int, specfile string, p *Interpreter, vm interfaces.VM) *Interpreter {
	this := &Interpreter{}
	this.CycleCounter = make([]interfaces.Countable, 0)
	this.History = make([]runestring.RuneString, 0)
	this.Name = name
	this.Dialect = dia
	this.Code = types.NewAlgorithm()
	this.CodeOptimizations = *types.NewAlgorithm()
	this.DirectAlgorithm = types.NewAlgorithm()
	this.Parent = parent
	this.Params = types.NewTokenList()
	this.Memory = mm
	this.MemIndex = memindex
	this.CommandBuffer = runestring.NewRuneString()
	this.SpecFile = ""
	this.PromptColor = 15
	this.UsePromptColor = false
	this.CurrentPage = "HGR1"
	this.DisplayPage = "HGR1"
	this.InputMapper = NewStdInputMatrix()
	this.clientSync = true
	this.Breakable = true
	this.Speed = 255 // default
	this.Labels = make(map[string]int)
	this.Triggers = NewTriggerTable(this)
	this.vm = vm

	if parent != nil {
		// inherit the uuid
		this.SetUUID(parent.GetUUID())
	}

	files.SetBlink0Callback(this.MemIndex, this.LED0)
	files.SetBlink1Callback(this.MemIndex, this.LED1)

	if this.Memory == nil {
		panic("memory is nil!")
	}

	this.StartTime = time.Now()
	this.NeedsPrompt = true

	//this.LoadSpec(specfile)

	//fmt.Printf("HUD layer count = %d\n", len(this.HUDLayers))

	//this.GetDialect().InitVDU(this, true)

	// Activate interp and mark layers as needing reconfig
	this.Memory.IntSetActiveState(this.MemIndex, 1)
	this.Memory.IntSetLayerState(this.MemIndex, 1)

	this.Zero(true)

	// txt := apple2helpers.GETHUD(this, "TEXT")
	// if txt != nil {
	// 	txt.Control.OnTBChange = this.TBChange
	// 	txt.Control.OnTBCheck = this.TBCheck
	// }

	this.SetMemory(103, 1)
	this.SetMemory(104, 8)
	this.clist = types.NewTokenList()

	if settings.Offline {
		this.SetWorkDir("/local/")
	} else if settings.IsRemInt {
		this.SetWorkDir("/appleii/")
	}

	return this
}

func (this *Interpreter) TBCheck(tb *types.TextBuffer) {
	return
}

func (this *Interpreter) BreakIntoVideo() {
	if this.Children != nil {
		this.Children.BreakIntoVideo()
		return
	}
	if !this.IsPlayingVideo() {
		return
	}
	this.p.RealTimeStop()
}

func (this *Interpreter) LED0(on bool) {
	var v uint64
	if on {
		v = 1
	}
	this.Memory.IntSetLED0(this.MemIndex, v)
}

func (this *Interpreter) LED1(on bool) {
	var v uint64
	if on {
		v = 1
	}
	this.Memory.IntSetLED1(this.MemIndex, v)
}

func (this *Interpreter) TBChange(tb *types.TextBuffer) {
	return
}

func (this *Interpreter) GetServer() *s8webclient.Client {
	return this.c
}

func (this *Interpreter) GetStartTime() time.Time {
	return this.StartTime
}

func (this *Interpreter) IsEmpty() bool {

	/* vars */
	var result bool

	result = (this.State == types.EMPTY)

	/* enforce non void return */
	return result

}

func (this *Interpreter) TokenListAsString(tokens types.TokenList) string {

	/* vars */
	var result string
	//Token tok;

	result = ""
	for _, tok := range tokens.Content {

		if tok == nil {
			continue
		}

		if result != "" {
			result = result + " "
		}

		if (tok.Type == types.KEYWORD) || (tok.Type == types.FUNCTION) || (tok.Type == types.DYNAMICKEYWORD) {
			result = result + tok.AsString()
		} else {
			result = result + tok.AsString()
		}
	}

	/* enforce non void return */
	return result

}

func (this *Interpreter) Jump(target types.CodeRef) {

	this.PC = target

}

func (this *Interpreter) StackDump() {
	//System.Err.Println("PC.Line = "+PC.Line);
	//System.Err.Println("PC.Statement = "+PC.Statement);
	for z := this.Stack.Size() - 1; z >= 0; z-- {
		//System.Err.Println(" from types.Line "+this.Stack.Get(z).PC.Line+", types.Statement "+this.Stack.Get(z).PC.Statement);
	}
}

func (this *Interpreter) LoopBase() int {
	if this.Stack.Size() == 0 {
		return -1
	}
	return this.Stack.Get(this.Stack.Size()-1).LoopStackSize - 1
}

func (this *Interpreter) ParseImm(s string) {
	this.SetDisabled(false)
	this.Parse(s)
	for this.IsRunningDirect() {
		this.RunStatementDirect()
	}
}

func (this *Interpreter) Parse(s string) {

	//fmt.Printf( "(:) Parse called for [%s], bytes (%v)\n", s, []byte(s) )

	for this.Children != nil {
		this.Children.Parse(s)
		return
	}

	this.CreatedTokens = types.NewTokenList()

	//	//fmt.Printf("Parse for [%s] by [%s]\n", s, target.GetName())

	//ss := time.Now()
	this.GetDialect().Parse(this, s)

	apple2helpers.TextHideCursor(this)

	////fmt.Printf("target.Dialect.Parse statement in %v\n", time.Since(ss))
	//	if target.IsRunningDirect() {
	//		target.PreOptimizer()
	//	}

	// update memory

}

func (this *Interpreter) CreateVar(n string, v types.Variable) {

	/* vars */

	if this.OuterVars {
		this.CreateVarLower(n, v)
		return
	}

	this.Local.Put(strings.ToLower(n), &v)
	if (v.Kind == types.VT_STRING) && (this.FirstString == "") {
		this.FirstString = n
	}

}

func (this *Interpreter) SplitOnTokenStartsWith(tokens types.TokenList, tok []types.TokenType) types.TokenListArray {

	/* vars */
	result := types.NewTokenListArray()
	var idx int
	//Token tt;

	idx = 0
	//SetLength(result, idx+1);
	result = result.Add(*types.NewTokenList())

	for _, tt := range tokens.Content {
		if tt.IsIn(tok) {
			if result[idx].Size() == 0 {
				result[idx].Push(tt)
			} else {
				idx++
				//SetLength(result, idx+1);
				result = result.Add(*types.NewTokenList())
				tl := result.Get(idx)
				tl.Push(tt)
			}
		} else {
			tl := result.Get(idx)
			tl.Push(tt)
		}
	}

	/* enforce non void return */
	return result

}

func (this *Interpreter) Reset() {

	/* vars */

	this.PC.Line = this.Code.GetLowIndex()
	this.PC.Statement = 0
	this.PC.Token = 0
	this.ErrorTrap.Line = -1

	fmt.Printf("After reset, first line is %d\n", this.PC.Line)

}

func (this *Interpreter) SetDialect(d interfaces.Dialecter, preserve bool, silent bool) {

	this.Dialect = d
	this.Dialect.InitVDU(this, silent)

	if !preserve {
		this.Clear()
		this.PurgeOwnedVariables()
		if s8webclient.CONN.IsAuthenticated() {
			this.SetWorkDir("/")
		} else {
			this.SetWorkDir("/local/")
		}

	}

	//settings.SetSubtitle(this.Dialect.GetTitle())

	_ = this.Halt()

}

func (this *Interpreter) ParseTokensForResult(tokens types.TokenList) types.Token {

	if this.Children != nil {

		return this.Children.ParseTokensForResult(tokens)

	}

	r, _ := this.Dialect.ParseTokensForResult(this, tokens)

	return *r

}

func (this *Interpreter) IsStopped() bool {

	/* vars */
	var result bool

	result = (this.State == types.STOPPED)

	/* enforce non void return */
	return result

}

func (this *Interpreter) ScreenReset() {
	// TextDrawBox(ent interfaces.Interpretable, x, y, w, h int, content string, shadow, window bool)
	apple2helpers.TEXT40(this)
}

func (this *Interpreter) SystemMessage(text string) {
	// TextDrawBox(ent interfaces.Interpretable, x, y, w, h int, content string, shadow, window bool)
	apple2helpers.SaveVSTATE(this)
	apple2helpers.TEXTMAX(this)

	cols := []uint64{15, 13, 6, 12, 3, 1, 2}
	for x := 0; x < 80; x++ {
		r := float32(x) / 80
		c := cols[int(r*float32(len(cols)))]
		apple2helpers.SetBGColor(this, c)
		apple2helpers.TextDrawBox(this, x, 0, 1, 48, "", false, false)
	}
	apple2helpers.SetBGColor(this, 0)
	apple2helpers.SetFGColor(this, 15)

	apple2helpers.SetCursorX(this, (80-len(text))/2)
	apple2helpers.SetCursorY(this, 23)

	apple2helpers.PutStr(this, text)
}

func (this *Interpreter) IsBlocked() bool {

	/* vars */
	var result bool
	var ct int64

	ct = time.Now().UnixNano() / int64(time.Millisecond)
	result = ((ct - this.LastExecuted) < int64(this.Dialect.GetIPS()))

	/* enforce non void return */
	return result

}

func (this *Interpreter) GetPrevStatement(current types.CodeRef) types.CodeRef {

	/* vars */
	var result types.CodeRef
	var ln types.Line
	var nl int
	var lCode *types.Algorithm

	if this.State == types.DIRECTRUNNING {
		lCode = this.DirectAlgorithm
	} else {
		lCode = this.Code
	}

	result = *types.NewCodeRef()
	result.Line = -1
	result.Statement = 0
	result.Token = 0

	if !this.IsCodeRefValid(current) {
		return result
	}

	/* current ref is valid */
	ln, _ = lCode.Get(current.Line)

	if current.Statement-1 >= 0 {
		result.Line = current.Line
		result.Statement = current.Statement - 1
		result.Token = 0
		return result
	}

	/* is there a next line */
	nl = lCode.PrevAfter(current.Line)

	if nl < 0 {
		return result
	}

	ln, _ = lCode.Get(nl)

	result.Line = nl
	result.Statement = len(ln) - 1
	result.Token = 0

	/* enforce non void return */
	return result

}

func (this *Interpreter) GetNextToken(current *types.CodeRef) types.Token {

	/* vars */
	var result types.Token
	var ln types.Line
	var tl types.Statement
	var ns types.CodeRef
	var lCode *types.Algorithm

	if this.State == types.DIRECTRUNNING {
		lCode = this.DirectAlgorithm
	} else {
		lCode = this.Code
	}

	result = *types.NewToken(types.INVALID, "")
	if !this.IsCodeRefValid(*current) {
		return result
	}

	ln, _ = lCode.Get(current.Line)
	tl = ln[current.Statement]

	/* can we just advance a token in the stream */
	if current.Token < tl.Size()-1 {
		current.Token = current.Token + 1
		result.Type = tl.Get(current.Token).Type
		result.Content = tl.Get(current.Token).Content
		return result
	}

	/* ok, can we advance a statement in the stream? */
	ns = this.GetNextStatement(*current)

	////fmt.Printf( "GetNextStatement(%d/%d) -> (%d/%d)\n", current.Line, current.Statement, ns.Line, ns.Statement )

	if ns.Line > -1 {
		current.Line = ns.Line
		current.Statement = ns.Statement
		current.Token = ns.Token
		ln, _ = lCode.Get(ns.Line)
		tl = ln[ns.Statement]
		result.Type = tl.Get(ns.Token).Type
		result.Content = tl.Get(ns.Token).Content
		return result
	}

	current.Line = -1

	/* enforce non void return */
	return result

}

func (this *Interpreter) Continue() {

	/* vars */

	if (this.PC.Line > 0) && (this.State == types.BREAK) {
		this.State = types.RUNNING
	}

}

func (this *Interpreter) IsWaiting() bool {

	if this.Children != nil {
		return this.Children.IsWaiting()
	}

	/* vars */
	var result bool
	var ct time.Time

	ct = time.Now()
	result = this.WaitUntil.After(ct)

	/* enforce non void return */
	return result

}

func (this *Interpreter) NextTokenInstance(current *types.CodeRef, ttype types.TokenType, tcontent string) bool {

	/* vars */
	var result bool
	var shadow types.CodeRef
	var tok types.Token

	result = false
	shadow = *types.NewCodeRefCopy(*current)

	for (shadow.Line != -1) && (result == false) {
		tok = this.GetNextToken(&shadow)

		////fmt.Printf("---> NextTokenInstance %d/%d/%d\n", shadow.Line, shadow.Statement, shadow.Token )

		//System.Out.Println( "* Looking for ",ttype,", ", tcontent );
		//System.Out.Println( "* line == \", shadow.Line, \", statement == \", shadow.Statement, \", token == ", shadow.Token;
		//readln;
		if tok.Type == ttype {
			if tcontent == "" {
				result = true
			} else {
				result = (strings.ToLower(tcontent) == strings.ToLower(tok.Content))
			}
		}
	}

	//if (result)
	//{
	/* update code reference */
	current.Line = shadow.Line
	current.Statement = shadow.Statement
	current.Token = shadow.Token
	current.SubIndex = 0
	//}

	/* enforce non void return */
	return result

}

func (this *Interpreter) IsPaused() bool {

	/* vars */
	var result bool

	result = (this.State == types.PAUSED)

	/* enforce non void return */
	return result

}

func (this *Interpreter) IsWaitingForWorld() bool {
	if this.Children != nil {
		return this.Children.IsWaitingForWorld()
	}
	return this.Paused
}

func (this *Interpreter) IsZ80Executing() bool {

	/* vars */
	var result bool

	result = (this.State == types.EXECZ80 || this.State == types.DIRECTEXECZ80)

	if !result {
		return false
	}

	cpu := apple2helpers.GetZ80CPU(this)

	if cpu.Halted {
		result = false
		this.State = this.PreCPUState
		if this.State == types.STOPPED {
			this.Halt()
		}
	}

	/* enforce non void return */
	return result

}

func (this *Interpreter) Is6502Executing() bool {

	/* vars */
	var result bool

	result = (this.State == types.EXEC6502 || this.State == types.DIRECTEXEC6502)

	if !result {
		return false
	}

	cpu := apple2helpers.GetCPU(this)

	if cpu.Halted {
		result = false
		this.State = this.PreCPUState
		if this.State == types.STOPPED {
			this.Halt()
		}
	}

	/* enforce non void return */
	return result

}

func (this *Interpreter) DoCycles6502() int {
	cpu := apple2helpers.GetCPU(this)

	r := cpu.ExecuteSliced()
	return int(r)
}

func (this *Interpreter) DoCyclesZ80() int {
	cpu := apple2helpers.GetZ80CPU(this)

	r := cpu.ExecuteSliced()
	return int(r)
}

func (this *Interpreter) Halt6502(r int) {

	if this.Children != nil {
		this.Children.Halt6502(r)
		return
	}

	cpu := apple2helpers.GetCPU(this)
	cpu.Halted = true
	//this.State = this.PreCPUState

}

func (this *Interpreter) HaltZ80(r int) {

	if this.Children != nil {
		this.Children.HaltZ80(r)
		return
	}

	cpu := apple2helpers.GetZ80CPU(this)
	cpu.Halted = true
	//this.State = this.PreCPUState

}

// Put us into 6502 mode
func (this *Interpreter) StartZ80(addr int, x, y, a, p, sp int) {

	this.PreCPUState = this.State // save current state
	if this.IsRunningDirect() {
		this.State = types.DIRECTEXECZ80
	} else {
		this.State = types.EXECZ80
	}
	cpu := apple2helpers.GetZ80CPU(this)
	// w, uw := cpu.HasUserWarp()

	// cpu.RAM = this.GetMemoryMap()
	// cpu.SetWarpUser(1.0)
	// cpu.CheckWarp()
	// if w {
	// 	cpu.SetWarpUser(uw)
	// }
	// cpu.CheckWarp()
	//apple2helpers.Exec6502CodeNB(this, a, x, y, addr, p, sp, true)
	cpu.EmuStartTime = time.Now().UnixNano()
	cpu.ResetSliced()

	cpu.Init()

	//log.Printf("In Start6502(%.4x)", addr)

}

// Put us into 6502 mode
func (this *Interpreter) Start6502(addr int, x, y, a, p, sp int) {

	this.PreCPUState = this.State // save current state
	if this.IsRunningDirect() {
		this.State = types.DIRECTEXEC6502
	} else {
		this.State = types.EXEC6502
	}
	cpu := apple2helpers.GetCPU(this)
	w, uw := cpu.HasUserWarp()

	cpu.RAM = this.GetMemoryMap()
	cpu.SetWarpUser(1.0)
	cpu.CheckWarp()
	if w {
		cpu.SetWarpUser(uw)
	}
	cpu.CheckWarp()
	apple2helpers.Exec6502CodeNB(this, a, x, y, addr, p, sp, true)
	cpu.EmuStartTime = time.Now().UnixNano()
	cpu.ResetSliced()

	cpu.Init()

	//log.Printf("In Start6502(%.4x)", addr)

}

func (this *Interpreter) RestorePrevState() {
	this.State = this.PreCPUState
	//log.Printf("Restore prev state to %s", this.State.String())
}

func (this *Interpreter) RunStatement() {

	this.WaitForWorld()

	if this.Children != nil {
		if this.Children.IsRunningDirect() {
			this.Children.RunStatementDirect()
		} else if this.Children.IsRunning() {
			this.Children.RunStatement()
		}
		return
	}

	/* vars */
	var opc types.CodeRef
	var npc types.CodeRef
	var ln types.Line
	var st types.Statement
	var cc types.TokenList
	var n string
	var ok bool

	this.CheckProfile(false)

	if this.Dialect.HasCBreak(this) && this.IsBreakable() {
		this.NeedsPrompt = true
		this.SetBuffer(runestring.NewRuneString())
		e := this.Halt()
		if e != nil {
			this.GetDialect().HandleException(this, e)
		}
		return
	}

	if this.State != types.RUNNING {
		debug.SetDebug(false)
		return
	}

	// Handle a trigger
	if len(this.Triggers.List) > 0 && !this.Triggers.InTrigger {
		if this.Triggers.Test() {
			return
		}
	}

	/* IPS throttling */
	if this.IsWaiting() {
		return
	}

	if this.InFunction() {
		/* In this case we are executing a dynamic function - we assume it has
		   already been set into this.FunctionPtr && this.FunctionPC points to
		   the current place in the function code */
		this.RunStatementFunction()
	}

	var e error
	var Scope *types.Algorithm
	/* Are we in a state machine based function */
	f := this.GetCommandState()
	if f != nil {
		ss := this.GetSubState()
		cmd := f.CurrentCommand

		switch ss {
		case types.ESS_INIT:
			_, e = cmd.StateInit(nil, this, f.Params, f.Scope, f.PC)
			if this.GetSubState() == types.ESS_INIT {
				this.SetSubState(types.ESS_EXEC) // so the init runs only once
			}
		case types.ESS_SLEEP:
			if f.SleepCounter > 0 {
				f.SleepCounter-- // tick-tock!
				this.Wait(1000000)
			} else {
				this.SetSubState(f.PostSleepState)
			}
			return
		case types.ESS_EXEC:
			_, e = cmd.StateExec(nil, this, f.Params, f.Scope, f.PC)
		case types.ESS_DONE:
			_, e = cmd.StateDone(nil, this, f.Params, f.Scope, f.PC)
			this.SetCommandState(nil)
			npc = this.GetNextStatement(this.PC)
			if npc.Line < 0 {
				//npc.Free;
				if this.CreatedTokens.Size() == 0 {
					this.State = types.STOPPED
					return
				}
			}
			this.PC.Line = npc.Line
			this.PC.Statement = npc.Statement
			this.PC.Token = npc.Token
			this.PC.SubIndex = 0
		}
		if e != nil {
			this.Dialect.HandleException(this, e)
		}
		return
	}

	/* check pc valid */
	if !this.IsCodeRefValid(this.PC) {
		debug.SetDebug(false)
		this.State = types.STOPPED
		this.Halt()
		return
	}

	/* get statement */
	opc = *types.NewCodeRef()
	opc.Line = this.PC.Line
	opc.Statement = this.PC.Statement
	opc.Token = this.PC.Token
	opc.SubIndex = 0
	n = this.Dialect.GetTitle()

	/* get appropriate tokens */
	ln, ok = this.CodeOptimizations.Get(this.PC.Line)
	Scope = this.Code
	if !ok {
		ln, _ = Scope.Get(this.PC.Line)
	}
	st = ln[this.PC.Statement]

	/* create a copy of the list */
	cc = *st.SubList(0, st.Size())

	/* now handle the statement */
	this.PC.SubIndex = 0
	e = this.Dialect.ExecuteDirectCommand(cc, this, this.Code, &this.PC)

	if e != nil {
		this.Dialect.HandleException(this, e)
	}

	if this.GetCommandState() != nil {
		return
	}

	/* Deassign temp list */
	/* now if (we are here there was no exception maybe */
	if this.State == types.EXEC6502 {
		npc = this.GetNextStatement(this.PC)
		this.PC.Line = npc.Line
		this.PC.Statement = npc.Statement
		this.PC.Token = npc.Token
		this.PC.SubIndex = 0
		if npc.Line < 0 {
			this.PreCPUState = types.STOPPED
		}
		return
	}

	/* now if (we are here there was no exception maybe */
	if this.State != types.RUNNING {
		files.DOSCLOSEALL()
		return
	}

	if (this.Dialect.GetTitle() == n) && (opc.Line == this.PC.Line) && (opc.Statement == this.PC.Statement) && (opc.SubIndex == this.PC.SubIndex) {
		npc = this.GetNextStatement(this.PC)
		if npc.Line < 0 {
			this.State = types.STOPPED
			files.DOSCLOSEALL()
			return
		}
		this.PC.Line = npc.Line
		this.PC.Statement = npc.Statement
		this.PC.Token = npc.Token
		this.PC.SubIndex = 0
	} else {
		this.PC.Token = 0
	}

	this.SaveCPOS()

}

func (this *Interpreter) GetNextStatement(current types.CodeRef) types.CodeRef {

	//dbg.PrintStack()

	/* vars */
	var result types.CodeRef
	var ln types.Line
	var nl int
	var lCode *types.Algorithm
	var ok bool

	if this.State == types.DIRECTRUNNING {
		lCode = this.DirectAlgorithm
	} else {
		lCode = this.Code
	}

	result = *types.NewCodeRef()
	result.Line = -1
	result.Statement = 0
	result.Token = 0

	// // don't advance PC with buffered commands
	// if this.CreatedTokens.Size() > 0 {
	// 	return current
	// }

	if !this.IsCodeRefValid(current) {
		return result
	}

	/* current ref is valid */
	if ln, ok = this.CodeOptimizations.Get(current.Line); !ok {
		ln, _ = lCode.Get(current.Line)
	}

	if current.Statement+1 < len(ln) {
		result.Line = current.Line
		result.Statement = current.Statement + 1
		result.Token = 0
		return result
	}

	/* is there a next line */
	nl = lCode.NextAfter(current.Line)

	if nl < 0 {
		return result
	}

	result.Line = nl
	result.Statement = 0
	result.Token = 0

	/* enforce non void return */
	return result

}

func (this *Interpreter) GetPrevToken(current *types.CodeRef) types.Token {

	/* vars */
	var result types.Token
	var ln types.Line
	var tl types.Statement
	var ns types.CodeRef
	var lCode *types.Algorithm

	if this.State == types.DIRECTRUNNING {
		lCode = this.DirectAlgorithm
	} else {
		lCode = this.Code
	}

	result = *types.NewToken(types.INVALID, "")
	if !this.IsCodeRefValid(*current) {
		return result
	}

	ln, _ = lCode.Get(current.Line)
	tl = ln[current.Statement]

	/* can we just advance a token in the stream */
	if current.Token > 0 {
		current.Token = current.Token - 1
		result.Type = tl.Get(current.Token).Type
		result.Content = tl.Get(current.Token).Content
		return result
	}

	/* ok, can we advance a statement in the stream? */
	ns = this.GetPrevStatement(*current)
	if ns.Line > -1 {
		current.Line = ns.Line
		current.Statement = ns.Statement
		ln, _ = lCode.Get(ns.Line)
		tl = ln[ns.Statement]
		ns.Token = tl.Size() - 1
		current.Token = ns.Token
		result.Type = tl.Get(ns.Token).Type
		result.Content = tl.Get(ns.Token).Content
		return result
	}

	current.Line = -1

	/* enforce non void return */
	return result

}

func (this *Interpreter) SetParams(p *types.TokenList) {
	this.Params = p.SubList(0, p.Size())
	//fmt.Printf("Entity passed params: %s\n", this.TokenListAsString(*this.Params))
}

// GetParams returns a copy of the arguments an entity was invoked with
func (this *Interpreter) GetParams() *types.TokenList {
	return this.Params
}

// ------------================[ Layers ]================-----------------

func (this *Interpreter) GetHUDLayerByID(name string) (*types.LayerSpecMapped, bool) {

	if len(this.HUDLayers) == 0 {
		this.HUDLayers, this.GFXLayers = this.vm.GetLayers()
	}

	// if this.MemIndex == 1 {
	// 	fmt.Printf("%d:GetHUDLayerByID(%s)\n", this.GetMemIndex(), name)
	// 	if this.HUDLayers[0] == nil {
	// 		panic("no layers in slot 1")
	// 	}
	// }
	var l *types.LayerSpecMapped

	for _, l = range this.HUDLayers {
		if l == nil {
			continue
		}
		// if this.MemIndex == 1 {
		// 	fmt.Printf("- %s vs %s\n", l.GetID(), name)
		// }
		if l.GetID() == name {
			return l, true
		}
	}

	return l, false

}

func (this *Interpreter) GetGFXLayerByID(name string) (*types.LayerSpecMapped, bool) {

	if len(this.HUDLayers) == 0 {
		this.HUDLayers, this.GFXLayers = this.vm.GetLayers()
	}

	var l *types.LayerSpecMapped

	for _, l = range this.GFXLayers {
		if l == nil {
			continue
		}
		if l.GetID() == name {
			return l, true
		}
	}

	return l, false

}

func (this *Interpreter) LoadSpec(specfile string) error {

	// if this.MemIndex == 1 {
	// 	fmt.Printf("=== Loading specfile %s to vm %d\n", specfile, this.MemIndex)
	// 	buf := make([]byte, 1<<16)
	// 	runtime.Stack(buf, false)
	// 	fmt.Printf("%s", buf)
	// }

	var e error
	if this.VM() != nil {
		this.HUDLayers, this.GFXLayers = this.VM().GetLayers() //hardware.LoadSpecToInterpreter(this, specfile)
	}
	this.SpecFile = specfile

	if len(this.HUDLayers) == 0 {
		panic(fmt.Sprintf("streuth no hud layers in %d", this.MemIndex))
	}

	// store spec at last 8 bytes memory
	data := types.PackName(specfile, 32)
	for i, v := range data {
		this.SetMemory(0x1fff8+i, v)
	}

	//	fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	//	fmt.Println(specfile)
	//	fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")

	//	//fmt.Printf("In INT Mappable structure:\n%v\n", this.Memory.InterpreterMappings)
	if specfile == settings.DefaultProfile {
		this.Memory.WriteInterpreterMemory(this.MemIndex, memory.OCTALYZER_INTERPRETER_PROFILE, 1)
	} else {
		this.Memory.WriteInterpreterMemory(this.MemIndex, memory.OCTALYZER_INTERPRETER_PROFILE, 0)
	}

	return e

}

func (this *Interpreter) ReadLayersFromMemory() {

	//index := this.MemIndex
	//mm := this.Memory

	//// Extract layers from memory
	//for i, l := range this.HUDLayers {
	//if l == nil {
	//l = &types.LayerSpec{}
	//this.HUDLayers[i] = l
	//}
	//hlbase := mm.MEMBASE(index) + memory.OCTALYZER_HUD_BASE
	//offset := memory.OCTALYZER_LAYERSPEC_SIZE * i
	//l.ReadFromMemory(mm, hlbase+offset)
	////		//fmt.Println("memory ->", l.String())
	//}

	//for i, l := range this.GFXLayers {
	//if l == nil {
	//l = &types.LayerSpec{}
	//this.GFXLayers[i] = l
	//}
	//hlbase := mm.MEMBASE(index) + memory.OCTALYZER_GFX_BASE
	//offset := memory.OCTALYZER_LAYERSPEC_SIZE * i
	//l.ReadFromMemory(mm, hlbase+offset)
	//}

}

func (this *Interpreter) WriteLayersToMemory() {

	//index := this.MemIndex
	//mm := this.Memory

	//// Extract layers from memory
	//for i, l := range this.HUDLayers {

	//if l == nil {
	//continue
	//}

	//hlbase := mm.MEMBASE(index) + memory.OCTALYZER_HUD_BASE
	//offset := memory.OCTALYZER_LAYERSPEC_SIZE * i
	//l.WriteToMemory(mm, hlbase+offset)

	//}

	//for i, l := range this.GFXLayers {
	//if l == nil {
	//continue
	//}
	//hlbase := mm.MEMBASE(index) + memory.OCTALYZER_GFX_BASE
	//offset := memory.OCTALYZER_LAYERSPEC_SIZE * i
	//l.WriteToMemory(mm, hlbase+offset)
	//}

	////this.WaitForLayers() // wait for zero state

	//// mark as changed
	//this.Memory.IntSetLayerState(this.MemIndex, 1)
}

func (this *Interpreter) GetHUDLayerSet() []*types.LayerSpecMapped {
	if len(this.HUDLayers) == 0 {
		this.HUDLayers, this.GFXLayers = this.vm.GetLayers()
	}

	return this.HUDLayers
}

func (this *Interpreter) GetGFXLayerSet() []*types.LayerSpecMapped {
	if len(this.GFXLayers) == 0 {
		this.HUDLayers, this.GFXLayers = this.vm.GetLayers()
	}
	return this.GFXLayers
}

func (this *Interpreter) WaitForLayers() {
	//if !this.clientSync {
	//return
	//}
	//for this.Memory.IntGetLayerState(this.MemIndex) > 0 {
	//time.Sleep(1 * time.Millisecond)
	//}
}

func (this *Interpreter) GetMemoryMap() *memory.MemoryMap {
	return this.Memory
}

func (this *Interpreter) UpdateCompletions() {
	// DID SOMETHING SO UPDATE COMPLETIONS
	if len(this.CommandBuffer.Runes) > 0 && this.InsertPos >= len(this.CommandBuffer.Runes) {
		amt, clist := this.Dialect.GetCompletions(this, this.CommandBuffer, this.InsertPos)
		//		for i, t := range clist.Content {
		//			//			fmt.Printf("(%d) %s %s (off %d)\n", i, t.Content, t.Type.String(), amt)
		//		}
		this.cprefix = amt
		this.clist = clist
		this.cptr = 0
		this.wantcompletion = false
	}
}

func (this *Interpreter) NextCompletion() {
	this.cptr++
	if this.cptr >= this.clist.Size() {
		this.cptr = 0
	}
	this.wantcompletion = false
}

func (this *Interpreter) PrevCompletion() {
	this.cptr--
	if this.cptr < 0 {
		this.cptr = this.clist.Size() - 1
	}
	this.wantcompletion = false
}

func getStops(text runestring.RuneString) []int {
	stops := make([]int, 0)
	//stops = append( stops, 0 ) // always stop at 0
	for i, ch := range text.Runes {
		if ch == 32 && i != 0 {
			stops = append(stops, i)
		}
	}
	//stops = append(stops, len(text.Runes))
	return stops
}

func wordToLeft(text runestring.RuneString, pos int) int {

	stops := getStops(text)

	start := pos - 1
	for i := len(stops) - 1; i >= 0; i-- {
		if stops[i] < start {
			return stops[i] + 1
		}
	}

	return 0

}

func wordToRight(text runestring.RuneString, pos int) int {
	stops := getStops(text)

	start := pos
	for i := 0; i < len(stops); i++ {
		if stops[i] > start {
			return stops[i] + 1
		}
	}

	return len(text.Runes)
}

func (this *Interpreter) DoCompletion() {
	line := this.CommandBuffer.SubString(0, len(this.CommandBuffer.Runes))
	linepos := this.InsertPos
	nrs := runestring.Cast("")
	nrs.AppendSlice(line.Runes[0:linepos])
	if this.compupper {
		nrs.Append(strings.ToUpper(this.clist.Content[this.cptr].Content[this.cprefix:]))
	} else {
		nrs.Append(strings.ToLower(this.clist.Content[this.cptr].Content[this.cprefix:]))
	}
	nrs.AppendSlice(line.Runes[linepos:])
	this.CommandBuffer = nrs
	this.InsertPos += len(this.clist.Content[this.cptr].Content[this.cprefix:])
	this.clist = types.NewTokenList()
	this.wantcompletion = false
}

func (this *Interpreter) SetUsePromptColor(b bool) {
	this.UsePromptColor = b
}

func (this *Interpreter) SetPromptColor(i int) {
	this.PromptColor = uint64(i & 15)
}

func (this *Interpreter) Interactive() {

	if this.Children != nil {

		this.Children.Interactive()
		return

	}

	if this.Remote != nil {
		this.NeedsPrompt = false
		return
	}

	settings.VideoSuspended = false

	this.WaitForWorld()

	////fmt.Printf("Interactive check for %d\n", this.MemIndex)
	apple2helpers.TextShowCursor(this)
	this.CheckProfile(false)

	if this.IsRunning() || this.IsRunningDirect() || settings.PureBoot(this.MemIndex) {
		return
	}

	//this.Dialect.UpdateRuntimeState(this)

	columns := this.GetColumns()
	rows := this.GetRows()

	if this.NeedsPrompt {
		//files.DOSCLOSEALL()
		this.NeedsPrompt = false
		text, ok := this.GetHUDLayerByID("TEXT")
		if ok && text.Control != nil {
			text.Control.NLIN()
			this.DoPrompt()
			this.RedoLine(true, columns, rows, this.CommandBuffer, this.InsertPos, -1, -1)
		}
		this.clist = types.NewTokenList()
		//		//fmt.Printf("Post Prompt PX=%d, PY=%d\n", this.Px, this.Py)
	}

	//this.Memory.IntSetAltChars(this.MemIndex, true)

	// We access the super 8 keybuffer directly in this case, as it is "implementation agnostic"
	if this.Memory.KeyBufferSize(this.MemIndex) > 0 || len(this.PasteBuffer.Runes) > 0 {

		//fmt.Printf("=====================================> Slot %d has a key...\n", this.MemIndex)

		var ch rune
		gap := 1000 * time.Millisecond / time.Duration(settings.PasteCPS)

		if len(this.PasteBuffer.Runes) > 0 && time.Since(this.LastPasteTime) > gap {

			ch = this.PasteBuffer.Runes[0]
			this.PasteBuffer = runestring.Delete(this.PasteBuffer, 1, 1)
			this.LastPasteTime = time.Now()

		} else {
			// Get scan code
			ch = rune(this.Memory.KeyBufferGet(this.MemIndex))
			////fmt.Printf("Got scancode %d from super8 buffer\n", ch)
		}

		switch {
		case ch == vduconst.PASTE || ch == vduconst.SHIFT_CTRL_V:
			text, err := clipboard.ReadAll()
			if err == nil {
				text = strings.Replace(text, "\r\n", "\r", -1)
				this.PasteBuffer = runestring.NewRuneString()
				this.PasteBuffer.Append(text)
			}
			this.wantcompletion = false
		case ch == vduconst.END:
			this.InsertPos = len(this.CommandBuffer.Runes)
			this.RedoLine(true, columns, rows, this.CommandBuffer, this.InsertPos, -1, -1)
			this.wantcompletion = false
		case ch == vduconst.HOME:
			this.InsertPos = 0
			this.RedoLine(true, columns, rows, this.CommandBuffer, this.InsertPos, -1, -1)
			this.wantcompletion = false
		case ch == vduconst.SHIFT_CSR_LEFT:
			this.InsertPos = wordToLeft(this.CommandBuffer, this.InsertPos)
			this.RedoLine(true, columns, rows, this.CommandBuffer, this.InsertPos, -1, -1)
		case ch == vduconst.SHIFT_CSR_RIGHT:
			this.InsertPos = wordToRight(this.CommandBuffer, this.InsertPos)
			this.RedoLine(true, columns, rows, this.CommandBuffer, this.InsertPos, -1, -1)
		case ch == vduconst.CSR_LEFT:
			if this.InsertPos > 0 {
				this.InsertPos--
				this.RedoLine(true, columns, rows, this.CommandBuffer, this.InsertPos, -1, -1)
			}
			this.wantcompletion = false
		case ch == 7:
			apple2helpers.Beep(this)
		case ch == 27:
			this.clist = types.NewTokenList()
			line := this.CommandBuffer.SubString(0, len(this.CommandBuffer.Runes))
			linepos := this.InsertPos
			this.RedoLine(true, columns, rows, line, linepos, -1, -1)
			this.wantcompletion = false
		case ch == 9:
			if this.clist != nil && this.clist.Size() > 0 {
				this.DoCompletion()
				this.RedoLine(true, columns, rows, this.CommandBuffer, this.InsertPos, -1, -1)
			}
		case ch == vduconst.CSR_RIGHT:
			if this.clist != nil && this.clist.Size() > 0 {
				this.DoCompletion()
				this.RedoLine(true, columns, rows, this.CommandBuffer, this.InsertPos, -1, -1)
			} else if this.InsertPos < len(this.CommandBuffer.Runes) {
				this.InsertPos++
				this.RedoLine(true, columns, rows, this.CommandBuffer, this.InsertPos, -1, -1)
			}
			this.wantcompletion = false
		case ch == vduconst.CSR_UP:

			if this.clist != nil && this.clist.Size() > 1 {
				this.PrevCompletion()
			} else {
				if this.InsertPos >= columns {
					this.InsertPos -= columns
				} else {
					this.CommandBuffer = this.BackHistory(this.CommandBuffer)
					this.InsertPos = len(this.CommandBuffer.Runes)
				}
			}

			line := this.CommandBuffer.SubString(0, len(this.CommandBuffer.Runes))
			linepos := this.InsertPos
			hs := -1
			he := -1
			if this.clist != nil && this.clist.Size() > 0 {
				nrs := runestring.Cast("")
				nrs.AppendSlice(line.Runes[0:linepos])
				hs = linepos
				if this.compupper {
					nrs.Append(strings.ToUpper(this.clist.Content[this.cptr].Content[this.cprefix:]))
				} else {
					nrs.Append(strings.ToLower(this.clist.Content[this.cptr].Content[this.cprefix:]))
				}
				he = len(nrs.Runes)
				nrs.AppendSlice(line.Runes[linepos:])
				line = nrs
				if this.wantcompletion {
					linepos += len(this.clist.Content[this.cptr].Content[this.cprefix:])
				}
			}

			this.RedoLine(true, columns, rows, line, linepos, hs, he)
			this.wantcompletion = false
		case ch == vduconst.CSR_DOWN:
			if this.clist != nil && this.clist.Size() > 1 {
				this.NextCompletion()
			} else {
				if this.InsertPos < len(this.CommandBuffer.Runes)-columns {
					this.InsertPos += columns
				} else {
					this.CommandBuffer = this.ForwardHistory(this.CommandBuffer)
					this.InsertPos = len(this.CommandBuffer.Runes)
				}
			}

			line := this.CommandBuffer.SubString(0, len(this.CommandBuffer.Runes))
			linepos := this.InsertPos
			hs := -1
			he := -1

			// modify with current completion
			if this.clist != nil && this.clist.Size() > 0 {
				hs = linepos
				nrs := runestring.Cast("")
				nrs.AppendSlice(line.Runes[0:linepos])
				if this.compupper {
					nrs.Append(strings.ToUpper(this.clist.Content[this.cptr].Content[this.cprefix:]))
				} else {
					nrs.Append(strings.ToLower(this.clist.Content[this.cptr].Content[this.cprefix:]))
				}
				he = len(nrs.Runes)
				nrs.AppendSlice(line.Runes[linepos:])
				line = nrs
				if this.wantcompletion {
					linepos += len(this.clist.Content[this.cptr].Content[this.cprefix:])
				}
			}

			this.RedoLine(true, columns, rows, line, linepos, hs, he)
			this.wantcompletion = false
		case ch == 10:
			{
				this.wantcompletion = false
				this.RedoLine(false, columns, rows, this.CommandBuffer, this.InsertPos, -1, -1)
				this.PutStr("\r\n")
				if len(this.CommandBuffer.Runes) > 0 {
					//s8webclient.CONN.LogMessage("UKL", string(this.CommandBuffer.Runes))
					this.Parse(string(this.CommandBuffer.Runes))
					this.AddToHistory(this.CommandBuffer)
					this.CommandBuffer.Assign("")
				}
				this.NeedsPrompt = true
				this.InsertPos = 0
			}
		case ch == 13:
			{
				this.wantcompletion = false
				this.RedoLine(false, columns, rows, this.CommandBuffer, this.InsertPos, -1, -1)
				this.PutStr("\r\n")

				if len(this.CommandBuffer.Runes) > 0 {
					this.NeedsPrompt = true
					//if s8webclient.CONN != nil {
					//s8webclient.CONN.LogMessage("UKL", string(this.CommandBuffer.Runes))
					//}
					this.Parse(string(this.CommandBuffer.Runes))
					this.NeedsPrompt = true
					this.AddToHistory(this.CommandBuffer)
					this.CommandBuffer.Assign("")
				}
				this.NeedsPrompt = true
				this.InsertPos = 0

			}
		case ch == 127:
			{
				this.clist = types.NewTokenList()
				this.wantcompletion = false
				if len(this.CommandBuffer.Runes) > 0 {
					if this.InsertPos >= len(this.CommandBuffer.Runes) {
						this.CommandBuffer = runestring.Copy(this.CommandBuffer, 1, len(this.CommandBuffer.Runes)-1)
						this.InsertPos = len(this.CommandBuffer.Runes)
						this.RedoLine(true, columns, rows, this.CommandBuffer, this.InsertPos, -1, -1)
					} else if this.InsertPos > 0 {
						a := runestring.Copy(this.CommandBuffer, 1, this.InsertPos-1)
						b := runestring.Delete(this.CommandBuffer, 1, this.InsertPos)
						a.AppendRunes(b)
						this.CommandBuffer.Assign(string(a.Runes))
						this.InsertPos--
						this.RedoLine(true, columns, rows, this.CommandBuffer, this.InsertPos, -1, -1)
					}
				}
				break
			}
		case ch < 32:
			break
		default:
			{
				if ch > 0x4000 {
					break
				}

				if ch >= vduconst.SHIFT_CTRL_A && ch <= vduconst.SHIFT_CTRL_Z {
					ch = 0
				}

				if this.Dialect.IsUpperOnly() && ch >= 'a' && ch <= 'z' {
					ch -= 32
				}

				if ch >= 'A' && ch <= 'Z' {
					this.compupper = true
				} else if ch >= 'a' && ch <= 'z' {
					this.compupper = false
				}

				// complete and continue
				if this.wantcompletion {
					this.DoCompletion()
				}

				if this.InsertPos < len(this.CommandBuffer.Runes) {
					a := runestring.Copy(this.CommandBuffer, 1, this.InsertPos)
					b := runestring.Delete(this.CommandBuffer, 1, this.InsertPos)
					a.AppendSlice([]rune{ch})
					a.AppendRunes(b)
					this.CommandBuffer.Assign(string(a.Runes))
				} else {
					this.CommandBuffer.Append(string(ch))
				}
				this.InsertPos++

				this.UpdateCompletions()

				line := this.CommandBuffer.SubString(0, len(this.CommandBuffer.Runes))
				linepos := this.InsertPos

				hs := -1
				he := -1

				// modify with current completion
				if this.clist != nil && this.clist.Size() > 0 {
					nrs := runestring.Cast("")
					nrs.AppendSlice(line.Runes[0:linepos])
					hs = linepos
					if this.compupper {
						nrs.Append(strings.ToUpper(this.clist.Content[this.cptr].Content[this.cprefix:]))
					} else {
						nrs.Append(strings.ToLower(this.clist.Content[this.cptr].Content[this.cprefix:]))
					}
					he = len(nrs.Runes)
					nrs.AppendSlice(line.Runes[linepos:])
					line = nrs
					if this.wantcompletion {
						linepos += len(this.clist.Content[this.cptr].Content[this.cprefix:])
					}
				}

				this.RedoLine(true, columns, rows, line, linepos, hs, he)
				break
			}
		}

	} else {
		// do nothing and return
		this.Wait(5000000)
		return
	}

}

func (this *Interpreter) RedoLine(pcursor bool, columns, rows int, line runestring.RuneString, linepos int, hs, he int) {
	sx, sy, ex, ey, ww, hh := this.Dialect.GetRealWindow(this)
	rc, _ := (ex - sx + 1), (ey - sy + 1)

	hh = 48 / apple2helpers.GetFullRows(this)
	ww = 80 / apple2helpers.GetFullColumns(this)

	if rc == 0 {
		apple2helpers.TEXT40(this)
		sx, sy, ex, ey, ww, hh = this.Dialect.GetRealWindow(this)
		rc, _ = (ex - sx + 1), (ey - sy + 1)
	}

	ln := ((this.Px + len(line.Runes)*ww) / rc) + 1
	// fix for long wrapped lines

	fl := (ey - this.Py + 1) / hh

	if fl < ln {
		diff := ln - fl
		apple2helpers.ScrollWindowBy(this, diff*hh)
		this.Py = this.Py - hh*diff
		fl = (ey - this.Py + 1) / hh
	}

	// hide cursor here
	apple2helpers.TextHideCursor(this)
	apple2helpers.Gotoxy(this, sx, this.Py)
	this.DoPrompt()
	this.ClearToBottom()

	var sh uint64 = 0
	var hsh uint64 = 3
	for i, r := range line.Runes {
		if i >= hs && i <= he {
			apple2helpers.SetShade(this, hsh)
		} else {
			apple2helpers.SetShade(this, sh)
		}
		apple2helpers.RealPut(this, r)
	}
	apple2helpers.SetShade(this, sh)

	if pcursor {

		vip := 0
		for i := 0; i < linepos && i < len(line.Runes); i++ {
			r := line.Runes[i]
			if r >= 32 && r < 2048 {
				vip++
			}
		}

		nx := sx + ((this.Px + vip*ww - sx) % rc)
		ny := this.Py + ((this.Px+vip*ww-sx)/rc)*hh

		this.Dialect.SetRealCursorPos(this, nx, ny)
	}
	apple2helpers.TextShowCursor(this)
}

func (this *Interpreter) GetBuffer() runestring.RuneString {
	return this.CommandBuffer
}

func (this *Interpreter) SetBuffer(r runestring.RuneString) {
	this.CommandBuffer = r
}

func (this *Interpreter) PutStr(s string) {
	this.Dialect.PutStr(this, s)
}

func (this *Interpreter) Backspace() {
	this.Dialect.Backspace(this)
}

func (this *Interpreter) RealPut(ch rune) {
	this.Dialect.RealPut(this, ch)
}

func (this *Interpreter) GetFeedBuffer() string {
	return this.FeedBuffer
}

func (this *Interpreter) PassWaveBuffer(channel int, data []float32, loop bool, rate int) {

	l := len(data)
	if l > memory.OCTALYZER_SPEAKER_MAX {
		l = memory.OCTALYZER_SPEAKER_MAX
	}

	// switch channel {
	// case 0:
	// 	for this.Memory.ReadGlobal(this.Memory.MEMBASE(this.GetMemIndex())+memory.OCTALYZER_SPEAKER_PLAYSTATE) != 0 {
	// 		time.Sleep(5 * time.Microsecond)
	// 	}
	// case 1:
	// 	for this.Memory.ReadGlobal(this.Memory.MEMBASE(this.GetMemIndex())+memory.OCTALYZER_CASSETTE_PLAYSTATE) != 0 {
	// 		time.Sleep(5 * time.Microsecond)
	// 	}
	// }

	chunk := types.FloatSlice2Uint(data)

	switch channel {
	case 0:
		this.Memory.BlockWrite(this.MemIndex, this.Memory.MEMBASE(this.GetMemIndex())+memory.OCTALYZER_SPEAKER_BUFFER, chunk)

		this.Memory.WriteGlobal(this.MemIndex, this.Memory.MEMBASE(this.GetMemIndex())+memory.OCTALYZER_SPEAKER_SAMPLECOUNT, uint64(l))
		this.Memory.WriteGlobal(this.MemIndex, this.Memory.MEMBASE(this.GetMemIndex())+memory.OCTALYZER_SPEAKER_SAMPLERATE, uint64(rate))
		this.Memory.WriteGlobal(this.MemIndex, this.Memory.MEMBASE(this.GetMemIndex())+memory.OCTALYZER_SPEAKER_PLAYSTATE, 128) // loading samples to buffer

		// for this.Memory.ReadGlobal(this.Memory.MEMBASE(this.GetMemIndex())+memory.OCTALYZER_SPEAKER_PLAYSTATE) != 0 {
		// 	time.Sleep(5 * time.Microsecond)
		// }
	case 1:
		this.Memory.BlockWrite(this.MemIndex, this.Memory.MEMBASE(this.GetMemIndex())+memory.OCTALYZER_CASSETTE_BUFFER, chunk)

		this.Memory.WriteGlobal(this.MemIndex, this.Memory.MEMBASE(this.GetMemIndex())+memory.OCTALYZER_CASSETTE_SAMPLECOUNT, uint64(l))
		this.Memory.WriteGlobal(this.MemIndex, this.Memory.MEMBASE(this.GetMemIndex())+memory.OCTALYZER_CASSETTE_SAMPLERATE, uint64(rate))
		this.Memory.WriteGlobal(this.MemIndex, this.Memory.MEMBASE(this.GetMemIndex())+memory.OCTALYZER_CASSETTE_PLAYSTATE, 128) // loading samples to buffer

		// for this.Memory.ReadGlobal(this.Memory.MEMBASE(this.GetMemIndex())+memory.OCTALYZER_CASSETTE_PLAYSTATE) != 0 {
		// 	time.Sleep(5 * time.Microsecond)
		// }
	}

}

func (this *Interpreter) MarkLayersDirty(list []string) {

	for _, l := range list {
		layer := apple2helpers.GETGFX(this, l)
		if layer != nil {
			layer.SetDirty(true)
		}
	}

}

func (this *Interpreter) PassWaveBufferNB(channel int, data []float32, loop bool, rate int) {

	chunk := types.FloatSlice2Uint(data)

	//log2.Printf("direct send audio channel %d", channel)

	this.Memory.DirectSendAudio(this.MemIndex, channel, chunk, rate)

}

func (this *Interpreter) PassRestBufferNB(data string) {

	this.Memory.DirectSendRestalgiaCmd(this.MemIndex, data)

}

func (this *Interpreter) PassMusicBufferNB(data []float32, loop bool, rate int, channel int) {

	chunk := types.FloatSlice2Uint(data)

	this.Memory.DirectSendDigitalMusic(this.MemIndex, chunk, rate, channel)

}

func (this *Interpreter) IsSharing() bool {

	return this.Memory.IsSlotShared(this.MemIndex)

}

func (this *Interpreter) PassWaveBufferCompressed(channel int, data []uint64, loop bool, rate int, directParallel bool) {

	this.Memory.RecordSendAudioPacked(this.MemIndex, channel, len(data), data, rate, true)

}

func (this *Interpreter) PassWaveBufferUncompressed(channel int, data []uint64, loop bool, rate int, directParallel bool) {

	this.Memory.RecordSendAudioPacked(this.MemIndex, channel, len(data), data, rate, false)

}

func (this *Interpreter) ClearToBottom() {
	this.Dialect.ClearToBottom(this)
}

func (this *Interpreter) SetCursorX(x int) {
	this.Dialect.SetCursorX(this, x)
}

func (this *Interpreter) SetCursorY(y int) {
	this.Dialect.SetCursorY(this, y)
}

func (this *Interpreter) DoPrompt() {

	txt := apple2helpers.TEXT(this)
	ofg := txt.FGColor
	if this.UsePromptColor {
		txt.FGColor = this.PromptColor
	}

	txt.PutStr(this.Prompt)
	this.SaveCPOS()

	txt.FGColor = ofg
}

func (this *Interpreter) GetColumns() int {
	return this.Dialect.GetColumns(this)
}

func (this *Interpreter) GetRows() int {
	return this.Dialect.GetRows(this)
}

func (this *Interpreter) Repos() {
	this.Dialect.Repos(this)
}

func (this *Interpreter) GetCursorX() int {
	return this.Dialect.GetCursorX(this)
}

func (this *Interpreter) GetCursorY() int {
	return this.Dialect.GetCursorY(this)
}

func (this *Interpreter) SaveCPOS() {
	this.Px, this.Py = this.Dialect.GetRealCursorPos(this)
}

func (this *Interpreter) Post() {
	if this.Producer != nil {
		this.Producer.Post(this.MemIndex)
	}
}

func (this *Interpreter) HUDLayerSetPos(name string, x, y, z float64) bool {

	this.ReadLayersFromMemory()
	ls, exists := this.GetHUDLayerByID(name)
	if exists {
		ls.SetPos(x, y, z)
		this.WriteLayersToMemory()
	}

	return exists

}

func (this *Interpreter) GFXLayerSetPos(name string, x, y, z float64) bool {

	this.ReadLayersFromMemory()
	ls, exists := this.GetGFXLayerByID(name)
	if exists {
		ls.SetPos(x, y, z)
		this.WriteLayersToMemory()
	}

	return exists

}

func (this *Interpreter) HUDLayerMovePos(name string, x, y, z float64) bool {

	this.ReadLayersFromMemory()
	ls, exists := this.GetHUDLayerByID(name)
	if exists {
		ox, oy, oz := ls.GetPos()
		ls.SetPos(ox+x, oy+y, oz+z)
	}

	return exists

}

func (this *Interpreter) GFXLayerMovePos(name string, x, y, z float64) bool {

	this.ReadLayersFromMemory()
	ls, exists := this.GetGFXLayerByID(name)
	if exists {
		ox, oy, oz := ls.GetPos()
		ls.SetPos(ox+x, oy+y, oz+z)
	}

	return exists

}

func (this *Interpreter) PositionAllLayers(x, y, z float64) {
	this.ReadLayersFromMemory()
	for _, ls := range this.HUDLayers {
		if ls == nil {
			continue
		}
		ls.SetPos(x, y, z)
	}
	for _, ls := range this.GFXLayers {
		if ls == nil {
			continue
		}
		ls.SetPos(x, y, z)
	}
	this.WriteLayersToMemory()
}

func (this *Interpreter) MoveAllLayers(x, y, z float64) {
	this.ReadLayersFromMemory()
	for _, ls := range this.HUDLayers {
		if ls == nil {
			continue
		}
		ox, oy, oz := ls.GetPos()
		ls.SetPos(ox+x, oy+y, oz+z)
	}
	for _, ls := range this.GFXLayers {
		if ls == nil {
			continue
		}
		ox, oy, oz := ls.GetPos()
		ls.SetPos(ox+x, oy+y, oz+z)
	}
	this.WriteLayersToMemory()
}
