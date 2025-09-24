package applesoft

import (
	"math"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type StandardFunctionSQR struct {
	dialect.CoreFunction
}

func NewStandardFunctionSQR(a int, b int, params types.TokenList) *StandardFunctionSQR {
	this := &StandardFunctionSQR{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "SQR"
	this.MaxParams = 1

	return this
}

func (this *StandardFunctionSQR) FunctionExecute(params *types.TokenList) error {

	/* vars */
	var value float64

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	value = this.Stack.Pop().AsExtended()

	this.Stack.Push(types.NewToken(types.NUMBER, utils.FormatFloat("", math.Sqrt(value))))

	return nil
}

func (this *StandardFunctionSQR) Syntax() string {

	/* vars */
	var result string

	result = "SQR(<number>)"

	/* enforce non void return */
	return result

}

func (this *StandardFunctionSQR) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	//SetLength( result, 1 );
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}
