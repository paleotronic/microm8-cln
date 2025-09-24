package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
	//"paleotronic.com/fmt"
)

type PlusImageMapClear struct {
	dialect.CoreFunction
}

func (this *PlusImageMapClear) FunctionExecute(params *types.TokenList) error {

	_ = this.CoreFunction.FunctionExecute(params)

	mm := types.NewInlineImageManager( this.Interpreter.GetMemIndex(), this.Interpreter.GetMemoryMap() )
	
	mm.Empty()

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusImageMapClear) Syntax() string {

	/* vars */
	var result string

	result = "Activate{name}"

	/* enforce non void return */
	return result

}

func (this *PlusImageMapClear) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewPlusImageMapClear(a int, b int, params types.TokenList) *PlusImageMapClear {
	this := &PlusImageMapClear{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "Activate"
	this.MinParams = 0
	this.MaxParams = 0

	return this
}
