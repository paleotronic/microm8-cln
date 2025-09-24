package plus

import (
	//	"strings"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusParamCount struct {
	dialect.CoreFunction
}

func (this *PlusParamCount) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	this.Stack.Push(types.NewToken(types.INTEGER, utils.IntToStr(this.Interpreter.GetParams().Size())))

	return nil
}

func (this *PlusParamCount) Syntax() string {

	/* vars */
	var result string

	result = "PARAMCOUNT{}"

	/* enforce non void return */
	return result

}

func (this *PlusParamCount) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewPlusParamCount(a int, b int, params types.TokenList) *PlusParamCount {
	this := &PlusParamCount{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "PARAMCOUNT"
	this.Raw = true
	this.MinParams = 0
	this.MaxParams = 0

	return this
}
