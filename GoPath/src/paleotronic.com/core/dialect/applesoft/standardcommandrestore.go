package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"strings"
)

type StandardCommandRESTORE struct {
	dialect.Command
}

func (this *StandardCommandRESTORE) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	if caller.GetState() != types.RUNNING {
		return 0, nil
	}

	/* vars */
	result := 0
	var ln int

	b := caller.GetCode()

	ln = b.GetLowIndex()
	caller.GetDataRef().Line = ln
	caller.GetDataRef().Token = 0
	caller.GetDataRef().Statement = 0
	caller.GetDataRef().SubIndex = 0

	t := caller.GetTokenAtCodeRef(*caller.GetDataRef())

	if (t.Type != types.KEYWORD) || (strings.ToLower(t.Content) != "data") {
		caller.NextTokenInstance(caller.GetDataRef(), types.KEYWORD, "data")
	}

	//System.Out.Println("After beforeRun() for data");

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandRESTORE) Syntax() string {

	/* vars */
	var result string

	result = "RESTORE"

	/* enforce non void return */
	return result

}
