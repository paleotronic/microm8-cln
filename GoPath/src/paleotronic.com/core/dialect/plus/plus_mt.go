package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/microtracker" //	"paleotronic.com/core/memory"
	//	"paleotronic.com/utils"
)

//var TrackerSong [settings.NUMSLOTS]*tracker.TSong

type PlusMicroTracker struct {
	dialect.CoreFunction
}

func (this *PlusMicroTracker) FunctionExecute(params *types.TokenList) error {

	t := microtracker.NewMicroTracker(this.Interpreter)
	if params.Size() == 1 && params.Get(0).Content != "" {
		filename := params.Get(0).Content
		t.Song.Load(filename)
	}
	t.Run(this.Interpreter)

	return nil

}

func (this *PlusMicroTracker) Syntax() string {

	/* vars */
	var result string

	result = "ALTCASE{n}"

	/* enforce non void return */
	return result

}

func (this *PlusMicroTracker) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusMicroTracker(a int, b int, params types.TokenList) *PlusMicroTracker {
	this := &PlusMicroTracker{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "VIDEOMODE"

	this.MinParams = 0
	this.MaxParams = 1

	return this
}
