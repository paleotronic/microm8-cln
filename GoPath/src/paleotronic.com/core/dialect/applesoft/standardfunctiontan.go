package applesoft

import (
	"math"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type StandardFunctionTAN struct {
	dialect.CoreFunction
}

func NewStandardFunctionTAN(a int, b int, params types.TokenList) *StandardFunctionTAN {
	this := &StandardFunctionTAN{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "TAN"

	return this
}

func (this *StandardFunctionTAN) FunctionExecute(params *types.TokenList) error {

	/* vars */
	var value float64

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	value = this.Stack.Pop().AsExtended()

	this.Stack.Push(types.NewToken(types.NUMBER, utils.FormatFloat("", math.Tan(value))))

	return nil
}

func (this *StandardFunctionTAN) Syntax() string {

	/* vars */
	var result string

	result = "TAN(<number>)"

	/* enforce non void return */
	return result

}

func (this *StandardFunctionTAN) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	//SetLength( result, 1 );
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}
