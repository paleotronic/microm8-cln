package logo

import (
	//	"strings"

	"errors"
	"strings"

	"paleotronic.com/core/dialect"
	//	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	//"errors"
)

type StandardCommandPOTS struct {
	dialect.Command
	needproc bool
}

func (this *StandardCommandPOTS) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0
	filter := ""
	d := this.Command.D.(*DialectLogo)
	if tokens.Size() > 0 {
		f, e := d.ParseTokensForResult(caller, tokens)
		if e != nil {
			return result, e
		}
		filter = f.Content
	}
	if this.needproc && filter == "" {
		return result, errors.New("PROC REQUIRED")
	}
	for _, k := range d.Driver.GetProcList() {
		if filter != "" && strings.ToLower(filter) != k {
			continue
		}
		p, _ := d.Driver.GetProc(k)
		caller.PutStr(p.GetCode()[0] + "\r\n")
	}

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandPOTS) Syntax() string {

	/* vars */
	var result string

	result = "TEXT"

	/* enforce non void return */
	return result

}
