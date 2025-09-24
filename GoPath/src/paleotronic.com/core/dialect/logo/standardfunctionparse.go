package logo

import (
	"strings"
	//	"math"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	//	"paleotronic.com/utils"
)

type StandardFunctionPARSE struct {
	dialect.CoreFunction
}

func (this *StandardFunctionPARSE) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	result = append(result, types.NUMBER)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewStandardFunctionPARSE(a int, b int, params types.TokenList) *StandardFunctionPARSE {
	this := &StandardFunctionPARSE{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "PARSE"
	this.MinParams = 1
	this.MaxParams = 1

	return this
}

func (this *StandardFunctionPARSE) FunctionExecute(params *types.TokenList) error {

	/* vars */
	//	var b, a float64

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	thing := this.Stack.Shift()

	if thing.Type == types.LIST {
		this.Stack.Push(thing)
		return nil
	}

	parts := strings.Split(thing.Content, " ")
	out := types.NewToken(types.LIST, "")
	out.List = types.NewTokenList()
	for _, word := range parts {
		out.List.Push(types.NewToken(types.STRING, word))
	}

	this.Stack.Push(out)

	return nil
}

func (this *StandardFunctionPARSE) Syntax() string {

	/* vars */
	var result string

	result = "MEMBER index word|list"

	/* enforce non void return */
	return result

}
