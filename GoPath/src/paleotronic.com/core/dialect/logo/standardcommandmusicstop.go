package logo

import (

	//	"errors"
	//	"paleotronic.com/fmt"

	"paleotronic.com/core/dialect"

	//"paleotronic.com/core/hires"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	//	"paleotronic.com/core/hardware/apple2helpers"
)

type StandardCommandMUSICSTOP struct {
	dialect.Command
}

func (this *StandardCommandMUSICSTOP) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	caller.StopMusic()
	if TrackerSong[caller.GetMemIndex()] != nil {
		TrackerSong[caller.GetMemIndex()].Stop()
		TrackerSong[caller.GetMemIndex()] = nil
	}

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandMUSICSTOP) Syntax() string {

	/* vars */
	var result string

	result = "GRAPHICS"

	/* enforce non void return */
	return result

}

func NewStandardCommandMUSICSTOP() *StandardCommandMUSICSTOP {
	return &StandardCommandMUSICSTOP{}
}
