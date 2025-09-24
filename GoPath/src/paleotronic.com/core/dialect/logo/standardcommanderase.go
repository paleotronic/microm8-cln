package logo

import (
	"errors"
	//	"errors"
	//	"paleotronic.com/fmt"

	"paleotronic.com/core/dialect"
	//"paleotronic.com/core/hires"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	//	"paleotronic.com/core/hardware/apple2helpers"
)

type StandardCommandERASE struct {
	dialect.Command
	Split bool
	Local bool
}

func (this *StandardCommandERASE) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	if tokens.Size() == 0 {
		return result, errors.New("NOT ENOUGH INPUTS")
	}

	if tokens.Size() > 1 {
		return result, errors.New("I DON'T KNOW WHAT TO DO WITH " + tokens.Get(1).AsString())
	}

	name := tokens.Shift().Content

	d := this.Command.D.(*DialectLogo)

	if _, ok := d.Driver.GetProc(name); !ok {
		return result, errors.New(name + " IS NOT DEFINED")
	}

	d.Driver.EraseProc(name)

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandERASE) Syntax() string {

	/* vars */
	var result string

	result = "GRAPHICS"

	/* enforce non void return */
	return result

}

func NewStandardCommandERASE() *StandardCommandERASE {
	return &StandardCommandERASE{
		Local: true,
	}
}
