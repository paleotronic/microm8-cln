package plus

import (
	"paleotronic.com/core/dialect" //	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusRenderHGR struct {
	dialect.CoreFunction
}

func (this *PlusRenderHGR) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	v := settings.VideoMode(params.Shift().AsInteger()) % settings.VM_MAX_MODE
	ov := v + 1%settings.VM_MAX_MODE

	//bus.SyncDo(
	//func() {
	settings.LastRenderModeHGR[this.Interpreter.GetMemIndex()] = ov

	//settings.RenderModeDHGR[this.Interpreter.GetMemIndex()] = v

	this.Interpreter.GetMemoryMap().IntSetHGRRender(this.Interpreter.GetMemIndex(), v)

	this.Interpreter.MarkLayersDirty([]string{"HGR1", "HGR2", "XGR1", "XGR2"})
	//time.Sleep(1 * time.Second)
	//},
	//)

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(0)))

	return nil
}

func (this *PlusRenderHGR) Syntax() string {

	/* vars */
	var result string

	result = "CGCOLOR{col}"

	/* enforce non void return */
	return result

}

func (this *PlusRenderHGR) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusRenderHGR(a int, b int, params types.TokenList) *PlusRenderHGR {
	this := &PlusRenderHGR{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "CGCOLOR"

	//this.NamedParams = []string{ "color" }
	//this.NamedDefaults = []types.Token{ *types.NewToken( types.INTEGER, "15" ) }
	//this.Raw = true
	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "mode", Default: *types.NewToken(types.NUMBER, "1")},
		},
	)

	return this
}
