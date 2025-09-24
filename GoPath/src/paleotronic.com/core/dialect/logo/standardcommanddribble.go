package logo

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/files"
)

type StandardCommandDRIBBLE struct {
	dialect.Command
}

func (this *StandardCommandDRIBBLE) Syntax() string {

	/* vars */
	var result string

	result = "DRIBBLE <filename>"

	/* enforce non void return */
	return result

}

func (this *StandardCommandDRIBBLE) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	rtok, e := this.Command.D.(*DialectLogo).ParseTokensForResult( caller, tokens )
	if e != nil {
		return result, e
	}

	// try open a files
	e = files.DOSDRIBBLE( caller.GetWorkDir(), rtok.Content)

	/* enforce non void return */
	return result, e

}
