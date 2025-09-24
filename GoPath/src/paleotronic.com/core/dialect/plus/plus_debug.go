package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusDebug struct {
	dialect.CoreFunction
}

func (this *PlusDebug) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	if !this.Query {
		q := this.ValueMap["enabled"]
		this.Interpreter.SetDebug( (q.AsInteger() == 1) )
	}

	if this.Interpreter.IsDebug() {
		this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))
	} else {
		this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(0)))
	}

	return nil
}

func (this *PlusDebug) Syntax() string {

	/* vars */
	var result string

	result = "DEBUG{}"

	/* enforce non void return */
	return result

}

func (this *PlusDebug) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewPlusDebug(a int, b int, params types.TokenList) *PlusDebug {
	this := &PlusDebug{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "SETCLASSICHGR"

	this.NamedDefaults = []types.Token{ *types.NewToken(types.NUMBER, "0") }
	this.NamedParams = []string{ "enabled" }
	this.Raw = true

	return this
}
