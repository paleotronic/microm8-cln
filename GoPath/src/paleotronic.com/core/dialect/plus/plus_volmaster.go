package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	//	"paleotronic.com/core/memory"
	"paleotronic.com/utils"
	//	"paleotronic.com/core/settings"
	//	"time"
	//	"paleotronic.com/fmt"
)

type PlusVolumeMaster struct {
	dialect.CoreFunction
}

func (this *PlusVolumeMaster) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	if this.Query {

		v := settings.MixerVolume
		this.Stack.Push(types.NewToken(types.NUMBER, utils.FloatToStr(v)))

	} else {

		l := this.ValueMap["level"]
		level := l.AsExtended()
		settings.MixerVolume = level

	}

	//fmt.Printf("Setting zoom to %f\n", zoom)

	return nil
}

func (this *PlusVolumeMaster) Syntax() string {

	/* vars */
	var result string

	result = "CAMZOOM{v}"

	/* enforce non void return */
	return result

}

func (this *PlusVolumeMaster) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusVolumeMaster(a int, b int, params types.TokenList) *PlusVolumeMaster {
	this := &PlusVolumeMaster{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "CAMZOOM"
	this.MinParams = 0
	this.MaxParams = 1
	this.Raw = true
	this.NamedParams = []string{"level"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.NUMBER, "0.5"),
	}
	this.EvalSingleParam = true

	return this
}
