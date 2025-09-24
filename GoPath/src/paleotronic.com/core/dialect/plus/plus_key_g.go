package plus

import (
	"time"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusKey struct {
	dialect.CoreFunction
	BaseKey rune
}

func (this *PlusKey) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	if !this.Query {

		val := rune(this.ValueMap["value"].Content[0])

		_, code := this.Interpreter.GetMemoryMap().MetaKeyPeek(this.Interpreter.GetMemIndex())
		for code != 0 {
			time.Sleep(1 * time.Millisecond)
			_, code = this.Interpreter.GetMemoryMap().MetaKeyPeek(this.Interpreter.GetMemIndex())
		}

		this.Interpreter.GetMemoryMap().MetaKeySet(this.Interpreter.GetMemIndex(), this.BaseKey, val)

	}

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusKey) Syntax() string {

	/* vars */
	var result string

	result = "Activate{name}"

	/* enforce non void return */
	return result

}

func (this *PlusKey) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusKey(a int, b int, key rune, params types.TokenList) *PlusKey {
	this := &PlusKey{BaseKey: key}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "Activate"
	this.MinParams = 1
	this.MaxParams = 1

	this.NamedParams = []string{"value"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.STRING, " "),
	}
	this.Raw = true

	return this
}
