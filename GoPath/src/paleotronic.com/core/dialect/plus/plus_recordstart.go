package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusRecordStart struct {
	dialect.CoreFunction
}

func (this *PlusRecordStart) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	fn := params.Shift().Content

	//	backend.REBOOT_NEEDED = true
	this.Interpreter.StartRecording(fn, false)

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusRecordStart) Syntax() string {

	/* vars */
	var result string

	result = "REBOOT{}"

	/* enforce non void return */
	return result

}

func (this *PlusRecordStart) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusRecordStart(a int, b int, params types.TokenList) *PlusRecordStart {
	this := &PlusRecordStart{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "REBOOT"
	this.MinParams = 1
	this.MaxParams = 1

	return this
}
