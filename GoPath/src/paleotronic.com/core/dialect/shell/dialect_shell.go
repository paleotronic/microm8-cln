package shell

import (
	"errors"
	"math"
	"regexp"
	"strings"
	"time"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/dialect/applesoft"
	"paleotronic.com/core/dialect/plus"
	"paleotronic.com/core/exception"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/files"
	"paleotronic.com/fmt" //"paleotronic.com/fmt"
	"paleotronic.com/log" //	"time"
	"paleotronic.com/runestring"
	"paleotronic.com/utils"
)

type DialectShell struct {
	dialect.Dialect
}

func NewDialectShell() *DialectShell {
	this := &DialectShell{}
	this.Dialect = *dialect.NewDialect()
	this.Init()
	this.Dialect.DefaultCost = 1000000000 / 1600
	this.Throttle = 100.0
	this.GenerateNumericTokens()
	return this
}

func (this *DialectShell) IsBreakingCharacter(ch rune, vs string) bool {

	items := " \r\n\t+-*^([]){}=:,;<>?"

	return (utils.Pos(string(ch), items) > 0)
}

func (this *DialectShell) Evaluate(chunk string, tokens *types.TokenList) bool {

	/* vars */
	var result bool
	var tok types.Token
	var ptok *types.Token

	result = true // continue

	//System.Out.Println("EVALUATE: ", chunk);
	//fmt.Printf("Evaluate token: (%s)\n", chunk)

	if len(chunk) == 0 {
		return result
	}

	tok = *types.NewToken(types.INVALID, "")

	//  if (this.IsBoolean(chunk))
	//  begin
	//    tok.Type = types.BOOLEAN;
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

		z := this.Commands[strings.ToLower(chunk)]

		result = !(z.HasNoTokens())
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
	} else if this.IsPlusVariableName(chunk) {
		tok.Type = types.PLUSVAR
		tok.Content = chunk
	} else if this.IsString(chunk) {
		tok.Type = types.STRING

		tok.Content = utils.Copy(chunk, 2, len(chunk)-2)
	} else {
		tok.Type = types.STRING
		tok.Content = chunk
	}

	if tok.Type != types.INVALID {
		//System.Out.Println("ADD: ", tok.Content);
		log.Printf("Yielded token: %q\n", tok)
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

func (this *DialectShell) Tokenize(s runestring.RuneString) *types.TokenList {

	/* vars */
	var result *types.TokenList
	var inq bool
	var inqq bool
	var cont bool
	var idx int
	var chunk string
	//pchunk := ""
	var ch rune
	var tt types.Token

	//	////fmt.Println("DialectShell.Tokenize()")

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
				inqq = false
			} else {
				chunk = chunk + string(ch)
			}
			inqq = !inqq
		} else {
			chunk = chunk + string(ch)

			/* break keywords out early */
			if this.Commands.ContainsKey(strings.ToLower(chunk)) {
				if (strings.ToLower(chunk) != "go") && (strings.ToLower(chunk) != "to") && (strings.ToLower(chunk) != "on") && (strings.ToLower(chunk) != "hgr") && (strings.ToLower(chunk) != "at") {
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

			tt = *types.NewToken(types.UNSTRING, strings.Trim(chunk, " "))
			//pchunk = chunk
			chunk = ""
			result.Push(&tt)
		}

	} /*while*/

	//System.Out.Println("chunk == ", chunk;

	if len(chunk) > 0 {
		//if inqq {
		//	chunk = chunk + "\""
		//}
		this.Evaluate(chunk, result)
		//pchunk = chunk
		chunk = ""
	}

	/* enforce non void return */
	return result

}

