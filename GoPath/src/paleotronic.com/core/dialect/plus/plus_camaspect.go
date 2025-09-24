package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	//	"paleotronic.com/core/memory"
	//	"paleotronic.com/utils"
	//	"paleotronic.com/core/settings"
	//	"time"
	//	"paleotronic.com/fmt"
)

type PlusCamAspect struct {
	dialect.CoreFunction
}

func (this *PlusCamAspect) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	aspect := params.Shift().AsExtended()
	index := params.Shift().AsInteger() - 1
	if index < 0 {
		index = 0
	}

	mm := this.Interpreter.GetMemoryMap()
	for cindex := 0; cindex < 9; cindex++ {
		//log.Printf("camaspect: slot=%d, camid=%d -> %.2f", index, cindex, aspect)
		control := types.NewOrbitController(mm, index, cindex-1)
		control.SetAspect(aspect)
		control.Update()
	}

	//fmt.Printf("Setting zoom to %f\n", zoom)

	return nil
}

func (this *PlusCamAspect) Syntax() string {

	/* vars */
	var result string

	result = "CAMZOOM{v}"

	/* enforce non void return */
	return result

}

func (this *PlusCamAspect) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusCamAspect(a int, b int, params types.TokenList) *PlusCamAspect {
	this := &PlusCamAspect{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "CAMZOOM"
	this.MinParams = 1
	this.MaxParams = 2

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "aspect", Default: *types.NewToken(types.NUMBER, "1.46")},
			dialect.FunctionParamDef{Name: "index", Default: *types.NewToken(types.NUMBER, "1")},
		},
	)

	return this
}
