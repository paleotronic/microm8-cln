package logo

import (
	"strings"
	//	"errors"
	//	"paleotronic.com/fmt"

	"paleotronic.com/core/dialect"
	//"paleotronic.com/core/hires"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"

	"paleotronic.com/log"
)

type StandardCommandSHARE struct {
	dialect.Command
	Split bool
	Local bool
}

func (this *StandardCommandSHARE) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

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

	log.Printf("Make: Token type: %s, Token Value: %s, Token List: %+v", rtok.Type, rtok.Content, rtok.List)

	//caller.SetData(strings.ToLower(name), *rtok, false)

	d := this.Command.D.(*DialectLogo)
	d.Driver.SetVarShared(name, rtok)

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandSHARE) Syntax() string {

	/* vars */
	var result string

	result = "GRAPHICS"

	/* enforce non void return */
	return result

}
