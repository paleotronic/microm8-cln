package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
//	"paleotronic.com/utils"
)

type PlusSplash struct {
	dialect.CoreFunction
}

var splash string

func (this *PlusSplash) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	this.Stack.Push(types.NewToken(types.STRING, splash))

	return nil
}

func (this *PlusSplash) Syntax() string {

	/* vars */
	var result string

	result = "SPLASH{image}"

	/* enforce non void return */
	return result

}

func (this *PlusSplash) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusSplash(a int, b int, params types.TokenList) *PlusSplash {
	this := &PlusSplash{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "SPLASH"

	this.NamedParams = []string{ "image" }
	this.NamedDefaults = []types.Token{ *types.NewToken( types.STRING, "" ) }
	this.Raw = true

	return this
}
