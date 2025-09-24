package plus

import (
	//"time"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusCubeGR struct {
	dialect.CoreFunction
}

func (this *PlusCubeGR) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	apple2helpers.HTab(this.Interpreter, 1)
	apple2helpers.VTab(this.Interpreter, 21)
	apple2helpers.ClearToBottom(this.Interpreter)

	// enable vector layer
	apple2helpers.CubeLayerEnableCustom(this.Interpreter, true, true, 80, 48)
	apple2helpers.CUBE(this.Interpreter).Clear()
	apple2helpers.CUBE(this.Interpreter).Render()

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusCubeGR) Syntax() string {

	/* vars */
	var result string

	result = "CubeGR{}"

	/* enforce non void return */
	return result

}

func (this *PlusCubeGR) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusCubeGR(a int, b int, params types.TokenList) *PlusCubeGR {
	this := &PlusCubeGR{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "CubeGR"

	return this
}
