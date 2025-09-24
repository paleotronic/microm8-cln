package interfaces

import (
	"bytes"
	"io"
	"time"

	s8webclient "paleotronic.com/api"
	"paleotronic.com/core/hardware/servicebus"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/syncmanager"
	"paleotronic.com/core/types"
	"paleotronic.com/ducktape"
	"paleotronic.com/filerecord"
	"paleotronic.com/runestring"
)

const REGCOUNT = 8

type AnalyzerFunc func(msg *ducktape.DuckTapeBundle) bool

type CommandState struct {
	CurrentCommand Commander
	Data           map[string]interface{}
	Params         types.TokenList
	Scope          *types.Algorithm
	PC             types.CodeRef
	SleepCounter   int
	// Step
	Step int
	// Registers for temp storage
	I [REGCOUNT]int
	F [REGCOUNT]float32
	R [REGCOUNT]rune
	S [REGCOUNT]runestring.RuneString
	B [REGCOUNT]bool
	T [REGCOUNT]types.TokenList
	L [REGCOUNT]types.TokenListArray
	// state to move to after
	PostSleepState types.EntitySubState
}

func NewCommandState(c Commander) *CommandState {
	this := &CommandState{
		CurrentCommand: c,
		Data:           make(map[string]interface{}),
	}
	return this
}

