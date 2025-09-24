package logo

import (
	//	"math"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type StandardFunctionKEYP struct {
	dialect.CoreFunction
}

func (this *StandardFunctionKEYP) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewStandardFunctionKEYP(a int, b int, params types.TokenList) *StandardFunctionKEYP {
	this := &StandardFunctionKEYP{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "KEYP"
	this.MinParams = 0
	this.MaxParams = 0

	return this
}

func (this *StandardFunctionKEYP) FunctionExecute(params *types.TokenList) error {

	/* vars */
	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	t := (this.Interpreter.GetMemoryMap().KeyBufferSize(this.Interpreter.GetMemIndex()) != 0)

	if t {
		this.Stack.Push(types.NewToken(types.NUMBER, utils.FormatFloat("", 1)))
	} else {
		this.Stack.Push(types.NewToken(types.NUMBER, utils.FormatFloat("", 0)))
	}

	return nil
}

func (this *StandardFunctionKEYP) Syntax() string {

	/* vars */
	var result string

	result = "KEYP a b"

	/* enforce non void return */
	return result

}
