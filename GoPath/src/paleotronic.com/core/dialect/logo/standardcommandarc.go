package logo

import (
	"errors"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/fmt"
)

type StandardCommandARC struct {
	dialect.Command
	solid   bool
	percent bool
}

func (this *StandardCommandARC) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

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

	p := float32(tt.List.Get(1).AsFloat())
	if this.percent {
		p = (p / 100) * 360
	}

	//cx := caller.GetVDU().GetCursorX()
	apple2helpers.VECTOR(caller).GetTurtle(this.Command.D.(*DialectLogo).Driver.GetTurtle()).Arc(float32(tt.List.Get(0).AsFloat()), p, this.solid)

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

func (this *StandardCommandARC) Syntax() string {

	/* vars */
	var result string

	result = "TEXT"

	/* enforce non void return */
	return result

}
