package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusPaletteDepth struct {
	dialect.CoreFunction
}

func (this *PlusPaletteDepth) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	//if !this.Query {
	c := params.Shift().AsInteger()
	o := params.Shift().AsInteger()

	if c != -1 {
		for _, id := range paletteList[this.Interpreter.GetMemIndex()] {
			apple2helpers.SetColorDepth(this.Interpreter, id, c, o)
		}
	} else {
		this.Interpreter.GetMemoryMap().IntSetVoxelDepth(this.Interpreter.GetMemIndex(), settings.VoxelDepth((o/10)-1))
	}

	//}
	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(0)))

	return nil
}

func (this *PlusPaletteDepth) Syntax() string {

	/* vars */
	var result string

	result = "BGCOLOR{col}"

	/* enforce non void return */
	return result

}

func (this *PlusPaletteDepth) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.INTEGER)
	result = append(result, types.INTEGER)

	/* enforce non void return */
	return result

}

func NewPlusPaletteDepth(a int, b int, params types.TokenList) *PlusPaletteDepth {
	this := &PlusPaletteDepth{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "BGCOLOR"

	//this.NamedParams = []string{ "color", "offset" }
	//this.NamedDefaults = []types.Token{ *types.NewToken( types.INTEGER, "0" ), *types.NewToken( types.INTEGER, "0" ) }
	//this.Raw = true

	this.MaxParams = 2
	this.MinParams = 1

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "color", Default: *types.NewToken(types.NUMBER, "-1")},
			dialect.FunctionParamDef{Name: "depth", Default: *types.NewToken(types.NUMBER, "10")},
		},
	)

	return this
}
