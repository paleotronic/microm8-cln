package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusSpriteScale struct {
	dialect.CoreFunction
}

func (this *PlusSpriteScale) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	var tmp types.Token
	tmp = this.ValueMap["sprite"]
	sprite := tmp.AsInteger()
	tmp = this.ValueMap["scale"]
	value := types.SpriteScale(tmp.AsInteger() - 1)
	controller := GetSpriteController(this.Interpreter)

	if !this.Query {
		controller.SetScale(sprite, value)
	}
	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(int(controller.GetScale(sprite)+1))))

	return nil
}

func (this *PlusSpriteScale) Syntax() string {

	/* vars */
	var result string

	result = "BGCOLOR{col}"

	/* enforce non void return */
	return result

}

func (this *PlusSpriteScale) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	//result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusSpriteScale(a int, b int, params types.TokenList) *PlusSpriteScale {
	this := &PlusSpriteScale{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "BGCOLOR"

	this.MaxParams = 3
	this.MinParams = 1

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "sprite", Default: *types.NewToken(types.NUMBER, "0")},
			dialect.FunctionParamDef{Name: "scale", Default: *types.NewToken(types.NUMBER, "1")},
		},
	)

	return this
}
