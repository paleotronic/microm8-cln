package plus

import (
	"strings"

	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	//	"strings"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
)

type PlusFillPoint struct {
	dialect.CoreFunction
}

func (this *PlusFillPoint) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	tmp := this.ValueMap["x"]
	x := tmp.AsInteger()
	tmp = this.ValueMap["y"]
	y := tmp.AsInteger()
	tmp = this.ValueMap["c"]
	c := tmp.AsInteger()

	var w, h int

	var get func(ent interfaces.Interpretable, x, y uint64) uint64
	var plot func(ent interfaces.Interpretable, x, y, c uint64)

	modes := apple2helpers.GetActiveVideoModes(this.Interpreter)
	switch modes[0] {
	case "LOGR":
		plot = apple2helpers.LOGRPlot40
		get = apple2helpers.GR40At
		w, h = 39, 47
	case "DLGR":
		plot = apple2helpers.LOGRPlot80
		get = apple2helpers.GR80At
		w, h = 79, 47
	default:
		switch {
		case strings.HasPrefix(modes[0], "HGR"):
			w, h = 279, 191
		case strings.HasPrefix(modes[0], "DHR"):
			w, h = 139, 191
		case strings.HasPrefix(modes[0], "SHR"):
			w, h = 319, 199
		}
		plot = apple2helpers.HGRPlot
		get = apple2helpers.HGRAt
	}

	apple2helpers.FloodFill(this.Interpreter, x, y, c, w, h, get, plot)

	return nil
}

func (this *PlusFillPoint) Syntax() string {

	/* vars */
	var result string

	result = "OPEN{filename}"

	/* enforce non void return */
	return result

}

func (this *PlusFillPoint) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusFillPoint(a int, b int, params types.TokenList) *PlusFillPoint {
	this := &PlusFillPoint{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "OPEN"

	this.NamedParams = []string{"x", "y", "c"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.NUMBER, "0"),
		*types.NewToken(types.NUMBER, "0"),
		*types.NewToken(types.NUMBER, "1"),
	}
	this.Raw = true

	return this
}
