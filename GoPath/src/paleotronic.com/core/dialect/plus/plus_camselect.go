package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	//	"paleotronic.com/core/memory"
	//	"paleotronic.com/utils"
	//"time"
)

type PlusCamSelect struct {
	dialect.CoreFunction
}

func (this *PlusCamSelect) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	c := params.Shift().AsInteger()

	index := this.Interpreter.GetMemIndex()
	mm := this.Interpreter.GetMemoryMap()
	mm.SetCameraConfigure(index, c)

	return nil
}

func (this *PlusCamSelect) Syntax() string {

	/* vars */
	var result string

	result = "CAM.SELECT{v}"

	/* enforce non void return */
	return result

}

func (this *PlusCamSelect) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusCamSelect(a int, b int, params types.TokenList) *PlusCamSelect {
	this := &PlusCamSelect{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "CAMERA.SELECT"

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "camera", Default: *types.NewToken(types.NUMBER, "0")},
		},
	)

	return this
}
