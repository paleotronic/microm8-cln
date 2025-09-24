package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/core/exception"
	"strings"
	"paleotronic.com/core/hardware/apple2helpers"
)

type StandardCommandDSP struct {
    dialect.Command
}

func (this *StandardCommandDSP) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

        /* vars */
        var result int
        var wtl types.TokenList
        //Token tok;


      result = 0

      wtl = *types.NewTokenList()
      for _, tok := range tokens.Content  {
        if (tok.Type == types.VARIABLE) {
          if (!caller.GetDialect().GetWatchVars().ContainsKey(strings.ToLower(tok.Content))) {
            caller.GetDialect().GetWatchVars().Push( strings.ToLower(tok.Content) )
            wtl.Push(tok)
          }
        } else {
          return result, exception.NewESyntaxError("SYNTAX ERROR")
        }
      }

      apple2helpers.PutStr(caller,"DSP ("+caller.TokenListAsString(wtl)+")")
      // removed free call here;


        /* enforce non void return */
        return result, nil

}

func (this *StandardCommandDSP) Syntax() string {

        /* vars */
        var result string

      result = "DSP"

        /* enforce non void return */
        return result

}

