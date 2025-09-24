package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/types"
	//	"paleotronic.com/runestring"
	//"paleotronic.com/log"
)

type PlusDrawBox struct {
	dialect.CoreFunction
	Fill bool
}

func (this *PlusDrawBox) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	if !this.Query {

		// stuff
		xt := this.ValueMap["x"]
		yt := this.ValueMap["y"]
		wt := this.ValueMap["w"]
		ht := this.ValueMap["h"]
		sht := this.ValueMap["shadow"]
		wit := this.ValueMap["window"]

		x := xt.AsInteger()
		y := yt.AsInteger()
		w := wt.AsInteger()
		h := ht.AsInteger()
		shadow := (sht.AsInteger() != 0)
		window := (wit.AsInteger() != 0)

		content := this.ValueMap["content"].Content

		apple2helpers.TextDrawBox(this.Interpreter, x, y, w, h, content, shadow, window)

	}

	this.Stack.Push(types.NewToken(types.STRING, ""))

	return nil
}

func (this *PlusDrawBox) Syntax() string {

	/* vars */
	var result string

	result = "WINDOW.ADD{name,sx,sy,ex,ey}"

	/* enforce non void return */
	return result

}

func (this *PlusDrawBox) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusDrawBox(a int, b int, params types.TokenList) *PlusDrawBox {
	this := &PlusDrawBox{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "WINDOW.ADD"

	this.NamedParams = []string{"x", "y", "w", "h", "content", "shadow", "window"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.NUMBER, "0"),
		*types.NewToken(types.NUMBER, "0"),
		*types.NewToken(types.NUMBER, "40"),
		*types.NewToken(types.NUMBER, "24"),
		*types.NewToken(types.STRING, ""),
		*types.NewToken(types.NUMBER, "0"),
		*types.NewToken(types.NUMBER, "0"),
	}
	this.Raw = true

	return this
}
