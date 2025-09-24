package plus

import (
	"errors"

	"paleotronic.com/core/dialect"
	//	"paleotronic.com/core/memory"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusDisableSlot struct {
	dialect.CoreFunction
}

func (this *PlusDisableSlot) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	st := params.Shift()
	slotid := st.AsInteger() - 1

	if slotid == 0 || slotid >= 7 {
		return errors.New("invalid slot")
	}

	e := this.Interpreter.GetProducer().GetInterpreter(slotid)

	e.EndRemote()

	this.Interpreter.GetMemoryMap().IntSetActiveState(slotid, 0)
	this.Interpreter.GetMemoryMap().IntSetLayerState(slotid, 0)

	this.Interpreter.GetProducer().DropInterpreter(slotid)

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusDisableSlot) Syntax() string {

	/* vars */
	var result string

	result = "SETCLASSICCPU{}"

	/* enforce non void return */
	return result

}

func (this *PlusDisableSlot) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewPlusDisableSlot(a int, b int, params types.TokenList) *PlusDisableSlot {
	this := &PlusDisableSlot{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "SETCLASSICCPU"

	this.MaxParams = 1
	this.MinParams = 1

	return this
}
