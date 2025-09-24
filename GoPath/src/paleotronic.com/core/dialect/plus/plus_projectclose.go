package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/files"
)

//var backdrop string

type PlusProjectClose struct {
	dialect.CoreFunction
}

func (this *PlusProjectClose) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	if !this.Query {
		files.SetProject("")
	} else {

		this.Stack.Push(types.NewToken(types.STRING, ""))

	}
	return nil
}

func (this *PlusProjectClose) Syntax() string {

	/* vars */
	var result string

	result = "PROJECT.USE{name}"

	/* enforce non void return */
	return result

}

func (this *PlusProjectClose) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusProjectClose(a int, b int, params types.TokenList) *PlusProjectClose {
	this := &PlusProjectClose{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "PROJECT.USE"

	this.NamedParams = []string{ "name" }
	this.NamedDefaults = []types.Token{ *types.NewToken( types.STRING, "" ) }
	this.Raw = true

	return this
}
