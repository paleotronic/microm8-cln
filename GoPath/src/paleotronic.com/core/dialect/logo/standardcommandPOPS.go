package logo

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types" //	"paleotronic.com/core/hardware/apple2helpers"
)

type StandardCommandPOPS struct {
	dialect.Command
	vars  bool
	procs bool
}

func (this *StandardCommandPOPS) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	// get procedure code
	d := this.Command.D.(*DialectLogo)

	code := d.Driver.GetWorkspaceBody(this.procs, this.vars)

	for _, l := range code {
		caller.PutStr(l + "\r\n")
	}

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandPOPS) Syntax() string {

	/* vars */
	var result string

	result = "TEXT"

	/* enforce non void return */
	return result

}
