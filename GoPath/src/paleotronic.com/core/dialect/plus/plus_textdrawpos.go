package plus

import (
	//"time"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/types"
	"paleotronic.com/log"
	"paleotronic.com/utils"
)

type PlusTextDrawPos struct {
	dialect.CoreFunction
}

func (this *PlusTextDrawPos) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		log.Println(e)
		return e
	}

	log.Println("In PlusTextDrawPos()")

	x := params.Get(0).AsInteger()
	y := params.Get(1).AsInteger()

	apple2helpers.PixelTextX = x
	apple2helpers.PixelTextY = y

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusTextDrawPos) Syntax() string {

	/* vars */
	var result string

	result = "CubePlot{x, y, z, c}"

	/* enforce non void return */
	return result

}

func (this *PlusTextDrawPos) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusTextDrawPos(a int, b int, params types.TokenList) *PlusTextDrawPos {
	this := &PlusTextDrawPos{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "CubePlot"
	this.MaxParams = 2
	this.MinParams = 2

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "x", Default: *types.NewToken(types.NUMBER, "0")},
			dialect.FunctionParamDef{Name: "y", Default: *types.NewToken(types.NUMBER, "0")},
		},
	)

	return this
}
