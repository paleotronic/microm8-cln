package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/core/hardware/apple2helpers"
//	"paleotronic.com/runestring"

	//"paleotronic.com/log"
)

type PlusUseWindow struct {
	dialect.CoreFunction
}

func (this *PlusUseWindow) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	if !this.Query {

		name := this.ValueMap["name"].Content

		apple2helpers.TextUseWindow(this.Interpreter, name)

	}

	this.Stack.Push(types.NewToken(types.STRING, ""))

	return nil
}

func (this *PlusUseWindow) Syntax() string {

	/* vars */
	var result string

	result = "WINDOW.USE{name}"

	/* enforce non void return */
	return result

}

func (this *PlusUseWindow) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusUseWindow(a int, b int, params types.TokenList) *PlusUseWindow {
	this := &PlusUseWindow{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "WINDOW.USE"

	this.NamedParams = []string{ "name" }
	this.NamedDefaults = []types.Token{
		*types.NewToken( types.STRING, "DEFAULT" ),
	}
	this.Raw = true

	return this
}
