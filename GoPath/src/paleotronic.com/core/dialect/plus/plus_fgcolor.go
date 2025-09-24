package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusFGColor struct {
	dialect.CoreFunction
}

func (this *PlusFGColor) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	if !this.Query {
		t := this.ValueMap["color"]
		c := t.AsInteger()
		apple2helpers.SetFGColor(this.Interpreter, uint64(c))
	}
	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(0)))

	return nil
}

func (this *PlusFGColor) Syntax() string {

	/* vars */
	var result string

	result = "FGCOLOR{col}"

	/* enforce non void return */
	return result

}

func (this *PlusFGColor) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusFGColor(a int, b int, params types.TokenList) *PlusFGColor {
	this := &PlusFGColor{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "FGCOLOR"

	this.NamedParams = []string{"color"}
	this.NamedDefaults = []types.Token{*types.NewToken(types.INTEGER, "15")}
	this.Raw = true

	return this
}
