package plus

import (
	"paleotronic.com/core/dialect" //	"strings"
	"paleotronic.com/core/types"
)

type PlusTurtleSendCommand struct {
	dialect.CoreFunction
	Command string
}

func (this *PlusTurtleSendCommand) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	if this.Interpreter.GetDialect().GetShortName() == "logo" {
		this.Interpreter.GetDialect().QueueCommand(this.Command)
	}

	return nil
}

func (this *PlusTurtleSendCommand) Syntax() string {

	/* vars */
	var result string

	result = "OPEN{filename}"

	/* enforce non void return */
	return result

}

func (this *PlusTurtleSendCommand) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusTurtleSendCommand(a int, b int, params types.TokenList, command string) *PlusTurtleSendCommand {
	this := &PlusTurtleSendCommand{Command: command}

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
