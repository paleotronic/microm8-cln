package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/core/hardware/apple2helpers"
)

type StandardCommandVIDEONORMAL struct {
	dialect.Command
}

func (this *StandardCommandVIDEONORMAL) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0
	apple2helpers.Attribute(caller, types.VA_NORMAL)

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandVIDEONORMAL) Syntax() string {

	/* vars */
	var result string

	result = "VIDEONORMAL"

	/* enforce non void return */
	return result

}
