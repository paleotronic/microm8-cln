package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusSlotSelect struct {
	dialect.CoreFunction
}

func (this *PlusSlotSelect) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	slotid := params.Shift().AsInteger() - 1
	if slotid < 0 {
		slotid = 0
	}

	if slotid == this.Interpreter.GetMemIndex() {
		slotid = 0
	} else {
		slotid = slotid | 128
	}

	this.Interpreter.GetMemoryMap().IntSetTargetSlot(this.Interpreter.GetMemIndex(), slotid)

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusSlotSelect) Syntax() string {

	/* vars */
	var result string

	result = "Select{index}"

	/* enforce non void return */
	return result

}

func (this *PlusSlotSelect) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusSlotSelect(a int, b int, params types.TokenList) *PlusSlotSelect {
	this := &PlusSlotSelect{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "Select"
	this.MinParams = 1
	this.MaxParams = 1
	this.CoreFunction.NoRedirect = true

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "vm", Default: *types.NewToken(types.NUMBER, "1")},
		},
	)

	return this
}
