package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	//	"paleotronic.com/core/memory"
	"paleotronic.com/utils"
	//	"paleotronic.com/core/settings"
	//	"time"
	//	"paleotronic.com/fmt"
)

type PlusCamZoom struct {
	dialect.CoreFunction
}

func (this *PlusCamZoom) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	index := this.Interpreter.GetMemIndex()
	mm := this.Interpreter.GetMemoryMap()
	cindex := mm.GetCameraConfigure(index)
	control := types.NewOrbitController(mm, index, cindex)

	if this.Query {

		v := control.GetZoom()
		this.Stack.Push(types.NewToken(types.NUMBER, utils.FloatToStr(v)))

	} else {

		zt := this.ValueMap["zoom"]
		zoom := zt.AsExtended()
		control.SetZoom(zoom)

	}

	//fmt.Printf("Setting zoom to %f\n", zoom)

	return nil
}

func (this *PlusCamZoom) Syntax() string {

	/* vars */
	var result string

	result = "CAMZOOM{v}"

	/* enforce non void return */
	return result

}

func (this *PlusCamZoom) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusCamZoom(a int, b int, params types.TokenList) *PlusCamZoom {
	this := &PlusCamZoom{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "CAMZOOM"
	this.MinParams = 0
	this.MaxParams = 1
	this.Raw = true
	this.NamedParams = []string{"zoom"}
	this.NamedDefaults = []types.Token{
		*types.NewToken(types.INTEGER, "1"),
	}
	this.EvalSingleParam = true

	return this
}
