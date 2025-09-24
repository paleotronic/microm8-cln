package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandCLEAR struct {
	dialect.Command
}

func (this *StandardCommandCLEAR) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int
	//var cr types.CodeRef

	result = 0

	if caller.GetStack().Size() == 0 {
		caller.PurgeOwnedVariables()
	}

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandCLEAR) Syntax() string {

	/* vars */
	var result string

	result = "CLEAR"

	/* enforce non void return */
	return result

}
