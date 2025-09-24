package logo

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types" //	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/types/glmath"
)

type StandardCommandCFOCUSTOITLE struct {
	dialect.Command
}

func (this *StandardCommandCFOCUSTOITLE) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	index := caller.GetMemIndex()
	mm := caller.GetMemoryMap()
	cindex := 1
	control := types.NewOrbitController(mm, index, cindex)
	control.SetLookAtV(
		&glmath.Vector3{
			apple2helpers.VECTOR(caller).GetTurtle(this.Command.D.(*DialectLogo).Driver.GetTurtle()).Position[0],
			apple2helpers.VECTOR(caller).GetTurtle(this.Command.D.(*DialectLogo).Driver.GetTurtle()).Position[1],
			apple2helpers.VECTOR(caller).GetTurtle(this.Command.D.(*DialectLogo).Driver.GetTurtle()).Position[2],
		},
	)
	control.Update()

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandCFOCUSTOITLE) Syntax() string {

	/* vars */
	var result string

	result = "TEXT"

	/* enforce non void return */
	return result

}
