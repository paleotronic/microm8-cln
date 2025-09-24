package logo

import (
	//	"math"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type StandardFunctionSUM struct {
	dialect.CoreFunction
}

func (this *StandardFunctionSUM) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewStandardFunctionSUM(a int, b int, params types.TokenList) *StandardFunctionSUM {
	this := &StandardFunctionSUM{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "SUM"
	this.MinParams = 2
	this.MaxParams = 2

	return this
}

func (this *StandardFunctionSUM) FunctionExecute(params *types.TokenList) error {

	/* vars */
	var b, a float64

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	b = this.Stack.Pop().AsExtended()
	a = this.Stack.Pop().AsExtended()

	this.Stack.Push(types.NewToken(types.NUMBER, utils.FormatFloat("", a+b)))

	return nil
}

func (this *StandardFunctionSUM) Syntax() string {

	/* vars */
	var result string

	result = "SUM a b"

	/* enforce non void return */
	return result

}
