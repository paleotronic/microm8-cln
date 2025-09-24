package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandEND struct {
	dialect.Command
}

func (this *StandardCommandEND) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	if caller.IsRunning() {
		caller.EndProgram()
	}

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandEND) Syntax() string {

	/* vars */
	var result string

	result = "END"

	/* enforce non void return */
	return result

}
