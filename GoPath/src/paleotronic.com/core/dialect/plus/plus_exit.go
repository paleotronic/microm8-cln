package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusExit struct {
	dialect.CoreFunction
}

func (this *PlusExit) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	this.Interpreter.PutStr("Trying to exit " + this.Interpreter.GetName())

	if this.Interpreter.GetParent() != nil {
		this.Interpreter.GetParent().SetChild(nil)
		this.Interpreter.SetParent(nil)
	}

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusExit) Syntax() string {

	/* vars */
	var result string

	result = "EXIT{}"

	/* enforce non void return */
	return result

}

func (this *PlusExit) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewPlusExit(a int, b int, params types.TokenList) *PlusExit {
	this := &PlusExit{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "EXIT"

	return this
}
