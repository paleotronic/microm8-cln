package logo

import (
	//	"strings"

	"paleotronic.com/core/dialect"
	"paleotronic.com/fmt"

	//	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	//"errors"
)

type StandardCommandLOGOLIST struct {
	dialect.Command
}

func (this *StandardCommandLOGOLIST) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	d := this.Command.D.(*DialectLogo)

	caller.PutStr(fmt.Sprintf("%6s  %-10s %-40s\r", "ID", "STATE", "ERR"))
	caller.PutStr(fmt.Sprintf("%6s  %-10s %-40s\r", "==", "=====", "==="))
	for id, cr := range d.Driver.Coroutines {
		err := ""
		if cr.err != nil {
			err = cr.err.Error()
		}
		caller.PutStr(fmt.Sprintf("%6s  %-10s %-40s\r", id, cr.State.String(), err))
	}

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandLOGOLIST) Syntax() string {

	/* vars */
	var result string

	result = "TEXT"

	/* enforce non void return */
	return result

}
