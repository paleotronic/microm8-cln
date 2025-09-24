package dialect

import (
	"errors" //"paleotronic.com/fmt"
	"math"
	"sort"
	"strconv"
	"strings" //	"time"

	"paleotronic.com/core/exception"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/log"
	"paleotronic.com/runestring"
	"paleotronic.com/utils"
)

// Used to assign value
const (
	TID_COMMAND_BASE  = 65536
	TID_FUNCTION_BASE = TID_COMMAND_BASE + 1024
	TID_LOGICAL_BASE  = TID_FUNCTION_BASE + 1024
)

type Dialect struct {
	interfaces.Dialecter
	CurrentNameSpace    string
	WatchVars           types.NameList
	VarPrefixes         string
	DynaFunctions       DynaMap
	DynaCommands        DynaMap
	ArrayDimDefault     int
	MaxVariableLength   int
	IPS                 int
	Trace               bool
	UserCommands        DynaMap
	ArrayDimMax         int
	ImpliedAssign       interfaces.Commander
	Functions           interfaces.FunctionList
	PlusFunctions       interfaces.NameSpaceFunctionList
	ShadowPlusFunctions interfaces.FunctionList
	Title               string
	VarSuffixes         string
	ReverseCase         bool
	UpperOnly           bool
	Throttle            float32
	Commands            interfaces.CommandList
	DefaultCost         int64
	PlusHandler         interfaces.Commander
	TokenMapping        map[string]uint64
	MappingToken        map[uint64]string
	Logicals            map[string]int
	fSkipMemParse       bool
	ShortName, LongName string
	SilentDefines       bool
}

func (this *Dialect) GetWorkspace(caller interfaces.Interpretable) []byte {

	b := caller.GetCode()

	l := b.GetLowIndex()
	h := b.GetHighIndex()

	f := make([]string, 0)

	//Str(h, s);
	s := utils.IntToStr(h)
	w := len(s) + 1
	if w < 4 {
		w = 4
	}

	if l < 0 {
		//close(f);
		return []byte{}
	}

	s = ""
	/* got code */
	for l != -1 {
		/* now formatted tokens */
		ln, _ := caller.GetCode().Get(l)
		s = utils.IntToStr(l) + "  "
		z := 0
		for _, stmt := range ln {

			ft := caller.TokenListAsString(stmt.TokenList)
			if z > 0 {
				s = s + " : "
			}

			s = s + ft

			z++
		}

		f = append(f, s)
		//f.Add(s);

		/* next line */
		b := caller.GetCode()
		l = b.NextAfter(l)
	}

	//close(f);

	//FileHan//dle fh = Gdx.Files.External(filename)
	//fh.WriteString(s, false)

	str := ""
	for _, ss := range f {
		if str != "" {
			str = str + "\r\n"
		}
		str = str + utils.Escape(ss)
	}

	return []byte(str)

}

func (this *Dialect) SetSilentDefines(b bool) {
	this.SilentDefines = b
}

func (this *Dialect) GetSilentDefines() bool {
	return this.SilentDefines
}

func (this *Dialect) GetShortName() string {
	return this.ShortName
}

func (this *Dialect) GetLongName() string {
	return this.LongName
}

func (this *Dialect) BeforeRun(caller interfaces.Interpretable) {
	// stub
}

func (this *Dialect) CheckOptimize(lno int, s string, OCode types.Algorithm) {
	// stub does nothing
}

func (this *Dialect) InitVarmap(ent interfaces.Interpretable, vm types.VarManager) {
	// does runtime dialect specific config of variable space.
	// override this in a subclass
}

func (this *Dialect) PreFreeze(ent interfaces.Interpretable) {
	// pre-freeze housekeeping
}

func (this *Dialect) PostThaw(ent interfaces.Interpretable) {
	// post-thaw housekeeping
}

func (this *Dialect) IsUpperOnly() bool {
	return this.UpperOnly
}

func (this *Dialect) SkipMemParse() bool {
	return this.fSkipMemParse
}

func (this *Dialect) SetSkipMemParse(v bool) {
	this.fSkipMemParse = v
}

func (this *Dialect) GenerateNumericTokens() {

	var tid uint64
	var names []string

	this.TokenMapping = make(map[string]uint64)
	this.MappingToken = make(map[uint64]string)

	// Commmands
	for n, _ := range this.Commands {
		names = append(names, n)
	}
	sort.Strings(names)

	for i, n := range names {
		tid = uint64(TID_COMMAND_BASE + i)
		this.TokenMapping[n] = tid
		this.MappingToken[tid] = n
	}

	// Functions
	names = make([]string, 0)
	for n, _ := range this.Functions {
		names = append(names, n)
	}
	sort.Strings(names)

	for i, n := range names {
		tid = uint64(TID_FUNCTION_BASE + i)
		this.TokenMapping[n] = tid
		this.MappingToken[tid] = n
	}

	// Logicals
	for n, _ := range this.Logicals {
		names = append(names, n)
	}
	sort.Strings(names)

	for i, n := range names {
		tid = uint64(TID_LOGICAL_BASE + i)
		this.TokenMapping[n] = tid
		this.MappingToken[tid] = n
	}

}

func (this *Dialect) SetThrottle(v float32) {
	if v != 0 {
		this.Throttle = v
	}
}

func (this *Dialect) GetFunctions() interfaces.FunctionList {
	return this.Functions
}

func (this *Dialect) GetPlusFunctions() interfaces.FunctionList {
	return this.ShadowPlusFunctions
}

func (this *Dialect) IsPlusVariableName(instr string) bool {

	/* vars */
	var result bool
	var ch rune
	var i int

	result = false

	if len(instr) == 0 {
		return result
	}

	//System.out.println("IsVariableName(",instr,")");

	if (!this.IsAlpha(rune(instr[0]))) && (rune(instr[0]) != '_') {
		return result
	}

	//System.out.println("- First is alpha");

	for i = 1; i <= len(instr)-2; i++ {
		ch = rune(instr[i])
		if !(this.IsDigit(ch) || this.IsAlpha(ch)) {
			return result
		}
	}

	//System.out.println("- Then alpha || numbers");

	ch = rune(instr[len(instr)-1])
	if ch == '@' {
		//System.out.println("- Last is digit, Alpha || VarSuffix");
		result = true
	}

	/* enforce non void return */
	return result

}

func (this *Dialect) SetTrace(b bool) {
	this.Trace = b
}

func (this *Dialect) GetTrace() bool {
	return this.Trace
}

func (this *Dialect) GetIPS() int {
	return this.IPS
}

func (this *Dialect) GetMaxVariableLength() int {
	return this.MaxVariableLength
}

func (this *Dialect) GetTitle() string {
	return this.Title
}

func (this *Dialect) SetTitle(s string) {
	this.Title = s
}

func (this *Dialect) ProcessDynamicCommand(ent interfaces.Interpretable, cmd string) error {
	return nil
}

func (this *Dialect) IsDigit(ch rune) bool {
	return (ch >= '0') && (ch <= '9')
}

func (this *Dialect) GetDefaultCost() int64 {
	return this.DefaultCost
}

func (this *Dialect) Tokenize(s runestring.RuneString) *types.TokenList {

	/* vars */
	var result *types.TokenList
	var inq bool
	var inqq bool
	var cont bool
	var idx int
	var chunk string
	var ch rune
	var tt *types.Token

	////fmt.Println("Dialect.Tokenize()")

	result = types.NewTokenList()

	inq = false
	inqq = false
	idx = 0
	chunk = ""
	cont = true

	for idx < len(s.Runes) {
		ch = s.Runes[idx]

		//System.Out.Println("Tokenizer sees ["+ch+"]");

		if this.IsWS(ch) && (inq || inqq) {
			chunk = chunk + string(ch)
		} else if this.IsVarSuffix(ch, this.VarSuffixes) && (!(inq || inqq)) {
			chunk = chunk + string(ch)
			if len(chunk) > 0 {
				cont = this.Evaluate(chunk, result)
				chunk = ""
			}
		} else if this.IsBreakingCharacter(ch, this.VarSuffixes) && (!(inq || inqq)) {

			// System.Out.Println("====== breaking char "+ch);

			/* special handling for x.Yyyye+/-xx notation */
			if ((ch == '+') || (ch == '-')) &&
				((len(chunk) >= 2) && (chunk[len(chunk)-1] == 'e') && (this.IsDigit(rune(chunk[0]))) && (this.IsDigit(rune(chunk[len(chunk)-1])))) {
				chunk = chunk + string(ch)
			} else {
				if len(chunk) > 0 {
					cont = this.Evaluate(chunk, result)
					chunk = ""
				}

				if !this.IsWS(ch) {
					chunk = chunk + string(ch)
					cont = this.Evaluate(chunk, result)
					chunk = ""
				}
			}

		} else if this.IsQ(ch) && (!inqq) {
			if (len(chunk) > 0) && (!inq) {
				this.Evaluate(chunk, result)
				chunk = ""
			}
			inq = !inq
			chunk = chunk + string(ch)
		} else if this.IsQQ(ch) && (!inq) {
			if (len(chunk) > 0) && (!inqq) {
				this.Evaluate(chunk, result)
				chunk = "\""
			} else if (len(chunk) > 0) && (inqq) {
				chunk = chunk + "\""
				this.Evaluate(chunk, result)
				chunk = ""
			} else {
				chunk = chunk + string(ch)
			}
			inqq = !inqq
		} else {
			chunk = chunk + string(ch)

			/* break keywords out early */
			if _, conkey := this.Commands[strings.ToLower(chunk)]; conkey == true {
				if (strings.ToLower(chunk) != "go") && (strings.ToLower(chunk) != "to") && (strings.ToLower(chunk) != "on") && (strings.ToLower(chunk) != "hgr") && (strings.ToLower(chunk) != "at") {
					cont = this.Evaluate(chunk, result)
					chunk = ""
				}
			}

		}

		idx++

		if !cont {
			chunk = ""

			for (idx < len(s.Runes)) && ((inqq || (s.Runes[idx] != ':')) || (strings.ToLower(result.RPeek().Content) == "rem")) {
				chunk = chunk + string(s.Runes[idx])
				if s.Runes[idx] == '"' {
					inqq = !inqq
				}
				idx++
			}

			tt = types.NewToken(types.UNSTRING, chunk)
			chunk = ""
			result.Push(tt)
		}

	} /*while*/

	//System.Out.Println("chunk == ", chunk;

	if len(chunk) > 0 {
		if inqq {
			chunk = chunk + "\""
		}
		this.Evaluate(chunk, result)
		chunk = ""
	}

	/* enforce non void return */
	return result

}

