package logo

import (
	"errors"
	"strings"

	//	"paleotronic.com/fmt"

	"paleotronic.com/core/dialect"
	//"paleotronic.com/core/hires"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	//	"paleotronic.com/core/hardware/apple2helpers"
)

type StandardCommandTHROW struct {
	dialect.Command
	Split bool
	Local bool
}

func (this *StandardCommandTHROW) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	rtok, _ := this.Command.D.(*DialectLogo).ParseTokensForResult(caller, tokens)

	//	var commands = make([]string, 0)

	// Throw generates an exception with the named text
	if rtok.Type == types.LIST {
		rtok = rtok.List.Shift()
	}

	if strings.ToLower(rtok.Content) == "toplevel" {
		d := this.Command.D.(*DialectLogo)
		d.Driver.ThrowTopLevel()
		return result, nil
	}

	return result, errors.New(rtok.Content)

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandTHROW) Syntax() string {

	/* vars */
	var result string

	result = "GRAPHICS"

	/* enforce non void return */
	return result

}
