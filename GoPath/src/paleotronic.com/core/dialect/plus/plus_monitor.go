package plus

import (
	"paleotronic.com/core/dialect"
	//	"paleotronic.com/core/memory"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/types"
	//	"paleotronic.com/utils"
)

type PlusMonitor struct {
	dialect.CoreFunction
}

func (this *PlusMonitor) FunctionExecute(params *types.TokenList) error {

	//	if e := this.CoreFunction.FunctionExecute(params); e != nil {
	//		return e
	//	}

	m := apple2helpers.NewMonitor(this.Interpreter)
	m.Manual("")

	return nil

}

func (this *PlusMonitor) Syntax() string {

	/* vars */
	var result string

	result = "ALTCASE{n}"

	/* enforce non void return */
	return result

}

func (this *PlusMonitor) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewPlusMonitor(a int, b int, params types.TokenList) *PlusMonitor {
	this := &PlusMonitor{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "VIDEOMODE"

	this.MinParams = 0
	this.MaxParams = 0

	return this
}
