package logo

import (
	//	"strings"
	//	"math"
	"errors"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	//	"paleotronic.com/utils"
)

type StandardFunctionBUTTONP struct {
	dialect.CoreFunction
}

func (this *StandardFunctionBUTTONP) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewStandardFunctionBUTTONP(a int, b int, params types.TokenList) *StandardFunctionBUTTONP {
	this := &StandardFunctionBUTTONP{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "LISTP"
	this.MinParams = 1
	this.MaxParams = 1

	return this
}

func (this *StandardFunctionBUTTONP) FunctionExecute(params *types.TokenList) error {

	/* vars */
	//	var b, a float64

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	o := params.Shift()
	if o == nil {
		return errors.New("I NEED AN OBJECT")
	}

	pnum := o.AsInteger()
	pressed := this.Interpreter.GetMemoryMap().IntGetPaddleButton(this.Interpreter.GetMemIndex(), pnum) != 0

	if pressed {
		this.Stack.Push(types.NewToken(types.NUMBER, "1"))
	} else {
		this.Stack.Push(types.NewToken(types.NUMBER, "0"))
	}

	return nil
}

func (this *StandardFunctionBUTTONP) Syntax() string {

	/* vars */
	var result string

	result = "LISTP object"

	/* enforce non void return */
	return result

}
