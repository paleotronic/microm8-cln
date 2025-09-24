package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	//	"paleotronic.com/core/memory"
	"strings"

	"paleotronic.com/utils"
	//	"paleotronic.com/log"
)

type PlusCamReset struct {
	dialect.CoreFunction
}

func (this *PlusCamReset) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	param := strings.ToUpper(params.Shift().Content)

	index := this.Interpreter.GetMemIndex()
	mm := this.Interpreter.GetMemoryMap()
	cindex := mm.GetCameraConfigure(index)
	control := types.NewOrbitController(mm, index, cindex)

	switch {
	case param == "ALL":
		control.ResetALL()
		control.Update()
	case param == "ANGLE":
		control.ResetLookAt()
		control.ResetOrientation()
		control.Update()
	case param == "LOC":
		control.ResetPosition()
		control.Update()
	case param == "ZOOM":
		control.ResetZoom()
		control.Update()
	}

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusCamReset) Syntax() string {

	/* vars */
	var result string

	result = "CAMRESET{}"

	/* enforce non void return */
	return result

}

func (this *PlusCamReset) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusCamReset(a int, b int, params types.TokenList) *PlusCamReset {
	this := &PlusCamReset{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "CAMRESET"

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "mode", Default: *types.NewToken(types.STRING, "ALL")},
		},
	)

	return this
}
