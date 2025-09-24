package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/servicebus"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusDiskIISwap struct {
	dialect.CoreFunction
}

func (this *PlusDiskIISwap) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	servicebus.SendServiceBusMessage(
		this.Interpreter.GetMemIndex(),
		servicebus.DiskIIExchangeDisks,
		"Disk swap",
	)
	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusDiskIISwap) Syntax() string {

	/* vars */
	var result string

	result = "Activate{name}"

	/* enforce non void return */
	return result

}

func (this *PlusDiskIISwap) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewPlusDiskIISwap(a int, b int, params types.TokenList) *PlusDiskIISwap {
	this := &PlusDiskIISwap{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "Activate"
	this.MinParams = 0
	this.MaxParams = 0

	return this
}
