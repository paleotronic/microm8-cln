package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type StandardFunctionRIGHTDollar struct {
	dialect.CoreFunction
}

func NewStandardFunctionRIGHTDollar(a int, b int, params types.TokenList) *StandardFunctionRIGHTDollar {
	this := &StandardFunctionRIGHTDollar{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "RIGHT$"
    
	this.MinParams = 2
	this.MaxParams = 2

	return this
}

func (this *StandardFunctionRIGHTDollar) FunctionExecute(params *types.TokenList) error {

	/* vars */
	var s string
	var n int

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	n = this.Stack.Pop().AsInteger()
	s = this.Stack.Pop().Content

	if n > len(s) {
		this.Stack.Push(types.NewToken(types.STRING, s))
	} else {
		this.Stack.Push(types.NewToken(types.STRING, utils.Copy(s, utils.Len(s)-n+1, n)))
	}

	return nil
}

func (this *StandardFunctionRIGHTDollar) Syntax() string {

	/* vars */
	var result string

	result = "RIGHT$(<string>,<number>)"

	/* enforce non void return */
	return result

}

func (this *StandardFunctionRIGHTDollar) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	//SetLength( result, 2 );
	result = append(result, types.STRING)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}
