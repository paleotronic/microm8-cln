package logo

import (
	//	"math"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type StandardFunctionOR struct {
	dialect.CoreFunction
}

func (this *StandardFunctionOR) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewStandardFunctionOR(a int, b int, params types.TokenList) *StandardFunctionOR {
	this := &StandardFunctionOR{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "SUM"
	this.MinParams = 2
	this.MaxParams = 2

	return this
}

func (this *StandardFunctionOR) FunctionExecute(params *types.TokenList) error {

	/* vars */
	var b, a int

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	b = this.Stack.Pop().AsInteger()
	a = this.Stack.Pop().AsInteger()

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(a|b)))

	return nil
}

func (this *StandardFunctionOR) Syntax() string {

	/* vars */
	var result string

	result = "SUM a b"

	/* enforce non void return */
	return result

}
