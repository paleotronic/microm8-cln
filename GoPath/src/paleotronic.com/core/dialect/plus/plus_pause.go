package plus

import (
	"time"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusPause struct {
	dialect.CoreFunction
}

func (this *PlusPause) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	c := params.Shift().AsInteger()

	//time.Sleep(time.Duration(c) * time.Millisecond)
	this.Interpreter.WaitAdd(time.Duration(c) * time.Millisecond)

	//    wu := time.Now().Add(time.Duration(c) * time.Millisecond)
	//    this.Interpreter.SetWaitUntil()

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusPause) Syntax() string {

	/* vars */
	var result string

	result = "PAUSE{v}"

	/* enforce non void return */
	return result

}

func (this *PlusPause) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusPause(a int, b int, params types.TokenList) *PlusPause {
	this := &PlusPause{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "PAUSE"
	this.NoRedirect = true

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "ms", Default: *types.NewToken(types.NUMBER, "1000")},
		},
	)

	return this
}

func NewPlusSlotPause(a int, b int, params types.TokenList) *PlusPause {
	this := &PlusPause{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "PAUSE"

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "ms", Default: *types.NewToken(types.NUMBER, "1000")},
		},
	)

	return this
}
