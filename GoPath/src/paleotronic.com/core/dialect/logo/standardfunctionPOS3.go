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

type StandardFunctionPOS3 struct {
	dialect.CoreFunction
	D *DialectLogo
}

func (this *StandardFunctionPOS3) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewStandardFunctionPOS3(a int, b int, params types.TokenList, d *DialectLogo) *StandardFunctionPOS3 {
	this := &StandardFunctionPOS3{D: d}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "POS3"
	this.MinParams = 0
	this.MaxParams = 0

	return this
}

func (this *StandardFunctionPOS3) FunctionExecute(params *types.TokenList) error {

	/* vars */
	//	var b, a float64

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	l := types.NewToken(types.LIST, "")
	l.List = types.NewTokenList()
	l.List.Push(types.NewToken(types.NUMBER, utils.FormatFloat("", apple2helpers.VECTOR(this.Interpreter).GetTurtle(this.D.Driver.GetTurtle()).Position[0])))
	l.List.Push(types.NewToken(types.NUMBER, utils.FormatFloat("", apple2helpers.VECTOR(this.Interpreter).GetTurtle(this.D.Driver.GetTurtle()).Position[1])))
	l.List.Push(types.NewToken(types.NUMBER, utils.FormatFloat("", apple2helpers.VECTOR(this.Interpreter).GetTurtle(this.D.Driver.GetTurtle()).Position[2])))

	this.Stack.Push(l)

	return nil
}

func (this *StandardFunctionPOS3) Syntax() string {

	/* vars */
	var result string

	result = "POS3 word list"

	/* enforce non void return */
	return result

}
