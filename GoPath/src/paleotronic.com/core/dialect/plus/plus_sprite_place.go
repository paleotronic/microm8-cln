package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusSpritePlace struct {
	dialect.CoreFunction
}

func (this *PlusSpritePlace) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	var tmp types.Token
	tmp = this.ValueMap["sprite"]
	sprite := tmp.AsInteger()
	tmp = this.ValueMap["x"]
	value1 := tmp.AsInteger()
	tmp = this.ValueMap["y"]
	value2 := tmp.AsInteger()
	controller := GetSpriteController(this.Interpreter)

	if !this.Query {
		controller.SetX(sprite, value1)
		controller.SetY(sprite, value2)
		controller.SetEnabled(sprite, true)
	}
	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(0)))

	return nil
}

func (this *PlusSpritePlace) Syntax() string {

	/* vars */
	var result string

	result = "BGCOLOR{col}"

	/* enforce non void return */
	return result

}

func (this *PlusSpritePlace) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	//result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusSpritePlace(a int, b int, params types.TokenList) *PlusSpritePlace {
	this := &PlusSpritePlace{}

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
