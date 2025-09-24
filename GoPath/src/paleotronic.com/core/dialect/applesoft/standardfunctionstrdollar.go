package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type StandardFunctionSTRDollar struct {
	dialect.CoreFunction
}

func (this *StandardFunctionSTRDollar) FunctionExecute(params *types.TokenList) error {

	/* vars */
	var e string

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	e = this.Stack.Pop().Content

	this.Stack.Push(types.NewToken(types.STRING, utils.StrToFloatStrApple(e) ))

	return nil
}

func (this *StandardFunctionSTRDollar) Syntax() string {

	/* vars */
	var result string

	result = "STR$(<string>)"

	/* enforce non void return */
	return result

}

func (this *StandardFunctionSTRDollar) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	//SetLength( result, 1 );
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewStandardFunctionSTRDollar(a int, b int, params types.TokenList) *StandardFunctionSTRDollar {
	this := &StandardFunctionSTRDollar{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "STR$"

	return this
}
