package plus

import (
	//"time"

	"paleotronic.com/core/dialect"
	//"paleotronic.com/core/memory"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
	//"paleotronic.com/core/settings"
)

type PlusCamDolly struct {
	dialect.CoreFunction
}

func (this *PlusCamDolly) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	c := params.Shift().AsExtended()

	index := this.Interpreter.GetMemIndex()
	mm := this.Interpreter.GetMemoryMap()
	cindex := mm.GetCameraConfigure(index)
	control := types.NewOrbitController(mm, index, cindex)
	control.SetDollyRate(c)

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusCamDolly) Syntax() string {

	/* vars */
	var result string

	result = "CAMDOLLY{v}"

	/* enforce non void return */
	return result

}

func (this *PlusCamDolly) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusCamDolly(a int, b int, params types.TokenList) *PlusCamDolly {
	this := &PlusCamDolly{}

	/* vars */
	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "CAMDOLLY"

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "rate", Default: *types.NewToken(types.NUMBER, "0.1")},
		},
	)

	return this
}
