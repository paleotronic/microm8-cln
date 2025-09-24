package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	//	"paleotronic.com/core/memory"
	//	"paleotronic.com/utils"
)

type PlusMicroTrackerPause struct {
	dialect.CoreFunction
}

func (this *PlusMicroTrackerPause) FunctionExecute(params *types.TokenList) error {

	//	if e := this.CoreFunction.FunctionExecute(params); e != nil {
	//		return e
	//	}

	if TrackerSong[this.Interpreter.GetMemIndex()] != nil {
		TrackerSong[this.Interpreter.GetMemIndex()].TogglePause()
	}

	return nil

}

func (this *PlusMicroTrackerPause) Syntax() string {

	/* vars */
	var result string

	result = "ALTCASE{n}"

	/* enforce non void return */
	return result

}

func (this *PlusMicroTrackerPause) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusMicroTrackerPause(a int, b int, params types.TokenList) *PlusMicroTrackerPause {
	this := &PlusMicroTrackerPause{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "VIDEOMODE"

	this.MinParams = 0
	this.MaxParams = 0

	return this
}
