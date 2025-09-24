package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/core/hardware/apple2helpers"
)

type StandardCommandCLS struct {
	dialect.Command
}

func (this *StandardCommandCLS) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0
	apple2helpers.Clearscreen(caller)

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandCLS) Syntax() string {

	/* vars */
	var result string

	result = "CLS"

	/* enforce non void return */
	return result

}
