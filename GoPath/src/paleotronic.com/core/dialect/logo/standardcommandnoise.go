package logo

import (
	"errors"
	"time"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/fmt"
)

type StandardCommandNOISE struct {
	dialect.Command
	solid bool
}

func (this *StandardCommandNOISE) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	// get result
	tt, e := this.Command.D.(*DialectLogo).ParseTokensForResult(caller, tokens)
	if e != nil {
		return result, e
	}

	if tt.Type != types.LIST || tt.List.Size() != 2 {
		return result, errors.New("I NEED 2 VALUES")
	}

	pitch := tt.List.Get(0).AsInteger()
	duration := tt.List.Get(1).AsInteger() * (1000 / 60)

	//cx := caller.GetVDU().GetCursorX()
	cmd := fmt.Sprintf(`
use mixer.voices.boom
set instrument "WAVE=NOISE:VOLUME=1.0:ADSR=0,0,%d,1"
set frequency %d
set volume 1.0
	`, duration, pitch)
	caller.PassRestBufferNB(cmd)

	time.Sleep(time.Millisecond * time.Duration(duration))
	caller.PassRestBufferNB("set volume 0.0")

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandNOISE) Syntax() string {

	/* vars */
	var result string

	result = "TEXT"

	/* enforce non void return */
	return result

}
