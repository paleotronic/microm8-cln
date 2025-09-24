package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusSpriteReset struct {
	dialect.CoreFunction
}

func (this *PlusSpriteReset) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	controller := GetSpriteController(this.Interpreter)

	if !this.Query {
		controller.Reset()
	}

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(0)))

	return nil
}

func (this *PlusSpriteReset) Syntax() string {

	/* vars */
	var result string

	result = "BGCOLOR{col}"

	/* enforce non void return */
	return result

}

func (this *PlusSpriteReset) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	//result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusSpriteReset(a int, b int, params types.TokenList) *PlusSpriteReset {
	this := &PlusSpriteReset{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "BGCOLOR"

	this.MaxParams = 0
	this.MinParams = 0

	this.InitNamedParams(
		[]dialect.FunctionParamDef{},
	)

	return this
}
