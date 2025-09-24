package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusBootTime struct {
	dialect.CoreFunction
}

func (this *PlusBootTime) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	c := int(this.Interpreter.GetStartTime().Unix())

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(c)))

	return nil
}

func (this *PlusBootTime) Syntax() string {

	/* vars */
	var result string

	result = "BOOTTIME{v}"

	/* enforce non void return */
	return result

}

func (this *PlusBootTime) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
//	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusBootTime(a int, b int, params types.TokenList) *PlusBootTime {
	this := &PlusBootTime{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "UPTIME"

	return this
}
