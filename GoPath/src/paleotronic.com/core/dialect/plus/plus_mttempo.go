package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusTrackerTempo struct {
	dialect.CoreFunction
}

func (this *PlusTrackerTempo) FunctionExecute(params *types.TokenList) error {

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

	if s != nil {
		s.Tempo = p.AsInteger()
	}

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(0)))

	return nil
}

func (this *PlusTrackerTempo) Syntax() string {

	/* vars */
	var result string

	result = "PLAY{soundfile}"

	/* enforce non void return */
	return result

}

func (this *PlusTrackerTempo) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusTrackerTempo(a int, b int, params types.TokenList) *PlusTrackerTempo {
	this := &PlusTrackerTempo{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "PLAY"
	this.MinParams = 1
	this.MaxParams = 1

	return this
}
