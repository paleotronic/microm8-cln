package logo

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
)

type StandardFunctionLASTREM struct {
	dialect.CoreFunction
	D *DialectLogo
}

func NewStandardFunctionLASTREM(a int, b int, params types.TokenList, d *DialectLogo) *StandardFunctionLASTREM {
	this := &StandardFunctionLASTREM{D: d}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "ROUTINE"
	this.MaxParams = 0
	this.MinParams = 0

	return this
}

func (this *StandardFunctionLASTREM) FunctionExecute(params *types.TokenList) error {

	/* vars */
	var value string

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	value = this.D.Driver.LastRem()
	this.Stack.Push(types.NewToken(types.STRING, value))

	return nil
}

func (this *StandardFunctionLASTREM) Syntax() string {

	/* vars */
	var result string

	result = "SIN(<number>)"

	/* enforce non void return */
	return result

}

func (this *StandardFunctionLASTREM) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	//SetLength( result, 1 );
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}
