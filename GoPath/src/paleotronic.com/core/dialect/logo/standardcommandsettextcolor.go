package logo

import (
	"errors"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandSETTEXTCOLOR struct {
	dialect.Command
}

func (this *StandardCommandSETTEXTCOLOR) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	// get result
	tt, e := this.Command.D.(*DialectLogo).ParseTokensForResult(caller, tokens)
	if e != nil {
		return result, e
	}

	if tt.Type != types.LIST || tt.List.Size() != 2 {
		return result, errors.New("I NEED 2 VALUES")
	}

	fg := tt.List.Get(0).AsInteger()
	bg := tt.List.Get(1).AsInteger()
	txt := apple2helpers.TEXT(caller)
	txt.FGColor = uint64(fg % 16)
	txt.BGColor = uint64(bg % 16)

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandSETTEXTCOLOR) Syntax() string {

	/* vars */
	var result string

	result = "TEXT"

	/* enforce non void return */
	return result

}
