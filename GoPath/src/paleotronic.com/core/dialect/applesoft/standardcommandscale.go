package applesoft

import (
	"errors"
//	"paleotronic.com/fmt"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandSCALE struct {
	dialect.Command
}

func (this *StandardCommandSCALE) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	if tokens.Size() == 0 {
		return result, errors.New("SYNTAX ERROR")
	}

	t := caller.ParseTokensForResult(tokens)

	apple2helpers.SetSCALE(caller, t.AsInteger())

	//fmt.Printf("SCALE= %d\n", t.AsInteger())

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandSCALE) Syntax() string {

	/* vars */
	var result string

	result = "SCALE="

	/* enforce non void return */
	return result

}
