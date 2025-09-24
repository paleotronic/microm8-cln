package logo

import (
	//	"strings"
	//	"math"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
	//	"paleotronic.com/utils"
)

type StandardFunctionSENTENCE struct {
	dialect.CoreFunction
}

func (this *StandardFunctionSENTENCE) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewStandardFunctionSENTENCE(a int, b int, params types.TokenList) *StandardFunctionSENTENCE {
	this := &StandardFunctionSENTENCE{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "SENTENCE"
	this.MinParams = 2
	this.MaxParams = 2

	return this
}

func (this *StandardFunctionSENTENCE) FunctionExecute(params *types.TokenList) error {

	/* vars */
	//	var b, a float64

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	var thing *types.Token

	thing = types.NewToken(types.LIST, "")
	thing.List = types.NewTokenList()

	for _, t := range params.Content {
		if t.Type == types.LIST {
			for _, tt := range t.List.Content {
				thing.List.Push(tt)
				if thing.List.RPeek().WSSuffix == "" {
					thing.List.RPeek().WSSuffix = " "
				}
			}
		} else {
			if t.Type == types.NUMBER {
				t = types.NewToken(types.WORD, utils.StrToFloatStrAppleLogo(t.Content))
			} else {
				t = types.NewToken(types.WORD, t.Content)
			}
			thing.List.Push(t)
			if thing.List.RPeek().WSSuffix == "" {
				thing.List.RPeek().WSSuffix = " "
			}
		}
	}

	this.Stack.Push(thing)

	return nil
}

func (this *StandardFunctionSENTENCE) Syntax() string {

	/* vars */
	var result string

	result = "SENTENCE word list"

	/* enforce non void return */
	return result

}
