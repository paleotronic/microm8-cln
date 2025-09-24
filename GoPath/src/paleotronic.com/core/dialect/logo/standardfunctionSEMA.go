package logo

import (
	//	"strings"
	//	"time"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	//	"paleotronic.com/core/interfaces"
	//	"paleotronic.com/runestring"
)

type StandardFunctionSEMA struct {
	dialect.CoreFunction
}

func (this *StandardFunctionSEMA) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewStandardFunctionSEMA(a int, b int, params types.TokenList) *StandardFunctionSEMA {
	this := &StandardFunctionSEMA{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "POS3"
	this.MinParams = 0
	this.MaxParams = 0

	return this
}

func (this *StandardFunctionSEMA) FunctionExecute(params *types.TokenList) error {

	/* vars */
	//	var b, a float64

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	this.Stack.Push(types.NewToken(types.STRING, this.Interpreter.GetSemaphore()))

	return nil
}

func (this *StandardFunctionSEMA) Syntax() string {

	/* vars */
	var result string

	result = "POS3 word list"

	/* enforce non void return */
	return result

}
