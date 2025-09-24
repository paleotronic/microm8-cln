package applesoft

import (
	"errors"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"

	s8webclient "paleotronic.com/api"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/dialect/plus"
	"paleotronic.com/core/exception"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/files"
	"paleotronic.com/fmt" //"os"
	"paleotronic.com/log"
	"paleotronic.com/runestring"
	"paleotronic.com/utils"
)

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
var reMultipleIF *regexp.Regexp

func init() {
	reSimpleLoop, _ = regexp.Compile("(.* )?for ([A-Za-z]+) = ([0-9]+) to ([0-9]+) : ([A-Za-z]+) = PEEK[(] - 16336 [)] : next( [A-Za-z]+)?( :.*)?")
	reAddSub, _ = regexp.Compile("(.* : )?([A-Za-z]+) = (PEEK[(] - 16336 [)])( [+-] PEEK[(] - 16336 [)])*( :.*)?")
	reNumberFloatE, _ = regexp.Compile("^[+-]?([0-9]+)([.]([0-9]+))?[eE]$")
	reNumberFloatES, _ = regexp.Compile("^[+-]?([0-9]+)([.]([0-9]+))?[eE][+-]$")
	reVarname, _ = regexp.Compile("^[@]?[a-zA-Z][a-zA-Z0-9.]*[$%#!]?$")
	reMultipleIF, _ = regexp.Compile("(?i)^IF[ ]+(.+)[ ]+THEN[ ]+(.+[:].+)$")
}

type DialectApplesoft struct {
	dialect.Dialect
	lastLineNumber int
}

func NewDialectApplesoft() *DialectApplesoft {
	this := &DialectApplesoft{}
	this.Dialect = *dialect.NewDialect()
	this.Init()
	this.Dialect.DefaultCost = 1000000000 / 800
	this.Throttle = 100.0
	this.GenerateNumericTokens()
	return this
}

func (this *DialectApplesoft) BeforeRun(caller interfaces.Interpretable) {
	fixMemoryPtrs(caller)
}

