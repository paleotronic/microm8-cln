package logo

import (
	"strings"
	//	"math"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	//	"paleotronic.com/utils"
)

type StandardFunctionLOWERCASE struct {
	dialect.CoreFunction
}

func (this *StandardFunctionLOWERCASE) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewStandardFunctionLOWERCASE(a int, b int, params types.TokenList) *StandardFunctionLOWERCASE {
	this := &StandardFunctionLOWERCASE{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "LOWERCASE"
	this.MinParams = 1
	this.MaxParams = 1

	return this
}

func (this *StandardFunctionLOWERCASE) FunctionExecute(params *types.TokenList) error {

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

	this.Stack.Push(types.NewToken(types.STRING, strings.ToLower(thing.Content)))

	//this.Stack.Push(types.NewToken(types.NUMBER, utils.FormatFloat("", a+b)))

	return nil
}

func (this *StandardFunctionLOWERCASE) Syntax() string {

	/* vars */
	var result string

	result = "LOWERCASE word | list"

	/* enforce non void return */
	return result

}
