package logo

import (
	//	"strings"
	//	"math"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	//	"paleotronic.com/utils"
)

type StandardFunctionLPUT struct {
	dialect.CoreFunction
}

func (this *StandardFunctionLPUT) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	result = append(result, types.NUMBER)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewStandardFunctionLPUT(a int, b int, params types.TokenList) *StandardFunctionLPUT {
	this := &StandardFunctionLPUT{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "LPUT"
	this.MinParams = 2
	this.MaxParams = 2

	return this
}

func (this *StandardFunctionLPUT) FunctionExecute(params *types.TokenList) error {

	/* vars */
	//	var b, a float64

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	thing := this.Stack.Pop()
	subthing := this.Stack.Pop()

	if thing == nil || subthing == nil {
		this.Stack.Push(types.NewToken(types.STRING, ""))
		return nil
	}

	if thing.Type != types.LIST {
		thing = types.NewToken(types.LIST, "")
		thing.List = types.NewTokenList()
	}

	if thing.Type == types.LIST {
		thing.List.Push(subthing.Copy())
	}

	this.Stack.Push(thing)

	return nil
}

func (this *StandardFunctionLPUT) Syntax() string {

	/* vars */
	var result string

	result = "LPUT word list"

	/* enforce non void return */
	return result

}
