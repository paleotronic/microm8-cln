package logo

import (
	"errors"

	"paleotronic.com/fmt"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandSHUFR struct {
	dialect.Command
}

func (this *StandardCommandSHUFR) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	if tokens.Size() < 1 {
		return result, errors.New("I NEED A VALUE")
	}

	// get result
	tt, e := this.Command.D.(*DialectLogo).ParseTokensForResult(caller, tokens)
	if e != nil {
		return result, e
	}

	//cx := caller.GetVDU().GetCursorX()
	apple2helpers.VECTOR(caller).GetTurtle(this.Command.D.(*DialectLogo).Driver.GetTurtle()).ShuffleRight(float32(tt.AsFloat()))

	if apple2helpers.IsVectorLayerEnabled(caller) == false {
		apple2helpers.HTab(caller, 1)
		apple2helpers.VTab(caller, 21)
		apple2helpers.ClearToBottom(caller)

		// enable vector layer
		apple2helpers.VectorLayerEnable(caller, true, true)
	}

	e = apple2helpers.VECTOR(caller).Render()

	if e != nil {
		fmt.Println("ERROR: " + e.Error())
		return result, e
	}

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandSHUFR) Syntax() string {

	/* vars */
	var result string

	result = "TEXT"

	/* enforce non void return */
	return result

}
