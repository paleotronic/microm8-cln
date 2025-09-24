package logo

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/files"
	"paleotronic.com/fmt"
)

type StandardCommandALLOPEN struct {
	dialect.Command
}

func (this *StandardCommandALLOPEN) Syntax() string {

	/* vars */
	var result string

	result = "ALLOPEN"

	/* enforce non void return */
	return result

}

func (this *StandardCommandALLOPEN) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	for i, v := range files.DOSBUFFERS() {
		caller.PutStr( fmt.Sprintf("%d %s\r\n", i+1, v) )
	}

	/* enforce non void return */
	return result, nil

}
