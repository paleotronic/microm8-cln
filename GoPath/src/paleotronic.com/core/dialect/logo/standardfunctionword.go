package logo

import (
	//	"strings"
	//	"math"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	//	"paleotronic.com/utils"
)

type StandardFunctionWORD struct {
	dialect.CoreFunction
}

func (this *StandardFunctionWORD) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	result = append(result, types.NUMBER)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewStandardFunctionWORD(a int, b int, params types.TokenList) *StandardFunctionWORD {
	this := &StandardFunctionWORD{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "WORD"
	this.MinParams = 2
	this.MaxParams = 2

	return this
}

func (this *StandardFunctionWORD) FunctionExecute(params *types.TokenList) error {

	/* vars */
	//	var b, a float64

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	var thing *types.Token

	thing = types.NewToken(types.WORD, "")

	for _, t := range params.Content {
		if t.Type == types.LIST {
			for _, vv := range t.List.Content {
				thing.Content += vv.Content
			}
		} else {
			thing.Content += t.Content
		}

	}

	this.Stack.Push(thing)

	return nil
}

func (this *StandardFunctionWORD) Syntax() string {

	/* vars */
	var result string

	result = "WORD word list"

	/* enforce non void return */
	return result

}
