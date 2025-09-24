package logo

import (
	"errors"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types" //	"paleotronic.com/core/hardware/apple2helpers"
)

type StandardCommandPON struct {
	dialect.Command
}

func (this *StandardCommandPON) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

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

	d := this.Command.D.(*DialectLogo)
	vv := d.Driver.Globals.Get(v.Content)
	if vv == nil {
		return result, errors.New("NO SUCH VARIABLE")
	}
	var txt string
	if vv.IsPropList {
		txt = d.Driver.DumpObjectStruct(vv, false, "PPROP \""+v.Content+" ")
	} else {
		txt = d.Driver.DumpObjectStruct(vv, false, "MAKE \""+v.Content+" ")
	}

	caller.PutStr(txt + "\r\n")

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandPON) Syntax() string {

	/* vars */
	var result string

	result = "TEXT"

	/* enforce non void return */
	return result

}
