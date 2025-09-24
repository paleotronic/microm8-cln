package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandGR4 struct {
	dialect.Command
}

func (this *StandardCommandGR4) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	apple2helpers.TEXT80(caller)
	apple2helpers.Clearscreen(caller)

	apple2helpers.CubeLayerEnable(caller, false, false)
	apple2helpers.VectorLayerEnable(caller, false, false)

	if mr, ok := caller.GetMemoryMap().InterpreterMappableAtAddress(caller.GetMemIndex(), 0xc000); ok {

		io := mr.(*apple2.Apple2IOChip)

		io.SetVidModeForce(
			apple2.VF_DHIRES | apple2.VF_80COL,
		)

		// caller.SetMemory(32, 0)
		// caller.SetMemory(33, 40)
		// caller.SetMemory(34, 20)
		// caller.SetMemory(35, 24)

		//apple2helpers.SetRealWindow(caller, 80, 48, 80, 48)

		apple2helpers.LOGRClear80(caller, 0, 48)

		//apple2helpers.Clearscreen(caller)
		//apple2helpers.SetRealCursorPos(caller, 0, 48)

	}

	caller.SetCurrentPage("DLGR")
	caller.SetDisplayPage("DLGR")

	//apple2helpers.Clearscreen(caller)
	//apple2helpers.SetRealCursorPos(caller, 0, 46)

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandGR4) Syntax() string {

	/* vars */
	var result string

	result = "GR"

	/* enforce non void return */
	return result

}
