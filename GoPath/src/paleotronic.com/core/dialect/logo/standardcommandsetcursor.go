package logo

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	//	"paleotronic.com/core/hardware/apple2helpers"
	"errors"
	//    "paleotronic.com/fmt"
)

type StandardCommandSETCURSOR struct {
	dialect.Command
}

func (this *StandardCommandSETCURSOR) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	if tokens.Size() < 1 {
		return result, errors.New("I NEED A VALUE")
	}

	//fmt.Println("params:", caller.TokenListAsString(tokens))

	list, _ := this.Command.D.(*DialectLogo).ParseTokensForResult(caller, tokens)

	//fmt.Printf("list = %s\n", list.List.AsString())

	if list.Type != types.LIST || list.List.Size() < 2 {
		return result, errors.New("I NEED 2 VALUES")
	}

	caller.SetMemory(36, uint64(list.List.Get(0).AsInteger()))
	caller.SetMemory(37, uint64(list.List.Get(1).AsInteger()))

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandSETCURSOR) Syntax() string {

	/* vars */
	var result string

	result = "TEXT"

	/* enforce non void return */
	return result

}
