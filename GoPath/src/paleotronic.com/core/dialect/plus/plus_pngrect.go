package plus

import (
	"image"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	//	"paleotronic.com/runestring"
	//"paleotronic.com/log"
)

type PlusPNGRect struct {
	dialect.CoreFunction
	Fill bool
}

func (this *PlusPNGRect) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	if !this.Query {

		// stuff
		xt := this.ValueMap["x0"]
		yt := this.ValueMap["y0"]
		wt := this.ValueMap["x1"]
		ht := this.ValueMap["y1"]

		x0 := xt.AsInteger()
		y0 := yt.AsInteger()
		x1 := wt.AsInteger()
		y1 := ht.AsInteger()

		r := image.Rect(x0, y0, x1, y1)

		if x0 != -1 {
			settings.ImageDrawRect[this.Interpreter.GetMemIndex()] = &r
		} else {
			settings.ImageDrawRect[this.Interpreter.GetMemIndex()] = nil
		}

	}

	this.Stack.Push(types.NewToken(types.STRING, ""))

	return nil
}

func (this *PlusPNGRect) Syntax() string {

	/* vars */
	var result string

	result = "WINDOW.ADD{name,sx,sy,ex,ey}"

	/* enforce non void return */
	return result

}

func (this *PlusPNGRect) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusPNGRect(a int, b int, params types.TokenList) *PlusPNGRect {
	this := &PlusPNGRect{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "WINDOW.ADD"

	this.NamedParams = []string{"x0", "y0", "x1", "y1"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.NUMBER, "-1"),
		*types.NewToken(types.NUMBER, "-1"),
		*types.NewToken(types.NUMBER, "-1"),
		*types.NewToken(types.NUMBER, "-1"),
	}
	this.Raw = true

	return this
}

func NewPlusPNGRectH(a int, b int, params types.TokenList) *PlusPNGRect {
	this := &PlusPNGRect{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "WINDOW.ADD"
	this.Hidden = true

	this.NamedParams = []string{"x0", "y0", "x1", "y1"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.NUMBER, "-1"),
		*types.NewToken(types.NUMBER, "-1"),
		*types.NewToken(types.NUMBER, "-1"),
		*types.NewToken(types.NUMBER, "-1"),
	}
	this.Raw = true

	return this
}
