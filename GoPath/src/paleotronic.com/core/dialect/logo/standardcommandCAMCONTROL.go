package logo

import (
	"errors"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types" //	"paleotronic.com/core/hardware/apple2helpers"
)

type StandardCommandCAMCONTROL struct {
	dialect.Command
}

func (this *StandardCommandCAMCONTROL) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	v, err := this.Command.D.ParseTokensForResult(caller, tokens)
	if err != nil {
		return result, err
	}
	if v == nil || v.Type == types.LIST {
		return result, errors.New("I NEED A VALUE")
	}

	settings.LogoCameraControl[caller.GetMemIndex()] = (v.AsInteger() != 0)

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandCAMCONTROL) Syntax() string {

	/* vars */
	var result string

	result = "TEXT"

	/* enforce non void return */
	return result

}
