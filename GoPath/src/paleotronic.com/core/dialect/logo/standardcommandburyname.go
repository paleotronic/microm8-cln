package logo

import (
	"strings"
	//	"errors"
	//	"paleotronic.com/fmt"

	"paleotronic.com/core/dialect"
	//"paleotronic.com/core/hires"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	//	"paleotronic.com/core/hardware/apple2helpers"
)

type StandardCommandBURYNAME struct {
	dialect.Command
	Split bool
	Local bool
}

func (this *StandardCommandBURYNAME) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	rtok, _ := this.Command.D.(*DialectLogo).ParseTokensForResult(caller, tokens)

	var commands = make([]string, 0)

	if rtok.Type == types.LIST {
		for _, v := range rtok.List.Content {
			if v.Content != "" {
				commands = append(commands, strings.ToLower(v.Content))
			}
		}
	} else {
		commands = append(commands, strings.ToLower(rtok.Content))
	}

	// Now do a bury
	d := this.Command.D.(*DialectLogo)
	for _, n := range commands {
		d.Driver.SetBuryVar(n, true)
	}

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandBURYNAME) Syntax() string {

	/* vars */
	var result string

	result = "GRAPHICS"

	/* enforce non void return */
	return result

}
