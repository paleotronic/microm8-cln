package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/exception"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/core/hardware/apple2helpers"
)

type StandardCommandSPAWN struct {
	dialect.Command
}

func (this *StandardCommandSPAWN) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	filename := ""

	if tokens.Size() < 1 {
		return 0, exception.NewESyntaxError("SPAWN expects a name")
	}

	filename = tokens.Shift().Content

	e := caller.NewChild(filename)
	apple2helpers.PutStr(caller,"Started " + e.GetName())
	caller.SetChild(e)
	e.SetParent(caller)

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandSPAWN) Syntax() string {

	/* vars */
	var result string

	result = "END"

	/* enforce non void return */
	return result

}
