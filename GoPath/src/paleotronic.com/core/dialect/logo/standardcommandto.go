package logo

import (
	"errors"
	//	"paleotronic.com/fmt"

	"paleotronic.com/core/dialect"
	//"paleotronic.com/core/hires"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	//	"paleotronic.com/core/hardware/apple2helpers"
)

type StandardCommandTO struct {
	dialect.Command
	Split bool
}

func (this *StandardCommandTO) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	//fmt.Printf("%d tokens in the list\n", tokens.Size())

	if tokens.Size() < 1 {
		return result, errors.New("TO EXPECTS A PROCEDURE NAME")
	}

	name := tokens.Shift()
	vlist := *types.NewTokenList()
	vlist.Add(name)
	var e error
	if name.Type == types.STRING {
		name, e = this.Command.D.(*DialectLogo).ParseTokensForResult(caller, vlist)
		if e != nil {
			return result, e
		}
	}

	// if caller.GetDialect().GetDynaCommand(name.Content) != nil {
	// 	return result, errors.New(name.Content + " IS ALREADY DEFINED")
	// }

	params := []string(nil)
	for tokens.LPeek() != nil && tokens.LPeek().Type == types.VARIABLE {
		params = append(params, tokens.Shift().Content)
	}

	d := this.Command.D.(*DialectLogo)

	d.Driver.PendingProcName = name.Content
	d.Driver.PendingProcArgs = params
	d.Driver.PendingProcStatements = []string{}
	d.OldPrompt = caller.GetPrompt()

	caller.SetPrompt(">")

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandTO) Syntax() string {

	/* vars */
	var result string

	result = "GRAPHICS"

	/* enforce non void return */
	return result

}
