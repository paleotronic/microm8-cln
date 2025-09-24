package logo

import (
	"errors"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types" //	"paleotronic.com/core/hardware/apple2helpers"
)

type StandardCommandTYPE struct {
	dialect.Command
}

func (this *StandardCommandTYPE) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	v, err := this.Command.D.ParseTokensForResult(caller, tokens)
	if err != nil {
		return result, err
	}
	if v == nil {
		return result, errors.New("I NEED A VALUE")
	}

	s := this.Command.D.(*DialectLogo).Driver.DumpObjectStruct(v, true, "")

	for _, ch := range s {
		apple2helpers.VECTOR(caller).GetTurtle(this.Command.D.(*DialectLogo).Driver.GetTurtle()).Glyph(ch)
	}

	apple2helpers.VECTOR(caller).Render()

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandTYPE) Syntax() string {

	/* vars */
	var result string

	result = "TEXT"

	/* enforce non void return */
	return result

}
