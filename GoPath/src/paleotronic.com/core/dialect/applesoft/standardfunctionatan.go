package applesoft

import (
	"math"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type StandardFunctionATAN struct {
	dialect.CoreFunction
}

func NewStandardFunctionATAN(a int, b int, params types.TokenList) *StandardFunctionATAN {
	this := &StandardFunctionATAN{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "ATAN"

	return this
}

func (this *StandardFunctionATAN) FunctionExecute(params *types.TokenList) error {

	/* vars */
	var value float64
	var z float64

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	value = this.Stack.Pop().AsExtended()

	z = math.Atan(value)
	//writeln( "====> ARCTAN (",value,") is ", z);

	this.Stack.Push(types.NewToken(types.NUMBER, utils.FloatToStr(z)))

	return nil
}

func (this *StandardFunctionATAN) Syntax() string {

	/* vars */
	var result string

	result = "ATAN(<number>)"

	/* enforce non void return */
	return result

}

func (this *StandardFunctionATAN) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}
