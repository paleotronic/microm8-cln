package logo

import (
	//	"math"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	//	"paleotronic.com/utils"
)

type StandardFunctionBUTFIRST struct {
	dialect.CoreFunction
}

func (this *StandardFunctionBUTFIRST) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewStandardFunctionBUTFIRST(a int, b int, params types.TokenList) *StandardFunctionBUTFIRST {
	this := &StandardFunctionBUTFIRST{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "BUTFIRST"
	this.MinParams = 1
	this.MaxParams = 1

	return this
}

func (this *StandardFunctionBUTFIRST) FunctionExecute(params *types.TokenList) error {

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
		this.Stack.Push(thing.SubListAsToken(1, thing.List.Size()))
	} else {
		n := len(thing.Content)
		this.Stack.Push(types.NewToken(types.STRING, thing.Content[1:n]))
	}

	//this.Stack.Push(types.NewToken(types.NUMBER, utils.FormatFloat("", a+b)))

	return nil
}

func (this *StandardFunctionBUTFIRST) Syntax() string {

	/* vars */
	var result string

	result = "BUTFIRST word | list"

	/* enforce non void return */
	return result

}
