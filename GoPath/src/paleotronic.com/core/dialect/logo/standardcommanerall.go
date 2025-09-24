package logo

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	//	"paleotronic.com/core/hardware/apple2helpers"
	//   "errors"
)

type StandardCommandERALL struct {
	dialect.Command
}

func (this *StandardCommandERALL) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	d := this.Command.D.(*DialectLogo)
	d.Driver.EraseAllProcs()

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandERALL) Syntax() string {

	/* vars */
	var result string

	result = "TEXT"

	/* enforce non void return */
	return result

}
