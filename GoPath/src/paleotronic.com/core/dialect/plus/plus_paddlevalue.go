package plus

import (
	"paleotronic.com/fmt"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusPaddleValue struct {
	dialect.CoreFunction
}

func (this *PlusPaddleValue) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	slotid := this.Interpreter.GetMemIndex()
	index := 0
	if params.Size() > 0 {
		index = params.Shift().AsInteger()
	}

	if params.Size() > 0 {
		num := params.Shift().AsExtended()
		if num < -1 {
			num = -1
		} else if num > 1 {
			num = 1
		}
		value := uint64(num*127 + 127)
		this.Interpreter.GetMemoryMap().IntSetPaddleValue(slotid, index, value)
	}

	fmt.Printf("Paddle index is %d\n", index)
	value := this.Interpreter.GetMemoryMap().IntGetPaddleValue(slotid, index)
	fmt.Printf("Paddle value is %d\n", value)

	v := 2*(float64(value)/255) - 1
	if v > 1 {
		v = 1
	}

	this.Stack.Push(types.NewToken(types.NUMBER, utils.FloatToStr(v)))

	return nil
}

func (this *PlusPaddleValue) Syntax() string {

	/* vars */
	var result string

	result = "Activate{name}"

	/* enforce non void return */
	return result

}

func (this *PlusPaddleValue) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusPaddleValue(a int, b int, params types.TokenList) *PlusPaddleValue {
	this := &PlusPaddleValue{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "Activate"
	this.MinParams = 0
	this.MaxParams = 2

	return this
}
