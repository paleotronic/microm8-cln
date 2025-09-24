package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type StandardFunctionMIDDollar struct {
	dialect.CoreFunction
}

func (this *StandardFunctionMIDDollar) Syntax() string {

	/* vars */
	var result string

	result = "MID$(<string>,<number>,<number>)"

	/* enforce non void return */
	return result

}

func (this *StandardFunctionMIDDollar) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	result = append(result, types.STRING)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewStandardFunctionMIDDollar(a int, b int, params types.TokenList) *StandardFunctionMIDDollar {
	this := &StandardFunctionMIDDollar{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "MID$"
	this.MinParams = 2
	this.MaxParams = 3

	return this
}

func (this *StandardFunctionMIDDollar) FunctionExecute(params *types.TokenList) error {

	/* vars */
	var s string
	var n int
	var c int

	c = -1

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	if this.Stack.Size() == 3 {
		c = this.Stack.Pop().AsInteger()
	}

	n = this.Stack.Pop().AsInteger()
	s = this.Stack.Pop().Content

	if c == -1 {
		c = utils.Len(s) - n + 1
	}

	this.Stack.Push(types.NewToken(types.STRING, utils.Copy(s, n, c)))

	return nil
}
