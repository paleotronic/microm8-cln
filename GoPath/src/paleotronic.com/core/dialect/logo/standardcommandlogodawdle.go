package logo

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandLOGODAWDLE struct {
	dialect.Command
	solid bool
}

func (this *StandardCommandLOGODAWDLE) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	d := caller.GetDialect().(*DialectLogo)
	err := d.Driver.WaitAllCoroutines()

	/* enforce non void return */
	return result, err

}

func (this *StandardCommandLOGODAWDLE) Syntax() string {

	/* vars */
	var result string

	result = "TEXT"

	/* enforce non void return */
	return result

}
