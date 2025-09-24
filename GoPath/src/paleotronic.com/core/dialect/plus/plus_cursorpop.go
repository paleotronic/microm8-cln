package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/types"
	//	"paleotronic.com/runestring"
	//"paleotronic.com/log"
)

type PlusPopCursor struct {
	dialect.CoreFunction
}

func (this *PlusPopCursor) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	if !this.Query {

		apple2helpers.TextPopCursor(this.Interpreter)

	}

	this.Stack.Push(types.NewToken(types.STRING, ""))

	return nil
}

func (this *PlusPopCursor) Syntax() string {

	/* vars */
	var result string

	result = "WINDOW.USE{name}"

	/* enforce non void return */
	return result

}

func (this *PlusPopCursor) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewPlusPopCursor(a int, b int, params types.TokenList) *PlusPopCursor {
	this := &PlusPopCursor{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "WINDOW.USE"

	this.NamedParams = []string{}
	this.NamedDefaults = []types.Token{}
	this.Raw = true

	return this
}
