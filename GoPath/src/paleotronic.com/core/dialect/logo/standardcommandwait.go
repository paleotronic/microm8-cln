package logo

import (
	"time"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandWAIT struct {
	dialect.Command
}

func (this *StandardCommandWAIT) Syntax() string {

	/* vars */
	var result string

	result = "WAIT"

	/* enforce non void return */
	return result

}

func (this *StandardCommandWAIT) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	t, _ := this.Command.D.(*DialectLogo).ParseTokensForResult(caller, tokens)

	if t.Type == types.LIST {
		t = t.List.Shift()
	}

	dur := 1000 / 60 * t.AsInteger()

	time.Sleep(time.Duration(dur) * time.Millisecond)

	/* enforce non void return */
	return result, nil

}
