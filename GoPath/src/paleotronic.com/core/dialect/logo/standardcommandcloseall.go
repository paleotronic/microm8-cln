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

type StandardCommandCLOSEALL struct {
	dialect.Command
	Split bool
	Local bool
}

func (this *StandardCommandCLOSEALL) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	e := files.DOSCLOSEALL()

	/* enforce non void return */
	return result, e

}

func (this *StandardCommandCLOSEALL) Syntax() string {

	/* vars */
	var result string

	result = "GRAPHICS"

	/* enforce non void return */
	return result

}
