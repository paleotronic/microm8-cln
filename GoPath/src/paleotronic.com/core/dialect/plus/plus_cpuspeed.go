package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/types"
	//	"paleotronic.com/core/memory"
	"paleotronic.com/utils"
	//	"paleotronic.com/core/settings"
	//	"time"
	//	"paleotronic.com/fmt"
)

type PlusCPUSpeed struct {
	dialect.CoreFunction
}

func (this *PlusCPUSpeed) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	cpu := apple2helpers.GetCPU(this.Interpreter)

	if this.Query {

		v := cpu.GetWarp()
		this.Stack.Push(types.NewToken(types.NUMBER, utils.FloatToStr(v)))

	} else {

		zt := this.ValueMap["warp"]
		zoom := zt.AsExtended()
		if zoom > 0.1 && zoom < 8 {
			cpu.SetWarpUser(zoom)
		}

	}

	//fmt.Printf("Setting zoom to %f\n", zoom)

	return nil
}

func (this *PlusCPUSpeed) Syntax() string {

	/* vars */
	var result string

	result = "CAMZOOM{v}"

	/* enforce non void return */
	return result

}

func (this *PlusCPUSpeed) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusCPUSpeed(a int, b int, params types.TokenList) *PlusCPUSpeed {
	this := &PlusCPUSpeed{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "CAMZOOM"
	this.MinParams = 0
	this.MaxParams = 1
	this.Raw = true
	this.NamedParams = []string{"warp"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.NUMBER, "1"),
	}
	this.EvalSingleParam = true

	return this
}
