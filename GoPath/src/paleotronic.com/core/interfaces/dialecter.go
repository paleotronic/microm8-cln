// dialect.go
package interfaces

import (
	"paleotronic.com/core/types"
	"paleotronic.com/runestring"
)

// We define an interface for anything that can be used as a dialect...
type Dialecter interface {
	GetLastCommand() string
	SetSilentDefines(b bool)
	GetSilentDefines() bool
	BeforeRun(caller Interpretable)
	GetShortName() string
	GetLongName() string
	GetCompletions(ent Interpretable, line runestring.RuneString, index int) (int, *types.TokenList)
	CleanDynaCommands()
	IsUpperOnly() bool
	Init()
	InitVDU(v Interpretable, promptonly bool)
	Tokenize(s runestring.RuneString) *types.TokenList
	ExecuteDirectCommand(tl types.TokenList, ent Interpretable, scope *types.Algorithm, lpc *types.CodeRef) error
	ParseTokensForResult(ent Interpretable, tokens types.TokenList) (*types.Token, error)
	Parse(ent Interpretable, s string) error
	SplitOnToken(tokens types.TokenList, tok types.Token) types.TokenListArray
	GetTitle() string
	SetTitle(s string)
	GetMaxVariableLength() int
	GetCommands() CommandList
	GetFunctions() FunctionList
	HandleException(ent Interpretable, e error)
	GetIPS() int
	GetWatchVars() types.NameList
	GetImpliedAssign() Commander
	GetArrayDimMax() int
	GetDefaultCost() int64
	GetArrayDimDefault() int
	SetTrace(v bool)
	SetThrottle(v float32)
	ProcessDynamicCommand(ent Interpretable, cmd string) error
	GetPlusFunctions() FunctionList
	Renumber(code types.Algorithm, start int, increment int) types.Algorithm
	PutStr(ent Interpretable, s string)
	RealPut(ent Interpretable, ch rune)
	Backspace(ent Interpretable)
	ClearToBottom(ent Interpretable)
	SetCursorX(ent Interpretable, x int)
	SetCursorY(ent Interpretable, y int)
	GetColumns(ent Interpretable) int
	GetRows(ent Interpretable) int
	Repos(ent Interpretable)
	GetCursorX(ent Interpretable) int
	GetCursorY(ent Interpretable) int
	RemoveProc(name string)
	ParseMemoryRepresentation(data []uint64) types.Algorithm
	GetMemoryRepresentation(a *types.Algorithm) []uint64
	AddPlusFunction(ns string, s string, cmd Functioner)
	AddHiddenPlusFunction(ns string, s string, cmd Functioner)
	SkipMemParse() bool
	SetSkipMemParse(v bool)
	PreFreeze(ent Interpretable)
	PostThaw(ent Interpretable)
	InitVarmap(ent Interpretable, vm types.VarManager)
	HasCBreak(ent Interpretable) bool
	UpdateRuntimeState(ent Interpretable)
	ThawVideoConfig(ent Interpretable)
	CheckOptimize(lno int, s string, OCode types.Algorithm)
	HomeLeft(ent Interpretable)
	GetRealCursorPos(ent Interpretable) (int, int)
	GetRealWindow(ent Interpretable) (int, int, int, int, int, int)
	SetRealCursorPos(ent Interpretable, x, y int)
	DefineProc(name string, params []string, code string)
	StartProc(name string, params []string, code string)
	GetDynamicCommands() []string
	GetDynamicFunctions() []string
	GetPublicDynamicCommands() []string
	GetDynamicCommandDef(name string) []string
	GetDynamicFunctionDef(name string) []string
	GetDynaCommand(name string) DynaCoder
	GetDynaFunction(name string) DynaCoder
	Decolon(code types.Algorithm, start int, increment int, iftosub bool) types.Algorithm
	CleanDynaCommandsByName(names []string)
	GetWorkspaceBody(vars bool, filterProc string) []string
	QueueCommand(command string)
	SaveState()
	RestoreState()
	SyntaxValid(s string) error
	GetWorkspace(caller Interpretable) []byte
}
