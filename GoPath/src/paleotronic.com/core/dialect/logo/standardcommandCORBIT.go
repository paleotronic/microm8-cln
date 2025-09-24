package logo

import (
	"errors"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types" //	"paleotronic.com/core/hardware/apple2helpers"
)

type StandardCommandCORBIT struct {
	dialect.Command
}

func (this *StandardCommandCORBIT) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

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

	if tt.Type != types.LIST || tt.List.Size() < 2 {
		return result, errors.New("I NEED 2 VALUES")
	}

	x := tt.List.Shift().AsExtended()
	y := tt.List.Shift().AsExtended()

	index := caller.GetMemIndex()
	mm := caller.GetMemoryMap()
	cindex := 1
	control := types.NewOrbitController(mm, index, cindex)
	control.Orbit(x, y)
	control.Update()

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandCORBIT) Syntax() string {

	/* vars */
	var result string

	result = "TEXT"

	/* enforce non void return */
	return result

}
