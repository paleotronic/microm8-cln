package logo

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	//	"paleotronic.com/core/hardware/apple2helpers"
	//    "errors"
	//   "time"
)

type StandardCommandCRESET struct {
	dialect.Command
}

func (this *StandardCommandCRESET) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	index := caller.GetMemIndex()
	mm := caller.GetMemoryMap()
	cindex := 1
	control := types.NewOrbitController(mm, index, cindex)
	control.ResetALL()
	control.Update()
	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandCRESET) Syntax() string {

	/* vars */
	var result string

	result = "TEXT"

	/* enforce non void return */
	return result

}
