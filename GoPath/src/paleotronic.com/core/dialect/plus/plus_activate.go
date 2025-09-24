package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusActivateSlot struct {
	dialect.CoreFunction
}

func (this *PlusActivateSlot) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	slotid := 1
	command := "LAYER.POS{X=0}"
	if this.Stack.Size() > 0 {
		slotid = this.Stack.Shift().AsInteger() - 1
	}
	if this.Stack.Size() > 0 {
		command = this.Stack.Shift().Content
	}
	e := this.Interpreter.GetProducer().GetInterpreter(slotid)

	e.SetDisabled(false)
	e.Parse(command)

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusActivateSlot) Syntax() string {

	/* vars */
	var result string

	result = "Activate{name}"

	/* enforce non void return */
	return result

}

func (this *PlusActivateSlot) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusActivateSlot(a int, b int, params types.TokenList) *PlusActivateSlot {
	this := &PlusActivateSlot{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "Activate"
	this.MinParams = 2
	this.MaxParams = 2

	return this
}
