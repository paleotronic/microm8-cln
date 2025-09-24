package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type StandardCommandAMPCALL struct {
	dialect.Command
}

func (this *StandardCommandAMPCALL) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int
	var addr int
	var tl types.TokenList

	result = 0

	addr = 1013

	tl = *types.NewTokenList()
	tl.Push(types.NewToken(types.KEYWORD, "call"))
	tl.Push(types.NewToken(types.NUMBER, utils.IntToStr(addr)))

	caller.GetDialect().ExecuteDirectCommand(tl, caller, Scope, &LPC)

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandAMPCALL) Syntax() string {

	/* vars */
	var result string

	result = "AMPCALL"

	/* enforce non void return */
	return result

}
