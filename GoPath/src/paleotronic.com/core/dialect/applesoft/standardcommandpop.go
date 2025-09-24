package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandPOP struct {
	dialect.Command
}

func (this *StandardCommandPOP) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	//caller.Pop(true)

	/* enforce non void return */
	return result, caller.Pop(true)

}

func (this *StandardCommandPOP) Syntax() string {

	/* vars */
	var result string

	result = "POP"

	/* enforce non void return */
	return result

}
