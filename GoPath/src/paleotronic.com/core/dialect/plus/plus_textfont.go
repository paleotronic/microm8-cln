package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/octalyzer/video/font"
	//	"paleotronic.com/core/memory"
	//	"paleotronic.com/core/settings"
	//	"time"
	//	"paleotronic.com/fmt"
)

type PlusTextFont struct {
	dialect.CoreFunction
}

func (this *PlusTextFont) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	if !this.Query {
		f := this.ValueMap["font"]
		fontid := f.AsInteger()
		index := this.Interpreter.GetMemIndex()
		if fontid >= 0 && fontid < len(settings.AuxFonts[index]) {
			fontName := settings.AuxFonts[index][fontid]
			f, err := font.LoadFromFile(fontName)
			if err == nil {
				settings.DefaultFont[index] = f
				settings.ForceTextVideoRefresh = true
			}
		}
	}

	//fmt.Printf("Setting zoom to %f\n", zoom)

	return nil
}

func (this *PlusTextFont) Syntax() string {

	/* vars */
	var result string

	result = "CAMZOOM{v}"

	/* enforce non void return */
	return result

}

func (this *PlusTextFont) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusTextFont(a int, b int, params types.TokenList) *PlusTextFont {
	this := &PlusTextFont{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "CAMZOOM"
	this.MinParams = 0
	this.MaxParams = 1
	this.Raw = true
	this.NamedParams = []string{"font"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.INTEGER, "0"),
	}
	this.EvalSingleParam = true

	return this
}
