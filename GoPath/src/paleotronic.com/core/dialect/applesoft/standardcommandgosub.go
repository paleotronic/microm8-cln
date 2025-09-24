package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/core/exception"
)

type StandardCommandGOSUB struct {
    dialect.Command
}

func (this *StandardCommandGOSUB) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

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

      if (!caller.IsCodeRefValid(cr)) {
        return result, exception.NewESyntaxError("LINE DOES NOT EXIST ("+vtok.Content+")")
      }

      if (caller.GetState() == types.RUNNING) {
		  a := caller.GetCode();
         caller.Call( cr, a, caller.GetState(), false, caller.GetVarPrefix(), *caller.GetTokenStack(), caller.GetDialect() )
      } else {
		  a := caller.GetDirectAlgorithm();
         caller.Call( cr, a, caller.GetState(), false, caller.GetVarPrefix(), *caller.GetTokenStack(), caller.GetDialect() )
      }

        /* enforce non void return */
        return result, nil

}

func (this *StandardCommandGOSUB) Syntax() string {

        /* vars */
        var result string

      result = "GOSUB <line>"

        /* enforce non void return */
        return result

}

