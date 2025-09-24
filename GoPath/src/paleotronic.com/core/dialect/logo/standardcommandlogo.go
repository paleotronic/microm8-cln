package logo

import (
	"errors"

	//"fmt"
	"paleotronic.com/fmt"

	"paleotronic.com/core/dialect"
	//"paleotronic.com/core/hires"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	//	"paleotronic.com/core/hardware/apple2helpers"
)

type StandardCommandLOGO struct {
	dialect.Command
	Split bool
}

func (this *StandardCommandLOGO) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	fmt.Println("REPEAT: ", caller.TokenListAsString(tokens))

	if tokens.Size() < 1 {
		return result, nil
	}

	listpos := 0

	if listpos == -1 {
		return result, errors.New("EXPECTED A LIST")
	}

	list := tokens.Get(listpos).List

	d := this.Command.D.(*DialectLogo)

	_, err := d.Driver.SpawnCoroutine(list)

	/* enforce non void return */
	return result, err

}

func (this *StandardCommandLOGO) Syntax() string {

	/* vars */
	var result string

	result = "GRAPHICS"

	/* enforce non void return */
	return result

}