func (this *Dialect) IsFloat(instr string) bool {

	/* vars */
	_, err := strconv.ParseFloat(instr, 32)

	return (err == nil)

}

func (this *Dialect) IsBoolean(instr string) bool {

	/* vars */
	_, err := strconv.ParseBool(instr)

	return (err == nil)

}

// NTokenize tokenize a group of tokens to uints
func (this *Dialect) NTokenize(tl types.TokenList) []uint64 {

	var values []uint64

	//var lasttok *types.Token

	for _, t := range tl.Content {
		switch t.Type {
		case types.LOGIC:
			values = append(values, this.TokenMapping[strings.ToLower(t.Content)])
		case types.KEYWORD:
			values = append(values, this.TokenMapping[strings.ToLower(t.Content)])
		case types.FUNCTION:
			values = append(values, this.TokenMapping[strings.ToLower(t.Content)])
		case types.STRING:
			values = append(values, 34)
			for _, ch := range t.Content {
				values = append(values, uint64(ch))
			}
			values = append(values, 34)
		default:
			//if lasttok != nil && lasttok.Type != types.KEYWORD && lasttok.Type != types.FUNCTION {
			//	values = append(values, 32)
			//}
			for _, ch := range t.Content {
				values = append(values, uint64(ch))
			}
		}
		//lasttok = t
	}

	return values

}

func (this *Dialect) GetMemoryRepresentation(a *types.Algorithm) []uint64 {
	data := make([]uint64, 0)

	// iterate through code
	l := a.GetLowIndex()
	h := a.GetHighIndex()

	for l <= h && l != -1 {
		ln, _ := a.Get(l)

		buffer := make([]uint64, 2) // allocate 2 slots
		buffer[0] = uint64(l)       // line number
		for i, s := range ln {
			tl := *s.SubList(0, s.Size())
			encoded := this.NTokenize(tl)
			buffer = append(buffer, encoded...)
			if i < len(ln)-1 {
				buffer = append(buffer, uint64(':'))
			}
		}
		buffer = append(buffer, 0)
		buffer[1] = uint64(len(data) + len(buffer))
		data = append(data, buffer...)

		// increment
		l = a.NextAfter(l)
	}

	return data
}

func (this *Dialect) HasCBreak(ent interfaces.Interpretable) bool {
	//~ b := (ent.GetMemory(49152)&0xffff == 131)
	//~ if b {
	//~ ent.SetMemory(49168, 0)
	//~ }

	b := ent.GetMemoryMap().KeyBufferHasBreak(ent.GetMemIndex())
	c := ent.GetMemoryMap().KeyBufferHasSpecialBreak(ent.GetMemIndex())

	return b || c
}

func (this *Dialect) ParseMemoryRepresentation(data []uint64) types.Algorithm {

	var lno int
	var pos, nextpos int
	var ll types.Line
	var st types.Statement

	if len(data) < 3 {
		return *types.NewAlgorithm()
	}

	a := *types.NewAlgorithm()

	if len(data) < 3 {
		return a
	}

	for pos < len(data) {
		lno = int(data[pos])
		nextpos = int(data[pos+1])
		buffer := data[pos+2 : nextpos-1]

		tl := this.UnNTokenize(buffer)

		cmdlist := this.SplitOnToken(*tl, *types.NewToken(types.SEPARATOR, ":"))

		////fmt.Println(len(cmdlist))

		ll = types.NewLine()

		for _, cl := range cmdlist {
			if cl.Size() > 0 {
				st = types.NewStatement()
				for _, tok1 := range cl.Content {
					st.Push(tok1)
				}
				ll = append(ll, st)
			} else {
				st = types.NewStatement()
				st.Push(types.NewToken(types.KEYWORD, "REM"))
				ll = append(ll, st)
			}
		}

		a.Put(lno, ll)

		pos = nextpos
	}

	return a

}

func (this *Dialect) UnNTokenize(values []uint64) *types.TokenList {

	s := ""
	var lastcode uint64 = 0

	var skipspace bool

	for _, v := range values {
		if v < TID_COMMAND_BASE {
			if lastcode >= TID_COMMAND_BASE && !skipspace {
				s = s + " "
			}
			skipspace = false
			if v > 0 {
				s = s + string(rune(v))
			}
		} else {
			s = s + " " + this.MappingToken[v]
			if this.MappingToken[v] == "rem" || this.MappingToken[v] == "data" {
				skipspace = true
			}
		}
		lastcode = v
	}

	tl := this.Tokenize(runestring.Cast(s))

	return tl
}

func NewDialect() *Dialect {
	this := &Dialect{}

	this.Commands = interfaces.NewCommandList()
	this.Functions = interfaces.NewFunctionList()
	this.PlusFunctions = interfaces.NewNameSpaceFunctionList()
	this.ShadowPlusFunctions = interfaces.NewFunctionList()
	this.DynaCommands = NewDynaMap()
	this.DynaFunctions = NewDynaMap()
	this.UserCommands = NewDynaMap()
	this.Logicals = make(map[string]int)
	this.ArrayDimDefault = 10
	this.ArrayDimMax = 65535
	this.IPS = 0
	this.Trace = false
	this.WatchVars = types.NewNameList()
	this.Init()

	return this
}

func (this *Dialect) GetArrayDimDefault() int {
	return this.ArrayDimDefault
}

func (this *Dialect) GetArrayDimMax() int {
	return this.ArrayDimMax
}

func (this *Dialect) GetWatchVars() types.NameList {
	return this.WatchVars
}

func (this *Dialect) GetImpliedAssign() interfaces.Commander {
	return this.ImpliedAssign
}

func (this *Dialect) Init() {
	// override this in your dialect
}

func (this *Dialect) AddCommand(s string, cmd interfaces.Commander) {

	//this.Commands.Put(strings.ToLower(s), cmd)
	this.Commands[strings.ToLower(s)] = cmd

}

func (this *Dialect) IsDynaFunction(instr string) bool {

	/* vars */
	var result bool

	_, result = this.DynaFunctions[strings.ToLower(instr)]

	/* enforce non void return */
	return result

}

func (this *Dialect) IsAlpha(ch rune) bool {
	return ((ch >= 'a') && (ch <= 'z')) || ((ch >= 'A') && (ch <= 'Z'))
}

func (this *Dialect) IsQQ(ch rune) bool {

	/* vars */
	var result bool

	result = (ch == '"')

	/* enforce non void return */
	return result

}

func (this *Dialect) GetDynamicCommandDef(name string) []string {
	return []string(nil)
}

func (this *Dialect) GetDynamicCommands() []string {
	out := make([]string, 0)
	for k, _ := range this.DynaCommands {
		out = append(out, strings.ToUpper(k))
	}
	return out
}

func (this *Dialect) GetDynamicFunctions() []string {
	out := make([]string, 0)
	for k, _ := range this.DynaFunctions {
		out = append(out, strings.ToUpper(k))
	}
	return out
}

func (this *Dialect) GetPublicDynamicCommands() []string {
	out := make([]string, 0)
	for k, v := range this.DynaCommands {
		if !v.IsHidden() {
			out = append(out, strings.ToUpper(k))
		}
	}
	return out
}

func (this *Dialect) TokenListAsString(tokens types.TokenList) string {

	/* vars */
	var result string

	result = ""
	for _, tok1 := range tokens.Content {
		if result != "" {
			result = result + " "
		}

		if (tok1.Type == types.KEYWORD) || (tok1.Type == types.FUNCTION) || (tok1.Type == types.DYNAMICKEYWORD) {
			result = result + strings.ToUpper(tok1.AsString())
		} else {
			result = result + tok1.AsString()
		}
	}

	/* enforce non void return */
	return result

}

func (this *Dialect) IsString(instr string) bool {

	/* vars */
	var result bool

	result = (instr[0] == '"') || (instr[0] == '\'')

	/* enforce non void return */
	return result

}

func (this *Dialect) IsQ(ch rune) bool {

	/* vars */
	var result bool

	result = (ch == '\'')

	/* enforce non void return */
	return result

}

