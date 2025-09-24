package plus

import (
	"paleotronic.com/core/dialect" //	"strings"
	"paleotronic.com/core/types"
)

type PlusTurtleExec struct {
	dialect.CoreFunction
}

func (this *PlusTurtleExec) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	cmd := this.ValueMap["cmd"].Content
	if cmd != "" {
		if this.Interpreter.GetDialect().GetShortName() == "logo" {
			this.Interpreter.Parse(cmd)
		}
	}

	return nil
}

func (this *PlusTurtleExec) Syntax() string {

	/* vars */
	var result string

	result = "OPEN{filename}"

	/* enforce non void return */
	return result

}

func (this *PlusTurtleExec) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusTurtleExec(a int, b int, params types.TokenList) *PlusTurtleExec {
	this := &PlusTurtleExec{}

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
