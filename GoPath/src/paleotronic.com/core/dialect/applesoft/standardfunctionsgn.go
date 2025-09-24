package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type StandardFunctionSGN struct {
	dialect.CoreFunction
}

func (this *StandardFunctionSGN) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	//SetLength( result, 1 );
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewStandardFunctionSGN(a int, b int, params types.TokenList) *StandardFunctionSGN {
	this := &StandardFunctionSGN{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "SGN"

	return this
}

func (this *StandardFunctionSGN) FunctionExecute(params *types.TokenList) error {

	/* vars */
	var value float64
	var r int

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	value = this.Stack.Pop().AsExtended()

	r = 0
	if value > 0 {
		r = 1
	} else if value < 0 {
		r = -1
	}

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(r)))

	return nil
}

func (this *StandardFunctionSGN) Syntax() string {

	/* vars */
	var result string

	result = "SGN(<number>)"

	/* enforce non void return */
	return result

}
