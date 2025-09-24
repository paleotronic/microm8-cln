package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/utils"
)

type PlusTXS struct {
	dialect.CoreFunction
}

func (this *PlusTXS) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	if !this.Query {
		q := this.ValueMap["enabled"]
        memory.WarmStart = true
		apple2helpers.GetCPU(this.Interpreter).IgnoreStackFallouts = (q.AsInteger() != 0)
        memory.WarmStart = false
	}

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusTXS) Syntax() string {

	/* vars */
	var result string

	result = "SETCLASSICCPU{}"

	/* enforce non void return */
	return result

}

func (this *PlusTXS) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewPlusTXS(a int, b int, params types.TokenList) *PlusTXS {
	this := &PlusTXS{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "SETCLASSICCPU"

	this.NamedDefaults = []types.Token{ *types.NewToken(types.NUMBER, "0") }
	this.NamedParams = []string{ "enabled" }
	this.Raw = true

	return this
}
