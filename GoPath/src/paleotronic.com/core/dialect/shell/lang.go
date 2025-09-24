package shell

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/exception"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/runestring"
	"strings"
)

type StandardCommandDIALECT struct {
	dialect.Command
}

func NewStandardCommandDIALECT() *StandardCommandDIALECT {
    this := &StandardCommandDIALECT{}
    this.ImmediateMode = true
    return this
}

func (this *StandardCommandDIALECT) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int
	//var tok types.Token
	//var vtok types.Token
	//var addr int
	//var hc int
	//var i types.TokenList
	var s string
	//var d interfaces.Dialecter

	result = 0

	if tokens.Size() == 0 {
		return result, exception.NewESyntaxError("SYNTAX ERROR")
	}

	s = strings.ToLower(caller.TokenListAsString(tokens))

    apple2helpers.TextSaveScreen(caller)
    
	e := caller.NewChild( strings.ToLower(s) )
	caller.SetChild(e)
	e.SetParent(caller)
    
	apple2helpers.TEXT40(e)
	apple2helpers.Attribute(e, types.VA_NORMAL)

	e.Bootstrap(s, false)
	e.SetBuffer(runestring.NewRuneString())
    memory.WarmStart = true
    e.LoadSpec( caller.GetSpec() )
    memory.WarmStart = false
	//e.GetVDU().PutStr("Entering shell. Type 'exit' to return.")

	//caller.Clear();

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandDIALECT) Syntax() string {

	/* vars */
	var result string

	result = "DIALECT"

	/* enforce non void return */
	return result

}
