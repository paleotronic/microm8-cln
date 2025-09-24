package logo

import (
	"errors"
	"strings"

	//	"errors"
	//	"paleotronic.com/fmt"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/settings"
	"paleotronic.com/files"
	"paleotronic.com/microtracker/tracker"

	//"paleotronic.com/core/hires"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	//	"paleotronic.com/core/hardware/apple2helpers"
)

type StandardCommandMUSICPLAY struct {
	dialect.Command
}

var TrackerSong [settings.NUMSLOTS]*tracker.TSong

func (this *StandardCommandMUSICPLAY) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	if tokens.Size() == 0 {
		return result, errors.New("NOT ENOUGH INPUTS")
	}

	if tokens.Size() > 1 {
		return result, errors.New("I DON'T KNOW WHAT TO DO WITH " + tokens.Get(1).AsString())
	}

	caller.StopMusic()
	if TrackerSong[caller.GetMemIndex()] != nil {
		TrackerSong[caller.GetMemIndex()].Stop()
		TrackerSong[caller.GetMemIndex()] = nil
	}

	name := tokens.Shift().Content
	switch strings.ToLower(files.GetExt(name)) {
	case "rst", "ogg", "sng":
		caller.PlayMusic(files.GetPath(name), files.GetFilename(name), 0, 0)
	}

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandMUSICPLAY) Syntax() string {

	/* vars */
	var result string

	result = "GRAPHICS"

	/* enforce non void return */
	return result

}

func NewStandardCommandMUSICPLAY() *StandardCommandMUSICPLAY {
	return &StandardCommandMUSICPLAY{}
}
