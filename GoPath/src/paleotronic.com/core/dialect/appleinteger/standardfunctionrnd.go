package appleinteger

import (
	"paleotronic.com/core/dialect"
	//	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type StandardFunctionRND struct {
	dialect.CoreFunction
}

func NewStandardFunctionRND(a int, b int, params types.TokenList) *StandardFunctionRND {
	this := &StandardFunctionRND{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "RND"
	this.MinParams = 1
	this.MaxParams = 1

	return this
}

func (this *StandardFunctionRND) FunctionExecute(params *types.TokenList) error {

	/* vars */
	var value float64

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	value = this.Stack.Pop().AsExtended()

	if value == 0 {
		this.Stack.Push(types.NewToken(types.NUMBER, utils.FormatFloat("", utils.Random())))
	} else {
		this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr((int)(utils.Random()*value))))
	}

	return nil

}

func (this *StandardFunctionRND) Syntax() string {

	/* vars */
	var result string

	result = "RND(<number>)"

	/* enforce non void return */
	return result

}

func (this *StandardFunctionRND) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	//SetLength( result, 1 );
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}
