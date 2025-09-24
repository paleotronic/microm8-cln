package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusReboot struct {
	dialect.CoreFunction
}

func (this *PlusReboot) FunctionExecute(params *types.TokenList) error {

	//if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	//	backend.REBOOT_NEEDED = true
	this.Interpreter.GetMemoryMap().IntSetSlotRestart(this.Interpreter.GetMemIndex(), true)

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusReboot) Syntax() string {

	/* vars */
	var result string

	result = "REBOOT{}"

	/* enforce non void return */
	return result

}

func (this *PlusReboot) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewPlusReboot(a int, b int, params types.TokenList) *PlusReboot {
	this := &PlusReboot{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "REBOOT"

	return this
}
