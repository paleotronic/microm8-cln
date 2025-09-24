package logo

import (
	// rlog "log"
	"errors"
	"regexp"
	"sort"
	"strings"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/dialect/appleinteger"
	"paleotronic.com/core/dialect/applesoft"
	"paleotronic.com/core/dialect/parcel"
	"paleotronic.com/core/exception"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/files"
	"paleotronic.com/fmt"
	"paleotronic.com/log"
	"paleotronic.com/runestring"
	"paleotronic.com/utils"
)

// TODO: Put this in config file.
var LOMEM = 24576

const (
	LOOPSTACK_ADDRESS  = 0xff00
	CALLSTACK_ADDRESS  = 0xfe00
	LOOPSTACK_MAX      = 255
	CALLSTACK_MAX      = 255
	LOOP_PEEK_FREQ     = 86
	ADDSUB_PEEK_FREQ   = 66
	ADDSUB_PEEK_MS_PER = 8
	LOOP_PEEK_MS_PER   = 6
	MEMTOP             = 38400
	MEMBOT             = 2049
)

var reSimpleLoop *regexp.Regexp
var reAddSub *regexp.Regexp
var reNumberFloatE *regexp.Regexp
var reNumberFloatES *regexp.Regexp
var reVarname *regexp.Regexp

func init() {
	reSimpleLoop, _ = regexp.Compile("(.* )?for ([A-Za-z]+) = ([0-9]+) to ([0-9]+) : ([A-Za-z]+) = PEEK[(] - 16336 [)] : next( [A-Za-z]+)?( :.*)?")
	reAddSub, _ = regexp.Compile("(.* : )?([A-Za-z]+) = (PEEK[(] - 16336 [)])( [+-] PEEK[(] - 16336 [)])*( :.*)?")
	reNumberFloatE, _ = regexp.Compile("^[+-]?([0-9]+)([.]([0-9]+))?[eE]$")
	reNumberFloatES, _ = regexp.Compile("^[+-]?([0-9]+)([.]([0-9]+))?[eE][+-]$")
	reVarname, _ = regexp.Compile("^[:\"][a-zA-Z][a-zA-Z0-9]*[?]?$")
}

type DialectLogo struct {
	dialect.Dialect
	lastLineNumber   int
	DefineProcedure  string
	DefineName       string
	DefineParams     []string
	DefineMode       bool
	CurrentProcedure string
	Driver           *LogoDriver
	Lexer            *parcel.Lexer
	LastCommand      string
	SuppressError    bool
	OldPrompt        string
}

func NewDialectLogo() *DialectLogo {
	this := &DialectLogo{}
	this.Dialect = *dialect.NewDialect()
	this.Init()
	this.Dialect.DefaultCost = 1000000000 / 3200
	this.Throttle = 100.0
	this.GenerateNumericTokens()
	this.Driver = NewLogoDriver(this)
	this.Lexer = parcel.NewLexer(this.Driver)
	this.Lexer.LoadString(GetSpec(this))
	return this
}

func (this *DialectLogo) IsSeparator(ch string) bool {

	/* vars */
	var result bool

	result = (ch == ";") || (ch == ",")

	/* enforce non void return */
	return result

}

func (this *DialectLogo) CheckOptimize(lno int, s string, OCode types.Algorithm) {
	// stub does nothing

	//fmt.Println("in applesoft match")

	if m := reSimpleLoop.FindStringSubmatch(s); len(m) > 0 {
		// (...:)? FOR ([A-Z]+) = ([0-9]+) TO ([0-9]+) : ([A-Z]+) = PEEK[(] -16336 [)] : NEXT( [A-Z]+)?( :.*)?
		//   1           2          3          4          5                                   6        7
		start := utils.StrToInt(m[3])
		end := utils.StrToInt(m[4])

		total_peeks := end - start + 1
		duration_ms := total_peeks * LOOP_PEEK_MS_PER
		freq_hz := LOOP_PEEK_FREQ

		alt_cmd := "audio.tone{" + utils.IntToStr(freq_hz) + "," + utils.IntToStr(duration_ms) + "}"
		rtl := this.Tokenize(runestring.Cast(alt_cmd))
		ln := types.NewLine()
		st := types.NewStatement()

		var joinLast bool

		if m[1] != "" {
			ss := m[1]

			ftl := *this.Tokenize(runestring.Cast(ss))

			if ftl.RPeek().Type == types.KEYWORD && strings.ToLower(ftl.RPeek().Content) == "then" {
				joinLast = true
			}

			tla := this.SplitOnToken(ftl, *types.NewToken(types.SEPARATOR, ":"))

			for i, tl := range tla {

				if tl.Size() == 0 {
					continue
				}

				if i == len(tla)-1 && joinLast {

					for _, tt := range rtl.Content {
						tl.Push(tt)
					}
					*rtl = tl

				} else {
					st = types.NewStatement()
					for _, t := range tl.Content {
						st.Push(t)
					}
					ln = append(ln, st)
				}
			}

			st = types.NewStatement()

		}

		for _, t := range rtl.Content {
			st.Push(t)
		}

		ln = append(ln, st)

		// final part
		if m[7] != "" {
			ss := m[7][3:]

			ftl := *this.Tokenize(runestring.Cast(ss))
			tla := this.SplitOnToken(ftl, *types.NewToken(types.SEPARATOR, ":"))

			for _, tl := range tla {
				st = types.NewStatement()

				for _, t := range tl.Content {
					st.Push(t)
				}

				ln = append(ln, st)
			}

		}

		// Add new version
		OCode.Put(lno, ln)

		//fmt.Println("[DEBUG] Original code     :", lno, s)
		//fmt.Println("[DEBUG] Optimizer suggests:", lno, ln.String())

	} else if m = reAddSub.FindStringSubmatch(s); len(m) > 0 {
		// (.+ : )?([A-Z]+) = (PEEK[(] -16336 [)])( [+-] (PEEK[(] -16336 [)]))*( :.*)?
		//    1       2                3              4             5             6
		//fmt.Printf("All submatches: %v\n", m)

		//		for ii, ss := range m {
		//fmt.Printf("%d) %s\n", ii, ss)
		//		}

		total_peeks := strings.Count(s, "PEEK( -16336 )")
		duration_ms := total_peeks * ADDSUB_PEEK_MS_PER
		freq_hz := ADDSUB_PEEK_FREQ

		alt_cmd := "audio.tone{" + utils.IntToStr(freq_hz) + "," + utils.IntToStr(duration_ms) + "}"
		tl := this.Tokenize(runestring.Cast(alt_cmd))
		ln := types.NewLine()
		st := types.NewStatement()

		if m[1] != "" {
			ss := m[1][0 : len(m[1])-3]

			ftl := *this.Tokenize(runestring.Cast(ss))
			tla := this.SplitOnToken(ftl, *types.NewToken(types.SEPARATOR, ":"))

			for _, tl := range tla {
				st = types.NewStatement()

				for _, t := range tl.Content {
					st.Push(t)
				}

				ln = append(ln, st)
			}

			st = types.NewStatement()

		}

		for _, t := range tl.Content {
			st.Push(t)
		}

		ln = append(ln, st)

		//
		if m[5] != "" {
			ss := m[5][3:]

			ftl := *this.Tokenize(runestring.Cast(ss))
			tla := this.SplitOnToken(ftl, *types.NewToken(types.SEPARATOR, ":"))

			for _, tl := range tla {
				st = types.NewStatement()

				for _, t := range tl.Content {
					st.Push(t)
				}

				ln = append(ln, st)
			}

		}

		// Add new version
		OCode.Put(lno, ln)

		//fmt.Println("Original code     :", lno, s)
		//fmt.Println("Optimizer suggests:", lno, ln.String())

	}
}

