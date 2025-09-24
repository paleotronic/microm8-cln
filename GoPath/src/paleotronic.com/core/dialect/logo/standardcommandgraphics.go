package logo

import (
	"paleotronic.com/core/dialect"
	//"paleotronic.com/core/hires"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandGRAPHICS struct {
	dialect.Command
	Split bool
}

func (this *StandardCommandGRAPHICS) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	//	caller.SetMemory(65536-16304,0)	// GFX
	//	caller.SetMemory(65536-16297,0) // Hires
	//	caller.SetMemory(65536-16300,0) // Page 1
	//	caller.SetMemory(65536-16301,0) // Splitscreen

	//	caller.SetCurrentPage("HGR1")
	//	caller.SetDisplayPage("HGR1")
	//	apple2helpers.HGRClear(caller)

	//apple2helpers.GRAPHICSDitherBayer4x4(caller, "woz.png")
	apple2helpers.HTab(caller, 1)
	apple2helpers.VTab(caller, 21)
	apple2helpers.ClearToBottom(caller)

	// enable vector layer
	apple2helpers.VectorLayerEnable(caller, true, this.Split)

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandGRAPHICS) Syntax() string {

	/* vars */
	var result string

	result = "GRAPHICS"

	/* enforce non void return */
	return result

}
