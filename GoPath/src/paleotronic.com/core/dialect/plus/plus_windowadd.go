package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/core/hardware/apple2helpers"
//	"paleotronic.com/runestring"

	//"paleotronic.com/log"
)

type PlusAddWindow struct {
	dialect.CoreFunction
}

func (this *PlusAddWindow) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	if !this.Query {

		// stuff
		sxt := this.ValueMap["sx"];
		syt := this.ValueMap["sy"];
		ext := this.ValueMap["ex"];
		eyt := this.ValueMap["ey"];

		sx := sxt.AsInteger()
		sy := syt.AsInteger()
		ex := ext.AsInteger()
		ey := eyt.AsInteger()

		name := this.ValueMap["name"].Content

		apple2helpers.TextAddWindow(this.Interpreter, name, sx, sy, ex, ey)

	}

	this.Stack.Push(types.NewToken(types.STRING, ""))

	return nil
}

func (this *PlusAddWindow) Syntax() string {

	/* vars */
	var result string

	result = "WINDOW.ADD{name,sx,sy,ex,ey}"

	/* enforce non void return */
	return result

}

func (this *PlusAddWindow) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusAddWindow(a int, b int, params types.TokenList) *PlusAddWindow {
	this := &PlusAddWindow{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "WINDOW.ADD"

	this.NamedParams = []string{ "name", "sx", "sy", "ex", "ey" }
	this.NamedDefaults = []types.Token{
		*types.NewToken( types.STRING, "DEFAULT" ),
		*types.NewToken( types.NUMBER, "0" ),
		*types.NewToken( types.NUMBER, "0" ),
		*types.NewToken( types.NUMBER, "79" ),
		*types.NewToken( types.NUMBER, "47" ),
	}
	this.Raw = true

	return this
}
