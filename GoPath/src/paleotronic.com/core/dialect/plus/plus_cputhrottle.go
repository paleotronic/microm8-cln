package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusCPUThrottle struct {
	dialect.CoreFunction
}

func (this *PlusCPUThrottle) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	c := params.Shift().AsExtended()

	this.Interpreter.GetDialect().SetThrottle(float32(c))
	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusCPUThrottle) Syntax() string {

	/* vars */
	var result string

	result = "CPUTHROTTLE{v}"

	/* enforce non void return */
	return result

}

func (this *PlusCPUThrottle) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusCPUThrottle(a int, b int, params types.TokenList) *PlusCPUThrottle {
	this := &PlusCPUThrottle{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "CPUTHROTTLE"

	return this
}
