package applesoft

import (
	"time"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/exception"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/files"
	"paleotronic.com/runestring"
	"paleotronic.com/utils"
)

func NewStandardCommandINPUT() *StandardCommandINPUT {
	this := &StandardCommandINPUT{}
	return this
}

func NewStandardCommandINPUTH() *StandardCommandINPUT {
	this := &StandardCommandINPUT{}
	//this.HighChars = true
	this.IgnoreComma = true
	return this
}

func NewStandardCommandINPUTM() *StandardCommandINPUT {
	this := &StandardCommandINPUT{}
	this.Masked = true
	return this
}

type StandardCommandINPUT struct {
	dialect.Command
	promptString string
	HighChars    bool
	Masked       bool
	FileMode     bool
	Break        bool
	IgnoreComma  bool
}

func (this *StandardCommandINPUT) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	//var promptString string
	result := 0

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
	tla := caller.SplitOnTokenWithBrackets(tokens, *types.NewToken(types.SEPARATOR, ","))

	buff := ""

	for _, tl := range tla {

		if this.Break {
			break
		}

		valid := false
		times := 0

		for !valid {

			if times > 0 {
				this.promptString = "REENTER?"
			}

			if tl.LPeek().Type != types.VARIABLE {
				return result, exception.NewESyntaxError("SYNTAX ERROR")
			}

			if len(buff) == 0 {
				buff = this.GetLine(caller)
			}

			if caller.GetMemoryMap().IntGetSlotRestart(caller.GetMemIndex()) {
				caller.Halt()
				return 0, nil
			}

			chunk := ""
			zz := 0
			if !this.IgnoreComma {
				zz = utils.PosRune(',', buff)
			}

			if zz > 0 {
				chunk = utils.Copy(buff, 1, zz-1)
				buff = utils.Delete(buff, 1, zz)
			} else {
				chunk = buff
				buff = ""
			}

			// now do an assignment
			c := types.NewTokenList()
			for _, t := range tl.Content {
				c.Push(t)
			}

			if utils.Pos("$", c.LPeek().Content) <= 0 {
				// not string
				if chunk == "" {
					chunk = "0"
				}
			}

			c.Push(types.NewToken(types.ASSIGNMENT, "="))
			if utils.Pos("$", c.LPeek().Content) <= 0 {
				c.Push(types.NewToken(types.NUMBER, chunk))
			} else {
				c.Push(types.NewToken(types.STRING, chunk))
			}

			//try {
			caller.GetDialect().GetImpliedAssign().Execute(env, caller, *c, Scope, LPC)
			valid = true
			//} catch (Exception e) {
			//	times += 1;
			//	buff = ""
			//}
			this.promptString = "??"

		}

		this.promptString = "?"

	}

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandINPUT) GetLine(caller interfaces.Interpretable) string {
	if caller.GetInChannel() != "" {
		return this.GetFileLine(caller)
	} else {
		return this.GetCRTLine(caller)
	}
}

func (this *StandardCommandINPUT) GetFileLine(caller interfaces.Interpretable) string {

	chunk := ""

	data, _ := files.DOSINPUT(
		files.GetPath(caller.GetInChannel()),
		files.GetFilename(caller.GetInChannel()),
	)

	chunk = string(data)

	return chunk
}

func (this *StandardCommandINPUT) GetCRTLine(caller interfaces.Interpretable) string {

	command := ""
	collect := true

	caller.SetBuffer(runestring.NewRuneString())

	caller.PutStr(this.promptString)
	this.promptString = "?"

	cb := caller.GetProducer().GetMemoryCallback(caller.GetMemIndex())

	lastKey := time.Now()
	gap := time.Millisecond * 1000 / time.Duration(settings.PasteCPS)

	for collect {

		caller.Post()

		apple2helpers.TextShowCursor(caller)

		if cb != nil {
			cb(caller.GetMemIndex())
		}

		for caller.GetMemory(49152) < 128 {

			if caller.GetMemoryMap().IntGetSlotInterrupt(caller.GetMemIndex()) {
				return command
			} else if caller.GetMemoryMap().IntGetSlotRestart(caller.GetMemIndex()) {
				return ""
			} else if caller.VM().IsDying() {
				return ""
			} else if caller.GetPasteBuffer().Length() > 0 && time.Since(lastKey) > gap {
				r := caller.GetPasteBuffer()
				v := r.Runes[0]
				r.Runes = r.Runes[1:]
				caller.SetPasteBuffer(r)
				caller.GetMemoryMap().KeyBufferAddNoRedirect(caller.GetMemIndex(), uint64(v))
				lastKey = time.Now()
			} else if !caller.VM().IsDying() {
				caller.WaitForWorld()
				time.Sleep(10 * time.Millisecond)
			}

		}

		apple2helpers.TextHideCursor(caller)

		//if len(caller.GetBuffer().Runes) > 0 {
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
					return command
				}
			}
		case 10:
			{
				//display.SetSuppressFormat(true)
				caller.PutStr("\r\n")
				//display.SetSuppressFormat(false)
				return command
			}
		case 13:
			{
				//display.SetSuppressFormat(true)
				caller.PutStr("\r\n")
				//display.SetSuppressFormat(false)
				return command
			}
		case 127:
			{
				if len(command) > 0 {
					command = utils.Copy(command, 1, len(command)-1)
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

				//             	if !caller.GetDialect().IsUpperOnly() {
				//			      if (ch >= 'a') && (ch <= 'z') {
				//				      ch -= 32
				//			      } else if (ch >= 'A') && (ch <= 'Z') {
				//				      ch += 32
				//			      }
				//                }

				//display.SetSuppressFormat(true)

				if this.Masked {
					caller.RealPut('*')
				} else {
					caller.RealPut(rune(ch))
				}
				//display.SetSuppressFormat(false)

				if this.HighChars {
					ch |= 128
				}

				command = command + string(ch)
				break
			}
		}
		//} else {
		//	time.Sleep(50 * time.Millisecond)
		//}

		if cb != nil {
			cb(caller.GetMemIndex())
		}
	}

	return command

}

func (this *StandardCommandINPUT) Syntax() string {

	/* vars */
	var result string

	result = "GET <var>[,<var>,...]"

	/* enforce non void return */
	return result

}

func (this *StandardCommandINPUT) SetHigh() {
	this.HighChars = true
}
