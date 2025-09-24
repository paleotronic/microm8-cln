package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/exception"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandVTAB struct {
	dialect.Command
}

func (this *StandardCommandVTAB) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int
	var tok types.Token
//	var rr int

	result = 0

	if tokens.Size() == 0 {
		return result, exception.NewESyntaxError("SYNTAX ERROR")
	}

	tok = caller.ParseTokensForResult(tokens)

	if (tok.Type != types.NUMBER) && (tok.Type != types.INTEGER) {
		return result, exception.NewESyntaxError("SYNTAX ERROR")
	}

	vpos := tok.AsInteger()
	//caller.SetMemory(37, uint(vpos-1))
    
    if vpos < 1 || vpos > 48 {
       return result, exception.NewESyntaxError("ILLEGAL QUANTITY ERROR")
    }

	//caller.GetDialect().SetCursorY(caller, vpos-1)
    
    apple2helpers.VTab(caller, vpos)

	//apple2helpers.VTab(caller, vpos)

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandVTAB) Syntax() string {

	/* vars */
	var result string

	result = "VTAB"

	/* enforce non void return */
	return result

}
