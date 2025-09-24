package plus

import (
	"paleotronic.com/core/dialect" //	"strings"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusTurtleProc struct {
	dialect.CoreFunction
}

func (this *PlusTurtleProc) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	name := this.ValueMap["name"].Content
	line := utils.StrToInt(this.ValueMap["line"].Content)

	if this.Query {

		var proc []string
		var linetext string

		if this.Interpreter.GetDialect().GetShortName() == "logo" {
			proc = this.Interpreter.GetDialect().GetDynamicCommandDef(name)
			if len(proc) == 0 {
				proc = this.Interpreter.GetDialect().GetDynamicFunctionDef(name)
			}

			if len(proc) > 0 && line >= 0 && line < len(proc) {
				linetext = proc[line]
			}

		}

		this.Stack.Push(types.NewToken(types.STRING, linetext))
	}

	return nil
}

func (this *PlusTurtleProc) Syntax() string {

	/* vars */
	var result string

	result = "OPEN{filename}"

	/* enforce non void return */
	return result

}

func (this *PlusTurtleProc) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusTurtleProc(a int, b int, params types.TokenList) *PlusTurtleProc {
	this := &PlusTurtleProc{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "OPEN"

	this.NamedParams = []string{"name", "line"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.NUMBER, ""),
		*types.NewToken(types.NUMBER, "0"),
	}
	this.Raw = true

	return this
}
