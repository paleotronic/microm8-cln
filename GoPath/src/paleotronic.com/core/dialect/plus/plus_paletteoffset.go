package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusPaletteOffset struct {
	dialect.CoreFunction
}

func (this *PlusPaletteOffset) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	//if !this.Query {
	c := params.Shift().AsInteger()
	o := params.Shift().AsInteger()
	// Set Palette offset
	//list := apple2helpers.GetActiveVideoModes(this.Interpreter)
	for _, id := range paletteList[this.Interpreter.GetMemIndex()] {
		apple2helpers.SetColorOffset(this.Interpreter, id, c, o)
	}
	//}
	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(0)))

	return nil
}

func (this *PlusPaletteOffset) Syntax() string {

	/* vars */
	var result string

	result = "BGCOLOR{col}"

	/* enforce non void return */
	return result

}

func (this *PlusPaletteOffset) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.INTEGER)
	result = append(result, types.INTEGER)

	/* enforce non void return */
	return result

}

func NewPlusPaletteOffset(a int, b int, params types.TokenList) *PlusPaletteOffset {
	this := &PlusPaletteOffset{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "BGCOLOR"

	this.MaxParams = 2
	this.MinParams = 2

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "color", Default: *types.NewToken(types.NUMBER, "1")},
			dialect.FunctionParamDef{Name: "offset", Default: *types.NewToken(types.NUMBER, "90")},
		},
	)

	return this
}
