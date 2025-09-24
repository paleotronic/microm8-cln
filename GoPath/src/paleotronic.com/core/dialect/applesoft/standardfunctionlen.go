package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type StandardFunctionLEN struct {
	dialect.CoreFunction
}

func (this *StandardFunctionLEN) Syntax() string {

	/* vars */
	var result string

	result = "LEN(<string>)"

	/* enforce non void return */
	return result

}

func (this *StandardFunctionLEN) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewStandardFunctionLEN(a int, b int, params types.TokenList) *StandardFunctionLEN {
	this := &StandardFunctionLEN{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "LEN"

	return this
}

func (this *StandardFunctionLEN) FunctionExecute(params *types.TokenList) error {

	/* vars */
	var value string

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	value = this.Stack.Pop().Content

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(utils.Len(value))))

	return nil
}
