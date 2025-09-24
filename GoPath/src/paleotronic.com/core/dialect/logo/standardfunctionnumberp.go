package logo

import (
	//	"strings"
	//	"math"
	"errors"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	//	"paleotronic.com/utils"
)

type StandardFunctionNUMBERP struct {
	dialect.CoreFunction
}

func (this *StandardFunctionNUMBERP) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewStandardFunctionNUMBERP(a int, b int, params types.TokenList) *StandardFunctionNUMBERP {
	this := &StandardFunctionNUMBERP{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "NUMBERP"
	this.MinParams = 1
	this.MaxParams = 1

	return this
}

func (this *StandardFunctionNUMBERP) FunctionExecute(params *types.TokenList) error {

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
		this.Stack.Push(types.NewToken(types.NUMBER, "0"))
	} else {
		if o.IsNumeric() || o.AsNumeric() != 0 {
			this.Stack.Push(types.NewToken(types.NUMBER, "1"))
		} else {
			this.Stack.Push(types.NewToken(types.NUMBER, "0"))
		}
	}

	return nil
}

func (this *StandardFunctionNUMBERP) Syntax() string {

	/* vars */
	var result string

	result = "NUMBERP object"

	/* enforce non void return */
	return result

}
