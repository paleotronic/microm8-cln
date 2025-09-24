package logo

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
)

type StandardFunctionROUTINE struct {
	dialect.CoreFunction
	D *DialectLogo
}

func NewStandardFunctionROUTINE(a int, b int, params types.TokenList, d *DialectLogo) *StandardFunctionROUTINE {
	this := &StandardFunctionROUTINE{D: d}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "ROUTINE"
	this.MaxParams = 1

	return this
}

func (this *StandardFunctionROUTINE) FunctionExecute(params *types.TokenList) error {

	/* vars */
	var value string

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	value = this.Stack.Shift().Content

	c, ok := this.D.Driver.GetCoroutine(value)
	if ok && c.IsRunning() {
		this.Stack.Push(types.NewToken(types.NUMBER, "1"))
	} else {
		this.Stack.Push(types.NewToken(types.NUMBER, "0"))
	}

	return nil
}

func (this *StandardFunctionROUTINE) Syntax() string {

	/* vars */
	var result string

	result = "SIN(<number>)"

	/* enforce non void return */
	return result

}

func (this *StandardFunctionROUTINE) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	//SetLength( result, 1 );
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}
