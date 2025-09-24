package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"strings"
)

type StandardCommandDATA struct {
	dialect.Command
}

func (this *StandardCommandDATA) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandDATA) Syntax() string {

	/* vars */
	var result string

	result = "DATA"

	/* enforce non void return */
	return result

}

func NewStandardCommandDATA() *StandardCommandDATA {
	this := &StandardCommandDATA{}
	this.NoTokens = true
	return this
}

func (this *StandardCommandDATA) BeforeRun(caller interfaces.Interpretable) {

	/* vars */
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

}
