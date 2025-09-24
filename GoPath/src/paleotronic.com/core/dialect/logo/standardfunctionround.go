package logo

import (
	"math"
	//	"math"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type StandardFunctionROUND struct {
	dialect.CoreFunction
}

func (this *StandardFunctionROUND) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewStandardFunctionROUND(a int, b int, params types.TokenList) *StandardFunctionROUND {
	this := &StandardFunctionROUND{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "ROUND"
	this.MinParams = 1
	this.MaxParams = 1

	return this
}

func (this *StandardFunctionROUND) FunctionExecute(params *types.TokenList) error {

	/* vars */
	var a float64

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	a = this.Stack.Pop().AsFloat()

	if a >= 0 {
		a = math.Floor(a + 0.5)
	} else {
		a = -math.Floor(math.Abs(a) + 0.5)
	}

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(int(a))))

	return nil
}

func (this *StandardFunctionROUND) Syntax() string {

	/* vars */
	var result string

	result = "ROUND a"

	/* enforce non void return */
	return result

}
