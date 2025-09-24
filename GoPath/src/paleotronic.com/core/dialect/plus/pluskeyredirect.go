package plus

import (
	"log"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusKeyRedirect struct {
	dialect.CoreFunction
}

func (this *PlusKeyRedirect) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	slotid := params.Shift().AsInteger() - 1
	if slotid < 0 {
		slotid = 0
	}

	tslotid := params.Shift().AsInteger() - 1
	if tslotid < 0 {
		tslotid = 0
	}

	log.Printf("enable key redirect from %d to %d", slotid, tslotid)

	this.Interpreter.GetMemoryMap().IntEnableKeyRedirect(slotid, tslotid, true)

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusKeyRedirect) Syntax() string {

	/* vars */
	var result string

	result = "Select{index}"

	/* enforce non void return */
	return result

}

func (this *PlusKeyRedirect) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusKeyRedirect(a int, b int, params types.TokenList) *PlusKeyRedirect {
	this := &PlusKeyRedirect{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "Select"
	this.MinParams = 2
	this.MaxParams = 2
	this.CoreFunction.NoRedirect = true

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "srcvm", Default: *types.NewToken(types.NUMBER, "1")},
			dialect.FunctionParamDef{Name: "targetvm", Default: *types.NewToken(types.NUMBER, "1")},
		},
	)

	return this
}
