package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces" //	"strings"
	"paleotronic.com/core/types"
)

type PlusDrawRect struct {
	dialect.CoreFunction
	Fill bool
}

func (this *PlusDrawRect) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	tmp := this.ValueMap["x0"]
	sx := tmp.AsInteger()
	tmp = this.ValueMap["y0"]
	sy := tmp.AsInteger()
	tmp = this.ValueMap["x1"]
	ex := tmp.AsInteger()
	tmp = this.ValueMap["y1"]
	ey := tmp.AsInteger()
	tmp = this.ValueMap["c"]
	c := tmp.AsInteger()

	modes := apple2helpers.GetActiveVideoModes(this.Interpreter)
	if len(modes) == 0 {
		return nil
	}

	var x0, y0, x1, y1 int

	for i := 0; i < 4; i++ {

		switch i {
		case 0:
			x0, y0 = sx, sy
			x1, y1 = ex, sy
		case 1:
			x0, y0 = ex, sy
			x1, y1 = ex, ey
		case 2:
			x0, y0 = ex, ey
			x1, y1 = sx, ey
		case 3:
			x0, y0 = sx, ey
			x1, y1 = sx, sy
		}

		switch modes[0] {
		case "LOGR":
			apple2helpers.BrenshamLine(this.Interpreter, x0, y0, x1, y1, c, apple2helpers.LOGRPlot40)
		case "DLGR":
			apple2helpers.BrenshamLine(this.Interpreter, x0, y0, x1, y1, c, apple2helpers.LOGRPlot80)
		case "HGR1":
			apple2helpers.BrenshamLine(this.Interpreter, x0, y0, x1, y1, c, apple2helpers.HGRPlot)
		case "HGR2":
			apple2helpers.BrenshamLine(this.Interpreter, x0, y0, x1, y1, c, apple2helpers.HGRPlot)
		case "DHR1":
			apple2helpers.BrenshamLine(this.Interpreter, x0, y0, x1, y1, c, apple2helpers.HGRPlot)
		case "XGR1":
			apple2helpers.BrenshamLine(this.Interpreter, x0, y0, x1, y1, c, apple2helpers.HGRPlot)
		case "XGR2":
			apple2helpers.BrenshamLine(this.Interpreter, x0, y0, x1, y1, c, apple2helpers.HGRPlot)
		case "SHR1":
			apple2helpers.BrenshamLine(this.Interpreter, x0, y0, x1, y1, c, apple2helpers.HGRPlot)
		}

	}

	//log2.Printf("fill = %v", this.Fill)
	if this.Fill {
		//var get func(ent interfaces.Interpretable, x, y uint64) uint64
		var plot func(ent interfaces.Interpretable, x, y, c uint64)

		modes := apple2helpers.GetActiveVideoModes(this.Interpreter)
		switch modes[0] {
		case "LOGR":
			plot = apple2helpers.LOGRPlot40
			//get = apple2helpers.GR40At
		case "DLGR":
			plot = apple2helpers.LOGRPlot80
			//get = apple2helpers.GR80At
		default:
			plot = apple2helpers.HGRPlot
			//get = apple2helpers.HGRAt
		}

		//log2.Printf("lines from %d,%d - %d,%d", sx, sy, ex, ey)
		if sy < ey {
			for y := sy; y <= ey; y++ {
				apple2helpers.BrenshamLine(this.Interpreter, sx, y, ex, y, c, plot)
			}
		} else {
			for y := ey; y <= sy; y++ {
				apple2helpers.BrenshamLine(this.Interpreter, sx, y, ex, y, c, plot)
			}
		}

		// x := (x1 + x0) / 2
		// y := (y1 + y0) / 2

		// apple2helpers.FloodFill(this.Interpreter, x, y, c, get, plot)
	}

	return nil
}

func (this *PlusDrawRect) Syntax() string {

	/* vars */
	var result string

	result = "OPEN{filename}"

	/* enforce non void return */
	return result

}

func (this *PlusDrawRect) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusDrawRect(a int, b int, params types.TokenList) *PlusDrawRect {
	this := &PlusDrawRect{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "OPEN"

	this.NamedParams = []string{"x0", "y0", "x1", "y1", "c"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.NUMBER, "0"),
		*types.NewToken(types.NUMBER, "0"),
		*types.NewToken(types.NUMBER, "0"),
		*types.NewToken(types.NUMBER, "0"),
		*types.NewToken(types.NUMBER, "0"),
	}
	this.Raw = true

	return this
}

func NewPlusDrawRectF(a int, b int, params types.TokenList) *PlusDrawRect {
	this := &PlusDrawRect{Fill: true}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "OPEN"

	this.NamedParams = []string{"x0", "y0", "x1", "y1", "c"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.NUMBER, "0"),
		*types.NewToken(types.NUMBER, "0"),
		*types.NewToken(types.NUMBER, "0"),
		*types.NewToken(types.NUMBER, "0"),
		*types.NewToken(types.NUMBER, "0"),
	}
	this.Raw = true

	return this
}
