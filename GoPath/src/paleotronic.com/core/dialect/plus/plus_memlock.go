package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusMemLock struct {
	dialect.CoreFunction
}

func (this *PlusMemLock) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	if !this.Query {

		var x types.Token
		x = this.ValueMap["address"]
		addr := x.AsInteger()
		x = this.ValueMap["value"]
		value := x.AsInteger()

		cpu := apple2helpers.GetCPU(this.Interpreter)
		if value == -1 {
			delete(cpu.LockValue, addr)
		} else {
			cpu.LockValue[addr] = uint64(value)
		}
	}

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(0)))

	return nil
}

func (this *PlusMemLock) Syntax() string {

	/* vars */
	var result string

	result = "Activate{name}"

	/* enforce non void return */
	return result

}

func (this *PlusMemLock) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusMemLock(a int, b int, params types.TokenList) *PlusMemLock {
	this := &PlusMemLock{}

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "Activate"
	this.MinParams = 2
	this.MaxParams = 2
	this.Raw = true

	this.NamedParams = []string{
		"address",
		"value",
	}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.NUMBER, "0"),
		*types.NewToken(types.NUMBER, "0"),
	}

	return this
}
