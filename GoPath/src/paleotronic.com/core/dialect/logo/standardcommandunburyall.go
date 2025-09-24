package logo

import (
	//	"strings"
	//	"errors"
	//	"paleotronic.com/fmt"

	"paleotronic.com/core/dialect"
	//"paleotronic.com/core/hires"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	//	"paleotronic.com/core/hardware/apple2helpers"
)

type StandardCommandUNBURYALL struct {
	dialect.Command
	Split bool
	Local bool
}

func (this *StandardCommandUNBURYALL) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	//rtok, _ := this.Command.D.(*DialectLogo).ParseTokensForResult(caller, tokens)

	// Now do a bury
	d := this.Command.D.(*DialectLogo)
	d.Driver.SetBuryAllProcs(false)
	d.Driver.SetBuryAllVars(false)

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandUNBURYALL) Syntax() string {

	/* vars */
	var result string

	result = "GRAPHICS"

	/* enforce non void return */
	return result

}
