package logo

import (
	"errors"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandSETY struct {
	dialect.Command
}

func (this *StandardCommandSETY) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

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
	apple2helpers.VECTOR(caller).GetTurtle( this.Command.D.(*DialectLogo).Driver.GetTurtle() ).AddCommand(types.TURTLE_POSY, float32(tt.AsExtended()), 0)
	apple2helpers.VECTOR(caller).Render()

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandSETY) Syntax() string {

	/* vars */
	var result string

	result = "TEXT"

	/* enforce non void return */
	return result

}
