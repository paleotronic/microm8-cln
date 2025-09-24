package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/core/hardware/apple2helpers"
    "errors"
//    "paleotronic.com/fmt"
)

type StandardCommandHCOLOR struct {
	dialect.Command
}

func (this *StandardCommandHCOLOR) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	if tokens.Size() == 0 {
       return result, errors.New( "SYNTAX ERROR" )
    }
    
    t := caller.ParseTokensForResult(tokens)
    
    apple2helpers.SetHCOLOR( caller, t.AsInteger() )
    
    //fmt.Printf("HCOLOR= %d\n", t.AsInteger())

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandHCOLOR) Syntax() string {

	/* vars */
	var result string

	result = "HCOLOR="

	/* enforce non void return */
	return result

}
