package logo

import (
	//	"math"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	//	"paleotronic.com/utils"
)

type StandardFunctionITEM struct {
	dialect.CoreFunction
}

func (this *StandardFunctionITEM) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	result = append(result, types.NUMBER)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewStandardFunctionITEM(a int, b int, params types.TokenList) *StandardFunctionITEM {
	this := &StandardFunctionITEM{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "ITEM"
	this.MinParams = 2
	this.MaxParams = 2

	return this
}

func (this *StandardFunctionITEM) FunctionExecute(params *types.TokenList) error {

	/* vars */
	//	var b, a float64

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	thing := this.Stack.Pop()
	index := this.Stack.Pop().AsInteger()

	if index < 1 || (thing.List != nil && index > thing.List.Size()) || (thing.List == nil && index > len(thing.Content)) {
		this.Stack.Push(types.NewToken(types.STRING, ""))
		return nil
	}

	if thing.Type == types.LIST {
		this.Stack.Push(thing.List.Get(index - 1))
	} else {
		//		n := len(thing.Content)
		this.Stack.Push(types.NewToken(types.STRING, thing.Content[index-1:index]))
	}

	//this.Stack.Push(types.NewToken(types.NUMBER, utils.FormatFloat("", a+b)))

	return nil
}

func (this *StandardFunctionITEM) Syntax() string {

	/* vars */
	var result string

	result = "ITEM index word|list"

	/* enforce non void return */
	return result

}
