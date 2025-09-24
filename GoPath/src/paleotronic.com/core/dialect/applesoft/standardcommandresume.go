package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandRESUME struct {
    dialect.Command
}

func (this *StandardCommandRESUME) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

        /* vars */
        var result int
        var cr types.CodeRef


      result = 0

      //if (caller.IsRunning);
      //  return types.NewESyntaxError("CANNOT RESUME");

      if (caller.GetBreakpoint().Line != -1) {
        //apple2helpers.PutStr(caller,"Currently at "+IntToStr(caller.PC.Line));
        caller.SetState(types.RUNNING)
        cr = caller.GetNextStatement(*caller.GetBreakpoint())
        //apple2helpers.PutStr(caller,"Resume from "+IntToStr(cr.Line));
        caller.GetPC().Line = cr.Line
        caller.GetPC().Statement = cr.Statement
        caller.GetPC().Token = 0
        //caller.Halt();
      }

        /* enforce non void return */
        return result, nil

}

func (this *StandardCommandRESUME) Syntax() string {

        /* vars */
        var result string

      result = "RESUME"

        /* enforce non void return */
        return result

}

