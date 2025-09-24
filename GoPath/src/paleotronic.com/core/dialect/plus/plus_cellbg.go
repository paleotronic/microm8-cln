package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusPositionBG struct {
	dialect.CoreFunction
}

func (this *PlusPositionBG) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	if !this.Query {
		x := params.Shift().AsInteger()
		y := params.Shift().AsInteger()
		c := params.Shift().AsInteger()
		apple2helpers.SetBGColorXY(this.Interpreter, x, y, uint64(c))
	}
	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(0)))

	return nil
}

func (this *PlusPositionBG) Syntax() string {

	/* vars */
	var result string

	result = "BGCOLOR{col}"

	/* enforce non void return */
	return result

}

func (this *PlusPositionBG) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusPositionBG(a int, b int, params types.TokenList) *PlusPositionBG {
	this := &PlusPositionBG{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "BGCOLOR"

	this.NamedParams = []string{"x", "y", "c"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.INTEGER, "0"),
		*types.NewToken(types.INTEGER, "0"),
		*types.NewToken(types.INTEGER, "15"),
	}
	this.Raw = true

	return this
}
