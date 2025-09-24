package logo

import (
	//	"math"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	//	"paleotronic.com/utils"
)

type StandardFunctionFIRST struct {
	dialect.CoreFunction
}

func (this *StandardFunctionFIRST) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewStandardFunctionFIRST(a int, b int, params types.TokenList) *StandardFunctionFIRST {
	this := &StandardFunctionFIRST{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "FIRST"
	this.MinParams = 1
	this.MaxParams = 1

	return this
}

func (this *StandardFunctionFIRST) FunctionExecute(params *types.TokenList) error {

	/* vars */
	//	var b, a float64

	//log.Printf("params size is %d", params.Size())

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	thing := this.Stack.Shift()

	if thing.Type == types.LIST {
		this.Stack.Push(thing.List.Get(0))
	} else {
		this.Stack.Push(types.NewToken(types.STRING, thing.Content[:1]))
	}

	//this.Stack.Push(types.NewToken(types.NUMBER, utils.FormatFloat("", a+b)))

	return nil
}

func (this *StandardFunctionFIRST) Syntax() string {

	/* vars */
	var result string

	result = "FIRST word | list"

	/* enforce non void return */
	return result

}
