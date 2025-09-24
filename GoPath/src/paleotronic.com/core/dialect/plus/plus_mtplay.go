package plus

import (
	"strings"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/microtracker" //	"paleotronic.com/core/memory"
	"paleotronic.com/microtracker/mock"
	"paleotronic.com/microtracker/tracker"
	//	"paleotronic.com/utils"
)

var TrackerSong [settings.NUMSLOTS]*tracker.TSong

type PlusMicroTrackerPlayer struct {
	dialect.CoreFunction
}

func (this *PlusMicroTrackerPlayer) FunctionExecute(params *types.TokenList) error {

	//	if e := this.CoreFunction.FunctionExecute(params); e != nil {
	//		return e
	//	}

	if TrackerSong[this.Interpreter.GetMemIndex()] != nil {
		TrackerSong[this.Interpreter.GetMemIndex()].Stop()
	}

	if params.Size() == 1 && params.Get(0).Content != "" {
		filename := params.Get(0).Content
		if !strings.HasPrefix(filename, "/") && this.Interpreter.GetWorkDir() != "" {
			filename = strings.Trim(this.Interpreter.GetWorkDir(), "/") + "/" + strings.Trim(filename, "/")
		}
		TrackerSong[this.Interpreter.GetMemIndex()] = tracker.NewSong(120, mock.New(this.Interpreter, 0xc400))
		err := TrackerSong[this.Interpreter.GetMemIndex()].Load(filename)
		if err != nil {
			return err
		}
		TrackerSong[this.Interpreter.GetMemIndex()].Start(tracker.PMLoopSong)
		TrackerSong[this.Interpreter.GetMemIndex()].SetPlayMode(tracker.PMLoopSong)
	} else {
		t := microtracker.NewMicroTracker(this.Interpreter)
		t.Run(this.Interpreter)
	}

	return nil

}

func (this *PlusMicroTrackerPlayer) Syntax() string {

	/* vars */
	var result string

	result = "ALTCASE{n}"

	/* enforce non void return */
	return result

}

func (this *PlusMicroTrackerPlayer) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusMicroTrackerPlayer(a int, b int, params types.TokenList) *PlusMicroTrackerPlayer {
	this := &PlusMicroTrackerPlayer{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "VIDEOMODE"

	this.MinParams = 0
	this.MaxParams = 1

	return this
}
