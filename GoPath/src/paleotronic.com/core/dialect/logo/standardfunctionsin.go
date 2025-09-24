package logo

import (
	"math"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type StandardFunctionSIN struct {
	dialect.CoreFunction
}

func NewStandardFunctionSIN(a int, b int, params types.TokenList) *StandardFunctionSIN {
	this := &StandardFunctionSIN{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "SIN"
	this.MaxParams = 1

	return this
}

func (this *StandardFunctionSIN) FunctionExecute(params *types.TokenList) error {

	/* vars */
	var value float64

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	value = this.Stack.Pop().AsExtended() * 0.0174533

	this.Stack.Push(types.NewToken(types.NUMBER, utils.FormatFloat("", math.Sin(value))))

	return nil
}

func (this *StandardFunctionSIN) Syntax() string {

	/* vars */
	var result string

	result = "SIN(<number>)"

	/* enforce non void return */
	return result

}

func (this *StandardFunctionSIN) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	//SetLength( result, 1 );
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}
