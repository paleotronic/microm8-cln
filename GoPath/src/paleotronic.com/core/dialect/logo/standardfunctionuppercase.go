package logo

import (
	"strings"
	//	"math"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	//	"paleotronic.com/utils"
)

type StandardFunctionUPPERCASE struct {
	dialect.CoreFunction
}

func (this *StandardFunctionUPPERCASE) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewStandardFunctionUPPERCASE(a int, b int, params types.TokenList) *StandardFunctionUPPERCASE {
	this := &StandardFunctionUPPERCASE{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "UPPERCASE"
	this.MinParams = 1
	this.MaxParams = 1

	return this
}

func (this *StandardFunctionUPPERCASE) FunctionExecute(params *types.TokenList) error {

	/* vars */
	//	var b, a float64

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	thing := this.Stack.Pop()

	if thing == nil {
		this.Stack.Push(types.NewToken(types.STRING, ""))
		return nil
	}

	this.Stack.Push(types.NewToken(types.STRING, strings.ToUpper(thing.Content)))

	//this.Stack.Push(types.NewToken(types.NUMBER, utils.FormatFloat("", a+b)))

	return nil
}

func (this *StandardFunctionUPPERCASE) Syntax() string {

	/* vars */
	var result string

	result = "UPPERCASE word | list"

	/* enforce non void return */
	return result

}
