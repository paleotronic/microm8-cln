package logo

import (
	//	"strings"
	//	"math"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	//	"paleotronic.com/utils"
)

type StandardFunctionFPUT struct {
	dialect.CoreFunction
}

func (this *StandardFunctionFPUT) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	result = append(result, types.NUMBER)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewStandardFunctionFPUT(a int, b int, params types.TokenList) *StandardFunctionFPUT {
	this := &StandardFunctionFPUT{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "FPUT"
	this.MinParams = 2
	this.MaxParams = 2

	return this
}

func (this *StandardFunctionFPUT) FunctionExecute(params *types.TokenList) error {

	/* vars */
	//	var b, a float64

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	subthing := this.Stack.Shift()
	thing := this.Stack.Shift()

	if thing == nil || subthing == nil {
		this.Stack.Push(types.NewToken(types.STRING, ""))
		return nil
	}

	if thing.Type != types.LIST {
		thing = types.NewToken(types.LIST, "")
		thing.List = types.NewTokenList()
	}

	if thing.Type == types.LIST {
		thing.List.UnShift(subthing.Copy())
	}

	this.Stack.Push(thing)

	return nil
}

func (this *StandardFunctionFPUT) Syntax() string {

	/* vars */
	var result string

	result = "FPUT word list"

	/* enforce non void return */
	return result

}
