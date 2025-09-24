package logo

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandLOGOKILLALL struct {
	dialect.Command
	solid bool
}

func (this *StandardCommandLOGOKILLALL) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	d := caller.GetDialect().(*DialectLogo)
	err := d.Driver.KillAllCoroutines()

	/* enforce non void return */
	return result, err

}

func (this *StandardCommandLOGOKILLALL) Syntax() string {

	/* vars */
	var result string

	result = "TEXT"

	/* enforce non void return */
	return result

}
