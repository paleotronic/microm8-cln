package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusPaletteTint struct {
	dialect.CoreFunction
}

func (this *PlusPaletteTint) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	//if !this.Query {
	if params.Size() == 3 {
		r := uint8(params.Shift().AsInteger())
		g := uint8(params.Shift().AsInteger())
		b := uint8(params.Shift().AsInteger())
		a := uint8(255)
		this.Interpreter.GetMemoryMap().IntSetVideoTintRGBA(this.Interpreter.GetMemIndex(), r, g, b, a)
	} else {
		r := settings.VideoPaletteTint(params.Shift().AsInteger())
		this.Interpreter.GetMemoryMap().IntSetVideoTint(this.Interpreter.GetMemIndex(), r)
	}

	//}
	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(0)))

	return nil
}

func (this *PlusPaletteTint) Syntax() string {

	/* vars */
	var result string

	result = "BGCOLOR{col}"

	/* enforce non void return */
	return result

}

func (this *PlusPaletteTint) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)
	//result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusPaletteTint(a int, b int, params types.TokenList) *PlusPaletteTint {
	this := &PlusPaletteTint{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "BGCOLOR"

	this.MaxParams = 3
	this.MinParams = 1

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "r", Default: *types.NewToken(types.NUMBER, "255")},
			dialect.FunctionParamDef{Name: "g", Default: *types.NewToken(types.NUMBER, "255")},
			dialect.FunctionParamDef{Name: "b", Default: *types.NewToken(types.NUMBER, "255")},
		},
	)

	return this
}
