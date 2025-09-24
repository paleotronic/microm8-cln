package logo

import (
	//	"strings"
	//	"math"
	"errors"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type StandardFunctionASCII struct {
	dialect.CoreFunction
}

func (this *StandardFunctionASCII) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewStandardFunctionASCII(a int, b int, params types.TokenList) *StandardFunctionASCII {
	this := &StandardFunctionASCII{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "ASCII"
	this.MinParams = 1
	this.MaxParams = 1

	return this
}

func (this *StandardFunctionASCII) FunctionExecute(params *types.TokenList) error {

	/* vars */
	//	var b, a float64

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	o := params.Shift()
	if o == nil {
		return errors.New("I NEED AN OBJECT")
	}

	if o.Type != types.LIST && o.Content != "" {
		this.Stack.Push(types.NewToken(types.WORD, utils.IntToStr(int(o.Content[0]))))
	} else {
		this.Stack.Push(types.NewToken(types.WORD, "0"))
	}

	return nil
}

func (this *StandardFunctionASCII) Syntax() string {

	/* vars */
	var result string

	result = "ASCII object"

	/* enforce non void return */
	return result

}
