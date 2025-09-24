package plus

import (
	//"time"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/types"
	"paleotronic.com/log"
	"paleotronic.com/utils"
)

type PlusTextDrawHeight struct {
	dialect.CoreFunction
}

func (this *PlusTextDrawHeight) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		log.Println(e)
		return e
	}

	log.Println("In PlusTextDrawHeight()")

	c := params.Get(0).AsInteger()

	apple2helpers.PixelTextHeight = c

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusTextDrawHeight) Syntax() string {

	/* vars */
	var result string

	result = "CubePlot{x, y, z, c}"

	/* enforce non void return */
	return result

}

func (this *PlusTextDrawHeight) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusTextDrawHeight(a int, b int, params types.TokenList) *PlusTextDrawHeight {
	this := &PlusTextDrawHeight{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "CubePlot"
	this.MaxParams = 1
	this.MinParams = 1

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "size", Default: *types.NewToken(types.NUMBER, "2")},
		},
	)

	return this
}
