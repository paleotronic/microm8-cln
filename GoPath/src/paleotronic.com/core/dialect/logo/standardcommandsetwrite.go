package logo

import (
	"paleotronic.com/files"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandSETWRITE struct {
	dialect.Command
}

func (this *StandardCommandSETWRITE) Syntax() string {

	/* vars */
	var result string

	result = "SETWRITE <filename>"

	/* enforce non void return */
	return result

}

func (this *StandardCommandSETWRITE) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	rtok, e := this.Command.D.(*DialectLogo).ParseTokensForResult(caller, tokens)
	if e != nil {
		return result, e
	}

	// try open a files
	e = files.DOSWRITE(caller.GetWorkDir(), rtok.Content, 0)
	if e == nil {
		files.Writer = caller.GetWorkDir() + "/" + rtok.Content
	}

	/* enforce non void return */
	return result, e

}