func (this *Dialect) Evaluate(chunk string, tokens *types.TokenList) bool {

	/* vars */
	var result bool
	var tok *types.Token
	var ptok *types.Token

	result = true // continue

	//System.Out.Println("EVALUATE: ", chunk);

	if len(chunk) == 0 {
		return result
	}

	tok = types.NewToken(types.INVALID, "")

	//  if (this.IsBoolean(chunk))
	//  begin
	//    tok.Type = BOOLEAN;
	//    tok.Content = chunk;
	//  }
	//  else
	if this.IsFunction(chunk) {
		tok.Type = types.FUNCTION
		tok.Content = strings.ToUpper(chunk)
	} else if this.IsDynaCommand(chunk) {
		tok.Type = types.DYNAMICKEYWORD
		tok.Content = strings.ToUpper(chunk)
	} else if this.IsDynaFunction(chunk) {
		tok.Type = types.DYNAMICFUNCTION
		tok.Content = strings.ToUpper(chunk)
	} else if this.IsKeyword(chunk) {
		tok.Type = types.KEYWORD
		tok.Content = strings.ToUpper(chunk)

		/* determine if (we should stop parsing */
		cc := this.Commands[strings.ToLower(chunk)]
		result = !cc.HasNoTokens()

	} else if this.IsLogic(chunk) {
		tok.Type = types.LOGIC
		tok.Content = chunk
	} else if this.IsFloat(chunk) {
		if (tokens.Size() > 0) && (tokens.RPeek().Type == types.OPERATOR) && (tokens.RPeek().Content == "-") {
			tokens.RPeek().Content = "-" + chunk
			tokens.RPeek().Type = types.NUMBER
		} else {
			tok.Type = types.NUMBER
			tok.Content = chunk
		}
	} else if this.IsInteger(chunk) {
		if (tokens.Size() > 0) && (tokens.RPeek().Type == types.OPERATOR) && (tokens.RPeek().Content == "-") {
			tokens.RPeek().Content = "-" + chunk
			tokens.RPeek().Type = types.INTEGER
		} else {
			tok.Type = types.INTEGER
			tok.Content = chunk
		}
	} else if this.IsSeparator(chunk) {
		tok.Type = types.SEPARATOR
		tok.Content = chunk
	} else if this.IsType(chunk) {
		tok.Type = types.TYPE
		tok.Content = chunk
	} else if this.IsOpenRBracket(chunk) || this.IsOpenSBracket(chunk) || this.IsOpenCBrace(chunk) {
		tok.Type = types.OBRACKET
		tok.Content = chunk
	} else if this.IsCloseRBracket(chunk) || this.IsCloseSBracket(chunk) || this.IsCloseCBrace(chunk) {
		tok.Type = types.CBRACKET
		tok.Content = chunk
	} else if this.IsOperator(chunk) {
		tok.Type = types.OPERATOR
		tok.Content = chunk
	} else if this.IsComparator(chunk) {
		if tokens.Size() > 0 {
			ptok = tokens.RPeek()
			if (ptok.Type == types.ASSIGNMENT) && (ptok.Content == "=") && ((chunk == ">") || (chunk == "<")) {
				ptok.Content = ptok.Content + chunk
			} else if ptok.Type == types.COMPARITOR {
				ptok.Content = ptok.Content + chunk
			} else {
				tok.Type = types.COMPARITOR
				tok.Content = chunk
			}
		} else {
			tok.Type = types.COMPARITOR
			tok.Content = chunk
		}
	} else if this.IsAssignment(rune(chunk[0])) {

		/* the the preceding token was a comparitor)we merge them */
		/* eg. ">" + "=" == ">=" */

		if tokens.Size() > 0 {
			ptok = tokens.RPeek()
			if (ptok.Type == types.COMPARITOR) && (chunk == "=") {
				ptok.Content = ptok.Content + "="
			} else {
				tok.Type = types.ASSIGNMENT
				tok.Content = chunk
			}
		} else {
			tok.Type = types.ASSIGNMENT
			tok.Content = chunk
		}
	} else if this.IsVariableName(chunk) {
		tok.Type = types.VARIABLE
		tok.Content = chunk
	} else if this.IsString(chunk) {
		tok.Type = types.STRING

		tok.Content = utils.Copy(chunk, 2, len(chunk)-2)
	}

	if tok.Type != types.INVALID {
		//System.Out.Println("ADD: ", tok.Content);
		log.Printf("Yielded token: %v\n", tok)

		/* shim for -No */
		if (tok.Type == types.NUMBER) && (tokens.Size() >= 2) {
			if (tokens.RPeek().Type == types.OPERATOR) &&
				(tokens.RPeek().Content == "-") &&
				((tokens.Get(tokens.Size()-2).Type == types.ASSIGNMENT) || (tokens.Get(tokens.Size()-2).Type == types.COMPARITOR)) {
				/* merge into last token, change to number type */
				tokens.Get(tokens.Size() - 1).Type = types.NUMBER
				tokens.Get(tokens.Size() - 1).Content = "-" + tok.Content
			} else {
				tokens.Push(tok)
			}
		} else {
			tokens.Push(tok)
		}
	}

	/* enforce non void return */
	return result

}

