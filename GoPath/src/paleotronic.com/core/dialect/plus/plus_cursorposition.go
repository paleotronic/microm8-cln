package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusCursorPosition struct {
	dialect.CoreFunction
}

func (this *PlusCursorPosition) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	if !this.Query {
		x := params.Shift().AsInteger()
		y := params.Shift().AsInteger()
		apple2helpers.Gotoxy(this.Interpreter, x, y)
	}
	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(0)))

	return nil
}

func (this *PlusCursorPosition) Syntax() string {

	/* vars */
	var result string

	result = "BGCOLOR{col}"

	/* enforce non void return */
	return result

}

func (this *PlusCursorPosition) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusCursorPosition(a int, b int, params types.TokenList) *PlusCursorPosition {
	this := &PlusCursorPosition{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "BGCOLOR"

	this.NamedParams = []string{"x", "y"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.INTEGER, "0"),
		*types.NewToken(types.INTEGER, "0"),
	}
	this.Raw = true

	return this
}
