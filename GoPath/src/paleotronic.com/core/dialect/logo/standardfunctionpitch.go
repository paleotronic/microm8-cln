package logo

import (
	//	"strings"
	//	"time"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"

	//	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/hardware/apple2helpers"
	//	"paleotronic.com/runestring"
	"paleotronic.com/utils"
)

type StandardFunctionPITCH struct {
	dialect.CoreFunction
	D *DialectLogo
}

func (this *StandardFunctionPITCH) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewStandardFunctionPITCH(a int, b int, params types.TokenList, d *DialectLogo) *StandardFunctionPITCH {
	this := &StandardFunctionPITCH{D: d}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "PITCH"
	this.MinParams = 0
	this.MaxParams = 0

	return this
}

func (this *StandardFunctionPITCH) FunctionExecute(params *types.TokenList) error {

	/* vars */
	//	var b, a float64

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	this.Stack.Push(types.NewToken(types.WORD, utils.FormatFloat("", apple2helpers.VECTOR(this.Interpreter).GetTurtle(this.D.Driver.GetTurtle()).Pitch)))

	return nil
}

func (this *StandardFunctionPITCH) Syntax() string {

	/* vars */
	var result string

	result = "PITCH"

	/* enforce non void return */
	return result

}
