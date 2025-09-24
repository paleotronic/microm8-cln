package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandMON struct {
	dialect.Command
}

func NewStandardCommandMON() *StandardCommandMON {
	this := &StandardCommandMON{}
	this.ImmediateMode = true
	return this
}

func (this *StandardCommandMON) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	m := apple2helpers.NewMonitor(caller)
	m.Manual("")

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandMON) Syntax() string {

	/* vars */
	var result string

	result = "NEW"

	/* enforce non void return */
	return result

}
