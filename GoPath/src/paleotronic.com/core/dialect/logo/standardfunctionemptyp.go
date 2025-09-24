package logo

import (
	//	"strings"
	//	"math"
	"errors"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	//	"paleotronic.com/utils"
)

type StandardFunctionEMPTYP struct {
	dialect.CoreFunction
}

func (this *StandardFunctionEMPTYP) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewStandardFunctionEMPTYP(a int, b int, params types.TokenList) *StandardFunctionEMPTYP {
	this := &StandardFunctionEMPTYP{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "EMPTYP"
	this.MinParams = 1
	this.MaxParams = 1

	return this
}

func (this *StandardFunctionEMPTYP) FunctionExecute(params *types.TokenList) error {

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
		if o.List.Size() == 0 {
			this.Stack.Push(types.NewToken(types.NUMBER, "1"))
		} else {
			this.Stack.Push(types.NewToken(types.NUMBER, "0"))
		}
	} else {
		if len(o.Content) == 0 {
			this.Stack.Push(types.NewToken(types.NUMBER, "1"))
		} else {
			this.Stack.Push(types.NewToken(types.NUMBER, "0"))
		}
	}

	return nil
}

func (this *StandardFunctionEMPTYP) Syntax() string {

	/* vars */
	var result string

	result = "EMPTYP object"

	/* enforce non void return */
	return result

}
