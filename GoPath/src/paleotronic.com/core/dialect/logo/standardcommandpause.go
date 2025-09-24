package logo

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandPAUSE struct {
	dialect.Command
}

func (this *StandardCommandPAUSE) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	//cx := caller.GetVDU().GetCursorX()
	d := this.Command.D.(*DialectLogo)
	d.Driver.PauseExecution() // this should save the stack and run state...

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandPAUSE) Syntax() string {

	/* vars */
	var result string

	result = "TEXT"

	/* enforce non void return */
	return result

}
