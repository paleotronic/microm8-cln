package appleinteger

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandWozPOP struct {
	dialect.Command
}

func (this *StandardCommandWozPOP) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	caller.Pop(false)

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandWozPOP) Syntax() string {

	/* vars */
	var result string

	result = "POP"

	/* enforce non void return */
	return result

}
