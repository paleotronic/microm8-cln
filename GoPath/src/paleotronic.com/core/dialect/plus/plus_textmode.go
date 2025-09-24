package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusTextMode struct {
	dialect.CoreFunction
}

func (this *PlusTextMode) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	if !this.Query {
		t := this.ValueMap["mode"]
		c := t.AsInteger()
		apple2helpers.SetTextSize(this.Interpreter, uint64(c%16))
	}

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(0)))

	return nil
}

func (this *PlusTextMode) Syntax() string {

	/* vars */
	var result string

	result = "TEXTMODE{n}"

	/* enforce non void return */
	return result

}

func (this *PlusTextMode) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusTextMode(a int, b int, params types.TokenList) *PlusTextMode {
	this := &PlusTextMode{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "TEXTMODE"

	this.NamedParams = []string{"mode"}
	this.NamedDefaults = []types.Token{*types.NewToken(types.INTEGER, "0")}
	this.Raw = true

	return this
}
