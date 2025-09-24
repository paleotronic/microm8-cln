package logo

import (
	"errors" //	"errors"
	//	"paleotronic.com/fmt"
	"paleotronic.com/core/dialect" //"paleotronic.com/core/hires"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	//	"paleotronic.com/core/hardware/apple2helpers"
)

type StandardCommandIFTRUE struct {
	dialect.Command
	Split bool
}

func (this *StandardCommandIFTRUE) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	if tokens.Size() < 1 {
		return result, nil
	}

	//fmt.Printf("%d tokens in the list\n", tokens.Size())

	tpos := tokens.IndexOfN(-1, types.LIST, "")

	if tpos == -1 {
		return result, errors.New("NEED COMMAND")
	}

	var d = this.Command.D.(*DialectLogo)
	rr := d.Driver.Globals.Get("__last_test__")
	if rr == nil {
		rr = types.NewToken(types.NUMBER, "0")
	}

	if rr.AsInteger() != 0 {
		d.Driver.CreateRepeatBlockScope(1, tokens.Get(tpos).List.Copy())
	}

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandIFTRUE) Syntax() string {

	/* vars */
	var result string

	result = "GRAPHICS"

	/* enforce non void return */
	return result

}
