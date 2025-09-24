package plus

import (
//	"strings"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
)

type PlusProgramDir struct {
	dialect.CoreFunction
}

func (this *PlusProgramDir) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	this.Stack.Push(types.NewToken(types.STRING, this.Interpreter.GetProgramDir()))

	return nil
}

func (this *PlusProgramDir) Syntax() string {

	/* vars */
	var result string

	result = "PROGRAMDIR{}"

	/* enforce non void return */
	return result

}

func (this *PlusProgramDir) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusProgramDir(a int, b int, params types.TokenList) *PlusProgramDir {
	this := &PlusProgramDir{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "PROGRAMDIR"

	this.NamedParams = []string{"path"}
	this.NamedDefaults = []types.Token{*types.NewToken(types.STRING, "")}
	this.Raw = true

	return this
}
