package logo

import (
	//	"math"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type StandardFunctionEQUALP struct {
	dialect.CoreFunction
}

func (this *StandardFunctionEQUALP) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewStandardFunctionEQUALP(a int, b int, params types.TokenList) *StandardFunctionEQUALP {
	this := &StandardFunctionEQUALP{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "EQUALP"
	this.MinParams = 2
	this.MaxParams = 2

	return this
}

func (this *StandardFunctionEQUALP) FunctionExecute(params *types.TokenList) error {

	/* vars */
	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	a := this.Stack.Shift()
	b := this.Stack.Shift()

	t := (b.Content == a.Content)

	if t {
		this.Stack.Push(types.NewToken(types.NUMBER, utils.FormatFloat("", 1)))
	} else {
		this.Stack.Push(types.NewToken(types.NUMBER, utils.FormatFloat("", 0)))
	}

	return nil
}

func (this *StandardFunctionEQUALP) Syntax() string {

	/* vars */
	var result string

	result = "EQUALP a b"

	/* enforce non void return */
	return result

}
