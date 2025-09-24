package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusSpriteYPos struct {
	dialect.CoreFunction
}

func (this *PlusSpriteYPos) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	var tmp types.Token
	tmp = this.ValueMap["sprite"]
	sprite := tmp.AsInteger()
	tmp = this.ValueMap["y"]
	value := tmp.AsInteger()
	controller := GetSpriteController(this.Interpreter)

	if !this.Query {
		controller.SetY(sprite, value)
	}
	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(controller.GetY(sprite))))

	return nil
}

func (this *PlusSpriteYPos) Syntax() string {

	/* vars */
	var result string

	result = "BGCOLOR{col}"

	/* enforce non void return */
	return result

}

func (this *PlusSpriteYPos) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	//result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusSpriteYPos(a int, b int, params types.TokenList) *PlusSpriteYPos {
	this := &PlusSpriteYPos{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "BGCOLOR"

	this.MaxParams = 3
	this.MinParams = 1

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "sprite", Default: *types.NewToken(types.NUMBER, "0")},
			dialect.FunctionParamDef{Name: "y", Default: *types.NewToken(types.NUMBER, "0")},
		},
	)

	return this
}
