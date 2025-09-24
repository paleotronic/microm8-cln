package logo

import (
	"errors"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandPENOPACITY struct {
	dialect.Command
}

func (this *StandardCommandPENOPACITY) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	if tokens.Size() < 1 {
		return result, errors.New("I NEED A VALUE")
	}

	d := this.Command.D.(*DialectLogo)

	// get result
	tt, e := d.ParseTokensForResult(caller, tokens)
	if e != nil {
		return result, e
	}

	//cx := caller.GetVDU().GetCursorX()
	apple2helpers.VECTOR(caller).GetTurtle(this.Command.D.(*DialectLogo).Driver.GetTurtle()).SetPenOpacity(int32(tt.AsInteger()))
	apple2helpers.VECTOR(caller).Render()

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandPENOPACITY) Syntax() string {

	/* vars */
	var result string

	result = "TEXT"

	/* enforce non void return */
	return result

}
