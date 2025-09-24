package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	//	"paleotronic.com/core/memory"
	//	"paleotronic.com/utils"
)

type PlusMicroTrackerPattern struct {
	dialect.CoreFunction
}

func (this *PlusMicroTrackerPattern) FunctionExecute(params *types.TokenList) error {

	//	if e := this.CoreFunction.FunctionExecute(params); e != nil {
	//		return e
	//	}

	if TrackerSong[this.Interpreter.GetMemIndex()] != nil {
		if params.Size() > 0 {
			pnum := params.Get(0).AsInteger()
			TrackerSong[this.Interpreter.GetMemIndex()].JumpPattern(pnum)
		}
	}

	return nil

}

func (this *PlusMicroTrackerPattern) Syntax() string {

	/* vars */
	var result string

	result = "ALTCASE{n}"

	/* enforce non void return */
	return result

}

func (this *PlusMicroTrackerPattern) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusMicroTrackerPattern(a int, b int, params types.TokenList) *PlusMicroTrackerPattern {
	this := &PlusMicroTrackerPattern{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "VIDEOMODE"

	this.MinParams = 1
	this.MaxParams = 1

	return this
}
