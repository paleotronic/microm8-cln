package plus

import (
	//"time"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/types"
	"paleotronic.com/log"
	"paleotronic.com/utils"
)

type PlusTextDrawPrint struct {
	dialect.CoreFunction
}

func (this *PlusTextDrawPrint) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		log.Println(e)
		return e
	}

	log.Println("In PlusTextDrawPrint()")

	txt := params.Get(0).Content

	apple2helpers.DrawGraphicsText(this.Interpreter, txt)

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusTextDrawPrint) Syntax() string {

	/* vars */
	var result string

	result = "CubePlot{x, y, z, c}"

	/* enforce non void return */
	return result

}

func (this *PlusTextDrawPrint) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusTextDrawPrint(a int, b int, params types.TokenList) *PlusTextDrawPrint {
	this := &PlusTextDrawPrint{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "CubePlot"
	this.MaxParams = 1
	this.MinParams = 1

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "str", Default: *types.NewToken(types.STRING, "")},
		},
	)

	return this
}
