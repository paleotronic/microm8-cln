package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/core/memory"
)

type StandardCommandFLUSH struct {
	dialect.Command
}

func (this *StandardCommandFLUSH) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	for i := 256; i < memory.OCTALYZER_INTERPRETER_SIZE; i++ {
		caller.SetMemory(i, 0)
	}

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandFLUSH) Syntax() string {

	/* vars */
	var result string

	result = "FLUSH"

	/* enforce non void return */
	return result

}
