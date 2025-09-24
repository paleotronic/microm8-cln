package logo

import (
	//	"math"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	//	"paleotronic.com/utils"
)

type StandardFunctionBUTLAST struct {
	dialect.CoreFunction
}

func (this *StandardFunctionBUTLAST) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewStandardFunctionBUTLAST(a int, b int, params types.TokenList) *StandardFunctionBUTLAST {
	this := &StandardFunctionBUTLAST{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "BUTLAST"
	this.MinParams = 1
	this.MaxParams = 1

	return this
}

func (this *StandardFunctionBUTLAST) FunctionExecute(params *types.TokenList) error {

	/* vars */
	//	var b, a float64

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	thing := this.Stack.Shift()

	if thing == nil {
		this.Stack.Push(types.NewToken(types.STRING, ""))
		return nil
	}

	if thing.Type == types.LIST {
		this.Stack.Push(thing.SubListAsToken(0, thing.List.Size()-1))
	} else {
		n := len(thing.Content)
		this.Stack.Push(types.NewToken(types.STRING, thing.Content[0:n-1]))
	}

	//this.Stack.Push(types.NewToken(types.NUMBER, utils.FormatFloat("", a+b)))

	return nil
}

func (this *StandardFunctionBUTLAST) Syntax() string {

	/* vars */
	var result string

	result = "BUTLAST word | list"

	/* enforce non void return */
	return result

}
