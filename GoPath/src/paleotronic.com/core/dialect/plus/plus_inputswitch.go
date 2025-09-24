package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusInputSwitch struct {
	dialect.CoreFunction
}

func (this *PlusInputSwitch) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	index := params.Shift().AsInteger()

	err := this.Interpreter.GetProducer().SetContext(index)

	if err != nil {
		this.Interpreter.PutStr(err.Error() + "\r\n")
		this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(0)))
		return err
	}

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusInputSwitch) Syntax() string {

	/* vars */
	var result string

	result = "Activate{name}"

	/* enforce non void return */
	return result

}

func (this *PlusInputSwitch) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusInputSwitch(a int, b int, params types.TokenList) *PlusInputSwitch {
	this := &PlusInputSwitch{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "Activate"

	return this
}
