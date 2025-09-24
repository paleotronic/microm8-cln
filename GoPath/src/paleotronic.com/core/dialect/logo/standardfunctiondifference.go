package logo

import (
	//	"math"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type StandardFunctionDIFFERENCE struct {
	dialect.CoreFunction
}

func (this *StandardFunctionDIFFERENCE) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewStandardFunctionDIFFERENCE(a int, b int, params types.TokenList) *StandardFunctionDIFFERENCE {
	this := &StandardFunctionDIFFERENCE{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "DIFFERENCE"
	this.MinParams = 2
	this.MaxParams = 2

	return this
}

func (this *StandardFunctionDIFFERENCE) FunctionExecute(params *types.TokenList) error {

	/* vars */
	var b, a float64

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	a = this.Stack.Shift().AsExtended()
	b = this.Stack.Shift().AsExtended()

	this.Stack.Push(types.NewToken(types.NUMBER, utils.FormatFloat("", a-b)))

	return nil
}

func (this *StandardFunctionDIFFERENCE) Syntax() string {

	/* vars */
	var result string

	result = "DIFFERENCE a b"

	/* enforce non void return */
	return result

}
