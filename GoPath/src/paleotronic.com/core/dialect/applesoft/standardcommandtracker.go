// +build !remint

package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/microtracker"
)

type StandardCommandTRACKER struct {
	dialect.Command
}

func NewStandardCommandTRACKER() *StandardCommandTRACKER {
	this := &StandardCommandTRACKER{}
	this.ImmediateMode = true
	return this
}

func (this *StandardCommandTRACKER) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	m := microtracker.NewMicroTracker(caller)
	m.Run(caller)

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandTRACKER) Syntax() string {

	/* vars */
	var result string

	result = "NEW"

	/* enforce non void return */
	return result

}
