package plus

import (
	"paleotronic.com/core/dialect" //	"strings"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusTurtlePots struct {
	dialect.CoreFunction
}

func (this *PlusTurtlePots) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	line := utils.StrToInt(this.ValueMap["line"].Content)

	if this.Query {

		var proc []string = this.Interpreter.GetDialect().GetDynamicCommands()
		var linetext string

		if len(proc) > 0 && line >= 0 && line < len(proc) {
			linetext = proc[line]
		}

		this.Stack.Push(types.NewToken(types.STRING, linetext))
	}

	return nil
}

func (this *PlusTurtlePots) Syntax() string {

	/* vars */
	var result string

	result = "OPEN{filename}"

	/* enforce non void return */
	return result

}

func (this *PlusTurtlePots) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusTurtlePots(a int, b int, params types.TokenList) *PlusTurtlePots {
	this := &PlusTurtlePots{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "OPEN"

	this.NamedParams = []string{"line"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.NUMBER, "0"),
	}
	this.Raw = true

	return this
}
