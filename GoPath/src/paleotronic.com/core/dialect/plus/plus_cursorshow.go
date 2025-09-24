package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/core/hardware/apple2helpers"
//	"paleotronic.com/runestring"

	//"paleotronic.com/log"
)

type PlusShowCursor struct {
	dialect.CoreFunction
}

func (this *PlusShowCursor) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	if !this.Query {

		apple2helpers.TextShowCursor(this.Interpreter)

	}

	this.Stack.Push(types.NewToken(types.STRING, ""))

	return nil
}

func (this *PlusShowCursor) Syntax() string {

	/* vars */
	var result string

	result = "WINDOW.USE{name}"

	/* enforce non void return */
	return result

}

func (this *PlusShowCursor) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusShowCursor(a int, b int, params types.TokenList) *PlusShowCursor {
	this := &PlusShowCursor{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "WINDOW.USE"

	this.NamedParams = []string{}
	this.NamedDefaults = []types.Token{
	}
	this.Raw = true

	return this
}
