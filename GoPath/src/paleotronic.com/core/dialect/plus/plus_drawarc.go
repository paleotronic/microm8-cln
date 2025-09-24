package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers" //	"strings"
	"paleotronic.com/core/types"
)

type PlusDrawArc struct {
	dialect.CoreFunction
}

func (this *PlusDrawArc) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	tmp := this.ValueMap["radius"]
	r := tmp.AsInteger()
	tmp = this.ValueMap["start"]
	s := tmp.AsInteger()
	tmp = this.ValueMap["end"]
	e := tmp.AsInteger()
	tmp = this.ValueMap["x"]
	x := tmp.AsInteger()
	tmp = this.ValueMap["y"]
	y := tmp.AsInteger()
	tmp = this.ValueMap["c"]
	c := tmp.AsInteger()

	modes := apple2helpers.GetActiveVideoModes(this.Interpreter)
	if len(modes) == 0 {
		return nil
	}

	switch modes[0] {
	case "LOGR":
		apple2helpers.Arc(this.Interpreter, r, s, e, x, y, c, apple2helpers.LOGRPlot40)
	case "DLGR":
		apple2helpers.Arc(this.Interpreter, r, s, e, x, y, c, apple2helpers.LOGRPlot80)
	case "HGR1":
		apple2helpers.Arc(this.Interpreter, r, s, e, x, y, c, apple2helpers.HGRPlot)
	case "HGR2":
		apple2helpers.Arc(this.Interpreter, r, s, e, x, y, c, apple2helpers.HGRPlot)
	case "DHR1":
		apple2helpers.Arc(this.Interpreter, r, s, e, x, y, c, apple2helpers.HGRPlot)
	case "XGR1":
		apple2helpers.Arc(this.Interpreter, r, s, e, x, y, c, apple2helpers.HGRPlot)
	case "XGR2":
		apple2helpers.Arc(this.Interpreter, r, s, e, x, y, c, apple2helpers.HGRPlot)
	case "SHR1":
		apple2helpers.Arc(this.Interpreter, r, s, e, x, y, c, apple2helpers.HGRPlot)
	}

	return nil
}

func (this *PlusDrawArc) Syntax() string {

	/* vars */
	var result string

	result = "OPEN{filename}"

	/* enforce non void return */
	return result

}

func (this *PlusDrawArc) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusDrawArc(a int, b int, params types.TokenList) *PlusDrawArc {
	this := &PlusDrawArc{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "OPEN"

	this.NamedParams = []string{"x", "y", "start", "end", "radius", "c"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.NUMBER, "20"),
		*types.NewToken(types.NUMBER, "20"),
		*types.NewToken(types.NUMBER, "90"),
		*types.NewToken(types.NUMBER, "180"),
		*types.NewToken(types.NUMBER, "20"),
		*types.NewToken(types.NUMBER, "1"),
	}
	this.Raw = true

	return this
}
