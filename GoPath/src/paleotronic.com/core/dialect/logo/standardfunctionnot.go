package logo

import (
	//	"math"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
)

type StandardFunctionNOT struct {
	dialect.CoreFunction
}

func (this *StandardFunctionNOT) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewStandardFunctionNOT(a int, b int, params types.TokenList) *StandardFunctionNOT {
	this := &StandardFunctionNOT{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "SUM"
	this.MinParams = 1
	this.MaxParams = 1

	return this
}

func (this *StandardFunctionNOT) FunctionExecute(params *types.TokenList) error {

	/* vars */
	var a int

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	a = this.Stack.Pop().AsInteger()

	if a == 0 {
		this.Stack.Push(types.NewToken(types.NUMBER, "1"))
	} else {
		this.Stack.Push(types.NewToken(types.NUMBER, "0"))
	}

	return nil
}

func (this *StandardFunctionNOT) Syntax() string {

	/* vars */
	var result string

	result = "SUM a b"

	/* enforce non void return */
	return result

}
