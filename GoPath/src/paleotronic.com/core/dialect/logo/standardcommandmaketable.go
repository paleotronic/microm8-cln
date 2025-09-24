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

	"paleotronic.com/log"
)

type StandardCommandMAKETABLE struct {
	dialect.Command
	Local  bool
	Global bool
	Shared bool
}

func (this *StandardCommandMAKETABLE) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

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
	if rtok.Type != types.LIST {
		return result, errors.New("Expected list for second arg")
	}

	if rtok.List.Size() != 2 {
		return result, errors.New("Expected 2 items in list")
	}

	tb := TableCreate(name, rtok.List.Get(0).AsInteger(), rtok.List.Get(1).AsInteger())

	d := this.Command.D.(*DialectLogo)
	if this.Local {
		d.Driver.SetVarLocal(name, tb)
	} else if this.Global {
		d.Driver.Globals.Set(name, tb)
	} else if this.Shared {
		SharedVars.Set(name, tb)
	}

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandMAKETABLE) Syntax() string {

	/* vars */
	var result string

	result = "GRAPHICS"

	/* enforce non void return */
	return result

}

func NewStandardCommandSHARETABLE() *StandardCommandMAKETABLE {
	return &StandardCommandMAKETABLE{
		Shared: true,
	}
}

func NewStandardCommandLOCALTABLE() *StandardCommandMAKETABLE {
	return &StandardCommandMAKETABLE{
		Local: true,
	}
}

func NewStandardCommandGLOBALTABLE() *StandardCommandMAKETABLE {
	return &StandardCommandMAKETABLE{
		Global: true,
	}
}
