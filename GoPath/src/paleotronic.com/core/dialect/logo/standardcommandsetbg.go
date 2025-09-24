package logo

import (
	"errors"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandSETBG struct {
	dialect.Command
}

func (this *StandardCommandSETBG) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

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
	vctr := apple2helpers.GETGFX(caller, "VCTR")
	i := tt.AsInteger() % vctr.GetPaletteSize()
	c := vctr.GetPaletteColor(i)
	caller.GetMemoryMap().SetBGColor(caller.GetMemIndex(), c.Red, c.Green, c.Blue, c.Alpha)

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandSETBG) Syntax() string {

	/* vars */
	var result string

	result = "TEXT"

	/* enforce non void return */
	return result

}
