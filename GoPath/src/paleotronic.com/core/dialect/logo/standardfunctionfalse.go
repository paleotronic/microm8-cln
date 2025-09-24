package logo

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
)

type StandardFunctionFALSE struct {
	dialect.CoreFunction
}

func (this *StandardFunctionFALSE) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewStandardFunctionFALSE(a int, b int, params types.TokenList) *StandardFunctionFALSE {
	this := &StandardFunctionFALSE{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "COS"
	this.MaxParams = 0
	this.MinParams = 0

	return this
}

func (this *StandardFunctionFALSE) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	this.Stack.Push(types.NewToken(types.BOOLEAN, "0"))

	return nil
}

func (this *StandardFunctionFALSE) Syntax() string {

	/* vars */
	var result string

	result = "COS(<number>)"

	/* enforce non void return */
	return result

}
