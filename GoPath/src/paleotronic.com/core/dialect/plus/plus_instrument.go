package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/files"
	"paleotronic.com/microtracker/mock"
	"paleotronic.com/microtracker/tracker"
	"paleotronic.com/utils"
)

type PlusInstrument struct {
	dialect.CoreFunction
}

func (this *PlusInstrument) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	filename := params.Shift().Content

	//SendInstrument(this.Interpreter, this.Interpreter.GetMemoryMap(), filename)

	if files.ExistsViaProvider(files.GetPath(filename), files.GetFilename(filename)) {
		// instrument is a file!!
		i := &tracker.TInstrument{}
		err := i.Load(filename)
		if err != nil {
			return err
		}
	}

	s := TrackerSong[this.Interpreter.GetMemIndex()]
	if s == nil {
		s = tracker.NewSong(120, mock.New(this.Interpreter, 0xc400))
		s.Start(tracker.PMBoundPattern)
		TrackerSong[this.Interpreter.GetMemIndex()] = s
	}

	i := &tracker.TInstrument{}
	if i.Load(filename) == nil {
		s.Instruments[0] = i
	} else {
		this.Interpreter.PutStr("Failed to load instrument\r\n")
	}

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusInstrument) Syntax() string {

	/* vars */
	var result string

	result = "PLAY{soundfile}"

	/* enforce non void return */
	return result

}

func (this *PlusInstrument) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusInstrument(a int, b int, params types.TokenList) *PlusInstrument {
	this := &PlusInstrument{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "PLAY"

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "instrumentdata", Default: *types.NewToken(types.STRING, "")},
		},
	)

	return this
}
