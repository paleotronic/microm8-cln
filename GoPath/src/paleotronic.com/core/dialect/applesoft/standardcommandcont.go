package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/exception"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandCONT struct {
	dialect.Command
}

func (this *StandardCommandCONT) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int
	var cr types.CodeRef

	result = 0

	if caller.IsRunning() {
		return result, exception.NewESyntaxError("CANNOT CONTINUE")
	}

	if caller.GetPC().Line != -1 {
		//apple2helpers.PutStr(caller,"Currently at "+IntToStr(caller.GetPC().Line));
		caller.SetState(types.RUNNING)
		caller.PreOptimizer()
		cr = caller.GetNextStatement(*caller.GetPC())
		//apple2helpers.PutStr(caller,"Resume from "+IntToStr(cr.Line));
		caller.GetPC().Line = cr.Line
		caller.GetPC().Statement = cr.Statement
		caller.GetPC().Token = 0
	}

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandCONT) Syntax() string {

	/* vars */
	var result string

	result = "CONT"

	/* enforce non void return */
	return result

}
