package logo

import (
	"errors"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types" //	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/microtracker/mock"
	"paleotronic.com/microtracker/tracker"
)

type StandardCommandSETINSTRUMENT struct {
	dialect.Command
}

func (this *StandardCommandSETINSTRUMENT) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	v, err := this.Command.D.ParseTokensForResult(caller, tokens)
	if err != nil {
		return result, err
	}
	if v == nil || v.Type == types.LIST {
		return result, errors.New("I NEED A VALUE")
	}

	filename := v.Content

	s := TrackerSong[caller.GetMemIndex()]
	if s == nil {
		s = tracker.NewSong(120, mock.New(caller, 0xc400))
		s.Start(tracker.PMBoundPattern)
		TrackerSong[caller.GetMemIndex()] = s
	}

	i := &tracker.TInstrument{}
	if i.Load(filename) == nil {
		s.Instruments[0] = i
	} else {
		return result, errors.New("Failed to load instrument")
	}

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandSETINSTRUMENT) Syntax() string {

	/* vars */
	var result string

	result = "TEXT"

	/* enforce non void return */
	return result

}
