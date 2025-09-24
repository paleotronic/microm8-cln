package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusSpriteOn struct {
	dialect.CoreFunction
}

func GetSpriteController(e interfaces.Interpretable) *types.SpriteController {
	return types.NewSpriteController(
		e.GetMemIndex(),
		e.GetMemoryMap(),
		memory.MICROM8_SPRITE_CONTROL_BASE,
	)
}

func (this *PlusSpriteOn) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	var tmp types.Token
	tmp = this.ValueMap["sprite"]
	sprite := tmp.AsInteger()
	controller := GetSpriteController(this.Interpreter)

	if !this.Query {
		controller.SetEnabled(sprite, true)
	}
	if controller.GetEnabled(sprite) {
		this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))
	} else {
		this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(0)))
	}

	return nil
}

func (this *PlusSpriteOn) Syntax() string {

	/* vars */
	var result string

	result = "BGCOLOR{col}"

	/* enforce non void return */
	return result

}

func (this *PlusSpriteOn) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	//result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusSpriteOn(a int, b int, params types.TokenList) *PlusSpriteOn {
	this := &PlusSpriteOn{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "BGCOLOR"

	this.MaxParams = 3
	this.MinParams = 1

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "sprite", Default: *types.NewToken(types.NUMBER, "0")},
		},
	)

	return this
}
