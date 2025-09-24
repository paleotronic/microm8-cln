package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
	//	"paleotronic.com/files"
	//	"paleotronic.com/core/exception"
	//	"paleotronic.com/core/hardware/apple2helpers"
	//	"paleotronic.com/api"
	//	"strings"
)

type PlusShareInput struct {
	dialect.CoreFunction
}

func (this *PlusShareInput) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	slotid := params.Shift().AsInteger() - 1
	command := params.Shift().Content

	e := this.Interpreter.GetProducer().GetInterpreter(slotid)
	e.SendRemoteText(command)

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusShareInput) Syntax() string {

	/* vars */
	var result string

	result = "Login{name,pass,onerror,var}"

	/* enforce non void return */
	return result

}

func (this *PlusShareInput) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusShareInput(a int, b int, params types.TokenList) *PlusShareInput {
	this := &PlusShareInput{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "Login"
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
