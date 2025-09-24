package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusSlotTarget struct {
	dialect.CoreFunction
}

func (this *PlusSlotTarget) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	t := this.Interpreter.GetMemoryMap().IntGetTargetSlot(this.Interpreter.GetMemIndex())
	if t&128 != 0 {
		t = t & 127
	} else {
		t = this.Interpreter.GetMemIndex()
	}

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(t+1)))

	return nil
}

func (this *PlusSlotTarget) Syntax() string {

	/* vars */
	var result string

	result = "Select{index}"

	/* enforce non void return */
	return result

}

func (this *PlusSlotTarget) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewPlusSlotTarget(a int, b int, params types.TokenList) *PlusSlotTarget {
	this := &PlusSlotTarget{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "Select"
	this.MinParams = 0
	this.MaxParams = 0
	this.CoreFunction.NoRedirect = true

	return this
}
