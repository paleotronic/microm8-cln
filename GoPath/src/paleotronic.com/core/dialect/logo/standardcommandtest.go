package logo

import (
	//	"errors"
	//	"paleotronic.com/fmt"
	"paleotronic.com/core/dialect" //"paleotronic.com/core/hires"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	//	"paleotronic.com/core/hardware/apple2helpers"
)

type StandardCommandTEST struct {
	dialect.Command
	Split bool
}

func (this *StandardCommandTEST) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	/*
		TEST ...........
	*/

	if tokens.Size() < 2 {
		return result, nil
	}

	test := tokens.SubList(0, tokens.Size())
	rr, e := this.Command.D.(*DialectLogo).ParseTokensForResult(caller, *test)
	if e != nil {
		return result, e
	}

	var d = this.Command.D.(*DialectLogo)
	d.Driver.Globals.Set("__last_test__", rr)

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandTEST) Syntax() string {

	/* vars */
	var result string

	result = "GRAPHICS"

	/* enforce non void return */
	return result

}
