package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusCGColor struct {
	dialect.CoreFunction
}

func (this *PlusCGColor) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	r := uint8(params.Shift().AsInteger())
	g := uint8(params.Shift().AsInteger())
	b := uint8(params.Shift().AsInteger())
	this.Interpreter.GetMemoryMap().SetBGColor(this.Interpreter.GetMemIndex(), r, g, b, 255)

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(0)))

	return nil
}

func (this *PlusCGColor) Syntax() string {

	/* vars */
	var result string

	result = "CGCOLOR{col}"

	/* enforce non void return */
	return result

}

func (this *PlusCGColor) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusCGColor(a int, b int, params types.TokenList) *PlusCGColor {
	this := &PlusCGColor{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "CGCOLOR"

	this.NamedParams = []string{"red", "green", "blue"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.INTEGER, "0"),
		*types.NewToken(types.INTEGER, "0"),
		*types.NewToken(types.INTEGER, "0"),
	}
	this.Raw = true

	return this
}
