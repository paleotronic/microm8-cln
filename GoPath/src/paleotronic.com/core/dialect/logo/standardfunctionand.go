package logo

import (
	//	"math"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type StandardFunctionAND struct {
	dialect.CoreFunction
}

func (this *StandardFunctionAND) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewStandardFunctionAND(a int, b int, params types.TokenList) *StandardFunctionAND {
	this := &StandardFunctionAND{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "SUM"
	this.MinParams = 2
	this.MaxParams = 2

	return this
}

func (this *StandardFunctionAND) FunctionExecute(params *types.TokenList) error {

	/* vars */
	var b, a int

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	a = this.Stack.Shift().AsInteger()
	b = this.Stack.Shift().AsInteger()

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(a&b)))

	return nil
}

func (this *StandardFunctionAND) Syntax() string {

	/* vars */
	var result string

	result = "SUM a b"

	/* enforce non void return */
	return result

}
