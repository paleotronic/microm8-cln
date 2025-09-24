package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandGR3 struct {
	dialect.Command
}

func (this *StandardCommandGR3) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	//log.Printf("GR3 called")

	apple2helpers.TEXT80(caller)
	apple2helpers.Clearscreen(caller)
	apple2helpers.CubeLayerEnable(caller, false, false)
	apple2helpers.VectorLayerEnable(caller, false, false)
	if mr, ok := caller.GetMemoryMap().InterpreterMappableAtAddress(caller.GetMemIndex(), 0xc000); ok {

		io := mr.(*apple2.Apple2IOChip)

		io.SetVidModeForce(
			apple2.VF_TEXT,
		)
		io.SetVidModeForce(
			apple2.VF_MIXED | apple2.VF_80COL | apple2.VF_DHIRES,
		)

		caller.SetMemory(32, 0)
		caller.SetMemory(33, 80)
		caller.SetMemory(34, 20)
		caller.SetMemory(35, 24)

		apple2helpers.SetRealWindow(caller, 0, 40, 79, 47)

		apple2helpers.LOGRClear80(caller, 0, 40)

		//apple2helpers.Clearscreen(caller)
		apple2helpers.SetRealCursorPos(caller, 0, 46)
	}

	caller.SetCurrentPage("DLGR")
	caller.SetDisplayPage("DLGR")

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandGR3) Syntax() string {

	/* vars */
	var result string

	result = "GR"

	/* enforce non void return */
	return result

}
