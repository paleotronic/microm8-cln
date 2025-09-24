package logo

import (
	"errors"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/fmt"
)

type StandardCommandQUADXY struct {
	dialect.Command
	solid bool
}

func (this *StandardCommandQUADXY) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	// get result
	tt, e := this.Command.D.(*DialectLogo).ParseTokensForResult(caller, tokens)
	if e != nil {
		return result, e
	}

	if tt.Type != types.LIST || tt.List.Size() != 2 {
		return result, errors.New("I NEED 2 VALUES")
	}

	//cx := caller.GetVDU().GetCursorX()
	apple2helpers.VECTOR(caller).GetTurtle( this.Command.D.(*DialectLogo).Driver.GetTurtle() ).Quad(float32(tt.List.Get(0).AsFloat()), float32(tt.List.Get(1).AsFloat()), this.solid)

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

func (this *StandardCommandQUADXY) Syntax() string {

	/* vars */
	var result string

	result = "TEXT"

	/* enforce non void return */
	return result

}
