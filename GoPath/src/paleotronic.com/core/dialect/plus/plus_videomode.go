package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusVideoMode struct {
	dialect.CoreFunction
}

func (this *PlusVideoMode) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusVideoMode) Syntax() string {

	/* vars */
	var result string

	result = "VIDEOMODE{n}"

	/* enforce non void return */
	return result

}

func (this *PlusVideoMode) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusVideoMode(a int, b int, params types.TokenList) *PlusVideoMode {
	this := &PlusVideoMode{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "VIDEOMODE"

	this.NamedParams = []string{ "mode" }
	this.NamedDefaults = []types.Token{ *types.NewToken(types.INTEGER, "0") }
	this.Raw = true

	return this
}
