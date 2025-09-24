package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/core/hardware/apple2helpers"
//	"paleotronic.com/runestring"

	//"paleotronic.com/log"
)

type PlusHideCursor struct {
	dialect.CoreFunction
}

func (this *PlusHideCursor) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	if !this.Query {

		apple2helpers.TextHideCursor(this.Interpreter)

	}

	this.Stack.Push(types.NewToken(types.STRING, ""))

	return nil
}

func (this *PlusHideCursor) Syntax() string {

	/* vars */
	var result string

	result = "WINDOW.USE{name}"

	/* enforce non void return */
	return result

}

func (this *PlusHideCursor) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusHideCursor(a int, b int, params types.TokenList) *PlusHideCursor {
	this := &PlusHideCursor{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "WINDOW.USE"

	this.NamedParams = []string{}
	this.NamedDefaults = []types.Token{
	}
	this.Raw = true

	return this
}
