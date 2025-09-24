package applesoft

import (
	"paleotronic.com/core/dialect"
	//"paleotronic.com/core/interfaces"
	"math"

	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type StandardFunctionABS struct {
	dialect.CoreFunction
}

func NewStandardFunctionABS(a int, b int, params types.TokenList) *StandardFunctionABS {
	this := &StandardFunctionABS{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "ABS"
	this.MaxParams = 1

	return this
}

func (this *StandardFunctionABS) FunctionExecute(params *types.TokenList) error {

	/* vars */
	var value float64

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	value = this.Stack.Pop().AsExtended()

	this.Stack.Push(types.NewToken(types.NUMBER, utils.FloatToStr(math.Abs(value))))

	return nil
}

func (this *StandardFunctionABS) Syntax() string {

	/* vars */
	var result string

	result = "ABS(<number>)"

	/* enforce non void return */
	return result

}

func (this *StandardFunctionABS) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}
