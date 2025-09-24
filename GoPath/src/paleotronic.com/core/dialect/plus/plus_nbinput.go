package plus

import (
	"time"

	"paleotronic.com/runestring"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/types"

	//"paleotronic.com/log"
	//	"paleotronic.com/fmt"
)

type PlusInput struct {
	dialect.CoreFunction
	InProgress bool
	CVar       string
	Content    runestring.RuneString
	Prompt     string
	cx, cy     int
	maxwidth   int
}

func (this *PlusInput) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	if !this.Query {

		varname := this.ValueMap["var"].Content
		this.Prompt = this.ValueMap["prompt"].Content

		//log.Printf("non-breaking input for %s\n", varname)

		// new or existing?
		if !this.InProgress {
			this.cx, this.cy = apple2helpers.GetRealCursorPos(this.Interpreter)

			this.InProgress = true
			this.CVar = varname
			this.Content = runestring.Cast("")
			nn := this.ValueMap["max"]
			this.maxwidth = nn.AsInteger()
		} else {
			// continuing

			//	ox := apple2helpers.GetCursorX(this.Interpreter)
			//	oy := apple2helpers.GetCursorY(this.Interpreter)

			ox, oy := apple2helpers.GetRealCursorPos(this.Interpreter)

			v := this.Interpreter.GetMemory(49152)
			if v&128 != 0 {
				// new char
				ch := rune(v & 0xff7f)

				switch {
				case ch == 8:
					// backspace
					if len(this.Content.Runes) > 0 {
						this.Content = runestring.Delete(this.Content, len(this.Content.Runes), 1)
					}
				case ch == 13:
					// enter
					a := this.Interpreter.GetCode()
					tl := types.NewTokenList()
					tl.Push(types.NewToken(types.VARIABLE, this.CVar))
					tl.Push(types.NewToken(types.ASSIGNMENT, "="))
					tl.Push(types.NewToken(types.STRING, string(this.Content.Runes)))
					//fmt.Println( this.Interpreter.TokenListAsString(*tl) )
					this.Interpreter.GetDialect().ExecuteDirectCommand(*tl, this.Interpreter, a, this.Interpreter.GetPC())
					//					log.Println(tl.AsString())
					this.InProgress = false
					apple2helpers.PutStr(this.Interpreter, "\r\n")
				case ch == 3:
					// do nothing here
				default:
					// just process char
					this.Content.AppendSlice([]rune{ch})
				}

				this.Interpreter.SetMemory(49168, 0) // clear keystrobe
			}

			//apple2helpers.Gotoxy(this.Interpreter, this.cx, this.cy)
			apple2helpers.SetRealCursorPos(this.Interpreter, this.cx, this.cy)
			apple2helpers.PutStr(this.Interpreter, this.Prompt)

			apple2helpers.ClearToBottom(this.Interpreter)

			s := 0
			if len(this.Content.Runes) > this.maxwidth {
				s = len(this.Content.Runes) - this.maxwidth
			}
			str := string(this.Content.Runes[s:])

			apple2helpers.PutStr(this.Interpreter, str)
			apple2helpers.Put(this.Interpreter, rune(127)) // pseudo cursor

			//apple2helpers.ClearToBottom(this.Interpreter)

			//	apple2helpers.Gotoxy(this.Interpreter, ox, oy)
			apple2helpers.SetRealCursorPos(this.Interpreter, ox, oy)

			time.Sleep(100 * time.Millisecond)

		}

	}

	this.Stack.Push(types.NewToken(types.STRING, ""))

	return nil
}

func (this *PlusInput) Syntax() string {

	/* vars */
	var result string

	result = "INPUT{var}"

	/* enforce non void return */
	return result

}

func (this *PlusInput) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusInput(a int, b int, params types.TokenList) *PlusInput {
	this := &PlusInput{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "INPUT"

	this.NamedParams = []string{"prompt", "var", "max"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.STRING, "?"),
		*types.NewToken(types.STRING, ""),
		*types.NewToken(types.NUMBER, "10"),
	}
	this.Raw = true

	this.InProgress = false

	return this
}
