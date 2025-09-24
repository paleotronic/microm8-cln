package applesoft

import (
	"paleotronic.com/core/dialect" //"paleotronic.com/core/hires"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandXGR2 struct {
	dialect.Command
}

func (this *StandardCommandXGR2) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	apple2helpers.SuperHiresEnable(caller, true)
	//apple2helpers.CubeLayerDisabcaller, false, false)
	cube := apple2helpers.GETGFX(caller, "CUBE")
	cube.SetActive(false)

	//caller.GetVDU().Clear;
	//caller.GetVDU().Home();
	caller.SetCurrentPage("SHR1")
	caller.SetDisplayPage("SHR1")
	//hires.GetAppleHiRES().HgrFill(caller.GetVDU().GetBitmapMemory()[caller.GetVDU().GetCurrentPage()%2], 0)
	apple2helpers.HGRClear(caller)

	//apple2helpers.HGRDitherBayer4x4(caller, "woz.png")
	apple2helpers.HTab(caller, 1)

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandXGR2) Syntax() string {

	/* vars */
	var result string

	result = "XGR"

	/* enforce non void return */
	return result

}
