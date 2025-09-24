package logo

import (
	"errors"
//	"paleotronic.com/fmt"
	"strings"

	"paleotronic.com/core/dialect"
	//"paleotronic.com/core/hires"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	//	"paleotronic.com/core/hardware/apple2helpers"
)

type StandardCommandCATCH struct {
	dialect.Command
	Split bool
}

func (this *StandardCommandCATCH) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	tokens, e := this.Expect(caller, tokens, 2)

	if e != nil {
		return result, e
	}

	if tokens.Size() < 2 {
		return result, nil
	}

	//fmt.Printf("%d tokens in the list\n", tokens.Size())

	listpos := tokens.IndexOf(types.LIST, "")

	if listpos == -1 {
		return result, errors.New("EXPECTED A LIST")
	}

	rptexpr := tokens.SubList(0, listpos)
	exp, e := this.Command.D.(*DialectLogo).ParseTokensForResult(caller, *rptexpr)
	if e != nil {
		return result, e
	}
	errtype := exp.Content

	list := tokens.Get(listpos).List

	// now lets prepend this list onto the buffer
	//	for i := 0; i < count; i++ {
	//		a := caller.GetCode()
	//		caller.GetDialect().ExecuteDirectCommand(*list, caller, &a, caller.GetLPC())
	//	}

	a := caller.GetDirectAlgorithm()
	
	e = caller.GetDialect().ExecuteDirectCommand( *list, caller, a, caller.GetLPC() )
	if e != nil && strings.ToLower(e.Error()) == strings.ToLower(errtype) {
		//fmt.Println("in catch", e.Error())
		e = nil
	}

	/* enforce non void return */
	return result, e

}

func (this *StandardCommandCATCH) Syntax() string {

	/* vars */
	var result string

	result = "GRAPHICS"

	/* enforce non void return */
	return result

}
