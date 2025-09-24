package logo

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandSTOP struct {
	dialect.Command
}

func (this *StandardCommandSTOP) Syntax() string {

	/* vars */
	var result string

	result = "STOP"

	/* enforce non void return */
	return result

}

func (this *StandardCommandSTOP) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	d := this.Command.D.(*DialectLogo)
	d.Driver.ReturnFromProc(nil)

	/* enforce non void return */
	return result, nil

}
