package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusSlotDo struct {
	dialect.CoreFunction
}

func (this *PlusSlotDo) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	slotid := 1
	command := "LAYER.POS{X=0}"
	if this.Stack.Size() > 0 {
		slotid = this.Stack.Shift().AsInteger() - 1
		command = this.Stack.Shift().Content
	}

	e := this.Interpreter.GetProducer().GetInterpreter(slotid)

	e.ResumeTheWorld()
	e.SetDisabled(false)
	e.Parse(command)

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusSlotDo) Syntax() string {

	/* vars */
	var result string

	result = "Activate{name}"

	/* enforce non void return */
	return result

}

func (this *PlusSlotDo) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusSlotDo(a int, b int, params types.TokenList) *PlusSlotDo {
	this := &PlusSlotDo{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "Activate"
	this.MinParams = 2
	this.MaxParams = 2

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "vm", Default: *types.NewToken(types.NUMBER, "1")},
			dialect.FunctionParamDef{Name: "command", Default: *types.NewToken(types.STRING, "")},
		},
	)

	return this
}
