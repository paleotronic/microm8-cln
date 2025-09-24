package logo

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"paleotronic.com/log"

	"paleotronic.com/core/dialect/parcel"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/core/types/glmath"
)

type TrackingParams struct {
	FollowPosition      bool // camera always centers on turtle
	FollowBehind        bool
	MaintainPerspective bool
	RelativePerspective glmath.Vector3
}

type LogoDriver struct {
	Buffer                []*parcel.TokenMatchResult
	d                     *DialectLogo
	ent                   interfaces.Interpretable
	Procs                 map[string]*LogoProc
	Globals               *LogoVarTable
	Stack                 *LogoStack
	S                     *LogoScope
	ErrScope              *LogoScope
	LastReturn            *types.Token
	State                 LogoState
	PendingProcName       string
	PendingProcArgs       []string
	PendingProcStatements []string
	NoResolveProcEntry    bool
	SuppressConsole       bool
	BreakFunc             func() bool
	InstDelay             time.Duration
	lastExec              time.Time
	savedRunState         *LogoRunState
	hasResumed            bool
	Commands              chan LogoDriverCommand
	Paused                bool
	DisableDefineMsgs     bool
	Tracking              TrackingParams
	CoroutineId           string
	Coroutines            map[string]*LogoCoroutine
	Channels              map[string]chan *types.Token
	crt                   sync.Mutex
	turtle                int
	lastRem               string
}

func NewLogoDriverInherit(parent *LogoDriver) *LogoDriver {
	d := NewDialectLogo()
	ld := d.Driver
	ld.Procs = parent.Procs
	ld.Globals = parent.Globals
	ld.Coroutines = parent.Coroutines
	ld.Channels = parent.Channels
	return ld
}

func NewLogoDriver(d *DialectLogo) *LogoDriver {
	ld := &LogoDriver{
		Buffer:          []*parcel.TokenMatchResult{},
		d:               d,
		Procs:           map[string]*LogoProc{},
		Globals:         NewLogoVarTable(),
		S:               NewLogoScope(nil, 0),
		Stack:           NewLogoStack(),
		SuppressConsole: false,
		InstDelay:       time.Second / 1000,
		Coroutines:      map[string]*LogoCoroutine{},
		turtle:          1,
		Channels:        map[string]chan *types.Token{},
	}
	ld.BreakFunc = ld.HasCBreak
	ld.ClearCommands()
	return ld
}

func (d *LogoDriver) GetTurtle() int {
	return d.turtle
}

func (d *LogoDriver) SetTurtle(i int) {
	d.turtle = i
}

func (d *LogoDriver) OnBeginStream(l *parcel.Lexer) error {
	//
	return nil
}

func (d *LogoDriver) OnEndStream(l *parcel.Lexer) error {
	//
	//return d.ExecBuffer()
	return nil
}

func (d *LogoDriver) OnTokenRecognized(t *parcel.TokenMatchResult) error {
	if t.Ignore {
		return nil
	}
	d.Buffer = append(d.Buffer, t)
	return nil
}

func (d *LogoDriver) OnTokenUnrecognized(line string, pos int, err error) error {
	log.Printf("Error: %v", err)
	return fmt.Errorf("I don't ")
}

func (d *LogoDriver) ParseMultiWithRules(l *parcel.Lexer, s string) (list []*types.TokenList, err error) {
	var tokens []*parcel.TokenMatchResult
	tokens, err = l.Tokenize(s)
	if err != nil {
		return list, err
	}

	pos := 0
	var matches []parcel.RuleMatchResult
	for err == nil && pos < len(tokens) {
		matches = l.RuleMatchesFromTokens(tokens, pos, false)
		if len(matches) == 0 {
			err = fmt.Errorf("i don't understand %s", s)
			return
		}
		var tl *types.TokenList
		tl, err = d.MatchesToTokenList(l, matches[0].Tokens)
		if err != nil {
			return
		}
		list = append(list, tl)
		pos += len(matches[0].Tokens)
	}

	return
}

