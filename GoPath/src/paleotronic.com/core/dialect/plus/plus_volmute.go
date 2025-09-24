package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	//	"paleotronic.com/core/memory"
	//	"paleotronic.com/core/settings"
	//	"time"
	//	"paleotronic.com/fmt"
)

type PlusVolumeMute struct {
	dialect.CoreFunction
}

func (this *PlusVolumeMute) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	if this.Query {

		v := settings.TemporaryMute
		if v {
			this.Stack.Push(types.NewToken(types.NUMBER, "1"))
		} else {
			this.Stack.Push(types.NewToken(types.NUMBER, "0"))
		}

	} else {

		l := this.ValueMap["mute"]
		mute := l.AsInteger()
		settings.TemporaryMute = (mute != 0)

	}

	//fmt.Printf("Setting zoom to %f\n", zoom)

	return nil
}

func (this *PlusVolumeMute) Syntax() string {

	/* vars */
	var result string

	result = "CAMZOOM{v}"

	/* enforce non void return */
	return result

}

func (this *PlusVolumeMute) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusVolumeMute(a int, b int, params types.TokenList) *PlusVolumeMute {
	this := &PlusVolumeMute{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "CAMZOOM"
	this.MinParams = 0
	this.MaxParams = 1
	this.Raw = true
	this.NamedParams = []string{"mute"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.INTEGER, "1"),
	}
	this.EvalSingleParam = true

	return this
}