type Interpretable interface {
	ReplayVideo() bool
	SetPromptColor(i int)
	SetUsePromptColor(b bool)
	MoveAllLayers(x, y, z float64)
	HUDLayerMovePos(name string, x, y, z float64) bool
	GFXLayerMovePos(name string, x, y, z float64) bool
	VM() VM
	Bind(vm VM)
	IsRecordingDiscVideo() bool
	ForwardVideo1x() bool
	HandleServiceBusRequest(r *servicebus.ServiceBusRequest) (*servicebus.ServiceBusResponse, bool)
	InjectServiceBusRequest(r *servicebus.ServiceBusRequest)
	HandleServiceBusInjection(handler func(r *servicebus.ServiceBusRequest) (*servicebus.ServiceBusResponse, bool))
	ServiceBusProcessPending()
	RestorePrevState()
	ResetView()
	IsNeedingInit() bool
	DoCatalog()
	SetStatusText(s string)
	GetStatusText() string
	SetVideoStatusText(s string)
	GetVideoStatusText() string
	PassWaveBufferUncompressed(channel int, data []uint64, loop bool, rate int, directParallel bool)
	PassWaveBufferCompressed(channel int, data []uint64, loop bool, rate int, directSent bool)
	WaitFrom(now time.Time, ms int64)
	IsWaitingForWorld() bool
	WaitDeduct(ms int64)
	SetDisabled(b bool)
	IsDisabled() bool
	StopTheWorld()
	ResumeTheWorld()
	WaitForWorld() bool
	PassRestBufferNB(data string)
	SliceRecording(fn string, newfn string, start int, end int) error
	PlayRecording(fn string)
	StartRecording(fn string, debugMode bool)
	TransferRemoteControl(target string) int
	SetControlProfile(target string, profile string) int
	StopRecording()
	StopRecordingHard()
	SetDiskImage(s string)
	GetDiskImage() string
	PeripheralReset()
	RecordToggle(debugMode bool)
	GetDataNames() []string
	GetPublicDataNames() []string
	GetDataKeys() []string
	BootCheck()
	GetPasteBuffer() runestring.RuneString
	IndexOfLoop(forvarname string) int
	Wait(ms int64)
	NeedsClock() bool
	PlayMusic(p, f string, leadin int, fadein int) error
	StopMusic()
	SendRemoteCommand(command string) int
	SendRemoteText(command string) int
	GetVar(n string) *types.Variable
	PurgeOwnedVariables()
	GetData(n string) *types.Token
	GetDataRef() *types.CodeRef
	Bootstrap(n string, silent bool) error
	ExistsVarLower(n string) bool
	Call(target types.CodeRef, c *types.Algorithm, s types.EntityState, iso bool, prefix string, stackstate types.TokenList, dia Dialecter)
	RunStatementFunction()
	Clear()
	CallUserFunction(funcname string, inputs types.TokenList) error
	SplitOnTokenWithBrackets(tokens types.TokenList, tok types.Token) types.TokenListArray
	VariableTypeFromString(s string) types.VariableType
	Halt() error
	Run(keepvars bool)
	IsRunning() bool
	IsBreak() bool
	RunStatementDirect()
	SplitOnToken(tokens types.TokenList, tok types.Token) types.TokenListArray
	SplitOnTokenType(tokens types.TokenList, tok types.TokenType) types.TokenListArray
	GetTokenAtCodeRef(current types.CodeRef) types.Token
	SeekForwards(current types.CodeRef, eType types.TokenType, eValue string, pairType types.TokenType, pairValue string) types.CodeRef
	EndProgram()
	TBCheck(tb *types.TextBuffer)
	TBChange(tb *types.TextBuffer)
	Pop(trimloops bool) error
	SetData(n string, v types.Token, local bool)
	HandleError() bool
	Zero(coldstart bool)
	CreateVarLower(n string, v types.Variable)
	Return(trimloops bool) error
	Deliver(im types.InfernalMessage)
	ExistsVar(n string) bool
	IndicesFromTokens(tl types.TokenList, ob string, cb string) ([]int, error)
	InFunction() bool
	SeekBackwards(current types.CodeRef, eType types.TokenType, eValue string, pairType types.TokenType, pairValue string) types.CodeRef
	IndexOfLoopFromBase(forvarname string) int
	ExistsData(n string) bool
	Send(t string, content string)
	GetVarLower(n string) *types.Variable
	IsRunningDirect() bool
	IsCodeRefValid(current types.CodeRef) bool
	IsEmpty() bool
	TokenListAsString(tokens types.TokenList) string
	Jump(target types.CodeRef)
	StackDump()
	LoopBase() int
	Parse(s string)
	CreateVar(n string, v types.Variable)
	SplitOnTokenStartsWith(tokens types.TokenList, tok []types.TokenType) types.TokenListArray
	Reset()
	SetDialect(d Dialecter, preserve bool, silent bool)
	ParseTokensForResult(tokens types.TokenList) types.Token
	IsStopped() bool
	IsBlocked() bool
	GetPrevStatement(current types.CodeRef) types.CodeRef
	GetNextToken(current *types.CodeRef) types.Token
	Continue()
	IsZ80Executing() bool
	DoCyclesZ80() int
	HaltZ80(r int)
	DoCycles6502() int
	Halt6502(r int)
	IsWaiting() bool
	NextTokenInstance(current *types.CodeRef, ttype types.TokenType, tcontent string) bool
	IsPaused() bool
	RunStatement()
	GetNextStatement(current types.CodeRef) types.CodeRef
	GetPrevToken(current *types.CodeRef) types.Token
	GetLocal() types.VarManager
	GetVarPrefix() string
	SetVarPrefix(s string)
	GetState() types.EntityState
	SetState(s types.EntityState)
	GetSubState() types.EntitySubState
	SetSubState(s types.EntitySubState)
	GetStack() *CallStack
	SetStack(s *CallStack)
	GetTokenStack() *types.TokenList
	GetPC() *types.CodeRef
	GetWaitUntil() time.Time
	SetWaitUntil(i time.Time)
	GetCode() *types.Algorithm
	SetCode(a *types.Algorithm)
	Start6502(addr int, x, y, a, p, sp int)
	Is6502Executing() bool
	GetDirectAlgorithm() *types.Algorithm
	SetDirectAlgorithm(a *types.Algorithm)
	GetName() string
	SetName(s string)
	SetLPC(c *types.CodeRef)
	GetLPC() *types.CodeRef
	SetProducer(v Producable)
	GetProducer() Producable
	GetDialect() Dialecter
	GetMemory(addr int) uint64
	GetLoopStack() *types.LoopStack
	SetLoopStep(f float64)
	SetLoopVariable(s string)
	GetLoopVariable() string
	GetLoopStep() float64
	GetLoopBase() int
	GetLoopStates() types.LoopStateMap
	SetOuterVars(v bool)
	SetPC(*types.CodeRef)
	GetErrorTrap() *types.CodeRef
	GetBreakpoint() *types.CodeRef
	GetFirstString() string
	SetLocal(types.VarManager)
	SetFirstString(s string)
	WaitAdd(d time.Duration)
	SetNeedsPrompt(v bool)
	SystemMessage(text string)
	ScreenReset()
	CheckProfile(force bool)
	GetMultiArgFunc() MafMap
	GetWorkDir() string
	SetWorkDir(s string)
	SetMemory(addr int, v uint64)
	SetMemorySilent(addr int, v uint64)
	RegisterWithParent()
	DeregisterWithParent()
	GetChild() Interpretable
	SetChild(a Interpretable)
	NewChild(name string) Interpretable
	SetParent(a Interpretable)
	GetParent() Interpretable
	LastHistory() runestring.RuneString
	AddToHistory(cmd runestring.RuneString)
	BackHistory(s runestring.RuneString) runestring.RuneString
	ForwardHistory(s runestring.RuneString) runestring.RuneString
	GetStartTime() time.Time
	GetServer() *s8webclient.Client
	GetVSync() *syncmanager.VariableSyncher
	SetMultiArgFunc(n string, maf MultiArgumentFunction)
	SetParams(p *types.TokenList)
	GetParams() *types.TokenList
	GetProgramDir() string
	SetProgramDir(s string)
	NewChildWithParamsAndTask(name string, dialectname string, params *types.TokenList, task string) Interpretable
	IsDebug() bool
	SetDebug(v bool)
	IsSilent() bool
	IsRemote() bool
	SetSilent(v bool)
	Log(component string, message string)
	GetHUDLayerByID(name string) (*types.LayerSpecMapped, bool)
	GetGFXLayerByID(name string) (*types.LayerSpecMapped, bool)
	ReadLayersFromMemory()
	WriteLayersToMemory()
	GetMemoryMap() *memory.MemoryMap
	SplitOnTokenTypeList(tokens types.TokenList, tok []types.TokenType) types.TokenListArray
	GetMemIndex() int
	SetPrompt(s string)
	GetPrompt() string
	GetUUID() uint64
	SetUUID(u uint64)
	SetMusicPaused(paused bool)
	SetTabWidth(s int)
	GetTabWidth() int
	Interactive()
	GetBuffer() runestring.RuneString
	SetBuffer(r runestring.RuneString)
	GetLastCommand() runestring.RuneString
	SetLastCommand(r runestring.RuneString)
	PutStr(s string)
	Backspace()
	RealPut(ch rune)
	GetFeedBuffer() string
	SetFeedBuffer(s string)
	GetDisplayPage() string
	GetCurrentPage() string
	SetDisplayPage(s string)
	SetCurrentPage(s string)
	SetBreakable(v bool)
	IsBreakable() bool
	GetDosBuffer() string
	SetDosBuffer(s string)
	SetDosCommand(v bool)
	IsDosCommand() bool
	SetNextByteColor(v bool)
	IsNextByteColor() bool
	GetSpeed() int
	SetSpeed(s int)
	GetOutChannel() string
	SetOutChannel(s string)
	GetInChannel() string
	SetInChannel(s string)
	GetCharacterCapture() string
	SetCharacterCapture(s string)
	GetLastChar() rune
	SetLastChar(s rune)
	SetSuppressFormat(b bool)
	IsSuppressFormat() bool
	PassWaveBuffer(channel int, data []float32, loop bool, rate int)
	PassWaveBufferNB(channel int, data []float32, loop bool, rate int)
	GetCursorX() int
	GetCursorY() int
	MarkLayersDirty(list []string)
	GetColumns() int
	SaveCPOS()
	Freeze(filename string) error
	Thaw(filename string) error
	WaitForLayers()
	GetGFXLayerState() []bool
	SetGFXLayerState(v []bool)
	ShouldSaveAndRestoreText() bool
	SetSaveAndRestoreText(v bool)
	ConnectRemote(ip, port string, slotid int)
	//ProcessRemote()
	FreezeBytes() ([]byte, error)
	ThawBytes(data []byte) error
	FreezeStreamLayers(f io.Writer)
	Post()
	EndRemote()
	SetPasteBuffer(r runestring.RuneString)
	PositionAllLayers(x, y, z float64)
	GFXLayerSetPos(name string, x, y, z float64) bool
	HUDLayerSetPos(name string, x, y, z float64) bool
	GetSpec() string
	//GetRemIntControl() *client.DuckTapeClient
	//SetRemIntControl(d *client.DuckTapeClient)
	GetRemIntIndex() int
	SetRemIntIndex(i int)
	//SendRemIntMessage(id string, payload []byte, binary bool)
	//SendChatMessage(message string)
	//ConnectControl(host, port string, slotid int) error
	//GetChatMessages(maxcount int) ([]string, []string, error)
	//GetControlState(control string) (string, error)
	SetClientSync(b bool)
	SetFileRecord(fr filerecord.FileRecord)
	GetFileRecord() filerecord.FileRecord
	PreOptimizer()
	LoadSpec(specfile string) error
	IsIgnoreSpecial() bool
	SetIgnoreSpecial(v bool)
	HandleEvent(e types.Event)
	GetVM() types.VarManager
	SetVM(vm types.VarManager)
	ThawStreamLayers(f io.Reader) error
	ReturnFromProc(value *types.Token) error
	BufferCommands(commands *types.TokenList, times int)
	BufferEmpty()
	IsBufferEmpty() bool
	GetCallReturnToken() *types.Token
	SetCallReturnToken(t *types.Token)
	ThawBytesNoPost(data []byte) error
	SetCommandState(cs *CommandState)
	GetCommandState() *CommandState
	SetPragma(name string)
	SetLabel(name string, line int)
	GetLabel(name string) int
	ClearLabels()
	ParseImm(s string)
	AddTrigger(slot int, condition *types.TokenList, line int)
	GetHUDLayerSet() []*types.LayerSpecMapped
	GetGFXLayerSet() []*types.LayerSpecMapped
	SetSpec(s string)
	SetCycleCounter(c Countable)
	GetCycleCounter() []Countable
	ClearCycleCounter(c Countable)
	DeleteCycleCounters()
	IgnoreMyAudio() bool
	SetIgnoreMyAudio(b bool)
	IsPlayingVideo() bool
	IsRecordingVideo() bool
	IncrementRecording(n int)
	BackVideo() bool
	ForwardVideo() bool
	GetPlayer() Playable
	BreakIntoVideo()
	BackstepVideo(ms int)
	PlayBlocks(blocks []*bytes.Buffer, backwards bool, backJumpMS int)
	ResumeRecording(fn string, blocks []*bytes.Buffer, debugMode bool)
	PlayRecordingCustom(fn string, backwards bool)
	GetLiveBlocks() []*bytes.Buffer
	StartRecordingWithBlocks(blocks []*bytes.Buffer, debugMode bool)
	WriteBlocks(blocks []*bytes.Buffer)
	SetVidMode(mode int)
	SetMemMode(mode int)
	PostJumpEvent(from, to int, context string)
	AnalyzeRecording(fn string, analyzeMap map[string]AnalyzerFunc) bool
	IsMicroControl() bool
	SetMicroControl(b bool)
	ClearAudioPorts()
	GetAudioPort(name string) int
	SetAudioPort(name string, port int)
	SetCurrentSubroutine(s string)
	GetCurrentSubroutine() string
	PBPaste()
	SetSemaphore(s string)
	GetSemaphore() string
}

type Playable interface {
	IsBackwards() bool
	Playback() PlayerExitMode
	SetBackwards(b bool)
	Slower()
	Faster()
	Pause()
	IsPaused() bool
	GetTimeShift() float64
	SetTimeShift(f float64)
	AddSeekDelta(ms int)
	IsSeeking() bool
	IsActive() bool
	Jump(syncs int)
	SetNoResume(b bool)
	GetNextNSyncs(count int, current int) []int
	GetLastNSyncs(count int, current int) []int
}

type PlayerExitMode int

const (
	PEM_NONE PlayerExitMode = iota
	PEM_RESUME_CPU
	PEM_RESET_SLOT
)