func (this *Dialect) ParseTokensForResult(ent interfaces.Interpretable, tokens types.TokenList) (*types.Token, error) {

	/* vars */
	var result *types.Token
	var ops *types.TokenList
	var values *types.TokenList
	var tidx int
	var rbc int
	var sbc int
	var i int
	var repeats int
	var blev int
	var tok *types.Token
	var lasttok *types.Token
	var ntok *types.Token
	var op *types.Token
	var a *types.Token
	var b *types.Token
	var n string
	var rs string
	var v *types.Variable
	var exptok *types.TokenList
	var subexpr *types.TokenList
	var par *types.TokenList
	var err bool
	var rr float64
	var aa float64
	var bb float64
	var dl []int
	var rrb bool
	var fun interfaces.Functioner
	var tla types.TokenListArray
	var defindex bool
	var lastop bool
	var hpop int
	//int  i;
	//Token tt;

	result = types.NewToken(types.INVALID, "")

	/* must be 1 || more tokens in list */
	if tokens.Size() == 0 {
		return result, nil
	}

	//System.Out.Println("*** Called to parse: ", ent.TokenListAsString(tokens));

	values = types.NewTokenList()
	ops = types.NewTokenList()

	// first fix: if (stream {s with an operator, prefix zero to the chain
	if tokens.Get(0).Type == types.OPERATOR {
		//System.Out.Println("BEEP");
		values.Push(types.NewToken(types.INTEGER, "0"))
	}

	/* init parser state */
	tidx = 0
	rbc = 0
	sbc = 0

	//fInterpreter.VDU.PutStr("Expression in: "+this.TokenListAsString(tokens)+"\r\n");
	lasttok = nil
	lastop = false
	/*main parse loop*/
	for tidx < tokens.Size() {
		tok = tokens.Get(tidx)

		//System.Out.Println( "--------------> type of token at tidx ", tidx, " is ", tok.Type );

		if (tok.Type == types.NUMBER) || (tok.Type == types.INTEGER) || (tok.Type == types.STRING) || (tok.Type == types.BOOLEAN) {
			// fix for missing + || separators
			if (lastop == false) && (lasttok != nil) {
				ops.Push(types.NewToken(types.OPERATOR, "+"))
			}

			values.Push(tok)
			lastop = false
		} else if tok.Type == types.ASSIGNMENT {
			ops.Push(tok)
			lastop = true
		} else if tok.Type == types.COMPARITOR {
			ops.Push(tok)
			lastop = true
		} else if tok.Type == types.LOGIC {
			ops.Push(tok)
			lastop = true
		} else if tok.Type == types.OPERATOR {
			if (lasttok != nil) && ((lasttok.Type == types.ASSIGNMENT) || (lasttok.Type == types.COMPARITOR)) && (tok.Content == "-") {
				values.Push(types.NewToken(types.NUMBER, "0"))
			}

			ops.Push(tok)
			lastop = true
		} else if tok.Type == types.FUNCTION {

			// fix for missing + || separators
			if (lastop == false) && (lasttok != nil) {
				ops.Push(types.NewToken(types.OPERATOR, "+"))
			}

			fun = this.Functions[strings.ToLower(tok.Content)]
			if fun == nil {
				return result, exception.NewESyntaxError("unknown function: " + tok.Content)
			}

			//fInterpreter.VDU.PutStr(fun.Name+"(\" + IntToStr(length(fun.FunctionParams))+\")"+"\r\n");

			if len(fun.FunctionParams()) > 0 {

				/* okay must be bracket after */
				tidx = tidx + 1
				if tidx >= tokens.Size() {
					return result, exception.NewESyntaxError(fun.GetName() + " requires params (size)")
				}
				tok = tokens.Get(tidx)
				if (tok.Type != types.OBRACKET) || (tok.Content != "(") {
					return result, exception.NewESyntaxError(fun.GetName() + " requires params: " + tok.Content)
				}

				// must be an index
				sbc = 1
				tidx = tidx + 1
				subexpr = types.NewTokenList()

				for (tidx < tokens.Size()) && (sbc > 0) {
					tok = tokens.Get(tidx)
					if (tok.Type == types.OBRACKET) && (tok.Content == "(") {
						sbc = sbc + 1
					}
					if (tok.Type == types.CBRACKET) && (tok.Content == "") {
						sbc = sbc - 1
					}
					if sbc > 0 {
						subexpr.Push(tok)
					}
					if (tidx < tokens.Size()) && (sbc > 0) {
						tidx = tidx + 1
					}
				}

				/* now condense this list down */
				tla = types.NewTokenListArray()
				exptok = types.NewTokenList()
				blev = 0

				for _, tok1 := range subexpr.Content {
					if (tok1.Type == types.SEPARATOR) && (tok1.Content == ",") && (blev == 0) {
						if exptok.Size() > 0 {
							/* don't need to grow this: SetLength(tla, tla.Size()+1); */
							tla = append(tla, *exptok)
							exptok = types.NewTokenList()
						}
					} else {

						if (tok1.Type == types.OBRACKET) && (tok1.Content == "(") {
							blev = blev + 1
						}

						if (tok1.Type == types.CBRACKET) && (tok1.Content == "") {
							blev = blev - 1
						}

						exptok.Push(tok1)
					}
				}

				if exptok.Size() > 0 {
					/* don't need to grow this: SetLength(tla, tla.Size()+1); */
					tla = append(tla, *exptok)
				}

				par = types.NewTokenList()
				for _, exptok1 := range tla {
					//System.Out.Println( fun.Name, ": ", this.TokenListAsString(exptok) );
					if fun.GetRaw() {
						if par.Size() > 0 {
							par.Push(types.NewToken(types.SEPARATOR, ","))
						}
						for _, tok1 := range exptok1.Content {
							par.Push(tok1)
						}
					} else {
						tok, _ := this.ParseTokensForResult(ent, exptok1)
						par.Push(tok)
					}
					//// FreeAndNil(exptok);
				}
			} else {
				par = types.NewTokenList()
			}

			fun.SetEntity(ent)
			fun.FunctionExecute(par)

			vv := fun.GetStack().Pop()

			values.Push(vv)
			lastop = false
		} else if tok.Type == types.VARIABLE {

			// fix for missing + || separators
			if (lastop == false) && (lasttok != nil) {
				ops.Push(types.NewToken(types.OPERATOR, "+"))
			}

			/* first try entity local */
			n = strings.ToLower(tok.Content)
			v = nil
			if ent.GetLocal().Contains(n) {
				v = ent.GetLocal().Get(n)
			} else
			//if (this.Producer.Global.ContainsKey(n))
			if ent.ExistsVar(n) {
				//v = this.Producer.Global.Get(n);
				v = ent.GetVar(n)
			}
			/* fall out if (var does not exist */
			if v == nil {
				//result.Type = INVALID;
				//return result;
				if n[len(n)-1] == '$' {
					values.Push(types.NewToken(types.STRING, ""))
				} else {
					values.Push(types.NewToken(types.NUMBER, "0"))
				}
			} else if v.IsArray() {
				defindex = false
				tidx = tidx + 1
				if tidx >= tokens.Size() {
					if v.AssumeLowIndex {
						defindex = true
					} else {
						return result, exception.NewESyntaxError(v.Name + " requires index")
					}
				} else {
					tok = tokens.Get(tidx)
				}

				if (tok.Type != types.OBRACKET) || (tok.Content != "(") {
					if v.AssumeLowIndex {
						defindex = true
					} else {
						return result, exception.NewESyntaxError(v.Name + " requires index")
					}
				}

				if defindex {
					tidx = tidx - 1
				}

				// must be an index
				subexpr = types.NewTokenList()
				if !defindex {
					sbc = 1
					tidx = tidx + 1
					subexpr.Push(types.NewToken(types.OBRACKET, "("))
					for (tidx < tokens.Size()) && (sbc > 0) {
						tok = tokens.Get(tidx)
						if (tok.Type == types.OBRACKET) && (tok.Content == "(") {
							sbc = sbc + 1
						}
						if (tok.Type == types.CBRACKET) && (tok.Content == "") {
							sbc = sbc - 1
						}
						if sbc > 0 {
							subexpr.Push(tok)
						}
						if (tidx < tokens.Size()) && (sbc > 0) {
							tidx = tidx + 1
						}
					}
				}

				if defindex {
					//SetLength(dl, length(v.Dimensions));
					dl = make([]int, len(v.Dimensions()))
					for i = 0; i < len(dl)-1; i++ {
						dl[i] = 0
					}
					//this.VDU.PutStr("[\"+v.Name+\" assume def index 0]")
				} else {
					subexpr.Push(types.NewToken(types.CBRACKET, ")"))
					//ent.GetVDU().PutStr("*** VAR INDICES ARE ["+this.TokenListAsString(subexpr)+"]"+"\r\n" );
					var derr error
					dl, derr = ent.IndicesFromTokens(*subexpr, "(", ")")
					// FreeAndNil(subexpr);
					if derr != nil {
						return result, derr
					}
				}

				/* var exists */
				var e error

				ss, e := v.GetContentByIndex(this.ArrayDimDefault, this.ArrayDimMax, dl)
				if e != nil {
					return result, e
				}

				ntok = types.NewToken(types.INVALID, ss)
				//  VariableType == (vtString, vtBoolean, vtFloat, vtInteger, vtExpression);
				switch v.Kind { /* FIXME: Switch statement needs cleanup */
				case types.VT_STRING:
					ntok.Type = types.STRING
					break
				case types.VT_FLOAT:
					ntok.Type = types.NUMBER
					break
				case types.VT_INTEGER:
					ntok.Type = types.INTEGER
					break
				case types.VT_BOOLEAN:
					ntok.Type = types.BOOLEAN
					break
				case types.VT_EXPRESSION:
					{
						ss, e = v.GetContentByIndex(this.ArrayDimDefault, this.ArrayDimMax, dl)
						if e != nil {
							return result, e
						}
						exptok = this.Tokenize(runestring.Cast(ss))
						ntok, e = this.ParseTokensForResult(ent, *exptok)
						if e != nil {
							return result, e
						}
						break
					}
				}
				//this.VDU.PutStr("// Adding value from array "+ntok.Content);
				values.Push(ntok)

			} else {

				/* var exists */
				ss, e := v.GetContentScalar()
				if e != nil {
					return result, e
				}
				ntok = types.NewToken(types.INVALID, ss)
				//  VariableType == (vtString, vtBoolean, vtFloat, vtInteger, vtExpression);
				switch v.Kind { /* FIXME: Switch statement needs cleanup */
				case types.VT_STRING:
					ntok.Type = types.STRING
					break
				case types.VT_FLOAT:
					ntok.Type = types.NUMBER
					break
				case types.VT_INTEGER:
					ntok.Type = types.INTEGER
					break
				case types.VT_BOOLEAN:
					ntok.Type = types.BOOLEAN
					break
				case types.VT_EXPRESSION:
					{
						ss, e = v.GetContentScalar()
						if e != nil {
							return result, e
						}
						exptok = this.Tokenize(runestring.Cast(ss))
						ntok, e = this.ParseTokensForResult(ent, *exptok)
						if e != nil {
							return result, e
						}
						break
					}
				}
				values.Push(ntok)

			}

			lastop = false
		} else if (tok.Type == types.OBRACKET) && (tok.Content == "(") {
			rbc = 1
			tidx = tidx + 1
			subexpr = types.NewTokenList()
			for (tidx < tokens.Size()) && (rbc > 0) {
				tok = tokens.Get(tidx)
				if (tok.Type == types.OBRACKET) && (tok.Content == "(") {
					rbc = rbc + 1
				}
				if (tok.Type == types.CBRACKET) && (tok.Content == "") {
					rbc = rbc - 1
				}
				if rbc > 0 {
					subexpr.Push(tok)
				}
				if (tidx < tokens.Size()) && (rbc > 0) {
					tidx = tidx + 1
				}
			}

			//System.Out.Println( "*** Must parse bracketed subexpression: ", this.TokenListAsString(subexpr) );
			//System.Out.Println( "=== rbc == ",rbc ;

			if rbc > 0 {
				return result, exception.NewESyntaxError("O_o")
			}

			/* solve the expression */
			ntok, e := this.ParseTokensForResult(ent, *subexpr)
			if e != nil {
				return result, e
			}
			// FreeAndNil(subexpr);
			values.Push(ntok)
			lastop = false
		}

		lasttok = tok
		tidx = tidx + 1
	}

	/* now we have ops, values etc - lets actually parse the expression */
	//System.Out.Println("Op stack");
	//for _, ntok := range ops.ToStringArray()
	//    Writeln("OP: ", ntok.Type, ", ", ntok.Content);
	//for _, ntok := range values.ToStringArray()
	//    Writeln("VALUE: ", ntok.Type, ", ", ntok.Content);
	//System.Out.Println("--------");

	/* This bit is some magic || something yea!!! Whoo ---------------- */
	/* End magic bits ------------------------------------------------- */

	/* process */
	err = false

	for (ops.Size() > 0) && (!err) {

		hpop = this.HPOpIndex(*ops)
		op = ops.Remove(hpop)

		if op.Type == types.LOGIC {

			if strings.ToLower(op.Content) == "not" {

				//a = values.Pop;
				a = values.Remove(hpop)

				//ent.GetVDU().PutStr(op.Content + " " + a.Content)

				/*if (a.AsInteger != 0)
				      values.Push( NewToken(NUMBER, "0") )
				  else {
				      values.Push( NewToken(NUMBER, "1") );*/
				//}

				if a.AsInteger() != 0 {
					ntok = types.NewToken(types.NUMBER, "0")
				} else {
					ntok = types.NewToken(types.NUMBER, "1")
				}

				values.Insert(hpop, ntok)
			} else if strings.ToLower(op.Content) == "and" {
				//b = values.Pop;
				//a = values.Pop;
				a = values.Remove(hpop)
				b = values.Remove(hpop)

				//ent.GetVDU().PutStr(a.Content + " " + op.Content + " " + b.Content + "\r\n")

				//values.Push( NewToken(NUMBER, IntToStr(a.AsInteger && b.AsInteger)) );
				ntok = types.NewToken(types.NUMBER, utils.IntToStr(a.AsInteger()&b.AsInteger()))
				values.Insert(hpop, ntok)
			} else if strings.ToLower(op.Content) == "or" {
				//b = values.Pop;
				//a = values.Pop;
				a = values.Remove(hpop)
				b = values.Remove(hpop)

				//ent.GetVDU().PutStr(a.Content + " " + op.Content + " " + b.Content + "\r\n")

				//values.Push( NewToken(NUMBER, IntToStr(a.AsInteger || b.AsInteger)) );
				ntok = types.NewToken(types.NUMBER, utils.IntToStr(a.AsInteger()|b.AsInteger()))
				values.Insert(hpop, ntok)
			} else if strings.ToLower(op.Content) == "xor" {
				//b = values.Pop;
				//a = values.Pop;
				a = values.Remove(hpop)
				b = values.Remove(hpop)

				//ent.GetVDU().PutStr(a.Content + " " + op.Content + " " + b.Content + "\r\n")

				//values.Push( NewToken(NUMBER, IntToStr(a.AsInteger xor b.AsInteger)) );
				ntok = types.NewToken(types.NUMBER, utils.IntToStr(a.AsInteger()^b.AsInteger()))
				values.Insert(hpop, ntok)
			}

		} else if (op.Type == types.COMPARITOR) || (op.Type == types.ASSIGNMENT) {

			if values.Size() < 2 {
				return result, exception.NewESyntaxError("Syntax Error")
			}

			//System.Out.Println("@@@@@@@@@@@@@@@@@@@@@ COMPARE");

			/*if (left) {
			      a = values.Left
			      b = values.Left
			  } else {
			      b = values.Right
			      a = values.Right
			  }*/
			a = values.Remove(hpop)
			b = values.Remove(hpop)

			// ent.GetVDU().PutStr(a.Content + " " + op.Content + " " + b.Content + "\r\n")

			if a.IsNumeric() && b.IsNumeric() {

				aa = float64(a.AsNumeric())
				bb = float64(b.AsNumeric())
				rrb = false

				//System.Out.Println("====================> NUMBER COMPARE a == [",a.Content,"], b == [",b.Content,"]");

				if op.Content == ">" {
					rrb = (aa > bb)
				} else if op.Content == "<" {
					rrb = (aa < bb)
				} else if op.Content == ">=" {
					rrb = (aa >= bb)
				} else if op.Content == "<=" {
					rrb = (aa <= bb)
				} else if op.Content == "!=" {
					rrb = (aa != bb)
				} else if op.Content == "=" {
					rrb = (aa == bb)
				} else {
					err = true
					break
				}

				n = "0"
				if rrb {
					n = "1"
				}
				ntok = types.NewToken(types.BOOLEAN, n)

				/*if (left)
				     values.Unshift(ntok)
				  else {
				      values.Push(ntok);*/
				//}
				values.Insert(hpop, ntok)
			} else if (!a.IsNumeric()) || (!b.IsNumeric()) {

				//tt := NUMBER /* most results are string */
				rrb = false

				//System.Out.Println("====================> STRING COMPARE a == [",a.Content,"], b == [",b.Content,"]");

				if op.Content == ">" {
					rrb = (a.Content > b.Content)
				} else if op.Content == "<" {
					rrb = (a.Content < b.Content)
				} else if op.Content == "!=" {
					rrb = (a.Content != b.Content)
				} else if op.Content == ">=" {
					rrb = (a.Content >= b.Content)
				} else if op.Content == "<=" {
					rrb = (a.Content <= b.Content)
				} else if op.Content == "=" {
					rrb = (a.Content == b.Content)
				} else {
					err = true
					break
				}

				n = "0"
				if rrb {
					n = "1"
				}
				ntok = types.NewToken(types.BOOLEAN, n)
				/*if (left)
				     values.Unshift(ntok)
				  else {
				      values.Push(ntok);*/
				//}
				values.Insert(hpop, ntok)

			}

		} else if op.Type == types.OPERATOR {

			//System.Out.Println("Currently ",values.Size()," values in stack... About to pop 2");

			if values.Size() < 2 {
				return result, exception.NewESyntaxError("Syntax Error")
			}

			/*if (left) {
			      a = values.Left
			      b = values.Left;
			  } else {
			      b = values.Right
			      a = values.Right
			  }*/
			a = values.Remove(hpop)
			b = values.Remove(hpop)

			//fInterpreter.VDU.PutStr("Op is: "+a.Content+op.Content+b.Content+"\r\n");
			//ent.GetVDU().PutStr(a.Content + " " + op.Content + " " + b.Content + "\r\n")

			if a.IsNumeric() && b.IsNumeric() {

				aa = float64(a.AsNumeric())
				bb = float64(b.AsNumeric())
				rr = 0

				//this.VDU.PutStr(a.Content+" "+op.Content+" "+b.Content+"\r\n");

				if op.Content == "+" {
					rr = aa + bb
				} else if op.Content == "-" {
					rr = aa - bb
				} else if op.Content == "*" {
					rr = aa * bb
				} else if op.Content == "/" {
					if bb == 0 {
						return result, exception.NewESyntaxError("DIVIDE BY ZERO ERROR")
					} else {
						rr = aa / bb
					}
				} else if op.Content == "^" {
					rr = math.Pow(aa, bb)
				} else if op.Content == "%" {
					rr = float64(int(aa) % int(bb))
				} else {
					err = true
					break
				}

				n = utils.FloatToStr(rr)
				ntok = types.NewToken(types.NUMBER, n)
				/*
				   if (trunc(rr) == rr) {
				       Str(trunc(rr), n)
				       ntok = NewToken( ttINTEGER, n )
				   } else {
				       ntok = NewToken( ttNUMBER, n )
				   }
				   var / *
				   /*if (left)
				      values.Unshift(ntok)
				   else {
				       values.Push(ntok);*/
				//}
				values.Insert(hpop, ntok)
			} else if (a.IsNumeric() && (!b.IsNumeric())) || (b.IsNumeric() && (!a.IsNumeric())) {

				tt := types.STRING /* most results are string */
				rs = ""

				if op.Content == "+" {
					rs = a.Content + b.Content
				} else if op.Content == "-" {
					// remove b from a, ignoring case
					//rs = StringReplace( a.Content, b.Content, '', [rfReplaceAll, rfIgnoreCase] );
					rs = strings.Replace(a.Content, b.Content, "", -1)
				} else if op.Content == "*" {
					repeats = 0
					rs = ""
					n = ""
					if a.IsNumeric() {
						repeats = int(a.AsNumeric())
						n = b.Content
					} else {
						repeats = int(b.AsNumeric())
						n = a.Content
					}
					for i = 1; i <= repeats; i++ {
						rs = rs + n
					}
				} else if op.Content == "/" {
					// divide string by substring - count occurrences
					//tt := INTEGER

					repeats = 0
					//i = 1;
					//for ((NPos(b.Content, a.Content, i) > 0))
					//{
					//    repeats++
					//    i++
					//}

					//Str( repeats, rs );
				} else {
					err = true
					break
				}

				ntok = types.NewToken(tt, rs)
				//System.Out.Println("CREATE RESULT: ", rs);
				/*if (left)
				     values.Unshift(ntok)
				  else {
				      values.Push(ntok);*/
				//}
				values.Insert(hpop, ntok)

			} else if (!a.IsNumeric()) && (!b.IsNumeric()) {

				tt := types.STRING /* most results are string */
				rs = ""

				if op.Content == "+" {
					rs = a.Content + b.Content
				} else if op.Content == "-" {
					rs = strings.Replace(a.Content, b.Content, "", -1)
				} else if op.Content == "/" {
					// divide string by substring - count occurrences
					tt = types.INTEGER

					repeats = 0
				} else {
					err = true
					break
				}

				ntok = types.NewToken(tt, rs)
				/*if (left)
				     values.Unshift(ntok)
				  else {
				      values.Push(ntok);*/
				//}
				values.Insert(hpop, ntok)
			}

		}

	}

	if err || (values.Size() > 1) {
		return result, exception.NewESyntaxError("Syntax Error")
	}

	result = values.Pop()

	//values.Free;
	//ops.Free;

	/* enforce non void return */
	return result, nil
}

