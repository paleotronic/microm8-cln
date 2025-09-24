package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusSlotDefault struct {
	dialect.CoreFunction
}

func (this *PlusSlotDefault) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	slotid := 0

	this.Interpreter.GetMemoryMap().IntSetTargetSlot(this.Interpreter.GetMemIndex(), slotid)

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusSlotDefault) Syntax() string {

	/* vars */
	var result string

	result = "Select{index}"

	/* enforce non void return */
	return result

}

func (this *PlusSlotDefault) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewPlusSlotDefault(a int, b int, params types.TokenList) *PlusSlotDefault {
	this := &PlusSlotDefault{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "Select"
	this.MinParams = 0
	this.MaxParams = 0
	this.CoreFunction.NoRedirect = true

	return this
}
