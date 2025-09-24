package logo

import (
	"errors"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	//"paleotronic.com/utils"

	"time"
)

type StandardFunctionREADCHARS struct {
	dialect.CoreFunction
}

func (this *StandardFunctionREADCHARS) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewStandardFunctionREADCHARS(a int, b int, params types.TokenList) *StandardFunctionREADCHARS {
	this := &StandardFunctionREADCHARS{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "READCHARS"
	this.MinParams = 1
	this.MaxParams = 1

	return this
}

func (this *StandardFunctionREADCHARS) FunctionExecute(params *types.TokenList) error {

	/* vars */
	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	if params.Size() == 0 {
		return errors.New("I NEED A VALUE")
	}

	needed := params.Shift().AsInteger()

	out := ""
	for needed > 0 {

		for this.Interpreter.GetMemory( 49152 ) < 128 {
			time.Sleep(5 * time.Millisecond)
		}

		ch := rune(this.Interpreter.GetMemory( 49152 ) & 0xff7f)
		out += string(ch)

		this.Interpreter.SetMemory(49168,0)

		needed --
	}

	// now got char
	

	this.Stack.Push(types.NewToken(types.WORD, out))
	

	return nil
}

func (this *StandardFunctionREADCHARS) Syntax() string {

	/* vars */
	var result string

	result = "READCHARS a b"

	/* enforce non void return */
	return result

}
