package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/core/hardware/apple2helpers"
)

type StandardCommandNODSP struct {
	dialect.Command
}

func (this *StandardCommandNODSP) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0
	caller.GetDialect().GetWatchVars().Clear()
	apple2helpers.PutStr(caller,"DSP OFF\r\n")

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandNODSP) Syntax() string {

	/* vars */
	var result string

	result = "NODSP"

	/* enforce non void return */
	return result

}
