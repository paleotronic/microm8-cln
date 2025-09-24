package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/microtracker/mock"
	"paleotronic.com/microtracker/tracker"
	"paleotronic.com/utils"
)

type PlusMusic struct {
	dialect.CoreFunction
}

func (this *PlusMusic) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	notes := params.Shift().Content

	s := TrackerSong[this.Interpreter.GetMemIndex()]
	if s == nil {
		s = tracker.NewSong(120, mock.New(this.Interpreter, 0xc400))
		s.Start(tracker.PMBoundPattern)
		TrackerSong[this.Interpreter.GetMemIndex()] = s
	}

	//this.Interpreter.GetVDU().SendRestalgiaEvent(types.RestalgiaPlayNoteStream, filename)
	s.EnterNotes(notes)

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusMusic) Syntax() string {

	/* vars */
	var result string

	result = "MUSIC{soundfile}"

	/* enforce non void return */
	return result

}

func (this *PlusMusic) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusMusic(a int, b int, params types.TokenList) *PlusMusic {
	this := &PlusMusic{}

	/* vars */
	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "PLAY"

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "notedata", Default: *types.NewToken(types.STRING, "A4B3")},
		},
	)

	return this
}
