package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	//	"paleotronic.com/core/memory"
	//	"paleotronic.com/utils"
	"strings"
)

type PlusCameraLoad struct {
	dialect.CoreFunction
}

func (this *PlusCameraLoad) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	s := params.Shift().Content

	s = strings.Replace(s, "'", "\"", -1)

	index := this.Interpreter.GetMemIndex()
	mm := this.Interpreter.GetMemoryMap()
	cindex := mm.GetCameraConfigure(index)
	control := types.NewOrbitController(mm, index, cindex)

	control.FromString(s)

	this.Stack.Push(types.NewToken(types.NUMBER, "0"))

	return nil
}

func (this *PlusCameraLoad) Syntax() string {

	/* vars */
	var result string

	result = "Activate{name}"

	/* enforce non void return */
	return result

}

func (this *PlusCameraLoad) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusCameraLoad(a int, b int, params types.TokenList) *PlusCameraLoad {
	this := &PlusCameraLoad{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "Activate"
	this.MinParams = 1
	this.MaxParams = 1

	return this
}