func (this *Dialect) AddFunction(s string, cmd interfaces.Functioner) {

	this.Functions[strings.ToLower(s)] = cmd

}

func (this *Dialect) AddPlusFunction(ns string, s string, cmd interfaces.Functioner) {

	m, ok := this.PlusFunctions[ns]
	if !ok {
		m = interfaces.NewFunctionList()
	}

	m[strings.ToLower(s)] = cmd
	this.PlusFunctions[ns] = m

	this.ShadowPlusFunctions[ns+"."+s] = cmd // for lookups

}

func (this *Dialect) AddHiddenPlusFunction(ns string, s string, cmd interfaces.Functioner) {

	cmd.SetHidden(true)

	m, ok := this.PlusFunctions[ns]
	if !ok {
		m = interfaces.NewFunctionList()
	}

	m[strings.ToLower(s)] = cmd
	this.PlusFunctions[ns] = m

	this.ShadowPlusFunctions[ns+"."+s] = cmd // for lookups

}

func (this *Dialect) AddUserCommand(s string, cmd *DynaCode) {

	/* vars */

	cmd.SetDialect(this)
	cmd.Init()
	this.UserCommands[strings.ToLower(s)] = cmd

}

func (this *Dialect) IsKeyword(instr string) bool {

	/* vars */
	var result bool

	_, result = this.Commands[strings.ToLower(instr)]

	/* enforce non void return */
	return result

}

func (this *Dialect) IsVariableName(instr string) bool {

	result := false
	var ch rune
	var i int

	if rune(instr[0]) == '@' {
		return true
	}

	for i = 0; i <= len(instr)-2; i++ {
		ch = rune(instr[i])
		if !(this.IsDigit(ch) || this.IsAlpha(ch) || ch == '.') {
			return result
		}
	}

	ch = rune(instr[len(instr)-1])
	if this.IsDigit(ch) || this.IsAlpha(ch) || (utils.Pos(string(ch), this.VarSuffixes) > 0) {
		//System.Out.Println("- Last is digit, Alpha || VarSuffix");
		result = true
	}

	return result

}

func (this *Dialect) IsVariableNameOld(instr string) bool {

	/* vars */
	var result bool
	var ch rune
	var i int

	result = false

	if len(instr) == 0 {
		return result
	}

	//System.Out.Println("IsVariableName(",instr,")");

	if !this.IsAlpha(rune(instr[0])) {
		return result
	}

	//System.Out.Println("- First is alpha");

	for i = 1; i <= len(instr)-2; i++ {
		ch = rune(instr[i])
		if !(this.IsDigit(ch) || this.IsAlpha(ch)) {
			return result
		}
	}

	//System.Out.Println("- Then alpha || numbers");

	ch = rune(instr[len(instr)-1])
	if this.IsDigit(ch) || this.IsAlpha(ch) || (utils.Pos(string(ch), this.VarSuffixes) > 0) {
		//System.Out.Println("- Last is digit, Alpha || VarSuffix");
		result = true
	}

	/* enforce non void return */
	return result

}

func (this *Dialect) IsLogic(ch string) bool {

	/* vars */
	var result bool

	ch = strings.ToLower(ch)
	result = (ch == "and") || (ch == "or") || (ch == "not") || (ch == "xor")

	/* enforce non void return */
	return result

}

func (this *Dialect) IsComparator(ch string) bool {

	/* vars */
	var result bool

	result = (ch == "<") || (ch == ">") || (ch == "#")

	/* enforce non void return */
	return result

}

func (this *Dialect) IsInteger(instr string) bool {

	//var i int64

	_, err := strconv.ParseInt(instr, 10, 32)
	return (err == nil)
}

func (this *Dialect) IsVarSuffix(ch rune, vs string) bool {

	var items string = vs

	return (utils.PosRune(ch, items) > 0)
}

func (this *Dialect) AddDynaCommand(s string, cmd *DynaCode) {

	if cmd.GetDialect() == nil {
		cmd.SetDialect(this)
	}
	cmd.Init()
	this.DynaCommands[strings.ToLower(s)] = cmd

}

func (this *Dialect) IsFunction(instr string) bool {

	/* vars */
	var result bool

	_, result = this.Functions[strings.ToLower(instr)]

	/* enforce non void return */
	return result

}

func (this *Dialect) IsOpenRBracket(instr string) bool {

	/* vars */
	var result bool

	result = (instr == "(")

	/* enforce non void return */
	return result

}

func (this *Dialect) IsCloseCBrace(instr string) bool {

	/* vars */
	var result bool

	result = (instr == "}")

	/* enforce non void return */
	return result

}

