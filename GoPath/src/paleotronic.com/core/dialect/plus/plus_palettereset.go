package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusPaletteReset struct {
	dialect.CoreFunction
}

func (this *PlusPaletteReset) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	for _, p := range paletteList[this.Interpreter.GetMemIndex()] {
		vp, e := hardware.LoadSpecPaletteData(this.Interpreter, settings.SpecFile[this.Interpreter.GetMemIndex()], p)
		if e == nil {
			ls, ok := this.Interpreter.GetGFXLayerByID(p)
			if ok && ls != nil {
				ls.SetPalette(*vp)
			}
		}
	}

	for _, p := range paletteTextList {
		vp, e := hardware.LoadSpecPaletteData(this.Interpreter, settings.SpecFile[this.Interpreter.GetMemIndex()], p)
		if e == nil {
			ls, ok := this.Interpreter.GetHUDLayerByID(p)
			if ok && ls != nil {
				ls.SetPalette(*vp)
			}
		}
	}

	//}
	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(0)))

	return nil
}

func (this *PlusPaletteReset) Syntax() string {

	/* vars */
	var result string

	result = "BGCOLOR{col}"

	/* enforce non void return */
	return result

}

func (this *PlusPaletteReset) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	//result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusPaletteReset(a int, b int, params types.TokenList) *PlusPaletteReset {
	this := &PlusPaletteReset{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "BGCOLOR"

	this.MaxParams = 1
	this.MinParams = 0

	return this
}
