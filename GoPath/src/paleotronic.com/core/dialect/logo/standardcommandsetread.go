package logo

import (
	"paleotronic.com/files"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandSETREAD struct {
	dialect.Command
}

func (this *StandardCommandSETREAD) Syntax() string {

	/* vars */
	var result string

	result = "SETREAD <filename>"

	/* enforce non void return */
	return result

}

func (this *StandardCommandSETREAD) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	rtok, e := this.Command.D.(*DialectLogo).ParseTokensForResult(caller, tokens)
	if e != nil {
		return result, e
	}

	// try open a files
	e = files.DOSREAD(caller.GetWorkDir(), rtok.Content, 0)
	if e == nil {
		files.Reader = caller.GetWorkDir() + "/" + rtok.Content
	}

	/* enforce non void return */
	return result, e

}
