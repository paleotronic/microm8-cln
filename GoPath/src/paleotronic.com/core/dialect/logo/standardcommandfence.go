package logo

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandFENCE struct {
	dialect.Command
}

func (this *StandardCommandFENCE) Syntax() string {

	/* vars */
	var result string

	result = "RETURN"

	/* enforce non void return */
	return result

}

func (this *StandardCommandFENCE) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	apple2helpers.VECTOR(caller).GetTurtle( this.Command.D.(*DialectLogo).Driver.GetTurtle() ).SetBoundsMode(types.FENCE)

	/* enforce non void return */
	return result, nil

}
