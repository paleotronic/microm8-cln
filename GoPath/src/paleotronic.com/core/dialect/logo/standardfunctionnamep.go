package logo

import (
	"strings"
	//	"math"
	"errors"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	//	"paleotronic.com/utils"
)

type StandardFunctionNAMEP struct {
	dialect.CoreFunction
}

func (this *StandardFunctionNAMEP) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewStandardFunctionNAMEP(a int, b int, params types.TokenList) *StandardFunctionNAMEP {
	this := &StandardFunctionNAMEP{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "NAMEP"
	this.MinParams = 1
	this.MaxParams = 1

	return this
}

func (this *StandardFunctionNAMEP) FunctionExecute(params *types.TokenList) error {

	/* vars */
	//	var b, a float64

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	o := params.Shift()
	if o == nil {
		return errors.New("I NEED AN OBJECT")
	}

	name := o.Content
	if !strings.HasPrefix(name, ":") {
		name = ":" + name
	}

	if this.Interpreter.GetData(name) != nil {
		this.Stack.Push(types.NewToken(types.NUMBER, "1"))
	} else {
		this.Stack.Push(types.NewToken(types.NUMBER, "0"))
	}

	return nil
}

func (this *StandardFunctionNAMEP) Syntax() string {

	/* vars */
	var result string

	result = "NAMEP object"

	/* enforce non void return */
	return result

}
