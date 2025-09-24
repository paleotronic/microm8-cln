package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusMouseSelect struct {
	dialect.CoreFunction
}

func (this *PlusMouseSelect) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	t := params.Get(0).AsInteger()

	settings.DisableTextSelect[this.Interpreter.GetMemIndex()] = (t == 0)

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(t)))

	return nil
}

func (this *PlusMouseSelect) Syntax() string {

	/* vars */
	var result string

	result = "Select{index}"

	/* enforce non void return */
	return result

}

func (this *PlusMouseSelect) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusMouseSelect(a int, b int, params types.TokenList) *PlusMouseSelect {
	this := &PlusMouseSelect{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "Select"
	this.MinParams = 1
	this.MaxParams = 1

	return this
}