func (this *DialectShell) DecideValueIndex(hpop int, count int, values types.TokenList) int {
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

func (this *DialectShell) HPOpIndex(tl types.TokenList) int {
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

func (this *DialectShell) ParseTokensForResult(ent interfaces.Interpretable, tokens types.TokenList) (*types.Token, error) {

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
	var lastop bool
	var hpop int
	//	var hs int
	//	var sc int

	result = types.NewToken(types.INVALID, "")

	/* must be 1 || more tokens in list */
	if tokens.Size() == 0 {
		return result, nil
	}

	log.Printf("Applesoft version of ParseTokensForResult() invoked for [%s]\n", ent.TokenListAsString(tokens))

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
	/*main parse loop*/
	for tidx < tokens.Size() {
		tok = tokens.Get(tidx)

		////fmt.Printf("tidx = %d, Type = %d, Content = %s\n", tidx, tok.Type, tok.Content)

		//System.Err.Println( "--------------> type of token at tidx "+tidx+" is "+tok.Type+"["+tok.Content+"]" );

		if (tok.Type == types.NUMBER) || (tok.Type == types.INTEGER) || (tok.Type == types.STRING) || (tok.Type == types.BOOLEAN) {
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
			} else if (lastop == true) && (lasttok != nil) && ((lasttok.Type == types.ASSIGNMENT) || (lasttok.Type == types.COMPARITOR)) && (tok.Content == "+") {
				tidx++
				continue
			}

			ops.Push(tok)
			lastop = true
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
						tok, _ = this.ParseTokensForResult(ent, exptok1)
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
							tok, _ = this.ParseTokensForResult(ent, exptok1)
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
						tok, _ = this.ParseTokensForResult(ent, exptok1)
						par.Push(tok)
						log.Println("Seeding to function params: ", tok)
					}
					//FreeAndNil(exptok);
				}
			} else {
				par = *types.NewTokenList()
			}

			fun.SetEntity(ent)
			fun.FunctionExecute(&par)
			values.Push(fun.GetStack().Pop())
			lastop = false

		} else if tok.Type == types.KEYWORD {

			values.Push(types.NewToken(types.STRING, tok.Content))
			lastop = false

		} else if tok.Type == types.VARIABLE {

			// fix for missing + || separators;
			if (lastop == false) && (lasttok != nil) {
				ops.Push(types.NewToken(types.OPERATOR, "+"))
			}

			/* first try entity local */
			n = strings.ToLower(tok.Content)
			v = nil
			if ent.GetLocal() != nil && ent.GetLocal().ContainsKey(n) {
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
				//caller.GetVDU().PutStr("after create");
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
					//ent.GetVDU().PutStr("*** VAR INDICES ARE ["+this.TokenListAsString(subexpr)+']'+PasUtil.CRLF );

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
						ntok, _ = this.ParseTokensForResult(ent, exptok)
					}
				}
				//this.VDU.PutStr("// Adding value from array "+ntok.Content);
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
						ntok, _ = this.ParseTokensForResult(ent, exptok)
					}
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
			ntok, _ = this.ParseTokensForResult(ent, subexpr)
			//FreeAndNil(subexpr);
			values.Push(ntok)
			lastop = false
		}

		lasttok = tok
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
		//ent.GetVDU().PutStr(">> Ops: "+ent.TokenListAsString(ops)+PasUtil.CRLF);
		//ent.GetVDU().PutStr(">> Val: "+ent.TokenListAsString(values)+PasUtil.CRLF);

		hpop = this.HPOpIndex(ops)
		op = *ops.Remove(hpop)
		/*        if ((hpop != 0) && (hpop != ops.Size()-1)) {
		              // rearrange op to be left most;
		              op = ops.Get(hpop)
		              if ((op.Type != types.LOGIC) && (values.Size() >= 2)) {
		                op = ops.Remove(hpop)
		                b = values.Remove(hpop)
		                a = values.Remove(hpop)
		                ops.UnShift(op)
		                values.UnShift(b)
		                values.UnShift(a)
		              }
		            }
		    var / *

		    /*
		            if (ops.LPeek().Content == "^") {
		                op = ops.Left
		                left = true
		            } else if (ops.RPeek().Content == "^") {
		                op = ops.Right
		                left = false
		            } else if ((ops.LPeek().Content == "*") || (ops.LPeek().Content == "/")) {
		                op = ops.Left
		                left = true
		            } else if ((ops.RPeek().Content == "*") || (ops.RPeek().Content == "/")) {
		                op = ops.Right
		                left = false
		            } else if ((ops.LPeek().Content == "+") || (ops.LPeek().Content == "-")) {
		                op = ops.Left
		                left = true
		            } else if ((ops.RPeek().Content == "+") || (ops.RPeek().Content == "-")) {
		                op = ops.Right
		                left = false
		            } else if ((ops.LPeek().Type == types.COMPARITOR) || (ops.LPeek().Type == types.ASSIGNMENT)) {
		                op = ops.Left
		                left = true
		            } else if ((ops.RPeek().Type == types.COMPARITOR) || (ops.RPeek().Type == types.ASSIGNMENT)) {
		                op = ops.Right
		                left = false
		            } else {
		                op = ops.Right
		                left = false
		            }
		    var / *

		            //ent.GetVDU().PutStr("Next Op is: "+op.Content+PasUtil.CRLF);

		            /*if (op.Type != types.LOGIC) {
		              if (left) {
		                this.VDU.PutStr("DEBUG: Evaluate "+values.Get(0].Content+op.Content+values[1).Content+"
		")
		              } else {
		                this.VDU.PutStr("DEBUG: Evaluate "+values.Get(values.Size()-2].Content+op.Content+values[values.Size()-1).Content+"
		")
		              }
		            } */
		//Writeln("Processing operator: ",op.Content);
		if op.Type == types.LOGIC {

			if strings.ToLower(op.Content) == "not" {

				//a = values.Pop();
				hpop = this.DecideValueIndex(hpop, 1, values)
				a = *values.Remove(hpop)

				//ent.GetVDU().PutStr(op.Content+" "+a.Content+PasUtil.CRLF);

				/*if (a.AsInteger() != 0)
				      values.Push( types.NewToken(types.NUMBER, "0") )
				  else {
				      values.Push( types.NewToken(types.NUMBER, "1") );*/
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

				//ent.GetVDU().PutStr(a.Content+" "+op.Content+" "+b.Content+PasUtil.CRLF);

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

				//ent.GetVDU().PutStr(a.Content+" "+op.Content+" "+b.Content+PasUtil.CRLF);

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

				//ent.GetVDU().PutStr(a.Content+" "+op.Content+" "+b.Content+PasUtil.CRLF);

				//values.Push( types.NewToken(types.NUMBER, IntToStr(a.AsInteger() xor b.AsInteger())) );
				ntok = types.NewToken(types.NUMBER, utils.IntToStr(a.AsInteger()^b.AsInteger()))
				values.Insert(hpop, ntok)
			}

		} else if (op.Type == types.COMPARITOR) || (op.Type == types.ASSIGNMENT) {

			if values.Size() < 2 {
				return result, nil
			}

			//writeln("@@@@@@@@@@@@@@@@@@@@@ COMPARE");

			/*if (left) {
			      a = values.Left
			      b = values.Left
			  } else {
			      b = values.Right
			      a = values.Right
			  }*/
			//if (hpop == values.Size()-1)
			//	hpop--;
			hpop = this.DecideValueIndex(hpop, 2, values)
			a = *values.Remove(hpop)
			b = *values.Remove(hpop)

			//ent.GetVDU().PutStr(a.Content+" "+op.Content+" "+b.Content+PasUtil.CRLF);

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

				/*if (left)
				     values.UnShift(ntok)
				  else {
				      values.Push(ntok);*/
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
				/*if (left)
				     values.Unshift(ntok)
				  else {
				      values.Push(ntok);*/
				//}
				values.Insert(hpop, ntok)

			}

		} else if op.Type == types.OPERATOR {

			//writeln("Currently ",values.Size()," values in stack... About to pop 2");

			if values.Size() < 2 {
				return result, nil
			}

			/*if (left) {
			      a = values.Left
			      b = values.Left;
			  } else {
			      b = values.Right
			      a = values.Right
			  }*/
			//if (hpop == values.Size()-1)
			//	hpop--;
			hpop = this.DecideValueIndex(hpop, 2, values)
			a = *values.Remove(hpop)
			b = *values.Remove(hpop)

			//ent.GetVDU().PutStr(a.Content+" "+op.Content+" "+b.Content+PasUtil.CRLF);

			//  ent.GetVDU().PutStr("Op is: "+a.Content+op.Content+b.Content+PasUtil.CRLF);

			if a.IsNumeric() && b.IsNumeric() {

				aa = float64(a.AsNumeric())
				bb = float64(b.AsNumeric())
				rr = 0

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

				n = utils.FormatFloat("", rr)
				ntok = types.NewToken(types.NUMBER, n)
				/*
				   if (trunc(rr) == rr) {
				       Str(trunc(rr), n)
				       ntok = types.NewToken( types.INTEGER, n )
				   } else {
				       ntok = types.NewToken( types.NUMBER, n )
				   }
				   var / *
				   /*if (left)
				      values.UnShift(ntok)
				   else {
				       values.Push(ntok);*/
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
				/*if (left)
				     values.UnShift(ntok)
				  else {
				      values.Push(ntok);*/
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
				/*if (left)
				     values.UnShift(ntok)
				  else {
				      values.Push(ntok);*/
				//}
				values.Insert(hpop, ntok)
			}

		}

	}

	if err || (values.Size() > 1) {
		return result, nil
	}

	result = values.Pop()
	result = types.NewToken(result.Type, ""+result.Content)

	if result.Type == types.NUMBER {
		// clip places
		//result.Content = types.NewDecimalFormat("#.########").Format(result.AsExtended())
		result.Content = utils.StrToFloatStr(result.Content)
	}

	//System.Err.Println("return "+result.Content);

	// removed free call here;
	// removed free call here;

	/* enforce non void return */
	return result, nil

}

