package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusSlotID struct {
	dialect.CoreFunction
	OB bool
}

func (this *PlusSlotID) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	slotid := this.Interpreter.GetMemIndex()
	//if this.OB {
	slotid += 1
	//}

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(slotid)))

	return nil
}

func (this *PlusSlotID) Syntax() string {

	/* vars */
	var result string

	result = "Activate{name}"

	/* enforce non void return */
	return result

}

func (this *PlusSlotID) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewPlusSlotID(a int, b int, params types.TokenList) *PlusSlotID {
	this := &PlusSlotID{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "Activate"
	this.MinParams = 0
	this.MaxParams = 0

	return this
}

func NewPlusSlotIDOB(a int, b int, params types.TokenList) *PlusSlotID {
	this := &PlusSlotID{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "Activate"
	this.MinParams = 0
	this.MaxParams = 0
	this.OB = true

	return this
}
