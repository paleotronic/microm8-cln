package plus

import (
	//	"paleotronic.com/log"
	//"errors"
	//	"strings"

	//"paleotronic.com/files"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
	//	"paleotronic.com/api"
	//"paleotronic.com/core/interfaces"
)

type PlusTransfer struct {
	dialect.CoreFunction
}

// params:
// (1) hostname
// (2) name

func (this *PlusTransfer) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	slotid := params.Shift().AsInteger() - 1
	target := params.Shift().Content

	e := this.Interpreter.GetProducer().GetInterpreter(slotid)
	for e.GetChild() != nil {
		e = e.GetChild()
	}

	resp := e.TransferRemoteControl(target)

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(resp)))

	return nil
}

func (this *PlusTransfer) Syntax() string {

	/* vars */
	var result string

	result = "TRANSFER{slot, target}"

	/* enforce non void return */
	return result

}

func (this *PlusTransfer) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusTransfer(a int, b int, params types.TokenList) *PlusTransfer {
	this := &PlusTransfer{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "TRANSFER"
	this.MinParams = 2
	this.MaxParams = 2

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "vm", Default: *types.NewToken(types.NUMBER, "1")},
			dialect.FunctionParamDef{Name: "user", Default: *types.NewToken(types.STRING, "")},
		},
	)

	return this
}
