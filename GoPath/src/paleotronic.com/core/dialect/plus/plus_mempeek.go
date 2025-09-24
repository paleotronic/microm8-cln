package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusMemPeek struct {
	dialect.CoreFunction
}

func (this *PlusMemPeek) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	address := this.Stack.Shift().AsInteger()

	value := this.Interpreter.GetMemoryMap().ReadInterpreterMemory(
		this.Interpreter.GetMemIndex(),
		address,
	)

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(int(value))))

	return nil
}

func (this *PlusMemPeek) Syntax() string {

	/* vars */
	var result string

	result = "Activate{name}"

	/* enforce non void return */
	return result

}

func (this *PlusMemPeek) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusMemPeek(a int, b int, params types.TokenList) *PlusMemPeek {
	this := &PlusMemPeek{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "Activate"
	this.MinParams = 1
	this.MaxParams = 1

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "address", Default: *types.NewToken(types.NUMBER, "1024")},
		},
	)

	return this
}
