package plus

import (
	"paleotronic.com/core/dialect" //	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusRenderSuspend struct {
	dialect.CoreFunction
}

func (this *PlusRenderSuspend) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	sus := params.Shift().AsInteger() != 0

	settings.VideoSuspended = sus
	//},
	//)

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(0)))

	return nil
}

func (this *PlusRenderSuspend) Syntax() string {

	/* vars */
	var result string

	result = "CGCOLOR{col}"

	/* enforce non void return */
	return result

}

func (this *PlusRenderSuspend) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusRenderSuspend(a int, b int, params types.TokenList) *PlusRenderSuspend {
	this := &PlusRenderSuspend{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "CGCOLOR"

	//this.NamedParams = []string{ "color" }
	//this.NamedDefaults = []types.Token{ *types.NewToken( types.INTEGER, "15" ) }
	//this.Raw = true
	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "mode", Default: *types.NewToken(types.NUMBER, "1")},
		},
	)

	return this
}
