package logo

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
)

type StandardCommandFAST struct {
	dialect.Command
}

func (this *StandardCommandFAST) Syntax() string {

	/* vars */
	var result string

	result = "RETURN"

	/* enforce non void return */
	return result

}

func (this *StandardCommandFAST) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	settings.LogoFastDraw[caller.GetMemIndex()] = true

	/* enforce non void return */
	return result, nil

}
