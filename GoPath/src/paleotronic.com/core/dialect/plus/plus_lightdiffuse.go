package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusLightDiffuse struct {
	dialect.CoreFunction
}

func (this *PlusLightDiffuse) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	if !this.Query {
		t := this.ValueMap["level"]
		c := t.AsExtended()

		this.Interpreter.GetMemoryMap().IntSetDiffuseLevel(this.Interpreter.GetMemIndex(), float32(c))
	}
	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusLightDiffuse) Syntax() string {

	/* vars */
	var result string

	result = "CPUTHROTTLE{v}"

	/* enforce non void return */
	return result

}

func (this *PlusLightDiffuse) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusLightDiffuse(a int, b int, params types.TokenList) *PlusLightDiffuse {
	this := &PlusLightDiffuse{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "CPUTHROTTLE"

	this.NamedParams = []string{"level"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.NUMBER, "1.0"),
	}
	this.Raw = true

	return this
}
