package plus

import (
	"paleotronic.com/core/dialect"
	//	"paleotronic.com/core/memory"
	"paleotronic.com/core/types"
	//	"paleotronic.com/utils"
)

type PlusUpperCase struct {
	dialect.CoreFunction
}

func (this *PlusUpperCase) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	if !this.Query {
		t := this.ValueMap["mode"]
		c := t.AsInteger() != 0
		this.Interpreter.GetMemoryMap().IntSetUppercaseOnly(this.Interpreter.GetMemIndex(), c)
	}

	return nil

}

func (this *PlusUpperCase) Syntax() string {

	/* vars */
	var result string

	result = "ALTCASE{n}"

	/* enforce non void return */
	return result

}

func (this *PlusUpperCase) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusUpperCase(a int, b int, params types.TokenList) *PlusUpperCase {
	this := &PlusUpperCase{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "VIDEOMODE"

	this.NamedParams = []string{"mode"}
	this.NamedDefaults = []types.Token{*types.NewToken(types.INTEGER, "0")}
	this.Raw = true

	return this
}
