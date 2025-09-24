package plus

import (
	//"time"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusCubeLine struct {
	dialect.CoreFunction
}

func (this *PlusCubeLine) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	x0 := params.Get(0).AsInteger()
	y0 := params.Get(1).AsInteger()
	z0 := params.Get(2).AsInteger()
	x1 := params.Get(3).AsInteger()
	y1 := params.Get(4).AsInteger()
	z1 := params.Get(5).AsInteger()
	c := apple2helpers.GetCOLOR(this.Interpreter)
	if params.Size() > 6 {
		c = params.Get(6).AsInteger()
	}

	// enable vector layer
	v := apple2helpers.CUBE(this.Interpreter)
	v.Line3d(
		float64(x0),
		float64(y0),
		float64(z0),
		float64(x1),
		float64(y1),
		float64(z1),
		uint8(c),
	)
	v.Render()
	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusCubeLine) Syntax() string {

	/* vars */
	var result string

	result = "CubeLine{x0, y0, z0, x1, y1, z1, c}"

	/* enforce non void return */
	return result

}

func (this *PlusCubeLine) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusCubeLine(a int, b int, params types.TokenList) *PlusCubeLine {
	this := &PlusCubeLine{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "CubeLine"
	this.MaxParams = 7
	this.MinParams = 6

	return this
}
