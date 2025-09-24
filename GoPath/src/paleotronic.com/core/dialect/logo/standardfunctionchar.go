package logo

import (
	//	"strings"
	//	"math"
	"errors"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
//	"paleotronic.com/utils"
)

type StandardFunctionCHAR struct {
	dialect.CoreFunction
}

func (this *StandardFunctionCHAR) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewStandardFunctionCHAR(a int, b int, params types.TokenList) *StandardFunctionCHAR {
	this := &StandardFunctionCHAR{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "CHAR"
	this.MinParams = 1
	this.MaxParams = 1

	return this
}

func (this *StandardFunctionCHAR) FunctionExecute(params *types.TokenList) error {

	/* vars */
	//	var b, a float64

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	o := params.Shift()
	if o == nil {
		return errors.New("I NEED AN OBJECT")
	}

	if o.IsNumeric() {
		this.Stack.Push(types.NewToken(types.WORD, string(rune(o.AsInteger()))))
	} else {
		this.Stack.Push(types.NewToken(types.WORD, " "))
	}

	return nil
}

func (this *StandardFunctionCHAR) Syntax() string {

	/* vars */
	var result string

	result = "CHAR object"

	/* enforce non void return */
	return result

}
