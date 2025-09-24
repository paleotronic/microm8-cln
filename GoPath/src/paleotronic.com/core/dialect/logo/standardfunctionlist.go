package logo

import (
	//	"strings"
	//	"math"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	//	"paleotronic.com/utils"
)

type StandardFunctionLIST struct {
	dialect.CoreFunction
}

func (this *StandardFunctionLIST) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	result = append(result, types.NUMBER)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewStandardFunctionLIST(a int, b int, params types.TokenList) *StandardFunctionLIST {
	this := &StandardFunctionLIST{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "LIST"
	this.MinParams = 2
	this.MaxParams = 2

	return this
}

func (this *StandardFunctionLIST) FunctionExecute(params *types.TokenList) error {

	/* vars */
	//	var b, a float64

	//log.Printf("params: %+v", params.AsString())

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	var thing *types.Token

	thing = types.NewToken(types.LIST, "")
	thing.List = types.NewTokenList()

	for _, t := range params.Content {
		if t.Type == types.LIST {
			thing.List.Push(t)
		} else {
			t = types.NewToken(types.WORD, t.Content)
			thing.List.Push(t)
		}
		if thing.List.RPeek().WSSuffix == "" {
			thing.List.RPeek().WSSuffix = " "
		}
	}

	this.Stack.Push(thing)

	return nil
}

func (this *StandardFunctionLIST) Syntax() string {

	/* vars */
	var result string

	result = "LIST word list"

	/* enforce non void return */
	return result

}
