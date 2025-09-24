package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	//"paleotronic.com/core/memory"
	"paleotronic.com/utils"
	//"paleotronic.com/core/settings"
	//"time"
)

type PlusCamRightDir struct {
	dialect.CoreFunction
}

func (this *PlusCamRightDir) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	index := this.Interpreter.GetMemIndex()
	mm := this.Interpreter.GetMemoryMap()
	cindex := mm.GetCameraConfigure(index)
	control := types.NewOrbitController(mm, index, cindex)

	if this.Query {

		v := float64(0)
		pos := control.GetLeftAxis().MulF(-1)
		switch this.QueryVar {
		case "x":
			v = pos[0]
		case "y":
			v = pos[1]
		case "z":
			v = pos[2]
		}

		this.Stack.Push(types.NewToken(types.NUMBER, utils.FloatToStr(v)))
	} else {
		// xt := this.ValueMap["x"]
		// x := xt.AsExtended()
		// yt := this.ValueMap["y"]
		// y := yt.AsExtended()
		// zt := this.ValueMap["z"]
		// z := zt.AsExtended()

		// control.SetRightDir(mgl64.Vec3{x, y, z})

		this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))
	}

	return nil
}

func (this *PlusCamRightDir) Syntax() string {

	/* vars */
	var result string

	result = "CAMLOC{x,y,z}"

	/* enforce non void return */
	return result

}

func (this *PlusCamRightDir) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusCamRightDir(a int, b int, params types.TokenList) *PlusCamRightDir {
	this := &PlusCamRightDir{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "CAMLOC"
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
