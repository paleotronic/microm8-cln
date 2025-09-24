package logo

import (
	//	"strings"
//	"time"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
//	"paleotronic.com/core/interfaces"
//	"paleotronic.com/core/hardware/apple2helpers"
//	"paleotronic.com/runestring"
	"paleotronic.com/utils"
	"paleotronic.com/files"
)

type StandardFunctionWRITEPOS struct {
	dialect.CoreFunction
}

func (this *StandardFunctionWRITEPOS) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewStandardFunctionWRITEPOS(a int, b int, params types.TokenList) *StandardFunctionWRITEPOS {
	this := &StandardFunctionWRITEPOS{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "WRITEPOS"
	this.MinParams = 0
	this.MaxParams = 0

	return this
}

func (this *StandardFunctionWRITEPOS) FunctionExecute(params *types.TokenList) error {

	/* vars */
	//	var b, a float64

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	p, e := files.DOSWRITEPOS( files.GetPath(files.Writer), files.GetFilename(files.Writer) )

	this.Stack.Push( types.NewToken( types.NUMBER, utils.FormatFloat("", float64(p)) ) )

	return e
}

func (this *StandardFunctionWRITEPOS) Syntax() string {

	/* vars */
	var result string

	result = "WRITEPOS word list"

	/* enforce non void return */
	return result

}
