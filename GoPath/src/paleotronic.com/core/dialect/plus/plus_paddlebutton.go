package plus

import (
	"paleotronic.com/fmt"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusPaddleButton struct {
	dialect.CoreFunction
}

func (this *PlusPaddleButton) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	slotid := this.Interpreter.GetMemIndex()
	index := 0
	if params.Size() > 0 {
		index = params.Shift().AsInteger()
	}

	if params.Size() > 0 {
		num := params.Shift().AsInteger()
		if num < -1 {
			num = -1
		} else if num > 1 {
			num = 1
		}
		value := uint64(num)
		this.Interpreter.GetMemoryMap().IntSetPaddleButton(slotid, index, value)
	}

	fmt.Printf("Paddle index is %d\n", index)
	value := this.Interpreter.GetMemoryMap().IntGetPaddleButton(slotid, index)

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(int(value))))

	return nil
}

func (this *PlusPaddleButton) Syntax() string {

	/* vars */
	var result string

	result = "Activate{name}"

	/* enforce non void return */
	return result

}

func (this *PlusPaddleButton) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusPaddleButton(a int, b int, params types.TokenList) *PlusPaddleButton {
	this := &PlusPaddleButton{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "Activate"
	this.MinParams = 0
	this.MaxParams = 2

	return this
}
