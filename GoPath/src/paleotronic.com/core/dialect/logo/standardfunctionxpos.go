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

type StandardFunctionXPOS struct {
	dialect.CoreFunction
	D *DialectLogo
}

func (this *StandardFunctionXPOS) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewStandardFunctionXPOS(a int, b int, params types.TokenList, d *DialectLogo) *StandardFunctionXPOS {
	this := &StandardFunctionXPOS{D: d}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "XPOS"
	this.MinParams = 0
	this.MaxParams = 0

	return this
}

func (this *StandardFunctionXPOS) FunctionExecute(params *types.TokenList) error {

	/* vars */
	//	var b, a float64

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	t := this.D.Driver.GetTurtle()

	this.Stack.Push(types.NewToken(types.WORD, utils.FormatFloat("", apple2helpers.VECTOR(this.Interpreter).GetTurtle(t).Position[0])))

	return nil
}

func (this *StandardFunctionXPOS) Syntax() string {

	/* vars */
	var result string

	result = "XPOS word list"

	/* enforce non void return */
	return result

}
