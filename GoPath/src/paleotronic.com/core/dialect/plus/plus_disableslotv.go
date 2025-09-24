package plus

import (
	"errors"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusDisableSlotV struct {
	dialect.CoreFunction
}

func (this *PlusDisableSlotV) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	st := params.Shift()
	slotid := st.AsInteger() - 1

	if slotid < 0 || slotid >= memory.OCTALYZER_NUM_INTERPRETERS {
		return errors.New("invalid slot")
	}

	this.Interpreter.GetMemoryMap().IntSetActiveState(slotid, 0)
	this.Interpreter.GetMemoryMap().IntSetLayerState(slotid, 0)

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusDisableSlotV) Syntax() string {

	/* vars */
	var result string

	result = "SETCLASSICCPU{}"

	/* enforce non void return */
	return result

}

func (this *PlusDisableSlotV) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewPlusDisableSlotV(a int, b int, params types.TokenList) *PlusDisableSlotV {
	this := &PlusDisableSlotV{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "SETCLASSICCPU"

	this.MaxParams = 1
	this.MinParams = 1

	return this
}
