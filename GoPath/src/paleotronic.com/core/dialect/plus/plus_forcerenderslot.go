package plus

import (
	"errors"

	"paleotronic.com/core/dialect"
	//	"paleotronic.com/core/memory"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusForceRenderSlot struct {
	dialect.CoreFunction
}

func (this *PlusForceRenderSlot) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	st := params.Shift()
	slotid := st.AsInteger() - 1
	en := params.Shift()
	enabled := en.AsInteger()

	if slotid < 0 || slotid >= memory.OCTALYZER_NUM_INTERPRETERS {
		return errors.New("invalid slot")
	}

	this.Interpreter.GetMemoryMap().IntSetLayerForceState(slotid, uint64(enabled))

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusForceRenderSlot) Syntax() string {

	/* vars */
	var result string

	result = "SETCLASSICCPU{}"

	/* enforce non void return */
	return result

}

func (this *PlusForceRenderSlot) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusForceRenderSlot(a int, b int, params types.TokenList) *PlusForceRenderSlot {
	this := &PlusForceRenderSlot{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "SETCLASSICCPU"

	this.MaxParams = 2
	this.MinParams = 2

	return this
}
