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

type StandardFunctionWRITER struct {
	dialect.CoreFunction
}

func (this *StandardFunctionWRITER) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewStandardFunctionWRITER(a int, b int, params types.TokenList) *StandardFunctionWRITER {
	this := &StandardFunctionWRITER{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "WRITER"
	this.MinParams = 0
	this.MaxParams = 0

	return this
}

func (this *StandardFunctionWRITER) FunctionExecute(params *types.TokenList) error {

	/* vars */
	//	var b, a float64

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	this.Stack.Push( types.NewToken( types.WORD, files.GetFilename(files.Writer) ) )

	return nil
}

func (this *StandardFunctionWRITER) Syntax() string {

	/* vars */
	var result string

	result = "WRITER word list"

	/* enforce non void return */
	return result

}
