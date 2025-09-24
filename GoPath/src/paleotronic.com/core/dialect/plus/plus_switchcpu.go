package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/core/memory"
	"paleotronic.com/utils"
)

type PlusSwitchCPU struct {
	dialect.CoreFunction
}

func (this *PlusSwitchCPU) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	if !this.Query {
		q := this.ValueMap["enabled"]
        memory.WarmStart = true
		if q.AsInteger() != 0 {
           // classic
           this.Interpreter.GetMemoryMap().IntSetProcessorState( this.Interpreter.GetMemIndex(), 128 )
        } else {
           this.Interpreter.GetMemoryMap().IntSetProcessorState( this.Interpreter.GetMemIndex(), 0 )
        }
        memory.WarmStart = false
	}

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusSwitchCPU) Syntax() string {

	/* vars */
	var result string

	result = "SETCLASSICCPU{}"

	/* enforce non void return */
	return result

}

func (this *PlusSwitchCPU) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewPlusSwitchCPU(a int, b int, params types.TokenList) *PlusSwitchCPU {
	this := &PlusSwitchCPU{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "SETCLASSICCPU"

	this.NamedDefaults = []types.Token{ *types.NewToken(types.NUMBER, "0") }
	this.NamedParams = []string{ "enabled" }
	this.Raw = true

	return this
}