func (this *DialectApplesoft) CheckOptimize(lno int, s string, OCode types.Algorithm) {
	// stub does nothing

	//fmt.Println("in applesoft match")

	if m := reMultipleIF.FindStringSubmatch(s); len(m) > 0 {

		fmt.Printf("Multiple IF at line %d...\n", lno)

	} else if m := reSimpleLoop.FindStringSubmatch(s); len(m) > 0 {
		// (...:)? FOR ([A-Z]+) = ([0-9]+) TO ([0-9]+) : ([A-Z]+) = PEEK[(] -16336 [)] : NEXT( [A-Z]+)?( :.*)?
		//   1           2          3          4          5                                   6        7
		start := utils.StrToInt(m[3])
		end := utils.StrToInt(m[4])

		total_peeks := end - start + 1
		duration_ms := total_peeks * LOOP_PEEK_MS_PER
		freq_hz := LOOP_PEEK_FREQ

		alt_cmd := "@music.tone{" + utils.IntToStr(freq_hz) + "," + utils.IntToStr(duration_ms) + "}"
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

		total_peeks := strings.Count(s, "PEEK( -16336 )")
		duration_ms := total_peeks * ADDSUB_PEEK_MS_PER
		freq_hz := ADDSUB_PEEK_FREQ

		alt_cmd := "@music.tone{" + utils.IntToStr(freq_hz) + "," + utils.IntToStr(duration_ms) + "}"
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

func (this *DialectApplesoft) Evaluate(chunk string, tokens *types.TokenList) bool {

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

	if this.IsLabel(chunk) {
		tok.Type = types.LABEL
		tok.Content = strings.ToUpper(chunk)
	} else if this.IsFunction(chunk) {
		tok.Type = types.FUNCTION
		tok.Content = strings.ToTitle(chunk)
	} else if this.IsDynaCommand(chunk) {
		tok.Type = types.DYNAMICKEYWORD
		tok.Content = strings.ToLower(chunk)
	} else if this.IsDynaFunction(chunk) {
		tok.Type = types.DYNAMICFUNCTION
		tok.Content = strings.ToTitle(chunk)
	} else if this.IsKeyword(chunk) {
		tok.Type = types.KEYWORD
		tok.Content = strings.ToLower(chunk)

		/* determine if (we should stop parsing */

		z := this.Commands[strings.ToLower(chunk)]

		result = !(z.HasNoTokens())
	} else if this.IsLogic(chunk) {
		tok.Type = types.LOGIC
		tok.Content = strings.ToLower(chunk)
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
	} else if this.IsString(chunk) {
		tok.Type = types.STRING

		tok.Content = utils.Copy(chunk, 2, utils.Len(chunk)-2)
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

func (this *DialectApplesoft) IsVariableName(instr string) bool {

	return reVarname.MatchString(instr)

}

func (this *DialectApplesoft) Tokenize(s runestring.RuneString) *types.TokenList {

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
			}

		} else if this.IsQ(ch) && (!inqq) {
			if (len(chunk) > 0) && (!inq) {
				this.Evaluate(chunk, result)
				//				pchunk = chunk
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
				//				pchunk = chunk
				chunk = ""
			} else {
				chunk = chunk + string(ch)
			}
			inqq = !inqq
		} else {
			chunk = chunk + string(ch)

			/* break keywords out early */
			if this.Commands.ContainsKey(strings.ToLower(chunk)) {
				if (strings.ToLower(chunk) != "cat") && (strings.ToLower(chunk) != "go") && (strings.ToLower(chunk) != "to") && (strings.ToLower(chunk) != "on") && (strings.ToLower(chunk) != "hgr") && (strings.ToLower(chunk) != "at") && (strings.ToLower(chunk) != "gr") {
					cont = this.Evaluate(chunk, result)
					//					pchunk = chunk
					chunk = ""
				}
			}

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

func (this *DialectApplesoft) DecideValueIndex(hpop int, count int, values types.TokenList) int {
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

func (this *DialectApplesoft) HPOpIndex(tl types.TokenList) int {
	result := -1
	hs := 1
	var tt types.Token
	var sc int

	for i := 0; i <= tl.Size()-1; i++ {
		tt = *tl.Get(i)
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

func (this *DialectApplesoft) ParseTokensForResult(ent interfaces.Interpretable, tokens types.TokenList) (*types.Token, error) {

	/* vars */
	var result *types.Token
	var ops types.TokenList
	var values types.TokenList
	var tidx int
	var rbc int
	var sbc int
	//	var cbc int
	var i int
	var repeats int
	var blev int
	var tok *types.Token
	var lasttok *types.Token
	var ntok *types.Token
	var op types.Token
	var a types.Token
	var b types.Token
	var n string
	var rs string
	var v *types.Variable
	var exptok types.TokenList
	var subexpr types.TokenList
	var par types.TokenList
	var err bool
	var rr float64
	var aa float64
	var bb float64
	var dl []int
	var rrb bool
	var fun interfaces.Functioner
	var mafun interfaces.MultiArgumentFunction
	var tla types.TokenListArray
	//	var left bool
	var defindex bool
	var lastop, nlastop, p2nlastop bool
	var hpop int
	//	var hs int
	//	var sc int
	var sexpr string
	var serr error

	result = types.NewToken(types.INVALID, "")

	/* must be 1 || more tokens in list */
	if tokens.Size() == 0 {
		return result, nil
	}

	if ent.IsDebug() && (tokens.Size() > 1 || tokens.Get(0).Type == types.VARIABLE) {
		sexpr = this.TokenListAsString(tokens)
	}

	////fmt.Printf("Applesoft version of ParseTokensForResult() invoked for [%s]\n", ent.TokenListAsString(tokens))

	//System.Err.Println("*** Called to parse: "+ent.TokenListAsString(tokens));

	values = *types.NewTokenList()
	ops = *types.NewTokenList()

	// (* first fix: if (stream {s with an operator, prefix zero to the chain *);
	if tokens.Get(0).Type == types.OPERATOR {
		//writeln("BEEP");
		values.Push(types.NewToken(types.INTEGER, "0"))
	}

	/* init parser state */
	tidx = 0
	rbc = 0
	//	cbc = 0
	sbc = 0

	//fInterpreter.VDU.PutStr("Expression in: "+this.TokenListAsString(tokens)+PasUtil.CRLF);
	lasttok = nil
	lastop = false
	//nlasttok = nil
	//    nlastop = false
	/*main parse loop*/
	for tidx < tokens.Size() {

		//nlasttok = lasttok
		lasttok = tok
		p2nlastop = nlastop
		nlastop = lastop

		tok = tokens.Get(tidx)

		////fmt.Printf("tidx = %d, Type = %d, Content = %s\n", tidx, tok.Type, tok.Content)

		//System.Err.Println( "--------------> type of token at tidx "+tidx+" is "+tok.Type+"["+tok.Content+"]" );

		if tok.Type == types.LABEL {

			if (lastop == false) && (lasttok != nil) {
				ops.Push(types.NewToken(types.OPERATOR, "+"))
			}

			if ent.GetLabel(tok.Content[1:]) != 0 {
				// label
				ntok := types.NewToken(types.NUMBER, utils.IntToStr(ent.GetLabel(tok.Content[1:])))
				values.Push(ntok)
			}

			lastop = false

		} else if (tok.Type == types.NUMBER) || (tok.Type == types.INTEGER) || (tok.Type == types.STRING) || (tok.Type == types.BOOLEAN) {
			// fix for missing + || separators;
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
			} else if (lastop == true) && (lasttok != nil) && (lasttok.Content == "+") && (tok.Content == "+") {
				tidx++
				continue
			} else if (lastop == true) && (lasttok != nil) && (lasttok.Content == "+") && (tok.Content == "-") {
				values.Push(types.NewToken(types.NUMBER, "0"))
			} else if (lastop == true) && (lasttok != nil) && ((lasttok.Type == types.ASSIGNMENT) || (lasttok.Type == types.COMPARITOR)) && (tok.Content == "+") {
				tidx++
				continue
			} /*else if (lastop == true) && (lasttok != nil) && (lasttok.Content == "+") && (tok.Content == "-") {
				values.Push(types.NewToken(types.NUMBER, "0"))
			} else if (lastop == true) && (lasttok != nil) && (lasttok.Content == "-") && (tok.Content == "+") {
				values.Push(types.NewToken(types.NUMBER, "0"))
			}*/

			ops.Push(tok)
			lastop = true
		} else if (tok.Type == types.KEYWORD) && (strings.ToLower(tok.Content) != "fn") {
			return result, errors.New("SYNTAX ERROR")
		} else if (tok.Type == types.KEYWORD) && (strings.ToLower(tok.Content) == "fn") {

			// USER DEFINED FUNCTION;
			tidx++
			tok = tokens.Get(tidx) // func name

			n = strings.ToLower(tok.Content)

			//writeln( "--- Evaluating function: "+n );

			// fix for missing + || separators;
			if (lastop == false) && (lasttok != nil) {
				ops.Push(types.NewToken(types.OPERATOR, "+"))
			}
			//			//fmt.Println("MAFMAP: ", ent.GetMultiArgFunc())
			if !ent.GetMultiArgFunc().ContainsKey(n) {
				return result, exception.NewESyntaxError("unknown user function: " + n)
			}

			mafun = ent.GetMultiArgFunc().Get(strings.ToLower(tok.Content))

			//fInterpreter.VDU.PutStr(fun.Name+"(\" + IntToStr(length(fun.FunctionParams))+\")"+PasUtil.CRLF);

			if mafun.Arguments.Size() > 0 {

				/* okay must be bracket after */
				tidx = tidx + 1
				if tidx >= tokens.Size() {
					return result, exception.NewESyntaxError(n + " requires params (size)")
				}
				tok = tokens.Get(tidx)
				if (tok.Type != types.OBRACKET) || (tok.Content != "(") {
					return result, exception.NewESyntaxError(n + " requires params: " + tok.Content)
				}

				// must be an index;
				sbc = 1
				tidx = tidx + 1
				subexpr = *types.NewTokenList()

				for (tidx < tokens.Size()) && (sbc > 0) {
					tok = tokens.Get(tidx)
					//					//fmt.Printf("----> Token = %s\n", tok.Content)
					if (tok.Type == types.OBRACKET) && (tok.Content == "(") {
						sbc = sbc + 1
					}
					if tok.Type == types.FUNCTION {
						sbc = sbc + 1
					}
					if (tok.Type == types.CBRACKET) && (tok.Content == ")") {
						sbc = sbc - 1
					}
					if sbc > 0 {
						subexpr.Push(tok)
						//						//fmt.Println("Func: Tokens in subexpr", subexpr.Size())
					}
					if (tidx < tokens.Size()) && (sbc > 0) {
						tidx = tidx + 1
					}
				}

				/* now condense this list down */
				tla = types.NewTokenListArray()
				exptok = *types.NewTokenList()
				blev = 0

				//				//fmt.Printf("====> %s\n", ent.TokenListAsString(subexpr))

				for _, tok1 := range subexpr.Content {
					if (tok1.Type == types.SEPARATOR) && (tok1.Content == ",") && (blev == 0) {
						if exptok.Size() > 0 {
							//SetLength(tla, tla.Size()+1);
							tla = tla.Add(exptok)
							exptok = *types.NewTokenList()
						}
					} else {

						if (tok1.Type == types.OBRACKET) && (tok1.Content == "(") {
							blev = blev + 1
						}

						if tok1.Type == types.FUNCTION {
							blev = blev + 1
						}

						if (tok1.Type == types.CBRACKET) && (tok1.Content == ")") {
							blev = blev - 1
						}

						exptok.Push(tok1)
					}
				}

				if exptok.Size() > 0 {
					//SetLength(tla, tla.Size()+1);
					tla = tla.Add(exptok)
				}

				par = *types.NewTokenList()
				for _, exptok1 := range tla {
					if (exptok1.Size() == 1) && (exptok1.RPeek().Type == types.VARIABLE) {
						tok = exptok1.RPeek() // preserve var ref
					} else {
						tok, serr = this.ParseTokensForResult(ent, exptok1)
						if serr != nil {
							return tok, serr
						}
					}
					par.Push(tok)
					//FreeAndNil(exptok);
				}
			} else {
				par = *types.NewTokenList()
			}

			a, _ := mafun.Evaluate(ent, par)
			values.Push(&a)
			//Writeln( "--- User func yielded result: "+values.RPeek().Content );
			lastop = false

		} else if tok.Type == types.PLUSFUNCTION {

			////fmt.Println(tok.Content + " is a function")

			// fix for missing + || separators;
			if (lastop == false) && (lasttok != nil) {
				ops.Push(types.NewToken(types.OPERATOR, "+"))
			}

			fun, exists, ns, err := this.PlusFunctions.GetFunctionByNameContext(this.CurrentNameSpace, strings.ToLower(tok.Content))
			//fun, exists := this.PlusFunctions[strings.ToLower(tok.Content)]

			if fun == nil || !exists {
				return result, err
			}

			this.CurrentNameSpace = ns

			qq := fun.IsQuery()

			fun.SetQuery(true)

			//fInterpreter.VDU.PutStr(fun.Name+"(\" + IntToStr(length(fun.FunctionParams))+\")"+PasUtil.CRLF);

			if len(fun.FunctionParams()) > 0 {

				/* okay must be bracket after */
				if tidx >= tokens.Size() {
					return result, exception.NewESyntaxError(fun.GetName() + " requires params (size)")
				}

				// must be an index;
				sbc = 1
				tidx = tidx + 1
				subexpr = *types.NewTokenList()

				for (tidx < tokens.Size()) && (sbc > 0) {
					tok = tokens.Get(tidx)
					if (tok.Type == types.OBRACKET) && (tok.Content == "{") {
						sbc = sbc + 1
					}
					if tok.Type == types.PLUSFUNCTION {
						sbc = sbc + 1
					}
					if (tok.Type == types.CBRACKET) && (tok.Content == "}") {
						sbc = sbc - 1
					}
					if sbc > 0 {
						subexpr.Push(tok)
						////fmt.Println("Func: Tokens in subexpr", subexpr.Size())
					}
					if (tidx < tokens.Size()) && (sbc > 0) {
						tidx = tidx + 1
					}
				}

				if !fun.GetRaw() {

					/* now condense this list down */
					tla = types.NewTokenListArray()
					exptok = *types.NewTokenList()
					blev = 0

					for _, tok1 := range subexpr.Content {
						if (tok1.Type == types.SEPARATOR) && (tok1.Content == ",") && (blev == 0) {
							if exptok.Size() > 0 {
								//SetLength(tla, tla.Size()+1);
								tla = tla.Add(exptok)
								exptok = *types.NewTokenList()
							}
						} else {

							if (tok1.Type == types.OBRACKET) && (tok1.Content == "(") {
								blev = blev + 1
							}

							if tok1.Type == types.FUNCTION {
								blev = blev + 1
							}

							if (tok1.Type == types.CBRACKET) && (tok1.Content == ")") {
								blev = blev - 1
							}

							exptok.Push(tok1)
						}
					}

					if exptok.Size() > 0 {
						//SetLength(tla, tla.Size()+1);
						tla = tla.Add(exptok)
						log.Println("exptok:", ent.TokenListAsString(exptok))
					}

					par = *types.NewTokenList()
					for _, exptok1 := range tla {
						log.Println(ent.TokenListAsString(exptok1))
						//writeln( fun.Name, ": ", this.TokenListAsString(exptok) );
						if fun.GetRaw() {
							if par.Size() > 0 {
								par.Push(types.NewToken(types.SEPARATOR, ","))
							}
							for _, tok1 := range exptok1.Content {
								par.Push(tok1)
							}
						} else {
							tok, serr = this.ParseTokensForResult(ent, exptok1)
							if serr != nil {
								return tok, serr
							}
							par.Push(tok)
							log.Println("Seeding to function params: ", tok)
						}
						//FreeAndNil(exptok);
					}

				} else {
					par = subexpr
				}
			} else {
				par = *types.NewTokenList()
			}

			fun.SetEntity(ent)
			fun.FunctionExecute(&par)
			fun.SetQuery(qq)
			values.Push(fun.GetStack().Pop())
			lastop = false

		} else if tok.Type == types.FUNCTION {

			////fmt.Println(tok.Content + " is a function")

			// fix for missing + || separators;
			if (lastop == false) && (lasttok != nil) {
				ops.Push(types.NewToken(types.OPERATOR, "+"))
			}

			fun = this.Functions.Get(strings.ToLower(tok.Content))
			if fun == nil {
				return result, exception.NewESyntaxError("unknown function: " + tok.Content)
			}

			//fInterpreter.VDU.PutStr(fun.Name+"(\" + IntToStr(length(fun.FunctionParams))+\")"+PasUtil.CRLF);

			if len(fun.FunctionParams()) > 0 {

				/* okay must be bracket after */
				if tidx >= tokens.Size() {
					return result, exception.NewESyntaxError(fun.GetName() + " requires params (size)")
				}

				// must be an index;
				sbc = 1
				tidx = tidx + 1
				subexpr = *types.NewTokenList()

				for (tidx < tokens.Size()) && (sbc > 0) {
					tok = tokens.Get(tidx)
					////fmt.Printf("----> Token = %s\n", tok.Content)
					if (tok.Type == types.OBRACKET) && (tok.Content == "(") {
						sbc = sbc + 1
					}
					if tok.Type == types.FUNCTION {
						sbc = sbc + 1
					}
					if (tok.Type == types.CBRACKET) && (tok.Content == ")") {
						sbc = sbc - 1
					}
					if sbc > 0 {
						subexpr.Push(tok)
						////fmt.Println("Func: Tokens in subexpr", subexpr.Size())
					}
					if (tidx < tokens.Size()) && (sbc > 0) {
						tidx = tidx + 1
					}
				}

				/* now condense this list down */
				tla = types.NewTokenListArray()
				exptok = *types.NewTokenList()
				blev = 0

				////fmt.Printf("====> %s\n", ent.TokenListAsString(subexpr))

				for _, tok1 := range subexpr.Content {
					if (tok1.Type == types.SEPARATOR) && (tok1.Content == ",") && (blev == 0) {
						if exptok.Size() > 0 {
							//SetLength(tla, tla.Size()+1);
							tla = tla.Add(exptok)
							exptok = *types.NewTokenList()
						}
					} else {

						if (tok1.Type == types.OBRACKET) && (tok1.Content == "(") {
							blev = blev + 1
						}

						if tok1.Type == types.FUNCTION {
							blev = blev + 1
						}

						if (tok1.Type == types.CBRACKET) && (tok1.Content == ")") {
							blev = blev - 1
						}

						exptok.Push(tok1)
					}
				}

				if exptok.Size() > 0 {
					//SetLength(tla, tla.Size()+1);
					tla = tla.Add(exptok)
					log.Println("exptok:", ent.TokenListAsString(exptok))
				}

				par = *types.NewTokenList()
				for _, exptok1 := range tla {
					log.Println(ent.TokenListAsString(exptok1))
					//writeln( fun.Name, ": ", this.TokenListAsString(exptok) );
					if fun.GetRaw() {
						if par.Size() > 0 {
							par.Push(types.NewToken(types.SEPARATOR, ","))
						}
						for _, tok1 := range exptok1.Content {
							par.Push(tok1)
						}
					} else {
						tok, serr = this.ParseTokensForResult(ent, exptok1)
						if serr != nil {
							return tok, serr
						}
						par.Push(tok)
						log.Println("Seeding to function params: ", tok)
					}
					//FreeAndNil(exptok);
				}
			} else {
				par = *types.NewTokenList()
			}

			fun.SetEntity(ent)
			e := fun.FunctionExecute(&par)
			if e != nil {
				return result, e
			}
			values.Push(fun.GetStack().Pop())
			lastop = false
		} else if tok.Type == types.VARIABLE {

			// fix for missing + || separators;
			if (lastop == false) && (lasttok != nil) {
				ops.Push(types.NewToken(types.OPERATOR, "+"))
			}

			//~ if ent.GetLabel( tok.Content ) != 0 {
			//~ // label
			//~ ntok = types.NewToken( types.NUMBER, utils.IntToStr(ent.GetLabel( tok.Content )) )
			//~ values.Push(ntok)
			//~ tidx = tidx + 1
			//~ continue
			//~ }

			/* first try entity local */
			n = strings.ToLower(tok.Content)
			v = nil
			if ent.GetLocal().ContainsKey(n) {
				v = ent.GetLocal().Get(n)
			} else if ent.ExistsVar(n) {
				//v = this.Producer.Global.Get(n);
				v = ent.GetVar(n)
			}
			/* fall out if (var does not exist */
			if v == nil {
				//System.Out.Println(n+" is undefined, defining it");
				//result.Type = types.INVALID;
				//return result;
				cvt := types.VT_STRING
				if n[len(n)-1] == '$' {
					cvt = types.VT_STRING
				} else {
					cvt = types.VT_FLOAT
				}
				//				d := ""
				//int adl[] = make([]int, 1);
				//adl[0] = this.ArrayDimDefault+1;
				//ent.CreateVarLower( strings.ToLower(n), types.NewVariable(strings.ToLower(n), cvt, d, true, adl) );
				//apple2helpers.PutStr(caller,"after create");
				//ent.GetVar( strings.ToLower(n) ).Owner = ent.Name;
				//v = ent.GetVar( strings.ToLower(n) );
				if cvt == types.VT_FLOAT {
					values.Push(types.NewToken(types.NUMBER, "0"))
				} else {
					values.Push(types.NewToken(types.STRING, ""))
				}

				//System.Err.Println("Defaulting var with name "+n+" to "+values.RPeek().Content);

				tidx = tidx + 1

				if (tidx < tokens.Size()) && (tokens.Get(tidx).Type == types.OBRACKET) {
					sbc = 1
					tidx = tidx + 1
					//subexpr.Push( types.NewToken(types.OBRACKET, '(') );
					for (tidx < tokens.Size()) && (sbc > 0) {
						tok = tokens.Get(tidx)
						if (tok.Type == types.OBRACKET) && (tok.Content == "(") {
							sbc = sbc + 1
						}
						if (tok.Type == types.CBRACKET) && (tok.Content == ")") {
							sbc = sbc - 1
						}
						//if ((sbc > 0))
						//	  subexpr.Push(tok);
						if (tidx < tokens.Size()) && (sbc > 0) {
							tidx = tidx + 1
						}
					}

					tidx = tidx + 1
				}

				lastop = false
				lasttok = values.RPeek()
				continue
			} else {
				log.Println("V is not NIL")
			}

			if v.IsArray() || !v.IsArray() {
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

				// must be an index;
				subexpr = *types.NewTokenList()
				if !defindex {
					log.Println("Moo frog soup")
					sbc = 1
					tidx = tidx + 1
					subexpr.Push(types.NewToken(types.OBRACKET, "("))
					for (tidx < tokens.Size()) && (sbc > 0) {
						tok = tokens.Get(tidx)
						if tok.Type == types.FUNCTION {
							sbc = sbc + 1
						}
						if (tok.Type == types.OBRACKET) && (tok.Content == "(") {
							sbc = sbc + 1
						}
						if (tok.Type == types.CBRACKET) && (tok.Content == ")") {
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

				dl = make([]int, 1)
				if !defindex {
					subexpr.Push(types.NewToken(types.CBRACKET, ")"))
					//ent.PutStr("*** VAR INDICES ARE ["+this.TokenListAsString(subexpr)+']'+PasUtil.CRLF );

					log.Println("Array Indices:", ent.TokenListAsString(subexpr))

					dl, _ = ent.IndicesFromTokens(subexpr, "(", ")")
					//FreeAndNil(subexpr);
					s, _ := v.GetContentByIndex(this.ArrayDimDefault, this.ArrayDimMax, dl)
					ntok = types.NewToken(types.INVALID, s)
				} else {
					s, _ := v.GetContentScalar()
					ntok = types.NewToken(types.INVALID, s)
				}

				/* var exists */
				//  VariableType == (vtString, vtBoolean, vtFloat, vtInteger, vtExpression);
				switch v.Kind { /* FIXME - Switch statement needs cleanup */
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
						s, _ := v.GetContentByIndex(this.ArrayDimDefault, this.ArrayDimMax, dl)
						exptok = *this.Tokenize(runestring.Cast(s))
						ntok, serr = this.ParseTokensForResult(ent, exptok)
						if serr != nil {
							return ntok, serr
						}
					}
				}
				//this.VDU.PutStr("// Adding value from array "+ntok.Content);

				/* ntok special case */

				if p2nlastop && ops.Size() >= 2 && ops.Get(ops.Size()-1).Content == "-" && strings.Contains("*/^", ops.Get(ops.Size()-2).Content) {
					_ = ops.Pop() // remove '-'
					ntok.Content = strconv.FormatFloat(
						0-ntok.AsExtended(),
						'f',
						-1,
						64,
					)
				}

				values.Push(ntok)

			} else {

				/* var exists */
				s, _ := v.GetContentScalar()
				ntok = types.NewToken(types.INVALID, s)
				//  VariableType == (vtString, vtBoolean, vtFloat, vtInteger, vtExpression);
				switch v.Kind { /* FIXME - Switch statement needs cleanup */
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
						s, _ = v.GetContentScalar()
						exptok = *this.Tokenize(runestring.Cast(s))
						ntok, serr = this.ParseTokensForResult(ent, exptok)
						if serr != nil {
							return ntok, serr
						}
					}
				}

				/* ntok special case */

				if p2nlastop && ops.Size() >= 2 && ops.Get(ops.Size()-1).Content == "-" && strings.Contains("*/^", ops.Get(ops.Size()-2).Content) {
					_ = ops.Pop() // remove '-'
					ntok.Content = strconv.FormatFloat(
						0-ntok.AsExtended(),
						'f',
						-1,
						64,
					)
				}

				values.Push(ntok)

			}

			lastop = false
		} else if (tok.Type == types.OBRACKET) && (tok.Content == "(") {
			rbc = 1
			tidx = tidx + 1
			subexpr = *types.NewTokenList()
			for (tidx < tokens.Size()) && (rbc > 0) {
				tok = tokens.Get(tidx)
				if (tok.Type == types.OBRACKET) && (tok.Content == "(") {
					rbc = rbc + 1
				}
				if tok.Type == types.FUNCTION {
					rbc = rbc + 1
				}
				if (tok.Type == types.CBRACKET) && (tok.Content == ")") {
					rbc = rbc - 1
				}
				if rbc > 0 {
					subexpr.Push(tok)
				}
				if (tidx < tokens.Size()) && (rbc > 0) {
					tidx = tidx + 1
				}
			}

			//writeln( "*** Must parse bracketed subexpression: ", this.TokenListAsString(subexpr) );
			//writeln( "=== rbc == ",rbc );

			if rbc > 0 {
				return result, exception.NewESyntaxError("SYNTAX ERROR")
			}

			/* solve the expression */
			ntok, serr = this.ParseTokensForResult(ent, subexpr)
			if serr != nil {
				return ntok, serr
			}
			//FreeAndNil(subexpr);
			values.Push(ntok)
			lastop = false
		}

		//nlasttok = lasttok
		//lasttok = tok
		tidx = tidx + 1
	}

	/* now we have ops, values etc - lets actually parse the expression */
	//writeln("Op stack");
	// for _, ntok := range ops
	//    Writeln("OP: ", ntok.Type, ", ", ntok.Content);
	// for _, ntok := range values
	//    Writeln("VALUE: ", ntok.Type, ", ", ntok.Content);
	//writeln("--------");

	/* This bit is some magic || something yea!!! Whoo ---------------- */
	/* End magic bits ------------------------------------------------- */

	/* process */
	err = false

	for (ops.Size() > 0) && (!err) {
		//ent.PutStr(">> Ops: "+ent.TokenListAsString(ops)+PasUtil.CRLF);
		//ent.PutStr(">> Val: "+ent.TokenListAsString(values)+PasUtil.CRLF);

		hpop = this.HPOpIndex(ops)
		op = *ops.Remove(hpop)

		//Writeln("Processing operator: ",op.Content);
		if op.Type == types.LOGIC {

			if strings.ToLower(op.Content) == "not" {

				//a = values.Pop();
				hpop = this.DecideValueIndex(hpop, 1, values)
				a = *values.Remove(hpop)

				//ent.PutStr(op.Content+" "+a.Content+PasUtil.CRLF);

				//}

				if a.AsInteger() != 0 {
					ntok = types.NewToken(types.NUMBER, "0")
				} else {
					ntok = types.NewToken(types.NUMBER, "1")
				}

				values.Insert(hpop, ntok)
			} else if strings.ToLower(op.Content) == "and" {
				//b = values.Pop();
				//a = values.Pop();
				hpop = this.DecideValueIndex(hpop, 2, values)
				a = *values.Remove(hpop)
				b = *values.Remove(hpop)

				//ent.PutStr(a.Content+" "+op.Content+" "+b.Content+PasUtil.CRLF);

				//this.VDU.PutStr(a.Content+" \"+op.Content+\" "+b.Content+PasUtil.CRLF);

				//vv := ((a.AsInteger() != 0) && (b.AsInteger() != 0)) ? 1 : 0

				var vv int
				if (a.AsInteger() != 0) && (b.AsInteger() != 0) {
					vv = 1
				}

				//values.Push( types.NewToken(types.NUMBER, IntToStr(a.AsInteger() && b.AsInteger())) );
				ntok = types.NewToken(types.NUMBER, utils.IntToStr(vv))
				values.Insert(hpop, ntok)
			} else if strings.ToLower(op.Content) == "or" {
				//b = values.Pop();
				//a = values.Pop();
				hpop = this.DecideValueIndex(hpop, 2, values)
				a = *values.Remove(hpop)
				b = *values.Remove(hpop)

				//ent.PutStr(a.Content+" "+op.Content+" "+b.Content+PasUtil.CRLF);

				//vv := ((a.AsInteger() != 0) || (b.AsInteger() != 0)) ? 1 : 0

				var vv int
				if (a.AsInteger() != 0) || (b.AsInteger() != 0) {
					vv = 1
				}

				//values.Push( types.NewToken(types.NUMBER, IntToStr(a.AsInteger() || b.AsInteger())) );
				ntok = types.NewToken(types.NUMBER, utils.IntToStr(vv))
				values.Insert(hpop, ntok)
			} else if strings.ToLower(op.Content) == "xor" {
				//b = values.Pop();
				//a = values.Pop();
				hpop = this.DecideValueIndex(hpop, 2, values)
				a = *values.Remove(hpop)
				b = *values.Remove(hpop)

				//ent.PutStr(a.Content+" "+op.Content+" "+b.Content+PasUtil.CRLF);

				//values.Push( types.NewToken(types.NUMBER, IntToStr(a.AsInteger() xor b.AsInteger())) );
				ntok = types.NewToken(types.NUMBER, utils.IntToStr(a.AsInteger()^b.AsInteger()))
				values.Insert(hpop, ntok)
			}

		} else if (op.Type == types.COMPARITOR) || (op.Type == types.ASSIGNMENT) {

			if values.Size() < 2 {
				return result, nil
			}

			//writeln("@@@@@@@@@@@@@@@@@@@@@ COMPARE");

			//if (hpop == values.Size()-1)
			//	hpop--;
			hpop = this.DecideValueIndex(hpop, 2, values)
			a = *values.Remove(hpop)
			b = *values.Remove(hpop)

			//ent.PutStr(a.Content+" "+op.Content+" "+b.Content+PasUtil.CRLF);

			//this.VDU.PutStr(a.Content+" \"+op.Content+\" "+b.Content+PasUtil.CRLF);

			if a.IsNumeric() && b.IsNumeric() {

				aa = float64(a.AsNumeric())
				bb = float64(b.AsNumeric())
				rrb = false

				//writeln("====================> NUMBER COMPARE a == [",a.Content,"], b == [",b.Content,']');

				if op.Content == ">" {
					rrb = (aa > bb)
				} else if op.Content == "<" {
					rrb = (aa < bb)
				} else if (op.Content == ">=") || (op.Content == "=>") {
					rrb = (aa >= bb)
				} else if (op.Content == "<=") || (op.Content == "=<") {
					rrb = (aa <= bb)
				} else if (op.Content == "<>") || (op.Content == "><") {
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
				ntok = types.NewToken(types.NUMBER, n)

				//}
				values.Insert(hpop, ntok)
			} else if (!a.IsNumeric()) || (!b.IsNumeric()) {

				//				tt := types.NUMBER /* most results are string */
				rrb = false

				if op.Content == ">" {
					rrb = (a.Content > b.Content)
				} else if op.Content == "<" {
					rrb = (a.Content < b.Content)
				} else if op.Content == "<>" || op.Content == "><" {
					rrb = (a.Content != b.Content)
				} else if op.Content == ">=" || op.Content == "=>" {
					rrb = (a.Content >= b.Content)
				} else if op.Content == "<=" || op.Content == "=<" {
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
				ntok = types.NewToken(types.NUMBER, n)

				//}
				values.Insert(hpop, ntok)

			}

		} else if op.Type == types.OPERATOR {

			//writeln("Currently ",values.Size()," values in stack... About to pop 2");

			if values.Size() < 2 {
				return result, nil
			}

			//if (hpop == values.Size()-1)
			//	hpop--;
			hpop = this.DecideValueIndex(hpop, 2, values)
			a = *values.Remove(hpop)
			b = *values.Remove(hpop)

			//ent.PutStr(a.Content+" "+op.Content+" "+b.Content+PasUtil.CRLF);

			if ent.IsDebug() {
				ent.Log("OP", a.Content+" "+op.Content+" "+b.Content)
			}

			//  ent.PutStr("Op is: "+a.Content+op.Content+b.Content+PasUtil.CRLF);

			if a.IsNumeric() && b.IsNumeric() {

				aa = a.AsNumeric64()
				bb = b.AsNumeric64()
				rr = 0

				////fmt.Println("OP: " + a.Content + " " + op.Content + " " + b.Content)

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

				//	n = utils.FormatFloat("", rr)

				n = utils.StrToFloatStrApple(strconv.FormatFloat(rr, 'f', -1, 64))

				ntok = types.NewToken(types.NUMBER, n)

				////fmt.Println("==: " + n)
				////fmt.Println("=~", rr)

				//}
				values.Insert(hpop, ntok)
			} else if (a.IsNumeric() && (!b.IsNumeric())) || (b.IsNumeric() && (!a.IsNumeric())) {

				tt := types.STRING /* most results are string */

				if op.Content == "+" {
					rs = a.Content + b.Content
				} else if op.Content == "-" {
					// remove b from a, ignoring case;
					// rs = StringReplace( a.Content, b.Content, "", [rfReplaceAll, rfIgnoreCase] );
					// rs = a.Content.ReplaceAll(b.Content, "")
				} else if op.Content == "*" {
					repeats = 0
					rs = ""
					n = ""
					if a.IsNumeric() {
						repeats = (int)(a.AsNumeric())
						n = b.Content
					} else {
						repeats = (int)(b.AsNumeric())
						n = a.Content
					}
					for i = 1; i <= repeats; i++ {
						rs = rs + n
					}
				} else if op.Content == "/" {
					// divide string by substring - count occurrences;
					tt = types.INTEGER

					repeats = 0

					rs = utils.IntToStr(repeats)
				} else {
					err = true
					break
				}

				ntok = types.NewToken(tt, rs)
				//writeln("CREATE RESULT: ", rs);

				//}
				values.Insert(hpop, ntok)

			} else if (!a.IsNumeric()) && (!b.IsNumeric()) {

				tt := types.STRING /* most results are string */

				if op.Content == "+" {
					rs = a.Content + b.Content
				} else if op.Content == "-" {
					// remove b from a, ignoring case;
					// rs = StringReplace( a.Content, b.Content, "", [rfReplaceAll, rfIgnoreCase] );
					//rs = a.Content.ReplaceAll(b.Content, "")
				} else if op.Content == "/" {
					// divide string by substring - count occurrences;
					tt = types.INTEGER

					repeats = 0

					rs = utils.IntToStr(repeats)
				} else {
					err = true
					break
				}

				ntok = types.NewToken(tt, rs)

				//}
				values.Insert(hpop, ntok)
			}

		}

	}

	if err || (values.Size() > 1) {
		if ent.IsDebug() && (tokens.Size() > 1 || tokens.Get(0).Type == types.VARIABLE) {
			ent.Log("EVAL", sexpr+" == Error!")
		}
		return result, exception.NewESyntaxError("Syntax error")
	}

	result = values.Pop()

	if result == nil {
		result = types.NewToken(types.NUMBER, "0")
	}

	result = types.NewToken(result.Type, ""+result.Content)

	//	if result.Type == types.NUMBER {
	//		 clip places
	//		result.Content = types.NewDecimalFormat("#.########").Format(result.AsExtended())
	//		//fmt.Printf("NRESULT = %s\n", result.Content)
	//        result.Content = utils.StrToFloatStrApple(result.Content)
	//	}

	//System.Err.Println("return "+result.Content);

	// removed free call here;
	// removed free call here;

	if ent.IsDebug() && (tokens.Size() > 1 || tokens.Get(0).Type == types.VARIABLE) {
		ent.Log("EVAL", sexpr+" == "+result.AsString())
		////fmt.Println(sexpr + " == " + result.AsString())
	}

	/* enforce non void return */
	return result, nil

}

func (this *DialectApplesoft) HandleException(ent interfaces.Interpretable, e error) {

	var msg string

	fmt.Printf("-> HandleException(%s) called from entity in slot %d\n", e.Error(), ent.GetMemIndex())

	apple2helpers.NLIN(ent)

	msg = e.Error()

	if (ent.GetState() == types.RUNNING) || (ent.GetState() == types.DIRECTRUNNING) || (ent.GetState() == types.STOPPED) {
		if !ent.HandleError() {
			apple2helpers.PutStr(ent, "?"+strings.ToUpper(msg))
			if (ent.GetState() == types.RUNNING) && (ent.GetPC().Line != 0) {
				apple2helpers.PutStr(ent, " AT LINE "+utils.IntToStr(ent.GetPC().Line))
			}
			ent.Halt()
		}
	}

	r, g, b, a := ent.GetMemoryMap().GetBGColor(ent.GetMemIndex())
	ent.GetMemoryMap().SetBGColor(ent.GetMemIndex(), 255, 0, 0, 255)
	apple2helpers.Beep(ent)
	ent.GetMemoryMap().SetBGColor(ent.GetMemIndex(), r, g, b, a)

	apple2helpers.PutStr(ent, "\r\n")

}

func (this *DialectApplesoft) Parse(ent interfaces.Interpretable, s string) error {

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

			fixMemoryPtrs(ent)
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
				}
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

			if settings.AutosaveFilename[ent.GetMemIndex()] != "" {
				data := this.GetWorkspace(ent)
				files.AutoSave(ent.GetMemIndex(), data)
			}

			data := this.GetMemoryRepresentation(ent.GetCode())

			// write to memory
			MEMBASE := this.GetProgramStart(ent)

			// write to memory
			for i, v := range data {
				ent.SetMemory(MEMBASE+1+i, v)
			}
			ent.SetMemory(MEMBASE, uint64(MEMBASE+1+len(data)))
			fixMemoryPtrs(ent)
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

	//ent.PutStr(fmt.Sprintf("Interpreter state is now %v\r\n", ent.GetState()))

	return nil

}

func (this *DialectApplesoft) InitVDU(v interfaces.Interpretable, promptonly bool) {

	/* vars */
	apple2helpers.TEXT40(v)

	v.SetPrompt("]")
	v.SetTabWidth(16)

	if !promptonly {
		apple2helpers.Clearscreen(v)
		apple2helpers.Gotoxy(v, 0, 0)
		apple2helpers.PutStr(v, "\r\n")
		apple2helpers.PutStr(v, "Floating Point microBASIC\r\n")
		apple2helpers.PutStr(v, "_  ^_^  _\r\n")
		apple2helpers.PutStr(v, " \\('_')/ \r\n")
		//		//fmt.Printf( "CX = %d, CY = %d\n", v.GetCursorX(), v.GetCursorY() )
		v.SetNeedsPrompt(true)
	}

	//settings.SlotZPEmu[v.GetMemIndex()] = !settings.PureBoot(v.GetMemIndex())

	v.SaveCPOS()

	//v.CreateVar(
	//	"speed",
	//	*types.NewVariableP("speed", types.VT_FLOAT, "255", true),
	//)

	v.SetMemory(228, 255)

	// set program pointers
	v.SetMemory(103, 1)
	v.SetMemory(104, 8)

	settings.SpecName[v.GetMemIndex()] = "Floating Point microBASIC"
	settings.SetSubtitle(settings.SpecName[v.GetMemIndex()])

	v.GetMemoryMap().WriteGlobal(v.GetMemIndex(), v.GetMemoryMap().MEMBASE(v.GetMemIndex())+memory.OCTALYZER_CAMERA_GFX_BASE+0, uint64(types.CC_ResetAll))

	v.SetNeedsPrompt(true)

}

func (this *DialectApplesoft) ProcessDynamicCommand(ent interfaces.Interpretable, cmd string) error {

	//	////fmt.Printf("In ProcessDynamicCommand for [%s]\n", cmd)

	s8webclient.CONN.LogMessage("UCL", "Dos command: "+cmd)

	if utils.Copy(cmd, 1, 4) == "BRUN" {

		fmt.Println(cmd)

		parts := strings.Split(strings.Trim(utils.Delete(cmd, 1, 4), " "), ",")
		addr := 16384
		length := -1
		offset := 0

		filename := parts[0]
		filename = strings.ToLower(filename)
		//System.Out.Println("will try to load "+filename);
		var hasaddr bool
		for i := 1; i < len(parts); i++ {
			tmp := strings.ToUpper(parts[i])
			if tmp[0] == 'A' {
				tmp = utils.Delete(tmp, 1, 1)
				addr, _ = utils.SuperStrToInt(tmp)
				hasaddr = true
			} else if tmp[0] == 'L' {
				tmp = utils.Delete(tmp, 1, 1)
				length, _ = utils.SuperStrToInt(tmp)
			} else if tmp[0] == 'B' {
				tmp = utils.Delete(tmp, 1, 1)
				offset, _ = utils.SuperStrToInt(tmp)
			}
		}

		//if !strings.HasSuffix(filename, ".s") {
		//filename += ".s"
		//}

		// try original file first, the other
		fmt.Printf("Checking for %s\n", filename)
		if !files.ExistsViaProvider(ent.GetWorkDir(), filename) {

			binEx := files.GetTypeBin()

			found := false
			for _, v := range binEx {

				fmt.Printf("Checking for %s.%s\n", filename, v.Ext)

				if files.ExistsViaProvider(ent.GetWorkDir(), filename+"."+v.Ext) {
					filename += "." + v.Ext
					found = true
					break
				}
			}

			if !found {
				return exception.NewESyntaxError("FILE NOT FOUND")
			}
		}

		fmt.Printf("Reading %s...\n", filename)
		data, err := files.ReadBytesViaProvider(ent.GetWorkDir(), filename)
		if offset > 0 && offset <= len(data.Content) {
			data.Content = data.Content[offset:]
		}
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
		if !hasaddr && data.Address != 0 {
			addr = data.Address
		}
		ent.GetMemoryMap().BlockWritePr(ent.GetMemIndex(), addr, rawdata)

		//System.Out.Println("OK");
		// disable zero page
		apple2helpers.GetCPU(ent).BasicMode = false
		apple2helpers.DoCall(addr, ent, true)

		return nil

	}

	if utils.Copy(cmd, 1, 5) == "BLOAD" {

		parts := strings.Split(strings.Trim(utils.Delete(cmd, 1, 5), " "), ",")
		addr := 16384
		length := -1
		offset := 0

		filename := parts[0]
		filename = strings.ToLower(filename)
		//System.Out.Println("will try to load "+filename);
		var hasaddr bool
		for i := 1; i < len(parts); i++ {
			tmp := strings.ToUpper(parts[i])
			if tmp[0] == 'A' {
				tmp = utils.Delete(tmp, 1, 1)
				addr, _ = utils.SuperStrToInt(tmp)
				hasaddr = true
			} else if tmp[0] == 'L' {
				tmp = utils.Delete(tmp, 1, 1)
				length, _ = utils.SuperStrToInt(tmp)
			} else if tmp[0] == 'B' {
				tmp = utils.Delete(tmp, 1, 1)
				offset, _ = utils.SuperStrToInt(tmp)
			}
		}

		//if !strings.HasSuffix(filename, ".s") {
		//filename += ".s"
		//}

		// try original file first, the other
		if !files.ExistsViaProvider(ent.GetWorkDir(), filename) {

			binEx := files.GetTypeBin()

			found := false
			for _, v := range binEx {
				if files.ExistsViaProvider(ent.GetWorkDir(), filename+"."+v.Ext) {
					filename += "." + v.Ext
					found = true
					break
				}
			}

			if !found {
				return exception.NewESyntaxError("FILE NOT FOUND")
			}
		}

		data, err := files.ReadBytesViaProvider(ent.GetWorkDir(), filename)
		if offset > 0 && offset <= len(data.Content) {
			data.Content = data.Content[offset:]
		}
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
		if !hasaddr && data.Address != 0 {
			addr = data.Address
		}
		//log2.Printf("blockwrite of %d bytes at %.6x", len(rawdata), addr)
		ent.GetMemoryMap().BlockWrite(ent.GetMemIndex(), addr, rawdata)

		//System.Out.Println("OK");

		return nil

	}

	if utils.Copy(cmd, 1, 6) == "VERIFY" {

		parts := strings.Split(strings.Trim(utils.Delete(cmd, 1, 6), " "), ",")
		//addr := 16384
		//length := -1
		offset := 0

		filename := parts[0]
		filename = strings.ToLower(filename)

		// try original file first, the other
		if !files.ExistsViaProvider(ent.GetWorkDir(), filename) {

			binEx := files.GetTypeAll()

			found := false
			for _, v := range binEx {
				if files.ExistsViaProvider(ent.GetWorkDir(), filename+"."+v.Ext) {
					filename += "." + v.Ext
					found = true
					break
				}
			}

			if !found {
				return exception.NewESyntaxError("FILE NOT FOUND")
			}
		}

		data, err := files.ReadBytesViaProvider(ent.GetWorkDir(), filename)
		if offset > 0 && offset <= len(data.Content) {
			data.Content = data.Content[offset:]
		}
		if err != nil {
			return err
		}

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
				addr, _ = utils.SuperStrToInt(tmp)
			} else if tmp[0] == 'L' {
				tmp = utils.Delete(tmp, 1, 1)
				length, _ = utils.SuperStrToInt(tmp)
			}
		}

		if !strings.HasSuffix(filename, ".bin") {
			filename += ".bin"
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
		//fn := ent.GetWorkDir() + "/" + strings.ToLower(strings.Trim(utils.Delete(cmd, 1, 5), " ")) + ".d"
		//fn := files.GetUserPath(files.BASEDIR, []string{ent.GetWorkDir(), strings.ToLower(parts[0]) + ".d"})

		parts := strings.Split(strings.Trim(utils.Delete(cmd, 1, 5), " "), ",")
		record := 0

		filename := parts[0] + ".dat"
		filename = ent.GetWorkDir() + "/" + strings.ToLower(filename)
		//System.Out.Println("will try to load "+filename);
		for i := 1; i < len(parts); i++ {
			tmp := strings.ToUpper(parts[i])
			if tmp[0] == 'R' {
				tmp = utils.Delete(tmp, 1, 1)
				record, _ = utils.SuperStrToInt(tmp)
				record -= 1
			}
		}

		e := files.DOSWRITE(files.GetPath(filename), files.GetFilename(filename), record)
		if e == nil {
			ent.SetOutChannel(filename)
		}
		return e
	}

	if utils.Copy(cmd, 1, 4) == "READ" {
		//		fn := ent.GetWorkDir() + "/" + strings.ToLower(strings.Trim(utils.Delete(cmd, 1, 4), " ")) + ".d"
		parts := strings.Split(strings.Trim(utils.Delete(cmd, 1, 4), " "), ",")
		record := 0

		filename := parts[0] + ".dat"
		filename = ent.GetWorkDir() + "/" + strings.ToLower(filename)
		//System.Out.Println("will try to load "+filename);
		for i := 1; i < len(parts); i++ {
			tmp := strings.ToUpper(parts[i])
			if tmp[0] == 'R' {
				tmp = utils.Delete(tmp, 1, 1)
				record, _ = utils.SuperStrToInt(tmp)
				record -= 1
			}
		}

		e := files.DOSREAD(files.GetPath(filename), files.GetFilename(filename), record)

		if e == nil {
			ent.SetInChannel(filename)
		}

		return e
	}

	if utils.Copy(cmd, 1, 6) == "APPEND" {
		fn := ent.GetWorkDir() + "/" + strings.ToLower(strings.Trim(utils.Delete(cmd, 1, 6), " ")) + ".dat"
		//fn := files.GetUserPath(files.BASEDIR, []string{ent.GetWorkDir(), strings.ToLower(parts[0]) + ".d"})
		e := files.DOSAPPEND(files.GetPath(fn), files.GetFilename(fn))
		if e == nil {
			ent.SetOutChannel(fn)
		}
		return e
	}

	if utils.Copy(cmd, 1, 5) == "CLOSE" {
		fn := ent.GetWorkDir() + "/" + strings.ToLower(strings.Trim(utils.Delete(cmd, 1, 5), " ")) + ".dat"

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
		parts := strings.Split(strings.Trim(utils.Delete(cmd, 1, 4), " "), ",")
		recsize := 0

		filename := parts[0] + ".dat"
		filename = ent.GetWorkDir() + "/" + strings.ToLower(filename)
		//System.Out.Println("will try to load "+filename);
		for i := 1; i < len(parts); i++ {
			tmp := strings.ToUpper(parts[i])
			if tmp[0] == 'L' {
				tmp = utils.Delete(tmp, 1, 1)
				recsize, _ = utils.SuperStrToInt(tmp)
			}
		}
		e := files.DOSOPEN(files.GetPath(filename), files.GetFilename(filename), recsize)
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
				apple2helpers.TEXT40(ent)
				ent.SetMemory(49152, 0)
				break
			}
		case 3:
			{
				apple2helpers.TEXT80(ent)
				ent.SetMemory(49153, 0)
				break
			}
		}

		return nil
	}

	return nil

}

func (this *DialectApplesoft) Init() {

	this.ShortName = "fp"
	this.LongName = "Floating Point BASIC"

	/* vars */
	//	var dn interfaces.DynaCoder
	//var diabc DiaBCODE

	// limit length
	this.MaxVariableLength = 2

	/* Dynacodez */
	//diabc = types.NewDiaBCODE()
	/*title*/

	//this.AddCommand("tracker", NewStandardCommandTRACKER())
	this.AddCommand("boop", NewStandardCommandBOOP())
	this.AddCommand("flush", &StandardCommandFLUSH{})
	//this.AddCommand("input", &StandardCommandINPUT{})
	this.AddCommand("input", NewStandardCommandINPUT())
	this.AddCommand("pinput", NewStandardCommandINPUTM())
	this.AddCommand("catalog", NewStandardCommandCAT())
	this.AddCommand("cat", NewStandardCommandCAT())
	//this.AddCommand("package", NewStandardCommandPACKAGE())
	this.AddCommand("exit", NewStandardCommandEXIT())
	//this.AddCommand("spawn", &StandardCommandSPAWN{})
	this.AddCommand("xlist", NewStandardCommandXLIST())
	this.AddCommand("edit", NewStandardCommandEDIT())
	this.AddCommand("feedback", NewStandardCommandFEEDBACK())
	this.AddCommand("print", &StandardCommandPRINT{})
	this.AddCommand("?", &StandardCommandPRINT{})
	this.AddCommand("declare", &StandardCommandDECLARE{})
	this.AddCommand("home", &StandardCommandCLS{})
	this.AddCommand("mon", NewStandardCommandMON())
	this.AddCommand("list", NewStandardCommandLIST())
	this.AddCommand("run", &StandardCommandRUN{})
	this.AddCommand("new", NewStandardCommandNEW())
	this.AddCommand("goto", &StandardCommandGOTO{})
	this.AddCommand("gosub", &StandardCommandGOSUB{})
	this.AddCommand("return", &StandardCommandRETURN{})
	this.AddCommand("rem", NewStandardCommandREM())
	this.AddCommand("end", &StandardCommandEND{})
	this.AddCommand("stop", &StandardCommandSTOP{})
	this.AddCommand("cont", &StandardCommandCONT{})
	this.AddCommand("else", &StandardCommandNOP{})
	this.AddCommand("then", &StandardCommandNOP{})
	this.AddCommand("trace", &StandardCommandNOP{})
	this.AddCommand("notrace", &StandardCommandNOP{})
	this.AddCommand("store", &StandardCommandNOP{})
	this.AddCommand("recall", &StandardCommandNOP{})
	this.AddCommand("pr#", &StandardCommandPR{})
	this.AddCommand("in#", &StandardCommandQNOP{})
	//this.AddCommand("usr", &StandardCommandQNOP{})
	this.AddCommand("call", &StandardCommandCALL{})
	this.AddCommand("at", &StandardCommandNOP{})
	this.AddCommand("to", &StandardCommandNOP{})
	this.AddCommand("def", &StandardCommandDEF{})
	this.AddCommand("fn", &StandardCommandNOP{})
	this.AddCommand("step", &StandardCommandNOP{})
	this.AddCommand("wait", &StandardCommandWAIT{})
	this.AddCommand("if", &StandardCommandIF{})
	this.AddCommand("save", &StandardCommandSAVE{})
	this.AddCommand("load", &StandardCommandLOAD{})
	this.AddCommand("for", &StandardCommandASFOR{})
	this.AddCommand("next", &StandardCommandASNEXT{})
	this.AddCommand("data", NewStandardCommandDATA())
	this.AddCommand("read", &StandardCommandREAD{})
	this.AddCommand("restore", &StandardCommandRESTORE{})
	this.AddCommand("dim", &StandardCommandDIM{})
	this.AddCommand("pop", &StandardCommandPOP{})
	this.AddCommand("text", &StandardCommandTEXT{})
	this.AddCommand("gr", &StandardCommandGR{})
	this.AddCommand("gr2", &StandardCommandGR2{})
	this.AddCommand("gr3", &StandardCommandGR3{})
	this.AddCommand("gr4", &StandardCommandGR4{})
	this.AddCommand("gr5", &StandardCommandGR5{})
	this.AddCommand("gr6", &StandardCommandGR6{})
	this.AddCommand("gr7", &StandardCommandGR7{})
	//this.AddCommand("gr8", &StandardCommandGR8{})
	this.AddCommand("hgr", &StandardCommandHGR{})
	this.AddCommand("hgr2", &StandardCommandHGR2{})
	this.AddCommand("hgr3", &StandardCommandHGR3{})
	this.AddCommand("hgr4", &StandardCommandHGR4{})
	this.AddCommand("plot", &StandardCommandPLOT{})
	this.AddCommand("hplot", &StandardCommandHPLOT{})
	this.AddCommand("hlin", &StandardCommandHLIN{})
	this.AddCommand("vlin", &StandardCommandVLIN{})
	this.AddCommand("poke", &StandardCommandPOKE{})
	this.AddCommand("htab", &StandardCommandHTAB{})
	this.AddCommand("vtab", &StandardCommandVTAB{})
	this.AddCommand("clear", &StandardCommandCLEAR{})
	this.AddCommand("del", &StandardCommandDEL{})
	this.AddCommand("onerr", &StandardCommandONERR{})
	this.AddCommand("resume", &StandardCommandRESUME{})
	this.AddCommand("on", &StandardCommandON{})
	this.AddCommand("xdraw", &StandardCommandXDRAW{})
	this.AddCommand("draw", &StandardCommandDRAW{})
	this.AddCommand("himem:", &StandardCommandHIMEM{})
	this.AddCommand("lomem:", &StandardCommandLOMEM{})
	this.AddCommand("&", &StandardCommandAMPCALL{})
	this.AddCommand("help", &StandardCommandHELP{})
	this.AddCommand("lang", NewStandardCommandDIALECT())
	this.AddCommand("let", &StandardCommandIMPLIEDASSIGN{})
	this.AddCommand("trace", &StandardCommandTRACE{})
	this.AddCommand("notrace", &StandardCommandNOTRACE{})
	//this.AddCommand( "dsp", &NewStandardCommandDSP());
	this.AddCommand("nodsp", &StandardCommandNODSP{})
	this.AddCommand("get", &StandardCommandGET{})
	this.AddCommand("renumber", &StandardCommandRENUMBER{})
	this.AddCommand("reorganize", &StandardCommandREORGANIZE{})

	this.AddCommand("autosave", &StandardCommandAUTOSAVE{})

	this.AddCommand("hgr5", &StandardCommandXGR{})
	this.AddCommand("hgr6", &StandardCommandXGR2{})

	/* dummies for now */
	this.AddCommand("inverse", &StandardCommandVIDEOINVERSE{})
	this.AddCommand("normal", &StandardCommandVIDEONORMAL{})
	this.AddCommand("flash", &StandardCommandVIDEOFLASH{})

	this.AddCommand("hcolor=", &StandardCommandHCOLOR{})
	this.AddCommand("color=", &StandardCommandCOLOR{})
	this.AddCommand("speed=", &StandardCommandSPEED{})
	this.AddCommand("scale=", &StandardCommandSCALE{})
	this.AddCommand("rot=", &StandardCommandROT{})

	this.ImpliedAssign = &StandardCommandIMPLIEDASSIGN{}

	/* math functions - TRS-80 LEVEL II */
	this.AddFunction("abs(", NewStandardFunctionABS(0, 0, *types.NewTokenList()))
	this.AddFunction("fix(", NewStandardFunctionINT(0, 0, *types.NewTokenList()))
	this.AddFunction("rnd(", NewStandardFunctionASRND(0, 0, *types.NewTokenList()))
	this.AddFunction("sqr(", NewStandardFunctionSQR(0, 0, *types.NewTokenList()))
	this.AddFunction("log(", NewStandardFunctionLOG(0, 0, *types.NewTokenList()))
	this.AddFunction("exp(", NewStandardFunctionEXP(0, 0, *types.NewTokenList()))
	this.AddFunction("int(", NewStandardFunctionINT(0, 0, *types.NewTokenList()))
	//this.AddFunction( "cint", NewStandardFunctionINT(0,0,nil) );
	//this.AddFunction( "csng", NewStandardFunctionFNOP(0,0,nil) );
	//this.AddFunction( "cdbl", NewStandardFunctionFNOP(0,0,nil) );
	this.AddFunction("sin(", NewStandardFunctionSIN(0, 0, *types.NewTokenList()))
	this.AddFunction("cos(", NewStandardFunctionCOS(0, 0, *types.NewTokenList()))
	this.AddFunction("tan(", NewStandardFunctionTAN(0, 0, *types.NewTokenList()))
	this.AddFunction("atn(", NewStandardFunctionATAN(0, 0, *types.NewTokenList()))
	this.AddFunction("sgn(", NewStandardFunctionSGN(0, 0, *types.NewTokenList()))

	this.AddFunction("asc(", NewStandardFunctionASC(0, 0, *types.NewTokenList()))
	this.AddFunction("chr$(", NewStandardFunctionCHRDollar(0, 0, *types.NewTokenList()))
	this.AddFunction("left$(", NewStandardFunctionLEFTDollar(0, 0, *types.NewTokenList()))
	this.AddFunction("right$(", NewStandardFunctionRIGHTDollar(0, 0, *types.NewTokenList()))
	this.AddFunction("mid$(", NewStandardFunctionMIDDollar(0, 0, *types.NewTokenList()))
	this.AddFunction("len(", NewStandardFunctionLEN(0, 0, *types.NewTokenList()))
	this.AddFunction("str$(", NewStandardFunctionSTRDollar(0, 0, *types.NewTokenList()))
	// this.AddFunction( "string$", NewStandardFunctionSTRINGDollar(0,0,nil) );
	this.AddFunction("val(", NewStandardFunctionVAL(0, 0, *types.NewTokenList()))
	//this.AddFunction( "inkey$", NewStandardFunctionINKEYDollar(0,0,nil) );
	this.AddFunction("tab(", NewStandardFunctionTAB(0, 0, *types.NewTokenList()))
	this.AddFunction("spc(", NewStandardFunctionSPC(0, 0, *types.NewTokenList()))
	this.AddFunction("pos(", NewStandardFunctionPOS(0, 0, *types.NewTokenList()))
	this.AddFunction("peek(", NewStandardFunctionPEEK())
	this.AddFunction("scrn(", NewStandardFunctionSCRN(0, 0, *types.NewTokenList()))
	this.AddFunction("pdl(", NewStandardFunctionPDL(0, 0, *types.NewTokenList()))
	this.AddFunction("fre(", NewStandardFunctionFRE(0, 0, *types.NewTokenList()))
	this.AddFunction("usr(", NewStandardFunctionUSR(0, 0, *types.NewTokenList()))

	plus.RegisterFunctions(this)
	// list := plus.GetList(this)

	this.Logicals["or"] = 1
	this.Logicals["and"] = 1
	this.Logicals["not"] = 1

	this.VarSuffixes = "%$&#"

	/* dynacode test shim */
	this.ReverseCase = true

	this.ArrayDimDefault = 10
	this.ArrayDimMax = 65535
	this.Title = "Applesoft"

	this.IPS = -1

}

func (this *DialectApplesoft) ExecuteDirectCommand(tl types.TokenList, ent interfaces.Interpretable, Scope *types.Algorithm, LPC *types.CodeRef) error {

	/* vars */
	var tok *types.Token
	var n string
	var cmd interfaces.Commander
	var cr *types.CodeRef
	var ss *types.TokenList

	// update random numbers...
	ent.SetMemory(78, uint64(utils.Random()*256))
	ent.SetMemory(79, uint64(utils.Random()*256))

	if this.NetBracketCount(tl) != 0 {
		return exception.NewESyntaxError("SYNTAX ERROR")
	}

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

	//log.Printf("DO %s", ent.TokenListAsString(tl))

	if this.Trace && (ent.GetState() == types.RUNNING) {
		ent.PutStr("#" + utils.IntToStr(LPC.Line) + " ")
	}

	/* process poop monster here (@^-^@) */
	tok = tl.Shift()

	if (tok.Type == types.NUMBER) || (tok.Type == types.INTEGER) {
	} else if tok.Type == types.DYNAMICKEYWORD {
		n = strings.ToLower(tok.Content)
		if dcmd, ok := this.DynaCommands[n]; ok {
			//dcmd = this.DynaCommands.Get(n)

			//ent.PutStr("Dynamic command parsing - Start at "+IntToStr(dcmd.Code.LowIndex)+"\r\n");

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
			this.HandleException(ent, exception.NewESyntaxError("SYNTAX ERROR"))
		}

	} else if tok.Type == types.KEYWORD {

		if (tl.Size() > 0) && (tl.LPeek().Type == types.ASSIGNMENT) {
			return exception.NewESyntaxError("SYNTAX ERROR")
		}

		n = strings.ToLower(tok.Content)
		if cmd, ok := this.Commands[n]; ok {
			//cmd = this.Commands.Get(n

			if !cmd.IsStateBased() {

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
				// Setup state based command
				cs := interfaces.NewCommandState(cmd)
				cs.Params = tl
				cs.Scope = Scope
				cs.PC = *LPC
				ent.SetSubState(types.ESS_INIT)
				ent.SetCommandState(cs)
				return nil
			}

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

	} else {
		return exception.NewESyntaxError("SYNTAX ERROR")
	}

	///tl.Free; /* clean up */
	return nil
}

func (this *DialectApplesoft) PutStr(ent interfaces.Interpretable, s string) {
	apple2helpers.PutStr(ent, s)
}

func (this *DialectApplesoft) RealPut(ent interfaces.Interpretable, ch rune) {
	apple2helpers.Put(ent, ch)
}

func (this *DialectApplesoft) Backspace(ent interfaces.Interpretable) {
	apple2helpers.Backspace(ent)
}

func (this *DialectApplesoft) ClearToBottom(ent interfaces.Interpretable) {
	apple2helpers.ClearToBottom(ent)
}

func (this *DialectApplesoft) SetCursorX(ent interfaces.Interpretable, xx int) {
	x := (80 / apple2helpers.GetFullColumns(ent)) * xx

	apple2helpers.SetCursorX(ent, x)
}

func (this *DialectApplesoft) SetCursorY(ent interfaces.Interpretable, yy int) {
	y := (48 / apple2helpers.GetFullRows(ent)) * yy

	apple2helpers.SetCursorY(ent, y)
}

func (this *DialectApplesoft) GetColumns(ent interfaces.Interpretable) int {
	return apple2helpers.GetColumns(ent)
}

func (this *DialectApplesoft) GetRows(ent interfaces.Interpretable) int {
	return apple2helpers.GetRows(ent)
}

func (this *DialectApplesoft) Repos(ent interfaces.Interpretable) {
	apple2helpers.Gotoxy(ent, int(ent.GetMemory(36)), int(ent.GetMemory(37)))
}

func (this *DialectApplesoft) GetCursorX(ent interfaces.Interpretable) int {
	return apple2helpers.GetCursorX(ent) / (80 / apple2helpers.GetFullColumns(ent))
}

func (this *DialectApplesoft) GetCursorY(ent interfaces.Interpretable) int {
	return apple2helpers.GetCursorY(ent) / (48 / apple2helpers.GetFullRows(ent))
}

func (this *DialectApplesoft) GetProgramStart(ent interfaces.Interpretable) int {
	v := int(ent.GetMemory(103)&0xff) + 256*int(ent.GetMemory(104)&0xff)
	if v == 0 || v > 49152 {
		v = 2049
	}
	return v
}

func (this *DialectApplesoft) InitVarmap(ent interfaces.Interpretable, vm types.VarManager) {

	MEMBASE := this.GetProgramStart(ent)

	//fmt.Println("MEMBASE =", MEMBASE)

	fretop := MEMTOP

	ent.SetMemory(105, uint64(ent.GetMemory(MEMBASE)%256))
	ent.SetMemory(106, uint64(ent.GetMemory(MEMBASE)/256))
	ent.SetMemory(107, uint64(ent.GetMemory(MEMBASE)%256))
	ent.SetMemory(108, uint64(ent.GetMemory(MEMBASE)/256))
	varmem := 256*int(ent.GetMemory(106)&0xff) + int(ent.GetMemory(105)&0xff)

	//fmt.Println("varmem =", varmem)

	// Create an Applesoft compatible memory map
	vmgr := types.NewVarManagerMSBIN(
		ent.GetMemoryMap(),
		ent.GetMemIndex(),
		105,
		107,
		111,
		109,
		115,
		types.VUR_QUIET,
	)

	vmgr.SetVector(vmgr.VARTAB, varmem)
	vmgr.SetVector(vmgr.ARRTAB, varmem)
	vmgr.SetVector(vmgr.STREND, varmem+1)
	vmgr.SetVector(vmgr.FRETOP, fretop)
	vmgr.SetVector(vmgr.MEMSIZ, fretop)

	// set the start of the table to zeroes to prevent spurious variable recognition`
	ent.SetMemory(varmem, 0)
	ent.SetMemory(varmem+1, 0)

	ent.SetVM(vmgr)

}

// NTokenize tokenize a group of tokens to uints
func (this *DialectApplesoft) NTokenize(tl types.TokenList) []uint64 {

	var values []uint64

	//var lasttok *types.Token

	for _, t := range tl.Content {
		switch t.Type {
		case types.LOGIC:
			//values = append(values, this.TokenMapping[strings.ToLower(t.Content)])
			values = append(values, TokenToCode[strings.ToUpper(t.Content)])
		case types.KEYWORD:
			//values = append(values, this.TokenMapping[strings.ToLower(t.Content)])
			values = append(values, TokenToCode[strings.ToUpper(t.Content)])
		case types.FUNCTION:
			//values = append(values, this.TokenMapping[strings.ToLower(t.Content)])
			values = append(values, TokenToCode[strings.ToUpper(t.Content)])
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

func (this *DialectApplesoft) GetMemoryRepresentation(a *types.Algorithm) []uint64 {
	data := make([]uint64, 0)
	return data

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

func (this *DialectApplesoft) ParseMemoryRepresentation(data []uint64) types.Algorithm {

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

func (this *DialectApplesoft) UnNTokenize(values []uint64) *types.TokenList {

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
			s = s + " " + CodeToToken[v]
			if CodeToToken[v] == "REM" || CodeToToken[v] == "DATA" {
				skipspace = true
			}
		}
		lastcode = v
	}

	tl := this.Tokenize(runestring.Cast(s))

	return tl
}

func (this *DialectApplesoft) UpdateRuntimeState(ent interfaces.Interpretable) {

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

func (this *DialectApplesoft) DumpState(ent interfaces.Interpretable) {

	////fmt.Printf("- Entity State: %v\n", ent.GetState())
	////fmt.Printf("- Entity PC   : line %d, stmt %d\n", ent.GetPC().Line, ent.GetPC().Statement)
	////fmt.Printf("- Error trap  : line %d, stmt %d\n", ent.GetErrorTrap().Line, ent.GetErrorTrap().Statement)
	////fmt.Printf("- Data pointer: line %d, stmt %d, token %d, subindex %d\n", ent.GetDataRef().Line, ent.GetDataRef().Statement, ent.GetDataRef().Token, ent.GetDataRef().SubIndex)
	////fmt.Printf("- Loopstack   : %d entries, %f, %s\n", ent.GetLoopStack().Size(), ent.GetLoopStep(), ent.GetLoopVariable())
	////fmt.Printf("- Callstack   : %d entries\n", ent.GetStack().Size())
	////fmt.Printf("- Variables   : %d, %v\n", len(ent.GetLocal().Keys()), ent.GetLocal().Keys())
	////fmt.Printf("- Program size: %d bytes\n", ent.GetMemory(2048)-2049)

}

func (this *DialectApplesoft) PreFreeze(ent interfaces.Interpretable) {

	this.UpdateRuntimeState(ent)
	this.DumpState(ent)

}

func (this *DialectApplesoft) PostThaw(ent interfaces.Interpretable) {

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

	fixMemoryPtrs(ent)

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

func (this *DialectApplesoft) ThawVideoConfig(ent interfaces.Interpretable) {
	apple2helpers.RestoreSoftSwitches(ent)
}

func (this *DialectApplesoft) HomeLeft(ent interfaces.Interpretable) {
	apple2helpers.HomeLeft(ent)
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

// Build completion list...
// Priority given to language keywords
func (this *DialectApplesoft) GetCompletions(ent interfaces.Interpretable, line runestring.RuneString, index int) (int, *types.TokenList) {
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

	nca := []string{"load", "save", "run", "next", "for", "help", "edit"}
	efn := []string{"load", "save", "run", "edit"}

	searchbase := last

	ctx := CTX_START_LINE

	if tl.Size() > 0 {
		t := tl.RPeek()
		f := tl.LPeek()

		if typeIn(t.Type, []types.TokenType{types.KEYWORD}) {
			ctx = CTX_AFTER_KW

		} else if typeIn(t.Type, []types.TokenType{types.FUNCTION, types.PLUSFUNCTION}) {
			ctx = CTX_AFTER_FUNC
		} else if typeIn(t.Type, []types.TokenType{types.ASSIGNMENT, types.COMPARITOR}) {
			ctx = CTX_AFTER_ASSIGN
		} else if typeIn(t.Type, []types.TokenType{types.SEPARATOR}) {
			ctx = CTX_START_LINE
		}

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
