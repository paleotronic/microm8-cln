package logo

import (
	//	"strings"
//	"time"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
//	"paleotronic.com/core/interfaces"
//	"paleotronic.com/core/hardware/apple2helpers"
//	"paleotronic.com/runestring"
//	"paleotronic.com/utils"
	"paleotronic.com/files"
)

type StandardFunctionREADER struct {
	dialect.CoreFunction
}

func (this *StandardFunctionREADER) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewStandardFunctionREADER(a int, b int, params types.TokenList) *StandardFunctionREADER {
	this := &StandardFunctionREADER{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "READER"
	this.MinParams = 0
	this.MaxParams = 0

	return this
}

func (this *StandardFunctionREADER) FunctionExecute(params *types.TokenList) error {

	/* vars */
	//	var b, a float64

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	this.Stack.Push( types.NewToken( types.WORD, files.GetFilename(files.Reader) ) )

	return nil
}

func (this *StandardFunctionREADER) Syntax() string {

	/* vars */
	var result string

	result = "READER word list"

	/* enforce non void return */
	return result

}
