package logo

import (
	"math"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type StandardFunctionCOS struct {
	dialect.CoreFunction
}

func (this *StandardFunctionCOS) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewStandardFunctionCOS(a int, b int, params types.TokenList) *StandardFunctionCOS {
	this := &StandardFunctionCOS{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "COS"
	this.MaxParams = 1

	return this
}

func (this *StandardFunctionCOS) FunctionExecute(params *types.TokenList) error {

	/* vars */
	var value float64

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	value = this.Stack.Shift().AsExtended() * 0.0174533

	this.Stack.Push(types.NewToken(types.NUMBER, utils.FormatFloat("", math.Cos(value))))

	return nil
}

func (this *StandardFunctionCOS) Syntax() string {

	/* vars */
	var result string

	result = "COS(<number>)"

	/* enforce non void return */
	return result

}
