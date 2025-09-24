package logo

import (
	//	"strings"

	"paleotronic.com/core/dialect"
	//	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	//"errors"
)

type StandardCommandPPS struct {
	dialect.Command
}

func (this *StandardCommandPPS) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	for _, k := range caller.GetDataKeys() {
		v := caller.GetData(k)
		if v.IsPropList {
			// prop list
			for i := 0; i<v.List.Size(); i+=2 {
				caller.PutStr("PPROP \"" + k[1:] + " \"" + v.List.Content[i].Content +" [" + v.List.Content[i+1].List.AsString() + "]\r\n" )
			}
		}
	}

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandPPS) Syntax() string {

	/* vars */
	var result string

	result = "TEXT"

	/* enforce non void return */
	return result

}
