package applesoft

import (
	"paleotronic.com/fmt"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"

	"paleotronic.com/core/hardware/apple2helpers"
)

type StandardFunctionUSR struct {
	dialect.CoreFunction
}

func NewStandardFunctionUSR(a int, b int, params types.TokenList) *StandardFunctionUSR {
	this := &StandardFunctionUSR{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "USR"

	return this
}

func (this *StandardFunctionUSR) FunctionExecute(params *types.TokenList) error {

	/* vars */
	var value float64

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		fmt.Println(e)
		return e
	}

	value = this.Stack.Shift().AsExtended()

	// load FAC with value
	f5 := types.NewFloat5b(value)
	f5.WriteMemoryFACFormat(this.Interpreter.GetMemoryMap(), this.Interpreter.GetMemIndex(), 0x9d)

	// invoke code call
	apple2helpers.DoCall(0x000a, this.Interpreter, true)

	// get fac value back
	f5.ReadMemoryFACFormat(this.Interpreter.GetMemoryMap(), this.Interpreter.GetMemIndex(), 0x9d)
	value = f5.GetValue()

	this.Stack.Push(types.NewToken(types.NUMBER, utils.FormatFloat("", value)))

	return nil
}

func (this *StandardFunctionUSR) Syntax() string {

	/* vars */
	var result string

	result = "USR(<number>)"

	/* enforce non void return */
	return result

}

func (this *StandardFunctionUSR) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	//SetLength( result, 1 );
	result = append(result, types.NUMBER)
	this.MinParams = 1
	this.MaxParams = 1

	/* enforce non void return */
	return result

}
