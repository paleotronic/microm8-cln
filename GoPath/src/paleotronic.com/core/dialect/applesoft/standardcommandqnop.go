package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
    //"errors"
)

type StandardCommandQNOP struct {
	dialect.Command
}

func (this *StandardCommandQNOP) Syntax() string {

	/* vars */
	var result string

	result = "NOP"

	/* enforce non void return */
	return result

}

func (this *StandardCommandQNOP) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	/* enforce non void return */
	return result, nil

}
