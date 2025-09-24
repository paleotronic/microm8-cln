package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type StandardFunctionPOS struct {
	dialect.CoreFunction
}

func (this *StandardFunctionPOS) Syntax() string {

	/* vars */
	var result string

	result = "POS(<number>)"

	/* enforce non void return */
	return result

}

func (this *StandardFunctionPOS) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	//( result, 1 );
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewStandardFunctionPOS(a int, b int, params types.TokenList) *StandardFunctionPOS {
	this := &StandardFunctionPOS{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "POS"

	return this
}

func (this *StandardFunctionPOS) FunctionExecute(params *types.TokenList) error {

	/* vars */
	//var value int

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }
	this.Stack.Clear()

	//value = this.Stack.Pop().AsInteger()

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(this.Interpreter.GetCursorX())))

	return nil
}
