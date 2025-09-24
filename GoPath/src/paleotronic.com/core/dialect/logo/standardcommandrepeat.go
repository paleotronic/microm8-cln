package logo

import (
	"errors"

	"paleotronic.com/fmt"

	"paleotronic.com/core/dialect"
	//"paleotronic.com/core/hires"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	//	"paleotronic.com/core/hardware/apple2helpers"
)

type StandardCommandREPEAT struct {
	dialect.Command
	Split bool
}

func (this *StandardCommandREPEAT) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	fmt.Println("REPEAT: ", caller.TokenListAsString(tokens))

	tokens, e := this.Expect(caller, tokens, 2)

	if e != nil {
		fmt.Println("error =", e)
		return result, e
	}

	fmt.Printf("tokens.Size() = %d\n", tokens.Size())

	if tokens.Size() < 2 {
		return result, nil
	}

	listpos := tokens.IndexOf(types.LIST, "")
	fmt.Printf("listpos = %d", listpos)

	if listpos == -1 {
		return result, errors.New("EXPECTED A LIST")
	}

	rptexpr := tokens.SubList(0, listpos)
	exp, e := this.Command.D.(*DialectLogo).ParseTokensForResult(caller, *rptexpr)
	if e != nil {
		fmt.Printf("ptfr err = %v\n", e)
		return result, e
	}
	count := exp.AsInteger()

	fmt.Printf("count = %d\n", count)

	list := tokens.Get(listpos).List

	d := this.Command.D.(*DialectLogo)

	d.Driver.CreateRepeatBlockScope(count, list)

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandREPEAT) Syntax() string {

	/* vars */
	var result string

	result = "GRAPHICS"

	/* enforce non void return */
	return result

}
