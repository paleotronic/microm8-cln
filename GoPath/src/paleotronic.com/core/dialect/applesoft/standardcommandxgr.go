package applesoft

import (
	"paleotronic.com/core/dialect" //"paleotronic.com/core/hires"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandXGR struct {
	dialect.Command
}

func (this *StandardCommandXGR) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	apple2helpers.SuperHiresEnable(caller, false)
	//apple2helpers.CubeLayerEnable(caller, false, false)
	cube := apple2helpers.GETGFX(caller, "CUBE")
	cube.SetActive(false)

	caller.SetCurrentPage("SHR1")
	caller.SetDisplayPage("SHR1")
	apple2helpers.HGRClear(caller)
	apple2helpers.HTab(caller, 1)

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandXGR) Syntax() string {

	/* vars */
	var result string

	result = "XGR"

	/* enforce non void return */
	return result

}
