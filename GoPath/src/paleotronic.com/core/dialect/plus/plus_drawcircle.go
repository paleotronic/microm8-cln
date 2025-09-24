package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces" //	"strings"
	"paleotronic.com/core/types"
)

type PlusDrawCircle struct {
	dialect.CoreFunction
	Fill bool
}

func (this *PlusDrawCircle) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	tmp := this.ValueMap["radius"]
	r := tmp.AsInteger()
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
	var w, h int

	switch modes[0] {
	case "LOGR":
		apple2helpers.Circle(this.Interpreter, r, x, y, c, apple2helpers.LOGRPlot40)
		w, h = 39, 47
	case "DLGR":
		apple2helpers.Circle(this.Interpreter, r, x, y, c, apple2helpers.LOGRPlot80)
		w, h = 79, 47
	case "HGR1":
		apple2helpers.Circle(this.Interpreter, r, x, y, c, apple2helpers.HGRPlot)
		w, h = 279, 191
	case "HGR2":
		apple2helpers.Circle(this.Interpreter, r, x, y, c, apple2helpers.HGRPlot)
		w, h = 279, 191
	case "DHR1":
		apple2helpers.Circle(this.Interpreter, r, x, y, c, apple2helpers.HGRPlot)
		w, h = 139, 191
	case "XGR1":
		apple2helpers.Circle(this.Interpreter, r, x, y, c, apple2helpers.HGRPlot)
	case "XGR2":
		apple2helpers.Circle(this.Interpreter, r, x, y, c, apple2helpers.HGRPlot)
	case "SHR1":
		apple2helpers.Circle(this.Interpreter, r, x, y, c, apple2helpers.HGRPlot)
		w, h = 319, 199
	}

	if this.Fill {
		var get func(ent interfaces.Interpretable, x, y uint64) uint64
		var plot func(ent interfaces.Interpretable, x, y, c uint64)

		switch modes[0] {
		case "LOGR":
			plot = apple2helpers.LOGRPlot40
			get = apple2helpers.GR40At
		case "DLGR":
			plot = apple2helpers.LOGRPlot80
			get = apple2helpers.GR80At
		default:
			plot = apple2helpers.HGRPlot
			get = apple2helpers.HGRAt
		}

		apple2helpers.FloodFill(this.Interpreter, x, y, c, w, h, get, plot)
	}

	return nil
}

func (this *PlusDrawCircle) Syntax() string {

	/* vars */
	var result string

	result = "OPEN{filename}"

	/* enforce non void return */
	return result

}

func (this *PlusDrawCircle) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusDrawCircle(a int, b int, params types.TokenList) *PlusDrawCircle {
	this := &PlusDrawCircle{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "OPEN"

	this.NamedParams = []string{"x", "y", "radius", "c"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.NUMBER, "20"),
		*types.NewToken(types.NUMBER, "20"),
		*types.NewToken(types.NUMBER, "20"),
		*types.NewToken(types.NUMBER, "1"),
	}
	this.Raw = true

	return this
}

func NewPlusDrawCircleF(a int, b int, params types.TokenList) *PlusDrawCircle {
	this := &PlusDrawCircle{Fill: true}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "OPEN"

	this.NamedParams = []string{"x", "y", "radius", "c"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.NUMBER, "20"),
		*types.NewToken(types.NUMBER, "20"),
		*types.NewToken(types.NUMBER, "20"),
		*types.NewToken(types.NUMBER, "1"),
	}
	this.Raw = true

	return this
}
