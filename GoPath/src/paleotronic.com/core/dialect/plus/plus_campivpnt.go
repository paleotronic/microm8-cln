package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/core/types/glmath"

	//	"paleotronic.com/core/memory"
	"paleotronic.com/utils"
	//	"time"
	//	"paleotronic.com/core/settings"
)

type PlusCamPivPnt struct {
	dialect.CoreFunction
}

func (this *PlusCamPivPnt) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}
	x := this.Stack.Shift().AsExtended()
	y := this.Stack.Shift().AsExtended()
	z := this.Stack.Shift().AsExtended()
	//this.Interpreter.GetVDU().CamMove(float32(x),float32(y),float32(z))
	index := this.Interpreter.GetMemIndex()
	mm := this.Interpreter.GetMemoryMap()
	cindex := mm.GetCameraConfigure(index)
	control := types.NewOrbitController(mm, index, cindex)
	control.SetTarget(&glmath.Vector3{x, y, z})
	//control.LookAtWithPosition(control.GetPosition(), &glmath.Vector3{x, y, z})
	control.Update()

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusCamPivPnt) Syntax() string {

	/* vars */
	var result string

	result = "CAMPIVPNT{x,y,z}"

	/* enforce non void return */
	return result

}

func (this *PlusCamPivPnt) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusCamPivPnt(a int, b int, params types.TokenList) *PlusCamPivPnt {
	this := &PlusCamPivPnt{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "CAMPIVPNT"
	this.MinParams = 3
	this.MaxParams = 3

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "x", Default: *types.NewToken(types.NUMBER, "0")},
			dialect.FunctionParamDef{Name: "y", Default: *types.NewToken(types.NUMBER, "0")},
			dialect.FunctionParamDef{Name: "z", Default: *types.NewToken(types.NUMBER, "0")},
		},
	)

	return this
}
