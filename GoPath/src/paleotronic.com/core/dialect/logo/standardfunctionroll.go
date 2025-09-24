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

type StandardFunctionROLL struct {
	dialect.CoreFunction
	D *DialectLogo
}

func (this *StandardFunctionROLL) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewStandardFunctionROLL(a int, b int, params types.TokenList, d *DialectLogo) *StandardFunctionROLL {
	this := &StandardFunctionROLL{D: d}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "ROLL"
	this.MinParams = 0
	this.MaxParams = 0

	return this
}

func (this *StandardFunctionROLL) FunctionExecute(params *types.TokenList) error {

	/* vars */
	//	var b, a float64

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	this.Stack.Push(types.NewToken(types.WORD, utils.FormatFloat("", apple2helpers.VECTOR(this.Interpreter).GetTurtle(this.D.Driver.GetTurtle()).Roll)))

	return nil
}

func (this *StandardFunctionROLL) Syntax() string {

	/* vars */
	var result string

	result = "ROLL"

	/* enforce non void return */
	return result

}
