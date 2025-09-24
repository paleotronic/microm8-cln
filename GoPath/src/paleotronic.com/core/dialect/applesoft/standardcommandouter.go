package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandOUTER struct {
    dialect.Command
}

func (this *StandardCommandOUTER) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) int {

        /* vars */
        var result int


      result = 0

      caller.SetOuterVars(true)
      caller.GetDialect().ExecuteDirectCommand(tokens, caller, Scope, &LPC)
      caller.SetOuterVars(false)


        /* enforce non void return */
        return result

}

func (this *StandardCommandOUTER) Syntax() string {

        /* vars */
        var result string

      result = "OUTER <command>"

        /* enforce non void return */
        return result

}

