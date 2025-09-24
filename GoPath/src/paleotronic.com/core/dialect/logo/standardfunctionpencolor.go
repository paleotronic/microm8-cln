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

type StandardFunctionPENCOLOR struct {
	dialect.CoreFunction
	D *DialectLogo
}

func (this *StandardFunctionPENCOLOR) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewStandardFunctionPENCOLOR(a int, b int, params types.TokenList, d *DialectLogo) *StandardFunctionPENCOLOR {
	this := &StandardFunctionPENCOLOR{D: d}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "PENCOLOR"
	this.MinParams = 0
	this.MaxParams = 0

	return this
}

func (this *StandardFunctionPENCOLOR) FunctionExecute(params *types.TokenList) error {

	/* vars */
	//	var b, a float64

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	this.Stack.Push(types.NewToken(types.WORD, utils.FormatFloat("", float64(apple2helpers.VECTOR(this.Interpreter).GetTurtle(this.D.Driver.GetTurtle()).PenColor))))

	return nil
}

func (this *StandardFunctionPENCOLOR) Syntax() string {

	/* vars */
	var result string

	result = "PENCOLOR word list"

	/* enforce non void return */
	return result

}
