package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type StandardFunctionLEFTDollar struct {
	dialect.CoreFunction
}

func NewStandardFunctionLEFTDollar(a int, b int, params types.TokenList) *StandardFunctionLEFTDollar {
	this := &StandardFunctionLEFTDollar{}

	/* vars */
	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "LEFT$"
    
	this.MinParams = 2
	this.MaxParams = 2

	return this
}

func (this *StandardFunctionLEFTDollar) FunctionExecute(params *types.TokenList) error {

	/* vars */
	var s string
	var n int

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	n = this.Stack.Pop().AsInteger()
	s = this.Stack.Pop().Content

	this.Stack.Push(types.NewToken(types.STRING, utils.Copy(s, 1, n)))

	return nil
}

func (this *StandardFunctionLEFTDollar) Syntax() string {

	/* vars */
	var result string

	result = "LEFT$(<string>,<number>)"

	/* enforce non void return */
	return result

}

func (this *StandardFunctionLEFTDollar) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	result = append(result, types.STRING)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}
