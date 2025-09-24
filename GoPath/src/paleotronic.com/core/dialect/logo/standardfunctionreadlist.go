package logo

import (
	//	"strings"
	"time"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/runestring"
	"paleotronic.com/utils"
)

type StandardFunctionREADLIST struct {
	dialect.CoreFunction
}

func (this *StandardFunctionREADLIST) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewStandardFunctionREADLIST(a int, b int, params types.TokenList) *StandardFunctionREADLIST {
	this := &StandardFunctionREADLIST{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "READLIST"
	this.MinParams = 0
	this.MaxParams = 0

	return this
}

func (this *StandardFunctionREADLIST) FunctionExecute(params *types.TokenList) error {

	/* vars */
	//	var b, a float64

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	thing := types.NewToken(types.LIST, "")
	thing.List = types.NewTokenList()

	// request input here
	line := this.GetCRTLine( this.Interpreter )
	thing.List = this.Interpreter.GetDialect().Tokenize( runestring.Cast(line) )
	for _, t := range thing.List.Content {
		if t.Type != types.LIST {
			t.Type = types.WORD
		}
	}

	this.Stack.Push(thing)

	return nil
}

func (this *StandardFunctionREADLIST) Syntax() string {

	/* vars */
	var result string

	result = "READLIST word list"

	/* enforce non void return */
	return result

}


func (this *StandardFunctionREADLIST) GetCRTLine(caller interfaces.Interpretable) string {

	command := ""
	collect := true

	caller.SetBuffer(runestring.NewRuneString())

	for collect {

		caller.Post()

        apple2helpers.TextShowCursor(caller)
        
		for caller.GetMemory(49152) < 128 {
			time.Sleep(10*time.Millisecond)
		}
        
        apple2helpers.TextHideCursor(caller)

		//if len(caller.GetBuffer().Runes) > 0 {
		ch := rune(caller.GetMemory(49152) & 127)
		caller.SetMemory(49168,0)
        
    	if caller.GetDialect().IsUpperOnly() && ch >= 'a' && ch <= 'z' {
           ch -= 32
        }

		switch ch {
		case 3:
			{
				//display.SetSuppressFormat(true)
				caller.SetMemory(49168,0)
				caller.PutStr("\r\n")
				//display.SetSuppressFormat(false)
				e := caller.Halt()
				if e != nil {
					caller.GetDialect().HandleException(caller, e)
				}
				return command
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
		case 8:
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
				caller.RealPut(rune(ch))
				//display.SetSuppressFormat(false)



				command = command + string(ch)
				break
			}
		}
		//} else {
		//	time.Sleep(50 * time.Millisecond)
		//}
	}

	return command

}
