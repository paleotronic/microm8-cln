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

type StandardFunctionZPOS struct {
	dialect.CoreFunction
	D *DialectLogo
}

func (this *StandardFunctionZPOS) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewStandardFunctionZPOS(a int, b int, params types.TokenList, d *DialectLogo) *StandardFunctionZPOS {
	this := &StandardFunctionZPOS{D: d}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "ZPOS"
	this.MinParams = 0
	this.MaxParams = 0

	return this
}

func (this *StandardFunctionZPOS) FunctionExecute(params *types.TokenList) error {

	/* vars */
	//	var b, a float64

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	this.Stack.Push(types.NewToken(types.WORD, utils.FormatFloat("", apple2helpers.VECTOR(this.Interpreter).GetTurtle(this.D.Driver.GetTurtle()).Position[2])))

	return nil
}

func (this *StandardFunctionZPOS) Syntax() string {

	/* vars */
	var result string

	result = "ZPOS word list"

	/* enforce non void return */
	return result

}
