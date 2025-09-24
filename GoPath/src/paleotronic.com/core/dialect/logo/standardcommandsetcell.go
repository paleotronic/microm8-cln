package logo

import (
	"errors"
	"strings"

	//	"errors"
	//	"paleotronic.com/fmt"

	"paleotronic.com/core/dialect"
	//"paleotronic.com/core/hires"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type StandardCommandSETCELL struct {
	dialect.Command
}

func (this *StandardCommandSETCELL) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	name := tokens.Shift().Content
	if !strings.HasPrefix(name, ":") {
		name = ":" + name
	}

	//tokens, e := this.Expect(caller, tokens, 2)
	rtok, e := this.Command.D.(*DialectLogo).ParseTokensForResult(caller, tokens)
	if e != nil {
		return result, e
	}

	if rtok.Type == types.LIST {
		d := this.Command.D.(*DialectLogo)
		if rtok.List.Size() != 2 {
			return result, errors.New("command expects 3 arguments")
		}
		index := rtok.List.Shift()
		value := rtok.List.Shift()
		if index.Type != types.LIST || index.List.Size() != 2 {
			return result, errors.New("second argument should be two element list")
		}
		r, c := index.List.Get(0).AsInteger(), index.List.Get(1).AsInteger()
		list, _ := d.Driver.GetVar(name)
		err := TableSetCell(list, r, c, value)
		if err != nil {
			return result, err
		}
	}

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandSETCELL) Syntax() string {

	/* vars */
	var result string

	result = "GRAPHICS"

	/* enforce non void return */
	return result

}
