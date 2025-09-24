package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/core/hardware/apple2helpers"
)

type StandardCommandNOTRACE struct {
	dialect.Command
}

func (this *StandardCommandNOTRACE) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0
	caller.GetDialect().SetTrace(false)
	apple2helpers.PutStr(caller,"TRACE OFF")

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandNOTRACE) Syntax() string {

	/* vars */
	var result string

	result = "NOTRACE"

	/* enforce non void return */
	return result

}
