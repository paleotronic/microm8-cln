package applesoft

import (
	"paleotronic.com/core/dialect"
	//"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardFunctionCHRDollar struct {
	dialect.CoreFunction
}

func NewStandardFunctionCHRDollar(a int, b int, params types.TokenList) *StandardFunctionCHRDollar {
	this := &StandardFunctionCHRDollar{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "CHR$"

	return this
}

func (this *StandardFunctionCHRDollar) FunctionExecute(params *types.TokenList) error {

	/* vars */
	var value int

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	value = this.Stack.Pop().AsInteger()

	this.Stack.Push(types.NewToken(types.STRING, string(value)))

	return nil
}

func (this *StandardFunctionCHRDollar) Syntax() string {

	/* vars */
	var result string

	result = "CHR$(<number>)"

	/* enforce non void return */
	return result

}

func (this *StandardFunctionCHRDollar) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}
