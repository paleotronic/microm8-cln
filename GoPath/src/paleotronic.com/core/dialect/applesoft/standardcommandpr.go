package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/exception"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandPR struct {
	dialect.Command
}

func (this *StandardCommandPR) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int
	var tok types.Token
	var i int

	result = 0

	if tokens.Size() == 0 {
		return result, exception.NewESyntaxError("SYNTAX ERROR")
	}

	/* now get number */
	tok = *tokens.Pop()
	if (tok.Type != types.INTEGER) && (tok.Type != types.NUMBER) {
		return result, exception.NewESyntaxError("SYNTAX ERROR")
	}

	i = tok.AsInteger()

	switch i { /* FIXME - Switch statement needs cleanup */
	case 0:
		{
			//caller.GetVDU().SetVideoMode(caller.GetVDU().GetVideoModes()[5])
			apple2helpers.TEXT40(caller)
			caller.SetMemory(49152, 0)
			//caller.GetVDU().RegenerateWindow(caller.GetMemory())
			break
		}
	case 3:
		{
			//caller.GetVDU().SetVideoMode(caller.GetVDU().GetVideoModes()[0])
			apple2helpers.TEXT80(caller)
			caller.SetMemory(49153, 0)
			//caller.GetVDU().RegenerateWindow(caller.GetMemory())
			break
		}
	}

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandPR) Syntax() string {

	/* vars */
	var result string

	result = "PR"

	/* enforce non void return */
	return result

}
