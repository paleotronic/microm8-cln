package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers" //	"strings"
	"paleotronic.com/core/types"
)

type PlusDrawLine struct {
	dialect.CoreFunction
}

func (this *PlusDrawLine) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	tmp := this.ValueMap["x0"]
	x0 := tmp.AsInteger()
	tmp = this.ValueMap["y0"]
	y0 := tmp.AsInteger()
	tmp = this.ValueMap["x1"]
	x1 := tmp.AsInteger()
	tmp = this.ValueMap["y1"]
	y1 := tmp.AsInteger()
	tmp = this.ValueMap["c"]
	c := tmp.AsInteger()

	modes := apple2helpers.GetActiveVideoModes(this.Interpreter)
	if len(modes) == 0 {
		return nil
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

	return nil
}

func (this *PlusDrawLine) Syntax() string {

	/* vars */
	var result string

	result = "OPEN{filename}"

	/* enforce non void return */
	return result

}

func (this *PlusDrawLine) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusDrawLine(a int, b int, params types.TokenList) *PlusDrawLine {
	this := &PlusDrawLine{}

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
