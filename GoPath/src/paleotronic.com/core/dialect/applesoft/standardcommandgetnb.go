package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/exception"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/log"
	"paleotronic.com/runestring" //   "paleotronic.com/files"
	//	"time"
	"paleotronic.com/utils"
)

type StandardCommandGETNB struct {
	dialect.Command
	promptString string
	HighChars    bool
	Masked       bool
	FileMode     bool
	Break        bool
}

func NewStandardCommandGETNB() *StandardCommandGETNB {
	this := &StandardCommandGETNB{}
	this.UseStates = true
	return this
}

func NewStandardCommandGETNBH() *StandardCommandGETNB {
	this := &StandardCommandGETNB{}
	this.UseStates = true
	this.HighChars = true
	return this
}

func NewStandardCommandGETNBM() *StandardCommandGETNB {
	this := &StandardCommandGETNB{}
	this.UseStates = true
	this.Masked = true
	return this
}

func (this *StandardCommandGETNB) StateInit(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	this.FileMode = (len(caller.GetFeedBuffer()) > 0)
	this.Break = false

	//interactive := (len(caller.GetVDU().GetFeedBuffer()) == 0)
	this.promptString = "?"

	scidx := tokens.IndexOf(types.SEPARATOR, "")
	if (scidx > -1) && (tokens.LPeek().Type == types.STRING) {
		prompt := tokens.SubList(0, scidx+1)
		tokens = *tokens.SubList(scidx+1, tokens.Size())

		//try {
		t, err := caller.GetDialect().ParseTokensForResult(caller, *prompt)
		if err != nil {
			return result, err
		}
		this.promptString = t.Content
		//} catch (Exception e) {
		//}
	}

	cs := caller.GetCommandState()

	// tla -> cs.L[0]
	cs.L[0] = caller.SplitOnTokenWithBrackets(tokens, *types.NewToken(types.SEPARATOR, ","))

	// buff -> cs.S[0]
	cs.S[0] = runestring.Cast("")

	cs.Step = NBI_NEXTVAR

	caller.SetSubState(types.ESS_EXEC)

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandGETNB) StateDone(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {
	// Not much
	log.Println("Done")
	return 0, nil
}

func (this *StandardCommandGETNB) StateExec(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	//var promptString string
	result := 0

	cs := caller.GetCommandState()

	switch cs.Step {
	case NBI_REDOVAR:

		log.Println("Redovar")

		this.promptString = "REENTER?"
		if len(cs.S[0].Runes) != 0 {
			cs.Step = NBI_GOTSTR
		} else {
			cs.Step = NBI_STARTLINE
		}

	case NBI_NEXTVAR:

		log.Println("Nextvar")

		/* fetch next var details from cs.L[0] (tla) */
		if this.Break || len(cs.L[0]) == 0 {
			caller.SetSubState(types.ESS_DONE)
			return result, nil
		}

		cs.B[0] = false // valie
		cs.I[0] = 0     // times

		// get next var details
		cs.T[0] = cs.L[0][0]
		cs.L[0] = cs.L[0][1:]

		// bail if next target is not a variable
		if cs.T[0].LPeek().Type != types.VARIABLE {
			caller.SetSubState(types.ESS_DONE)
			return result, exception.NewESyntaxError("SYNTAX ERROR")
		}

		if len(cs.S[0].Runes) != 0 {
			cs.Step = NBI_GOTSTR
		} else {
			cs.Step = NBI_READCHARS
		}

	case NBI_READCHARS:
		/* Collecting the value from input stream */

		//log.Println("Readchar")

		if caller.GetInChannel() != "" {
			// scoop from file
			//
			cs.Step = NBI_GOTSTR
		} else {
			// scoop from user
			if caller.GetMemory(49152) < 128 {
				apple2helpers.TextShowCursor(caller)
				// no char yet..
				cs.SleepCounter = 5
				cs.PostSleepState = types.ESS_EXEC
				caller.SetSubState(types.ESS_SLEEP)
			} else {
				apple2helpers.TextHideCursor(caller)
				// got a char
				ch := rune(caller.GetMemory(49152) & 127)
				caller.SetMemory(49168, 0)

				if caller.GetDialect().IsUpperOnly() && ch >= 'a' && ch <= 'z' {
					ch -= 32
				}

				switch ch {
				case 3:
					{
						//display.SetSuppressFormat(true)
						if caller.IsBreakable() {
							caller.SetMemory(49168, 0)
							caller.PutStr("\r\n")
							//display.SetSuppressFormat(false)
							e := caller.Halt()
							if e != nil {
								caller.GetDialect().HandleException(caller, e)
							}
							this.Break = true
							cs.Step = NBI_GOTSTR
						}
					}
				case 10:
					{
						//display.SetSuppressFormat(true)
						caller.PutStr("\r\n")
						//display.SetSuppressFormat(false)
						cs.Step = NBI_GOTSTR
					}
				case 13:
					{
						//display.SetSuppressFormat(true)
						caller.PutStr("\r\n")
						//display.SetSuppressFormat(false)
						cs.Step = NBI_GOTSTR
					}
				case 8:
					{
						if len(cs.S[0].Runes) > 0 {
							cs.S[0] = runestring.Copy(cs.S[0], 1, len(cs.S[0].Runes)-1)
							caller.Backspace()
							//						display.SetSuppressFormat(true)
							caller.PutStr(" ")
							//display.SetSuppressFormat(false)
							caller.Backspace()
						}
						break
					}
				default:
					{

						if this.Masked {
							caller.RealPut('*')
						} else {
							caller.RealPut(rune(ch))
						}
						//display.SetSuppressFormat(false)

						if this.HighChars {
							ch |= 128
						}

						cs.S[0].Runes = append(cs.S[0].Runes, ch)
						break
					}
				}

			}
		}

	case NBI_GOTSTR:
		/* Got a complete string - cs.S */

		log.Println("Gotstr")

		chunk := runestring.Cast("")
		zz := runestring.Pos(',', cs.S[0])

		if zz > 0 {
			chunk = runestring.Copy(cs.S[0], 1, zz-1)
			cs.S[0] = runestring.Delete(cs.S[0], 1, zz)
		} else {
			chunk = cs.S[0]
			cs.S[0] = runestring.Cast("")
		}

		// now do an assignment
		c := types.NewTokenList()
		tl := cs.T[0]
		for _, t := range tl.Content {
			c.Push(t)
		}

		if utils.Pos("$", c.LPeek().Content) <= 0 {
			// not string
			if len(chunk.Runes) == 0 {
				chunk = runestring.Cast("0")
			}
		}

		c.Push(types.NewToken(types.ASSIGNMENT, "="))
		if utils.Pos("$", c.LPeek().Content) <= 0 {
			c.Push(types.NewToken(types.NUMBER, chunk.String()))
		} else {
			c.Push(types.NewToken(types.STRING, chunk.String()))
		}

		cs.B[0] = true
		_, e := caller.GetDialect().GetImpliedAssign().Execute(env, caller, *c, Scope, LPC)
		cs.Step = NBI_NEXTVAR
		if e != nil {
			cs.B[0] = false
			cs.Step = NBI_REDOVAR
		}
		this.promptString = "??"

	default:
		/* something else? */
	}

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandGETNB) Syntax() string {

	/* vars */
	var result string

	result = "GET <var>[,<var>,...]"

	/* enforce non void return */
	return result

}

func (this *StandardCommandGETNB) SetHigh() {
	this.HighChars = true
}
