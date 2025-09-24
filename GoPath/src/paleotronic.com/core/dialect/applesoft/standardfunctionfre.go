package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type StandardFunctionFRE struct {
	dialect.CoreFunction
}

func NewStandardFunctionFRE(a int, b int, params types.TokenList) *StandardFunctionFRE {
	this := &StandardFunctionFRE{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "FRE"

	return this
}

func (this *StandardFunctionFRE) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

    var v int

    mode := this.Stack.Shift().AsInteger()

    if mode == 0 {
	   v = this.Interpreter.GetLocal().CleanStrings()
    } else {
       v = this.Interpreter.GetLocal().GetFree()
    }

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(v)))

	return nil
}

func (this *StandardFunctionFRE) Syntax() string {

	/* vars */
	var result string

	result = "FRE(<number>)"

	/* enforce non void return */
	return result

}

func (this *StandardFunctionFRE) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}
