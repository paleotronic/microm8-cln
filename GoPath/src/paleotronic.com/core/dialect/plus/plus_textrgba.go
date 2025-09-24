package plus

import (
	"paleotronic.com/fmt"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

var paletteTextList = []string{"TEXT", "TXT2"}

type PlusTextRGBA struct {
	dialect.CoreFunction
}

func (this *PlusTextRGBA) FunctionExecute(params *types.TokenList) error {

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
	for _, id := range paletteTextList {
		apple2helpers.SetTextColorRGBA(this.Interpreter, id, c, r, g, b, a)
	}
	//}
	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(0)))

	return nil
}

func (this *PlusTextRGBA) Syntax() string {

	/* vars */
	var result string

	result = "BGCOLOR{col}"

	/* enforce non void return */
	return result

}

func (this *PlusTextRGBA) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusTextRGBA(a int, b int, params types.TokenList) *PlusTextRGBA {
	this := &PlusTextRGBA{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "BGCOLOR"

	//this.NamedParams = []string{ "color", "red", "green", "blue", "alpha" }
	//this.NamedDefaults = []types.Token{ *types.NewToken( types.INTEGER, "0" ), *types.NewToken( types.INTEGER, "0" ), *types.NewToken( types.INTEGER, "0"), *types.NewToken( types.INTEGER, "0"), *types.NewToken( types.INTEGER, "255") }
	//this.Raw = true

	this.MinParams = 4
	this.MaxParams = 5

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "c", Default: *types.NewToken(types.NUMBER, "15")},
			dialect.FunctionParamDef{Name: "r", Default: *types.NewToken(types.NUMBER, "255")},
			dialect.FunctionParamDef{Name: "g", Default: *types.NewToken(types.NUMBER, "255")},
			dialect.FunctionParamDef{Name: "b", Default: *types.NewToken(types.NUMBER, "255")},
			dialect.FunctionParamDef{Name: "a", Default: *types.NewToken(types.NUMBER, "255")},
		},
	)

	return this
}
