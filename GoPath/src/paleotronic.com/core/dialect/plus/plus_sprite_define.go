package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusSpriteDefine struct {
	dialect.CoreFunction
}

func (this *PlusSpriteDefine) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	var tmp types.Token
	tmp = this.ValueMap["sprite"]
	sprite := tmp.AsInteger()
	tmp = this.ValueMap["def"]
	value := tmp.Content
	controller := GetSpriteController(this.Interpreter)

	if !this.Query {
		controller.SetDefinition(sprite, utils.Flatten7Bit(value))
	}
	this.Stack.Push(types.NewToken(types.STRING, controller.GetDefinition(sprite)))

	return nil
}

func (this *PlusSpriteDefine) Syntax() string {

	/* vars */
	var result string

	result = "BGCOLOR{col}"

	/* enforce non void return */
	return result

}

func (this *PlusSpriteDefine) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	//result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusSpriteDefine(a int, b int, params types.TokenList) *PlusSpriteDefine {
	this := &PlusSpriteDefine{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "BGCOLOR"

	this.MaxParams = 3
	this.MinParams = 1

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "sprite", Default: *types.NewToken(types.NUMBER, "0")},
			dialect.FunctionParamDef{Name: "def", Default: *types.NewToken(types.STRING, "")},
		},
	)

	return this
}
