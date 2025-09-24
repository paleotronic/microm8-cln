package plus

import (
	"paleotronic.com/core/dialect" //	"strings"
	"paleotronic.com/core/types"
)

type PlusTurtleLast struct {
	dialect.CoreFunction
}

func (this *PlusTurtleLast) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	var last string
	last = this.Interpreter.GetDialect().GetLastCommand()

	this.Stack.Push(types.NewToken(types.STRING, last))

	return nil
}

func (this *PlusTurtleLast) Syntax() string {

	/* vars */
	var result string

	result = "OPEN{filename}"

	/* enforce non void return */
	return result

}

func (this *PlusTurtleLast) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusTurtleLast(a int, b int, params types.TokenList) *PlusTurtleLast {
	this := &PlusTurtleLast{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "OPEN"
	this.MaxParams = 0
	this.MinParams = 0
	this.Raw = true

	return this
}
