package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/core/exception"
)

type StandardCommandGOTO struct {
    dialect.Command
}

func (this *StandardCommandGOTO) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

        /* vars */
        var result int
        var vtok types.Token
        var cr types.CodeRef


      result = 0

      if ((tokens.Size() == 0)) {
        return result, exception.NewESyntaxError("LINE NUMBER REQUIRED")
      }

      vtok = caller.ParseTokensForResult(tokens)

      if ((vtok.Type != types.NUMBER) && (vtok.Type != types.INTEGER)) {
        return result, exception.NewESyntaxError("LINE NUMBER MUST BE NUMERIC ("+vtok.Content+")")
      }

      cr = *types.NewCodeRef()
      cr.Line = vtok.AsInteger()
      cr.Statement = 0
      cr.Token = 0

      if ((caller.GetState() != types.RUNNING) && (caller.GetStack().Size() == 0)) {
        caller.SetState(types.RUNNING)
      }

      if (!caller.IsCodeRefValid(cr)) {
        return result, exception.NewESyntaxError("LINE DOES NOT EXIST ("+vtok.Content+")")
      }

      cr.SubIndex = -1

      if (caller.GetState() == types.RUNNING) {
         caller.SetPC(&cr)
      } else {
          caller.SetLPC(&cr)
      }

        /* enforce non void return */
        return result, nil

}

func (this *StandardCommandGOTO) Syntax() string {

        /* vars */
        var result string

      result = "GOTO <line>"

        /* enforce non void return */
        return result

}

