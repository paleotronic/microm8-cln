package logo

import (
	"errors"

	"paleotronic.com/core/dialect"
	//"paleotronic.com/core/hires"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	//	"paleotronic.com/core/hardware/apple2helpers"
)

type StandardCommandWHILE struct {
	dialect.Command
	Split bool
}

func (this *StandardCommandWHILE) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	//tokens, e := this.Expect(caller, tokens, 2)

	// if e != nil {
	// 	fmt.Println("error =", e)
	// 	return result, e
	// }

	if tokens.Size() < 2 {
		return result, nil
	}

	listpos := tokens.IndexOf(types.LIST, "")

	if listpos == -1 {
		return result, errors.New("EXPECTED A LIST")
	}

	ex := tokens.SubList(0, listpos).Copy()
	//log.Printf("ex0 = %s", tlistStr("", ex))
	rptexpr := tokens.SubList(0, listpos).Copy()
	exp, e := this.Command.D.(*DialectLogo).ParseTokensForResult(caller, *rptexpr.Copy())
	if e != nil {
		return result, e
	}

	if exp.AsInteger() != 0 {
		// expr evals true
		list := tokens.Get(listpos).List.Copy()
		d := this.Command.D.(*DialectLogo)
		//log.Printf("ex = %s", tlistStr("", ex))
		d.Driver.CreateWhileBlockScope(ex, list)
	}

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandWHILE) Syntax() string {

	/* vars */
	var result string

	result = "GRAPHICS"

	/* enforce non void return */
	return result

}
