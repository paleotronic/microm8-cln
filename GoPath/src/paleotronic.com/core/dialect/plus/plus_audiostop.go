package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusAudioStop struct {
	dialect.CoreFunction
}

func (this *PlusAudioStop) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	// r := this.Interpreter.GetMemoryMap()
	// SendSongStop(this.Interpreter, r)
	if TrackerSong[this.Interpreter.GetMemIndex()] != nil {
		TrackerSong[this.Interpreter.GetMemIndex()].Stop()
		TrackerSong[this.Interpreter.GetMemIndex()] = nil
	}

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(0)))

	return nil
}

func (this *PlusAudioStop) Syntax() string {

	/* vars */
	var result string

	result = "PLAY{soundfile}"

	/* enforce non void return */
	return result

}

func (this *PlusAudioStop) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewPlusAudioStop(a int, b int, params types.TokenList) *PlusAudioStop {
	this := &PlusAudioStop{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "PLAY"

	return this
}