func (this *Dialect) ExecuteDynamicFunction(ent interfaces.Interpretable, funcname string, values types.TokenList) (*types.Token, error) {

	/* vars */
	var result *types.Token
	var pc int
	var i int
	var stacklevel int
	var dcmd interfaces.DynaCoder
	var arglist *types.TokenList
	var cr *types.CodeRef

	result = types.NewToken(types.INVALID, "0")
	_, ok := this.DynaFunctions[strings.ToLower(funcname)]
	if !ok {
		return result, exception.NewESyntaxError("Undefined function: " + funcname)
	}

	/* Established function exists */
	dcmd, _ = this.DynaFunctions[strings.ToLower(funcname)]
	pc = dcmd.GetParamCount()

	arglist = types.NewTokenList()
	for i = 1; i <= pc; i++ {
		t := values.Pop() // off right
		if t == nil {
			return result, exception.NewESyntaxError("Syntax Error")
		}
		arglist.UnShift(t)
	}

	//System.Out.Println( "Would execute dynamic function "+funcname+" with args "+this.TokenListAsString(arglist));

	/*
	 * What we want to do here is create another stack frame based on the input tokens from
	 * the values list && run it to completion.
	 *
	 * At that point we will have our function result.
	 *
	 * */

	defer func() {
		if r := recover(); r != nil {
			this.HandleException(ent, errors.New(r.(string)))
		}
	}()

	/* its actually a hidden subroutine call */
	cr = types.NewCodeRef()
	cr.Line = dcmd.GetCode().GetLowIndex()
	cr.Statement = 0
	cr.Token = 0
	if cr.Line != -1 {

		/* something to do */
		ent.Call(*cr, dcmd.GetCode(), ent.GetState(), false, funcname+ent.GetVarPrefix(), *arglist, dcmd.GetDialect()) // call with isolation off

		/* at this point we save the current stack level */
		stacklevel = ent.GetStack().Size() - 1

		for (ent.GetStack().Size() > stacklevel) && (ent.IsRunning() || ent.IsRunningDirect()) {
			if ent.GetState() == types.RUNNING {
				ent.RunStatement()
			} else {
				ent.RunStatementDirect()
			}
		}

		/* we have our result now - should be in arglist */
		//System.Out.Println( "Args in list after execute function.Equals("+ent.TokenStack.Size()+"): " + this.TokenListAsString(ent.TokenStack) );
		result = ent.GetTokenStack().Shift()
		if result == nil {
			return result, exception.NewESyntaxError("Expected token")
		}
	} else {
		return result, exception.NewESyntaxError("Dynamic Code Hook has no content")
	}

	return result, nil

}

func (this *Dialect) IsOpenSBracket(instr string) bool {

	/* vars */
	var result bool

	result = (instr == "[")

	/* enforce non void return */
	return result

}

func (this *Dialect) IsAssignment(ch rune) bool {

	/* vars */
	var result bool

	result = (ch == '=')

	/* enforce non void return */
	return result

}

func (this *Dialect) HPOpIndex(tl types.TokenList) int {
	var result int = -1
	var hs int = 1
	var tt *types.Token
	var sc int

	for i := 0; i <= tl.Size()-1; i++ {
		tt = tl.Get(i)
		sc = 0

		if tt.Content == "^" {
			sc = 500
		} else if (tt.Content == "*") || (tt.Content == "/") {
			sc = 400
		} else if (tt.Content == "+") || (tt.Content == "-") {
			sc = 300
		} else if (tt.Type == types.COMPARITOR) || (tt.Type == types.ASSIGNMENT) {
			sc = 200
		} else if tt.Type == types.LOGIC {
			s := strings.ToLower(tt.Content)
			if s == "not" {
				sc = 150
			} else if s == "and" {
				sc = 140
			} else {
				sc = 130
			}
		} else {
			sc = 100
		}

		if sc > hs {
			hs = sc
			result = i
		}
	}

	return result
}

func (this *Dialect) InitVDU(vDU interfaces.Interpretable, promptonly bool) {
	// TODO Auto-generated method stub

}

func (this *Dialect) HandleException(ent interfaces.Interpretable, e error) {

	//	if ent.GetVDU().GetCursorX() != 0 {
	//		ent.GetVDU().PutStr("\r\n")
	//	}

	//ent.GetVDU().PutStr("? " + e.Error())

	if (ent.GetState() == types.RUNNING) || (ent.GetState() == types.DIRECTRUNNING) {
		if !ent.HandleError() {

			if ent.GetState() == types.RUNNING {
				//				ent.GetVDU().PutStr(" at line " + utils.IntToStr(ent.GetPC().Line))
			}
			ent.Halt()
		}
	}

	//ent.GetVDU().PutStr("\r\n")

}

func (this *Dialect) ExecuteDirectCommand(tl types.TokenList, ent interfaces.Interpretable, Scope *types.Algorithm, LPC *types.CodeRef) error {

	/* vars */
	var tok *types.Token
	var n string
	var cmd interfaces.Commander
	var cr *types.CodeRef
	var ss *types.TokenList

	if this.NetBracketCount(tl) != 0 {
		return exception.NewESyntaxError("SYNTAX ERROR")
	}

	//System.Out.Println( "DEBUG: -------------> ["+this.Title+"]: "+LPC.Line+" "+ent.TokenListAsString(tl) );

	if ent.IsDebug() && ent.IsRunning() {

		if tl.Get(0).Content == "next" {
			if ent.GetLoopStack().Size() > 0 {
				zz := ent.GetLoopStack().Get(ent.GetLoopStack().Size() - 1).Entry

				if zz.Line == ent.GetPC().Line && zz.Statement == ent.GetPC().Statement {
					ent.SetSilent(true)
					defer ent.SetSilent(false)
				}
			}
		}

		ent.Log("DO", ent.TokenListAsString(tl))

	}

	if this.Trace && (ent.GetState() == types.RUNNING) {
		//ent.GetVDU().PutStr("#" + utils.IntToStr(LPC.Line) + " ")
	}

	/* process poop monster here (@^-^@) */
	tok = tl.Shift()

	if (tok.Type == types.NUMBER) || (tok.Type == types.INTEGER) {
	} else if tok.Type == types.DYNAMICKEYWORD {
		n = strings.ToLower(tok.Content)
		if dcmd, ok := this.DynaCommands[n]; ok {
			//dcmd = this.DynaCommands.Get(n)

			//ent.GetVDU().PutStr("Dynamic command parsing - Start at "+IntToStr(dcmd.Code.LowIndex)+"\r\n");

			defer func() {
				if r := recover(); r != nil {
					this.HandleException(ent, errors.New(r.(string)))
				}
			}()

			//try {
			/* its actually a hidden subroutine call */
			cr = types.NewCodeRef()
			cr.Line = dcmd.GetCode().GetLowIndex()
			cr.Statement = 0
			cr.Token = 0
			if cr.Line != -1 {
				/* something to do */
				ss = tl.SubList(0, tl.Size())
				ent.Call(*cr, dcmd.GetCode(), ent.GetState(), false, n+ent.GetVarPrefix(), *ss, dcmd.GetDialect()) // call with isolation off
			} else {
				return exception.NewESyntaxError("Dynamic Code Hook has no content")
			}
			//}
			//catch (Exception e) {

			//}

		}

	} else if tok.Type == types.PLUSFUNCTION {

		// Handle plus function execution here
		fun, exists, ns, err := this.PlusFunctions.GetFunctionByNameContext(this.CurrentNameSpace, strings.ToLower(tok.Content))
		//fun, exists := this.PlusFunctions[strings.ToLower(tok.Content)]

		if !exists {
			this.HandleException(ent, err)
			return nil
		}

		this.CurrentNameSpace = ns

		////fmt.Println(ent.TokenListAsString(tl))

		if tl.RPeek() != nil && tl.RPeek().Content == "}" {

			sl := tl.SubList(0, tl.Size()-1)
			////fmt.Println(ent.TokenListAsString(*sl))

			tla := ent.SplitOnTokenWithBrackets(*sl, *types.NewToken(types.SEPARATOR, ","))

			subexpr := types.NewTokenList()

			if fun.GetRaw() {
				subexpr = sl
			} else {

				for _, tt := range tla {
					v, _ := this.ParseTokensForResult(ent, tt)
					subexpr.Push(v)
				}
				////fmt.Println(ent.TokenListAsString(*subexpr))
			}

			fun.SetEntity(ent)
			fun.SetQuery(false)
			fun.FunctionExecute(subexpr)

		} else {
			this.HandleException(ent, exception.NewESyntaxError("SYNTAX ERROR"))
		}

	} else if tok.Type == types.KEYWORD {

		if (tl.Size() > 0) && (tl.LPeek().Type == types.ASSIGNMENT) {
			return exception.NewESyntaxError("SYNTAX ERROR")
		}

		n = strings.ToLower(tok.Content)
		if cmd, ok := this.Commands[n]; ok {
			//cmd = this.Commands.Get(n

			_, err := cmd.Execute(nil, ent, tl, Scope, *LPC)
			if err != nil {
				this.HandleException(ent, err)
			}
			cost := cmd.GetCost()
			if cost == 0 {
				cost = this.GetDefaultCost()
			}
			//ent.SetWaitUntil(time.Now().UnixNano() + (int64)(float32(cost)*(100/this.Throttle)))
			ent.Wait((int64)(float32(cost) * (100 / this.Throttle)))
		} else {
			//System.Out.Println("DOES NOT EXIST!")
		}

	} else if (tok.Type == types.VARIABLE) && (this.ImpliedAssign != nil) {

		/* assign variable here */
		tl.UnShift(tok)
		cmd = this.ImpliedAssign

		_, err := cmd.Execute(nil, ent, tl, Scope, *LPC)
		if err != nil {
			this.HandleException(ent, err)
		}

		cost := cmd.GetCost()
		if cost == 0 {
			cost = this.GetDefaultCost()
		}
		ent.Wait((int64)(float32(cost) * (100 / this.Throttle)))

	} else if (tok.Type == types.PLUSVAR) && (this.PlusHandler != nil) {

		/* assign variable here */
		tl.UnShift(tok)
		cmd = this.PlusHandler

		_, err := cmd.Execute(nil, ent, tl, Scope, *LPC)
		if err != nil {
			this.HandleException(ent, err)
		}

		cost := cmd.GetCost()
		if cost == 0 {
			cost = this.GetDefaultCost()
		}
		ent.Wait((int64)(float32(cost) * (100 / this.Throttle)))

	} else {
		return exception.NewESyntaxError("SYNTAX ERROR")
	}

	///tl.Free; /* clean up */
	return nil
}

func (this *Dialect) GetCommands() interfaces.CommandList {
	return this.Commands
}

