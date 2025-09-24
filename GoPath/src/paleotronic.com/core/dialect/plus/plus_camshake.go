package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	//	"paleotronic.com/core/memory"
	//	"paleotronic.com/utils"
	//	"paleotronic.com/core/settings"
	//	"time"
)

type PlusCamShake struct {
	dialect.CoreFunction
}

func (this *PlusCamShake) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	frames := int(this.Stack.Shift().AsInteger())
	max := this.Stack.Shift().AsExtended() / 280

	index := this.Interpreter.GetMemIndex()
	mm := this.Interpreter.GetMemoryMap()
	cindex := mm.GetCameraConfigure(index)
	control := types.NewOrbitController(mm, index, cindex)
	control.SetShakeFrames(frames)
	control.SetShakeMax(max)

	return nil
}

func (this *PlusCamShake) Syntax() string {

	/* vars */
	var result string

	result = "CAM.SHAKE{frames,maxdiff}"

	/* enforce non void return */
	return result

}

func (this *PlusCamShake) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusCamShake(a int, b int, params types.TokenList) *PlusCamShake {
	this := &PlusCamShake{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "CAMLOC"
	this.MinParams = 2
	this.MaxParams = 2

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "frames", Default: *types.NewToken(types.NUMBER, "25")},
			dialect.FunctionParamDef{Name: "maxpixels", Default: *types.NewToken(types.NUMBER, "8")},
		},
	)

	return this
}
