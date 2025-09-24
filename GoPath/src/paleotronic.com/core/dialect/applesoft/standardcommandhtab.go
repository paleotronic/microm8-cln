package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/exception"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	//"paleotronic.com/fmt"
)

type StandardCommandHTAB struct {
	dialect.Command
}

func (this *StandardCommandHTAB) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int
	var tok types.Token

	result = 0

	if tokens.Size() == 0 {
		return result, exception.NewESyntaxError("SYNTAX ERROR")
	}

	tok = caller.ParseTokensForResult(tokens)

	if (tok.Type != types.NUMBER) && (tok.Type != types.INTEGER) {
		return result, exception.NewESyntaxError("SYNTAX ERROR")
	}

	hpos := tok.AsInteger()

	if hpos < 1 {
		return result, exception.NewESyntaxError("ILLEGAL QUANTITY ERROR")
	}

	//sx, sy, ex, ey, ww, hh := apple2helpers.GetRealWindow(caller)
	////fmt.Printf("SX=%d, SY=%d, EX=%d, EY=%d, W=%d, H=%d\n", sx, sy, ex, ey, ww, hh)

	apple2helpers.HTab(caller, hpos)

	/* enforce non void return */
	return result, nil
}

func (this *StandardCommandHTAB) Syntax() string {

	/* vars */
	var result string

	result = "HTAB"

	/* enforce non void return */
	return result

}
