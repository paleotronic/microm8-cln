package logo

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	//"errors"
)

type StandardCommandHOMETURTLE struct {
	dialect.Command
}

func (this *StandardCommandHOMETURTLE) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	//    if tokens.Size() < 1 {
	//       return result, errors.New( "I NEED A VALUE" )
	//    }
	//
	//get result
	//    tt, e := this.Command.D.(*DialectLogo).ParseTokensForResult( caller, tokens )
	//    if e != nil {
	//       return result, e
	//    }

	//cx := caller.GetVDU().GetCursorX()
	apple2helpers.VECTOR(caller).GetTurtle(this.Command.D.(*DialectLogo).Driver.GetTurtle()).DoHome()
	if apple2helpers.IsVectorLayerEnabled(caller) == false {
		apple2helpers.HTab(caller, 1)
		apple2helpers.VTab(caller, 21)
		apple2helpers.ClearToBottom(caller)

		// enable vector layer
		apple2helpers.VectorLayerEnable(caller, true, true)
	}
	apple2helpers.VECTOR(caller).Render()

	// caller.GetMemoryMap().WriteGlobal(caller.GetMemoryMap().MEMBASE(caller.GetMemIndex())+memory.OCTALYZER_CAMERA_GFX_BASE+0, uint64(types.CC_ResetAll))

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandHOMETURTLE) Syntax() string {

	/* vars */
	var result string

	result = "TEXT"

	/* enforce non void return */
	return result

}
