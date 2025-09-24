package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/types"
	//	"paleotronic.com/utils"
)

type PlusVarSlot struct {
	dialect.CoreFunction
}

func (this *PlusVarSlot) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	slotid := 0
	varname := ""
	if this.Stack.Size() > 0 {
		slotid = (this.Stack.Shift().AsInteger() - 1) % memory.OCTALYZER_NUM_INTERPRETERS
		varname = this.Stack.Shift().Content
	}

	e := this.Interpreter.GetProducer().GetInterpreter(slotid)

	tl := types.NewTokenList()
	tl.Push(types.NewToken(types.VARIABLE, varname))

	rtok := e.ParseTokensForResult(*tl)

	this.Stack.Push(&rtok)

	return nil
}

func (this *PlusVarSlot) Syntax() string {

	/* vars */
	var result string

	result = "Activate{name}"

	/* enforce non void return */
	return result

}

func (this *PlusVarSlot) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusVarSlot(a int, b int, params types.TokenList) *PlusVarSlot {
	this := &PlusVarSlot{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "Activate"
	this.MinParams = 2
	this.MaxParams = 2

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "vm", Default: *types.NewToken(types.NUMBER, "1")},
			dialect.FunctionParamDef{Name: "varname", Default: *types.NewToken(types.NUMBER, "255")},
		},
	)

	return this
}
