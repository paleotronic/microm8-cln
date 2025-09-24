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

type StandardCommandBURYALL struct {
	dialect.Command
	Split bool
	Local bool
}

func (this *StandardCommandBURYALL) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	//rtok, _ := this.Command.D.(*DialectLogo).ParseTokensForResult(caller, tokens)

	// var commands = caller.GetDialect().GetDynamicCommands()
	// var data = caller.GetDataNames()

	d := this.Command.D.(*DialectLogo)

	// Now do a bury
	d.Driver.SetBuryAllProcs(true)
	d.Driver.SetBuryAllVars(true)

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandBURYALL) Syntax() string {

	/* vars */
	var result string

	result = "GRAPHICS"

	/* enforce non void return */
	return result

}
