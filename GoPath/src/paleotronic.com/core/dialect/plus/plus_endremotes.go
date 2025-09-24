package plus

import (
	"paleotronic.com/core/dialect"
	//	"paleotronic.com/core/memory"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusEndRemotes struct {
	dialect.CoreFunction
}

func (this *PlusEndRemotes) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	this.Interpreter.GetProducer().CallEndRemotes()

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusEndRemotes) Syntax() string {

	/* vars */
	var result string

	result = "SETCLASSICCPU{}"

	/* enforce non void return */
	return result

}

func (this *PlusEndRemotes) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewPlusEndRemotes(a int, b int, params types.TokenList) *PlusEndRemotes {
	this := &PlusEndRemotes{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "SETCLASSICCPU"

	this.MaxParams = 0
	this.MinParams = 0

	return this
}
