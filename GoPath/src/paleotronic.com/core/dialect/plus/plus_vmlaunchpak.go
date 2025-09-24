package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusVMLaunchPAK struct {
	dialect.CoreFunction
}

func (this *PlusVMLaunchPAK) FunctionExecute(params *types.TokenList) error {

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

	settings.MicroPakPath = lang
	this.Interpreter.GetMemoryMap().IntSetSlotRestart(0, true)

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusVMLaunchPAK) Syntax() string {

	/* vars */
	var result string

	result = "SPAWN{name}"

	/* enforce non void return */
	return result

}

func (this *PlusVMLaunchPAK) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusVMLaunchPAK(a int, b int, params types.TokenList) *PlusVMLaunchPAK {
	this := &PlusVMLaunchPAK{}

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
