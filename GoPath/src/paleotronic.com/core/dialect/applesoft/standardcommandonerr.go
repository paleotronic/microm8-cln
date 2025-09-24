package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/core/exception"
	"strings"
)

type StandardCommandONERR struct {
    dialect.Command
}

func (this *StandardCommandONERR) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

        /* vars */
        var result int
        //var vtok types.Token
        //var cr types.CodeRef
        //var tla types.TokenListArray
        //var tl types.TokenList
        //var i types.TokenList


      result = 0

      if (tokens.Size() != 2) {
        return result, exception.NewESyntaxError("SYNTAX ERROR")
      }

      if ((tokens.Get(0).Type != types.KEYWORD) || (strings.ToLower(tokens.Get(0).Content) != "goto")) {
        return result, exception.NewESyntaxError("SYNTAX ERROR")
      }

      if ((tokens.Get(1).Type != types.NUMBER)) {
        return result, exception.NewESyntaxError("SYNTAX ERROR")
      }

      caller.GetErrorTrap().Line = tokens.Get(1).AsInteger()
      caller.GetErrorTrap().Statement = 0
      caller.GetErrorTrap().Token = 0


        /* enforce non void return */
        return result, nil

}

func (this *StandardCommandONERR) Syntax() string {

        /* vars */
        var result string

      result = "ONERR GOTO <line>"

        /* enforce non void return */
        return result

}

