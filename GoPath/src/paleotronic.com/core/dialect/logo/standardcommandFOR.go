package logo

import (
	"errors"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandFOR struct {
	dialect.Command
}

func (this *StandardCommandFOR) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	// get result
	tt, e := this.Command.D.(*DialectLogo).ParseTokensForResult(caller, tokens)
	if e != nil {
		return result, e
	}

	if tt.Type != types.LIST || tt.List.Size() != 3 {
		return result, errors.New("I NEED 3 VALUES")
	}

	d := this.Command.D.(*DialectLogo)

	varname := tt.List.Get(0).Content
	value := tt.List.Get(1).AsExtended()
	list := tt.List.Get(2).List
	step := float64(1)

	if varname == "" {
		return result, errors.New("FOR EXPECTS VAR")
	}
	if list == nil {
		return result, errors.New("FOR NEEDS COMMAND LIST")
	}

	vtok, scope := d.Driver.GetVar(varname)
	if vtok == nil {
		// set global
		vtok = types.NewToken(types.NUMBER, "0")
	}
	if scope == nil {
		d.Driver.Globals.Set(varname, vtok)
	} else {
		scope.Vars.Set(varname, vtok)
	}

	// handle neg steps
	if value < vtok.AsExtended() {
		step = -step
	} else if value == vtok.AsExtended() {
		step = 0
	}

	//log.Printf("for varname=%s, to=%f, step=%f", varname, value, step)
	d.Driver.CreateForBlockScope(varname, value, step, list)

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandFOR) Syntax() string {

	/* vars */
	var result string

	result = "TEXT"

	/* enforce non void return */
	return result

}
