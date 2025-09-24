package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusMemCounterBump struct {
	dialect.CoreFunction
}

func (this *PlusMemCounterBump) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	var counter int

	counter = params.Shift().AsInteger()

	this.Interpreter.GetMemoryMap().IntBumpCounter(this.Interpreter.GetMemIndex(), counter)

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(int(counter))))

	return nil
}

func (this *PlusMemCounterBump) Syntax() string {

	/* vars */
	var result string

	result = "Activate{name}"

	/* enforce non void return */
	return result

}

func (this *PlusMemCounterBump) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusMemCounterBump(a int, b int, params types.TokenList) *PlusMemCounterBump {
	this := &PlusMemCounterBump{}

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "Activate"
	this.MinParams = 1
	this.MaxParams = 1

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "counter", Default: *types.NewToken(types.NUMBER, "1")},
		},
	)

	return this
}
