package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type StandardFunctionPDL struct {
	dialect.CoreFunction
}

func (this *StandardFunctionPDL) Syntax() string {

	/* vars */
	var result string

	result = "PDL(<number>)"

	/* enforce non void return */
	return result

}

func (this *StandardFunctionPDL) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewStandardFunctionPDL(a int, b int, params types.TokenList) *StandardFunctionPDL {
	this := &StandardFunctionPDL{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "PDL"
	this.MinParams = 1
	this.MaxParams = 1

	return this
}

func (this *StandardFunctionPDL) FunctionExecute(params *types.TokenList) error {

	/* vars */
	var value int

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	settings.LogoCameraControl[this.Interpreter.GetMemIndex()] = false

	value = this.Stack.Pop().AsInteger() % 4

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(int(this.Interpreter.GetMemoryMap().IntGetPaddleValue(this.Interpreter.GetMemIndex(), value)))))

	return nil
}