func (this *Dialect) Parse(ent interfaces.Interpretable, s string) error {

	/* vars */
	var tl *types.TokenList
	//TokenList  cl;
	var cmdlist types.TokenListArray
	//Token tok;
	var lno int
	var ll types.Line
	var st types.Statement

	tl = this.Tokenize(runestring.Cast(s))

	log.Printf("Dialect [%s] received [%s] to parse from [%s]\n", this.GetTitle(), s, ent.GetName())

	//for _, tok := range tl
	//{
	//  Str( tok.Type, n );
	//  Producer.GetInstance().VDU.PutStr("T: "+n+", V: "+tok.Content+"\r\n");
	//}

	if tl.Size() == 0 {
		return nil
	}

	tok := tl.Get(0)

	if tok.Type == types.NUMBER {

		tok = tl.Shift()

		if tl.Size() > 0 {

			cmdlist = ent.SplitOnToken(*tl, *types.NewToken(types.SEPARATOR, ":"))

			////fmt.Println(len(cmdlist))

			lno = tok.AsInteger()
			ll = types.NewLine()

			for _, cl := range cmdlist {
				if cl.Size() > 0 {
					st = types.NewStatement()
					for _, tok1 := range cl.Content {
						st.Push(tok1)
					}
					ll = append(ll, st)
				} else {
					st = types.NewStatement()
					st.Push(types.NewToken(types.KEYWORD, "REM"))
					ll = append(ll, st)
				}
			}

			z := ent.GetCode()

			z.Put(lno, ll)

			//ent.GetVDU().PutStr("+" + utils.IntToStr(lno) + "\r\n")
			ent.SetState(types.STOPPED)
			//fInterpreter.VDU.Dump;
		} else {
			lno = tok.AsInteger()
			if _, ok := ent.GetCode().Get(lno); ok {
				//delete(ent.GetCode(), lno)
				z := ent.GetCode()
				z.Remove(lno)
				//ent.GetVDU().PutStr("-"+PasUtil.IntToStr(lno)+"\r\n");
			}
		}
		return nil
	}

	/* at this point we break it into statements */
	cmdlist = ent.SplitOnToken(*tl, *types.NewToken(types.SEPARATOR, ":"))
	lno = 999
	ent.SetDirectAlgorithm(types.NewAlgorithm())
	ent.SetLPC(types.NewCodeRef())

	ll = types.NewLine()

	for zz, cl := range cmdlist {
		log.Println(zz)
		if cl.Size() > 0 {
			st = types.NewStatement()
			for _, zz := range cl.Content {
				st.Add(zz)
			}
			ll.Push(st)
			log.Println("Added:", st)
			log.Println("Statement count ", len(ll))
		} else {
			log.Println("cmdlist has no content")
		}
	}

	//ent.GetDirectAlgorithm()[lno] = ll
	a := ent.GetDirectAlgorithm()
	a.Put(lno, ll)
	ent.SetDirectAlgorithm(a)
	ent.GetLPC().Line = lno
	ent.GetLPC().Statement = 0
	ent.GetLPC().Token = 0

	//a := ent.GetDirectAlgorithm()
	//	ent.GetVDU().PutStr(fmt.Sprintf("Direct has %d lines\r\n", len(a)))

	//str(ent.GetState(), n);
	//fInterpreter.VDU.PutStr(n+": run direct with "+IntToStr(ll.Size())+" commands: "+s+"\r\n");

	// start running
	ent.SetState(types.DIRECTRUNNING)

	//	ent.GetVDU().PutStr(fmt.Sprintf("Interpreter state is now %v\r\n", ent.GetState()))

	return nil

}

func (this *Dialect) AddDynaFunction(s string, cmd *DynaCode) {

	/* vars */

	if cmd.Dialect == nil {
		cmd.Dialect = this
	}
	cmd.Init()
	this.DynaFunctions.Put(strings.ToLower(s), cmd)

}

func (this *Dialect) IsType(instr string) bool {

	/* vars */
	var result bool

	result = (strings.ToLower(instr) == "integer") || (strings.ToLower(instr) == "float") ||
		(strings.ToLower(instr) == "string") ||
		(strings.ToLower(instr) == "expression")

	/* enforce non void return */
	return result

}

func (this *Dialect) LoadFromFile(filename string, rootdia interfaces.Dialecter) {

	/* vars */
	var ini utils.TINIFile
	var kw string
	var wd string
	var fn string
	var tmp string
	var i int
	var argc int
	var dc *DynaCode

	ini = *utils.NewTINIFile(filename)
	this.Title = ini.ReadString("Definition", "Title", "untitled")
	kw = ini.ReadString("Definition", "Keywords", "")

	//System.Out.Println("Keywords = "+kw);

	for len(kw) > 0 {
		i = utils.Pos(",", kw)
		if i > 0 {
			wd = utils.Copy(kw, 1, i-1)
			kw = utils.Delete(kw, 1, i)
		} else {
			wd = kw
			kw = ""
		}
		/* process keyword */
		fn = ini.ReadString("Keyword", wd, "")
		if fn != "" {
			//System.Out.Println(fn);
			dc = NewDynaCodeWithRootDia(wd, rootdia, fn)
			//dc.ParamCount = argc;
			this.AddDynaCommand(wd, dc)
			//System.Out.Println("Add command "+wd+" to dialect "+this.Title);
		}
	}

	/* functions */
	kw = ini.ReadString("Definition", "Functions", "")

	for len(kw) > 0 {
		i = utils.Pos(",", kw)
		if i > 0 {
			wd = utils.Copy(kw, 1, i-1)
			kw = utils.Delete(kw, 1, i)
		} else {
			wd = kw
			kw = ""
		}
		/* process keyword */
		fn = ini.ReadString("Function", wd, "")

		argc = 0
		i = utils.Pos(":", fn)
		if i > 0 {
			tmp = utils.Copy(fn, i+1, len(fn)-i)
			fn = utils.Delete(fn, i, len(fn)-i+1)
			argc = utils.StrToInt(tmp)
		}

		if fn != "" {
			//System.Out.Println(fn);
			dc = NewDynaCodeWithRootDia(wd, rootdia, fn)
			dc.ParamCount = argc
			this.AddDynaFunction(wd, dc)
		}
	}

}

func (this *Dialect) IsCloseRBracket(instr string) bool {

	/* vars */
	var result bool

	result = (instr == ")")

	/* enforce non void return */
	return result

}

func (this *Dialect) IsOpenCBrace(instr string) bool {

	/* vars */
	var result bool

	result = (instr == "{")

	/* enforce non void return */
	return result

}

func (this *Dialect) IsOperator(ch string) bool {

	/* vars */
	var result bool

	result = (ch == "+") || (ch == "-") || (ch == "/") || (ch == "*") ||
		(ch == "^") || (ch == "&") || (ch == "|") || (ch == "%") ||
		(ch == ";") || (ch == "@")

	/* enforce non void return */
	return result

}

func (this *Dialect) IsWS(ch rune) bool {

	/* vars */
	var result bool

	result = (ch == '\r') || (ch == '\n') || (ch == '\t') || (ch == ' ')

	/* enforce non void return */
	return result

}

// IsDynaCommand returns true if instr is a dynamic command.
func (this *Dialect) IsDynaCommand(instr string) bool {

	/* vars */
	var result bool

	result = this.DynaCommands.ContainsKey(strings.ToLower(instr))

	/* enforce non void return */
	return result

}

// IsCloseSBracket returns true if instr is a ']'
func (this *Dialect) IsCloseSBracket(instr string) bool {

	/* vars */
	var result bool

	result = (instr == "]")

	/* enforce non void return */
	return result

}

// IsSeparator returns true if ch is a separator
func (this *Dialect) IsSeparator(ch string) bool {

	/* vars */
	var result bool

	result = (ch == ";") || (ch == ",") || (ch == ":")

	/* enforce non void return */
	return result

}

// SplitOnToken splits a token list into a TokenListArray
func (this *Dialect) SplitOnToken(tokens types.TokenList, tok types.Token) types.TokenListArray {

	/* vars */
	result := types.NewTokenListArray()
	var idx int

	idx = 0
	//SetLength(result, idx+1);
	result = result.Add(*types.NewTokenList())

	for _, tt1 := range tokens.Content {
		if (tt1.Type == tok.Type) && (strings.ToLower(tt1.Content) == strings.ToLower(tok.Content)) {
			idx++
			//SetLength(result, idx+1);
			result = result.Add(*types.NewTokenList())
		} else {
			tl := result.Get(idx)
			tl.Push(tt1)
		}
	}

	/* enforce non void return */
	return result

}

func (this *Dialect) NetBracketCount(tokens types.TokenList) int {
	nbc := 0
	for _, t := range tokens.Content {
		if t.Type == types.OBRACKET {
			var v int
			switch t.Content[0] {
			case '(':
				v = 1
				break
			case '[':
				v = 2
				break
			case '{':
				v = 4
				break
			default:
				v = 1
				break
			}
			nbc += v
		} else if t.Type == types.CBRACKET {
			var v int
			switch t.Content[0] {
			case ')':
				v = 1
				break
			case ']':
				v = 2
				break
			case '}':
				v = 4
				break
			default:
				v = 1
				break
			}
			nbc -= v
		} else if t.Type == types.FUNCTION {
			nbc += 1
		} else if t.Type == types.PLUSFUNCTION {
			nbc += 4
		}
	}
	return nbc
}

func (this *Dialect) FindAssignmentSymbol(tokens types.TokenList) int {
	eqidx := -1
	idx := 0
	bc := 0
	for (idx < tokens.Size()) && (eqidx == -1) {
		tt := tokens.Get(idx)
		if tt.Type == types.CBRACKET {
			bc--
		} else if tt.Type == types.OBRACKET {
			bc++
		} else if (tt.Type == types.ASSIGNMENT) && (bc == 0) {
			eqidx = idx
		}
		idx++
	}
	return eqidx
}

func (this *Dialect) IsUserCommand(instr string) bool {

	/* vars */
	var result bool

	result = this.UserCommands.ContainsKey(strings.ToLower(instr))

	/* enforce non void return */
	return result

}

func (this *Dialect) DefineProc(name string, params []string, code string) {

	//dc := dialect.NewDynaCodeWithRootDia()

}

func (this *Dialect) StartProc(name string, params []string, code string) {

	//dc := dialect.NewDynaCodeWithRootDia()

}

func (this *Dialect) RemoveProc(name string) {

	//dc := dialect.NewDynaCodeWithRootDia()
	delete(this.DynaCommands, name)

}

func (this *Dialect) IsLabel(vs string) bool {
	return strings.HasPrefix(vs, "!")
}

func (this *Dialect) IsBreakingCharacter(ch rune, vs string) bool {

	items := " \r\n\t+-/*^([]){}=:,;<>?"

	return (utils.Pos(string(ch), items) > 0)
}

