package plus

import (
//	"strings"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
)

type PlusParam struct {
	dialect.CoreFunction
}

func (this *PlusParam) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }


	t := this.Stack.Shift()
	i := t.AsInteger()
	if i<0 || i >= this.Interpreter.GetParams().Size() {
		this.Stack.Push(types.NewToken(types.STRING, ""))
	} else {
		this.Stack.Push(types.NewToken(types.STRING, this.Interpreter.GetParams().Get(i).Content))
	}


	return nil
}

func (this *PlusParam) Syntax() string {

	/* vars */
	var result string

	result = "PARAM{index}"

	/* enforce non void return */
	return result

}

func (this *PlusParam) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusParam(a int, b int, params types.TokenList) *PlusParam {
	this := &PlusParam{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "PARAM"

	this.NamedParams = []string{"index"}
	this.NamedDefaults = []types.Token{*types.NewToken(types.INTEGER, "0")}
	this.Raw = false

	return this
}
