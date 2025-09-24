package plus

import (
	"strings"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
)

type PlusStringLower struct {
	dialect.CoreFunction
}

func (this *PlusStringLower) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	this.Stack.Push(types.NewToken(types.STRING, strings.ToLower(this.ValueMap["str"].Content)))

	return nil
}

func (this *PlusStringLower) Syntax() string {

	/* vars */
	var result string

	result = "BGCOLOR{col}"

	/* enforce non void return */
	return result

}

func (this *PlusStringLower) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusStringLower(a int, b int, params types.TokenList) *PlusStringLower {
	this := &PlusStringLower{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "BGCOLOR"

	this.NamedParams = []string{"str"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.STRING, ""),
	}
	this.Raw = true

	return this
}
