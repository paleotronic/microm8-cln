package logo

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/files"
    "errors"
)

type StandardCommandSETWRITEPOS struct {
	dialect.Command
}

func (this *StandardCommandSETWRITEPOS) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

    if tokens.Size() < 1 {
       return result, errors.New( "I NEED A VALUE" )
    }
    
    // get result
    tt, e := this.Command.D.(*DialectLogo).ParseTokensForResult( caller, tokens )
    if e != nil {
       return result, e
    }
    
	e = files.DOSSEEK( files.GetPath(files.Writer), files.GetFilename(files.Writer), tt.AsInteger()  )

	/* enforce non void return */
	return result, e

}

func (this *StandardCommandSETWRITEPOS) Syntax() string {

	/* vars */
	var result string

	result = "TEXT"

	/* enforce non void return */
	return result

}
