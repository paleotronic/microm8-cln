package logo

import (
//	"strings"
	//	"errors"
//	"paleotronic.com/fmt"

	"paleotronic.com/core/dialect"
	//"paleotronic.com/core/hires"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/files"
)

type StandardCommandNODRIBBLE struct {
	dialect.Command
	Split bool
	Local bool
}

func (this *StandardCommandNODRIBBLE) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	e := files.DOSNODRIBBLE()

	/* enforce non void return */
	return result, e

}

func (this *StandardCommandNODRIBBLE) Syntax() string {

	/* vars */
	var result string

	result = "GRAPHICS"

	/* enforce non void return */
	return result

}
