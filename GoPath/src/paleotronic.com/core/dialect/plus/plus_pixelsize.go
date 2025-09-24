package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	//	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/utils"
)

type PlusPixelSize struct {
	dialect.CoreFunction
}

func (this *PlusPixelSize) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	if !this.Query {
		t := this.ValueMap["size"]
		c := t.AsInteger()
		this.Interpreter.GetMemoryMap().SetPixelSize(uint64(c))
	}
	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(0)))

	return nil
}

func (this *PlusPixelSize) Syntax() string {

	/* vars */
	var result string

	result = "BGCOLOR{col}"

	/* enforce non void return */
	return result

}

func (this *PlusPixelSize) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusPixelSize(a int, b int, params types.TokenList) *PlusPixelSize {
	this := &PlusPixelSize{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "BGCOLOR"

	this.NamedParams = []string{"size"}
	this.NamedDefaults = []types.Token{*types.NewToken(types.INTEGER, "4")}
	this.Raw = true

	return this
}
