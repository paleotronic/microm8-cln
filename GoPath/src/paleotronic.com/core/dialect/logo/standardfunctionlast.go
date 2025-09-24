package logo

import (
	//	"math"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	//	"paleotronic.com/utils"
)

type StandardFunctionLAST struct {
	dialect.CoreFunction
}

func (this *StandardFunctionLAST) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewStandardFunctionLAST(a int, b int, params types.TokenList) *StandardFunctionLAST {
	this := &StandardFunctionLAST{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "LAST"
	this.MinParams = 1
	this.MaxParams = 1

	return this
}

func (this *StandardFunctionLAST) FunctionExecute(params *types.TokenList) error {

	/* vars */
	//	var b, a float64

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	thing := this.Stack.Pop()

	if thing.Type == types.LIST {
		this.Stack.Push(thing.List.RPeek())
	} else {
		n := len(thing.Content)
		this.Stack.Push(types.NewToken(types.STRING, thing.Content[n-1:n]))
	}

	//this.Stack.Push(types.NewToken(types.NUMBER, utils.FormatFloat("", a+b)))

	return nil
}

func (this *StandardFunctionLAST) Syntax() string {

	/* vars */
	var result string

	result = "LAST word | list"

	/* enforce non void return */
	return result

}
