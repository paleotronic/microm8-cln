package logo

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types" //	"paleotronic.com/core/hardware/apple2helpers"
)

type StandardCommandCAMTURTLE struct {
	dialect.Command
}

func (this *StandardCommandCAMTURTLE) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	d := this.Command.D.(*DialectLogo)
	d.Driver.Track()

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandCAMTURTLE) Syntax() string {

	/* vars */
	var result string

	result = "TEXT"

	/* enforce non void return */
	return result

}
