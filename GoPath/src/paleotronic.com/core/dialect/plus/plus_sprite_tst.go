package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusSpriteTest struct {
	dialect.CoreFunction
}

func (this *PlusSpriteTest) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	var tmp types.Token
	tmp = this.ValueMap["sprite"]
	sprite := tmp.AsInteger()
	tmp = this.ValueMap["x"]
	x := tmp.AsInteger()
	tmp = this.ValueMap["y"]
	y := tmp.AsInteger()
	controller := GetSpriteController(this.Interpreter)

	_, _, _, _, scl, bounds, _ := controller.GetSpriteAttr(sprite)
	ids := controller.FindCollidingSprites(x, y, bounds, scl, sprite)

	if len(ids) > 0 {
		this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))
	} else {
		this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(0)))
	}

	return nil
}

func (this *PlusSpriteTest) Syntax() string {

	/* vars */
	var result string

	result = "BGCOLOR{col}"

	/* enforce non void return */
	return result

}

func (this *PlusSpriteTest) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	//result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusSpriteTest(a int, b int, params types.TokenList) *PlusSpriteTest {
	this := &PlusSpriteTest{}

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
