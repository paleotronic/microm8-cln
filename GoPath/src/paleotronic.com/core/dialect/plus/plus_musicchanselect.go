package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/microtracker/mock"
	"paleotronic.com/microtracker/tracker"
	"paleotronic.com/utils"
)

type PlusTrackerChanSelect struct {
	dialect.CoreFunction
}

func (this *PlusTrackerChanSelect) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	p := params.Get(0)
	if p.Type == types.VARIABLE {
		tl := types.NewTokenList()
		tl.Push(p.Copy())
		*p = this.Interpreter.ParseTokensForResult(*tl)
	}

	s := TrackerSong[this.Interpreter.GetMemIndex()]
	if s == nil {
		s = tracker.NewSong(120, mock.New(this.Interpreter, 0xc400))
		s.Start(tracker.PMBoundPattern)
		TrackerSong[this.Interpreter.GetMemIndex()] = s
	}

	if p.AsInteger() > 1 && p.AsInteger() <= 6 {
		s.SelectEntryTrack(p.AsInteger() - 1)
	}

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(0)))

	return nil
}

func (this *PlusTrackerChanSelect) Syntax() string {

	/* vars */
	var result string

	result = "PLAY{soundfile}"

	/* enforce non void return */
	return result

}

func (this *PlusTrackerChanSelect) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusTrackerChanSelect(a int, b int, params types.TokenList) *PlusTrackerChanSelect {
	this := &PlusTrackerChanSelect{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "PLAY"
	this.MinParams = 1
	this.MaxParams = 1

	return this
}
