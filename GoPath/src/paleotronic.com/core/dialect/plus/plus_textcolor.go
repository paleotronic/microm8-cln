package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusTextColor struct {
	dialect.CoreFunction
}

func (this *PlusTextColor) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	if !this.Query {
		fg := params.Shift().AsInteger()
		bg := params.Shift().AsInteger()
		if fg != -1 {
			apple2helpers.SetFGColor(this.Interpreter, uint64(fg))
		}
		if bg != -1 {
			apple2helpers.SetBGColor(this.Interpreter, uint64(bg))
		}
	}
	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(0)))

	return nil
}

func (this *PlusTextColor) Syntax() string {

	/* vars */
	var result string

	result = "BGCOLOR{col}"

	/* enforce non void return */
	return result

}

func (this *PlusTextColor) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusTextColor(a int, b int, params types.TokenList) *PlusTextColor {
	this := &PlusTextColor{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "BGCOLOR"

	this.NamedParams = []string{"fg", "bg"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.INTEGER, "-1"),
		*types.NewToken(types.INTEGER, "-1"),
	}
	this.Raw = true

	return this
}
