package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/core/hardware/apple2helpers"
)

type StandardCommandTRACE struct {
	dialect.Command
}

func (this *StandardCommandTRACE) Syntax() string {

	/* vars */
	var result string

	result = "TRACE"

	/* enforce non void return */
	return result

}

func (this *StandardCommandTRACE) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0
	caller.GetDialect().SetTrace(true)
	apple2helpers.PutStr(caller,"TRACE ON\r\n")

	/* enforce non void return */
	return result, nil

}