func (this *DialectShell) HandleException(ent interfaces.Interpretable, e error) {

	/* vars */
	var msg string

	apple2helpers.NLIN(ent)

	msg = e.Error()

	//CoreException.ErrorRecord ed = ex.StringToErrorDetail(msg)
	//msg = ed.Msg;

	//e.PrintStackTrace()
	//System.Out.Println("MOO");

	//ent.Memory[222] = (ed.Code & 0xff)

	if (ent.GetState() == types.RUNNING) || (ent.GetState() == types.DIRECTRUNNING) || (ent.GetState() == types.STOPPED) {
		//System.Err.Println("PIG");
		if !ent.HandleError() {
			ent.PutStr("error: " + msg)
			if (ent.GetState() == types.RUNNING) && (ent.GetPC().Line != 0) {
				ent.PutStr(" AT LINE " + utils.IntToStr(ent.GetPC().Line))
			}
			//try {
			ent.Halt()
			//} catch (Exception ee) {
			//}
		}
	}

	ent.PutStr("\r\n")
	apple2helpers.Beep(ent)
	//ent.GetVDU().DoPrompt();
	//System.Out.Println("COW");

}

func (this *DialectShell) Parse(ent interfaces.Interpretable, s string) error {

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
				z := ent.GetCode()
				z.Remove(lno)
				//ent.GetVDU().PutStr("-"+PasUtil.IntToStr(lno)+"\r\n");
			}
		}
		//types.RebuildAlgo()
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
	//ent.GetVDU().PutStr(fmt.Sprintf("Direct has %d lines\r\n", len(a)))

	//str(ent.GetState(), n);
	//fInterpreter.VDU.PutStr(n+": run direct with "+IntToStr(ll.Size())+" commands: "+s+"\r\n");

	// start running
	ent.SetState(types.DIRECTRUNNING)

	//ent.GetVDU().PutStr(fmt.Sprintf("Interpreter state is now %v\r\n", ent.GetState()))

	return nil

}

