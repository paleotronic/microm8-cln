package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	//	"paleotronic.com/core/memory"
	//	"paleotronic.com/core/settings"
	//	"time"
	//	"paleotronic.com/fmt"
)

type PlusTextDrawFont struct {
	dialect.CoreFunction
}

func (this *PlusTextDrawFont) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	if !this.Query {
		f := this.ValueMap["font"]
		fontid := f.AsInteger()
		index := this.Interpreter.GetMemIndex()
		if fontid >= 0 && fontid < len(settings.AuxFonts[index]) {
			apple2helpers.LoadGraphicsFont(index, fontid)
		}
	}

	//fmt.Printf("Setting zoom to %f\n", zoom)

	return nil
}

func (this *PlusTextDrawFont) Syntax() string {

	/* vars */
	var result string

	result = "CAMZOOM{v}"

	/* enforce non void return */
	return result

}

func (this *PlusTextDrawFont) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusTextDrawFont(a int, b int, params types.TokenList) *PlusTextDrawFont {
	this := &PlusTextDrawFont{}

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
