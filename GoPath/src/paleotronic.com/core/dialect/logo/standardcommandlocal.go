package logo

import (
	"errors"
	"strings"

	//	"errors"
	//	"paleotronic.com/fmt"

	"paleotronic.com/core/dialect"
	//"paleotronic.com/core/hires"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	//	"paleotronic.com/core/hardware/apple2helpers"
)

type StandardCommandLOCAL struct {
	dialect.Command
	Split bool
	Local bool
}

func (this *StandardCommandLOCAL) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

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
	if !strings.HasPrefix(name, ":") {
		name = ":" + name
	}

	//tokens, e := this.Expect(caller, tokens, 2)

	rtok := types.NewToken(types.STRING, "")

	d := this.Command.D.(*DialectLogo)
	d.Driver.S.Vars.Set(name, rtok)

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandLOCAL) Syntax() string {

	/* vars */
	var result string

	result = "GRAPHICS"

	/* enforce non void return */
	return result

}

func NewStandardCommandLOCAL() *StandardCommandLOCAL {
	return &StandardCommandLOCAL{
		Local: true,
	}
}