func (this *DialectShell) InitVDU(v interfaces.Interpretable, promptonly bool) {

	/* vars */

	v.SetPrompt("$")
	v.SetTabWidth(8)

	if !promptonly {
		apple2helpers.Clearscreen(v)
		apple2helpers.Gotoxy(v, 0, 0)
		apple2helpers.PutStr(v, "\r\n")
		apple2helpers.PutStr(v, "Welcome to microShell!\r\n")
		apple2helpers.PutStr(v, "\r\n")
		apple2helpers.PutStr(v, "Type 'fp' for floating point basic,\r\n")
		apple2helpers.PutStr(v, "'int' for integer basic, or type 'logo'\r\n")
		apple2helpers.PutStr(v, "for 3D Logo!\r\n")
		apple2helpers.PutStr(v, "\r\n")
		v.SetNeedsPrompt(true)
	}

	settings.SpecName[v.GetMemIndex()] = "microShell"
	settings.SetSubtitle(settings.SpecName[v.GetMemIndex()])

	v.SaveCPOS()

	v.SetNeedsPrompt(true)
}

func (this *DialectShell) InitVarmap(ent interfaces.Interpretable, vm types.VarManager) {

	MEMBASE := 2048
	MEMTOP := 32768

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

func (this *DialectShell) ProcessDynamicCommand(ent interfaces.Interpretable, cmd string) error {

	//	////fmt.Printf("In ProcessDynamicCommand for [%s]\n", cmd)

	if utils.Copy(cmd, 1, 5) == "BLOAD" {

		parts := strings.Split(strings.Trim(utils.Delete(cmd, 1, 5), " "), ",")
		addr := 16384
		length := -1

		filename := parts[0]
		filename = strings.ToLower(filename)
		//System.Out.Println("will try to load "+filename);
		for i := 1; i < len(parts); i++ {
			tmp := parts[i]
			if tmp[0] == 'A' {
				tmp = utils.Delete(tmp, 1, 1)
				addr = utils.StrToInt(tmp)
			} else if tmp[0] == 'L' {
				tmp = utils.Delete(tmp, 1, 1)
				length = utils.StrToInt(tmp)
			}
		}

		if !files.ExistsViaProvider(ent.GetWorkDir(), filename+".s") {
			return exception.NewESyntaxError("FILE NOT FOUND")
		}

		data, err := files.ReadBytesViaProvider(ent.GetWorkDir(), filename+".s")
		if err != nil {
			return err
		}

		rl := len(data.Content)
		if (length != -1) && (length <= rl) {
			rl = length
		}
		//System.Out.Println("BLOAD A="+addr+", L="+rl);
		for i := 0; i < rl; i++ {
			ent.SetMemory((addr+i)%65536, uint64(data.Content[i]&0xff))
		}

		//System.Out.Println("OK");

		return nil

	}

	// Run shim...

	if utils.Copy(cmd, 1, 3) == "RUN" {
		return nil
	}

	if utils.Copy(cmd, 1, 5) == "WRITE" {
		parts := strings.Split(strings.Trim(utils.Delete(cmd, 1, 5), " "), ",")
		fn := files.GetUserPath(files.BASEDIR, []string{ent.GetWorkDir(), strings.ToLower(parts[0]) + ".d"})
		ent.SetOutChannel(fn)
		files.WriteString(fn, "", false)
		return nil
	}

	if utils.Copy(cmd, 1, 4) == "READ" {
		parts := strings.Split(strings.Trim(utils.Delete(cmd, 1, 4), " "), ",")
		fn := files.GetUserPath(files.BASEDIR, []string{ent.GetWorkDir(), strings.ToLower(parts[0]) + ".d"})

		readFrom := 1

		for i := 1; i < len(parts); i++ {
			tmp := parts[i]
			if tmp[0] == 'R' {
				tmp = utils.Delete(tmp, 1, 1)
				readFrom = utils.StrToInt(tmp)
			}
		}

		if !files.Exists(fn) {
			return exception.NewESyntaxError("FILE NOT FOUND")
		}

		text, err := utils.ReadTextFile(fn)
		if err != nil {
			return exception.NewESyntaxError("I/O ERROR")
		}

		var t string
		var zzz int
		for _, l := range text {
			if zzz >= readFrom {
				if t != "" {
					t = t + "\r\n"
				}
				t = t + l
			}
			zzz++
		}
		ent.SetFeedBuffer(t)

		return nil
	}

	if utils.Copy(cmd, 1, 6) == "APPEND" {
		parts := strings.Split(strings.Trim(utils.Delete(cmd, 1, 6), " "), ",")
		fn := files.GetUserPath(files.BASEDIR, []string{ent.GetWorkDir(), strings.ToLower(parts[0]) + ".d"})
		ent.SetOutChannel(fn)
		files.WriteString(fn, "", false)
		return nil
	}

	if utils.Copy(cmd, 1, 5) == "CLOSE" {
		//		parts := strings.Split(strings.Trim(utils.Delete(cmd, 1, 5), " "), ",")
		ent.SetOutChannel("")
		ent.SetFeedBuffer("")
		return nil
	}

	if utils.Copy(cmd, 1, 4) == "OPEN" {
		//		parts := strings.Split(strings.Trim(utils.Delete(cmd, 1, 4), " "), ",")
		ent.SetOutChannel("")
		ent.SetFeedBuffer("")
		return nil
	}

	if utils.Copy(cmd, 1, 3) == "PR#" {
		cmd = strings.Trim(utils.Delete(cmd, 1, 3), " ")
		//		mode := utils.StrToInt(cmd)

		//		switch mode { /* FIXME - Switch statement needs cleanup */
		//		case 0:
		//			{
		//				ent.GetVDU().SetVideoMode(ent.GetVDU().GetVideoModes()[5])
		//				ent.GetVDU().ClrHome()
		////				ent.GetVDU().RegenerateWindow(ent.GetMemory())
		//				break
		//			}
		//		case 3:
		//			{
		//				ent.GetVDU().SetVideoMode(ent.GetVDU().GetVideoModes()[0])
		//				ent.GetVDU().ClrHome()
		//				//ent.GetVDU().RegenerateWindow(ent.GetMemory())
		//				break
		//			}
		//		}

		return nil
	}

	return nil

}

func (this *DialectShell) Init() {

	this.ShortName = "shell"
	this.LongName = "Shell"

	// limit length
	this.MaxVariableLength = 32

	this.ImpliedAssign = nil

	this.AddCommand("cat", &CommandWrapper{WrappedName: "cat", WrappedCommand: &applesoft.StandardCommandCAT{}})
	this.AddCommand("display", &CommandWrapper{WrappedName: "display", WrappedCommand: &applesoft.StandardCommandPRINT{}})
	this.AddCommand("?", &CommandWrapper{WrappedName: "?", WrappedCommand: &applesoft.StandardCommandPRINT{}})
	this.AddCommand("return", &CommandWrapper{WrappedName: "return", WrappedCommand: &applesoft.StandardCommandRETURN{}})
	this.AddCommand("exit", &CommandWrapper{WrappedName: "exit", WrappedCommand: &applesoft.StandardCommandEXIT{}})
	this.AddCommand("halt", &CommandWrapper{WrappedName: "halt", WrappedCommand: &applesoft.StandardCommandSTOP{}})
	this.AddCommand("feedback", &CommandWrapper{WrappedName: "feedback", WrappedCommand: &applesoft.StandardCommandFEEDBACK{}})
	this.AddCommand("cls", &CommandWrapper{WrappedName: "cls", WrappedCommand: &applesoft.StandardCommandCLS{}})
	this.AddCommand("eighty", &CommandWrapper{WrappedName: "eighty", WrappedCommand: &applesoft.StandardCommandPR{}, WrappedParam: []*types.Token{types.NewToken(types.INTEGER, "3")}})
	this.AddCommand("forty", &CommandWrapper{WrappedName: "forty", WrappedCommand: &applesoft.StandardCommandPR{}, WrappedParam: []*types.Token{types.NewToken(types.INTEGER, "0")}})
	this.AddCommand("run", &CommandWrapper{WrappedName: "run", WrappedCommand: &applesoft.StandardCommandRUN{}})
	this.AddCommand("edit", &CommandWrapper{WrappedName: "edit", WrappedCommand: &applesoft.StandardCommandEDIT{}})
	this.AddCommand("prompt", &CommandWrapper{WrappedName: "prompt", WrappedCommand: &applesoft.StandardCommandINPUT{}})
	this.AddCommand("lowgr", &CommandWrapper{WrappedName: "lowgr", WrappedCommand: &applesoft.StandardCommandGR{}})
	this.AddCommand("help", &CommandWrapper{WrappedName: "help", WrappedCommand: &applesoft.StandardCommandHELP{}})
	this.AddCommand("#", &CommandWrapper{WrappedName: "#", WrappedCommand: applesoft.NewStandardCommandREM()})
	this.AddCommand("load", &CommandWrapper{WrappedName: "load", WrappedCommand: &applesoft.StandardCommandLOAD{}})
	this.AddCommand("text", &CommandWrapper{WrappedName: "text", WrappedCommand: &applesoft.StandardCommandTEXT{}})

	this.AddCommand("fp", &CommandWrapper{WrappedName: "applesoft", WrappedCommand: NewStandardCommandDIALECT(), WrappedParam: []*types.Token{types.NewToken(types.VARIABLE, "fp")}})
	this.AddCommand("logo", &CommandWrapper{WrappedName: "logo", WrappedCommand: NewStandardCommandDIALECT(), WrappedParam: []*types.Token{types.NewToken(types.VARIABLE, "logo")}})
	this.AddCommand("int", &CommandWrapper{WrappedName: "integer", WrappedCommand: NewStandardCommandDIALECT(), WrappedParam: []*types.Token{types.NewToken(types.VARIABLE, "int")}})
	this.AddCommand("lang", &CommandWrapper{WrappedName: "lang", WrappedCommand: NewStandardCommandDIALECT()})

	this.AddCommand("splash", &PlusWrapper{WrappedName: "splash", WrappedCommand: plus.NewPlusSplash(0, 0, *types.NewTokenList())})

	this.AddCommand("textcolor", &PlusWrapper{WrappedName: "textcolor", WrappedCommand: plus.NewPlusFGColor(0, 0, *types.NewTokenList())})
	this.AddCommand("bgcolor", &PlusWrapper{WrappedName: "bgcolor", WrappedCommand: plus.NewPlusBGColor(0, 0, *types.NewTokenList())})
	this.AddCommand("spacecolor", &PlusWrapper{WrappedName: "spacecolor", WrappedCommand: plus.NewPlusCGColor(0, 0, *types.NewTokenList())})

	this.AddCommand("ls", &PlusWrapper{
		WrappedName:    "ls",
		WrappedCommand: plus.NewPlusDir(0, 0, *types.NewTokenList()),
		Matchers: []PlusRegMatcher{
			PlusRegMatcher{
				RegExp: regexp.MustCompile("^(.*)[/]([^/]+)$"),
				RegToToken: []*types.Token{
					types.NewToken(types.STRING, "$1:${cwd}"),
					types.NewToken(types.STRING, "$2:*.*"),
				},
			},
			PlusRegMatcher{
				RegExp: regexp.MustCompile("^([^/]+)$"),
				RegToToken: []*types.Token{
					types.NewToken(types.STRING, ":${cwd}"),
					types.NewToken(types.STRING, "$1:*.*"),
				},
			},
			PlusRegMatcher{
				RegExp: regexp.MustCompile("^$"),
				RegToToken: []*types.Token{
					types.NewToken(types.STRING, ":${cwd}"),
					types.NewToken(types.STRING, ":*.*"),
				},
			},
		},
	})

	this.AddCommand("dir", &PlusWrapper{WrappedName: "dir", WrappedCommand: plus.NewPlusDir(0, 0, *types.NewTokenList())})
	this.AddCommand("cd", &PlusWrapper{WrappedName: "cd", WrappedCommand: plus.NewPlusCd(0, 0, *types.NewTokenList())})
	this.AddCommand("rm", &PlusWrapper{WrappedName: "rm", WrappedCommand: plus.NewPlusDelete(0, 0, *types.NewTokenList())})
	this.AddCommand("del", &PlusWrapper{WrappedName: "rm", WrappedCommand: plus.NewPlusDelete(0, 0, *types.NewTokenList())})
	this.AddCommand("mkdir", &PlusWrapper{WrappedName: "mkdir", WrappedCommand: plus.NewPlusMkDir(0, 0, *types.NewTokenList())})
	//this.AddCommand("copy", &PlusWrapper{WrappedName: "copy", WrappedCommand: plus.NewPlusCopy(0, 0, *types.NewTokenList())})
	this.AddCommand("cp", &PlusWrapper{
		WrappedName:    "copy",
		WrappedCommand: plus.NewPlusCopy(0, 0, *types.NewTokenList()),
	})

	this.AddCommand("nobreak", &PlusWrapper{WrappedName: "nobreak", WrappedCommand: plus.NewPlusNoBreak(0, 0, *types.NewTokenList())})
	this.AddCommand("norestore", &PlusWrapper{WrappedName: "norestore", WrappedCommand: plus.NewPlusNoRestore(0, 0, *types.NewTokenList())})

	this.AddCommand("set", &CommandWrapper{WrappedName: "set", WrappedCommand: &applesoft.StandardCommandIMPLIEDASSIGN{}})

	this.AddCommand("motd", &PlusWrapper{WrappedName: "motd", WrappedCommand: plus.NewPlusDisplayMOTD(0, 0, *types.NewTokenList())})

	this.AddFunction("abs(", applesoft.NewStandardFunctionABS(0, 0, *types.NewTokenList()))
	this.AddFunction("fix(", applesoft.NewStandardFunctionINT(0, 0, *types.NewTokenList()))
	this.AddFunction("rnd(", applesoft.NewStandardFunctionASRND(0, 0, *types.NewTokenList()))
	this.AddFunction("sqr(", applesoft.NewStandardFunctionSQR(0, 0, *types.NewTokenList()))
	this.AddFunction("log(", applesoft.NewStandardFunctionLOG(0, 0, *types.NewTokenList()))
	this.AddFunction("exp(", applesoft.NewStandardFunctionEXP(0, 0, *types.NewTokenList()))
	this.AddFunction("int(", applesoft.NewStandardFunctionINT(0, 0, *types.NewTokenList()))
	//this.AddFunction( "cint", NewStandardFunctionINT(0,0,nil) );
	//this.AddFunction( "csng", NewStandardFunctionFNOP(0,0,nil) );
	//this.AddFunction( "cdbl", NewStandardFunctionFNOP(0,0,nil) );
	this.AddFunction("sin(", applesoft.NewStandardFunctionSIN(0, 0, *types.NewTokenList()))
	this.AddFunction("cos(", applesoft.NewStandardFunctionCOS(0, 0, *types.NewTokenList()))
	this.AddFunction("tan(", applesoft.NewStandardFunctionTAN(0, 0, *types.NewTokenList()))
	this.AddFunction("atn(", applesoft.NewStandardFunctionATAN(0, 0, *types.NewTokenList()))
	this.AddFunction("sgn(", applesoft.NewStandardFunctionSGN(0, 0, *types.NewTokenList()))

	this.AddFunction("ordinal(", applesoft.NewStandardFunctionASC(0, 0, *types.NewTokenList()))
	this.AddFunction("char(", applesoft.NewStandardFunctionCHRDollar(0, 0, *types.NewTokenList()))
	this.AddFunction("left(", applesoft.NewStandardFunctionLEFTDollar(0, 0, *types.NewTokenList()))
	this.AddFunction("right(", applesoft.NewStandardFunctionRIGHTDollar(0, 0, *types.NewTokenList()))
	this.AddFunction("substr(", applesoft.NewStandardFunctionMIDDollar(0, 0, *types.NewTokenList()))
	this.AddFunction("length(", applesoft.NewStandardFunctionLEN(0, 0, *types.NewTokenList()))
	this.AddFunction("string(", applesoft.NewStandardFunctionSTRDollar(0, 0, *types.NewTokenList()))
	// this.AddFunction( "string$", NewStandardFunctionSTRINGDollar(0,0,nil) );
	this.AddFunction("value(", applesoft.NewStandardFunctionVAL(0, 0, *types.NewTokenList()))
	//this.AddFunction( "inkey$", NewStandardFunctionINKEYDollar(0,0,nil) );
	this.AddFunction("tab(", applesoft.NewStandardFunctionTAB(0, 0, *types.NewTokenList()))
	this.AddFunction("spc(", applesoft.NewStandardFunctionSPC(0, 0, *types.NewTokenList()))
	this.AddFunction("pos(", applesoft.NewStandardFunctionPOS(0, 0, *types.NewTokenList()))
	this.AddFunction("peek(", applesoft.NewStandardFunctionPEEK())
	this.AddFunction("scrn(", applesoft.NewStandardFunctionSCRN(0, 0, *types.NewTokenList()))
	//this.AddFunction("hscrn(", applesoft.NewStandardFunctionHSCRN(0, 0, *types.NewTokenList()))
	this.AddFunction("pdl(", applesoft.NewStandardFunctionPDL(0, 0, *types.NewTokenList()))
	this.AddFunction("fre(", applesoft.NewStandardFunctionABS(0, 0, *types.NewTokenList()))

	this.AddPlusFunction("color", "bg{", plus.NewPlusBGColor(0, 0, *types.NewTokenList()))
	this.AddPlusFunction("color", "fg{", plus.NewPlusFGColor(0, 0, *types.NewTokenList()))
	this.AddPlusFunction("color", "cg{", plus.NewPlusCGColor(0, 0, *types.NewTokenList()))

	this.AddPlusFunction("system", "spawn{", plus.NewPlusSpawn(0, 0, *types.NewTokenList()))
	this.AddPlusFunction("system", "echo{", plus.NewPlusEcho(0, 0, *types.NewTokenList()))
	this.AddPlusFunction("system", "nobreak{", plus.NewPlusNoBreak(0, 0, *types.NewTokenList()))
	this.AddPlusFunction("system", "exit{", plus.NewPlusExit(0, 0, *types.NewTokenList()))
	this.AddPlusFunction("mode", "video{", plus.NewPlusVideoMode(0, 0, *types.NewTokenList()))
	this.AddPlusFunction("mode", "font{", plus.NewPlusTextMode(0, 0, *types.NewTokenList()))
	this.AddPlusFunction("mode", "hgr{", plus.NewPlusSwitchHGR(0, 0, *types.NewTokenList()))

	/* added functions for fun */
	//this.AddFunction( "time$", NewStandardFunctionTIMEDollar(0,0,nil) );
	this.AddPlusFunction("system", "cputhrottle{", plus.NewPlusCPUThrottle(0, 0, *types.NewTokenList()))
	this.AddPlusFunction("system", "uptime{", plus.NewPlusUpTime(0, 0, *types.NewTokenList()))
	this.AddPlusFunction("system", "boottime{", plus.NewPlusBootTime(0, 0, *types.NewTokenList()))
	this.AddPlusFunction("system", "pause{", plus.NewPlusPause(0, 0, *types.NewTokenList()))
	this.AddPlusFunction("system", "connect{", plus.NewPlusConnect(0, 0, *types.NewTokenList()))
	//	this.AddPlusFunction("system", "stack{", plus.NewPlusStack(0, 0, *types.NewTokenList()))
	//	this.AddPlusFunction("system", "publish{", plus.NewPlusPublish(0, 0, *types.NewTokenList()))
	this.AddPlusFunction("system", "shim{", plus.NewPlusShim(0, 0, *types.NewTokenList()))

	this.AddPlusFunction("dos", "mkdir{", plus.NewPlusMkDir(0, 0, *types.NewTokenList()))
	this.AddPlusFunction("dos", "cd{", plus.NewPlusCd(0, 0, *types.NewTokenList()))
	this.AddPlusFunction("dos", "dir{", plus.NewPlusDir(0, 0, *types.NewTokenList()))
	this.AddPlusFunction("dos", "ls{", plus.NewPlusDir(0, 0, *types.NewTokenList()))
	this.AddPlusFunction("dos", "del{", plus.NewPlusDelete(0, 0, *types.NewTokenList()))
	this.AddPlusFunction("dos", "rm{", plus.NewPlusDelete(0, 0, *types.NewTokenList()))
	this.AddPlusFunction("dos", "copy{", plus.NewPlusCopy(0, 0, *types.NewTokenList()))
	this.AddPlusFunction("dos", "cp{", plus.NewPlusCopy(0, 0, *types.NewTokenList()))
	this.AddPlusFunction("dos", "grant{", plus.NewPlusGrant(0, 0, *types.NewTokenList()))
	this.AddPlusFunction("dos", "revoke{", plus.NewPlusRevoke(0, 0, *types.NewTokenList()))

	this.AddPlusFunction("dos", "paramcount{", plus.NewPlusParamCount(0, 0, *types.NewTokenList()))
	this.AddPlusFunction("dos", "param{", plus.NewPlusParam(0, 0, *types.NewTokenList()))
	this.AddPlusFunction("dos", "programdir{", plus.NewPlusProgramDir(0, 0, *types.NewTokenList()))

	this.VarSuffixes = "%$!&#"

	/* dynacode test shim */
	this.ReverseCase = true

	this.ArrayDimDefault = 10
	this.ArrayDimMax = 65535
	this.Title = "Shell"

	this.IPS = -1

}

func (this *DialectShell) ExecuteDirectCommand(tl types.TokenList, ent interfaces.Interpretable, Scope *types.Algorithm, LPC *types.CodeRef) error {

	/* vars */
	var tok *types.Token
	var n string
	var cmd interfaces.Commander
	var cr *types.CodeRef
	var ss *types.TokenList

	ent.SetPrompt("$")

	if this.NetBracketCount(tl) != 0 {
		return exception.NewESyntaxError("SYNTAX ERROR")
	}

	//System.Out.Println( "DEBUG: -------------> ["+this.Title+"]: "+LPC.Line+" "+ent.TokenListAsString(tl) );

	// if this.Trace && (ent.GetState() == types.RUNNING) {
	// 	ent.GetVDU().PutStr("#" + utils.IntToStr(LPC.Line) + " ")
	// }

	/* process poop monster here (@^-^@) */
	tok = tl.Shift()

	if (tok.Type == types.NUMBER) || (tok.Type == types.INTEGER) {
		// nothing
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

	} else if ((tok.Type == types.STRING) || (tok.Type == types.VARIABLE)) && (this.ImpliedAssign == nil) {

		path := ent.GetWorkDir()
		if utils.Pos("/", tok.Content) > 0 {

			path = files.GetPath(tok.Content)
			tok.Content = files.GetFilename(tok.Content)

			if path == "." {
				path = ent.GetWorkDir()
			}

		}

		/* is there a system/command entry */
		fmt.Printf("Shell looking up command: %s\n", tok.Content)

		codeTypes := files.GetTypeCode()

		found := false
		for _, info := range codeTypes {
			fmt.Printf("Trying ext %s for %s\n", info.Ext, tok.Content)
			if files.ExistsViaProvider(path, strings.ToLower(tok.Content)+"."+info.Ext) {
				e := ent.NewChildWithParamsAndTask(strings.ToLower(tok.Content), info.Dialect, &tl, path+"/"+strings.ToLower(tok.Content))
				ent.SetChild(e)
				e.SetParent(ent)
				e.Run(false)
				found = true

				apple2helpers.HGRClear(e)
				apple2helpers.TextMode(e)

				time.Sleep(50 * time.Millisecond)

				index := e.GetMemIndex()
				mm := e.GetMemoryMap()
				cindex := mm.GetCameraConfigure(index)
				for i := 0; i < 8; i++ {
					control := types.NewOrbitController(mm, index, cindex)
					control.ResetALL()
					control.SetZoom(types.GFXMULT)
				}
				break
			}
		}

		if !found {
			this.HandleException(ent, errors.New("Unrecognised command: "+strings.ToLower(tok.Content)))
		}

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

func (this *DialectShell) PutStr(ent interfaces.Interpretable, s string) {
	apple2helpers.PutStr(ent, s)
}

func (this *DialectShell) RealPut(ent interfaces.Interpretable, ch rune) {
	apple2helpers.Put(ent, ch)
}

func (this *DialectShell) Backspace(ent interfaces.Interpretable) {
	apple2helpers.Backspace(ent)
}

func (this *DialectShell) ClearToBottom(ent interfaces.Interpretable) {
	apple2helpers.ClearToBottom(ent)
}

func (this *DialectShell) SetCursorX(ent interfaces.Interpretable, xx int) {
	x := (80 / apple2helpers.GetFullColumns(ent)) * xx

	apple2helpers.SetCursorX(ent, x)
}

func (this *DialectShell) SetCursorY(ent interfaces.Interpretable, yy int) {
	y := (48 / apple2helpers.GetFullRows(ent)) * yy

	apple2helpers.SetCursorY(ent, y)
}

func (this *DialectShell) GetColumns(ent interfaces.Interpretable) int {
	return apple2helpers.GetColumns(ent)
}

func (this *DialectShell) GetRows(ent interfaces.Interpretable) int {
	return apple2helpers.GetRows(ent)
}

func (this *DialectShell) Repos(ent interfaces.Interpretable) {
	apple2helpers.Gotoxy(ent, int(ent.GetMemory(36)), int(ent.GetMemory(37)))
}

func (this *DialectShell) GetCursorX(ent interfaces.Interpretable) int {
	return apple2helpers.GetCursorX(ent) / (80 / apple2helpers.GetFullColumns(ent))
}

func (this *DialectShell) GetCursorY(ent interfaces.Interpretable) int {
	return apple2helpers.GetCursorY(ent) / (48 / apple2helpers.GetFullRows(ent))
}

var reVarname *regexp.Regexp

func init() {
	reVarname, _ = regexp.Compile("^[$][a-zA-Z][a-zA-Z0-9]*$")
}

func (this *DialectShell) IsVariableName(instr string) bool {

	return reVarname.MatchString(instr)

}
