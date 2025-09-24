package logo

import (
	"paleotronic.com/files"
	"paleotronic.com/log"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandOPEN struct {
	dialect.Command
}

func (this *StandardCommandOPEN) Syntax() string {

	/* vars */
	var result string

	result = "OPEN <filename>"

	/* enforce non void return */
	return result

}

func (this *StandardCommandOPEN) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	rtok, e := this.Command.D.(*DialectLogo).ParseTokensForResult(caller, tokens)
	if e != nil {
		return result, e
	}

	// try open a files
	e = files.DOSOPEN(caller.GetWorkDir(), rtok.Content, 0)

	log.Println(files.Buffers)

	/* enforce non void return */
	return result, e

}
