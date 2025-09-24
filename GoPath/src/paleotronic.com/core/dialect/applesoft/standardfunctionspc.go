package applesoft

import (
	"paleotronic.com/core/dialect"
	//"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardFunctionSPC struct {
	dialect.CoreFunction
}

func (this *StandardFunctionSPC) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	//SetLength( result, 1 );
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewStandardFunctionSPC(a int, b int, params types.TokenList) *StandardFunctionSPC {
	this := &StandardFunctionSPC{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "SPC"

	return this
}

func (this *StandardFunctionSPC) FunctionExecute(params *types.TokenList) error {

	/* vars */
	var value int
	var i int
	var s string

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	value = this.Stack.Pop().AsInteger()

	s = ""
	for i = 1; i <= value; i++ {
		s = s + " "
	}

	this.Stack.Push(types.NewToken(types.STRING, s))

	return nil
}

func (this *StandardFunctionSPC) Syntax() string {

	/* vars */
	var result string

	result = "SPC(<number>)"

	/* enforce non void return */
	return result

}
