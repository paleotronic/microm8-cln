package logo

import (
	"errors"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandCHANNELTRANSMIT struct {
	dialect.Command
}

func (this *StandardCommandCHANNELTRANSMIT) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

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

	if tt == nil || tt.Type != types.LIST || tt.List.Size() != 2 {
		return result, errors.New("I NEED 2 VALUES")
	}

	this.Command.D.(*DialectLogo).Driver.ChannelSend(tt.List.Get(0).Content, tt.List.Get(1))

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandCHANNELTRANSMIT) Syntax() string {

	/* vars */
	var result string

	result = "TEXT"

	/* enforce non void return */
	return result

}
