package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandLOMEM struct {
	dialect.Command
}

func (this *StandardCommandLOMEM) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int
	//var cr types.CodeRef

	result = 0

	t, e := caller.GetDialect().ParseTokensForResult(caller, tokens)
	if e != nil {
		return 0, e
	}

	addr := t.AsInteger()

	if addr < 0 {
		addr = 65536 + addr
	}

	caller.GetLocal().SetLoBound(addr)

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandLOMEM) Syntax() string {

	/* vars */
	var result string

	result = "LOMEM:"

	/* enforce non void return */
	return result

}
