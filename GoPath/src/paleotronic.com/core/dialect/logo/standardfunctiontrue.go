package logo

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
)

type StandardFunctionTRUE struct {
	dialect.CoreFunction
}

func (this *StandardFunctionTRUE) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewStandardFunctionTRUE(a int, b int, params types.TokenList) *StandardFunctionTRUE {
	this := &StandardFunctionTRUE{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "COS"
	this.MaxParams = 0
	this.MinParams = 0

	return this
}

func (this *StandardFunctionTRUE) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	this.Stack.Push(types.NewToken(types.BOOLEAN, "1"))

	return nil
}

func (this *StandardFunctionTRUE) Syntax() string {

	/* vars */
	var result string

	result = "COS(<number>)"

	/* enforce non void return */
	return result

}
