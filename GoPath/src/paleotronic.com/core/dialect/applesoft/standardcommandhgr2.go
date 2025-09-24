package applesoft

import (
	"paleotronic.com/core/dialect" //"paleotronic.com/core/hires"
	"paleotronic.com/core/hardware/apple2"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandHGR2 struct {
	dialect.Command
}

func (this *StandardCommandHGR2) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	var result int = 0

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
			apple2.VF_HIRES | apple2.VF_PAGE2,
		)

		caller.SetCurrentPage("HGR2")
		caller.SetDisplayPage("HGR2")

		caller.SetMemory(32, 0)
		caller.SetMemory(33, 40)
		caller.SetMemory(34, 0)
		caller.SetMemory(35, 24)

		apple2helpers.SetRealWindow(caller, 0, 0, 79, 47)

		apple2helpers.HGRClear(caller)

		//apple2helpers.Clearscreen(caller)
		//apple2helpers.SetRealCursorPos(caller, 0, 46)
	}

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandHGR2) Syntax() string {

	/* vars */
	var result string

	result = "HGR2"

	/* enforce non void return */
	return result

}
