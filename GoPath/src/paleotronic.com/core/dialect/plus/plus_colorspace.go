package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusColorSpace struct {
	dialect.CoreFunction
}

func (this *PlusColorSpace) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	if !this.Query {
		q := this.ValueMap["mode"]
		PNGColorSpace = byte(q.AsInteger() % 3)
	}

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(int(PNGColorSpace))))

	return nil
}

func (this *PlusColorSpace) Syntax() string {

	/* vars */
	var result string

	result = "COLORSPACE{}"

	/* enforce non void return */
	return result

}

func (this *PlusColorSpace) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewPlusColorSpace(a int, b int, params types.TokenList) *PlusColorSpace {
	this := &PlusColorSpace{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "COLORSPACE"

	this.NamedDefaults = []types.Token{*types.NewToken(types.NUMBER, "0")}
	this.NamedParams = []string{"mode"}
	this.Raw = true

	return this
}
