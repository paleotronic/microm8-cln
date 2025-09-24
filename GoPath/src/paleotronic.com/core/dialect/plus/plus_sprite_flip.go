package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusSpriteFlip struct {
	dialect.CoreFunction
}

func (this *PlusSpriteFlip) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	var tmp types.Token
	tmp = this.ValueMap["sprite"]
	sprite := tmp.AsInteger()
	tmp = this.ValueMap["flip"]
	value := types.SpriteFlip(tmp.AsInteger())
	controller := GetSpriteController(this.Interpreter)

	if !this.Query {
		controller.SetFlip(sprite, value)
	}
	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(int(controller.GetFlip(sprite)))))

	return nil
}

func (this *PlusSpriteFlip) Syntax() string {

	/* vars */
	var result string

	result = "BGCOLOR{col}"

	/* enforce non void return */
	return result

}

func (this *PlusSpriteFlip) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	//result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusSpriteFlip(a int, b int, params types.TokenList) *PlusSpriteFlip {
	this := &PlusSpriteFlip{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "BGCOLOR"

	this.MaxParams = 3
	this.MinParams = 1

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "sprite", Default: *types.NewToken(types.NUMBER, "0")},
			dialect.FunctionParamDef{Name: "flip", Default: *types.NewToken(types.NUMBER, "0")},
		},
	)

	return this
}
