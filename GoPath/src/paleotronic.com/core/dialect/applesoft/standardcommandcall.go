package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/exception"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"time"
)

type StandardCommandCALL struct {
	dialect.Command
}

func (this *StandardCommandCALL) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int
	var tok types.Token
	//var vtok types.Token
	var addr int
	//var hc int
	//var i types.TokenList

	result = 0

	this.Cost = 0

	if tokens.Size() == 0 {
		return result, exception.NewESyntaxError("SYNTAX ERROR")
	}

	tok = caller.ParseTokensForResult(tokens)

	if tok.Type != types.NUMBER && tok.Type != types.INTEGER {
		return result, exception.NewESyntaxError("SYNTAX ERROR")
	}

	addr = tok.AsInteger()

	if addr < 0 {
		addr = 65536 + addr
	}

	apple2helpers.DoCall(addr, caller, true)
	
	for time.Now().Before( caller.GetWaitUntil() ) {
		// hold the line
	}

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandCALL) Syntax() string {

	/* vars */
	var result string

	result = "CALL"

	/* enforce non void return */
	return result

}
