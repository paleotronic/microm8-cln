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

type PlusCamOrbit struct {
	dialect.CoreFunction
}

func (this *PlusCamOrbit) FunctionExecute(params *types.TokenList) error {

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

	a := control.GetAngle()
	if z == -999 {
		z = a[2]
	}

	control.SetRotation(&glmath.Vector3{x, y, z})
	control.Update()

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusCamOrbit) Syntax() string {

	/* vars */
	var result string

	result = "CAMORBIT{pitch,yaw}"

	/* enforce non void return */
	return result

}

func (this *PlusCamOrbit) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusCamOrbit(a int, b int, params types.TokenList) *PlusCamOrbit {
	this := &PlusCamOrbit{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "CAMORBIT"
	this.MinParams = 2
	this.MaxParams = 3

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "pitch", Default: *types.NewToken(types.NUMBER, "0")},
			dialect.FunctionParamDef{Name: "yaw", Default: *types.NewToken(types.NUMBER, "0")},
			dialect.FunctionParamDef{Name: "roll", Default: *types.NewToken(types.NUMBER, "-999")},
		},
	)

	return this
}
