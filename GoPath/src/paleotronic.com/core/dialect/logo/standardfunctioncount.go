package logo

import (
	//	"strings"
	//	"math"
	"errors"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type StandardFunctionCOUNT struct {
	dialect.CoreFunction
}

func (this *StandardFunctionCOUNT) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewStandardFunctionCOUNT(a int, b int, params types.TokenList) *StandardFunctionCOUNT {
	this := &StandardFunctionCOUNT{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "COUNT"
	this.MinParams = 1
	this.MaxParams = 1

	return this
}

func (this *StandardFunctionCOUNT) FunctionExecute(params *types.TokenList) error {

	/* vars */
	//	var b, a float64

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	o := params.Shift()
	if o == nil {
		return errors.New("I NEED AN OBJECT")
	}

	if o.Type == types.LIST {
		this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(o.List.Size())))
	} else {
		this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(len(o.Content))))
	}

	return nil
}

func (this *StandardFunctionCOUNT) Syntax() string {

	/* vars */
	var result string

	result = "COUNT object"

	/* enforce non void return */
	return result

}
