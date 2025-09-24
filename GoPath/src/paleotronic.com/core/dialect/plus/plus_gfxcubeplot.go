package plus

import (
	//"time"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/types"
	"paleotronic.com/log"
	"paleotronic.com/utils"
)

type PlusCubePlot struct {
	dialect.CoreFunction
}

func (this *PlusCubePlot) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		log.Println(e)
		return e
	}

	log.Println("In PlusCubePlot()")

	x := params.Get(0).AsInteger()
	y := params.Get(1).AsInteger()
	z := params.Get(2).AsInteger()
	c := params.Get(3).AsInteger()

	// enable cube layer
	v := apple2helpers.CUBE(this.Interpreter)
	v.Plot(uint8(x), uint8(y), uint8(z), uint8(c))

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusCubePlot) Syntax() string {

	/* vars */
	var result string

	result = "CubePlot{x, y, z, c}"

	/* enforce non void return */
	return result

}

func (this *PlusCubePlot) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusCubePlot(a int, b int, params types.TokenList) *PlusCubePlot {
	this := &PlusCubePlot{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "CubePlot"
	this.MaxParams = 4
	this.MinParams = 3

	return this
}
