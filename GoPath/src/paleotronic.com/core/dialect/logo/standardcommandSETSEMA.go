package logo

import (
	"errors"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandSEMA struct {
	dialect.Command
}

func (this *StandardCommandSEMA) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	t, _ := this.Command.D.(*DialectLogo).ParseTokensForResult(caller, tokens)

	if t == nil {
		return result, errors.New("I NEED A VALUE")
	}

	sema := t.Content

	/* enforce non void return */
	caller.SetSemaphore(sema)

	return result, nil

}

func (this *StandardCommandSEMA) Syntax() string {

	/* vars */
	var result string

	result = "TEXT"

	/* enforce non void return */
	return result

}
