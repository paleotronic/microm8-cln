package plus

import (
	"paleotronic.com/fmt"

	"paleotronic.com/core/settings"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

var paletteList [settings.NUMSLOTS][]string

func ResetPaletteList(i int) {
	paletteList[i] = []string{"HGR1", "HGR2"}
}

func init() {
	for i := 0; i < settings.NUMSLOTS; i++ {
		ResetPaletteList(i)
	}
}

type PlusPaletteRGBA struct {
	dialect.CoreFunction
}

func (this *PlusPaletteRGBA) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		fmt.Println(e)
		return e
	}

	//if !this.Query {
	c := params.Shift().AsInteger()
	r := params.Shift().AsInteger()
	g := params.Shift().AsInteger()
	b := params.Shift().AsInteger()
	a := 255
	if params.Size() > 0 {
		a = params.Shift().AsInteger()
	}

	// Set Palette offset
	//list := apple2helpers.GetActiveVideoModes(this.Interpreter)
	for _, id := range paletteList[this.Interpreter.GetMemIndex()] {
		apple2helpers.SetColorRGBA(this.Interpreter, id, c, r, g, b, a)
	}
	//}
	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(0)))

	return nil
}

func (this *PlusPaletteRGBA) Syntax() string {

	/* vars */
	var result string

	result = "BGCOLOR{col}"

	/* enforce non void return */
	return result

}

func (this *PlusPaletteRGBA) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusPaletteRGBA(a int, b int, params types.TokenList) *PlusPaletteRGBA {
	this := &PlusPaletteRGBA{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "BGCOLOR"

	this.NamedParams = []string{"color", "red", "green", "blue", "alpha"}
	this.NamedDefaults = []types.Token{*types.NewToken(types.INTEGER, "0"), *types.NewToken(types.INTEGER, "0"), *types.NewToken(types.INTEGER, "0"), *types.NewToken(types.INTEGER, "0"), *types.NewToken(types.INTEGER, "255")}
	this.Raw = true

	this.MinParams = 4
	this.MaxParams = 5

	return this
}
