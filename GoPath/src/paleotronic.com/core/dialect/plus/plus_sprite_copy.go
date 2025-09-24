package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusSpriteCopy struct {
	dialect.CoreFunction
}

func (this *PlusSpriteCopy) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	var tmp types.Token
	tmp = this.ValueMap["src"]
	src := tmp.AsInteger()
	tmp = this.ValueMap["dest"]
	dest := tmp.AsInteger()
	controller := GetSpriteController(this.Interpreter)

	if !this.Query {
		data := controller.GetSpriteData(src)
		controller.SetSpriteData(dest, data)
	}
	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(0)))

	return nil
}

func (this *PlusSpriteCopy) Syntax() string {

	/* vars */
	var result string

	result = "BGCOLOR{col}"

	/* enforce non void return */
	return result

}

func (this *PlusSpriteCopy) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	//result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusSpriteCopy(a int, b int, params types.TokenList) *PlusSpriteCopy {
	this := &PlusSpriteCopy{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "BGCOLOR"

	this.MaxParams = 3
	this.MinParams = 1

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "src", Default: *types.NewToken(types.NUMBER, "0")},
			dialect.FunctionParamDef{Name: "dest", Default: *types.NewToken(types.NUMBER, "0")},
		},
	)

	return this
}