func (this *Dialect) Decolon(code types.Algorithm, start int, increment int, iftosub bool) types.Algorithm {

	code = this.Renumber(code, 100, 100)
	newcode := types.NewAlgorithm()

	highbase := code.GetHighIndex()

	for lno, ln := range code.C {

		if len(ln) > 1 {
			// has colons...

			for i, st := range ln {

				tt := st.LPeek()
				if strings.ToLower(tt.Content) == "if" && tt.Type == types.KEYWORD && iftosub {

					// make a subroutine...
					ifstmt := ln[i]
					rest := ln[i+1:]

					if len(rest) > 0 {

						usegosub := true
						last := rest[len(rest)-1]
						if last.LPeek().Type == types.KEYWORD && strings.ToLower(last.LPeek().Content) == "goto" {
							usegosub = false
						}

						startline := ((highbase + 100) / 100) * 100

						// add dangling bit first
						tla := this.SplitOnToken(*ifstmt.SubList(0, ifstmt.Size()), *types.NewToken(types.KEYWORD, "then"))
						prepart := tla[0]
						postpart := tla[1]

						if postpart.Size() == 1 && (postpart.LPeek().Type == types.NUMBER || postpart.LPeek().Type == types.INTEGER) {
							postpart.UnShift(types.NewToken(types.KEYWORD, "goto"))
							usegosub = false
						}

						tla = tla[2:]
						for _, tailpart := range tla {
							postpart.Content = append(postpart.Content, types.NewToken(types.KEYWORD, "then"))
							postpart.Content = append(postpart.Content, tailpart.Content...)
						}

						prepart.Push(types.NewToken(types.KEYWORD, "then"))

						if usegosub {
							prepart.Push(types.NewToken(types.KEYWORD, "gosub"))
						} else {
							prepart.Push(types.NewToken(types.KEYWORD, "goto"))
						}

						prepart.Push(types.NewToken(types.NUMBER, utils.IntToStr(startline)))

						pp := types.NewStatement()
						for _, ttt := range prepart.Content {
							pp.Push(ttt)
						}

						newcode.C[lno+i] = types.Line{pp}

						pp2 := types.NewStatement()
						for _, ttt := range postpart.Content {
							pp2.Push(ttt)
						}

						// build subroutine
						newcode.C[startline] = types.Line{pp2}
						startline++

						for ii, sst := range rest {
							newcode.C[startline+ii] = types.Line{sst}
						}

						// add return statement
						if usegosub {
							sst := types.NewStatement()
							sst.Push(types.NewToken(types.KEYWORD, "return"))
							newcode.C[startline+len(rest)] = types.Line{sst}
						}

						highbase = startline + 100 // bump to next line block of 100 after

					} else {
						// stay inline
						newcode.C[lno+i] = types.Line{ifstmt}
					}

					break // drop out of loop for now

				} else {
					newcode.C[lno+i] = types.Line{st}
				}

			}

		} else {
			newcode.C[lno] = ln
		}

	}

	// renumber it at the end
	return this.Renumber(*newcode, start, increment)
}

func (this *Dialect) Renumber(code types.Algorithm, start int, increment int) types.Algorithm {

	current := start

	old2new := make(map[int]int)

	l := code.GetLowIndex()
	h := code.GetHighIndex()

	newcode := *types.NewAlgorithm()

	for (l <= h) && (l != -1) {

		nn := current
		on := l

		old2new[on] = nn
		t, _ := code.Get(on)
		newcode.Put(nn, t)

		l = code.NextAfter(l)
		current = current + increment

	}

	// now newcode is keyed by the new line numbers but we need to scan and update tokens
	for _, line := range newcode.C {

		for _, stmt := range line {

			// each statement
			replaceNumber := false
			var lt *types.Token = nil
			for _, t := range stmt.Content {

				if t.Type == types.KEYWORD {

					if strings.ToLower(t.Content) == "goto" || strings.ToLower(t.Content) == "gosub" {
						replaceNumber = true
					}

				} else if (t.Type == types.NUMBER || t.Type == types.INTEGER) && replaceNumber {

					on := t.AsInteger()
					if float64(on) == t.AsExtended() {
						if _, ok := code.Get(on); ok {
							// replace it
							t.Content = utils.IntToStr(old2new[on])
						}
					}

				} else if lt != nil && lt.Type == types.KEYWORD && strings.ToLower(lt.Content) == "then" && (t.Type == types.NUMBER || t.Type == types.INTEGER) {
					on := t.AsInteger()
					if float64(on) == t.AsExtended() {
						if _, ok := code.Get(on); ok {
							// replace it
							t.Content = utils.IntToStr(old2new[on])
						}
					}
				}

				lt = t

			}

		}

	}

	return newcode

}

func (this *Dialect) Reorganize(code types.Algorithm, start int, increment int) types.Algorithm {

	current := start

	old2new := make(map[int]int)
	lenofnew := make(map[int]int)

	l := code.GetLowIndex()
	h := code.GetHighIndex()

	newcode := *types.NewAlgorithm()

	for (l <= h) && (l != -1) {

		nn := current
		on := l

		old2new[on] = nn
		t, _ := code.Get(on)

		lenofnew[nn] = len(t)

		for _, st := range t {
			// each statement gets its own line
			line := types.NewLine()
			line.Push(st)
			newcode.Put(nn, line)
			nn += increment
			current = nn
		}

		//newcode.Put(nn, t)

		l = code.NextAfter(l)
		//current = current + increment

	}

	// now newcode is keyed by the new line numbers but we need to scan and update tokens
	for _, line := range newcode.C {

		for _, stmt := range line {

			// each statement
			replaceNumber := false
			var lt *types.Token = nil
			for _, t := range stmt.Content {

				if t.Type == types.KEYWORD {

					if t.Content == "goto" || t.Content == "gosub" {
						replaceNumber = true
					}

				} else if (t.Type == types.NUMBER || t.Type == types.INTEGER) && replaceNumber {

					on := t.AsInteger()
					if float64(on) == t.AsExtended() {
						if _, ok := code.Get(on); ok {
							// replace it
							t.Content = utils.IntToStr(old2new[on])
						}
					}

				} else if lt != nil && lt.Type == types.KEYWORD && lt.Content == "then" && (t.Type == types.NUMBER || t.Type == types.INTEGER) {
					on := t.AsInteger()
					if float64(on) == t.AsExtended() {
						if _, ok := code.Get(on); ok {
							// replace it
							t.Content = utils.IntToStr(old2new[on])
						}
					}
				}

				lt = t

			}

		}

	}

	return newcode

}

func (this *Dialect) PutStr(ent interfaces.Interpretable, s string) {

}

func (this *Dialect) RealPut(ent interfaces.Interpretable, ch rune) {

}

func (this *Dialect) Backspace(ent interfaces.Interpretable) {

}

func (this *Dialect) ClearToBottom(ent interfaces.Interpretable) {
	//
}

func (this *Dialect) Repos(ent interfaces.Interpretable) {
	//
}

func (this *Dialect) SetCursorX(ent interfaces.Interpretable, x int) {
	//
}

func (this *Dialect) SetCursorY(ent interfaces.Interpretable, y int) {
	//
}

func (this *Dialect) GetColumns(ent interfaces.Interpretable) int {
	return 40
}

func (this *Dialect) GetRows(ent interfaces.Interpretable) int {
	return 24
}

func (this *Dialect) GetCursorX(ent interfaces.Interpretable) int {
	return 0
}

func (this *Dialect) GetCursorY(ent interfaces.Interpretable) int {
	return 2
}

func (this *Dialect) HomeLeft(ent interfaces.Interpretable) {
	// stub
}

func (this *Dialect) UpdateRuntimeState(ent interfaces.Interpretable) {

}

func (this *Dialect) GetRealCursorPos(ent interfaces.Interpretable) (int, int) {
	return apple2helpers.GetRealCursorPos(ent)
}

func (this *Dialect) GetRealWindow(ent interfaces.Interpretable) (int, int, int, int, int, int) {
	return apple2helpers.GetRealWindow(ent)
}

func (this *Dialect) SetRealCursorPos(ent interfaces.Interpretable, x, y int) {
	apple2helpers.SetRealCursorPos(ent, x, y)
}

func (this *Dialect) CleanDynaCommands() {
	this.DynaCommands = make(DynaMap)
	this.DynaFunctions = make(DynaMap)
}

func (this *Dialect) CleanDynaCommandsByName(names []string) {
	for _, n := range names {
		if _, ok := this.DynaCommands[strings.ToLower(n)]; ok {
			delete(this.DynaCommands, strings.ToLower(n))
		}
		if _, ok := this.DynaFunctions[strings.ToLower(n)]; ok {
			delete(this.DynaFunctions, strings.ToLower(n))
		}
	}
}

func (this *Dialect) HideDynaCommandsByName(names []string) {
	for _, n := range names {
		if dc, ok := this.DynaCommands[strings.ToLower(n)]; ok {
			dc.SetHidden(true)
		}
		if dc, ok := this.DynaFunctions[strings.ToLower(n)]; ok {
			dc.SetHidden(true)
		}
	}
}

func (this *Dialect) UnhideDynaCommandsByName(names []string) {
	for _, n := range names {
		if dc, ok := this.DynaCommands[strings.ToLower(n)]; ok {
			dc.SetHidden(false)
		}
		if dc, ok := this.DynaFunctions[strings.ToLower(n)]; ok {
			dc.SetHidden(false)
		}
	}
}

func (this *Dialect) GetDynaCommand(name string) interfaces.DynaCoder {
	return this.DynaCommands[strings.ToLower(name)]
}

func (this *Dialect) GetDynaFunction(name string) interfaces.DynaCoder {
	return this.DynaFunctions[strings.ToLower(name)]
}

func (this *Dialect) GetCompletions(ent interfaces.Interpretable, line runestring.RuneString, index int) (int, *types.TokenList) {
	return 0, types.NewTokenList()
}

func (this *Dialect) GetWorkspaceBody(vars bool, filterProc string) []string {
	return []string{}
}

func (this *Dialect) QueueCommand(command string) {
	//
}

func (this *Dialect) SaveState() {

}

func (this *Dialect) RestoreState() {

}

func (this *Dialect) SyntaxValid(s string) error {
	return nil
}

func (this *Dialect) GetLastCommand() string {
	return ""
}
