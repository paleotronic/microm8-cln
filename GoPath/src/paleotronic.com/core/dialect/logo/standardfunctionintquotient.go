package logo

import (
	//	"math"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type StandardFunctionINTQUOTIENT struct {
	dialect.CoreFunction
}

func (this *StandardFunctionINTQUOTIENT) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewStandardFunctionINTQUOTIENT(a int, b int, params types.TokenList) *StandardFunctionINTQUOTIENT {
	this := &StandardFunctionINTQUOTIENT{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "INTQUOTIENT"
	this.MinParams = 2
	this.MaxParams = 2

	return this
}

func (this *StandardFunctionINTQUOTIENT) FunctionExecute(params *types.TokenList) error {

	/* vars */
	var b, a int

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	a = this.Stack.Shift().AsInteger()
	b = this.Stack.Shift().AsInteger()

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(a/b)))

	return nil
}

func (this *StandardFunctionINTQUOTIENT) Syntax() string {

	/* vars */
	var result string

	result = "INTQUOTIENT a b"

	/* enforce non void return */
	return result

}
