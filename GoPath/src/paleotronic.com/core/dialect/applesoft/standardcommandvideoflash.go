package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/core/hardware/apple2helpers"
)

type StandardCommandVIDEOFLASH struct {
	dialect.Command
}

func (this *StandardCommandVIDEOFLASH) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0
	apple2helpers.Attribute(caller, types.VA_BLINK)

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandVIDEOFLASH) Syntax() string {

	/* vars */
	var result string

	result = "VIDEOFLASH"

	/* enforce non void return */
	return result

}
