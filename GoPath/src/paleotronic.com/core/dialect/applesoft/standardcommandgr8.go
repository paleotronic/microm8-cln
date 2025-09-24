package applesoft

import (
	"time"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandGR8 struct {
	dialect.Command
}

func (this *StandardCommandGR8) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	apple2helpers.TEXT40(caller)

	apple2helpers.HTab(caller, 1)
	apple2helpers.VTab(caller, 21)
	apple2helpers.ClearToBottom(caller)

	// enable vector layer
	apple2helpers.CubeLayerEnableCustom(caller, true, true, 256, 256)
	apple2helpers.CUBE(caller).Clear()
	apple2helpers.CUBE(caller).Render()
	time.Sleep(100 * time.Millisecond)

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandGR8) Syntax() string {

	/* vars */
	var result string

	result = "GR"

	/* enforce non void return */
	return result

}
