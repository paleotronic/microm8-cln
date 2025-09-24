package logo

import (
	"errors"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandSETHOME struct {
	dialect.Command
}

func (this *StandardCommandSETHOME) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	d := this.Command.D.(*DialectLogo)
	v, e := d.ParseTokensForResult(caller, tokens)
	if e != nil {
		return result, e
	}

	if v.Type != types.LIST && v.List.Size() < 3 {
		return result, errors.New("I NEED A LIST OF 3")
	}

	x := v.List.Get(0).AsExtended()
	y := v.List.Get(1).AsExtended()
	z := v.List.Get(2).AsExtended()

	/* enforce non void return */
	apple2helpers.VECTOR(caller).GetTurtle(this.Command.D.(*DialectLogo).Driver.GetTurtle()).SetHome(x, y, z)

	return result, nil

}

func (this *StandardCommandSETHOME) Syntax() string {

	/* vars */
	var result string

	result = "TEXT"

	/* enforce non void return */
	return result

}
