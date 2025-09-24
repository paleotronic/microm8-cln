package plus

import (
	"paleotronic.com/core/dialect" //	"strings"
	"paleotronic.com/core/types"
)

type PlusTurtleSemaphore struct {
	dialect.CoreFunction
}

func (this *PlusTurtleSemaphore) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	sema := this.ValueMap["semaphore"].Content

	if this.Query {
		this.Stack.Push(types.NewToken(types.STRING, this.Interpreter.GetSemaphore()))
	} else {

		this.Interpreter.SetSemaphore(sema)

	}

	return nil
}

func (this *PlusTurtleSemaphore) Syntax() string {

	/* vars */
	var result string

	result = "OPEN{filename}"

	/* enforce non void return */
	return result

}

func (this *PlusTurtleSemaphore) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusTurtleSemaphore(a int, b int, params types.TokenList) *PlusTurtleSemaphore {
	this := &PlusTurtleSemaphore{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "OPEN"

	this.NamedParams = []string{"semaphore"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.NUMBER, ""),
	}
	this.Raw = true

	return this
}
