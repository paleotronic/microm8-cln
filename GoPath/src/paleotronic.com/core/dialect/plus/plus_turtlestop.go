package plus

import (
	"paleotronic.com/core/dialect" //	"strings"
	"paleotronic.com/core/types"
)

type PlusTurtleStop struct {
	dialect.CoreFunction
}

func (this *PlusTurtleStop) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	if this.Interpreter.GetDialect().GetShortName() == "logo" {
		this.Interpreter.GetDialect().QueueCommand("stop")
	}

	return nil
}

func (this *PlusTurtleStop) Syntax() string {

	/* vars */
	var result string

	result = "OPEN{filename}"

	/* enforce non void return */
	return result

}

func (this *PlusTurtleStop) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusTurtleStop(a int, b int, params types.TokenList) *PlusTurtleStop {
	this := &PlusTurtleStop{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "OPEN"

	this.NamedParams = []string{"cmd"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.STRING, ""),
	}
	this.Raw = true

	return this
}
