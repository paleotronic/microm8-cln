package applesoft

import (
	"time"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandGR5 struct {
	dialect.Command
}

func (this *StandardCommandGR5) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	apple2helpers.TEXT40(caller)

	apple2helpers.HTab(caller, 1)
	apple2helpers.VTab(caller, 21)
	apple2helpers.ClearToBottom(caller)

	// soft switches
	if mr, ok := caller.GetMemoryMap().InterpreterMappableAtAddress(caller.GetMemIndex(), 0xc000); ok {
		io := mr.(*apple2.Apple2IOChip)
		mode := io.GetVidMode()
		mode &= (apple2.VF_MASK ^ apple2.VF_DHIRES)
		mode &= (apple2.VF_MASK ^ apple2.VF_HIRES)
		mode |= apple2.VF_TEXT
		io.SetVidModeForce(
			mode,
		)
	}

	// enable vector layer
	apple2helpers.SuperHiresDisable(caller, false)
	apple2helpers.CubeLayerEnableCustom(caller, true, true, 48, 48)
	apple2helpers.CUBE(caller).Clear()
	apple2helpers.CUBE(caller).Render()
	time.Sleep(100 * time.Millisecond)

	caller.SetCurrentPage("CUBE")
	caller.SetDisplayPage("CUBE")

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandGR5) Syntax() string {

	/* vars */
	var result string

	result = "GR"

	/* enforce non void return */
	return result

}
