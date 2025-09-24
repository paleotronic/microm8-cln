package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusVMLaunch struct {
	dialect.CoreFunction
}

func (this *PlusVMLaunch) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	lang := params.Shift().Content

	// e := this.Interpreter.NewChild(filename)
	// this.Interpreter.PutStr("Started " + e.GetName())
	// this.Interpreter.SetChild(e)
	// e.SetParent(this.Interpreter)
	// e.Bootstrap(filename, false)
	// if !s8webclient.CONN.IsAuthenticated() {
	// 	e.SetWorkDir("/local/")
	// 	this.Interpreter.SetWorkDir("/local/")
	// }

	apple2helpers.SwitchToDialect(this.Interpreter, lang)
	this.Interpreter.SetWorkDir("/local/")

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusVMLaunch) Syntax() string {

	/* vars */
	var result string

	result = "SPAWN{name}"

	/* enforce non void return */
	return result

}

func (this *PlusVMLaunch) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusVMLaunch(a int, b int, params types.TokenList) *PlusVMLaunch {
	this := &PlusVMLaunch{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "SPAWN"

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "dialect", Default: *types.NewToken(types.STRING, "FP")},
		},
	)

	return this
}
