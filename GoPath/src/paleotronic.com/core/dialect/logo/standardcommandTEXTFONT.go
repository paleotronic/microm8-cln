package logo

import (
	"errors"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types" //	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/octalyzer/video/font"
)

type StandardCommandTEXTFONT struct {
	dialect.Command
}

func (this *StandardCommandTEXTFONT) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	v, err := this.Command.D.ParseTokensForResult(caller, tokens)
	if err != nil {
		return result, err
	}
	if v == nil || v.Type == types.LIST {
		return result, errors.New("I NEED A VALUE")
	}

	fontid := v.AsInteger()
	index := caller.GetMemIndex()
	if fontid >= 0 && fontid < len(settings.AuxFonts[index]) {
		fontName := settings.AuxFonts[index][fontid]
		f, err := font.LoadFromFile(fontName)
		if err == nil {
			settings.DefaultFont[index] = f
			settings.ForceTextVideoRefresh = true
		}
	}

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandTEXTFONT) Syntax() string {

	/* vars */
	var result string

	result = "TEXT"

	/* enforce non void return */
	return result

}
