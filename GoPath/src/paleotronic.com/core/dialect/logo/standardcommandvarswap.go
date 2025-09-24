package logo

import (
	"errors"
	"fmt"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandVARSWAP struct {
	dialect.Command
}

func (this *StandardCommandVARSWAP) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	if tokens.Size() < 1 {
		return result, errors.New("I NEED A VALUE")
	}

	// get result
	tt, e := this.Command.D.(*DialectLogo).ParseTokensForResult(caller, tokens)
	if e != nil {
		return result, e
	}

	if tt == nil || tt.Type != types.LIST || tt.List.Size() != 2 {
		return result, errors.New("I NEED 2 VALUES")
	}

	v1 := tt.List.Get(0).Content
	v2 := tt.List.Get(1).Content

	d := this.Command.D.(*DialectLogo)
	val1, scp1 := d.Driver.GetVar(v1)
	if val1 == nil {
		return result, fmt.Errorf("no such var %s", v1)
	}
	val2, scp2 := d.Driver.GetVar(v2)
	if val2 == nil {
		return result, fmt.Errorf("no such var %s", v2)
	}
	scp1.Vars.Set(v1, val2)
	scp2.Vars.Set(v2, val1)

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandVARSWAP) Syntax() string {

	/* vars */
	var result string

	result = "TEXT"

	/* enforce non void return */
	return result

}
