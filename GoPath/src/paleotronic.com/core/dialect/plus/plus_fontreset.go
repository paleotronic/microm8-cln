package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusFontReset struct {
	dialect.CoreFunction
}

func (this *PlusFontReset) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		//log.Printf("@text.reset -> %v", e)
		return e
	}

	if !this.Query {
		apple2helpers.SetFGColor(this.Interpreter, 15)
		apple2helpers.SetBGColor(this.Interpreter, 0)
		apple2helpers.SetTextSize(this.Interpreter, 5)
		var huds = []string{"TEXT", "TXT2"}
		for _, hud := range huds {
			//log.Printf("resetting palette for HUD %s", hud)
			vp, e := hardware.LoadSpecPaletteData(this.Interpreter, settings.SpecFile[this.Interpreter.GetMemIndex()], hud)
			//log.Printf("load result: e = %v", e)
			if e == nil {
				ls, ok := this.Interpreter.GetHUDLayerByID(hud)
				if ok && ls != nil {
					ls.SetPalette(*vp)
					ls.SetDirty(true)
				}
			}
		}
	}
	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(0)))

	return nil
}

func (this *PlusFontReset) Syntax() string {

	/* vars */
	var result string

	result = "FGCOLOR{col}"

	/* enforce non void return */
	return result

}

func (this *PlusFontReset) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusFontReset(a int, b int, params types.TokenList) *PlusFontReset {
	this := &PlusFontReset{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "FGCOLOR"

	this.NamedParams = []string{"color"}
	this.NamedDefaults = []types.Token{*types.NewToken(types.INTEGER, "15")}
	this.Raw = true

	return this
}
