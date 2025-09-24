package logo

import (
	"errors"
	//	"errors"
	//	"paleotronic.com/fmt"

	"paleotronic.com/core/dialect"
	//"paleotronic.com/core/hires"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	//	"paleotronic.com/core/hardware/apple2helpers"
)

type StandardCommandIF struct {
	dialect.Command
	Split bool
}

func (this *StandardCommandIF) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	/*
		IF ........... [TRUE STEPS] [FALSE STEPS]
	*/

	//	tokens, e := this.Expect(caller, tokens, 3)

	//	if e != nil {
	//		return result, e
	//	}

	if tokens.Size() < 2 {
		return result, nil
	}

	//fmt.Printf("%d tokens in the list\n", tokens.Size())

	lists := tokens.FindCommandLists()
	if len(lists) == 0 {
		return result, errors.New("NEED TRUE CONDITION")
	}

	tpos := lists[0]

	epos := -1
	if len(lists) > 1 {
		epos = lists[1]
	}

	hasElse := (epos != -1)

	test := tokens.SubList(0, tpos)
	rr, e := this.Command.D.(*DialectLogo).ParseTokensForResult(caller, *test)
	if e != nil {
		return result, e
	}

	var d = this.Command.D.(*DialectLogo)

	if rr.AsInteger() != 0 {
		// TRUE
		//c := caller.GetCode()
		//caller.BufferCommands(tokens.Get(tpos).List.Copy(), 1)
		d.Driver.CreateRepeatBlockScope(1, tokens.Get(tpos).List.Copy())
		//caller.GetDialect().ExecuteDirectCommand(*tokens.Get(tpos).List, caller, c, caller.GetLPC())
	} else {
		// FALSE
		if hasElse {
			//caller.BufferCommands(tokens.Get(epos).List.Copy(), 1)
			d.Driver.CreateRepeatBlockScope(1, tokens.Get(epos).List.Copy())
			//caller.GetDialect().ExecuteDirectCommand(*tokens.Get(epos).List, caller, c, caller.GetLPC())
		}
	}

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandIF) Syntax() string {

	/* vars */
	var result string

	result = "GRAPHICS"

	/* enforce non void return */
	return result

}
