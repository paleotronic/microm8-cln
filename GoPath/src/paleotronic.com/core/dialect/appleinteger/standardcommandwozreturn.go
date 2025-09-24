package appleinteger

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandWozRETURN struct {
	dialect.Command
}

func (this *StandardCommandWozRETURN) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	caller.Return(false)

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandWozRETURN) Syntax() string {

	/* vars */
	var result string

	result = "RETURN"

	/* enforce non void return */
	return result

}
