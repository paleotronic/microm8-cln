package plus

import (
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	//	"strings"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
)

type PlusFillScreen struct {
	dialect.CoreFunction
}

func (this *PlusFillScreen) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	tmp := this.ValueMap["c"]
	c := tmp.AsInteger()

	var fill func(ent interfaces.Interpretable, c uint64)

	modes := apple2helpers.GetActiveVideoModes(this.Interpreter)
	switch modes[0] {
	case "LOGR":
		fill = apple2helpers.GR40Fill
	case "DLGR":
		fill = apple2helpers.GR80Fill
	default:
		fill = apple2helpers.HGRFill
	}

	apple2helpers.Fill(this.Interpreter, c, fill)

	return nil
}

func (this *PlusFillScreen) Syntax() string {

	/* vars */
	var result string

	result = "OPEN{filename}"

	/* enforce non void return */
	return result

}

func (this *PlusFillScreen) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusFillScreen(a int, b int, params types.TokenList) *PlusFillScreen {
	this := &PlusFillScreen{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "OPEN"

	this.NamedParams = []string{"c"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.NUMBER, "1"),
	}
	this.Raw = true

	return this
}
