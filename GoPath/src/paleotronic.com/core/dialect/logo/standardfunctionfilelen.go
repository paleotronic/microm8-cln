package logo

import (
	//	"strings"
	//	"time"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	//	"paleotronic.com/core/interfaces"
	//	"paleotronic.com/core/hardware/apple2helpers"
	//	"paleotronic.com/runestring"
	"paleotronic.com/files"
	"paleotronic.com/utils"
)

type StandardFunctionFILELEN struct {
	dialect.CoreFunction
}

func (this *StandardFunctionFILELEN) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)
	result = append(result, types.WORD)

	/* enforce non void return */
	return result

}

func NewStandardFunctionFILELEN(a int, b int, params types.TokenList) *StandardFunctionFILELEN {
	this := &StandardFunctionFILELEN{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "FILELEN"
	this.MinParams = 1
	this.MaxParams = 1

	return this
}

func (this *StandardFunctionFILELEN) FunctionExecute(params *types.TokenList) error {

	/* vars */
	//	var b, a float64

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	rtok, _ := this.Interpreter.GetDialect().ParseTokensForResult(this.Interpreter, *params)

	s, e := files.DOSLEN(this.Interpreter.GetWorkDir(), rtok.Content)

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(s)))

	return e
}

func (this *StandardFunctionFILELEN) Syntax() string {

	/* vars */
	var result string

	result = "FILELEN word list"

	/* enforce non void return */
	return result

}