func (this *DialectLogo) Evaluate(chunk string, tokens *types.TokenList) bool {

	// rlog.Printf("Evaluate called for [%s]", chunk)

	/* vars */
	var result bool
	var tok types.Token
	var ptok *types.Token

	result = true // continue

	if len(chunk) == 0 {
		return result
	}

	tok = *types.NewToken(types.INVALID, "")

	if tokens.Size() > 0 {
		ptok = tokens.RPeek()
	} else {
		ptok = nil
	}

	if this.IsFunction(chunk) {
		tok.Type = types.FUNCTION
		tok.Content = strings.ToTitle(chunk)
	} else if this.IsDynaCommand(chunk) {
		if tokens.Size() > 0 && tokens.RPeek().Type == types.LOGOVARQUOTE {
			tokens.RPeek().Type = types.STRING
			tokens.RPeek().Content = chunk
		} else {
			tok.Type = types.DYNAMICKEYWORD
			tok.Content = chunk
		}
	} else if this.IsDynaFunction(chunk) {
		tok.Type = types.DYNAMICFUNCTION
		tok.Content = strings.ToTitle(chunk)
	} else if this.IsKeyword(chunk) {

		/* determine if (we should stop parsing */

		if tokens.Size() > 0 && tokens.RPeek().Type == types.LOGOVARQUOTE {
			tokens.RPeek().Type = types.STRING
			tokens.RPeek().Content = chunk
		} else {
			tok.Type = types.KEYWORD
			tok.Content = chunk
			z := this.Commands[strings.ToLower(chunk)]
			result = !(z.HasNoTokens())
		}

		// } else if this.IsLogic(chunk) {
		// 	tok.Type = types.LOGIC
		// 	tok.Content = strings.ToLower(chunk)
	} else if this.IsFloat(chunk) {
		if (tokens.Size() > 1) && (tokens.Get(tokens.Size()-1).Type == types.OPERATOR) && (tokens.Get(tokens.Size()-2).Type == types.OPERATOR) {
			tokens.RPeek().Content = tokens.RPeek().Content + chunk
			tokens.RPeek().Type = types.NUMBER
		} else {

			if (ptok != nil) && (reNumberFloatES.MatchString(ptok.Content)) {
				ptok.Content += chunk
			} else {
				tok.Type = types.NUMBER
				tok.Content = chunk
			}
		}
	} else if this.IsInteger(chunk) {
		if (tokens.Size() > 1) && (tokens.Get(tokens.Size()-1).Type == types.OPERATOR) && (tokens.Get(tokens.Size()-2).Type == types.OPERATOR) {
			tokens.RPeek().Content = tokens.RPeek().Content + chunk
			tokens.RPeek().Type = types.INTEGER
		} else {
			if (ptok != nil) && (reNumberFloatES.MatchString(ptok.Content)) {
				ptok.Content += chunk
			} else {
				tok.Type = types.INTEGER
				tok.Content = chunk
			}
		}
	} else if this.IsSeparator(chunk) {
		if chunk == ":" && (tokens.Size() > 0) && (strings.ToLower(tokens.RPeek().Content) == "himem" || strings.ToLower(tokens.RPeek().Content) == "lomem") {
			tokens.RPeek().Content += ":"
			tokens.RPeek().Type = types.KEYWORD
		} else {
			tok.Type = types.SEPARATOR
			tok.Content = chunk
		}
	} else if this.IsType(chunk) {
		tok.Type = types.TYPE
		tok.Content = chunk
	} else if this.IsOpenSBracket(chunk) {
		tok.Type = types.LIST
		tok.Content = ""
		tok.List = types.NewTokenList()
		tok.List.Open = true
	} else if this.IsCloseSBracket(chunk) {
		tokens.Open = false
	} else if this.IsOpenRBracket(chunk) || this.IsOpenSBracket(chunk) || this.IsOpenCBrace(chunk) {
		ptok = tokens.RPeek()
		if (tokens.Size() >= 1) && (ptok != nil) && (ptok.Type == types.VARIABLE) {
			ss := strings.ToLower(ptok.Content + chunk)
			if this.Functions[ss] != nil {
				ptok.Content = ptok.Content + chunk
				ptok.Type = types.FUNCTION
			} else if _, ex, _, _ := this.PlusFunctions.GetFunctionByNameContext(this.CurrentNameSpace, ss); ex {
				ptok.Content = ptok.Content + chunk
				ptok.Type = types.PLUSFUNCTION
			} else {
				tok.Type = types.OBRACKET
				tok.Content = chunk
			}
		} else {
			tok.Type = types.OBRACKET
			tok.Content = chunk
		}
	} else if this.IsCloseRBracket(chunk) || this.IsCloseSBracket(chunk) || this.IsCloseCBrace(chunk) {
		tok.Type = types.CBRACKET
		tok.Content = chunk
	} else if this.IsOperator(chunk) {

		/* Fix for -/+  */
		if (chunk == "+" || chunk == "-") && (ptok != nil) && (reNumberFloatE.MatchString(ptok.Content)) {
			// part of a continuing number
			ptok.Type = types.NUMBER
			ptok.Content = ptok.Content + chunk
		} else {
			tok.Type = types.OPERATOR
			tok.Content = chunk
		}

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
			posscmd := strings.ToLower(ptok.Content + chunk)
			if this.IsKeyword(posscmd) {
				ptok.Content = ptok.Content + chunk
				ptok.Type = types.KEYWORD
			} else if (ptok.Type == types.COMPARITOR) && (chunk == "=") {
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
		tok.Content = strings.ToUpper(chunk)
	} else if this.IsPlusVariableName(chunk) {
		tok.Type = types.PLUSVAR
		tok.Content = chunk
	} else if this.IsLogoParamPrefix(chunk) {
		tok.Type = types.LOGOPARAMPREFIX
		tok.Content = chunk
	} else if this.IsLogoVarQuote(chunk) {
		tok.Type = types.LOGOVARQUOTE
		tok.Content = chunk
	} else {
		if tokens.Size() > 0 && tokens.RPeek().Type == types.LOGOPARAMPREFIX {
			tokens.RPeek().Type = types.VARIABLE
			tokens.RPeek().Content = tokens.RPeek().Content + chunk
		} else if tokens.Size() > 0 && tokens.RPeek().Type == types.LOGOVARQUOTE {
			tokens.RPeek().Type = types.STRING
			tokens.RPeek().Content = chunk
		} else {
			tok.Type = types.DYNAMICFUNCTION
			tok.Content = chunk
		}
	}

	if tok.Type != types.INVALID {
		//System.Out.Println("ADD: ", tok.Content);
		//fmt.Printf("*** Yielded token: %s %s\n", tok.Content, tok.Type.String())
		/* shim for -No */
		if (tok.Type == types.NUMBER) && (tokens.Size() >= 2) {
			if (tokens.RPeek().Type == types.OPERATOR) &&
				(tokens.RPeek().Content == "-") &&
				((tokens.Get(tokens.Size()-2).Type == types.ASSIGNMENT) || (tokens.Get(tokens.Size()-2).Type == types.COMPARITOR)) {
				/* merge into last token, change to number type */
				tokens.Get(tokens.Size() - 1).Type = types.NUMBER
				tokens.Get(tokens.Size() - 1).Content = "-" + tok.Content
			} else {
				tokens.Push(&tok)
			}
		} else {
			tokens.Push(&tok)
		}
	}

	/* enforce non void return */
	return result

}

func (this *DialectLogo) IsVariableName(instr string) bool {

	return reVarname.MatchString(instr)

}

func (this *DialectLogo) IsLogoParamPrefix(instr string) bool {

	return (rune(instr[0]) == ':')

}

func (this *DialectLogo) IsLogoVarQuote(instr string) bool {

	return (rune(instr[0]) == '"')

}

func (this *DialectLogo) IsBreakingCharacter(ch rune, vs string) bool {

	items := " \r\n\t+*^([]){}=\",;<>"

	return (utils.Pos(string(ch), items) > 0)
}

func (this *DialectLogo) IsVerb(instr string) bool {

	return this.IsKeyword(instr) || this.IsUserCommand(instr)

}

func (this *DialectLogo) Tokenize(s runestring.RuneString) *types.TokenList {

	/* vars */
	var result *types.TokenList
	var inq bool
	var inqq bool
	var cont bool
	var idx int
	var chunk string
	var ch rune

	result = types.NewTokenList()

	inq = false
	inqq = false
	idx = 0
	chunk = ""
	cont = true

	tlstack := make([]*types.TokenList, 0)

	for idx < len(s.Runes) {
		ch = s.Runes[idx]

		//System.Out.Println("Tokenizer sees ["+ch+"]");

		if this.IsWS(ch) && (inq || inqq) {
			chunk = chunk + string(ch)
		} else if this.IsVarSuffix(ch, this.VarSuffixes) && (!(inq || inqq)) {
			chunk = chunk + string(ch)
			if len(chunk) > 0 {
				cont = this.Evaluate(chunk, result)
				//				pchunk = chunk
				chunk = ""
			}
		} else if this.IsBreakingCharacter(ch, this.VarSuffixes) && (!(inq || inqq)) {

			//System.Out.Println("====== breaking char "+ch+" with chunk ["+chunk+"]");

			/* special handling for x.Yyyye+/-xx notation */
			if ((ch == '+') || (ch == '-')) &&
				((len(chunk) >= 2) &&
					((chunk[len(chunk)-1] == 'e') || (chunk[len(chunk)-1] == 'E')) &&
					(this.IsDigit(rune(chunk[0]))) &&
					(this.IsDigit(rune(chunk[len(chunk)-2])))) {
				chunk = chunk + string(ch)
			} else {
				if len(chunk) > 0 {
					cc := result.Size()
					oc := ""
					if result.Size() > 0 {
						oc = result.RPeek().Content
					}
					cont = this.Evaluate(chunk, result)
					if (result.Size() > cc) || (result.RPeek().Content != oc) || (!this.IsWS(ch)) {
						//						pchunk = chunk
						chunk = ""
					} else {
						//System.Err.Println("Line: "+s)
						//System.Err.Println("Keeping chunk for next cycle: "+chunk)
					}
				}

				if !this.IsWS(ch) {
					chunk = chunk + string(ch)
					cont = this.Evaluate(chunk, result)
					//					pchunk = chunk
					chunk = ""
				}

				if result.Size() > 0 && result.RPeek().Type == types.LIST && result.RPeek().List.IsOpen() {
					// list has opened
					tlstack = append(tlstack, result)
					result = result.RPeek().List
				} else if len(tlstack) > 0 && result.IsClosed() {
					result = tlstack[len(tlstack)-1]
					tlstack = tlstack[0 : len(tlstack)-1]
				}

			}

			//		} else if this.IsQ(ch) && (!inqq) {
			//			if (len(chunk) > 0) && (!inq) {
			//				this.Evaluate(chunk, result)
			//				//				pchunk = chunk
			//				chunk = ""
			//			}
			//			inq = !inq
			//			chunk = chunk + string(ch)
			//		} else if this.IsQQ(ch) && (!inq) {
			//			if (len(chunk) > 0) && (!inqq) {
			//				this.Evaluate(chunk, result)
			//				chunk = "\""
			//			} else if (len(chunk) > 0) && (inqq) {
			//				chunk = chunk + "\""
			//				this.Evaluate(chunk, result)
			//				//				pchunk = chunk
			//				chunk = ""
			//			} else {
			//				chunk = chunk + string(ch)
			//			}
			//			inqq = !inqq
		} else {
			chunk = chunk + string(ch)

			/* break keywords out early */
			//			if this.Commands.ContainsKey(strings.ToLower(chunk)) {
			//				if (strings.ToLower(chunk) != "go") && (strings.ToLower(chunk) != "to") && (strings.ToLower(chunk) != "on") && (strings.ToLower(chunk) != "hgr") && (strings.ToLower(chunk) != "at") {
			//					cont = this.Evaluate(chunk, result)
			//					//					pchunk = chunk
			//					chunk = ""
			//				}
			//			}

		}

		idx++

		if !cont {
			//			pchunk = chunk
			chunk = ""

			for (idx < len(s.Runes)) && ((inqq || (s.Runes[idx] != ':')) || (strings.ToLower(result.RPeek().Content) == "rem")) {
				chunk = chunk + string(s.Runes[idx])
				if s.Runes[idx] == '"' {
					inqq = !inqq
				}
				idx++
			}

			ttt := types.NewToken(types.UNSTRING, strings.Trim(chunk, " "))
			//pchunk = chunk
			chunk = ""
			result.Push(ttt)
		}

	} /*while*/

	//System.Out.Println("chunk == ", chunk;

	if len(chunk) > 0 {
		if inqq {
			chunk = chunk + "\""
		}
		this.Evaluate(chunk, result)
		//pchunk = chunk
		chunk = ""
	}

	/* enforce non void return */
	return result

}

func (this *DialectLogo) DecideValueIndex(hpop int, count int, values types.TokenList) int {
	// decide the lowest index to stake from

	// 5 6 7
	// + *
	// 0 1 2

	// low end
	if hpop == 0 {
		return 0
	}

	// high end
	if hpop+count-1 >= values.Size() {
		return values.Size() - count
	}

	// leaning ... between 1 and size - 2
	//if (ops.Size()+1 == values.Size()-1) {
	//	return hpop;
	//}

	return hpop
}

func (this *DialectLogo) HPOpIndex(tl types.TokenList) int {
	result := -1
	hs := 1
	var tt types.Token
	var sc int

	for i := 0; i <= tl.Size()-1; i++ {
		tt = *tl.Get(i)
		sc = 0

		if tt.Type == types.FUNCTION {
			sc = 600 + i // Functions highest precedence
		} else if tt.Content == "^" {
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
				sc = 600
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

func isTokenNumeric(t *types.Token) bool {
	if t.Type != types.LIST {
		return false
	}
	return utils.StrToFloat64(t.Content) != 0
}

func (this *DialectLogo) ParseTokensForResult(ent interfaces.Interpretable, tokens types.TokenList) (*types.Token, error) {
	this.Driver.ent = ent
	t, err := this.Driver.ParseExprRLCollapse(&tokens, false)
	if err != nil {
		return nil, err
	}
	if len(t) == 1 {
		return t[0], err
	}
	tt := types.NewToken(types.LIST, "")
	tt.List = types.NewTokenList()
	tt.List.Content = t
	//tt, err := this.Driver.ParseExprRLCollapse(&tokens)
	return tt, nil
}

func (this *DialectLogo) HandleException(ent interfaces.Interpretable, e error) {

	if this.SuppressError {
		return
	}

	var msg string

	apple2helpers.NLIN(ent)

	msg = e.Error()

	if (ent.GetState() == types.RUNNING) || (ent.GetState() == types.DIRECTRUNNING) || (ent.GetState() == types.STOPPED) {
		if !ent.HandleError() {
			apple2helpers.PutStr(ent, strings.ToUpper(msg))
			proc, line := this.Driver.CurrentProc()
			if proc != "" && !strings.HasPrefix(proc, "_") {
				apple2helpers.PutStr(ent, strings.ToUpper(" IN "+proc+" ("+utils.IntToStr(line)+")"))
			}
			ent.Halt()
		}
	}

	apple2helpers.Beep(ent)

	apple2helpers.PutStr(ent, "\r\n")

}

func (this *DialectLogo) Parse(ent interfaces.Interpretable, s string) error {
	//this.Driver.Printf("################ Parsing line: %s", s)
	this.LastCommand = strings.ToLower(strings.Trim(s, " "))

	this.Driver.ent = ent

	this.Driver.Reset()

	if this.Driver.PendingProcName != "" {
		//this.Driver.Printf("###################### PENDING PROC DEFINITION: %s", s)
		if strings.ToLower(strings.Trim(s, " ")) != "end" {
			this.Driver.PendingProcStatements = append(this.Driver.PendingProcStatements, strings.Trim(s, " "))
			return nil
		}
	}

	this.Driver.ReresolveSymbols(this.Lexer)
	err := this.Driver.CreateCommandScope(this.Lexer, []string{s})
	if err != nil {
		this.HandleException(ent, err)
		return err
	}

	_, err = this.Driver.ExecTillReturn()
	if err != nil {
		this.HandleException(ent, err)
		return err
	}

	return err
}

func (this *DialectLogo) ParseOld(ent interfaces.Interpretable, s string) error {

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

	if tl.Size() == 0 {
		return nil
	}

	tok := tl.Get(0)

	if tok.Type == types.SEPARATOR && tok.Content == ":" {

		if tl.Size() > 0 {

			// fix for immediate only commands
			for i := 0; i < tl.Size(); i++ {
				if tl.Get(i).Type == types.KEYWORD {
					cmd := this.Commands[strings.ToLower(tl.Get(i).Content)]
					if cmd.ImmediateModeOnly() {
						tl.Get(i).Type = types.VARIABLE
					}
				}
			}

			cmdlist = ent.SplitOnToken(*tl, *types.NewToken(types.SEPARATOR, ":"))

			//			////fmt.Println(len(cmdlist))

			lno = this.lastLineNumber

			ll, exists := ent.GetCode().Get(lno)

			if !exists {
				ll = types.NewLine()
			}

			for _, cl := range cmdlist {
				if cl.Size() > 0 {
					st = types.NewStatement()
					for _, tok1 := range cl.Content {
						st.Push(tok1)
					}
					ll = append(ll, st)
				}
			}

			z := ent.GetCode()
			z.Put(lno, ll)
			this.lastLineNumber = lno

			//ent.PutStr("+" + utils.IntToStr(lno) + "\r\n")
			ent.SetState(types.STOPPED)
		}

		//data := this.GetMemoryRepresentation(ent.GetCode())

		if !this.SkipMemParse() {
			data := this.GetMemoryRepresentation(ent.GetCode())

			MEMBASE := this.GetProgramStart(ent)

			// write to memory
			for i, v := range data {
				ent.SetMemory(MEMBASE+1+i, v)
			}
			ent.SetMemory(MEMBASE, uint64(MEMBASE+1+len(data)))

			//			fixMemoryPtrs(ent)
		}

		//        types.RebuildAlgo()

		return nil

	} else if tok.Type == types.NUMBER {

		tok = tl.Shift()

		if tl.Size() > 0 {

			// fix for immediate only commands
			for i := 0; i < tl.Size(); i++ {
				if tl.Get(i).Type == types.KEYWORD {
					cmd := this.Commands[strings.ToLower(tl.Get(i).Content)]
					if cmd.ImmediateModeOnly() {
						tl.Get(i).Type = types.VARIABLE
					}
				}
			}

			cmdlist = ent.SplitOnToken(*tl, *types.NewToken(types.SEPARATOR, ":"))

			//			////fmt.Println(len(cmdlist))

			lno = tok.AsInteger()
			ll = types.NewLine()

			for _, cl := range cmdlist {
				if cl.Size() > 0 {
					st = types.NewStatement()
					for _, tok1 := range cl.Content {
						st.Push(tok1)
					}
					ll = append(ll, st)
				} /*else {
					st = types.NewStatement()
					st.Push(types.NewToken(types.KEYWORD, "REM"))
					ll = append(ll, st)
				}*/
			}

			z := ent.GetCode()
			z.Put(lno, ll)
			this.lastLineNumber = lno

			//ent.PutStr("+" + utils.IntToStr(lno) + "\r\n")
			ent.SetState(types.STOPPED)
			//fInterpreter.VDU.Dump;
		} else {
			lno = tok.AsInteger()
			if _, ok := ent.GetCode().Get(lno); ok {
				z := ent.GetCode()
				z.Remove(lno)
				//ent.PutStr("-"+PasUtil.IntToStr(lno)+"\r\n");
			}
		}

		if !this.SkipMemParse() {
			data := this.GetMemoryRepresentation(ent.GetCode())

			// write to memory
			MEMBASE := this.GetProgramStart(ent)

			// write to memory
			for i, v := range data {
				ent.SetMemory(MEMBASE+1+i, v)
			}
			ent.SetMemory(MEMBASE, uint64(MEMBASE+1+len(data)))
			//			fixMemoryPtrs(ent)
		}

		//        types.RebuildAlgo()

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
	a := types.NewAlgorithm()
	a.Put(lno, ll)
	ent.SetDirectAlgorithm(a)
	ent.GetLPC().Line = lno
	ent.GetLPC().Statement = 0
	ent.GetLPC().Token = 0

	//	//fmt.Printf("Setting up direct run for code [%s]\n", ent.GetName())

	//a := ent.GetDirectAlgorithm()
	//ent.PutStr(fmt.Sprintf("Direct has %d lines\r\n", len(a)))

	//str(ent.GetState(), n);
	//fInterpreter.VDU.PutStr(n+": run direct with "+IntToStr(ll.Size())+" commands: "+s+"\r\n");

	// start running
	ent.SetState(types.DIRECTRUNNING)

	// back resolve
	if strings.Trim(strings.ToLower(s), " ") == "end" {
		// 	// back propagate
		// lines := make([]string, 0)
		// for _, procname := range this.GetDynamicFunctions() {
		// 	// procname
		// 	def := this.GetDynamicFunctionDef(procname)
		// 	if len(lines) > 0 {
		// 		lines = append(lines, "")
		// 	}
		// 	lines = append(lines, def...)
		// }
		// lines = append(lines, "")
		// for _, procname := range this.GetDynamicCommands() {
		// 	// procname
		// 	def := this.GetDynamicCommandDef(procname)
		// 	if len(lines) > 0 {
		// 		lines = append(lines, "")
		// 	}
		// 	lines = append(lines, def...)
		// }

		// for _, l := range lines {
		// 	if l != "" {
		// 		fmt.Printf("*** DEF DUMP: %s\n", l)
		// 		tl := this.Tokenize(runestring.Cast(l))
		// 		scope := ent.GetDirectAlgorithm()
		// 		this.SetSilentDefines(true)
		// 		this.ExecuteDirectCommand(*tl, ent, scope, ent.GetLPC())
		// 		this.SetSilentDefines(false)
		// 	}
		// }

	}

	//ent.PutStr(fmt.Sprintf("Interpreter state is now %v\r\n", ent.GetState()))

	return nil

}

func (this *DialectLogo) InitVDU(v interfaces.Interpretable, promptonly bool) {

	/* vars */

	apple2helpers.TEXT40(v)

	v.SetPrompt("?")
	v.SetTabWidth(16)

	if !promptonly {
		apple2helpers.Clearscreen(v)
		apple2helpers.Gotoxy(v, 0, 0)
		apple2helpers.PutStr(v, "\r\n")
		apple2helpers.PutStr(v, "Welcome to microLOGO!\r\n")
		apple2helpers.PutStr(v, " _  .----.\r\n")
		apple2helpers.PutStr(v, "(_\\/      \\_,\r\n")
		apple2helpers.PutStr(v, "  'uu----uu~'\r\n")

		advert := `
Psst... hey, want to see what a
Logo turtle can REALLY do?

Check out turtleSpaces, a full 3D
implementation of Apple Logo at
https://turtlespaces.org

Our turtles will amaze you!

Now back to your regularly
scheduled Logo...

`

		advert = strings.Join(strings.Split(advert, "\n"), "\r\n")

		apple2helpers.SetFGColor(v, 13)
		apple2helpers.PutStr(v, advert)
		apple2helpers.SetFGColor(v, 15)

		//		//fmt.Printf( "CX = %d, CY = %d\n", v.GetCursorX(), v.GetCursorY() )
	}

	v.SaveCPOS()

	//v.CreateVar(
	//	"speed",
	//	*types.NewVariableP("speed", types.VT_FLOAT, "255", true),
	//)
	settings.SlotZPEmu[v.GetMemIndex()] = true

	v.SetMemory(228, 255)

	// set program pointers
	v.SetMemory(103, 1)
	v.SetMemory(104, 8)

	settings.SpecName[v.GetMemIndex()] = "microLOGO"
	settings.SetSubtitle(settings.SpecName[v.GetMemIndex()])

	v.GetMemoryMap().WriteGlobal(v.GetMemIndex(), v.GetMemoryMap().MEMBASE(v.GetMemIndex())+memory.OCTALYZER_CAMERA_GFX_BASE+0, uint64(types.CC_ResetAll))

	v.SetNeedsPrompt(true)

	//settings.LogoCameraControl[v.GetMemIndex()] = true

}

func (this *DialectLogo) ProcessDynamicCommand(ent interfaces.Interpretable, cmd string) error {

	//	////fmt.Printf("In ProcessDynamicCommand for [%s]\n", cmd)

	if utils.Copy(cmd, 1, 5) == "BLOAD" {

		parts := strings.Split(strings.Trim(utils.Delete(cmd, 1, 5), " "), ",")
		addr := 16384
		length := -1

		filename := parts[0]
		filename = strings.ToLower(filename)
		//System.Out.Println("will try to load "+filename);
		for i := 1; i < len(parts); i++ {
			tmp := strings.ToUpper(parts[i])
			if tmp[0] == 'A' {
				tmp = utils.Delete(tmp, 1, 1)
				addr = utils.StrToInt(tmp)
			} else if tmp[0] == 'L' {
				tmp = utils.Delete(tmp, 1, 1)
				length = utils.StrToInt(tmp)
			}
		}

		if !strings.HasSuffix(filename, ".s") {
			filename += ".s"
		}

		if !files.ExistsViaProvider(ent.GetWorkDir(), filename) {
			return exception.NewESyntaxError("FILE NOT FOUND")
		}

		data, err := files.ReadBytesViaProvider(ent.GetWorkDir(), filename)
		if err != nil {
			return err
		}

		rl := len(data.Content)
		if (length != -1) && (length <= rl) {
			rl = length
		}
		//System.Out.Println("BLOAD A="+addr+", L="+rl);
		rawdata := make([]uint64, rl)
		for i := 0; i < rl; i++ {
			//ent.SetMemory((addr+i)%65536, uint64(data.Content[i]&0xff))
			rawdata[i] = uint64(data.Content[i])
		}
		ent.GetMemoryMap().BlockWritePr(ent.GetMemIndex(), addr, rawdata)

		//System.Out.Println("OK");

		return nil

	}

	if utils.Copy(cmd, 1, 5) == "BSAVE" {

		parts := strings.Split(strings.Trim(utils.Delete(cmd, 1, 5), " "), ",")
		addr := 16384
		length := -1

		filename := parts[0]
		filename = strings.ToLower(filename)
		//System.Out.Println("will try to load "+filename);
		for i := 1; i < len(parts); i++ {
			tmp := strings.ToUpper(parts[i])
			if tmp[0] == 'A' {
				tmp = utils.Delete(tmp, 1, 1)
				addr = utils.StrToInt(tmp)
			} else if tmp[0] == 'L' {
				tmp = utils.Delete(tmp, 1, 1)
				length = utils.StrToInt(tmp)
			}
		}

		if !strings.HasSuffix(filename, ".s") {
			filename += ".s"
		}

		if length == -1 {
			return nil
		}

		data := make([]byte, length)

		for i, _ := range data {
			data[i] = byte(ent.GetMemory(addr + i))
		}

		e := files.WriteBytesViaProvider(ent.GetWorkDir(), filename, data)

		//System.Out.Println("OK");

		return e

	}

	// Run shim...

	if utils.Copy(cmd, 1, 3) == "RUN" {
		fn := ent.GetWorkDir() + "/" + strings.ToLower(strings.Trim(utils.Delete(cmd, 1, 3), " "))
		tl := types.NewTokenList()
		tl.Push(types.NewToken(types.STRING, fn))
		a := ent.GetCode()
		_, e := ent.GetDialect().GetCommands()["load"].Execute(nil, ent, *tl, a, *ent.GetLPC())
		if e != nil {
			return e
		}
		ent.Run(false)
	}

	if utils.Copy(cmd, 1, 5) == "CHAIN" {
		fn := ent.GetWorkDir() + "/" + strings.ToLower(strings.Trim(utils.Delete(cmd, 1, 5), " "))
		tl := types.NewTokenList()
		tl.Push(types.NewToken(types.STRING, fn))
		a := ent.GetCode()
		_, e := ent.GetDialect().GetCommands()["load"].Execute(nil, ent, *tl, a, *ent.GetLPC())
		if e != nil {
			return e
		}
		a = ent.GetCode()
		pc := ent.GetPC()
		pc.Line = a.GetLowIndex()
		pc.Statement = 0
		pc.Token = 0
		ent.SetPC(pc)
		ent.SetState(types.RUNNING)
	}

	if utils.Copy(cmd, 1, 5) == "WRITE" {
		fn := ent.GetWorkDir() + "/" + strings.ToLower(strings.Trim(utils.Delete(cmd, 1, 5), " ")) + ".d"
		//fn := files.GetUserPath(files.BASEDIR, []string{ent.GetWorkDir(), strings.ToLower(parts[0]) + ".d"})
		e := files.DOSWRITE(files.GetPath(fn), files.GetFilename(fn), 0)
		if e == nil {
			ent.SetOutChannel(fn)
		}
		return e
	}

	if utils.Copy(cmd, 1, 4) == "READ" {
		fn := ent.GetWorkDir() + "/" + strings.ToLower(strings.Trim(utils.Delete(cmd, 1, 4), " ")) + ".d"
		e := files.DOSREAD(files.GetPath(fn), files.GetFilename(fn), 0)

		if e == nil {
			ent.SetInChannel(fn)
		}

		return e
	}

	if utils.Copy(cmd, 1, 6) == "APPEND" {
		fn := ent.GetWorkDir() + "/" + strings.ToLower(strings.Trim(utils.Delete(cmd, 1, 6), " ")) + ".d"
		//fn := files.GetUserPath(files.BASEDIR, []string{ent.GetWorkDir(), strings.ToLower(parts[0]) + ".d"})
		e := files.DOSAPPEND(files.GetPath(fn), files.GetFilename(fn))
		if e == nil {
			ent.SetOutChannel(fn)
		}
		return e
	}

	if utils.Copy(cmd, 1, 5) == "CLOSE" {
		fn := ent.GetWorkDir() + "/" + strings.ToLower(strings.Trim(utils.Delete(cmd, 1, 5), " ")) + ".d"

		var e error
		if cmd == "CLOSE" {
			e = files.DOSCLOSEALL()
		} else {
			e = files.DOSCLOSE(files.GetPath(fn), files.GetFilename(fn))
		}

		ent.SetOutChannel("")
		ent.SetInChannel("")
		ent.SetFeedBuffer("")

		return e
	}

	if utils.Copy(cmd, 1, 4) == "OPEN" {
		fn := ent.GetWorkDir() + "/" + strings.ToLower(strings.Trim(utils.Delete(cmd, 1, 4), " ")) + ".d"

		e := files.DOSOPEN(files.GetPath(fn), files.GetFilename(fn), 0)
		ent.SetOutChannel("")
		ent.SetFeedBuffer("")
		return e
	}

	if utils.Copy(cmd, 1, 3) == "PR#" {
		cmd = strings.Trim(utils.Delete(cmd, 1, 3), " ")
		mode := utils.StrToInt(cmd)

		switch mode { /* FIXME - Switch statement needs cleanup */
		case 0:
			{
				//ent.GetVDU().SetVideoMode(ent.GetVDU().GetVideoModes()[5])
				//ent.GetVDU().ClrHome()
				//				ent.GetVDU().RegenerateWindow(ent.GetMemory())
				break
			}
		case 3:
			{
				//ent.GetVDU().SetVideoMode(ent.GetVDU().GetVideoModes()[0])
				//ent.GetVDU().ClrHome()
				//				ent.GetVDU().RegenerateWindow(ent.GetMemory())
				break
			}
		}

		return nil
	}

	return nil

}

func (this *DialectLogo) GetDynamicFunctionDef(name string) []string {
	c, ok := this.DynaFunctions[strings.ToLower(name)]
	if !ok {
		return []string{
			"to " + name,
			"",
			"end",
		}
	}

	n, params := c.GetFunctionSpec()

	out := make([]string, 0)
	header := "TO " + n
	for _, v := range params.Content {
		header += " " + v.Content
	}
	out = append(out, header)
	for _, l := range c.GetRawCode() {
		if strings.ToLower(l) != "end" {
			l += "    "
		}
		out = append(out, l)
	}

	return out

}

func (this *DialectLogo) GetDynamicCommandDef(name string) []string {

	if p, ok := this.Driver.GetProc(name); ok {
		return p.GetCode()
	}

	return []string{
		"TO " + name,
		"",
		"END",
	}

}

func (this *DialectLogo) GetDynamicCommands() []string {
	out := this.Driver.GetProcList()
	return out
}

func (this *DialectLogo) Init() {

	this.ShortName = "logo"
	this.LongName = "microLOGO"

	/* vars */
	//	var dn interfaces.DynaCoder
	//var diabc DiaBCODE

	// limit length
	this.MaxVariableLength = 2

	/* Dynacodez */
	//diabc = types.NewDiaBCODE()
	/*title*/
	this.AddCommand("sharetable", NewStandardCommandSHARETABLE())
	this.AddCommand("localtable", NewStandardCommandLOCALTABLE())
	this.AddCommand("table", NewStandardCommandGLOBALTABLE())
	this.AddCommand("setcell", &StandardCommandSETCELL{})

	this.AddCommand("share", &StandardCommandSHARE{})

	this.AddCommand("musicplay", &StandardCommandMUSICPLAY{})
	this.AddCommand("musicstop", &StandardCommandMUSICSTOP{})

	this.AddCommand("wait", &StandardCommandWAIT{})
	this.AddCommand("throw", &StandardCommandTHROW{})
	this.AddCommand("catch", &StandardCommandCATCH{})

	this.AddCommand("dribble", &StandardCommandDRIBBLE{})
	this.AddCommand("nodribble", &StandardCommandNODRIBBLE{})
	this.AddCommand("setread", &StandardCommandSETREAD{})
	this.AddCommand("setwrite", &StandardCommandSETWRITE{})
	this.AddCommand("open", &StandardCommandOPEN{})
	this.AddCommand("close", &StandardCommandCLOSE{})
	this.AddCommand("closeall", &StandardCommandCLOSEALL{})
	this.AddCommand("setreadpos", &StandardCommandSETREADPOS{})
	this.AddCommand("setwritepos", &StandardCommandSETWRITEPOS{})
	//this.AddCommand("allopen", &StandardCommandALLOPEN{})

	this.AddCommand("bury", &StandardCommandBURY{})
	this.AddCommand("unbury", &StandardCommandUNBURY{})
	this.AddCommand("buryall", &StandardCommandBURYALL{})
	this.AddCommand("unburyall", &StandardCommandUNBURYALL{})
	this.AddCommand("buryname", &StandardCommandBURYNAME{})
	this.AddCommand("unburyname", &StandardCommandUNBURYNAME{})

	this.AddCommand("show", &StandardCommandSHOW{})
	this.AddCommand("toot", &StandardCommandTOOT{})

	//this.AddCommand("list", &StandardCommandLIST{})
	this.AddCommand("window", &StandardCommandWINDOW{})
	this.AddCommand("fence", &StandardCommandFENCE{})
	this.AddCommand("wrap", &StandardCommandWRAP{})

	this.AddCommand("pprop", &StandardCommandPPROP{})
	this.AddCommand("remprop", &StandardCommandREMPROP{})
	this.AddCommand("pps", &StandardCommandPPS{})

	this.AddCommand("iftrue", &StandardCommandIFTRUE{})
	this.AddCommand("iffalse", &StandardCommandIFFALSE{})
	this.AddCommand("ift", &StandardCommandIFTRUE{})
	this.AddCommand("iff", &StandardCommandIFFALSE{})
	this.AddCommand("test", &StandardCommandTEST{})

	this.AddCommand("setturtle", &StandardCommandSETTURTLE{})

	this.AddCommand("setheading", &StandardCommandSETH{})
	this.AddCommand("setpitch", &StandardCommandSETP{})
	this.AddCommand("setroll", &StandardCommandSETR{})
	this.AddCommand("setbg", &StandardCommandSETBG{})
	this.AddCommand("setpos", &StandardCommandSETPOS{})
	this.AddCommand("setxy", &StandardCommandSETPOS{})
	this.AddCommand("setpos3d", &StandardCommandSETPOS3{})
	this.AddCommand("sethome", &StandardCommandSETHOME{})
	this.AddCommand("setcursor", &StandardCommandSETCURSOR{})
	this.AddCommand("seth", &StandardCommandSETH{})
	this.AddCommand("setp", &StandardCommandSETP{})
	this.AddCommand("setr", &StandardCommandSETR{})
	this.AddCommand("setx", &StandardCommandSETX{})
	this.AddCommand("sety", &StandardCommandSETY{})
	this.AddCommand("edit", &StandardCommandEDIT{})
	this.AddCommand("print", &StandardCommandSHOW{PrintMode: true})
	this.AddCommand("type", &StandardCommandPRINT{NoNL: true})
	this.AddCommand("pr", &StandardCommandSHOW{PrintMode: true})
	this.AddCommand("cleartext", &applesoft.StandardCommandCLS{})
	this.AddCommand("ct", &applesoft.StandardCommandCLS{})
	this.AddCommand("textscreen", &StandardCommandTEXT{})
	this.AddCommand("ts", &StandardCommandTEXT{})
	this.AddCommand("splitscreen", &StandardCommandGRAPHICS{Split: true})
	this.AddCommand("ss", &StandardCommandGRAPHICS{Split: true})
	this.AddCommand("fullscreen", &StandardCommandGRAPHICS{Split: false})
	this.AddCommand("fs", &StandardCommandGRAPHICS{Split: false})
	this.AddCommand("fd", &StandardCommandFD{})
	this.AddCommand("forward", &StandardCommandFD{})
	this.AddCommand("bk", &StandardCommandBK{})
	this.AddCommand("back", &StandardCommandBK{})
	this.AddCommand("lt", &StandardCommandLT{})
	this.AddCommand("left", &StandardCommandLT{})
	this.AddCommand("rt", &StandardCommandRT{})
	this.AddCommand("right", &StandardCommandRT{})
	this.AddCommand("up", &StandardCommandUP{})
	this.AddCommand("dn", &StandardCommandDN{})
	this.AddCommand("down", &StandardCommandDN{})
	this.AddCommand("setpc", &StandardCommandPENCOLOR{})
	this.AddCommand("home", &StandardCommandHOMETURTLE{})
	this.AddCommand("clearscreen", &StandardCommandHOME{})
	this.AddCommand("cs", &StandardCommandHOME{})
	this.AddCommand("os", &StandardCommandOVERLAYSCREEN{})
	this.AddCommand("overlayscreen", &StandardCommandOVERLAYSCREEN{})
	this.AddCommand("clean", &StandardCommandCLEAN{})
	this.AddCommand("showturtle", &StandardCommandSHOWTURTLE{})
	this.AddCommand("st", &StandardCommandSHOWTURTLE{})
	this.AddCommand("hideturtle", &StandardCommandHIDETURTLE{})
	this.AddCommand("ht", &StandardCommandHIDETURTLE{})
	this.AddCommand("penup", &StandardCommandPENUP{})
	this.AddCommand("pu", &StandardCommandPENUP{})
	this.AddCommand("pendown", &StandardCommandPENDOWN{})
	this.AddCommand("pd", &StandardCommandPENDOWN{})

	this.AddCommand("repeat", &StandardCommandREPEAT{})

	this.AddCommand("run", &StandardCommandRUN{})
	this.AddCommand("to", &StandardCommandTO{})
	this.AddCommand("end", &StandardCommandEND{})
	this.AddCommand("stop", &StandardCommandSTOP{})
	this.AddCommand("if", &StandardCommandIF{})
	this.AddCommand("exit", &StandardCommandEXIT{})
	this.AddCommand("setpenwidth", &StandardCommandSETWIDTH{})
	this.AddCommand("make", &StandardCommandMAKE{})
	this.AddCommand("local", &StandardCommandLOCAL{Local: true})
	this.AddCommand("pots", &StandardCommandPOTS{})
	this.AddCommand("save", &StandardCommandSAVE{})
	this.AddCommand("load", &StandardCommandLOAD{})
	this.AddCommand("rl", &StandardCommandRL{})
	this.AddCommand("rr", &StandardCommandRR{})
	this.AddCommand("rollleft", &StandardCommandRL{})
	this.AddCommand("rollright", &StandardCommandRR{})
	this.AddCommand("output", &StandardCommandOUTPUT{})
	this.AddCommand("op", &StandardCommandOUTPUT{})
	this.AddCommand("erall", &StandardCommandERALL{})
	this.AddCommand("erase", &StandardCommandERASE{})
	this.AddCommand("backtrack", &StandardCommandBACKTRACK{})

	this.AddCommand("bt", &StandardCommandBACKTRACK{})

	this.AddCommand("camreset", &StandardCommandCRESET{})
	this.AddCommand("camorbit", &StandardCommandCORBIT{})
	this.AddCommand("camzoom", &StandardCommandCZOOM{})
	this.AddCommand("camrotate", &StandardCommandCROT{})
	this.AddCommand("camfocus", &StandardCommandCFOCUSTOITLE{})
	this.AddCommand("setcamheading", &StandardCommandCHEADING{})
	this.AddCommand("setcampos", &StandardCommandSETCAMPOS{})
	this.AddCommand("cammove", &StandardCommandCAMMOVE{})

	this.AddCommand("raise", &StandardCommandRAISE{})
	this.AddCommand("ra", &StandardCommandRAISE{})
	this.AddCommand("lower", &StandardCommandLOWER{})
	this.AddCommand("lo", &StandardCommandLOWER{})

	this.AddCommand("shufl", &StandardCommandSHUFL{})
	this.AddCommand("sl", &StandardCommandSHUFL{})
	this.AddCommand("shufr", &StandardCommandSHUFR{})
	this.AddCommand("sr", &StandardCommandSHUFR{})

	this.AddCommand("po", &StandardCommandPO{})
	this.AddCommand("pops", &StandardCommandPOPS{procs: true})
	this.AddCommand("poall", &StandardCommandPOPS{procs: true, vars: true})
	this.AddCommand("pons", &StandardCommandPOPS{procs: false, vars: true})
	this.AddCommand("pot", &StandardCommandPOTS{needproc: true})
	this.AddCommand("pon", &StandardCommandPON{})

	this.AddCommand("define", &StandardCommandDEFINE{})

	this.AddCommand("ft", &StandardCommandFT{})
	this.AddCommand("filledtriangle", &StandardCommandFT{})
	this.AddCommand("iso", &StandardCommandFT{})
	this.AddCommand("cube", &StandardCommandCUBE{})
	this.AddCommand("cuboid", &StandardCommandCUBEXYZ{})
	this.AddCommand("voxel", &StandardCommandCUBE{solid: true})
	this.AddCommand("voxeloid", &StandardCommandCUBEXYZ{solid: true})
	this.AddCommand("square", &StandardCommandQUAD{solid: true})
	this.AddCommand("quad", &StandardCommandQUADXY{solid: true})
	this.AddCommand("box", &StandardCommandQUAD{})
	this.AddCommand("rect", &StandardCommandQUADXY{})
	this.AddCommand("circle", &StandardCommandCIRCLE{})
	this.AddCommand("spot", &StandardCommandCIRCLE{solid: true})
	this.AddCommand("sphere", &StandardCommandSPHERE{solid: true})
	this.AddCommand("pyramid", &StandardCommandPYRAMID{solid: true})

	this.AddCommand("colors", &StandardCommandCOLORS{})
	this.AddCommand("settextfont", &StandardCommandTEXTFONT{})
	this.AddCommand("setpromptstring", &StandardCommandPROMPTSTRING{})
	this.AddCommand("setpromptcolor", &StandardCommandPROMPTCOLOR{})
	this.AddCommand("setwidth", &StandardCommandSETTEXTWIDTH{})

	this.AddCommand("noise", &StandardCommandNOISE{})
	this.AddCommand("playnotes", &StandardCommandPLAYNOTES{})
	this.AddCommand("setinstrument", &StandardCommandSETINSTRUMENT{})

	this.AddCommand("poly", &StandardCommandPOLY{})
	this.AddCommand("polyspot", &StandardCommandPOLY{solid: true})

	this.AddCommand("tagshow", &StandardCommandTAGVISIBLE{})
	this.AddCommand("taghide", &StandardCommandTAGVISIBLE{hidden: true})
	this.AddCommand("tagswitch", &StandardCommandTAGSWITCH{})

	this.AddCommand("arc", &StandardCommandARC{})
	this.AddCommand("pie", &StandardCommandARC{solid: true, percent: true})

	this.AddCommand("settextcolor", &StandardCommandSETTEXTCOLOR{})
	this.AddCommand("settextsize", &StandardCommandSETTEXTSIZE{})
	this.AddCommand("setts", &StandardCommandSETTEXTSIZE{})
	//this.AddCommand("settc", &StandardCommandSETTEXTCOLOR{})

	this.AddCommand("fc", &StandardCommandFILLCOLOR{})
	this.AddCommand("ftcolor", &StandardCommandFILLCOLOR{})
	this.AddCommand("setfc", &StandardCommandFILLCOLOR{})

	this.AddCommand("loadpic", &StandardCommandLOADPIC{})
	this.AddCommand("savepic", &StandardCommandSAVEPIC{})

	this.AddCommand("fast", &StandardCommandFAST{})
	this.AddCommand("slow", &StandardCommandSLOW{})

	this.AddCommand("tagbegin", &StandardCommandTAGBEGIN{})
	this.AddCommand("tagend", &StandardCommandTAGEND{})
	this.AddCommand("tagerase", &StandardCommandTAGDELETE{})
	this.AddCommand("setmodel", &StandardCommandSETMODEL{})

	this.AddCommand("pause", &StandardCommandPAUSE{})
	this.AddCommand("co", &StandardCommandCONT{})

	this.AddCommand("setsemaphore", &StandardCommandSEMA{})

	this.AddCommand("setrender", &StandardCommandSETRENDER{})

	this.AddCommand("type", &StandardCommandTYPE{})
	this.AddCommand("settypesize", &StandardCommandTYPESIZE{})
	this.AddCommand("settypedepth", &StandardCommandTYPEDEPTH{})
	this.AddCommand("settypestretch", &StandardCommandTYPESTRETCH{})
	this.AddCommand("settypeface", &StandardCommandTYPEFACE{})
	this.AddCommand("settypefilled", &StandardCommandTYPEFILLED{})

	// commands
	this.AddCommand("logo", &StandardCommandLOGO{})
	this.AddCommand("routines", &StandardCommandLOGOLIST{})
	this.AddCommand("kill", &StandardCommandLOGOKILL{})
	this.AddCommand("killall", &StandardCommandLOGOKILLALL{})
	this.AddCommand("dawdle", &StandardCommandLOGODAWDLE{})
	this.AddCommand("channel", &StandardCommandCHANNEL{})
	this.AddCommand("transmit", &StandardCommandCHANNELTRANSMIT{})

	this.AddCommand("swap", &StandardCommandVARSWAP{})

	this.AddCommand("wedge", &StandardCommandTRIANGLE{solid: true})
	this.AddCommand("triangle", &StandardCommandTRIANGLE{solid: false})

	this.AddCommand("camturtle", &StandardCommandCAMTURTLE{})
	this.AddCommand("camfollow", &StandardCommandCAMFOLLOW{})
	this.AddCommand("camcontrol", &StandardCommandCAMCONTROL{})
	this.AddCommand("joystick", &StandardCommandJOYSTICK{})

	this.AddCommand("rem", &StandardCommandREM{})

	this.AddCommand("while", &StandardCommandWHILE{})

	this.AddCommand("setfopa", &StandardCommandFILLOPACITY{})
	this.AddCommand("setpopa", &StandardCommandPENOPACITY{})

	this.AddCommand("step", &StandardCommandSTEP{})
	this.AddCommand("for", &StandardCommandFOR{})

	// functions
	this.AddFunction("lastrem", NewStandardFunctionLASTREM(0, 0, *types.NewTokenList(), this))

	this.AddFunction("true", NewStandardFunctionTRUE(0, 0, *types.NewTokenList()))
	this.AddFunction("false", NewStandardFunctionFALSE(0, 0, *types.NewTokenList()))
	this.AddFunction("on", NewStandardFunctionTRUE(0, 0, *types.NewTokenList()))
	this.AddFunction("off", NewStandardFunctionFALSE(0, 0, *types.NewTokenList()))

	this.AddFunction("routine", NewStandardFunctionROUTINE(0, 0, *types.NewTokenList(), this))
	this.AddFunction("receive", NewStandardFunctionCHANNELRECEIVE(0, 0, *types.NewTokenList(), this))
	this.AddFunction("logoid", NewStandardFunctionLOGOID(0, 0, *types.NewTokenList(), this))

	this.AddFunction("cell", NewStandardFunctionCELL(0, 0, *types.NewTokenList()))
	this.AddFunction("srow", NewStandardFunctionSROW(0, 0, *types.NewTokenList()))
	this.AddFunction("scol", NewStandardFunctionSCOL(0, 0, *types.NewTokenList()))

	this.AddFunction("abs", applesoft.NewStandardFunctionABS(0, 0, *types.NewTokenList()))
	this.AddFunction("random", appleinteger.NewStandardFunctionRND(0, 0, *types.NewTokenList()))
	this.AddFunction("sqrt", applesoft.NewStandardFunctionSQR(0, 0, *types.NewTokenList()))

	this.AddFunction("int", applesoft.NewStandardFunctionINT(0, 0, *types.NewTokenList()))

	this.AddFunction("not", NewStandardFunctionNOT(0, 0, *types.NewTokenList()))
	this.AddFunction("and", NewStandardFunctionAND(0, 0, *types.NewTokenList()))
	this.AddFunction("or", NewStandardFunctionOR(0, 0, *types.NewTokenList()))

	this.AddFunction("sin", NewStandardFunctionSIN(0, 0, *types.NewTokenList()))
	this.AddFunction("cos", NewStandardFunctionCOS(0, 0, *types.NewTokenList()))
	this.AddFunction("tan", NewStandardFunctionTAN(0, 0, *types.NewTokenList()))
	this.AddFunction("arctan", NewStandardFunctionATAN(0, 0, *types.NewTokenList()))
	this.AddFunction("paddle", applesoft.NewStandardFunctionPDL(0, 0, *types.NewTokenList()))
	this.AddFunction("buttonp", NewStandardFunctionBUTTONP(0, 0, *types.NewTokenList()))

	this.AddFunction("sum", NewStandardFunctionSUM(0, 0, *types.NewTokenList()))
	this.AddFunction("difference", NewStandardFunctionDIFFERENCE(0, 0, *types.NewTokenList()))
	this.AddFunction("product", NewStandardFunctionPRODUCT(0, 0, *types.NewTokenList()))
	this.AddFunction("quotient", NewStandardFunctionQUOTIENT(0, 0, *types.NewTokenList()))
	this.AddFunction("intquotient", NewStandardFunctionINTQUOTIENT(0, 0, *types.NewTokenList()))
	this.AddFunction("remainder", NewStandardFunctionREMAINDER(0, 0, *types.NewTokenList()))
	this.AddFunction("first", NewStandardFunctionFIRST(0, 0, *types.NewTokenList()))
	this.AddFunction("last", NewStandardFunctionLAST(0, 0, *types.NewTokenList()))
	this.AddFunction("butfirst", NewStandardFunctionBUTFIRST(0, 0, *types.NewTokenList()))
	this.AddFunction("butlast", NewStandardFunctionBUTLAST(0, 0, *types.NewTokenList()))
	this.AddFunction("bf", NewStandardFunctionBUTFIRST(0, 0, *types.NewTokenList()))
	this.AddFunction("bl", NewStandardFunctionBUTLAST(0, 0, *types.NewTokenList()))
	this.AddFunction("round", NewStandardFunctionROUND(0, 0, *types.NewTokenList()))
	this.AddFunction("uppercase", NewStandardFunctionUPPERCASE(0, 0, *types.NewTokenList()))
	this.AddFunction("lowercase", NewStandardFunctionLOWERCASE(0, 0, *types.NewTokenList()))
	this.AddFunction("item", NewStandardFunctionITEM(0, 0, *types.NewTokenList()))
	this.AddFunction("member", NewStandardFunctionMEMBER(0, 0, *types.NewTokenList()))
	this.AddFunction("memberp", NewStandardFunctionMEMBERP(0, 0, *types.NewTokenList()))
	this.AddFunction("parse", NewStandardFunctionPARSE(0, 0, *types.NewTokenList()))
	this.AddFunction("fput", NewStandardFunctionFPUT(0, 0, *types.NewTokenList()))
	this.AddFunction("lput", NewStandardFunctionLPUT(0, 0, *types.NewTokenList()))
	this.AddFunction("list", NewStandardFunctionLIST(0, 0, *types.NewTokenList()))
	this.AddFunction("sentence", NewStandardFunctionSENTENCE(0, 0, *types.NewTokenList()))
	this.AddFunction("se", NewStandardFunctionSENTENCE(0, 0, *types.NewTokenList()))
	this.AddFunction("word", NewStandardFunctionWORD(0, 0, *types.NewTokenList()))
	this.AddFunction("rw", NewStandardFunctionREADWORD(0, 0, *types.NewTokenList()))
	this.AddFunction("readword", NewStandardFunctionREADWORD(0, 0, *types.NewTokenList()))
	//this.AddFunction("rl", NewStandardFunctionREADLIST(0, 0, *types.NewTokenList()))
	this.AddFunction("readlist", NewStandardFunctionREADLIST(0, 0, *types.NewTokenList()))
	this.AddFunction("rc", NewStandardFunctionREADCHAR(0, 0, *types.NewTokenList()))
	this.AddFunction("readchar", NewStandardFunctionREADCHAR(0, 0, *types.NewTokenList()))
	this.AddFunction("rcs", NewStandardFunctionREADCHARS(0, 0, *types.NewTokenList()))
	this.AddFunction("readchars", NewStandardFunctionREADCHARS(0, 0, *types.NewTokenList()))
	this.AddFunction("equalp", NewStandardFunctionEQUALP(0, 0, *types.NewTokenList()))

	this.AddFunction("wordp", NewStandardFunctionWORDP(0, 0, *types.NewTokenList()))
	this.AddFunction("listp", NewStandardFunctionLISTP(0, 0, *types.NewTokenList()))
	this.AddFunction("numberp", NewStandardFunctionNUMBERP(0, 0, *types.NewTokenList()))
	this.AddFunction("emptyp", NewStandardFunctionEMPTYP(0, 0, *types.NewTokenList()))
	this.AddFunction("count", NewStandardFunctionCOUNT(0, 0, *types.NewTokenList()))
	this.AddFunction("namep", NewStandardFunctionNAMEP(0, 0, *types.NewTokenList()))

	this.AddFunction("char", NewStandardFunctionCHAR(0, 0, *types.NewTokenList()))
	this.AddFunction("ascii", NewStandardFunctionASCII(0, 0, *types.NewTokenList()))

	this.AddFunction("xpos", NewStandardFunctionXPOS(0, 0, *types.NewTokenList(), this))
	this.AddFunction("ypos", NewStandardFunctionYPOS(0, 0, *types.NewTokenList(), this))
	this.AddFunction("zpos", NewStandardFunctionZPOS(0, 0, *types.NewTokenList(), this))

	this.AddFunction("xcor", NewStandardFunctionXPOS(0, 0, *types.NewTokenList(), this))
	this.AddFunction("ycor", NewStandardFunctionYPOS(0, 0, *types.NewTokenList(), this))
	this.AddFunction("zcor", NewStandardFunctionZPOS(0, 0, *types.NewTokenList(), this))

	this.AddFunction("heading", NewStandardFunctionHEADING(0, 0, *types.NewTokenList(), this))
	this.AddFunction("pitch", NewStandardFunctionPITCH(0, 0, *types.NewTokenList(), this))
	this.AddFunction("roll", NewStandardFunctionROLL(0, 0, *types.NewTokenList(), this))

	this.AddFunction("pencolor", NewStandardFunctionPENCOLOR(0, 0, *types.NewTokenList(), this))

	this.AddFunction("pos", NewStandardFunctionPOS(0, 0, *types.NewTokenList(), this))
	this.AddFunction("pos3d", NewStandardFunctionPOS3(0, 0, *types.NewTokenList(), this))

	this.AddFunction("camheading", NewStandardFunctionCAMHEADING(0, 0, *types.NewTokenList()))
	this.AddFunction("campos", NewStandardFunctionCAMPOS(0, 0, *types.NewTokenList()))

	this.AddFunction("gprop", NewStandardFunctionGPROP(0, 0, *types.NewTokenList()))

	this.AddFunction("keyp", NewStandardFunctionKEYP(0, 0, *types.NewTokenList()))

	this.AddFunction("readpos", NewStandardFunctionREADPOS(0, 0, *types.NewTokenList()))
	this.AddFunction("writepos", NewStandardFunctionWRITEPOS(0, 0, *types.NewTokenList()))
	this.AddFunction("reader", NewStandardFunctionREADER(0, 0, *types.NewTokenList()))
	this.AddFunction("writer", NewStandardFunctionWRITER(0, 0, *types.NewTokenList()))
	this.AddFunction("filelen", NewStandardFunctionFILELEN(0, 0, *types.NewTokenList()))
	this.AddFunction("allopen", NewStandardFunctionALLOPEN(0, 0, *types.NewTokenList()))

	this.AddFunction("thing", NewStandardFunctionTHING(0, 0, *types.NewTokenList(), this))

	this.AddFunction("semaphore", NewStandardFunctionSEMA(0, 0, *types.NewTokenList()))

	this.AddFunction("repcount", NewStandardFunctionREPCOUNT(0, 0, *types.NewTokenList(), this))

	this.AddCommand("cat", applesoft.NewStandardCommandCAT())

	// this.Logicals["or"] = 1
	// this.Logicals["and"] = 1
	// this.Logicals["not"] = 1

	this.VarSuffixes = "%$!&#"

	/* dynacode test shim */
	this.ReverseCase = true

	this.ArrayDimDefault = 10
	this.ArrayDimMax = 65535
	this.Title = "Logo"
	this.DefaultCost = 1000000000 / 5

	this.IPS = -1

}

func (this *DialectLogo) DefineProc(name string, params []string, code string) {

	// create dynacode process
	dc := dialect.NewDynaCodeWithRootDiaS(name, this, code, true)
	p := *types.NewTokenList()
	for _, pname := range params {
		p.Add(types.NewToken(types.VARIABLE, pname))
	}
	dc.Params = p

	if dc.HasToken(types.KEYWORD, "output") || dc.HasToken(types.KEYWORD, "OUTPUT") {
		this.AddDynaFunction(strings.ToLower(name), dc)
		//fmt.Println("FUNCTION")
		fmt.Printf("*** Registered dynamic function %s\n", name)
	} else {
		this.AddDynaCommand(strings.ToLower(name), dc)
		//fmt.Println("KEYWORD")
		fmt.Printf("*** Registered dynamic procedure %s\n", name)
	}
}

func (this *DialectLogo) StartProc(name string, params []string, code string) {

	//dc := dialect.NewDynaCodeWithRootDia()
	this.DefineMode = true
	this.DefineName = name
	this.DefineParams = params
	this.DefineProcedure = code

}

func (this *DialectLogo) NetBracketCount(tokens types.TokenList) int {
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
		}
	}
	return nbc
}

func (this *DialectLogo) SplitOnTokenTypeList(tokens types.TokenList, tok []types.TokenType) types.TokenListArray {

	/* vars */
	result := types.NewTokenListArray()
	var idx int
	var bc int
	//Token tt;

	idx = 0
	//SetLength(result, idx+1);
	result = result.Add(*types.NewTokenList())

	for _, tt := range tokens.Content {
		if tt.Type == types.OBRACKET && tt.Content == "(" {
			if bc > 0 {
				tl := result.Get(idx)
				tl.Push(tt)
			}
			bc++
		} else if tt.Type == types.OBRACKET && tt.Content == ")" {
			bc--
			if bc > 0 {
				tl := result.Get(idx)
				tl.Push(tt)
			}
		} else if tt.IsIn(tok) && bc == 0 {
			idx++
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

func (this *DialectLogo) ExecuteDirectCommand(tl types.TokenList, ent interfaces.Interpretable, Scope *types.Algorithm, LPC *types.CodeRef) error {

	/* vars */
	var tok *types.Token
	var n string
	var cmd interfaces.Commander
	var cr *types.CodeRef
	var ss *types.TokenList

	panic(errors.New("edc called :("))

	if this.NetBracketCount(tl) != 0 {
		return exception.NewESyntaxError("SYNTAX ERROR 1")
	}

	//fmt.Println("DEBUG: -------------> [" + this.Title + "]: " + utils.IntToStr(LPC.Line) + " " + ent.TokenListAsString(tl))

	if ent.IsDebug() && ent.IsRunning() {

		if tl.Get(0).Content == "next" {
			if ent.GetLoopStack().Size() > 0 {
				zz := ent.GetLoopStack().Get(ent.GetLoopStack().Size() - 1).Entry

				if zz.Line == ent.GetPC().Line && zz.Statement == ent.GetPC().Statement {
					//ent.SetSilent(true)
					defer ent.SetSilent(false)
				}
			}
		}

		ent.Log("DO", ent.TokenListAsString(tl))

	}

	if this.Trace && (ent.GetState() == types.RUNNING) {
		ent.PutStr("#" + utils.IntToStr(LPC.Line) + " ")
	}

	//fmt.Println("---------------------------------------:", ent.TokenListAsString(tl))

	// BREAK TOKENLIST ON KEYWORDS
	var tla types.TokenListArray
	if tl.LPeek().Type == types.KEYWORD && strings.ToLower(tl.LPeek().Content) == "to" {
		tla = types.TokenListArray{tl}
	} else {
		tla = this.SplitOnTokenTypeList(tl, []types.TokenType{types.KEYWORD, types.DYNAMICKEYWORD})
	}

	for _, tl := range tla {

		//fmt.Println(ii, "------------------>", ent.TokenListAsString(tl))

		/* process poop monster here (@^-^@) */
		tok = tl.Shift()
		if tok == nil {
			continue
		}

		if this.DefineMode {
			if this.DefineProcedure != "" {
				this.DefineProcedure += "\r\n"
			}
			tl.UnShift(tok)
			this.DefineProcedure += this.TokenListAsString(tl)

			if tok.Type == types.KEYWORD && strings.ToLower(tok.Content) == "end" {
				this.DefineMode = false
				// Call Proc creation here
				this.DefineProc(
					this.DefineName,
					this.DefineParams,
					this.DefineProcedure,
				)

				if !this.GetSilentDefines() {
					this.PutStr(ent, "DEFINED "+strings.ToUpper(this.DefineName)+"\r\n")

					lines := make([]string, 0)
					for _, procname := range this.GetDynamicFunctions() {
						// procname
						def := this.GetDynamicFunctionDef(procname)
						if len(lines) > 0 {
							lines = append(lines, "")
						}
						lines = append(lines, def...)
					}
					lines = append(lines, "")
					for _, procname := range this.GetDynamicCommands() {
						// procname
						def := this.GetDynamicCommandDef(procname)
						if len(lines) > 0 {
							lines = append(lines, "")
						}
						lines = append(lines, def...)
					}

					for _, l := range lines {
						if l != "" {
							fmt.Printf("*** DEF DUMP: %s\n", l)
							tl := this.Tokenize(runestring.Cast(l))
							scope := ent.GetDirectAlgorithm()
							this.SetSilentDefines(true)
							this.ExecuteDirectCommand(*tl, ent, scope, ent.GetLPC())
							this.SetSilentDefines(false)
						}
					}
				}
				ent.SetPrompt(this.OldPrompt)
				return nil
			}

			// continue, not return as we want the other lines - AA
			// return nil
			continue
		}

		if (tok.Type == types.NUMBER) || (tok.Type == types.INTEGER) {
			//
			//
		} else if tok.Type == types.DYNAMICKEYWORD {
			n = strings.ToLower(tok.Content)
			if dcmd, ok := this.DynaCommands[n]; ok {
				cr = types.NewCodeRef()
				cr.Line = dcmd.GetCode().GetLowIndex()
				cr.Statement = 0
				cr.Token = 0
				if cr.Line != -1 {
					oldProcName := ent.GetCurrentSubroutine()
					procName := n
					isRecursive := (procName == oldProcName)
					//fmt.Printf("%s -> %s (rec: %v)\n", oldProcName, procName, isRecursive)

					/* something to do */
					ss = tl.SubList(0, tl.Size())

					//fmt.Printf("Tokens before parse: %s", ent.TokenListAsString(*ss))

					rtok, e := this.ParseTokensForResult(ent, *ss)
					if e != nil {
						return e
					}
					ss.Clear()
					if rtok.Type == types.LIST {
						for _, tt := range rtok.List.Content {
							ss.Add(tt)
						}
					} else {
						ss.Add(rtok)
					}

					if ss.Size() < dcmd.GetParamCount() {
						return exception.NewESyntaxError(strings.ToUpper(procName) + " REQUIRES MORE PARAMETERS")
					}

					ent.Call(*cr, dcmd.GetCode(), ent.GetState(), false, n+ent.GetVarPrefix(), *ss, dcmd.GetDialect()) // call with isolation off
					ent.SetCurrentSubroutine(procName)
					dcmd.SeedParamsData(*ss, ent)
					if isRecursive {
						ent.Pop(false)
					}
					//fmt.Printf("Stack size: %d\n", ent.GetStack().Size())
				} else {
					return exception.NewESyntaxError("Dynamic Code Hook has no content")
				}
				//}
				//catch (Exception e) {

				//}

			} else {
				return exception.NewESyntaxError("I DON'T KNOW HOW TO " + strings.ToUpper(tok.Content))
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
						v, serr := this.ParseTokensForResult(ent, tt)
						if serr != nil {
							this.HandleException(ent, serr)
						}
						subexpr.Push(v)
					}
					////fmt.Println(ent.TokenListAsString(*subexpr))
				}

				fun.SetEntity(ent)
				fun.SetQuery(false)
				fun.FunctionExecute(subexpr)

			} else {
				this.HandleException(ent, exception.NewESyntaxError("SYNTAX ERROR 2"))
			}

		} else if tok.Type == types.KEYWORD {

			if (tl.Size() > 0) && (tl.LPeek().Type == types.ASSIGNMENT) {
				return exception.NewESyntaxError("SYNTAX ERROR 3")
			}

			n = strings.ToLower(tok.Content)
			if cmd, ok := this.Commands[n]; ok {
				//cmd = this.Commands.Get(n

				_, err := cmd.Execute(nil, ent, tl, Scope, *LPC)
				if err != nil {
					ent.BufferEmpty()
					//this.HandleException(ent, err)
					return err
				}
				cost := cmd.GetCost()
				if cost == 0 {
					cost = this.GetDefaultCost()
				}
				if settings.LogoFastDraw[ent.GetMemIndex()] {
					cost = 0
				}
				ent.Wait(int64(float32(cost) * (100 / this.Throttle)))

			} else {
				return exception.NewESyntaxError("I DON'T KNOW HOW TO " + strings.ToUpper(tok.Content))
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
			if settings.LogoFastDraw[ent.GetMemIndex()] {
				cost = 0
			}
			ent.Wait(int64(float32(cost) * (100 / this.Throttle)))

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
			if settings.LogoFastDraw[ent.GetMemIndex()] {
				cost = 0
			}
			ent.Wait(int64(float32(cost) * (100 / this.Throttle)))

		} else {
			return exception.NewESyntaxError("I DON'T KNOW HOW TO " + strings.ToUpper(tok.Content))
		}

	}

	if this.DefineMode {
		ent.SetPrompt(">")
	} else {
		ent.SetPrompt("?")
	}

	///tl.Free; /* clean up */
	return nil
}

func (this *DialectLogo) PutStr(ent interfaces.Interpretable, s string) {
	apple2helpers.PutStr(ent, s)
}

func (this *DialectLogo) RealPut(ent interfaces.Interpretable, ch rune) {
	apple2helpers.Put(ent, ch)
}

func (this *DialectLogo) Backspace(ent interfaces.Interpretable) {
	apple2helpers.Backspace(ent)
}

func (this *DialectLogo) ClearToBottom(ent interfaces.Interpretable) {
	apple2helpers.ClearToBottom(ent)
}

func (this *DialectLogo) SetCursorX(ent interfaces.Interpretable, xx int) {
	x := (80 / apple2helpers.GetFullColumns(ent)) * xx

	apple2helpers.SetCursorX(ent, x)
}

func (this *DialectLogo) SetCursorY(ent interfaces.Interpretable, yy int) {
	y := (48 / apple2helpers.GetFullRows(ent)) * yy

	apple2helpers.SetCursorY(ent, y)
}

func (this *DialectLogo) GetColumns(ent interfaces.Interpretable) int {
	return apple2helpers.GetColumns(ent)
}

func (this *DialectLogo) GetRows(ent interfaces.Interpretable) int {
	return apple2helpers.GetRows(ent)
}

func (this *DialectLogo) Repos(ent interfaces.Interpretable) {
	apple2helpers.Gotoxy(ent, int(ent.GetMemory(36)), int(ent.GetMemory(37)))
}

func (this *DialectLogo) GetCursorX(ent interfaces.Interpretable) int {
	return apple2helpers.GetCursorX(ent) / (80 / apple2helpers.GetFullColumns(ent))
}

func (this *DialectLogo) GetCursorY(ent interfaces.Interpretable) int {
	return apple2helpers.GetCursorY(ent) / (48 / apple2helpers.GetFullRows(ent))
}

func (this *DialectLogo) GetProgramStart(ent interfaces.Interpretable) int {
	return int(ent.GetMemory(103)) + 256*int(ent.GetMemory(104))
}

func (this *DialectLogo) InitVarmap(ent interfaces.Interpretable, vm types.VarManager) {

	//ent.SetMemory(103, 1)
	//ent.SetMemory(104, 8)

	//MEMBASE := this.GetProgramStart(ent)

	//fretop := MEMTOP

	//ent.SetMemory(105, uint64(ent.GetMemory(MEMBASE)%256))
	//ent.SetMemory(106, uint64(ent.GetMemory(MEMBASE)/256))
	//varmem := 256*int(ent.GetMemory(106)) + int(ent.GetMemory(105))

	//// Create an Applesoft compatible memory map
	//vmgr := types.NewVarManagerMSBIN(
	//	ent.GetMemoryMap(),
	//	ent.GetMemIndex(),
	//	105,
	//	107,
	//	111,
	//	109,
	//	115,
	//	types.VUR_QUIET,
	//)

	//vmgr.SetVector(vmgr.VARTAB, varmem)
	//vmgr.SetVector(vmgr.ARRTAB, varmem)
	//vmgr.SetVector(vmgr.STREND, varmem+1)
	//vmgr.SetVector(vmgr.FRETOP, fretop)
	//vmgr.SetVector(vmgr.MEMSIZ, fretop)

	//// set the start of the table to zeroes to prevent spurious variable recognition`
	//ent.SetMemory(varmem, 0)
	//ent.SetMemory(varmem+1, 0)

	//ent.SetVM(vmgr)

}

// NTokenize tokenize a group of tokens to uints
func (this *DialectLogo) NTokenize(tl types.TokenList) []uint64 {

	var values []uint64

	//var lasttok *types.Token

	for _, t := range tl.Content {
		switch t.Type {
		case types.LOGIC:
			//values = append(values, this.TokenMapping[strings.ToLower(t.Content)])
			//values = append(values, TokenToCode[strings.ToUpper(t.Content)])
		case types.KEYWORD:
			//values = append(values, this.TokenMapping[strings.ToLower(t.Content)])
			//values = append(values, TokenToCode[strings.ToUpper(t.Content)])
		case types.FUNCTION:
			//values = append(values, this.TokenMapping[strings.ToLower(t.Content)])
			//values = append(values, TokenToCode[strings.ToUpper(t.Content)])
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

func (this *DialectLogo) GetMemoryRepresentation(a *types.Algorithm) []uint64 {
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

func (this *DialectLogo) ParseMemoryRepresentation(data []uint64) types.Algorithm {

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

func (this *DialectLogo) UnNTokenize(values []uint64) *types.TokenList {

	s := ""
	var lastcode uint64 = 0

	var skipspace bool

	for _, v := range values {
		if v < 128 {
			if lastcode >= 128 && !skipspace {
				s = s + " "
			}
			skipspace = false
			if v > 0 {
				s = s + string(rune(v))
			}
		} else {
			//			s = s + " " + CodeToToken[v]
			//			if CodeToToken[v] == "REM" || CodeToToken[v] == "DATA" {
			//				skipspace = true
			//			}
		}
		lastcode = v
	}

	tl := this.Tokenize(runestring.Cast(s))

	return tl
}

func (this *DialectLogo) UpdateRuntimeState(ent interfaces.Interpretable) {

	// Run state (227)
	ent.SetMemory(227, uint64(ent.GetState()))

	// Line/Stmt
	ent.SetMemory(117, uint64(ent.GetPC().Line))
	ent.SetMemory(118, uint64(ent.GetPC().Statement))

	// Data/Pointer
	ent.SetMemory(123, uint64(ent.GetDataRef().Line))
	ent.SetMemory(124, uint64(ent.GetDataRef().Statement))
	ent.SetMemory(125, uint64(ent.GetDataRef().Token))
	ent.SetMemory(126, uint64(ent.GetDataRef().SubIndex))

	// Onerr address
	ent.SetMemory(220, uint64(ent.GetErrorTrap().Line))

	// Loop stack
	data, _ := ent.GetLoopStack().MarshalBinary()
	for i, v := range data {
		ent.SetMemory(LOOPSTACK_ADDRESS+i, v)
	}

	// Call stack
	data, _ = ent.GetStack().MarshalBinary()
	for i, v := range data {
		ent.SetMemory(CALLSTACK_ADDRESS+i, v)
	}

	// Loop vars
	data = types.PackName(ent.GetLoopVariable(), 16)
	data = append(data, types.Float2uint(float32(ent.GetLoopStep())))
	for i, v := range data {
		ent.SetMemory(0xfd00+i, v)
	}

}

func (this *DialectLogo) DumpState(ent interfaces.Interpretable) {

	////fmt.Printf("- Entity State: %v\n", ent.GetState())
	////fmt.Printf("- Entity PC   : line %d, stmt %d\n", ent.GetPC().Line, ent.GetPC().Statement)
	////fmt.Printf("- Error trap  : line %d, stmt %d\n", ent.GetErrorTrap().Line, ent.GetErrorTrap().Statement)
	////fmt.Printf("- Data pointer: line %d, stmt %d, token %d, subindex %d\n", ent.GetDataRef().Line, ent.GetDataRef().Statement, ent.GetDataRef().Token, ent.GetDataRef().SubIndex)
	////fmt.Printf("- Loopstack   : %d entries, %f, %s\n", ent.GetLoopStack().Size(), ent.GetLoopStep(), ent.GetLoopVariable())
	////fmt.Printf("- Callstack   : %d entries\n", ent.GetStack().Size())
	////fmt.Printf("- Variables   : %d, %v\n", len(ent.GetLocal().Keys()), ent.GetLocal().Keys())
	////fmt.Printf("- Program size: %d bytes\n", ent.GetMemory(2048)-2049)

}

func (this *DialectLogo) PreFreeze(ent interfaces.Interpretable) {

	this.UpdateRuntimeState(ent)
	this.DumpState(ent)

}

func (this *DialectLogo) PostThaw(ent interfaces.Interpretable) {

	// reload softswitches
	_ = ent.GetMemory(65536 - 16304)

	data := make([]uint64, 0)
	MEMBASE := this.GetProgramStart(ent)
	e := int(ent.GetMemory(MEMBASE))
	for i := MEMBASE + 1; i < e; i++ {
		data = append(data, ent.GetMemory(i))
	}

	//	//fmt.Printf("PostThaw, says program ends at %d\n", e)

	// now unpack program again
	if len(data) > 0 {
		a := this.ParseMemoryRepresentation(data)
		ent.SetCode(&a)
	} else {
		ent.SetCode(types.NewAlgorithm())
	}

	//	fixMemoryPtrs(ent)

	//ent.GetLocal().Mgr = ent
	//ent.GetLocal().Defrost(true)

	// Cursor
	apple2helpers.ResyncCursor(ent)
	ent.SaveCPOS()

	// Now state registers
	ent.GetErrorTrap().Line = int(ent.GetMemory(220))

	ent.GetDataRef().Line = int(ent.GetMemory(123))
	ent.GetDataRef().Statement = int(ent.GetMemory(124))
	ent.GetDataRef().Token = int(ent.GetMemory(125))
	ent.GetDataRef().SubIndex = int(ent.GetMemory(126))

	// PC
	ent.GetPC().Line = int(ent.GetMemory(117))
	ent.GetPC().Statement = int(ent.GetMemory(118))

	// stacks here
	data = make([]uint64, 0)
	for i := 0; i < LOOPSTACK_MAX; i++ {
		data = append(data, ent.GetMemory(LOOPSTACK_ADDRESS+i))
	}
	_ = ent.GetLoopStack().UnmarshalBinary(data)

	data = make([]uint64, 0)
	for i := 0; i < CALLSTACK_MAX; i++ {
		data = append(data, ent.GetMemory(CALLSTACK_ADDRESS+i))
	}
	_ = ent.GetStack().UnmarshalBinary(data)

	// Fix stuff on stack
	c := ent.GetCode()
	for i := 0; i < ent.GetStack().Size(); i++ {
		ent.GetStack().Get(i).State = ent.GetState()
		ent.GetStack().Get(i).Locals = ent.GetLocal()
		ent.GetStack().Get(i).Code = c
	}

	// Get loop var
	data = make([]uint64, 0)
	for i := 0; i < 5; i++ {
		data = append(data, ent.GetMemory(0xfd00+i))
	}
	ent.SetLoopVariable(types.UnpackName(data[0:4]))
	ent.SetLoopStep(float64(types.Uint2Float(data[4])))

	// PC
	ent.SetState(types.EntityState(ent.GetMemory(227)))

	this.DumpState(ent)
}

func (this *DialectLogo) ThawVideoConfig(ent interfaces.Interpretable) {
	apple2helpers.RestoreSoftSwitches(ent)
}

func (this *DialectLogo) HomeLeft(ent interfaces.Interpretable) {
	apple2helpers.HomeLeft(ent)
}

func (this *DialectLogo) TokenListAsString(tokens types.TokenList) string {

	/* vars */
	var result string

	result = ""
	for _, tok1 := range tokens.Content {
		if result != "" {
			result = result + " "
		}
		if tok1.Type == types.LIST {
			// recurse!
			s := "[" + this.TokenListAsString(*tok1.List) + "]"
			result = result + s
		} else if tok1.Type == types.COMMANDLIST {
			// recurse!
			s := "(" + this.TokenListAsString(*tok1.List) + ")"
			result = result + s
		} else if (tok1.Type == types.KEYWORD) || (tok1.Type == types.FUNCTION) || (tok1.Type == types.DYNAMICKEYWORD) {
			result = result + strings.ToUpper(tok1.AsString())
		} else if tok1.Type == types.STRING {
			result = result + "\"" + tok1.Content
		} else {
			result = result + tok1.AsString()
		}
	}

	/* enforce non void return */
	return result

}

func (d *DialectLogo) GetWorkspaceBody(vars bool, filterProc string) []string {
	if filterProc != "" {
		p, ok := d.Driver.GetProc(filterProc)
		if !ok {
			return []string{
				"TO " + filterProc,
				"",
				"END",
			}
		}
		return p.GetCode()
	}

	lines := d.Driver.GetWorkspaceBody(true, vars)
	return lines
}

func (this *DialectLogo) QueueCommand(command string) {
	this.Driver.SendCommand(LogoDriverCommand(command))
}

func (this *DialectLogo) SaveState() {
	this.Driver.PauseExecution()
}

func (this *DialectLogo) RestoreState() {
	this.Driver.ResumeExecution()
	this.Driver.hasResumed = false
}

func (this *DialectLogo) SyntaxValid(s string) error {
	_, err := this.Driver.Parse(this.Lexer, s)
	return err
}

func (this *DialectLogo) GetLastCommand() string {
	v := this.LastCommand
	this.LastCommand = ""
	return v
}

func Fields(str string) []string {

	sepkeepers := ":=+*"

	chunk := ""
	parts := make([]string, 0)
	for _, ch := range str {
		switch {
		case ch == ' ':
			if chunk != "" {
				parts = append(parts, chunk)
				chunk = ""
			}
		case strings.IndexRune(sepkeepers, ch) > -1:
			if chunk != "" {
				parts = append(parts, chunk)
				chunk = ""
			}
			chunk = string(ch)
		default:
			if len(chunk) == 1 && strings.Index(sepkeepers, chunk) > -1 {
				parts = append(parts, chunk)
				chunk = ""
			}
			chunk += string(ch)
		}
	}

	if chunk != "" {
		parts = append(parts, chunk)
	}

	return parts
}

func in(str string, list []string) bool {
	for _, v := range list {
		if strings.ToLower(v) == strings.ToLower(str) {
			return true
		}
	}
	return false
}

func typeIn(t types.TokenType, list []types.TokenType) bool {
	for _, v := range list {
		if v == t {
			return true
		}
	}
	return false
}

const CTX_START_LINE = 0
const CTX_AFTER_KW = 1
const CTX_AFTER_ASSIGN = 2
const CTX_AFTER_FUNC = 3
const CTX_WANT_FILE = 4

func (this *DialectLogo) GetCompletions(ent interfaces.Interpretable, line runestring.RuneString, index int) (int, *types.TokenList) {
	r := types.NewTokenList()

	// 10 pr/int/#/
	if index > len(line.Runes) {
		return 0, r
	}
	tmp := string(line.Runes[0:index])

	if len(tmp) > 0 && tmp[len(tmp)-1] == 32 {
		return 0, r
	}

	var inqq bool
	var bc int
	for _, ch := range line.Runes[0:index] {
		switch {
		case ch == '{' || ch == '[' || ch == '(':
			bc++
		case ch == '}' || ch == ']' || ch == ')':
			bc--
		case ch == '"':
			inqq = !inqq
		}
	}
	if bc > 0 || inqq {
		return 0, r
	}

	p := Fields(tmp)
	if len(p) == 0 {
		return 0, r
	}
	// get last word part before insert pos
	last := strings.ToLower(p[len(p)-1])

	// now tokenize the rest - we need to know what if any the last token was
	rest := tmp[0 : len(tmp)-len(last)]
	tl := this.Tokenize(runestring.Cast(rest))

	tla := ent.SplitOnToken(*tl, *types.NewToken(types.SEPARATOR, ":"))
	tl = &tla[len(tla)-1]

	nca := []string{}
	efn := []string{}

	searchbase := last

	ctx := CTX_START_LINE

	if tl.Size() > 0 {
		t := tl.RPeek()
		f := tl.LPeek()

		// if typeIn(t.Type, []types.TokenType{types.KEYWORD}) {
		// 	ctx = CTX_AFTER_KW

		// } else if typeIn(t.Type, []types.TokenType{types.FUNCTION, types.PLUSFUNCTION}) {
		// 	ctx = CTX_AFTER_FUNC
		// } else if typeIn(t.Type, []types.TokenType{types.ASSIGNMENT, types.COMPARITOR}) {
		// 	ctx = CTX_AFTER_ASSIGN
		// } else if typeIn(t.Type, []types.TokenType{types.SEPARATOR}) {
		// 	ctx = CTX_START_LINE
		// }

		ctx = CTX_START_LINE

		if f.Type == types.KEYWORD && in(f.Content, efn) {
			ctx = CTX_WANT_FILE

			searchbase = strings.Replace(searchbase, "\"", "", -1)

			base := searchbase

			var addworkdir bool

			if !strings.HasPrefix(base, "/") {
				base = ent.GetWorkDir() + "/" + base
				addworkdir = true
			}

			// Do the search
			fmt.Printf("I will try autocomplete for files based on a path of [%s]\n", base)
			ditems, fitems := files.GetCompletions(base)
			for _, i := range ditems {
				isDir := true

				// strip auto prefix
				if addworkdir {
					i = i[len(ent.GetWorkDir()+"/"):]
				}

				if i != "/" && i != "" && !strings.HasSuffix(i, "..") && i != searchbase {
					if isDir {
						r.Push(types.NewToken(types.STRING, i+"/"))
					} else {
						r.Push(types.NewToken(types.STRING, i))
					}
					fmt.Printf("dir: %s\n", i)
				}
			}
			for _, i := range fitems {
				isDir := false

				// strip auto prefix
				if addworkdir {
					i = i[len(ent.GetWorkDir()+"/"):]
				}

				if i != "/" && i != "" && !strings.HasSuffix(i, "..") && i != searchbase {
					if isDir {
						r.Push(types.NewToken(types.STRING, i+"/"))
					} else {
						r.Push(types.NewToken(types.STRING, i))
					}
					fmt.Printf("file: %s\n", i)
				}
			}

			return len(searchbase), r

		} else if t.Type == types.KEYWORD {
			// last was keyword
			if strings.ToLower(t.Content) == "then" || strings.ToLower(t.Content) == "def" {
				ctx = CTX_START_LINE
			}

			if in(t.Content, nca) {
				return 0, r
			}
		} else {
			t = tl.LPeek()
			if in(t.Content, nca) {
				return 0, r
			}
		}
	} else {

	}

	// now we look for it as the start of something
	//
	// #1 keywords

	scom := make([]string, 0)

	if ctx == CTX_START_LINE {
		for k, _ := range this.Commands {
			scom = append(scom, k)
		}
		sort.Strings(scom)

		for _, k := range scom {
			//fmt.Printf("[%s] vs [%s]\n", k, last)
			if strings.HasPrefix(k, last) {
				// Add to list
				r.Push(types.NewToken(types.KEYWORD, k))
			}
		}
	}
	//
	// #2 functions
	scom = make([]string, 0)
	if ctx == CTX_AFTER_ASSIGN || ctx == CTX_AFTER_FUNC || ctx == CTX_AFTER_KW {
		for k, _ := range this.Functions {
			scom = append(scom, k)
		}
		sort.Strings(scom)
		for _, k := range scom {
			if strings.HasPrefix(k, last) {
				// Add to list
				r.Push(types.NewToken(types.FUNCTION, k))
			}
		}
	}
	// #3 Plus functions
	scom = make([]string, 0)
	for k, v := range this.PlusFunctions {
		//scom = append(scom, k)
		for kk, vv := range v {
			if !vv.IsHidden() {
				scom = append(scom, k+"."+kk)
			}
		}
	}
	sort.Strings(scom)
	for _, k := range scom {
		if strings.HasPrefix(k, last) {
			// Add to list
			r.Push(types.NewToken(types.PLUSFUNCTION, k))
		}
	}

	return len(last), r
}
