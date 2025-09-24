package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusSpriteStamp struct {
	dialect.CoreFunction
	ClearStamp bool
}

func (this *PlusSpriteStamp) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	var tmp types.Token
	tmp = this.ValueMap["sprite"]
	sprite := tmp.AsInteger()
	tmp = this.ValueMap["x"]
	x := tmp.AsInteger()
	tmp = this.ValueMap["y"]
	y := tmp.AsInteger()
	tmp = this.ValueMap["c"]
	oc := tmp.AsInteger()
	controller := GetSpriteController(this.Interpreter)
	var width, height int

	if !this.Query {

		mode := apple2helpers.GetVideoMode(this.Interpreter)
		switch mode {
		case "LOGR":
			width, height = 40, 48
		case "DLGR":
			width, height = 80, 48
		case "HGR1", "HGR2":
			width, height = 280, 192
		case "DHR1", "DHR2":
			width, height = 140, 192
		case "SHR1":
			width, height = 320, 200
		}
		data := controller.GetSpriteData(sprite)
		_, _, _, _, scl, bounds, _ := controller.GetSpriteAttr(sprite)
		l := apple2helpers.GETGFX(this.Interpreter, mode)

		xs := bounds.X
		ys := bounds.Y
		ye := ys + bounds.Size - 1
		xe := xs + bounds.Size - 1
		sval := int(scl) + 1
		for yy := ys * sval; yy < ye*sval; yy++ {
			for xx := xs * sval; xx < xe*sval; xx++ {
				px, py := x+xx, y+yy
				c := int(data[xx/sval][yy/sval])
				if oc != -1 {
					c = oc
				}
				if this.ClearStamp {
					c = 0
				}
				if int(data[xx/sval][yy/sval]) != 0 && px >= 0 && px < width && py >= 0 && py < height {
					mode := apple2helpers.GetVideoMode(this.Interpreter)
					switch mode {
					case "LOGR":
						apple2helpers.LOGRPlot40(this.Interpreter, uint64(px), uint64(py), uint64(c))
					case "DLGR":
						apple2helpers.LOGRPlot80(this.Interpreter, uint64(px), uint64(py), uint64(c))
					case "HGR1", "HGR2":
						l.HControl.Plot(px, py, c)
					case "DHR1", "DHR2":
						l.HControl.Plot(px, py, c)
					case "SHR1":
						l.HControl.Plot(px, py, c)
					}
				}
			}
		}

	}
	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(0)))

	return nil
}

func (this *PlusSpriteStamp) Syntax() string {

	/* vars */
	var result string

	result = "BGCOLOR{col}"

	/* enforce non void return */
	return result

}

func (this *PlusSpriteStamp) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	//result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusSpriteStamp(a int, b int, params types.TokenList) *PlusSpriteStamp {
	this := &PlusSpriteStamp{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "BGCOLOR"

	this.MaxParams = 3
	this.MinParams = 1

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "sprite", Default: *types.NewToken(types.NUMBER, "0")},
			dialect.FunctionParamDef{Name: "x", Default: *types.NewToken(types.NUMBER, "0")},
			dialect.FunctionParamDef{Name: "y", Default: *types.NewToken(types.NUMBER, "0")},
			dialect.FunctionParamDef{Name: "c", Default: *types.NewToken(types.NUMBER, "-1")},
		},
	)

	return this
}

func NewPlusSpriteUnstamp(a int, b int, params types.TokenList) *PlusSpriteStamp {
	this := &PlusSpriteStamp{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "BGCOLOR"

	this.MaxParams = 3
	this.MinParams = 1
	this.ClearStamp = true

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "sprite", Default: *types.NewToken(types.NUMBER, "0")},
			dialect.FunctionParamDef{Name: "x", Default: *types.NewToken(types.NUMBER, "0")},
			dialect.FunctionParamDef{Name: "y", Default: *types.NewToken(types.NUMBER, "0")},
		},
	)

	return this
}
