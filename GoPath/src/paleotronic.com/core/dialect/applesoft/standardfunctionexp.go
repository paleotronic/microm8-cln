package applesoft

import (
	"math"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type StandardFunctionEXP struct {
	dialect.CoreFunction
}

func NewStandardFunctionEXP(a int, b int, params types.TokenList) *StandardFunctionEXP {
	this := &StandardFunctionEXP{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "EXP"

	return this
}

func (this *StandardFunctionEXP) FunctionExecute(params *types.TokenList) error {

	/* vars */
	var value float64

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	value = this.Stack.Pop().AsExtended()

	this.Stack.Push(types.NewToken(types.NUMBER, utils.FormatFloat("", math.Exp(value))))

	return nil
}

func (this *StandardFunctionEXP) Syntax() string {

	/* vars */
	var result string

	result = "EXP(<number>)"

	/* enforce non void return */
	return result

}

func (this *StandardFunctionEXP) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	//SetLength( result, 1 );
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}
