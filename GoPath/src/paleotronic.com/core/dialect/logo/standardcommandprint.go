package logo

import (
	"math"
	"strings"
	"time"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type StandardCommandPRINT struct {
	dialect.Command
	NoNL bool
}

func f(t types.Token, c interfaces.Interpretable) string {
	if t.Type == types.NUMBER {
		return utils.StrToFloatStrApple(t.Content)
	} else if t.Type == types.LIST {
		return c.TokenListAsString(*t.List)
	}
	return t.Content
}

func (this *StandardCommandPRINT) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int
	var t types.Token
	//Token  tt;
	var tl types.TokenList
	var nl bool
	var bc int

	/* print assumes an expression, let's parse it */

	/* process tokens */
	nl = true
	result = 0
	bc = 0
	tl = *types.NewTokenList()
	for _, tt := range tokens.Content {
		//    	//fmt.Printf("Checking token: %d) %s / %s / %v\n", i, tt.Type, tt.Content, nl)
		if bc > 0 {
			if tt.Type == types.OBRACKET || tt.Type == types.FUNCTION || tt.Type == types.PLUSFUNCTION {
				bc++
				//tl.Push(tt);
			} else if tt.Type == types.CBRACKET {
				bc--
			}
			tl.Push(tt)
		} else {
			if (tt.Type == types.FUNCTION) && ((strings.ToLower(tt.Content) == "tab(") || (strings.ToLower(tt.Content) == "spc(")) {
				nl = false
				if tl.Size() > 0 {
					tt, err := this.Command.D.(*DialectLogo).ParseTokensForResult(caller, tl)
					if err != nil {
						return result, err
					}

					this.PutStr(f(*tt, caller), caller)
				}
				tl = *types.NewTokenList()
				tl.Push(tt)
				bc++
				//                //fmt.Println("SPC / TAB")
			} else if tt.Type == types.OBRACKET || tt.Type == types.FUNCTION || tt.Type == types.PLUSFUNCTION {
				bc++
				tl.Push(tt)
			} else if tt.Type == types.SEPARATOR {
				if tt.Content == ";" {
					t = caller.ParseTokensForResult(tl)
					this.PutStr(f(t, caller), caller)
					nl = false
					tl = *types.NewTokenList()
				} else if tt.Content == "," {
					t = caller.ParseTokensForResult(tl)
					this.PutStr(f(t, caller), caller)
					this.PutStr("\t", caller)
					nl = false
					tl = *types.NewTokenList()
				}
			} else {
				if tl.Size() > 0 && tl.RPeek().Type == types.STRING {
					t = caller.ParseTokensForResult(tl)
					this.PutStr(f(t, caller), caller)
					tl = *types.NewTokenList()
					tl.Push(tt)
				} else {
					tl.Push(tt)
					if bc == 0 {
						nl = true
					}
				}
			}
		}
	}

	if tl.Size() > 0 {
		tt, err := this.Command.D.(*DialectLogo).ParseTokensForResult(caller, tl)

		if err != nil {
			return result, err
		}
		// removed free call here /*free tokens*/
		this.PutStr(f(*tt, caller), caller)

		zt := tl.LPeek()
		//        rt := tl.RPeek()
		nl = ((tl.RPeek().Type == types.STRING) || ((zt.Type != types.SEPARATOR) && (strings.ToLower(zt.Content) != "spc(") && (strings.ToLower(zt.Content) != "tab(")))

	}

	////fmt.Printf( "PRINT NL=%v\n", nl )

	if nl && !this.NoNL {
		this.PutStr("\r\n", caller)
	}

	//this.Cost = (1000000000 / 800) * 10 * tokens.Size();

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandPRINT) PutStr(s string, caller interfaces.Interpretable) {

	speed := float64(caller.GetSpeed())

	cps := -128*math.Log((256-speed)/256) + 2

	delay := 1000000 / cps

	if speed >= 255 {
		delay = 0
	}

	for _, ch := range s {
		apple2helpers.RealPut(caller, ch)
		time.Sleep(time.Duration(delay) * time.Microsecond)
	}

	//apple2helpers.PutStr(caller, s)
}

func (this *StandardCommandPRINT) Syntax() string {

	/* vars */
	var result string

	result = "PRINT <expression> [<;|>, <expression>]"

	/* enforce non void return */
	return result

}
