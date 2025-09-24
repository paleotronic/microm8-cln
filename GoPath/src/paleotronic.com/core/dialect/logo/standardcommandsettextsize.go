package logo

import (
	"errors"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandSETTEXTSIZE struct {
	dialect.Command
}

func (this *StandardCommandSETTEXTSIZE) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	// get result
	tt, e := this.Command.D.(*DialectLogo).ParseTokensForResult(caller, tokens)
	if e != nil {
		return result, e
	}

	if tt.Type == types.LIST || !tt.IsNumeric() {
		return result, errors.New("I NEED A VALUE")
	}

	size := tt.AsInteger()
	txt := apple2helpers.TEXT(caller)
	txt.Font = types.TextSize(size)

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandSETTEXTSIZE) Syntax() string {

	/* vars */
	var result string

	result = "TEXT"

	/* enforce non void return */
	return result

}
