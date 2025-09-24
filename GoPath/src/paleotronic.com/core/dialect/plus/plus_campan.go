package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	//	"paleotronic.com/core/memory"
	"paleotronic.com/utils"
	//	"time"
	//	"paleotronic.com/core/settings"
)

type PlusCamPan struct {
	dialect.CoreFunction
}

func (this *PlusCamPan) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	index := this.Interpreter.GetMemIndex()
	mm := this.Interpreter.GetMemoryMap()
	cindex := mm.GetCameraConfigure(index)
	control := types.NewOrbitController(mm, index, cindex)

	if this.Query {

		var v float64
		switch this.QueryVar {
		case "x":
			v = control.GetPanX()
		case "y":
			v = control.GetPanY()
		}

		this.Stack.Push(types.NewToken(types.NUMBER, utils.FloatToStr(v)))

	} else {

		xt := this.ValueMap["x"]
		x := xt.AsExtended()
		yt := this.ValueMap["y"]
		y := yt.AsExtended()

		control.Pan(x*10, y*10, 0)

	}

	return nil
}

func (this *PlusCamPan) Syntax() string {

	/* vars */
	var result string

	result = "ANG{ax,ay,az}"

	/* enforce non void return */
	return result

}

func (this *PlusCamPan) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusCamPan(a int, b int, params types.TokenList) *PlusCamPan {
	this := &PlusCamPan{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "CAMPAN"
	this.MinParams = 1
	this.MaxParams = 2
	this.Raw = true
	this.NamedParams = []string{"x", "y"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.INTEGER, "0"),
		*types.NewToken(types.INTEGER, "0"),
	}
	this.EvalSingleParam = true

	return this
}