func (d *LogoDriver) MatchesToTokenList(l *parcel.Lexer, tokens []*parcel.TokenMatchResult) (*types.TokenList, error) {
	var list = types.NewTokenList()
	var lasttoken *types.Token
	for i := 0; i < len(tokens); i++ {
		t := tokens[i]
		switch t.Token.Label {
		case "whitespace":
			if lasttoken != nil {
				lasttoken.WSSuffix += t.Content
			}
		case "blockbeg", "listbeg":
			list.Push(types.NewToken(types.OBRACKET, t.Content))
		case "blockend", "listend":
			list.Push(types.NewToken(types.CBRACKET, t.Content))
		case "block":
			// [ xxx ... ]
			//log.Printf("Recursively processing block: %s", t.Content)
			newtokens, err := d.Parse(l, t.Content[1:len(t.Content)-1])
			if err != nil {
				return list, err
			}
			list.Push(types.NewToken(types.LIST, ""))
			list.RPeek().List = newtokens
		case "list":
			// ( ... )
			//log.Printf("Recursively processing list: %s", t.Content)
			newtokens, err := d.Parse(l, t.Content[1:len(t.Content)-1])
			if err != nil {
				return list, err
			}
			list.Push(types.NewToken(types.COMMANDLIST, ""))
			list.RPeek().List = newtokens
		case "expr":
			// ( ... )
			//log.Printf("Recursively processing expression list: %s", t.Content)
			newtokens, err := d.Parse(l, t.Content[1:len(t.Content)-1])
			if err != nil {
				return list, err
			}
			list.Push(types.NewToken(types.EXPRESSIONLIST, ""))
			list.RPeek().List = newtokens
		case "string":
			list.Push(types.NewToken(types.WORD, t.Content[1:]))
		case "comparison":
			list.Push(types.NewToken(types.COMPARITOR, t.Content))
		case "assign":
			list.Push(types.NewToken(types.ASSIGNMENT, t.Content))
		case "float", "int":
			list.Push(types.NewToken(types.NUMBER, t.Content))
		case "muldiv", "addsub":
			list.Push(types.NewToken(types.OPERATOR, t.Content))
		case "function":
			list.Push(types.NewToken(types.FUNCTION, t.Content))
		case "identifier", "symbols":
			/*if strings.ToLower(t.Content) == "and" || strings.ToLower(t.Content) == "or" || strings.ToLower(t.Content) == "not" {
				list.Push(types.NewToken(types.LOGIC, t.Content))
			} else */if d.HasFunc(t.Content) {
				list.Push(types.NewToken(types.DYNAMICFUNCTION, t.Content))
			} else if d.HasProc(t.Content) {
				list.Push(types.NewToken(types.DYNAMICKEYWORD, t.Content))
			} else if _, ok := d.d.GetCommands()[strings.ToLower(t.Content)]; ok {
				list.Push(types.NewToken(types.KEYWORD, t.Content))
			} else if _, ok := d.d.GetFunctions()[strings.ToLower(t.Content)]; ok {
				list.Push(types.NewToken(types.FUNCTION, t.Content))
			} else {
				d.Printf("warning: Parsed expression has as yet unresolved identifier %s", t.Content)
				list.Push(types.NewToken(types.IDENTIFIER, t.Content))
			}
		case "command":
			list.Push(types.NewToken(types.KEYWORD, t.Content))
		case "varref":
			list.Push(types.NewToken(types.VARIABLE, t.Content))
		case "boolop":
			list.Push(types.NewToken(types.LOGIC, t.Content))
		// case "symbols":
		// 	list.Push(types.NewToken(types.IDENTIFIER, t.Content))
		case "escapedchar":
			list.Push(types.NewToken(types.WORD, t.Content))
		default:
		}
		if list.Size() > 0 {
			lasttoken = list.RPeek()
		}
	}

	return d.CrunchLists(list), nil
}

// Build Tree constructions an abstract syntax tree from the tokens involved
func (d *LogoDriver) Parse(l *parcel.Lexer, s string) (*types.TokenList, error) {

	tokens, err := l.Tokenize(s)
	if err != nil {
		return nil, err
	}

	return d.MatchesToTokenList(l, tokens)
}

func (d *LogoDriver) CombineMulti(items []*types.TokenList) *types.TokenList {
	n := types.NewTokenList()
	for _, tl := range items {
		for _, t := range tl.Content {
			n.Push(t.Copy())
		}
	}
	return n
}

func (d *LogoDriver) SplitOnTokenTypeList(tokens *types.TokenList, tok []types.TokenType) []*types.TokenList {
	result := []*types.TokenList{}
	var bc int

	//d.Printf("split sees %d tokens", tokens.Size())

	var buffer = types.NewTokenList()
	for _, tt := range tokens.Content {
		if tt.Type == types.OBRACKET && tt.Content == "(" {
			if bc > 0 {
				buffer.Push(tt)
			}
			bc++
		} else if tt.Type == types.OBRACKET && tt.Content == ")" {
			bc--
			if bc > 0 {
				buffer.Push(tt)
			}
		} else if tt.IsIn(tok) && bc == 0 {
			if buffer.Size() > 0 {
				result = append(result, buffer)
				buffer = types.NewTokenList()
			}
			buffer.Push(tt)
		} else {
			buffer.Push(tt)
		}
	}
	if buffer.Size() > 0 {
		result = append(result, buffer)
	}

	return result
}

func (d *LogoDriver) SplitMulti(list *types.TokenList) []*types.TokenList {
	if list == nil || list.Size() == 0 {
		return []*types.TokenList{}
	}
	if list.LPeek().Type == types.KEYWORD && strings.ToLower(list.LPeek().Content) == "to" {
		return []*types.TokenList{list.Copy()}
	}
	return d.SplitOnTokenTypeList(list.Copy(), []types.TokenType{types.KEYWORD, types.DYNAMICKEYWORD, types.COMMANDLIST})
}

func (d *LogoDriver) ParseMulti(l *parcel.Lexer, s string) ([]*types.TokenList, error) {
	var stmts []*types.TokenList
	list, err := d.Parse(l, s)
	if err != nil {
		return stmts, err
	}

	stmts = d.SplitMulti(list)

	return stmts, nil
}

func (d *LogoDriver) Resolver(statement *types.TokenList) (*types.TokenList, error) {
	// create a working copy to modify
	work := statement.Copy()

	for i := 0; i < len(work.Content); i++ {
		t := work.Get(i)
		if t.List != nil {
			r, e := d.Resolver(t.List)
			if e != nil {
				return work, e
			}
			t.List = r
		} else if t.Type == types.VARIABLE {
			v := d.ent.GetData(strings.Trim(t.Content, ":"))
			if v == nil {
				return work, fmt.Errorf("No such variable %s", strings.Trim(t.Content, ":"))
			}
			t.Type = v.Type
			t.Content = v.Content
			if v.List != nil {
				t.List = v.List.Copy()
			}
		}
	}

	return work, nil
}
