package plus

import (
	"paleotronic.com/log"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusTest struct {
	dialect.CoreFunction
}

func (this *PlusTest) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	for k, v := range this.ValueMap {
		log.Printf("%s == %s\n", k, v.AsString())
	}

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusTest) Syntax() string {

	/* vars */
	var result string

	result = "TEST{x,y,z,frog}"

	/* enforce non void return */
	return result

}

func (this *PlusTest) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusShim(a int, b int, params types.TokenList) *PlusTest {
	this := &PlusTest{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "TEST"

	this.NamedParams = []string{ "x", "y", "z", "frog" }
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.NUMBER, "1"),
		*types.NewToken(types.NUMBER, "2"),
		*types.NewToken(types.NUMBER, "3"),
		*types.NewToken(types.STRING, "frog"),
	}
	this.Raw = true

	return this
}
