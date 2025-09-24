package applesoft

import (
	"math"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type StandardFunctionLOG struct {
	dialect.CoreFunction
}

func NewStandardFunctionLOG(a int, b int, params types.TokenList) *StandardFunctionLOG {
	this := &StandardFunctionLOG{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "LOG"

	return this
}

func (this *StandardFunctionLOG) FunctionExecute(params *types.TokenList) error {

	/* vars */
	var value float64

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	value = this.Stack.Pop().AsExtended()

	this.Stack.Push(types.NewToken(types.NUMBER, utils.FormatFloat("", math.Log(value))))

	return nil
}

func (this *StandardFunctionLOG) Syntax() string {

	/* vars */
	var result string

	result = "LOG(<number>)"

	/* enforce non void return */
	return result

}

func (this *StandardFunctionLOG) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}
