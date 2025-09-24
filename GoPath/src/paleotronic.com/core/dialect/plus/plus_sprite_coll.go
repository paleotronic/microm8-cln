package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusSpriteCollision struct {
	dialect.CoreFunction
}

func (this *PlusSpriteCollision) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	var tmp types.Token
	tmp = this.ValueMap["sprite"]
	sprite := tmp.AsInteger()
	controller := GetSpriteController(this.Interpreter)

	x, y, _, _, scl, bounds, _ := controller.GetSpriteAttr(sprite)
	ids := controller.FindCollidingSprites(x, y, bounds, scl, sprite)

	if len(ids) > 0 {
		this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(ids[0])))
	} else {
		this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(-1)))
	}

	return nil
}

func (this *PlusSpriteCollision) Syntax() string {

	/* vars */
	var result string

	result = "BGCOLOR{col}"

	/* enforce non void return */
	return result

}

func (this *PlusSpriteCollision) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	//result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusSpriteCollision(a int, b int, params types.TokenList) *PlusSpriteCollision {
	this := &PlusSpriteCollision{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "BGCOLOR"

	this.MaxParams = 3
	this.MinParams = 1

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "sprite", Default: *types.NewToken(types.NUMBER, "0")},
			dialect.FunctionParamDef{Name: "x", Default: *types.NewToken(types.NUMBER, "0")},
			dialect.FunctionParamDef{Name: "y", Default: *types.NewToken(types.NUMBER, "0")},
		},
	)

	return this
}
