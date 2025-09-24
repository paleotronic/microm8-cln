package applesoft

import (
	"paleotronic.com/core/dialect" //"paleotronic.com/core/hires"
	"paleotronic.com/core/hardware/apple2"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandHGR struct {
	dialect.Command
}

func (this *StandardCommandHGR) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	// caller.SetMemory(65536-16304, 0) // GFX
	// caller.SetMemory(65536-16297, 0) // Hires
	// caller.SetMemory(65536-16300, 0) // Page 1
	// caller.SetMemory(65536-16301, 0) // Splitscreen
	// apple2helpers.SuperHiresDisable(caller, false)

	// //caller.GetVDU().Clear;
	// //caller.GetVDU().Home();
	// caller.SetCurrentPage("HGR1")
	// caller.SetDisplayPage("HGR1")
	// //hires.GetAppleHiRES().HgrFill(caller.GetVDU().GetBitmapMemory()[caller.GetVDU().GetCurrentPage()%2], 0)
	// apple2helpers.HGRClear(caller)

	// //apple2helpers.HGRDitherBayer4x4(caller, "woz.png")
	// apple2helpers.HTab(caller, 1)

	apple2helpers.TEXT40(caller)
	apple2helpers.Clearscreen(caller)
	apple2helpers.CubeLayerEnable(caller, false, false)
	apple2helpers.VectorLayerEnable(caller, false, false)
	if mr, ok := caller.GetMemoryMap().InterpreterMappableAtAddress(caller.GetMemIndex(), 0xc000); ok {

		io := mr.(*apple2.Apple2IOChip)

		io.SetVidModeForce(
			apple2.VF_TEXT,
		)
		io.SetVidModeForce(
			apple2.VF_MIXED | apple2.VF_HIRES,
		)

		caller.SetCurrentPage("HGR1")
		caller.SetDisplayPage("HGR1")

		caller.SetMemory(32, 0)
		caller.SetMemory(33, 40)
		caller.SetMemory(34, 0)
		caller.SetMemory(35, 24)

		apple2helpers.SetRealWindow(caller, 0, 40, 79, 47)

		apple2helpers.HGRClear(caller)

		//apple2helpers.Clearscreen(caller)
		//apple2helpers.SetRealCursorPos(caller, 0, 46)
	}

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandHGR) Syntax() string {

	/* vars */
	var result string

	result = "HGR"

	/* enforce non void return */
	return result

}
