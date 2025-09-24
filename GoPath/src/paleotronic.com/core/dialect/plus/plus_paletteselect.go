package plus

import (
	"strings"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusPaletteSelect struct {
	dialect.CoreFunction
}

func (this *PlusPaletteSelect) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	mode := params.Shift().Content
	switch strings.ToUpper(mode) {
	case "HGR":
		paletteList[this.Interpreter.GetMemIndex()] = []string{"HGR1", "HGR2"}
		apple2helpers.SetZoneFormat(this.Interpreter, types.LF_HGR_WOZ, "HGR1")
	case "SHR":
		paletteList[this.Interpreter.GetMemIndex()] = []string{"SHR1"}
		apple2helpers.SetZoneFormat(this.Interpreter, types.LF_SUPER_HIRES, "SHR1")
	case "GR":
		paletteList[this.Interpreter.GetMemIndex()] = []string{"LOGR", "LGR2", "DLGR"}
		apple2helpers.SetZoneFormat(this.Interpreter, types.LF_LOWRES_WOZ, "LOGR")
	case "DGR":
		paletteList[this.Interpreter.GetMemIndex()] = []string{"DLGR"}
		apple2helpers.SetZoneFormat(this.Interpreter, types.LF_LOWRES_WOZ, "DLGR")
	case "DHGR":
		paletteList[this.Interpreter.GetMemIndex()] = []string{"DHR1", "DHR2"}
		apple2helpers.SetZoneFormat(this.Interpreter, types.LF_DHGR_WOZ, "DHR1")
	case "VECTOR":
		paletteList[this.Interpreter.GetMemIndex()] = []string{"VCTR"}
		apple2helpers.SetZoneFormat(this.Interpreter, types.LF_VECTOR, "VCTR")
	case "ZX":
		paletteList[this.Interpreter.GetMemIndex()] = []string{"SCRN", "SCR2"}
		apple2helpers.SetZoneFormat(this.Interpreter, types.LF_SPECTRUM_0, "SCRN")
	case "CUBE":
		paletteList[this.Interpreter.GetMemIndex()] = []string{"CUBE"}
		apple2helpers.SetZoneFormat(this.Interpreter, types.LF_CUBE_PACKED, "CUBE")
	case "TEXT":
		paletteTextList[this.Interpreter.GetMemIndex()] = "TEXT"
		//apple2helpers.SetZoneFormat(this.Interpreter, types.LF_CUBE_PACKED, "CUBE")
	default:
		this.Interpreter.PutStr("Invalid mode: " + mode + "\r\n")
	}

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(0)))

	return nil
}

func (this *PlusPaletteSelect) Syntax() string {

	/* vars */
	var result string

	result = "BGCOLOR{col}"

	/* enforce non void return */
	return result

}

func (this *PlusPaletteSelect) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusPaletteSelect(a int, b int, params types.TokenList) *PlusPaletteSelect {
	this := &PlusPaletteSelect{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "BGCOLOR"

	this.MaxParams = 1
	this.MinParams = 1

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "palette", Default: *types.NewToken(types.STRING, "HGR")},
		},
	)

	return this
}
