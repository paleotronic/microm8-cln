package plus

import (
	"time"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusUpTime struct {
	dialect.CoreFunction
}

func (this *PlusUpTime) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	c := int(time.Since(this.Interpreter.GetStartTime()).Nanoseconds()/1000000)

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(c)))

	return nil
}

func (this *PlusUpTime) Syntax() string {

	/* vars */
	var result string

	result = "UPTIME{v}"

	/* enforce non void return */
	return result

}

func (this *PlusUpTime) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
//	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusUpTime(a int, b int, params types.TokenList) *PlusUpTime {
	this := &PlusUpTime{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "UPTIME"

	return this
}
