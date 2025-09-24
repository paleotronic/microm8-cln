package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandNEW struct {
	dialect.Command
}

func NewStandardCommandNEW() *StandardCommandNEW {
	this := &StandardCommandNEW{}
	this.ImmediateMode = true
	return this
}

func (this *StandardCommandNEW) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0
	caller.PurgeOwnedVariables()
	caller.Clear()

	//caller.CreateVar(
	//   "speed",
	//   *types.NewVariableP( "speed", types.VT_FLOAT, "255", true ),
	//)

	// now thaw
	//caller.Thaw("machine.tmp")
	fixMemoryPtrs(caller)

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandNEW) Syntax() string {

	/* vars */
	var result string

	result = "NEW"

	/* enforce non void return */
	return result

}
