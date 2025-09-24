package logo

import (
	"errors"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandCHANNEL struct {
	dialect.Command
}

func (this *StandardCommandCHANNEL) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

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

	if tt == nil || tt.Type == types.LIST {
		return result, errors.New("I NEED A VALUE")
	}

	this.Command.D.(*DialectLogo).Driver.ChannelCreate(tt.Content)

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandCHANNEL) Syntax() string {

	/* vars */
	var result string

	result = "TEXT"

	/* enforce non void return */
	return result

}
