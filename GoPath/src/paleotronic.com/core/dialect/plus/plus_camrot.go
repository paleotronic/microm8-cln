package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types" //	"paleotronic.com/core/memory"
	"paleotronic.com/utils"      //	"time"
	//	"paleotronic.com/core/settings"
)

type PlusCamAng struct {
	dialect.CoreFunction
}

func (this *PlusCamAng) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	index := this.Interpreter.GetMemIndex()
	mm := this.Interpreter.GetMemoryMap()
	cindex := mm.GetCameraConfigure(index)
	control := types.NewOrbitController(mm, index, cindex)

	if this.Query {

		angle := control.GetAngle()

		var v float64

		switch this.QueryVar {
		case "x":
			v = angle.X()
		case "y":
			v = angle.Y()
		case "z":
			v = angle.Z()
		}

		this.Stack.Push(types.NewToken(types.NUMBER, utils.FloatToStr(v)))

	} else {

		xt := this.ValueMap["x"]
		x := xt.AsExtended()
		yt := this.ValueMap["y"]
		y := yt.AsExtended()
		zt := this.ValueMap["z"]
		z := zt.AsExtended()

		//		angle := control.GetAngle()
		//		fmt.Printf("Angle is %v\n", angle)

		//		angle[0] += x
		//		angle[1] += y
		//		angle[2] += z

		//		for angle[0] < 0 {
		//			angle[0] += 360
		//		}
		//		for angle[0] >= 360 {
		//			angle[0] -= 360
		//		}

		//		for angle[1] < 0 {
		//			angle[1] += 360
		//		}
		//		for angle[1] >= 360 {
		//			angle[1] -= 360
		//		}

		//		for angle[2] < 0 {
		//			angle[2] += 360
		//		}
		//		for angle[2] >= 360 {
		//			angle[2] -= 360
		//		}

		//		fmt.Printf("Angle after adjustment is %v\n", angle)
		//		control.SetRotation(angle)
		//		control.Update()

		control.Rotate(x, y, z)
		control.Update()

	}

	return nil
}

func (this *PlusCamAng) Syntax() string {

	/* vars */
	var result string

	result = "ANG{ax,ay,az}"

	/* enforce non void return */
	return result

}

func (this *PlusCamAng) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusCamAng(a int, b int, params types.TokenList) *PlusCamAng {
	this := &PlusCamAng{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "CAMANG"
	this.MinParams = 1
	this.MaxParams = 3
	this.Raw = true
	this.NamedParams = []string{"x", "y", "z"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.INTEGER, "0"),
		*types.NewToken(types.INTEGER, "0"),
		*types.NewToken(types.INTEGER, "0"),
	}
	this.EvalSingleParam = true

	return this
}
