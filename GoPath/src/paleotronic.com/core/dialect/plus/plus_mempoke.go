package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusMemPoke struct {
	dialect.CoreFunction
}

func (this *PlusMemPoke) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	address := this.Stack.Shift().AsInteger()
	value := uint64(this.Stack.Shift().AsInteger())

	this.Interpreter.GetMemoryMap().WriteInterpreterMemory(
		this.Interpreter.GetMemIndex(),
		address,
		value,
	)

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusMemPoke) Syntax() string {

	/* vars */
	var result string

	result = "Activate{name}"

	/* enforce non void return */
	return result

}

func (this *PlusMemPoke) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusMemPoke(a int, b int, params types.TokenList) *PlusMemPoke {
	this := &PlusMemPoke{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "Activate"
	this.MinParams = 2
	this.MaxParams = 2

	this.NamedParams = []string{
		"address",
		"value",
	}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.NUMBER, "1024"),
		*types.NewToken(types.NUMBER, "0"),
	}
	this.Raw = true

	return this
}
